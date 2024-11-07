package updater

import "time"

type UpdaterOptions struct {
	DownloadTimeout time.Duration
	CheckTimeout    time.Duration
}

type Updater interface {
	CheckForUpdate(currentTag string) (*Update, error)
	DownloadAndInstall(update *Update) error
}
