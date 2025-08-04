package nativeplayer

import (
	"context"
	"seanime/internal/mkvparser"
	"time"

	"github.com/goccy/go-json"
)

type ServerEvent string

const (
	ServerEventOpenAndAwait     ServerEvent = "open-and-await"
	ServerEventWatch            ServerEvent = "watch"
	ServerEventSubtitleEvent    ServerEvent = "subtitle-event"
	ServerEventSetTracks        ServerEvent = "set-tracks"
	ServerEventPause            ServerEvent = "pause"
	ServerEventResume           ServerEvent = "resume"
	ServerEventSeek             ServerEvent = "seek"
	ServerEventError            ServerEvent = "error"
	ServerEventAddSubtitleTrack ServerEvent = "add-subtitle-track"
	ServerEventTerminate        ServerEvent = "terminate"
)

// OpenAndAwait opens the player and waits for the client to send the watch event.
func (p *NativePlayer) OpenAndAwait(clientId string, loadingState string) {
	p.sendPlayerEventTo(clientId, string(ServerEventOpenAndAwait), loadingState)
}

// Watch sends the watch event to the client.
func (p *NativePlayer) Watch(clientId string, playbackInfo *PlaybackInfo) {
	p.sendPlayerEventTo(clientId, string(ServerEventWatch), playbackInfo, true)
}

// SubtitleEvent sends the subtitle event to the client.
func (p *NativePlayer) SubtitleEvent(clientId string, event *mkvparser.SubtitleEvent) {
	p.sendPlayerEventTo(clientId, string(ServerEventSubtitleEvent), event, true)
}

// SetTracks sends the set tracks event to the client.
func (p *NativePlayer) SetTracks(clientId string, tracks []*mkvparser.TrackInfo) {
	p.sendPlayerEventTo(clientId, string(ServerEventSetTracks), tracks)
}

// Pause sends the pause event to the client.
func (p *NativePlayer) Pause(clientId string) {
	p.sendPlayerEventTo(clientId, string(ServerEventPause), nil)
}

// Resume sends the resume event to the client.
func (p *NativePlayer) Resume(clientId string) {
	p.sendPlayerEventTo(clientId, string(ServerEventResume), nil)
}

// Seek sends the seek event to the client.
func (p *NativePlayer) Seek(clientId string, time float64) {
	p.sendPlayerEventTo(clientId, string(ServerEventSeek), time)
}

// Error stops the playback and displays an error message.
func (p *NativePlayer) Error(clientId string, err error) {
	p.sendPlayerEventTo(clientId, string(ServerEventError), struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

// AddSubtitleTrack sends the subtitle track added event to the client.
func (p *NativePlayer) AddSubtitleTrack(clientId string, track *mkvparser.TrackInfo) {
	p.sendPlayerEventTo(clientId, string(ServerEventAddSubtitleTrack), track)
}

// Stop emits a VideoTerminatedEvent to all subscribers.
// It should only be called by a module.
func (p *NativePlayer) Stop() {
	p.logger.Debug().Msg("nativeplayer: Stopping playback, notifying subscribers")
	p.notifySubscribers(&VideoTerminatedEvent{
		BaseVideoEvent: BaseVideoEvent{ClientId: p.playbackStatus.ClientId},
	})
	p.sendPlayerEvent(string(ServerEventTerminate), nil)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Client Events
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ClientEvent string

const (
	PlayerEventVideoPaused          ClientEvent = "video-paused"
	PlayerEventVideoResumed         ClientEvent = "video-resumed"
	PlayerEventVideoCompleted       ClientEvent = "video-completed"
	PlayerEventVideoEnded           ClientEvent = "video-ended"
	PlayerEventVideoSeeked          ClientEvent = "video-seeked"
	PlayerEventVideoError           ClientEvent = "video-error"
	PlayerEventVideoLoadedMetadata  ClientEvent = "loaded-metadata"
	PlayerEventSubtitleFileUploaded ClientEvent = "subtitle-file-uploaded"
	PlayerEventVideoTerminated      ClientEvent = "video-terminated"
	PlayerEventVideoTimeUpdate      ClientEvent = "video-time-update"
)

type (
	// PlayerEvent is an event coming from the client player.
	PlayerEvent struct {
		ClientId string      `json:"clientId"`
		Type     ClientEvent `json:"type"`
		Payload  interface{} `json:"payload"`
	}
	VideoEvent interface {
		GetClientId() string
	}
	BaseVideoEvent struct {
		ClientId string `json:"clientId"`
	}
	VideoPausedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	VideoResumedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	VideoEndedEvent struct {
		BaseVideoEvent
		AutoNext bool `json:"autoNext"`
	}
	VideoErrorEvent struct {
		BaseVideoEvent
		Error string `json:"error"`
	}
	VideoSeekedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	VideoStatusEvent struct {
		BaseVideoEvent
		Status PlaybackStatus `json:"status"`
	}
	VideoLoadedMetadataEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	SubtitleFileUploadedEvent struct {
		BaseVideoEvent
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	VideoTerminatedEvent struct {
		BaseVideoEvent
	}
	VideoCompletedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
)

// Client event payloads
type (
	videoStartedPayload struct {
		Url         string  `json:"url"`
		Paused      bool    `json:"paused"`
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	videoPausedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	videoResumedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	videoLoadedMetadataPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	videoSeekedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	subtitleFileUploadedPayload struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	videoErrorPayload struct {
		Error string `json:"error"`
	}
	videoEndedPayload struct {
		AutoNext bool `json:"autoNext"`
	}
	videoTerminatedPayload struct {
	}
	videoTimeUpdatePayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	videoCompletedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
)

// listenToPlayerEvents listens to client events and notifies subscribers.
func (p *NativePlayer) listenToPlayerEvents() {
	// Start a goroutine to listen to native player events
	go func() {
		for {
			select {
			// Listen to native player events from the client
			case clientEvent := <-p.clientPlayerEventSubscriber.Channel:
				playerEvent := &PlayerEvent{}
				marshaled, _ := json.Marshal(clientEvent.Payload)
				// Unmarshal the player event
				if err := json.Unmarshal(marshaled, &playerEvent); err == nil {
					// Handle events
					switch playerEvent.Type {

					// case PlayerEventVideoStarted:
					// 	p.setPlaybackStatus(func() {
					// 		event := &videoStartedPayload{}
					// 		if err := playerEvent.UnmarshalAs(&event); err != nil {
					// 			p.notifySubscribers(&VideoStartedEvent{
					// 				BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
					// 			})
					// 		}
					// 	})
					case PlayerEventVideoPaused:
						payload := &videoPausedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
								p.playbackStatus.Paused = true
								p.playbackStatus.CurrentTime = payload.CurrentTime
								p.playbackStatus.Duration = payload.Duration
							})
							p.notifySubscribers(&VideoPausedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								CurrentTime:    payload.CurrentTime,
								Duration:       payload.Duration,
							})
							p.notifySubscribers(&VideoStatusEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								Status:         *p.playbackStatus,
							})
						}
					case PlayerEventVideoResumed:
						payload := &videoResumedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
								p.playbackStatus.Paused = false
								p.playbackStatus.CurrentTime = payload.CurrentTime
								p.playbackStatus.Duration = payload.Duration
							})
							p.notifySubscribers(&VideoResumedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								CurrentTime:    payload.CurrentTime,
								Duration:       payload.Duration,
							})
							p.notifySubscribers(&VideoStatusEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								Status:         *p.playbackStatus,
							})
						}
					case PlayerEventVideoCompleted:
						payload := &videoCompletedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
								p.playbackStatus.CurrentTime = payload.CurrentTime
								p.playbackStatus.Duration = payload.Duration
							})
							p.notifySubscribers(&VideoCompletedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								CurrentTime:    payload.CurrentTime,
								Duration:       payload.Duration,
							})
						}
					case PlayerEventVideoEnded:
						payload := &videoEndedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
							})
							p.notifySubscribers(&VideoEndedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								AutoNext:       payload.AutoNext,
							})
						}
					case PlayerEventVideoError:
						payload := &videoErrorPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
							})
							p.notifySubscribers(&VideoErrorEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								Error:          payload.Error,
							})
						}
					case PlayerEventVideoSeeked:
						payload := &videoSeekedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
							})
							if p.seekedEventCancelFunc != nil {
								p.seekedEventCancelFunc()
							}
							var ctx context.Context
							ctx, p.seekedEventCancelFunc = context.WithCancel(context.Background())
							// Debounce the event
							go func() {
								defer func() {
									if r := recover(); r != nil {
									}
									if p.seekedEventCancelFunc != nil {
										p.seekedEventCancelFunc()
										p.seekedEventCancelFunc = nil
									}
								}()
								select {
								case <-ctx.Done():
								case <-time.After(time.Millisecond * 150):
									p.setPlaybackStatus(func() {
										p.playbackStatus.CurrentTime = payload.CurrentTime
										p.playbackStatus.Duration = payload.Duration
									})
									p.notifySubscribers(&VideoSeekedEvent{
										BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
										CurrentTime:    payload.CurrentTime,
										Duration:       payload.Duration,
									})
									return
								}
							}()
						} else {
							// Log error: util.Logger.Error().Err(err).Msg("nativeplayer: Failed to unmarshal video seeked payload")
						}
					case PlayerEventVideoLoadedMetadata:
						payload := &videoLoadedMetadataPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
								p.playbackStatus.CurrentTime = payload.CurrentTime
								p.playbackStatus.Duration = payload.Duration
							})
							p.notifySubscribers(&VideoLoadedMetadataEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								CurrentTime:    payload.CurrentTime,
								Duration:       payload.Duration,
							})
						}
					case PlayerEventSubtitleFileUploaded:
						payload := &subtitleFileUploadedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
							})
							p.notifySubscribers(&SubtitleFileUploadedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								Filename:       payload.Filename,
								Content:        payload.Content,
							})
						}
					case PlayerEventVideoTerminated:
						p.setPlaybackStatus(func() {
							p.playbackStatus.ClientId = playerEvent.ClientId
						})
						p.notifySubscribers(&VideoTerminatedEvent{
							BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
						})
					case PlayerEventVideoTimeUpdate:
						payload := &videoTimeUpdatePayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.ClientId = playerEvent.ClientId
								p.playbackStatus.CurrentTime = payload.CurrentTime
								p.playbackStatus.Duration = payload.Duration
								p.playbackStatus.Paused = payload.Paused
							})
							p.notifySubscribers(&VideoStatusEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								Status:         *p.playbackStatus,
							})
						}
					}
				}
			}
		}
	}()
}

// Events returns the event channel for the subscriber.
func (s *Subscriber) Events() <-chan VideoEvent {
	return s.eventCh
}

func (e *PlayerEvent) UnmarshalAs(dest interface{}) error {
	marshaled, _ := json.Marshal(e.Payload)
	return json.Unmarshal(marshaled, dest)
}

func (e *BaseVideoEvent) GetClientId() string {
	return e.ClientId
}
