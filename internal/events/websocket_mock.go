package events

import (
	"seanime/internal/util/result"
	"sync"

	"github.com/rs/zerolog"
)

type (
	MockWSEventManager struct {
		Conn                    interface{}
		Logger                  *zerolog.Logger
		ClientEventSubscribers  *result.Map[string, *ClientEventSubscriber]
		videoCoreSubscribers    *result.Map[string, *ClientEventSubscriber]
		mpvCoreSubscribers      *result.Map[string, *ClientEventSubscriber]
		nativePlayerSubscribers *result.Map[string, *ClientEventSubscriber]
		mu                      sync.Mutex
		sentEvents              []MockWSEvent
	}

	MockWSEvent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

func NewMockWSEventManager(logger *zerolog.Logger) *MockWSEventManager {
	return &MockWSEventManager{
		Logger:                  logger,
		ClientEventSubscribers:  result.NewMap[string, *ClientEventSubscriber](),
		videoCoreSubscribers:    result.NewMap[string, *ClientEventSubscriber](),
		mpvCoreSubscribers:      result.NewMap[string, *ClientEventSubscriber](),
		nativePlayerSubscribers: result.NewMap[string, *ClientEventSubscriber](),
	}
}

// SendEvent sends a websocket event to the client.
func (m *MockWSEventManager) SendEvent(t string, payload interface{}) {
	m.mu.Lock()
	m.sentEvents = append(m.sentEvents, MockWSEvent{Type: t, Payload: payload})
	m.mu.Unlock()
	m.Logger.Trace().Any("payload", payload).Str("type", t).Msg("ws: Sent message")
}

func (m *MockWSEventManager) SendEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	m.mu.Lock()
	m.sentEvents = append(m.sentEvents, MockWSEvent{Type: t, Payload: payload})
	m.mu.Unlock()
	if len(noLog) == 0 || !noLog[0] {
		m.Logger.Trace().Any("payload", payload).Str("type", t).Str("clientId", clientId).Msg("ws: Sent message to client")
	}
}

func (m *MockWSEventManager) Events() []MockWSEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	ret := make([]MockWSEvent, len(m.sentEvents))
	copy(ret, m.sentEvents)
	return ret
}

func (m *MockWSEventManager) GetClientIds() []string {
	ids := make([]string, 0)
	m.ClientEventSubscribers.Range(func(key string, subscriber *ClientEventSubscriber) bool {
		if subscriber != nil {
			ids = append(ids, key)
		}
		return true
	})
	return ids
}

func (m *MockWSEventManager) GetClientPlatform(clientId string) string {
	return ""
}

func (m *MockWSEventManager) SubscribeToClientEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.ClientEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) SubscribeToClientVideoCoreEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.videoCoreSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) SubscribeToClientMpvCoreEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.mpvCoreSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) SubscribeToClientNativePlayerEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.nativePlayerSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) SubscribeToClientNakamaEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.ClientEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) SubscribeToClientPlaylistEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.ClientEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) UnsubscribeFromClientEvents(id string) {
	m.ClientEventSubscribers.Delete(id)
	m.videoCoreSubscribers.Delete(id)
	m.mpvCoreSubscribers.Delete(id)
	m.nativePlayerSubscribers.Delete(id)
}

////

func (m *MockWSEventManager) MockSendClientEvent(event *WebsocketClientEvent) {
	subscribers := m.ClientEventSubscribers
	switch event.Type {
	case VideoCoreEventType:
		subscribers = m.videoCoreSubscribers
	case MpvCoreEventType:
		subscribers = m.mpvCoreSubscribers
	case NativePlayerEventType:
		subscribers = m.nativePlayerSubscribers
	}
	subscribers.Range(func(key string, subscriber *ClientEventSubscriber) bool {
		subscriber.Channel <- event
		return true
	})
}
