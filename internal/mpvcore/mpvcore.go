package mpvcore

import (
	"encoding/json"
	"errors"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/models"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type MpvCore struct {
	wsEventManager              events.WSEventManagerInterface
	clientPlayerEventSubscriber *events.ClientEventSubscriber

	continuityManager          *continuity.Manager
	metadataProviderRef        *util.Ref[metadata_provider.Provider]
	discordPresence            *discordrpc_presence.Presence
	platformRef                *util.Ref[platform.Platform]
	refreshAnimeCollectionFunc func()
	isOfflineRef               *util.Ref[bool]

	pendingMu    sync.RWMutex
	pendingState *PlaybackState
	stateMu      sync.RWMutex
	state        *PlaybackState
	statusMu     sync.RWMutex
	status       *PlaybackStatus

	subscribers *result.Map[string, *Subscriber]
	eventBus    chan VideoEvent
	stopCh      chan struct{}
	startOnce   sync.Once

	settingsMu sync.RWMutex
	settings   *models.Settings
	logger     *zerolog.Logger
	inSight    *InSight
}

type Subscriber struct {
	id        string
	eventCh   chan VideoEvent
	closed    atomic.Bool
	closeOnce sync.Once
}

type NewMpvCoreOptions struct {
	WsEventManager             events.WSEventManagerInterface
	Logger                     *zerolog.Logger
	MetadataProviderRef        *util.Ref[metadata_provider.Provider]
	ContinuityManager          *continuity.Manager
	DiscordPresence            *discordrpc_presence.Presence
	PlatformRef                *util.Ref[platform.Platform]
	RefreshAnimeCollectionFunc func()
	IsOfflineRef               *util.Ref[bool]
}

func New(opts NewMpvCoreOptions) *MpvCore {
	mc := &MpvCore{
		wsEventManager:              opts.WsEventManager,
		clientPlayerEventSubscriber: opts.WsEventManager.SubscribeToClientMpvCoreEvents("mpvcore"),
		continuityManager:           opts.ContinuityManager,
		metadataProviderRef:         opts.MetadataProviderRef,
		discordPresence:             opts.DiscordPresence,
		platformRef:                 opts.PlatformRef,
		refreshAnimeCollectionFunc:  opts.RefreshAnimeCollectionFunc,
		isOfflineRef:                opts.IsOfflineRef,
		subscribers:                 result.NewMap[string, *Subscriber](),
		eventBus:                    make(chan VideoEvent, 100),
		stopCh:                      make(chan struct{}),
		logger:                      opts.Logger,
	}
	mc.inSight = NewInSight(opts.Logger, mc)
	mc.Start()
	mc.inSight.Start()
	return mc
}

func (mc *MpvCore) Start() {
	mc.startOnce.Do(func() {
		go mc.listenToClientEvents()
		go func() {
			for {
				select {
				case <-mc.stopCh:
					return
				case event := <-mc.eventBus:
					mc.dispatch(event)
				}
			}
		}()
		//mc.setupEffects()
	})
}

func (mc *MpvCore) Shutdown() {
	select {
	case <-mc.stopCh:
	default:
		close(mc.stopCh)
	}
}

func (mc *MpvCore) SetSettings(settings *models.Settings) {
	mc.settingsMu.Lock()
	mc.settings = settings
	mc.settingsMu.Unlock()
}

func (mc *MpvCore) InSight() *InSight { return mc.inSight }

func (mc *MpvCore) Subscribe(id string) *Subscriber {
	sub := &Subscriber{id: id, eventCh: make(chan VideoEvent, 100)}
	if previous, ok := mc.subscribers.Pop(id); ok {
		previous.closed.Store(true)
		previous.closeOnce.Do(func() { close(previous.eventCh) })
	}
	mc.subscribers.Set(id, sub)
	return sub
}

func (mc *MpvCore) Unsubscribe(id string) {
	if sub, ok := mc.subscribers.Pop(id); ok {
		sub.closed.Store(true)
		sub.closeOnce.Do(func() { close(sub.eventCh) })
	}
}

func (s *Subscriber) Events() <-chan VideoEvent { return s.eventCh }
func (s *Subscriber) GetID() string             { return s.id }

func (mc *MpvCore) RegisterEventCallback(callback func(VideoEvent) bool) func() {
	id := uuid.NewString()
	sub := mc.Subscribe(id)
	var once sync.Once
	cancel := func() { once.Do(func() { mc.Unsubscribe(id) }) }
	go func() {
		defer cancel()
		for event := range sub.Events() {
			if !callback(event) {
				return
			}
		}
	}()
	return cancel
}

func (mc *MpvCore) dispatch(event VideoEvent) {
	mc.subscribers.Range(func(id string, sub *Subscriber) bool {
		if sub.closed.Load() {
			return true
		}
		if event.IsCritical() {
			select {
			case sub.eventCh <- event:
			case <-time.After(time.Second):
				mc.logger.Warn().Str("subscriber", id).Msg("mpvcore: subscriber blocked a critical event")
			}
		} else {
			select {
			case sub.eventCh <- event:
			default:
			}
		}
		return true
	})
}

func (mc *MpvCore) PushEvent(event VideoEvent) {
	state, ok := mc.GetPlaybackState()
	if !ok {
		return
	}
	event.identify(state.PlaybackInfo.ID, state.ClientID, state.PlaybackInfo.PlaybackType)
	select {
	case mc.eventBus <- event:
	default:
		mc.logger.Warn().Msg("mpvcore: event bus full")
	}
}

func (mc *MpvCore) sendTo(clientID string, event ServerEvent, payload interface{}, noLog ...bool) {
	envelope := struct {
		Type    ServerEvent `json:"type"`
		Payload interface{} `json:"payload"`
	}{Type: event, Payload: payload}
	if clientID != "" {
		mc.wsEventManager.SendEventTo(clientID, string(events.MpvCoreEventType), envelope, noLog...)
		return
	}
	mc.wsEventManager.SendEvent(string(events.MpvCoreEventType), envelope)
}

func (mc *MpvCore) OpenAndAwait(clientID, loadingState string) {
	mc.sendTo(clientID, ServerEventOpenAndAwait, loadingState)
}

func (mc *MpvCore) AbortOpen(clientID, reason string) {
	mc.sendTo(clientID, ServerEventAbortOpen, reason)
}

func (mc *MpvCore) Watch(clientID string, info *PlaybackInfo) {
	if info == nil {
		return
	}
	state := &PlaybackState{ClientID: clientID, PlaybackInfo: info}
	mc.pendingMu.Lock()
	mc.pendingState = state
	mc.pendingMu.Unlock()
	mc.sendTo(clientID, ServerEventWatch, info, true)
}

func (mc *MpvCore) Error(clientID string, err error) {
	if err == nil {
		return
	}
	mc.sendTo(clientID, ServerEventStreamError, errorPayload{Error: err.Error()})
	mc.clearPlayback()
}

func (mc *MpvCore) Stop() { mc.Terminate() }

func (mc *MpvCore) withStateEvent(event ServerEvent, payload interface{}) {
	state, ok := mc.GetPlaybackState()
	if !ok {
		return
	}
	mc.sendTo(state.ClientID, event, payload)
}

func (mc *MpvCore) Pause()                       { mc.withStateEvent(ServerEventPause, nil) }
func (mc *MpvCore) Resume()                      { mc.withStateEvent(ServerEventResume, nil) }
func (mc *MpvCore) Seek(seconds float64)         { mc.withStateEvent(ServerEventSeek, seconds) }
func (mc *MpvCore) SeekTo(seconds float64)       { mc.withStateEvent(ServerEventSeekTo, seconds) }
func (mc *MpvCore) SetFullscreen(value bool)     { mc.withStateEvent(ServerEventSetFullscreen, value) }
func (mc *MpvCore) SetPip(value bool)            { mc.withStateEvent(ServerEventSetPip, value) }
func (mc *MpvCore) SetAudioTrack(id interface{}) { mc.withStateEvent(ServerEventSetAudioTrack, id) }
func (mc *MpvCore) SetSubtitleTrack(id interface{}) {
	mc.withStateEvent(ServerEventSetSubtitleTrack, id)
}
func (mc *MpvCore) AddSubtitleTrack(t *SubtitleTrack) {
	mc.withStateEvent(ServerEventAddSubtitleTrack, t)
}
func (mc *MpvCore) PlayPlaylistEpisode(which string) {
	mc.withStateEvent(ServerEventPlayPlaylistEntry, which)
}
func (mc *MpvCore) SendInSightData(data *InSightData) {
	mc.withStateEvent(ServerEventInSightData, data)
}

func (mc *MpvCore) ShowMessage(message string, duration int) {
	mc.withStateEvent(ServerEventShowMessage, struct {
		Message  string `json:"message"`
		Duration int    `json:"duration"`
	}{message, duration})
}

func (mc *MpvCore) Terminate() {
	mc.pendingMu.RLock()
	pending := mc.pendingState
	mc.pendingMu.RUnlock()
	if state, ok := mc.GetPlaybackState(); ok {
		mc.sendTo(state.ClientID, ServerEventTerminate, nil)
	} else if pending != nil {
		mc.sendTo(pending.ClientID, ServerEventTerminate, nil)
	}
	mc.clearPlayback()
}

func (mc *MpvCore) GetPlaybackState() (*PlaybackState, bool) {
	mc.stateMu.RLock()
	defer mc.stateMu.RUnlock()
	return mc.state, mc.state != nil && mc.state.PlaybackInfo != nil
}

func (mc *MpvCore) GetPlaybackStatus() (*PlaybackStatus, bool) {
	mc.statusMu.RLock()
	defer mc.statusMu.RUnlock()
	return mc.status, mc.status != nil && mc.status.ID != ""
}

func (mc *MpvCore) GetCurrentPlaybackType() (PlaybackType, bool) {
	state, ok := mc.GetPlaybackState()
	if !ok {
		return "", false
	}
	return state.PlaybackInfo.PlaybackType, true
}

func (mc *MpvCore) clearPlayback() {
	mc.pendingMu.Lock()
	mc.pendingState = nil
	mc.pendingMu.Unlock()
	mc.stateMu.Lock()
	mc.state = nil
	mc.stateMu.Unlock()
	mc.statusMu.Lock()
	mc.status = nil
	mc.statusMu.Unlock()
}

func (mc *MpvCore) setStatus(payload statusPayload) {
	state, ok := mc.GetPlaybackState()
	if !ok {
		return
	}
	mc.statusMu.Lock()
	mc.status = &PlaybackStatus{
		ID:          state.PlaybackInfo.ID,
		ClientID:    state.ClientID,
		CurrentTime: payload.CurrentTime,
		Duration:    payload.Duration,
		Paused:      payload.Paused,
	}
	mc.statusMu.Unlock()
}

func (mc *MpvCore) updateStatus(payload statusPayload) {
	mc.statusMu.Lock()
	if mc.status != nil {
		mc.status.CurrentTime = payload.CurrentTime
		mc.status.Duration = payload.Duration
		mc.status.Paused = payload.Paused
	}
	mc.statusMu.Unlock()
}

func (mc *MpvCore) PullStatus() (StatusEvent, bool) {
	if _, ok := mc.GetPlaybackState(); !ok {
		return StatusEvent{}, false
	}
	done := make(chan struct{})
	var ret StatusEvent
	cancel := mc.RegisterEventCallback(func(event VideoEvent) bool {
		if status, ok := event.(*StatusEvent); ok {
			ret = *status
			close(done)
			return false
		}
		return true
	})
	defer cancel()
	mc.withStateEvent(ServerEventGetStatus, nil)
	select {
	case <-done:
		return ret, true
	case <-time.After(5 * time.Second):
		return StatusEvent{}, false
	}
}

func (mc *MpvCore) GetPlaylist() (*PlaylistState, bool) {
	done := make(chan struct{})
	var ret *PlaylistState
	cancel := mc.RegisterEventCallback(func(event VideoEvent) bool {
		if value, ok := event.(*PlaylistStateEvent); ok {
			ret = value.Playlist
			close(done)
			return false
		}
		return true
	})
	defer cancel()
	mc.withStateEvent(ServerEventGetPlaylist, nil)
	select {
	case <-done:
		return ret, ret != nil
	case <-time.After(5 * time.Second):
		return nil, false
	}
}

func (mc *MpvCore) GetSkipData() (*SkipData, bool) {
	done := make(chan struct{})
	var ret *SkipData
	cancel := mc.RegisterEventCallback(func(event VideoEvent) bool {
		if value, ok := event.(*SkipDataEvent); ok {
			ret = value.SkipData
			close(done)
			return false
		}
		return true
	})
	defer cancel()
	mc.withStateEvent(ServerEventGetSkipData, nil)
	select {
	case <-done:
		return ret, true
	case <-time.After(5 * time.Second):
		return nil, false
	}
}

func (mc *MpvCore) SetSkipData(data *SkipData) { mc.withStateEvent(ServerEventSetSkipData, data) }
func (mc *MpvCore) ClearSkipData()             { mc.withStateEvent(ServerEventSetSkipData, nil) }

func (mc *MpvCore) listenToClientEvents() {
	for raw := range mc.clientPlayerEventSubscriber.Channel {
		event := &ClientEvent{}
		payload, _ := json.Marshal(raw.Payload)
		if json.Unmarshal(payload, event) != nil {
			continue
		}
		clientID := event.ClientID
		if clientID == "" {
			clientID = raw.ClientID
		}

		if event.Type == ClientEventPlaybackLoaded {
			mc.handlePlaybackLoaded(event, clientID)
			continue
		}
		if event.Type == ClientEventTerminated {
			mc.handleTerminated(event, clientID)
			continue
		}

		state, ok := mc.GetPlaybackState()
		if !ok || (clientID != "" && state.ClientID != "" && clientID != state.ClientID) {
			continue
		}

		switch event.Type {
		case ClientEventLoadedMetadata, ClientEventCanPlay, ClientEventPaused, ClientEventResumed,
			ClientEventStatus, ClientEventSeeked, ClientEventCompleted:
			var p statusPayload
			if event.UnmarshalAs(&p) != nil || (p.ID != "" && p.ID != state.PlaybackInfo.ID) {
				continue
			}
			if event.Type == ClientEventLoadedMetadata {
				mc.setStatus(p)
				mc.PushEvent(&LoadedMetadataEvent{CurrentTime: p.CurrentTime, Duration: p.Duration, Paused: p.Paused})
			} else {
				mc.updateStatus(p)
				switch event.Type {
				case ClientEventCanPlay:
					mc.PushEvent(&CanPlayEvent{CurrentTime: p.CurrentTime, Duration: p.Duration, Paused: p.Paused})
				case ClientEventPaused:
					mc.PushEvent(&PausedEvent{CurrentTime: p.CurrentTime, Duration: p.Duration})
				case ClientEventResumed:
					mc.PushEvent(&ResumedEvent{CurrentTime: p.CurrentTime, Duration: p.Duration})
				case ClientEventStatus:
					mc.PushEvent(&StatusEvent{CurrentTime: p.CurrentTime, Duration: p.Duration, Paused: p.Paused})
				case ClientEventSeeked:
					mc.PushEvent(&SeekedEvent{CurrentTime: p.CurrentTime, Duration: p.Duration, Paused: p.Paused})
				case ClientEventCompleted:
					mc.PushEvent(&CompletedEvent{CurrentTime: p.CurrentTime, Duration: p.Duration})
				}
			}
		case ClientEventEnded:
			var p endedPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&EndedEvent{AutoNext: p.AutoNext})
			}
		case ClientEventPlayerError:
			var p errorPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&ErrorEvent{Error: p.Error})
			}
		case ClientEventFullscreenChanged:
			var p struct {
				Fullscreen bool `json:"fullscreen"`
			}
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&FullscreenChangedEvent{Fullscreen: p.Fullscreen})
			}
		case ClientEventPipChanged:
			var p struct {
				Pip bool `json:"pip"`
			}
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&PipChangedEvent{Pip: p.Pip})
			}
		case ClientEventAudioTrackChanged:
			var p trackChangedPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&AudioTrackChangedEvent{TrackID: p.TrackID})
			}
		case ClientEventSubtitleTrackChanged:
			var p trackChangedPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&SubtitleTrackChangedEvent{TrackID: p.TrackID})
			}
		case ClientEventPlaylistState:
			var p playlistPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&PlaylistStateEvent{Playlist: p.Playlist})
			}
		case ClientEventSkipData:
			var p skipDataPayload
			if event.UnmarshalAs(&p) == nil {
				mc.PushEvent(&SkipDataEvent{SkipData: p.SkipData})
			}
		}
	}
}

func (mc *MpvCore) handlePlaybackLoaded(event *ClientEvent, clientID string) {
	var p playbackLoadedPayload
	if event.UnmarshalAs(&p) != nil {
		return
	}
	if p.ClientID != "" {
		clientID = p.ClientID
	}
	mc.pendingMu.Lock()
	pending := mc.pendingState
	if pending == nil || (clientID != "" && pending.ClientID != "" && clientID != pending.ClientID) ||
		(p.ID != "" && pending.PlaybackInfo.ID != p.ID) {
		mc.pendingMu.Unlock()
		return
	}
	mc.pendingState = nil
	mc.pendingMu.Unlock()
	mc.stateMu.Lock()
	mc.state = pending
	mc.stateMu.Unlock()
	mc.PushEvent(&PlaybackLoadedEvent{State: *pending})
}

func (mc *MpvCore) handleTerminated(event *ClientEvent, clientID string) {
	var p terminatedPayload
	_ = event.UnmarshalAs(&p)
	if p.ClientID != "" {
		clientID = p.ClientID
	}

	var id string
	var playbackType PlaybackType
	if state, ok := mc.GetPlaybackState(); ok {
		if clientID != "" && state.ClientID != "" && clientID != state.ClientID {
			return
		}
		id = state.PlaybackInfo.ID
		playbackType = state.PlaybackInfo.PlaybackType
	} else {
		mc.pendingMu.RLock()
		pending := mc.pendingState
		mc.pendingMu.RUnlock()
		if pending != nil && (clientID == "" || pending.ClientID == "" || clientID == pending.ClientID) {
			id = pending.PlaybackInfo.ID
			playbackType = pending.PlaybackInfo.PlaybackType
		} else if clientID != "" {
			id = p.ID
			playbackType = p.PlaybackType
		} else {
			return
		}
	}
	if p.ID != "" && id != "" && p.ID != id {
		return
	}
	eventValue := &TerminatedEvent{}
	eventValue.identify(id, clientID, playbackType)
	select {
	case mc.eventBus <- eventValue:
	default:
		mc.logger.Warn().Msg("mpvcore: event bus full while terminating")
	}
	mc.clearPlayback()
}

var errNoPlayback = errors.New("mpvcore playback is not active")

func (mc *MpvCore) CurrentPlaybackInfo() (*PlaybackInfo, error) {
	state, ok := mc.GetPlaybackState()
	if !ok {
		return nil, errNoPlayback
	}
	return state.PlaybackInfo, nil
}
