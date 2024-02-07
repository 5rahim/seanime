package autodownloader

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"time"
)

const (
	NyaaProvider   = "nyaa"
	NyaaRSSFeedURL = "https://nyaa.si/?page=rss&c=1_2"
)

type (
	AutoDownloader struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
		Rules             []*entities.AutoDownloaderRule
		Settings          *models.AutoDownloaderSettings
		settingsUpdatedCh chan struct{}
		stopCh            chan struct{}
		startCh           chan struct{}
		active            bool
	}

	NewAutoDownloaderOptions struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
		Rules             []*entities.AutoDownloaderRule
	}
)

func NewAutoDownloader(opts *NewAutoDownloaderOptions) *AutoDownloader {
	return &AutoDownloader{
		Logger:            opts.Logger,
		QbittorrentClient: opts.QbittorrentClient,
		WSEventManager:    opts.WSEventManager,
		Rules:             opts.Rules,
		Settings:          &models.AutoDownloaderSettings{},
		settingsUpdatedCh: make(chan struct{}, 1),
		stopCh:            make(chan struct{}, 1),
		startCh:           make(chan struct{}, 1),
		active:            false,
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings) {
	ad.Settings = settings
	ad.settingsUpdatedCh <- struct{}{} // Notify that the settings have been updated
	if ad.Settings.Enabled && !ad.active {
		ad.startCh <- struct{}{} // Start the auto downloader
	} else if !ad.Settings.Enabled && ad.active {
		ad.stopCh <- struct{}{} // Stop the auto downloader
	}
}

// Start will start the auto downloader.
// This should be run in a goroutine.
func (ad *AutoDownloader) Start() {
	ad.Logger.Debug().Msg("autodownloader: Starting auto downloader")

	// Start the auto downloader
	ad.start()
}

func (ad *AutoDownloader) start() {

	for {
		interval := 10
		if ad.Settings != nil && ad.Settings.Interval > 0 {
			interval = ad.Settings.Interval
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		select {
		case <-ad.settingsUpdatedCh:
			break // Restart the loop
		case <-ad.stopCh:
			ad.active = false
			ad.Logger.Debug().Msg("autodownloader: Stopping auto downloader")
		case <-ad.startCh:
			ad.active = true
			ad.Logger.Debug().Msg("autodownloader: Starting auto downloader")
		case <-ticker.C:
			if ad.active {
				ad.Logger.Debug().Msg("autodownloader: Checking for new episodes")
				ad.checkForNewEpisodes()
			}
		}
		ticker.Stop()
	}

}

func (ad *AutoDownloader) checkForNewEpisodes() {

}
