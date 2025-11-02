package nakama

import (
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/nativeplayer"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"

	"sync/atomic"
)

type MediaControllerType string

const (
	MediaControllerTypePlaybackManager MediaControllerType = "playbackmanager"
	MediaControllerTypeNativePlayer    MediaControllerType = "nativeplayer"
)

// MediaController is an agnostic interface for controlling both the PlaybackManager (system player) and the NativePlayer (Denshi player).
// It converts native player events to playback manager events.
type MediaController struct {
	manager *Manager
	current MediaControllerType

	nativePlayerSubscribers *result.Map[string, *NativePlayerToPlaybackManagerSubscriber]
}

func NewMediaController(manager *Manager) *MediaController {
	return &MediaController{
		manager:                 manager,
		current:                 MediaControllerTypePlaybackManager,
		nativePlayerSubscribers: result.NewResultMap[string, *NativePlayerToPlaybackManagerSubscriber](),
	}
}

func (m *MediaController) SetType(t MediaControllerType) {
	m.current = t
}

func (m *MediaController) PullStatus() (*mediaplayer.PlaybackStatus, bool) {
	if m.current == MediaControllerTypePlaybackManager {
		return m.manager.playbackManager.PullStatus()
	}
	ps := m.manager.nativePlayer.GetPlaybackStatus()
	if ps == nil || ps.Duration == 0 {
		return nil, false
	}
	status := m.toPlaybackManagerStatus(ps)
	return &status, ps.Url != ""
}

func (m *MediaController) toPlaybackManagerStatus(ps *nativeplayer.PlaybackStatus) mediaplayer.PlaybackStatus {
	if ps == nil || ps.Duration == 0 {
		return mediaplayer.PlaybackStatus{}
	}
	return mediaplayer.PlaybackStatus{
		Duration:             int(ps.Duration * 1000), // convert to ms
		CompletionPercentage: ps.CurrentTime / ps.Duration,
		Playing:              !ps.Paused,
		Filename:             ps.Url,
		Path:                 ps.Url,
		Filepath:             ps.Url,
		CurrentTimeInSeconds: ps.CurrentTime,
		DurationInSeconds:    ps.Duration,
		PlaybackType:         "stream",
	}
}

func (m *MediaController) toPlaybackManagerState(info *nativeplayer.PlaybackInfo, status *nativeplayer.PlaybackStatus) playbackmanager.PlaybackState {
	if info == nil {
		return playbackmanager.PlaybackState{}
	}
	return playbackmanager.PlaybackState{
		EpisodeNumber:        info.Episode.EpisodeNumber,
		AniDbEpisode:         info.Episode.AniDBEpisode,
		MediaTitle:           info.Media.GetPreferredTitle(),
		MediaTotalEpisodes:   info.Media.GetTotalEpisodeCount(),
		Filename:             info.StreamUrl,
		CompletionPercentage: status.CurrentTime / status.Duration,
		CanPlayNext:          false,
		ProgressUpdated:      false,
		MediaId:              info.Media.GetID(),
	}
}

func (m *MediaController) Pause() {
	if m.current == MediaControllerTypePlaybackManager {
		_ = m.manager.playbackManager.Pause()
		return
	}
	m.manager.nativePlayer.Pause("")
}

func (m *MediaController) Resume() {
	if m.current == MediaControllerTypePlaybackManager {
		_ = m.manager.playbackManager.Resume()
		return
	}
	m.manager.nativePlayer.Resume("")
}

func (m *MediaController) Cancel() {
	if m.current == MediaControllerTypePlaybackManager {
		_ = m.manager.playbackManager.Cancel()
		return
	}
	m.manager.nativePlayer.Stop()
}

func (m *MediaController) SeekTo(time float64) {
	if m.current == MediaControllerTypePlaybackManager {
		_ = m.manager.playbackManager.SeekTo(time)
		return
	}
	m.manager.nativePlayer.SeekTo("", time)
}

type NativePlayerToPlaybackManagerSubscriber struct {
	subscriber             *playbackmanager.PlaybackStatusSubscriber
	nativePlayerSubscriber *nativeplayer.Subscriber
	closeOnce              sync.Once
	closeCh                chan struct{}
}

// SubscribeToPlaybackStatus subscribes to the playback status of the media controller
// It will convert native player events to playback manager events
func (m *MediaController) SubscribeToPlaybackStatus(id string) *playbackmanager.PlaybackStatusSubscriber {
	defer util.HandlePanicInModuleThen("nakama/SubscribeToPlaybackStatus", func() {})

	// Playback manager
	if m.current == MediaControllerTypePlaybackManager {
		return m.manager.playbackManager.SubscribeToPlaybackStatus(id)
	}

	// Native player
	nativePlayerSubscriber := m.manager.nativePlayer.Subscribe(id)
	sub := &NativePlayerToPlaybackManagerSubscriber{
		subscriber: &playbackmanager.PlaybackStatusSubscriber{
			EventCh:  make(chan playbackmanager.PlaybackEvent, 100),
			Canceled: atomic.Bool{},
		},
		nativePlayerSubscriber: nativePlayerSubscriber,
		closeOnce:              sync.Once{},
		closeCh:                make(chan struct{}),
	}
	m.nativePlayerSubscribers.Set(id, sub)

	// Convert native player events to playback manager events
	go func() {
		defer util.HandlePanicInModuleThen("nakama/nativePlayerSubscriber", func() {})

		for {
			select {
			case event := <-nativePlayerSubscriber.Events():
				if sub.subscriber.Canceled.Load() {
					return
				}
				nativePlayerStatus := m.manager.nativePlayer.GetPlaybackStatus()
				nativePlayerInfo, _ := m.manager.nativePlayer.GetPlaybackInfo()
				status := m.toPlaybackManagerStatus(nativePlayerStatus)
				state := m.toPlaybackManagerState(nativePlayerInfo, nativePlayerStatus)
				switch event.(type) {
				case *nativeplayer.VideoLoadedMetadataEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
					sub.subscriber.EventCh <- &playbackmanager.StreamStartedEvent{
						Filename: status.Filename,
						Filepath: status.Filepath,
					}
				case *nativeplayer.VideoCompletedEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
					sub.subscriber.EventCh <- &playbackmanager.StreamCompletedEvent{
						Filename: status.Filename,
					}
				case *nativeplayer.VideoTerminatedEvent:
					sub.subscriber.EventCh <- &playbackmanager.StreamStoppedEvent{
						Reason: "Player closed",
					}
				case *nativeplayer.VideoPausedEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
				case *nativeplayer.VideoResumedEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
				case *nativeplayer.VideoSeekedEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
				case *nativeplayer.VideoStatusEvent:
					sub.subscriber.EventCh <- &playbackmanager.PlaybackStatusChangedEvent{
						Status: status,
						State:  state,
					}
				}
			case <-sub.closeCh:
				// Terminate the goroutine when the subscriber is closed
				return
			}
		}
	}()

	return sub.subscriber
}

func (m *MediaController) UnsubscribeFromPlaybackStatus(id string) {
	defer util.HandlePanicInModuleThen("nakama/UnsubscribeFromPlaybackStatus", func() {})

	// Playback manager
	if m.current == MediaControllerTypePlaybackManager {
		m.manager.playbackManager.UnsubscribeFromPlaybackStatus(id)
		return
	}

	// Native player
	subscriber, ok := m.nativePlayerSubscribers.Get(id)
	if !ok {
		return
	}
	subscriber.closeOnce.Do(func() {
		close(subscriber.closeCh)
	})
	m.manager.nativePlayer.Unsubscribe(id)
	m.nativePlayerSubscribers.Delete(id)
}
