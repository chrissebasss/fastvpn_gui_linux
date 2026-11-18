package ovpn_test

import (
	"testing"
	"time"

	"github.com/fastvpn/fastvpn-gui-linux/backend/ovpn"
)

func TestParseServerPrefixFromLine(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{
			line:     "VERIFY OK: depth=1, C=US, ST=VPN, L=VPN, O=VPN, OU=VPN, CN=VPN, name=VPN, emailAddress=VPN",
			expected: "",
		},
		{
			line:     "1793727977,,VERIFY X509NAME OK: C=US, ST=VPN, L=VPN, O=VPN, OU=VPN, CN=osl-c05.wlvpn.com, emailAddress=VPN",
			expected: "osl-c05",
		},
		{
			line:     "VERIFY OK: depth=0, C=US, ST=VPN, L=VPN, O=VPN, OU=VPN, CN=tia-c08.wlvpn.com, emailAddress=VPN",
			expected: "tia-c08",
		},
		{
			line:     "VERIFY OK: depth=0, CN=adv-c01.wlvpn.com",
			expected: "adv-c01",
		},
	}

	for i, tt := range tests {
		got := ovpn.ParseServerPrefixFromLine(tt.line)
		if got != tt.expected {
			t.Errorf("test %d: for line %q, expected %q, got %q", i, tt.line, tt.expected, got)
		}
	}
}

func TestBuildOpenvpnArgs(t *testing.T) {
	args := ovpn.BuildOpenvpnArgs("/path/to/test.ovpn", nil, "/path/to/auth.tmp", 11194)
	if len(args) == 0 {
		t.Fatal("expected non-empty args")
	}

	hasConfig := false
	hasManagement := false
	hasWritePid := false
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) && args[i+1] == "/path/to/test.ovpn" {
			hasConfig = true
		}
		if arg == "--management" && i+2 < len(args) && args[i+1] == "127.0.0.1" && args[i+2] == "11194" {
			hasManagement = true
		}
		if arg == "--writepid" && i+1 < len(args) && args[i+1] == "/tmp/fastvpn.pid" {
			hasWritePid = true
		}
	}

	if !hasConfig {
		t.Error("expected --config /path/to/test.ovpn in args")
	}
	if !hasManagement {
		t.Error("expected --management 127.0.0.1 11194 in args")
	}
	if !hasWritePid {
		t.Error("expected --writepid /tmp/fastvpn.pid in args")
	}

	for _, arg := range args {
		if arg == "--management-query-passwords" {
			t.Error("args should not contain --management-query-passwords when using --auth-user-pass")
		}
	}

	// Test with explicit endpoint
	customEp := "osl-c01.vpn.wlvpn.com:1194"
	argsCustom := ovpn.BuildOpenvpnArgs("/path/to/test.ovpn", &customEp, "/path/to/auth.tmp", 11194)
	hasRemote := false
	for i, arg := range argsCustom {
		if arg == "--remote" && i+2 < len(argsCustom) && argsCustom[i+1] == "osl-c01.vpn.wlvpn.com" && argsCustom[i+2] == "1194" {
			hasRemote = true
		}
	}
	if !hasRemote {
		t.Error("expected --remote osl-c01.vpn.wlvpn.com 1194 in custom args")
	}
}

func TestDisconnectNoDeadlock(t *testing.T) {
	pm := ovpn.NewProcessManager()
	var logs []string
	var statuses []ovpn.ConnectionStatus

	pm.SetCallbacks(
		func(l string) {
			logs = append(logs, l)
		},
		func(s ovpn.ConnectionStatus) {
			statuses = append(statuses, s)
		},
	)

	// Call Disconnect - should never deadlock
	done := make(chan bool, 1)
	go func() {
		_ = pm.Disconnect()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("pm.Disconnect() deadlocked!")
	}

	if pm.GetStatus().Status != "Disconnected" {
		t.Errorf("expected status 'Disconnected', got %q", pm.GetStatus().Status)
	}
	if len(pm.GetLogs()) == 0 {
		t.Errorf("expected logs to be recorded")
	}
}

