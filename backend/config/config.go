package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type AppConfig struct {
	Username             string   `json:"username,omitempty"`
	Password             string   `json:"password,omitempty"`
	Favorites            []string `json:"favorites"`
	LastSelectedProfile  string   `json:"last_selected_profile,omitempty"`
	LastSelectedEndpoint string   `json:"last_selected_endpoint,omitempty"`
	Protocol             string   `json:"protocol,omitempty"`               // "udp" or "tcp", default "udp"
	ServerListSource     string   `json:"server_list_source,omitempty"`     // "downloaded" or "custom"
	LastServerListUpdate int64    `json:"last_server_list_update,omitempty"` // Unix timestamp in seconds
	CustomUdpPath        string   `json:"custom_udp_path,omitempty"`
	CustomTcpPath        string   `json:"custom_tcp_path,omitempty"`
}

type VpnSession struct {
	ProfileID        string  `json:"profile_id"`
	Endpoint         *string `json:"endpoint,omitempty"`
	ResolvedEndpoint *string `json:"resolved_endpoint,omitempty"`
	AssignedIP       *string `json:"assigned_ip,omitempty"`
	PublicIP         *string `json:"public_ip,omitempty"`
	ConnectedAt      int64   `json:"connected_at"`
	PID              *int    `json:"pid,omitempty"`
	MgmtPort         int     `json:"mgmt_port"`
}

func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "/tmp/fastvpn-gui"
	}
	return filepath.Join(configDir, "fastvpn-gui")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func GetSessionPath() string {
	return filepath.Join(GetConfigDir(), "session.json")
}

func GetProfilesBaseDir() string {
	return filepath.Join(GetConfigDir(), "profiles")
}

func GetProfilesDir(protocol string) string {
	if protocol == "" {
		protocol = "udp"
	}
	return filepath.Join(GetProfilesBaseDir(), strings.ToLower(protocol))
}

func LoadConfigFromPath(path string) (*AppConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &AppConfig{
			Favorites:        []string{},
			Protocol:         "udp",
			ServerListSource: "downloaded",
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		Favorites:        []string{},
		Protocol:         "udp",
		ServerListSource: "downloaded",
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return &AppConfig{
			Favorites:        []string{},
			Protocol:         "udp",
			ServerListSource: "downloaded",
		}, nil
	}
	if cfg.Favorites == nil {
		cfg.Favorites = []string{}
	}
	if cfg.Protocol == "" {
		cfg.Protocol = "udp"
	}
	if cfg.ServerListSource == "" {
		cfg.ServerListSource = "downloaded"
	}
	return cfg, nil
}

func SaveConfigToPath(path string, cfg *AppConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Write with 0600 permissions
	return os.WriteFile(path, data, 0600)
}

func LoadConfig() (*AppConfig, error) {
	return LoadConfigFromPath(GetConfigPath())
}

func SaveConfig(cfg *AppConfig) error {
	return SaveConfigToPath(GetConfigPath(), cfg)
}

func LoadSessionFromPath(path string) (*VpnSession, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sess VpnSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func SaveSessionToPath(path string, sess *VpnSession) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func ClearSessionFromPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func LoadSession() (*VpnSession, error) {
	return LoadSessionFromPath(GetSessionPath())
}

func SaveSession(sess *VpnSession) error {
	return SaveSessionToPath(GetSessionPath(), sess)
}

func ClearSession() error {
	return ClearSessionFromPath(GetSessionPath())
}

