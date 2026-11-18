package ovpn_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastvpn/fastvpn-gui-linux/backend/ovpn"
)

func createTestZip(t *testing.T) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	files := []struct {
		Name    string
		Content string
	}{
		{"udp/test-us-udp.ovpn", "client\ndev tun\nproto udp\nremote 1.2.3.4 1194\n"},
		{"udp/test-uk-udp.ovpn", "client\ndev tun\nproto udp\nremote 5.6.7.8 1194\n"},
		{"tcp/test-de-tcp.ovpn", "client\ndev tun\nproto tcp\nremote 9.10.11.12 443\n"},
		{"other/ignored.txt", "should be ignored"},
		{"../slip.ovpn", "should be prevented by zip slip protection"},
	}

	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			t.Fatalf("failed to create zip file entry: %v", err)
		}
		_, err = f.Write([]byte(file.Content))
		if err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func TestExtractZipBytes(t *testing.T) {
	tempDir := t.TempDir()
	zipData := createTestZip(t)

	udpCount, tcpCount, err := ovpn.ExtractZipBytes(zipData, tempDir)
	if err != nil {
		t.Fatalf("ExtractZipBytes failed: %v", err)
	}

	if udpCount != 2 {
		t.Errorf("expected 2 UDP profiles, got %d", udpCount)
	}
	if tcpCount != 1 {
		t.Errorf("expected 1 TCP profile, got %d", tcpCount)
	}

	// Verify files on disk
	if !ovpn.HasProfiles(tempDir) {
		t.Errorf("expected HasProfiles to be true")
	}

	udpOnDisk := ovpn.CountProfiles(filepath.Join(tempDir, "udp"))
	if udpOnDisk != 2 {
		t.Errorf("expected 2 udp files on disk, got %d", udpOnDisk)
	}

	tcpOnDisk := ovpn.CountProfiles(filepath.Join(tempDir, "tcp"))
	if tcpOnDisk != 1 {
		t.Errorf("expected 1 tcp file on disk, got %d", tcpOnDisk)
	}

	// Verify zip slip file was not created outside tempDir
	slipPath := filepath.Join(filepath.Dir(tempDir), "slip.ovpn")
	if _, err := os.Stat(slipPath); !os.IsNotExist(err) {
		t.Errorf("zip slip vulnerability detected: file exists at %s", slipPath)
	}
}

func TestCountProfilesEmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	if ovpn.CountProfiles(tempDir) != 0 {
		t.Errorf("expected 0 profiles in empty dir")
	}
	if ovpn.HasProfiles(tempDir) {
		t.Errorf("expected HasProfiles to be false for empty dir")
	}
}

