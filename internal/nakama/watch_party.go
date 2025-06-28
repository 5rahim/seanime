package nakama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/util"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

const (
	// Host -> Peer
	MessageTypeWatchPartyCreated             = "watch_party_created"                // Host creates a watch party
	MessageTypeWatchPartyStateChanged        = "watch_party_state_changed"          // Host or peer changes the state of the watch party
	MessageTypeWatchPartyStopped             = "watch_party_stopped"                // Host stops a watch party
	MessageTypeWatchPartyPlaybackInfo        = "watch_party_playback_info"          // Host is ready, sends playback info to peers
	MessageTypeWatchPartyPlaybackStatus      = "watch_party_playback_status"        // Host or peer sends playback status to peers (seek, play, pause, etc)
	MessageTypeWatchPartyPlaybackStopped     = "watch_party_playback_stopped"       // Peer sends playback stopped to host
	MessageTypeWatchPartyRelayModePeersReady = "watch_party_relay_mode_peers_ready" // Relay server signals that all peers are ready to origin
	// Peer -> Host
	MessageTypeWatchPartyJoin                          = "watch_party_join"                              // Peer joins a watch party
	MessageTypeWatchPartyLeave                         = "watch_party_leave"                             // Peer leaves a watch party
	MessageTypeWatchPartyPeerStatus                    = "watch_party_peer_status"                       // Peer reports their current status to host
	MessageTypeWatchPartyBufferUpdate                  = "watch_party_buffer_update"                     // Peer reports buffering state to host
	MessageTypeWatchPartyRelayModeOriginPlaybackStatus = "watch_party_relay_mode_origin_playback_status" // Relay origin sends playback status to relay server
)

const (
	// Drift detection and sync thresholds
	MinSyncThreshold         = 0.8 // Minimum sync threshold to prevent excessive seeking
	MaxSyncThreshold         = 5.0 // Maximum sync threshold for loose synchronization
	AggressiveSyncMultiplier = 0.4 // Multiplier for large drift (>3s) to sync aggressively
	ModerateSyncMultiplier   = 0.6 // Multiplier for medium drift (>1.5s) to sync more frequently

	// Sync timing and delays
	MinSeekDelay        = 200 * time.Millisecond // Minimum delay for seek operations
	MaxSeekDelay        = 600 * time.Millisecond // Maximum delay for seek operations
	DefaultSeekCooldown = 1 * time.Second        // Cooldown between consecutive seeks

	// Message staleness and processing
	MaxMessageAge             = 1.5 // Seconds to ignore stale sync messages
	PendingSeekWaitMultiplier = 1.0 // Multiplier for pending seek wait time

	// Position and state detection
	SignificantPositionJump      = 3.0 // Seconds to detect seeking vs normal playback
	ResumePositionDriftThreshold = 1.0 // Seconds of drift before syncing on resume
	ResumeAheadTolerance         = 2.0 // Seconds ahead tolerance to prevent jitter on resume
	PausePositionSyncThreshold   = 0.7 // Seconds of drift threshold for pause sync

	// Catch-up and buffering
	CatchUpBehindThreshold    = 2.0                    // Seconds behind before starting catch-up
	CatchUpToleranceThreshold = 0.5                    // Seconds within target to stop catch-up
	MaxCatchUpDuration        = 4 * time.Second        // Maximum duration for catch-up operations
	CatchUpTickInterval       = 200 * time.Millisecond // Interval for catch-up progress checks

	// Buffer detection (peer-side)
	BufferDetectionMinInterval    = 1.5  // Seconds between buffer health checks
	BufferDetectionTolerance      = 0.6  // Tolerance for playback progress detection
	BufferDetectionStallThreshold = 2    // Consecutive stalls before buffering detection
	BufferHealthDecrement         = 0.15 // Buffer health decrease per stall
	EndOfContentThreshold         = 2.0  // Seconds from end to disable buffering detection

	// Network and timing compensation
	MinDynamicDelay = 200 * time.Millisecond // Minimum network delay compensation
	MaxDynamicDelay = 500 * time.Millisecond // Maximum network delay compensation
)

type WatchPartyManager struct {
	logger  *zerolog.Logger
	manager *Manager

	currentSession   mo.Option[*WatchPartySession] // Current watch party session
	sessionCtx       context.Context               // Context for the current watch party session
	sessionCtxCancel context.CancelFunc            // Cancel function for the current watch party session
	mu               sync.RWMutex                  // Mutex for the watch party manager

	// Seek management to prevent choppy playback
	lastSeekTime time.Time     // Time of last seek operation
	seekCooldown time.Duration // Minimum time between seeks

	// Catch-up management
	catchUpCancel context.CancelFunc // Cancel function for catch-up operations
	catchUpMu     sync.Mutex         // Mutex for catch-up operations

	// Seek management
	pendingSeekTime     time.Time  // When a seek was initiated
	pendingSeekPosition float64    // Position we're seeking to
	seekMu              sync.Mutex // Mutex for seek state

	// Buffering management (host only)
	bufferWaitStart     time.Time          // When we started waiting for peers to buffer
	isWaitingForBuffers bool               // Whether we're currently waiting for peers to be ready
	bufferMu            sync.Mutex         // Mutex for buffer state changes
	statusReportTicker  *time.Ticker       // Ticker for peer status reporting
	statusReportCancel  context.CancelFunc // Cancel function for status reporting
	waitForPeersCancel  context.CancelFunc // Cancel function for waitForPeersReady goroutine

	// Buffering detection (peer only)
	bufferDetectionMu sync.Mutex // Mutex for buffering detection state
	lastPosition      float64    // Last known playback position
	lastPositionTime  time.Time  // When we last updated the position
	stallCount        int        // Number of consecutive stalls detected

	lastPlayState     bool      // Last known play/pause state to detect rapid changes
	lastPlayStateTime time.Time // When we last changed play state
}

type WatchPartySession struct {
	ID               string                                   `json:"id"`
	Participants     map[string]*WatchPartySessionParticipant `json:"participants"`
	Settings         *WatchPartySessionSettings               `json:"settings"`
	CreatedAt        time.Time                                `json:"createdAt"`
	CurrentMediaInfo *WatchPartySessionMediaInfo              `json:"currentMediaInfo"` // can be nil if not set
	IsRelayMode      bool                                     `json:"isRelayMode"`      // Whether this session is in relay mode
	mu               sync.RWMutex                             `json:"-"`
}

type WatchPartySessionParticipant struct {
	ID         string    `json:"id"`       // PeerID (UUID) for unique identification
	Username   string    `json:"username"` // Display name
	IsHost     bool      `json:"isHost"`
	CanControl bool      `json:"canControl"`
	IsReady    bool      `json:"isReady"`
	LastSeen   time.Time `json:"lastSeen"`
	Latency    int64     `json:"latency"` // in milliseconds
	// Buffering state
	IsBuffering    bool                        `json:"isBuffering"`
	BufferHealth   float64                     `json:"bufferHealth"`             // 0.0 to 1.0, how much buffer is available
	PlaybackStatus *mediaplayer.PlaybackStatus `json:"playbackStatus,omitempty"` // Current playback status
	// Relay mode
	IsRelayOrigin bool `json:"isRelayHost"` // Whether this peer is the origin for relay mode
}

type WatchPartySessionMediaInfo struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	AniDBEpisode  string `json:"aniDbEpisode"`
	StreamType    string `json:"streamType"` // "file", "torrent", "debrid"
	StreamPath    string `json:"streamPath"` // URL for stream playback (e.g. /api/v1/nakama/stream?type=file&path=...)
}

type WatchPartySessionSettings struct {
	SyncThreshold     float64 `json:"syncThreshold"`     // Seconds of desync before forcing sync
	MaxBufferWaitTime int     `json:"maxBufferWaitTime"` // Max time to wait for buffering peers (seconds)
}

// Events
type (
	WatchPartyCreatedPayload struct {
		Session *WatchPartySession `json:"session"`
	}

	WatchPartyJoinPayload struct {
		PeerId   string `json:"peerId"`
		Username string `json:"username"`
	}

	WatchPartyLeavePayload struct {
		PeerId string `json:"peerId"`
	}

	WatchPartyPlaybackInfoPayload struct {
		MediaInfo *WatchPartySessionMediaInfo `json:"mediaInfo"`
	}

	WatchPartyPlaybackStatusPayload struct {
		PlaybackStatus mediaplayer.PlaybackStatus `json:"playbackStatus"`
		Timestamp      time.Time                  `json:"timestamp"`     // Client timestamp
		EpisodeNumber  int                        `json:"episodeNumber"` // For episode changes
	}

	WatchPartyStateChangedPayload struct {
		Session *WatchPartySession `json:"session"`
	}

	WatchPartyPeerStatusPayload struct {
		PeerId         string                     `json:"peerId"`
		PlaybackStatus mediaplayer.PlaybackStatus `json:"playbackStatus"`
		IsBuffering    bool                       `json:"isBuffering"`
		BufferHealth   float64                    `json:"bufferHealth"` // 0.0 to 1.0
		Timestamp      time.Time                  `json:"timestamp"`
	}

	WatchPartyBufferUpdatePayload struct {
		PeerId       string    `json:"peerId"`
		IsBuffering  bool      `json:"isBuffering"`
		BufferHealth float64   `json:"bufferHealth"`
		Timestamp    time.Time `json:"timestamp"`
	}
)

func NewWatchPartyManager(manager *Manager) *WatchPartyManager {
	return &WatchPartyManager{
		logger:       manager.logger,
		manager:      manager,
		seekCooldown: DefaultSeekCooldown,
	}
}

// Cleanup stops all goroutines and cleans up resources to prevent memory leaks
func (wpm *WatchPartyManager) Cleanup() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Msg("nakama: Cleaning up watch party manager")

	// Stop status reporting (peer side)
	wpm.stopStatusReporting()

	// Cancel any ongoing catch-up operations
	wpm.cancelCatchUp()

	// Clean up seek management state
	wpm.seekMu.Lock()
	wpm.pendingSeekTime = time.Time{}
	wpm.pendingSeekPosition = 0
	wpm.seekMu.Unlock()

	// Cancel waitForPeersReady goroutine (host side)
	wpm.bufferMu.Lock()
	if wpm.waitForPeersCancel != nil {
		wpm.waitForPeersCancel()
		wpm.waitForPeersCancel = nil
	}
	wpm.isWaitingForBuffers = false
	wpm.bufferMu.Unlock()

	// Cancel session context (stops all session-related goroutines)
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
	}

	// Clear session
	wpm.currentSession = mo.None[*WatchPartySession]()

	wpm.logger.Debug().Msg("nakama: Watch party manager cleanup completed")
}

// GetCurrentSession returns the current watch party session if it exists
func (wpm *WatchPartyManager) GetCurrentSession() (*WatchPartySession, bool) {
	wpm.mu.RLock()
	defer wpm.mu.RUnlock()

	session, ok := wpm.currentSession.Get()
	return session, ok
}

func (wpm *WatchPartyManager) handleMessage(message *Message, senderID string) error {
	marshaledPayload, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	// wpm.logger.Debug().Str("type", string(message.Type)).Interface("payload", message.Payload).Msg("nakama: Received watch party message")

	switch message.Type {
	case MessageTypeWatchPartyStateChanged:
		// wpm.logger.Debug().Msg("nakama: Received watch party state changed message")
		var payload WatchPartyStateChangedPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyStateChangedEvent(&payload)
	case MessageTypeWatchPartyCreated:
		wpm.logger.Debug().Msg("nakama: Received watch party created message")
		var payload WatchPartyCreatedPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyCreatedEvent(&payload)

	case MessageTypeWatchPartyStopped:
		wpm.logger.Debug().Msg("nakama: Received watch party stopped message")
		wpm.handleWatchPartyStoppedEvent()

	case MessageTypeWatchPartyJoin:
		wpm.logger.Debug().Msg("nakama: Received watch party join message")
		var payload WatchPartyJoinPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerJoinedEvent(&payload, message.Timestamp)

	case MessageTypeWatchPartyLeave:
		wpm.logger.Debug().Msg("nakama: Received watch party leave message")
		var payload WatchPartyLeavePayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerLeftEvent(&payload)

	case MessageTypeWatchPartyPeerStatus:
		wpm.logger.Debug().Msg("nakama: Received watch party peer status message")
		var payload WatchPartyPeerStatusPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerStatusEvent(&payload)

	case MessageTypeWatchPartyBufferUpdate:
		wpm.logger.Debug().Msg("nakama: Received watch party buffer update message")
		var payload WatchPartyBufferUpdatePayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyBufferUpdateEvent(&payload)

	case MessageTypeWatchPartyPlaybackStatus:
		// wpm.logger.Debug().Msg("nakama: Received watch party playback status message")
		var payload WatchPartyPlaybackStatusPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPlaybackStatusEvent(&payload)

	case MessageTypeWatchPartyRelayModePeersReady:
		// TODO: Implement
	case MessageTypeWatchPartyRelayModeOriginPlaybackStatus:
		// TODO: Implement
	}

	return nil
}

func (mi *WatchPartySessionMediaInfo) Equals(other *WatchPartySessionMediaInfo) bool {
	if mi == nil || other == nil {
		return false
	}

	return mi.MediaId == other.MediaId &&
		mi.EpisodeNumber == other.EpisodeNumber &&
		mi.AniDBEpisode == other.AniDBEpisode &&
		mi.StreamType == other.StreamType &&
		mi.StreamPath == other.StreamPath
}

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
		Username:   wpm.manager.settings.Username,
		IsHost:     true,
		CanControl: true,
		IsReady:    true,
		LastSeen:   time.Now(),
		Latency:    0,
	}

	wpm.currentSession = mo.Some(session)

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

	return session, nil
}

// PromotePeerToRelayModeOrigin promotes a peer to be the origin for relay mode
// TODO: To implement
func (wpm *WatchPartyManager) PromotePeerToRelayModeOrigin(peerId string) {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

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

				session, ok := wpm.currentSession.Get()
				if !ok {
					continue
				}

				wpm.manager.playbackManager.PullStatus()

				switch event := event.(type) {
				case playbackmanager.PlaybackStatusChangedEvent:
					if event.State.MediaId == 0 {
						continue
					}

					streamType := "file"
					if event.Status.PlaybackType == mediaplayer.PlaybackTypeStream {
						if strings.Contains(event.Status.Filepath, "/api/v1/torrentstream") {
							streamType = "torrent"
						} else {
							streamType = "debrid"
						}
					}

					streamPath := event.Status.Filepath
					newCurrentMediaInfo := &WatchPartySessionMediaInfo{
						MediaId:       event.State.MediaId,
						EpisodeNumber: event.State.EpisodeNumber,
						AniDBEpisode:  event.State.AniDbEpisode,
						StreamType:    streamType,
						StreamPath:    streamPath,
					}
					// Video playback has started, send the media info to the peers
					if session.CurrentMediaInfo.Equals(newCurrentMediaInfo) && event.State.MediaId != 0 {
						_ = wpm.manager.SendMessage(MessageTypeWatchPartyPlaybackStatus, WatchPartyPlaybackStatusPayload{
							PlaybackStatus: event.Status,
							Timestamp:      time.Now(),
							EpisodeNumber:  event.State.EpisodeNumber,
						})
					} else {
						wpm.logger.Debug().Msgf("nakama: Playback changed or started: %s", streamPath)
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

						wpm.broadcastSessionStateToPeers()

						// Start checking peer readiness
						go wpm.waitForPeersReady()
					}
				}
			}
		}
	}()
}

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

	// Send session state
	wpm.broadcastSessionStateToPeers()

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
	wpm.broadcastSessionStateToPeers()

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
	wpm.broadcastSessionStateToPeers()

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
	wpm.checkAndManageBuffering()
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
	wpm.checkAndManageBuffering()

	// Broadcast updated session state
	wpm.broadcastSessionStateToPeers()
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
func (wpm *WatchPartyManager) waitForPeersReady() {
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
				_ = wpm.manager.playbackManager.Resume()
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
				if !participant.IsHost {
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
				_ = wpm.manager.playbackManager.Resume()
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

	// Send join message to host
	_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyJoin, WatchPartyJoinPayload{
		PeerId:   hostConn.PeerId,
		Username: wpm.manager.settings.Username,
	})

	// Start status reporting to host
	wpm.startStatusReporting()

	// Send websocket event to update the UI
	wpm.sendSessionStateToClient()

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

// +----------------
// | Events
// +----------------

// handleWatchPartyStateChangedEvent is called when the host updates the session state.
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

	//
	// Starting playback / Peer joined / Video changed
	//

	// If the payload session has a media info but the current session doesn't,
	// and the peer is a participant, we need to start playback
	isParticipant := payload.Session.Participants[hostConn.PeerId] != nil
	newPlayback := payload.Session.CurrentMediaInfo != nil && currentSession.CurrentMediaInfo == nil
	playbackChanged := payload.Session.CurrentMediaInfo != nil && !payload.Session.CurrentMediaInfo.Equals(currentSession.CurrentMediaInfo)

	if (newPlayback || playbackChanged) && isParticipant {
		wpm.logger.Debug().Bool("newPlayback", newPlayback).Bool("playbackChanged", playbackChanged).Msg("nakama: Starting playback due to new media info")

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
		case "torrent", "debrid":
			err = wpm.manager.PlayHostAnimeStream(payload.Session.CurrentMediaInfo.StreamType, "seanime/nakama", media, payload.Session.CurrentMediaInfo.AniDBEpisode)
		case "file":
			err = wpm.manager.PlayHostAnimeLibraryFile(payload.Session.CurrentMediaInfo.StreamPath, "seanime/nakama", media, payload.Session.CurrentMediaInfo.AniDBEpisode)
		}
		if err != nil {
			wpm.logger.Error().Err(err).Msg("nakama: Failed to play watch party media")
			wpm.manager.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Watch party: Failed to play media: %s", err.Error()))
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
		_ = wpm.manager.playbackManager.Cancel()
		canceledPlayback = true
	}

	//
	// Session stopped
	//

	// If the host stopped the session, we need to cancel playback
	if payload.Session.CurrentMediaInfo == nil && payload.Session.CurrentMediaInfo != nil && !canceledPlayback {
		wpm.logger.Debug().Msg("nakama: Canceling playback due to host stopping session")
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
	timeSinceMessage := time.Since(payload.Timestamp).Seconds()
	if timeSinceMessage > 5.0 {
		wpm.logger.Debug().Float64("timeSinceMessage", timeSinceMessage).Msg("nakama: Ignoring stale playback status message")
		return
	}

	// Handle play/pause state changes
	if payloadStatus.Playing != playbackStatus.Playing {
		if payloadStatus.Playing {
			// Cancel any ongoing catch-up operation
			wpm.cancelCatchUp()

			// When host resumes, sync position before resuming if there's significant drift
			timeSinceMessage := time.Since(payload.Timestamp).Seconds()
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
			wpm.handleHostPause(payloadStatus, *playbackStatus, payload.Timestamp)
		}
	}

	// Handle position sync for different state combinations
	if payloadStatus.Playing == playbackStatus.Playing {
		// Both in same state, use normal sync
		wpm.syncPlaybackPosition(payloadStatus, *playbackStatus, payload.Timestamp, session)
	} else if payloadStatus.Playing && !playbackStatus.Playing {
		// Host playing, peer paused, sync position and resume
		timeSinceMessage := time.Since(payload.Timestamp).Seconds()
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
		wpm.handleHostPause(payloadStatus, *playbackStatus, payload.Timestamp)
	}
}

// handleHostPause handles when the host pauses playback
func (wpm *WatchPartyManager) handleHostPause(hostStatus mediaplayer.PlaybackStatus, peerStatus mediaplayer.PlaybackStatus, hostTimestamp time.Time) {
	// Cancel any ongoing catch-up operation
	wpm.cancelCatchUp()

	now := time.Now()
	timeSinceMessage := now.Sub(hostTimestamp).Seconds()

	// Calculate where the host actually paused based on dynamic timing
	hostActualPausePosition := hostStatus.CurrentTimeInSeconds
	// Don't add time compensation for pause position, the host has already paused

	// Calculate time difference considering message delay
	timeDifference := hostActualPausePosition - peerStatus.CurrentTimeInSeconds

	// If peer is significantly behind the host, let it catch up before pausing
	if timeDifference > CatchUpBehindThreshold {
		wpm.logger.Debug().Msgf("nakama: Host paused, peer behind by %.2f seconds, catching up", timeDifference)
		wpm.startCatchUp(hostActualPausePosition, hostTimestamp)
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
func (wpm *WatchPartyManager) startCatchUp(hostPausePosition float64, hostTimestamp time.Time) {
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
func (wpm *WatchPartyManager) syncPlaybackPosition(hostStatus mediaplayer.PlaybackStatus, peerStatus mediaplayer.PlaybackStatus, hostTimestamp time.Time, session *WatchPartySession) {
	now := time.Now()
	timeSinceMessage := now.Sub(hostTimestamp).Seconds()

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
