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
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
)

type SelectionMode string

const (
	SelectionModeTorrent SelectionMode = "torrent"
	SelectionModeDebrid  SelectionMode = "debrid"
)

var (
	ErrNoFileFound     = fmt.Errorf("no file found")
	ErrNoTorrentsFound = fmt.Errorf("no torrents found")
)

type (
	AutoSelect struct {
		logger            *zerolog.Logger
		torrentRepository *itorrent.Repository
		metadataProvider  *util.Ref[metadata_provider.Provider]
		platform          *util.Ref[platform.Platform]
		onEvent           func(string)
		onStatus          func(StreamAutoSelectStatusPayload)
		statusMu          sync.Mutex
	}

	NewAutoSelectOptions struct {
		Logger            *zerolog.Logger
		TorrentRepository *itorrent.Repository
		MetadataProvider  *util.Ref[metadata_provider.Provider]
		Platform          *util.Ref[platform.Platform]
		OnEvent           func(string)
		OnStatus          func(StreamAutoSelectStatusPayload)
	}

	Result struct {
		Torrent         *torrent.Torrent // For torrent client
		File            *torrent.File    // For torrent client
		AnalysisFile    *torrent_analyzer.File
		DebridTorrent   *debrid.TorrentInfo         // For debrid
		DebridFileID    string                      // For debrid
		OriginalTorrent *hibiketorrent.AnimeTorrent // The original torrent object
	}

	StreamAutoSelectStatusPayload struct {
		Active       bool                  `json:"active"`
		MediaTitle   string                `json:"mediaTitle"`
		Episode      int                   `json:"episode"`
		Resolutions  []string              `json:"resolutions"`
		MinSeeders   int                   `json:"minSeeders"`
		Step         string                `json:"step"`       // "searching", "ranking", "analyzing", "completed"
		StepDetail   string                `json:"stepDetail"` // Description of the current action
		Candidates   []AutoSelectCandidate `json:"candidates"`
		SelectedFile string                `json:"selectedFile"`
	}

	AutoSelectCandidate struct {
		Name     string `json:"name"`
		Provider string `json:"provider"`
		Seeders  int    `json:"seeders"`
		Score    int    `json:"score"`
		Status   string `json:"status"` // "waiting", "analyzing", "skipped", "selected"
	}
)

func New(opts *NewAutoSelectOptions) *AutoSelect {
	return &AutoSelect{
		logger:            opts.Logger,
		torrentRepository: opts.TorrentRepository,
		metadataProvider:  opts.MetadataProvider,
		platform:          opts.Platform,
		onEvent:           opts.OnEvent,
		onStatus:          opts.OnStatus,
	}
}

type TorrentClient interface {
	AddTorrent(ctx context.Context, magnet string) (*torrent.Torrent, error)
	RemoveTorrent(hash string) error
}

type DebridClient interface {
	GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error)
}

type contextKey string

const statusKey contextKey = "autoselect-status"

const freshSearchKey contextKey = "autoselect-fresh-search"

func (s *AutoSelect) updateStatus(status StreamAutoSelectStatusPayload) {
	if s.onStatus != nil {
		s.onStatus(status)
	}
}

func (s *AutoSelect) updateStep(ctx context.Context, step string, detail string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	if status, ok := ctx.Value(statusKey).(*StreamAutoSelectStatusPayload); ok {
		status.Step = step
		status.StepDetail = detail
		s.updateStatus(*status)
	}
}

func (s *AutoSelect) updateCandidates(ctx context.Context, list []AutoSelectCandidate) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	if status, ok := ctx.Value(statusKey).(*StreamAutoSelectStatusPayload); ok {
		status.Candidates = list
		s.updateStatus(*status)
	}
}

func (s *AutoSelect) updateCandidateStatus(ctx context.Context, name string, statusStr string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	if status, ok := ctx.Value(statusKey).(*StreamAutoSelectStatusPayload); ok {
		for i, c := range status.Candidates {
			if c.Name == name {
				status.Candidates[i].Status = statusStr
				s.updateStatus(*status)
				break
			}
		}
	}
}

func (s *AutoSelect) completeStatus(ctx context.Context, selectedFile string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	if status, ok := ctx.Value(statusKey).(*StreamAutoSelectStatusPayload); ok {
		status.Step = "completed"
		status.StepDetail = "Best file selected!"
		status.SelectedFile = selectedFile
		s.updateStatus(*status)
	}
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
) (res *Result, err error) {

	resolutions := []string{"1080p"}
	minSeeders := 0
	if profile != nil {
		if len(profile.Resolutions) > 0 {
			resolutions = profile.Resolutions
		}
		minSeeders = profile.MinSeeders
	}

	status := StreamAutoSelectStatusPayload{
		Active:      true,
		MediaTitle:  media.GetTitleSafe(),
		Episode:     episodeNumber,
		Resolutions: resolutions,
		MinSeeders:  minSeeders,
		Step:        "searching",
		StepDetail:  "Starting auto-select search...",
	}

	ctx = context.WithValue(ctx, statusKey, &status)
	s.updateStatus(status)

	defer func() {
		status.Active = false
		s.updateStatus(status)
	}()

	// 1. Search
	s.log("Searching for torrents")
	torrents, err := s.search(ctx, media, episodeNumber, profile)
	if err != nil {
		s.log(fmt.Sprintf("Search failed: %v", err))
		return nil, err
	}

	// 2. Filter & sort
	s.log("Filtering and sorting candidates")
	s.updateStep(ctx, "ranking", "Filtering and sorting candidates...")
	torrents = s.filterAndSort(ctx, torrents, profile, postSearchSort)

	// 3. Select file (iterate top 3)
	s.log("Selecting best file from top candidates")
	s.updateStep(ctx, "analyzing", "Selecting best file from top candidates...")
	res, err = s.selectFile(ctx, media, episodeNumber, torrents, mode, torrentClient, debridClient)
	if err != nil {
		return nil, err
	}

	if res != nil {
		fileName := ""
		if res.File != nil {
			fileName = res.File.DisplayPath()
		} else if res.DebridTorrent != nil {
			fileName = res.DebridFileID
		} else if res.AnalysisFile != nil {
			fileName = res.AnalysisFile.GetPath()
		}
		s.completeStatus(ctx, fileName)
	}

	return res, nil
}

func (s *AutoSelect) log(msg string) {
	if s.onEvent != nil {
		s.onEvent(msg)
	}
}
