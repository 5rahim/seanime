package nakama

import (
	"context"
	"math"
	"seanime/internal/mediaplayers/mediaplayer"
	"time"
)

// handleWatchPartyPlaybackStatusEvent is called when the host sends a playback status.
//
// We check if the peer is a participant in the session.
// If yes, we will update the playback status and sync the playback position.
func (wpm *WatchPartyManager) handleWatchPartyPlaybackStatusEvent(payload *WatchPartyPlaybackStatusPayload) {
	if wpm.manager.IsHost() {
		return
	}

	// wpm.logger.Debug().Msg("nakama: Received playback status from watch party")

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	hostConn, ok := wpm.manager.GetHostConnection()
	if !ok {
		return
	}

	if participant, isParticipant := session.Participants[hostConn.PeerId]; !isParticipant || participant.IsRelayOrigin {
		return
	}

	payloadStatus := payload.PlaybackStatus

	// If the peer's session doesn't have a media info, do nothing
	if session.CurrentMediaInfo == nil {
		return
	}

	// If the playback manager doesn't have a status, do nothing
	playbackStatus, ok := wpm.manager.playbackManager.PullStatus()
	if !ok {
		return
	}

	// Check if the message is too old to prevent acting on stale data
	wpm.sequenceMu.Lock()
	isStale := payload.SequenceNumber != 0 && payload.SequenceNumber <= wpm.lastRxSequence
	if payload.SequenceNumber > wpm.lastRxSequence {
		wpm.lastRxSequence = payload.SequenceNumber
	}
	wpm.sequenceMu.Unlock()

	if isStale {
		wpm.logger.Debug().Uint64("messageSeq", payload.SequenceNumber).Uint64("lastSeq", wpm.lastRxSequence).Msg("nakama: Ignoring stale playback status message (old sequence)")
		return
	}

	now := time.Now().UnixNano()
	driftNs := now - payload.Timestamp
	timeSinceMessage := float64(driftNs) / 1e9 // Convert to seconds
	if timeSinceMessage > 5 {                  // Clamp to a reasonable maximum delay
		timeSinceMessage = 0 // If it's more than 5 seconds, treat it as no delay
	}

	// Handle play/pause state changes
	if payloadStatus.Playing != playbackStatus.Playing {
		if payloadStatus.Playing {
			// Cancel any ongoing catch-up operation
			wpm.cancelCatchUp()

			// When host resumes, sync position before resuming if there's significant drift
			// Calculate where the host should be NOW, not when they resumed
			hostCurrentPosition := payloadStatus.CurrentTimeInSeconds + timeSinceMessage
			positionDrift := hostCurrentPosition - playbackStatus.CurrentTimeInSeconds

			// Check if we need to seek
			shouldSeek := false
			if positionDrift < 0 {
				// Peer is behind, always seek if beyond threshold
				shouldSeek = math.Abs(positionDrift) > ResumePositionDriftThreshold
			} else {
				// Peer is ahead, only seek backward if significantly ahead to prevent jitter
				// This prevents backward seeks when peer is slightly ahead due to pause message delay
				shouldSeek = positionDrift > ResumeAheadTolerance
			}

			if shouldSeek {
				// Calculate dynamic seek delay based on message timing
				dynamicDelay := time.Duration(timeSinceMessage*1000) * time.Millisecond
				if dynamicDelay < MinSeekDelay {
					dynamicDelay = MinSeekDelay
				}
				if dynamicDelay > MaxDynamicDelay {
					dynamicDelay = MaxDynamicDelay
				}

				// Predict where host will be when our seek takes effect
				seekPosition := hostCurrentPosition + dynamicDelay.Seconds()

				wpm.logger.Debug().
					Float64("positionDrift", positionDrift).
					Float64("hostCurrentPosition", hostCurrentPosition).
					Float64("seekPosition", seekPosition).
					Float64("peerPosition", playbackStatus.CurrentTimeInSeconds).
					Float64("dynamicDelay", dynamicDelay.Seconds()).
					Bool("peerAhead", positionDrift > 0).
					Msg("nakama: Host resumed, syncing position before resume")

				// Track pending seek
				now := time.Now()
				wpm.seekMu.Lock()
				wpm.pendingSeekTime = now
				wpm.pendingSeekPosition = seekPosition
				wpm.seekMu.Unlock()

				_ = wpm.manager.playbackManager.Seek(seekPosition)
			} else if positionDrift > 0 && positionDrift <= ResumeAheadTolerance {
				wpm.logger.Debug().
					Float64("positionDrift", positionDrift).
					Float64("hostCurrentPosition", hostCurrentPosition).
					Float64("peerPosition", playbackStatus.CurrentTimeInSeconds).
					Msg("nakama: Host resumed, peer slightly ahead, not seeking yet")
			}

			wpm.logger.Debug().Msg("nakama: Host resumed, resuming peer playback")
			_ = wpm.manager.playbackManager.Resume()
		} else {
			wpm.logger.Debug().Msg("nakama: Host paused, handling peer pause")
			wpm.handleHostPause(payloadStatus, *playbackStatus, timeSinceMessage)
		}
	}

	// Handle position sync for different state combinations
	if payloadStatus.Playing == playbackStatus.Playing {
		// Both in same state, use normal sync
		wpm.syncPlaybackPosition(payloadStatus, *playbackStatus, timeSinceMessage, session)
	} else if payloadStatus.Playing && !playbackStatus.Playing {
		// Host playing, peer paused, sync position and resume
		hostExpectedPosition := payloadStatus.CurrentTimeInSeconds + timeSinceMessage

		wpm.logger.Debug().
			Float64("hostPosition", hostExpectedPosition).
			Float64("peerPosition", playbackStatus.CurrentTimeInSeconds).
			Msg("nakama: Host is playing but peer is paused, syncing and resuming")

		// Resume and sync to host position
		_ = wpm.manager.playbackManager.Resume()

		// Track pending seek
		now := time.Now()
		wpm.seekMu.Lock()
		wpm.pendingSeekTime = now
		wpm.pendingSeekPosition = hostExpectedPosition
		wpm.seekMu.Unlock()

		_ = wpm.manager.playbackManager.Seek(hostExpectedPosition)
	} else if !payloadStatus.Playing && playbackStatus.Playing {
		// Host paused, peer playing, pause immediately
		wpm.logger.Debug().Msg("nakama: Host is paused but peer is playing, pausing immediately")

		// Cancel catch-up and pause
		wpm.cancelCatchUp()
		wpm.handleHostPause(payloadStatus, *playbackStatus, timeSinceMessage)
	}
}

// handleHostPause handles when the host pauses playback
func (wpm *WatchPartyManager) handleHostPause(hostStatus mediaplayer.PlaybackStatus, peerStatus mediaplayer.PlaybackStatus, timeSinceMessage float64) {
	// Cancel any ongoing catch-up operation
	wpm.cancelCatchUp()

	now := time.Now()

	// Calculate where the host actually paused based on dynamic timing
	hostActualPausePosition := hostStatus.CurrentTimeInSeconds
	// Don't add time compensation for pause position, the host has already paused

	// Calculate time difference considering message delay
	timeDifference := hostActualPausePosition - peerStatus.CurrentTimeInSeconds

	// If peer is significantly behind the host, let it catch up before pausing
	if timeDifference > CatchUpBehindThreshold {
		wpm.logger.Debug().Msgf("nakama: Host paused, peer behind by %.2f seconds, catching up", timeDifference)
		wpm.startCatchUp(hostActualPausePosition, timeSinceMessage)
	} else {
		// Peer is close enough or ahead, pause immediately with position correction
		// Use more aggressive sync threshold for pause operations
		if math.Abs(timeDifference) > PausePositionSyncThreshold {
			wpm.logger.Debug().
				Float64("hostPausePosition", hostActualPausePosition).
				Float64("peerPosition", peerStatus.CurrentTimeInSeconds).
				Float64("timeDifference", timeDifference).
				Float64("timeSinceMessage", timeSinceMessage).
				Msg("nakama: Host paused, syncing position before pause")

			// Track pending seek
			wpm.seekMu.Lock()
			wpm.pendingSeekTime = now
			wpm.pendingSeekPosition = hostActualPausePosition
			wpm.seekMu.Unlock()

			_ = wpm.manager.playbackManager.Seek(hostActualPausePosition)
		}
		_ = wpm.manager.playbackManager.Pause()
		wpm.logger.Debug().Msgf("nakama: Host paused, peer paused immediately (diff: %.2f)", timeDifference)
	}
}

// startCatchUp starts a catch-up operation to sync with the host's pause position
func (wpm *WatchPartyManager) startCatchUp(hostPausePosition float64, timeSinceMessage float64) {
	wpm.catchUpMu.Lock()
	defer wpm.catchUpMu.Unlock()

	// Cancel any existing catch-up
	if wpm.catchUpCancel != nil {
		wpm.catchUpCancel()
	}

	// Create a new context for this catch-up operation
	ctx, cancel := context.WithCancel(context.Background())
	wpm.catchUpCancel = cancel

	go func() {
		defer cancel()

		ticker := time.NewTicker(CatchUpTickInterval)
		defer ticker.Stop()

		maxCatchUpTime := MaxCatchUpDuration
		startTime := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// If catch-up is taking too long, force sync to host position
				if time.Since(startTime) > maxCatchUpTime {
					wpm.logger.Debug().Msg("nakama: Catch-up timeout, seeking to host position and pausing")

					// Seek to host position and pause
					now := time.Now()
					wpm.seekMu.Lock()
					wpm.pendingSeekTime = now
					wpm.pendingSeekPosition = hostPausePosition
					wpm.seekMu.Unlock()

					_ = wpm.manager.playbackManager.Seek(hostPausePosition)
					_ = wpm.manager.playbackManager.Pause()
					return
				}

				// Get current playback status
				currentStatus, ok := wpm.manager.playbackManager.PullStatus()
				if !ok {
					continue
				}

				// Check if we've reached or passed the host's pause position (with tighter tolerance)
				positionDiff := hostPausePosition - currentStatus.CurrentTimeInSeconds
				if positionDiff <= CatchUpToleranceThreshold {
					wpm.logger.Debug().Msgf("nakama: Caught up to host position %.2f (current: %.2f), pausing", hostPausePosition, currentStatus.CurrentTimeInSeconds)

					// Track pending seek
					now := time.Now()
					wpm.seekMu.Lock()
					wpm.pendingSeekTime = now
					wpm.pendingSeekPosition = hostPausePosition
					wpm.seekMu.Unlock()

					_ = wpm.manager.playbackManager.Seek(hostPausePosition)
					_ = wpm.manager.playbackManager.Pause()
					return
				}

				// Continue trying to catch up to host position

				wpm.logger.Debug().
					Float64("positionDiff", positionDiff).
					Float64("currentPosition", currentStatus.CurrentTimeInSeconds).
					Float64("hostPausePosition", hostPausePosition).
					Msg("nakama: Still catching up to host pause position")
			}
		}
	}()
}

// cancelCatchUp cancels any ongoing catch-up operation
func (wpm *WatchPartyManager) cancelCatchUp() {
	wpm.catchUpMu.Lock()
	defer wpm.catchUpMu.Unlock()

	if wpm.catchUpCancel != nil {
		wpm.catchUpCancel()
		wpm.catchUpCancel = nil
	}
}

// syncPlaybackPosition synchronizes playback position when both host and peer are in the same play/pause state
func (wpm *WatchPartyManager) syncPlaybackPosition(hostStatus mediaplayer.PlaybackStatus, peerStatus mediaplayer.PlaybackStatus, timeSinceMessage float64, session *WatchPartySession) {
	now := time.Now()

	// Ignore very old messages to prevent stale syncing
	if timeSinceMessage > MaxMessageAge {
		return
	}

	// Check if we have a pending seek operation, use dynamic compensation
	wpm.seekMu.Lock()
	hasPendingSeek := !wpm.pendingSeekTime.IsZero()
	timeSincePendingSeek := now.Sub(wpm.pendingSeekTime)
	pendingSeekPosition := wpm.pendingSeekPosition
	wpm.seekMu.Unlock()

	// Use dynamic compensation, if we have a pending seek, wait for at least the message delay time
	dynamicSeekDelay := time.Duration(timeSinceMessage*1000) * time.Millisecond
	if dynamicSeekDelay < MinSeekDelay {
		dynamicSeekDelay = MinSeekDelay // Minimum delay
	}
	if dynamicSeekDelay > MaxSeekDelay {
		dynamicSeekDelay = MaxSeekDelay // Maximum delay
	}

	// If we have a pending seek that's still in progress, don't sync
	if hasPendingSeek && timeSincePendingSeek < dynamicSeekDelay {
		wpm.logger.Debug().
			Float64("timeSincePendingSeek", timeSincePendingSeek.Seconds()).
			Float64("dynamicSeekDelay", dynamicSeekDelay.Seconds()).
			Float64("pendingSeekPosition", pendingSeekPosition).
			Msg("nakama: Ignoring sync, pending seek in progress")
		return
	}

	// Clear pending seek if it's been long enough
	if hasPendingSeek && timeSincePendingSeek >= dynamicSeekDelay {
		wpm.seekMu.Lock()
		wpm.pendingSeekTime = time.Time{}
		wpm.pendingSeekPosition = 0
		wpm.seekMu.Unlock()
	}

	// Dynamic compensation: Calculate where the host should be NOW based on their timestamp
	hostCurrentPosition := hostStatus.CurrentTimeInSeconds
	if hostStatus.Playing {
		// Add the exact time that has passed since the host's status was captured
		hostCurrentPosition += timeSinceMessage
	}

	// Calculate drift between peer and host's current position
	drift := hostCurrentPosition - peerStatus.CurrentTimeInSeconds
	driftAbs := drift
	if driftAbs < 0 {
		driftAbs = -driftAbs
	}

	// Get sync threshold from session settings
	syncThreshold := session.Settings.SyncThreshold
	// Clamp
	if syncThreshold < MinSyncThreshold {
		syncThreshold = MinSyncThreshold
	} else if syncThreshold > MaxSyncThreshold {
		syncThreshold = MaxSyncThreshold
	}

	// Check if we're in seek cooldown period
	timeSinceLastSeek := now.Sub(wpm.lastSeekTime)
	inCooldown := timeSinceLastSeek < wpm.seekCooldown

	// Use more aggressive thresholds for different drift ranges
	effectiveThreshold := syncThreshold
	if driftAbs > 3.0 { // Large drift - be very aggressive
		effectiveThreshold = syncThreshold * AggressiveSyncMultiplier
	} else if driftAbs > 1.5 { // Medium drift - be more aggressive
		effectiveThreshold = syncThreshold * ModerateSyncMultiplier
	}

	// Only sync if drift exceeds threshold and we're not in cooldown
	if driftAbs > effectiveThreshold && !inCooldown {
		// For the seek position, predict where the host will be when our seek takes effect
		// Use the dynamic delay we calculated based on actual network conditions
		seekPosition := hostCurrentPosition
		if hostStatus.Playing {
			// Add compensation for the time it will take for our seek to take effect
			seekPosition += dynamicSeekDelay.Seconds()
		}

		wpm.logger.Debug().
			Float64("drift", drift).
			Float64("hostOriginalPosition", hostStatus.CurrentTimeInSeconds).
			Float64("hostCurrentPosition", hostCurrentPosition).
			Float64("seekPosition", seekPosition).
			Float64("peerPosition", peerStatus.CurrentTimeInSeconds).
			Float64("timeSinceMessage", timeSinceMessage).
			Float64("dynamicSeekDelay", dynamicSeekDelay.Seconds()).
			Float64("effectiveThreshold", effectiveThreshold).
			Msg("nakama: Syncing playback position with dynamic compensation")

		// Track pending seek
		wpm.seekMu.Lock()
		wpm.pendingSeekTime = now
		wpm.pendingSeekPosition = seekPosition
		wpm.seekMu.Unlock()

		_ = wpm.manager.playbackManager.Seek(seekPosition)
		wpm.lastSeekTime = now
	}
}
