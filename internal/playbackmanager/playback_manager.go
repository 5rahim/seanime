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
	VideoPlaybackTracking  VideoPlaybackStateType = "tracking"
	VideoPlaybackCompleted VideoPlaybackStateType = "completed"
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
		mu                           sync.Mutex
		ctx                          context.Context
		cancel                       context.CancelFunc
		history                      []VideoPlaybackState            // This is used to keep track of the user's completed video playbacks
		currentMediaListEntry        *anilist.MediaListEntry         // List Entry for the current video playback (can be nil)
		currentLocalFile             *entities.LocalFile             // Local file for the current video playback (can be nil)
		currentLocalFileWrapperEntry *entities.LocalFileWrapperEntry // This contains the current media entry local file data
	}

	VideoPlaybackStateType string

	VideoPlaybackState struct {
		State                VideoPlaybackStateType `json:"state"`                // The state of the video playback
		EpisodeNumber        int                    `json:"episodeNumber"`        // The episode number
		MediaTitle           string                 `json:"mediaTitle"`           // The title of the media
		MediaTotalEpisodes   int                    `json:"MediaTotalEpisodes"`   // The total number of episodes
		Filename             string                 `json:"filename"`             // The filename
		CompletionPercentage float64                `json:"completionPercentage"` // The completion percentage
		CanPlayNext          bool                   `json:"canPlayNext"`          // Whether the next episode can be played
	}

	Playlist struct {
		localFiles []*entities.LocalFile
		media      *anilist.BaseMedia
	}

	NewProgressManagerOptions struct {
		WSEventManager       events.IWSEventManager
		Logger               *zerolog.Logger
		AnilistClientWrapper *anilist.ClientWrapper
		AnilistCollection    *anilist.AnimeCollection
		Database             *db.Database
	}
)

func New(opts *NewProgressManagerOptions) *PlaybackManager {
	return &PlaybackManager{
		Logger:               opts.Logger,
		Database:             opts.Database,
		wsEventManager:       opts.WSEventManager,
		anilistClientWrapper: opts.AnilistClientWrapper,
		anilistCollection:    opts.AnilistCollection,
		mu:                   sync.Mutex{},
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
func (pm *PlaybackManager) PlayNextEpisode() {
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

func (pm *PlaybackManager) listenToMediaPlayerEvents() {
	// Listen for media player events
	go func() {
		for {
			select {
			case <-pm.ctx.Done(): // Context has been cancelled
				return
			case status := <-pm.mediaPlayerRepoSubscriber.TrackingStartedCh: // New video has started playing
				// Send event to the client
				vps := pm.getVideoPlaybackState(status)
				pm.Logger.Debug().Msg("playback manager: Tracking started, extracting metadata...")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStarted, vps)

				// Retrieve data about the current video playback
				// Set PlaybackManager.currentMediaListEntry to the list entry of the current video
				var err error
				pm.currentMediaListEntry, pm.currentLocalFile, pm.currentLocalFileWrapperEntry, err = pm.getListEntryFromLocalFilePath(status.Filename)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: failed to get media data")
					// Send error event to the client
					pm.wsEventManager.SendEvent(events.PlaybackManagerMetadataError, err.Error())
				}

				pm.Logger.Debug().Msgf("playback manager: Watching %s - Episode %d", pm.currentMediaListEntry.GetMedia().GetPreferredTitle(), pm.currentLocalFile.GetEpisodeNumber())

			case status := <-pm.mediaPlayerRepoSubscriber.VideoCompletedCh: // Video has been watched completely but still tracking
				vps := pm.getVideoPlaybackState(status)
				pm.Logger.Debug().Msg("playback manager: Video completed")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressCompleted, vps)
				// Push the video playback state to the history
				pm.history = append(pm.history, vps)

			case path := <-pm.mediaPlayerRepoSubscriber.TrackingStoppedCh: // Tracking has stopped completely
				pm.Logger.Debug().Msg("playback manager: Received tracking stopped event")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStopped, path)

			case playbackStatus := <-pm.mediaPlayerRepoSubscriber.PlaybackStatusCh: // Playback status has changed
				pm.wsEventManager.SendEvent(events.MediaPlayerPlaybackStatus, playbackStatus)

			case _ = <-pm.mediaPlayerRepoSubscriber.TrackingRetryCh: // Error occurred while starting tracking
				// DEVNOTE: This event is not sent to the client
			}
		}
	}()
}

func (pm *PlaybackManager) getVideoPlaybackState(status *mediaplayer.PlaybackStatus) VideoPlaybackState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentLocalFileWrapperEntry == nil || pm.currentLocalFile == nil || pm.currentMediaListEntry == nil {
		return VideoPlaybackState{}
	}

	state := VideoPlaybackTracking
	if status.CompletionPercentage > 0.9 {
		state = VideoPlaybackCompleted
	}
	// Find the following episode
	_, canPlayNext := pm.currentLocalFileWrapperEntry.FindNextEpisode(pm.currentLocalFile)
	return VideoPlaybackState{
		State:                state,
		EpisodeNumber:        pm.currentLocalFile.GetEpisodeNumber(),
		MediaTitle:           pm.currentMediaListEntry.GetMedia().GetPreferredTitle(),
		MediaTotalEpisodes:   pm.currentMediaListEntry.GetMedia().GetCurrentEpisodeCount(),
		Filename:             status.Filename,
		CompletionPercentage: status.CompletionPercentage,
		CanPlayNext:          canPlayNext,
	}
}

func (pm *PlaybackManager) StartPlayingUsingMediaPlayer(videopath string) error {
	err := pm.MediaPlayerRepository.Play(videopath)
	if err != nil {
		return err
	}

	pm.MediaPlayerRepository.StartTracking()

	return nil
}
