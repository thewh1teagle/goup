package main

import (
	"log"

	"github.com/thewh1teagle/goup/updater"
)

// go run -ldflags="-X 'main.Tag=v0.0.0'" cmd/main.go
// go build -ldflags="-X 'main.Tag=v0.0.0'" cmd

var Version string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	u, err := createUpdater()
	if err != nil {
		log.Fatal(err)
	}
	if err := checkAndUpdate(u); err != nil {
		log.Fatal(err)
	}
}

func createUpdater() (*updater.GitHubUpdater, error) {
	options := updater.GitHubUpdaterOptions{
		User:       "thewh1teagle",
		Repo:       "goup",
		CurrentTag: Version,
		Patterns: updater.PlatformAssets{
			Windows: "goup_windows_$arch$ext",
			Linux:   "goup_linux_$arch",
			MacOS:   "goup_darwin_$arch",
		},
	}
	return updater.NewGitHubUpdater(options)
}

func checkAndUpdate(updater *updater.GitHubUpdater) error {
	newUpdate, err := updater.CheckForUpdate()
	if err != nil {
		return err
	}
	if newUpdate == nil {
		log.Println("No update available")
		return nil
	}
	log.Printf("Installing update: %s", newUpdate.URL)
	return updater.DownloadAndInstall(newUpdate, func(part, total int64) {
		log.Printf("Downloaded (%.2f%%)\n", float64(part)*100/float64(total))
	})
}
