package playbackmanager

import (
	"cmp"
	"context"
	"errors"
	"github.com/samber/mo"
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

				// Set the playback type
				pm.currentPlaybackType = LocalFilePlayback

				// Reset the history map
				pm.historyMap = make(map[string]PlaybackState)

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getLocalFilePlaybackState(status)
				// Log
				pm.Logger.Debug().Msg("playback manager: Tracking started, extracting metadata...")
				// Send event to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStarted, _ps)

				// Retrieve data about the current video playback
				// Set PlaybackManager.currentMediaListEntry to the list entry of the current video
				currentMediaListEntry, currentLocalFile, currentLocalFileWrapperEntry, err := pm.getLocalFilePlaybackDetails(status.Filename)
				if err != nil {
					pm.Logger.Error().Err(err).Msg("playback manager: Failed to get media data")
					// Send error event to the client
					pm.wsEventManager.SendEvent(events.ErrorToast, err.Error())
					//
					pm.MediaPlayerRepository.Cancel()

					pm.eventMu.Unlock()
					continue
				}

				pm.currentMediaListEntry = mo.Some(currentMediaListEntry)
				pm.currentLocalFile = mo.Some(currentLocalFile)
				pm.currentLocalFileWrapperEntry = mo.Some(currentLocalFileWrapperEntry)
				pm.Logger.Debug().
					Str("media", pm.currentMediaListEntry.MustGet().GetMedia().GetPreferredTitle()).
					Int("episode", pm.currentLocalFile.MustGet().GetEpisodeNumber()).
					Msg("playback manager: Playback started")

				// ------- Playlist ------- //
				go pm.playlistHub.onVideoStart(pm.currentMediaListEntry.MustGet(), pm.currentLocalFile.MustGet(), _ps)

				// ------- Discord ------- //
				if pm.discordPresence != nil && !pm.isOffline {
					go pm.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
						Title:         pm.currentMediaListEntry.MustGet().GetMedia().GetPreferredTitle(),
						Image:         pm.currentMediaListEntry.MustGet().GetMedia().GetCoverImageSafe(),
						IsMovie:       pm.currentMediaListEntry.MustGet().GetMedia().IsMovie(),
						EpisodeNumber: pm.currentLocalFileWrapperEntry.MustGet().GetProgressNumber(pm.currentLocalFile.MustGet()),
					})
				}

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.VideoCompletedCh: // Video has been watched completely but still tracking
				pm.eventMu.Lock()
				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getLocalFilePlaybackState(status)
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
				if pm.currentMediaListEntry.IsPresent() && pm.currentLocalFile.IsPresent() {
					go pm.playlistHub.onVideoCompleted(pm.currentMediaListEntry.MustGet(), pm.currentLocalFile.MustGet(), _ps)
				}

				pm.eventMu.Unlock()
			case reason := <-pm.mediaPlayerRepoSubscriber.TrackingStoppedCh: // Tracking has stopped completely
				pm.eventMu.Lock()

				pm.Logger.Debug().Msg("playback manager: Received tracking stopped event")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStopped, reason)

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
				_ps := pm.getLocalFilePlaybackState(status)
				// If the same PlaybackState is in the history, update the ProgressUpdated flag
				// PlaybackStatusCh has no way of knowing if the progress has been updated
				if h, ok := pm.historyMap[status.Filename]; ok {
					_ps.ProgressUpdated = h.ProgressUpdated
				}
				// Send the playback state to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressPlaybackState, _ps)

				// ------- Playlist ------- //
				if pm.currentMediaListEntry.IsPresent() && pm.currentLocalFile.IsPresent() {
					go pm.playlistHub.onPlaybackStatus(pm.currentMediaListEntry.MustGet(), pm.currentLocalFile.MustGet(), _ps)
				}

				pm.eventMu.Unlock()
			case _ = <-pm.mediaPlayerRepoSubscriber.TrackingRetryCh: // Error occurred while starting tracking
				// DEVNOTE: This event is not sent to the client
				// We notify the playlist hub, so it can play the next episode (it's assumed that the user closed the player)

				// ------- Playlist ------- //
				go pm.playlistHub.onTrackingError()

				//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			case status := <-pm.mediaPlayerRepoSubscriber.StreamingTrackingStartedCh:
				pm.eventMu.Lock()
				if pm.currentStreamEpisode.IsAbsent() || pm.currentStreamMedia.IsAbsent() {
					pm.eventMu.Unlock()
					continue
				}

				// Get the media list entry
				// Note that it might be absent if the user is watching a stream that is not in the library
				pm.currentMediaListEntry = pm.getStreamPlaybackDetails(pm.currentStreamMedia.MustGet().ID)

				// Set the playback type
				pm.currentPlaybackType = StreamPlayback

				// Reset the history map
				pm.historyMap = make(map[string]PlaybackState)

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getStreamPlaybackState(status)

				// Log
				pm.Logger.Debug().Msg("playback manager: Tracking started for stream")
				// Send event to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStarted, _ps)

				// ------- Discord ------- //
				if pm.discordPresence != nil && !pm.isOffline {
					go pm.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
						Title:         pm.currentStreamMedia.MustGet().GetPreferredTitle(),
						Image:         pm.currentStreamMedia.MustGet().GetCoverImageSafe(),
						IsMovie:       pm.currentStreamMedia.MustGet().IsMovie(),
						EpisodeNumber: pm.currentStreamEpisode.MustGet().GetProgressNumber(), // GetEpisodeNumber maybe?
					})
				}

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.StreamingPlaybackStatusCh:
				pm.eventMu.Lock()
				if pm.currentStreamEpisode.IsAbsent() {
					pm.eventMu.Unlock()
					continue
				}

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getStreamPlaybackState(status)
				// If the same PlaybackState is in the history, update the ProgressUpdated flag
				// PlaybackStatusCh has no way of knowing if the progress has been updated
				if h, ok := pm.historyMap[status.Filename]; ok {
					_ps.ProgressUpdated = h.ProgressUpdated
				}
				// Send the playback state to the client
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressPlaybackState, _ps)

				pm.eventMu.Unlock()
			case status := <-pm.mediaPlayerRepoSubscriber.StreamingVideoCompletedCh:
				pm.eventMu.Lock()
				if pm.currentStreamEpisode.IsAbsent() {
					pm.eventMu.Unlock()
					continue
				}

				// Set the current media playback status
				pm.currentMediaPlaybackStatus = status
				// Get the playback state
				_ps := pm.getStreamPlaybackState(status)
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

				pm.eventMu.Unlock()
			case reason := <-pm.mediaPlayerRepoSubscriber.StreamingTrackingStoppedCh:
				pm.eventMu.Lock()
				if pm.currentStreamEpisode.IsAbsent() {
					pm.eventMu.Unlock()
					continue
				}

				pm.Logger.Debug().Msg("playback manager: Received tracking stopped event")
				pm.wsEventManager.SendEvent(events.PlaybackManagerProgressTrackingStopped, reason)

				// ------- Discord ------- //
				if pm.discordPresence != nil && !pm.isOffline {
					go pm.discordPresence.Close()
				}

				pm.eventMu.Unlock()
			case _ = <-pm.mediaPlayerRepoSubscriber.StreamingTrackingRetryCh:
				// Do nothing
			}
		}
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local File
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getLocalFilePlaybackState returns a new PlaybackState
func (pm *PlaybackManager) getLocalFilePlaybackState(status *mediaplayer.PlaybackStatus) PlaybackState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentLocalFileWrapperEntry.IsAbsent() || pm.currentLocalFile.IsAbsent() || pm.currentMediaListEntry.IsAbsent() {
		return PlaybackState{}
	}

	// Find the following episode
	_, canPlayNext := pm.currentLocalFileWrapperEntry.MustGet().FindNextEpisode(pm.currentLocalFile.MustGet())
	return PlaybackState{
		EpisodeNumber:        pm.currentLocalFileWrapperEntry.MustGet().GetProgressNumber(pm.currentLocalFile.MustGet()),
		MediaTitle:           pm.currentMediaListEntry.MustGet().GetMedia().GetPreferredTitle(),
		MediaTotalEpisodes:   pm.currentMediaListEntry.MustGet().GetMedia().GetCurrentEpisodeCount(),
		MediaId:              pm.currentMediaListEntry.MustGet().GetMedia().GetID(),
		Filename:             status.Filename,
		CompletionPercentage: status.CompletionPercentage,
		CanPlayNext:          canPlayNext,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Stream
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getStreamPlaybackState returns a new PlaybackState
func (pm *PlaybackManager) getStreamPlaybackState(status *mediaplayer.PlaybackStatus) PlaybackState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.currentStreamEpisodeCollection.IsAbsent() || pm.currentStreamEpisode.IsAbsent() || pm.currentStreamMedia.IsAbsent() {
		return PlaybackState{}
	}

	return PlaybackState{
		EpisodeNumber:        pm.currentStreamEpisode.MustGet().GetProgressNumber(),
		MediaTitle:           pm.currentStreamMedia.MustGet().GetPreferredTitle(),
		MediaTotalEpisodes:   pm.currentStreamMedia.MustGet().GetCurrentEpisodeCount(),
		MediaId:              pm.currentStreamMedia.MustGet().GetID(),
		Filename:             cmp.Or(status.Filename, "Stream"),
		CompletionPercentage: status.CompletionPercentage,
		CanPlayNext:          false, // DEVNOTE: This is not used for streams
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// autoSyncCurrentProgress syncs the current video playback progress with providers.
// This is called once when a "video complete" event is heard.
func (pm *PlaybackManager) autoSyncCurrentProgress(_ps *PlaybackState) {

	shouldUpdate, err := pm.Database.AutoUpdateProgressIsEnabled()
	if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Failed to check if auto update progress is enabled")
		return
	}

	if !shouldUpdate {
		return
	}

	switch pm.currentPlaybackType {
	case LocalFilePlayback:
		// Note :currentMediaListEntry MUST be defined since we assume that the media is in the user's library
		if pm.currentMediaListEntry.IsAbsent() || pm.currentLocalFileWrapperEntry.IsAbsent() || pm.currentLocalFile.IsAbsent() {
			return
		}
		// Check if we should update the progress
		// If the current progress is lower than the episode progress number
		epProgressNum := pm.currentLocalFileWrapperEntry.MustGet().GetProgressNumber(pm.currentLocalFile.MustGet())
		if *pm.currentMediaListEntry.MustGet().Progress >= epProgressNum {
			return
		}

	case StreamPlayback:
		if pm.currentStreamEpisode.IsAbsent() || pm.currentStreamMedia.IsAbsent() {
			return
		}
		// Do not auto update progress is the media is in the library AND the progress is higher than the current episode
		epProgressNum := pm.currentStreamEpisode.MustGet().GetProgressNumber()
		if pm.currentMediaListEntry.IsPresent() && *pm.currentMediaListEntry.MustGet().Progress >= epProgressNum {
			return
		}
	}

	// Update the progress on AniList
	pm.Logger.Debug().Msg("playback manager: Updating progress on AniList")
	err = pm.updateProgress()

	if err != nil {
		_ps.ProgressUpdated = false
		pm.wsEventManager.SendEvent(events.ErrorToast, "Failed to update progress on AniList")
	} else {
		_ps.ProgressUpdated = true
		pm.wsEventManager.SendEvent(events.PlaybackManagerProgressUpdated, _ps)
	}

}

// SyncCurrentProgress syncs the current video playback progress with providers
// This method is called when the user manually requests to sync the progress
//   - This method will return an error only if the progress update fails on AniList
//   - This method will refresh the anilist collection
func (pm *PlaybackManager) SyncCurrentProgress() error {
	pm.eventMu.Lock()

	err := pm.updateProgress()
	if err != nil {
		pm.eventMu.Unlock()
		return err
	}

	// Push the current playback state to the history
	if pm.currentMediaPlaybackStatus != nil {
		var _ps PlaybackState
		switch pm.currentPlaybackType {
		case LocalFilePlayback:
			pm.getLocalFilePlaybackState(pm.currentMediaPlaybackStatus)
		case StreamPlayback:
			pm.getStreamPlaybackState(pm.currentMediaPlaybackStatus)
		}
		_ps.ProgressUpdated = true
		pm.historyMap[pm.currentMediaPlaybackStatus.Filename] = _ps
		pm.wsEventManager.SendEvent(events.PlaybackManagerProgressUpdated, _ps)
	}

	pm.refreshAnimeCollectionFunc()

	pm.eventMu.Unlock()
	return nil
}

// updateProgress updates the progress of the current video playback on AniList and MyAnimeList.
// This only returns an error if the progress update fails on AniList
//   - /!\ When this is called, the PlaybackState should have been pushed to the history
func (pm *PlaybackManager) updateProgress() (err error) {

	var mediaId int
	var epNum int
	var totalEpisodes int

	switch pm.currentPlaybackType {
	case LocalFilePlayback:
		//
		// Local File
		//
		if pm.currentLocalFileWrapperEntry.IsAbsent() || pm.currentLocalFile.IsAbsent() || pm.currentMediaListEntry.IsAbsent() {
			return errors.New("no video is being watched")
		}

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
		mediaId = pm.currentMediaListEntry.MustGet().GetMedia().GetID()
		epNum = pm.currentLocalFileWrapperEntry.MustGet().GetProgressNumber(pm.currentLocalFile.MustGet())
		totalEpisodes = pm.currentMediaListEntry.MustGet().GetMedia().GetTotalEpisodeCount()

	case StreamPlayback:
		//
		// Stream
		//
		// Last sanity check
		if pm.currentStreamEpisode.IsAbsent() || pm.currentStreamMedia.IsAbsent() {
			return errors.New("no video is being watched")
		}

		mediaId = pm.currentStreamMedia.MustGet().ID
		epNum = pm.currentStreamEpisode.MustGet().GetProgressNumber()
		totalEpisodes = pm.currentStreamMedia.MustGet().GetTotalEpisodeCount()
	default:
		return errors.New("unknown playback type")
	}

	if mediaId == 0 { // Sanity check
		return errors.New("media ID not found")
	}

	// Update the progress on AniList
	err = pm.platform.UpdateEntryProgress(
		mediaId,
		epNum,
		&totalEpisodes,
	)
	if err != nil {
		pm.Logger.Error().Err(err).Msg("playback manager: Error occurred while updating progress on AniList")
		return ErrProgressUpdateAnilist
	}

	pm.refreshAnimeCollectionFunc() // Refresh the AniList collection

	pm.Logger.Info().Msg("playback manager: Updated progress on AniList")

	go func() {
		defer util.HandlePanicThen(func() {})
		var malId *int
		// Get the MAL ID from the current media list entry or the current stream media
		if pm.currentMediaListEntry.IsPresent() {
			malId = pm.currentMediaListEntry.MustGet().GetMedia().GetIDMal()
		}
		if malId == nil && pm.currentStreamMedia.IsPresent() {
			malId = pm.currentStreamMedia.MustGet().GetIDMal()
		}
		if malId == nil {
			return
		}
		// Update the progress on MAL if an account is linked
		malInfo, _ := pm.Database.GetMalInfo()
		if malInfo == nil || malInfo.AccessToken == "" {
			return
		}
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
	}()

	return nil
}

func (pm *PlaybackManager) updateProgressOffline() (err error) {
	if pm.currentLocalFileWrapperEntry.IsAbsent() || pm.currentLocalFile.IsAbsent() || pm.currentMediaListEntry.IsAbsent() {
		return errors.New("no video is being watched")
	}

	mediaId := pm.currentMediaListEntry.MustGet().GetMedia().GetID()
	epNum := pm.currentLocalFileWrapperEntry.MustGet().GetProgressNumber(pm.currentLocalFile.MustGet())
	totalEpisodes := pm.currentMediaListEntry.MustGet().GetMedia().GetTotalEpisodeCount()

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
