package nativeplayer

import (
	"seanime/internal/util"

	"github.com/goccy/go-json"
)

type ServerEvent string

var (
	ServerEventWatch          ServerEvent = "watch"
	ServerEventMediaContainer ServerEvent = "media-container"
	ServerEventSubtitleEvent  ServerEvent = "subtitle-event"
	ServerEventSetTracks      ServerEvent = "set-tracks"
	ServerEventPause          ServerEvent = "pause"
	ServerEventResume         ServerEvent = "resume"
	ServerEventSeek           ServerEvent = "seek"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Client Events
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	ClientPlayerEventCanPlay         = "can-play"
	ClientPlayerEventVideoStarted    = "video-started"
	ClientPlayerEventVideoPaused     = "video-paused"
	ClientPlayerEventVideoResumed    = "video-resumed"
	ClientPlayerEventVideoEnded      = "video-ended"
	ClientPlayerEventVideoSeeked     = "video-seeked"
	ClientPlayerEventVideoError      = "video-error"
	ClientPlayerEventVideoTimeUpdate = "video-time-update"
	ClientPlayerEventVideoMetadata   = "video-metadata"
)

type (
	ClientEvent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}

	VideoEvent interface {
		Unmarshal(dest interface{}) error
	}
	VideoStartedEvent struct {
	}
	VideoPausedEvent struct {
	}
	VideoResumedEvent struct {
	}
	VideoEndedEvent struct {
	}
	VideoSeekedEvent struct {
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
)

// listenToClientEvents listens to client events and notifies subscribers.
func (p *NativePlayer) listenToClientEvents() {
	// Start a goroutine to listen to native player events
	go func() {
		for {
			select {
			// Listen to native player events from the client
			case clientEvent := <-p.clientPlayerEventSubscriber.Channel:
				playerEvent := ClientEvent{}
				marshaled, _ := json.Marshal(clientEvent.Payload)
				// Unmarshal the player event
				if err := json.Unmarshal(marshaled, playerEvent); err != nil {
					util.Spew(playerEvent) // todo remove
					// Handle events
					switch playerEvent.Type {
					case ClientPlayerEventCanPlay:
					case ClientPlayerEventVideoStarted:

						p.setPlaybackStatus(func() {
							event := &videoStartedPayload{}
							if err := playerEvent.UnmarshalAs(&event); err != nil {

							}
						})
					case ClientPlayerEventVideoPaused:
						p.setPlaybackStatus(func() {
							p.playbackStatus.Paused = true
						})
					case ClientPlayerEventVideoResumed:
						p.setPlaybackStatus(func() {
							p.playbackStatus.Paused = false
						})
					case ClientPlayerEventVideoEnded:
						p.setPlaybackStatus(func() {
							p.playbackStatus = &PlaybackStatus{}
						})
					case ClientPlayerEventVideoSeeked:
					case ClientPlayerEventVideoError:
					case ClientPlayerEventVideoTimeUpdate:
					case ClientPlayerEventVideoMetadata:
					}
				}
			}
		}
	}()
}

// Events returns the event channel for the subscriber.
func (s *Subscriber) Events() <-chan interface{} {
	return s.eventCh
}

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	marshaled, _ := json.Marshal(e.Payload)
	return json.Unmarshal(marshaled, dest)
}
