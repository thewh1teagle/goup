package main

import (
	"log"

	"github.com/thewh1teagle/goup/updater"
)

// go run -ldflags="-X 'main.Tag=v0.0.0'" main.go
// go build -ldflags="-X 'main.Tag=v0.0.0'" main.go
var Tag string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	patterns := updater.PlatformBinaries{
		Windows: "goup_windows_%arch", // x86_64
		Linux:   "goup_linux_%arch",   // x86_64
		MacOS:   "goup_darwin_%arch",  // x86_64, aarch64
	}
	updater, err := updater.NewGitHubUpdater("thewh1teagle", "goup", Tag, patterns)
	if err != nil {
		log.Fatal(err)
	}
	update, err := updater.CheckForUpdate()
	if err != nil {
		log.Println(err)
	}
	if update != nil {
		log.Printf("Installing update: %s", update.URL)
		if err := updater.DownloadAndInstall(update); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("No update available")
	}

}
