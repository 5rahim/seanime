package autodownloader

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/qbittorrent"
)

type (
	AutoDownloader struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
	}

	NewAutoDownloaderOptions struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
	}
)

func NewAutoDownloader(opts *NewAutoDownloaderOptions) *AutoDownloader {
	return &AutoDownloader{
		Logger:            opts.Logger,
		QbittorrentClient: opts.QbittorrentClient,
		WSEventManager:    opts.WSEventManager,
	}
}
