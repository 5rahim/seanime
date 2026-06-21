package mpvcore

import (
	"encoding/json"
	"seanime/internal/player"
)

type PlaybackType = player.PlaybackType

const (
	PlaybackTypeLocalFile = player.PlaybackTypeLocalFile
	PlaybackTypeTorrent   = player.PlaybackTypeTorrent
	PlaybackTypeDebrid    = player.PlaybackTypeDebrid
	PlaybackTypeNakama    = player.PlaybackTypeNakama
	PlaybackTypeURL       = player.PlaybackTypeURL
)

type StreamType = PlaybackType

const (
	StreamTypeFile    StreamType = player.PlaybackTypeLocalFile
	StreamTypeTorrent StreamType = player.PlaybackTypeTorrent
	StreamTypeDebrid  StreamType = player.PlaybackTypeDebrid
	StreamTypeNakama  StreamType = player.PlaybackTypeNakama
	StreamTypeURL     StreamType = player.PlaybackTypeURL
)

type SubtitleTrack = player.SubtitleTrack
type VideoSource = player.VideoSource
type InitialState = player.InitialState
type SkipInterval = player.SkipInterval
type SkipDataEntry = player.SkipDataEntry
type SkipData = player.SkipData
type PlaybackInfo = player.PlaybackInfo
type PlaybackState = player.PlaybackState
type PlaybackStatus = player.PlaybackStatus
type PlaylistState = player.PlaylistState

type ClientEventType string

const (
	ClientEventPlaybackLoaded       ClientEventType = "playback-loaded"
	ClientEventLoadedMetadata       ClientEventType = "loaded-metadata"
	ClientEventCanPlay              ClientEventType = "can-play"
	ClientEventPaused               ClientEventType = "paused"
	ClientEventResumed              ClientEventType = "resumed"
	ClientEventStatus               ClientEventType = "status"
	ClientEventSeeked               ClientEventType = "seeked"
	ClientEventCompleted            ClientEventType = "completed"
	ClientEventEnded                ClientEventType = "ended"
	ClientEventPlayerError          ClientEventType = "player-error"
	ClientEventTerminated           ClientEventType = "terminated"
	ClientEventFullscreenChanged    ClientEventType = "fullscreen-changed"
	ClientEventPipChanged           ClientEventType = "pip-changed"
	ClientEventAudioTrackChanged    ClientEventType = "audio-track-changed"
	ClientEventSubtitleTrackChanged ClientEventType = "subtitle-track-changed"
	ClientEventPlaylistState        ClientEventType = "playlist-state"
	ClientEventSkipData             ClientEventType = "skip-data"
)

type ServerEvent string

const (
	ServerEventOpenAndAwait      ServerEvent = "open-and-await"
	ServerEventAbortOpen         ServerEvent = "abort-open"
	ServerEventWatch             ServerEvent = "watch"
	ServerEventStreamError       ServerEvent = "stream-error"
	ServerEventPause             ServerEvent = "pause"
	ServerEventResume            ServerEvent = "resume"
	ServerEventSeek              ServerEvent = "seek"
	ServerEventSeekTo            ServerEvent = "seek-to"
	ServerEventTerminate         ServerEvent = "terminate"
	ServerEventSetFullscreen     ServerEvent = "set-fullscreen"
	ServerEventSetPip            ServerEvent = "set-pip"
	ServerEventSetAudioTrack     ServerEvent = "set-audio-track"
	ServerEventSetSubtitleTrack  ServerEvent = "set-subtitle-track"
	ServerEventAddSubtitleTrack  ServerEvent = "add-subtitle-track"
	ServerEventShowMessage       ServerEvent = "show-message"
	ServerEventGetStatus         ServerEvent = "get-status"
	ServerEventGetPlaylist       ServerEvent = "get-playlist"
	ServerEventGetSkipData       ServerEvent = "get-skip-data"
	ServerEventSetSkipData       ServerEvent = "set-skip-data"
	ServerEventPlayPlaylistEntry ServerEvent = "play-playlist-episode"
	ServerEventInSightData       ServerEvent = "in-sight-data"
)

type ClientEvent struct {
	ClientID string          `json:"clientId"`
	Type     ClientEventType `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	return json.Unmarshal(e.Payload, dest)
}

type statusPayload struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"clientId"`
	CurrentTime float64 `json:"currentTime"`
	Duration    float64 `json:"duration"`
	Paused      bool    `json:"paused"`
}

type playbackLoadedPayload struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}

type endedPayload struct {
	AutoNext bool `json:"autoNext"`
}

type errorPayload struct {
	Error string `json:"error"`
}

type terminatedPayload struct {
	ID           string       `json:"id"`
	ClientID     string       `json:"clientId"`
	PlaybackType PlaybackType `json:"playbackType"`
}

type trackChangedPayload struct {
	TrackID interface{} `json:"trackId"`
}

type playlistPayload struct {
	Playlist *PlaylistState `json:"playlist"`
}

type skipDataPayload struct {
	SkipData *SkipData `json:"skipData"`
}

type VideoEvent interface {
	GetPlaybackType() PlaybackType
	GetPlaybackID() string
	GetPlaybackId() string
	GetClientID() string
	GetClientId() string
	IsCritical() bool
	identify(id string, clientID string, playbackType PlaybackType)
}

type BaseVideoEvent struct {
	PlaybackType PlaybackType `json:"playbackType"`
	PlaybackID   string       `json:"playbackId"`
	ClientID     string       `json:"clientId"`
}

func (e *BaseVideoEvent) GetPlaybackType() PlaybackType { return e.PlaybackType }
func (e *BaseVideoEvent) GetPlaybackID() string         { return e.PlaybackID }
func (e *BaseVideoEvent) GetPlaybackId() string         { return e.PlaybackID }
func (e *BaseVideoEvent) GetClientID() string           { return e.ClientID }
func (e *BaseVideoEvent) GetClientId() string           { return e.ClientID }
func (e *BaseVideoEvent) IsCritical() bool              { return true }
func (e *BaseVideoEvent) identify(id string, clientID string, playbackType PlaybackType) {
	e.PlaybackID = id
	e.ClientID = clientID
	e.PlaybackType = playbackType
}

type (
	PlaybackLoadedEvent struct {
		BaseVideoEvent
		State PlaybackState `json:"state"`
	}
	LoadedMetadataEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	CanPlayEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	PausedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	ResumedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	StatusEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	SeekedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	CompletedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	EndedEvent struct {
		BaseVideoEvent
		AutoNext bool `json:"autoNext"`
	}
	ErrorEvent struct {
		BaseVideoEvent
		Error string `json:"error"`
	}
	TerminatedEvent        struct{ BaseVideoEvent }
	FullscreenChangedEvent struct {
		BaseVideoEvent
		Fullscreen bool `json:"fullscreen"`
	}
	PipChangedEvent struct {
		BaseVideoEvent
		Pip bool `json:"pip"`
	}
	AudioTrackChangedEvent struct {
		BaseVideoEvent
		TrackID interface{} `json:"trackId"`
	}
	SubtitleTrackChangedEvent struct {
		BaseVideoEvent
		TrackID interface{} `json:"trackId"`
	}
	PlaylistStateEvent struct {
		BaseVideoEvent
		Playlist *PlaylistState `json:"playlist"`
	}
	SkipDataEvent struct {
		BaseVideoEvent
		SkipData *SkipData `json:"skipData"`
	}
)

func (e *StatusEvent) IsCritical() bool { return false }
