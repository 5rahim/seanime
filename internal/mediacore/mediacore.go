package mediacore

import (
	"context"
	"fmt"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/models"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Coordinator struct {
	logger                     *zerolog.Logger
	continuityManager          *continuity.Manager
	metadataProviderRef        *util.Ref[metadata_provider.Provider]
	discordPresence            *discordrpc_presence.Presence
	platformRef                *util.Ref[platform.Platform]
	refreshAnimeCollectionFunc func()
	isOfflineRef               *util.Ref[bool]

	backends map[Target]Backend

	mu                   sync.RWMutex
	session              SessionKey
	activeTarget         Target
	activePlaybackState  *PlaybackState
	activePlaybackStatus *PlaybackStatus
	activePlaybackInfo   *PlaybackInfo

	settingsMu sync.RWMutex
	settings   *models.Settings

	subscribers *result.Map[string, *Subscriber]
	eventBus    chan Event
	stopCh      chan struct{}
	startOnce   sync.Once
	effectsOnce sync.Once
}

type Subscriber struct {
	id        string
	eventCh   chan Event
	closed    atomic.Bool
	closeOnce sync.Once
}

func (s *Subscriber) Events() <-chan Event { return s.eventCh }
func (s *Subscriber) GetID() string        { return s.id }

type NewCoordinatorOptions struct {
	Logger                     *zerolog.Logger
	MetadataProviderRef        *util.Ref[metadata_provider.Provider]
	ContinuityManager          *continuity.Manager
	DiscordPresence            *discordrpc_presence.Presence
	PlatformRef                *util.Ref[platform.Platform]
	RefreshAnimeCollectionFunc func()
	IsOfflineRef               *util.Ref[bool]
	Backends                   map[Target]Backend
}

func NewCoordinator(opts NewCoordinatorOptions) *Coordinator {
	c := &Coordinator{
		logger:                     opts.Logger,
		metadataProviderRef:        opts.MetadataProviderRef,
		continuityManager:          opts.ContinuityManager,
		discordPresence:            opts.DiscordPresence,
		platformRef:                opts.PlatformRef,
		refreshAnimeCollectionFunc: opts.RefreshAnimeCollectionFunc,
		isOfflineRef:               opts.IsOfflineRef,
		backends:                   opts.Backends,
		subscribers:                result.NewMap[string, *Subscriber](),
		eventBus:                   make(chan Event, 100),
		stopCh:                     make(chan struct{}),
	}

	c.Start()
	return c
}

func (c *Coordinator) Start() {
	c.startOnce.Do(func() {
		// Listen to each backend's events in a separate goroutine
		for target, b := range c.backends {
			go c.listenToBackendEvents(target, b)
		}

		go func() {
			for {
				select {
				case <-c.stopCh:
					return
				case event := <-c.eventBus:
					c.dispatch(event)
				}
			}
		}()
	})
}

func (c *Coordinator) Close() error {
	select {
	case <-c.stopCh:
		return nil
	default:
		close(c.stopCh)
	}

	// Close all backends
	for _, b := range c.backends {
		_ = b.Close()
	}

	return nil
}

func (c *Coordinator) SetSettings(settings *models.Settings) {
	c.settingsMu.Lock()
	c.settings = settings
	c.settingsMu.Unlock()
}

func (c *Coordinator) Subscribe(id string) *Subscriber {
	sub := &Subscriber{id: id, eventCh: make(chan Event, 100)}
	if previous, ok := c.subscribers.Pop(id); ok {
		previous.closed.Store(true)
		previous.closeOnce.Do(func() { close(previous.eventCh) })
	}
	c.subscribers.Set(id, sub)
	return sub
}

func (c *Coordinator) Unsubscribe(id string) {
	if sub, ok := c.subscribers.Pop(id); ok {
		sub.closed.Store(true)
		sub.closeOnce.Do(func() { close(sub.eventCh) })
	}
}

func (c *Coordinator) RegisterEventCallback(callback func(Event) bool) func() {
	id := uuid.NewString()
	sub := c.Subscribe(id)
	var once sync.Once
	cancel := func() { once.Do(func() { c.Unsubscribe(id) }) }
	go func() {
		defer cancel()
		for event := range sub.Events() {
			if !callback(event) {
				return
			}
		}
	}()
	return cancel
}

func (c *Coordinator) GetActiveSession() (SessionKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session, c.session.PlaybackID != ""
}

func (c *Coordinator) GetActivePlaybackState() (PlaybackState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.activePlaybackState == nil {
		return PlaybackState{}, false
	}
	return *c.activePlaybackState, true
}

func (c *Coordinator) GetActivePlaybackStatus() (PlaybackStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.activePlaybackStatus == nil {
		return PlaybackStatus{}, false
	}
	return *c.activePlaybackStatus, true
}

func (c *Coordinator) Execute(session SessionKey, cmd Command) error {
	backend, ok := c.backends[session.Target]
	if !ok {
		return fmt.Errorf("unknown target backend: %s", session.Target)
	}

	c.mu.RLock()
	activeSession := c.session
	c.mu.RUnlock()

	if activeSession.PlaybackID != "" && activeSession.PlaybackID != session.PlaybackID {
		return fmt.Errorf("session mismatch or stale command: expected playback ID %s, got %s", activeSession.PlaybackID, session.PlaybackID)
	}

	return backend.Execute(session, cmd)
}

func (c *Coordinator) Terminate(session SessionKey) {
	backend, ok := c.backends[session.Target]
	if !ok {
		return
	}
	backend.Terminate(session)
}

func (c *Coordinator) GetSession() (SessionKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session, c.session.ClientID != ""
}

func (c *Coordinator) GetActivePlaybackInfo() (*PlaybackInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.activePlaybackInfo != nil {
		return c.activePlaybackInfo, true
	}
	if c.activePlaybackState != nil {
		return c.activePlaybackState.PlaybackInfo, true
	}
	return nil, false
}

func (c *Coordinator) OpenAndAwait(target Target, clientID, state string) {
	c.mu.Lock()
	c.activeTarget = target
	c.session = SessionKey{
		Target:     target,
		ClientID:   clientID,
		PlaybackID: "",
	}
	c.activePlaybackState = nil
	c.activePlaybackStatus = nil
	c.activePlaybackInfo = nil
	c.mu.Unlock()

	backend, ok := c.backends[target]
	if ok {
		backend.OpenAndAwait(clientID, state)
	}
}

func (c *Coordinator) AbortOpen(target Target, clientID, reason string) {
	backend, ok := c.backends[target]
	if ok {
		backend.AbortOpen(clientID, reason)
	}
}

func (c *Coordinator) Watch(target Target, clientID string, info *PlaybackInfo) {
	if info == nil {
		return
	}
	c.restoreContinuity(info)
	c.mu.Lock()
	c.activeTarget = target
	c.session = SessionKey{
		Target:     target,
		ClientID:   clientID,
		PlaybackID: info.ID,
	}
	c.activePlaybackInfo = info
	c.mu.Unlock()

	backend, ok := c.backends[target]
	if ok {
		backend.Watch(clientID, info)
	}
}

func (c *Coordinator) Error(target Target, clientID string, err error) {
	backend, ok := c.backends[target]
	if ok {
		backend.Error(clientID, err)
	}
}

func (c *Coordinator) PullStatus() (PlaybackStatus, bool) {
	c.mu.RLock()
	target := c.activeTarget
	c.mu.RUnlock()

	backend, ok := c.backends[target]
	if !ok {
		return PlaybackStatus{}, false
	}
	return backend.PullStatus()
}

func (c *Coordinator) GetPlaylist() (*PlaylistState, bool) {
	c.mu.RLock()
	target := c.activeTarget
	c.mu.RUnlock()

	backend, ok := c.backends[target]
	if !ok {
		return nil, false
	}
	return backend.GetPlaylist()
}

func (c *Coordinator) GetSkipData() (*SkipData, bool) {
	c.mu.RLock()
	target := c.activeTarget
	c.mu.RUnlock()

	backend, ok := c.backends[target]
	if !ok {
		return nil, false
	}
	return backend.GetSkipData()
}

func (c *Coordinator) SetSkipData(data *SkipData) {
	c.mu.RLock()
	target := c.activeTarget
	session := c.session
	c.mu.RUnlock()

	if target != "" && session.ClientID != "" {
		_ = c.Execute(session, Command{Type: CommandSetSkipData, Payload: data})
	}
}

func (c *Coordinator) ClearSkipData() {
	c.mu.RLock()
	target := c.activeTarget
	session := c.session
	c.mu.RUnlock()

	if target != "" && session.ClientID != "" {
		_ = c.Execute(session, Command{Type: CommandClearSkipData, Payload: nil})
	}
}

func (c *Coordinator) dispatch(event Event) {
	c.subscribers.Range(func(id string, sub *Subscriber) bool {
		if sub.closed.Load() {
			return true
		}
		if event.IsCritical() {
			select {
			case sub.eventCh <- event:
			case <-time.After(time.Second):
				c.logger.Warn().Str("subscriber", id).Msg("mediacore: subscriber blocked a critical event")
			}
		} else {
			select {
			case sub.eventCh <- event:
			default:
			}
		}
		return true
	})
}

func (c *Coordinator) listenToBackendEvents(target Target, b Backend) {
	for ev := range b.Events() {
		key := ev.GetSessionKey()
		var finalState *PlaybackState
		var finalStatus *PlaybackStatus

		c.mu.Lock()
		if c.session.Target != key.Target || c.session.ClientID != key.ClientID {
			_, isLoaded := ev.(*PlaybackLoadedEvent)
			if c.session.Target == "" && isLoaded {
				c.activeTarget = key.Target
				c.session = SessionKey{
					Target:     key.Target,
					ClientID:   key.ClientID,
					PlaybackID: key.PlaybackID,
				}
			} else {
				c.mu.Unlock()
				continue
			}
		}

		if c.session.PlaybackID != "" && key.PlaybackID != "" && c.session.PlaybackID != key.PlaybackID {
			c.mu.Unlock()
			continue
		}

		if c.session.PlaybackID == "" && key.PlaybackID != "" {
			c.session.PlaybackID = key.PlaybackID
		}

		// Update cached state/status
		switch event := ev.(type) {
		case *PlaybackLoadedEvent:
			c.activePlaybackState = &event.State
			c.activePlaybackInfo = event.State.PlaybackInfo
		case *LoadedMetadataEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, event.Paused)
		case *CanPlayEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, event.Paused)
		case *PausedEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, true)
		case *ResumedEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, false)
		case *StatusEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, event.Paused)
		case *SeekedEvent:
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, event.Paused)
		case *CompletedEvent:
			paused := false
			if c.activePlaybackStatus != nil {
				paused = c.activePlaybackStatus.Paused
			}
			c.activePlaybackStatus = playbackStatusFromEvent(event.BaseEvent, event.CurrentTime, event.Duration, paused)
		case *TerminatedEvent:
			if c.activePlaybackState != nil {
				finalState = new(*c.activePlaybackState)
			}
			if c.activePlaybackStatus != nil {
				finalStatus = new(*c.activePlaybackStatus)
			}
			c.activePlaybackState = nil
			c.activePlaybackStatus = nil
			c.activePlaybackInfo = nil
			c.session = SessionKey{}
		}
		c.mu.Unlock()

		if finalState != nil && finalStatus != nil {
			c.updateContinuityState(*finalState, finalStatus.CurrentTime, finalStatus.Duration)
		}

		select {
		case c.eventBus <- ev:
		default:
			c.logger.Warn().Msg("mediacore: coordinator event bus full")
		}
	}
}

func (c *Coordinator) SetupSharedEffects() {
	c.effectsOnce.Do(func() {
		sub := c.Subscribe("coordinator:effects")
		go func() {
			for event := range sub.Events() {
				switch value := event.(type) {
				case *PausedEvent:
					c.updateContinuity(value.CurrentTime, value.Duration)
					if c.discordPresence != nil && !c.isOfflineRef.Get() {
						go c.discordPresence.UpdateAnimeActivity(int(value.CurrentTime), int(value.Duration), true)
					}
				case *ResumedEvent:
					if c.discordPresence != nil && !c.isOfflineRef.Get() {
						go c.discordPresence.UpdateAnimeActivity(int(value.CurrentTime), int(value.Duration), false)
					}
				case *LoadedMetadataEvent:
					state, ok := c.GetActivePlaybackState()
					if !ok || state.PlaybackInfo.Media == nil || state.PlaybackInfo.Episode == nil {
						continue
					}
					if c.discordPresence != nil && !c.isOfflineRef.Get() {
						c.logger.Debug().Msgf("mediacore: Setting Discord presence for %s", state.PlaybackInfo.Media.GetPreferredTitle())
						go c.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
							ID:            state.PlaybackInfo.Media.GetID(),
							Title:         state.PlaybackInfo.Media.GetPreferredTitle(),
							Image:         state.PlaybackInfo.Media.GetCoverImageSafe(),
							IsMovie:       state.PlaybackInfo.Media.IsMovie(),
							EpisodeNumber: state.PlaybackInfo.Episode.EpisodeNumber,
							Progress:      int(value.CurrentTime),
							Duration:      int(value.Duration),
						})
					}
				case *StatusEvent:
					state, ok := c.GetActivePlaybackState()
					if !ok || state.PlaybackInfo.Media == nil || state.PlaybackInfo.Episode == nil {
						continue
					}
					c.updateContinuityState(state, value.CurrentTime, value.Duration)
					if c.discordPresence != nil && !c.isOfflineRef.Get() {
						go c.discordPresence.UpdateAnimeActivity(int(value.CurrentTime), int(value.Duration), value.Paused)
					}
				case *SeekedEvent:
					c.updateContinuity(value.CurrentTime, value.Duration)
				case *CompletedEvent:
					state, ok := c.GetActivePlaybackState()
					if !ok || state.PlaybackInfo.Media == nil || state.PlaybackInfo.Episode == nil || c.platformRef == nil {
						continue
					}
					c.settingsMu.RLock()
					shouldUpdate := c.settings != nil && c.settings.GetLibrary().AutoUpdateProgress
					c.settingsMu.RUnlock()
					if !shouldUpdate {
						continue
					}

					mediaID := state.PlaybackInfo.Media.GetID()
					progress := state.PlaybackInfo.Episode.GetProgressNumber()
					total := state.PlaybackInfo.Media.Episodes

					collection, err := c.platformRef.Get().GetAnimeCollection(context.Background(), false)
					if err == nil {
						if listEntry, hasEntry := collection.GetListEntryFromAnimeId(mediaID); hasEntry {
							if listEntry.Progress != nil && progress <= *listEntry.Progress {
								continue
							}
						}
					}

					err = c.platformRef.Get().UpdateEntryProgress(context.Background(), mediaID, progress, total)
					if err == nil && c.refreshAnimeCollectionFunc != nil {
						c.refreshAnimeCollectionFunc()
					} else if err != nil {
						c.logger.Error().Err(err).Msgf("mediacore: Failed to update progress for media %d", mediaID)
					}
				case *EndedEvent, *ErrorEvent, *TerminatedEvent:
					if c.discordPresence != nil && !c.isOfflineRef.Get() {
						go c.discordPresence.Close()
					}
				}
			}
		}()
	})
}

func playbackStatusFromEvent(base BaseEvent, currentTime, duration float64, paused bool) *PlaybackStatus {
	return &PlaybackStatus{
		ID:          base.Session.PlaybackID,
		ClientID:    base.Session.ClientID,
		Paused:      paused,
		CurrentTime: currentTime,
		Duration:    duration,
	}
}

func (c *Coordinator) restoreContinuity(info *PlaybackInfo) {
	if c.continuityManager == nil || info.InitialState != nil || info.Media == nil || info.Episode == nil || info.IsNakamaWatchParty {
		return
	}
	if info.DisableRestoreFromContinuity != nil && *info.DisableRestoreFromContinuity {
		return
	}

	settings := c.continuityManager.GetSettings()
	if settings == nil || !settings.WatchContinuityEnabled {
		return
	}

	history := c.continuityManager.GetExternalPlayerEpisodeWatchHistoryItem(
		info.StreamPath,
		true,
		info.Episode.GetEpisodeNumber(),
		info.Media.GetID(),
	)
	if history == nil || !history.Found || history.Item == nil || history.Item.CurrentTime <= 0 {
		return
	}

	info.InitialState = &InitialState{CurrentTime: new(history.Item.CurrentTime)}
}

func (c *Coordinator) updateContinuity(currentTime, duration float64) {
	state, ok := c.GetActivePlaybackState()
	if !ok {
		return
	}
	c.updateContinuityState(state, currentTime, duration)
}

func (c *Coordinator) updateContinuityState(state PlaybackState, currentTime, duration float64) {
	if c.continuityManager == nil || state.PlaybackInfo == nil || state.PlaybackInfo.Media == nil || state.PlaybackInfo.Episode == nil || duration <= 0 {
		return
	}
	settings := c.continuityManager.GetSettings()
	if settings == nil || !settings.WatchContinuityEnabled {
		return
	}

	kind := continuity.MediastreamKind
	if state.PlaybackInfo.PlaybackType == PlaybackTypeOnlinestream {
		kind = continuity.OnlinestreamKind
	}
	_ = c.continuityManager.UpdateWatchHistoryItem(&continuity.UpdateWatchHistoryItemOptions{
		CurrentTime:   currentTime,
		Duration:      duration,
		MediaId:       state.PlaybackInfo.Media.GetID(),
		EpisodeNumber: state.PlaybackInfo.Episode.GetEpisodeNumber(),
		Kind:          kind,
	})
}
