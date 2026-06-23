package player

import (
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"

	"github.com/samber/mo"
)

type Target string

const (
	TargetVideoCore Target = "videocore"
	TargetMpvCore   Target = "mpvcore"
)

type Renderer string

const (
	RendererWeb    Renderer = "web"
	RendererNative Renderer = "native"
	RendererMpv    Renderer = "mpv"
)

type PlaybackType string

const (
	PlaybackTypeLocalFile    PlaybackType = "localfile"
	PlaybackTypeTorrent      PlaybackType = "torrent"
	PlaybackTypeDebrid       PlaybackType = "debrid"
	PlaybackTypeNakama       PlaybackType = "nakama"
	PlaybackTypeOnlinestream PlaybackType = "onlinestream"
	PlaybackTypeURL          PlaybackType = "url"
)

type SessionKey struct {
	Target     Target
	ClientID   string
	PlaybackID string
}

type SubtitleTrack struct {
	Index     int     `json:"index"`
	URI       *string `json:"uri,omitempty"`
	SourceURL *string `json:"sourceUrl,omitempty"`
	Content   *string `json:"content,omitempty"`
	Label     string  `json:"label"`
	Language  string  `json:"language"`
	Format    *string `json:"format,omitempty"`
	Default   *bool   `json:"default,omitempty"`

	// Compatibility fields
	Src               *string `json:"src,omitempty"`
	Type              *string `json:"type,omitempty"`
	UseLibassRenderer *bool   `json:"useLibassRenderer,omitempty"`
}

type VideoSource struct {
	Index      int     `json:"index"`
	Resolution string  `json:"resolution"`
	URL        *string `json:"url,omitempty"`
	Label      *string `json:"label,omitempty"`
	MoreInfo   *string `json:"moreInfo,omitempty"`
}

type InitialState struct {
	CurrentTime *float64 `json:"currentTime,omitempty"`
	Paused      *bool    `json:"paused,omitempty"`
}

type SkipInterval struct {
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
}

type SkipDataEntry struct {
	Interval SkipInterval `json:"interval"`
}

type OnlinestreamParams struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	Provider      string `json:"provider"`
	Server        string `json:"server"`
	Quality       string `json:"quality"`
	Dubbed        bool   `json:"dubbed"`
}

type SkipData struct {
	Op *SkipDataEntry `json:"op,omitempty"`
	Ed *SkipDataEntry `json:"ed,omitempty"`
}

type LibassFont struct {
	Name *string `json:"name,omitempty"`
	Src  string  `json:"src"`
}

type PlaybackInfo struct {
	ID                             string                               `json:"id"`
	Target                         Target                               `json:"target"`
	Renderer                       Renderer                             `json:"renderer"`
	PlaybackType                   PlaybackType                         `json:"playbackType"`
	PlaybackURI                    string                               `json:"playbackUri,omitempty"`
	StreamURL                      string                               `json:"streamUrl"`
	StreamPath                     string                               `json:"streamPath,omitempty"`
	MimeType                       string                               `json:"mimeType,omitempty"`
	ContentLength                  int64                                `json:"contentLength,omitempty"`
	MkvMetadata                    *mkvparser.Metadata                  `json:"mkvMetadata,omitempty"`
	SubtitleTracks                 []*SubtitleTrack                     `json:"subtitleTracks,omitempty"`
	VideoSources                   []*VideoSource                       `json:"videoSources,omitempty"`
	SelectedVideoSource            *int                                 `json:"selectedVideoSource,omitempty"`
	PlaylistExternalEpisodeNumbers []int                                `json:"playlistExternalEpisodeNumbers,omitempty"`
	DisableRestoreFromContinuity   *bool                                `json:"disableRestoreFromContinuity,omitempty"`
	InitialState                   *InitialState                        `json:"initialState,omitempty"`
	EntryListData                  *anime.EntryListData                 `json:"entryListData,omitempty"`
	Media                          *anilist.BaseAnime                   `json:"media"`
	Episode                        *anime.Episode                       `json:"episode"`
	LocalFile                      *anime.LocalFile                     `json:"localFile,omitempty"`
	OnlinestreamParams             *OnlinestreamParams                  `json:"onlinestreamParams,omitempty"`
	IsNakamaWatchParty             bool                                 `json:"isNakamaWatchParty,omitempty"`
	MkvMetadataParser              mo.Option[*mkvparser.MetadataParser] `json:"-"`

	// Compatibility fields
	StreamType  string        `json:"streamType,omitempty"`
	LibassFonts []*LibassFont `json:"libassFonts,omitempty"`
}

type PlaybackState struct {
	ClientID     string        `json:"clientId"`
	PlaybackInfo *PlaybackInfo `json:"playbackInfo"`

	// Compatibility fields
	PlayerType      string `json:"playerType,omitempty"`
	CurrentProgress int    `json:"currentProgress,omitempty"`
}

type PlaybackStatus struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"clientId"`
	Paused      bool    `json:"paused"`
	CurrentTime float64 `json:"currentTime"`
	Duration    float64 `json:"duration"`
}

type PlaylistState struct {
	Type            PlaybackType     `json:"type"`
	Episodes        []*anime.Episode `json:"episodes"`
	PreviousEpisode *anime.Episode   `json:"previousEpisode,omitempty"`
	NextEpisode     *anime.Episode   `json:"nextEpisode,omitempty"`
	CurrentEpisode  *anime.Episode   `json:"currentEpisode"`
	AnimeEntry      *anime.Entry     `json:"animeEntry,omitempty"`
}

type Event interface {
	GetSessionKey() SessionKey
	IsCritical() bool
}

type BaseEvent struct {
	Session SessionKey
}

func (e *BaseEvent) GetSessionKey() SessionKey { return e.Session }
func (e *BaseEvent) IsCritical() bool          { return true }

type (
	PlaybackLoadedEvent struct {
		BaseEvent
		State PlaybackState
	}
	LoadedMetadataEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
		Paused      bool
	}
	CanPlayEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
		Paused      bool
	}
	PausedEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
	}
	ResumedEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
	}
	StatusEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
		Paused      bool
	}
	SeekedEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
		Paused      bool
	}
	CompletedEvent struct {
		BaseEvent
		CurrentTime float64
		Duration    float64
	}
	EndedEvent struct {
		BaseEvent
		AutoNext bool
	}
	ErrorEvent struct {
		BaseEvent
		Error string
	}
	TerminatedEvent struct {
		BaseEvent
	}
	FullscreenChangedEvent struct {
		BaseEvent
		Fullscreen bool
	}
	PipChangedEvent struct {
		BaseEvent
		Pip bool
	}
	AudioTrackChangedEvent struct {
		BaseEvent
		TrackID interface{}
	}
	SubtitleTrackChangedEvent struct {
		BaseEvent
		TrackID interface{}
	}
	SubtitleFileUploadedEvent struct {
		BaseEvent
		Filename string
		Content  string
	}
	PlaylistStateEvent struct {
		BaseEvent
		Playlist *PlaylistState
	}
	SkipDataEvent struct {
		BaseEvent
		SkipData *SkipData
	}
)

func (e *StatusEvent) IsCritical() bool { return false }

type CommandType string

const (
	CommandPause                       CommandType = "pause"
	CommandResume                      CommandType = "resume"
	CommandSeek                        CommandType = "seek"
	CommandSeekTo                      CommandType = "seek-to"
	CommandSetFullscreen               CommandType = "set-fullscreen"
	CommandSetPip                      CommandType = "set-pip"
	CommandSetAudioTrack               CommandType = "set-audio-track"
	CommandSetSubtitleTrack            CommandType = "set-subtitle-track"
	CommandAddSubtitleTrack            CommandType = "add-subtitle-track"
	CommandAddExternalSubtitleTrack    CommandType = "add-external-subtitle-track"
	CommandSetMediaCaptionTrack        CommandType = "set-media-caption-track"
	CommandAddMediaCaptionTrack        CommandType = "add-media-caption-track"
	CommandPlayPlaylistEpisode         CommandType = "play-playlist-episode"
	CommandShowMessage                 CommandType = "show-message"
	CommandSetSkipData                 CommandType = "set-skip-data"
	CommandClearSkipData               CommandType = "clear-skip-data"
	CommandStartOnlinestreamWatchParty CommandType = "start-onlinestream-watch-party"
)

type ShowMessagePayload struct {
	Message  string
	Duration int
}

type Command struct {
	Type    CommandType
	Payload interface{}
}
