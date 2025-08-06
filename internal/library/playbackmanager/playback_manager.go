package playbackmanager

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

const (
	LocalFilePlayback      PlaybackType = "localfile"
	StreamPlayback         PlaybackType = "stream"
	ManualTrackingPlayback PlaybackType = "manual"
)

var playbackStatePool = sync.Pool{
	New: func() interface{} {
		return &PlaybackState{}
	},
}

type (
	PlaybackType string

	// PlaybackManager manages video playback (local and stream) and progress tracking for desktop media players.
	// It receives and dispatch appropriate events for:
	//  - Syncing progress with AniList, etc.
	//  - Sending notifications to the client
	PlaybackManager struct {
		Logger                *zerolog.Logger
		Database              *db.Database
		MediaPlayerRepository *mediaplayer.Repository // MediaPlayerRepository is used to control the media player
		continuityManager     *continuity.Manager

		settings *Settings

		discordPresence            *discordrpc_presence.Presence     // DiscordPresence is used to update the user's Discord presence
		mediaPlayerRepoSubscriber  *mediaplayer.RepositorySubscriber // Used to listen for media player events
		wsEventManager             events.WSEventManagerInterface
		platform                   platform.Platform
		metadataProvider           metadata.Provider
		refreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		mu                         sync.Mutex
		eventMu                    sync.RWMutex
		cancel                     context.CancelFunc

		// historyMap stores a PlaybackState whose state is "completed"
		// Since PlaybackState is sent to client continuously, once a PlaybackState is stored in historyMap, only IT will be sent to the client.
		// This is so when the user seeks back to a video, the client can show the last known "completed" state of the video
		historyMap                 map[string]PlaybackState
		currentPlaybackType        PlaybackType
		currentMediaPlaybackStatus *mediaplayer.PlaybackStatus // The current video playback status (can be nil)

		autoPlayMu           sync.Mutex
		nextEpisodeLocalFile mo.Option[*anime.LocalFile] // The next episode's local file (for local file playback)

		// currentMediaListEntry for Local file playback & stream playback
		// For Local file playback, it MUST be set
		// For Stream playback, it is optional
		// See [progress_tracking.go] for how it is handled
		currentMediaListEntry mo.Option[*anilist.AnimeListEntry] // List Entry for the current video playback

		// \/ Local file playback
		currentLocalFile             mo.Option[*anime.LocalFile]             // Local file for the current video playback
		currentLocalFileWrapperEntry mo.Option[*anime.LocalFileWrapperEntry] // This contains the current media entry local file data

		// \/ Stream playback
		// The current episode being streamed, set in [StartStreamingUsingMediaPlayer] by finding the episode in currentStreamEpisodeCollection
		currentStreamEpisode mo.Option[*anime.Episode]
		// The current media being streamed, set in [StartStreamingUsingMediaPlayer]
		currentStreamMedia        mo.Option[*anilist.BaseAnime]
		currentStreamAniDbEpisode mo.Option[string]

		// \/ Manual progress tracking (non-integrated external player)
		manualTrackingCtx           context.Context
		manualTrackingCtxCancel     context.CancelFunc
		manualTrackingPlaybackState PlaybackState
		currentManualTrackingState  mo.Option[*ManualTrackingState]
		manualTrackingWg            sync.WaitGroup

		// \/ Playlist
		playlistHub *playlistHub // The playlist hub

		isOffline       *bool
		animeCollection mo.Option[*anilist.AnimeCollection]

		playbackStatusSubscribers *result.Map[string, *PlaybackStatusSubscriber]
	}

	// PlaybackStatusSubscriber provides a single event channel for all playback events
	PlaybackStatusSubscriber struct {
		EventCh  chan PlaybackEvent
		canceled atomic.Bool
	}

	// PlaybackEvent is the base interface for all playback events
	PlaybackEvent interface {
		Type() string
	}

	PlaybackStartingEvent struct {
		Filepath      string
		PlaybackType  PlaybackType
		Media         *anilist.BaseAnime
		AniDbEpisode  string
		EpisodeNumber int
		WindowTitle   string
	}

	// Local file playback events

	PlaybackStatusChangedEvent struct {
		Status mediaplayer.PlaybackStatus
		State  PlaybackState
	}

	VideoStartedEvent struct {
		Filename string
		Filepath string
	}

	VideoStoppedEvent struct {
		Reason string
	}

	VideoCompletedEvent struct {
		Filename string
	}

	// Stream playback events
	StreamStateChangedEvent struct {
		State PlaybackState
	}

	StreamStatusChangedEvent struct {
		Status mediaplayer.PlaybackStatus
	}

	StreamStartedEvent struct {
		Filename string
		Filepath string
	}

	StreamStoppedEvent struct {
		Reason string
	}

	StreamCompletedEvent struct {
		Filename string
	}

	PlaybackStateType string

	// PlaybackState is used to keep track of the user's current video playback
	// It is sent to the client each time the video playback state is picked up -- this is used to update the client's UI
	PlaybackState struct {
		EpisodeNumber        int     `json:"episodeNumber"`        // The episode number
		AniDbEpisode         string  `json:"aniDbEpisode"`         // The AniDB episode number
		MediaTitle           string  `json:"mediaTitle"`           // The title of the media
		MediaCoverImage      string  `json:"mediaCoverImage"`      // The cover image of the media
		MediaTotalEpisodes   int     `json:"mediaTotalEpisodes"`   // The total number of episodes
		Filename             string  `json:"filename"`             // The filename
		CompletionPercentage float64 `json:"completionPercentage"` // The completion percentage
		CanPlayNext          bool    `json:"canPlayNext"`          // Whether the next episode can be played
		ProgressUpdated      bool    `json:"progressUpdated"`      // Whether the progress has been updated
		MediaId              int     `json:"mediaId"`              // The media ID
	}

	NewPlaybackManagerOptions struct {
		WSEventManager             events.WSEventManagerInterface
		Logger                     *zerolog.Logger
		Platform                   platform.Platform
		MetadataProvider           metadata.Provider
		Database                   *db.Database
		RefreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		DiscordPresence            *discordrpc_presence.Presence
		IsOffline                  *bool
		ContinuityManager          *continuity.Manager
	}

	Settings struct {
		AutoPlayNextEpisode bool
	}
)

// Event type implementations
func (e PlaybackStatusChangedEvent) Type() string { return "playback_status_changed" }
func (e VideoStartedEvent) Type() string          { return "video_started" }
func (e VideoStoppedEvent) Type() string          { return "video_stopped" }
func (e VideoCompletedEvent) Type() string        { return "video_completed" }
func (e StreamStateChangedEvent) Type() string    { return "stream_state_changed" }
func (e StreamStatusChangedEvent) Type() string   { return "stream_status_changed" }
func (e StreamStartedEvent) Type() string         { return "stream_started" }
func (e StreamStoppedEvent) Type() string         { return "stream_stopped" }
func (e StreamCompletedEvent) Type() string       { return "stream_completed" }
func (e PlaybackStartingEvent) Type() string      { return "playback_starting" }

func New(opts *NewPlaybackManagerOptions) *PlaybackManager {
	pm := &PlaybackManager{
		Logger:                       opts.Logger,
		Database:                     opts.Database,
		settings:                     &Settings{},
		discordPresence:              opts.DiscordPresence,
		wsEventManager:               opts.WSEventManager,
		platform:                     opts.Platform,
		metadataProvider:             opts.MetadataProvider,
		refreshAnimeCollectionFunc:   opts.RefreshAnimeCollectionFunc,
		mu:                           sync.Mutex{},
		autoPlayMu:                   sync.Mutex{},
		eventMu:                      sync.RWMutex{},
		historyMap:                   make(map[string]PlaybackState),
		isOffline:                    opts.IsOffline,
		nextEpisodeLocalFile:         mo.None[*anime.LocalFile](),
		currentStreamEpisode:         mo.None[*anime.Episode](),
		currentStreamMedia:           mo.None[*anilist.BaseAnime](),
		currentStreamAniDbEpisode:    mo.None[string](),
		animeCollection:              mo.None[*anilist.AnimeCollection](),
		currentManualTrackingState:   mo.None[*ManualTrackingState](),
		currentLocalFile:             mo.None[*anime.LocalFile](),
		currentLocalFileWrapperEntry: mo.None[*anime.LocalFileWrapperEntry](),
		currentMediaListEntry:        mo.None[*anilist.AnimeListEntry](),
		continuityManager:            opts.ContinuityManager,
		playbackStatusSubscribers:    result.NewResultMap[string, *PlaybackStatusSubscriber](),
	}

	pm.playlistHub = newPlaylistHub(pm)

	return pm
}

func (pm *PlaybackManager) SetAnimeCollection(ac *anilist.AnimeCollection) {
	pm.animeCollection = mo.Some(ac)
}

func (pm *PlaybackManager) SetSettings(s *Settings) {
	pm.settings = s
}

// SetMediaPlayerRepository sets the media player repository and starts listening to media player events
// - This method is called when the media player is mounted (due to settings change or when the app starts)
func (pm *PlaybackManager) SetMediaPlayerRepository(mediaPlayerRepository *mediaplayer.Repository) {
	go func() {
		// If a previous context exists, cancel it
		if pm.cancel != nil {
			pm.cancel()
		}

		pm.playlistHub.reset()

		// Create a new context for listening to the MediaPlayer instance's event
		// When this is canceled above, the previous listener goroutine will stop -- this is done to prevent multiple listeners
		var ctx context.Context
		ctx, pm.cancel = context.WithCancel(context.Background())

		pm.mu.Lock()
		// Set the new media player repository instance
		pm.MediaPlayerRepository = mediaPlayerRepository
		// Set up event listeners for the media player instance
		pm.mediaPlayerRepoSubscriber = pm.MediaPlayerRepository.Subscribe("playbackmanager")
		pm.mu.Unlock()

		// Start listening to new media player events
		pm.listenToMediaPlayerEvents(ctx)

		// DEVNOTE: pm.listenToClientPlayerEvents()
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StartPlayingOptions struct {
	Payload   string // url or path
	UserAgent string
	ClientId  string
}

func (pm *PlaybackManager) StartPlayingUsingMediaPlayer(opts *StartPlayingOptions) error {

	event := &LocalFilePlaybackRequestedEvent{
		Path: opts.Payload,
	}
	err := hook.GlobalHookManager.OnLocalFilePlaybackRequested().Trigger(event)
	if err != nil {
		return err
	}
	opts.Payload = event.Path

	if event.DefaultPrevented {
		pm.Logger.Debug().Msg("playback manager: Local file playback prevented by hook")
		return nil
	}

	pm.playlistHub.reset()
	if err := pm.checkOrLoadAnimeCollection(); err != nil {
		return err
	}

	// Cancel manual tracking if active
	if pm.manualTrackingCtxCancel != nil {
		pm.manualTrackingCtxCancel()
	}

	// Send the media file to the media player
	err = pm.MediaPlayerRepository.Play(opts.Payload)
	if err != nil {
		return err
	}

	trackingEvent := &PlaybackBeforeTrackingEvent{
		IsStream: false,
	}
	err = hook.GlobalHookManager.OnPlaybackBeforeTracking().Trigger(trackingEvent)
	if err != nil {
		return err
	}

	if trackingEvent.DefaultPrevented {
		return nil
	}

	// Start tracking
	pm.MediaPlayerRepository.StartTracking()

	return nil
}

// StartUntrackedStreamingUsingMediaPlayer starts a stream using the media player without any tracking.
func (pm *PlaybackManager) StartUntrackedStreamingUsingMediaPlayer(windowTitle string, opts *StartPlayingOptions) (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/StartUntrackedStreamingUsingMediaPlayer", &err)

	event := &StreamPlaybackRequestedEvent{
		WindowTitle:  windowTitle,
		Payload:      opts.Payload,
		Media:        nil,
		AniDbEpisode: "",
	}
	err = hook.GlobalHookManager.OnStreamPlaybackRequested().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		pm.Logger.Debug().Msg("playback manager: Stream playback prevented by hook")
		return nil
	}

	pm.Logger.Trace().Msg("playback manager: Starting the media player")

	pm.mu.Lock()
	defer pm.mu.Unlock()

	episodeNumber := 0

	err = pm.MediaPlayerRepository.Stream(opts.Payload, episodeNumber, 0, windowTitle)
	if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Failed to start streaming")
		return err
	}

	pm.Logger.Trace().Msg("playback manager: Sent stream to media player")

	return nil
}

// StartStreamingUsingMediaPlayer starts streaming a video using the media player.
// This sets PlaybackManager.currentStreamMedia and PlaybackManager.currentStreamEpisode used for progress tracking.
// Note that PlaybackManager.currentStreamEpisodeCollection is not required to start streaming but is needed for progress tracking.
func (pm *PlaybackManager) StartStreamingUsingMediaPlayer(windowTitle string, opts *StartPlayingOptions, media *anilist.BaseAnime, aniDbEpisode string) (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/StartStreamingUsingMediaPlayer", &err)

	event := &StreamPlaybackRequestedEvent{
		WindowTitle:  windowTitle,
		Payload:      opts.Payload,
		Media:        media,
		AniDbEpisode: aniDbEpisode,
	}
	err = hook.GlobalHookManager.OnStreamPlaybackRequested().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		pm.Logger.Debug().Msg("playback manager: Stream playback prevented by hook")
		return nil
	}

	pm.playlistHub.reset()
	if *pm.isOffline {
		return errors.New("cannot stream when offline")
	}

	if media == nil || aniDbEpisode == "" {
		pm.Logger.Error().Msg("playback manager: cannot start streaming, missing options [StartStreamingUsingMediaPlayer]")
		return errors.New("cannot start streaming, not enough data provided")
	}

	pm.Logger.Trace().Msg("playback manager: Starting the media player")

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Cancel manual tracking if active
	if pm.manualTrackingCtxCancel != nil {
		pm.manualTrackingCtxCancel()
	}

	pm.currentStreamMedia = mo.Some(media)
	episodeNumber := 0

	// Find the current episode being stream
	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:    nil,
		Media:            media,
		MetadataProvider: pm.metadataProvider,
		Logger:           pm.Logger,
	})

	pm.currentStreamAniDbEpisode = mo.Some(aniDbEpisode)

	if episode, ok := episodeCollection.FindEpisodeByAniDB(aniDbEpisode); ok {
		episodeNumber = episode.EpisodeNumber
		pm.currentStreamEpisode = mo.Some(episode)
	} else {
		pm.Logger.Warn().Str("episode", aniDbEpisode).Msg("playback manager: Failed to find episode in episode collection")
	}

	err = pm.MediaPlayerRepository.Stream(opts.Payload, episodeNumber, media.ID, windowTitle)
	if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Failed to start streaming")
		return err
	}

	pm.Logger.Trace().Msg("playback manager: Sent stream to media player")

	trackingEvent := &PlaybackBeforeTrackingEvent{
		IsStream: true,
	}
	err = hook.GlobalHookManager.OnPlaybackBeforeTracking().Trigger(trackingEvent)
	if err != nil {
		return err
	}

	if trackingEvent.DefaultPrevented {
		return nil
	}

	pm.MediaPlayerRepository.StartTrackingTorrentStream()

	pm.Logger.Trace().Msg("playback manager: Started tracking torrent stream")

	return nil
}

// PlayNextEpisode plays the next episode of the local media that is being watched
//   - Called when the user clicks on the "Next" button in the client
//   - Should not be called when the user is watching a playlist
//   - Should not be called when no next episode is available
func (pm *PlaybackManager) PlayNextEpisode() (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/PlayNextEpisode", &err)

	switch pm.currentPlaybackType {
	case LocalFilePlayback:
		if pm.currentLocalFile.IsAbsent() || pm.currentMediaListEntry.IsAbsent() || pm.currentLocalFileWrapperEntry.IsAbsent() {
			return errors.New("could not play next episode")
		}

		nextLf, found := pm.currentLocalFileWrapperEntry.MustGet().FindNextEpisode(pm.currentLocalFile.MustGet())
		if !found {
			return errors.New("could not play next episode")
		}

		err = pm.MediaPlayerRepository.Play(nextLf.Path)
		if err != nil {
			return err
		}
		// Start tracking the video
		pm.MediaPlayerRepository.StartTracking()

	case StreamPlayback:
		// TODO: Implement it for torrentstream
		// Check if torrent stream etc...
	}

	return nil
}

// GetNextEpisode gets the next [anime.LocalFile] of the local media that is being watched.
// It will return nil if there is no next episode.
// This is used by the client's "Auto Play" feature.
func (pm *PlaybackManager) GetNextEpisode() (ret *anime.LocalFile) {
	defer util.HandlePanicInModuleThen("library/playbackmanager/GetNextEpisode", func() {
		ret = nil
	})

	switch pm.currentPlaybackType {
	case LocalFilePlayback:
		if lf, found := pm.nextEpisodeLocalFile.Get(); found {
			ret = lf
		}
		return
	}

	return nil
}

// AutoPlayNextEpisode will play the next episode of the local media that is being watched.
// This calls [PlaybackManager.PlayNextEpisode] only once if multiple clients made the request.
func (pm *PlaybackManager) AutoPlayNextEpisode() error {
	pm.autoPlayMu.Lock()
	defer pm.autoPlayMu.Unlock()

	pm.Logger.Trace().Msg("playback manager: Auto play request received")

	if !pm.settings.AutoPlayNextEpisode {
		return nil
	}

	lf := pm.GetNextEpisode()
	// This shouldn't happen because the client should check if there is a next episode before sending the request.
	// However, it will happen if there are multiple clients launching the request.
	if lf == nil {
		pm.Logger.Warn().Msg("playback manager: No next episode to play")
		return nil
	}

	if err := pm.PlayNextEpisode(); err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Failed to auto play next episode")
		return fmt.Errorf("failed to auto play next episode: %w", err)
	}

	// Remove the next episode from the queue
	pm.nextEpisodeLocalFile = mo.None[*anime.LocalFile]()

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause pauses the current media player playback.
func (pm *PlaybackManager) Pause() error {
	return pm.MediaPlayerRepository.Pause()
}

// Resume resumes the current media player playback.
func (pm *PlaybackManager) Resume() error {
	return pm.MediaPlayerRepository.Resume()
}

// Seek seeks to the specified time in the current media.
func (pm *PlaybackManager) Seek(seconds float64) error {
	return pm.MediaPlayerRepository.Seek(seconds)
}

// PullStatus pulls the current media player playback status at the time of the call.
func (pm *PlaybackManager) PullStatus() (*mediaplayer.PlaybackStatus, bool) {
	return pm.MediaPlayerRepository.PullStatus()
}

// Cancel stops the current media player playback and publishes a "normal" event.
func (pm *PlaybackManager) Cancel() error {
	pm.MediaPlayerRepository.Stop()
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Playlist
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// CancelCurrentPlaylist cancels the current playlist.
// This is an action triggered by the client.
func (pm *PlaybackManager) CancelCurrentPlaylist() error {
	go pm.playlistHub.reset()
	return nil
}

// RequestNextPlaylistFile will play the next file in the playlist.
// This is an action triggered by the client.
func (pm *PlaybackManager) RequestNextPlaylistFile() error {
	go pm.playlistHub.playNextFile()
	return nil
}

// StartPlaylist starts a playlist.
// This action is triggered by the client.
func (pm *PlaybackManager) StartPlaylist(playlist *anime.Playlist) (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/StartPlaylist", &err)

	pm.playlistHub.loadPlaylist(playlist)

	_ = pm.checkOrLoadAnimeCollection()

	// Play the first video in the playlist
	firstVidPath := playlist.LocalFiles[0].Path
	err = pm.MediaPlayerRepository.Play(firstVidPath)
	if err != nil {
		return err
	}

	// Start tracking the video
	pm.MediaPlayerRepository.StartTracking()

	// Create a new context for the playlist hub
	var ctx context.Context
	ctx, pm.playlistHub.cancel = context.WithCancel(context.Background())

	// Listen to new play requests
	go func() {
		pm.Logger.Debug().Msg("playback manager: Listening for new file requests")
		for {
			select {
			// When the playlist hub context is cancelled (No playlist is being played)
			case <-ctx.Done():
				pm.Logger.Debug().Msg("playback manager: Playlist context cancelled")
				// Send event to the client -- nil signals that no playlist is being played
				pm.wsEventManager.SendEvent(events.PlaybackManagerPlaylistState, nil)
				return
			case path := <-pm.playlistHub.requestNewFileCh:
				// requestNewFileCh receives the path of the next video to play
				// The channel is fed when it's time to play the next video or when the client requests the next video
				// see: RequestNextPlaylistFile, playlistHub code
				pm.Logger.Debug().Str("path", path).Msg("playback manager: Playing next file")
				// Send notification to the client
				pm.wsEventManager.SendEvent(events.InfoToast, "Playing next file in playlist")
				// Play the requested video
				err := pm.MediaPlayerRepository.Play(path)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: Failed to play next file in playlist")
					pm.playlistHub.cancel()
					return
				}
				// Start tracking the video
				pm.MediaPlayerRepository.StartTracking()
			case <-pm.playlistHub.endOfPlaylistCh:
				pm.Logger.Debug().Msg("playback manager: End of playlist")
				pm.wsEventManager.SendEvent(events.InfoToast, "End of playlist")
				// Send event to the client -- nil signals that no playlist is being played
				pm.wsEventManager.SendEvent(events.PlaybackManagerPlaylistState, nil)
				go pm.MediaPlayerRepository.Stop()
				pm.playlistHub.cancel()
				return
			}
		}
	}()

	// Delete playlist in goroutine
	go func() {
		err := db_bridge.DeletePlaylist(pm.Database, playlist.DbId)
		if err != nil {
			pm.Logger.Error().Err(err).Str("name", playlist.Name).Msgf("playback manager: Failed to delete playlist")
			return
		}
		pm.Logger.Debug().Str("name", playlist.Name).Msgf("playback manager: Deleted playlist")
	}()

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pm *PlaybackManager) checkOrLoadAnimeCollection() (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/checkOrLoadAnimeCollection", &err)

	if pm.animeCollection.IsAbsent() {
		// If the anime collection is not present, we retrieve it from the platform
		collection, err := pm.platform.GetAnimeCollection(context.Background(), false)
		if err != nil {
			return err
		}
		pm.animeCollection = mo.Some(collection)
	}
	return nil
}

func (pm *PlaybackManager) SubscribeToPlaybackStatus(id string) *PlaybackStatusSubscriber {
	subscriber := &PlaybackStatusSubscriber{
		EventCh: make(chan PlaybackEvent, 100),
	}
	pm.playbackStatusSubscribers.Set(id, subscriber)
	return subscriber
}

func (pm *PlaybackManager) RegisterMediaPlayerCallback(callback func(event PlaybackEvent, cancelFunc func())) (cancel func()) {
	id := uuid.NewString()
	playbackSubscriber := pm.SubscribeToPlaybackStatus(id)
	cancel = func() {
		pm.UnsubscribeFromPlaybackStatus(id)
	}
	go func(playbackSubscriber *PlaybackStatusSubscriber) {
		for event := range playbackSubscriber.EventCh {
			callback(event, cancel)
		}
	}(playbackSubscriber)

	return cancel
}

func (pm *PlaybackManager) UnsubscribeFromPlaybackStatus(id string) {
	defer func() {
		if r := recover(); r != nil {
			pm.Logger.Warn().Msg("playback manager: Failed to unsubscribe from playback status")
		}
	}()
	subscriber, ok := pm.playbackStatusSubscribers.Get(id)
	if !ok {
		return
	}
	subscriber.canceled.Store(true)
	pm.playbackStatusSubscribers.Delete(id)
	close(subscriber.EventCh)
}
