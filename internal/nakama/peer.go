package nakama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"seanime/internal/constants"
	"seanime/internal/events"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// connectToHost establishes a connection to the Nakama host
func (m *Manager) connectToHost() {
	if m.settings == nil || !m.settings.Enabled || m.settings.RemoteServerURL == "" || m.settings.RemoteServerPassword == "" {
		return
	}

	m.logger.Info().Str("url", m.settings.RemoteServerURL).Msg("nakama: Connecting to host")

	// Cancel any existing connection attempts
	m.hostMu.Lock()
	if m.hostConnectionCancel != nil {
		m.hostConnectionCancel()
	}

	// Create new context for this connection attempt
	m.hostConnectionCtx, m.hostConnectionCancel = context.WithCancel(m.ctx)

	// Prevent multiple concurrent connection attempts
	if m.reconnecting {
		m.hostMu.Unlock()
		return
	}
	m.reconnecting = true
	m.hostMu.Unlock()

	go m.connectToHostAsync()
}

// disconnectFromHost disconnects from the Nakama host
func (m *Manager) disconnectFromHost() {
	m.hostMu.Lock()
	defer m.hostMu.Unlock()

	// Cancel any ongoing connection attempts
	if m.hostConnectionCancel != nil {
		m.hostConnectionCancel()
		m.hostConnectionCancel = nil
	}

	if m.hostConnection != nil {
		m.logger.Info().Msg("nakama: Disconnecting from host")

		// Cancel any reconnection timer
		if m.hostConnection.reconnectTimer != nil {
			m.hostConnection.reconnectTimer.Stop()
		}

		m.hostConnection.Close()
		m.hostConnection = nil

		// Send event to client about disconnection
		m.wsEventManager.SendEvent(events.NakamaHostDisconnected, map[string]interface{}{
			"connected": false,
		})
	}

	// Reset reconnecting flag
	m.reconnecting = false
}

// connectToHostAsync handles the actual connection logic with retries
func (m *Manager) connectToHostAsync() {
	defer func() {
		m.hostMu.Lock()
		m.reconnecting = false
		m.hostMu.Unlock()
	}()

	if m.settings == nil || !m.settings.Enabled || m.settings.RemoteServerURL == "" || m.settings.RemoteServerPassword == "" {
		return
	}

	// Get the connection context
	m.hostMu.RLock()
	connCtx := m.hostConnectionCtx
	m.hostMu.RUnlock()

	if connCtx == nil {
		return
	}

	maxRetries := 5
	retryDelay := 5 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-connCtx.Done():
			m.logger.Info().Msg("nakama: Connection attempt cancelled")
			return
		case <-m.ctx.Done():
			return
		default:
		}

		if err := m.attemptHostConnection(connCtx); err != nil {
			m.logger.Error().Err(err).Int("attempt", attempt+1).Msg("nakama: Failed to connect to host")

			if attempt < maxRetries-1 {
				select {
				case <-connCtx.Done():
					m.logger.Info().Msg("nakama: Connection attempt cancelled")
					return
				case <-m.ctx.Done():
					return
				case <-time.After(retryDelay):
					retryDelay *= 2 // Exponential backoff
					continue
				}
			}
		} else {
			// Success
			m.logger.Info().Msg("nakama: Successfully connected to host")
			return
		}
	}

	// Only log error if not cancelled
	select {
	case <-connCtx.Done():
		m.logger.Info().Msg("nakama: Connection attempts cancelled")
	default:
		m.logger.Error().Msg("nakama: Failed to connect to host after all retries")
		m.wsEventManager.SendEvent(events.ErrorToast, "Failed to connect to Nakama host after multiple attempts.")
	}
}

// attemptHostConnection makes a single connection attempt to the host
func (m *Manager) attemptHostConnection(connCtx context.Context) error {
	// Parse URL
	u, err := url.Parse(m.settings.RemoteServerURL)
	if err != nil {
		return err
	}

	// Convert HTTP to WebSocket scheme
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	// Add Nakama WebSocket path
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Path += "api/v1/nakama/ws"

	// Generate UUID for this peer instance
	peerID := uuid.New().String()

	username := m.username
	// Generate a random username if username is not set
	if username == "" {
		username = "Peer_" + util.RandomStringWithAlphabet(8, "bcdefhijklmnopqrstuvwxyz0123456789")
	}

	// Set up headers for authentication
	headers := http.Header{}
	headers.Set("X-Seanime-Nakama-Token", m.settings.RemoteServerPassword)
	headers.Set("X-Seanime-Nakama-Username", username)
	headers.Set("X-Seanime-Nakama-Server-Version", constants.Version)
	headers.Set("X-Seanime-Nakama-Peer-Id", peerID)

	// Create a dialer with the connection context
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// Connect with context
	conn, _, err := dialer.DialContext(connCtx, u.String(), headers)
	if err != nil {
		return err
	}

	hostConn := &HostConnection{
		URL:           u.String(),
		Conn:          conn,
		Authenticated: false,
		LastPing:      time.Now(),
		PeerId:        peerID, // Store our generated PeerID
	}

	// Authenticate
	authMessage := &Message{
		Type: MessageTypeAuth,
		Payload: AuthPayload{
			Password: m.settings.RemoteServerPassword,
			PeerId:   peerID, // Include PeerID in auth payload
		},
		Timestamp: time.Now(),
	}

	if err := hostConn.SendMessage(authMessage); err != nil {
		_ = conn.Close()
		return err
	}

	// Wait for auth response with timeout
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var authResponse Message
	if err := conn.ReadJSON(&authResponse); err != nil {
		_ = conn.Close()
		return err
	}

	if authResponse.Type != MessageTypeAuthReply {
		_ = conn.Close()
		return errors.New("unexpected auth response type")
	}

	// Parse auth response
	authReplyData, err := json.Marshal(authResponse.Payload)
	if err != nil {
		_ = conn.Close()
		return err
	}

	var authReply AuthReplyPayload
	if err := json.Unmarshal(authReplyData, &authReply); err != nil {
		_ = conn.Close()
		return err
	}

	if !authReply.Success {
		_ = conn.Close()
		return errors.New("authentication failed: " + authReply.Message)
	}

	// Verify that the host echoed back our PeerID
	if authReply.PeerId != peerID {
		m.logger.Warn().Str("expectedPeerID", peerID).Str("receivedPeerID", authReply.PeerId).Msg("nakama: Host returned different PeerID")
	}

	hostConn.Username = authReply.Username
	if hostConn.Username == "" {
		hostConn.Username = "Host_" + util.RandomStringWithAlphabet(8, "bcdefhijklmnopqrstuvwxyz0123456789")
	}
	hostConn.Authenticated = true

	// Set the connection and cancel any existing reconnection timer
	m.hostMu.Lock()
	if m.hostConnection != nil && m.hostConnection.reconnectTimer != nil {
		m.hostConnection.reconnectTimer.Stop()
	}
	m.hostConnection = hostConn
	m.hostMu.Unlock()

	// Send event to client about successful connection
	m.wsEventManager.SendEvent(events.NakamaHostConnected, map[string]interface{}{
		"connected":     true,
		"authenticated": true,
		"url":           hostConn.URL,
		"peerID":        peerID, // Include our PeerID in the event
	})

	// Start handling the connection
	go m.handleHostConnection(hostConn)

	// Start client ping routine
	go m.clientPingRoutine()

	return nil
}

// handleHostConnection handles messages from the host
func (m *Manager) handleHostConnection(hostConn *HostConnection) {
	defer func() {
		m.logger.Info().Msg("nakama: Host connection closed")

		m.hostMu.Lock()
		if m.hostConnection == hostConn {
			m.hostConnection = nil
		}
		m.hostMu.Unlock()

		// Send event to client about disconnection
		m.wsEventManager.SendEvent(events.NakamaHostDisconnected, map[string]interface{}{
			"connected": false,
		})

		// Attempt reconnection after a delay if settings are still valid and not already reconnecting
		m.hostMu.Lock()
		shouldReconnect := m.settings != nil && m.settings.RemoteServerURL != "" && m.settings.RemoteServerPassword != "" && !m.reconnecting
		if shouldReconnect {
			m.reconnecting = true
			hostConn.reconnectTimer = time.AfterFunc(10*time.Second, func() {
				m.connectToHostAsync()
			})
		}
		m.hostMu.Unlock()
	}()

	// Set up ping/pong handler
	hostConn.Conn.SetPongHandler(func(appData string) error {
		hostConn.LastPing = time.Now()
		return nil
	})

	// Set read deadline
	_ = hostConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			var message Message
			err := hostConn.Conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					m.logger.Error().Err(err).Msg("nakama: Unexpected close error from host")
				}
				return
			}

			// Handle the message
			if err := m.handleMessage(&message, "host"); err != nil {
				m.logger.Error().Err(err).Str("messageType", string(message.Type)).Msg("nakama: Failed to handle message from host")
			}

			// Reset read deadline
			_ = hostConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

// clientPingRoutine sends ping messages to the host
func (m *Manager) clientPingRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.hostMu.RLock()
			if m.hostConnection == nil || !m.hostConnection.Authenticated {
				m.hostMu.RUnlock()
				return
			}

			// Check if host is still alive
			if time.Since(m.hostConnection.LastPing) > 90*time.Second {
				m.logger.Warn().Msg("nakama: Host connection timeout")
				m.hostConnection.Close()
				m.hostMu.RUnlock()
				return
			}

			// Send ping
			message := &Message{
				Type:      MessageTypePing,
				Payload:   nil,
				Timestamp: time.Now(),
			}

			if err := m.hostConnection.SendMessage(message); err != nil {
				m.logger.Error().Err(err).Msg("nakama: Failed to send ping to host")
				m.hostConnection.Close()
				m.hostMu.RUnlock()
				return
			}
			m.hostMu.RUnlock()
		}
	}
}
