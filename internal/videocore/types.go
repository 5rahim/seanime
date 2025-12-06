package videocore

import (
	"encoding/json"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
)

type ClientEventType string

const (
	PlayerEventVideoPaused          ClientEventType = "video-paused"
	PlayerEventVideoResumed         ClientEventType = "video-resumed"
	PlayerEventVideoCompleted       ClientEventType = "video-completed"
	PlayerEventVideoEnded           ClientEventType = "video-ended"
	PlayerEventVideoSeeked          ClientEventType = "video-seeked"
	PlayerEventVideoError           ClientEventType = "video-error"
	PlayerEventVideoLoadedMetadata  ClientEventType = "loaded-metadata" // Acts as PlayerEventVideoStarted
	PlayerEventSubtitleFileUploaded ClientEventType = "subtitle-file-uploaded"
	PlayerEventVideoTerminated      ClientEventType = "video-terminated"
	PlayerEventVideoTimeUpdate      ClientEventType = "video-time-update"
)

type PlayerType string

const (
	NativePlayer PlayerType = "native"
	WebPlayer    PlayerType = "web"
)

// PlaybackType is the playback method.
type PlaybackType string

const (
	PlaybackTypeLocalFile    PlaybackType = "localfile"    // NativePlayer only
	PlaybackTypeTorrent      PlaybackType = "torrent"      // NativePlayer only
	PlaybackTypeDebrid       PlaybackType = "debrid"       // NativePlayer only
	PlaybackTypeNakama       PlaybackType = "nakama"       // NativePlayer only
	PlaybackTypeOnlinestream PlaybackType = "onlinestream" // WebPlayer only
)

// VideoSubtitleTrack is an external subtitle track.
type VideoSubtitleTrack struct {
	Index             int     `json:"index"`
	Src               string  `json:"src"`
	Label             string  `json:"label"`
	Language          string  `json:"language"`
	Type              *string `json:"type"` // "srt" | "vtt" | "ass" | "ssa"
	Default           *bool   `json:"default"`
	UseLibassRenderer *bool   `json:"useLibassRenderer"`
}

// VideoSource is an alternative video stream source (e.g., resolution options).
type VideoSource struct {
	Index      int     `json:"index"`
	Resolution string  `json:"resolution"`
	URL        *string `json:"url"`
	Label      *string `json:"label"`
	MoreInfo   *string `json:"moreInfo"`
}

// VideoInitialState specifies the initial state for the player.
type VideoInitialState struct {
	CurrentTime *float64 `json:"currentTime"`
	Paused      *bool    `json:"paused"`
}

// VideoPlaybackInfo contains detailed information about the currently played media.
type VideoPlaybackInfo struct {
	Id           string       `json:"id"`
	PlaybackType PlaybackType `json:"playbackType"`
	StreamURL    string       `json:"streamUrl"`
	// MkvMetadata is only set for NativePlayer playbacks. Parsed by mkvparser.MetadataParser for directstream.Manager.
	MkvMetadata                    *mkvparser.Metadata   `json:"mkvMetadata"` // NativePlayer only
	SubtitleTracks                 []*VideoSubtitleTrack `json:"subtitleTracks"`
	VideoSources                   []*VideoSource        `json:"videoSources"`
	SelectedVideoSource            *int                  `json:"selectedVideoSource"` // index of VideoSource
	PlaylistExternalEpisodeNumbers []int                 `json:"playlistExternalEpisodeNumbers"`
	DisableRestoreFromContinuity   *bool                 `json:"disableRestoreFromContinuity"`
	EnableDiscordRichPresence      *bool                 `json:"enableDiscordRichPresence"`
	InitialState                   *VideoInitialState    `json:"initialState"`
	TrackContinuity                *bool                 `json:"trackContinuity"`
	Media                          *anilist.BaseAnime    `json:"media"`
	Episode                        *anime.Episode        `json:"episode"`
	StreamType                     string                `json:"streamType"` // "native" | "hls" | "unknown"
}

type (
	PlaybackStatus struct {
		Id          string  `json:"id"`
		ClientId    string  `json:"clientId"`
		Paused      bool    `json:"paused"`
		CurrentTime float64 `json:"currentTime"` // in seconds
		Duration    float64 `json:"duration"`    // in seconds
		Fullscreen  bool    `json:"fullscreen"`
	}
	// PlaybackState is sent once when the video starts.
	PlaybackState struct {
		PlayerType      PlayerType         `json:"playerType"`
		PlaybackInfo    *VideoPlaybackInfo `json:"playbackInfo"`
		CurrentProgress int                `json:"currentProgress"`
	}
	ClientEvent struct {
		ClientId string          `json:"clientId"`
		Type     ClientEventType `json:"type"`
		Payload  json.RawMessage `json:"payload"`
	}
)

// Client event payloads
type (
	clientPausedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	clientResumedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	clientLoadedMetadataPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	clientSeekedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	clientSubtitleFileUploadedPayload struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	clientErrorPayload struct {
		Error string `json:"error"`
	}
	clientEndedPayload struct {
		AutoNext bool `json:"autoNext"`
	}
	clientTerminatedPayload struct {
	}
	clientTimeUpdatePayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	clientCompletedPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
)

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	return json.Unmarshal(e.Payload, dest)
}

func (e *BaseVideoEvent) GetId() string {
	return e.Id
}
func (e *BaseVideoEvent) GetClientId() string {
	return e.ClientId
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// VideoEvent is an event coming from the NativePlayer or WebPlayer.
// This interface is used by the backend modules.
type VideoEvent interface {
	GetId() string
	GetClientId() string
	IsCritical() bool
}

type BaseVideoEvent struct {
	Id       string `json:"id"`
	ClientId string `json:"clientId"`
}

func (e *BaseVideoEvent) IsCritical() bool { return true }

type (
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
	// VideoTerminatedEvent is sent when the video playback is terminated.
	// For the Native Player, this happens when the user closes the player.
	// For the Web Player, this happens when the video player unmounts (user navigates away from the page).
	VideoTerminatedEvent struct {
		BaseVideoEvent
	}
	VideoCompletedEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
	}
	VideoTimeUpdateEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
	}
)

func (e *VideoStatusEvent) IsCritical() bool     { return false }
func (e *VideoTimeUpdateEvent) IsCritical() bool { return false }
