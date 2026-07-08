package nakama

import (
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediacore"
	"seanime/internal/player"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"

	"go.uber.org/atomic"
)

// WatchPartyGenericPlayerType is the type of player that the WatchPartyGenericPlayer is actively using.
type WatchPartyGenericPlayerType string

const (
	WatchPartyPlaybackManager WatchPartyGenericPlayerType = "playbackmanager"
	WatchPartyVideoCore       WatchPartyGenericPlayerType = "videocore"
	WatchPartyMpvCore         WatchPartyGenericPlayerType = "mpvcore"
)

// WatchPartyGenericPlayer is a player-agnostic interface for controlling
// both the playbackmanager.PlaybackManager (system player) and the mediacore.Coordinator.
type WatchPartyGenericPlayer struct {
	manager       *Manager
	current       atomic.String
	defaultPlayer atomic.String
	subscribers   *result.Map[string, *WatchPartyPlaybackSubscriber]
}

func NewWatchPartyGenericPlayer(manager *Manager) *WatchPartyGenericPlayer {
	ret := &WatchPartyGenericPlayer{
		manager:     manager,
		subscribers: result.NewMap[string, *WatchPartyPlaybackSubscriber](),
	}
	ret.defaultPlayer.Store(string(WatchPartyPlaybackManager))
	ret.current.Store("")
	return ret
}

// SetType sets the current player type.
func (m *WatchPartyGenericPlayer) SetType(t WatchPartyGenericPlayerType) {
	m.current.Store(string(t))
}

// SetDefaultType sets the default player type.
// It is called periodically in NewManager to update the default player type (Video Playback setting) used by the client.
func (m *WatchPartyGenericPlayer) SetDefaultType(t WatchPartyGenericPlayerType) {
	m.defaultPlayer.Store(string(t))
}

func (m *WatchPartyGenericPlayer) getCurrentType() WatchPartyGenericPlayerType {
	if m.current.Load() != "" {
		return WatchPartyGenericPlayerType(m.current.Load())
	}
	return WatchPartyGenericPlayerType(m.defaultPlayer.Load())
}

func (m *WatchPartyGenericPlayer) isPlaybackManager() bool {
	return m.getCurrentType() == WatchPartyPlaybackManager
}

func (m *WatchPartyGenericPlayer) isVideoCore() bool {
	return m.getCurrentType() == WatchPartyVideoCore
}

func (m *WatchPartyGenericPlayer) isMpvCore() bool {
	return m.getCurrentType() == WatchPartyMpvCore
}

func (m *WatchPartyGenericPlayer) Reset() {
	m.current.Store("")
}

// PullStatus returns the current playback status of whatever media player is currently in use.
func (m *WatchPartyGenericPlayer) PullStatus() (*WatchPartyPlaybackStatus, bool) {
	// Playback manager
	if m.isPlaybackManager() {
		status, ok := m.manager.playbackManager.PullStatus()
		if !ok {
			return nil, false
		}

		return &WatchPartyPlaybackStatus{
			Paused:      !status.Playing,
			CurrentTime: status.CurrentTimeInSeconds,
			Duration:    status.DurationInSeconds,
		}, true
	}

	// Pull from Mediacore Coordinator
	status, ok := m.manager.mediacoreCoordinator.PullStatus()
	if !ok {
		return nil, false
	}

	return &WatchPartyPlaybackStatus{
		Paused:      status.Paused,
		CurrentTime: status.CurrentTime,
		Duration:    status.Duration,
	}, true
}

func (m *WatchPartyGenericPlayer) Pause() {
	if m.isPlaybackManager() {
		_ = m.manager.playbackManager.Pause()
		return
	}
	if session, ok := m.manager.mediacoreCoordinator.GetActiveSession(); ok {
		_ = m.manager.mediacoreCoordinator.Execute(session, player.Command{Type: player.CommandPause})
	}
}

func (m *WatchPartyGenericPlayer) Resume() {
	if m.isPlaybackManager() {
		_ = m.manager.playbackManager.Resume()
		return
	}
	if session, ok := m.manager.mediacoreCoordinator.GetActiveSession(); ok {
		_ = m.manager.mediacoreCoordinator.Execute(session, player.Command{Type: player.CommandResume})
	}
}

func (m *WatchPartyGenericPlayer) Cancel() {
	defer m.Reset()
	if m.isPlaybackManager() {
		_ = m.manager.playbackManager.Cancel()
		return
	}
	if session, ok := m.manager.mediacoreCoordinator.GetActiveSession(); ok {
		m.manager.mediacoreCoordinator.Terminate(session)
	}
}

func (m *WatchPartyGenericPlayer) SeekTo(time float64) {
	if m.isPlaybackManager() {
		_ = m.manager.playbackManager.SeekTo(time)
		return
	}
	if session, ok := m.manager.mediacoreCoordinator.GetActiveSession(); ok {
		_ = m.manager.mediacoreCoordinator.Execute(session, player.Command{Type: player.CommandSeekTo, Payload: time})
	}
}

type (
	WatchPartyPlaybackEvent interface {
		IsWatchPartyPlaybackEvent() bool
		Type() string
	}
	WatchPartyPlaybackSubscriber struct {
		id                        string
		EventCh                   chan WatchPartyPlaybackEvent
		closeOnce                 sync.Once
		playbackManagerSubscriber *playbackmanager.PlaybackStatusSubscriber
		mediacoreSubscriber       *mediacore.Subscriber
	}

	WatchPartyPlayerBaseEvent struct{}

	WatchPartyPlayerVideoStarted struct {
		WatchPartyPlayerBaseEvent
		StreamType WatchPartyStreamType
	}
	WatchPartyPlayerVideoStatus struct {
		WatchPartyPlayerBaseEvent
		Status   *WatchPartyPlaybackStatus
		State    *WatchPartyPlaybackState
		Filename string
		Filepath string
	}
	WatchPartyPlayerVideoEnded struct {
		WatchPartyPlayerBaseEvent
	}
)

func (e *WatchPartyPlayerBaseEvent) IsWatchPartyPlaybackEvent() bool {
	return true
}

func (e *WatchPartyPlayerVideoStarted) Type() string {
	return "video-started"
}

func (e *WatchPartyPlayerVideoStatus) Type() string {
	return "video-status"
}

func (e *WatchPartyPlayerVideoEnded) Type() string {
	return "video-ended"
}

func (m *WatchPartyGenericPlayer) Unsubscribe(id string) {
	defer util.HandlePanicInModuleThen("nakama/UnsubscribeToPlaybackStatus", func() {})

	if subscriber, ok := m.subscribers.Pop(id); ok {
		// Playback manager
		if subscriber.playbackManagerSubscriber != nil {
			m.manager.playbackManager.UnsubscribeFromPlaybackStatus(subscriber.id)
		}
		// Mediacore
		if subscriber.mediacoreSubscriber != nil {
			m.manager.mediacoreCoordinator.Unsubscribe(subscriber.id)
		}
		subscriber.closeOnce.Do(func() {
			close(subscriber.EventCh)
		})
	}
}

func fromPlaybackManagerStatus(event playbackmanager.PlaybackStatusChangedEvent) *WatchPartyPlayerVideoStatus {
	streamType := WatchPartyStreamTypeFile
	if strings.Contains(event.Status.Filepath, "type=file") {
		streamType = WatchPartyStreamTypeFile
	} else if strings.Contains(event.Status.Filepath, "/api/v1/torrentstream") {
		streamType = WatchPartyStreamTypeTorrent
	} else {
		streamType = WatchPartyStreamTypeDebrid
	}

	return &WatchPartyPlayerVideoStatus{
		Status: &WatchPartyPlaybackStatus{
			Paused:      !event.Status.Playing,
			CurrentTime: event.Status.CurrentTimeInSeconds,
			Duration:    event.Status.DurationInSeconds,
		},
		State: &WatchPartyPlaybackState{
			MediaId:       event.State.MediaId,
			EpisodeNumber: event.State.EpisodeNumber,
			AniDBEpisode:  event.State.AniDbEpisode,
			StreamType:    streamType,
		},
		Filename: event.Status.Filename,
		Filepath: event.Status.Filepath,
	}
}

func fromMediacoreStatus(event *player.StatusEvent, state *player.PlaybackState) *WatchPartyPlayerVideoStatus {
	streamType := WatchPartyStreamTypeFile
	filename := ""
	filepath := ""
	if state.PlaybackInfo != nil {
		switch state.PlaybackInfo.PlaybackType {
		case player.PlaybackTypeLocalFile:
			streamType = WatchPartyStreamTypeFile
			if state.PlaybackInfo.LocalFile != nil {
				filename = state.PlaybackInfo.LocalFile.Name
				filepath = state.PlaybackInfo.LocalFile.Path
			}
		case player.PlaybackTypeTorrent:
			streamType = WatchPartyStreamTypeTorrent
		case player.PlaybackTypeDebrid:
			streamType = WatchPartyStreamTypeDebrid
		case player.PlaybackTypeOnlinestream:
			streamType = WatchPartyStreamTypeOnlinestream
		case player.PlaybackTypeNakama:
			streamType = WatchPartyStreamTypeFile
			filepath = state.PlaybackInfo.StreamPath
		}
	}

	ret := &WatchPartyPlayerVideoStatus{
		Status: &WatchPartyPlaybackStatus{
			Paused:      event.Paused,
			CurrentTime: event.CurrentTime,
			Duration:    event.Duration,
		},
		State:    &WatchPartyPlaybackState{StreamType: streamType},
		Filename: filename,
		Filepath: filepath,
	}
	if state.PlaybackInfo != nil {
		if state.PlaybackInfo.Media != nil {
			ret.State.MediaId = state.PlaybackInfo.Media.GetID()
		}
		if state.PlaybackInfo.Episode != nil {
			ret.State.EpisodeNumber = state.PlaybackInfo.Episode.EpisodeNumber
			ret.State.AniDBEpisode = state.PlaybackInfo.Episode.AniDBEpisode
		}
	}
	return ret
}

// Subscribe is a generic subscriber to PlaybackManager and Mediacore Coordinator.
func (m *WatchPartyGenericPlayer) Subscribe(id string) *WatchPartyPlaybackSubscriber {
	defer util.HandlePanicInModuleThen("nakama/Subscribe", func() {})
	subscriber := &WatchPartyPlaybackSubscriber{
		id:      id,
		EventCh: make(chan WatchPartyPlaybackEvent, 100),
	}

	m.subscribers.Set(id, subscriber)

	playbackManagerSubscriber := m.manager.playbackManager.SubscribeToPlaybackStatus(id)
	subscriber.playbackManagerSubscriber = playbackManagerSubscriber

	go func() {
		defer util.HandlePanicInModuleThen("nakama/Subscribe", func() {})
		for e := range playbackManagerSubscriber.EventCh {
			switch e.(type) {
			case playbackmanager.StreamStartedEvent, playbackmanager.VideoStartedEvent:
				m.SetType(WatchPartyPlaybackManager)
			}
			if !m.isPlaybackManager() {
				continue
			}
			switch event := e.(type) {
			case playbackmanager.StreamStartedEvent:
				streamType := WatchPartyStreamTypeFile
				if strings.Contains(event.Filepath, "type=file") {
					streamType = WatchPartyStreamTypeFile
				} else if strings.Contains(event.Filepath, "/api/v1/torrentstream") {
					streamType = WatchPartyStreamTypeTorrent
				} else {
					streamType = WatchPartyStreamTypeDebrid
				}

				subscriber.EventCh <- &WatchPartyPlayerVideoStarted{
					StreamType: streamType,
				}
			case playbackmanager.VideoStartedEvent:
				subscriber.EventCh <- &WatchPartyPlayerVideoStarted{
					StreamType: WatchPartyStreamTypeFile,
				}
			case playbackmanager.PlaybackStatusChangedEvent:
				subscriber.EventCh <- fromPlaybackManagerStatus(event)
			case playbackmanager.StreamStoppedEvent, playbackmanager.VideoStoppedEvent:
				subscriber.EventCh <- &WatchPartyPlayerVideoEnded{}
			}
		}
	}()

	mediacoreSubscriber := m.manager.mediacoreCoordinator.Subscribe(id)
	subscriber.mediacoreSubscriber = mediacoreSubscriber

	go func() {
		defer util.HandlePanicInModuleThen("nakama/Subscribe", func() {})
		for e := range mediacoreSubscriber.Events() {
			target := e.GetSessionKey().Target
			switch e.(type) {
			case *player.PlaybackLoadedEvent, *player.LoadedMetadataEvent, *player.StatusEvent:
				if target == player.TargetVideoCore {
					m.SetType(WatchPartyVideoCore)
				} else if target == player.TargetMpvCore {
					m.SetType(WatchPartyMpvCore)
				}
			}
			if m.isPlaybackManager() {
				continue
			}
			if target == player.TargetVideoCore && !m.isVideoCore() {
				continue
			}
			if target == player.TargetMpvCore && !m.isMpvCore() {
				continue
			}

			switch event := e.(type) {
			case *player.LoadedMetadataEvent:
				streamType := WatchPartyStreamTypeFile
				playbackState, ok := m.manager.mediacoreCoordinator.GetActivePlaybackState()
				if ok && playbackState.PlaybackInfo != nil {
					switch playbackState.PlaybackInfo.PlaybackType {
					case player.PlaybackTypeLocalFile:
						streamType = WatchPartyStreamTypeFile
					case player.PlaybackTypeTorrent:
						streamType = WatchPartyStreamTypeTorrent
					case player.PlaybackTypeDebrid:
						streamType = WatchPartyStreamTypeDebrid
					case player.PlaybackTypeOnlinestream:
						streamType = WatchPartyStreamTypeOnlinestream
					}
				}

				subscriber.EventCh <- &WatchPartyPlayerVideoStarted{
					StreamType: streamType,
				}
			case *player.StatusEvent:
				state, ok := m.manager.mediacoreCoordinator.GetActivePlaybackState()
				if !ok {
					continue
				}
				subscriber.EventCh <- fromMediacoreStatus(event, &state)
			case *player.EndedEvent, *player.TerminatedEvent:
				subscriber.EventCh <- &WatchPartyPlayerVideoEnded{}
			}
		}
	}()

	return subscriber
}
