package ioutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Download downloads a file from url to path, showing progress
func Download(client *http.Client, url, path string) error {
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading: HTTP status %s", resp.Status)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	contentLength := resp.ContentLength
	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64
	var lastPercent int
	var lastPrintedBytes int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("error writing to file: %w", writeErr)
			}
			downloaded += int64(n)

			// Progress reporting
			if contentLength > 0 {
				percent := int((float64(downloaded) / float64(contentLength)) * 100)
				if percent != lastPercent {
					lastPercent = percent
					bar := buildProgressBar(percent, 50)
					fmt.Printf("\r%s %.1f%% %s / %s", bar, float64(percent), formatBytes(downloaded), formatBytes(contentLength))
				}
			} else {
				// Update every 1MB when content length unknown
				if downloaded-lastPrintedBytes >= 1024*1024 {
					lastPrintedBytes = downloaded
					fmt.Printf("\r%s", formatBytes(downloaded))
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("error reading response: %w", readErr)
		}
	}

	fmt.Println() // New line after progress
	return nil
}

// buildProgressBar creates a progress bar string
func buildProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	empty := width - filled
	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ValidateSHA256 validates a file's SHA256 checksum
func ValidateSHA256(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file for validation: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("could not compute hash: %w", err)
	}

	computed := hex.EncodeToString(hash.Sum(nil))
	if computed != strings.ToLower(expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, computed)
	}

	return nil
}
