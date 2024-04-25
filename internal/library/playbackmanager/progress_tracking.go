package playbackmanager

import (
	"context"
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/mal"
	"github.com/seanime-app/seanime/internal/discordrpc/presence"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/util"
)

var (
	ErrProgressUpdateAnilist = errors.New("playback manager: Failed to update progress on AniList")
	ErrProgressUpdateMAL     = errors.New("playback manager: Failed to update progress on MyAnimeList")
)

func (pm *PlaybackManager) listenToMediaPlayerEvents(ctx context.Context) {
	// Listen for media player events
	go func() {
		for {
			select {
			// Stop listening when the context is cancelled -- meaning a new MediaPlayer instance is set
			case <-ctx.Done():
				return
			case status := <-pm.mediaPlayerRepoSubscriber.TrackingStartedCh: // New video has started playing
				pm.eventMu.Lock()

				// Reset the history map
				pm.historyMap = make(map[string]PlaybackState)

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
					pm.wsEventManager.SendEvent(events.ErrorToast, err.Error())
					//
					pm.MediaPlayerRepository.Cancel()
				} else {
					pm.Logger.Debug().
						Str("media", pm.currentMediaListEntry.GetMedia().GetPreferredTitle()).
						Int("episode", pm.currentLocalFile.GetEpisodeNumber()).
						Msg("playback manager: Playback started")
				}

				// ------- Playlist ------- //
				go pm.playlistHub.onVideoStart(pm.currentMediaListEntry, pm.currentLocalFile, pm.anilistCollection, _ps)

				// ------- Discord ------- //
				if pm.discordPresence != nil && !pm.isOffline {
					go pm.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
						Title:         pm.currentMediaListEntry.GetMedia().GetPreferredTitle(),
						Image:         pm.currentMediaListEntry.GetMedia().GetCoverImageSafe(),
						IsMovie:       pm.currentMediaListEntry.GetMedia().IsMovie(),
						EpisodeNumber: pm.currentLocalFileWrapperEntry.GetProgressNumber(pm.currentLocalFile),
					})
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
				pm.historyMap[status.Filename] = _ps

				// ------- Playlist ------- //
				go pm.playlistHub.onVideoCompleted(pm.currentMediaListEntry, pm.currentLocalFile, _ps)

				pm.eventMu.Unlock()
			case path := <-pm.mediaPlayerRepoSubscriber.TrackingStoppedCh: // Tracking has stopped completely
				pm.eventMu.Lock()

				pm.Logger.Debug().Msg("playback manager: Received tracking stopped event")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStopped, path)

				// ------- Playlist ------- //
				go pm.playlistHub.onTrackingStopped()

				// ------- Discord ------- //
				if pm.discordPresence != nil && !pm.isOffline {
					go pm.discordPresence.Close()
				}

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.PlaybackStatusCh: // Playback status has changed
				pm.eventMu.Lock()

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getPlaybackState(status)
				// If the same PlaybackState is in the history, update the ProgressUpdated flag
				// PlaybackStatusCh has no way of knowing if the progress has been updated
				if h, ok := pm.historyMap[status.Filename]; ok {
					_ps.ProgressUpdated = h.ProgressUpdated
				}
				// Send the playback state to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressPlaybackState, _ps)

				// ------- Playlist ------- //
				go pm.playlistHub.onPlaybackStatus(pm.currentMediaListEntry, pm.currentLocalFile, _ps)

				pm.eventMu.Unlock()
			case _ = <-pm.mediaPlayerRepoSubscriber.TrackingRetryCh: // Error occurred while starting tracking
				// DEVNOTE: This event is not sent to the client

				// ------- Playlist ------- //
				go pm.playlistHub.onTrackingError()
			}
		}
	}()
}

// getPlaybackState returns a new PlaybackState
func (pm *PlaybackManager) getPlaybackState(status *mediaplayer.PlaybackStatus) PlaybackState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentLocalFileWrapperEntry == nil || pm.currentLocalFile == nil || pm.currentMediaListEntry == nil {
		return PlaybackState{}
	}

	// Find the following episode
	_, canPlayNext := pm.currentLocalFileWrapperEntry.FindNextEpisode(pm.currentLocalFile)
	return PlaybackState{
		EpisodeNumber:        pm.currentLocalFileWrapperEntry.GetProgressNumber(pm.currentLocalFile),
		MediaTitle:           pm.currentMediaListEntry.GetMedia().GetPreferredTitle(),
		MediaTotalEpisodes:   pm.currentMediaListEntry.GetMedia().GetCurrentEpisodeCount(),
		MediaId:              pm.currentMediaListEntry.GetMedia().GetID(),
		Filename:             status.Filename,
		CompletionPercentage: status.CompletionPercentage,
		CanPlayNext:          canPlayNext,
	}
}

// autoSyncCurrentProgress syncs the current video playback progress with providers.
// This is called once when a "video complete" event is heard.
func (pm *PlaybackManager) autoSyncCurrentProgress(_ps *PlaybackState) {
	if pm.currentLocalFile == nil || pm.currentMediaListEntry == nil {
		return
	}

	if shouldUpdate, err := pm.Database.AutoUpdateProgressIsEnabled(); err == nil && shouldUpdate {

		// Check if we should update the progress
		// If the current progress is lower than the episode progress number
		epProgressNum := pm.currentLocalFileWrapperEntry.GetProgressNumber(pm.currentLocalFile)
		if *pm.currentMediaListEntry.Progress >= epProgressNum {
			return
		}

		// Update the progress on AniList
		pm.Logger.Debug().Msg("playback manager: Updating progress on AniList")
		err := pm.updateProgress()

		if err != nil {
			_ps.ProgressUpdated = false
			pm.wsEventManager.SendEvent(events.ErrorToast, "Failed to update progress on AniList")
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
	if pm.currentMediaListEntry == nil || pm.currentLocalFile == nil {
		return errors.New("no video is being watched")
	}

	err := pm.updateProgress()
	if err != nil {
		pm.eventMu.Unlock()
		return err
	}

	// Push the current playback state to the history
	if pm.currentMediaPlaybackStatus != nil {
		_ps := pm.getPlaybackState(pm.currentMediaPlaybackStatus)
		_ps.ProgressUpdated = true
		pm.historyMap[pm.currentMediaPlaybackStatus.Filename] = _ps
		pm.wsEventManager.SendEvent(events.PlaybackManagerProgressUpdated, _ps)
	}

	pm.refreshAnilistCollectionFunc()

	pm.eventMu.Unlock()
	return nil
}

// updateProgress updates the progress of the current video playback on AniList and MyAnimeList.
// This only returns an error if the progress update fails on AniList
//   - /!\ When this is called, the PlaybackState should have been pushed to the history
func (pm *PlaybackManager) updateProgress() (err error) {

	defer util.HandlePanicInModuleWithError("playbackmanager/updateProgress", &err)

	//
	// Offline
	//
	if pm.isOffline {
		return pm.updateProgressOffline()
	}

	//
	// Online
	//

	mediaId := pm.currentMediaListEntry.GetMedia().GetID()
	epNum := pm.currentLocalFileWrapperEntry.GetProgressNumber(pm.currentLocalFile)
	totalEpisodes := pm.currentMediaListEntry.GetMedia().GetTotalEpisodeCount()

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

	go func() {
		defer util.HandlePanicThen(func() {})
		malId := pm.currentMediaListEntry.GetMedia().GetIDMal()
		if malId != nil {
			// Update the progress on MAL if an account is linked
			malInfo, _ := pm.Database.GetMalInfo()
			if malInfo != nil && malInfo.AccessToken != "" {
				// Verify MAL auth
				malInfo, err = mal.VerifyMALAuth(malInfo, pm.Database, pm.Logger)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: Error occurred while verifying MAL auth")
					return
				}

				client := mal.NewWrapper(malInfo.AccessToken, pm.Logger)

				err = client.UpdateAnimeProgress(&mal.AnimeListProgressParams{
					NumEpisodesWatched: &epNum,
				}, *malId)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: Error occurred while updating progress on MyAnimeList")
					return
				}
				pm.Logger.Info().Msg("playback manager: Updated progress on MyAnimeList")
			}
		} else {
			pm.Logger.Debug().Msg("playback manager: MAL ID not found, skipping update on MyAnimeList")
		}
	}()

	return nil
}

func (pm *PlaybackManager) updateProgressOffline() (err error) {

	mediaId := pm.currentMediaListEntry.GetMedia().GetID()
	epNum := pm.currentLocalFileWrapperEntry.GetProgressNumber(pm.currentLocalFile)
	totalEpisodes := pm.currentMediaListEntry.GetMedia().GetTotalEpisodeCount()

	totalEp := 0
	if totalEpisodes != 0 && totalEpisodes > 0 {
		totalEp = totalEpisodes
	}

	status := anilist.MediaListStatusCurrent
	if totalEp > 0 && epNum >= totalEp {
		status = anilist.MediaListStatusCompleted
	}

	if totalEp > 0 && epNum > totalEp {
		epNum = totalEp
	}

	return pm.offlineHub.UpdateAnimeListStatus(mediaId, epNum, status)
}
