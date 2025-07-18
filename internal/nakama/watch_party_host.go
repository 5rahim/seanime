package nakama

import (
	"context"
	"errors"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/torrentstream"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Host
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type CreateWatchOptions struct {
	Settings *WatchPartySessionSettings `json:"settings"`
}

// CreateWatchParty creates a new watch party (host only)
func (wpm *WatchPartyManager) CreateWatchParty(options *CreateWatchOptions) (*WatchPartySession, error) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	if !wpm.manager.IsHost() {
		return nil, errors.New("only hosts can create watch parties")
	}

	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	wpm.logger.Debug().Msg("nakama: Creating watch party")

	wpm.sessionCtx, wpm.sessionCtxCancel = context.WithCancel(context.Background())

	// Generate unique ID
	sessionID := uuid.New().String()

	session := &WatchPartySession{
		ID:               sessionID,
		Participants:     make(map[string]*WatchPartySessionParticipant),
		CurrentMediaInfo: nil,
		Settings:         options.Settings,
		CreatedAt:        time.Now(),
	}

	// Add host as participant
	session.Participants["host"] = &WatchPartySessionParticipant{
		ID:         "host",
		Username:   wpm.manager.username,
		IsHost:     true,
		CanControl: true,
		IsReady:    true,
		LastSeen:   time.Now(),
		Latency:    0,
	}

	wpm.currentSession = mo.Some(session)

	// Reset sequence numbers for new session
	wpm.sequenceMu.Lock()
	wpm.sendSequence = 0
	wpm.lastRxSequence = 0
	wpm.sequenceMu.Unlock()

	// Notify all peers about the new watch party
	_ = wpm.manager.SendMessage(MessageTypeWatchPartyCreated, WatchPartyCreatedPayload{
		Session: session,
	})

	wpm.logger.Debug().Str("sessionId", sessionID).Msg("nakama: Watch party created")

	// Send websocket event to update the UI
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-wpm.sessionCtx.Done():
				wpm.logger.Debug().Msg("nakama: Watch party periodic broadcast stopped")
				return
			case <-ticker.C:
				// Broadcast the session state to all peers every 5 seconds
				// This is useful for peers that will join later
				wpm.broadcastSessionStateToPeers()
			}
		}
	}()

	go wpm.listenToPlaybackManager()
	// go wpm.listenToOnlineStreaming() // TODO

	return session, nil
}

// PromotePeerToRelayModeOrigin promotes a peer to be the origin for relay mode
func (wpm *WatchPartyManager) PromotePeerToRelayModeOrigin(peerId string) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	if !wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Str("peerId", peerId).Msg("nakama: Promoting peer to relay mode origin")

	session, ok := wpm.currentSession.Get()
	if !ok {
		wpm.logger.Warn().Msg("nakama: Cannot promote peer to relay mode origin, no active watch party session")
		return
	}

	// Check if the peer exists in the session
	participant, exists := session.Participants[peerId]
	if !exists {
		wpm.logger.Warn().Str("peerId", peerId).Msg("nakama: Cannot promote peer to relay mode origin, peer not found in session")
		return
	}

	// Set the IsRelayOrigin flag to true
	participant.IsRelayOrigin = true
	// Broadcast the updated session state to all peers
	session.mu.Lock()
	session.IsRelayMode = true
	session.mu.Unlock()

	wpm.logger.Debug().Str("peerId", peerId).Msg("nakama: Peer promoted to relay mode origin")

	wpm.broadcastSessionStateToPeers()
	wpm.sendSessionStateToClient()
}

func (wpm *WatchPartyManager) StopWatchParty() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	if !wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Stopping watch party")

	// Cancel any ongoing catch-up operations
	wpm.cancelCatchUp()

	// Reset buffering state and cancel any waitForPeersReady goroutine
	wpm.bufferMu.Lock()
	wpm.isWaitingForBuffers = false
	if wpm.waitForPeersCancel != nil {
		wpm.waitForPeersCancel()
		wpm.waitForPeersCancel = nil
	}
	wpm.bufferMu.Unlock()

	// Broadcast the stop event to all peers
	_ = wpm.manager.SendMessage(MessageTypeWatchPartyStopped, nil)

	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	wpm.broadcastSessionStateToPeers()
	wpm.sendSessionStateToClient()
}

// listenToPlaybackManager listens to the playback manager
func (wpm *WatchPartyManager) listenToPlaybackManager() {
	playbackSubscriber := wpm.manager.playbackManager.SubscribeToPlaybackStatus("nakama_watch_party")

	go func() {
		defer util.HandlePanicInModuleThen("nakama/listenToPlaybackManager", func() {})
		defer func() {
			wpm.logger.Debug().Msg("nakama: Stopping playback manager listener")
			go wpm.manager.playbackManager.UnsubscribeFromPlaybackStatus("nakama_watch_party")
		}()

		for {
			select {
			case <-wpm.sessionCtx.Done():
				wpm.logger.Debug().Msg("nakama: Stopping playback manager listener")
				return
			case event := <-playbackSubscriber.EventCh:
				_, ok := wpm.currentSession.Get()
				if !ok {
					continue
				}

				switch event := event.(type) {
				case playbackmanager.VideoStoppedEvent, playbackmanager.StreamStoppedEvent:
					// Reset
					wpm.logger.Debug().Msg("nakama: Playback stopped event received")

					wpm.bufferMu.Lock()
					wpm.isWaitingForBuffers = true
					wpm.bufferWaitStart = time.Now()
					// Cancel existing waitForPeersReady goroutine
					if wpm.waitForPeersCancel != nil {
						wpm.waitForPeersCancel()
						wpm.waitForPeersCancel = nil
					}
					wpm.bufferMu.Unlock()

					// Reset the current session media info
					wpm.mu.Lock()
					session, ok := wpm.currentSession.Get()
					if !ok {
						wpm.mu.Unlock()
						return
					}
					session.CurrentMediaInfo = nil
					wpm.mu.Unlock()

					// Broadcast the session state to all peers
					go wpm.broadcastSessionStateToPeers()

				case playbackmanager.PlaybackStatusChangedEvent:
					if event.State.MediaId == 0 {
						continue
					}

					go func(event playbackmanager.PlaybackStatusChangedEvent) {
						wpm.manager.playbackManager.PullStatus()

						streamType := "file"
						if event.Status.PlaybackType == mediaplayer.PlaybackTypeStream {
							if strings.Contains(event.Status.Filepath, "/api/v1/torrentstream") {
								streamType = "torrent"
							} else {
								streamType = "debrid"
							}
						}

						optionalTorrentStreamStartOptions, _ := wpm.manager.torrentstreamRepository.GetPreviousStreamOptions()

						streamPath := event.Status.Filepath
						newCurrentMediaInfo := &WatchPartySessionMediaInfo{
							MediaId:                           event.State.MediaId,
							EpisodeNumber:                     event.State.EpisodeNumber,
							AniDBEpisode:                      event.State.AniDbEpisode,
							StreamType:                        streamType,
							StreamPath:                        streamPath,
							OptionalTorrentStreamStartOptions: optionalTorrentStreamStartOptions,
						}

						wpm.mu.Lock()
						session, ok := wpm.currentSession.Get()
						if !ok {
							wpm.mu.Unlock()
							return
						}

						// If this is the same media, just send the playback status
						if session.CurrentMediaInfo.Equals(newCurrentMediaInfo) && event.State.MediaId != 0 {
							wpm.mu.Unlock()

							// Get next sequence number for message ordering
							wpm.sequenceMu.Lock()
							wpm.sendSequence++
							sequenceNum := wpm.sendSequence
							wpm.sequenceMu.Unlock()

							// Send message
							_ = wpm.manager.SendMessage(MessageTypeWatchPartyPlaybackStatus, WatchPartyPlaybackStatusPayload{
								PlaybackStatus: event.Status,
								Timestamp:      time.Now().UnixNano(),
								SequenceNumber: sequenceNum,
								EpisodeNumber:  event.State.EpisodeNumber,
							})

						} else {
							// For new playback, update the session
							wpm.logger.Debug().Msgf("nakama: Playback changed or started: %s", streamPath)
							session.CurrentMediaInfo = newCurrentMediaInfo
							wpm.mu.Unlock()

							// Pause immediately and wait for peers to be ready
							_ = wpm.manager.playbackManager.Pause()

							// Reset buffering state for new playback
							wpm.bufferMu.Lock()
							wpm.isWaitingForBuffers = true
							wpm.bufferWaitStart = time.Now()

							// Cancel existing waitForPeersReady goroutine
							if wpm.waitForPeersCancel != nil {
								wpm.waitForPeersCancel()
								wpm.waitForPeersCancel = nil
							}
							wpm.bufferMu.Unlock()

							go wpm.broadcastSessionStateToPeers()

							// Start checking peer readiness
							go wpm.waitForPeersReady(func() {
								if !session.IsRelayMode {
									// resume playback
									_ = wpm.manager.playbackManager.Resume()
								} else {
									// in relay mode, just signal to the origin
									_ = wpm.manager.SendMessage(MessageTypeWatchPartyRelayModePeersReady, nil)
								}
							})
						}
					}(event)
				}
			}
		}
	}()
}

// broadcastSessionStateToPeers broadcasts the session state to all peers
func (wpm *WatchPartyManager) broadcastSessionStateToPeers() {
	session, ok := wpm.currentSession.Get()
	if !ok {
		_ = wpm.manager.SendMessage(MessageTypeWatchPartyStateChanged, WatchPartyStateChangedPayload{
			Session: nil,
		})
		return
	}

	_ = wpm.manager.SendMessage(MessageTypeWatchPartyStateChanged, WatchPartyStateChangedPayload{
		Session: session,
	})
}

func (wpm *WatchPartyManager) sendSessionStateToClient() {
	session, ok := wpm.currentSession.Get()
	if !ok {
		wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)
		return
	}

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)
}

// handleWatchPartyPeerJoinedEvent is called when a peer joins a watch party
func (wpm *WatchPartyManager) handleWatchPartyPeerJoinedEvent(payload *WatchPartyJoinPayload, timestamp time.Time) {
	if !wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Peer joined watch party")

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	session.mu.Lock()
	// Add the peer to the session
	session.Participants[payload.PeerId] = &WatchPartySessionParticipant{
		ID:         payload.PeerId,
		Username:   payload.Username,
		IsHost:     false,
		CanControl: false,
		IsReady:    false,
		LastSeen:   timestamp,
		Latency:    0,
		// Initialize buffering state
		IsBuffering:    false,
		BufferHealth:   1.0,
		PlaybackStatus: nil,
	}
	session.mu.Unlock()

	// Send session state
	go wpm.broadcastSessionStateToPeers()

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Updated watch party state after peer joined")

	wpm.sendSessionStateToClient()
}

// handleWatchPartyPeerLeftEvent is called when a peer leaves a watch party
func (wpm *WatchPartyManager) handleWatchPartyPeerLeftEvent(payload *WatchPartyLeavePayload) {
	if !wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Peer left watch party")

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	// Remove the peer from the session
	delete(session.Participants, payload.PeerId)

	// Send session state
	go wpm.broadcastSessionStateToPeers()

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Updated watch party state after peer left")

	wpm.sendSessionStateToClient()
}

// HandlePeerDisconnected handles peer disconnections and removes them from the watch party
func (wpm *WatchPartyManager) HandlePeerDisconnected(peerID string) {
	if !wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	// Check if the peer is in the watch party
	if _, exists := session.Participants[peerID]; !exists {
		return
	}

	wpm.logger.Debug().Str("peerId", peerID).Msg("nakama: Peer disconnected, removing from watch party")

	// Remove the peer from the session
	delete(session.Participants, peerID)

	// Send session state to remaining peers
	go wpm.broadcastSessionStateToPeers()

	wpm.logger.Debug().Str("peerId", peerID).Msg("nakama: Updated watch party state after peer disconnected")

	// Send websocket event to update the UI
	wpm.sendSessionStateToClient()
}

// handleWatchPartyPeerStatusEvent handles regular status reports from peers
func (wpm *WatchPartyManager) handleWatchPartyPeerStatusEvent(payload *WatchPartyPeerStatusPayload) {
	if !wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	session, ok := wpm.currentSession.Get()
	if !ok {
		wpm.mu.Unlock()
		return
	}

	// Update peer status
	if participant, exists := session.Participants[payload.PeerId]; exists {
		participant.PlaybackStatus = &payload.PlaybackStatus
		participant.IsBuffering = payload.IsBuffering
		participant.BufferHealth = payload.BufferHealth
		participant.LastSeen = payload.Timestamp
		participant.IsReady = !payload.IsBuffering && payload.BufferHealth > 0.1 // Consider ready if not buffering and has some buffer

		wpm.logger.Debug().
			Str("peerId", payload.PeerId).
			Bool("isBuffering", payload.IsBuffering).
			Float64("bufferHealth", payload.BufferHealth).
			Bool("isReady", participant.IsReady).
			Msg("nakama: Updated peer status")
	}
	wpm.mu.Unlock()

	// Check if we should start/resume playback based on peer states (call after releasing mutex)
	// Run this asynchronously to avoid blocking the event processing
	go wpm.checkAndManageBuffering()

	// Send session state to client to update the UI
	wpm.sendSessionStateToClient()
}

// handleWatchPartyBufferUpdateEvent handles buffer state changes from peers
func (wpm *WatchPartyManager) handleWatchPartyBufferUpdateEvent(payload *WatchPartyBufferUpdatePayload) {
	if !wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	session, ok := wpm.currentSession.Get()
	if !ok {
		wpm.mu.Unlock()
		return
	}

	// Update peer buffer status
	if participant, exists := session.Participants[payload.PeerId]; exists {
		participant.IsBuffering = payload.IsBuffering
		participant.BufferHealth = payload.BufferHealth
		participant.LastSeen = payload.Timestamp
		participant.IsReady = !payload.IsBuffering && payload.BufferHealth > 0.1

		wpm.logger.Debug().
			Str("peerId", payload.PeerId).
			Bool("isBuffering", payload.IsBuffering).
			Float64("bufferHealth", payload.BufferHealth).
			Bool("isReady", participant.IsReady).
			Msg("nakama: Updated peer buffer status")
	}
	wpm.mu.Unlock()

	// Immediately check if we need to pause/resume based on buffer state (call after releasing mutex)
	// Run this asynchronously to avoid blocking the event processing
	go wpm.checkAndManageBuffering()

	// Broadcast updated session state
	go wpm.broadcastSessionStateToPeers()

	// Send session state to client to update the UI
	wpm.sendSessionStateToClient()
}

// checkAndManageBuffering manages playback based on peer buffering states
// NOTE: This function should NOT be called while holding wpm.mu as it may need to acquire bufferMu
func (wpm *WatchPartyManager) checkAndManageBuffering() {
	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	// Get current playback status
	playbackStatus, hasPlayback := wpm.manager.playbackManager.PullStatus()
	if !hasPlayback {
		return
	}

	// Count peer states
	var totalPeers, readyPeers, bufferingPeers int
	for _, participant := range session.Participants {
		if !participant.IsHost {
			totalPeers++
			if participant.IsReady {
				readyPeers++
			}
			if participant.IsBuffering {
				bufferingPeers++
			}
		}
	}

	// No peers means no buffering management needed
	if totalPeers == 0 {
		return
	}

	wpm.bufferMu.Lock()
	defer wpm.bufferMu.Unlock()

	maxWaitTime := time.Duration(session.Settings.MaxBufferWaitTime) * time.Second

	// If any peer is buffering and we're playing, pause and wait
	if bufferingPeers > 0 && playbackStatus.Playing {
		if !wpm.isWaitingForBuffers {
			wpm.logger.Debug().
				Int("bufferingPeers", bufferingPeers).
				Int("totalPeers", totalPeers).
				Msg("nakama: Pausing playback due to peer buffering")

			_ = wpm.manager.playbackManager.Pause()
			wpm.isWaitingForBuffers = true
			wpm.bufferWaitStart = time.Now()
		}
		return
	}

	// If we're waiting for buffers
	if wpm.isWaitingForBuffers {
		waitTime := time.Since(wpm.bufferWaitStart)

		// Resume if all peers are ready or max wait time exceeded
		if bufferingPeers == 0 || waitTime > maxWaitTime {
			wpm.logger.Debug().
				Int("readyPeers", readyPeers).
				Int("totalPeers", totalPeers).
				Int("bufferingPeers", bufferingPeers).
				Float64("waitTimeSeconds", waitTime.Seconds()).
				Bool("maxWaitExceeded", waitTime > maxWaitTime).
				Msg("nakama: Resuming playback after buffer wait")

			_ = wpm.manager.playbackManager.Resume()
			wpm.isWaitingForBuffers = false
		}
	}
}

// waitForPeersReady waits for peers to be ready before resuming playback
func (wpm *WatchPartyManager) waitForPeersReady(onReady func()) {
	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	// Create cancellable context for this goroutine
	ctx, cancel := context.WithCancel(context.Background())

	wpm.bufferMu.Lock()
	wpm.waitForPeersCancel = cancel
	wpm.bufferMu.Unlock()

	defer func() {
		wpm.bufferMu.Lock()
		wpm.waitForPeersCancel = nil
		wpm.bufferMu.Unlock()
	}()

	maxWaitTime := time.Duration(session.Settings.MaxBufferWaitTime) * time.Second
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	wpm.logger.Debug().Msg("nakama: Waiting for peers to be ready")

	for {
		select {
		case <-ctx.Done():
			wpm.logger.Debug().Msg("nakama: waitForPeersReady cancelled")
			return
		case <-wpm.sessionCtx.Done():
			wpm.logger.Debug().Msg("nakama: Session ended while waiting for peers")
			return
		case <-ticker.C:
			wpm.bufferMu.Lock()

			// Check if we've been waiting too long
			waitTime := time.Since(wpm.bufferWaitStart)
			if waitTime > maxWaitTime {
				wpm.logger.Debug().Float64("waitTimeSeconds", waitTime.Seconds()).Msg("nakama: Max wait time exceeded, resuming playback")

				onReady()

				wpm.isWaitingForBuffers = false
				wpm.bufferMu.Unlock()
				return
			}

			// Count ready peers
			session, ok := wpm.currentSession.Get()
			if !ok {
				wpm.bufferMu.Unlock()
				return
			}

			var totalPeers, readyPeers int
			for _, participant := range session.Participants {
				if !participant.IsHost && !participant.IsRelayOrigin {
					totalPeers++
					if participant.IsReady {
						readyPeers++
					}
				}
			}

			// If no peers or all peers are ready, resume playback
			if totalPeers == 0 || readyPeers == totalPeers {
				wpm.logger.Debug().
					Int("readyPeers", readyPeers).
					Int("totalPeers", totalPeers).
					Msg("nakama: All peers are ready, resuming playback")

				onReady()

				wpm.isWaitingForBuffers = false
				wpm.bufferMu.Unlock()
				return
			}

			wpm.logger.Debug().
				Int("readyPeers", readyPeers).
				Int("totalPeers", totalPeers).
				Float64("waitTimeSeconds", waitTime.Seconds()).
				Msg("nakama: Still waiting for peers to be ready")

			wpm.bufferMu.Unlock()
		}
	}
}

func (wpm *WatchPartyManager) EnableRelayMode(peerId string) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Str("peerId", peerId).Msg("nakama: Enabling relay mode")

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	session.mu.Lock()
	participant, exists := session.Participants[peerId]
	if !exists {
		wpm.logger.Warn().Str("peerId", peerId).Msg("nakama: Cannot enable relay mode, peer not found in session")
		wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Peer not found in session")
		return
	}
	session.IsRelayMode = true
	participant.IsRelayOrigin = true
	session.mu.Unlock()

	wpm.logger.Debug().Str("peerId", peerId).Msg("nakama: Relay mode enabled")

	wpm.broadcastSessionStateToPeers()
	wpm.sendSessionStateToClient()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Relay mode
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// handleWatchPartyRelayModeOriginStreamStartedEvent is called when the relay origin sends us (the host) a new stream.
// It starts the same stream as the origin on the host by using the same options as the origin.
func (wpm *WatchPartyManager) handleWatchPartyRelayModeOriginStreamStartedEvent(payload *WatchPartyRelayModeOriginStreamStartedPayload) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Str("filepath", payload.Filepath).Msg("nakama: Relay mode origin stream started")

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	session.Settings.MaxBufferWaitTime = 60 // higher buffer wait time for relay mode

	event := payload

	// Load the stream on the host
	// Playback won't actually be started
	switch event.StreamType {
	case "file":
		// Do nothing, the file is already available
	case "torrent":
		// Start the torrent stream and wait for it to be ready
		options := *event.OptionalTorrentStreamStartOptions
		options.PlaybackType = torrentstream.PlaybackTypeNoneAndAwait
		err := wpm.manager.torrentstreamRepository.StartStream(context.Background(), &options)
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to start torrent stream")
		}
	case "debrid":
		// Start the debrid stream and wait for it to be ready
		options := *event.OptionalDebridStreamStartOptions
		options.PlaybackType = debrid_client.PlaybackTypeNoneAndAwait
		err := wpm.manager.debridClientRepository.StartStream(context.Background(), &options)
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to start debrid stream")
		}
	}

	// Update the current media info
	streamPath := event.Status.Filepath
	if event.StreamType == "file" {
		// For file streams, we should use the file path directly
		streamPath = event.OptionalLocalPath
	}
	newCurrentMediaInfo := &WatchPartySessionMediaInfo{
		MediaId:                           event.State.MediaId,
		EpisodeNumber:                     event.State.EpisodeNumber,
		AniDBEpisode:                      event.State.AniDbEpisode,
		StreamType:                        event.StreamType,
		StreamPath:                        streamPath,
		OptionalTorrentStreamStartOptions: event.OptionalTorrentStreamStartOptions,
	}

	// Video playback has started, send the media info to the peers
	session.CurrentMediaInfo = newCurrentMediaInfo

	// Pause immediately and wait for peers to be ready
	_ = wpm.manager.playbackManager.Pause()

	// Reset buffering state for new playback
	wpm.bufferMu.Lock()
	wpm.isWaitingForBuffers = true
	wpm.bufferWaitStart = time.Now()

	// Cancel existing waitForPeersReady goroutine
	if wpm.waitForPeersCancel != nil {
		wpm.waitForPeersCancel()
		wpm.waitForPeersCancel = nil
	}
	wpm.bufferMu.Unlock()

	// broadcast the session state to the peers
	// this will not include the relay origin
	wpm.broadcastSessionStateToPeers()

	// Start checking peer readiness
	go wpm.waitForPeersReady(func() {
		if !session.IsRelayMode {
			// not in relay mode, resume playback
			_ = wpm.manager.playbackManager.Resume()
		} else {
			// in relay mode, just signal to the origin
			_ = wpm.manager.SendMessage(MessageTypeWatchPartyRelayModePeersReady, nil)
		}
	})

}

// handleWatchPartyRelayModeOriginPlaybackStatusEvent is called when the relay origin sends us (the host) a playback status update
func (wpm *WatchPartyManager) handleWatchPartyRelayModeOriginPlaybackStatusEvent(payload *WatchPartyRelayModeOriginPlaybackStatusPayload) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	// wpm.logger.Debug().Msg("nakama: Relay mode origin playback status")

	// Send the playback status immediately to the peers
	// Get next sequence number for relayed message
	wpm.sequenceMu.Lock()
	wpm.sendSequence++
	sequenceNum := wpm.sendSequence
	wpm.sequenceMu.Unlock()

	_ = wpm.manager.SendMessage(MessageTypeWatchPartyPlaybackStatus, WatchPartyPlaybackStatusPayload{
		PlaybackStatus: payload.Status,
		Timestamp:      payload.Timestamp, // timestamp of the origin
		SequenceNumber: sequenceNum,
		EpisodeNumber:  payload.State.EpisodeNumber,
	})
}

// handleWatchPartyRelayModeOriginPlaybackStoppedEvent is called when the relay origin sends us (the host) a playback stopped event
func (wpm *WatchPartyManager) handleWatchPartyRelayModeOriginPlaybackStoppedEvent() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Msg("nakama: Relay mode origin playback stopped")

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	session.mu.Lock()
	session.CurrentMediaInfo = nil
	session.mu.Unlock()

	wpm.broadcastSessionStateToPeers()
	wpm.sendSessionStateToClient()
}
