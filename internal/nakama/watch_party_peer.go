package nakama

import (
	"context"
	"errors"
	"fmt"
	"math"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/samber/mo"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Peer
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (wpm *WatchPartyManager) JoinWatchParty() error {
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

	wpm.sessionCtx, wpm.sessionCtxCancel = context.WithCancel(context.Background())

	// Reset sequence numbers for new session participation
	wpm.sequenceMu.Lock()
	wpm.sendSequence = 0
	wpm.lastRxSequence = 0
	wpm.sequenceMu.Unlock()

	// Send join message to host
	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyJoin, WatchPartyJoinPayload{
		PeerId:   hostConn.PeerId,
		Username: wpm.manager.username,
	})

	// Start status reporting to host
	wpm.startStatusReporting()

	// Send websocket event to update the UI
	wpm.sendSessionStateToClient()

	// Start listening to playback manager
	wpm.relayModeListenToPlaybackManager()

	return nil
}

// startStatusReporting starts sending status updates to the host every 2 seconds
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
	playbackStatus, hasPlayback := wpm.manager.playbackManager.PullStatus()
	if !hasPlayback {
		return
	}

	// Calculate buffer health and buffering state
	isBuffering, bufferHealth := wpm.calculateBufferState(playbackStatus)

	// Send peer status update
	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyPeerStatus, WatchPartyPeerStatusPayload{
		PeerId:         peerId,
		PlaybackStatus: *playbackStatus,
		IsBuffering:    isBuffering,
		BufferHealth:   bufferHealth,
		Timestamp:      time.Now(),
	})
}

// calculateBufferState calculates buffering state and buffer health from playback status
func (wpm *WatchPartyManager) calculateBufferState(status *mediaplayer.PlaybackStatus) (bool, float64) {
	if status == nil {
		return true, 0.0 // No status means we're probably buffering
	}

	wpm.bufferDetectionMu.Lock()
	defer wpm.bufferDetectionMu.Unlock()

	now := time.Now()
	currentPosition := status.CurrentTimeInSeconds

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
	isAtEnd := currentPosition >= (status.DurationInSeconds - EndOfContentThreshold)
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
	if status.Playing {
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

	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyLeave, WatchPartyLeavePayload{
		PeerId: hostConn.PeerId,
	})

	// Send websocket event to update the UI (nil indicates session left)
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Events
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
			_ = wpm.manager.playbackManager.Cancel()
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

		// Fetch the media info
		media, err := wpm.manager.platform.GetAnime(context.Background(), payload.Session.CurrentMediaInfo.MediaId)
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to fetch media info for watch party")
			return
		}

		// Play the media
		wpm.logger.Debug().Int("mediaId", payload.Session.CurrentMediaInfo.MediaId).Msg("nakama: Playing watch party media")

		switch payload.Session.CurrentMediaInfo.StreamType {
		case "torrent":
			if payload.Session.CurrentMediaInfo.OptionalTorrentStreamStartOptions == nil {
				wpm.logger.Error().Msg("nakama: No torrent stream start options found")
				wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Watch party: Failed to play media: Host did not return torrent stream start options")
				return
			}
			if !wpm.manager.torrentstreamRepository.IsEnabled() {
				wpm.logger.Error().Msg("nakama: Torrent streaming is not enabled")
				wpm.manager.wsEventManager.SendEvent(events.ErrorToast, "Watch party: Failed to play media: Torrent streaming is not enabled")
				return
			}
			// Start the torrent
			err = wpm.manager.torrentstreamRepository.StartStream(wpm.sessionCtx, payload.Session.CurrentMediaInfo.OptionalTorrentStreamStartOptions)
		case "debrid":
			err = wpm.manager.PlayHostAnimeStream(payload.Session.CurrentMediaInfo.StreamType, "seanime/nakama", media, payload.Session.CurrentMediaInfo.AniDBEpisode)
		case "file":
			err = wpm.manager.PlayHostAnimeLibraryFile(payload.Session.CurrentMediaInfo.StreamPath, "seanime/nakama", media, payload.Session.CurrentMediaInfo.AniDBEpisode)
		case "online":
			wpm.sendCommandToOnlineStream(OnlineStreamCommandStart, payload.Session.CurrentMediaInfo.OnlineStreamParams)
		}
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to play watch party media")
			wpm.manager.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Watch party: Failed to play media: %s", err.Error()))
		}

		// Auto-leave the watch party when playback stops
		// The user will have to re-join to start the stream again
		if payload.Session.CurrentMediaInfo.StreamType != "online" && !participant.IsRelayOrigin {
			wpm.peerPlaybackListener = wpm.manager.playbackManager.SubscribeToPlaybackStatus("nakama_peer_playback_listener")
			go func() {
				defer util.HandlePanicInModuleThen("nakama/handleWatchPartyStateChangedEvent/autoLeaveWatchParty", func() {})

				for {
					select {
					case <-wpm.sessionCtx.Done():
						wpm.manager.playbackManager.UnsubscribeFromPlaybackStatus("nakama_peer_playback_listener")
						return
					case event, ok := <-wpm.peerPlaybackListener.EventCh:
						if !ok {
							return
						}

						switch event.(type) {
						case playbackmanager.StreamStoppedEvent:
							_ = wpm.LeaveWatchParty()
							return
						}
					}
				}
			}()
		}
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
			wpm.manager.playbackManager.UnsubscribeFromPlaybackStatus("nakama_peer_playback_listener")
			wpm.peerPlaybackListener = nil
		}
		_ = wpm.manager.playbackManager.Cancel()
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
			wpm.manager.playbackManager.UnsubscribeFromPlaybackStatus("nakama_peer_playback_listener")
			wpm.peerPlaybackListener = nil
		}
		_ = wpm.manager.playbackManager.Cancel()
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
			_ = wpm.manager.playbackManager.Cancel()
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Relay mode
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// relayModeListenToPlaybackManager starts listening to the playback manager when in relay mode
func (wpm *WatchPartyManager) relayModeListenToPlaybackManager() {
	go func() {
		defer util.HandlePanicInModuleThen("nakama/relayModeListenToPlaybackManager", func() {})

		wpm.logger.Debug().Msg("nakama: Started listening to playback manager for relay mode")

		playbackSubscriber := wpm.manager.playbackManager.SubscribeToPlaybackStatus("nakama_peer_relay_mode")
		defer wpm.manager.playbackManager.UnsubscribeFromPlaybackStatus("nakama_peer_relay_mode")

		newStream := false
		streamStartedPayload := WatchPartyRelayModeOriginStreamStartedPayload{}

		for {
			select {
			case <-wpm.sessionCtx.Done():
				wpm.logger.Debug().Msg("nakama: Stopped listening to playback manager")
				return
			case event := <-playbackSubscriber.EventCh:
				currentSession, ok := wpm.currentSession.Get() // should always be ok
				if !ok {
					return
				}

				hostConn, ok := wpm.manager.GetHostConnection() // should always be ok
				if !ok {
					return
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

				switch event := event.(type) {
				// 1. Stream started
				case playbackmanager.StreamStartedEvent:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream started")

					newStream = true
					streamStartedPayload = WatchPartyRelayModeOriginStreamStartedPayload{}

					// immediately pause the playback
					_ = wpm.manager.playbackManager.Pause()

					streamStartedPayload.Filename = event.Filename
					streamStartedPayload.Filepath = event.Filepath

					if strings.Contains(streamStartedPayload.Filepath, "type=file") {
						streamStartedPayload.OptionalLocalPath = wpm.manager.previousPath
						streamStartedPayload.StreamType = "file"
					} else if strings.Contains(streamStartedPayload.Filepath, "/api/v1/torrentstream") {
						streamStartedPayload.StreamType = "torrent"
						streamStartedPayload.OptionalTorrentStreamStartOptions, _ = wpm.manager.torrentstreamRepository.GetPreviousStreamOptions()
					} else {
						streamStartedPayload.StreamType = "debrid"
						streamStartedPayload.OptionalDebridStreamStartOptions, _ = wpm.manager.debridClientRepository.GetPreviousStreamOptions()
					}

				// 2. Stream status changed
				case playbackmanager.PlaybackStatusChangedEvent:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream status changed")

					if newStream {
						newStream = false

						// this is a new stream, send the stream started payload
						_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginStreamStarted, WatchPartyRelayModeOriginStreamStartedPayload{
							Filename:                          streamStartedPayload.Filename,
							Filepath:                          streamStartedPayload.Filepath,
							StreamType:                        streamStartedPayload.StreamType,
							OptionalLocalPath:                 streamStartedPayload.OptionalLocalPath,
							OptionalTorrentStreamStartOptions: streamStartedPayload.OptionalTorrentStreamStartOptions,
							OptionalDebridStreamStartOptions:  streamStartedPayload.OptionalDebridStreamStartOptions,
							Status:                            event.Status,
							State:                             event.State,
						})
						currentSession.mu.Unlock()
						continue
					}

					// send the playback status to the relay host
					_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginPlaybackStatus, WatchPartyRelayModeOriginPlaybackStatusPayload{
						Status:    event.Status,
						State:     event.State,
						Timestamp: time.Now().UnixNano(),
					})

				// 3. Stream stopped
				case playbackmanager.StreamStoppedEvent:
					wpm.logger.Debug().Msg("nakama: Relay mode origin stream stopped")
					_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyRelayModeOriginPlaybackStopped, nil)
				}
				currentSession.mu.Unlock()
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
	_ = wpm.manager.playbackManager.Resume()
}
