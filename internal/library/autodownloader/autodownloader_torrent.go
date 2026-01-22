package autodownloader

import (
	"errors"
	"seanime/internal/api/metadata"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/util/limiter"
	"sync"
	"time"

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

	mu := sync.Mutex{}

	rateLimiter := limiter.NewLimiter(time.Second, 2)

	// For enhanced queries we use smart search queries for each rule
	// If a rule cannot be resolved, it is added to the rulesRest slice
	if ad.settings.EnableEnhancedQueries {
		settings := providerExtension.GetProvider().GetSettings()
		if settings.CanSmartSearch {
			wg := sync.WaitGroup{}
			wg.Add(len(rules))
			for _, rule := range rules {
				go func() {
					defer wg.Done()

					rateLimiter.Wait()

					ac, ok := ad.animeCollection.Get()
					if !ok {
						return
					}

					media, ok := ac.FindAnime(rule.MediaId)
					if !ok || media == nil || media.GetStatus() == nil || media.GetFormat() == nil {
						return
					}

					mediaMetadata, err := ad.metadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, rule.MediaId)
					if err != nil {
						return
					}

					queryMedia := hibiketorrent.Media{
						ID:                   media.GetID(),
						IDMal:                media.GetIDMal(),
						Status:               string(*media.GetStatus()),
						Format:               string(*media.GetFormat()),
						EnglishTitle:         media.GetTitle().GetEnglish(),
						RomajiTitle:          media.GetRomajiTitleSafe(),
						EpisodeCount:         media.GetTotalEpisodeCount(),
						AbsoluteSeasonOffset: 0,
						Synonyms:             media.GetSynonymsContainingSeason(),
						IsAdult:              *media.GetIsAdult(),
						StartDate:            &hibiketorrent.FuzzyDate{},
					}

					if media.GetStartDate() != nil && media.GetStartDate().GetYear() != nil {
						queryMedia.StartDate.Year = *media.GetStartDate().GetYear()
						queryMedia.StartDate.Month = media.GetStartDate().GetMonth()
						queryMedia.StartDate.Day = media.GetStartDate().GetDay()
					}

					resolution := ""
					if len(rule.Resolutions) > 0 {
						resolution = rule.Resolutions[0]
					}

					res, err := providerExtension.GetProvider().SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
						Media:      queryMedia,
						Resolution: resolution,
						AnidbAID:   mediaMetadata.GetMappings().AnidbId,
					})
					if err != nil {
						return
					}

					mu.Lock()
					torrents = append(torrents, res...)
					mu.Unlock()
				}()
			}
			wg.Wait()
		}

	}

	releaseGroupToResolutions := GetReleaseGroupToResolutionsMap(rules) // e.g. Subsplease -> []{"1080p","720p"}

	for releaseGroup, resolutions := range releaseGroupToResolutions {
		wg := sync.WaitGroup{}
		var resultForReleaseGroup []*hibiketorrent.AnimeTorrent
		for i, resolution := range resolutions {
			if i >= 2 { // Only search for the first 2 resolutions
				break
			}
			wg.Add(1)
			go func(releaseGroup string) {
				defer wg.Done()

				rateLimiter.Wait()

				res, err := providerExtension.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
					Media: hibiketorrent.Media{},
					Query: releaseGroup + " " + resolution,
				})
				if err != nil {
					return
				}
				mu.Lock()
				resultForReleaseGroup = append(resultForReleaseGroup, res...)
				mu.Unlock()
			}(releaseGroup)
		}
		wg.Wait()
		// Launch a query without resolution if both failed to return anything
		if len(resultForReleaseGroup) == 0 {
			res, err := providerExtension.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
				Media: hibiketorrent.Media{},
				Query: releaseGroup,
			})
			if err != nil {
				continue
			}
			mu.Lock()
			resultForReleaseGroup = append(resultForReleaseGroup, res...)
			mu.Unlock()
		}
		// Add the results to the torrents
		mu.Lock()
		torrents = append(torrents, resultForReleaseGroup...)
		mu.Unlock()
	}

	// Deduplicate
	torrents = lo.UniqBy(torrents, func(t *hibiketorrent.AnimeTorrent) string {
		return t.Name
	})

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
