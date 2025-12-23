package nakama

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

// buildRoomHostWsURL constructs the WebSocket URL with properly escaped parameters
func buildRoomHostWsURL(baseURL, password, version string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("password", password)
	q.Set("version", version)
	u.RawQuery = q.Encode()
	return u.String(), nil
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

// RoomsAvailable returns true if the Rooms API is available.
func (m *Manager) RoomsAvailable() bool {
	resp, err := m.reqClient.R().
		Get(constants.SeanimeRoomsApiUrl + "/health")
	if err != nil {
		return false
	}

	if !resp.IsSuccessState() {
		return false
	}

	var createResp struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(resp.Bytes(), &createResp); err != nil {
		return false
	}

	if createResp.Status != "healthy" && createResp.Version != constants.SeanimeRoomsVersion {
		return false
	}

	return true
}

// CreateAndJoinRoom creates a room using Seanime Rooms API and connects as host
func (m *Manager) CreateAndJoinRoom() error {
	if !m.IsHost() {
		return errors.New("not in host mode")
	}

	if m.settings == nil || m.settings.HostPassword == "" {
		return errors.New("host password not set")
	}

	// Check if Rooms API is available
	if !m.RoomsAvailable() {
		return errors.New("rooms API not available")
	}

	// Create room
	room, err := m.createRoom(m.settings.HostPassword)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	m.logger.Info().Str("roomId", room.ID).Msg("nakama: Room created successfully")

	// Store the room and change the connection mode
	m.roomMu.Lock()
	m.currentRoom = room
	m.connectionMode = ConnectionModeRooms
	m.roomMu.Unlock()

	// Connect to room as host
	go m.connectToRoomAsHost(room)

	return nil
}

// connectToRoomAsHost establishes a WebSocket connection to the room as host
func (m *Manager) connectToRoomAsHost(room *Room) {
	// Websocket URL to connect to
	wsURL, err := buildRoomHostWsURL(room.HostWsUrl, room.Password, constants.Version)
	if err != nil {
		m.logger.Error().Err(err).Msg("nakama: Failed to build room host WS URL")
		m.wsEventManager.SendEvent(events.ErrorToast, "Failed to parse room URL")
		return
	}
	m.logger.Info().Str("roomId", room.ID).Str("wsUrl", wsURL).Msg("nakama: Connecting to room as host")

	// Create connection
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		m.logger.Error().Err(err).Msg("nakama: Failed to connect to room as host")
		m.wsEventManager.SendEvent(events.ErrorToast, "Failed to connect to room server")
		return
	}

	m.logger.Info().Str("roomId", room.ID).Msg("nakama: Connected to room as host")

	// Store connection info (reuse hostConnection for simplicity)
	m.hostMu.Lock()
	m.hostConnection = &HostConnection{
		URL:            wsURL,
		Conn:           conn,
		Authenticated:  true,
		LastPing:       time.Now(),
		PeerId:         m.peerId,
		Username:       m.username,
		ConnectionMode: ConnectionModeRooms,
	}
	m.hostMu.Unlock()

	// Send event about room creation
	m.wsEventManager.SendEvent(events.NakamaRoomCreated, map[string]interface{}{
		"roomId":      room.ID,
		"peerJoinUrl": room.PeerJoinUrl,
		"expiresAt":   room.ExpiresAt,
	})

	// Handle room messages
	go m.handleRoomHostConnection(conn, room)
}

// handleRoomHostConnection handles messages from the room relay server when acting as host
func (m *Manager) handleRoomHostConnection(conn *websocket.Conn, room *Room) {
	defer func() {
		m.logger.Info().Str("roomId", room.ID).Msg("nakama: Room host connection closed")
		_ = conn.Close()

		m.hostMu.Lock()
		if m.hostConnection != nil && m.hostConnection.Conn == conn {
			m.hostConnection = nil
		}
		m.hostMu.Unlock()

		// Attempt reconnection instead of immediately clearing room
		// This preserves watch party state during temporary relay server reboots
		m.roomMu.RLock()
		isCurrentRoom := m.currentRoom != nil && m.currentRoom.ID == room.ID
		m.roomMu.RUnlock()

		m.hostMu.Lock()
		shouldReconnect := isCurrentRoom && m.settings != nil && m.settings.IsHost && m.settings.Enabled && !m.reconnecting
		if shouldReconnect {
			m.reconnecting = true
			m.hostMu.Unlock()

			m.logger.Info().Str("roomId", room.ID).Msg("nakama: Scheduling reconnection to room")
			// Reconnect after a short delay
			time.AfterFunc(5*time.Second, func() {
				m.reconnectToRoomAsHost(room)
			})
		} else {
			m.hostMu.Unlock()

			if !isCurrentRoom {
				m.logger.Info().Str("roomId", room.ID).Msg("nakama: Not reconnecting, this is not the current room")
				return
			}

			// Only clear room if we're not reconnecting
			m.roomMu.Lock()
			if m.currentRoom != nil && m.currentRoom.ID == room.ID {
				m.currentRoom = nil
			}
			m.connectionMode = ConnectionModeDirect
			m.roomMu.Unlock()

			// Send event about room closure
			m.wsEventManager.SendEvent(events.NakamaRoomClosed, map[string]interface{}{
				"roomId": room.ID,
			})
		}
	}()

	// Set up pong handler and reset read deadline
	conn.SetPongHandler(func(appData string) error {
		m.hostMu.Lock()
		// Only update if this is still the active connection
		if m.hostConnection != nil && m.hostConnection.Conn == conn {
			m.hostConnection.LastPing = time.Now()
			//m.logger.Debug().Msg("nakama: Received pong from room, updated LastPing")
		}
		m.hostMu.Unlock()
		// Reset read deadline when pong is received
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			m.logger.Warn().Err(err).Msg("nakama: Failed to set read deadline in pong handler")
		}
		return nil
	})

	// Set initial read deadline
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			var message Message
			err := conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					m.logger.Error().Err(err).Msg("nakama: Unexpected close error from room")
				}
				return
			}

			// Handle message from relay (sent by peers)
			if err := m.handleMessage(&message, "room_peer"); err != nil {
				m.logger.Error().Err(err).Str("messageType", string(message.Type)).Msg("nakama: Failed to handle room message")
			}

			// Reset read deadline
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

// SendMessageToRoom sends a message to all peers via the room relay
func (m *Manager) SendMessageToRoom(msgType MessageType, payload interface{}) error {
	m.hostMu.RLock()
	defer m.hostMu.RUnlock()

	// No host connection or connect mode isn't rooms
	if m.hostConnection == nil || m.hostConnection.ConnectionMode != ConnectionModeRooms {
		// Might be reconnecting, so just skip the message
		m.logger.Warn().Str("messageType", string(msgType)).Msg("nakama: Cannot send message to room (may be reconnecting)")
		return nil
	}

	message := &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	err := m.hostConnection.SendMessage(message)
	if err != nil {
		// Just skip
		m.logger.Error().Err(err).Str("messageType", string(msgType)).Msg("nakama: Failed to send message to room (may be reconnecting)")
	}
	return err
}

// reconnectToRoomAsHost attempts to reconnect to a room after disconnection
func (m *Manager) reconnectToRoomAsHost(room *Room) {
	defer func() {
		m.hostMu.Lock()
		m.reconnecting = false
		m.hostMu.Unlock()
	}()

	m.logger.Info().Str("roomId", room.ID).Msg("nakama: Attempting to reconnect to room as host")

	maxRetries := 5
	retryDelay := 5 * time.Second

	var conn *websocket.Conn
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-m.ctx.Done():
			m.logger.Info().Msg("nakama: Room reconnection cancelled")
			return
		default:
		}

		wsURL, parseErr := buildRoomHostWsURL(room.HostWsUrl, room.Password, constants.Version)
		if parseErr != nil {
			m.logger.Error().Err(parseErr).Msg("nakama: Failed to build room host WS URL")
			return
		}

		// Create connection
		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}

		conn, _, err = dialer.Dial(wsURL, nil)
		if err != nil {
			m.logger.Error().Err(err).Int("attempt", attempt+1).Msg("nakama: Failed to reconnect to room")

			if attempt < maxRetries-1 {
				select {
				case <-m.ctx.Done():
					return
				case <-time.After(retryDelay):
					continue
				}
			}
		} else {
			// Success
			m.logger.Info().Str("roomId", room.ID).Msg("nakama: Successfully reconnected to room as host")

			// Store connection info
			m.hostMu.Lock()
			m.hostConnection = &HostConnection{
				URL:            wsURL,
				Conn:           conn,
				Authenticated:  true,
				LastPing:       time.Now(),
				PeerId:         m.peerId,
				Username:       m.username,
				ConnectionMode: ConnectionModeRooms,
			}
			m.hostMu.Unlock()

			// Send event about successful reconnection
			m.wsEventManager.SendEvent(events.NakamaRoomReconnected, map[string]interface{}{
				"roomId": room.ID,
			})

			// Handle room messages
			go m.handleRoomHostConnection(conn, room)
			return
		}
	}

	// Failed to reconnect after all retries
	m.logger.Error().Str("roomId", room.ID).Msg("nakama: Failed to reconnect to room after all retries")

	// Devnote: Do not clear room and watch party state
	//go m.watchPartyManager.StopWatchParty()
	m.roomMu.Lock()
	if m.currentRoom != nil && m.currentRoom.ID == room.ID {
		m.currentRoom = nil
	}
	m.connectionMode = ConnectionModeDirect
	m.roomMu.Unlock()

	// Send events
	m.wsEventManager.SendEvent(events.NakamaRoomClosed, map[string]interface{}{
		"roomId": room.ID,
	})
	m.wsEventManager.SendEvent(events.ErrorToast, "Failed to reconnect to room. Please create a new room.")
}

// DisconnectFromRoom disconnects the host from the current room
func (m *Manager) DisconnectFromRoom() error {
	if !m.IsHost() {
		return errors.New("not in host mode")
	}

	m.roomMu.RLock()
	connectionMode := m.connectionMode
	currentRoom := m.currentRoom
	m.roomMu.RUnlock()

	if connectionMode != ConnectionModeRooms {
		return errors.New("not connected to a room")
	}

	if currentRoom == nil {
		return errors.New("no active room")
	}

	m.logger.Info().Str("roomId", currentRoom.ID).Msg("nakama: Disconnecting from room")

	// Stop any active watch party
	go m.watchPartyManager.StopWatchParty()

	// Close the room connection
	m.hostMu.Lock()
	if m.hostConnection != nil && m.hostConnection.ConnectionMode == ConnectionModeRooms {
		m.hostConnection.Close()
		m.hostConnection = nil
	}
	m.hostMu.Unlock()

	// Clear room state
	m.roomMu.Lock()
	m.currentRoom = nil
	m.connectionMode = ConnectionModeDirect
	m.roomMu.Unlock()

	// Send event about room disconnection
	m.wsEventManager.SendEvent(events.NakamaRoomClosed, map[string]interface{}{
		"roomId": currentRoom.ID,
	})

	m.logger.Info().Str("roomId", currentRoom.ID).Msg("nakama: Successfully disconnected from room")
	return nil
}

// stopHostServices stops the host services
func (m *Manager) stopHostServices() {
	m.logger.Info().Msg("nakama: Stopping host services")

	// Close room connection if in room mode
	m.roomMu.Lock()
	if m.currentRoom != nil {
		m.currentRoom = nil
	}
	m.connectionMode = ConnectionModeDirect
	m.roomMu.Unlock()

	m.hostMu.Lock()
	if m.hostConnection != nil && m.hostConnection.ConnectionMode == ConnectionModeRooms {
		m.hostConnection.Close()
		m.hostConnection = nil
	}
	m.hostMu.Unlock()

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

	m.roomMu.RLock()
	connectionMode := m.connectionMode
	m.roomMu.RUnlock()
	if connectionMode != ConnectionModeDirect {
		http.Error(w, "Host is currently in room mode, cannot accept direct peer connections", http.StatusForbidden)
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
			// Check connection mode
			m.roomMu.RLock()
			connectionMode := m.connectionMode
			m.roomMu.RUnlock()

			if connectionMode == ConnectionModeRooms {
				// In rooms mode, ping the relay server using WebSocket ping frames
				m.hostMu.RLock()
				if m.hostConnection != nil && m.hostConnection.ConnectionMode == ConnectionModeRooms {
					lastPing := m.hostConnection.LastPing
					timeSince := time.Since(lastPing)

					//m.logger.Debug().
					//	Dur("timeSinceLastPing", timeSince).
					//	Time("lastPing", lastPing).
					//	Msg("nakama: Room connection health check")

					// Check if room connection is still alive
					if timeSince > 90*time.Second {
						m.logger.Warn().
							Dur("timeSinceLastPing", timeSince).
							Msg("nakama: Room connection timeout, no pong received")
						m.hostConnection.Close()
						m.hostMu.RUnlock()
						continue
					}

					// Send WebSocket ping frame to keep connection alive
					//m.logger.Debug().Msg("nakama: Sending ping to room")
					if err := m.hostConnection.Conn.WriteControl(
						websocket.PingMessage,
						[]byte{},
						time.Now().Add(10*time.Second),
					); err != nil {
						m.logger.Error().Err(err).Msg("nakama: Failed to send ping to room")
						m.hostConnection.Close()
					}
				}
				m.hostMu.RUnlock()
			} else {
				// In direct mode, ping each peer connection
				m.peerConnections.Range(func(id string, conn *PeerConnection) bool {
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
}

// staleConnectionCleanupRoutine periodically removes stale connections
func (m *Manager) staleConnectionCleanupRoutine() {
	ticker := time.NewTicker(120 * time.Second)
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
