package autodownloader

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/seanime-parser"
)

type (
	// NormalizedTorrent is a struct built from torrent from a provider.
	// It is used to normalize the data from different providers so that it can be used by the AutoDownloader.
	NormalizedTorrent struct {
		hibiketorrent.AnimeTorrent
		ParsedData *seanime_parser.Metadata
		magnet     string // Access using GetMagnet()
	}
)

func (ad *AutoDownloader) getLatestTorrents() (ret []*NormalizedTorrent, err error) {
	ad.logger.Debug().Msg("autodownloader: Checking for new episodes from Nyaa")

	providerExtension, ok := ad.torrentRepository.GetDefaultAnimeProviderExtension()
	if !ok {
		ad.logger.Warn().Msg("autodownloader: No default torrent provider found")
		return []*NormalizedTorrent{}, nil
	}

	// Get the latest torrents
	torrents, err := providerExtension.GetProvider().GetLatest()
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to get latest torrents")
		return nil, err
	}

	// Normalize the torrents
	ret = make([]*NormalizedTorrent, 0, len(torrents))
	for _, t := range torrents {
		parsedData := seanime_parser.Parse(t.Name)
		ret = append(ret, &NormalizedTorrent{
			AnimeTorrent: *t,
			ParsedData:   parsedData,
		})
	}

	return ret, nil
}

func (t *NormalizedTorrent) GetMagnet(providerExtension hibiketorrent.AnimeProvider) (string, error) {
	if t.magnet == "" {
		magnet, err := providerExtension.GetTorrentMagnetLink(&t.AnimeTorrent)
		if err != nil {
			return "", err
		}
		t.magnet = magnet
		return t.magnet, nil
	}
	return t.magnet, nil
}
