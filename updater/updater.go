package updater

import "fmt"

type Update struct {
	URL      string
	Filename string
	Version  string
}

type ProgressCallback func(current int64, total int64)

type Updater interface {
	CheckForUpdate() (*Update, error)
	DownloadAndInstall(update *Update, progressCallback ProgressCallback) error
}

func NewGitHubUpdater(options GitHubUpdaterOptions) (*GitHubUpdater, error) {
	// Ensure currentTag is not empty
	if options.CurrentTag == "" {
		return nil, fmt.Errorf("currentTag cannot be empty")
	}
	updater := GitHubUpdater(options)
	return &updater, nil
}
