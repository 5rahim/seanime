package videocore

import (
	"context"
	"encoding/json"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/models"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/mkvparser"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

type (
	// VideoCore represents the built-in HTML5 video player.
	// It can be the NativePlayer (Seanime Denshi player) or the WebPlayer.
	VideoCore struct {
		wsEventManager              events.WSEventManagerInterface
		clientPlayerEventSubscriber *events.ClientEventSubscriber

		translatorService *TranslatorService

		continuityManager          *continuity.Manager
		metadataProviderRef        *util.Ref[metadata_provider.Provider]
		discordPresence            *discordrpc_presence.Presence
		platformRef                *util.Ref[platform.Platform]
		refreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection
		isOfflineRef               *util.Ref[bool]

		playbackStatusMu  sync.RWMutex
		playbackStatus    *PlaybackStatus
		playbackStateMu   sync.RWMutex
		playbackState     *PlaybackState
		playbackMkvEvents *result.Map[uint64, []*mkvparser.SubtitleEvent]

		inSight *InSight

		subscribers *result.Map[string, *Subscriber]

		eventBus       chan VideoEvent
		dispatcherStop chan struct{}
		startOnce      sync.Once

		logger     *zerolog.Logger
		settingsMu sync.RWMutex
		settings   *models.Settings
	}

	// Subscriber listens to the player events
	Subscriber struct {
		id        string
		eventCh   chan VideoEvent
		isClosed  atomic.Bool
		closeOnce sync.Once
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
		playbackMkvEvents:           result.NewMap[uint64, []*mkvparser.SubtitleEvent](),
	}
	vc.Start()
	vc.inSight = NewInSight(opts.Logger, vc)
	vc.inSight.Start()
	return vc
}

func (vc *VideoCore) SetSettings(settings *models.Settings) {
	if settings == nil {
		return
	}
	vc.logger.Trace().Msgf("videocore: Setting settings")
	vc.settingsMu.Lock()
	vc.settings = settings
	vc.settingsMu.Unlock()

	if vc.translatorService != nil {
		vc.translatorService.Shutdown()
	}
	vc.translatorService = nil
	if settings.GetMediaPlayer().VcTranslate {
		vc.logger.Trace().Msgf("videocore: Setting up translator service %s", settings.GetMediaPlayer().VcTranslateProvider)
		vc.translatorService = NewTranslatorService(vc, settings.GetMediaPlayer().VcTranslateApiKey, settings.GetMediaPlayer().VcTranslateProvider, settings.GetMediaPlayer().VcTranslateTargetLanguage)
	}
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
	// Before pushing the event, identify it with the playback state.
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	event.identify(state.PlaybackInfo.Id, state.ClientId, state.PlayerType, state.PlaybackInfo.PlaybackType)
	select {
	case vc.eventBus <- event:
	default:
		vc.logger.Warn().Msgf("videcore: Event bus full, dropping event %s", event.GetPlaybackId())
	}
}

func (vc *VideoCore) dispatchEvent(event VideoEvent) {
	//if _, ok := event.(*VideoStatusEvent); !ok {
	//	vc.logger.Debug().Msgf("videocore: Dispatching event %T", event)
	//} else {
	//	//vc.logger.Trace().Msgf("videocore: Dispatching status, playbackId: %s, clientId: %s", event.GetPlaybackId(), event.GetClientId())
	//}
	vc.subscribers.Range(func(id string, subscriber *Subscriber) bool {
		if subscriber.isClosed.Load() {
			return true
		}
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

	if len(noLog) == 0 || !noLog[0] {
		vc.logger.Trace().Msgf("videocore: Sending event %s to client %s", t, clientId)
	}

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
		id:      id,
		eventCh: make(chan VideoEvent, 100),
	}
	vc.subscribers.Set(id, subscriber)

	return subscriber
}

// Unsubscribe removes a subscriber from the player.
func (vc *VideoCore) Unsubscribe(id string) {
	if subscriber, ok := vc.subscribers.Pop(id); ok {
		subscriber.isClosed.Store(true)
		subscriber.closeOnce.Do(func() {
			close(subscriber.eventCh)
		})
	}
}

// Events returns the event channel for the subscriber.
func (s *Subscriber) Events() <-chan VideoEvent {
	return s.eventCh
}

// GetId returns the subscriber id
func (s *Subscriber) GetId() string {
	return s.id
}

func (vc *VideoCore) RegisterEventCallback(callback func(event VideoEvent) bool) (cancel func()) {
	id := uuid.NewString()
	sub := vc.Subscribe(id)
	cancel = func() {
		vc.Unsubscribe(id)
	}
	go func(sub *Subscriber) {
		defer vc.Unsubscribe(id)
		for event := range sub.Events() {
			cont := callback(event)
			if !cont {
				cancel()
				return
			}
		}
	}(sub)

	return cancel
}

func (vc *VideoCore) GetPlaybackStatus() (*PlaybackStatus, bool) {
	vc.playbackStatusMu.RLock()
	defer vc.playbackStatusMu.RUnlock()
	return vc.playbackStatus, vc.playbackStatus != nil && len(vc.playbackStatus.Id) > 0 && vc.playbackStatus.Duration > 0
}

// GetPlaybackState returns the current playback state of the player.
// This will return nil right after VideoTerminatedEvent is received.
func (vc *VideoCore) GetPlaybackState() (*PlaybackState, bool) {
	vc.playbackStateMu.RLock()
	defer vc.playbackStateMu.RUnlock()
	return vc.playbackState, vc.playbackState != nil && vc.playbackState.PlaybackInfo != nil && vc.playbackState.PlaybackInfo.Episode != nil
}

// GetCurrentPlaybackInfo returns the current playback info of the player.
// This will return nil right after VideoTerminatedEvent is received.
func (vc *VideoCore) GetCurrentPlaybackInfo() (*VideoPlaybackInfo, bool) {
	vc.playbackStateMu.RLock()
	defer vc.playbackStateMu.RUnlock()
	if vc.playbackState == nil {
		return nil, false
	}
	return vc.playbackState.PlaybackInfo, true
}

// GetCurrentMedia returns the current media.
// This will return nil right after VideoTerminatedEvent is received.
func (vc *VideoCore) GetCurrentMedia() (*anilist.BaseAnime, bool) {
	info, ok := vc.GetCurrentPlaybackInfo()
	if !ok {
		return nil, false
	}
	return info.Media, true
}

// GetCurrentClientId returns the current client id.
// This will return an empty string right after VideoTerminatedEvent is received, use VideoEvent.GetClientId() instead.
func (vc *VideoCore) GetCurrentClientId() string {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return ""
	}
	return state.ClientId
}

// GetCurrentPlayerType returns the current player type.
// This will return false right after VideoTerminatedEvent is received, use VideoEvent.GetPlayerType() instead.
func (vc *VideoCore) GetCurrentPlayerType() (PlayerType, bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return "", false
	}
	return state.PlayerType, true
}

// GetCurrentPlaybackType returns the current playback type.
// This will return false right after VideoTerminatedEvent is received, use VideoEvent.GetPlaybackType() instead.
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
	vc.playbackMkvEvents.Clear()
	vc.inSight.Clear()
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

func (vc *VideoCore) ShowMessage(message string, milliseconds int) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventShowMessage), struct {
		Message  string `json:"message"`
		Duration int    `json:"duration"`
	}{
		Message:  message,
		Duration: milliseconds,
	})
}

// PlayPlaylistEpisode sends a play-episode command to the video player.
// which is "next", "previous", or the AniDB episode ID.
func (vc *VideoCore) PlayPlaylistEpisode(which string) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventPlayPlaylistEpisode), which)
}

// SendInSightData sends InSight data for a playback session.
func (vc *VideoCore) SendInSightData(data *InSightData) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventInSightData), data)
}

// Terminate sends a terminate command to the video player and clears the playback state.
// The video player should stop on the client.
func (vc *VideoCore) Terminate() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventTerminate), nil)
	vc.clearPlayback()
}

// Reset clears the current playback state without sending a terminate command to the video player.
// This will cause further client events to be ignored until a new playback state is set.
func (vc *VideoCore) Reset() {
	_, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.clearPlayback()
}

// StartOnlinestreamPlayback sends a start-onlinestream-playback command to the video player.
func (vc *VideoCore) StartOnlinestreamWatchParty(params *OnlinestreamParams) {
	// devnote: dispatch to all connected clients with sendPlayerEvent
	vc.sendPlayerEvent(string(ServerEventStartOnlinestreamWatchParty), params)
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

// SendGetSubtitleTrackContent sends a get-subtitle-track-content request to the video player.
func (vc *VideoCore) SendGetSubtitleTrackContent() {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return
	}
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetSubtitleTrackContent), nil)
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

// GetPlaylist sends a get-text-tracks request to the video player and returns the text tracks.
func (vc *VideoCore) GetTextTracks() (ret []*VideoTextTrack, ok bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return nil, false
	}
	done := make(chan struct{})
	cancel := vc.RegisterEventCallback(func(e VideoEvent) bool {
		switch event := e.(type) {
		case *VideoTextTracksEvent:
			ret = event.TextTracks
			close(done)
			return false // stop
		}
		return true // keep listening
	})
	go func(cancel func()) {
		defer cancel()
		<-time.After(5 * time.Second)
	}(cancel)
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetTextTracks), nil)
	<-done
	return ret, ret != nil
}

// GetPlaylist sends a get-playlist request to the video player and returns the playlist state.
func (vc *VideoCore) GetPlaylist() (ret *VideoPlaylistState, ok bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return nil, false
	}
	done := make(chan struct{})
	cancel := vc.RegisterEventCallback(func(e VideoEvent) bool {
		switch event := e.(type) {
		case *VideoPlaylistEvent:
			ret = event.Playlist
			close(done)
			return false // stop
		}
		return true // keep listening
	})
	go func(cancel func()) {
		defer cancel()
		<-time.After(5 * time.Second)
	}(cancel)
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetPlaylist), nil)
	<-done
	return ret, ret != nil
}

// PullStatus pulls the current playback status from the video player.
func (vc *VideoCore) PullStatus() (ret VideoStatusEvent, ok bool) {
	state, ok := vc.GetPlaybackState()
	if !ok {
		return VideoStatusEvent{}, false
	}
	done := make(chan struct{})
	cancel := vc.RegisterEventCallback(func(e VideoEvent) bool {
		switch event := e.(type) {
		case *VideoStatusEvent:
			ret = *event
			close(done)
			return false // stop
		}
		return true // keep listening
	})
	go func(cancel func()) {
		defer cancel()
		<-time.After(5 * time.Second)
	}(cancel)
	vc.sendPlayerEventTo(state.ClientId, string(ServerEventGetStatus), nil, true)
	<-done
	return ret, true
}

func (vc *VideoCore) RecordEvent(event *mkvparser.SubtitleEvent) {
	// todo
	//track, found := vc.playbackMkvEvents.Get(event.TrackNumber)
	//if !found {
	//	vc.playbackMkvEvents.Set(event.TrackNumber, []*mkvparser.SubtitleEvent{event})
	//	return
	//}
	//track = append(track, event)
	//vc.playbackMkvEvents.Set(event.TrackNumber, track)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (vc *VideoCore) listenToClientEvents() {
	// Start a goroutine to listen to video core events
	go func() {
		// Listen to video core events from the client
		for clientEvent := range vc.clientPlayerEventSubscriber.Channel {
			playerEvent := &ClientEvent{}
			marshaled, _ := json.Marshal(clientEvent.Payload)
			// Unmarshal the player event
			if err := json.Unmarshal(marshaled, &playerEvent); err == nil {
				// Validate that the event is from the current client
				currentState, hasState := vc.GetPlaybackState()
				if hasState && clientEvent.ClientID != "" && clientEvent.ClientID != currentState.ClientId {
					continue
				}

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
						vc.playbackMkvEvents.Clear()
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
				case PlayerEventVideoSubtitleTrackContent:
					payload := &clientVideoSubtitleTrackContentPayload{}
					if err := playerEvent.UnmarshalAs(&payload); err == nil {
						vc.PushEvent(&VideoSubtitleTrackContentEvent{
							TrackNumber: payload.TrackNumber,
							Content:     payload.Content,
							Type:        payload.Type,
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
				case PlayerEventVideoPlaylist:
					payload := &clientVideoPlaylistPayload{}
					if err := playerEvent.UnmarshalAs(payload); err == nil {
						vc.PushEvent(&VideoPlaylistEvent{
							Playlist: &payload.Playlist,
						})
					}
				case PlayerEventVideoTextTracks:
					payload := &clientVideoTextTracksPayload{}
					if err := playerEvent.UnmarshalAs(payload); err == nil {
						vc.PushEvent(&VideoTextTracksEvent{
							TextTracks: payload.TextTracks,
						})
					}
				case PlayerEventTranslateText:
					payload := &clientTranslateTextPayload{}
					if err := playerEvent.UnmarshalAs(payload); err == nil {
						// Translate in a goroutine
						go func() {
							state, ok := vc.GetPlaybackState()
							if !ok {
								return
							}
							translated := vc.TranslateText(context.Background(), payload.Text)
							// Send the result
							vc.sendPlayerEventTo(state.ClientId, string(ServerEventTranslatedText), struct {
								Original   string `json:"original"`
								Translated string `json:"translated"`
							}{
								Original:   payload.Text,
								Translated: translated,
							}, true)
						}()
					}
				case PlayerEventTranslateSubtitleFileTrack:
					payload := &VideoSubtitleTrack{}
					if err := playerEvent.UnmarshalAs(payload); err == nil {
						// Translate in a goroutine
						go func() {
							vc.logger.Trace().Msgf("videocore: Received subtitle track translation request")
							state, ok := vc.GetPlaybackState()
							if !ok {
								return
							}
							var translated string
							if payload.Src != nil && len(*payload.Src) > 0 {
								client := req.C()
								client.SetTimeout(30 * time.Second)
								resp := client.Get(*payload.Src).Do()

								if resp.IsErrorState() {
									vc.logger.Error().Err(resp.Err).Msgf("videocore: Failed to download subtitle file %s", *payload.Src)
									return
								}

								content := resp.String()

								from := mkvparser.DetectSubtitleType(content)
								translated = vc.TranslateContent(context.Background(), content, from)

							} else if payload.Content != nil && len(*payload.Content) > 0 {
								content := *payload.Content
								from := mkvparser.DetectSubtitleType(content)
								translated = vc.TranslateContent(context.Background(), content, from)
							}
							if translated != "" {
								// Modify the payload but keep the same index
								payload.Content = &translated
								payload.Src = nil
								payload.Label = payload.Label + " (translated)"
								payload.Language = strings.ToLower(vc.GetTranslationTargetLanguage())
								// Send the result
								vc.logger.Debug().Str("clientId", state.ClientId).Int("length", len(*payload.Content)).Msgf("videocore: Sending translated subtitle track")
								vc.sendPlayerEventTo(state.ClientId, string(ServerEventAddExternalSubtitleTrack), payload, true)
							} else {
								vc.logger.Error().Msgf("videocore: Failed to translate subtitle track")
							}
						}()
					}
				}
			}
		}
	}()
}
