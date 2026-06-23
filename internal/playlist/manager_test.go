package playlist

import (
	"encoding/json"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediacore"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mpvcore"
	"seanime/internal/platforms/platform"
	"seanime/internal/player"
	"seanime/internal/testmocks"
	"seanime/internal/testutil"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func TestPlaylistManagerSendCurrentPlaylistToClient(t *testing.T) {
	t.Run("no playlist sends nil playlist", func(t *testing.T) {
		// this keeps the UI bootstrap path honest when nothing is playing yet.
		h := newPlaylistTestWrapper(t)
		h.manager.clientId = "web"

		h.manager.sendCurrentPlaylistToClient()

		event := h.wsEventManager.lastPlaylistServerEvent(t, ServerEventCurrentPlaylist)
		payload := decodeCurrentPlaylistPayload(t, event)
		require.Nil(t, payload.Playlist)
		require.Nil(t, payload.PlaylistEpisode)
	})

	t.Run("active playlist sends playlist and episode", func(t *testing.T) {
		// once a playlist is active, the client should receive both the queue and the selected episode.
		h := newPlaylistTestWrapper(t)
		h.manager.clientId = "web"
		episodeOne := newStreamPlaylistEpisode(101, 1, "1")
		playlist := newPlaylistFixture("queue", episodeOne)
		h.manager.currentPlaylistData = mo.Some(&playlistData{
			playlist: playlist,
			options:  newClientPlaylistOptions("web"),
		})
		h.manager.currentEpisode = mo.Some(episodeOne)

		h.manager.sendCurrentPlaylistToClient()

		event := h.wsEventManager.lastPlaylistServerEvent(t, ServerEventCurrentPlaylist)
		payload := decodeCurrentPlaylistPayload(t, event)
		require.NotNil(t, payload.Playlist)
		require.Equal(t, playlist.Name, payload.Playlist.Name)
		require.Len(t, payload.Playlist.Episodes, 1)
		require.NotNil(t, payload.PlaylistEpisode)
		require.Equal(t, episodeOne.Episode.BaseAnime.ID, payload.PlaylistEpisode.Episode.BaseAnime.ID)
		require.Equal(t, episodeOne.Episode.AniDBEpisode, payload.PlaylistEpisode.Episode.AniDBEpisode)
	})
}

func TestPlaylistManagerPlayEpisodeNextWithoutCurrentEpisode(t *testing.T) {
	// calling next with no current episode should pick the first incomplete entry instead of hanging on the mutex.
	h := newPlaylistTestWrapper(t)
	h.manager.clientId = "web"
	episodeOne := newStreamPlaylistEpisode(201, 1, "1")
	episodeTwo := newStreamPlaylistEpisode(201, 2, "2")
	playlist := newPlaylistFixture("queue", episodeOne, episodeTwo)
	h.manager.currentPlaylistData = mo.Some(&playlistData{
		playlist: playlist,
		options:  newClientPlaylistOptions("web"),
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.manager.PlayEpisode("next", false)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("PlayEpisode(next) deadlocked without a current episode")
	}

	require.True(t, h.manager.currentEpisode.IsPresent())
	require.Same(t, episodeOne, h.manager.currentEpisode.MustGet())

	event := h.wsEventManager.lastPlaylistServerEvent(t, ServerEventPlayEpisode)
	payload, ok := event.Payload.(playEpisodePayload)
	require.True(t, ok)
	require.Same(t, episodeOne, payload.PlaylistEpisode)
}

func TestPlaylistManagerPlayEpisodePrevious(t *testing.T) {
	// previous should switch back to the earlier entry and notify the client with that episode.
	h := newPlaylistTestWrapper(t)
	h.manager.clientId = "web"
	episodeOne := newStreamPlaylistEpisode(301, 1, "1")
	episodeTwo := newStreamPlaylistEpisode(301, 2, "2")
	playlist := newPlaylistFixture("queue", episodeOne, episodeTwo)
	h.manager.currentPlaylistData = mo.Some(&playlistData{
		playlist: playlist,
		options:  newClientPlaylistOptions("web"),
	})
	h.manager.currentEpisode = mo.Some(episodeTwo)

	h.manager.PlayEpisode("previous", false)

	require.True(t, h.manager.currentEpisode.IsPresent())
	require.Same(t, episodeOne, h.manager.currentEpisode.MustGet())

	event := h.wsEventManager.lastPlaylistServerEvent(t, ServerEventPlayEpisode)
	payload, ok := event.Payload.(playEpisodePayload)
	require.True(t, ok)
	require.Same(t, episodeOne, payload.PlaylistEpisode)
}

func TestPlaylistManagerMarkCurrentAsCompletedPersistsAndUpdatesProgress(t *testing.T) {
	// completing an episode should update both the stored playlist and the AniList progress bridge.
	h := newPlaylistTestWrapper(t)
	h.manager.clientId = "web"
	episodeOne := newStreamPlaylistEpisode(401, 4, "4")
	playlist := newPlaylistFixture("queue", episodeOne)
	h.persistPlaylist(t, playlist)
	h.manager.currentPlaylistData = mo.Some(&playlistData{
		playlist: playlist,
		options:  newClientPlaylistOptions("web"),
	})
	h.manager.currentEpisode = mo.Some(episodeOne)

	h.manager.markCurrentAsCompleted()

	require.True(t, episodeOne.IsCompleted)
	require.Eventually(t, func() bool {
		calls := h.platform.UpdateEntryProgressCalls()
		if len(calls) != 1 {
			return false
		}
		storedPlaylist, err := db_bridge.GetPlaylist(h.database, playlist.DbId)
		if err != nil || len(storedPlaylist.Episodes) != 1 {
			return false
		}
		return storedPlaylist.Episodes[0].IsCompleted
	}, time.Second, 10*time.Millisecond)

	calls := h.platform.UpdateEntryProgressCalls()
	require.Len(t, calls, 1)
	require.Equal(t, 401, calls[0].MediaID)
	require.Equal(t, 4, calls[0].Progress)
	require.NotNil(t, calls[0].TotalEpisodes)
	require.Equal(t, 12, *calls[0].TotalEpisodes)
}

func TestPlaylistManagerStopPlaylistDeletesCompletedPlaylistAndResetsState(t *testing.T) {
	// stopping an already-finished playlist should clear in-memory state and remove the stored queue.
	h := newPlaylistTestWrapper(t)
	h.manager.clientId = "web"
	episodeOne := newStreamPlaylistEpisode(501, 1, "1")
	episodeOne.IsCompleted = true
	playlist := newPlaylistFixture("queue", episodeOne)
	h.persistPlaylist(t, playlist)
	var canceled bool
	h.manager.currentPlaylistData = mo.Some(&playlistData{
		playlist: playlist,
		options:  newClientPlaylistOptions("web"),
	})
	h.manager.currentEpisode = mo.Some(episodeOne)
	h.manager.cancel = func() {
		canceled = true
	}

	h.manager.StopPlaylist("done")

	require.True(t, canceled)
	require.True(t, h.manager.currentPlaylistData.IsAbsent())
	require.True(t, h.manager.currentEpisode.IsAbsent())
	require.Nil(t, h.manager.cancel)
	require.Eventually(t, func() bool {
		_, err := db_bridge.GetPlaylist(h.database, playlist.DbId)
		return err != nil
	}, time.Second, 10*time.Millisecond)

	require.Contains(t, h.wsEventManager.eventTypesForClient("web"), events.InvalidateQueries)
	require.Contains(t, h.wsEventManager.eventTypesForClient("web"), events.InfoToast)
	invalidatePayload := h.wsEventManager.lastDirectedEvent(t, events.InvalidateQueries).payload
	queries, ok := invalidatePayload.([]string)
	require.True(t, ok)
	require.Equal(t, []string{events.GetPlaylistsEndpoint}, queries)
}

func TestPlaylistManagerListenToEventsStartsPlaylistAndServesCurrentPlaylist(t *testing.T) {
	// the common client flow is start playlist first, then ask for the current queue state.
	h := newPlaylistTestWrapper(t)
	episodeOne := newStreamPlaylistEpisode(601, 1, "1")
	episodeTwo := newStreamPlaylistEpisode(601, 2, "2")
	playlist := newPlaylistFixture("queue", episodeOne, episodeTwo)
	h.persistPlaylist(t, playlist)

	h.sendPlaylistClientEvent(t, "web", ClientEvent{Type: ClientEventStart, Payload: startPlaylistPayload{
		DbId:                    playlist.DbId,
		ClientId:                "web",
		LocalFilePlaybackMethod: ClientPlaybackMethodNativePlayer,
		StreamPlaybackMethod:    ClientPlaybackMethodNativePlayer,
	}})

	require.Eventually(t, func() bool {
		return h.manager.currentEpisode.IsPresent() && h.manager.currentEpisode.MustGet().Episode.AniDBEpisode == "1"
	}, time.Second, 10*time.Millisecond)

	playEvent := h.wsEventManager.waitForPlaylistServerEvent(t, ServerEventPlayEpisode)
	playPayload, ok := playEvent.Payload.(playEpisodePayload)
	require.True(t, ok)
	require.Equal(t, episodeOne.Episode.BaseAnime.ID, playPayload.PlaylistEpisode.Episode.BaseAnime.ID)
	require.Equal(t, episodeOne.Episode.AniDBEpisode, playPayload.PlaylistEpisode.Episode.AniDBEpisode)

	h.sendPlaylistClientEvent(t, "web", ClientEvent{Type: ClientEventCurrentPlaylist})

	currentEvent := h.wsEventManager.waitForPlaylistServerEvent(t, ServerEventCurrentPlaylist)
	currentPayload := decodeCurrentPlaylistPayload(t, currentEvent)
	require.NotNil(t, currentPayload.Playlist)
	require.Equal(t, playlist.Name, currentPayload.Playlist.Name)
	require.NotNil(t, currentPayload.PlaylistEpisode)
	require.Equal(t, episodeOne.Episode.AniDBEpisode, currentPayload.PlaylistEpisode.Episode.AniDBEpisode)

	h.manager.StopPlaylist("done")
}

func TestPlaylistManagerNativeLifecycleAdvancesToNextEpisode(t *testing.T) {
	// the normal native-player loop is: load metadata, complete the episode, then move to the next one on ended.
	h := newPlaylistTestWrapper(t)
	episodeOne := newStreamPlaylistEpisode(701, 1, "1")
	episodeTwo := newStreamPlaylistEpisode(701, 2, "2")
	playlist := newPlaylistFixture("queue", episodeOne, episodeTwo)
	h.manager.clientId = "web"
	h.persistPlaylist(t, playlist)
	h.manager.startPlaylist(playlist, newClientPlaylistOptions("web"))

	require.Eventually(t, func() bool {
		return h.manager.currentEpisode.IsPresent() && h.manager.currentEpisode.MustGet().Episode.AniDBEpisode == "1"
	}, time.Second, 10*time.Millisecond)

	h.sendNativeLoadedSequence(t, "web", episodeOne)

	require.Eventually(t, func() bool {
		return h.manager.playerType.Load() == MpvCorePlayer && h.manager.state.Load() == StateStarted
	}, time.Second, 10*time.Millisecond)

	h.sendMpvCoreClientEvent("web", mpvcore.ClientEventCompleted, map[string]any{
		"id":          "playback-1",
		"clientId":    "web",
		"currentTime": 1200.0,
		"duration":    1200.0,
		"paused":      true,
	})

	require.Eventually(t, func() bool {
		return h.manager.state.Load() == StateCompleted && episodeOne.IsCompleted
	}, time.Second, 10*time.Millisecond)

	h.sendMpvCoreClientEvent("web", mpvcore.ClientEventEnded, map[string]any{"autoNext": false})

	require.Eventually(t, func() bool {
		return h.manager.currentEpisode.IsPresent() && h.manager.currentEpisode.MustGet().Episode.AniDBEpisode == "2"
	}, time.Second, 10*time.Millisecond)
	playEvent := h.wsEventManager.lastPlaylistServerEvent(t, ServerEventPlayEpisode)
	playPayload, ok := playEvent.Payload.(playEpisodePayload)
	require.True(t, ok)
	require.Same(t, episodeTwo, playPayload.PlaylistEpisode)
	require.True(t, episodeOne.IsCompleted)

	h.manager.StopPlaylist("done")
}

func TestPlaylistManagerNativeTerminationStopsPlaylist(t *testing.T) {
	// when the native player closes during playback, the playlist should stop and clear its state.
	h := newPlaylistTestWrapper(t)
	episodeOne := newStreamPlaylistEpisode(801, 1, "1")
	playlist := newPlaylistFixture("queue", episodeOne)
	h.manager.clientId = "web"
	h.persistPlaylist(t, playlist)
	h.manager.startPlaylist(playlist, newClientPlaylistOptions("web"))

	require.Eventually(t, func() bool {
		return h.manager.currentEpisode.IsPresent() && h.manager.currentEpisode.MustGet().Episode.AniDBEpisode == "1"
	}, time.Second, 10*time.Millisecond)

	h.sendNativeLoadedSequence(t, "web", episodeOne)
	require.Eventually(t, func() bool {
		return h.manager.playerType.Load() == MpvCorePlayer && h.manager.state.Load() == StateStarted
	}, time.Second, 10*time.Millisecond)

	h.sendMpvCoreClientEvent("web", mpvcore.ClientEventTerminated, map[string]any{
		"id":           "playback-1",
		"clientId":     "web",
		"playbackType": mpvcore.PlaybackTypeTorrent,
	})

	require.Eventually(t, func() bool {
		return h.manager.currentPlaylistData.IsAbsent() && h.manager.currentEpisode.IsAbsent()
	}, time.Second, 10*time.Millisecond)
	require.Contains(t, h.wsEventManager.eventTypesForClient("web"), events.InfoToast)
}

func TestPlaylistManagerListenToEventsReopensCurrentEpisode(t *testing.T) {
	// reopen should send the currently selected episode back to the client without changing selection.
	h := newPlaylistTestWrapper(t)
	episodeOne := newStreamPlaylistEpisode(901, 1, "1")
	playlist := newPlaylistFixture("queue", episodeOne)
	h.manager.currentPlaylistData = mo.Some(&playlistData{playlist: playlist, options: newClientPlaylistOptions("web")})
	h.manager.currentEpisode = mo.Some(episodeOne)
	h.manager.clientId = "web"

	h.sendPlaylistClientEvent(t, "web", ClientEvent{Type: ClientEventReopenEpisode})

	playEvent := h.wsEventManager.waitForPlaylistServerEvent(t, ServerEventPlayEpisode)
	playPayload, ok := playEvent.Payload.(playEpisodePayload)
	require.True(t, ok)
	require.Same(t, episodeOne, playPayload.PlaylistEpisode)
}

type playlistTestWrapper struct {
	database        *db.Database
	wsEventManager  *recordingPlaylistWSEventManager
	platform        *testmocks.FakePlatform
	playbackManager *playbackmanager.PlaybackManager
	mpvCore         *mpvcore.MpvCore
	manager         *Manager
}

func newPlaylistTestWrapper(t *testing.T) *playlistTestWrapper {
	t.Helper()

	env := testutil.NewTestEnv(t)
	logger := util.NewLogger()
	database := env.MustNewDatabase(logger)
	wsEventManager := &recordingPlaylistWSEventManager{MockWSEventManager: events.NewMockWSEventManager(logger)}
	platformImpl := testmocks.NewFakePlatformBuilder().Build()
	platformInterface := platform.Platform(platformImpl)
	var provider metadata_provider.Provider
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: env.NewCacher("playlist-continuity"),
		Logger:     logger,
		Database:   database,
	})
	continuityManager.SetSettings(&continuity.Settings{WatchContinuityEnabled: false})
	mpvCore := mpvcore.New(mpvcore.NewMpvCoreOptions{
		WsEventManager:      wsEventManager,
		Logger:              logger,
		MetadataProviderRef: util.NewRef(provider),
		ContinuityManager:   continuityManager,
		PlatformRef:         util.NewRef(platformInterface),
		IsOfflineRef:        util.NewRef(false),
	})
	mpvCore.SetSettings(&models.Settings{
		Library:     &models.LibrarySettings{AutoUpdateProgress: false},
		MediaPlayer: &models.MediaPlayerSettings{},
	})
	t.Cleanup(mpvCore.Shutdown)
	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:              logger,
		Database:            database,
		WSEventManager:      wsEventManager,
		PlatformRef:         util.NewRef(platformInterface),
		MetadataProviderRef: util.NewRef(provider),
		IsOfflineRef:        util.NewRef(false),
	})
	playbackManager.SetMediaPlayerRepository(mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
		Logger:         logger,
		Default:        "",
		WSEventManager: wsEventManager,
	}))

	mediacoreCoordinator := mediacore.NewCoordinator(mediacore.NewCoordinatorOptions{
		Logger:              logger,
		MetadataProviderRef: util.NewRef(provider),
		ContinuityManager:   continuityManager,
		PlatformRef:         util.NewRef(platformInterface),
		IsOfflineRef:        util.NewRef(false),
		Backends: map[player.Target]mediacore.Backend{
			player.TargetMpvCore: mpvcore.NewAdapter(mpvCore),
		},
	})
	t.Cleanup(func() { _ = mediacoreCoordinator.Close() })

	manager := NewManager(&NewManagerOptions{
		PlaybackManager:      playbackManager,
		MediacoreCoordinator: mediacoreCoordinator,
		Logger:               logger,
		PlatformRef:          util.NewRef(platformInterface),
		WSEventManager:       wsEventManager,
		Database:             database,
	})
	manager.currentPlaylistData = mo.None[*playlistData]()
	manager.currentEpisode = mo.None[*anime.PlaylistEpisode]()
	manager.state.Store(StateIdle)
	manager.playerType.Store("")

	return &playlistTestWrapper{
		database:        database,
		wsEventManager:  wsEventManager,
		platform:        platformImpl,
		playbackManager: playbackManager,
		mpvCore:         mpvCore,
		manager:         manager,
	}
}

func (h *playlistTestWrapper) sendPlaylistClientEvent(t *testing.T, clientID string, payload ClientEvent) {
	t.Helper()
	h.waitForClientSubscriber(t, "playlist-manager")

	h.wsEventManager.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: clientID,
		Type:     events.PlaylistEvent,
		Payload:  payload,
	})
}

func (h *playlistTestWrapper) sendMpvCoreClientEvent(clientID string, eventType mpvcore.ClientEventType, payload interface{}) {
	h.wsEventManager.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: clientID,
		Type:     events.MpvCoreEventType,
		Payload: map[string]interface{}{
			"clientId": clientID,
			"type":     eventType,
			"payload":  payload,
		},
	})
}

func (h *playlistTestWrapper) sendNativeLoadedSequence(t *testing.T, clientID string, episode *anime.PlaylistEpisode) {
	t.Helper()

	h.manager.mediacoreCoordinator.Watch(player.TargetMpvCore, clientID, &player.PlaybackInfo{
		ID:           "playback-1",
		PlaybackType: player.PlaybackTypeTorrent,
		Media:        episode.Episode.BaseAnime,
		Episode:      episode.Episode,
	})
	h.sendMpvCoreClientEvent(clientID, mpvcore.ClientEventPlaybackLoaded, map[string]interface{}{
		"id":       "playback-1",
		"clientId": clientID,
	})
	h.sendMpvCoreClientEvent(clientID, mpvcore.ClientEventLoadedMetadata, map[string]interface{}{
		"id":          "playback-1",
		"clientId":    clientID,
		"currentTime": 0.0,
		"duration":    1200.0,
		"paused":      false,
	})
}

func (h *playlistTestWrapper) waitForClientSubscriber(t *testing.T, id string) {
	t.Helper()

	require.Eventually(t, func() bool {
		return h.wsEventManager.ClientEventSubscribers.Has(id)
	}, time.Second, 10*time.Millisecond)
}

func (h *playlistTestWrapper) persistPlaylist(t *testing.T, playlist *anime.Playlist) {
	t.Helper()

	data, err := json.Marshal(playlist.Episodes)
	require.NoError(t, err)
	entry := &models.Playlist{
		Name:  playlist.Name,
		Value: data,
	}
	require.NoError(t, h.database.Gorm().Create(entry).Error)
	playlist.DbId = entry.ID
}

type recordingDirectedEvent struct {
	clientID  string
	eventType string
	payload   interface{}
}

type recordingPlaylistWSEventManager struct {
	*events.MockWSEventManager
	mu        sync.Mutex
	directed  []recordingDirectedEvent
	broadcast []recordingDirectedEvent
}

func (m *recordingPlaylistWSEventManager) SendEvent(t string, payload interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcast = append(m.broadcast, recordingDirectedEvent{eventType: t, payload: payload})
}

func (m *recordingPlaylistWSEventManager) SendEventTo(clientID string, t string, payload interface{}, _ ...bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.directed = append(m.directed, recordingDirectedEvent{clientID: clientID, eventType: t, payload: payload})
}

func (m *recordingPlaylistWSEventManager) lastPlaylistServerEvent(t *testing.T, eventType PlaylistServerEventType) ServerEvent {
	t.Helper()

	event, ok := m.tryLastPlaylistServerEvent(eventType)
	if ok {
		return event
	}
	t.Fatalf("playlist server event %s not found", eventType)
	return ServerEvent{}
}

func (m *recordingPlaylistWSEventManager) waitForPlaylistServerEvent(t *testing.T, eventType PlaylistServerEventType) ServerEvent {
	t.Helper()

	var event ServerEvent
	require.Eventually(t, func() bool {
		var ok bool
		event, ok = m.tryLastPlaylistServerEvent(eventType)
		return ok
	}, time.Second, 10*time.Millisecond)
	return event
}

func (m *recordingPlaylistWSEventManager) tryLastPlaylistServerEvent(eventType PlaylistServerEventType) (ServerEvent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.directed) - 1; i >= 0; i-- {
		if m.directed[i].eventType != string(events.PlaylistEvent) {
			continue
		}
		event, ok := m.directed[i].payload.(ServerEvent)
		if ok && event.Type == eventType {
			return event, true
		}
	}
	return ServerEvent{}, false
}

func (m *recordingPlaylistWSEventManager) lastDirectedEvent(t *testing.T, eventType string) recordingDirectedEvent {
	t.Helper()

	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.directed) - 1; i >= 0; i-- {
		if m.directed[i].eventType == eventType {
			return m.directed[i]
		}
	}
	t.Fatalf("directed event %s not found", eventType)
	return recordingDirectedEvent{}
}

func (m *recordingPlaylistWSEventManager) eventTypesForClient(clientID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	ret := make([]string, 0)
	for _, event := range m.directed {
		if event.clientID == clientID {
			ret = append(ret, event.eventType)
		}
	}
	return ret
}

type currentPlaylistPayload struct {
	PlaylistEpisode *anime.PlaylistEpisode `json:"playlistEpisode"`
	Playlist        *anime.Playlist        `json:"playlist"`
}

func decodeCurrentPlaylistPayload(t *testing.T, event ServerEvent) currentPlaylistPayload {
	t.Helper()

	data, err := json.Marshal(event.Payload)
	require.NoError(t, err)
	var payload currentPlaylistPayload
	require.NoError(t, json.Unmarshal(data, &payload))
	return payload
}

func newClientPlaylistOptions(clientID string) *startPlaylistPayload {
	return &startPlaylistPayload{
		ClientId:                clientID,
		LocalFilePlaybackMethod: ClientPlaybackMethodNativePlayer,
		StreamPlaybackMethod:    ClientPlaybackMethodNativePlayer,
	}
}

func newPlaylistFixture(name string, episodes ...*anime.PlaylistEpisode) *anime.Playlist {
	playlist := anime.NewPlaylist(name)
	playlist.SetEpisodes(episodes)
	return playlist
}

func newStreamPlaylistEpisode(mediaID int, episodeNumber int, aniDBEpisode string) *anime.PlaylistEpisode {
	title := "playlist anime"
	media := testmocks.NewBaseAnimeBuilder(mediaID, title).
		WithUserPreferredTitle(title).
		WithEpisodes(12).
		Build()
	media.IDMal = nil

	return &anime.PlaylistEpisode{
		Episode: &anime.Episode{
			BaseAnime:      media,
			EpisodeNumber:  episodeNumber,
			ProgressNumber: episodeNumber,
			AniDBEpisode:   aniDBEpisode,
			DisplayTitle:   "Episode",
		},
		WatchType: anime.WatchTypeTorrent,
	}
}
