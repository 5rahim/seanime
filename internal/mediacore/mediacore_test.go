package mediacore

import (
	"seanime/internal/api/anilist"
	"seanime/internal/continuity"
	"seanime/internal/library/anime"
	"seanime/internal/player"
	"seanime/internal/testutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	target      player.Target
	eventsCh    chan player.Event
	watchedInfo *player.PlaybackInfo
}

var _ Backend = (*mockBackend)(nil)

func newMockBackend(target player.Target) *mockBackend {
	return &mockBackend{
		target:   target,
		eventsCh: make(chan player.Event, 10),
	}
}

func (m *mockBackend) Target() player.Target {
	return m.target
}

func (m *mockBackend) OpenAndAwait(clientID, state string) {}
func (m *mockBackend) AbortOpen(clientID, reason string)   {}
func (m *mockBackend) Watch(clientID string, info *player.PlaybackInfo) {
	m.watchedInfo = info
}
func (m *mockBackend) Error(clientID string, err error) {}
func (m *mockBackend) Execute(session player.SessionKey, cmd player.Command) error {
	return nil
}
func (m *mockBackend) Terminate(session player.SessionKey) {}
func (m *mockBackend) Events() <-chan player.Event {
	return m.eventsCh
}
func (m *mockBackend) Close() error {
	close(m.eventsCh)
	return nil
}
func (m *mockBackend) PullStatus() (player.PlaybackStatus, bool) {
	return player.PlaybackStatus{}, false
}
func (m *mockBackend) GetPlaylist() (*player.PlaylistState, bool) {
	return nil, false
}
func (m *mockBackend) GetSkipData() (*player.SkipData, bool) {
	return nil, false
}

func TestCoordinatorRoutingAndStaleRejection(t *testing.T) {
	mbMpv := newMockBackend(player.TargetMpvCore)
	mbVc := newMockBackend(player.TargetVideoCore)

	backends := map[player.Target]Backend{
		player.TargetMpvCore:   mbMpv,
		player.TargetVideoCore: mbVc,
	}

	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger:   new(zerolog.Nop()),
		Backends: backends,
	})
	defer coordinator.Close()

	sub := coordinator.Subscribe("test")
	defer coordinator.Unsubscribe("test")

	sess, ok := coordinator.GetActiveSession()
	require.False(t, ok)
	require.Empty(t, sess.PlaybackID)

	// mpvcore, client-1
	coordinator.OpenAndAwait(player.TargetMpvCore, "client-1", "Opening...")
	sess, ok = coordinator.GetActiveSession()
	require.False(t, ok) // session shouldnt be active
	require.Equal(t, player.TargetMpvCore, sess.Target)
	require.Equal(t, "client-1", sess.ClientID)
	require.Empty(t, sess.PlaybackID)

	// 3. send stale status event from VideoCore
	mbVc.eventsCh <- &player.StatusEvent{
		BaseEvent: player.BaseEvent{
			Session: player.SessionKey{
				Target:     player.TargetVideoCore,
				ClientID:   "client-1",
				PlaybackID: "stale-1",
			},
		},
		CurrentTime: 10.0,
		Duration:    100.0,
	}

	mbMpv.eventsCh <- &player.PlaybackLoadedEvent{
		BaseEvent: player.BaseEvent{
			Session: player.SessionKey{
				Target:     player.TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "play-1",
			},
		},
		State: player.PlaybackState{
			ClientID: "client-1",
			PlaybackInfo: &player.PlaybackInfo{
				ID:           "play-1",
				Target:       player.TargetMpvCore,
				PlaybackType: player.PlaybackTypeLocalFile,
			},
		},
	}

	var lastEvent player.Event
	select {
	case ev := <-sub.Events():
		lastEvent = ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for event")
	}

	loadedEv, ok := lastEvent.(*player.PlaybackLoadedEvent)
	require.True(t, ok)
	require.Equal(t, "play-1", loadedEv.State.PlaybackInfo.ID)

	sess, ok = coordinator.GetActiveSession()
	require.True(t, ok)
	require.Equal(t, "play-1", sess.PlaybackID)

	mbMpv.eventsCh <- &player.StatusEvent{
		BaseEvent: player.BaseEvent{
			Session: player.SessionKey{
				Target:     player.TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "play-1",
			},
		},
		CurrentTime: 20.0,
		Duration:    100.0,
		Paused:      false,
	}

	select {
	case ev := <-sub.Events():
		lastEvent = ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for status event")
	}

	statusEv, ok := lastEvent.(*player.StatusEvent)
	require.True(t, ok)
	require.Equal(t, 20.0, statusEv.CurrentTime)

	mbMpv.eventsCh <- &player.StatusEvent{
		BaseEvent: player.BaseEvent{
			Session: player.SessionKey{
				Target:     player.TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "stale-2",
			},
		},
		CurrentTime: 30.0,
		Duration:    100.0,
	}

	mbMpv.eventsCh <- &player.StatusEvent{
		BaseEvent: player.BaseEvent{
			Session: player.SessionKey{
				Target:     player.TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "play-1",
			},
		},
		CurrentTime: 40.0,
		Duration:    100.0,
	}

	select {
	case ev := <-sub.Events():
		lastEvent = ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for final status event")
	}

	statusEv, ok = lastEvent.(*player.StatusEvent)
	require.True(t, ok)
	require.Equal(t, 40.0, statusEv.CurrentTime)
}

func TestCoordinatorStaleCommandRejection(t *testing.T) {
	mbMpv := newMockBackend(player.TargetMpvCore)

	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger: new(zerolog.Nop()),
		Backends: map[player.Target]Backend{
			player.TargetMpvCore: mbMpv,
		},
	})
	defer coordinator.Close()

	coordinator.Watch(player.TargetMpvCore, "client-1", &player.PlaybackInfo{ID: "play-1"})

	err := coordinator.Execute(player.SessionKey{
		Target:     player.TargetMpvCore,
		ClientID:   "client-1",
		PlaybackID: "play-1",
	}, player.Command{Type: player.CommandPause})
	require.NoError(t, err)

	err = coordinator.Execute(player.SessionKey{
		Target:     player.TargetMpvCore,
		ClientID:   "client-1",
		PlaybackID: "stale-session",
	}, player.Command{Type: player.CommandPause})
	require.Error(t, err)
	require.Contains(t, err.Error(), "session mismatch or stale command")
}

func TestCoordinatorRestoresContinuityBeforeWatch(t *testing.T) {
	env := testutil.NewTestEnv(t)
	logger := env.Logger()
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: env.NewCacher("continuity-restore"),
		Logger:     logger,
	})
	continuityManager.SetSettings(&continuity.Settings{WatchContinuityEnabled: true})
	require.NoError(t, continuityManager.UpdateWatchHistoryItem(&continuity.UpdateWatchHistoryItemOptions{
		MediaId:       42,
		EpisodeNumber: 3,
		CurrentTime:   45,
		Duration:      100,
		Kind:          continuity.MediastreamKind,
	}))

	backend := newMockBackend(player.TargetMpvCore)
	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger:            logger,
		ContinuityManager: continuityManager,
		Backends: map[player.Target]Backend{
			player.TargetMpvCore: backend,
		},
	})
	defer coordinator.Close()

	coordinator.Watch(player.TargetMpvCore, "client-1", &player.PlaybackInfo{
		ID:      "play-1",
		Media:   &anilist.BaseAnime{ID: 42},
		Episode: &anime.Episode{EpisodeNumber: 3},
	})

	require.NotNil(t, backend.watchedInfo)
	require.NotNil(t, backend.watchedInfo.InitialState)
	require.NotNil(t, backend.watchedInfo.InitialState.CurrentTime)
	require.Equal(t, 45.0, *backend.watchedInfo.InitialState.CurrentTime)
}

func TestCoordinatorPersistsContinuityOnPause(t *testing.T) {
	env := testutil.NewTestEnv(t)
	logger := env.Logger()
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: env.NewCacher("continuity-persist"),
		Logger:     logger,
	})
	continuityManager.SetSettings(&continuity.Settings{WatchContinuityEnabled: true})

	backend := newMockBackend(player.TargetMpvCore)
	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger:            logger,
		ContinuityManager: continuityManager,
		Backends: map[player.Target]Backend{
			player.TargetMpvCore: backend,
		},
	})
	defer coordinator.Close()
	coordinator.SetupSharedEffects()
	coordinator.OpenAndAwait(player.TargetMpvCore, "client-1", "Opening...")

	session := player.SessionKey{Target: player.TargetMpvCore, ClientID: "client-1", PlaybackID: "play-1"}
	backend.eventsCh <- &player.PlaybackLoadedEvent{
		BaseEvent: player.BaseEvent{Session: session},
		State: player.PlaybackState{
			ClientID: "client-1",
			PlaybackInfo: &player.PlaybackInfo{
				ID:           "play-1",
				PlaybackType: player.PlaybackTypeLocalFile,
				Media:        &anilist.BaseAnime{ID: 84},
				Episode:      &anime.Episode{EpisodeNumber: 6},
			},
		},
	}
	backend.eventsCh <- &player.PausedEvent{
		BaseEvent:   player.BaseEvent{Session: session},
		CurrentTime: 61,
		Duration:    100,
	}

	require.Eventually(t, func() bool {
		history := continuityManager.GetWatchHistoryItem(84)
		return history != nil && history.Found && history.Item != nil && history.Item.CurrentTime == 61
	}, time.Second, 10*time.Millisecond)
}
