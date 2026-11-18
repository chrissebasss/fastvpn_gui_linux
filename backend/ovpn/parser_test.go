package ovpn_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chrissebasss/fastvpn-gui-linux/backend/ovpn"
)

func TestLoadProfiles(t *testing.T) {
	tempDir := t.TempDir()
	sampleContent := `client
dev tun
proto udp
remote oslo.fastvpn.com 1194
remote 192.168.1.1 1194
resolv-retry infinite
nobind
ca ca.crt
`
	err := os.WriteFile(filepath.Join(tempDir, "NCVPN-NO-Oslo-UDP.ovpn"), []byte(sampleContent), 0644)
	if err != nil {
		t.Fatalf("failed to write sample profile: %v", err)
	}

	// Write Sweden (SE) and Albania (AL) to test A-Z sorting
	_ = os.WriteFile(filepath.Join(tempDir, "NCVPN-SE-Stockholm-UDP.ovpn"), []byte(sampleContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "NCVPN-AL-Tirana-UDP.ovpn"), []byte(sampleContent), 0644)

	profiles, err := ovpn.LoadProfiles(tempDir)
	if err != nil {
		t.Fatalf("failed to load profiles: %v", err)
	}

	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}

	// Should be sorted Albania -> Norway -> Sweden
	if profiles[0].Country != "Albania" || profiles[1].Country != "Norway" || profiles[2].Country != "Sweden" {
		t.Errorf("profiles not sorted A-Z by country: got %s, %s, %s",
			profiles[0].Country, profiles[1].Country, profiles[2].Country)
	}

	var foundOslo bool
	for _, p := range profiles {
		if p.ID == "NCVPN-NO-Oslo-UDP" {
			foundOslo = true
			if p.CountryCode != "NO" {
				t.Errorf("expected CountryCode 'NO', got %q", p.CountryCode)
			}
			if p.Country != "Norway" {
				t.Errorf("expected Country 'Norway', got %q", p.Country)
			}
			if p.City != "Oslo" {
				t.Errorf("expected City 'Oslo', got %q", p.City)
			}
			if len(p.Endpoints) == 0 {
				t.Errorf("expected endpoints in profile, got 0")
			}
		}
	}

	if !foundOslo {
		t.Error("expected to find NCVPN-NO-Oslo-UDP profile")
	}
}
