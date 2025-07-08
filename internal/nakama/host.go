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

	// Clean up any existing watch party session
	m.watchPartyManager.Cleanup()

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
	// Generate a random username if username is not set (this shouldn't be the case because the peer will generate its own username)
	if username == "" {
		username = "Peer_" + util.RandomStringWithAlphabet(8, "bcdefhijklmnopqrstuvwxyz0123456789")
	}

	peerID := r.Header.Get("X-Seanime-Nakama-Peer-Id")
	if peerID == "" {
		m.logger.Error().Msg("nakama: Peer connection missing PeerID header")
		http.Error(w, "Missing PeerID header", http.StatusBadRequest)
		return
	}

	serverVersion := r.Header.Get("X-Seanime-Nakama-Server-Version")
	if serverVersion != constants.Version {
		http.Error(w, "Server version mismatch", http.StatusBadRequest)
		return
	}

	// Check for existing connection with the same PeerID (reconnection scenario)
	var existingConnID string
	m.peerConnections.Range(func(id string, existingConn *PeerConnection) bool {
		if existingConn.PeerId == peerID {
			existingConnID = id
			return false // Stop iteration
		}
		return true
	})

	// Remove existing connection for this PeerID to handle reconnection
	if existingConnID != "" {
		if oldConn, exists := m.peerConnections.Get(existingConnID); exists {
			m.logger.Info().Str("peerID", peerID).Str("oldConnID", existingConnID).Msg("nakama: Removing old connection for reconnecting peer")
			m.peerConnections.Delete(existingConnID)
			oldConn.Close()
		}
	}

	// Generate new internal connection ID
	internalConnID := generateConnectionID()
	peerConn := &PeerConnection{
		ID:             internalConnID,
		PeerId:         peerID,
		Username:       username,
		Conn:           conn,
		ConnectionType: ConnectionTypePeer,
		Authenticated:  false,
		LastPing:       time.Now(),
	}

	m.logger.Info().Str("internalConnID", internalConnID).Str("peerID", peerID).Str("username", username).Msg("nakama: New peer connection")

	// Add to connections using internal connection ID as key
	m.peerConnections.Set(internalConnID, peerConn)

	// Handle the connection in a goroutine
	go m.handlePeerConnection(peerConn)
}

// handlePeerConnection handles messages from a specific peer
func (m *Manager) handlePeerConnection(peerConn *PeerConnection) {
	defer func() {
		m.logger.Info().Str("peerId", peerConn.PeerId).Str("internalConnID", peerConn.ID).Msg("nakama: Peer disconnected")

		// Remove from connections (safe to call multiple times)
		if _, exists := m.peerConnections.Get(peerConn.ID); exists {
			m.peerConnections.Delete(peerConn.ID)

			// Remove peer from watch party if they were participating
			m.watchPartyManager.HandlePeerDisconnected(peerConn.PeerId)

			// Send event to client about peer disconnection (only if we actually removed it)
			m.wsEventManager.SendEvent(events.NakamaPeerDisconnected, map[string]interface{}{
				"peerId": peerConn.PeerId,
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
					m.logger.Error().Err(err).Str("peerId", peerConn.PeerId).Msg("nakama: Unexpected close error")
				}
				return
			}

			// Handle the message using internal connection ID for message routing
			if err := m.handleMessage(&message, peerConn.ID); err != nil {
				m.logger.Error().Err(err).Str("peerId", peerConn.PeerId).Str("messageType", string(message.Type)).Msg("nakama: Failed to handle message")

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
					m.logger.Error().Err(err).Str("peerId", conn.PeerId).Msg("nakama: Failed to send ping")
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
