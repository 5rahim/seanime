package nativeplayer

import (
	"seanime/internal/mediastream/mkvparser"

	"github.com/goccy/go-json"
)

type ServerEvent string

const (
	ServerEventOpenAndAwait  ServerEvent = "open-and-await"
	ServerEventWatch         ServerEvent = "watch"
	ServerEventSubtitleEvent ServerEvent = "subtitle-event"
	ServerEventSetTracks     ServerEvent = "set-tracks"
	ServerEventPause         ServerEvent = "pause"
	ServerEventResume        ServerEvent = "resume"
	ServerEventSeek          ServerEvent = "seek"
	ServerEventError         ServerEvent = "error"
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Client Events
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ClientEvent string

const (
	PlayerEventCanPlay         ClientEvent = "can-play"
	PlayerEventVideoStarted    ClientEvent = "video-started"
	PlayerEventVideoPaused     ClientEvent = "video-paused"
	PlayerEventVideoResumed    ClientEvent = "video-resumed"
	PlayerEventVideoEnded      ClientEvent = "video-ended"
	PlayerEventVideoSeeked     ClientEvent = "video-seeked"
	PlayerEventVideoError      ClientEvent = "video-error"
	PlayerEventVideoTimeUpdate ClientEvent = "video-time-update"
	PlayerEventVideoMetadata   ClientEvent = "video-metadata"
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
	VideoStartedEvent struct {
		BaseVideoEvent
	}
	VideoPausedEvent struct {
		BaseVideoEvent
	}
	VideoResumedEvent struct {
		BaseVideoEvent
	}
	VideoEndedEvent struct {
		BaseVideoEvent
	}
	VideoSeekedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
	}
	VideoTimeUpdateEvent struct {
		BaseVideoEvent
	}
	VideoStatusEvent struct {
		BaseVideoEvent
		Status PlaybackStatus `json:"status"`
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

	videoSeekedPayload struct {
		CurrentTime float64 `json:"currentTime"`
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
					case PlayerEventCanPlay:

					case PlayerEventVideoStarted:
						p.setPlaybackStatus(func() {
							event := &videoStartedPayload{}
							if err := playerEvent.UnmarshalAs(&event); err != nil {
								p.NotifySubscribers(&VideoStartedEvent{
									BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								})
							}
						})
					case PlayerEventVideoPaused:
						p.setPlaybackStatus(func() {
							p.playbackStatus.Paused = true
						})
					case PlayerEventVideoResumed:
						p.setPlaybackStatus(func() {
							p.playbackStatus.Paused = false
						})
					case PlayerEventVideoEnded:
						p.setPlaybackStatus(func() {
							p.playbackStatus = &PlaybackStatus{}
						})
					case PlayerEventVideoSeeked:
						payload := &videoSeekedPayload{}
						if err := playerEvent.UnmarshalAs(&payload); err == nil {
							p.setPlaybackStatus(func() {
								p.playbackStatus.CurrentTime = payload.CurrentTime
							})
							p.NotifySubscribers(&VideoSeekedEvent{
								BaseVideoEvent: BaseVideoEvent{ClientId: playerEvent.ClientId},
								CurrentTime:    payload.CurrentTime,
							})
						} else {
							// Log error: util.Logger.Error().Err(err).Msg("nativeplayer: Failed to unmarshal video seeked payload")
						}
					case PlayerEventVideoError:
					case PlayerEventVideoTimeUpdate:
					case PlayerEventVideoMetadata:
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
