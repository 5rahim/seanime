package torrentstream

import (
	"context"
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
			case _ = <-r.mediaPlayerRepositorySubscriber.TrackingStartedCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.TrackingRetryCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.VideoCompletedCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.TrackingStoppedCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.PlaybackStatusCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.StreamingTrackingStartedCh:
				// Reset the current video duration, as the video has stopped
				// DEVNOTE: This is changed in client.go as well when the duration is updated over 0
				r.playback.currentVideoDuration = 0
			case _ = <-r.mediaPlayerRepositorySubscriber.StreamingVideoCompletedCh:
			case _ = <-r.mediaPlayerRepositorySubscriber.StreamingTrackingStoppedCh:
				if r.client.currentTorrent.IsPresent() {
					go func() {
						defer func() {
							if r := recover(); r != nil {
							}
						}()
						r.logger.Debug().Msg("torrentstream: Media player stopped event received")
						// Stop the stream
						_ = r.StopStream()
						// Stop the server
						//r.serverManager.stopServer()
						//// Signal to client.go that the media player has stopped
						//close(r.client.stopCh)
					}()
				}
			case status := <-r.mediaPlayerRepositorySubscriber.StreamingPlaybackStatusCh:
				go func() {
					if status != nil && r.client.currentTorrent.IsPresent() {
						r.client.mediaPlayerPlaybackStatusCh <- status
					}
				}()
			case _ = <-r.mediaPlayerRepositorySubscriber.StreamingTrackingRetryCh:
				// ignored
			}
		}
	}(ctx)
}
