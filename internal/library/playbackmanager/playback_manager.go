package playbackmanager

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
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

	// PlaybackManager is used as an interface between the video playback and progress tracking.
	// It can receive progress updates and dispatch appropriate events for:
	//  - syncing progress with AniList, MAL, etc.
	//  - sending notifications to the client
	//  - DEVNOTE: in the future, it could also be used to implement w2g, handle built-in player or allow multiple watchers
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
		refreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		mu                         sync.Mutex
		eventMu                    sync.Mutex
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
		// DEVNOTE: currentStreamEpisodeCollection and currentStreamEpisode can be absent when the user is streaming a video,
		// we will just not track the progress in that case
		// This is set by [SetStreamEpisodeCollection]
		currentStreamEpisodeCollection mo.Option[*anime.EpisodeCollection]
		// The current episode being streamed, set in [StartStreamingUsingMediaPlayer] by finding the episode in currentStreamEpisodeCollection
		currentStreamEpisode mo.Option[*anime.Episode]
		// The current media being streamed, set in [StartStreamingUsingMediaPlayer]
		currentStreamMedia mo.Option[*anilist.BaseAnime]

		// \/ Manual progress tracking (non-integrated external player)
		manualTrackingCtx           context.Context
		manualTrackingCtxCancel     context.CancelFunc
		manualTrackingPlaybackState PlaybackState
		currentManualTrackingState  mo.Option[*ManualTrackingState]
		manualTrackingWg            sync.WaitGroup

		// \/ Playlist
		playlistHub *playlistHub // The playlist hub

		isOffline       bool
		animeCollection mo.Option[*anilist.AnimeCollection]

		playbackStatusSubscribers *result.Map[string, *PlaybackStatusSubscriber]
	}

	PlaybackStatusSubscriber struct {
		PlaybackStateCh  chan PlaybackState
		PlaybackStatusCh chan mediaplayer.PlaybackStatus
		VideoStartedCh   chan string
		VideoStoppedCh   chan string
		VideoCompletedCh chan string

		StreamStateCh     chan PlaybackState
		StreamStatusCh    chan mediaplayer.PlaybackStatus
		StreamStartedCh   chan string
		StreamStoppedCh   chan string
		StreamCompletedCh chan string
	}

	PlaybackStateType string

	// PlaybackState is used to keep track of the user's current video playback
	// It is sent to the client each time the video playback state is picked up -- this is used to update the client's UI
	PlaybackState struct {
		EpisodeNumber        int     `json:"episodeNumber"`        // The episode number
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
		Database                   *db.Database
		RefreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		DiscordPresence            *discordrpc_presence.Presence
		IsOffline                  bool
		ContinuityManager          *continuity.Manager
	}

	Settings struct {
		AutoPlayNextEpisode bool
	}
)

func New(opts *NewPlaybackManagerOptions) *PlaybackManager {
	pm := &PlaybackManager{
		Logger:                         opts.Logger,
		Database:                       opts.Database,
		settings:                       &Settings{},
		discordPresence:                opts.DiscordPresence,
		wsEventManager:                 opts.WSEventManager,
		platform:                       opts.Platform,
		refreshAnimeCollectionFunc:     opts.RefreshAnimeCollectionFunc,
		mu:                             sync.Mutex{},
		autoPlayMu:                     sync.Mutex{},
		eventMu:                        sync.Mutex{},
		historyMap:                     make(map[string]PlaybackState),
		isOffline:                      opts.IsOffline,
		nextEpisodeLocalFile:           mo.None[*anime.LocalFile](),
		currentStreamEpisodeCollection: mo.None[*anime.EpisodeCollection](),
		currentStreamEpisode:           mo.None[*anime.Episode](),
		currentStreamMedia:             mo.None[*anilist.BaseAnime](),
		animeCollection:                mo.None[*anilist.AnimeCollection](),
		currentManualTrackingState:     mo.None[*ManualTrackingState](),
		currentLocalFile:               mo.None[*anime.LocalFile](),
		currentLocalFileWrapperEntry:   mo.None[*anime.LocalFileWrapperEntry](),
		currentMediaListEntry:          mo.None[*anilist.AnimeListEntry](),
		continuityManager:              opts.ContinuityManager,
		playbackStatusSubscribers:      result.NewResultMap[string, *PlaybackStatusSubscriber](),
	}

	pm.playlistHub = newPlaylistHub(pm)

	return pm
}

func (pm *PlaybackManager) SetStreamEpisodeCollection(ec []*anime.Episode) {
	// DEVNOTE: This is called from the torrentstream repository instance

	// Commented to fix potential deadlock
	//pm.mu.Lock()
	//defer pm.mu.Unlock()

	pm.Logger.Trace().Msg("playback manager: Setting stream episode collection")

	pm.currentStreamEpisodeCollection = mo.Some(&anime.EpisodeCollection{
		Episodes: ec,
	})
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
	if pm.isOffline {
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

	// Set the current episode being streamed
	// If the episode collection is not set, we'll still let the stream start. The progress will just not be tracked
	if pm.currentStreamEpisodeCollection.IsPresent() {
		for _, episode := range pm.currentStreamEpisodeCollection.MustGet().Episodes {
			if episode.AniDBEpisode == aniDbEpisode {
				episodeNumber = episode.EpisodeNumber
				pm.currentStreamEpisode = mo.Some(episode)
				break
			}
		}
	} else {
		//DEVNOTE: This will happen if the server is restarted and the stream page is not reloaded
		pm.Logger.Warn().Msg("playback manager: Stream episode collection is not set, no tracking will be done. Please refresh the page and try again if this shouldn't happen")
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

func (pm *PlaybackManager) Pause() error {
	return pm.MediaPlayerRepository.Pause()
}

func (pm *PlaybackManager) Resume() error {
	return pm.MediaPlayerRepository.Resume()
}

func (pm *PlaybackManager) Seek(seconds float64) error {
	return pm.MediaPlayerRepository.Seek(seconds)
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
			default:
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
		collection, err := pm.platform.GetAnimeCollection(false)
		if err != nil {
			return err
		}
		pm.animeCollection = mo.Some(collection)
	}
	return nil
}

func (pm *PlaybackManager) SubscribeToPlaybackStatus(id string) *PlaybackStatusSubscriber {
	subscriber := &PlaybackStatusSubscriber{
		PlaybackStateCh:  make(chan PlaybackState),
		PlaybackStatusCh: make(chan mediaplayer.PlaybackStatus),
		VideoStartedCh:   make(chan string),
		VideoStoppedCh:   make(chan string),
		VideoCompletedCh: make(chan string),

		StreamStateCh:     make(chan PlaybackState),
		StreamStatusCh:    make(chan mediaplayer.PlaybackStatus),
		StreamStartedCh:   make(chan string),
		StreamStoppedCh:   make(chan string),
		StreamCompletedCh: make(chan string),
	}
	pm.playbackStatusSubscribers.Set(id, subscriber)
	return subscriber
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
	close(subscriber.PlaybackStateCh)
	close(subscriber.PlaybackStatusCh)
	close(subscriber.VideoStartedCh)
	close(subscriber.VideoStoppedCh)
	close(subscriber.VideoCompletedCh)
	close(subscriber.StreamStateCh)
	close(subscriber.StreamStatusCh)
	close(subscriber.StreamStartedCh)
	close(subscriber.StreamStoppedCh)
	close(subscriber.StreamCompletedCh)
	pm.playbackStatusSubscribers.Delete(id)
}
