package nativeplayer

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type StreamType string

const (
	StreamTypeTorrent StreamType = "torrent"
	StreamTypeFile    StreamType = "localfile"
	StreamTypeDebrid  StreamType = "debrid"
	StreamTypeURL     StreamType = "url"
	StreamTypeNakama  StreamType = "nakama"
)

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

type VideoCoreDelegate interface {
	RecordEvent(event *mkvparser.SubtitleEvent)
	Reset()
	Terminate()
}

type (
	PlaybackInfo struct {
		ID                 string                `json:"id"`
		StreamType         StreamType            `json:"streamType"`
		StreamPath         string                `json:"streamPath"`
		MimeType           string                `json:"mimeType"`              // e.g. "video/mp4", "video/webm"
		StreamUrl          string                `json:"streamUrl"`             // URL of the stream
		ContentLength      int64                 `json:"contentLength"`         // Size of the stream in bytes
		MkvMetadata        *mkvparser.Metadata   `json:"mkvMetadata,omitempty"` // nil if not ebml
		SubtitleTracks     []*VideoSubtitleTrack `json:"subtitleTracks,omitempty"`
		EntryListData      *anime.EntryListData  `json:"entryListData,omitempty"` // nil if not in list
		Episode            *anime.Episode        `json:"episode"`
		Media              *anilist.BaseAnime    `json:"media"`
		IsNakamaWatchParty bool                  `json:"isNakamaWatchParty"` // Is the stream from Nakama Watch Party
		LocalFile          *anime.LocalFile      `json:"localFile,omitempty"`

		MkvMetadataParser mo.Option[*mkvparser.MetadataParser] `json:"-"`
	}
)

type (
	// NativePlayer is the built-in HTML5 video player in Seanime.
	// There can only be one instance of this player at a time.
	NativePlayer struct {
		wsEventManager        events.WSEventManagerInterface
		videoCore             VideoCoreDelegate
		seekedEventCancelFunc context.CancelFunc

		logger *zerolog.Logger
	}

	PlaybackStatus struct {
		ClientId    string
		Url         string
		Paused      bool
		CurrentTime float64
		Duration    float64
	}

	NewNativePlayerOptions struct {
		WsEventManager events.WSEventManagerInterface
		Logger         *zerolog.Logger
		VideoCore      VideoCoreDelegate
	}
)

// New returns a new instance of NativePlayer.
// There should be only one for the lifetime of the app.
func New(options NewNativePlayerOptions) *NativePlayer {
	np := &NativePlayer{
		wsEventManager: options.WsEventManager,
		logger:         options.Logger,
		videoCore:      options.VideoCore,
	}

	return np
}

func (p *NativePlayer) VideoCore() VideoCoreDelegate {
	return p.videoCore
}

// sendPlayerEventTo sends an event of type events.NativePlayerEventType to the client.
func (p *NativePlayer) sendPlayerEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	if clientId != "" {
		p.wsEventManager.SendEventTo(clientId, string(events.NativePlayerEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		}, noLog...)
	} else {
		p.wsEventManager.SendEvent(string(events.NativePlayerEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		})
	}
}

func (p *NativePlayer) sendPlayerEvent(t string, payload interface{}) {
	p.wsEventManager.SendEvent(string(events.NativePlayerEventType), struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    t,
		Payload: payload,
	})
}
