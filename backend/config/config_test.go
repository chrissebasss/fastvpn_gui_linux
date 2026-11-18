package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastvpn/fastvpn-gui-linux/backend/config"
)

func TestConfigPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fastvpn-cfg-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "config.json")
	cfg := &config.AppConfig{
		Username:  "testuser",
		Password:  "testpass",
		Favorites: []string{"NCVPN-NO-Oslo-UDP"},
	}

	if err := config.SaveConfigToPath(cfgPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file permissions 0600
	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 permissions, got %04o", perm)
	}

	loaded, err := config.LoadConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Username != "testuser" || loaded.Password != "testpass" || len(loaded.Favorites) != 1 {
		t.Errorf("config mismatch: %+v", loaded)
	}
}

func TestSessionPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fastvpn-sess-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sessPath := filepath.Join(tmpDir, "session.json")
	resolved := "osl-c05"
	pubIP := "84.247.50.239"
	sess := &config.VpnSession{
		ProfileID:        "NCVPN-NO-Oslo-UDP",
		ResolvedEndpoint: &resolved,
		PublicIP:         &pubIP,
		ConnectedAt:      1793727978,
		MgmtPort:         11194,
	}

	if err := config.SaveSessionToPath(sessPath, sess); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	loaded, err := config.LoadSessionFromPath(sessPath)
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	if loaded.ProfileID != "NCVPN-NO-Oslo-UDP" || loaded.ResolvedEndpoint == nil || *loaded.ResolvedEndpoint != "osl-c05" {
		t.Errorf("session mismatch: %+v", loaded)
	}

	if err := config.ClearSessionFromPath(sessPath); err != nil {
		t.Fatalf("failed to clear session: %v", err)
	}

	if _, err := os.Stat(sessPath); !os.IsNotExist(err) {
		t.Errorf("expected session file to be deleted")
	}
}

