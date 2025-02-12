package main

import (
	"net/http"
	"time"

	bypass "github.com/5rahim/hibike/pkg/util/bypass"
	"github.com/rs/zerolog"
	torrent "seanime/internal/extension/hibike/torrent"
)

type (
	MyAnimeTorrentProvider struct {
		url    string
		client *http.Client
		logger *zerolog.Logger
	}
)

func NewProvider(logger *zerolog.Logger) torrent.AnimeProvider {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = bypass.AddCloudFlareByPass(c.Transport)
	return &MyAnimeTorrentProvider{
		url:    "https://example.com",
		client: c,
		logger: logger,
	}
}

func (m *MyAnimeTorrentProvider) Search(opts torrent.AnimeSearchOptions) ([]*torrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MyAnimeTorrentProvider) SmartSearch(opts torrent.AnimeSmartSearchOptions) ([]*torrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MyAnimeTorrentProvider) GetTorrentInfoHash(torrent *torrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MyAnimeTorrentProvider) GetTorrentMagnetLink(torrent *torrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MyAnimeTorrentProvider) GetLatest() ([]*torrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MyAnimeTorrentProvider) GetSettings() torrent.AnimeProviderSettings {
	return torrent.AnimeProviderSettings{
		CanSmartSearch: true,
		SmartSearchFilters: []torrent.AnimeProviderSmartSearchFilter{
			torrent.AnimeProviderSmartSearchFilterEpisodeNumber,
			torrent.AnimeProviderSmartSearchFilterResolution,
			torrent.AnimeProviderSmartSearchFilterQuery,
			torrent.AnimeProviderSmartSearchFilterBatch,
		},
		SupportsAdult: false,
		Type:          "main",
	}
}
