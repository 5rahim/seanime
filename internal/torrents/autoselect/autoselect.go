package autoselect

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/debrid/debrid"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/platform"
	torrent_analyzer "seanime/internal/torrents/analyzer"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"

	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
)

type SelectionMode string

const (
	SelectionModeTorrent SelectionMode = "torrent"
	SelectionModeDebrid  SelectionMode = "debrid"
)

var (
	ErrNoFileFound = fmt.Errorf("no file found")
)

type (
	AutoSelect struct {
		logger            *zerolog.Logger
		torrentRepository *itorrent.Repository
		metadataProvider  *util.Ref[metadata_provider.Provider]
		platform          *util.Ref[platform.Platform]
		onEvent           func(string)
	}

	NewAutoSelectOptions struct {
		Logger            *zerolog.Logger
		TorrentRepository *itorrent.Repository
		MetadataProvider  *util.Ref[metadata_provider.Provider]
		Platform          *util.Ref[platform.Platform]
		OnEvent           func(string)
	}

	Result struct {
		Torrent         *torrent.Torrent // For torrent client
		File            *torrent.File    // For torrent client
		AnalysisFile    *torrent_analyzer.File
		DebridTorrent   *debrid.TorrentInfo         // For debrid
		DebridFileID    string                      // For debrid
		OriginalTorrent *hibiketorrent.AnimeTorrent // The original torrent object
	}
)

func New(opts *NewAutoSelectOptions) *AutoSelect {
	return &AutoSelect{
		logger:            opts.Logger,
		torrentRepository: opts.TorrentRepository,
		metadataProvider:  opts.MetadataProvider,
		platform:          opts.Platform,
		onEvent:           opts.OnEvent,
	}
}

type TorrentClient interface {
	AddTorrent(magnet string) (*torrent.Torrent, error)
	RemoveTorrent(hash string) error
}

type DebridClient interface {
	GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error)
}

func (s *AutoSelect) FindBestTorrent(
	ctx context.Context,
	media *anilist.CompleteAnime,
	episodeNumber int,
	profile *anime.AutoSelectProfile,
	mode SelectionMode,
	postSearchSort func([]*hibiketorrent.AnimeTorrent) []*TorrentWithCacheStatus,
	torrentClient TorrentClient,
	debridClient debrid.Provider,
) (*Result, error) {

	// 1. Search
	s.log("Searching for torrents")
	torrents, err := s.search(ctx, media, episodeNumber, profile)
	if err != nil {
		s.log(fmt.Sprintf("Search failed: %v", err))
		return nil, err
	}

	// 2. Filter & sort
	s.log("Filtering and sorting candidates")
	torrents = s.filterAndSort(torrents, profile, postSearchSort)

	// 3. Select file (iterate top 3)
	s.log("Selecting best file from top candidates")
	return s.selectFile(media, episodeNumber, torrents, mode, torrentClient, debridClient)
}

func (s *AutoSelect) log(msg string) {
	if s.onEvent != nil {
		s.onEvent(msg)
	}
}
