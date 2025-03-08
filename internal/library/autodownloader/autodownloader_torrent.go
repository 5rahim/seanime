package autodownloader

import (
	"errors"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"sync"

	"github.com/5rahim/habari"
	"github.com/samber/lo"
)

type (
	// NormalizedTorrent is a struct built from torrent from a provider.
	// It is used to normalize the data from different providers so that it can be used by the AutoDownloader.
	NormalizedTorrent struct {
		hibiketorrent.AnimeTorrent
		ParsedData *habari.Metadata `json:"parsedData"`
		magnet     string           // Access using GetMagnet()
	}
)

func (ad *AutoDownloader) getLatestTorrents(rules []*anime.AutoDownloaderRule) (ret []*NormalizedTorrent, err error) {
	ad.logger.Debug().Msg("autodownloader: Checking for new episodes")

	providerExtension, ok := ad.torrentRepository.GetDefaultAnimeProviderExtension()
	if !ok {
		ad.logger.Warn().Msg("autodownloader: No default torrent provider found")
		return nil, errors.New("no default torrent provider found")
	}

	// Get the latest torrents
	torrents, err := providerExtension.GetProvider().GetLatest()
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to get latest torrents")
		return nil, err
	}

	if ad.settings.EnableEnhancedQueries {
		// Get unique release groups
		uniqueReleaseGroups := GetUniqueReleaseGroups(rules)
		// Filter the torrents
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		wg.Add(len(uniqueReleaseGroups))

		for _, releaseGroup := range uniqueReleaseGroups {
			go func(releaseGroup string) {
				defer wg.Done()
				filteredTorrents, err := providerExtension.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
					Media: hibiketorrent.Media{},
					Query: releaseGroup,
				})
				if err != nil {
					return
				}
				mu.Lock()
				torrents = append(torrents, filteredTorrents...)
				mu.Unlock()
			}(releaseGroup)
		}
		wg.Wait()
		// Remove duplicates
		torrents = lo.UniqBy(torrents, func(t *hibiketorrent.AnimeTorrent) string {
			return t.Name
		})
	}

	// Normalize the torrents
	ret = make([]*NormalizedTorrent, 0, len(torrents))
	for _, t := range torrents {
		parsedData := habari.Parse(t.Name)
		ret = append(ret, &NormalizedTorrent{
			AnimeTorrent: *t,
			ParsedData:   parsedData,
		})
	}

	return ret, nil
}

// GetMagnet returns the magnet link for the torrent.
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
