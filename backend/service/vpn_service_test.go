package service_test

import (
	"testing"

	"github.com/chrissebasss/fastvpn-gui-linux/backend/service"
)

func TestServiceInitialization(t *testing.T) {
	svc := service.NewVpnService()
	if svc == nil {
		t.Fatal("expected non-nil VpnService")
	}

	profiles, err := svc.GetProfiles()
	if err != nil {
		t.Fatalf("failed to get profiles: %v", err)
	}
	_ = profiles

	cfg, err := svc.GetConfig()
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	status := svc.GetConnectionStatus()
	if status.Status == "" {
		t.Fatal("expected non-empty status string")
	}
}

func TestPublicIPDetection(t *testing.T) {
	svc := service.NewVpnService()
	ip, err := svc.GetPublicIP()
	if err != nil {
		t.Logf("network offline or restricted: %v", err)
		return
	}
	if len(ip) < 7 {
		t.Errorf("invalid IP format: %s", ip)
	}
}

func TestParseIPFromBody(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "wmip json minified",
			input:    `{"ip":"93.89.122.161","type":"IPv4"}`,
			expected: "93.89.122.161",
		},
		{
			name:     "wmip json spaced",
			input:    `{ "ip": "93.89.122.161", "type": "IPv4" }`,
			expected: "93.89.122.161",
		},
		{
			name:     "plain text with trailing newline",
			input:    "1.1.1.1\n",
			expected: "1.1.1.1",
		},
		{
			name:     "plain text with spaces",
			input:    "   8.8.8.8   ",
			expected: "8.8.8.8",
		},
		{
			name:     "json with query key",
			input:    `{"query":"104.244.42.1"}`,
			expected: "104.244.42.1",
		},
		{
			name:     "html error response",
			input:    "<html><body>502 Bad Gateway</body></html>",
			expected: "",
		},
		{
			name:     "empty response",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := service.ParseIPFromBody([]byte(tc.input))
			if res != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, res)
			}
		})
	}
}

func TestServerListMethods(t *testing.T) {
	svc := service.NewVpnService()
	info := svc.GetServerListInfo()
	if info == nil {
		t.Fatal("expected non-nil server list info")
	}

	if _, ok := info["protocol"]; !ok {
		t.Error("expected protocol key in info")
	}
	if _, ok := info["has_profiles"]; !ok {
		t.Error("expected has_profiles key in info")
	}

	// Test protocol switching
	profs, err := svc.SetProtocol("tcp")
	if err != nil {
		t.Fatalf("SetProtocol(tcp) failed: %v", err)
	}
	info = svc.GetServerListInfo()
	if info["protocol"] != "tcp" {
		t.Errorf("expected protocol to be tcp, got %v", info["protocol"])
	}

	// Switch back to udp
	profs, err = svc.SetProtocol("udp")
	if err != nil {
		t.Fatalf("SetProtocol(udp) failed: %v", err)
	}
	_ = profs
	info = svc.GetServerListInfo()
	if info["protocol"] != "udp" {
		t.Errorf("expected protocol to be udp, got %v", info["protocol"])
	}
}

func TestDownloadServerLists(t *testing.T) {
	svc := service.NewVpnService()
	var emitted bool
	svc.SetEventEmitter(func(event string, data interface{}) {
		if event == "serverlist:updated" {
			emitted = true
		}
	})

	info, err := svc.DownloadServerLists("")
	if err != nil {
		t.Fatalf("DownloadServerLists failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil info from DownloadServerLists")
	}
	if !emitted {
		t.Errorf("expected serverlist:updated event to be emitted")
	}
	if info["udp_count"].(int) == 0 {
		t.Errorf("expected >0 UDP profiles downloaded")
	}
	if info["tcp_count"].(int) == 0 {
		t.Errorf("expected >0 TCP profiles downloaded")
	}
}

func TestConnectFlow(t *testing.T) {
	svc := service.NewVpnService()
	profs, err := svc.GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles failed: %v", err)
	}
	t.Logf("Loaded %d profiles", len(profs))
	if len(profs) == 0 {
		t.Skip("no profiles loaded")
	}

	pID := profs[0].ID
	t.Logf("Connecting to profile %s...", pID)
	ep := "auto"
	err = svc.ConnectVPN(pID, &ep)
	t.Logf("ConnectVPN result: %v", err)
	_ = svc.DisconnectVPN()
}

func TestFavoriteSyncAcrossProtocols(t *testing.T) {
	svc := service.NewVpnService()
	testKey := "TEST-LOC-ALPHA-TCP"
	// Ensure test key is cleared first
	_, _ = svc.ToggleFavorite(testKey)
	cfg, _ := svc.GetConfig()
	for _, f := range cfg.Favorites {
		if f == "TEST-LOC-ALPHA-TCP" || f == "TEST-LOC-ALPHA-UDP" {
			_, _ = svc.ToggleFavorite(testKey)
			break
		}
	}

	// Toggle TCP profile as favorite
	favs, err := svc.ToggleFavorite("TEST-LOC-ALPHA-TCP")
	if err != nil {
		t.Fatalf("ToggleFavorite failed: %v", err)
	}

	hasUDP := false
	hasTCP := false
	for _, f := range favs {
		if f == "TEST-LOC-ALPHA-UDP" {
			hasUDP = true
		}
		if f == "TEST-LOC-ALPHA-TCP" {
			hasTCP = true
		}
	}
	if !hasUDP || !hasTCP {
		t.Errorf("expected both UDP and TCP variants in favorites, got %v", favs)
	}

	// Toggle again using UDP variant - should remove both
	favsAfter, err := svc.ToggleFavorite("TEST-LOC-ALPHA-UDP")
	if err != nil {
		t.Fatalf("ToggleFavorite second call failed: %v", err)
	}
	for _, f := range favsAfter {
		if f == "TEST-LOC-ALPHA-UDP" || f == "TEST-LOC-ALPHA-TCP" {
			t.Errorf("expected TEST-LOC-ALPHA to be removed, but found %s", f)
		}
	}
}
