package main

import (
	"log"
	"time"

	"github.com/thewh1teagle/goup/updater"
)

// go run -ldflags="-X 'main.Tag=v0.0.0'" main.go
// go build -ldflags="-X 'main.Tag=v0.0.0'" main.go
var Tag string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	options := updater.GitHubUpdaterOptions{
		Username:   "thewh1teagle",
		Repo:       "goup",
		CurrentTag: Tag,
		Patterns: updater.PlatformBinaries{
			Windows: "goup_windows_$arch$ext", // x86_64
			Linux:   "goup_linux_$arch",       // x86_64
			MacOS:   "goup_darwin_$arch",      // x86_64, aarch64
		},
		DownloadTimeout: 30 * time.Second,
		CheckTimeout:    2 * time.Second,
	}
	updater, err := updater.NewGitHubUpdater(options)
	if err != nil {
		log.Fatal(err)
	}
	update, err := updater.CheckForUpdate()
	if err != nil {
		log.Fatal(err)
	}
	if update != nil {
		log.Printf("Installing update: %s", update.URL)
		err := updater.DownloadAndInstall(update, func(part int64, total int64) {
			log.Printf("Downloaded %d of %d bytes (%.2f%%)\n", part, total, float64(part)*100/float64(total))
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("No update available")
	}
}
