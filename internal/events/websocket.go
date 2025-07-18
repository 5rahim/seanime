package events

import (
	"os"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type WSEventManagerInterface interface {
	SendEvent(t string, payload interface{})
	SendEventTo(clientId string, t string, payload interface{}, noLog ...bool)
	SubscribeToClientEvents(id string) *ClientEventSubscriber
	SubscribeToClientNativePlayerEvents(id string) *ClientEventSubscriber
	SubscribeToClientNakamaEvents(id string) *ClientEventSubscriber
	UnsubscribeFromClientEvents(id string)
}

type GlobalWSEventManagerWrapper struct {
	WSEventManager WSEventManagerInterface
}

var GlobalWSEventManager *GlobalWSEventManagerWrapper

func (w *GlobalWSEventManagerWrapper) SendEvent(t string, payload interface{}) {
	if w.WSEventManager == nil {
		return
	}
	w.WSEventManager.SendEvent(t, payload)
}

func (w *GlobalWSEventManagerWrapper) SendEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	if w.WSEventManager == nil {
		return
	}
	w.WSEventManager.SendEventTo(clientId, t, payload, noLog...)
}

type (
	// WSEventManager holds the websocket connection instance.
	// It is attached to the App instance, so it is available to other handlers.
	WSEventManager struct {
		Conns                              []*WSConn
		Logger                             *zerolog.Logger
		hasHadConnection                   bool
		mu                                 sync.Mutex
		eventMu                            sync.RWMutex
		clientEventSubscribers             *result.Map[string, *ClientEventSubscriber]
		clientNativePlayerEventSubscribers *result.Map[string, *ClientEventSubscriber]
		nakamaEventSubscribers             *result.Map[string, *ClientEventSubscriber]
	}

	ClientEventSubscriber struct {
		Channel chan *WebsocketClientEvent
		mu      sync.RWMutex
		closed  bool
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
	ret := &WSEventManager{
		Logger:                             logger,
		Conns:                              make([]*WSConn, 0),
		clientEventSubscribers:             result.NewResultMap[string, *ClientEventSubscriber](),
		clientNativePlayerEventSubscribers: result.NewResultMap[string, *ClientEventSubscriber](),
		nakamaEventSubscribers:             result.NewResultMap[string, *ClientEventSubscriber](),
	}
	GlobalWSEventManager = &GlobalWSEventManagerWrapper{
		WSEventManager: ret,
	}
	return ret
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

		for range ticker.C {
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
func (m *WSEventManager) SendEventTo(clientId string, t string, payload interface{}, noLog ...bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.Conns {
		if conn.ID == clientId {
			if t != "pong" {
				if len(noLog) == 0 || !noLog[0] {
					truncated := spew.Sprint(payload)
					if len(truncated) > 500 {
						truncated = truncated[:500] + "..."
					}
					m.Logger.Trace().Str("to", clientId).Str("type", t).Str("payload", truncated).Msg("ws: Sending message")
				}
			}
			_ = conn.Conn.WriteJSON(WSEvent{
				Type:    t,
				Payload: payload,
			})
		}
	}
}

func (m *WSEventManager) SendStringTo(clientId string, s string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.Conns {
		if conn.ID == clientId {
			_ = conn.Conn.WriteMessage(websocket.TextMessage, []byte(s))
		}
	}
}

func (m *WSEventManager) OnClientEvent(event *WebsocketClientEvent) {
	m.eventMu.RLock()
	defer m.eventMu.RUnlock()

	onEvent := func(key string, subscriber *ClientEventSubscriber) bool {
		go func() {
			defer util.HandlePanicInModuleThen("events/OnClientEvent/clientNativePlayerEventSubscribers", func() {})
			subscriber.mu.RLock()
			defer subscriber.mu.RUnlock()
			if !subscriber.closed {
				select {
				case subscriber.Channel <- event:
				default:
					// Channel is blocked, skip sending
					m.Logger.Warn().Msg("ws: Client event channel is blocked, event dropped")
				}
			}
		}()
		return true
	}

	switch event.Type {
	case NativePlayerEventType:
		m.clientNativePlayerEventSubscribers.Range(onEvent)
	case NakamaEventType:
		m.nakamaEventSubscribers.Range(onEvent)
	default:
		m.clientEventSubscribers.Range(onEvent)
	}
}

func (m *WSEventManager) SubscribeToClientEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent, 900),
	}
	m.clientEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *WSEventManager) SubscribeToClientNativePlayerEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent, 100),
	}
	m.clientNativePlayerEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *WSEventManager) SubscribeToClientNakamaEvents(id string) *ClientEventSubscriber {
	subscriber := &ClientEventSubscriber{
		Channel: make(chan *WebsocketClientEvent, 100),
	}
	m.nakamaEventSubscribers.Set(id, subscriber)
	return subscriber
}

func (m *WSEventManager) UnsubscribeFromClientEvents(id string) {
	m.eventMu.Lock()
	defer m.eventMu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			m.Logger.Warn().Msg("ws: Failed to unsubscribe from client events")
		}
	}()
	subscriber, ok := m.clientEventSubscribers.Get(id)
	if !ok {
		subscriber, ok = m.clientNativePlayerEventSubscribers.Get(id)
	}
	if ok {
		subscriber.mu.Lock()
		defer subscriber.mu.Unlock()
		subscriber.closed = true
		m.clientEventSubscribers.Delete(id)
		close(subscriber.Channel)
	}
}
