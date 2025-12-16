package torrentstream

import (
	"context"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/videocore"
)

type (
	playback struct {
		mediaPlayerCtxCancelFunc context.CancelFunc
		// Stores the video duration returned by the media player
		// When this is greater than 0, the video is considered to be playing
		currentVideoDuration int
	}
)

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
					if settings, ok := r.settings.Get(); ok {
						r.shouldPreloadStream.Store(settings.PreloadNextStream)
					}
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

func (r *Repository) listenToNativePlayerEvents() {
	r.nativePlayer.VideoCore().Unsubscribe("torrentstream")
	r.logger.Trace().Msg("torrentstream: Subscribing to video core events")
	videoCoreSubscriber := r.nativePlayer.VideoCore().Subscribe("torrentstream")

	go func(sub *videocore.Subscriber) {
		defer func() {
			r.logger.Trace().Msg("torrentstream: Stopping video core listener")
		}()
		for e := range sub.Events() {
			// get the player type from the event instead of the instance
			if e.GetPlayerType() != videocore.NativePlayer {
				continue
			}

			switch event := e.(type) {
			case *videocore.VideoLoadedEvent:
				r.logger.Debug().Msg("torrentstream: Native player loaded event received")
				r.playback.currentVideoDuration = 0
				if settings, ok := r.settings.Get(); ok {
					r.shouldPreloadStream.Store(settings.PreloadNextStream)
				}
			case *videocore.VideoLoadedMetadataEvent:
				go func() {
					if r.client.currentFile.IsPresent() && r.playback.currentVideoDuration == 0 {
						// If the stored video duration is 0 but the media player status shows a duration that is not 0
						// we know that the video has been loaded and is playing
						if r.playback.currentVideoDuration == 0 && event.Duration > 0 {
							// The media player has started playing the video
							r.logger.Debug().Msg("torrentstream: Media player started playing the video, sending event")
							r.sendStateEvent(eventTorrentStartedPlaying)
							// Update the stored video duration
							r.playback.currentVideoDuration = int(event.Duration)
						}
					}
				}()
			case *videocore.VideoStatusEvent:
				if event.CurrentTime/event.Duration >= 0.5 && r.shouldPreloadStream.Load() {
					r.shouldPreloadStream.Store(false)
					r.sendStateEvent(eventPreloadNextStream)
				}
			case *videocore.VideoTerminatedEvent:
				r.logger.Debug().Msg("torrentstream: Native player terminated event received")
				r.playback.currentVideoDuration = 0
				// Only handle the event if we actually have a current torrent to avoid unnecessary cleanup
				if r.client.currentTorrent.IsPresent() {
					go func() {
						defer func() {
							if rec := recover(); rec != nil {
								r.logger.Error().Msg("torrentstream: Recovered from panic in VideoTerminatedEvent handler")
							}
						}()
						r.logger.Debug().Msg("torrentstream: Stopping stream due to native player termination")
						// Stop the stream
						_ = r.StopStream()
					}()
				}
			}

		}
	}(videoCoreSubscriber)
}
