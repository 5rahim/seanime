package autoselect

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"slices"
	"sync"
	"time"

	"github.com/samber/lo"
)

func (s *AutoSelect) search(ctx context.Context, media *anilist.CompleteAnime, episodeNumber int, profile *anime.AutoSelectProfile) ([]*hibiketorrent.AnimeTorrent, error) {
	s.log("Starting auto-select search")
	s.logger.Debug().Msgf("autoselect: Searching for episode %d of %s", episodeNumber, media.GetTitleSafe())

	// 1. Get providers to search
	providers := s.getProvidersToSearch(profile)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	s.logger.Debug().Strs("providers", providers).Msg("autoselect: Using providers")
	s.log(fmt.Sprintf("Searching with providers: %v", providers))

	// 2. Determine initial batch search capability
	shouldSearchBatch := s.shouldSearchBatch(media)

	// 3. Search concurrently from all providers
	allTorrents, err := s.searchFromProviders(ctx, providers, media, episodeNumber, shouldSearchBatch, profile)
	if err != nil {
		return nil, err
	}

	if len(allTorrents) == 0 {
		s.logger.Warn().Msg("autoselect: No torrents found")
		s.log("No torrents found")
		return nil, fmt.Errorf("no torrents found")
	}

	s.logger.Debug().Int("count", len(allTorrents)).Msg("autoselect: Total unique torrents found")
	s.log(fmt.Sprintf("Total unique torrents: %d", len(allTorrents)))

	return allTorrents, nil
}

// getProvidersToSearch returns the list of providers to search.
func (s *AutoSelect) getProvidersToSearch(profile *anime.AutoSelectProfile) []string {
	// Use profile providers if available
	if profile != nil && len(profile.Providers) > 0 {
		// Take 3 max
		maxProviders := 3
		if len(profile.Providers) < maxProviders {
			maxProviders = len(profile.Providers)
		}
		return profile.Providers[:maxProviders]
	}

	// Fall back to default provider
	defaultProviderExtension, ok := s.torrentRepository.GetDefaultAnimeProviderExtension()
	if !ok {
		s.logger.Error().Msg("autoselect: Default provider extension not found")
		return nil
	}
	return []string{defaultProviderExtension.GetID()}
}

// searchFromProviders searches concurrently from all providers and deduplicates results.
func (s *AutoSelect) searchFromProviders(
	ctx context.Context,
	providers []string,
	media *anilist.CompleteAnime,
	episodeNumber int,
	shouldSearchBatch bool,
	profile *anime.AutoSelectProfile,
) ([]*hibiketorrent.AnimeTorrent, error) {

	type providerResult struct {
		torrents []*hibiketorrent.AnimeTorrent
		err      error
	}

	results := make(chan providerResult, len(providers))
	var wg sync.WaitGroup

	// Search from each provider concurrently
	for _, provider := range providers {
		wg.Add(1)
		go func(providerID string) {
			defer wg.Done()

			torrents, err := s.searchFromProvider(ctx, providerID, media, episodeNumber, shouldSearchBatch, profile)
			results <- providerResult{
				torrents: torrents,
				err:      err,
			}
		}(provider)
	}

	// Close results channel when all searches are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and deduplicate results
	infohashes := make(map[string]struct{})
	var allTorrents []*hibiketorrent.AnimeTorrent
	var lastErr error

	for result := range results {
		if result.err != nil {
			lastErr = result.err
			continue
		}

		for _, t := range result.torrents {
			if _, exists := infohashes[t.InfoHash]; !exists {
				allTorrents = append(allTorrents, t)
				infohashes[t.InfoHash] = struct{}{}
			}
		}
	}

	// If no torrents found from any provider, return the last error
	if len(allTorrents) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return allTorrents, nil
}

// searchFromProvider searches from a single provider with batch/single fallback logic.
func (s *AutoSelect) searchFromProvider(
	ctx context.Context,
	provider string,
	media *anilist.CompleteAnime,
	episodeNumber int,
	shouldSearchBatch bool,
	profile *anime.AutoSelectProfile,
) ([]*hibiketorrent.AnimeTorrent, error) {

	s.logger.Debug().Str("provider", provider).Msg("autoselect: Searching from provider")

	resolutions := []string{""}
	if profile != nil && len(profile.Resolutions) > 0 {
		resolutions = profile.Resolutions
	}

	// Try each resolution until we get results
	for _, resolution := range resolutions {
		if resolution != "" {
			s.logger.Debug().Str("provider", provider).Str("resolution", resolution).Msg("autoselect: Trying resolution")
		}

		// Build search options for this resolution
		searchOptions, err := s.buildSearchOptions(provider, media, episodeNumber, shouldSearchBatch, resolution)
		if err != nil {
			s.logger.Warn().Err(err).Str("provider", provider).Msg("autoselect: Failed to build search options")
			continue
		}

		// Search loop with batch/single fallback
		var allTorrents []*hibiketorrent.AnimeTorrent

		for {
			s.logger.Debug().
				Str("provider", provider).
				Bool("batch", searchOptions.Batch).
				Int("episode", episodeNumber).
				Interface("type", searchOptions.Type).
				Str("resolution", resolution).
				Msg("autoselect: Executing search")

			data, err := s.torrentRepository.SearchAnime(ctx, searchOptions)

			logFound := 0
			if data != nil {
				logFound = len(data.Torrents)
			}
			s.logger.Debug().Str("provider", provider).Int("found", logFound).Msg("autoselect: Search completed")

			// error
			if err != nil {
				if searchOptions.Batch {
					// Batch search failed, retry without batch
					s.logger.Warn().Err(err).Str("provider", provider).Msg("autoselect: Batch search failed, retrying without batch")
					searchOptions.Batch = false
					continue
				}
				// Single search failed, break to try next resolution
				break
			}

			// Handle batch results
			if searchOptions.Batch {
				if data == nil || !s.validateBatchResults(data.Torrents) {
					s.logger.Warn().Str("provider", provider).Msg("autoselect: Batch results insufficient, retrying without batch")
					searchOptions.Batch = false
					continue
				}

				// Found valid batch, add to results
				s.logger.Debug().Str("provider", provider).Int("count", len(data.Torrents)).Msg("autoselect: Found valid batch torrents")
				allTorrents = append(allTorrents, data.Torrents...)

				// Also search for single episodes to maximize results
				s.logger.Debug().Str("provider", provider).Msg("autoselect: Searching for single episodes")
				singleOpts := searchOptions
				singleOpts.Batch = false

				data2, err2 := s.torrentRepository.SearchAnime(ctx, singleOpts)
				if err2 == nil && data2 != nil && len(data2.Torrents) > 0 {
					s.logger.Debug().Str("provider", provider).Int("count", len(data2.Torrents)).Msg("autoselect: Found single episode torrents")
					allTorrents = append(allTorrents, data2.Torrents...)
				}
				break
			}

			// Single episode search results
			if data != nil && len(data.Torrents) > 0 {
				allTorrents = append(allTorrents, data.Torrents...)
			}
			break
		}

		// If we found results, return them
		if len(allTorrents) > 0 {
			s.logger.Debug().Str("provider", provider).Str("resolution", resolution).Int("count", len(allTorrents)).Msg("autoselect: Found torrents with resolution")
			return allTorrents, nil
		}

		// no results with this resolution, try next one
		if resolution != "" {
			s.logger.Debug().Str("provider", provider).Str("resolution", resolution).Msg("autoselect: No results with this resolution, trying next")
		}
	}

	// no results found with any resolution
	return nil, fmt.Errorf("no torrents found with any resolution")
}

// shouldSearchBatch determines if we should initially attempt to search for batches.
func (s *AutoSelect) shouldSearchBatch(media *anilist.CompleteAnime) bool {
	if media.IsMovie() || !media.IsFinished() {
		return false
	}

	// Check if 2 weeks have passed since the anime ended
	// This helps avoid unnecessary batch searches for recently ended series to maximize results
	endDate := media.GetEndDate()
	if endDate != nil && endDate.GetYear() != nil && endDate.GetMonth() != nil && endDate.GetDay() != nil {
		endTime := time.Date(*endDate.GetYear(), time.Month(*endDate.GetMonth()), *endDate.GetDay(), 0, 0, 0, 0, time.UTC)
		twoWeeksAgo := time.Now().UTC().AddDate(0, 0, -14)

		if endTime.After(twoWeeksAgo) {
			return false
		}
	}

	return true
}

// buildSearchOptions constructs the search options based on the provider capabilities and resolution.
func (s *AutoSelect) buildSearchOptions(
	provider string,
	media *anilist.CompleteAnime,
	episodeNumber int,
	batch bool,
	resolution string,
) (itorrent.AnimeSearchOptions, error) {

	ext, ok := s.torrentRepository.GetAnimeProviderExtension(provider)
	if !ok {
		return itorrent.AnimeSearchOptions{}, fmt.Errorf("provider %s not found", provider)
	}

	settings := ext.GetProvider().GetSettings()

	searchType := itorrent.AnimeSearchTypeSmart
	query := ""

	if !settings.CanSmartSearch {
		searchType = itorrent.AnimeSearchTypeSimple
		// Use sanitized romaji title for simple search
		query = util.CleanMediaTitle(media.ToBaseAnime().GetRomajiTitleSafe())
	}

	return itorrent.AnimeSearchOptions{
		Provider:      provider,
		Type:          searchType,
		Media:         media.ToBaseAnime(),
		Query:         query,
		Batch:         batch,
		EpisodeNumber: episodeNumber,
		BestReleases:  false,
		Resolution:    resolution,
		SkipPreviews:  true,
	}, nil
}

// validateBatchResults checks if the batch results are sufficient.
func (s *AutoSelect) validateBatchResults(torrents []*hibiketorrent.AnimeTorrent) bool {
	nbFound := len(torrents)
	seedersArr := lo.Map(torrents, func(t *hibiketorrent.AnimeTorrent, _ int) int {
		return t.Seeders
	})

	if len(seedersArr) == 0 {
		return false
	}

	maxSeeders := slices.Max(seedersArr)
	// Conditions for a "good" batch search result
	if maxSeeders >= 15 || nbFound > 2 {
		return true
	}
	return false
}
