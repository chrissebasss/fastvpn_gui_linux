package ovpn

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultServerListURL = "https://vpn.ncapi.io/groupedServerList.zip"

// DownloadAndExtractServerLists downloads the grouped server list zip archive
// (or reads a local zip file / file:// URL) and safely extracts its 'udp' and 'tcp' directories into destBaseDir.
// Returns (udpCount, tcpCount, error).
func DownloadAndExtractServerLists(url string, destBaseDir string) (int, int, error) {
	if url == "" {
		url = DefaultServerListURL
	}

	var bodyBytes []byte

	// 1. Check if url is a local file or file:// URI
	if strings.HasPrefix(url, "file://") || strings.HasPrefix(url, "/") || strings.HasPrefix(url, "~") {
		filePath := strings.TrimPrefix(url, "file://")
		if strings.HasPrefix(filePath, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				filePath = filepath.Join(home, strings.TrimPrefix(filePath, "~"))
			}
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read local zip file %q: %w", filePath, err)
		}
		bodyBytes = data
	} else {
		client := &http.Client{
			Timeout: 45 * time.Second,
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to create download request: %w", err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) FastVPN-Linux/1.0")

		resp, err := client.Do(req)
		if err != nil {
			// If upstream server is temporarily unreachable or refused connection, check if groupedServerList.zip exists in ~/Downloads
			if home, herr := os.UserHomeDir(); herr == nil {
				fallbackZip := filepath.Join(home, "Downloads", "groupedServerList.zip")
				if data, rerr := os.ReadFile(fallbackZip); rerr == nil {
					return ExtractZipBytes(data, destBaseDir)
				}
			}
			return 0, 0, fmt.Errorf("failed to download from %s: %w", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, 0, fmt.Errorf("download from %s returned HTTP status %d", url, resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read response body: %w", err)
		}
		bodyBytes = data
	}

	return ExtractZipBytes(bodyBytes, destBaseDir)
}

// ExtractZipBytes extracts zip archive bytes into destBaseDir, ensuring zip slip protection.
func ExtractZipBytes(zipData []byte, destBaseDir string) (int, int, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid zip archive: %w", err)
	}

	cleanBase := filepath.Clean(destBaseDir)
	if err := os.MkdirAll(cleanBase, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create base profiles directory: %w", err)
	}

	udpCount := 0
	tcpCount := 0

	for _, file := range zipReader.File {
		name := filepath.Clean(file.Name)

		// Disallow directory traversal (Zip Slip protection)
		if strings.HasPrefix(name, "..") || strings.Contains(name, "/../") || strings.Contains(name, `\..\`) {
			continue
		}

		// Ensure file is within destBaseDir
		targetPath := filepath.Join(cleanBase, name)
		if !strings.HasPrefix(filepath.Clean(targetPath), cleanBase) {
			continue
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return 0, 0, fmt.Errorf("failed to create directory %q: %w", targetPath, err)
			}
			continue
		}

		// Only extract files into udp/ or tcp/ directories, specifically .ovpn files
		lowerName := strings.ToLower(name)
		if !strings.HasPrefix(lowerName, "udp"+string(filepath.Separator)) &&
			!strings.HasPrefix(lowerName, "udp/") &&
			!strings.HasPrefix(lowerName, "tcp"+string(filepath.Separator)) &&
			!strings.HasPrefix(lowerName, "tcp/") {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return 0, 0, fmt.Errorf("failed to create parent dir for %q: %w", targetPath, err)
		}

		rc, err := file.Open()
		if err != nil {
			return 0, 0, fmt.Errorf("failed to open zipped file %q: %w", file.Name, err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			rc.Close()
			return 0, 0, fmt.Errorf("failed to create output file %q: %w", targetPath, err)
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if copyErr != nil {
			return 0, 0, fmt.Errorf("failed to extract file %q: %w", targetPath, copyErr)
		}

		if strings.HasSuffix(lowerName, ".ovpn") {
			if strings.HasPrefix(lowerName, "udp") {
				udpCount++
			} else if strings.HasPrefix(lowerName, "tcp") {
				tcpCount++
			}
		}
	}

	return udpCount, tcpCount, nil
}

// CountProfiles returns the number of .ovpn profiles found in dir.
func CountProfiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".ovpn") {
			count++
		}
	}
	return count
}

// HasProfiles checks if either udp or tcp directories under baseDir contain profiles.
func HasProfiles(baseDir string) bool {
	return CountProfiles(filepath.Join(baseDir, "udp")) > 0 ||
		CountProfiles(filepath.Join(baseDir, "tcp")) > 0
}

