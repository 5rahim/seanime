package continuity

import (
	"github.com/rs/zerolog"
	"seanime/internal/util/filecache"
	"sync"
	"time"
)

const (
	OnlinestreamKind   Kind = "onlinestream"
	MediastreamKind    Kind = "mediastream"
	ExternalPlayerKind Kind = "external_player"
)

type (
	// Manager is used to manage the user's viewing history across different media types.
	Manager struct {
		fileCacher *filecache.Cacher
		// Permanent bucket
		watchHistoryFileCacheBucket *filecache.Bucket
		logger                      *zerolog.Logger

		settings *Settings
		mu       sync.RWMutex
	}

	Settings struct {
		WatchContinuityEnabled bool
	}

	Kind string
)

type (
	NewManagerOptions struct {
		FileCacher *filecache.Cacher
		Logger     *zerolog.Logger
	}
)

// NewManager creates a new Manager, it should be initialized once.
func NewManager(opts *NewManagerOptions) *Manager {
	watchHistoryFileCacheBucket := filecache.NewBucket(WatchHistoryBucketName, time.Hour*24*99999)

	ret := &Manager{
		fileCacher:                  opts.FileCacher,
		logger:                      opts.Logger,
		watchHistoryFileCacheBucket: &watchHistoryFileCacheBucket,
	}

	ret.logger.Info().Msg("continuity: Initialized manager")

	return ret
}

// SetSettings should be called after initializing the Manager.
func (m *Manager) SetSettings(settings *Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = settings
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
