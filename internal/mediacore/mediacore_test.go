package mediacore

import (
	"seanime/internal/api/anilist"
	"seanime/internal/continuity"
	"seanime/internal/library/anime"
	"seanime/internal/testutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	target      Target
	eventsCh    chan Event
	watchedInfo *PlaybackInfo
}

var _ Backend = (*mockBackend)(nil)

func newMockBackend(target Target) *mockBackend {
	return &mockBackend{
		target:   target,
		eventsCh: make(chan Event, 10),
	}
}

func (m *mockBackend) Target() Target {
	return m.target
}

func (m *mockBackend) OpenAndAwait(clientID, state string) {}
func (m *mockBackend) AbortOpen(clientID, reason string)   {}
func (m *mockBackend) Watch(clientID string, info *PlaybackInfo) {
	m.watchedInfo = info
}
func (m *mockBackend) Error(clientID string, err error) {}
func (m *mockBackend) Execute(session SessionKey, cmd Command) error {
	return nil
}
func (m *mockBackend) Terminate(session SessionKey) {}
func (m *mockBackend) Events() <-chan Event {
	return m.eventsCh
}
func (m *mockBackend) Close() error {
	close(m.eventsCh)
	return nil
}
func (m *mockBackend) PullStatus() (PlaybackStatus, bool) {
	return PlaybackStatus{}, false
}
func (m *mockBackend) GetPlaylist() (*PlaylistState, bool) {
	return nil, false
}
func (m *mockBackend) GetSkipData() (*SkipData, bool) {
	return nil, false
}

func TestCoordinatorRoutingAndStaleRejection(t *testing.T) {
	mbMpv := newMockBackend(TargetMpvCore)
	mbVc := newMockBackend(TargetVideoCore)

	backends := map[Target]Backend{
		TargetMpvCore:   mbMpv,
		TargetVideoCore: mbVc,
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
	coordinator.OpenAndAwait(TargetMpvCore, "client-1", "Opening...")
	sess, ok = coordinator.GetActiveSession()
	require.False(t, ok) // session shouldnt be active
	require.Equal(t, TargetMpvCore, sess.Target)
	require.Equal(t, "client-1", sess.ClientID)
	require.Empty(t, sess.PlaybackID)

	// 3. send stale status event from VideoCore
	mbVc.eventsCh <- &StatusEvent{
		BaseEvent: BaseEvent{
			Session: SessionKey{
				Target:     TargetVideoCore,
				ClientID:   "client-1",
				PlaybackID: "stale-1",
			},
		},
		CurrentTime: 10.0,
		Duration:    100.0,
	}

	mbMpv.eventsCh <- &PlaybackLoadedEvent{
		BaseEvent: BaseEvent{
			Session: SessionKey{
				Target:     TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "play-1",
			},
		},
		State: PlaybackState{
			ClientID: "client-1",
			PlaybackInfo: &PlaybackInfo{
				ID:           "play-1",
				Target:       TargetMpvCore,
				PlaybackType: PlaybackTypeLocalFile,
			},
		},
	}

	var lastEvent Event
	select {
	case ev := <-sub.Events():
		lastEvent = ev
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for event")
	}

	loadedEv, ok := lastEvent.(*PlaybackLoadedEvent)
	require.True(t, ok)
	require.Equal(t, "play-1", loadedEv.State.PlaybackInfo.ID)

	sess, ok = coordinator.GetActiveSession()
	require.True(t, ok)
	require.Equal(t, "play-1", sess.PlaybackID)

	mbMpv.eventsCh <- &StatusEvent{
		BaseEvent: BaseEvent{
			Session: SessionKey{
				Target:     TargetMpvCore,
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

	statusEv, ok := lastEvent.(*StatusEvent)
	require.True(t, ok)
	require.Equal(t, 20.0, statusEv.CurrentTime)

	mbMpv.eventsCh <- &StatusEvent{
		BaseEvent: BaseEvent{
			Session: SessionKey{
				Target:     TargetMpvCore,
				ClientID:   "client-1",
				PlaybackID: "stale-2",
			},
		},
		CurrentTime: 30.0,
		Duration:    100.0,
	}

	mbMpv.eventsCh <- &StatusEvent{
		BaseEvent: BaseEvent{
			Session: SessionKey{
				Target:     TargetMpvCore,
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

	statusEv, ok = lastEvent.(*StatusEvent)
	require.True(t, ok)
	require.Equal(t, 40.0, statusEv.CurrentTime)
}

func TestCoordinatorStaleCommandRejection(t *testing.T) {
	mbMpv := newMockBackend(TargetMpvCore)

	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger: new(zerolog.Nop()),
		Backends: map[Target]Backend{
			TargetMpvCore: mbMpv,
		},
	})
	defer coordinator.Close()

	coordinator.Watch(TargetMpvCore, "client-1", &PlaybackInfo{ID: "play-1"})

	err := coordinator.Execute(SessionKey{
		Target:     TargetMpvCore,
		ClientID:   "client-1",
		PlaybackID: "play-1",
	}, Command{Type: CommandPause})
	require.NoError(t, err)

	err = coordinator.Execute(SessionKey{
		Target:     TargetMpvCore,
		ClientID:   "client-1",
		PlaybackID: "stale-session",
	}, Command{Type: CommandPause})
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

	backend := newMockBackend(TargetMpvCore)
	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger:            logger,
		ContinuityManager: continuityManager,
		Backends: map[Target]Backend{
			TargetMpvCore: backend,
		},
	})
	defer coordinator.Close()

	coordinator.Watch(TargetMpvCore, "client-1", &PlaybackInfo{
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

	backend := newMockBackend(TargetMpvCore)
	coordinator := NewCoordinator(NewCoordinatorOptions{
		Logger:            logger,
		ContinuityManager: continuityManager,
		Backends: map[Target]Backend{
			TargetMpvCore: backend,
		},
	})
	defer coordinator.Close()
	coordinator.SetupSharedEffects()
	coordinator.OpenAndAwait(TargetMpvCore, "client-1", "Opening...")

	session := SessionKey{Target: TargetMpvCore, ClientID: "client-1", PlaybackID: "play-1"}
	backend.eventsCh <- &PlaybackLoadedEvent{
		BaseEvent: BaseEvent{Session: session},
		State: PlaybackState{
			ClientID: "client-1",
			PlaybackInfo: &PlaybackInfo{
				ID:           "play-1",
				PlaybackType: PlaybackTypeLocalFile,
				Media:        &anilist.BaseAnime{ID: 84},
				Episode:      &anime.Episode{EpisodeNumber: 6},
			},
		},
	}
	backend.eventsCh <- &PausedEvent{
		BaseEvent:   BaseEvent{Session: session},
		CurrentTime: 61,
		Duration:    100,
	}

	require.Eventually(t, func() bool {
		history := continuityManager.GetWatchHistoryItem(84)
		return history != nil && history.Found && history.Item != nil && history.Item.CurrentTime == 61
	}, time.Second, 10*time.Millisecond)
}
