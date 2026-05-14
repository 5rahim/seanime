package directstream

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"seanime/internal/videocore"
	"sync"
	"testing"
	"time"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type testStream struct {
	BaseStream
	handler http.Handler
}

func (s *testStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeTorrent
}

func (s *testStream) GetStreamHandler() http.Handler {
	return s.handler
}

func (s *testStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return s.playbackInfo, s.playbackInfoErr
}

type trackingReadSeekCloser struct {
	closed bool
}

type blockingStream struct {
	clientID       string
	loadPlaybackCh chan struct{}
	loadStartedCh  chan struct{}
	terminatedCh   chan struct{}
	terminated     bool
	startOnce      sync.Once
}

func (s *blockingStream) Type() nativeplayer.StreamType               { return nativeplayer.StreamTypeTorrent }
func (s *blockingStream) LoadContentType() string                     { return "video/webm" }
func (s *blockingStream) ClientId() string                            { return s.clientID }
func (s *blockingStream) Media() *anilist.BaseAnime                   { return nil }
func (s *blockingStream) Episode() *anime.Episode                     { return nil }
func (s *blockingStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *blockingStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *blockingStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	s.startOnce.Do(func() {
		if s.loadStartedCh != nil {
			close(s.loadStartedCh)
		}
	})
	<-s.loadPlaybackCh
	return &nativeplayer.PlaybackInfo{ID: "blocked"}, nil
}
func (s *blockingStream) GetAttachmentByName(string) (*mkvparser.AttachmentInfo, bool) {
	return nil, false
}
func (s *blockingStream) GetStreamHandler() http.Handler { return http.NewServeMux() }
func (s *blockingStream) StreamError(error)              {}
func (s *blockingStream) Terminate() {
	if s.terminated {
		return
	}
	s.terminated = true
	close(s.terminatedCh)
}
func (s *blockingStream) GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent] {
	return result.NewMap[string, *mkvparser.SubtitleEvent]()
}
func (s *blockingStream) OnSubtitleFileUploaded(string, string) {}

type prevTerminateStream struct {
	manager       *Manager
	clientID      string
	terminatedCh  chan struct{}
	terminateOnce sync.Once
}

func (s *prevTerminateStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeTorrent
}
func (s *prevTerminateStream) LoadContentType() string                     { return "video/webm" }
func (s *prevTerminateStream) ClientId() string                            { return s.clientID }
func (s *prevTerminateStream) Media() *anilist.BaseAnime                   { return nil }
func (s *prevTerminateStream) Episode() *anime.Episode                     { return nil }
func (s *prevTerminateStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *prevTerminateStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *prevTerminateStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return &nativeplayer.PlaybackInfo{ID: "previous-playback"}, nil
}
func (s *prevTerminateStream) GetAttachmentByName(string) (*mkvparser.AttachmentInfo, bool) {
	return nil, false
}
func (s *prevTerminateStream) GetStreamHandler() http.Handler { return http.NewServeMux() }
func (s *prevTerminateStream) StreamError(error)              {}
func (s *prevTerminateStream) Terminate() {
	s.terminateOnce.Do(func() {
		close(s.terminatedCh)
		_ = s.manager.CloseOpen("")
	})
}
func (s *prevTerminateStream) GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent] {
	return result.NewMap[string, *mkvparser.SubtitleEvent]()
}
func (s *prevTerminateStream) OnSubtitleFileUploaded(string, string) {}

type eventStream struct {
	clientID      string
	playbackInfo  *nativeplayer.PlaybackInfo
	terminatedCh  chan struct{}
	terminateOnce sync.Once
}

func (s *eventStream) Type() nativeplayer.StreamType               { return nativeplayer.StreamTypeTorrent }
func (s *eventStream) LoadContentType() string                     { return "video/webm" }
func (s *eventStream) ClientId() string                            { return s.clientID }
func (s *eventStream) Media() *anilist.BaseAnime                   { return nil }
func (s *eventStream) Episode() *anime.Episode                     { return nil }
func (s *eventStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *eventStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *eventStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return s.playbackInfo, nil
}
func (s *eventStream) GetAttachmentByName(string) (*mkvparser.AttachmentInfo, bool) {
	return nil, false
}
func (s *eventStream) GetStreamHandler() http.Handler { return http.NewServeMux() }
func (s *eventStream) StreamError(error)              {}
func (s *eventStream) Terminate() {
	s.terminateOnce.Do(func() {
		close(s.terminatedCh)
	})
}
func (s *eventStream) GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent] {
	return result.NewMap[string, *mkvparser.SubtitleEvent]()
}
func (s *eventStream) OnSubtitleFileUploaded(string, string) {}

func mustMarshalRawMessage(t *testing.T, value interface{}) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return data
}

func (r *trackingReadSeekCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (r *trackingReadSeekCloser) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

func (r *trackingReadSeekCloser) Close() error {
	r.closed = true
	return nil
}

func TestGetStreamHandlerRejectsMismatchedPlaybackID(t *testing.T) {
	called := false
	stream := &testStream{
		BaseStream: BaseStream{
			clientId: "client-1",
			playbackInfo: &nativeplayer.PlaybackInfo{
				ID: "expected-playback-id",
			},
		},
		handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		}),
	}

	manager := &Manager{
		currentStream: mo.Some[Stream](stream),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/directstream/stream?id=stale-playback-id", nil)
	rec := httptest.NewRecorder()

	manager.getStreamHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.False(t, called)
}

func TestGetStreamHandlerForwardsMatchingPlaybackID(t *testing.T) {
	called := false
	stream := &testStream{
		BaseStream: BaseStream{
			clientId: "client-1",
			playbackInfo: &nativeplayer.PlaybackInfo{
				ID: "playback-id",
			},
		},
		handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		}),
	}

	manager := &Manager{
		currentStream: mo.Some[Stream](stream),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/directstream/stream?id=playback-id", nil)
	rec := httptest.NewRecorder()

	manager.getStreamHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.True(t, called)
}

func TestStartSubtitleStreamPClosesReaderWhenParserMissing(t *testing.T) {
	reader := &trackingReadSeekCloser{}
	stream := &BaseStream{
		logger: util.NewLogger(),
		playbackInfo: &nativeplayer.PlaybackInfo{
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
		},
		activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
	}

	stream.StartSubtitleStreamP(stream, context.Background(), reader, 0, 1024)

	require.True(t, reader.closed)
}

func TestListenToPlayerEventsTerminatesWithoutWaitingForPlaybackInfo(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := videocore.New(videocore.NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	np := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      vc,
	})
	manager := NewManager(NewManagerOptions{
		Logger:         logger,
		WSEventManager: ws,
		NativePlayer:   np,
		VideoCore:      vc,
	})

	stream := &blockingStream{
		clientID:       "player-client",
		loadPlaybackCh: make(chan struct{}),
		loadStartedCh:  make(chan struct{}),
		terminatedCh:   make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)

	t.Cleanup(func() {
		close(stream.loadPlaybackCh)
		vc.Shutdown()
	})

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: videocore.ClientEvent{
			ClientId: "player-client",
			Type:     videocore.PlayerEventVideoTerminated,
		},
	})

	select {
	case <-stream.terminatedCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected terminate to bypass playback info loading")
	}
}

// ensures that if a new stream is started while the previous stream is still loading playback info,
// the previous stream will be terminated without waiting for the playback info to finish loading
func TestStream_beginOpenTerminatesPreviousStream(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := videocore.New(videocore.NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	np := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      vc,
	})
	manager := NewManager(NewManagerOptions{
		Logger:         logger,
		WSEventManager: ws,
		NativePlayer:   np,
		VideoCore:      vc,
	})

	stream := &prevTerminateStream{
		manager:      manager,
		clientID:     "player-client",
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)

	t.Cleanup(func() {
		vc.Shutdown()
	})

	done := make(chan bool, 1)
	go func() {
		done <- manager.BeginOpen("player-client", "opening", nil)
	}()

	select {
	case <-stream.terminatedCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected previous stream to terminate")
	}

	select {
	case ok := <-done:
		require.True(t, ok)
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected BeginOpen to return without deadlocking")
	}
}

// the old player's terminate event can arrive after the new stream starts opening
func TestStream_beginOpenIgnoresReplacedPlaybackTermination(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := videocore.New(videocore.NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	np := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      vc,
	})
	manager := NewManager(NewManagerOptions{
		Logger:         logger,
		WSEventManager: ws,
		NativePlayer:   np,
		VideoCore:      vc,
	})

	stream := &eventStream{
		clientID:     "player-client",
		playbackInfo: &nativeplayer.PlaybackInfo{ID: "previous-playback-id"},
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)
	manager.currentPlaybackId = "previous-playback-id"
	manager.currentPlaybackClient = "player-client"

	t.Cleanup(func() {
		vc.Shutdown()
	})

	require.True(t, manager.BeginOpen("player-client", "opening", nil))

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: videocore.ClientEvent{
			ClientId: "player-client",
			Type:     videocore.PlayerEventVideoTerminated,
			Payload: mustMarshalRawMessage(t, map[string]interface{}{
				"id":           "previous-playback-id",
				"clientId":     "player-client",
				"playerType":   "native",
				"playbackType": "torrent",
			}),
		},
	})

	require.Eventually(t, func() bool {
		return manager.IsOpenActive("player-client")
	}, time.Second, 10*time.Millisecond)

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: videocore.ClientEvent{
			ClientId: "player-client",
			Type:     videocore.PlayerEventVideoTerminated,
			Payload: mustMarshalRawMessage(t, map[string]interface{}{
				"clientId":     "player-client",
				"playerType":   "native",
				"playbackType": "torrent",
			}),
		},
	})

	require.Eventually(t, func() bool {
		return !manager.IsOpenActive("player-client")
	}, time.Second, 10*time.Millisecond)
}

func TestStream_closeOpenReturnsWhileLoadPlaybackInfoIsBlocked(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := videocore.New(videocore.NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	np := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      vc,
	})
	manager := NewManager(NewManagerOptions{
		Logger:         logger,
		WSEventManager: ws,
		NativePlayer:   np,
		VideoCore:      vc,
	})

	stream := &blockingStream{
		clientID:       "player-client",
		loadPlaybackCh: make(chan struct{}),
		loadStartedCh:  make(chan struct{}),
		terminatedCh:   make(chan struct{}),
	}

	t.Cleanup(func() {
		close(stream.loadPlaybackCh)
		vc.Shutdown()
	})

	require.True(t, manager.BeginOpen("player-client", "opening", nil))

	go manager.loadStream(stream)

	select {
	case <-stream.loadStartedCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected stream loading to reach playback info")
	}

	done := make(chan bool, 1)
	go func() {
		done <- manager.CloseOpen("player-client")
	}()

	select {
	case ok := <-done:
		require.True(t, ok)
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected CloseOpen to return while metadata is still loading")
	}
}

func TestStream_listenToPlayerEventsIgnoresStalePlaybackTermination(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	vc := videocore.New(videocore.NewVideoCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	np := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      vc,
	})
	manager := NewManager(NewManagerOptions{
		Logger:         logger,
		WSEventManager: ws,
		NativePlayer:   np,
		VideoCore:      vc,
	})

	stream := &eventStream{
		clientID:     "player-client",
		playbackInfo: &nativeplayer.PlaybackInfo{ID: "current-playback-id"},
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)
	manager.currentPlaybackId = "current-playback-id"
	manager.currentPlaybackClient = "player-client"

	t.Cleanup(func() {
		vc.Shutdown()
	})

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: videocore.ClientEvent{
			ClientId: "player-client",
			Type:     videocore.PlayerEventVideoTerminated,
			Payload: mustMarshalRawMessage(t, map[string]interface{}{
				"id":           "stale-playback-id",
				"clientId":     "player-client",
				"playerType":   "native",
				"playbackType": "torrent",
			}),
		},
	})

	select {
	case <-stream.terminatedCh:
		t.Fatal("expected stale terminated event to be ignored")
	case <-time.After(250 * time.Millisecond):
	}

	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "socket-client",
		Type:     events.VideoCoreEventType,
		Payload: videocore.ClientEvent{
			ClientId: "player-client",
			Type:     videocore.PlayerEventVideoTerminated,
			Payload: mustMarshalRawMessage(t, map[string]interface{}{
				"id":           "current-playback-id",
				"clientId":     "player-client",
				"playerType":   "native",
				"playbackType": "torrent",
			}),
		},
	})

	select {
	case <-stream.terminatedCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected matching terminated event to stop the stream")
	}
}
