package torrentstream

import (
	"cmp"
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/anacrolix/torrent"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"slices"
	"time"
)

var (
	ErrNoTorrentsFound = fmt.Errorf("no torrents found")
	ErrNoEpisodeFound  = fmt.Errorf("could not select episode from torrents")
)

type (
	playbackTorrent struct {
		Torrent *torrent.Torrent
		File    *torrent.File
	}
)

func (r *Repository) findBestTorrent(media *anilist.CompleteAnime, anizipMedia *anizip.Media, anizipEpisode *anizip.Episode, episodeNumber int) (ret *playbackTorrent, err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/findBestTorrent", &err)

	r.logger.Debug().Msgf("torrentstream: Finding best torrent for %s, Episode %d", media.GetTitleSafe(), episodeNumber)

	providerId := itorrent.ProviderAnimeTosho // todo: get provider from settings

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
searchLoop:
	for {
		var err error
		data, err = r.torrentRepository.SearchAnime(itorrent.AnimeSearchOptions{
			Provider:      providerId,
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
			return nil, err
		} else if err != nil {
			searchBatch = false
			continue
		}

		// This whole thing below just means that
		// If we are looking for batches, there should be at least 3 torrents found or the max seeders should be at least 15
		if searchBatch == true {
			nbFound := len(data.Torrents)
			seedersArr := lo.Map(data.Torrents, func(t *hibiketorrent.AnimeTorrent, _ int) int {
				return t.Seeders
			})
			if seedersArr == nil || len(seedersArr) == 0 {
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
		r.logger.Error().Msg("torrentstream: No torrents found")
		return nil, ErrNoTorrentsFound
	}

	// Sort by seeders from highest to lowest
	slices.SortStableFunc(data.Torrents, func(a, b *hibiketorrent.AnimeTorrent) int {
		return cmp.Compare(b.Seeders, a.Seeders)
	})

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
			t.DownloadAll()
			firstPieceIdx := t.Files()[0].Offset() * int64(t.NumPieces()) / t.Length()
			endPieceIdx := (t.Files()[0].Offset() + t.Length()) * int64(t.NumPieces()) / t.Length()
			for idx := firstPieceIdx; idx <= endPieceIdx*5/100; idx++ {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
			}
			r.logger.Debug().Msgf("torrentstream: Found single file torrent: %s", t.Files()[0].DisplayPath())

			return &playbackTorrent{
				Torrent: t,
				File:    t.Files()[0],
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
			Logger:    r.logger,
			Filepaths: filepaths,
			Media:     media,
			Platform:  r.platform,
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

		analysisFile, found := analysis.GetFileByAniDBEpisode(anizipEpisode.Episode)
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

		r.logger.Debug().Msgf("torrentstream: Found corresponding file for episode %s: %s", anizipEpisode.Episode, analysisFile.GetLocalFile().Name)

		// Download the file and unselect the rest
		for i, f := range t.Files() {
			if i != analysisFile.GetIndex() {
				f.SetPriority(torrent.PiecePriorityNone)
			}
		}
		tFile := t.Files()[analysisFile.GetIndex()]
		// Select the first 5% of the pieces
		firstPieceIdx := tFile.Offset() * int64(t.NumPieces()) / t.Length()
		endPieceIdx := (tFile.Offset() + tFile.Length()) * int64(t.NumPieces()) / t.Length()
		for idx := firstPieceIdx; idx <= endPieceIdx*5/100; idx++ {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
		r.logger.Debug().Msgf("torrentstream: Selected torrent %s", tFile.DisplayPath())
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
func (r *Repository) findBestTorrentFromManualSelection(t *hibiketorrent.AnimeTorrent, media *anilist.CompleteAnime, anizipEpisode *anizip.Episode, episodeNumber int) (*playbackTorrent, error) {

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
		selectedTorrent.DownloadAll()
		firstPieceIdx := selectedTorrent.Files()[0].Offset() * int64(selectedTorrent.NumPieces()) / selectedTorrent.Length()
		endPieceIdx := (selectedTorrent.Files()[0].Offset() + selectedTorrent.Length()) * int64(selectedTorrent.NumPieces()) / selectedTorrent.Length()
		for idx := firstPieceIdx; idx <= endPieceIdx*5/100; idx++ {
			selectedTorrent.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
		return &playbackTorrent{
			Torrent: selectedTorrent,
			File:    selectedTorrent.Files()[0],
		}, nil
	}

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
		Logger:    r.logger,
		Filepaths: filepaths,
		Media:     media,
		Platform:  r.platform,
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

	analysisFile, found := analysis.GetFileByAniDBEpisode(anizipEpisode.Episode)
	// Check if analyzer found the episode
	if !found {
		r.logger.Error().Msgf("torrentstream: Failed to auto-select episode from torrent %s", selectedTorrent.Info().Name)
		// Remove torrent on failure
		go func() {
			_ = r.client.RemoveTorrent(selectedTorrent.InfoHash().AsString())
		}()
		return nil, ErrNoEpisodeFound
	}

	r.logger.Debug().Msgf("torrentstream: Found corresponding file for episode %s: %s", anizipEpisode.Episode, analysisFile.GetLocalFile().Name)

	// Download the file and unselect the rest
	for i, f := range selectedTorrent.Files() {
		if i != analysisFile.GetIndex() {
			f.SetPriority(torrent.PiecePriorityNone)
		}
	}
	//selectedTorrent.Files()[analysisFile.GetIndex()].SetPriority(torrent.PiecePriorityNormal)
	r.logger.Debug().Msgf("torrentstream: Selected torrent %s", selectedTorrent.Files()[analysisFile.GetIndex()].DisplayPath())

	tFile := selectedTorrent.Files()[analysisFile.GetIndex()]
	tFile.Download()
	// Select the first 5% of the pieces
	firstPieceIdx := tFile.Offset() * int64(selectedTorrent.NumPieces()) / selectedTorrent.Length()
	endPieceIdx := (tFile.Offset() + tFile.Length()) * int64(selectedTorrent.NumPieces()) / selectedTorrent.Length()
	for idx := firstPieceIdx; idx <= endPieceIdx*5/100; idx++ {
		selectedTorrent.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
	}

	ret := &playbackTorrent{
		Torrent: selectedTorrent,
		File:    selectedTorrent.Files()[analysisFile.GetIndex()],
	}

	return ret, nil
}
