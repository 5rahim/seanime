package directstream

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/mpvcore"
	"seanime/internal/nativeplayer"
	"seanime/internal/player"
	"seanime/internal/util"
	httputil "seanime/internal/util/http"
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

type blockingRangeReader struct {
	closed chan struct{}
	once   sync.Once
}

func newBlockingRangeReader() *blockingRangeReader {
	return &blockingRangeReader{closed: make(chan struct{})}
}

func (r *blockingRangeReader) Read([]byte) (int, error) {
	<-r.closed
	return 0, io.ErrClosedPipe
}

func (r *blockingRangeReader) Seek(int64, int) (int64, error) {
	return 0, nil
}

func (r *blockingRangeReader) Close() error {
	r.once.Do(func() {
		close(r.closed)
	})
	return nil
}

func TestServeContentRange(t *testing.T) {
	reader := newBlockingRangeReader()
	ctx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	done := make(chan struct{})

	go func() {
		defer close(done)
		serveContentRange(rec, req, ctx, reader, "video.mkv", 1024, "video/webm", httputil.Range{Start: 0, Length: 512})
	}()

	// canceled stream contexts should close blocked readers and free the http request
	cancel()

	select {
	case <-reader.closed:
	case <-time.After(time.Second):
		t.Fatal("expected reader to close")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected range serve to return")
	}
}

func TestDirectStreamPlaybackTargetsRemainIndependent(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	videoCore := videocore.New(videocore.NewVideoCoreOptions{WsEventManager: ws, Logger: logger})
	nativePlayer := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      videoCore,
	})
	mpvCore := mpvcore.New(mpvcore.NewMpvCoreOptions{WsEventManager: ws, Logger: logger})

	vcAdapter := videocore.NewAdapter(videoCore, nativePlayer)
	mcAdapter := mpvcore.NewAdapter(mpvCore)
	coordinator := mediacore.NewCoordinator(mediacore.NewCoordinatorOptions{
		Logger:       logger,
		IsOfflineRef: util.NewRef(false),
		Backends: map[player.Target]mediacore.Backend{
			player.TargetVideoCore: vcAdapter,
			player.TargetMpvCore:   mcAdapter,
		},
	})

	manager := NewManager(NewManagerOptions{
		Logger:               logger,
		WSEventManager:       ws,
		NativePlayer:         nativePlayer,
		VideoCore:            videoCore,
		MediacoreCoordinator: coordinator,
	})
	t.Cleanup(videoCore.Shutdown)
	t.Cleanup(mpvCore.Shutdown)
	t.Cleanup(func() {
		_ = coordinator.Close()
	})

	manager.SetPlaybackTarget(PlaybackTargetVideoCore)
	require.True(t, manager.BeginOpen("client", "video-core", nil))
	require.Equal(t, string(events.NativePlayerEventType), ws.Events()[len(ws.Events())-1].Type)
	require.True(t, manager.CloseOpen("client"))

	manager.SetPlaybackTarget(PlaybackTargetMpvCore)
	require.True(t, manager.BeginOpen("client", "mpv-core", nil))
	require.Equal(t, string(events.MpvCoreEventType), ws.Events()[len(ws.Events())-1].Type)
}

func (s *testStream) Type() player.PlaybackType {
	return player.PlaybackTypeTorrent
}

func (s *testStream) GetStreamHandler() http.Handler {
	return s.handler
}

func (s *testStream) LoadPlaybackInfo() (*player.PlaybackInfo, error) {
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

func (s *blockingStream) Type() player.PlaybackType                   { return player.PlaybackTypeTorrent }
func (s *blockingStream) LoadContentType() string                     { return "video/webm" }
func (s *blockingStream) ClientId() string                            { return s.clientID }
func (s *blockingStream) Media() *anilist.BaseAnime                   { return nil }
func (s *blockingStream) Episode() *anime.Episode                     { return nil }
func (s *blockingStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *blockingStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *blockingStream) LoadPlaybackInfo() (*player.PlaybackInfo, error) {
	s.startOnce.Do(func() {
		if s.loadStartedCh != nil {
			close(s.loadStartedCh)
		}
	})
	<-s.loadPlaybackCh
	return &player.PlaybackInfo{ID: "blocked", PlaybackType: player.PlaybackTypeTorrent}, nil
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

func (s *prevTerminateStream) Type() player.PlaybackType {
	return player.PlaybackTypeTorrent
}
func (s *prevTerminateStream) LoadContentType() string                     { return "video/webm" }
func (s *prevTerminateStream) ClientId() string                            { return s.clientID }
func (s *prevTerminateStream) Media() *anilist.BaseAnime                   { return nil }
func (s *prevTerminateStream) Episode() *anime.Episode                     { return nil }
func (s *prevTerminateStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *prevTerminateStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *prevTerminateStream) LoadPlaybackInfo() (*player.PlaybackInfo, error) {
	return &player.PlaybackInfo{ID: "previous-playback", PlaybackType: player.PlaybackTypeTorrent}, nil
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
	playbackInfo  *player.PlaybackInfo
	terminatedCh  chan struct{}
	terminateOnce sync.Once
}

func (s *eventStream) Type() player.PlaybackType                   { return player.PlaybackTypeTorrent }
func (s *eventStream) LoadContentType() string                     { return "video/webm" }
func (s *eventStream) ClientId() string                            { return s.clientID }
func (s *eventStream) Media() *anilist.BaseAnime                   { return nil }
func (s *eventStream) Episode() *anime.Episode                     { return nil }
func (s *eventStream) ListEntryData() *anime.EntryListData         { return nil }
func (s *eventStream) EpisodeCollection() *anime.EpisodeCollection { return nil }
func (s *eventStream) LoadPlaybackInfo() (*player.PlaybackInfo, error) {
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

func newDirectstreamMpvTestManager(t *testing.T) (*Manager, *events.MockWSEventManager, *mediacore.Coordinator) {
	t.Helper()
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	mpvCore := mpvcore.New(mpvcore.NewMpvCoreOptions{
		WsEventManager: ws,
		Logger:         logger,
	})
	mcAdapter := mpvcore.NewAdapter(mpvCore)
	coordinator := mediacore.NewCoordinator(mediacore.NewCoordinatorOptions{
		Logger:       logger,
		IsOfflineRef: util.NewRef(false),
		Backends: map[player.Target]mediacore.Backend{
			player.TargetMpvCore: mcAdapter,
		},
	})
	manager := NewManager(NewManagerOptions{
		Logger:               logger,
		WSEventManager:       ws,
		MediacoreCoordinator: coordinator,
	})
	t.Cleanup(func() {
		_ = coordinator.Close()
		mpvCore.Shutdown()
	})
	return manager, ws, coordinator
}

func activateMpvPlayback(t *testing.T, ws *events.MockWSEventManager, core *mediacore.Coordinator, clientID, playbackID string) {
	t.Helper()
	core.Watch(player.TargetMpvCore, clientID, &player.PlaybackInfo{ID: playbackID, PlaybackType: player.PlaybackTypeTorrent})
	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: clientID,
		Type:     events.MpvCoreEventType,
		Payload: map[string]interface{}{
			"clientId": clientID,
			"type":     mpvcore.ClientEventPlaybackLoaded,
			"payload": map[string]interface{}{
				"id":       playbackID,
				"clientId": clientID,
			},
		},
	})
	require.Eventually(t, func() bool {
		state, ok := core.GetActivePlaybackState()
		return ok && state.PlaybackInfo.ID == playbackID
	}, time.Second, 10*time.Millisecond)
}

func sendMpvTerminated(ws *events.MockWSEventManager, clientID, playbackID string) {
	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: clientID,
		Type:     events.MpvCoreEventType,
		Payload: map[string]interface{}{
			"clientId": clientID,
			"type":     mpvcore.ClientEventTerminated,
			"payload": map[string]interface{}{
				"id":           playbackID,
				"clientId":     clientID,
				"playbackType": mpvcore.PlaybackTypeTorrent,
			},
		},
	})
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
			playbackInfo: &player.PlaybackInfo{
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
			playbackInfo: &player.PlaybackInfo{
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
		playbackInfo: &player.PlaybackInfo{
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
		},
		activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
	}

	stream.StartSubtitleStreamP(stream, context.Background(), reader, 0, 1024)

	require.True(t, reader.closed)
}

func TestListenToPlayerEventsTerminatesWithoutWaitingForPlaybackInfo(t *testing.T) {
	manager, ws, core := newDirectstreamMpvTestManager(t)

	stream := &blockingStream{
		clientID:       "player-client",
		loadPlaybackCh: make(chan struct{}),
		loadStartedCh:  make(chan struct{}),
		terminatedCh:   make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)
	manager.currentPlaybackId = "blocked"
	manager.currentPlaybackClient = "player-client"
	core.Watch(player.TargetMpvCore, "player-client", &player.PlaybackInfo{ID: "blocked", PlaybackType: player.PlaybackTypeTorrent})

	t.Cleanup(func() {
		close(stream.loadPlaybackCh)
	})

	sendMpvTerminated(ws, "player-client", "blocked")

	select {
	case <-stream.terminatedCh:
	case <-time.After(time.Second):
		t.Fatal("expected terminate to bypass playback info loading")
	}
}

// ensures that if a new stream is started while the previous stream is still loading playback info,
// the previous stream will be terminated without waiting for the playback info to finish loading
func TestStream_beginOpenTerminatesPreviousStream(t *testing.T) {
	manager, _, _ := newDirectstreamMpvTestManager(t)

	stream := &prevTerminateStream{
		manager:      manager,
		clientID:     "player-client",
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)

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
	manager, ws, core := newDirectstreamMpvTestManager(t)

	stream := &eventStream{
		clientID:     "player-client",
		playbackInfo: &player.PlaybackInfo{ID: "previous-playback-id", PlaybackType: player.PlaybackTypeTorrent},
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)
	manager.currentPlaybackId = "previous-playback-id"
	manager.currentPlaybackClient = "player-client"

	activateMpvPlayback(t, ws, core, "player-client", "previous-playback-id")

	require.True(t, manager.BeginOpen("player-client", "opening", nil))

	sendMpvTerminated(ws, "player-client", "previous-playback-id")

	time.Sleep(100 * time.Millisecond)

	require.True(t, manager.IsOpenActive("player-client"))

	require.True(t, manager.CloseOpen("player-client"))
}

func TestStream_closeOpenReturnsWhileLoadPlaybackInfoIsBlocked(t *testing.T) {
	manager, _, _ := newDirectstreamMpvTestManager(t)

	stream := &blockingStream{
		clientID:       "player-client",
		loadPlaybackCh: make(chan struct{}),
		loadStartedCh:  make(chan struct{}),
		terminatedCh:   make(chan struct{}),
	}

	t.Cleanup(func() {
		close(stream.loadPlaybackCh)
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
	manager, ws, core := newDirectstreamMpvTestManager(t)

	stream := &eventStream{
		clientID:     "player-client",
		playbackInfo: &player.PlaybackInfo{ID: "current-playback-id", PlaybackType: player.PlaybackTypeTorrent},
		terminatedCh: make(chan struct{}),
	}
	manager.currentStream = mo.Some[Stream](stream)
	manager.currentPlaybackId = "current-playback-id"
	manager.currentPlaybackClient = "player-client"

	activateMpvPlayback(t, ws, core, "player-client", "current-playback-id")
	sendMpvTerminated(ws, "player-client", "stale-playback-id")

	select {
	case <-stream.terminatedCh:
		t.Fatal("expected stale terminated event to be ignored")
	case <-time.After(250 * time.Millisecond):
	}

	sendMpvTerminated(ws, "player-client", "current-playback-id")

	select {
	case <-stream.terminatedCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected matching terminated event to stop the stream")
	}
}
