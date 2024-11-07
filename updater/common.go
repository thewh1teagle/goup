package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetPlatformBinary returns the filename pattern based on the current OS,
// with architecture placeholders expanded.
func (p PlatformBinaries) GetPlatformBinary() (string, error) {
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
	return expandPattern(pattern), nil
}

// expandPattern replaces %arch in the pattern with the actual architecture name.
func expandPattern(pattern string) string {
	arch := getArchName()
	return strings.ReplaceAll(pattern, "%arch", arch)
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
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}
