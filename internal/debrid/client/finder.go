package debrid_client

import (
	"cmp"
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/debrid/debrid"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"slices"
	"strconv"
)

func (r *Repository) findBestTorrent(provider debrid.Provider, media *anilist.CompleteAnime, episodeNumber int) (selectedTorrent *hibiketorrent.AnimeTorrent, fileId string, err error) {

	defer util.HandlePanicInModuleWithError("debridstream/findBestTorrent", &err)

	r.logger.Debug().Msgf("debridstream: Finding best torrent for %s, Episode %d", media.GetTitleSafe(), episodeNumber)

	providerId := itorrent.ProviderAnimeTosho // todo: get provider from settings

	// Get AnimeTosho provider extension
	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(providerId)
	if !ok {
		r.logger.Error().Str("provider", itorrent.ProviderAnimeTosho).Msg("debridstream: AnimeTosho provider extension not found")
		return nil, "", fmt.Errorf("provider extension not found")
	}

	searchBatch := false
	canSearchBatch := !media.IsMovie() && media.IsFinished()
	if canSearchBatch {
		searchBatch = true
	}

	loopCount := 0

	var data *itorrent.SearchData
searchLoop:
	for {
		data, err = r.torrentRepository.SearchAnime(itorrent.AnimeSearchOptions{
			Provider:      providerId,
			Type:          itorrent.AnimeSearchTypeSmart,
			Media:         media.ToBaseAnime(),
			Query:         "",
			Batch:         searchBatch,
			EpisodeNumber: episodeNumber,
			BestReleases:  false,
			Resolution:    r.settings.StreamPreferredResolution,
		})
		// If we are searching for batches, we don't want to return an error if no torrents are found
		// We will just search again without the batch flag
		if err != nil {
			if !searchBatch {
				r.logger.Error().Err(err).Msg("debridstream: Error searching torrents")
				return nil, "", err
			}
			searchBatch = false
			continue
		}

		// Get cached
		hashes := make([]string, 0)
		for _, t := range data.Torrents {
			if t.InfoHash == "" {
				continue
			}
			hashes = append(hashes, t.InfoHash)
		}
		instantAvail := provider.GetInstantAvailability(hashes)
		data.DebridInstantAvailability = instantAvail

		// If we are searching for batches, we want to filter out torrents that are not cached
		if searchBatch {
			// Nothing found, search again without the batch flag
			if len(data.Torrents) == 0 {
				searchBatch = false
				loopCount++
				continue
			}
			if len(data.DebridInstantAvailability) > 0 {
				r.logger.Debug().Msg("debridstream: Found cached instant availability")
				data.Torrents = lo.Filter(data.Torrents, func(t *hibiketorrent.AnimeTorrent, i int) bool {
					_, isCached := data.DebridInstantAvailability[t.InfoHash]
					return isCached
				})
				break searchLoop
			}
			// If we didn't find any cached batches, we will search again without the batch flag
			searchBatch = false
			loopCount++
			continue
		}

		// If on the first try were looking for file torrents but found no cached ones, we will search again for batches
		if loopCount == 0 && canSearchBatch && len(data.DebridInstantAvailability) == 0 {
			searchBatch = true
			loopCount++
			continue
		}

		// Stop looking if either we found cached torrents or no cached batches were found
		break searchLoop
	}

	if data == nil || len(data.Torrents) == 0 {
		r.logger.Error().Msg("debridstream: No torrents found")
		return nil, "", fmt.Errorf("no torrents found")
	}

	// Sort by seeders from highest to lowest
	slices.SortStableFunc(data.Torrents, func(a, b *hibiketorrent.AnimeTorrent) int {
		return cmp.Compare(b.Seeders, a.Seeders)
	})

	r.logger.Debug().Msgf("debridstream: Found %d torrents", len(data.Torrents))

	hashes := make([]string, 0)
	for _, t := range data.Torrents {
		if t.InfoHash == "" {
			continue
		}
		hashes = append(hashes, t.InfoHash)
	}

	// Find cached torrent
	instantAvail := provider.GetInstantAvailability(hashes)
	data.DebridInstantAvailability = instantAvail

	// Filter out torrents that are not cached if we have cached instant availability
	if len(data.DebridInstantAvailability) > 0 {
		r.logger.Debug().Msg("debridstream: Found cached instant availability")
		data.Torrents = lo.Filter(data.Torrents, func(t *hibiketorrent.AnimeTorrent, i int) bool {
			_, isCached := data.DebridInstantAvailability[t.InfoHash]
			return isCached
		})
	}

	tries := 0

	for _, searchT := range data.Torrents {
		if tries >= 2 {
			break
		}

		r.logger.Trace().Msgf("debridstream: Getting torrent magnet for %s", searchT.Name)
		magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(searchT)
		if err != nil {
			r.logger.Warn().Err(err).Msgf("debridstream: Error scraping magnet link for %s", searchT.Link)
			tries++
			continue
		}

		// Set the magnet link
		searchT.MagnetLink = magnet

		r.logger.Debug().Msgf("debridstream: Adding torrent %s from magnet", searchT.Link)

		// Get the torrent info
		// On Real-Debrid, this will add the torrent
		info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{
			MagnetLink: searchT.MagnetLink,
			InfoHash:   searchT.InfoHash,
		})
		if err != nil {
			r.logger.Warn().Err(err).Msgf("debridstream: Error adding torrent %s", searchT.Link)
			tries++
			continue
		}

		filepaths := lo.Map(info.Files, func(f *debrid.TorrentItemFile, _ int) string {
			return f.Path
		})

		if len(filepaths) == 0 {
			r.logger.Error().Msg("debridstream: No files found in the torrent")
			return nil, "", fmt.Errorf("no files found in the torrent")
		}

		// Create a new Torrent Analyzer
		analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
			Logger:           r.logger,
			Filepaths:        filepaths,
			Media:            media,
			Platform:         r.platform,
			MetadataProvider: r.metadataProvider,
		})

		r.logger.Debug().Msgf("debridstream: Analyzing torrent %s", searchT.Link)

		// Analyze torrent files
		analysis, err := analyzer.AnalyzeTorrentFiles()
		if err != nil {
			r.logger.Warn().Err(err).Msg("debridstream: Error analyzing torrent files")
			// Remove torrent on failure (if it was added)
			if info.ID != nil {
				go func() {
					_ = provider.DeleteTorrent(*info.ID)
				}()
			}
			tries++
			continue
		}

		analysisFile, found := analysis.GetFileByAniDBEpisode(strconv.Itoa(episodeNumber))
		// Check if analyzer found the episode
		if !found {
			r.logger.Error().Msgf("debridstream: Failed to auto-select episode from torrent %s", searchT.Link)
			// Remove torrent on failure
			if info.ID != nil {
				go func() {
					_ = provider.DeleteTorrent(*info.ID)
				}()
			}
			tries++
			continue
		}

		r.logger.Debug().Msgf("debridstream: Found corresponding file for episode %s: %s", strconv.Itoa(episodeNumber), analysisFile.GetLocalFile().Name)

		tFile := info.Files[analysisFile.GetIndex()]
		r.logger.Debug().Msgf("debridstream: Selected torrent %s", searchT.Name)
		selectedTorrent = searchT
		fileId = tFile.ID
		break
	}

	if selectedTorrent == nil {
		return nil, "", fmt.Errorf("failed to find torrent")
	}

	return
}
