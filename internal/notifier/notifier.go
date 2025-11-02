package notifier

import (
	"path/filepath"
	"seanime/internal/database/models"
	"sync"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	Notifier struct {
		dataDir  mo.Option[string]
		settings mo.Option[*models.NotificationSettings]
		mu       sync.Mutex
		logoPath string
		logger   mo.Option[*zerolog.Logger]
	}

	Notification string
)

const (
	AutoDownloader Notification = "Auto Downloader"
	AutoScanner    Notification = "Auto Scanner"
	Debrid         Notification = "Debrid"
)

var GlobalNotifier = NewNotifier()

func init() {
	GlobalNotifier = NewNotifier()
}

func NewNotifier() *Notifier {
	return &Notifier{
		dataDir:  mo.None[string](),
		settings: mo.None[*models.NotificationSettings](),
		mu:       sync.Mutex{},
		logger:   mo.None[*zerolog.Logger](),
	}
}

func (n *Notifier) SetSettings(datadir string, settings *models.NotificationSettings, logger *zerolog.Logger) {
	if datadir == "" || settings == nil {
		return
	}

	n.mu.Lock()
	n.dataDir = mo.Some(datadir)
	n.settings = mo.Some(settings)
	n.logoPath = filepath.Join(datadir, "seanime-logo.png")
	n.logger = mo.Some(logger)
	n.mu.Unlock()
}

func (n *Notifier) canProceed(id Notification) bool {
	if !n.dataDir.IsPresent() || !n.settings.IsPresent() {
		return false
	}

	if n.settings.MustGet().DisableNotifications {
		return false
	}

	//switch id {
	//case AutoDownloader:
	//	return !n.settings.MustGet().DisableAutoDownloaderNotifications
	//case AutoScanner:
	//	return !n.settings.MustGet().DisableAutoScannerNotifications
	//}

	return false
}
