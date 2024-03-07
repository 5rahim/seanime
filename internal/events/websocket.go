package events

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog"
)

type IWSEventManager interface {
	SendEvent(t string, payload interface{})
}

type (
	// WSEventManager holds the websocket connection instance.
	// It is attached to the App instance, so it is available to other handlers.
	WSEventManager struct {
		Conn   *websocket.Conn
		Logger *zerolog.Logger
	}

	WSEvent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

// NewWSEventManager creates a new WSEventManager instance for App.
func NewWSEventManager(logger *zerolog.Logger) *WSEventManager {
	return &WSEventManager{
		Logger: logger,
	}
}

// SendEvent sends a websocket event to the client.
func (m *WSEventManager) SendEvent(t string, payload interface{}) {
	// If there's no connection, do nothing
	if m.Conn == nil {
		return
	}

	err := m.Conn.WriteJSON(WSEvent{
		Type:    t,
		Payload: payload,
	})
	if err != nil {
		m.Logger.Err(err).Msg("ws: Failed to send message")
	}
	//m.Logger.Trace().Str("type", t).Msg("ws: Sent message")
}
