package videocore

import (
	"encoding/json"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/mkvparser"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type (
	// VideoCore allows to interact with the built-in HTML5 video player in Seanime.
	// It can be the NativePlayer (Seanime Denshi player) or the WebPlayer.
	VideoCore struct {
		wsEventManager              events.WSEventManagerInterface
		clientPlayerEventSubscriber *events.ClientEventSubscriber

		continuityManager          *continuity.Manager
		metadataProviderRef        *util.Ref[metadata_provider.Provider]
		discordPresence            *discordrpc_presence.Presence
		platformRef                *util.Ref[platform.Platform]
		refreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		isOfflineRef               *util.Ref[bool]

		playbackStatusMu sync.RWMutex
		playbackStatus   *PlaybackStatus
		playbackStateMu  sync.RWMutex
		playbackState    *PlaybackState

		subscribers *result.Map[string, *Subscriber]

		eventBus       chan VideoEvent
		dispatcherStop chan struct{}
		startOnce      sync.Once

		logger *zerolog.Logger
	}

	// Subscriber listens to the player events
	Subscriber struct {
		eventCh chan VideoEvent
	}

	NewVideoCoreOptions struct {
		WsEventManager             events.WSEventManagerInterface
		Logger                     *zerolog.Logger
		MetadataProviderRef        *util.Ref[metadata_provider.Provider]
		ContinuityManager          *continuity.Manager
		DiscordPresence            *discordrpc_presence.Presence
		PlatformRef                *util.Ref[platform.Platform]
		RefreshAnimeCollectionFunc func()
		IsOfflineRef               *util.Ref[bool]
	}
)

// New returns a new instance of VideoCore. There should be only one for the lifetime of the app.
func New(opts NewVideoCoreOptions) *VideoCore {
	vc := &VideoCore{
		wsEventManager:              opts.WsEventManager,
		continuityManager:           opts.ContinuityManager,
		discordPresence:             opts.DiscordPresence,
		metadataProviderRef:         opts.MetadataProviderRef,
		platformRef:                 opts.PlatformRef,
		refreshAnimeCollectionFunc:  opts.RefreshAnimeCollectionFunc,
		isOfflineRef:                opts.IsOfflineRef,
		subscribers:                 result.NewMap[string, *Subscriber](),
		clientPlayerEventSubscriber: opts.WsEventManager.SubscribeToClientVideoCoreEvents("videocore"),
		logger:                      opts.Logger,
		eventBus:                    make(chan VideoEvent, 100),
		dispatcherStop:              make(chan struct{}),
	}
	vc.Start()
	return vc
}

func (vc *VideoCore) Start() {
	vc.startOnce.Do(func() {
		vc.listenToClientEvents()
		go func() {
			for {
				select {
				case <-vc.dispatcherStop:
					return
				case event := <-vc.eventBus:
					vc.dispatchEvent(event)
				}
			}
		}()
		vc.setupEffects()
	})
}

// Shutdown gracefully stops the dispatcher.
func (vc *VideoCore) Shutdown() {
	close(vc.dispatcherStop)
}

func (vc *VideoCore) PushEvent(event VideoEvent) {
	select {
	case vc.eventBus <- event:
	default:
		vc.logger.Warn().Msgf("videcore: Event bus full, dropping event %s", event.GetPlaybackId())
	}
}

func (vc *VideoCore) dispatchEvent(event VideoEvent) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	event.Identify(state.PlaybackInfo.Id, state.ClientId, state.PlayerType, state.PlaybackInfo.PlaybackType)
	//if _, ok := event.(*VideoStatusEvent); !ok {
	//	vc.logger.Debug().Msgf("videocore: Dispatching event %T", event)
	//} else {
	//	//vc.logger.Trace().Msgf("videocore: Dispatching status, playbackId: %s, clientId: %s", event.GetPlaybackId(), event.GetClientId())
	//}
	vc.subscribers.Range(func(id string, subscriber *Subscriber) bool {
		if event.IsCritical() {
			select {
			case subscriber.eventCh <- event:
			case <-time.After(1 * time.Second):
				vc.logger.Warn().Msgf("videocore: Subscriber %s blocked critical event %T", id, event)
			}
		} else {
			// Drop non-critical events if busy
			select {
			case subscriber.eventCh <- event:
			default:
				//vc.logger.Warn().Msgf("videocore: Subscriber %s dropped non-critical event %T", id, event)
			}
		}
		return true
	})
}

// sendPlayerEventTo sends an event of type events.VideoCoreEventType to the client.
func (vc *VideoCore) sendPlayerEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	vc.playbackStatusMu.RLock()
	if vc.playbackStatus != nil && len(vc.playbackStatus.Id) > 0 && vc.playbackStatus.Duration > 0 && clientId == "" {
		clientId = vc.playbackStatus.ClientId
	}
	vc.playbackStatusMu.RUnlock()

	vc.logger.Trace().Msgf("videocore: Sending event %s to client %s", t, clientId)

	if clientId != "" {
		vc.wsEventManager.SendEventTo(clientId, string(events.VideoCoreEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		}, noLog...)
	} else {
		vc.wsEventManager.SendEvent(string(events.VideoCoreEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		})
	}
}

func (vc *VideoCore) sendPlayerEvent(t string, payload interface{}) {
	vc.wsEventManager.SendEvent(string(events.VideoCoreEventType), struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    t,
		Payload: payload,
	})
}

// Subscribe lets other modules subscribe to the native player events
func (vc *VideoCore) Subscribe(id string) *Subscriber {
	subscriber := &Subscriber{
		eventCh: make(chan VideoEvent, 100),
	}
	vc.subscribers.Set(id, subscriber)

	return subscriber
}

// Unsubscribe removes a subscriber from the player.
func (vc *VideoCore) Unsubscribe(id string) {
	if subscriber, ok := vc.subscribers.Get(id); ok {
		close(subscriber.eventCh)
		vc.subscribers.Delete(id)
	}
}

// Events returns the event channel for the subscriber.
func (s *Subscriber) Events() <-chan VideoEvent {
	return s.eventCh
}

func (vc *VideoCore) RegisterEventCallback(callback func(event VideoEvent, cancelFunc func())) (cancel func()) {
	id := uuid.NewString()
	sub := vc.Subscribe(id)
	cancel = func() {
		vc.Unsubscribe(id)
	}
	go func(sub *Subscriber) {
		for event := range sub.Events() {
			callback(event, cancel)
		}
	}(sub)

	return cancel
}

func (vc *VideoCore) GetPlaybackStatus() (*PlaybackStatus, bool) {
	vc.playbackStatusMu.RLock()
	defer vc.playbackStatusMu.RUnlock()
	return vc.playbackStatus, vc.playbackStatus != nil && len(vc.playbackStatus.Id) > 0 && vc.playbackStatus.Duration > 0
}

func (vc *VideoCore) GetPlaybackState() (*PlaybackState, bool) {
	vc.playbackStateMu.RLock()
	defer vc.playbackStateMu.RUnlock()
	return vc.playbackState, vc.playbackState != nil && vc.playbackState.PlaybackInfo != nil && vc.playbackState.PlaybackInfo.Episode != nil
}

func (vc *VideoCore) GetCurrentPlaybackInfo() (*VideoPlaybackInfo, bool) {
	vc.playbackStateMu.RLock()
	defer vc.playbackStateMu.RUnlock()
	if vc.playbackState == nil {
		return nil, false
	}
	return vc.playbackState.PlaybackInfo, true
}

func (vc *VideoCore) GetCurrentMedia() (*anilist.BaseAnime, bool) {
	info, ok := vc.GetCurrentPlaybackInfo()
	if !ok {
		return nil, false
	}
	return info.Media, true
}

func (vc *VideoCore) GetCurrentClientId() string {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return ""
	}
	return state.ClientId
}

func (vc *VideoCore) GetCurrentPlayerType() (PlayerType, bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return "", false
	}
	return state.PlayerType, true
}

func (vc *VideoCore) GetCurrentPlaybackType() (PlaybackType, bool) {
	info, ok := vc.GetCurrentPlaybackInfo()
	if !ok {
		return "", false
	}
	return info.PlaybackType, true
}

func (vc *VideoCore) clearPlayback() {
	vc.setPlaybackStatus(nil)
	vc.setPlaybackState(nil)
}

func (vc *VideoCore) setPlaybackState(state *PlaybackState) {
	vc.playbackStateMu.Lock()
	defer vc.playbackStateMu.Unlock()
	vc.playbackState = state
}

func (vc *VideoCore) setPlaybackStatus(status *PlaybackStatus) {
	vc.setPlaybackStatusFn(status)
}

// setPlaybackStatus sets the current playback status of the player.
// and notifies all subscribers of the change (if it exists).
func (vc *VideoCore) setPlaybackStatusFn(status *PlaybackStatus) {
	vc.playbackStatusMu.Lock()
	vc.playbackStatus = status
	shouldNotify := vc.playbackStatus != nil && len(vc.playbackStatus.Id) > 0 && vc.playbackStatus.Duration > 0
	var currentTime, duration float64
	var paused bool
	if shouldNotify {
		currentTime = vc.playbackStatus.CurrentTime
		duration = vc.playbackStatus.Duration
		paused = vc.playbackStatus.Paused
	}
	vc.playbackStatusMu.Unlock()

	if shouldNotify {
		vc.PushEvent(&VideoStatusEvent{
			CurrentTime: currentTime,
			Duration:    duration,
			Paused:      paused,
		})
	}
}

// updatePlaybackStatus updates the current playback status of the player only if it exists.
// and notifies all subscribers of the change.
func (vc *VideoCore) updatePlaybackStatusFn(do func()) {
	vc.playbackStatusMu.Lock()
	if vc.playbackStatus == nil || len(vc.playbackStatus.Id) == 0 || vc.playbackStatus.Duration <= 0 {
		vc.playbackStatusMu.Unlock()
		return
	}
	do()
	currentTime := vc.playbackStatus.CurrentTime
	duration := vc.playbackStatus.Duration
	paused := vc.playbackStatus.Paused
	vc.playbackStatusMu.Unlock()

	vc.PushEvent(&VideoStatusEvent{
		CurrentTime: currentTime,
		Duration:    duration,
		Paused:      paused,
	})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Server Events
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause sends a pause command to the video player.
func (vc *VideoCore) Pause() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventPause), nil)
}

// Resume sends a resume command to the video player.
func (vc *VideoCore) Resume() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventResume), nil)
}

// Seek sends a seek command to the video player.
// seconds is the amount to seek forward (positive) or backward (negative).
func (vc *VideoCore) Seek(seconds float64) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSeek), seconds)
}

// SeekTo sends a seek-to command to the video player.
// seconds is the absolute time position to seek to.
func (vc *VideoCore) SeekTo(seconds float64) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSeekTo), seconds)
}

// SetFullscreen sends a set-fullscreen command to the video player.
func (vc *VideoCore) SetFullscreen(fullscreen bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSetFullscreen), fullscreen)
}

// SetPip sends a set-pip command to the video player.
func (vc *VideoCore) SetPip(pip bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSetPip), pip)
}

// SetSubtitleTrack sends a set-subtitle-track command to the video player.
func (vc *VideoCore) SetSubtitleTrack(trackNumber int) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSetSubtitleTrack), trackNumber)
}

// AddSubtitleTrack sends an add-subtitle-track command to the video player.
func (vc *VideoCore) AddSubtitleTrack(track *mkvparser.TrackInfo) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventAddSubtitleTrack), track)
}

// AddSubtitleTrack sends an add-external-subtitle-track command to the video player.
func (vc *VideoCore) AddExternalSubtitleTrack(track *VideoSubtitleTrack) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventAddExternalSubtitleTrack), track)
}

// SetMediaCaptionTrack sends a set-media-caption-track command to the video player.
func (vc *VideoCore) SetMediaCaptionTrack(trackIndex int) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSetMediaCaptionTrack), trackIndex)
}

// AddMediaCaptionTrack sends an add-media-caption-track command to the video player.
func (vc *VideoCore) AddMediaCaptionTrack(track interface{}) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventAddMediaCaptionTrack), track)
}

// SetAudioTrack sends a set-audio-track command to the video player.
func (vc *VideoCore) SetAudioTrack(trackNumber int) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventSetAudioTrack), trackNumber)
}

// Terminate sends a terminate command to the video player.
func (vc *VideoCore) Terminate() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventTerminate), nil)
	vc.clearPlayback()
}

// StartOnlinestreamPlayback sends a start-onlinestream-playback command to the video player.
func (vc *VideoCore) StartOnlinestreamWatchParty(params *OnlinestreamParams) {
	vc.sendPlayerEventTo("", string(ServerEventStartOnlinestreamWatchParty), params)
}

// SendGetFullscreen sends a get-fullscreen request to the video player.
func (vc *VideoCore) SendGetFullscreen() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetFullscreen), nil)
}

// SendGetPip sends a get-pip request to the video player.
func (vc *VideoCore) SendGetPip() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetPip), nil)
}

// SendGetAnime4K sends a get-anime-4k request to the video player.
func (vc *VideoCore) SendGetAnime4K() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetAnime4K), nil)
}

// SendGetSubtitleTrack sends a get-subtitle-track request to the video player.
func (vc *VideoCore) SendGetSubtitleTrack() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetSubtitleTrack), nil)
}

// SendGetAudioTrack sends a get-audio-track request to the video player.
func (vc *VideoCore) SendGetAudioTrack() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetAudioTrack), nil)
}

// SendGetMediaCaptionTrack sends a get-media-caption-track request to the video player.
func (vc *VideoCore) SendGetMediaCaptionTrack() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetMediaCaptionTrack), nil)
}

// SendGetPlaybackState sends a get-playback-state request to the video player.
func (vc *VideoCore) SendGetPlaybackState() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetPlaybackState), nil)
}

// PullStatus pulls the current playback status from the video player.
func (vc *VideoCore) PullStatus() (ret VideoStatusEvent, ok bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return VideoStatusEvent{}, false
	}
	done := make(chan struct{})
	vc.RegisterEventCallback(func(e VideoEvent, cancel func()) {
		switch event := e.(type) {
		case *VideoStatusEvent:
			ret = *event
			cancel()
			close(done)
			return
		}
	})
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetStatus), nil)
	<-done
	return ret, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (vc *VideoCore) listenToClientEvents() {
	// Start a goroutine to listen to video core events
	go func() {
		for {
			select {
			// Listen to video core events from the client
			case clientEvent := <-vc.clientPlayerEventSubscriber.Channel:
				playerEvent := &ClientEvent{}
				marshaled, _ := json.Marshal(clientEvent.Payload)
				// Unmarshal the player event
				if err := json.Unmarshal(marshaled, &playerEvent); err == nil {
					// Handle events
					switch playerEvent.Type {
					case PlayerEventVideoLoaded:
						payload := &clientVideoLoadedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.setPlaybackState(&payload.State)
							vc.PushEvent(&VideoLoadedEvent{
								State: payload.State,
							})
						}
					case PlayerEventVideoPlaybackState:
						payload := &clientVideoLoadedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.setPlaybackState(&payload.State)
							vc.PushEvent(&VideoPlaybackStateEvent{
								State: payload.State,
							})
						}
					case PlayerEventVideoLoadedMetadata:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							ps, ok := vc.GetPlaybackState()
							if !ok {
								continue
							}
							vc.setPlaybackStatus(&PlaybackStatus{
								Id:          ps.PlaybackInfo.Id,
								ClientId:    ps.ClientId,
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
								Paused:      payload.Paused,
							})
							vc.PushEvent(&VideoLoadedMetadataEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
								Paused:      payload.Paused,
							})
						}
					case PlayerEventVideoCanPlay:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = payload.Paused
							})
							vc.PushEvent(&VideoCanPlayEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
								Paused:      payload.Paused,
							})
						}
					case PlayerEventVideoSeeked:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = payload.Paused
							})
							vc.PushEvent(&VideoSeekedEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
								Paused:      payload.Paused,
							})
						}
					case PlayerEventVideoPaused:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = true
							})
							vc.PushEvent(&VideoPausedEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
							})
						}
					case PlayerEventVideoResumed:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = false
							})
							vc.PushEvent(&VideoResumedEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
							})
						}
					case PlayerEventVideoEnded:
						payload := &clientVideoEndedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.CurrentTime = vc.playbackStatus.Duration
								vc.playbackStatus.Paused = true
							})
							vc.PushEvent(&VideoEndedEvent{
								AutoNext: payload.AutoNext,
							})
						}
					case PlayerEventVideoStatus:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = payload.Paused
							})
						}
					case PlayerEventVideoCompleted:
						payload := &clientVideoStatusPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.updatePlaybackStatusFn(func() {
								vc.playbackStatus.Duration = payload.Duration
								vc.playbackStatus.CurrentTime = payload.CurrentTime
								vc.playbackStatus.Paused = payload.Paused
							})
							vc.PushEvent(&VideoCompletedEvent{
								CurrentTime: payload.CurrentTime,
								Duration:    payload.Duration,
							})
						}
					case PlayerEventVideoFullscreen:
						payload := &clientVideoFullscreenPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoFullscreenEvent{
								Fullscreen: payload.Fullscreen,
							})
						}
					case PlayerEventVideoSubtitleTrack:
						payload := &clientVideoSubtitleTrackPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoSubtitleTrackEvent{
								TrackNumber: payload.TrackNumber,
								Kind:        payload.Kind,
							})
						}
					case PlayerEventMediaCaptionTrack:
						payload := &clientVideoMediaCaptionTrackPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoMediaCaptionTrackEvent{
								TrackIndex: payload.TrackIndex,
							})
						}
					case PlayerEventVideoAudioTrack:
						payload := &clientVideoAudioTrackPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoAudioTrackEvent{
								TrackNumber: payload.TrackNumber,
								IsHls:       payload.IsHls,
							})
						}
					case PlayerEventAnime4K:
						payload := &clientVideoAnime4KPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoAnime4KEvent{
								Option: payload.Option,
							})
						}
					case PlayerEventVideoPip:
						payload := &clientVideoPipPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoPipEvent{
								Pip: payload.Pip,
							})
						}
					case PlayerEventVideoError:
						payload := &clientVideoErrorPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&VideoErrorEvent{
								Error: payload.Error,
							})
						}
					case PlayerEventVideoTerminated:
						// No payload
						vc.PushEvent(&VideoTerminatedEvent{})
						vc.clearPlayback()
					case PlayerEventSubtitleFileUploaded:
						payload := &clientSubtitleFileUploadedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							vc.PushEvent(&SubtitleFileUploadedEvent{
								Filename: payload.Filename,
								Content:  payload.Content,
							})
						}
					}
				}
			}
		}
	}()
}
