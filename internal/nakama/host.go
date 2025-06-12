package nakama

import (
	"net/http"
	"seanime/internal/constants"
	"seanime/internal/events"
	"seanime/internal/util"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// startHostServices initializes the host services
func (m *Manager) startHostServices() {
	if m.settings == nil || !m.settings.IsHost || !m.settings.Enabled {
		return
	}

	m.logger.Info().Msg("nakama: Starting host services")

	// Start ping routine for connected peers
	go m.hostPingRoutine()

	// Start stale connection cleanup routine
	go m.staleConnectionCleanupRoutine()

	// Send event to client about host mode being enabled
	m.wsEventManager.SendEvent(events.NakamaHostStarted, map[string]interface{}{
		"enabled": true,
	})
}

// stopHostServices stops the host services
func (m *Manager) stopHostServices() {
	m.logger.Info().Msg("nakama: Stopping host services")

	// Disconnect all peers
	m.peerConnections.Range(func(id string, conn *PeerConnection) bool {
		conn.Close()
		return true
	})
	m.peerConnections.Clear()

	// Send event to client about host mode being disabled
	m.wsEventManager.SendEvent(events.NakamaHostStopped, map[string]interface{}{
		"enabled": false,
	})
}

// HandlePeerConnection handles incoming WebSocket connections from peers
func (m *Manager) HandlePeerConnection(w http.ResponseWriter, r *http.Request) {
	if m.settings == nil || !m.settings.IsHost || !m.settings.Enabled {
		http.Error(w, "Host mode not enabled", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error().Err(err).Msg("nakama: Failed to upgrade WebSocket connection")
		return
	}

	username := r.Header.Get("X-Seanime-Nakama-Username")
	if username == "" {
		username = "Peer_" + util.RandomStringWithAlphabet(8, "bcdefhijklmnopqrstuvwxyz0123456789")
	}

	serverVersion := r.Header.Get("X-Seanime-Nakama-Server-Version")
	if serverVersion != constants.Version {
		http.Error(w, "Server version mismatch", http.StatusForbidden)
		return
	}

	// Clean up any existing connections from the same username to prevent duplicates
	var oldConnections []string
	m.peerConnections.Range(func(id string, existingConn *PeerConnection) bool {
		if existingConn.Username == username {
			oldConnections = append(oldConnections, id)
		}
		return true
	})

	// Remove old connections from the same user
	for _, oldID := range oldConnections {
		if oldConn, exists := m.peerConnections.Get(oldID); exists {
			m.logger.Info().Str("oldPeerId", oldID).Str("username", username).Msg("nakama: Removing old connection for reconnecting peer")
			m.peerConnections.Delete(oldID)
			oldConn.Close()
		}
	}

	peerID := generateConnectionID()
	peerConn := &PeerConnection{
		ID:             peerID,
		Username:       username,
		Conn:           conn,
		ConnectionType: ConnectionTypePeer,
		Authenticated:  false,
		LastPing:       time.Now(),
	}

	m.logger.Info().Str("peerId", peerID).Str("username", username).Msg("nakama: New peer connection")

	// Add to connections
	m.peerConnections.Set(peerID, peerConn)

	// Handle the connection in a goroutine
	go m.handlePeerConnection(peerConn)
}

// handlePeerConnection handles messages from a specific peer
func (m *Manager) handlePeerConnection(peerConn *PeerConnection) {
	defer func() {
		m.logger.Info().Str("peerId", peerConn.ID).Msg("nakama: Peer disconnected")

		// Remove from connections (safe to call multiple times)
		if _, exists := m.peerConnections.Get(peerConn.ID); exists {
			m.peerConnections.Delete(peerConn.ID)

			// Send event to client about peer disconnection (only if we actually removed it)
			m.wsEventManager.SendEvent(events.NakamaPeerDisconnected, map[string]interface{}{
				"peerId": peerConn.ID,
			})
		}

		// Close connection (safe to call multiple times)
		peerConn.Close()
	}()

	// Set up ping/pong handler
	peerConn.Conn.SetPongHandler(func(appData string) error {
		peerConn.LastPing = time.Now()
		return nil
	})

	// Set read deadline
	peerConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			var message Message
			err := peerConn.Conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					m.logger.Error().Err(err).Str("peerId", peerConn.ID).Msg("nakama: Unexpected close error")
				}
				return
			}

			// Handle the message
			if err := m.handleMessage(&message, peerConn.ID); err != nil {
				m.logger.Error().Err(err).Str("peerId", peerConn.ID).Str("messageType", string(message.Type)).Msg("nakama: Failed to handle message")

				// Send error response
				errorMsg := &Message{
					Type: MessageTypeError,
					Payload: ErrorPayload{
						Message: err.Error(),
					},
					Timestamp: time.Now(),
				}
				peerConn.SendMessage(errorMsg)
			}

			// Reset read deadline
			peerConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

// hostPingRoutine sends ping messages to all connected peers
func (m *Manager) hostPingRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.peerConnections.Range(func(id string, conn *PeerConnection) bool {
				// Send ping
				message := &Message{
					Type:      MessageTypePing,
					Payload:   nil,
					Timestamp: time.Now(),
				}

				if err := conn.SendMessage(message); err != nil {
					m.logger.Error().Err(err).Str("peerId", id).Msg("nakama: Failed to send ping")
					// Don't close here, let the stale connection cleanup handle it
				}
				return true
			})
		}
	}
}

// staleConnectionCleanupRoutine periodically removes stale connections
func (m *Manager) staleConnectionCleanupRoutine() {
	ticker := time.NewTicker(120 * time.Second) // Run every 2 minutes
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.RemoveStaleConnections()
		}
	}
}
