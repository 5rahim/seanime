package torrentstream

import (
	"cmp"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	torrentanalyzer "github.com/seanime-app/seanime/internal/torrents/analyzer"
	itorrent "github.com/seanime-app/seanime/internal/torrents/torrent"
	"slices"
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

func (r *Repository) findBestTorrent(media *anilist.BaseMedia, anizipMedia *anizip.Media, anizipEpisode *anizip.Episode, episodeNumber int) (*playbackTorrent, error) {

	r.logger.Debug().Msgf("torrentstream: Finding best torrent for %s, episode %d", media.GetTitleSafe(), episodeNumber)

	searchBatch := false
	// Search batch if not a movie and finished
	if media.IsMovie() && // Not a movie
		media.Status != nil && *media.Status == anilist.MediaStatusFinished { // Finished
		searchBatch = true
	}

	data, err := itorrent.NewSmartSearch(&itorrent.SmartSearchOptions{
		SmartSearchQueryOptions: itorrent.SmartSearchQueryOptions{
			SmartSearch:    lo.ToPtr(true),
			Query:          lo.ToPtr(""),
			EpisodeNumber:  &episodeNumber,
			Batch:          &searchBatch,
			Media:          media,
			AbsoluteOffset: lo.ToPtr(anizipMedia.GetOffset()),
			Resolution:     lo.ToPtr(r.settings.MustGet().PreferredResolution),
			Provider:       "animetosho",
			Best:           lo.ToPtr(true),
		},
		NyaaSearchCache:       r.nyaaSearchCache,
		AnimeToshoSearchCache: r.animeToshoSearchCache,
		AnizipCache:           r.anizipCache,
		Logger:                r.logger,
		MetadataProvider:      r.metadataProvider,
	})
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: Error searching torrents")
		return nil, err
	}

	if data == nil || len(data.Torrents) == 0 {
		r.logger.Error().Msg("torrentstream: No torrents found")
		return nil, ErrNoTorrentsFound
	}

	// Sort by seeders from highest to lowest
	slices.SortStableFunc(data.Torrents, func(a, b *itorrent.AnimeTorrent) int {
		return cmp.Compare(b.Seeders, a.Seeders)
	})

	r.logger.Debug().Msgf("torrentstream: Analyzing %d torrents", len(data.Torrents))

	// Go through the top 3 torrents
	// - For each torrent, add it, get the files, and check if it has the episode
	// - If it does, return the magnet link
	var selectedTorrent *torrent.Torrent
	var selectedFile *torrent.File
	try := 0
	for _, searchT := range data.Torrents {
		if try >= 3 {
			break
		}
		magnet, err := itorrent.ScrapeMagnet(searchT.Link)
		if err != nil {
			r.logger.Warn().Err(err).Msgf("torrentstream: Error scraping magnet link for %s", searchT.Link)
			try++
			continue
		}
		t, err := r.client.AddTorrent(magnet)
		if err != nil {
			r.logger.Warn().Err(err).Msgf("torrentstream: Error adding torrent %s", searchT.Link)
			try++
			continue
		}

		filepaths := lo.Map(t.Files(), func(f *torrent.File, _ int) string {
			return f.DisplayPath()
		})

		if len(filepaths) == 0 {
			r.logger.Error().Msg("torrentstream: No files found in the torrent")
			return nil, fmt.Errorf("no files found in the torrent")
		}

		// Create a new Torrent Analyzer
		analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
			Logger:               r.logger,
			Filepaths:            filepaths,
			Media:                media,
			AnilistClientWrapper: r.anilistClientWrapper,
		})

		// Analyze torrent files
		analysis, err := analyzer.AnalyzeTorrentFiles()
		if err != nil {
			r.logger.Warn().Err(err).Msg("torrentstream: Error analyzing torrent files")
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(t.InfoHash().AsString())
			}()
			try++
			continue
		}

		selectedFilesMap := analysis.GetCorrespondingMainFiles()
		selectedIndices := analysis.GetIndices(selectedFilesMap)

		// Check if analyzer found the episode
		if len(selectedIndices) == 0 || len(selectedIndices) > 1 {
			r.logger.Error().Msgf("torrentstream: Failed to auto-select episode from torrent %s", searchT.Link)
			// Remove torrent on failure
			go func() {
				_ = r.client.RemoveTorrent(t.InfoHash().AsString())
			}()
			try++
			continue
		}

		// Download the file and unselect the rest
		for i, f := range t.Files() {
			if i != selectedIndices[0] {
				f.SetPriority(torrent.PiecePriorityNone)
			}
		}
		t.Files()[selectedIndices[0]].SetPriority(torrent.PiecePriorityNow)
		selectedTorrent = t
		selectedFile = t.Files()[selectedIndices[0]]
		break
	}

	if selectedTorrent == nil {
		return nil, ErrNoEpisodeFound
	}

	ret := &playbackTorrent{
		Torrent: selectedTorrent,
		File:    selectedFile,
	}

	return ret, nil
}

// findBestTorrentFromManualSelection is like findBestTorrent but no need to search for the best torrent first
func (r *Repository) findBestTorrentFromManualSelection(torrentLink string, media *anilist.BaseMedia, episodeNumber int) (*playbackTorrent, error) {

	r.logger.Debug().Msgf("torrentstream: Analyzing torrent from %s for %s", torrentLink, media.GetTitleSafe())

	// First, add the torrent
	torrentId, err := itorrent.ScrapeMagnet(torrentLink)
	if err != nil {
		r.logger.Error().Err(err).Msgf("torrentstream: Error scraping magnet link for %s", torrentLink)
		return nil, fmt.Errorf("could not get magnet link from %s", torrentLink)
	}
	selectedTorrent, err := r.client.AddTorrent(torrentId)
	if err != nil {
		r.logger.Error().Err(err).Msgf("torrentstream: Error adding torrent %s", torrentLink)
		return nil, err
	}

	// If the torrent has only one file, return it
	if len(selectedTorrent.Files()) == 1 {
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
		Logger:               r.logger,
		Filepaths:            filepaths,
		Media:                media,
		AnilistClientWrapper: r.anilistClientWrapper,
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

	selectedFilesMap := analysis.GetCorrespondingMainFiles()
	selectedIndices := analysis.GetIndices(selectedFilesMap)

	// Check if analyzer found the episode
	if len(selectedIndices) == 0 || len(selectedIndices) > 1 {
		r.logger.Error().Msgf("torrentstream: Failed to auto-select episode from torrent %s", torrentLink)
		// Remove torrent on failure
		go func() {
			_ = r.client.RemoveTorrent(selectedTorrent.InfoHash().AsString())
		}()
		return nil, ErrNoEpisodeFound
	}

	// Download the file and unselect the rest
	for i, f := range selectedTorrent.Files() {
		if i != selectedIndices[0] {
			f.SetPriority(torrent.PiecePriorityNone)
		}
	}
	selectedTorrent.Files()[selectedIndices[0]].SetPriority(torrent.PiecePriorityNow)

	ret := &playbackTorrent{
		Torrent: selectedTorrent,
		File:    selectedTorrent.Files()[selectedIndices[0]],
	}

	return ret, nil
}
