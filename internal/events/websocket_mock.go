package events

import (
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
)

type (
	MockWSEventManager struct {
		Conn                   interface{}
		Logger                 *zerolog.Logger
		ClientEventSubscribers *result.Map[string, *ClientEventSubscriber]
	}

	MockWSEvent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

func NewMockWSEventManager(logger *zerolog.Logger) *MockWSEventManager {
	return &MockWSEventManager{
		Logger:                 logger,
		ClientEventSubscribers: result.NewResultMap[string, *ClientEventSubscriber](),
	}
}

// SendEvent sends a websocket event to the client.
func (m *MockWSEventManager) SendEvent(t string, payload interface{}) {
	m.Logger.Trace().Any("payload", payload).Str("type", t).Msg("ws: Sent message")
}

func (m *MockWSEventManager) SendEventTo(clientId string, t string, payload interface{}) {
	m.Logger.Trace().Any("payload", payload).Str("type", t).Str("clientId", clientId).Msg("ws: Sent message to client")
}

func (m *MockWSEventManager) SubscribeToClientEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent),
	}
	m.ClientEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *MockWSEventManager) UnsubscribeFromClientEvents(id string) {
	m.ClientEventSubscribers.Delete(id)
}

////

func (m *MockWSEventManager) MockSendClientEvent(event *WebsocketClientEvent) {
	m.ClientEventSubscribers.Range(func(key string, subscriber *ClientEventSubscriber) bool {
		subscriber.Channel <- event
		return true
	})
}
