package playbackmanager

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/database/db"
	discordrpc_presence "github.com/seanime-app/seanime/internal/discordrpc/presence"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/offline"
	"sync"
)

type (
	// PlaybackManager is used as an interface between the video playback and progress tracking.
	// It can receive progress updates and dispatch appropriate events for:
	//  - syncing progress with AniList, MAL, etc.
	//  - sending notifications to the client
	//  - DEVNOTE: in the future, it could also be used to implement w2g, handle built-in player or allow multiple watchers
	PlaybackManager struct {
		Logger                       *zerolog.Logger
		Database                     *db.Database
		MediaPlayerRepository        *mediaplayer.Repository           // MediaPlayerRepository is used to control the media player
		discordPresence              *discordrpc_presence.Presence     // DiscordPresence is used to update the user's Discord presence
		mediaPlayerRepoSubscriber    *mediaplayer.RepositorySubscriber // Used to listen for media player events
		wsEventManager               events.IWSEventManager
		anilistClientWrapper         anilist.ClientWrapperInterface
		anilistCollection            *anilist.AnimeCollection
		refreshAnilistCollectionFunc func() // This function is called to refresh the AniList collection
		mu                           sync.Mutex
		eventMu                      sync.Mutex
		cancel                       context.CancelFunc
		// historyMap stores PlaybackState whose state is "completed"
		// Since PlaybackState is sent to client, once it is stored in historyMap, only the one stored in historyMap will be sent to client
		historyMap                   map[string]PlaybackState
		currentMediaPlaybackStatus   *mediaplayer.PlaybackStatus  // The current video playback status (can be nil)
		currentMediaListEntry        *anilist.MediaListEntry      // List Entry for the current video playback (can be nil)
		currentLocalFile             *anime.LocalFile             // Local file for the current video playback (can be nil)
		currentLocalFileWrapperEntry *anime.LocalFileWrapperEntry // This contains the current media entry local file data (can be nil)
		playlistHub                  *playlistHub                 // The playlist hub

		isOffline  bool
		offlineHub *offline.Hub
	}

	PlaybackStateType string

	// PlaybackState is used to keep track of the user's current video playback
	// It is sent to the client each time the video playback state is picked up -- this is used to update the client's UI
	PlaybackState struct {
		EpisodeNumber        int     `json:"episodeNumber"`        // The episode number
		MediaTitle           string  `json:"mediaTitle"`           // The title of the media
		MediaTotalEpisodes   int     `json:"mediaTotalEpisodes"`   // The total number of episodes
		Filename             string  `json:"filename"`             // The filename
		CompletionPercentage float64 `json:"completionPercentage"` // The completion percentage
		CanPlayNext          bool    `json:"canPlayNext"`          // Whether the next episode can be played
		ProgressUpdated      bool    `json:"progressUpdated"`      // Whether the progress has been updated
		MediaId              int     `json:"mediaId"`              // The media ID
	}

	NewPlaybackManagerOptions struct {
		WSEventManager               events.IWSEventManager
		Logger                       *zerolog.Logger
		AnilistClientWrapper         anilist.ClientWrapperInterface
		AnilistCollection            *anilist.AnimeCollection
		Database                     *db.Database
		RefreshAnilistCollectionFunc func() // This function is called to refresh the AniList collection
		DiscordPresence              *discordrpc_presence.Presence
		IsOffline                    bool
		OfflineHub                   *offline.Hub
	}
)

func New(opts *NewPlaybackManagerOptions) *PlaybackManager {
	return &PlaybackManager{
		Logger:                       opts.Logger,
		Database:                     opts.Database,
		discordPresence:              opts.DiscordPresence,
		wsEventManager:               opts.WSEventManager,
		anilistClientWrapper:         opts.AnilistClientWrapper,
		anilistCollection:            opts.AnilistCollection,
		refreshAnilistCollectionFunc: opts.RefreshAnilistCollectionFunc,
		playlistHub:                  newPlaylistHub(opts.Logger, opts.WSEventManager),
		mu:                           sync.Mutex{},
		historyMap:                   make(map[string]PlaybackState),
		isOffline:                    opts.IsOffline,
		offlineHub:                   opts.OfflineHub,
	}
}

func (pm *PlaybackManager) SetAnilistClientWrapper(anilistClientWrapper anilist.ClientWrapperInterface) {
	pm.anilistClientWrapper = anilistClientWrapper
}

func (pm *PlaybackManager) SetAnilistCollection(anilistCollection *anilist.AnimeCollection) {
	go func() {
		pm.mu.Lock()
		defer pm.mu.Unlock()
		pm.anilistCollection = anilistCollection
	}()
}

// PlayNextEpisode plays the next episode of the media that has been watched
//   - Called when the user clicks on the "Next" button in the client
//   - Should not be called when the user is watching a playlist
//   - Should not be called when no next episode is available
func (pm *PlaybackManager) PlayNextEpisode() error {
	if pm.currentLocalFile == nil || pm.currentMediaListEntry == nil || pm.currentLocalFileWrapperEntry == nil {
		return errors.New("could not play next episode")
	}

	nextLf, found := pm.currentLocalFileWrapperEntry.FindNextEpisode(pm.currentLocalFile)
	if !found {
		return errors.New("could not play next episode")
	}

	err := pm.MediaPlayerRepository.Play(nextLf.Path)
	if err != nil {
		return err
	}
	// Start tracking the video
	pm.MediaPlayerRepository.StartTracking()

	return nil
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

func (pm *PlaybackManager) StartPlayingUsingMediaPlayer(videopath string) error {
	pm.playlistHub.reset()

	// When offline, pm.anilistCollection is nil because SetAnilistCollection is not called
	// So, when starting a video, we retrieve the AnimeCollection from the OfflineHub
	if pm.isOffline && pm.anilistCollection == nil {
		snapshot, found := pm.offlineHub.RetrieveCurrentSnapshot()
		if !found {
			return errors.New("could not retrieve anime collection")
		}
		pm.anilistCollection = snapshot.Collections.AnimeCollection
	}

	err := pm.MediaPlayerRepository.Play(videopath)
	if err != nil {
		return err
	}

	pm.MediaPlayerRepository.StartTracking()

	return nil
}

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
func (pm *PlaybackManager) StartPlaylist(playlist *anime.Playlist) error {
	pm.playlistHub.loadPlaylist(playlist)

	// When offline, pm.anilistCollection is nil because SetAnilistCollection is not called
	// So, when starting a video, we retrieve the AnimeCollection from the OfflineHub
	if pm.isOffline && pm.anilistCollection == nil {
		snapshot, found := pm.offlineHub.RetrieveCurrentSnapshot()
		if !found {
			return errors.New("could not retrieve anime collection")
		}
		pm.anilistCollection = snapshot.Collections.AnimeCollection
	}

	// Play the first video in the playlist
	firstVidPath := playlist.LocalFiles[0].Path
	err := pm.MediaPlayerRepository.Play(firstVidPath)
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
					pm.Logger.Error().Err(err).Msg("playback manager: failed to play next file in playlist")
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
		err := pm.Database.DeletePlaylist(playlist.DbId)
		if err != nil {
			pm.Logger.Error().Err(err).Str("name", playlist.Name).Msgf("playback manager: Failed to delete playlist")
			return
		}
		pm.Logger.Debug().Str("name", playlist.Name).Msgf("playback manager: Deleted playlist")
	}()

	return nil
}
