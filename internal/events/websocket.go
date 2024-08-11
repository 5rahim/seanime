package events

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog"
	"sync"
)

type WSEventManagerInterface interface {
	SendEvent(t string, payload interface{})
	SendEventTo(clientId string, t string, payload interface{})
}

type (
	// WSEventManager holds the websocket connection instance.
	// It is attached to the App instance, so it is available to other handlers.
	WSEventManager struct {
		//Conn   *websocket.Conn // DEPRECATED
		Conns  []*WSConn
		Logger *zerolog.Logger
		mu     sync.Mutex
	}

	WSConn struct {
		ID   string
		Conn *websocket.Conn
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
		Conns:  make([]*WSConn, 0),
	}
}

func (m *WSEventManager) AddConn(id string, conn *websocket.Conn) {
	m.Conns = append(m.Conns, &WSConn{
		ID:   id,
		Conn: conn,
	})
}

func (m *WSEventManager) RemoveConn(id string) {
	for i, conn := range m.Conns {
		if conn.ID == id {
			m.Conns = append(m.Conns[:i], m.Conns[i+1:]...)
			break
		}
	}
}

// SendEvent sends a websocket event to the client.
func (m *WSEventManager) SendEvent(t string, payload interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// If there's no connection, do nothing
	//if m.Conn == nil {
	//	return
	//}

	if t != PlaybackManagerProgressPlaybackState && payload == nil {
		m.Logger.Trace().Str("type", t).Msg("ws: Sending message")
	}

	for _, conn := range m.Conns {
		err := conn.Conn.WriteJSON(WSEvent{
			Type:    t,
			Payload: payload,
		})
		if err != nil {
			// TODO NaN error coming from [progress_tracking.go]
			//m.Logger.Err(err).Msg("ws: Failed to send message")
		}
		//m.Logger.Trace().Str("type", t).Msg("ws: Sent message")
	}

	//err := m.Conn.WriteJSON(WSEvent{
	//	Type:    t,
	//	Payload: payload,
	//})
	//if err != nil {
	//	m.Logger.Err(err).Msg("ws: Failed to send message")
	//}
	//m.Logger.Trace().Str("type", t).Msg("ws: Sent message")
}

// SendEventTo sends a websocket event to the specified client.
func (m *WSEventManager) SendEventTo(clientId string, t string, payload interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.Conns {
		spew.Dump(conn.ID, clientId)
		if conn.ID == clientId {
			m.Logger.Trace().Str("to", clientId).Str("type", t).Str("payload", spew.Sprint(payload)).Msg("ws: Sending message")
			_ = conn.Conn.WriteJSON(WSEvent{
				Type:    t,
				Payload: payload,
			})
		}
	}
}
