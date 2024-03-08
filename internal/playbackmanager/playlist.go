package playbackmanager

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/entities"
	"sync"
)

type (
	playlistHub struct {
		logger          *zerolog.Logger
		currentPlaylist *entities.Playlist  // The current playlist that is being played (can be nil)
		nextLocalFile   *entities.LocalFile // The next episode that will be played (can be nil)
		cancel          context.CancelFunc  // The cancel function for the current playlist
		mu              sync.Mutex          // The mutex
	}
)

func newPlaylistHub(logger *zerolog.Logger) *playlistHub {
	return &playlistHub{
		logger: logger,
	}
}

func (h *playlistHub) loadPlaylist(playlist *entities.Playlist) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentPlaylist = playlist
	h.logger.Debug().Msg("playlist hub: Playlist loaded")

}

func (h *playlistHub) onVideoStart() {
	h.mu.Lock()
	defer h.mu.Unlock()

}

func (h *playlistHub) onVideoCompleted() {
	h.mu.Lock()
	defer h.mu.Unlock()

}

func (h *playlistHub) onTrackingStopped() {
	h.mu.Lock()
	defer h.mu.Unlock()

}

func (h *playlistHub) onTrackingError() {
	h.mu.Lock()
	defer h.mu.Unlock()

}
