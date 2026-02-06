package autoselect

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/debrid/debrid"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	torrentanalyzer "seanime/internal/torrents/analyzer"
	"seanime/internal/util"
	"strconv"

	"github.com/anacrolix/torrent"
	"github.com/samber/lo"
)

const (
	MaxTorrentCandidatesToCheck = 3
	MaxAnalyzedTorrents         = 3
)

func (s *AutoSelect) selectFile(
	media *anilist.CompleteAnime,
	episodeNumber int,
	torrents []*hibiketorrent.AnimeTorrent,
	mode SelectionMode,
	torrentClient TorrentClient,
	debridClient debrid.Provider,
) (*Result, error) {
	// Go through the top torrents
	limit := MaxTorrentCandidatesToCheck
	if len(torrents) < limit {
		limit = len(torrents)
	}

	analyzedCount := 0

	for i := 0; i < limit; i++ {
		t := torrents[i]

		if analyzedCount >= MaxAnalyzedTorrents {
			break
		}

		s.logger.Debug().Msgf("autoselect: Checking torrent candidate: %s", t.Name)
		s.log(fmt.Sprintf("Checking torrent candidate: %s", t.Name))

		providerExt, ok := s.torrentRepository.GetAnimeProviderExtension(t.Provider)
		if !ok {
			s.logger.Warn().Str("provider", t.Provider).Msg("autoselect: Provider not found")
			continue
		}

		var res *Result
		var err error

		switch mode {
		case SelectionModeDebrid:
			if debridClient != nil {
				res, err = s.selectFileFromDebrid(media, episodeNumber, t, providerExt, debridClient)
			} else {
				s.logger.Error().Msg("autoselect: Debrid client is nil but mode is Debrid")
				continue
			}
		case SelectionModeTorrent:
			if torrentClient != nil {
				res, err = s.selectFileFromTorrentClient(media, episodeNumber, t, providerExt, torrentClient)
			} else {
				s.logger.Error().Msg("autoselect: Torrent client is nil but mode is Torrent")
				continue
			}
		}

		if err == nil && res != nil {
			return res, nil
		}

		if err != nil {
			s.logger.Warn().Err(err).Msgf("autoselect: Could not select file for %s", t.Name)
		}

		// Count the analysis attempt if we actually tried
		analyzedCount++
	}

	return nil, ErrNoFileFound
}

func (s *AutoSelect) selectFileFromTorrentClient(
	media *anilist.CompleteAnime,
	episodeNumber int,
	t *hibiketorrent.AnimeTorrent,
	providerExt extension.AnimeTorrentProviderExtension,
	client TorrentClient,
) (res *Result, err error) {
	defer util.HandlePanicInModuleWithError("autoselect/selectFileFromTorrentClient", &err)

	s.logger.Trace().Msgf("autoselect: Getting torrent magnet")
	magnet, err := providerExt.GetProvider().GetTorrentMagnetLink(t)
	if err != nil {
		s.logger.Warn().Err(err).Msgf("autoselect: Error scraping magnet link for %s", t.Link)
		return nil, err
	}

	s.logger.Debug().Msgf("autoselect: Adding torrent %s from magnet", t.Link)

	addedTorrent, err := client.AddTorrent(magnet)
	if err != nil {
		s.logger.Warn().Err(err).Msgf("autoselect: Error adding torrent %s", t.Link)
		return nil, err
	}

	// Override magnet link
	t.MagnetLink = magnet

	// Only one file, use it
	if len(addedTorrent.Files()) == 1 {
		tFile := addedTorrent.Files()[0]
		addedTorrent.DownloadAll()
		return &Result{
			Torrent:         addedTorrent,
			File:            tFile,
			OriginalTorrent: t,
		}, nil
	}

	// get file paths
	filepaths := lo.Map(addedTorrent.Files(), func(f *torrent.File, _ int) string {
		return f.DisplayPath()
	})

	if len(filepaths) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	// Remove the torrent
	cancel := func() {
		go func() {
			_ = client.RemoveTorrent(addedTorrent.InfoHash().AsString())
		}()
	}

	analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
		Logger:              s.logger,
		Filepaths:           filepaths,
		Media:               media,
		PlatformRef:         s.platform,
		MetadataProviderRef: s.metadataProvider,
		ForceMatch:          true,
	})

	analysis, err := analyzer.AnalyzeTorrentFiles()
	if err != nil {
		cancel()
		return nil, err
	}

	analysisFile, found := analysis.GetFileByAniDBEpisode(strconv.Itoa(episodeNumber))
	if !found {
		cancel()
		return nil, fmt.Errorf("episode not found")
	}

	// Download the file and unselect the rest
	for i, f := range addedTorrent.Files() {
		if i != analysisFile.GetIndex() {
			f.SetPriority(torrent.PiecePriorityNone)
		}
	}
	tFile := addedTorrent.Files()[analysisFile.GetIndex()]
	tFile.Download()

	s.log(fmt.Sprintf("Selected file: %s", tFile.DisplayPath()))

	return &Result{
		Torrent:         addedTorrent,
		File:            tFile,
		AnalysisFile:    analysisFile,
		OriginalTorrent: t,
	}, nil
}

func (s *AutoSelect) selectFileFromDebrid(
	media *anilist.CompleteAnime,
	episodeNumber int,
	t *hibiketorrent.AnimeTorrent,
	providerExt extension.AnimeTorrentProviderExtension,
	client debrid.Provider,
) (*Result, error) {

	s.logger.Trace().Msgf("autoselect: Getting torrent magnet")
	magnet, err := providerExt.GetProvider().GetTorrentMagnetLink(t)
	if err != nil {
		s.logger.Warn().Err(err).Msgf("autoselect: Error scraping magnet link for %s", t.Link)
		return nil, err
	}

	// Override magnet link
	t.MagnetLink = magnet

	s.logger.Debug().Msgf("autoselect: Checking debrid info for %s", t.Link)

	info, err := client.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: magnet,
		InfoHash:   t.InfoHash,
	})
	if err != nil {
		s.logger.Warn().Err(err).Msgf("autoselect: Error getting debrid info %s", t.Link)
		return nil, err
	}

	filepaths := lo.Map(info.Files, func(f *debrid.TorrentItemFile, _ int) string {
		return f.Path
	})

	if len(filepaths) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	analyzer := torrentanalyzer.NewAnalyzer(&torrentanalyzer.NewAnalyzerOptions{
		Logger:              s.logger,
		Filepaths:           filepaths,
		Media:               media,
		PlatformRef:         s.platform,
		MetadataProviderRef: s.metadataProvider,
		ForceMatch:          true,
	})

	analysis, err := analyzer.AnalyzeTorrentFiles()
	if err != nil {
		return nil, err
	}

	analysisFile, found := analysis.GetFileByAniDBEpisode(strconv.Itoa(episodeNumber))
	if !found {
		return nil, fmt.Errorf("episode not found")
	}

	tFile := info.Files[analysisFile.GetIndex()]
	s.logger.Debug().Msgf("autoselect: Selected debrid file %s", tFile.Name)
	s.log(fmt.Sprintf("Selected debrid file: %s", tFile.Name))

	return &Result{
		DebridTorrent:   info,
		DebridFileID:    tFile.ID,
		AnalysisFile:    analysisFile,
		OriginalTorrent: t,
	}, nil
}
