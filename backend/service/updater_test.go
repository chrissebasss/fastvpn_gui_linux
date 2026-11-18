package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCompareSemVer(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int // -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
	}{
		{"1.1.0", "1.2.0", -1},
		{"1.2.0", "1.1.0", 1},
		{"1.1.0", "1.1.0", 0},
		{"v1.1.0", "1.1.0", 0},
		{"1.1.0", "v1.1.0", 0},
		{"v1.1.0", "v1.2.0", -1},
		{"1.9.0", "1.10.0", -1},
		{"1.10.0", "1.9.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"2.0.0", "1.99.99", 1},
		{"1.1.0-beta", "1.1.0", -1},
		{"1.1.0", "1.1.0-beta", 1},
	}

	for _, tt := range tests {
		got := CompareSemVer(tt.v1, tt.v2)
		if got != tt.expected {
			t.Errorf("CompareSemVer(%q, %q) = %d; want %d", tt.v1, tt.v2, got, tt.expected)
		}
	}
}

func TestCheckGitHubUpdateSuccess(t *testing.T) {
	mockRelease := githubReleaseResponse{
		TagName:     "v1.2.0",
		Name:        "FastVPN GUI v1.2.0",
		Body:        "### Added\n- In-app update notifications\n- Debian .deb packaging",
		HTMLURL:     "https://github.com/chrissebasss/fastvpn_gui_linux/releases/tag/v1.2.0",
		PublishedAt: "2026-11-18T14:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	info, err := fetchReleaseFromURL(server.URL, "1.1.0", 2*time.Second)
	if err != nil {
		t.Fatalf("fetchReleaseFromURL failed: %v", err)
	}

	if !info.HasUpdate {
		t.Errorf("expected HasUpdate to be true for current 1.1.0 vs latest 1.2.0")
	}
	if info.LatestVersion != "1.2.0" {
		t.Errorf("expected LatestVersion to be 1.2.0, got %q", info.LatestVersion)
	}
	if info.CurrentVersion != "1.1.0" {
		t.Errorf("expected CurrentVersion to be 1.1.0, got %q", info.CurrentVersion)
	}
	if info.ReleaseNotes != mockRelease.Body {
		t.Errorf("expected ReleaseNotes to match, got %q", info.ReleaseNotes)
	}
	if info.ReleaseURL != mockRelease.HTMLURL {
		t.Errorf("expected ReleaseURL to match, got %q", info.ReleaseURL)
	}
}

func TestCheckGitHubUpdateNoUpdate(t *testing.T) {
	mockRelease := githubReleaseResponse{
		TagName:     "v1.1.0",
		Name:        "FastVPN GUI v1.1.0",
		Body:        "Current release notes",
		HTMLURL:     "https://github.com/chrissebasss/fastvpn_gui_linux/releases/tag/v1.1.0",
		PublishedAt: "2026-11-18T12:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	info, err := fetchReleaseFromURL(server.URL, "1.1.0", 2*time.Second)
	if err != nil {
		t.Fatalf("fetchReleaseFromURL failed: %v", err)
	}

	if info.HasUpdate {
		t.Errorf("expected HasUpdate to be false for identical versions (1.1.0 vs 1.1.0)")
	}
}

func TestCheckGitHubUpdateErrorHandling(t *testing.T) {
	// Server returning 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := fetchReleaseFromURL(server.URL, "1.1.0", 2*time.Second)
	if err == nil {
		t.Errorf("expected error for 404 response, got nil")
	}
}
