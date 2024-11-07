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

type GitHubUpdater struct {
	Username        string
	Repo            string
	CurrentTag      string
	Patterns        PlatformBinaries
	DownloadTimeout time.Duration
	CheckTimeout    time.Duration
}

type PlatformBinaries struct {
	Windows string
	Linux   string
	MacOS   string
}

type Update struct {
	URL      string
	Filename string
	Version  string
}

type ProgressCallback func(current int64, total int64)

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

type ProgressWriter struct {
	Writer           io.Writer
	ProgressCallback ProgressCallback
	TotalSize        int64
	CurrentSize      int64
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if err == nil {
		pw.CurrentSize += int64(n)
		if pw.ProgressCallback != nil {
			pw.ProgressCallback(pw.CurrentSize, pw.TotalSize)
		}
	}
	return n, err
}

func (g *GitHubUpdater) CheckForUpdate() (*Update, error) {
	filename, err := g.Patterns.GetPlatformBinary()
	if err != nil {
		return nil, err
	}

	latestTag, err := getLatestTag(g.Username, g.Repo, g.CheckTimeout)
	if err != nil {
		return nil, err
	}

	// Check if there is a new version
	if latestTag == g.CurrentTag {
		return nil, nil // No update available
	}
	log.Printf("latest tag %s current %s", latestTag, g.CurrentTag)

	// Construct the URL for the latest binary
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", g.Username, g.Repo, latestTag, filename)

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

type GitHubUpdaterOptions struct {
	Username        string
	Repo            string
	CurrentTag      string
	Patterns        PlatformBinaries
	DownloadTimeout time.Duration
	CheckTimeout    time.Duration
}

func NewGitHubUpdater(options GitHubUpdaterOptions) (*GitHubUpdater, error) {
	// Ensure currentTag is not empty
	if options.CurrentTag == "" {
		return nil, fmt.Errorf("currentTag cannot be empty")
	}

	updater := GitHubUpdater{
		Username:        options.Username,
		Repo:            options.Repo,
		CurrentTag:      options.CurrentTag,
		Patterns:        options.Patterns,
		DownloadTimeout: options.DownloadTimeout,
		CheckTimeout:    options.CheckTimeout,
	}
	return &updater, nil
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

	// Handle file replacement differently on Windows
	if runtime.GOOS == "windows" {
		// Move the current file to a temporary location
		oldPath := filepath.Join(os.TempDir(), "old_"+filepath.Base(currentPath))
		log.Printf("Move current executable to %s", oldPath)
		if err := os.Rename(currentPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename current file: %w", err)
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

func copyFile(src string, dst string) error {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Write data to dst
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return err
	}
	return nil
}
