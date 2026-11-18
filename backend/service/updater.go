package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseNotes   string `json:"release_notes"`
	ReleaseURL     string `json:"release_url"`
	PublishedAt    string `json:"published_at"`
}

type githubReleaseResponse struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
}

var (
	updateCacheMu   sync.Mutex
	cachedUpdate    *UpdateInfo
	lastUpdateCheck time.Time
	cacheTTL        = 30 * time.Minute
)

// CompareSemVer compares two SemVer versions (e.g., "1.1.0", "v1.2.0").
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func CompareSemVer(v1, v2 string) int {
	clean1 := strings.TrimPrefix(strings.TrimSpace(v1), "v")
	clean2 := strings.TrimPrefix(strings.TrimSpace(v2), "v")

	parts1, pre1 := splitVersionAndPrerelease(clean1)
	parts2, pre2 := splitVersionAndPrerelease(clean2)

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		num1 := 0
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		num2 := 0
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	// SemVer spec: normal version has higher precedence than pre-release version
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 != "" && pre2 == "" {
		return -1
	}
	if pre1 < pre2 {
		return -1
	}
	if pre1 > pre2 {
		return 1
	}

	return 0
}

func splitVersionAndPrerelease(v string) ([]string, string) {
	var prerelease string
	base := v
	if idx := strings.Index(v, "-"); idx != -1 {
		base = v[:idx]
		prerelease = v[idx+1:]
	}
	parts := strings.Split(base, ".")
	return parts, prerelease
}

func fetchReleaseFromURL(apiURL, currentVersion string, timeout time.Duration) (*UpdateInfo, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("User-Agent", "FastVPN-Linux-Client/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(body))
	}

	var release githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := CompareSemVer(currentVersion, latestVersion) < 0

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
		ReleaseNotes:   release.Body,
		ReleaseURL:     release.HTMLURL,
		PublishedAt:    release.PublishedAt,
	}, nil
}

// CheckGitHubUpdate checks for updates against GitHub Releases.
// Caches results in memory for cacheTTL unless force is true.
func CheckGitHubUpdate(repoOwner, repoName, currentVersion string, force bool) (*UpdateInfo, error) {
	updateCacheMu.Lock()
	if !force && cachedUpdate != nil && time.Since(lastUpdateCheck) < cacheTTL {
		cached := *cachedUpdate
		updateCacheMu.Unlock()
		return &cached, nil
	}
	updateCacheMu.Unlock()

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	info, err := fetchReleaseFromURL(apiURL, currentVersion, 5*time.Second)
	if err != nil {
		log.Printf("[FastVPN Updater] Check failed: %v", err)
		return nil, err
	}

	updateCacheMu.Lock()
	cachedUpdate = info
	lastUpdateCheck = time.Now()
	updateCacheMu.Unlock()

	return info, nil
}
