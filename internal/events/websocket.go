package events

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog"
	"os"
	"seanime/internal/util"
	"sync"
	"time"
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
		Conns            []*WSConn
		Logger           *zerolog.Logger
		hasHadConnection bool
		mu               sync.Mutex
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

// ExitIfNoConnsAsDesktopSidecar monitors the websocket connection as a desktop sidecar.
// It checks for a connection every 5 seconds. If a connection is lost, it starts a countdown a waits for 15 seconds.
// If a connection is not established within 15 seconds, it will exit the app.
func (m *WSEventManager) ExitIfNoConnsAsDesktopSidecar() {
	go func() {
		defer util.HandlePanicInModuleThen("events/ExitIfNoConnsAsDesktopSidecar", func() {})

		m.Logger.Info().Msg("ws: Monitoring connection as desktop sidecar")
		// Create a ticker to check connection every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Track connection loss time
		var connectionLostTime time.Time
		exitTimeout := 10 * time.Second

		for {
			select {
			case <-ticker.C:
				// Check WebSocket connection status
				if len(m.Conns) == 0 && m.hasHadConnection {
					// If not connected and first detection of connection loss
					if connectionLostTime.IsZero() {
						m.Logger.Warn().Msg("ws: No connection detected. Starting countdown...")
						connectionLostTime = time.Now()
					}

					// Check if connection has been lost for more than 15 seconds
					if time.Since(connectionLostTime) > exitTimeout {
						m.Logger.Warn().Msg("ws: No connection detected for 10 seconds. Exiting...")
						os.Exit(1)
					}
				} else {
					// Connection is active, reset connection lost time
					connectionLostTime = time.Time{}
				}
			}
		}
	}()
}

func (m *WSEventManager) AddConn(id string, conn *websocket.Conn) {
	m.hasHadConnection = true
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
			// Note: NaN error coming from [progress_tracking.go]
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
		if conn.ID == clientId {
			m.Logger.Trace().Str("to", clientId).Str("type", t).Str("payload", spew.Sprint(payload)).Msg("ws: Sending message")
			_ = conn.Conn.WriteJSON(WSEvent{
				Type:    t,
				Payload: payload,
			})
		}
	}
}
