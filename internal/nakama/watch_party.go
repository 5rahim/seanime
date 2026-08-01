package nakama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/customsource"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/events"
	"seanime/internal/player"
	"seanime/internal/torrentstream"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

const (
	// Host -> Peer
	MessageTypeWatchPartyCreated         = "watch_party_created"          // Host creates a watch party
	MessageTypeWatchPartyStateChanged    = "watch_party_state_changed"    // Host or peer changes the state of the watch party
	MessageTypeWatchPartyStopped         = "watch_party_stopped"          // Host stops a watch party
	MessageTypeWatchPartyPlaybackStatus  = "watch_party_playback_status"  // Host or peer sends playback status to peers (seek, play, pause, etc)
	MessageTypeWatchPartyPlaybackStopped = "watch_party_playback_stopped" // Peer sends playback stopped to host
	// MessageTypeWatchPartyRelayModeStreamReady   = "watch_party_relay_mode_stream_ready"   // Relay server signals to origin that the stream is ready
	MessageTypeWatchPartyRelayModePeersReady    = "watch_party_relay_mode_peers_ready"    // Relay server signals to origin that all peers are ready
	MessageTypeWatchPartyRelayModePeerBuffering = "watch_party_relay_mode_peer_buffering" // Relay server signals to origin the buffering status (tells origin to pause/unpause)
	// Peer -> Host
	MessageTypeWatchPartyJoin         = "watch_party_join"          // Peer joins a watch party
	MessageTypeWatchPartyLeave        = "watch_party_leave"         // Peer leaves a watch party
	MessageTypeWatchPartyPeerStatus   = "watch_party_peer_status"   // Peer reports their current status to host
	MessageTypeWatchPartyBufferUpdate = "watch_party_buffer_update" // Peer reports buffering state to host
	// Relay mode, Origin (Peer) -> Relay (Host) -> Peers
	MessageTypeWatchPartyRelayModeOriginStreamStarted   = "watch_party_relay_mode_origin_stream_started"   // Relay origin sends is starting a stream, the host will start it too
	MessageTypeWatchPartyRelayModeOriginPlaybackStatus  = "watch_party_relay_mode_origin_playback_status"  // Relay origin sends playback status to relay server
	MessageTypeWatchPartyRelayModeOriginPlaybackStopped = "watch_party_relay_mode_origin_playback_stopped" // Relay origin sends playback stopped to relay server
	// Seanime Watch Party Rooms, Host -> Seanime Watch Party Room API -> Peers

	// Chat
	MessageTypeWatchPartyChatMessage = "watch_party_chat_message" // Chat message sent by any participant
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
	clientId string // used for native player

	logger  *zerolog.Logger
	manager *Manager

	playbackController watchPartyPlaybackController

	currentSession   mo.Option[*WatchPartySession] // Current watch party session
	sessionCtx       context.Context               // Context for the current watch party session
	sessionCtxCancel context.CancelFunc            // Cancel function for the current watch party session
	mu               sync.RWMutex                  // Mutex for the watch party manager

	// SeekToSlow management to prevent choppy player
	lastSeekTime time.Time     // Time of last seek operation
	seekCooldown time.Duration // Minimum time between seeks

	// Catch-up management
	catchUpCancel context.CancelFunc // Cancel function for catch-up operations
	catchUpMu     sync.Mutex         // Mutex for catch-up operations

	// SeekToSlow management
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
	lastPosition      float64    // Last known player position
	lastPositionTime  time.Time  // When we last updated the position
	stallCount        int        // Number of consecutive stalls detected

	lastPlayState     bool      // Last known play/pause state to detect rapid changes
	lastPlayStateTime time.Time // When we last changed play state

	// Sequence-based message ordering
	sequenceMu     sync.Mutex // Mutex for sequence number operations
	sendSequence   uint64     // Current sequence number for outgoing messages
	lastRxSequence uint64     // Latest received sequence number

	// Peer
	peerPlaybackListener *WatchPartyPlaybackSubscriber // Listener for player status changes (can be nil)
}

type watchPartyPlaybackController interface {
	PullStatus() (*WatchPartyPlaybackStatus, bool)
	Pause()
	Resume()
	SeekTo(float64)
}

type WatchPartySession struct {
	ID               string                                   `json:"id"`
	Participants     map[string]*WatchPartySessionParticipant `json:"participants"`
	Settings         *WatchPartySessionSettings               `json:"settings"`
	CreatedAt        time.Time                                `json:"createdAt"`
	CurrentMediaInfo *WatchPartySessionMediaInfo              `json:"currentMediaInfo"` // can be nil if not set
	// Whether this session is in relay mode
	// In this case, the host will act as a relay server and relay status from the origin (a chosen peer) to all other peers
	IsRelayMode bool `json:"isRelayMode"`
	// Whether this session is using the Seanime Watch Party Rooms API
	// In this case, the host will broadcast playback status via a  relay server
	IsRoom bool         `json:"isRoom"`
	mu     sync.RWMutex `json:"-"`
}

type WatchPartySessionParticipant struct {
	ID         string    `json:"id"`       // PeerID (UUID) for unique identification
	Username   string    `json:"username"` // Display name
	IsHost     bool      `json:"isHost"`
	CanControl bool      `json:"canControl"`
	IsReady    bool      `json:"isReady"`
	LastSeen   time.Time `json:"lastSeen"`
	Latency    int64     `json:"latency"` // in milliseconds
	// Player settings
	UseDenshiPlayer bool `json:"useDenshiPlayer"` // Whether this participant uses Denshi player
	// Buffering state
	IsBuffering    bool                      `json:"isBuffering"`
	BufferHealth   float64                   `json:"bufferHealth"`             // 0.0 to 1.0, how much buffer is available
	PlaybackStatus *WatchPartyPlaybackStatus `json:"playbackStatus,omitempty"` // Current playback status
	// Relay mode
	IsRelayOrigin bool `json:"isRelayOrigin"` // Whether this peer is the origin for relay mode
}

type WatchPartyStreamType string

const (
	WatchPartyStreamTypeFile         WatchPartyStreamType = "file"
	WatchPartyStreamTypeTorrent      WatchPartyStreamType = "torrent"
	WatchPartyStreamTypeDebrid       WatchPartyStreamType = "debrid"
	WatchPartyStreamTypeOnlinestream WatchPartyStreamType = "onlinestream"
)

type WatchPartySessionMediaInfo struct {
	MediaId       int                  `json:"mediaId"`
	EpisodeNumber int                  `json:"episodeNumber"`
	AniDBEpisode  string               `json:"aniDbEpisode"`
	StreamType    WatchPartyStreamType `json:"streamType"`
	LocalFilePath string               `json:"localFilePath"` // Path to local file if StreamType is file
	Media         *anilist.BaseAnime   `json:"media,omitempty"`
	// OnlinestreamParams used by peers to start the same stream
	OnlinestreamParams *player.OnlinestreamParams `json:"onlinestreamParams,omitempty"`
	// OnlinestreamParams used by peers to start the same stream
	TorrentStreamParams *torrentstream.StartStreamOptions `json:"torrentStreamParams,omitempty"`
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

	WatchPartyPlaybackStatus struct {
		Paused      bool    `json:"paused"`
		CurrentTime float64 `json:"currentTime"` // in seconds
		Duration    float64 `json:"duration"`    // in seconds
	}
	WatchPartyPlaybackState struct {
		MediaId       int                  `json:"mediaId"`
		EpisodeNumber int                  `json:"episodeNumber"`
		AniDBEpisode  string               `json:"aniDbEpisode"`
		StreamType    WatchPartyStreamType `json:"streamType"`
	}

	WatchPartyPlaybackStatusPayload struct {
		PlaybackStatus *WatchPartyPlaybackStatus `json:"playbackStatus"`
		Timestamp      int64                     `json:"timestamp"` // Unix nano timestamp
		SequenceNumber uint64                    `json:"sequenceNumber"`
		EpisodeNumber  int                       `json:"episodeNumber"` // For episode changes
	}

	WatchPartyStateChangedPayload struct {
		Session *WatchPartySession `json:"session"`
	}

	WatchPartyPeerStatusPayload struct {
		PeerId          string                    `json:"peerId"`
		PlaybackStatus  *WatchPartyPlaybackStatus `json:"playbackStatus"`
		IsBuffering     bool                      `json:"isBuffering"`
		BufferHealth    float64                   `json:"bufferHealth"` // 0.0 to 1.0
		UseDenshiPlayer bool                      `json:"useDenshiPlayer"`
		Timestamp       time.Time                 `json:"timestamp"`
	}

	WatchPartyBufferUpdatePayload struct {
		PeerId       string    `json:"peerId"`
		IsBuffering  bool      `json:"isBuffering"`
		BufferHealth float64   `json:"bufferHealth"`
		Timestamp    time.Time `json:"timestamp"`
	}

	WatchPartyEnableRelayModePayload struct {
		PeerId string `json:"peerId"` // PeerID of the peer to promote to origin
	}

	WatchPartyRelayModeOriginStreamStartedPayload struct {
		Filename            string                            `json:"filename"`
		Filepath            string                            `json:"filepath"`
		StreamType          WatchPartyStreamType              `json:"streamType"`
		LocalFilePath       string                            `json:"localFilePath,omitempty"`
		Media               *anilist.BaseAnime                `json:"media,omitempty"`
		TorrentStreamParams *torrentstream.StartStreamOptions `json:"torrentStreamParams,omitempty"`
		DebridStreamParams  *debrid_client.StartStreamOptions `json:"debridStreamParams,omitempty"`
		OnlinestreamParams  *player.OnlinestreamParams        `json:"onlinestreamParams,omitempty"`
		Status              *WatchPartyPlaybackStatus         `json:"status"`
		State               *WatchPartyPlaybackState          `json:"state"`
	}

	WatchPartyRelayModeOriginPlaybackStatusPayload struct {
		Status    *WatchPartyPlaybackStatus `json:"status"`
		State     *WatchPartyPlaybackState  `json:"state"`
		Timestamp int64                     `json:"timestamp"`
	}

	WatchPartyChatMessagePayload struct {
		PeerId    string    `json:"peerId"`
		Username  string    `json:"username"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		MessageId string    `json:"messageId"` // Unique id
	}
)

func NewWatchPartyManager(manager *Manager) *WatchPartyManager {
	return &WatchPartyManager{
		logger:       manager.logger,
		manager:      manager,
		seekCooldown: DefaultSeekCooldown,
	}
}

func (wpm *WatchPartyManager) player() watchPartyPlaybackController {
	if wpm.playbackController != nil {
		return wpm.playbackController
	}
	if wpm.manager == nil {
		return nil
	}
	return wpm.manager.genericPlayer
}

// Cleanup stops all goroutines and cleans up resources to prevent memory leaks
func (wpm *WatchPartyManager) Cleanup() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	if wpm.currentSession.IsPresent() {
		go wpm.LeaveWatchParty()
		go wpm.StopWatchParty()
	}

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
		//wpm.logger.Debug().Msg("nakama: Received watch party peer status message")
		var payload WatchPartyPeerStatusPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerStatusEvent(&payload)

	case MessageTypeWatchPartyBufferUpdate:
		//wpm.logger.Debug().Msg("nakama: Received watch party buffer update message")
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

	case MessageTypeWatchPartyRelayModeOriginStreamStarted:
		wpm.logger.Debug().Msg("nakama: Received relay mode stream from origin message")
		var payload WatchPartyRelayModeOriginStreamStartedPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyRelayModeOriginStreamStartedEvent(&payload)

	case MessageTypeWatchPartyRelayModePeerBuffering:
		// TODO: Implement

	case MessageTypeWatchPartyRelayModePeersReady:
		wpm.logger.Debug().Msg("nakama: Received relay mode peers ready message")
		wpm.handleWatchPartyRelayModePeersReadyEvent()

	case MessageTypeWatchPartyRelayModeOriginPlaybackStatus:
		// wpm.logger.Debug().Msg("nakama: Received relay mode origin playback status message")
		var payload WatchPartyRelayModeOriginPlaybackStatusPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyRelayModeOriginPlaybackStatusEvent(&payload)

	case MessageTypeWatchPartyRelayModeOriginPlaybackStopped:
		wpm.logger.Debug().Msg("nakama: Received relay mode origin playback stopped message")
		wpm.handleWatchPartyRelayModeOriginPlaybackStoppedEvent()

	case MessageTypeWatchPartyChatMessage:
		var payload WatchPartyChatMessagePayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyChatMessageEvent(&payload)
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
		mi.LocalFilePath == other.LocalFilePath
}

func (wpm *WatchPartyManager) getSessionMedia(ctx context.Context, info *WatchPartySessionMediaInfo) (*anilist.BaseAnime, error) {
	if info == nil {
		return nil, errors.New("missing media info")
	}
	if info.Media != nil {
		return info.Media, nil
	}
	if wpm.manager == nil || wpm.manager.platformRef.IsAbsent() {
		return nil, errors.New("platform is not available")
	}
	return wpm.manager.platformRef.Get().GetAnime(ctx, info.MediaId)
}

func (m *Manager) currentPlaybackMedia() (*anilist.BaseAnime, bool) {
	if m == nil {
		return nil, false
	}
	if m.genericPlayer != nil && m.genericPlayer.isPlaybackManager() && m.playbackManager != nil {
		if media, ok := m.playbackManager.GetCurrentMedia(); ok {
			return media, true
		}
	}
	if m.mediacoreCoordinator != nil {
		if state, ok := m.mediacoreCoordinator.GetActivePlaybackState(); ok && state.PlaybackInfo != nil && state.PlaybackInfo.Media != nil {
			return state.PlaybackInfo.Media, true
		}
	}
	if m.playbackManager != nil {
		return m.playbackManager.GetCurrentMedia()
	}
	return nil, false
}

// SendChatMessage sends a chat message to all participants in the watch party
func (wpm *WatchPartyManager) SendChatMessage(message string) error {
	wpm.mu.RLock()
	session, ok := wpm.currentSession.Get()
	wpm.mu.RUnlock()

	if !ok {
		return errors.New("no active watch party session")
	}

	// Get current participant info
	var peerId, username string
	if wpm.manager.IsHost() {
		peerId = "host"
		session.mu.RLock()
		if host, exists := session.Participants["host"]; exists {
			username = host.Username
		}
		session.mu.RUnlock()
	} else {
		hostConn, ok := wpm.manager.GetHostConnection()
		if !ok {
			return errors.New("no host connection")
		}
		peerId = hostConn.PeerId
		username = wpm.manager.username
	}

	// Create chat message payload
	payload := WatchPartyChatMessagePayload{
		PeerId:    peerId,
		Username:  username,
		Message:   message,
		Timestamp: time.Now(),
		MessageId: fmt.Sprintf("%s-%d", peerId, time.Now().UnixNano()),
	}

	// Send the message
	if wpm.manager.IsHost() {
		// Host broadcasts to all peers
		_ = wpm.manager.SendMessage(MessageTypeWatchPartyChatMessage, payload)
		// Send local event since SendMessage doesn't send to self
		wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyChatMessage, &payload)
	} else {
		// Peer sends to host, host will broadcast it back to all peers including sender
		_ = wpm.manager.SendMessageToHost(MessageTypeWatchPartyChatMessage, payload)
	}

	return nil
}

// format: ext_custom_source_{extensionId}|END|{siteUrl}
func getExtensionIdFromSiteUrl(siteUrl *string) (string, bool) {
	if siteUrl == nil {
		return "", false
	}
	s := *siteUrl
	if !strings.HasPrefix(s, "ext_custom_source_") {
		return "", false
	}
	s = strings.Replace(s, "ext_custom_source_", "", 1)
	parts := strings.Split(s, "|END|")
	return parts[0], true
}

func (wpm *WatchPartyManager) getLocalMediaIdOfCustomSource(mediaId int, media *anilist.BaseAnime) int {
	if media == nil || media.SiteURL == nil {
		return mediaId
	}
	if !customsource.IsExtensionId(mediaId) {
		return mediaId
	}

	var customSourceManager *customsource.Manager
	if wpm.manager != nil && !wpm.manager.platformRef.IsAbsent() {
		if p, ok := wpm.manager.platformRef.Get().(interface {
			GetCustomSourceManager() *customsource.Manager
		}); ok {
			customSourceManager = p.GetCustomSourceManager()
		}
	}
	if customSourceManager == nil {
		return mediaId
	}

	_, localId := customsource.ExtractExtensionData(mediaId)

	extId, ok := getExtensionIdFromSiteUrl(media.SiteURL)
	if !ok {
		return mediaId
	}

	provider, ok := customSourceManager.GetProviderFromExtensionId(extId)
	if !ok {
		return mediaId
	}

	return customsource.GenerateMediaId(provider.GetExtensionIdentifier(), localId)
}

func (wpm *WatchPartyManager) translateMediaInfo(info *WatchPartySessionMediaInfo) {
	if info == nil || info.Media == nil {
		return
	}
	if !customsource.IsExtensionId(info.MediaId) {
		return
	}

	localMediaId := wpm.getLocalMediaIdOfCustomSource(info.MediaId, info.Media)
	if localMediaId == info.MediaId {
		return
	}

	wpm.logger.Debug().Int("oldMediaId", info.MediaId).Int("newMediaId", localMediaId).Msg("nakama: Translated custom source MediaId")

	info.MediaId = localMediaId
	info.Media.ID = localMediaId
	if info.TorrentStreamParams != nil {
		info.TorrentStreamParams.MediaId = localMediaId
	}
}

func (wpm *WatchPartyManager) translateOriginStreamStartedPayload(payload *WatchPartyRelayModeOriginStreamStartedPayload) {
	if payload == nil || payload.Media == nil {
		return
	}

	mediaId := 0
	if payload.State != nil {
		mediaId = payload.State.MediaId
	} else if payload.DebridStreamParams != nil {
		mediaId = payload.DebridStreamParams.MediaId
	} else if payload.TorrentStreamParams != nil {
		mediaId = payload.TorrentStreamParams.MediaId
	} else {
		mediaId = payload.Media.ID
	}

	if !customsource.IsExtensionId(mediaId) {
		return
	}

	localMediaId := wpm.getLocalMediaIdOfCustomSource(mediaId, payload.Media)
	if localMediaId == mediaId {
		return
	}

	wpm.logger.Debug().Int("oldMediaId", mediaId).Int("newMediaId", localMediaId).Msg("nakama: Translated origin stream payload custom source MediaId")

	payload.Media.ID = localMediaId
	if payload.State != nil {
		payload.State.MediaId = localMediaId
	}
	if payload.DebridStreamParams != nil {
		payload.DebridStreamParams.MediaId = localMediaId
	}
	if payload.TorrentStreamParams != nil {
		payload.TorrentStreamParams.MediaId = localMediaId
	}
}
