package nativeplayer

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/util/result"
	"sync"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type StreamType string

const (
	StreamTypeTorrent StreamType = "torrent"
	StreamTypeFile    StreamType = "localfile"
	StreamTypeDebrid  StreamType = "debrid"
)

type (
	PlaybackInfo struct {
		ID            string               `json:"id"`
		StreamType    StreamType           `json:"streamType"`
		MimeType      string               `json:"mimeType"`                // e.g. "video/mp4", "video/webm"
		StreamUrl     string               `json:"streamUrl"`               // URL of the stream
		ContentLength int64                `json:"contentLength"`           // Size of the stream in bytes
		MkvMetadata   *mkvparser.Metadata  `json:"mkvMetadata,omitempty"`   // nil if not ebml
		EntryListData *anime.EntryListData `json:"entryListData,omitempty"` // nil if not in list
		Episode       *anime.Episode       `json:"episode"`
		Media         *anilist.BaseAnime   `json:"media"`

		MkvMetadataParser mo.Option[*mkvparser.MetadataParser] `json:"-"`
	}
)

type (
	// NativePlayer is the built-in HTML5 video player in Seanime.
	// There can only be one instance of this player at a time.
	NativePlayer struct {
		wsEventManager              events.WSEventManagerInterface
		clientPlayerEventSubscriber *events.ClientEventSubscriber

		playbackStatusMu sync.RWMutex
		playbackStatus   *PlaybackStatus

		seekedEventCancelFunc context.CancelFunc

		subscribers *result.Map[string, *Subscriber]

		logger *zerolog.Logger
	}

	PlaybackStatus struct {
		ClientId    string
		Url         string
		Paused      bool
		CurrentTime float64
		Duration    float64
	}

	// Subscriber listens to the player events
	Subscriber struct {
		eventCh chan VideoEvent
	}

	NewNativePlayerOptions struct {
		WsEventManager events.WSEventManagerInterface
		Logger         *zerolog.Logger
	}
)

// New returns a new instance of NativePlayer.
func New(options NewNativePlayerOptions) *NativePlayer {
	np := &NativePlayer{
		playbackStatus:              &PlaybackStatus{},
		wsEventManager:              options.WsEventManager,
		clientPlayerEventSubscriber: options.WsEventManager.SubscribeToClientNativePlayerEvents("nativeplayer"),
		subscribers:                 result.NewResultMap[string, *Subscriber](),
		logger:                      options.Logger,
	}

	np.listenToPlayerEvents()

	return np
}

// sendPlayerEventTo sends an event of type events.NativePlayerEventType to the client.
func (p *NativePlayer) sendPlayerEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	p.wsEventManager.SendEventTo(clientId, string(events.NativePlayerEventType), struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    t,
		Payload: payload,
	}, noLog...)
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Subscribe lets other modules subscribe to the native player events
func (p *NativePlayer) Subscribe(id string) *Subscriber {
	subscriber := &Subscriber{
		eventCh: make(chan VideoEvent, 10),
	}
	p.subscribers.Set(id, subscriber)

	return subscriber
}

// Unsubscribe removes a subscriber from the player.
func (p *NativePlayer) Unsubscribe(id string) {
	p.subscribers.Delete(id)
}

func (p *NativePlayer) notifySubscribers(event VideoEvent) {
	p.subscribers.Range(func(id string, subscriber *Subscriber) bool {
		select {
		case subscriber.eventCh <- event:
		default:
			// If the channel is full, skip sending the event
		}
		return true
	})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetPlaybackStatus returns the current playback status of the player.
func (p *NativePlayer) GetPlaybackStatus() *PlaybackStatus {
	p.playbackStatusMu.RLock()
	defer p.playbackStatusMu.RUnlock()
	return p.playbackStatus
}

func (p *NativePlayer) SetPlaybackStatus(status *PlaybackStatus) {
	p.setPlaybackStatus(func() {
		p.playbackStatus = status
	})
}

// setPlaybackStatus sets the current playback status of the player
// and notifies all subscribers of the change.
func (p *NativePlayer) setPlaybackStatus(do func()) {
	p.playbackStatusMu.Lock()
	defer p.playbackStatusMu.Unlock()
	do()
	p.notifySubscribers(&VideoStatusEvent{
		BaseVideoEvent: BaseVideoEvent{
			ClientId: p.playbackStatus.ClientId,
		},
		Status: *p.playbackStatus,
	})
}
