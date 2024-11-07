package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type PlatformAssets struct {
	Windows string
	Linux   string
	MacOS   string
}

type GitHubUpdater struct {
	User            string
	Repo            string
	CurrentTag      string
	Patterns        PlatformAssets
	DownloadTimeout time.Duration
	CheckTimeout    time.Duration
}

type GitHubUpdaterOptions struct {
	User            string
	Repo            string
	CurrentTag      string
	Patterns        PlatformAssets
	DownloadTimeout time.Duration
	CheckTimeout    time.Duration
}

func NewGitHubUpdater(options GitHubUpdaterOptions) (*GitHubUpdater, error) {
	initCleanup()
	// Ensure currentTag is not empty
	if options.CurrentTag == "" {
		return nil, fmt.Errorf("currentTag cannot be empty")
	}

	// Defaults
	if options.DownloadTimeout == 0 {
		options.DownloadTimeout = DefaultDownloadTimeout
	}
	if options.CheckTimeout == 0 {
		options.CheckTimeout = DefaultCheckTimeout
	}

	updater := GitHubUpdater(options)
	return &updater, nil
}

func initCleanup() error {
	if execPath, err := os.Executable(); err == nil {
		tmpDir := filepath.Join(filepath.Dir(execPath), ".tmp")
		if _, err := os.Stat(tmpDir); err == nil {
			return os.RemoveAll(tmpDir)
		}
	}
	return nil
}

// GetPlatformBinary returns the filename pattern based on the current OS,
// with architecture placeholders expanded.
func (p PlatformAssets) GetPlatformBinary(currentVersion string) (string, error) {
	var pattern string
	switch runtime.GOOS {
	case "windows":
		pattern = p.Windows
	case "linux":
		pattern = p.Linux
	case "darwin":
		pattern = p.MacOS
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return expandPattern(pattern, currentVersion), nil
}

func (u *Update) Download(path string, progressCallback *ProgressCallback, timeout time.Duration) error {
	// Get the file
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(u.URL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: received status code %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	// Write the file
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	var writer io.Writer
	if progressCallback != nil {
		progressWriter := &ProgressWriter{
			Writer:           outFile,
			ProgressCallback: *progressCallback,
			TotalSize:        totalSize,
		}
		_, err = io.Copy(progressWriter, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to save update: %w", err)
		}
		writer = progressWriter
	} else {
		writer = outFile
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}
	return nil
}

func (g *GitHubUpdater) CheckForUpdate() (*Update, error) {
	filename, err := g.Patterns.GetPlatformBinary(g.CurrentTag)
	if err != nil {
		return nil, err
	}

	latestTag, err := getLatestTag(g.User, g.Repo, g.CheckTimeout)
	if err != nil {
		return nil, err
	}

	// Check if there is a new version
	if latestTag == g.CurrentTag {
		return nil, nil // No update available
	}
	log.Printf("latest tag %s current %s", latestTag, g.CurrentTag)

	// Construct the URL for the latest binary
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", g.User, g.Repo, latestTag, filename)

	// Make a HEAD request to check if the file exists without downloading it
	client := http.Client{Timeout: g.CheckTimeout}
	resp, err := client.Head(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the file exists by confirming a 200 OK status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update binary not found at %s", url)
	}
	return &Update{URL: url, Filename: filename, Version: latestTag}, nil
}

func getLatestTag(username string, repository string, timeout time.Duration) (string, error) {
	// Construct the URL for the latest release
	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", username, repository)

	// Perform the HTTP GET request
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get latest tag: %w", err)
	}
	defer resp.Body.Close()

	// Check the final URL after redirects, which should contain the latest tag
	latestTagURL := resp.Request.URL.String()

	tag := path.Base(latestTagURL)
	if tag == "" || strings.HasSuffix(tag, "latest") {
		return "", fmt.Errorf("failed to retrieve tag from URL")
	}
	return tag, nil
}

func (g *GitHubUpdater) DownloadAndInstall(update *Update, progressCallback ProgressCallback) error {
	currentPath, err := GetCurrentFilePath()
	if err != nil {
		return fmt.Errorf("failed to get current file path: %w", err)
	}

	// Download the update to a temporary location
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, update.Filename)
	defer os.Remove(tempPath)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}
	if err := update.Download(tempPath, &progressCallback, g.DownloadTimeout); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		return fmt.Errorf("downloaded update file not found at %s", tempPath)
	}

	if runtime.GOOS == "linux" {
		log.Printf("set permission to 0755 %s", tempPath)
		if err := setExecutablePermission(tempPath); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	// Handle file replacement differently on Windows
	if runtime.GOOS == "windows" {
		// Attempt to move the current file to a temporary location
		oldPath := filepath.Join(os.TempDir(), "old_"+filepath.Base(currentPath))
		log.Printf("Attempting to move current executable to %s", oldPath)
		if err := os.Rename(currentPath, oldPath); err != nil {
			log.Printf("Failed to move to temp location, attempting fallback to .tmp directory within current path")

			// Fallback: Get the directory of the current executable
			currentDir := filepath.Dir(currentPath)
			// Define the .tmp directory within the current executable's directory
			tmpDir := filepath.Join(currentDir, ".tmp")

			// Ensure the .tmp directory exists
			if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create .tmp directory: %w", err)
			}
			// Set the fallback path to move the current executable to the .tmp directory
			oldPath = filepath.Join(tmpDir, filepath.Base(currentPath))
			if err := os.Rename(currentPath, oldPath); err != nil {
				return fmt.Errorf("failed to rename current file to .tmp directory: %w", err)
			}
		}
		defer os.Remove(oldPath)
	}

	// Replace the current file with the downloaded update
	log.Printf("Renaming %s to %s", tempPath, currentPath)
	if err := os.Rename(tempPath, currentPath); err != nil {
		log.Printf("Rename failed: %v. Attempting to copy file instead.", err)
		if err := copyFile(tempPath, currentPath); err != nil {
			return fmt.Errorf("failed to replace current file with copy: %w", err)
		}
	}

	log.Println("Update installed successfully")
	return nil
}
