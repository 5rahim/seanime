package autodownloader

import (
	"context"
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
		*hibiketorrent.AnimeTorrent
		ParsedData  *habari.Metadata `json:"parsedData"`
		magnet      string           // Access using GetMagnet()
		ExtensionID string
	}
)

func (ad *AutoDownloader) fetchTorrentsFromProviders(
	ctx context.Context,
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

	// Find which provider to use for rules/profiles that don't specify one
	// If the default provider exists, we'll use it,
	// if it doesn't, we get the provider with the most rules
	var hasDefault bool
	defaultProv, foundDefault := ad.torrentRepository.GetAnimeProviderExtension(ad.settings.Provider)
	if !foundDefault {
		ad.logger.Warn().Msg("autodownloader: Default provider not found, it might be uninstalled")
		// Get the provider with the most rules?
		providerByRules := make(map[string]int)
		for _, rule := range rules {
			// Check rule providers
			if len(rule.Providers) > 0 {
				for _, p := range rule.Providers {
					providerByRules[p]++
				}
			} else {
				// Check profile providers
				if rule.ProfileID != nil {
					profile, found := lo.Find(profiles, func(p *anime.AutoDownloaderProfile) bool {
						return p.DbID == *rule.ProfileID
					})
					if found && len(profile.Providers) > 0 {
						for _, p := range profile.Providers {
							providerByRules[p]++
						}
					}
				}
			}
		}
		if len(providerByRules) > 0 {
			mostRulesProvider := lo.MaxBy(lo.Keys(providerByRules), func(a, b string) bool {
				return providerByRules[a] < providerByRules[b]
			})
			defaultProv, foundDefault = ad.torrentRepository.GetAnimeProviderExtension(mostRulesProvider)
			if foundDefault {
				hasDefault = true
			}
		}
		if !hasDefault {
			defaultProv, foundDefault = ad.torrentRepository.GetAnimeProviderExtensionOrDefault(ad.settings.Provider)
			hasDefault = foundDefault
		}
	} else {
		hasDefault = true
	}
	ad.logger.Debug().Str("extension", defaultProv.GetName()).Bool("hasDefault", hasDefault).Msg("autodownloader: Checked for default provider")

	// go through all providers concurrently
	for _, providerExt := range providers {
		wg.Add(1)
		go func(pExt extension.AnimeTorrentProviderExtension) {
			defer wg.Done()

			// Set up a rate limiter for a single provider
			rateLimiter := limiter.NewLimiter(time.Second, 2) // 2 reqs per sec

			// Step 1: Get all latest torrents
			ad.logger.Debug().Str("provider", pExt.GetName()).Msg("autodownloader: Getting latest torrents")
			latest, err := pExt.GetProvider().GetLatest()
			if err != nil {
				ad.logger.Error().Err(err).Str("provider", pExt.GetName()).Msg("autodownloader: Failed to get latest torrents")
			} else {
				for _, t := range latest {
					parsedData := habari.Parse(t.Name)
					mu.Lock()
					torrents = append(torrents, &NormalizedTorrent{
						AnimeTorrent: t,
						ParsedData:   parsedData,
						ExtensionID:  pExt.GetID(),
					})
					mu.Unlock()
				}
			}

			// Step 2: Identify rules relevant to this provider
			// If this provider is assigned to rules (directly or via profiles),
			// we get them in order to retrieve the release groups/resolutions combos
			// This will be used to launch more precise searches

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

			// Get deduplicated map of Release groups to resolutions from rules or rules' profiles
			// e.g. "SubsPlease" -> []string{"1080p", "720p"}
			releaseGroupToResolutions := ad.getReleaseGroupToResolutionsMap(relevantRules, profiles)
			ad.logger.Debug().Interface("releaseGroups", releaseGroupToResolutions).Msg("autodownloader: Found release groups to search for")

			// For each release group, search for torrents with resolutions concurrently
			// e.g. "SubsPlease 1080p"
			pWg := sync.WaitGroup{}
			pWg.Add(len(releaseGroupToResolutions))
			for releaseGroup, resolutions := range releaseGroupToResolutions {
				go func(rg string, res []string) {
					defer pWg.Done()
					foundForGroup := false

					// For each release group, search with a specific resolution
					// Devnote: Would be better to use OR operators for resolutions but not all providers support it, so we'll limit to 2 searches
					for i, resolution := range res {
						if i >= 2 || resolution == "-" {
							break
						}
						rateLimiter.Wait()
						ad.logger.Debug().Str("extensionId", pExt.GetID()).Str("releaseGroup", rg).Str("resolution", resolution).Msg("autodownloader: Searching for torrents")
						result, err := pExt.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
							Media: hibiketorrent.Media{},
							Query: rg + " " + resolution,
						})
						if err == nil {
							if len(result) > 0 {
								foundForGroup = true
							}
							for _, t := range result {
								parsedData := habari.Parse(t.Name)
								mu.Lock()
								torrents = append(torrents, &NormalizedTorrent{
									AnimeTorrent: t,
									ParsedData:   parsedData,
									ExtensionID:  pExt.GetID(),
								})
								mu.Unlock()
							}
						}
					}

					// Search without resolution as a fallback if nothing found for specific resolutions
					if !foundForGroup {
						rateLimiter.Wait()
						ad.logger.Debug().Str("extensionId", pExt.GetID()).Str("releaseGroup", rg).Msg("autodownloader: Searching for torrents without resolution")
						result, err := pExt.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
							Media: hibiketorrent.Media{},
							Query: rg,
						})
						if err == nil {
							for _, t := range result {
								parsedData := habari.Parse(t.Name)
								mu.Lock()
								torrents = append(torrents, &NormalizedTorrent{
									AnimeTorrent: t,
									ParsedData:   parsedData,
									ExtensionID:  pExt.GetID(),
								})
								mu.Unlock()
							}
						}
					}
				}(releaseGroup, resolutions)
			}
			pWg.Wait()

		}(providerExt)
	}
	wg.Wait()

	// Deduplicate
	ret = lo.Filter(torrents, func(t *NormalizedTorrent, _ int) bool {
		return t.InfoHash != ""
	})
	ret = lo.UniqBy(ret, func(t *NormalizedTorrent) string {
		return t.InfoHash
	})

	ad.logger.Debug().Int("torrents", len(ret)).Msg("autodownloader: Found torrents")

	return ret, nil
}

// GetMagnet returns the magnet link for the torrent.
func (t *NormalizedTorrent) GetMagnet(providerExtension hibiketorrent.AnimeProvider) (string, error) {
	if t.magnet == "" {
		magnet, err := providerExtension.GetTorrentMagnetLink(t.AnimeTorrent)
		if err != nil {
			return "", err
		}
		t.magnet = magnet
		return t.magnet, nil
	}
	return t.magnet, nil
}
