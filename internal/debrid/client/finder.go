package debrid_client

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/debrid/debrid"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	"seanime/internal/torrents/autoselect"
	"seanime/internal/util"
	"strconv"

	"github.com/samber/lo"
)

type (
	playbackTorrent struct {
		torrent  *hibiketorrent.AnimeTorrent
		fileId   string
		filepath string
	}
)

func (r *Repository) findBestTorrent(provider debrid.Provider, media *anilist.CompleteAnime, episodeNumber int) (ret *playbackTorrent, err error) {

	defer util.HandlePanicInModuleWithError("debridstream/findBestTorrent", &err)

	r.logger.Debug().Msgf("debridstream: Finding best torrent for %s, Episode %d", media.GetTitleSafe(), episodeNumber)

	profile, found := db_bridge.FindAutoSelectProfile(r.db)
	if !found {
		resolution := r.settings.StreamPreferredResolution
		if resolution == "" {
			resolution = "1080p"
		}
		profile = &anime.AutoSelectProfile{
			Resolutions: []string{resolution},
			MinSeeders:  0,
		}
	}

	// Prioritize cached torrents
	postSearchSort := func(torrents []*hibiketorrent.AnimeTorrent) []*autoselect.TorrentWithCacheStatus {
		if len(torrents) == 0 {
			return []*autoselect.TorrentWithCacheStatus{}
		}

		// Check cached status
		hashes := make([]string, 0)
		for _, t := range torrents {
			if t.InfoHash != "" {
				hashes = append(hashes, t.InfoHash)
			}
		}

		instantAvail := provider.GetInstantAvailability(hashes)

		result := make([]*autoselect.TorrentWithCacheStatus, 0, len(torrents))
		for _, t := range torrents {
			_, isCached := instantAvail[t.InfoHash]
			result = append(result, &autoselect.TorrentWithCacheStatus{
				Torrent:  t,
				IsCached: isCached,
			})
		}

		return result
	}

	result, err := r.autoSelect.FindBestTorrent(
		context.Background(),
		media,
		episodeNumber,
		profile,
		autoselect.SelectionModeDebrid,
		postSearchSort,
		nil,
		provider,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("debridstream: Auto-select failed")
		if err.Error() == "no torrents found" {
			return nil, fmt.Errorf("no torrents found, please select manually")
		}
		return nil, err
	}

	if result.DebridTorrent == nil {
		return nil, fmt.Errorf("failed to find torrent")
	}

	// Log success
	r.logger.Info().Msgf("debridstream: Auto-selected torrent: %s", result.OriginalTorrent.Name)
	r.logger.Debug().Msgf("debridstream: Selected file ID: %s", result.DebridFileID)

	ret = &playbackTorrent{
		torrent:  result.OriginalTorrent,
		fileId:   result.DebridFileID,
		filepath: result.AnalysisFile.GetPath(),
	}

	return ret, nil
}

// findBestTorrentFromManualSelection is like findBestTorrent but for a pre-selected torrent
func (r *Repository) findBestTorrentFromManualSelection(provider debrid.Provider, t *hibiketorrent.AnimeTorrent, media *anilist.CompleteAnime, episodeNumber int, chosenFileIndex *int) (ret *playbackTorrent, err error) {

	r.logger.Debug().Msgf("debridstream: Analyzing torrent from %s for %s", t.Link, media.GetTitleSafe())

	// Get the torrent's provider extension
	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(t.Provider)
	if !ok {
		r.logger.Error().Str("provider", t.Provider).Msg("debridstream: provider extension not found")
		return nil, fmt.Errorf("provider extension not found")
	}

	// Check if the torrent is cached
	if t.InfoHash != "" {
		instantAvail := provider.GetInstantAvailability([]string{t.InfoHash})
		if len(instantAvail) == 0 {
			r.logger.Warn().Msg("debridstream: Torrent is not cached")
			// We'll still continue since the user specifically selected this torrent
		}
	}

	// Get the magnet link
	magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(t)
	if err != nil {
		r.logger.Error().Err(err).Msgf("debridstream: Error scraping magnet link for %s", t.Link)
		return nil, fmt.Errorf("could not get magnet link from %s", t.Link)
	}

	// Set the magnet link
	t.MagnetLink = magnet

	// Get the torrent info from the debrid provider
	info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: t.MagnetLink,
		InfoHash:   t.InfoHash,
	})
	if err != nil {
		r.logger.Error().Err(err).Msgf("debridstream: Error adding torrent %s", t.Link)
		return nil, err
	}

	// If the torrent has only one file, return it
	if len(info.Files) == 1 {
		return &playbackTorrent{torrent: t, fileId: info.Files[0].ID, filepath: info.Files[0].Path}, nil
	}

	var fileIndex int

	// If the file index is already selected
	if chosenFileIndex != nil {
		fileIndex = *chosenFileIndex
	} else {
		// We know the torrent has multiple files, so we'll need to analyze it
		filepaths := lo.Map(info.Files, func(f *debrid.TorrentItemFile, _ int) string {
			return f.Path
		})

		if len(filepaths) == 0 {
			r.logger.Error().Msg("debridstream: No files found in the torrent")
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
			r.logger.Warn().Err(err).Msg("debridstream: Error analyzing torrent files")
			return nil, err
		}

		analysisFile, found := analysis.GetFileByAniDBEpisode(strconv.Itoa(episodeNumber))
		// Check if analyzer found the episode
		if !found {
			r.logger.Error().Msgf("debridstream: Failed to auto-select episode from torrent %s", t.Name)
			return nil, fmt.Errorf("could not find episode %d in torrent", episodeNumber)
		}

		r.logger.Debug().Msgf("debridstream: Found corresponding file for episode %s: %s", strconv.Itoa(episodeNumber), analysisFile.GetLocalFile().Name)

		fileIndex = analysisFile.GetIndex()
	}

	tFile := info.Files[fileIndex]
	r.logger.Debug().Str("file", util.SpewT(tFile)).Msgf("debridstream: Selected file %s", tFile.Name)
	r.logger.Debug().Msgf("debridstream: Selected torrent %s", t.Name)

	return &playbackTorrent{torrent: t, fileId: tFile.ID, filepath: tFile.Path}, nil
}
