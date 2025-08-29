package playlist

import (
	"context"
	"encoding/json"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/directstream"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/nativeplayer"
	"seanime/internal/platforms/platform"
	"seanime/internal/torrentstream"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type ClientEvent struct {
	Type    PlaylistClientEventType `json:"type"`
	Payload interface{}             `json:"payload"`
}

type PlaylistClientEventType string

const (
	PlaylistClientEventStart PlaylistClientEventType = "start-playlist"
)

type PlaylistStateItem struct {
	Name       string `json:"name"`
	MediaImage string `json:"mediaImage"`
}

type (
	Manager struct {
		// Playlist being played currently
		currentPlaylist mo.Option[*anime.Playlist]
		currentEpisode  mo.Option[*anime.PlaylistEpisode]
		db              *db.Database
		platform        platform.Platform
		wsEventManager  events.WSEventManagerInterface

		directstreamManager     *directstream.Manager
		playbackManager         *playbackmanager.PlaybackManager
		nativePlayer            *nativeplayer.NativePlayer
		torrentstreamRepository *torrentstream.Repository
		debridClientRepository  *debrid_client.Repository

		mu     sync.Mutex
		logger *zerolog.Logger

		isStartingPlaylist   atomic.Bool
		isLoadingNextEpisode atomic.Bool

		ctx    context.Context
		cancel context.CancelFunc
	}

	NewManagerOptions struct {
		DirectStreamManager     *directstream.Manager
		PlaybackManager         *playbackmanager.PlaybackManager
		TorrentstreamRepository *torrentstream.Repository
		DebridClientRepository  *debrid_client.Repository
		NativePlayer            *nativeplayer.NativePlayer
		Logger                  *zerolog.Logger
		Platform                platform.Platform
		WSEventManager          events.WSEventManagerInterface
		Database                *db.Database
	}
)

func NewManager(opts *NewManagerOptions) *Manager {
	ret := &Manager{
		directstreamManager:     opts.DirectStreamManager,
		playbackManager:         opts.PlaybackManager,
		logger:                  opts.Logger,
		torrentstreamRepository: opts.TorrentstreamRepository,
		debridClientRepository:  opts.DebridClientRepository,
		nativePlayer:            opts.NativePlayer,
		platform:                opts.Platform,
		db:                      opts.Database,
		wsEventManager:          opts.WSEventManager,
	}

	ret.listenToEvents()

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	startPlaylistPayload struct {
		DbId uint `json:"dbId"`
	}
)

func (m *Manager) listenToEvents() {
	go func() {
		subscriber := m.wsEventManager.SubscribeToClientPlaylistEvents("playlist-manager")

		for clientEvent := range subscriber.Channel {
			marshaledPayload, err := json.Marshal(clientEvent.Payload)
			if err != nil {
				continue
			}
			event := ClientEvent{}
			err = json.Unmarshal(marshaledPayload, &event)
			if err != nil {
				continue
			}
			switch event.Type {
			case PlaylistClientEventStart:
				// User is starting a new playlist
				m.logger.Debug().Msg("playlist: New playlist requested")

				if m.isStartingPlaylist.Load() {
					continue
				}
				m.isStartingPlaylist.Store(true)

				// cancel any existing playback
				if m.cancel != nil {
					m.cancel()
				}
				payload := startPlaylistPayload{}
				if err := event.UnmarshalAs(&payload); err == nil {
					// Get the playlist
					playlist, err := db_bridge.GetPlaylist(m.db, payload.DbId)
					if err != nil {
						m.logger.Error().Err(err).Msg("playlist: failed to get playlist")
						m.wsEventManager.SendEvent(events.ErrorToast, "Failed to retrieve playlist info")
						m.isStartingPlaylist.Store(false)
						continue
					}
					// Start playlist
					go m.startPlaylist(playlist)
				}
				m.isStartingPlaylist.Store(false)
			}
		}
	}()
}

func (m *Manager) startPlaylist(playlist *anime.Playlist) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Debug().Uint("dbId", playlist.DbId).Msg("playlist: Starting playlist")

	// store the playlist
	m.currentPlaylist = mo.Some(playlist)

	// create a new context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	playbackManagerSubscriber := m.playbackManager.SubscribeToPlaybackStatus("playlist-manager")
	nativePlayerSubscriber := m.nativePlayer.Subscribe("playlist-manager")

	episodeCompleted := atomic.Bool{}

	// continue in goroutine
	go func() {
		select {
		case <-m.ctx.Done():
			m.logger.Trace().Uint("dbId", playlist.DbId).Msg("playlist: Current playlist context done")
			m.resetPlaylist()
			m.playbackManager.UnsubscribeFromPlaybackStatus("playlist-manager")
			m.nativePlayer.Unsubscribe("playlist-manager")
		case event := <-playbackManagerSubscriber.EventCh:
			switch e := event.(type) {
			case playbackmanager.PlaybackStatusChangedEvent:
				// check if video is done
				if e.Status.CompletionPercentage < 99.9 {
					if e.Status.CompletionPercentage >= 80.0 {
						episodeCompleted.Store(true)
					}
					break
				}
				m.markCurrentAsCompleted()
				m.playNextEpisode()
			case playbackmanager.PlaybackErrorEvent:
				if episodeCompleted.Load() {
					m.markCurrentAsCompleted()
					m.playNextEpisode()
				}
			case playbackmanager.VideoStartedEvent, playbackmanager.StreamStartedEvent:
				episodeCompleted.Store(false)
			}
		case event := <-nativePlayerSubscriber.Events():
			switch event.(type) {
			case *nativeplayer.VideoLoadedMetadataEvent:
				episodeCompleted.Store(false)
			case *nativeplayer.VideoCompletedEvent:
				m.markCurrentAsCompleted()
				episodeCompleted.Store(true)
			case *nativeplayer.VideoEndedEvent:
				m.markCurrentAsCompleted()
				m.playNextEpisode()
			}
		}
	}()

	// Continue playlist
	go m.playNextEpisode()

}

func (m *Manager) playNextEpisode() {
	if m.isLoadingNextEpisode.Load() {
		return
	}
	m.isLoadingNextEpisode.Store(true)
	defer m.isLoadingNextEpisode.Store(false)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().Msg("playlist: Playing next episode")

	playlist, ok := m.currentPlaylist.Get()
	if !ok {
		m.logger.Error().Msg("playlist: Cannot play next episode, no playlist is currently playing")
		return
	}

	// find episode
	var episode *anime.PlaylistEpisode
	for _, playlistEp := range playlist.Episodes {
		if playlistEp.IsCompleted {
			continue
		}
		episode = playlistEp
		break
	}

	if episode == nil {
		m.logger.Error().Msg("playlist: Cannot play next episode, no episodes in playlist")
		return
	}

	// store pointer to episode
	m.currentEpisode = mo.Some(episode)
}

func (m *Manager) hasNextEpisode() bool {
	playlist, ok := m.currentPlaylist.Get()
	if !ok {
		return false
	}

	var found bool
	for _, playlistEp := range playlist.Episodes {
		if playlistEp.IsCompleted {
			continue
		}
		found = true
		break
	}

	return found
}

func (m *Manager) markCurrentAsCompleted() {
	m.logger.Trace().Msg("playlist: Marking current episode as completed")

	playlist, ok := m.currentPlaylist.Get()
	if !ok {
		return
	}

	currentEpisode, ok := m.currentEpisode.Get()
	if !ok {
		return
	}

	if currentEpisode.IsCompleted {
		return
	}

	currentEpisode.IsCompleted = true

	go func() {
		// update the playlist in db
		err := db_bridge.UpdatePlaylist(m.db, playlist)
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to update playlist")
		}
		// update the progress
		err = m.platform.UpdateEntryProgress(context.Background(), currentEpisode.Episode.BaseAnime.GetID(), currentEpisode.Episode.ProgressNumber, currentEpisode.Episode.BaseAnime.Episodes)
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to update progress")
		}
	}()

	return
}

func (m *Manager) isSameEpisode(a *anime.Episode, b *anime.Episode) bool {
	if a == nil || b == nil {
		return false
	}

	// If one file is a local file, use progress number for comparison
	if a.LocalFile != nil || b.LocalFile != nil {
		return a.BaseAnime.ID == b.BaseAnime.ID && a.ProgressNumber == b.ProgressNumber
	}

	// Otherwise, use AniDB episode number for comparison
	return a.BaseAnime.ID == b.BaseAnime.ID && a.AniDBEpisode == b.AniDBEpisode
}

func (m *Manager) resetPlaylist() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Trace().Msg("playlist: Stopping current playlist")
	m.currentPlaylist = mo.None[*anime.Playlist]()
	m.cancel = nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// PlayNextEpisode plays the next episode in the playlist
// isEpisodeCompleted is true if the current episode is completed (used for manual tracking)
func (m *Manager) PlayNextEpisode(isEpisodeCompleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	marshaled, _ := json.Marshal(e.Payload)
	return json.Unmarshal(marshaled, dest)
}
