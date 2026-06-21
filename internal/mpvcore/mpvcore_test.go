package mpvcore

import (
	"encoding/json"
	"seanime/internal/events"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type directedMpvEvent struct {
	clientID string
	event    ServerEvent
	payload  interface{}
}

type mpvCoreTestWS struct {
	*events.MockWSEventManager
	mu        sync.Mutex
	directed  []directedMpvEvent
	mpvClient *events.ClientEventSubscriber
}

func newMpvCoreTestWS() *mpvCoreTestWS {
	return &mpvCoreTestWS{MockWSEventManager: events.NewMockWSEventManager(util.NewLogger())}
}

func (w *mpvCoreTestWS) SubscribeToClientMpvCoreEvents(_ string) *events.ClientEventSubscriber {
	w.mpvClient = &events.ClientEventSubscriber{Channel: make(chan *events.WebsocketClientEvent, 32)}
	return w.mpvClient
}

func (w *mpvCoreTestWS) SendEventTo(clientID string, eventType string, payload interface{}, noLog ...bool) {
	w.record(clientID, eventType, payload)
	w.MockWSEventManager.SendEventTo(clientID, eventType, payload, noLog...)
}

func (w *mpvCoreTestWS) SendEvent(eventType string, payload interface{}) {
	w.record("", eventType, payload)
	w.MockWSEventManager.SendEvent(eventType, payload)
}

func (w *mpvCoreTestWS) record(clientID, eventType string, payload interface{}) {
	if eventType != string(events.MpvCoreEventType) {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	var envelope struct {
		Type    ServerEvent `json:"type"`
		Payload interface{} `json:"payload"`
	}
	if json.Unmarshal(data, &envelope) != nil {
		return
	}
	w.mu.Lock()
	w.directed = append(w.directed, directedMpvEvent{clientID: clientID, event: envelope.Type, payload: envelope.Payload})
	w.mu.Unlock()
}

func (w *mpvCoreTestWS) send(clientID string, eventType ClientEventType, payload interface{}) {
	w.mpvClient.Channel <- &events.WebsocketClientEvent{
		ClientID: clientID,
		Type:     events.MpvCoreEventType,
		Payload: map[string]interface{}{
			"clientId": clientID,
			"type":     eventType,
			"payload":  payload,
		},
	}
}

func (w *mpvCoreTestWS) hasDirected(clientID string, event ServerEvent) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, value := range w.directed {
		if value.clientID == clientID && value.event == event {
			return true
		}
	}
	return false
}

func newMpvCoreForTest(t *testing.T) (*MpvCore, *mpvCoreTestWS) {
	t.Helper()
	ws := newMpvCoreTestWS()
	core := New(NewMpvCoreOptions{WsEventManager: ws, Logger: util.NewLogger()})
	t.Cleanup(core.Shutdown)
	return core, ws
}

func activatePlayback(t *testing.T, core *MpvCore, ws *mpvCoreTestWS, clientID, playbackID string) {
	t.Helper()
	core.Watch(clientID, &PlaybackInfo{ID: playbackID, PlaybackType: PlaybackTypeTorrent})
	ws.send(clientID, ClientEventPlaybackLoaded, map[string]interface{}{"id": playbackID, "clientId": clientID})
	require.Eventually(t, func() bool {
		state, ok := core.GetPlaybackState()
		return ok && state.ClientID == clientID && state.PlaybackInfo.ID == playbackID
	}, time.Second, 10*time.Millisecond)
}

func TestMpvCorePendingPromotionAndMetadata(t *testing.T) {
	core, ws := newMpvCoreForTest(t)
	core.Watch("client-a", &PlaybackInfo{ID: "playback-a", PlaybackType: PlaybackTypeTorrent})

	_, active := core.GetPlaybackState()
	require.False(t, active)

	ws.send("client-a", ClientEventStatus, map[string]interface{}{
		"id": "playback-a", "currentTime": 10.0, "duration": 100.0, "paused": false,
	})
	time.Sleep(20 * time.Millisecond)
	_, hasStatus := core.GetPlaybackStatus()
	require.False(t, hasStatus)

	ws.send("client-a", ClientEventPlaybackLoaded, map[string]interface{}{"id": "playback-a", "clientId": "client-a"})
	require.Eventually(t, func() bool {
		_, ok := core.GetPlaybackState()
		return ok
	}, time.Second, 10*time.Millisecond)

	ws.send("client-a", ClientEventLoadedMetadata, map[string]interface{}{
		"id": "playback-a", "clientId": "client-a", "currentTime": 12.0, "duration": 120.0, "paused": true,
	})
	require.Eventually(t, func() bool {
		status, ok := core.GetPlaybackStatus()
		return ok && status.CurrentTime == 12 && status.Duration == 120 && status.Paused
	}, time.Second, 10*time.Millisecond)
}

func TestMpvCoreClientAndPlaybackIsolation(t *testing.T) {
	core, ws := newMpvCoreForTest(t)
	core.Watch("client-a", &PlaybackInfo{ID: "playback-a", PlaybackType: PlaybackTypeDebrid})

	ws.send("client-b", ClientEventPlaybackLoaded, map[string]interface{}{"id": "playback-a", "clientId": "client-b"})
	time.Sleep(20 * time.Millisecond)
	_, ok := core.GetPlaybackState()
	require.False(t, ok)

	activatePlayback(t, core, ws, "client-a", "playback-a")
	ws.send("client-b", ClientEventLoadedMetadata, map[string]interface{}{
		"id": "playback-a", "clientId": "client-b", "currentTime": 25.0, "duration": 100.0,
	})
	ws.send("client-a", ClientEventLoadedMetadata, map[string]interface{}{
		"id": "wrong-playback", "clientId": "client-a", "currentTime": 25.0, "duration": 100.0,
	})
	time.Sleep(20 * time.Millisecond)
	_, hasStatus := core.GetPlaybackStatus()
	require.False(t, hasStatus)
}

func TestMpvCoreStatusLifecycleAndPullStatus(t *testing.T) {
	core, ws := newMpvCoreForTest(t)
	activatePlayback(t, core, ws, "client-a", "playback-a")
	ws.send("client-a", ClientEventLoadedMetadata, map[string]interface{}{
		"id": "playback-a", "clientId": "client-a", "currentTime": 0.0, "duration": 200.0, "paused": false,
	})

	resultCh := make(chan StatusEvent, 1)
	go func() {
		status, ok := core.PullStatus()
		if ok {
			resultCh <- status
		}
	}()
	require.Eventually(t, func() bool {
		return ws.hasDirected("client-a", ServerEventGetStatus)
	}, time.Second, 10*time.Millisecond)
	ws.send("client-a", ClientEventStatus, map[string]interface{}{
		"id": "playback-a", "clientId": "client-a", "currentTime": 33.0, "duration": 200.0, "paused": true,
	})

	select {
	case status := <-resultCh:
		require.Equal(t, 33.0, status.CurrentTime)
		require.True(t, status.Paused)
	case <-time.After(time.Second):
		t.Fatal("PullStatus did not receive the renderer response")
	}
}

func TestMpvCoreTerminationWorksForPendingAndActivePlayback(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		core, ws := newMpvCoreForTest(t)
		sub := core.Subscribe("termination-test")
		core.Watch("client-a", &PlaybackInfo{ID: "pending", PlaybackType: PlaybackTypeNakama})
		ws.send("client-a", ClientEventTerminated, map[string]interface{}{
			"id": "pending", "clientId": "client-a", "playbackType": PlaybackTypeNakama,
		})
		select {
		case event := <-sub.Events():
			terminated, ok := event.(*TerminatedEvent)
			require.True(t, ok)
			require.Equal(t, "pending", terminated.GetPlaybackID())
		case <-time.After(time.Second):
			t.Fatal("pending termination was not dispatched")
		}
	})

	t.Run("active", func(t *testing.T) {
		core, ws := newMpvCoreForTest(t)
		sub := core.Subscribe("termination-test")
		activatePlayback(t, core, ws, "client-a", "active")
		<-sub.Events() // PlaybackLoadedEvent
		ws.send("client-a", ClientEventTerminated, map[string]interface{}{
			"id": "active", "clientId": "client-a", "playbackType": PlaybackTypeTorrent,
		})
		select {
		case event := <-sub.Events():
			_, ok := event.(*TerminatedEvent)
			require.True(t, ok)
		case <-time.After(time.Second):
			t.Fatal("active termination was not dispatched")
		}
		require.Eventually(t, func() bool {
			_, ok := core.GetPlaybackState()
			return !ok
		}, time.Second, 10*time.Millisecond)
	})
}

func TestMpvCoreServerControlsTargetOnlyActiveClient(t *testing.T) {
	core, ws := newMpvCoreForTest(t)
	activatePlayback(t, core, ws, "client-a", "playback-a")

	core.Pause()
	core.SeekTo(42)
	core.SetAudioTrack(2)

	require.True(t, ws.hasDirected("client-a", ServerEventPause))
	require.True(t, ws.hasDirected("client-a", ServerEventSeekTo))
	require.True(t, ws.hasDirected("client-a", ServerEventSetAudioTrack))
	require.False(t, ws.hasDirected("client-b", ServerEventPause))
}

func TestMpvCoreCriticalEventDeliveryUnderPressure(t *testing.T) {
	core, ws := newMpvCoreForTest(t)
	sub := core.Subscribe("slow-subscriber")
	activatePlayback(t, core, ws, "client-a", "playback-a")
	<-sub.Events() // PlaybackLoadedEvent

	for i := 0; i < cap(sub.eventCh); i++ {
		sub.eventCh <- &StatusEvent{}
	}
	go func() {
		time.Sleep(40 * time.Millisecond)
		<-sub.eventCh
	}()
	ws.send("client-a", ClientEventTerminated, map[string]interface{}{
		"id": "playback-a", "clientId": "client-a", "playbackType": PlaybackTypeTorrent,
	})

	require.Eventually(t, func() bool {
		for len(sub.eventCh) > 0 {
			if _, ok := (<-sub.eventCh).(*TerminatedEvent); ok {
				return true
			}
		}
		return false
	}, time.Second, 20*time.Millisecond)
}

func TestMpvCoreUsesIndependentClientSubscription(t *testing.T) {
	_, ws := newMpvCoreForTest(t)
	require.NotNil(t, ws.mpvClient)
	require.Empty(t, ws.ClientEventSubscribers.Keys())
}
