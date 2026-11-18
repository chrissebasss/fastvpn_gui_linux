package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fastvpn/fastvpn-gui-linux/backend/config"
	"github.com/fastvpn/fastvpn-gui-linux/backend/ovpn"
)

type VpnService struct {
	mu         sync.RWMutex
	pm         *ovpn.ProcessManager
	profiles   []ovpn.VpnProfile
	profileMap map[string]ovpn.VpnProfile
	emitFunc   func(event string, data interface{})
}

func NewVpnService() *VpnService {
	pm := ovpn.NewProcessManager()
	cfg, _ := config.LoadConfig()

	svc := &VpnService{
		pm:         pm,
		profileMap: make(map[string]ovpn.VpnProfile),
	}

	_ = svc.reloadProfilesLocked(cfg)

	pm.SetCallbacks(
		func(logLine string) {
			svc.emitEvent("vpn:log", logLine)
		},
		func(status ovpn.ConnectionStatus) {
			svc.emitEvent("vpn:status", status)
		},
	)

	// Check if background connection is already active
	pm.CheckRestoredSession()

	return svc
}

func (s *VpnService) resolveActiveProfilesDir(cfg *config.AppConfig) string {
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil || cfg == nil {
			return config.GetProfilesDir("udp")
		}
	}

	proto := strings.ToLower(cfg.Protocol)
	if proto != "tcp" {
		proto = "udp"
	}

	if cfg.ServerListSource == "custom" {
		if proto == "tcp" && cfg.CustomTcpPath != "" {
			if info, err := os.Stat(cfg.CustomTcpPath); err == nil && info.IsDir() {
				return cfg.CustomTcpPath
			}
		} else if proto != "tcp" && cfg.CustomUdpPath != "" {
			if info, err := os.Stat(cfg.CustomUdpPath); err == nil && info.IsDir() {
				return cfg.CustomUdpPath
			}
		}
	}

	// Downloaded profiles directory (~/.config/fastvpn-gui/profiles/udp or profiles/tcp)
	return config.GetProfilesDir(proto)
}

func (s *VpnService) reloadProfilesLocked(cfg *config.AppConfig) error {
	dir := s.resolveActiveProfilesDir(cfg)
	profs, err := ovpn.LoadProfiles(dir)
	if err != nil {
		s.profiles = []ovpn.VpnProfile{}
		s.profileMap = make(map[string]ovpn.VpnProfile)
		return nil
	}
	s.profiles = profs
	s.profileMap = make(map[string]ovpn.VpnProfile, len(profs))
	for _, p := range profs {
		s.profileMap[p.ID] = p
	}
	return nil
}

func (s *VpnService) SetEventEmitter(emit func(event string, data interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emitFunc = emit
}

func (s *VpnService) emitEvent(event string, data interface{}) {
	s.mu.RLock()
	fn := s.emitFunc
	s.mu.RUnlock()

	if fn != nil {
		fn(event, data)
	}
}

func (s *VpnService) GetProfiles() ([]ovpn.VpnProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.profiles) == 0 {
		cfg, _ := config.LoadConfig()
		_ = s.reloadProfilesLocked(cfg)
	}

	res := make([]ovpn.VpnProfile, len(s.profiles))
	copy(res, s.profiles)
	return res, nil
}

func (s *VpnService) GetServerListInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.AppConfig{
			Protocol:         "udp",
			ServerListSource: "downloaded",
		}
	}

	udpCount := 0
	tcpCount := 0

	if cfg.ServerListSource == "custom" {
		if cfg.CustomUdpPath != "" {
			udpCount = ovpn.CountProfiles(cfg.CustomUdpPath)
		}
		if cfg.CustomTcpPath != "" {
			tcpCount = ovpn.CountProfiles(cfg.CustomTcpPath)
		}
	} else {
		baseDir := config.GetProfilesBaseDir()
		udpCount = ovpn.CountProfiles(filepath.Join(baseDir, "udp"))
		tcpCount = ovpn.CountProfiles(filepath.Join(baseDir, "tcp"))
	}

	hasProfiles := (cfg.Protocol == "tcp" && tcpCount > 0) || (cfg.Protocol != "tcp" && udpCount > 0)
	if !hasProfiles && (udpCount > 0 || tcpCount > 0) {
		hasProfiles = true
	}

	return map[string]interface{}{
		"protocol":                cfg.Protocol,
		"server_list_source":      cfg.ServerListSource,
		"last_server_list_update": cfg.LastServerListUpdate,
		"udp_count":               udpCount,
		"tcp_count":               tcpCount,
		"has_profiles":            hasProfiles,
		"custom_udp_path":         cfg.CustomUdpPath,
		"custom_tcp_path":         cfg.CustomTcpPath,
	}
}

func (s *VpnService) DownloadServerLists(customURL string) (map[string]interface{}, error) {
	baseDir := config.GetProfilesBaseDir()
	udpCount, tcpCount, err := ovpn.DownloadAndExtractServerLists(strings.TrimSpace(customURL), baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download server lists: %w", err)
	}

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.AppConfig{
			Protocol: "udp",
		}
	}
	cfg.ServerListSource = "downloaded"
	cfg.LastServerListUpdate = time.Now().Unix()
	_ = config.SaveConfig(cfg)

	s.mu.Lock()
	_ = s.reloadProfilesLocked(cfg)
	s.mu.Unlock()

	info := map[string]interface{}{
		"protocol":                cfg.Protocol,
		"server_list_source":      cfg.ServerListSource,
		"last_server_list_update": cfg.LastServerListUpdate,
		"udp_count":               udpCount,
		"tcp_count":               tcpCount,
		"has_profiles":            (cfg.Protocol == "tcp" && tcpCount > 0) || (cfg.Protocol != "tcp" && udpCount > 0),
		"custom_udp_path":         cfg.CustomUdpPath,
		"custom_tcp_path":         cfg.CustomTcpPath,
	}

	s.emitEvent("serverlist:updated", info)
	return info, nil
}

func (s *VpnService) SetProtocol(protocol string) ([]ovpn.VpnProfile, error) {
	proto := strings.ToLower(protocol)
	if proto != "tcp" {
		proto = "udp"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.AppConfig{}
	}
	cfg.Protocol = proto
	_ = config.SaveConfig(cfg)

	_ = s.reloadProfilesLocked(cfg)

	res := make([]ovpn.VpnProfile, len(s.profiles))
	copy(res, s.profiles)
	return res, nil
}

func (s *VpnService) SetCustomServerListPaths(udpPath string, tcpPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.AppConfig{Protocol: "udp"}
	}
	cfg.CustomUdpPath = strings.TrimSpace(udpPath)
	cfg.CustomTcpPath = strings.TrimSpace(tcpPath)
	cfg.ServerListSource = "custom"
	_ = config.SaveConfig(cfg)

	_ = s.reloadProfilesLocked(cfg)
	return nil
}

func (s *VpnService) SetServerListSource(source string) ([]ovpn.VpnProfile, error) {
	src := strings.ToLower(source)
	if src != "custom" {
		src = "downloaded"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.AppConfig{Protocol: "udp"}
	}
	cfg.ServerListSource = src
	_ = config.SaveConfig(cfg)

	_ = s.reloadProfilesLocked(cfg)

	res := make([]ovpn.VpnProfile, len(s.profiles))
	copy(res, s.profiles)
	return res, nil
}

func (s *VpnService) GetConfig() (*config.AppConfig, error) {
	return config.LoadConfig()
}

func (s *VpnService) SaveCredentials(username, password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.AppConfig{Favorites: []string{}}
	}
	cfg.Username = username
	cfg.Password = password
	return config.SaveConfig(cfg)
}

func normalizeFavoriteKey(id string) string {
	k := strings.TrimSuffix(id, "-UDP")
	k = strings.TrimSuffix(k, "-TCP")
	k = strings.TrimSuffix(k, "-udp")
	k = strings.TrimSuffix(k, "-tcp")
	return k
}

func (s *VpnService) ToggleFavorite(profileID string) ([]string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.AppConfig{Favorites: []string{}}
	}

	baseKey := normalizeFavoriteKey(profileID)
	found := false
	var updated []string
	for _, fav := range cfg.Favorites {
		if normalizeFavoriteKey(fav) == baseKey {
			found = true
		} else {
			updated = append(updated, fav)
		}
	}

	if !found {
		// Add both UDP and TCP variants so any direct ID lookup matches across protocols
		updated = append(updated, baseKey+"-UDP", baseKey+"-TCP")
	}

	cfg.Favorites = updated
	if err := config.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *VpnService) ConnectVPN(profileID string, endpoint *string) error {
	s.mu.RLock()
	profile, exists := s.profileMap[profileID]
	s.mu.RUnlock()

	if !exists {
		// Try reloading profiles
		_, _ = s.GetProfiles()
		s.mu.RLock()
		profile, exists = s.profileMap[profileID]
		s.mu.RUnlock()
		if !exists {
			return fmt.Errorf("profile %s not found", profileID)
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg.Username == "" || cfg.Password == "" {
		return errors.New("VPN credentials not configured. Please set username and password in Settings.")
	}

	// Save last selected profile & endpoint
	cfg.LastSelectedProfile = profileID
	if endpoint != nil {
		cfg.LastSelectedEndpoint = *endpoint
	} else {
		cfg.LastSelectedEndpoint = ""
	}
	_ = config.SaveConfig(cfg)

	return s.pm.StartConnection(profile, endpoint, cfg.Username, cfg.Password)
}

func (s *VpnService) DisconnectVPN() error {
	return s.pm.Disconnect()
}

func (s *VpnService) GetConnectionStatus() ovpn.ConnectionStatus {
	s.pm.CheckRestoredSession()
	return s.pm.GetStatus()
}

func (s *VpnService) GetLogs() []string {
	return s.pm.GetLogs()
}

func (s *VpnService) CheckPrivileges() bool {
	ok := ovpn.CheckPasswordlessPrivilege()
	log.Printf("[FastVPN Privilege] CheckPrivileges() -> %v", ok)
	return ok
}

func (s *VpnService) GetPrivilegeSetupCommand() string {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "scripts", "setup-privileges.sh"),
			filepath.Join(exeDir, "..", "scripts", "setup-privileges.sh"),
		)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".local", "share", "fastvpn-gui", "scripts", "setup-privileges.sh"),
		)
	}
	candidates = append(candidates,
		"/opt/fastvpn-gui/scripts/setup-privileges.sh",
		"./scripts/setup-privileges.sh",
	)

	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			if abs, err := filepath.Abs(cand); err == nil {
				return fmt.Sprintf("sudo bash %s", abs)
			}
			return fmt.Sprintf("sudo bash %s", cand)
		}
	}

	return "sudo bash scripts/setup-privileges.sh"
}

func (s *VpnService) RequestPrivilegeSetup() (bool, error) {
	log.Printf("[FastVPN Privilege] ========================================")
	log.Printf("[FastVPN Privilege] RequestPrivilegeSetup() invoked")
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		if u, err := user.Current(); err == nil {
			currentUser = u.Username
		}
	}

	candidates := []string{
		"./scripts/setup-privileges.sh",
		"../scripts/setup-privileges.sh",
		"/opt/fastvpn-gui/scripts/setup-privileges.sh",
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "scripts", "setup-privileges.sh"),
			filepath.Join(exeDir, "..", "scripts", "setup-privileges.sh"),
		)
	}

	scriptPath := ""
	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			scriptPath = cand
			break
		}
	}

	if scriptPath == "" {
		tmpPath := filepath.Join(os.TempDir(), "fastvpn_setup_privileges.sh")
		scriptContent := `#!/usr/bin/env bash
set -e
CURRENT_USER="${1:-${SUDO_USER:-$USER}}"
if [ -n "$PKEXEC_UID" ] && [ -z "$1" ]; then
  CURRENT_USER=$(id -un "$PKEXEC_UID" 2>/dev/null || echo "$CURRENT_USER")
fi
OPENVPN_BIN=$(type -P openvpn || which openvpn 2>/dev/null || echo "/usr/bin/openvpn")
PKILL_BIN=$(type -P pkill || which pkill 2>/dev/null || echo "/usr/bin/pkill")
KILL_BIN=$(type -P kill || which kill 2>/dev/null || echo "/usr/bin/kill")
KILLALL_BIN=$(type -P killall || which killall 2>/dev/null || echo "/usr/bin/killall")
[[ "$OPENVPN_BIN" != /* ]] && OPENVPN_BIN="/usr/bin/openvpn"
[[ "$PKILL_BIN" != /* ]] && PKILL_BIN="/usr/bin/pkill"
[[ "$KILL_BIN" != /* ]] && KILL_BIN="/usr/bin/kill"
[[ "$KILLALL_BIN" != /* ]] && KILLALL_BIN="/usr/bin/killall"

POLKIT_RULES_DIR="/etc/polkit-1/rules.d"
if [ -d "$POLKIT_RULES_DIR" ]; then
  cat <<EOF > "$POLKIT_RULES_DIR/50-fastvpn-openvpn.rules"
polkit.addRule(function(action, subject) {
    if (action.id == "org.freedesktop.policykit.exec" &&
        (action.lookup("program") == "$OPENVPN_BIN" ||
         action.lookup("program") == "$PKILL_BIN" ||
         action.lookup("program") == "$KILL_BIN" ||
         action.lookup("program") == "$KILLALL_BIN") &&
        subject.user == "$CURRENT_USER") {
        return polkit.Result.YES;
    }
});
EOF
  chmod 644 "$POLKIT_RULES_DIR/50-fastvpn-openvpn.rules"
fi

SUDOERS_FILE="/etc/sudoers.d/fastvpn-openvpn"
cat <<EOF > "$SUDOERS_FILE"
$CURRENT_USER ALL=(ALL) NOPASSWD: $OPENVPN_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $PKILL_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $KILL_BIN
$CURRENT_USER ALL=(ALL) NOPASSWD: $KILLALL_BIN
EOF
chmod 0440 "$SUDOERS_FILE"
if visudo -cf "$SUDOERS_FILE" >/dev/null 2>&1; then
  exit 0
else
  rm -f "$SUDOERS_FILE"
  exit 1
fi
`
		if err := os.WriteFile(tmpPath, []byte(scriptContent), 0755); err != nil {
			return false, fmt.Errorf("failed to write setup script: %w", err)
		}
		scriptPath = tmpPath
		defer os.Remove(tmpPath)
	}

	if absP, err := filepath.Abs(scriptPath); err == nil {
		scriptPath = absP
	}

	log.Printf("[FastVPN Privilege] Target script: '%s', User: '%s'", scriptPath, currentUser)
	setupCmdStr := fmt.Sprintf("bash '%s' '%s'", scriptPath, currentUser)

	// 1. Try KDE native kdesu (Plasma 6 / Plasma 5)
	kdesuPaths := []string{"/usr/lib/kf6/kdesu", "/usr/lib/kdesu", "kdesu"}
	for _, kp := range kdesuPaths {
		if bin, err := exec.LookPath(kp); err == nil {
			log.Printf("[FastVPN Privilege] Attempting kdesu at '%s' with cmd: %s", bin, setupCmdStr)
			cmd := exec.Command(bin, "-i", "fastvpn-gui", "-c", setupCmdStr)
			cmd.Env = os.Environ()
			out, err := cmd.CombinedOutput()
			log.Printf("[FastVPN Privilege] kdesu result: err=%v, output=%s", err, string(out))
			if err == nil && s.CheckPrivileges() {
				log.Printf("[FastVPN Privilege] Elevation verified via kdesu! Success.")
				log.Printf("[FastVPN Privilege] ========================================")
				return true, nil
			}
		}
	}

	// 2. Try kdialog native Qt password prompt with sudo -S
	if kdialogBin, err := exec.LookPath("kdialog"); err == nil {
		log.Printf("[FastVPN Privilege] Attempting kdialog at '%s'", kdialogBin)
		passCmd := exec.Command(kdialogBin, "--title", "FastVPN Authorization", "--password", "FastVPN requires authorization to configure passwordless OpenVPN permissions:")
		passCmd.Env = os.Environ()
		passOut, err := passCmd.Output()
		log.Printf("[FastVPN Privilege] kdialog finished: err=%v, enteredBytes=%d", err, len(passOut))
		if err == nil && len(passOut) > 0 {
			pass := strings.TrimRight(string(passOut), "\r\n")
			sudoCmd := exec.Command("sudo", "-S", "-p", "", "bash", scriptPath, currentUser)
			sudoCmd.Env = os.Environ()
			sudoCmd.Stdin = strings.NewReader(pass + "\n")
			out, sudoErr := sudoCmd.CombinedOutput()
			log.Printf("[FastVPN Privilege] sudo execution: err=%v, output=%s", sudoErr, string(out))
			if sudoErr == nil && s.CheckPrivileges() {
				log.Printf("[FastVPN Privilege] Elevation verified via kdialog + sudo! Success.")
				log.Printf("[FastVPN Privilege] ========================================")
				return true, nil
			}
		}
	}

	// 3. Try standard PolicyKit pkexec
	if pkexecBin, err := exec.LookPath("pkexec"); err == nil {
		log.Printf("[FastVPN Privilege] Attempting pkexec at '%s'", pkexecBin)
		cmd := exec.Command(pkexecBin, "bash", scriptPath, currentUser)
		cmd.Env = os.Environ()
		out, err := cmd.CombinedOutput()
		log.Printf("[FastVPN Privilege] pkexec result: err=%v, output=%s", err, string(out))
		if err == nil && s.CheckPrivileges() {
			log.Printf("[FastVPN Privilege] Elevation verified via pkexec! Success.")
			log.Printf("[FastVPN Privilege] ========================================")
			return true, nil
		}
	}

	// 4. Try zenity (GNOME / GTK) password prompt
	if zenityBin, err := exec.LookPath("zenity"); err == nil {
		log.Printf("[FastVPN Privilege] Attempting zenity at '%s'", zenityBin)
		passCmd := exec.Command(zenityBin, "--password", "--title=FastVPN Authorization")
		passCmd.Env = os.Environ()
		passOut, err := passCmd.Output()
		log.Printf("[FastVPN Privilege] zenity finished: err=%v, enteredBytes=%d", err, len(passOut))
		if err == nil && len(passOut) > 0 {
			pass := strings.TrimRight(string(passOut), "\r\n")
			sudoCmd := exec.Command("sudo", "-S", "-p", "", "bash", scriptPath, currentUser)
			sudoCmd.Env = os.Environ()
			sudoCmd.Stdin = strings.NewReader(pass + "\n")
			out, sudoErr := sudoCmd.CombinedOutput()
			log.Printf("[FastVPN Privilege] sudo execution: err=%v, output=%s", sudoErr, string(out))
			if sudoErr == nil && s.CheckPrivileges() {
				log.Printf("[FastVPN Privilege] Elevation verified via zenity + sudo! Success.")
				log.Printf("[FastVPN Privilege] ========================================")
				return true, nil
			}
		}
	}

	log.Printf("[FastVPN Privilege] All elevation methods completed. CheckPrivileges() = %v", s.CheckPrivileges())
	log.Printf("[FastVPN Privilege] ========================================")
	if s.CheckPrivileges() {
		return true, nil
	}
	return false, errors.New("authorization was cancelled or failed")
}

func (s *VpnService) ForceExit(disconnect bool, publicIP *string) {
	if disconnect {
		_ = s.pm.Disconnect()
	} else {
		s.pm.PreserveSessionOnExit(publicIP)
	}
}

// ParseIPFromBody extracts a clean IP address from an HTTP response body,
// handling raw IP text, JSON payloads (e.g. {"ip": "93.89.122.161", "type": "IPv4"}),
// and extracting valid IPv4/IPv6 addresses.
func ParseIPFromBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}

	// 1. If JSON payload, try unmarshaling into common IP keys
	if strings.HasPrefix(trimmed, "{") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
			for _, key := range []string{"ip", "ip_address", "query", "origin"} {
				if val, ok := obj[key]; ok {
					if strVal, isStr := val.(string); isStr {
						candidate := strings.TrimSpace(strVal)
						if net.ParseIP(candidate) != nil {
							return candidate
						}
					}
				}
			}
		}
	}

	// 2. Direct IP check (plain text response)
	if net.ParseIP(trimmed) != nil {
		return trimmed
	}

	// 3. Fallback: regex search for IPv4 pattern
	re := regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	if match := re.FindString(trimmed); match != "" {
		if net.ParseIP(match) != nil {
			return match
		}
	}

	return ""
}

func (s *VpnService) GetPublicIP() (string, error) {
	urls := []string{
		"https://wmip.dev/api/json",
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{
		Timeout: 3500 * time.Millisecond,
	}

	for _, u := range urls {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			cancel()
			continue
		}
		req.Header.Set("User-Agent", "curl/7.88.1")

		resp, err := client.Do(req)
		if err != nil {
			cancel()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		cancel()

		if err == nil && resp.StatusCode == http.StatusOK {
			ip := ParseIPFromBody(body)
			if ip != "" {
				return ip, nil
			}
		}
	}

	return "", errors.New("failed to determine public IP from upstream providers")
}
