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
)

type GitHubUpdater struct {
	Username   string
	Repo       string
	CurrentTag string
	Patterns   PlatformBinaries
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

func (u *Update) Download(destinationDir string) error {
	// Get the file
	resp, err := http.Get(u.URL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: received status code %d", resp.StatusCode)
	}

	// Write the file
	destinationPath := filepath.Join(destinationDir, u.Filename)
	outFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}
	return nil
}

func GetUpdate(username string, repository string, currentTag string, binaries *PlatformBinaries) (*Update, error) {
	filename, error := binaries.GetPlatformBinary()
	if error != nil {
		return nil, error
	}
	tag, _ := getLatestTag(username, repository)

	currentFilePath, error := GetCurrentFilePath()
	if error != nil {
		return nil, error
	}
	log.Println(tag, filename, currentFilePath)
	return nil, nil
}

func (g *GitHubUpdater) CheckForUpdate() (*Update, error) {
	filename, err := g.Patterns.GetPlatformBinary()
	if err != nil {
		return nil, err
	}

	latestTag, err := getLatestTag(g.Username, g.Repo)
	if err != nil {
		return nil, err
	}

	// Check if there is a new version
	if latestTag == g.CurrentTag {
		return nil, nil // No update available
	}

	// Construct the URL for the latest binary
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", g.Username, g.Repo, latestTag, filename)

	// Make a HEAD request to check if the file exists without downloading it
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the file exists by confirming a 200 OK status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update binary not found at %s", url)
	}

	return NewUpdate(url, filename, latestTag), nil
}

func getLatestTag(username string, repository string) (string, error) {
	// Construct the URL for the latest release
	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", username, repository)

	// Perform the HTTP GET request
	resp, err := http.Get(url)
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

func NewGitHubUpdater(username, repo, currentTag string, patterns PlatformBinaries) (*GitHubUpdater, error) {
	// Ensure currentTag is not empty
	if currentTag == "" {
		return nil, fmt.Errorf("currentTag cannot be empty")
	}

	return &GitHubUpdater{
		Username: username,
		Repo:     repo,
		Patterns: patterns,
	}, nil
}

func (g *GitHubUpdater) DownloadAndInstall(update *Update) error {
	currentPath, err := GetCurrentFilePath()
	if err != nil {
		return fmt.Errorf("failed to get current file path: %w", err)
	}

	// Download the update to a temporary location
	tempPath := filepath.Join(os.TempDir(), update.Filename)
	if err := update.Download(os.TempDir()); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Handle file replacement differently on Windows
	if runtime.GOOS == "windows" {
		// Move the current file to a temporary location
		oldPath := filepath.Join(os.TempDir(), "old_"+filepath.Base(currentPath))
		if err := os.Rename(currentPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename current file: %w", err)
		}
		// TODO: clean it
		// defer os.Remove(oldPath)
	}

	// Replace the current file with the downloaded update
	if err := os.Rename(tempPath, currentPath); err != nil {
		return fmt.Errorf("failed to replace current file: %w", err)
	}

	log.Println("Update installed successfully. Restarting application...")
	return nil
}
