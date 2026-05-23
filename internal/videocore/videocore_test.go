package videocore

import (
	"encoding/json"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type recordedWSEvent struct {
	clientId  string
	eventType string
	payload   interface{}
}

type recordingWSEventManager struct {
	videoCoreSubscriber *events.ClientEventSubscriber
	sent                []recordedWSEvent
}

func newRecordingWSEventManager() *recordingWSEventManager {
	return &recordingWSEventManager{
		videoCoreSubscriber: &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 1)},
	}
}

func (m *recordingWSEventManager) SendEvent(string, interface{}) {}

func (m *recordingWSEventManager) SendEventTo(clientId string, eventType string, payload interface{}, _ ...bool) {
	m.sent = append(m.sent, recordedWSEvent{clientId: clientId, eventType: eventType, payload: payload})
}

func (m *recordingWSEventManager) GetClientIds() []string { return nil }

func (m *recordingWSEventManager) GetClientPlatform(string) string { return "" }

func (m *recordingWSEventManager) SubscribeToClientEvents(string) *events.ClientEventSubscriber {
	return &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 1)}
}

func (m *recordingWSEventManager) SubscribeToClientNativePlayerEvents(string) *events.ClientEventSubscriber {
	return &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 1)}
}

func (m *recordingWSEventManager) SubscribeToClientVideoCoreEvents(string) *events.ClientEventSubscriber {
	return m.videoCoreSubscriber
}

func (m *recordingWSEventManager) SubscribeToClientNakamaEvents(string) *events.ClientEventSubscriber {
	return &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 1)}
}

func (m *recordingWSEventManager) SubscribeToClientPlaylistEvents(string) *events.ClientEventSubscriber {
	return &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 1)}
}

func (m *recordingWSEventManager) UnsubscribeFromClientEvents(string) {}

func (m *recordingWSEventManager) MockSendVideoCoreEvent(event ClientEvent) {
	m.videoCoreSubscriber.Channel <- &events.WebsocketClientEvent{
		ClientID: event.ClientId,
		Type:     events.VideoCoreEventType,
		Payload:  event,
	}
}

func decodeVideoCoreEnvelope(t *testing.T, payload interface{}) map[string]interface{} {
	t.Helper()

	marshaled, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(marshaled, &decoded))

	return decoded
}

func mustMarshalRaw(t *testing.T, payload interface{}) json.RawMessage {
	t.Helper()

	marshaled, err := json.Marshal(payload)
	require.NoError(t, err)

	return marshaled
}

func newPlaybackState(playbackID string) *PlaybackState {
	return &PlaybackState{
		ClientId:   "player-client",
		PlayerType: WebPlayer,
		PlaybackInfo: &VideoPlaybackInfo{
			Id:           playbackID,
			PlaybackType: PlaybackTypeOnlinestream,
			Episode:      &anime.Episode{},
		},
	}
}

func TestVideoTerminatedEventUsesPayloadClientIDWithoutPlaybackState(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	sub := vc.Subscribe("test")
	t.Cleanup(func() {
		vc.Unsubscribe("test")
		vc.Shutdown()
	})

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: ClientEvent{
			ClientId: "player-client",
			Type:     PlayerEventVideoTerminated,
		},
	})

	select {
	case rawEvent := <-sub.Events():
		event, ok := rawEvent.(*VideoTerminatedEvent)
		require.True(t, ok)
		require.Equal(t, "player-client", event.GetClientId())
		require.Equal(t, NativePlayer, event.GetPlayerType())
	case <-time.After(time.Second):
		t.Fatal("expected terminated event")
	}
}

func TestSetSkipDataSendsSanitizedOverride(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))
	vc.SetSkipData(&SkipData{
		Op: &SkipDataEntry{Interval: SkipInterval{StartTime: 12, EndTime: 42}},
		Ed: &SkipDataEntry{Interval: SkipInterval{StartTime: 20, EndTime: 60}},
	})

	// overlapping ed ranges should be dropped before they reach the player.
	require.Len(t, ws.sent, 1)
	require.Equal(t, "player-client", ws.sent[0].clientId)
	require.Equal(t, string(events.VideoCoreEventType), ws.sent[0].eventType)

	envelope := decodeVideoCoreEnvelope(t, ws.sent[0].payload)
	require.Equal(t, string(ServerEventSetSkipData), envelope["type"])

	sentSkipData, ok := envelope["payload"].(map[string]interface{})
	require.True(t, ok)
	require.NotNil(t, sentSkipData["op"])
	require.Nil(t, sentSkipData["ed"])
}

func TestSetSkipDataKeepsExplicitEmptyOverride(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))
	vc.SetSkipData(&SkipData{})

	// an empty override should stay distinct from clearing so plugins can disable AniSkip fallback.
	require.Len(t, ws.sent, 1)
	envelope := decodeVideoCoreEnvelope(t, ws.sent[0].payload)
	require.Equal(t, string(ServerEventSetSkipData), envelope["type"])
	require.NotNil(t, envelope["payload"])
}

func TestClearSkipDataSendsNilOverride(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))
	vc.SetSkipData(&SkipData{Op: &SkipDataEntry{Interval: SkipInterval{StartTime: 12, EndTime: 42}}})
	ws.sent = nil

	vc.ClearSkipData()

	require.Len(t, ws.sent, 1)

	envelope := decodeVideoCoreEnvelope(t, ws.sent[0].payload)
	require.Equal(t, string(ServerEventSetSkipData), envelope["type"])
	require.Nil(t, envelope["payload"])
}

func TestGetSkipDataReturnsClientOwnedState(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))
	type result struct {
		skipData *SkipData
		ok       bool
	}
	resultCh := make(chan result, 1)

	go func() {
		skipData, ok := vc.GetSkipData()
		resultCh <- result{skipData: skipData, ok: ok}
	}()

	require.Eventually(t, func() bool {
		return len(ws.sent) == 1
	}, time.Second, 10*time.Millisecond)

	envelope := decodeVideoCoreEnvelope(t, ws.sent[0].payload)
	require.Equal(t, string(ServerEventGetSkipData), envelope["type"])
	require.Nil(t, envelope["payload"])

	ws.MockSendVideoCoreEvent(ClientEvent{
		ClientId: "player-client",
		Type:     PlayerEventVideoSkipData,
		Payload: mustMarshalRaw(t, clientVideoSkipDataPayload{SkipData: &SkipData{
			Op: &SkipDataEntry{Interval: SkipInterval{StartTime: 12, EndTime: 42}},
		}}),
	})

	select {
	case ret := <-resultCh:
		require.True(t, ret.ok)
		require.NotNil(t, ret.skipData)
		require.NotNil(t, ret.skipData.Op)
		require.Equal(t, 12.0, ret.skipData.Op.Interval.StartTime)
	case <-time.After(time.Second):
		t.Fatal("expected skip data result")
	}
}

func TestGetSkipDataAllowsEmptyClientState(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))
	type result struct {
		skipData *SkipData
		ok       bool
	}
	resultCh := make(chan result, 1)

	go func() {
		skipData, ok := vc.GetSkipData()
		resultCh <- result{skipData: skipData, ok: ok}
	}()

	require.Eventually(t, func() bool {
		return len(ws.sent) == 1
	}, time.Second, 10*time.Millisecond)

	ws.MockSendVideoCoreEvent(ClientEvent{
		ClientId: "player-client",
		Type:     PlayerEventVideoSkipData,
		Payload:  mustMarshalRaw(t, clientVideoSkipDataPayload{}),
	})

	select {
	case ret := <-resultCh:
		require.True(t, ret.ok)
		require.Nil(t, ret.skipData)
	case <-time.After(time.Second):
		t.Fatal("expected skip data result")
	}
}

func TestPlayerStateRequestsTimeout(t *testing.T) {
	previousTimeout := playerEventResponseTimeout
	playerEventResponseTimeout = 10 * time.Millisecond
	t.Cleanup(func() {
		playerEventResponseTimeout = previousTimeout
	})

	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	vc.setPlaybackState(newPlaybackState("playback-1"))

	tests := []struct {
		name      string
		eventType ServerEvent
		call      func() bool
	}{
		{
			name:      "text tracks",
			eventType: ServerEventGetTextTracks,
			call: func() bool {
				_, ok := vc.GetTextTracks()
				return ok
			},
		},
		{
			name:      "playlist",
			eventType: ServerEventGetPlaylist,
			call: func() bool {
				_, ok := vc.GetPlaylist()
				return ok
			},
		},
		{
			name:      "status",
			eventType: ServerEventGetStatus,
			call: func() bool {
				_, ok := vc.PullStatus()
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws.sent = nil
			resultCh := make(chan bool, 1)

			go func() {
				resultCh <- tt.call()
			}()

			require.Eventually(t, func() bool {
				return len(ws.sent) == 1
			}, time.Second, 10*time.Millisecond)

			// missing client responses should release
			envelope := decodeVideoCoreEnvelope(t, ws.sent[0].payload)
			require.Equal(t, string(tt.eventType), envelope["type"])

			select {
			case ok := <-resultCh:
				require.False(t, ok)
			case <-time.After(time.Second):
				t.Fatal("expected player state request to time out")
			}
		})
	}
}

func TestVideoStatusRecoversAfterZeroDurationLoadedMetadata(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager()
	vc := New(NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})

	t.Cleanup(vc.Shutdown)

	statusEventCh := make(chan *VideoStatusEvent, 1)
	cancel := vc.RegisterEventCallback(func(e VideoEvent) bool {
		statusEvent, ok := e.(*VideoStatusEvent)
		if !ok {
			return true
		}

		statusEventCh <- statusEvent
		return false
	})
	t.Cleanup(cancel)

	state := newPlaybackState("playback-1")
	ws.MockSendVideoCoreEvent(ClientEvent{
		ClientId: "player-client",
		Type:     PlayerEventVideoLoaded,
		Payload:  mustMarshalRaw(t, clientVideoLoadedPayload{State: *state}),
	})

	require.Eventually(t, func() bool {
		playbackState, ok := vc.GetPlaybackState()
		return ok && playbackState.PlaybackInfo != nil && playbackState.PlaybackInfo.Id == "playback-1"
	}, time.Second, 10*time.Millisecond)

	// making sure duration is zero to simulate an edge case
	ws.MockSendVideoCoreEvent(ClientEvent{
		ClientId: "player-client",
		Type:     PlayerEventVideoLoadedMetadata,
		Payload: mustMarshalRaw(t, clientVideoStatusPayload{
			CurrentTime: 24,
			Duration:    0,
			Paused:      false,
		}),
	})

	require.Eventually(t, func() bool {
		vc.playbackStatusMu.RLock()
		defer vc.playbackStatusMu.RUnlock()
		return vc.playbackStatus != nil && vc.playbackStatus.Duration == 0
	}, time.Second, 10*time.Millisecond)

	ws.MockSendVideoCoreEvent(ClientEvent{
		ClientId: "player-client",
		Type:     PlayerEventVideoStatus,
		Payload: mustMarshalRaw(t, clientVideoStatusPayload{
			CurrentTime: 32,
			Duration:    120,
			Paused:      false,
		}),
	})

	require.Eventually(t, func() bool {
		status, ok := vc.GetPlaybackStatus()
		return ok && status.CurrentTime == 32 && status.Duration == 120
	}, time.Second, 10*time.Millisecond)

	select {
	case statusEvent := <-statusEventCh:
		require.Equal(t, 32.0, statusEvent.CurrentTime)
		require.Equal(t, 120.0, statusEvent.Duration)
	case <-time.After(time.Second):
		t.Fatal("expected recovered video status event")
	}
}
