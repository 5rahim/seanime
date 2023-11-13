package events

import (
	"github.com/rs/zerolog"
)

type (
	MockWSEventManager struct {
		Conn   interface{}
		Logger *zerolog.Logger
	}

	MockWSEvent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

func NewMockWSEventManager(logger *zerolog.Logger) *MockWSEventManager {
	return &MockWSEventManager{
		Logger: logger,
	}
}

// SendEvent sends a websocket event to the client.
func (m *MockWSEventManager) SendEvent(t string, payload interface{}) {
	m.Logger.Trace().Any("payload", payload).Str("type", t).Msg("ws: Sent message")
}
