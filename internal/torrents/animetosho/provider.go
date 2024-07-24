package animetosho

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/rs/zerolog"
)

type Provider struct {
	logger *zerolog.Logger
}

func NewProvider(logger *zerolog.Logger) hibiketorrent.Provider {
	return &Provider{
		logger: logger,
	}
}

func (at *Provider) Search(opts hibiketorrent.SearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) SmartSearch(opts hibiketorrent.SmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) CanSmartSearch() bool {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) CanFindBestRelease() bool {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) SupportsAdult() bool {
	//TODO implement me
	panic("implement me")
}
