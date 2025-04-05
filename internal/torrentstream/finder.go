package torrentstream

import (
	"cmp"
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"slices"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/samber/lo"
)

var (
	ErrNoTorrentsFound = fmt.Errorf("no torrents found, please select manually")
	ErrNoEpisodeFound  = fmt.Errorf("could not select episode from torrents, please select manually")
)

type (
	playbackTorrent struct {
		Torrent *torrent.Torrent
		File    *torrent.File
	}
)

// setPriorityDownloadStrategy sets piece priorities for optimal streaming experience
// This helps to optimize initial buffering, seeking, and end-of-file playback
func (r *Repository) setPriorityDownloadStrategy(t *torrent.Torrent, file *torrent.File) {
	// Calculate file's pieces
	firstPieceIdx := file.Offset() * int64(t.NumPieces()) / t.Length()
	endPieceIdx := (file.Offset() + file.Length()) * int64(t.NumPieces()) / t.Length()

	// Prioritize more pieces at the beginning for faster initial loading (3% for beginning)
	numPiecesForStart := (endPieceIdx - firstPieceIdx + 1) * 3 / 100
	r.logger.Debug().Msgf("torrentstream: Setting high priority for first 3%% - pieces %d to %d (total %d)",
		firstPieceIdx, firstPieceIdx+numPiecesForStart, numPiecesForStart)
	for idx := firstPieceIdx; idx <= firstPieceIdx+numPiecesForStart; idx++ {
		t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
	}

	// // Also prioritize pieces in the middle of the file for seeking
	// midPieceIdx := (firstPieceIdx + endPieceIdx) / 2
	// numPiecesForMiddle := (endPieceIdx - firstPieceIdx + 1) * 2 / 100
	// r.logger.Debug().Msgf("torrentstream: Setting priority for middle pieces %d to %d",
	// 	midPieceIdx-numPiecesForMiddle/2, midPieceIdx+numPiecesForMiddle/2)
	// for idx := midPieceIdx - numPiecesForMiddle/2; idx <= midPieceIdx+numPiecesForMiddle/2; idx++ {
	// 	if idx >= 0 && int(idx) < t.NumPieces() {
	// 		t.Piece(int(idx)).SetPriority(torrent.PiecePriorityHigh)
	// 	}
	// }

	// Also prioritize the last few pieces
	numPiecesForEnd := (endPieceIdx - firstPieceIdx + 1) * 1 / 100
	r.logger.Debug().Msgf("torrentstream: Setting priority for last pieces %d to %d (total %d)",
		endPieceIdx-numPiecesForEnd, endPieceIdx, numPiecesForEnd)
	for idx := endPieceIdx - numPiecesForEnd; idx <= endPieceIdx; idx++ {
		if idx >= 0 && int(idx) < t.NumPieces() {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
	}

	// // Set some additional keyframe positions to high priority based on common seek points
	// // This helps when users seek to 25%, 50%, and 75% positions
	// for seekPercent := 25; seekPercent <= 75; seekPercent += 25 {
	// 	seekPieceIdx := firstPieceIdx + ((endPieceIdx - firstPieceIdx) * int64(seekPercent) / 100)
	// 	numPiecesForSeek := (endPieceIdx - firstPieceIdx + 1) / 100 // 1% of pieces at each seek point
	// 	r.logger.Debug().Msgf("torrentstream: Setting priority for %d%% seek point pieces %d to %d",
	// 		seekPercent, seekPieceIdx, seekPieceIdx+numPiecesForSeek)
	// 	for idx := seekPieceIdx; idx <= seekPieceIdx+numPiecesForSeek; idx++ {
	// 		if idx >= 0 && int(idx) < t.NumPieces() {
	// 			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityHigh)
	// 		}
	// 	}
	// }
}

func (r *Repository) findBestTorrent(media *anilist.CompleteAnime, aniDbEpisode string, episodeNumber int) (ret *playbackTorrent, err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/findBestTorrent", &err)

	r.logger.Debug().Msgf("torrentstream: Finding best torrent for %s, Episode %d", media.GetTitleSafe(), episodeNumber)

	providerId := itorrent.ProviderAnimeTosho // todo: get provider from settings
	fallbackProviderId := itorrent.ProviderNyaa

	// Get AnimeTosho provider extension
	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(providerId)
	if !ok {
		r.logger.Error().Str("provider", itorrent.ProviderAnimeTosho).Msg("torrentstream: AnimeTosho provider extension not found")
		return nil, fmt.Errorf("provider extension not found")
	}

	searchBatch := false
	// Search batch if not a movie and finished
	yearsSinceStart := 999
	if media.StartDate != nil && *media.StartDate.Year > 0 {
		yearsSinceStart = time.Now().Year() - *media.StartDate.Year // e.g. 2024 - 2020 = 4
	}
	if !media.IsMovie() && media.IsFinished() && yearsSinceStart > 4 {
		searchBatch = true
	}

	r.sendTorrentLoadingStatus(TLSStateSearchingTorrents, "")

	var data *itorrent.SearchData
	var currentProvider string = providerId
searchLoop:
	for {
		var err error
		data, err = r.torrentRepository.SearchAnime(context.Background(), itorrent.AnimeSearchOptions{
			Provider:      currentProvider,
			Type:          itorrent.AnimeSearchTypeSmart,
			Media:         media.ToBaseAnime(),
			Query:         "",
			Batch:         searchBatch,
			EpisodeNumber: episodeNumber,
			BestReleases:  false,
			Resolution:    r.settings.MustGet().PreferredResolution,
		})
		// If we are searching for batches, we don't want to return an error if no torrents are found
		// We will just search again without the batch flag
		if err != nil && !searchBatch {
			r.logger.Error().Err(err).Msg("torrentstream: Error searching torrents")

			// Try fallback provider if we're still on primary provider
			if currentProvider == providerId {
				r.logger.Debug().Msgf("torrentstream: Primary provider failed, trying fallback provider %s", fallbackProviderId)
				currentProvider = fallbackProviderId
				// Get fallback provider extension
				providerExtension, ok = r.torrentRepository.GetAnimeProviderExtension(currentProvider)
				if !ok {
					r.logger.Error().Str("provider", fallbackProviderId).Msg("torrentstream: Fallback provider extension not found")
					return nil, fmt.Errorf("fallback provider extension not found")
				}
				continue
			}

			return nil, err
		} else if err != nil {
			searchBatch = false
			continue
		}

		// This whole thing below just means that
		// If we are looking for batches, there should be at least 3 torrents found or the max seeders should be at least 15
		if searchBatch {
			nbFound := len(data.Torrents)
			seedersArr := lo.Map(data.Torrents, func(t *hibiketorrent.AnimeTorrent, _ int) int {
				return t.Seeders
			})
			if len(seedersArr) == 0 {
				searchBatch = false
				continue
			}
			maxSeeders := slices.Max(seedersArr)
			if maxSeeders >= 15 || nbFound > 2 {
				break searchLoop
			} else {
				searchBatch = false
			}
		} else {
			break searchLoop
		}
	}

	if data == nil || len(data.Torrents) == 0 {
		// Try fallback provider if we're still on primary provider
		if currentProvider == providerId {
			r.logger.Debug().Msgf("torrentstream: No torrents found with primary provider, trying fallback provider %s", fallbackProviderId)
			currentProvider = fallbackProviderId
			// Get fallback provider extension
			providerExtension, ok = r.torrentRepository.GetAnimeProviderExtension(currentProvider)
			if !ok {
				r.logger.Error().Str("provider", fallbackProviderId).Msg("torrentstream: Fallback provider extension not found")
				return nil, fmt.Errorf("fallback provider extension not found")
			}

			// Try searching with fallback provider (reset searchBatch)
			searchBatch = false
			if !media.IsMovie() && media.IsFinished() && yearsSinceStart > 4 {
				searchBatch = true
			}

			// Restart the search with fallback provider
			goto searchLoop
		}

		r.logger.Error().Msg("torrentstream: No torrents found")
		return nil, ErrNoTorrentsFound
	}

	// Sort by seeders from highest to lowest
	slices.SortStableFunc(data.Torrents, func(a, b *hibiketorrent.AnimeTorrent) int {
		return cmp.Compare(b.Seeders, a.Seeders)
	})

	// Trigger hook
	fetchedEvent := &TorrentStreamAutoSelectTorrentsFetchedEvent{
		Torrents: data.Torrents,
	}
	_ = hook.GlobalHookManager.OnTorrentStreamAutoSelectTorrentsFetched().Trigger(fetchedEvent)
	data.Torrents = fetchedEvent.Torrents

	r.logger.Debug().Msgf("torrentstream: Found %d torrents", len(data.Torrents))

	// Go through the top 3 torrents
	// - For each torrent, add it, get the files, and check if it has the episode
	// - If it does, return the magnet link
	var selectedTorrent *torrent.Torrent
	var selectedFile *torrent.File
	tries := 0

	for _, searchT := range data.Torrents {
		if tries >= 2 {
			break
		}
		r.sendTorrentLoadingStatus(TLSStateAddingTorrent, searchT.Name)
		r.logger.Trace().Msgf("torrentstream: Getting torrent magnet")
		magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(searchT)
		if err != nil {
			r.logger.Warn().Err(err).Msgf("torrentstream: Error scraping magnet link for %s", searchT.Link)
			tries++
			continue
		}
		r.logger.Debug().Msgf("torrentstream: Adding torrent %s from magnet", searchT.Link)

		t, err := r.client.AddTorrent(magnet)
		if err != nil {
			r.logger.Warn().Err(err).Msgf("torrentstream: Error adding torrent %s", searchT.Link)
			tries++
			continue
		}

		r.sendTorrentLoadingStatus(TLSStateCheckingTorrent, searchT.Name)

		// If the torrent has only one file, return it
		if len(t.Files()) == 1 {
			tFile := t.Files()[0]
			tFile.Download()
			r.setPriorityDownloadStrategy(t, tFile)
			r.logger.Debug().Msgf("torrentstream: Found single file torrent: %s", tFile.DisplayPath())

			return &playbackTorrent{
				Torrent: t,
				File:    tFile,
			}, nil
		}

		r.sendTorrentLoadingStatus(TLSStateSelectingFile, searchT.Name)

		// DEVNOTE: The gap between adding the torrent and file analysis causes some pieces to be downloaded
		// We currently can't Pause/Resume torrents so :shrug:

		filepaths := lo.Map(t.Files(), func(f *torrent.File, _ int) string {
			return f.DisplayPath()
		})

		if len(filepaths) == 0 {
			r.logger.Error().Msg("torrentstream: No files found in the torrent")
			return nil, fmt.Errorf("no files found in the torrent")
		}

		// Create a new Torrent Analyzer
		analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
			Logger:           r.logger,
			Filepaths:        filepaths,
			Media:            media,
			Platform:         r.platform,
			MetadataProvider: r.metadataProvider,
			ForceMatch:       true,
		})

		r.logger.Debug().Msgf("torrentstream: Analyzing torrent %s", searchT.Link)

		// Analyze torrent files
		analysis, err := analyzer.AnalyzeTorrentFiles()
		if err != nil {
			r.logger.Warn().Err(err).Msg("torrentstream: Error analyzing torrent files")
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(t.InfoHash().AsString())
			}()
			tries++
			continue
		}

		analysisFile, found := analysis.GetFileByAniDBEpisode(aniDbEpisode)
		// Check if analyzer found the episode
		if !found {
			r.logger.Error().Msgf("torrentstream: Failed to auto-select episode from torrent %s", searchT.Link)
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(t.InfoHash().AsString())
			}()
			tries++
			continue
		}

		r.logger.Debug().Msgf("torrentstream: Found corresponding file for episode %s: %s", aniDbEpisode, analysisFile.GetLocalFile().Name)

		// Download the file and unselect the rest
		for i, f := range t.Files() {
			if i != analysisFile.GetIndex() {
				f.SetPriority(torrent.PiecePriorityNone)
			}
		}
		tFile := t.Files()[analysisFile.GetIndex()]
		r.logger.Debug().Msgf("torrentstream: Selecting file %s", tFile.DisplayPath())
		r.setPriorityDownloadStrategy(t, tFile)

		selectedTorrent = t
		selectedFile = tFile
		break
	}

	if selectedTorrent == nil {
		return nil, ErrNoEpisodeFound
	}

	ret = &playbackTorrent{
		Torrent: selectedTorrent,
		File:    selectedFile,
	}

	return ret, nil
}

// findBestTorrentFromManualSelection is like findBestTorrent but no need to search for the best torrent first
func (r *Repository) findBestTorrentFromManualSelection(t *hibiketorrent.AnimeTorrent, media *anilist.CompleteAnime, aniDbEpisode string, chosenFileIndex *int) (*playbackTorrent, error) {

	r.logger.Debug().Msgf("torrentstream: Analyzing torrent from %s for %s", t.Link, media.GetTitleSafe())

	// Get the torrent's provider extension
	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(t.Provider)
	if !ok {
		r.logger.Error().Str("provider", t.Provider).Msg("torrentstream: provider extension not found")
		return nil, fmt.Errorf("provider extension not found")
	}

	// First, add the torrent
	magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(t)
	if err != nil {
		r.logger.Error().Err(err).Msgf("torrentstream: Error scraping magnet link for %s", t.Link)
		return nil, fmt.Errorf("could not get magnet link from %s", t.Link)
	}
	selectedTorrent, err := r.client.AddTorrent(magnet)
	if err != nil {
		r.logger.Error().Err(err).Msgf("torrentstream: Error adding torrent %s", t.Link)
		return nil, err
	}

	// If the torrent has only one file, return it
	if len(selectedTorrent.Files()) == 1 {
		tFile := selectedTorrent.Files()[0]
		tFile.Download()
		r.setPriorityDownloadStrategy(selectedTorrent, tFile)
		r.logger.Debug().Msgf("torrentstream: Found single file torrent: %s", tFile.DisplayPath())

		return &playbackTorrent{
			Torrent: selectedTorrent,
			File:    tFile,
		}, nil
	}

	var fileIndex int

	// If the file index is already selected
	if chosenFileIndex != nil {

		fileIndex = *chosenFileIndex

	} else {

		// We know the torrent has multiple files, so we'll need to analyze it
		filepaths := lo.Map(selectedTorrent.Files(), func(f *torrent.File, _ int) string {
			return f.DisplayPath()
		})

		if len(filepaths) == 0 {
			r.logger.Error().Msg("torrentstream: No files found in the torrent")
			return nil, fmt.Errorf("no files found in the torrent")
		}

		// Create a new Torrent Analyzer
		analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
			Logger:           r.logger,
			Filepaths:        filepaths,
			Media:            media,
			Platform:         r.platform,
			MetadataProvider: r.metadataProvider,
			ForceMatch:       true,
		})

		// Analyze torrent files
		analysis, err := analyzer.AnalyzeTorrentFiles()
		if err != nil {
			r.logger.Warn().Err(err).Msg("torrentstream: Error analyzing torrent files")
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(selectedTorrent.InfoHash().AsString())
			}()
			return nil, err
		}

		analysisFile, found := analysis.GetFileByAniDBEpisode(aniDbEpisode)
		// Check if analyzer found the episode
		if !found {
			r.logger.Error().Msgf("torrentstream: Failed to auto-select episode from torrent %s", selectedTorrent.Info().Name)
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(selectedTorrent.InfoHash().AsString())
			}()
			return nil, ErrNoEpisodeFound
		}

		r.logger.Debug().Msgf("torrentstream: Found corresponding file for episode %s: %s", aniDbEpisode, analysisFile.GetLocalFile().Name)

		fileIndex = analysisFile.GetIndex()

	}

	// Download the file and unselect the rest
	for i, f := range selectedTorrent.Files() {
		if i != fileIndex {
			f.SetPriority(torrent.PiecePriorityNone)
		}
	}
	//selectedTorrent.Files()[fileIndex].SetPriority(torrent.PiecePriorityNormal)
	r.logger.Debug().Msgf("torrentstream: Selected torrent %s", selectedTorrent.Files()[fileIndex].DisplayPath())

	tFile := selectedTorrent.Files()[fileIndex]
	tFile.Download()
	r.setPriorityDownloadStrategy(selectedTorrent, tFile)

	ret := &playbackTorrent{
		Torrent: selectedTorrent,
		File:    selectedTorrent.Files()[fileIndex],
	}

	return ret, nil
}
