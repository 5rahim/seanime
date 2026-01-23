package autodownloader

import (
	"errors"
	"seanime/internal/extension"
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
		ParsedData  *habari.Metadata `json:"parsedData"`
		magnet      string           // Access using GetMagnet()
		ExtensionID string
	}
)

func (ad *AutoDownloader) getTorrentsFromProviders(
	providers []extension.AnimeTorrentProviderExtension,
	rules []*anime.AutoDownloaderRule,
	profiles []*anime.AutoDownloaderProfile,
) (ret []*NormalizedTorrent, err error) {
	ad.logger.Debug().Msg("autodownloader: Checking for new episodes")

	if len(providers) == 0 {
		return nil, errors.New("no providers found")
	}

	mu := sync.Mutex{}
	torrents := make([]*NormalizedTorrent, 0)
	wg := sync.WaitGroup{}
	rateLimiter := limiter.NewLimiter(time.Second, 2)

	// Check if we should use the default provider for rules/profiles that don't specify one
	defaultProv, hasDefault := ad.torrentRepository.GetDefaultAnimeProviderExtension()

	for _, providerExt := range providers {
		wg.Add(1)
		go func(pExt extension.AnimeTorrentProviderExtension) {
			defer wg.Done()

			// Get all latest torrents
			latest, err := pExt.GetProvider().GetLatest()
			if err != nil {
				ad.logger.Error().Err(err).Str("provider", pExt.GetName()).Msg("autodownloader: Failed to get latest torrents")
			} else {
				mu.Lock()
				for _, t := range latest {
					parsedData := habari.Parse(t.Name)
					torrents = append(torrents, &NormalizedTorrent{
						AnimeTorrent: *t,
						ParsedData:   parsedData,
						ExtensionID:  pExt.GetID(),
					})
				}
				mu.Unlock()
			}

			// Release Groups + Resolutions
			// Identify rules relevant to this provider
			relevantRules := make([]*anime.AutoDownloaderRule, 0)
			for _, rule := range rules {
				isRelevant := false

				// Check rule providers
				if len(rule.Providers) > 0 {
					if lo.Contains(rule.Providers, pExt.GetID()) {
						isRelevant = true
					}
				} else {
					// Check profile providers
					hasProfileProviders := false
					if rule.ProfileID != nil {
						profile, found := lo.Find(profiles, func(p *anime.AutoDownloaderProfile) bool {
							return p.DbID == *rule.ProfileID
						})
						if found && len(profile.Providers) > 0 {
							hasProfileProviders = true
							if lo.Contains(profile.Providers, pExt.GetID()) {
								isRelevant = true
							}
						}
					}

					// If neither rule nor profile has providers, and this is the default provider, it's relevant
					if !hasProfileProviders && hasDefault && defaultProv.GetID() == pExt.GetID() {
						isRelevant = true
					}
				}

				if isRelevant {
					relevantRules = append(relevantRules, rule)
				}
			}

			if len(relevantRules) == 0 {
				return
			}

			// Get deduplicated map of "Release Group" -> ["Resolutions"], e.g. "SubsPlease" -> []string{"1080p", "720p"}
			// We pass 'profiles' so resolutions can be inherited if missing from the rule
			releaseGroupToResolutions := ad.getReleaseGroupToResolutionsMap(relevantRules, profiles)

			for releaseGroup, resolutions := range releaseGroupToResolutions {
				foundForGroup := false

				// Search with resolution (limit 2)
				for i, resolution := range resolutions {
					if i >= 2 {
						break
					}
					rateLimiter.Wait()
					res, err := pExt.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
						Media: hibiketorrent.Media{},
						Query: releaseGroup + " " + resolution,
					})
					if err == nil {
						if len(res) > 0 {
							foundForGroup = true
						}
						mu.Lock()
						for _, t := range res {
							t := t // Capture loop variable
							parsedData := habari.Parse(t.Name)
							torrents = append(torrents, &NormalizedTorrent{
								AnimeTorrent: *t,
								ParsedData:   parsedData,
								ExtensionID:  pExt.GetID(),
							})
						}
						mu.Unlock()
					}
				}

				// Search without resolution as a fallback if nothing found for specific resolutions
				if !foundForGroup {
					rateLimiter.Wait()
					res, err := pExt.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
						Media: hibiketorrent.Media{},
						Query: releaseGroup,
					})
					if err == nil {
						mu.Lock()
						for _, t := range res {
							t := t
							parsedData := habari.Parse(t.Name)
							torrents = append(torrents, &NormalizedTorrent{
								AnimeTorrent: *t,
								ParsedData:   parsedData,
								ExtensionID:  pExt.GetID(),
							})
						}
						mu.Unlock()
					}
				}
			}

		}(providerExt)
	}
	wg.Wait()

	// Deduplicate
	ret = lo.UniqBy(torrents, func(t *NormalizedTorrent) string {
		return t.Name
	})

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
