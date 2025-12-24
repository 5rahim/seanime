package nakama

import (
	"context"
	"errors"
	"fmt"
	"math"
	"seanime/internal/events"
	"seanime/internal/torrentstream"
	"seanime/internal/util"
	"time"

	"github.com/samber/mo"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Peer
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (wpm *WatchPartyManager) JoinWatchParty(clientId string) error {
	if wpm.manager.IsHost() {
		return errors.New("only peers can join watch parties")
	}

	wpm.logger.Debug().Msg("nakama: Joining watch party")

	hostConn, ok := wpm.manager.GetHostConnection()
	if !ok {
		return errors.New("no host connection found")
	}

	_, ok = wpm.currentSession.Get() // session should exist
	if !ok {
		return errors.New("no watch party found")
	}

	// update the client
	wpm.clientId = clientId

	// Cancel any existing session context before creating a new one
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
	}

	wpm.sessionCtx, wpm.sessionCtxCancel = context.WithCancel(context.Background())

	// Reset sequence numbers for new session participation
	wpm.sequenceMu.Lock()
	wpm.sendSequence = 0
	wpm.lastRxSequence = 0
	wpm.sequenceMu.Unlock()

	// Send join message to host
	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyJoin, &WatchPartyJoinPayload{
		PeerId:   hostConn.PeerId,
		Username: wpm.manager.username,
	})

	// Start status reporting to host
	wpm.startStatusReporting()

	// Send websocket event to update the UI
	wpm.sendSessionStateToClient()

	// Start listening to players for relay mode
	wpm.relayModeListenToPlayerAsOrigin()

	return nil
}

// startStatusReporting starts sending status updates to the host every 2 seconds
// When a watch party is started, it'll tell the host that the peer is ready
func (wpm *WatchPartyManager) startStatusReporting() {
	if wpm.manager.IsHost() {
		return
	}

	// Stop any existing status reporting
	wpm.stopStatusReporting()

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	// Reset buffering detection state
	wpm.bufferDetectionMu.Lock()
	wpm.lastPosition = 0
	wpm.lastPositionTime = time.Time{}
	wpm.stallCount = 0
	wpm.bufferDetectionMu.Unlock()

	// Create context for status reporting
	ctx, cancel := context.WithCancel(context.Background())
	wpm.statusReportCancel = cancel

	// Start ticker for regular status reports
	wpm.statusReportTicker = time.NewTicker(2 * time.Second)

	go func() {
		defer util.HandlePanicInModuleThen("nakama/startStatusReporting", func() {})
		defer wpm.statusReportTicker.Stop()

		hostConn, ok := wpm.manager.GetHostConnection()
		if !ok {
			return
		}

		wpm.logger.Debug().Msg("nakama: Started status reporting to host")

		for {
			select {
			case <-ctx.Done():
				wpm.logger.Debug().Msg("nakama: Stopped status reporting")
				return
			case <-wpm.statusReportTicker.C:
				wpm.sendStatusToHost(hostConn.PeerId)
			}
		}
	}()
}

// stopStatusReporting stops sending status updates to the host
func (wpm *WatchPartyManager) stopStatusReporting() {
	if wpm.statusReportCancel != nil {
		wpm.statusReportCancel()
		wpm.statusReportCancel = nil
	}

	if wpm.statusReportTicker != nil {
		wpm.statusReportTicker.Stop()
		wpm.statusReportTicker = nil
	}
}

// sendStatusToHost sends current playback status and buffer state to the host
func (wpm *WatchPartyManager) sendStatusToHost(peerId string) {
	playbackStatus, hasPlayback := wpm.manager.genericPlayer.PullStatus()
	if !hasPlayback {
		return
	}

	// Calculate buffer health and buffering state
	isBuffering, bufferHealth := wpm.calculateBufferState(playbackStatus)

	// Send peer status update
	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyPeerStatus, &WatchPartyPeerStatusPayload{
		PeerId:          peerId,
		PlaybackStatus:  playbackStatus,
		IsBuffering:     isBuffering,
		BufferHealth:    bufferHealth,
		UseDenshiPlayer: wpm.manager.GetUseDenshiPlayer(),
		Timestamp:       time.Now(),
	})
}

// calculateBufferState calculates buffering state and buffer health from playback status
func (wpm *WatchPartyManager) calculateBufferState(status *WatchPartyPlaybackStatus) (bool, float64) {
	if status == nil {
		return true, 0.0 // No status means we're probably buffering
	}

	wpm.bufferDetectionMu.Lock()
	defer wpm.bufferDetectionMu.Unlock()

	now := time.Now()
	currentPosition := status.CurrentTime

	// Initialize tracking on first call
	if wpm.lastPositionTime.IsZero() {
		wpm.lastPosition = currentPosition
		wpm.lastPositionTime = now
		wpm.stallCount = 0
		return false, 1.0 // Assume good state initially
	}

	// Time since last position check
	timeDelta := now.Sub(wpm.lastPositionTime).Seconds()
	positionDelta := currentPosition - wpm.lastPosition

	// Update tracking
	wpm.lastPosition = currentPosition
	wpm.lastPositionTime = now

	// Don't check too frequently to avoid false positives
	if timeDelta < BufferDetectionMinInterval {
		return false, 1.0 // Return good state if checking too soon
	}

	// Check if we're at the end of the content
	isAtEnd := currentPosition >= (status.Duration - EndOfContentThreshold)
	if isAtEnd {
		// Reset stall count when at end
		wpm.stallCount = 0
		return false, 1.0 // Not buffering if we're at the end
	}

	// Handle seeking, if position jumped significantly, reset tracking
	if math.Abs(positionDelta) > SignificantPositionJump { // Detect seeking vs normal playback
		wpm.logger.Debug().
			Float64("positionDelta", positionDelta).
			Float64("currentPosition", currentPosition).
			Msg("nakama: Position change detected, likely seeking, resetting stall tracking")
		wpm.stallCount = 0
		return false, 1.0 // Reset state after seeking
	}

	// If the player is playing but position hasn't advanced significantly
	if !status.Paused {
		// Expected minimum position change
		expectedMinChange := timeDelta * BufferDetectionTolerance

		if positionDelta < expectedMinChange {
			// Position hasn't advanced as expected while playing, likely buffering
			wpm.stallCount++

			// Consider buffering after threshold consecutive stalls to avoid false positives
			isBuffering := wpm.stallCount >= BufferDetectionStallThreshold

			// Buffer health decreases with consecutive stalls
			bufferHealth := math.Max(0.0, 1.0-(float64(wpm.stallCount)*BufferHealthDecrement))

			if isBuffering {
				wpm.logger.Debug().
					Int("stallCount", wpm.stallCount).
					Float64("positionDelta", positionDelta).
					Float64("expectedMinChange", expectedMinChange).
					Float64("bufferHealth", bufferHealth).
					Msg("nakama: Buffering detected, position not advancing while playing")
			}

			return isBuffering, bufferHealth
		} else {
			// Position is advancing normally, reset stall count
			if wpm.stallCount > 0 {
				wpm.logger.Debug().
					Int("previousStallCount", wpm.stallCount).
					Float64("positionDelta", positionDelta).
					Msg("nakama: Playback resumed normally, resetting stall count")
			}
			wpm.stallCount = 0
			return false, 0.95 // good buffer health when playing normally
		}
	} else {
		// Player is paused, reset stall count and return good buffer state
		if wpm.stallCount > 0 {
			wpm.logger.Debug().Msg("nakama: Player paused, resetting stall count")
		}
		wpm.stallCount = 0
		return false, 1.0
	}
}

// resetBufferingState resets the buffering detection state (useful when playback changes)
func (wpm *WatchPartyManager) resetBufferingState() {
	wpm.bufferDetectionMu.Lock()
	defer wpm.bufferDetectionMu.Unlock()

	wpm.lastPosition = 0
	wpm.lastPositionTime = time.Time{}
	wpm.stallCount = 0
	wpm.logger.Debug().Msg("nakama: Reset buffering detection state")
}

// LeaveWatchParty signals to the host that the peer is leaving the watch party.
// The host will remove the peer from the session and the peer will receive a new session state.
// DEVNOTE: We don't remove the session from the manager, it should still exist.
func (wpm *WatchPartyManager) LeaveWatchParty() error {
	if wpm.manager.IsHost() {
		return errors.New("only peers can leave watch parties")
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Msg("nakama: Leaving watch party")

	// Stop status reporting
	wpm.stopStatusReporting()

	// Cancel the session context
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
	}

	hostConn, ok := wpm.manager.GetHostConnection()
	if !ok {
		return errors.New("no host connection found")
	}

	_, ok = wpm.currentSession.Get() // session should exist
	if !ok {
		return errors.New("no watch party found")
	}

	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyLeave, &WatchPartyLeavePayload{
		PeerId: hostConn.PeerId,
	})

	wpm.currentSession = mo.None[*WatchPartySession]()

	// Send websocket event to update the UI (nil indicates session left)
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Events
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var NakamaPeerListenerID = "nakama_peer_playback_listener"

// handleWatchPartyStateChangedEvent is called when the host updates the session state.
// It starts a stream on the peer if there's a new media info.
func (wpm *WatchPartyManager) handleWatchPartyStateChangedEvent(payload *WatchPartyStateChangedPayload) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	hostConn, ok := wpm.manager.GetHostConnection() // should always be ok
	if !ok {
		return
	}

	//
	// Session didn't exist
	//

	// Immediately update the session if it doesn't exist
	if _, exists := wpm.currentSession.Get(); !exists && payload.Session != nil {
		wpm.currentSession = mo.Some(&WatchPartySession{}) // Add a placeholder session
	}

	currentSession, exists := wpm.currentSession.Get()
	if !exists {
		return
	}

	//
	// Session destroyed
	//

	if payload.Session == nil {
		wpm.logger.Debug().Msg("nakama: Session destroyed")
		if wpm.sessionCtxCancel != nil {
			wpm.sessionCtxCancel()
			wpm.sessionCtx = nil
			wpm.sessionCtxCancel = nil
		}
		// Stop playback if it's playing
		if _, ok := currentSession.Participants[hostConn.PeerId]; ok {
			wpm.logger.Debug().Msg("nakama: Stopping playback due to session destroyed")
			wpm.manager.genericPlayer.Cancel()
		}
		wpm.currentSession = mo.None[*WatchPartySession]()
		wpm.sendSessionStateToClient()
		return
	}

	// \/ Below, session should exist

	participant, isParticipant := payload.Session.Participants[hostConn.PeerId]

	//
	// Starting playback / Peer joined / Video changed
	//

	// If the payload session has a media info but the current session doesn't,
	// and the peer is a participant, we need to start playback
	newPlayback := payload.Session.CurrentMediaInfo != nil && currentSession.CurrentMediaInfo == nil
	playbackChanged := payload.Session.CurrentMediaInfo != nil && !payload.Session.CurrentMediaInfo.Equals(currentSession.CurrentMediaInfo)

	// Check if peer is newly a participant - they should start playback even if media info hasn't changed
	wasParticipant := currentSession.Participants != nil && currentSession.Participants[hostConn.PeerId] != nil
	peerJoined := isParticipant && !wasParticipant && payload.Session.CurrentMediaInfo != nil

	if (newPlayback || playbackChanged || peerJoined) &&
		isParticipant &&
		!participant.IsRelayOrigin {
		wpm.logger.Debug().Bool("newPlayback", newPlayback).Bool("playbackChanged", playbackChanged).Bool("peerJoined", peerJoined).Msg("nakama: Starting playback due to new media info")

		// Reset buffering detection state for new media
		wpm.resetBufferingState()
		// Reset the player params
		wpm.manager.genericPlayer.Reset()

		// Fetch the media info
		media, err := wpm.manager.platformRef.Get().GetAnime(context.Background(), payload.Session.CurrentMediaInfo.MediaId)
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to fetch media info for watch party")
			return
		}

		// Start the media on the peer
		wpm.logger.Debug().Int("mediaId", payload.Session.CurrentMediaInfo.MediaId).Msg("nakama: Starting watch party media")

		switch payload.Session.CurrentMediaInfo.StreamType {
		case WatchPartyStreamTypeTorrent:
			// If the params are missing, the host didn't return them
			if payload.Session.CurrentMediaInfo.TorrentStreamParams == nil {
				wpm.logger.Error().Msg("nakama: No torrent stream start options found")
				wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Watch party: Failed to play media: Host did not return torrent stream start options")
				return
			}
			if !wpm.manager.torrentstreamRepository.IsEnabled() {
				wpm.logger.Error().Msg("nakama: Torrent streaming is not enabled")
				wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Watch party: Failed to play media: Torrent streaming is not enabled")
				return
			}
			// Overwrite the player used and client ID
			payload.Session.CurrentMediaInfo.TorrentStreamParams.ClientId = wpm.clientId
			if wpm.manager.GetUseDenshiPlayer() {
				payload.Session.CurrentMediaInfo.TorrentStreamParams.PlaybackType = torrentstream.PlaybackTypeNativePlayer
			}

			wpm.logger.Debug().Interface("params", payload.Session.CurrentMediaInfo.TorrentStreamParams).Msg("nakama: Starting torrent stream")

			// Start the torrent
			err = wpm.manager.torrentstreamRepository.StartStream(wpm.sessionCtx, payload.Session.CurrentMediaInfo.TorrentStreamParams)
		case WatchPartyStreamTypeDebrid:
			// Start the debrid stream, which is just the current stream the host is playing
			err = wpm.manager.PlayHostAnimeStream(payload.Session.CurrentMediaInfo.StreamType, "seanime/nakama", wpm.clientId, media, payload.Session.CurrentMediaInfo.AniDBEpisode)
		case WatchPartyStreamTypeFile:
			// Start the local file stream off of the host using the file path
			err = wpm.manager.PlayHostAnimeLibraryFile(payload.Session.CurrentMediaInfo.LocalFilePath, "seanime/nakama", wpm.clientId, media, payload.Session.CurrentMediaInfo.AniDBEpisode, "")
		case WatchPartyStreamTypeOnlinestream:
			if payload.Session.CurrentMediaInfo.OnlinestreamParams == nil {
				wpm.logger.Error().Msg("nakama: No onlinestream params found")
				wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Watch party: Failed to play media: Host did not return onlinestream params")
				return
			}
			// Since it's an online stream force the current player to VideoCore
			wpm.manager.genericPlayer.SetType(WatchPartyVideoCore)
			// Start the onlinestream using the params
			wpm.manager.videoCore.StartOnlinestreamWatchParty(payload.Session.CurrentMediaInfo.OnlinestreamParams)
		}
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to play watch party media")
			wpm.manager.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Watch party: Failed to play media: %s", err.Error()))
		}

		//// Auto-leave the watch party when playback stops
		//// The user will have to re-join to start the stream again
		//if payload.Session.CurrentMediaInfo.StreamType != WatchPartyStreamTypeOnlinestream && !participant.IsRelayOrigin {
		//	// Clean up old listener
		//	if wpm.peerPlaybackListener != nil {
		//		wpm.manager.genericPlayer.Unsubscribe(NakamaPeerListenerID)
		//		wpm.peerPlaybackListener = nil
		//	}
		//
		//	wpm.peerPlaybackListener = wpm.manager.genericPlayer.Subscribe(NakamaPeerListenerID)
		//	go func() {
		//		defer util.HandlePanicInModuleThen("nakama/handleWatchPartyStateChangedEvent/autoLeaveWatchParty", func() {})
		//
		//		for {
		//			select {
		//			case <-wpm.sessionCtx.Done():
		//				wpm.manager.genericPlayer.Unsubscribe(NakamaPeerListenerID)
		//				return
		//			case event, ok := <-wpm.peerPlaybackListener.EventCh:
		//				if !ok {
		//					return
		//				}
		//				switch event.(type) {
		//				case *WatchPartyPlayerVideoEnded:
		//					_ = wpm.LeaveWatchParty()
		//					return
		//				}
		//			}
		//		}
		//	}()
		//}
	}

	//
	// Peer left
	//

	canceledPlayback := false

	// If the peer is a participant in the current session but the new session doesn't have them,
	// we need to stop playback and status reporting
	if _, ok := currentSession.Participants[hostConn.PeerId]; ok && payload.Session.Participants[hostConn.PeerId] == nil {
		wpm.logger.Debug().Msg("nakama: Removing peer from session due to new session state")
		// Stop status reporting when removed from session
		wpm.stopStatusReporting()
		// Before stopping playback, unsubscribe from the playback listener
		// This is to prevent the peer from auto-leaving the watch party when host stops playback
		if wpm.peerPlaybackListener != nil {
			wpm.manager.genericPlayer.Unsubscribe(NakamaPeerListenerID)
			wpm.peerPlaybackListener = nil
		}
		wpm.manager.genericPlayer.Cancel()
		canceledPlayback = true
	}

	//
	// Session stopped
	//

	// If the host stopped the session, we need to cancel playback
	if payload.Session.CurrentMediaInfo == nil && currentSession.CurrentMediaInfo != nil && !canceledPlayback {
		wpm.logger.Debug().Msg("nakama: Canceling playback due to host stopping session")
		// Before stopping playback, unsubscribe from the playback listener
		// This is to prevent the peer from auto-leaving the watch party when host stops playback
		if wpm.peerPlaybackListener != nil {
			wpm.manager.genericPlayer.Unsubscribe(NakamaPeerListenerID)
			wpm.peerPlaybackListener = nil
		}
		wpm.manager.genericPlayer.Cancel()
		canceledPlayback = true
	}

	// Update the session
	wpm.currentSession = mo.Some(payload.Session)
	wpm.sendSessionStateToClient()
}

// handleWatchPartyCreatedEvent is called when a host creates a watch party
// We cancel any existing session
// We just store the session in the manager, and the peer will decide whether to join or not
func (wpm *WatchPartyManager) handleWatchPartyCreatedEvent(payload *WatchPartyCreatedPayload) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Host created watch party")

	// Cancel any existing session
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	// Load the session into the manager
	// even if the peer isn't a participant
	wpm.currentSession = mo.Some(payload.Session)

	wpm.sendSessionStateToClient()
}

// handleWatchPartyStoppedEvent is called when the host stops a watch party.
//
// We check if the user was a participant in an active watch party session.
// If yes, we will cancel playback.
func (wpm *WatchPartyManager) handleWatchPartyStoppedEvent() {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Host stopped watch party")

	// Stop status reporting
	wpm.stopStatusReporting()

	// Cancel any ongoing catch-up operations
	wpm.cancelCatchUp()

	hostConn, ok := wpm.manager.GetHostConnection() // should always be ok
	if !ok {
		return
	}

	// Cancel playback if the user was a participant in any previous session
	currentSession, ok := wpm.currentSession.Get()
	if ok {
		if _, ok := currentSession.Participants[hostConn.PeerId]; ok {
			wpm.manager.genericPlayer.Cancel()
		}
	}

	// Cancel any existing session
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Relay mode
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// relayModeListenToPlayerAsOrigin starts listening to players when in relay mode.
// If the user is the relay origin, we listen to playback started events to send it to the relay host.
func (wpm *WatchPartyManager) relayModeListenToPlayerAsOrigin() {
	go func() {
		id := "nakama:relay-origin"
		defer util.HandlePanicInModuleThen("nakama/relayModeListenToPlayerAsOrigin", func() {})

		wpm.logger.Debug().Msg("nakama: Started listening to players for relay mode")

		// Subscribe to players
		playbackSubscriber := wpm.manager.genericPlayer.Subscribe(id)
		defer wpm.manager.genericPlayer.Unsubscribe(id)

		newStream := false
		streamStartedPayload := WatchPartyRelayModeOriginStreamStartedPayload{}

		for {
			select {
			case <-wpm.sessionCtx.Done():
				wpm.logger.Debug().Msg("nakama: Stopped listening to players for relay mode")
				return
			case e, ok := <-playbackSubscriber.EventCh:
				if !ok {
					// Channel closed, exit
					wpm.logger.Debug().Msg("nakama: Playback subscriber channel closed")
					return
				}

				currentSession, ok := wpm.currentSession.Get()
				if !ok {
					continue
				}

				hostConn, ok := wpm.manager.GetHostConnection()
				if !ok {
					continue
				}

				currentSession.mu.Lock()
				if !currentSession.IsRelayMode {
					currentSession.mu.Unlock()
					continue
				}

				participant, ok := currentSession.Participants[hostConn.PeerId]
				if !ok {
					currentSession.mu.Unlock()
					continue
				}

				if !participant.IsRelayOrigin {
					currentSession.mu.Unlock()
					continue
				}

				switch event := e.(type) {
				// 1. Relay origin stream started
				case *WatchPartyPlayerVideoStarted:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream started")

					newStream = true
					streamStartedPayload = WatchPartyRelayModeOriginStreamStartedPayload{}

					// immediately pause the playback
					wpm.manager.genericPlayer.Pause()

					if event.StreamType == WatchPartyStreamTypeFile {
						streamStartedPayload.LocalFilePath = wpm.manager.previousPath
					} else if event.StreamType == WatchPartyStreamTypeTorrent {
						streamStartedPayload.TorrentStreamParams, _ = wpm.manager.torrentstreamRepository.GetPreviousStreamOptions()
					} else if event.StreamType == WatchPartyStreamTypeDebrid {
						streamStartedPayload.DebridStreamParams, _ = wpm.manager.debridClientRepository.GetPreviousStreamOptions()
					} else if event.StreamType == WatchPartyStreamTypeOnlinestream {
						state, ok := wpm.manager.videoCore.GetPlaybackState()
						if !ok {
							wpm.logger.Error().Msg("nakama: Failed to get playback state for online stream")
							currentSession.mu.Unlock()
							continue
						}
						params := state.PlaybackInfo.OnlinestreamParams
						if params == nil {
							wpm.logger.Error().Msg("nakama: Online stream playback state missing params")
							currentSession.mu.Unlock()
							continue
						}
						streamStartedPayload.OnlinestreamParams = params
					}
					currentSession.mu.Unlock()

				// 2. Stream status changed
				case *WatchPartyPlayerVideoStatus:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream status changed")

					if newStream {
						newStream = false

						// relay origin started a new stream, send the payload to the relay host
						_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginStreamStarted, &WatchPartyRelayModeOriginStreamStartedPayload{
							Filename:            event.Filename,
							Filepath:            event.Filepath,
							StreamType:          event.State.StreamType,
							LocalFilePath:       streamStartedPayload.LocalFilePath,
							TorrentStreamParams: streamStartedPayload.TorrentStreamParams,
							DebridStreamParams:  streamStartedPayload.DebridStreamParams,
							OnlinestreamParams:  streamStartedPayload.OnlinestreamParams,
							Status:              event.Status,
							State:               event.State,
						})
						currentSession.mu.Unlock()
						continue
					}

					// send the playback status to the relay host
					_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginPlaybackStatus, &WatchPartyRelayModeOriginPlaybackStatusPayload{
						Status:    event.Status,
						State:     event.State,
						Timestamp: time.Now().UnixNano(),
					})
					currentSession.mu.Unlock()

				// 3. Stream stopped
				case *WatchPartyPlayerVideoEnded:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream stopped")
					_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginPlaybackStopped, nil)
					currentSession.mu.Unlock()
				}
			}
		}
	}()
}

// handleWatchPartyRelayModePeersReadyEvent is called when the host signals that the peers are ready in relay mode
func (wpm *WatchPartyManager) handleWatchPartyRelayModePeersReadyEvent() {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Relay mode peers ready")

	// resume playback
	wpm.manager.genericPlayer.Resume()
}
