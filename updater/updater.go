package updater

import (
	"time"
)

const (
	DefaultDownloadTimeout = 30 * time.Second
	DefaultCheckTimeout    = 2 * time.Second
)

type Update struct {
	URL      string
	Filename string
	Version  string
}

type Updater interface {
	CheckForUpdate() (*Update, error)
	DownloadAndInstall(update *Update, progressCallback ProgressCallback) error
}
