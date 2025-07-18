package playbackmanager

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/util"
	"time"

	"github.com/samber/mo"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manual progress tracking
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ManualTrackingState struct {
	EpisodeNumber   int
	MediaId         int
	CurrentProgress int
	TotalEpisodes   int
}

type StartManualProgressTrackingOptions struct {
	ClientId      string
	MediaId       int
	EpisodeNumber int
}

func (pm *PlaybackManager) CancelManualProgressTracking() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.manualTrackingCtxCancel != nil {
		pm.manualTrackingCtxCancel()
		pm.currentManualTrackingState = mo.None[*ManualTrackingState]()
	}
}

func (pm *PlaybackManager) StartManualProgressTracking(opts *StartManualProgressTrackingOptions) (err error) {
	defer util.HandlePanicInModuleWithError("library/playbackmanager/StartManualProgressTracking", &err)

	ctx := context.Background()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Logger.Trace().Msg("playback manager: Starting manual progress tracking")

	// Cancel manual tracking if active
	if pm.manualTrackingCtxCancel != nil {
		pm.Logger.Trace().Msg("playback manager: Cancelling previous manual tracking context")
		pm.manualTrackingCtxCancel()
		pm.manualTrackingWg.Wait()
	}

	// Get the media
	// - Find the media in the collection
	animeCollection, err := pm.platform.GetAnimeCollection(ctx, false)
	if err != nil {
		return err
	}

	var media *anilist.BaseAnime
	var currentProgress int
	var totalEpisodes int

	listEntry, found := animeCollection.GetListEntryFromAnimeId(opts.MediaId)

	if found {
		media = listEntry.Media
	} else {
		// Fetch the media from AniList
		media, err = pm.platform.GetAnime(ctx, opts.MediaId)
	}
	if media == nil {
		pm.Logger.Error().Msg("playback manager: Media not found for manual tracking")
		return fmt.Errorf("media not found")
	}

	currentProgress = 0
	if listEntry != nil && listEntry.GetProgress() != nil {
		currentProgress = *listEntry.GetProgress()
	}
	totalEpisodes = media.GetTotalEpisodeCount()

	// Set the current playback type (for progress update later on)
	pm.currentPlaybackType = ManualTrackingPlayback

	// Set the manual tracking state (for progress update later on)
	pm.currentManualTrackingState = mo.Some(&ManualTrackingState{
		EpisodeNumber:   opts.EpisodeNumber,
		MediaId:         opts.MediaId,
		CurrentProgress: currentProgress,
		TotalEpisodes:   totalEpisodes,
	})

	pm.Logger.Trace().
		Int("episode_number", opts.EpisodeNumber).
		Int("mediaId", opts.MediaId).
		Int("currentProgress", currentProgress).
		Int("totalEpisodes", totalEpisodes).
		Msg("playback manager: Starting manual progress tracking")

	// Start sending the manual tracking events
	pm.manualTrackingWg.Add(1)
	go func() {
		defer pm.manualTrackingWg.Done()
		// Create a new context
		pm.manualTrackingCtx, pm.manualTrackingCtxCancel = context.WithCancel(context.Background())
		defer func() {
			if pm.manualTrackingCtxCancel != nil {
				pm.manualTrackingCtxCancel()
			}
		}()

		for {
			select {
			case <-pm.manualTrackingCtx.Done():
				pm.Logger.Debug().Msg("playback manager: Manual progress tracking canceled")
				pm.wsEventManager.SendEvent(events.PlaybackManagerManualTrackingStopped, nil)
				return
			default:
				ps := playbackStatePool.Get().(*PlaybackState)
				ps.EpisodeNumber = opts.EpisodeNumber
				ps.MediaTitle = *media.GetTitle().GetUserPreferred()
				ps.MediaTotalEpisodes = totalEpisodes
				ps.Filename = ""
				ps.CompletionPercentage = 0
				ps.CanPlayNext = false
				ps.ProgressUpdated = false
				ps.MediaId = opts.MediaId
				pm.wsEventManager.SendEvent(events.PlaybackManagerManualTrackingPlaybackState, ps)
				playbackStatePool.Put(ps)
				// Continuously send the progress to the client
				time.Sleep(3 * time.Second)
			}
		}
	}()

	return nil
}
