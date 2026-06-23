package torrentstream

import (
	"context"
	"seanime/internal/mediacore"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/player"
)

type (
	playback struct {
		mediaPlayerCtxCancelFunc context.CancelFunc
		// Stores the video duration returned by the media player
		// When this is greater than 0, the video is considered to be playing
		currentVideoDuration int
	}
)

func (r *Repository) resetPreloadFlag() {
	settings, ok := r.settings.Get()
	if !ok {
		r.shouldPreloadStream.Store(false)
		return
	}
	r.shouldPreloadStream.Store(settings.PreloadNextStream)
}

func (r *Repository) listenToMediaPlayerEvents() {
	r.mediaPlayerRepositorySubscriber = r.mediaPlayerRepository.Subscribe("torrentstream")

	if r.playback.mediaPlayerCtxCancelFunc != nil {
		r.playback.mediaPlayerCtxCancelFunc()
	}

	var ctx context.Context
	ctx, r.playback.mediaPlayerCtxCancelFunc = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				r.logger.Debug().Msg("torrentstream: Media player context cancelled")
				return
			case event := <-r.mediaPlayerRepositorySubscriber.EventCh:
				switch e := event.(type) {
				case mediaplayer.StreamingTrackingStartedEvent:
					// Reset the current video duration, as the video has stopped
					// DEVNOTE: This is changed in client.go as well when the duration is updated over 0
					r.playback.currentVideoDuration = 0
					r.resetPreloadFlag()
				case mediaplayer.StreamingVideoCompletedEvent:
				case mediaplayer.StreamingTrackingStoppedEvent:
					if r.client.currentTorrent.IsPresent() {
						go func() {
							defer func() {
								if r := recover(); r != nil {
								}
							}()
							r.logger.Debug().Msg("torrentstream: Media player stopped event received")
							// Stop the stream
							_ = r.StopStream()
						}()
					}
				case mediaplayer.StreamingPlaybackStatusEvent:
					go func() {
						if e.Status != nil && r.client.currentTorrent.IsPresent() {
							r.client.mediaPlayerPlaybackStatusCh <- e.Status
						}
						if r.shouldPreloadStream.Load() && e.Status.CompletionPercentage >= 0.5 {
							r.shouldPreloadStream.Store(false)
							r.sendStateEvent(eventPreloadNextStream)
						}
					}()
				}
			}
		}
	}(ctx)
}

func (r *Repository) listenToMediacoreEvents() {
	if r.mediacoreCoordinator == nil {
		return
	}
	r.mediacoreCoordinator.Unsubscribe("torrentstream")
	r.logger.Trace().Msg("torrentstream: Subscribing to mediacore events")
	subscriber := r.mediacoreCoordinator.Subscribe("torrentstream")

	go func(sub *mediacore.Subscriber) {
		defer func() {
			r.logger.Trace().Msg("torrentstream: Stopping mediacore listener")
		}()
		for e := range sub.Events() {
			key := e.GetSessionKey()

			playbackState, ok := r.mediacoreCoordinator.GetActivePlaybackState()
			if !ok || playbackState.PlaybackInfo.PlaybackType != player.PlaybackTypeTorrent {
				continue
			}

			playbackID, clientID, ok := r.directStreamManager.GetCurrentPlaybackIdentity()
			if !ok {
				continue
			}
			if clientID != "" && key.ClientID != "" && key.ClientID != clientID {
				continue
			}
			if playbackID != "" && key.PlaybackID != "" && key.PlaybackID != playbackID {
				continue
			}

			switch event := e.(type) {
			case *player.PlaybackLoadedEvent:
				r.logger.Debug().Msg("torrentstream: PlaybackLoaded event received")
				r.playback.currentVideoDuration = 0
				r.resetPreloadFlag()
			case *player.LoadedMetadataEvent:
				go func() {
					if r.client.currentFile.IsPresent() && r.playback.currentVideoDuration == 0 {
						if event.Duration > 0 {
							r.logger.Debug().Msg("torrentstream: Media player started playing the video, sending event")
							r.sendStateEvent(eventTorrentStartedPlaying)
							r.playback.currentVideoDuration = int(event.Duration)
							r.resetPreloadFlag()
						}
					}
				}()
			case *player.StatusEvent:
				if event.Duration > 0 && event.CurrentTime/event.Duration >= 0.5 && r.shouldPreloadStream.Load() {
					r.shouldPreloadStream.Store(false)
					r.sendStateEvent(eventPreloadNextStream)
				}
			case *player.TerminatedEvent:
				r.logger.Debug().Msg("torrentstream: Playback terminated event received")
				r.playback.currentVideoDuration = 0
			}
		}
	}(subscriber)
}
