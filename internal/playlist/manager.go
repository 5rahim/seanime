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

type PlaylistClientEventType string

type ClientEvent struct {
	Type    PlaylistClientEventType `json:"type"`
	Payload interface{}             `json:"payload"`
}

const (
	ClientEventCurrentPlaylist PlaylistClientEventType = "current-playlist"
	ClientEventStart           PlaylistClientEventType = "start-playlist"
	ClientEventStop            PlaylistClientEventType = "stop-playlist"
	ClientEventPlayEpisode     PlaylistClientEventType = "play-episode"
	ClientEventReopenEpisode   PlaylistClientEventType = "reopen-episode"
)

type ClientPlaybackMethod string

const (
	ClientPlaybackMethodNone               ClientPlaybackMethod = ""
	ClientPlaybackMethodDefault            ClientPlaybackMethod = "default" // desktop media player
	ClientPlaybackMethodTranscode          ClientPlaybackMethod = "transcode"
	ClientPlaybackMethodExternalPlayerLink ClientPlaybackMethod = "externalPlayerLink"
	ClientPlaybackMethodNativePlayer       ClientPlaybackMethod = "nativePlayer"
)

func (m ClientPlaybackMethod) String() string {
	return string(m)
}

//--------------------------------------------------------------------------------------------------------------------------------------------------//

type PlaylistServerEventType string

type ServerEvent struct {
	Type    PlaylistServerEventType `json:"type"`
	Payload interface{}             `json:"payload"`
}

const (
	ServerEventCurrentPlaylist PlaylistServerEventType = "current-playlist"
	ServerEventPlayEpisode     PlaylistServerEventType = "play-episode"
)

//--------------------------------------------------------------------------------------------------------------------------------------------------//

type playlistData struct {
	playlist *anime.Playlist
	options  *startPlaylistPayload
}

type (
	Manager struct {
		// Playlist being played currently
		currentPlaylistData   mo.Option[*playlistData]
		currentEpisode        mo.Option[*anime.PlaylistEpisode]
		currentPlaybackMethod ClientPlaybackMethod
		db                    *db.Database
		platform              platform.Platform
		wsEventManager        events.WSEventManagerInterface

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
		DbId                    uint                 `json:"dbId"`
		LocalFilePlaybackMethod ClientPlaybackMethod `json:"localFilePlaybackMethod"`
		StreamPlaybackMethod    ClientPlaybackMethod `json:"streamPlaybackMethod"`
	}

	episodeRequestedPayload struct {
		Which              string `json:"which"`              // "next", "previous", or index
		IsCurrentCompleted bool   `json:"isCurrentCompleted"` // Whether to mark the current episode as completed
	}
)

func (m *Manager) sendCurrentPlaylistToClient() {
	playlistEpisode, _ := m.currentEpisode.Get()

	data, ok := m.currentPlaylistData.Get()
	if !ok {
		m.wsEventManager.SendEvent(string(events.PlaylistEvent), ServerEvent{
			Type: ServerEventCurrentPlaylist,
			Payload: struct {
				PlaylistEpisode *anime.PlaylistEpisode `json:"playlistEpisode"`
				Playlist        *anime.Playlist        `json:"playlist"`
			}{
				PlaylistEpisode: playlistEpisode,
				Playlist:        nil,
			},
		})
		return
	}
	m.wsEventManager.SendEvent(string(events.PlaylistEvent), ServerEvent{
		Type: ServerEventCurrentPlaylist,
		Payload: struct {
			PlaylistEpisode *anime.PlaylistEpisode `json:"playlistEpisode"`
			Playlist        *anime.Playlist        `json:"playlist"`
		}{
			PlaylistEpisode: playlistEpisode,
			Playlist:        data.playlist,
		},
	})
}

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
			case ClientEventCurrentPlaylist:
				// UI requested current playlist
				m.sendCurrentPlaylistToClient()
			case ClientEventStart:
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
					go m.startPlaylist(playlist, &payload)
				}
				m.isStartingPlaylist.Store(false)
			case ClientEventStop:
				m.logger.Debug().Msg("playlist: Stop requested")
				m.StopPlaylist("Playlist stopped")
			case ClientEventPlayEpisode:
				payload := episodeRequestedPayload{}
				if err := event.UnmarshalAs(&payload); err == nil {
					m.PlayEpisode(payload.Which, payload.IsCurrentCompleted)
				}
			case ClientEventReopenEpisode:
				m.ReopenEpisode()
			default:
				m.logger.Error().Msgf("playlist: Unknown event type: %s", event.Type)
			}
		}
	}()
}

func (m *Manager) startPlaylist(playlist *anime.Playlist, options *startPlaylistPayload) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}

	m.logger.Debug().Uint("dbId", playlist.DbId).Msg("playlist: Starting playlist")

	// store the playlist
	m.currentPlaylistData = mo.Some(&playlistData{
		playlist: playlist,
		options:  options,
	})

	m.sendCurrentPlaylistToClient()

	// create a new context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	playbackManagerSubscriber := m.playbackManager.SubscribeToPlaybackStatus("playlist-manager")
	nativePlayerSubscriber := m.nativePlayer.Subscribe("playlist-manager")

	episodeCompleted := atomic.Bool{}
	isTransitioning := atomic.Bool{}

	// continue in goroutine
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				m.logger.Trace().Uint("dbId", playlist.DbId).Msg("playlist: Current playlist context done")
				m.resetPlaylist()
				m.playbackManager.UnsubscribeFromPlaybackStatus("playlist-manager")
				m.nativePlayer.Unsubscribe("playlist-manager")
				return
			case event := <-playbackManagerSubscriber.EventCh:
				switch e := event.(type) {
				case playbackmanager.PlaybackStatusChangedEvent:
					// check if video is done
					if e.Status.CompletionPercentage < 99 {
						if e.Status.CompletionPercentage >= 80.0 {
							episodeCompleted.Store(true)
						}
						continue
					}
					m.markCurrentAsCompleted()
					m.playNextEpisode()

				case playbackmanager.VideoCompletedEvent, playbackmanager.StreamCompletedEvent:
					episodeCompleted.Store(true)

				case playbackmanager.PlaybackErrorEvent:
					if episodeCompleted.Load() {
						m.markCurrentAsCompleted()
						m.playNextEpisode()
						// player is closed before starting the next episode
						isTransitioning.Store(true)
						continue
					}
					// Otherwise, stop the playlist
					m.StopPlaylist("Playlist stopped")

				case playbackmanager.VideoStartedEvent, playbackmanager.StreamStartedEvent:
					episodeCompleted.Store(false)

					// Check if the episode has changed to the next one without completion/error events
					if !isTransitioning.Load() {
						currentEpisode, ok := m.currentEpisode.Get()
						if ok {
							data, ok := m.currentPlaylistData.Get()
							if !ok {
								continue
							}
							nextEpisode, found := data.playlist.NextEpisode(currentEpisode)
							if found {
								m.currentEpisode = mo.Some(nextEpisode)
							}
						}
					}
					isTransitioning.Store(false)
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

				case *nativeplayer.VideoTerminatedEvent:
					m.StopPlaylist("Playlist stopped")
				}
			}
		}
	}()

	// Continue playlist
	go m.playNextEpisode()

}

type (
	playEpisodePayload struct {
		PlaylistEpisode *anime.PlaylistEpisode `json:"playlistEpisode"`
	}
)

func (m *Manager) playNextEpisode() {
	if m.isLoadingNextEpisode.Load() {
		return
	}
	m.isLoadingNextEpisode.Store(true)
	defer m.isLoadingNextEpisode.Store(false)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().Msg("playlist: Playing next episode")

	data, ok := m.currentPlaylistData.Get()
	if !ok {
		m.logger.Error().Msg("playlist: Cannot play next episode, no playlist is currently playing")
		return
	}

	// find episode
	var episode *anime.PlaylistEpisode
	for _, playlistEp := range data.playlist.Episodes {
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

	m.sendCurrentPlaylistToClient()

	m.playEpisode(episode)

	m.prepareNextEpisode()
}

func (m *Manager) hasNextEpisode() bool {
	data, ok := m.currentPlaylistData.Get()
	if !ok {
		return false
	}

	var found bool
	for _, playlistEp := range data.playlist.Episodes {
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

	data, ok := m.currentPlaylistData.Get()
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

	go func(currentEpisode anime.PlaylistEpisode) {
		// update the playlist in db
		err := db_bridge.UpdatePlaylist(m.db, data.playlist)
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to update playlist")
		}
		// update the progress
		err = m.platform.UpdateEntryProgress(context.Background(), currentEpisode.Episode.BaseAnime.GetID(), currentEpisode.Episode.ProgressNumber, currentEpisode.Episode.BaseAnime.Episodes)
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to update progress")
		}
	}(*currentEpisode)

	m.sendCurrentPlaylistToClient()

	return
}

func (m *Manager) resetPlaylist() {
	m.currentPlaylistData = mo.None[*playlistData]()
	m.cancel = nil
	m.sendCurrentPlaylistToClient()
}

func (m *Manager) playEpisode(episode *anime.PlaylistEpisode) {
	data, ok := m.currentPlaylistData.Get()
	if !ok {
		return
	}

	// play the file
	// - if external player link, do nothing
	// - if nakama stream, launch it from client
	// - if local file & desktop player, launch it from server
	// - if torrent/debrid stream, launch it from client

	isLf := isLocalFile(episode)
	isNakama := episode.IsNakama
	isTorrentOrDebridStream := !isLf && !isNakama

	// it's a local file and user uses an external player link, do nothing
	if (isLf && data.options.LocalFilePlaybackMethod == ClientPlaybackMethodExternalPlayerLink) ||
		(isNakama && data.options.LocalFilePlaybackMethod == ClientPlaybackMethodExternalPlayerLink) ||
		(isTorrentOrDebridStream && data.options.StreamPlaybackMethod == ClientPlaybackMethodExternalPlayerLink) {
		m.logger.Trace().Msg("playlist: External player link, skipping")

		m.currentPlaybackMethod = ClientPlaybackMethodExternalPlayerLink

		return
	}

	// local file and desktop media player, play it from server
	if isLf && data.options.LocalFilePlaybackMethod == ClientPlaybackMethodDefault {
		err := m.playbackManager.StartPlayingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
			Payload:   episode.Episode.LocalFile.Path,
			UserAgent: "",
			ClientId:  "",
		})
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to start playing local file")
			m.StopPlaylist("Failed to start playing local file")
		}

		m.currentPlaybackMethod = ClientPlaybackMethodDefault

		return
	}

	m.wsEventManager.SendEvent(string(events.PlaylistEvent), ServerEvent{
		Type: ServerEventPlayEpisode,
		Payload: playEpisodePayload{
			PlaylistEpisode: episode,
		},
	})
}

func (m *Manager) prepareNextEpisode() {
	m.logger.Trace().Msg("playlist: Preparing next episode")

	data, ok := m.currentPlaylistData.Get()
	if !ok {
		return
	}

	currentEpisode, ok := m.currentEpisode.Get()
	if !ok {
		return
	}

	nextEpisode, found := data.playlist.NextEpisode(currentEpisode)
	if !found {
		return
	}

	// Append local file to media player
	if isLocalFile(nextEpisode) {
		if data.options.LocalFilePlaybackMethod != ClientPlaybackMethodDefault {
			return
		}

		err := m.playbackManager.AppendToMediaPlayer(&playbackmanager.AppendToMediaPlayerOptions{
			Payload: nextEpisode.Episode.LocalFile.Path,
		})
		if err != nil {
			m.logger.Error().Err(err).Msg("playlist: Failed to append to media player")
			m.StopPlaylist("Failed to append to media player")
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Manager) StopPlaylist(reason string) {
	m.logger.Trace().Msg("playlist: Stopping current playlist")
	if m.cancel != nil {
		m.cancel()
	}
	m.isStartingPlaylist.Store(false)
	m.resetPlaylist()
	m.wsEventManager.SendEvent(events.InfoToast, reason)
}

// PlayEpisode plays the next episode in the playlist
// isEpisodeCompleted is true if the current episode is completed (used for manual tracking)
func (m *Manager) PlayEpisode(which string, isCurrentCompleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isCurrentCompleted {
		m.markCurrentAsCompleted()
	}

}

func (m *Manager) ReopenEpisode() {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.currentPlaylistData.Get()
	if !ok {
		return
	}

	currentEpisode, ok := m.currentEpisode.Get()
	if !ok {
		return
	}

	m.playEpisode(currentEpisode)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	marshaled, _ := json.Marshal(e.Payload)
	return json.Unmarshal(marshaled, dest)
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
