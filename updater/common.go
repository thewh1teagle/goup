package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// expandPattern replaces $arch, $ext, $version in the pattern with the actual architecture name.
func expandPattern(pattern string, currentVersion string) string {
	pattern = strings.ReplaceAll(pattern, "$arch", getArchName())
	pattern = strings.ReplaceAll(pattern, "$ext", getExt())
	pattern = strings.ReplaceAll(pattern, "$version", currentVersion)
	return pattern
}

func GetCurrentFilePath() (string, error) {
	// Get the path of the executable
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Convert it to an absolute path
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return exePath, nil
}

func getArchName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	default:
		return runtime.GOARCH
	}
}

func getExt() string {
	switch runtime.GOOS {
	case "windows":
		return ".exe"
	default:
		return ""
	}
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

func setExecutablePermission(path string) error {
	return os.Chmod(path, 0755)
}
