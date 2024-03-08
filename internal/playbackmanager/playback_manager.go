package playbackmanager

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayer"
	"sync"
)

const (
	VideoPlaybackTracking  PlaybackStateType = "tracking"
	VideoPlaybackCompleted PlaybackStateType = "completed"
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
		mediaPlayerRepoSubscriber    *mediaplayer.RepositorySubscriber // Used to listen for media player events
		wsEventManager               events.IWSEventManager
		anilistClientWrapper         *anilist.ClientWrapper
		anilistCollection            *anilist.AnimeCollection
		refreshAnilistCollectionFunc func() // This function is called to refresh the AniList collection
		mu                           sync.Mutex
		eventMu                      sync.Mutex
		ctx                          context.Context
		cancel                       context.CancelFunc
		history                      []PlaybackState                 // This is used to keep track of the user's completed video playbacks
		historyMap                   map[string]PlaybackState        // This is used to keep track of the user's completed video playbacks (keyed by filename)
		currentMediaPlaybackStatus   *mediaplayer.PlaybackStatus     // The current video playback status (can be nil)
		currentMediaListEntry        *anilist.MediaListEntry         // List Entry for the current video playback (can be nil)
		currentLocalFile             *entities.LocalFile             // Local file for the current video playback (can be nil)
		currentLocalFileWrapperEntry *entities.LocalFileWrapperEntry // This contains the current media entry local file data (can be nil)
		playlistHub                  *playlistHub                    // The playlist hub
	}

	PlaybackStateType string

	PlaybackState struct {
		State                PlaybackStateType `json:"state"`                // The state of the video playback
		EpisodeNumber        int               `json:"episodeNumber"`        // The episode number
		MediaTitle           string            `json:"mediaTitle"`           // The title of the media
		MediaTotalEpisodes   int               `json:"mediaTotalEpisodes"`   // The total number of episodes
		Filename             string            `json:"filename"`             // The filename
		CompletionPercentage float64           `json:"completionPercentage"` // The completion percentage
		CanPlayNext          bool              `json:"canPlayNext"`          // Whether the next episode can be played
		ProgressUpdated      bool              `json:"progressUpdated"`      // Whether the progress has been updated
		MediaId              int               `json:"mediaId"`              // The media ID
	}

	NewProgressManagerOptions struct {
		WSEventManager               events.IWSEventManager
		Logger                       *zerolog.Logger
		AnilistClientWrapper         *anilist.ClientWrapper
		AnilistCollection            *anilist.AnimeCollection
		Database                     *db.Database
		RefreshAnilistCollectionFunc func() // This function is called to refresh the AniList collection
	}
)

func New(opts *NewProgressManagerOptions) *PlaybackManager {
	return &PlaybackManager{
		Logger:                       opts.Logger,
		Database:                     opts.Database,
		wsEventManager:               opts.WSEventManager,
		anilistClientWrapper:         opts.AnilistClientWrapper,
		anilistCollection:            opts.AnilistCollection,
		refreshAnilistCollectionFunc: opts.RefreshAnilistCollectionFunc,
		playlistHub:                  newPlaylistHub(opts.Logger, opts.WSEventManager),
		mu:                           sync.Mutex{},
		history:                      make([]PlaybackState, 0),
		historyMap:                   make(map[string]PlaybackState),
	}
}

func (pm *PlaybackManager) SetAnilistClientWrapper(anilistClientWrapper *anilist.ClientWrapper) {
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
// - This method is called when the user clicks on the "Next" button in the client
func (pm *PlaybackManager) PlayNextEpisode() error {
	panic("not implemented")
	// devnote: make sure not to relaunch the media player
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

		// Create a new context
		pm.ctx, pm.cancel = context.WithCancel(context.Background())

		pm.mu.Lock()
		// Set the media player repository
		pm.MediaPlayerRepository = mediaPlayerRepository
		// Set up event listeners for the media player
		pm.mediaPlayerRepoSubscriber = pm.MediaPlayerRepository.Subscribe("playbackmanager")
		pm.mu.Unlock()

		// Start listening to media player events
		pm.listenToMediaPlayerEvents()

		// DEVNOTE: pm.listenToClientPlayerEvents()
	}()
}

func (pm *PlaybackManager) StartPlayingUsingMediaPlayer(videopath string) error {
	pm.playlistHub.reset()

	err := pm.MediaPlayerRepository.Play(videopath)
	if err != nil {
		return err
	}

	pm.MediaPlayerRepository.StartTracking()

	return nil
}

func (pm *PlaybackManager) CancelCurrentPlaylist() error {
	go pm.playlistHub.reset()
	return nil
}

func (pm *PlaybackManager) RequestNextPlaylistFile() error {
	go pm.playlistHub.playNextFile()
	return nil
}

func (pm *PlaybackManager) StartPlaylist(playlist *entities.Playlist) error {
	pm.playlistHub.loadPlaylist(playlist)

	firstVidPath := playlist.LocalFiles[0].Path

	err := pm.MediaPlayerRepository.Play(firstVidPath)
	if err != nil {
		return err
	}

	pm.MediaPlayerRepository.StartTracking()

	var ctx context.Context
	ctx, pm.playlistHub.cancel = context.WithCancel(context.Background())

	// Listen to new play requests
	go func() {
		pm.Logger.Debug().Msg("playback manager: Listening for new file requests")
		for {
			select {
			case <-ctx.Done():
				pm.Logger.Debug().Msg("playback manager: Playlist context cancelled")
				pm.wsEventManager.SendEvent(events.PlaybackManagerPlaylistState, nil)
				return
			case path := <-pm.playlistHub.requestNewFileCh:
				pm.Logger.Debug().Str("path", path).Msg("playback manager: Playing next file")
				pm.wsEventManager.SendEvent(events.PlaybackManagerNotifyInfo, "Playing next file in playlist")
				err := pm.MediaPlayerRepository.Play(path)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: failed to play next file in playlist")
					pm.playlistHub.cancel()
					return
				}
				pm.MediaPlayerRepository.StartTracking()
			case <-pm.playlistHub.endOfPlaylistCh:
				pm.Logger.Debug().Msg("playback manager: End of playlist")
				pm.wsEventManager.SendEvent(events.PlaybackManagerNotifyInfo, "End of playlist")
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
