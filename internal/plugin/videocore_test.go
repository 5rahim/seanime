package plugin

import (
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/mediacore"
	"seanime/internal/player"
	gojautil "seanime/internal/util/goja"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMediacoreBackend struct {
	target      player.Target
	eventsCh    chan player.Event
	executedCmd []player.Command
}

var _ mediacore.Backend = (*mockMediacoreBackend)(nil)

func newMockMediacoreBackend(target player.Target) *mockMediacoreBackend {
	return &mockMediacoreBackend{
		target:   target,
		eventsCh: make(chan player.Event, 100),
	}
}

func (m *mockMediacoreBackend) Target() player.Target                            { return m.target }
func (m *mockMediacoreBackend) OpenAndAwait(clientID, state string)              {}
func (m *mockMediacoreBackend) AbortOpen(clientID, reason string)                {}
func (m *mockMediacoreBackend) Watch(clientID string, info *player.PlaybackInfo) {}
func (m *mockMediacoreBackend) Error(clientID string, err error)                 {}
func (m *mockMediacoreBackend) Terminate(session player.SessionKey)              {}
func (m *mockMediacoreBackend) Events() <-chan player.Event                      { return m.eventsCh }
func (m *mockMediacoreBackend) Close() error                                     { return nil }
func (m *mockMediacoreBackend) PullStatus() (player.PlaybackStatus, bool) {
	return player.PlaybackStatus{
		ID:          "playback-1",
		ClientID:    "client-1",
		Paused:      false,
		CurrentTime: 42.5,
		Duration:    100.0,
	}, true
}
func (m *mockMediacoreBackend) GetPlaylist() (*player.PlaylistState, bool) {
	return &player.PlaylistState{
		Type:           player.PlaybackTypeLocalFile,
		CurrentEpisode: &anime.Episode{EpisodeNumber: 3},
	}, true
}
func (m *mockMediacoreBackend) GetSkipData() (*player.SkipData, bool) {
	return &player.SkipData{
		Op: &player.SkipDataEntry{Interval: player.SkipInterval{StartTime: 10, EndTime: 30}},
	}, true
}
func (m *mockMediacoreBackend) Execute(session player.SessionKey, cmd player.Command) error {
	m.executedCmd = append(m.executedCmd, cmd)
	return nil
}

func newVideoCoreTestContext(t *testing.T) (*AppContextImpl, *mediacore.Coordinator, *mockMediacoreBackend) {
	mb := newMockMediacoreBackend(player.TargetVideoCore)
	backends := map[player.Target]mediacore.Backend{
		player.TargetVideoCore: mb,
	}
	coordinator := mediacore.NewCoordinator(mediacore.NewCoordinatorOptions{
		Backends: backends,
	})
	t.Cleanup(func() {
		coordinator.Close()
	})

	appCtx := NewAppContext().(*AppContextImpl)
	appCtx.SetModulesPartial(AppContextModules{
		MediacoreCoordinator: coordinator,
	})

	return appCtx, coordinator, mb
}

func bindTestVideoCore(t *testing.T, appCtx *AppContextImpl) (*goja.Runtime, *goja.Object, *gojautil.Scheduler) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	obj := vm.NewObject()
	scheduler := gojautil.NewScheduler()
	ext := &extension.Extension{ID: "test-ext", Name: "Test"}

	appCtx.BindVideoCoreToContextObj(vm, obj, nil, ext, scheduler)
	t.Cleanup(scheduler.Stop)

	return vm, obj, scheduler
}

func runJSScriptAsync(t *testing.T, vm *goja.Runtime, scheduler *gojautil.Scheduler, script string) map[string]interface{} {
	t.Helper()
	resultChan := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	// set up the reportResults callback
	err := scheduler.Schedule(func() error {
		return vm.Set("reportResults", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				resultChan <- nil
				return goja.Undefined()
			}
			var res map[string]interface{}
			err := vm.ExportTo(call.Arguments[0], &res)
			if err != nil {
				errChan <- err
			} else {
				resultChan <- res
			}
			return goja.Undefined()
		})
	})
	require.NoError(t, err)

	// run script inside scheduler asynchronously
	scheduler.ScheduleAsync(func() error {
		_, err := vm.RunString(script)
		if err != nil {
			errChan <- err
		}
		return nil
	})

	// wait for callback or error
	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		t.Fatalf("JS runtime execution error: %v", err)
		return nil
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for JS execution results")
		return nil
	}
}

func assertNumberEqual(t *testing.T, expected float64, actual interface{}) {
	t.Helper()
	var val float64
	switch v := actual.(type) {
	case int:
		val = float64(v)
	case int32:
		val = float64(v)
	case int64:
		val = float64(v)
	case float32:
		val = float64(v)
	case float64:
		val = float64(v)
	default:
		t.Fatalf("expected a number type, got %T", actual)
	}
	assert.Equal(t, expected, val)
}

func TestVideoCoreAllMethodsRegression(t *testing.T) {
	appCtx, coordinator, mb := newVideoCoreTestContext(t)
	vm, obj, scheduler := bindTestVideoCore(t, appCtx)

	// make videoCore globally accessible in script
	err := vm.Set("videoCore", obj.Get("videoCore"))
	require.NoError(t, err)

	// prepare script that executes every method on videoCore
	testScript := `
		(async function() {
			const vc = videoCore;
			let results = {};

			// 1. event listeners
			try {
				vc.addEventListener("video-loaded", () => {});
				vc.removeEventListener("video-loaded");
				results.events = "ok";
			} catch(e) {
				results.events = e.message;
			}

			// 2. playback control (inactive or loading/active)
			try {
				vc.pause();
				vc.resume();
				vc.seek(10.5);
				vc.seekTo(42);
				vc.terminate();
				vc.playEpisodeFromPlaylist("next");
				results.playback = "ok";
			} catch(e) {
				results.playback = e.message;
			}

			// 3. ui control
			try {
				vc.setFullscreen(true);
				vc.setPip(false);
				vc.showMessage("Loading...", 1500);
				vc.setSkipData({
					op: {
						interval: {
							startTime: 10,
							endTime: 40
						}
					}
				});
				vc.clearSkipData();
				results.ui = "ok";
			} catch(e) {
				results.ui = e.message;
			}

			// 4. track control
			try {
				vc.setSubtitleTrack(1);
				vc.addSubtitleTrack({index: 1, label: "English"});
				vc.addExternalSubtitleTrack({src: "http://example.com/sub.vtt", label: "External"});
				vc.setMediaCaptionTrack(1);
				vc.addMediaCaptionTrack({index: 1});
				vc.setAudioTrack(2);
				results.tracks = "ok";
			} catch(e) {
				results.tracks = e.message;
			}

			// 5. state requests
			try {
				vc.sendGetFullscreen();
				vc.sendGetPip();
				vc.sendGetAnime4K();
				vc.sendGetSubtitleTrack();
				vc.sendGetAudioTrack();
				vc.sendGetMediaCaptionTrack();
				vc.sendGetPlaybackState();
				results.stateRequests = "ok";
			} catch(e) {
				results.stateRequests = e.message;
			}

			// 6. promises & async getters (awaiting them to check values)
			try {
				// getSkipData
				let skipData = await vc.getSkipData();
				if (skipData) {
					results.skipData = JSON.parse(JSON.stringify(skipData));
				} else {
					results.skipData = null;
				}

				// getPlaylist
				let playlist = await vc.getPlaylist();
				if (playlist) {
					results.playlist = JSON.parse(JSON.stringify(playlist));
				} else {
					results.playlist = null;
				}

				// pullStatus
				let status = await vc.pullStatus();
				if (status) {
					results.status = JSON.parse(JSON.stringify(status));
				} else {
					results.status = null;
				}

				// getTextTracks
				let textTracks = await vc.getTextTracks();
				results.textTracks = textTracks;

				results.asyncGetters = "ok";
			} catch(e) {
				results.asyncGetters = e.message;
			}

			// 7. sync getters
			try {
				const status = vc.getPlaybackStatus();
				if (status) {
					results.syncStatus = JSON.parse(JSON.stringify(status));
				}

				const state = vc.getPlaybackState();
				if (state) {
					results.syncState = JSON.parse(JSON.stringify(state));
				}

				const info = vc.getCurrentPlaybackInfo();
				if (info) {
					results.syncInfo = JSON.parse(JSON.stringify(info));
				}

				const media = vc.getCurrentMedia();
				if (media) {
					results.syncMedia = JSON.parse(JSON.stringify(media));
				}

				results.clientId = vc.getCurrentClientId();
				results.playerType = vc.getCurrentPlayerType();
				results.playbackType = vc.getCurrentPlaybackType();
				results.syncGetters = "ok";
			} catch(e) {
				results.syncGetters = e.message;
			}

			// 8. play initiate (returns promise, caught to prevent unhandled rejection)
			try {
				vc.playStream("http://url", "1", {}).catch(() => {});
				vc.playLocalFile("path").catch(() => {});
				results.playInitiate = "ok";
			} catch(e) {
				results.playInitiate = e.message;
			}

			reportResults(results);
		})().catch(e => {
			reportResults({ error: e.message });
		});
	`

	t.Run("Lifecycle: Inactive Session", func(t *testing.T) {
		res := runJSScriptAsync(t, vm, scheduler, testScript)

		// check that no control or getter method threw an exception
		assert.Equal(t, "ok", res["events"])
		assert.Equal(t, "ok", res["playback"])
		assert.Equal(t, "ok", res["ui"])
		assert.Equal(t, "ok", res["tracks"])
		assert.Equal(t, "ok", res["stateRequests"])
		assert.Equal(t, "ok", res["asyncGetters"])
		assert.Equal(t, "ok", res["syncGetters"])
		assert.Equal(t, "ok", res["playInitiate"])

		// since inactive, sync/async getters should resolve to null/undefined/empty values
		assert.Nil(t, res["status"])
		assert.Nil(t, res["skipData"])
		assert.Nil(t, res["playlist"])
		assert.Nil(t, res["syncStatus"])
		assert.Nil(t, res["syncState"])
		assert.Nil(t, res["syncInfo"])
		assert.Nil(t, res["syncMedia"])
		assert.Equal(t, "", res["clientId"])
		assert.Equal(t, "", res["playerType"])
		assert.Equal(t, "", res["playbackType"])
	})

	t.Run("Lifecycle: Loading Session", func(t *testing.T) {
		// mock loading state
		coordinator.OpenAndAwait(player.TargetVideoCore, "client-123", "loading")

		mockInfo := &player.PlaybackInfo{
			ID:           "playback-123",
			Target:       player.TargetVideoCore,
			Renderer:     player.RendererWeb,
			PlaybackType: player.PlaybackTypeOnlinestream,
			PlaybackURI:  "http://example.com/video.m3u8",
			StreamURL:    "http://example.com/video.m3u8",
			SubtitleTracks: []*player.SubtitleTrack{
				{
					Index:  0,
					URI:    new("http://example.com/sub.vtt"),
					Label:  "English",
					Format: new("vtt"),
				},
			},
			Media: &anilist.BaseAnime{
				ID: 42,
				Title: &anilist.BaseAnime_Title{
					Romaji: new("My Anime Title"),
				},
			},
			Episode: &anime.Episode{
				EpisodeNumber: 1,
			},
		}
		coordinator.Watch(player.TargetVideoCore, "client-123", mockInfo)

		res := runJSScriptAsync(t, vm, scheduler, testScript)

		// check basic bindings did not fail
		assert.Equal(t, "ok", res["syncGetters"])
		assert.Equal(t, "client-123", res["clientId"])
		assert.Equal(t, "onlinestream", res["playbackType"])
		assert.Equal(t, "web", res["playerType"]) // maps renderer web to web in string

		// verify plugin properties returned in syncInfo
		info := res["syncInfo"].(map[string]interface{})
		assert.Equal(t, "playback-123", info["id"])
		assert.Equal(t, "onlinestream", info["playbackType"])
		assert.Equal(t, "hls", info["streamType"])

		subTracks := info["subtitleTracks"].([]interface{})
		require.Len(t, subTracks, 1)
		subTrack := subTracks[0].(map[string]interface{})
		assert.Equal(t, "http://example.com/sub.vtt", subTrack["src"])
		assert.Equal(t, "vtt", subTrack["type"])
		assert.Equal(t, true, subTrack["useLibassRenderer"])

		// verify media info returned
		media := res["syncMedia"].(map[string]interface{})
		assertNumberEqual(t, 42, media["id"])

		// verify resolved async getters values
		// getSkipData
		skipData := res["skipData"].(map[string]interface{})
		require.NotNil(t, skipData)
		op := skipData["op"].(map[string]interface{})
		interval := op["interval"].(map[string]interface{})
		assertNumberEqual(t, 10, interval["startTime"])
		assertNumberEqual(t, 30, interval["endTime"])

		// getPlaylist
		playlist := res["playlist"].(map[string]interface{})
		require.NotNil(t, playlist)
		assert.Equal(t, "localfile", playlist["type"])
		currentEpisode := playlist["currentEpisode"].(map[string]interface{})
		assertNumberEqual(t, 3, currentEpisode["episodeNumber"])

		// pullStatus
		status := res["status"].(map[string]interface{})
		require.NotNil(t, status)
		assert.Equal(t, "playback-1", status["id"])
		assert.Equal(t, "client-1", status["clientId"])
		assert.Equal(t, false, status["paused"])
		assertNumberEqual(t, 42.5, status["currentTime"])
		assertNumberEqual(t, 100, status["duration"])

		// verify preparation/loading phase command propagation
		var skipDataCmdFound bool
		var clearSkipDataCmdFound bool
		var showMessageCmdFound bool

		for _, cmd := range mb.executedCmd {
			switch cmd.Type {
			case player.CommandSetSkipData:
				skipDataCmdFound = true
			case player.CommandClearSkipData:
				clearSkipDataCmdFound = true
			case player.CommandShowMessage:
				showMessageCmdFound = true
				payload := cmd.Payload.(player.ShowMessagePayload)
				assert.Equal(t, "Loading...", payload.Message)
				assert.Equal(t, 1500, payload.Duration)
			}
		}
		assert.True(t, skipDataCmdFound, "setSkipData command should execute during loading phase")
		assert.True(t, clearSkipDataCmdFound, "clearSkipData command should execute during loading phase")
		assert.True(t, showMessageCmdFound, "showMessage command should execute during loading phase")
	})

	t.Run("Lifecycle: Active Session Controls", func(t *testing.T) {
		// mock playback loaded -> session becomes fully active
		mb.executedCmd = nil

		// dispatch loaded event to coordinator
		coordinator.OpenAndAwait(player.TargetVideoCore, "client-123", "loading")
		mockInfo := &player.PlaybackInfo{
			ID:           "playback-123",
			Target:       player.TargetVideoCore,
			Renderer:     player.RendererWeb,
			PlaybackType: player.PlaybackTypeOnlinestream,
			PlaybackURI:  "http://example.com/video.m3u8",
			StreamURL:    "http://example.com/video.m3u8",
			SubtitleTracks: []*player.SubtitleTrack{
				{
					Index:  0,
					URI:    new("http://example.com/sub.vtt"),
					Label:  "English",
					Format: new("vtt"),
				},
			},
		}
		coordinator.Watch(player.TargetVideoCore, "client-123", mockInfo)

		// run js script that executes control methods
		_ = runJSScriptAsync(t, vm, scheduler, testScript)

		// assert all control commands propagated to mock backend
		expectedCommands := map[player.CommandType]bool{
			player.CommandPause:                    false,
			player.CommandResume:                   false,
			player.CommandSeek:                     false,
			player.CommandSeekTo:                   false,
			player.CommandSetFullscreen:            false,
			player.CommandSetPip:                   false,
			player.CommandSetSubtitleTrack:         false,
			player.CommandAddSubtitleTrack:         false,
			player.CommandAddExternalSubtitleTrack: false,
			player.CommandSetMediaCaptionTrack:     false,
			player.CommandAddMediaCaptionTrack:     false,
			player.CommandSetAudioTrack:            false,
			player.CommandPlayPlaylistEpisode:      false,
		}

		for _, cmd := range mb.executedCmd {
			expectedCommands[cmd.Type] = true
		}

		for cmdType, executed := range expectedCommands {
			assert.True(t, executed, "Command %s should have executed when session is active", cmdType)
		}
	})
}

func TestVideoCoreEventListenersPluginFormat(t *testing.T) {
	appCtx, coordinator, mb := newVideoCoreTestContext(t)
	vm, obj, scheduler := bindTestVideoCore(t, appCtx)

	err := vm.Set("videoCore", obj.Get("videoCore"))
	require.NoError(t, err)

	// we register listeners, subscribe to coordinator, send backend events, and capture JS event properties
	testScript := `
		(async function() {
			const vc = videoCore;
			let receivedEvents = {};

			vc.addEventListener("video-paused", (e) => {
				receivedEvents.pausedEvent = {
					playerType: e.playerType,
					playbackType: e.playbackType,
					playbackId: e.playbackId,
					clientId: e.clientId,
					currentTime: e.currentTime,
					duration: e.duration
				};
				if (receivedEvents.pausedEvent && receivedEvents.statusEvent) {
					reportResults(receivedEvents);
				}
			});

			vc.addEventListener("video-status", (e) => {
				receivedEvents.statusEvent = {
					playerType: e.playerType,
					playbackType: e.playbackType,
					playbackId: e.playbackId,
					clientId: e.clientId,
					currentTime: e.currentTime,
					duration: e.duration,
					paused: e.paused
				};
				if (receivedEvents.pausedEvent && receivedEvents.statusEvent) {
					reportResults(receivedEvents);
				}
			});
		})().catch(e => {
			reportResults({ error: e.message });
		});
	`

	resultChan := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	err = scheduler.Schedule(func() error {
		return vm.Set("reportResults", func(call goja.FunctionCall) goja.Value {
			var res map[string]interface{}
			err := vm.ExportTo(call.Arguments[0], &res)
			if err != nil {
				errChan <- err
			} else {
				resultChan <- res
			}
			return goja.Undefined()
		})
	})
	require.NoError(t, err)

	scheduler.ScheduleAsync(func() error {
		_, err := vm.RunString(testScript)
		if err != nil {
			errChan <- err
		}
		return nil
	})

	// wait a tiny bit for the JS to bind event listeners
	time.Sleep(100 * time.Millisecond)

	// mock active session so events are not discarded by session check
	coordinator.OpenAndAwait(player.TargetVideoCore, "client-123", "loading")
	mockInfo := &player.PlaybackInfo{
		ID:           "playback-123",
		Target:       player.TargetVideoCore,
		Renderer:     player.RendererWeb,
		PlaybackType: player.PlaybackTypeOnlinestream,
		PlaybackURI:  "http://example.com/video.m3u8",
		StreamURL:    "http://example.com/video.m3u8",
		SubtitleTracks: []*player.SubtitleTrack{
			{
				Index:  0,
				URI:    new("http://example.com/sub.vtt"),
				Label:  "English",
				Format: new("vtt"),
			},
		},
	}
	coordinator.Watch(player.TargetVideoCore, "client-123", mockInfo)

	// dispatch events through mock backend channel
	sessKey := player.SessionKey{
		Target:     player.TargetVideoCore,
		ClientID:   "client-123",
		PlaybackID: "playback-123",
	}

	baseEv := player.BaseEvent{Session: sessKey}

	mb.eventsCh <- &player.PausedEvent{
		BaseEvent:   baseEv,
		CurrentTime: 12.3,
		Duration:    120.0,
	}

	mb.eventsCh <- &player.StatusEvent{
		BaseEvent:   baseEv,
		CurrentTime: 45.6,
		Duration:    120.0,
		Paused:      true,
	}

	// wait for results
	select {
	case res := <-resultChan:
		require.NotNil(t, res["pausedEvent"])
		paused := res["pausedEvent"].(map[string]interface{})
		assert.Equal(t, "web", paused["playerType"])
		assert.Equal(t, "onlinestream", paused["playbackType"])
		assert.Equal(t, "playback-123", paused["playbackId"])
		assert.Equal(t, "client-123", paused["clientId"])
		assertNumberEqual(t, 12.3, paused["currentTime"])
		assertNumberEqual(t, 120.0, paused["duration"])

		require.NotNil(t, res["statusEvent"])
		status := res["statusEvent"].(map[string]interface{})
		assert.Equal(t, "web", status["playerType"])
		assert.Equal(t, "onlinestream", status["playbackType"])
		assert.Equal(t, "playback-123", status["playbackId"])
		assert.Equal(t, "client-123", status["clientId"])
		assertNumberEqual(t, 45.6, status["currentTime"])
		assertNumberEqual(t, 120.0, status["duration"])
		assert.Equal(t, true, status["paused"])

	case err := <-errChan:
		t.Fatalf("JS runtime execution error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for JS execution results")
	}
}
