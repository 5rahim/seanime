package playbackmanager

import (
	"context"
	"errors"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mal"
	"github.com/seanime-app/seanime/internal/mediaplayer"
	"github.com/seanime-app/seanime/internal/util"
)

var (
	ErrProgressUpdateAnilist = errors.New("playback manager: Failed to update progress on AniList")
	ErrProgressUpdateMAL     = errors.New("playback manager: Failed to update progress on MyAnimeList")
)

func (pm *PlaybackManager) listenToMediaPlayerEvents() {
	// Listen for media player events
	go func() {
		for {
			select {
			case <-pm.ctx.Done(): // Context has been cancelled
				return
			case status := <-pm.mediaPlayerRepoSubscriber.TrackingStartedCh: // New video has started playing
				pm.eventMu.Lock()
				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getPlaybackState(status)
				// Log
				pm.Logger.Debug().Msg("playback manager: Tracking started, extracting metadata...")
				// Send event to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStarted, _ps)

				// Retrieve data about the current video playback
				// Set PlaybackManager.currentMediaListEntry to the list entry of the current video
				var err error
				pm.currentMediaListEntry, pm.currentLocalFile, pm.currentLocalFileWrapperEntry, err = pm.getListEntryFromLocalFilePath(status.Filename)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: failed to get media data")
					// Send error event to the client
					pm.wsEventManager.SendEvent(events.PlaybackManagerProgressMetadataError, err.Error())
				} else {
					pm.Logger.Debug().Msgf("playback manager: Watching %s - Episode %d", pm.currentMediaListEntry.GetMedia().GetPreferredTitle(), pm.currentLocalFile.GetEpisodeNumber())
				}

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.VideoCompletedCh: // Video has been watched completely but still tracking
				pm.eventMu.Lock()
				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getPlaybackState(status)
				// Log
				pm.Logger.Debug().Msg("playback manager: Received video completed event")

				//
				// Update the progress on AniList if auto update progress is enabled
				//
				pm.autoSyncCurrentProgress(&_ps)

				// Send the playback state with the `ProgressUpdated` flag
				// The client will use this to notify the user if the progress has been updated
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressVideoCompleted, _ps)
				// Push the video playback state to the history
				pm.history = append(pm.history, _ps)

				pm.eventMu.Unlock()
			case path := <-pm.mediaPlayerRepoSubscriber.TrackingStoppedCh: // Tracking has stopped completely
				pm.eventMu.Lock()

				pm.Logger.Debug().Msg("playback manager: Received tracking stopped event")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStopped, path)

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.PlaybackStatusCh: // Playback status has changed
				pm.eventMu.Lock()

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getPlaybackState(status)
				// Update the playback state if the filename is in the history
				// This is done so the completion status of the PlaybackState is not overwritten
				for _, h := range pm.history {
					if h.Filename == _ps.Filename {
						_ps = h
						break
					}
				}
				// Send the playback state to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressPlaybackState, _ps)

				pm.eventMu.Unlock()
			case _ = <-pm.mediaPlayerRepoSubscriber.TrackingRetryCh: // Error occurred while starting tracking
				// DEVNOTE: This event is not sent to the client
			}
		}
	}()
}

func (pm *PlaybackManager) getPlaybackState(status *mediaplayer.PlaybackStatus) PlaybackState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentLocalFileWrapperEntry == nil || pm.currentLocalFile == nil || pm.currentMediaListEntry == nil {
		return PlaybackState{}
	}

	state := VideoPlaybackTracking
	if status.CompletionPercentage > 0.9 {
		state = VideoPlaybackCompleted
	}
	// Find the following episode
	_, canPlayNext := pm.currentLocalFileWrapperEntry.FindNextEpisode(pm.currentLocalFile)
	return PlaybackState{
		State:                state,
		EpisodeNumber:        pm.currentLocalFile.GetEpisodeNumber(),
		MediaTitle:           pm.currentMediaListEntry.GetMedia().GetPreferredTitle(),
		MediaTotalEpisodes:   pm.currentMediaListEntry.GetMedia().GetCurrentEpisodeCount(),
		MediaId:              pm.currentMediaListEntry.GetMedia().GetID(),
		Filename:             status.Filename,
		CompletionPercentage: status.CompletionPercentage,
		CanPlayNext:          canPlayNext,
	}
}

// SyncCurrentProgress syncs the current video playback progress with providers
func (pm *PlaybackManager) autoSyncCurrentProgress(_ps *PlaybackState) {
	if shouldUpdate, err := pm.Database.AutoUpdateProgressIsEnabled(); err == nil && shouldUpdate {
		// Update the progress on AniList
		pm.Logger.Debug().Msg("playback manager: Updating progress on AniList")
		err := pm.updateProgress(pm.currentMediaListEntry, pm.currentLocalFile)

		if err != nil && errors.Is(err, ErrProgressUpdateMAL) {
			_ps.ProgressUpdated = false
		} else {
			_ps.ProgressUpdated = true
			pm.wsEventManager.SendEvent(events.PlaybackManagerProgressUpdated, _ps)
		}
	} else if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Failed to check if auto update progress is enabled")
	}
}

// SyncCurrentProgress syncs the current video playback progress with providers
// This method is called when the user manually requests to sync the progress
//   - This method will return an error only if the progress update fails on AniList
//   - This method will refresh the anilist collection
func (pm *PlaybackManager) SyncCurrentProgress() error {
	pm.eventMu.Lock()
	defer pm.eventMu.Unlock()
	if pm.currentMediaListEntry == nil || pm.currentLocalFile == nil {
		return errors.New("no video is being watched")
	}

	err := pm.updateProgress(pm.currentMediaListEntry, pm.currentLocalFile)
	if err != nil && errors.Is(err, ErrProgressUpdateAnilist) {
		return err
	}

	// Push the current playback state to the history
	if pm.currentMediaPlaybackStatus != nil {
		_ps := pm.getPlaybackState(pm.currentMediaPlaybackStatus)
		_ps.ProgressUpdated = true
		pm.history = append(pm.history, _ps)
		pm.wsEventManager.SendEvent(events.PlaybackManagerProgressUpdated, _ps)
	}

	pm.refreshAnilistCollectionFunc()

	return nil
}

// updateProgress updates the progress of the current video playback on AniList and MyAnimeList
//   - /!\ When this is called, the PlaybackState should have been pushed to the history
func (pm *PlaybackManager) updateProgress(listEntry *anilist.MediaListEntry, localFile *entities.LocalFile) (err error) {

	defer util.HandlePanicInModuleWithError("playbackmanager/updateProgress", &err)

	mediaId := listEntry.GetMedia().GetID()
	epNum := localFile.GetEpisodeNumber()
	totalEpisodes := listEntry.GetMedia().GetTotalEpisodeCount()

	// Update the progress on AniList
	err = pm.anilistClientWrapper.UpdateMediaListEntryProgress(
		context.Background(),
		&mediaId,
		&epNum,
		&totalEpisodes,
	)
	if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Error occurred while updating progress on AniList")
		return ErrProgressUpdateAnilist
	}

	pm.refreshAnilistCollectionFunc() // Refresh the AniList collection

	pm.Logger.Info().Msg("playback manager: Updated progress on AniList")

	// Update the progress on MAL if an account is linked
	malInfo, _ := pm.Database.GetMalInfo()
	if malInfo != nil && malInfo.AccessToken != "" {
		client := mal.NewWrapper(malInfo.AccessToken)
		err = client.UpdateAnimeProgress(&mal.AnimeListProgressParams{
			NumEpisodesWatched: &epNum,
		}, mediaId)
		if err != nil {
			pm.Logger.Error().Err(err).Msg("playback manager: Error occurred while updating progress on MyAnimeList")
			return ErrProgressUpdateMAL
		}
		pm.Logger.Info().Msg("playback manager: Updated progress on MyAnimeList")
	}

	return nil
}
