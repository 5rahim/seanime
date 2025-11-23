package torrent_client

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/platforms/platform"
	torrent_analyzer "seanime/internal/torrents/analyzer"
	"seanime/internal/util"
	"time"
)

type (
	DeselectAndDownloadParams struct {
		Torrent          *hibiketorrent.AnimeTorrent
		FileIndices      []int // indices of the files to deselect
		Destination      string
		ShouldAddTorrent bool
	}
)

func (r *Repository) DeselectAndDownload(p *DeselectAndDownloadParams) error {
	if p.Torrent == nil || r.torrentRepository == nil {
		r.logger.Error().Msg("torrent client: torrent is nil (deselect)")
		return errors.New("torrent is nil")
	}

	if len(p.FileIndices) == 0 {
		r.logger.Error().Msg("torrent client: no file indices provided (deselect)")
		return errors.New("no file indices provided")
	}

	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(p.Torrent.Provider)
	if !ok {
		r.logger.Error().Str("provider", p.Torrent.Provider).Msg("torrent client: provider extension not found (simple select)")
		return errors.New("provider extension not found")
	}

	if p.ShouldAddTorrent {
		r.logger.Info().Msg("torrent client: adding torrent (simple select)")
		// Get magnet
		magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(p.Torrent)
		if err != nil {
			return err
		}
		// Add the torrent
		err = r.AddMagnets([]string{magnet}, p.Destination)
		if err != nil {
			return err
		}

		_, _ = r.GetFiles(p.Torrent.InfoHash)
	}

	// Pause the torrent
	_ = r.PauseTorrents([]string{p.Torrent.InfoHash})

	err := r.DeselectFiles(p.Torrent.InfoHash, p.FileIndices)
	if err != nil {
		r.logger.Err(err).Msg("torrent client: error while deselecting files (simple select)")
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return fmt.Errorf("error while deselecting files: %w", err)
	}

	// Unpause the torrent
	_ = r.ResumeTorrents([]string{p.Torrent.InfoHash})

	return nil
}

type (
	SmartSelectParams struct {
		Torrent          *hibiketorrent.AnimeTorrent
		EpisodeNumbers   []int
		Media            *anilist.CompleteAnime
		Destination      string
		ShouldAddTorrent bool
		PlatformRef      *util.Ref[platform.Platform]
	}
)

// SmartSelect will automatically the provided episode files from the torrent.
// If the torrent has not been added yet, set SmartSelect.ShouldAddTorrent to true.
// The torrent will NOT be removed if the selection fails.
func (r *Repository) SmartSelect(p *SmartSelectParams) error {
	if p.Media == nil || p.PlatformRef.IsAbsent() || r.torrentRepository == nil {
		r.logger.Error().Msg("torrent client: media or platform is nil (smart select)")
		return errors.New("media or anilist client wrapper is nil")
	}

	providerExtension, ok := r.torrentRepository.GetAnimeProviderExtension(p.Torrent.Provider)
	if !ok {
		r.logger.Error().Str("provider", p.Torrent.Provider).Msg("torrent client: provider extension not found (smart select)")
		return errors.New("provider extension not found")
	}

	if p.Media.IsMovieOrSingleEpisode() {
		return errors.New("smart select is not supported for movies or single-episode series")
	}

	if len(p.EpisodeNumbers) == 0 {
		r.logger.Error().Msg("torrent client: no episode numbers provided (smart select)")
		return errors.New("no episode numbers provided")
	}

	if p.ShouldAddTorrent {
		r.logger.Info().Msg("torrent client: adding torrent (smart select)")
		// Get magnet
		magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(p.Torrent)
		if err != nil {
			return err
		}
		// Add the torrent
		err = r.AddMagnets([]string{magnet}, p.Destination)
		if err != nil {
			return err
		}
	}

	filepaths, err := r.GetFiles(p.Torrent.InfoHash)
	if err != nil {
		r.logger.Err(err).Msg("torrent client: error getting files (smart select)")
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return fmt.Errorf("error getting files, torrent still added: %w", err)
	}

	// Pause the torrent
	err = r.PauseTorrents([]string{p.Torrent.InfoHash})
	if err != nil {
		r.logger.Err(err).Msg("torrent client: error while pausing torrent (smart select)")
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return fmt.Errorf("error while selecting files: %w", err)
	}

	// AnalyzeTorrentFiles the torrent files
	analyzer := torrent_analyzer.NewAnalyzer(&torrent_analyzer.NewAnalyzerOptions{
		Logger:              r.logger,
		Filepaths:           filepaths,
		Media:               p.Media,
		PlatformRef:         p.PlatformRef,
		MetadataProviderRef: r.metadataProviderRef,
	})

	r.logger.Debug().Msg("torrent client: analyzing torrent files (smart select)")

	analysis, err := analyzer.AnalyzeTorrentFiles()
	if err != nil {
		r.logger.Err(err).Msg("torrent client: error while analyzing torrent files (smart select)")
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return fmt.Errorf("error while analyzing torrent files: %w", err)
	}

	r.logger.Debug().Msg("torrent client: finished analyzing torrent files (smart select)")

	mainFiles := analysis.GetCorrespondingMainFiles()

	// find episode number duplicates
	dup := make(map[int]int) // map[episodeNumber]count
	for _, f := range mainFiles {
		if _, ok := dup[f.GetLocalFile().GetEpisodeNumber()]; ok {
			dup[f.GetLocalFile().GetEpisodeNumber()]++
		} else {
			dup[f.GetLocalFile().GetEpisodeNumber()] = 1
		}
	}
	dupCount := 0
	for _, count := range dup {
		if count > 1 {
			dupCount++
		}
	}
	if dupCount > 2 {
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return errors.New("failed to select files, can't tell seasons apart")
	}

	selectedFiles := make(map[int]*torrent_analyzer.File)
	selectedCount := 0
	for idx, f := range mainFiles {
		for _, ep := range p.EpisodeNumbers {
			if f.GetLocalFile().GetEpisodeNumber() == ep {
				selectedCount++
				selectedFiles[idx] = f
			}
		}
	}

	if selectedCount == 0 || selectedCount < len(p.EpisodeNumbers) {
		_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
		return errors.New("failed to select files, could not find the right season files")
	}

	indicesToRemove := analysis.GetUnselectedIndices(selectedFiles)

	if len(indicesToRemove) > 0 {
		// Deselect files
		err = r.DeselectFiles(p.Torrent.InfoHash, indicesToRemove)
		if err != nil {
			r.logger.Err(err).Msg("torrent client: error while deselecting files (smart select)")
			_ = r.RemoveTorrents([]string{p.Torrent.InfoHash})
			return fmt.Errorf("error while deselecting files: %w", err)
		}
	}

	time.Sleep(1 * time.Second)

	// Resume the torrent
	_ = r.ResumeTorrents([]string{p.Torrent.InfoHash})

	return nil
}
