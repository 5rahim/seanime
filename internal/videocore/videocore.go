package videocore

import (
	"seanime/internal/events"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	// VideoCore is the built-in HTML5 video player in Seanime.
	// It can be the NativePlayer (Seanime Denshi player) or the WebPlayer.
	VideoCore struct {
		wsEventManager              events.WSEventManagerInterface
		clientPlayerEventSubscriber *events.ClientEventSubscriber

		playbackStatusMu sync.RWMutex
		playbackStatus   mo.Option[*PlaybackStatus]
		playbackStateMu  sync.RWMutex
		playbackState    mo.Option[*PlaybackState]

		subscribers *result.Map[string, *Subscriber]

		eventBus       chan VideoEvent
		dispatcherStop chan struct{}
		dispatcherOnce sync.Once

		logger *zerolog.Logger
	}

	// Subscriber listens to the player events
	Subscriber struct {
		eventCh chan VideoEvent
	}
)

func NewVideoCore(wsManager events.WSEventManagerInterface, logger *zerolog.Logger) *VideoCore {
	vc := &VideoCore{
		wsEventManager: wsManager,
		subscribers:    result.NewMap[string, *Subscriber](),
		logger:         logger,
		eventBus:       make(chan VideoEvent, 100),
		dispatcherStop: make(chan struct{}),
	}
	vc.Start()
	return vc
}

// Start spins up the background dispatcher.
func (vc *VideoCore) Start() {
	vc.dispatcherOnce.Do(func() {
		go func() {
			for {
				select {
				case <-vc.dispatcherStop:
					return
				case event := <-vc.eventBus:
					vc.distributeEvent(event)
				}
			}
		}()
	})
}

// Shutdown gracefully stops the dispatcher.
func (vc *VideoCore) Shutdown() {
	close(vc.dispatcherStop)
}

func (vc *VideoCore) PushEvent(event VideoEvent) {
	select {
	case vc.eventBus <- event:
	default:
		vc.logger.Warn().Msgf("VideoCore: Event bus full, dropping event %s", event.GetId())
	}
}

func (vc *VideoCore) distributeEvent(event VideoEvent) {
	vc.subscribers.Range(func(id string, subscriber *Subscriber) bool {

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
			}
		}
		return true
	})
}

// sendPlayerEventTo sends an event of type events.NativePlayerEventType to the client.
func (vc *VideoCore) sendPlayerEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	vc.playbackStatusMu.RLock()
	defer vc.playbackStatusMu.RUnlock()
	if playbackStatus, ok := vc.playbackStatus.Get(); ok && clientId == "" {
		clientId = playbackStatus.ClientId
	}

	if clientId != "" {
		vc.wsEventManager.SendEventTo(clientId, string(events.NativePlayerEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		}, noLog...)
	} else {
		vc.wsEventManager.SendEvent(string(events.NativePlayerEventType), struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    t,
			Payload: payload,
		})
	}
}

func (vc *VideoCore) sendPlayerEvent(t string, payload interface{}) {
	vc.wsEventManager.SendEvent(string(events.NativePlayerEventType), struct {
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
		eventCh: make(chan VideoEvent, 50),
	}
	vc.subscribers.Set(id, subscriber)

	return subscriber
}

// Unsubscribe removes a subscriber from the player.
func (vc *VideoCore) Unsubscribe(id string) {
	if subscriber, ok := vc.subscribers.Get(id); ok {
		close(subscriber.eventCh)
		vc.subscribers.Delete(id)
	}
}

// Events returns the event channel for the subscriber.
func (s *Subscriber) Events() <-chan VideoEvent {
	return s.eventCh
}

func (vc *VideoCore) RegisterEventCallback(callback func(event VideoEvent, cancelFunc func())) (cancel func()) {
	id := uuid.NewString()
	sub := vc.Subscribe(id)
	cancel = func() {
		vc.Unsubscribe(id)
	}
	go func(sub *Subscriber) {
		for event := range sub.Events() {
			callback(event, cancel)
		}
	}(sub)

	return cancel
}
