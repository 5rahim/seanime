package playbackmanager

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/entities"
	"sync"
)

type (
	playlistHub struct {
		requestNewFileCh chan string
		endOfPlaylistCh  chan struct{}

		wsEventManager  events.IWSEventManager
		logger          *zerolog.Logger
		currentPlaylist *entities.Playlist  // The current playlist that is being played (can be nil)
		nextLocalFile   *entities.LocalFile // The next episode that will be played (can be nil)
		cancel          context.CancelFunc  // The cancel function for the current playlist
		mu              sync.Mutex          // The mutex

		playingLf             *entities.LocalFile     // The currently playing local file
		playingMediaListEntry *anilist.MediaListEntry // The currently playing media entry
		completedCurrent      bool                    // Whether the current episode has been completed

		currentState *PlaylistState // This is sent to the client to show the current playlist state
	}

	PlaylistState struct {
		Current   *PlaylistStateItem `json:"current"`
		Next      *PlaylistStateItem `json:"next"`
		Remaining int                `json:"remaining"`
	}

	PlaylistStateItem struct {
		Name       string `json:"name"`
		MediaImage string `json:"mediaImage"`
	}
)

func newPlaylistHub(logger *zerolog.Logger, wsEventManager events.IWSEventManager) *playlistHub {
	return &playlistHub{
		logger:           logger,
		wsEventManager:   wsEventManager,
		requestNewFileCh: make(chan string, 1),
		endOfPlaylistCh:  make(chan struct{}, 1),
	}
}

func (h *playlistHub) loadPlaylist(playlist *entities.Playlist) {
	if playlist == nil {
		h.logger.Error().Msg("playlist hub: Playlist is nil")
		return
	}
	h.reset()
	h.currentPlaylist = playlist
	h.logger.Debug().Str("name", playlist.Name).Msg("playlist hub: Playlist loaded")
	return
}

func (h *playlistHub) reset() {
	if h.cancel != nil {
		h.cancel()
	}
	h.currentPlaylist = nil
	h.playingLf = nil
	h.playingMediaListEntry = nil
	h.currentState = nil
	h.wsEventManager.SendEvent(events.PlaybackManagerPlaylistState, h.currentState)
	return
}

func (h *playlistHub) check(currListEntry *anilist.MediaListEntry, currLf *entities.LocalFile, ps PlaybackState) bool {
	if h.currentPlaylist == nil || currLf == nil || currListEntry == nil {
		h.currentPlaylist = nil
		h.playingLf = nil
		h.playingMediaListEntry = nil
		return false
	}
	return true
}

func (h *playlistHub) findNextFile() (*entities.LocalFile, bool) {
	if h.currentPlaylist == nil || h.playingLf == nil {
		return nil, false
	}

	for i, lf := range h.currentPlaylist.LocalFiles {
		if lf.GetNormalizedPath() == h.playingLf.GetNormalizedPath() {
			if i+1 < len(h.currentPlaylist.LocalFiles) {
				return h.currentPlaylist.LocalFiles[i+1], true
			}
			break
		}
	}

	return nil, false
}

func (h *playlistHub) playNextFile() (*entities.LocalFile, bool) {
	if h.currentPlaylist == nil || h.playingLf == nil || h.nextLocalFile == nil {
		return nil, false
	}

	h.logger.Debug().Str("path", h.nextLocalFile.Path).Str("cmd", "playNextFile").Msg("playlist hub: Requesting next file")
	h.requestNewFileCh <- h.nextLocalFile.Path
	h.completedCurrent = false

	return nil, false
}

func (h *playlistHub) onVideoStart(currListEntry *anilist.MediaListEntry, currLf *entities.LocalFile, anilistCollection *anilist.AnimeCollection, ps PlaybackState) {
	if !h.check(currListEntry, currLf, ps) {
		return
	}

	h.completedCurrent = false
	h.playingLf = currLf
	h.playingMediaListEntry = currListEntry

	h.nextLocalFile, _ = h.findNextFile()

	// Refresh current playlist state
	playlistState := &PlaylistState{}
	playlistState.Current = &PlaylistStateItem{
		Name:       fmt.Sprintf("%s - Episode %d", currListEntry.GetMedia().GetPreferredTitle(), currLf.GetEpisodeNumber()),
		MediaImage: currListEntry.GetMedia().GetCoverImageSafe(),
	}
	if h.nextLocalFile != nil {
		lfe, found := anilistCollection.GetListEntryFromMediaId(h.nextLocalFile.MediaId)
		if found {
			playlistState.Next = &PlaylistStateItem{
				Name:       fmt.Sprintf("%s - Episode %d", lfe.GetMedia().GetPreferredTitle(), h.nextLocalFile.GetEpisodeNumber()),
				MediaImage: lfe.GetMedia().GetCoverImageSafe(),
			}
		}
	}
	remaining := 0
	for i, lf := range h.currentPlaylist.LocalFiles {
		if lf.GetNormalizedPath() == currLf.GetNormalizedPath() {
			remaining = len(h.currentPlaylist.LocalFiles) - 1 - i
			break
		}
	}
	playlistState.Remaining = remaining
	h.currentState = playlistState

	h.logger.Debug().Str("path", currLf.Path).Msgf("playlist hub: Video started")

	return
}

func (h *playlistHub) onVideoCompleted(currListEntry *anilist.MediaListEntry, currLf *entities.LocalFile, ps PlaybackState) {
	if !h.check(currListEntry, currLf, ps) {
		return
	}

	h.completedCurrent = true

	return
}

func (h *playlistHub) onPlaybackStatus(currListEntry *anilist.MediaListEntry, currLf *entities.LocalFile, ps PlaybackState) {
	if !h.check(currListEntry, currLf, ps) {
		return
	}

	h.wsEventManager.SendEvent(events.PlaybackManagerPlaylistState, h.currentState)

	return
}

func (h *playlistHub) onTrackingStopped() {
	if h.currentPlaylist == nil || h.playingLf == nil { // Return if no playlist
		return
	}

	// When tracking has stopped, request next file
	//if h.nextLocalFile != nil {
	//	h.logger.Debug().Str("path", h.nextLocalFile.Path).Msg("playlist hub: Requesting next file")
	//	h.requestNewFileCh <- h.nextLocalFile.Path
	//} else {
	//	h.logger.Debug().Msg("playlist hub: End of playlist")
	//	h.endOfPlaylistCh <- struct{}{}
	//}

	if !h.completedCurrent {
		h.reset()
	}

	return
}

func (h *playlistHub) onTrackingError() {
	if h.currentPlaylist == nil { // Return if no playlist
		return
	}

	// When tracking has stopped, request next file
	if h.completedCurrent {
		if h.nextLocalFile != nil {
			h.logger.Debug().Str("path", h.nextLocalFile.Path).Msg("playlist hub: Requesting next file")
			h.requestNewFileCh <- h.nextLocalFile.Path
			h.completedCurrent = false
		} else {
			h.logger.Debug().Msg("playlist hub: End of playlist")
			h.endOfPlaylistCh <- struct{}{}
			h.completedCurrent = false
		}
	}

	return
}
