package nakama

import (
	"context"
	"encoding/json"
	"errors"
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
	MessageTypeWatchPartyCreated         = "watch_party_created"          // Host creates a watch party
	MessageTypeWatchPartyStateChanged    = "watch_party_state_changed"    // Host or peer changes the state of the watch party
	MessageTypeWatchPartyStopped         = "watch_party_stopped"          // Host stops a watch party
	MessageTypeWatchPartyPlaybackInfo    = "watch_party_playback_info"    // Host is ready, sends playback info to peers
	MessageTypeWatchPartyPlaybackStatus  = "watch_party_playback_status"  // Host or peer sends playback status to peers (seek, play, pause, etc)
	MessageTypeWatchPartyPlaybackStopped = "watch_party_playback_stopped" // Peer sends playback stopped to host
	// Peer -> Host
	MessageTypeWatchPartyJoin  = "watch_party_join"  // Peer joins a watch party
	MessageTypeWatchPartyLeave = "watch_party_leave" // Peer leaves a watch party
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
}

type WatchPartySession struct {
	ID               string                                   `json:"id"`
	Participants     map[string]*WatchPartySessionParticipant `json:"participants"`
	Settings         *WatchPartySessionSettings               `json:"settings"`
	CreatedAt        time.Time                                `json:"createdAt"`
	CurrentMediaInfo *WatchPartySessionMediaInfo              `json:"currentMediaInfo"` // can be nil if not set
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
}

type WatchPartySessionMediaInfo struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	AniDBEpisode  string `json:"aniDbEpisode"`
	StreamType    string `json:"streamType"` // "file", "torrent", "debrid"
	StreamPath    string `json:"streamPath"` // URL for stream playback (e.g. /api/v1/nakama/stream?type=file&path=...)
}

type WatchPartySessionSettings struct {
	AllowParticipantControl bool    `json:"allowParticipantControl"`
	SyncThreshold           float64 `json:"syncThreshold"`     // Seconds of desync before forcing sync
	MaxBufferWaitTime       int     `json:"maxBufferWaitTime"` // Max time to wait for buffering peers (seconds)
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
)

func NewWatchPartyManager(manager *Manager) *WatchPartyManager {
	return &WatchPartyManager{
		logger:       manager.logger,
		manager:      manager,
		seekCooldown: 1 * time.Second, // cooldown between seeks to prevent choppy playback
	}
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

	wpm.logger.Debug().Str("type", string(message.Type)).Interface("payload", message.Payload).Msg("nakama: Received watch party message")

	switch message.Type {
	case MessageTypeWatchPartyStateChanged:
		var payload WatchPartyStateChangedPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyStateChangedEvent(&payload, message.Timestamp, senderID)
	case MessageTypeWatchPartyCreated:
		var payload WatchPartyCreatedPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyCreatedEvent(&payload, message.Timestamp, senderID)

	case MessageTypeWatchPartyStopped:
		wpm.handleWatchPartyStoppedEvent(message.Timestamp, senderID)

	case MessageTypeWatchPartyJoin:
		var payload WatchPartyJoinPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerJoinedEvent(&payload, message.Timestamp, senderID)

	case MessageTypeWatchPartyLeave:
		var payload WatchPartyLeavePayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPeerLeftEvent(&payload, message.Timestamp, senderID)

	case MessageTypeWatchPartyPlaybackInfo:
		var payload WatchPartyPlaybackInfoPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPlaybackInfoEvent(&payload, message.Timestamp, senderID)

	case MessageTypeWatchPartyPlaybackStatus:
		var payload WatchPartyPlaybackStatusPayload
		err := json.Unmarshal(marshaledPayload, &payload)
		if err != nil {
			return err
		}
		wpm.handleWatchPartyPlaybackStatusEvent(&payload, message.Timestamp, senderID)
	}

	return nil
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
	wpm.manager.SendMessage(MessageTypeWatchPartyCreated, WatchPartyCreatedPayload{
		Session: session,
	})

	wpm.logger.Debug().Str("sessionId", sessionID).Msg("nakama: Watch party created")

	// Send websocket event to update the UI
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)

	go func() {
		select {
		case <-wpm.sessionCtx.Done():
			wpm.logger.Debug().Msg("nakama: Watch party stopped")
			return
		}
	}()

	go wpm.listenToPlaybackManager()

	return session, nil
}

func (wpm *WatchPartyManager) StopWatchParty() {
	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Msg("nakama: Stopping watch party")

	// Broadcast the stop event to all peers
	wpm.manager.SendMessage(MessageTypeWatchPartyStopped, nil)

	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	// Send websocket event to update the UI (null indicates session stopped)
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)
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

				switch event := event.(type) {
				case playbackmanager.PlaybackStartingEvent:
					session.CurrentMediaInfo = nil
				case playbackmanager.PlaybackStatusChangedEvent:
					// Video playback has started, send the media info to the peers
					if session.CurrentMediaInfo == nil && event.State.MediaId != 0 {
						streamType := "file"
						if event.Status.PlaybackType == mediaplayer.PlaybackTypeStream {
							if strings.Contains(event.Status.Filepath, "/api/v1/torrentstream") {
								streamType = "torrent"
							} else {
								streamType = "debrid"
							}
						}

						wpm.manager.playbackManager.Pause()

						streamPath := event.Status.Filepath

						wpm.logger.Debug().Msgf("nakama: Playback started: %s", streamPath)

						session.CurrentMediaInfo = &WatchPartySessionMediaInfo{
							MediaId:       event.State.MediaId,
							EpisodeNumber: event.State.EpisodeNumber,
							AniDBEpisode:  event.State.AniDbEpisode,
							StreamType:    streamType,
							StreamPath:    streamPath,
						}

						wpm.manager.SendMessage(MessageTypeWatchPartyStateChanged, WatchPartyStateChangedPayload{
							Session: session,
						})
						wpm.manager.SendMessage(MessageTypeWatchPartyPlaybackInfo, WatchPartyPlaybackInfoPayload{
							MediaInfo: session.CurrentMediaInfo,
						})
					} else {
						wpm.manager.SendMessage(MessageTypeWatchPartyPlaybackStatus, WatchPartyPlaybackStatusPayload{
							PlaybackStatus: event.Status,
							Timestamp:      time.Now(),
							EpisodeNumber:  event.State.EpisodeNumber,
						})
					}
				}
			}
		}
	}()
}

// handleWatchPartyPeerJoinedEvent is called when a peer joins a watch party
func (wpm *WatchPartyManager) handleWatchPartyPeerJoinedEvent(payload *WatchPartyJoinPayload, timestamp time.Time, senderID string) {
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
	}

	// Send session state
	wpm.manager.SendMessage(MessageTypeWatchPartyStateChanged, WatchPartyStateChangedPayload{
		Session: session,
	})

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Updated watch party state after peer joined")

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)
}

func (wpm *WatchPartyManager) handleWatchPartyPeerLeftEvent(payload *WatchPartyLeavePayload, timestamp time.Time, senderID string) {
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
	wpm.manager.SendMessage(MessageTypeWatchPartyStateChanged, WatchPartyStateChangedPayload{
		Session: session,
	})

	wpm.logger.Debug().Str("peerId", payload.PeerId).Msg("nakama: Updated watch party state after peer left")

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)
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

	session, ok := wpm.currentSession.Get()
	if !ok {
		return errors.New("no watch party found")
	}

	// Send join message to host
	wpm.manager.SendMessageToHost(MessageTypeWatchPartyJoin, WatchPartyJoinPayload{
		PeerId:   hostConn.PeerId,
		Username: wpm.manager.settings.Username,
	})

	// Send websocket event to update the UI
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, session)

	return nil
}

func (wpm *WatchPartyManager) LeaveWatchParty() error {
	if wpm.manager.IsHost() {
		return errors.New("only peers can leave watch parties")
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	wpm.logger.Debug().Msg("nakama: Leaving watch party")

	hostConn, ok := wpm.manager.GetHostConnection()
	if !ok {
		return errors.New("no host connection found")
	}

	_, ok = wpm.currentSession.Get()
	if !ok {
		return errors.New("no watch party found")
	}

	wpm.manager.SendMessageToHost(MessageTypeWatchPartyLeave, WatchPartyLeavePayload{
		PeerId: hostConn.PeerId,
	})

	// Send websocket event to update the UI (null indicates session left)
	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)

	return nil
}

// Events

func (wpm *WatchPartyManager) handleWatchPartyStateChangedEvent(payload *WatchPartyStateChangedPayload, timestamp time.Time, senderID string) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	_, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	// // Create a copy of the session
	// marshaledSession, err := json.Marshal(session)
	// if err != nil {
	// 	return
	// }
	// var previousSession WatchPartySession
	// err = json.Unmarshal(marshaledSession, &previousSession)
	// if err != nil {
	// 	return
	// }

	// Update the session
	wpm.currentSession = mo.Some(payload.Session)

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, wpm.currentSession.MustGet())
}

// handleWatchPartyCreatedEvent is called when a host creates a watch party
func (wpm *WatchPartyManager) handleWatchPartyCreatedEvent(payload *WatchPartyCreatedPayload, timestamp time.Time, senderID string) {
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

	// Add the session to the manager
	wpm.currentSession = mo.Some(payload.Session)

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, wpm.currentSession.MustGet())
}

func (wpm *WatchPartyManager) handleWatchPartyPlaybackInfoEvent(payload *WatchPartyPlaybackInfoPayload, timestamp time.Time, senderID string) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Received playback info from watch party")

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	session, ok := wpm.currentSession.Get()
	if !ok {
		wpm.logger.Debug().Msg("nakama: No watch party found")
		return
	}

	// Update the session
	session.CurrentMediaInfo = payload.MediaInfo

	// streamPath := payload.MediaInfo.StreamPath
	// streamPath = strings.Replace(streamPath, "{{SERVER_URL}}", wpm.manager.getBaseServerURL(), 1)

	// wpm.logger.Debug().Str("streamPath", streamPath).Int("mediaId", payload.MediaInfo.MediaId).Msg("nakama: Fetching media info")

	// Fetch the media info
	media, err := wpm.manager.platform.GetAnime(context.Background(), payload.MediaInfo.MediaId)
	if err != nil {
		wpm.logger.Error().Err(err).Msg("nakama: Failed to fetch media info")
		return
	}

	// Play the media
	wpm.logger.Debug().Msg("nakama: Playing media")

	switch payload.MediaInfo.StreamType {
	case "torrent", "debrid":
		wpm.manager.PlayHostAnimeStream(payload.MediaInfo.StreamType, "seanime/nakama", media, payload.MediaInfo.AniDBEpisode)
	case "file":
		wpm.manager.PlayHostAnimeLibraryFile(payload.MediaInfo.StreamPath, "seanime/nakama", media, payload.MediaInfo.AniDBEpisode)
	}
}

func (wpm *WatchPartyManager) handleWatchPartyStoppedEvent(timestamp time.Time, senderID string) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Host stopped watch party")

	// Cancel any existing session
	if wpm.sessionCtxCancel != nil {
		wpm.sessionCtxCancel()
		wpm.sessionCtx = nil
		wpm.sessionCtxCancel = nil
		wpm.currentSession = mo.None[*WatchPartySession]()
	}

	wpm.manager.wsEventManager.SendEvent(events.NakamaWatchPartyState, nil)
}

func (wpm *WatchPartyManager) handleWatchPartyPlaybackStatusEvent(payload *WatchPartyPlaybackStatusPayload, timestamp time.Time, senderID string) {
	if wpm.manager.IsHost() {
		return
	}

	wpm.logger.Debug().Msg("nakama: Received playback status from watch party")

	wpm.mu.Lock()
	defer wpm.mu.Unlock()

	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	payloadStatus := payload.PlaybackStatus

	playbackStatus, ok := wpm.manager.playbackManager.PullStatus()
	if !ok {
		return
	}

	// Handle play/pause state changes immediately
	if payloadStatus.Playing != playbackStatus.Playing {
		if payloadStatus.Playing {
			wpm.manager.playbackManager.Resume()
		} else {
			wpm.manager.playbackManager.Pause()
		}
	}

	// Calculate drift for position synchronization
	if payloadStatus.CurrentTimeInSeconds != playbackStatus.CurrentTimeInSeconds {
		// Calculate time elapsed since the message was sent
		now := time.Now()
		timeSinceMessage := now.Sub(payload.Timestamp).Seconds()

		// Adjust host position based on time elapsed (if playing)
		hostExpectedPosition := payloadStatus.CurrentTimeInSeconds
		if payloadStatus.Playing {
			hostExpectedPosition += timeSinceMessage
		}

		// Calculate drift between peer and host
		drift := hostExpectedPosition - playbackStatus.CurrentTimeInSeconds
		driftAbs := drift
		if driftAbs < 0 {
			driftAbs = -driftAbs
		}

		// Get sync threshold from session settings, cap at 2 second
		syncThreshold := session.Settings.SyncThreshold
		if syncThreshold > 2.0 {
			syncThreshold = 2.0
		}

		// Check if we're in seek cooldown period
		timeSinceLastSeek := now.Sub(wpm.lastSeekTime)
		inCooldown := timeSinceLastSeek < wpm.seekCooldown

		//wpm.logger.Debug().
		//	Float64("drift", drift).
		//	Float64("driftAbs", driftAbs).
		//	Float64("syncThreshold", syncThreshold).
		//	Float64("hostPosition", hostExpectedPosition).
		//	Float64("peerPosition", playbackStatus.CurrentTimeInSeconds).
		//	Float64("timeSinceMessage", timeSinceMessage).
		//	Float64("timeSinceLastSeek", timeSinceLastSeek.Seconds()).
		//	Bool("inCooldown", inCooldown).
		//	Msg("nakama: Calculated playback drift")

		// Only sync if drift exceeds threshold and we're not in cooldown
		if driftAbs > syncThreshold && !inCooldown {
			//wpm.logger.Debug().
			//	Float64("drift", drift).
			//	Float64("newPosition", hostExpectedPosition).
			//	Msg("nakama: Syncing playback position due to drift")

			wpm.manager.playbackManager.Seek(hostExpectedPosition)
			wpm.lastSeekTime = now // Update last seek time
		} else if driftAbs > syncThreshold && inCooldown {
			//wpm.logger.Debug().
			//	Float64("drift", drift).
			//	Float64("cooldownRemaining", (wpm.seekCooldown - timeSinceLastSeek).Seconds()).
			//	Msg("nakama: Sync needed but in cooldown period, skipping seek")
		}
	}
}
