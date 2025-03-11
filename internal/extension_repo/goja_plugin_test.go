package extension_repo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seanime/internal/api/anilist"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/plugin"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// TestPluginOptions contains options for initializing a test plugin
type TestPluginOptions struct {
	ID          string
	Payload     string
	Language    extension.Language
	Permissions []extension.PluginPermission
	PoolSize    int
	SetupHooks  bool
}

// DefaultTestPluginOptions returns default options for a test plugin
func DefaultTestPluginOptions() TestPluginOptions {
	return TestPluginOptions{
		ID:          "dummy-plugin",
		Payload:     "",
		Language:    extension.LanguageJavascript,
		Permissions: nil,
		PoolSize:    15,
		SetupHooks:  true,
	}
}

// InitTestPlugin initializes a test plugin with the given options
func InitTestPlugin(t testing.TB, opts TestPluginOptions) (*GojaPlugin, *zerolog.Logger, *goja_runtime.Manager, *anilist_platform.AnilistPlatform, events.WSEventManagerInterface) {
	if opts.SetupHooks {
		test_utils.SetTwoLevelDeep()
		if tPtr, ok := t.(*testing.T); ok {
			test_utils.InitTestProvider(tPtr, test_utils.Anilist())
		}
	}

	ext := &extension.Extension{
		ID:       opts.ID,
		Payload:  opts.Payload,
		Language: opts.Language,
	}

	if len(opts.Permissions) > 0 {
		ext.Plugin = &extension.PluginManifest{
			Permissions: opts.Permissions,
		}
	}

	logger := util.NewLogger()
	wsEventManager := events.NewMockWSEventManager(logger)
	anilistClient := anilist.NewMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger).(*anilist_platform.AnilistPlatform)

	// Initialize hook manager if needed
	if opts.SetupHooks {
		hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
		hook.SetGlobalHookManager(hm)
	}

	manager := goja_runtime.NewManager(logger, int32(opts.PoolSize))
	loader := NewGojaPluginLoader(ext, logger, manager)

	plugin, _, err := NewGojaPlugin(loader, ext, opts.Language, logger, manager, wsEventManager)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

	return plugin, logger, manager, anilistPlatform, wsEventManager
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginMpv(t *testing.T) {
	payload := `
function init() {

	$ui.register((ctx) => {

		console.log("Testing MPV");

		const conn = $mpv.newConnection("/tmp/mpv_socket")
		conn.open()


		console.log("Connection created", conn)

		conn.call("observe_property", 42, "time-pos")

		const cancel = $mpv.registerEventListener(conn, (event) => {
			console.log("Event received", event)
		})

		// conn.call("set_property", "pause", true)

		ctx.setTimeout(() => {
			console.log("Cancelling event listener")
			cancel()
		}, 1000)
	});

}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = []extension.PluginPermission{
		extension.PluginPermissionPlayback,
	}

	_, _, manager, _, _ := InitTestPlugin(t, opts)

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(8 * time.Second)
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginPlaybackEvents(t *testing.T) {
	payload := `
function init() {

	$ui.register((ctx) => {
		console.log("Testing Playback");

		const cancel = $playback.registerEventListener("mySubscriber", (event) => {
			console.log("Event received", event)
		})

		ctx.setTimeout(() => {
			console.log("Cancelling event listener")
			cancel()
		}, 1000)
	});

}
	`

	playbackManager, _, err := getPlaybackManager(t)
	require.NoError(t, err)

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		PlaybackManager: playbackManager,
	})

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = []extension.PluginPermission{
		extension.PluginPermissionPlayback,
	}

	_, _, manager, _, _ := InitTestPlugin(t, opts)

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(8 * time.Second)
}

func TestNewGojaPluginUI(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		time.Sleep(1000 * time.Millisecond)
		fmt.Fprint(w, `{"test": "data"}`)
	}))
	defer server.Close()

	payload := fmt.Sprintf(`
	function init() {

		$app.onGetAnime(async (e) => {
			const url = "%s"

			// const res = $await(fetch(url))
			const res = await fetch(url)
			const data = res.json()
			console.log("fetched results in hook", data)
			$store.set("data", data)

			console.log("first hook fired");

			e.next();
		});

		$app.onGetAnime(async (e) => {
			console.log("results from first hook", $store.get("data"));

			e.next();
		});

		$ui.register((ctx) => {
			console.log("%s");
			console.log("this is the start");

			const count = ctx.state(0)

			ctx.effect(async () => {
				console.log("running effect that takes 1s")
				ctx.setTimeout(() => {
					console.log("1s elapsed since first effect called")
				}, 1000)
				const [a, b, c] = await Promise.all([
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/1"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/2"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
				])
				console.log("fetch results", a.json(), b.json(), c.json())
			}, [count])

			ctx.effect(() => {
				console.log("running effect that runs fast ran second")
			}, [count])

			count.set(p => p+1)

			console.log("this is the end");
		});

	}
	`, server.URL, server.URL)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	_, _, manager, anilistPlatform, _ := InitTestPlugin(t, opts)

	go func() {
		time.Sleep(time.Second)
		_, err := anilistPlatform.GetAnime(178022)
		if err != nil {
			t.Errorf("GetAnime returned error: %v", err)
		}

		// _, err = anilistPlatform.GetAnime(177709)
		// if err != nil {
		// 	t.Errorf("GetAnime returned error: %v", err)
		// }
	}()

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(8 * time.Second)
}

func TestNewGojaPluginContext(t *testing.T) {
	payload := `
	function init() {

		$app.onGetAnime((e) => {

			$store.set("anime", e.anime);
			$store.set("value", 42);

			e.next();
		});

		$app.onGetAnime((e) => {

			console.log("Hook 2, value", $store.get("value"));
			console.log("Hook 2, value 2", $store.get("value2"));

			e.next();
		});

	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	_, _, manager, anilistPlatform, _ := InitTestPlugin(t, opts)

	m, err := anilistPlatform.GetAnime(178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)
	util.Spew(m.Synonyms)

	m, err = anilistPlatform.GetAnime(177709)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)

	manager.PrintPluginPoolMetrics(opts.ID)
}

func TestNewGojaPlugin(t *testing.T) {
	payload := `
	function init() {

		$app.onGetAnime((e) => {

			if(e.anime.id === 178022) {
				e.anime.id = 21;
				$replace(e.anime.title, { "english": "The One Piece is Real" })
				$replace(e.anime.synonyms, ["The One Piece is Real"])
				e.anime.synonyms[0] = "The One Piece"
			}

			e.next();
		});

		$app.onGetAnime((e) => {
			console.log("$app.onGetAnime(2) fired")
			console.log(e.anime.id)
			console.log(e.anime.synonyms[0])
			console.log(e.anime.title)
		});
	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	_, _, manager, anilistPlatform, _ := InitTestPlugin(t, opts)

	m, err := anilistPlatform.GetAnime(178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)
	util.Spew(m.Synonyms)

	m, err = anilistPlatform.GetAnime(177709)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)

	manager.PrintPluginPoolMetrics(opts.ID)
}

func BenchmarkAllHooks(b *testing.B) {
	b.Run("BaselineNoHook", BenchmarkBaselineNoHook)
	b.Run("HookInvocation", BenchmarkHookInvocation)
	b.Run("HookInvocationParallel", BenchmarkHookInvocationParallel)
	b.Run("HookInvocationWithWork", BenchmarkHookInvocationWithWork)
	b.Run("HookInvocationWithWorkParallel", BenchmarkHookInvocationWithWorkParallel)
	b.Run("NoHookInvocation", BenchmarkNoHookInvocation)
	b.Run("NoHookInvocationParallel", BenchmarkNoHookInvocationParallel)
	b.Run("NoHookInvocationWithWork", BenchmarkNoHookInvocationWithWork)
}

func BenchmarkHookInvocation(b *testing.B) {
	b.ReportAllocs()

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	// Create a dummy anime event that we'll reuse
	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hook.GlobalHookManager.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

func BenchmarkNoHookInvocation(b *testing.B) {
	b.ReportAllocs()

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onMissingEpisodes(function(e) {
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	// Create a dummy anime event that we'll reuse
	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hook.GlobalHookManager.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

// Add a parallel version to see how it performs under concurrent load
func BenchmarkHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()

	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	title := "Test Anime"
	event := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := hook.GlobalHookManager.OnGetAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

func BenchmarkNoHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()

	payload := `
		function init() {
			$app.onMissingEpisodes(function(e) {
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	title := "Test Anime"
	event := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := hook.GlobalHookManager.OnGetAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

// BenchmarkBaselineNoHook measures the baseline performance without any hooks
func BenchmarkBaselineNoHook(b *testing.B) {
	b.ReportAllocs()
	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dummyEvent.Next()
	}
}

// BenchmarkHookInvocationWithWork measures performance with a hook that does some actual work
func BenchmarkHookInvocationWithWork(b *testing.B) {
	b.ReportAllocs()

	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				// Do some work
				if (e.anime.id === 1234) {
					e.anime.id = 5678;
					e.anime.title.english = "Modified Title";
					e.anime.idMal = 9012;
				}
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hook.GlobalHookManager.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

// BenchmarkHookParallel measures parallel performance with a hook that does some work
func BenchmarkHookInvocationWithWorkParallel(b *testing.B) {
	b.ReportAllocs()

	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				// Do some work
				if (e.anime.id === 1234) {
					e.anime.id = 5678;
					e.anime.title.english = "Modified Title";
					e.anime.idMal = 9012;
				}
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := hook.GlobalHookManager.OnGetAnime().Trigger(dummyEvent); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

func BenchmarkNoHookInvocationWithWork(b *testing.B) {
	b.ReportAllocs()

	payload := `
		function init() {
			$app.onMissingEpisodes(function(e) {
				// Do some work
				if (e.anime.id === 1234) {
					e.anime.id = 5678;
					e.anime.title.english = "Modified Title";
					e.anime.idMal = 9012;
				}
				e.next();
			});
		}
	`

	opts := DefaultTestPluginOptions()
	opts.ID = "dummy-hook-benchmark"
	opts.Payload = payload
	opts.SetupHooks = true

	_, _, runtimeManager, _, _ := InitTestPlugin(b, opts)

	title := "Test Anime"
	dummyEvent := &anilist_platform.GetAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hook.GlobalHookManager.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(opts.ID)
}

/////////////////////////////////////////////////////////////////////////////////////////////s

func getPlaybackManager(t *testing.T) (*playbackmanager.PlaybackManager, *anilist.AnimeCollection, error) {

	logger := util.NewLogger()

	wsEventManager := events.NewMockWSEventManager(logger)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)

	if err != nil {
		t.Fatalf("error while creating database, %v", err)
	}

	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollection(true)
	require.NoError(t, err)
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: filecacher,
		Logger:     logger,
		Database:   database,
	})

	return playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager: wsEventManager,
		Logger:         logger,
		Platform:       anilistPlatform,
		Database:       database,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		DiscordPresence:   nil,
		IsOffline:         false,
		ContinuityManager: continuityManager,
	}), animeCollection, nil
}
