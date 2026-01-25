package torrentstream

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	"seanime/internal/torrents/autoselect"
	"seanime/internal/util"
	"seanime/internal/util/torrentutil"

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
	torrentutil.PrioritizeDownloadPieces(t, file, r.logger)
}

func (r *Repository) findBestTorrent(media *anilist.CompleteAnime, aniDbEpisode string, episodeNumber int) (ret *playbackTorrent, err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/findBestTorrent", &err)

	r.logger.Debug().Msgf("torrentstream: Finding best torrent for %s, Episode %d", media.GetTitleSafe(), episodeNumber)

	if r.settings.IsAbsent() {
		return nil, fmt.Errorf("torrent streaming is disabled")
	}

	r.sendStateEvent(eventLoading, TLSStateSearchingTorrents)

	profile, found := db_bridge.FindAutoSelectProfile(r.db)
	if !found {
		resolution := r.settings.MustGet().PreferredResolution
		if resolution == "" {
			resolution = "1080p"
		}
		profile = &anime.AutoSelectProfile{
			Resolutions: []string{resolution},
			MinSeeders:  0,
		}
	}

	result, err := r.autoSelect.FindBestTorrent(
		context.Background(),
		media,
		episodeNumber,
		profile,
		autoselect.SelectionModeTorrent,
		nil,
		r.client,
		nil,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: Auto-select failed")
		if err.Error() == "no torrents found" {
			return nil, ErrNoTorrentsFound
		}
		return nil, err
	}

	if result.Torrent == nil || result.File == nil {
		return nil, ErrNoEpisodeFound
	}

	r.logger.Info().Msgf("torrentstream: Auto-selected torrent: %s", result.Torrent.Name())
	r.logger.Debug().Msgf("torrentstream: Selected file: %s", result.File.DisplayPath())

	// Set priority
	r.setPriorityDownloadStrategy(result.Torrent, result.File)

	ret = &playbackTorrent{
		Torrent: result.Torrent,
		File:    result.File,
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
			Logger:              r.logger,
			Filepaths:           filepaths,
			Media:               media,
			PlatformRef:         r.platformRef,
			MetadataProviderRef: r.metadataProviderRef,
			ForceMatch:          true,
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
