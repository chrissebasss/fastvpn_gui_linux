package ovpn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fastvpn/fastvpn-gui-linux/backend/config"
)

type ConnectionStatus struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type ProcessManager struct {
	mu               sync.Mutex
	status           ConnectionStatus
	cmd              *exec.Cmd
	authFile         string
	logs             []string
	isActive         atomic.Bool
	keepAliveOnExit  atomic.Bool
	onLogCallback    func(string)
	onStatusCallback func(ConnectionStatus)
}

func NewProcessManager() *ProcessManager {
	pm := &ProcessManager{
		status: ConnectionStatus{
			Status: "Disconnected",
		},
		logs: make([]string, 0, 500),
	}
	return pm
}

func (pm *ProcessManager) SetCallbacks(onLog func(string), onStatus func(ConnectionStatus)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onLogCallback = onLog
	pm.onStatusCallback = onStatus
}

func (pm *ProcessManager) IsActive() bool {
	return pm.isActive.Load()
}

func (pm *ProcessManager) IsKeepAlive() bool {
	return pm.keepAliveOnExit.Load()
}

func (pm *ProcessManager) GetStatus() ConnectionStatus {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If marked connected, check if process died outside
	if pm.status.Status == "Connected" && pm.cmd == nil {
		if !isProcessRunning("openvpn") {
			pm.isActive.Store(false)
			_ = config.ClearSession()
			pm.status = ConnectionStatus{Status: "Disconnected"}
		}
	}
	return pm.status
}

func (pm *ProcessManager) GetLogs() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	logsCopy := make([]string, len(pm.logs))
	copy(logsCopy, pm.logs)
	return logsCopy
}

func (pm *ProcessManager) AddLog(line string) {
	log.Println("[FastVPN]", line)

	pm.mu.Lock()
	if len(pm.logs) >= 500 {
		pm.logs = pm.logs[100:]
	}
	pm.logs = append(pm.logs, line)
	cb := pm.onLogCallback
	pm.mu.Unlock()

	if cb != nil {
		cb(line)
	}
}

func (pm *ProcessManager) setStatus(status ConnectionStatus) {
	pm.mu.Lock()
	pm.status = status
	cb := pm.onStatusCallback
	pm.mu.Unlock()

	if cb != nil {
		cb(status)
	}
}

func ParseServerPrefixFromLine(line string) string {
	if strings.Contains(line, "CN=") && (strings.Contains(line, "depth=0") || strings.Contains(line, "VERIFY X509NAME") || strings.Contains(line, "VERIFY OK")) {
		idx := strings.Index(line, "CN=")
		if idx != -1 {
			rest := line[idx+3:]
			for _, sep := range []string{",", "/", " ", "\r", "\n"} {
				if end := strings.Index(rest, sep); end != -1 {
					rest = rest[:end]
				}
			}
			rawCN := strings.TrimSpace(rest)
			prefix := strings.TrimSpace(strings.Split(rawCN, ".")[0])
			if prefix != "" && prefix != "VPN" {
				return prefix
			}
		}
	}
	return ""
}

func BuildOpenvpnArgs(ovpnPath string, endpoint *string, authFile string, mgmtPort int) []string {
	args := []string{
		"--config", ovpnPath,
	}

	if endpoint != nil && *endpoint != "" && *endpoint != "auto" {
		parts := strings.Split(*endpoint, ":")
		host := parts[0]
		port := "1194"
		if len(parts) > 1 {
			port = parts[1]
		}
		args = append(args, "--remote", host, port)
	}

	args = append(args,
		"--auth-user-pass", authFile,
		"--management", "127.0.0.1", strconv.Itoa(mgmtPort),
		"--writepid", "/tmp/fastvpn.pid",
		"--dev", "tun",
		"--nobind",
		"--verb", "3",
	)

	return args
}

func CheckPasswordlessPrivilege() bool {
	cmd := exec.Command("sudo", "-n", "openvpn", "--version")
	return cmd.Run() == nil
}

func (pm *ProcessManager) StartConnection(profile VpnProfile, endpoint *string, username, password string) error {
	pm.Disconnect()

	pm.isActive.Store(true)
	pm.keepAliveOnExit.Store(false)

	statusData := map[string]interface{}{
		"profile_id": profile.ID,
	}
	if endpoint != nil {
		statusData["endpoint"] = *endpoint
	}
	pm.setStatus(ConnectionStatus{
		Status: "Connecting",
		Data:   statusData,
	})

	// Write temporary credentials file with 0600
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/tmp"
	}
	authFile := filepath.Join(runtimeDir, fmt.Sprintf("fastvpn_auth_%d.tmp", os.Getpid()))
	authContent := fmt.Sprintf("%s\n%s\n", username, password)
	if err := os.WriteFile(authFile, []byte(authContent), 0600); err != nil {
		pm.isActive.Store(false)
		return fmt.Errorf("failed to write auth file: %w", err)
	}
	pm.authFile = authFile

	mgmtPort := 11194
	openvpnArgs := BuildOpenvpnArgs(profile.FullPath, endpoint, authFile, mgmtPort)

	var cmdBin string
	var cmdArgs []string
	if CheckPasswordlessPrivilege() {
		cmdBin = "sudo"
		cmdArgs = append([]string{"openvpn"}, openvpnArgs...)
	} else if _, err := exec.LookPath("pkexec"); err == nil {
		cmdBin = "pkexec"
		cmdArgs = append([]string{"openvpn"}, openvpnArgs...)
	} else {
		cmdBin = "sudo"
		cmdArgs = append([]string{"openvpn"}, openvpnArgs...)
	}

	cmd := exec.Command(cmdBin, cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		pm.isActive.Store(false)
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		pm.isActive.Store(false)
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	pm.AddLog(fmt.Sprintf("==> Launching: %s openvpn [args hidden]", cmdBin))

	if err := cmd.Start(); err != nil {
		pm.isActive.Store(false)
		_ = os.Remove(authFile)
		return fmt.Errorf("failed to start openvpn: %w", err)
	}

	pm.mu.Lock()
	pm.cmd = cmd
	pm.mu.Unlock()

	// Spawn stdout listener
	go func() {
		scanner := bufio.NewScanner(stdout)
		var assignedIP string
		var resolvedEndpoint string
		if endpoint != nil && *endpoint != "" && *endpoint != "auto" {
			resolvedEndpoint = *endpoint
		}

		for scanner.Scan() {
			line := scanner.Text()
			pm.AddLog(line)

			if prefix := ParseServerPrefixFromLine(line); prefix != "" {
				resolvedEndpoint = prefix
			}

			if strings.Contains(line, "Initialization Sequence Completed") {
				if resolvedEndpoint == "" {
					_, prefix := QueryMgmtForSession(mgmtPort)
					if prefix != "" {
						resolvedEndpoint = prefix
					}
				}

				now := time.Now().Unix()
				data := map[string]interface{}{
					"profile_id":   profile.ID,
					"connected_at": now,
				}
				if endpoint != nil {
					data["endpoint"] = *endpoint
				}
				if resolvedEndpoint != "" {
					data["resolved_endpoint"] = resolvedEndpoint
				}
				if assignedIP != "" {
					data["assigned_ip"] = assignedIP
				}

				pm.setStatus(ConnectionStatus{
					Status: "Connected",
					Data:   data,
				})
				pm.AddLog("==> VPN Connected Successfully!")
			} else if strings.Contains(line, "AUTH_FAILED") {
				pm.isActive.Store(false)
				pm.setStatus(ConnectionStatus{
					Status: "Error",
					Data: map[string]interface{}{
						"error": "Authentication failed: check username/password",
					},
				})
			} else if strings.Contains(line, "/sbin/ip addr add dev tun") || strings.Contains(line, "tun0") {
				for _, part := range strings.Fields(line) {
					if strings.Contains(part, ".") && len(strings.Split(part, ".")) == 4 {
						assignedIP = part
					}
				}
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
			pm.AddLog(fmt.Sprintf("[ERROR] stdout scanner: %v", err))
		}
	}()

	// Spawn stderr listener
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			pm.AddLog("[STDERR] " + scanner.Text())
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
			pm.AddLog(fmt.Sprintf("[ERROR] stderr scanner: %v", err))
		}
	}()

	return nil
}

func (pm *ProcessManager) Disconnect() error {
	pm.isActive.Store(false)
	pm.keepAliveOnExit.Store(false)
	_ = config.ClearSession()

	pm.setStatus(ConnectionStatus{
		Status: "Disconnecting",
	})
	pm.AddLog("==> Disconnecting VPN...")

	// Graceful shutdown via management socket
	conn, err := net.DialTimeout("tcp", "127.0.0.1:11194", 500*time.Millisecond)
	if err == nil {
		_, _ = conn.Write([]byte("signal SIGTERM\r\n"))
		_ = conn.Close()
	}

	pm.mu.Lock()
	if pm.cmd != nil && pm.cmd.Process != nil {
		_ = pm.cmd.Process.Kill()
		pm.cmd = nil
	}
	authFile := pm.authFile
	pm.authFile = ""
	pm.mu.Unlock()

	if authFile != "" {
		_ = os.Remove(authFile)
	}

	// Kill via PID file passwordlessly
	if _, err := os.Stat("/tmp/fastvpn.pid"); err == nil {
		_ = exec.Command("sudo", "-n", "pkill", "-SIGTERM", "-F", "/tmp/fastvpn.pid").Run()
		_ = exec.Command("pkill", "-SIGTERM", "-F", "/tmp/fastvpn.pid").Run()
		_ = os.Remove("/tmp/fastvpn.pid")
	}

	// Terminate any remaining openvpn instances passwordlessly
	_ = exec.Command("sudo", "-n", "pkill", "-SIGTERM", "-x", "openvpn").Run()
	_ = exec.Command("pkill", "-SIGTERM", "-x", "openvpn").Run()

	pm.setStatus(ConnectionStatus{
		Status: "Disconnected",
	})
	pm.AddLog("==> Disconnected.")
	return nil
}

func (pm *ProcessManager) PreserveSessionOnExit(publicIP *string) {
	pm.keepAliveOnExit.Store(true)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.status.Status == "Connected" && pm.status.Data != nil {
		profileID, _ := pm.status.Data["profile_id"].(string)
		var ep *string
		if e, ok := pm.status.Data["endpoint"].(string); ok {
			ep = &e
		}
		var resolved *string
		if r, ok := pm.status.Data["resolved_endpoint"].(string); ok {
			resolved = &r
		}
		var assigned *string
		if a, ok := pm.status.Data["assigned_ip"].(string); ok {
			assigned = &a
		}
		var connectedAt int64
		if c, ok := pm.status.Data["connected_at"].(int64); ok {
			connectedAt = c
		} else if cFloat, ok := pm.status.Data["connected_at"].(float64); ok {
			connectedAt = int64(cFloat)
		}

		sess := &config.VpnSession{
			ProfileID:        profileID,
			Endpoint:         ep,
			ResolvedEndpoint: resolved,
			AssignedIP:       assigned,
			PublicIP:         publicIP,
			ConnectedAt:      connectedAt,
			MgmtPort:         11194,
		}
		_ = config.SaveSession(sess)
	}
}

func (pm *ProcessManager) CheckRestoredSession() {
	pm.mu.Lock()
	if pm.status.Status != "Disconnected" {
		pm.mu.Unlock()
		return
	}
	pm.mu.Unlock()

	sess, err := config.LoadSession()
	if err == nil && sess != nil {
		if isProcessRunning("openvpn") {
			pm.isActive.Store(true)
			resolved := sess.ResolvedEndpoint
			if resolved == nil {
				_, p := QueryMgmtForSession(sess.MgmtPort)
				if p != "" {
					resolved = &p
				}
			}

			data := map[string]interface{}{
				"profile_id":   sess.ProfileID,
				"connected_at": sess.ConnectedAt,
			}
			if sess.Endpoint != nil {
				data["endpoint"] = *sess.Endpoint
			}
			if resolved != nil {
				data["resolved_endpoint"] = *resolved
			}
			if sess.AssignedIP != nil {
				data["assigned_ip"] = *sess.AssignedIP
			}
			if sess.PublicIP != nil {
				data["public_ip"] = *sess.PublicIP
			}

			pm.setStatus(ConnectionStatus{
				Status: "Connected",
				Data:   data,
			})
			pm.AddLog("==> Restored active VPN session from background")
			return
		}
		_ = config.ClearSession()
	}

	// Fallback: check if an orphaned fastvpn process is running
	if isProcessRunning("openvpn") {
		assigned, prefix := QueryMgmtForSession(11194)
		if prefix != "" || assigned != "" {
			pm.isActive.Store(true)
			data := map[string]interface{}{
				"profile_id":   "FastVPN",
				"connected_at": time.Now().Unix(),
			}
			if prefix != "" {
				data["resolved_endpoint"] = prefix
			}
			if assigned != "" {
				data["assigned_ip"] = assigned
			}
			pm.setStatus(ConnectionStatus{
				Status: "Connected",
				Data:   data,
			})
			pm.AddLog("==> Detected active background OpenVPN connection")
		}
	}
}

func QueryMgmtForSession(mgmtPort int) (assignedIP, serverPrefix string) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", mgmtPort), 600*time.Millisecond)
	if err != nil {
		return "", ""
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(1200 * time.Millisecond))
	reader := bufio.NewReader(conn)

	// Read banner
	_, _ = reader.ReadString('\n')

	// Query state
	_, _ = conn.Write([]byte("state\r\n"))
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.HasPrefix(line, "END") {
			break
		}
		if strings.Contains(line, "CONNECTED,SUCCESS") {
			parts := strings.Split(line, ",")
			if len(parts) > 3 && parts[3] != "" {
				assignedIP = parts[3]
			}
		}
	}

	// Query log 50
	_, _ = conn.Write([]byte("log 50\r\n"))
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.HasPrefix(line, "END") {
			break
		}
		if p := ParseServerPrefixFromLine(line); p != "" {
			serverPrefix = p
			break
		}
	}

	return assignedIP, serverPrefix
}

func isProcessRunning(name string) bool {
	if _, err := os.Stat("/tmp/fastvpn.pid"); err == nil {
		data, err := os.ReadFile("/tmp/fastvpn.pid")
		if err == nil {
			pidStr := strings.TrimSpace(string(data))
			if pid, err := strconv.Atoi(pidStr); err == nil {
				if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err == nil {
					return true
				}
			}
		}
	}

	out, err := exec.Command("pgrep", "-x", name).Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return true
	}
	return false
}

// Ensure unused io import check
var _ io.Reader = nil

