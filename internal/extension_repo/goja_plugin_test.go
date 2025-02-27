package extension_repo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestNewGojaPluginUI(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		time.Sleep(2000 * time.Millisecond)
		fmt.Fprint(w, `{"test": "data"}`)
	}))
	defer server.Close()

	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())
	payload := fmt.Sprintf(`
	function init() {

		$app.onGetAnime(async (e) => {

			$store.set("anime", e.anime);
			$store.set("value", 42);

			console.log("onGetAnime fired", $store.get("value"));

			e.next();
		});

		$ui.register((ctx) => {
			console.log("%s");
			console.log("this is the start");

			const count = ctx.state(0)

			ctx.effect(() => {
				console.log("running effect that takes 1s")
				ctx.setTimeout(() => {
					console.log("1s elapsed since first effect called")
				}, 1000)
			}, [count])

			ctx.effect(() => {
				console.log("running effect that runs fast ran second")
			}, [count])

			count.set(p => p+1)

			console.log("this is the end");
		});

	}
	`, server.URL)

	ext := &extension.Extension{
		ID:      "dummy-plugin",
		Payload: payload,
	}

	logger := util.NewLogger()

	wsEventManager := events.NewMockWSEventManager(logger)
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewMockAnilistClient(), logger)
	_ = anilistPlatform

	manager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, manager)

	//go func() {
	//	time.Sleep(time.Second)
	//	_, err := anilistPlatform.GetAnime(178022)
	//	if err != nil {
	//		t.Errorf("GetAnime returned error: %v", err)
	//	}
	//
	//	_, err = anilistPlatform.GetAnime(177709)
	//	if err != nil {
	//		t.Errorf("GetAnime returned error: %v", err)
	//	}
	//}()

	_, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, manager, wsEventManager)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

	manager.PrintPluginPoolMetrics(ext.ID)

	time.Sleep(6 * time.Second)
}

func TestNewGojaPluginContext(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())
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

	ext := &extension.Extension{
		ID:      "dummy-plugin",
		Payload: payload,
	}

	logger := util.NewLogger()

	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewMockAnilistClient(), logger)

	manager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, manager)

	wsEventManager := events.NewMockWSEventManager(logger)
	_, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, manager, wsEventManager)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

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

	manager.PrintPluginPoolMetrics(ext.ID)

}

func TestNewGojaPlugin(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())
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

	ext := &extension.Extension{
		ID:      "dummy-plugin",
		Payload: payload,
	}

	logger := util.NewLogger()

	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewMockAnilistClient(), logger)

	manager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, manager)

	wsEventManager := events.NewMockWSEventManager(logger)
	_, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, manager, wsEventManager)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

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

	manager.PrintPluginPoolMetrics(ext.ID)

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
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				e.next();
			});
		}
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	wsEventManager := events.NewMockWSEventManager(logger)

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager, wsEventManager)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin // keep the plugin reference alive

	title := "Test Anime"
	// Create a dummy anime event that we'll reuse
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
		if err := hm.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
		//b.ReportMetric(b.Elapsed().Seconds(), "s/op")
	}

	runtimeManager.PrintPluginPoolMetrics(ext.ID)

}

func BenchmarkNoHookInvocation(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onMissingEpisodes(function(e) {
				e.next();
			});
		}
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	wsEventManager := events.NewMockWSEventManager(logger)
	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager, wsEventManager)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin // keep the plugin reference alive

	title := "Test Anime"
	// Create a dummy anime event that we'll reuse
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
		if err := hm.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(ext.ID)
}

// Add a parallel version to see how it performs under concurrent load
func BenchmarkHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	wsEventManager := events.NewMockWSEventManager(logger)

	payload := `
		function init() {
			$app.onGetAnime(function(e) {
				e.next();
			});
		}
	`

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		wsEventManager,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

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
			if err := hm.OnGetAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})

	runtimeManager.PrintPluginPoolMetrics(ext.ID)
}

func BenchmarkNoHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	wsEventManager := events.NewMockWSEventManager(logger)

	payload := `
		function init() {
			$app.onMissingEpisodes(function(e) {
				e.next();
			});
		}
	`

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		wsEventManager,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

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
			if err := hm.OnGetAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})
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
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

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
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)

	loader := NewGojaPluginLoader(ext, logger, runtimeManager)
	wsEventManager := events.NewMockWSEventManager(logger)
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager, wsEventManager)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

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
		if err := hm.OnGetAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}

	runtimeManager.PrintPluginPoolMetrics(ext.ID)
}

// BenchmarkHookParallel measures parallel performance with a hook that does some work
func BenchmarkHookInvocationWithWorkParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()

	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	wsEventManager := events.NewMockWSEventManager(logger)

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

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		wsEventManager,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

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

	runtimeManager.PrintPluginPoolMetrics(ext.ID)
}

func BenchmarkNoHookInvocationWithWork(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()

	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hm)

	wsEventManager := events.NewMockWSEventManager(logger)

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
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	runtimeManager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(ext, logger, runtimeManager)

	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager, wsEventManager)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

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

	runtimeManager.PrintPluginPoolMetrics(ext.ID)
}
