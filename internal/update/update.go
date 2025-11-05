package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	gover "github.com/hashicorp/go-version"
)

const (
	repoOwner = "coollabsio"
	repoName  = "jean"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateJean checks for and installs the latest version of jean
func UpdateJean(currentVersion string) error {
	ctx := context.Background()

	// Parse current version
	current, err := gover.NewVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}

	// Fetch latest release from GitHub API
	latestRelease, err := fetchLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	if latestRelease == nil {
		return fmt.Errorf("no releases found for %s/%s", repoOwner, repoName)
	}

	// Strip 'v' prefix from release version for comparison
	releaseVersion := latestRelease.TagName
	if len(releaseVersion) > 0 && releaseVersion[0] == 'v' {
		releaseVersion = releaseVersion[1:]
	}

	latestVer, err := gover.NewVersion(releaseVersion)
	if err != nil {
		return fmt.Errorf("failed to parse latest version: %w", err)
	}

	// Check if update is needed
	if !latestVer.GreaterThan(current) {
		fmt.Printf("jean is already up to date (version %s)\n", currentVersion)
		return nil
	}

	// Get the executable path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Printf("Updating jean from %s to %s...\n", currentVersion, latestRelease.TagName)

	// Download and install the new binary
	if err := downloadAndInstall(ctx, latestRelease, exe); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully updated to %s\n", latestRelease.TagName)
	return nil
}

// fetchLatestRelease fetches the latest release from GitHub API
func fetchLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadAndInstall downloads the latest binary and replaces the current executable
func downloadAndInstall(ctx context.Context, release *Release, exePath string) error {
	// Find the appropriate asset for this platform
	assetURL := findAssetURL(release)
	if assetURL == "" {
		return fmt.Errorf("no suitable release asset found for this platform")
	}

	// Create a temporary directory for the download
	tmpDir, err := os.MkdirTemp("", "jean-update-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the binary
	tmpFile := filepath.Join(tmpDir, "jean.tar.gz")
	if err := downloadFile(ctx, assetURL, tmpFile); err != nil {
		return err
	}

	// Extract the binary from tar.gz
	extractedPath, err := extractBinary(tmpFile, tmpDir)
	if err != nil {
		return err
	}

	// Check if we need sudo (file is in a protected location)
	needsSudo := !isWritable(filepath.Dir(exePath))

	// Replace the current executable
	// First, make a backup
	backup := exePath + ".backup"
	if err := renameFile(backup, exePath, needsSudo); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Copy the new binary
	if err := copyFileTo(extractedPath, exePath, needsSudo); err != nil {
		// Restore the backup
		renameFile(exePath, backup, needsSudo)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Make it executable
	if err := chmodFile(exePath, 0755, needsSudo); err != nil {
		// Restore the backup
		renameFile(exePath, backup, needsSudo)
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Remove the backup
	removeFile(backup, needsSudo)

	return nil
}

// findAssetURL finds the download URL for the appropriate platform
func findAssetURL(release *Release) string {
	// Look for tar.gz asset matching the current platform
	// The asset name should be like: jean_0.1.0_linux_amd64.tar.gz
	for _, asset := range release.Assets {
		// For now, we'll just return the first tar.gz file
		// In production, you'd want to check the OS and ARCH
		if len(asset.BrowserDownloadURL) > 0 && contains(asset.Name, ".tar.gz") {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

// downloadFile downloads a file from a URL
func downloadFile(ctx context.Context, url, filePath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

// extractBinary extracts the jean binary from a tar.gz file
func extractBinary(tarPath, destDir string) (string, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	var extractedPath string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Only extract the jean binary file
		if header.Name == "jean" || filepath.Base(header.Name) == "jean" {
			outPath := filepath.Join(destDir, "jean")
			out, err := os.Create(outPath)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}

			if _, err := io.Copy(out, tarReader); err != nil {
				out.Close()
				return "", fmt.Errorf("failed to extract binary: %w", err)
			}
			out.Close()

			// Make the extracted binary executable
			if err := os.Chmod(outPath, 0755); err != nil {
				return "", fmt.Errorf("failed to make extracted binary executable: %w", err)
			}

			extractedPath = outPath
			break
		}
	}

	if extractedPath == "" {
		return "", fmt.Errorf("jean binary not found in archive")
	}

	return extractedPath, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// isWritable checks if a directory is writable
func isWritable(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return false
	}

	// Try to write a temporary file
	tmpFile := filepath.Join(dir, ".write-test")
	err = os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	os.Remove(tmpFile)
	return true
}

// renameFile renames a file, using sudo if necessary
func renameFile(newPath, oldPath string, useSudo bool) error {
	if !useSudo {
		return os.Rename(oldPath, newPath)
	}

	// Use sudo mv command
	cmd := exec.Command("sudo", "mv", oldPath, newPath)
	return cmd.Run()
}

// copyFileTo copies a file, using sudo if necessary
func copyFileTo(src, dst string, useSudo bool) error {
	if !useSudo {
		return copyFile(src, dst)
	}

	// Use sudo cp command
	cmd := exec.Command("sudo", "cp", src, dst)
	return cmd.Run()
}

// chmodFile changes file permissions, using sudo if necessary
func chmodFile(path string, perm os.FileMode, useSudo bool) error {
	if !useSudo {
		return os.Chmod(path, perm)
	}

	// Use sudo chmod command
	cmd := exec.Command("sudo", "chmod", fmt.Sprintf("%o", perm), path)
	return cmd.Run()
}

// removeFile removes a file, using sudo if necessary
func removeFile(path string, useSudo bool) error {
	if !useSudo {
		return os.Remove(path)
	}

	// Use sudo rm command
	cmd := exec.Command("sudo", "rm", "-f", path)
	return cmd.Run()
}
