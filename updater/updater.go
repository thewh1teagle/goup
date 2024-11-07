package updater

type Updater interface {
	CheckForUpdate(currentTag string) (*Update, error)
	DownloadAndInstall(update *Update) error
}

func NewUpdate(url, filename string, version string) *Update {
	return &Update{
		Filename: filename,
		URL:      url,
		Version:  version,
	}
}
