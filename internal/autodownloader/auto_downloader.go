package autodownloader

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/qbittorrent"
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
