package autodownloader

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/qbittorrent"
)

const (
	NyaaRSSFeedURL = "https://nyaa.si/?page=rss&c=1_2"
)

type (
	AutoDownloader struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
		Rules             []*entities.AutoDownloaderRule
		Settings          *models.AutoDownloaderSettings
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
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings) {
	ad.Settings = settings
}
