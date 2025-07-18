package nakama

import (
	"encoding/json"
	"errors"
	"seanime/internal/events"
	"time"
)

// registerDefaultHandlers registers the default message handlers
func (m *Manager) registerDefaultHandlers() {
	m.messageHandlers[MessageTypeAuth] = m.handleAuthMessage
	m.messageHandlers[MessageTypeAuthReply] = m.handleAuthReplyMessage
	m.messageHandlers[MessageTypePing] = m.handlePingMessage
	m.messageHandlers[MessageTypePong] = m.handlePongMessage
	m.messageHandlers[MessageTypeError] = m.handleErrorMessage
	m.messageHandlers[MessageTypeCustom] = m.handleCustomMessage

	// Watch party handlers
	m.messageHandlers[MessageTypeWatchPartyCreated] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyStopped] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyJoin] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyLeave] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyStateChanged] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyPlaybackStatus] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyPlaybackStopped] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyPeerStatus] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyBufferUpdate] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyRelayModeOriginStreamStarted] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyRelayModeOriginPlaybackStatus] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyRelayModePeersReady] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyRelayModePeerBuffering] = m.handleWatchPartyMessage
	m.messageHandlers[MessageTypeWatchPartyRelayModeOriginPlaybackStopped] = m.handleWatchPartyMessage
}

// handleMessage routes messages to the appropriate handler
func (m *Manager) handleMessage(message *Message, senderID string) error {
	m.handlerMu.RLock()
	handler, exists := m.messageHandlers[message.Type]
	m.handlerMu.RUnlock()

	if !exists {
		return errors.New("unknown message type: " + string(message.Type))
	}

	return handler(message, senderID)
}

// handleAuthMessage handles authentication requests from peers
func (m *Manager) handleAuthMessage(message *Message, senderID string) error {
	if !m.settings.IsHost {
		return errors.New("not acting as host")
	}

	// Parse auth payload
	authData, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	var authPayload AuthPayload
	if err := json.Unmarshal(authData, &authPayload); err != nil {
		return err
	}

	// Get peer connection
	peerConn, exists := m.peerConnections.Get(senderID)
	if !exists {
		return errors.New("peer connection not found")
	}

	// Verify password
	success := authPayload.Password == m.settings.HostPassword
	var replyMessage string
	if success {
		// Update the peer connection with the PeerID from auth payload if not already set
		if peerConn.PeerId == "" && authPayload.PeerId != "" {
			peerConn.PeerId = authPayload.PeerId
		}

		peerConn.Authenticated = true
		replyMessage = "Authentication successful"
		m.logger.Info().Str("peerID", peerConn.PeerId).Str("senderID", senderID).Msg("nakama: Peer authenticated successfully")

		// Send event to client about new peer connection
		m.wsEventManager.SendEvent(events.NakamaPeerConnected, map[string]interface{}{
			"peerId":        peerConn.PeerId, // Use PeerID for events
			"authenticated": true,
		})
	} else {
		replyMessage = "Authentication failed"
		m.logger.Warn().Str("peerId", peerConn.PeerId).Str("senderID", senderID).Msg("nakama: Peer authentication failed")
	}

	// Send auth reply
	authReply := &Message{
		Type: MessageTypeAuthReply,
		Payload: AuthReplyPayload{
			Success:  success,
			Message:  replyMessage,
			Username: m.username,
			PeerId:   peerConn.PeerId, // Echo back the peer's UUID
		},
		Timestamp: time.Now(),
	}

	return peerConn.SendMessage(authReply)
}

// handleAuthReplyMessage handles authentication replies from hosts
func (m *Manager) handleAuthReplyMessage(message *Message, senderID string) error {
	// This should only be received by clients, and is handled in the client connection logic
	// We can log it here for debugging purposes
	m.logger.Debug().Str("senderID", senderID).Msg("nakama: Received auth reply")
	return nil
}

// handlePingMessage handles ping messages
func (m *Manager) handlePingMessage(message *Message, senderID string) error {
	// Send pong response
	pongMessage := &Message{
		Type:      MessageTypePong,
		Payload:   nil,
		Timestamp: time.Now(),
	}

	if m.settings.IsHost {
		// We're the host, send pong to peer
		peerConn, exists := m.peerConnections.Get(senderID)
		if !exists {
			return errors.New("peer connection not found")
		}
		return peerConn.SendMessage(pongMessage)
	} else {
		// We're a client, send pong to host
		m.hostMu.RLock()
		defer m.hostMu.RUnlock()
		if m.hostConnection == nil {
			return errors.New("not connected to host")
		}
		return m.hostConnection.SendMessage(pongMessage)
	}
}

// handlePongMessage handles pong messages
func (m *Manager) handlePongMessage(message *Message, senderID string) error {
	// Update last ping time
	if m.settings.IsHost {
		// Update peer's last ping time
		peerConn, exists := m.peerConnections.Get(senderID)
		if exists {
			peerConn.LastPing = time.Now()
		}
	} else {
		// Update host's last ping time
		m.hostMu.Lock()
		if m.hostConnection != nil {
			m.hostConnection.LastPing = time.Now()
		}
		m.hostMu.Unlock()
	}
	return nil
}

// handleErrorMessage handles error messages
func (m *Manager) handleErrorMessage(message *Message, senderID string) error {
	// Parse error payload
	errorData, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	var errorPayload ErrorPayload
	if err := json.Unmarshal(errorData, &errorPayload); err != nil {
		return err
	}

	m.logger.Error().Str("senderID", senderID).Str("errorMessage", errorPayload.Message).Str("errorCode", errorPayload.Code).Msg("nakama: Received error message")

	// Send event to client about the error
	m.wsEventManager.SendEvent(events.NakamaError, map[string]interface{}{
		"senderID": senderID,
		"message":  errorPayload.Message,
		"code":     errorPayload.Code,
	})

	return nil
}

// handleCustomMessage handles custom messages
func (m *Manager) handleCustomMessage(message *Message, senderID string) error {
	m.logger.Debug().Str("senderID", senderID).Msg("nakama: Received custom message")

	// Send event to client with the custom message
	m.wsEventManager.SendEvent(events.NakamaCustomMessage, map[string]interface{}{
		"senderID":  senderID,
		"payload":   message.Payload,
		"requestID": message.RequestID,
		"timestamp": message.Timestamp,
	})

	return nil
}

func (m *Manager) handleWatchPartyMessage(message *Message, senderID string) error {
	return m.watchPartyManager.handleMessage(message, senderID)
}
