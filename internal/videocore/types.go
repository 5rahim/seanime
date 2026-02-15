package videocore

import (
	"encoding/json"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
)

type ClientEventType string

const (
	// Player is mounted, playback is about to start
	PlayerEventVideoLoaded ClientEventType = "video-loaded"
	// Player loaded metadata for playback
	PlayerEventVideoLoadedMetadata ClientEventType = "video-loaded-metadata"
	// Player is ready to play
	PlayerEventVideoCanPlay    ClientEventType = "video-can-play"
	PlayerEventVideoPaused     ClientEventType = "video-paused"
	PlayerEventVideoResumed    ClientEventType = "video-resumed"
	PlayerEventVideoStatus     ClientEventType = "video-status"
	PlayerEventVideoCompleted  ClientEventType = "video-completed"
	PlayerEventVideoFullscreen ClientEventType = "video-fullscreen"
	PlayerEventVideoPip        ClientEventType = "video-pip"
	// Subtitle track is selected
	PlayerEventVideoSubtitleTrack ClientEventType = "video-subtitle-track"
	// Caption track is selected
	PlayerEventMediaCaptionTrack ClientEventType = "video-media-caption-track"
	// Subtitle track content is sent
	PlayerEventVideoSubtitleTrackContent ClientEventType = "video-subtitle-track-content"
	// Anime4K option is changed
	PlayerEventAnime4K ClientEventType = "video-anime-4k"
	// Audio track is selected
	PlayerEventVideoAudioTrack ClientEventType = "video-audio-track"
	// Playback reached the end
	PlayerEventVideoEnded  ClientEventType = "video-ended"
	PlayerEventVideoSeeked ClientEventType = "video-seeked"
	PlayerEventVideoError  ClientEventType = "video-error"
	// Player unmounted (gracefully or fatal)
	PlayerEventVideoTerminated ClientEventType = "video-terminated"
	// Player sent type and playback info
	PlayerEventVideoPlaybackState ClientEventType = "video-playback-state"
	// Subtitle file was uploaded
	PlayerEventSubtitleFileUploaded ClientEventType = "subtitle-file-uploaded"
	PlayerEventVideoPlaylist        ClientEventType = "video-playlist"
	// Player sent all text tracks
	PlayerEventVideoTextTracks ClientEventType = "video-text-tracks"
	// Request to translate text
	PlayerEventTranslateText ClientEventType = "translate-text"
	// Request to translate subtitle file track
	PlayerEventTranslateSubtitleFileTrack ClientEventType = "translate-subtitle-file-track"
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
	Src               *string `json:"src"`
	Content           *string `json:"content"`
	Label             string  `json:"label"`
	Language          string  `json:"language"`
	Type              *string `json:"type"` // "srt" | "vtt" | "ass" | "ssa"
	Default           *bool   `json:"default"`
	UseLibassRenderer *bool   `json:"useLibassRenderer"`
}

type VideoLibassFont struct {
	Name *string `json:"name,omitempty"`
	Src  string  `json:"src"`
}

type VideoTextTrack struct {
	Number   int    `json:"number"`
	Type     string `json:"type"` // "subtitles" | "captions"
	Label    string `json:"label"`
	Language string `json:"language"`
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

type OnlinestreamParams struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	Provider      string `json:"provider"`
	Server        string `json:"server"`
	Quality       string `json:"quality"`
	Dubbed        bool   `json:"dubbed"`
}

// VideoPlaybackInfo contains detailed information about the currently played media.
// It is filled by the client, passed to the player and sent to the server during playback.
type VideoPlaybackInfo struct {
	Id           string       `json:"id"`
	PlaybackType PlaybackType `json:"playbackType"`
	StreamURL    string       `json:"streamUrl"`
	StreamPath   string       `json:"streamPath,omitempty"` // e.g. /anime/episode 01.mkv
	// MkvMetadata is only set for NativePlayer playbacks. Parsed by mkvparser.MetadataParser for directstream.Manager.
	MkvMetadata *mkvparser.Metadata `json:"mkvMetadata"` // NativePlayer only
	// LocalFile is only set for local file streams. NativePlayer
	LocalFile *anime.LocalFile `json:"localFile"`
	// Set by WebPlayer when online stream starts. Used for Nakama watch parties.
	OnlinestreamParams             *OnlinestreamParams   `json:"onlinestreamParams"`
	SubtitleTracks                 []*VideoSubtitleTrack `json:"subtitleTracks"`
	LibassFonts                    []*VideoLibassFont    `json:"libassFonts"`
	VideoSources                   []*VideoSource        `json:"videoSources"`
	SelectedVideoSource            *int                  `json:"selectedVideoSource"` // index of VideoSource
	PlaylistExternalEpisodeNumbers []int                 `json:"playlistExternalEpisodeNumbers"`
	DisableRestoreFromContinuity   *bool                 `json:"disableRestoreFromContinuity"`
	InitialState                   *VideoInitialState    `json:"initialState"`
	Media                          *anilist.BaseAnime    `json:"media"`
	Episode                        *anime.Episode        `json:"episode"`
	StreamType                     string                `json:"streamType"` // "native" | "hls" | "unknown"
	IsNakamaWatchParty             bool                  `json:"isNakamaWatchParty,omitempty"`
}

// VideoPlaylistState holds the state for the video player's playlist and playback.
type VideoPlaylistState struct {
	Type            PlaybackType     `json:"type"`
	Episodes        []*anime.Episode `json:"episodes"`
	PreviousEpisode *anime.Episode   `json:"previousEpisode,omitempty"`
	NextEpisode     *anime.Episode   `json:"nextEpisode,omitempty"`
	CurrentEpisode  *anime.Episode   `json:"currentEpisode"`
	AnimeEntry      *anime.Entry     `json:"animeEntry,omitempty"`
}

type (
	PlaybackStatus struct {
		Id          string  `json:"id"`
		ClientId    string  `json:"clientId"`
		Paused      bool    `json:"paused"`
		CurrentTime float64 `json:"currentTime"` // in seconds
		Duration    float64 `json:"duration"`    // in seconds
	}
	// PlaybackState is sent once when the video starts.
	PlaybackState struct {
		ClientId     string             `json:"clientId"`
		PlayerType   PlayerType         `json:"playerType"`
		PlaybackInfo *VideoPlaybackInfo `json:"playbackInfo"`
	}
	ClientEvent struct {
		ClientId string          `json:"clientId"`
		Type     ClientEventType `json:"type"`
		Payload  json.RawMessage `json:"payload"`
	}
)

// Client event payloads
type (
	clientSubtitleFileUploadedPayload struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	clientVideoLoadedPayload struct {
		State PlaybackState `json:"state"`
	}
	clientVideoPlaylistPayload struct {
		Playlist VideoPlaylistState `json:"playlist"`
	}
	clientVideoErrorPayload struct {
		Error string `json:"error"`
	}
	clientVideoEndedPayload struct {
		AutoNext bool `json:"autoNext"`
	}
	clientVideoStatusPayload struct {
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	clientVideoFullscreenPayload struct {
		Fullscreen bool `json:"fullscreen"`
	}
	clientVideoPipPayload struct {
		Pip bool `json:"pip"`
	}
	clientVideoSubtitleTrackPayload struct {
		TrackNumber int    `json:"trackNumber"`
		Kind        string `json:"kind"` // file | event
	}
	clientVideoSubtitleTrackContentPayload struct {
		TrackNumber int    `json:"trackNumber"`
		Content     string `json:"content"`
		Type        string `json:"type"`
	}
	clientVideoMediaCaptionTrackPayload struct {
		TrackIndex int `json:"trackIndex"`
	}
	clientVideoAudioTrackPayload struct {
		TrackNumber int  `json:"trackNumber"`
		IsHls       bool `json:"isHLS"`
	}
	clientVideoAnime4KPayload struct {
		Option string `json:"option"`
	}
	clientVideoTextTracksPayload struct {
		TextTracks []*VideoTextTrack `json:"textTracks"`
	}
	clientTranslateTextPayload struct {
		Text string `json:"text"`
	}
)

func (e *ClientEvent) UnmarshalAs(dest interface{}) error {
	return json.Unmarshal(e.Payload, dest)
}

func (e *BaseVideoEvent) GetPlaybackId() string {
	return e.PlaybackId
}
func (e *BaseVideoEvent) GetClientId() string {
	return e.ClientId
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// VideoEvent is an event coming from the NativePlayer or WebPlayer.
// This interface is used by the backend modules.
type VideoEvent interface {
	IsWebPlayer() bool
	IsNativePlayer() bool
	IsOnlinestream() bool
	IsTorrent() bool
	IsNakama() bool
	IsDebrid() bool
	GetPlayerType() PlayerType
	GetPlaybackType() PlaybackType
	GetPlaybackId() string
	GetClientId() string
	IsCritical() bool
	identify(id string, clientId string, playerType PlayerType, playbackType PlaybackType)
}

type BaseVideoEvent struct {
	PlayerType   PlayerType   `json:"playerType"`
	PlaybackType PlaybackType `json:"playbackType"`
	PlaybackId   string       `json:"playbackId"`
	ClientId     string       `json:"clientId"`
}

func (e *BaseVideoEvent) GetPlayerType() PlayerType     { return e.PlayerType }
func (e *BaseVideoEvent) GetPlaybackType() PlaybackType { return e.PlaybackType }
func (e *BaseVideoEvent) IsNativePlayer() bool          { return e.PlayerType == NativePlayer }
func (e *BaseVideoEvent) IsWebPlayer() bool             { return e.PlayerType == WebPlayer }
func (e *BaseVideoEvent) IsOnlinestream() bool          { return e.PlaybackType == PlaybackTypeOnlinestream }
func (e *BaseVideoEvent) IsTorrent() bool               { return e.PlaybackType == PlaybackTypeTorrent }
func (e *BaseVideoEvent) IsNakama() bool                { return e.PlaybackType == PlaybackTypeNakama }
func (e *BaseVideoEvent) IsDebrid() bool                { return e.PlaybackType == PlaybackTypeDebrid }
func (e *BaseVideoEvent) IsCritical() bool              { return true }
func (e *BaseVideoEvent) identify(id string, clientId string, playerType PlayerType, playbackType PlaybackType) {
	e.PlaybackId = id
	e.ClientId = clientId
	e.PlayerType = playerType
	e.PlaybackType = playbackType
}

type (
	VideoLoadedEvent struct {
		BaseVideoEvent
		ClientId string        `json:"clientId"`
		State    PlaybackState `json:"state"`
	}
	VideoPlaybackStateEvent struct {
		BaseVideoEvent
		ClientId string        `json:"clientId"`
		State    PlaybackState `json:"state"`
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
		Paused      bool    `json:"paused"`
	}
	VideoStatusEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	VideoLoadedMetadataEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
	}
	VideoCanPlayEvent struct {
		BaseVideoEvent
		CurrentTime float64 `json:"currentTime"`
		Duration    float64 `json:"duration"`
		Paused      bool    `json:"paused"`
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
	VideoAudioTrackEvent struct {
		BaseVideoEvent
		TrackNumber int  `json:"trackNumber"`
		IsHls       bool `json:"isHLS"`
	}
	VideoSubtitleTrackEvent struct {
		BaseVideoEvent
		TrackNumber int    `json:"trackNumber"`
		Kind        string `json:"kind"` // "file" | "event"
	}
	VideoSubtitleTrackContentEvent struct {
		BaseVideoEvent
		TrackNumber int    `json:"trackNumber"`
		Content     string `json:"content"`
		Type        string `json:"type"`
	}
	VideoMediaCaptionTrackEvent struct {
		BaseVideoEvent
		TrackIndex int `json:"trackIndex"`
	}
	VideoFullscreenEvent struct {
		BaseVideoEvent
		Fullscreen bool `json:"fullscreen"`
	}
	VideoPipEvent struct {
		BaseVideoEvent
		Pip bool `json:"pip"`
	}
	VideoAnime4KEvent struct {
		BaseVideoEvent
		Option string `json:"string"` // name or "off"
	}
	VideoPlaylistEvent struct {
		BaseVideoEvent
		Playlist *VideoPlaylistState `json:"playlist"`
	}
	VideoTextTracksEvent struct {
		BaseVideoEvent
		TextTracks []*VideoTextTrack `json:"textTracks"`
	}
)

func (e *VideoStatusEvent) IsCritical() bool     { return false }
func (e *VideoTimeUpdateEvent) IsCritical() bool { return false }

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ServerEvent string

const (
	ServerEventPause                       ServerEvent = "pause"
	ServerEventResume                      ServerEvent = "resume"
	ServerEventSeek                        ServerEvent = "seek"
	ServerEventSeekTo                      ServerEvent = "seek-to"
	ServerEventSetFullscreen               ServerEvent = "set-fullscreen"
	ServerEventSetPip                      ServerEvent = "set-pip"
	ServerEventSetSubtitleTrack            ServerEvent = "set-subtitle-track"
	ServerEventAddSubtitleTrack            ServerEvent = "add-subtitle-track"
	ServerEventAddExternalSubtitleTrack    ServerEvent = "add-external-subtitle-track"
	ServerEventSetMediaCaptionTrack        ServerEvent = "set-media-caption-track"
	ServerEventAddMediaCaptionTrack        ServerEvent = "add-media-caption-track"
	ServerEventSetAudioTrack               ServerEvent = "set-audio-track"
	ServerEventTerminate                   ServerEvent = "terminate"
	ServerEventStartOnlinestreamWatchParty ServerEvent = "start-onlinestream-watch-party"
	ServerEventGetStatus                   ServerEvent = "get-status"
	ServerEventShowMessage                 ServerEvent = "show-message"
	ServerEventPlayPlaylistEpisode         ServerEvent = "play-playlist-episode"
	ServerEventGetTextTracks               ServerEvent = "get-text-tracks"
	ServerEventRequestPlayEpisode          ServerEvent = "request-play-episode"
	ServerEventTranslatedText              ServerEvent = "translated-text"
	ServerEventInSightData                 ServerEvent = "in-sight-data"
	// State requests
	ServerEventGetFullscreen           ServerEvent = "get-fullscreen"
	ServerEventGetPip                  ServerEvent = "get-pip"
	ServerEventGetAnime4K              ServerEvent = "get-anime-4k"
	ServerEventGetSubtitleTrack        ServerEvent = "get-subtitle-track"
	ServerEventGetSubtitleTrackContent ServerEvent = "get-subtitle-track-content"
	ServerEventGetAudioTrack           ServerEvent = "get-audio-track"
	ServerEventGetMediaCaptionTrack    ServerEvent = "get-media-caption-track"
	ServerEventGetPlaybackState        ServerEvent = "get-playback-state"
	ServerEventGetPlaylist             ServerEvent = "get-playlist"
)
