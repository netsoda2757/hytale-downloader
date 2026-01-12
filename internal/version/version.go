package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"hytale-downloader/internal/buildinfo"
)

// Manifest represents the version manifest response
type Manifest struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

// UpdateInfo represents the downloader update info
type UpdateInfo struct {
	Latest string `json:"latest"`
}

// getBaseURL returns the appropriate base URL based on build branch
func getBaseURL() string {
	if buildinfo.Branch == "release" {
		return "https://downloader.hytale.com"
	}
	return "https://downloader-dev.hytale.com"
}

// getAccountDataURL returns the appropriate account-data URL based on build branch
func getAccountDataURL() string {
	if buildinfo.Branch == "release" {
		return "https://account-data.hytale.com"
	}
	return "https://account-data-dev.hytale.com"
}

// Get fetches the version manifest for a patchline
func Get(client *http.Client, patchline string) (*Manifest, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := fmt.Sprintf("%s/version/%s.json", getBaseURL(), patchline)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// GetSignedURL fetches the signed download URL for a patchline
func GetSignedURL(client *http.Client, patchline string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	url := fmt.Sprintf("%s/game-assets/%s", getAccountDataURL(), patchline)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP status: %s\nResponse: %s", resp.Status, string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// CheckForUpdates checks if a newer version of the downloader is available
func CheckForUpdates() (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/version.json", getBaseURL())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetDownloaderURL returns the URL to download the latest downloader
func GetDownloaderURL() string {
	return fmt.Sprintf("%s/hytale-downloader.zip", getBaseURL())
}
