package extension_repo

import (
	"runtime"
	"testing"

	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/hook_event"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
)

func TestNewGojaPlugin(t *testing.T) {

	payload := `
	function init() {
		console.log("init called");
		$app.onGetBaseAnime((e) => {
			console.log("$app.onGetBaseAnime fired")
			if(e.anime.id === 178022) {
				e.anime.id = 22;
				e.anime.idMal = 22;
				e.anime.title.english = "The One Piece is Real";
			}

			// Store a value
			$ctx.cache.set("myKey", 42);

			// Retrieve it later in another hook
			const value = $ctx.cache.get("myKey"); // 42
			console.log(value)

			e.next();
		});

		$app.onGetBaseAnime((e) => {
			console.log("$app.onGetBaseAnime(2) fired")
			console.log(e.anime.id)

			// Check if exists
			const value = $ctx.cache.get("myKey"); // 42
			console.log(value)
		});
	}
	`

	ext := &extension.Extension{
		ID:      "dummy-plugin",
		Payload: payload,
	}

	lang := extension.Language("typescript")

	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewAnilistClient(""), logger, hm)
	anilistPlatform.SetAnilistClient(anilist.NewAnilistClient(""))

	// Use a single runtimeManager for both loader and plugin
	manager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(logger, manager, hm)

	// Here the plugin is being initialized using the same manager
	_, err := NewGojaPlugin(loader, ext, lang, logger, manager, hm)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

	m, err := anilistPlatform.GetAnime(178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.GetTitleSafe())

	m, err = anilistPlatform.GetAnime(177709)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.GetTitleSafe())

}

func BenchmarkAllHooks(b *testing.B) {
	b.Run("BaselineNoHook", BenchmarkBaselineNoHook)
	b.Run("HookInvocation", BenchmarkHookInvocation)
	b.Run("HookInvocationParallel", BenchmarkHookInvocationParallel)
	b.Run("NoHookInvocation", BenchmarkNoHookInvocation)
	b.Run("NoHookInvocationParallel", BenchmarkNoHookInvocationParallel)
	b.Run("HookWithWork", BenchmarkHookWithWork)
	b.Run("HookWithWorkParallel", BenchmarkHookWithWorkParallel)
	b.Run("NoHookInvocationWithWork", BenchmarkNoHookInvocationWithWork)
}

func BenchmarkHookInvocation(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) {
				e.next();
			});
		}
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin // keep the plugin reference alive

	title := "Test Anime"
	// Create a dummy anime event that we'll reuse
	dummyEvent := &hook_event.GetBaseAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hm.OnGetBaseAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
		//b.ReportMetric(b.Elapsed().Seconds(), "s/op")
	}

}

func BenchmarkNoHookInvocation(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onGetBaseAnimeError(function(e) {
				e.next();
			});
		}
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin // keep the plugin reference alive

	title := "Test Anime"
	// Create a dummy anime event that we'll reuse
	dummyEvent := &hook_event.GetBaseAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hm.OnGetBaseAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}
}

// Add a parallel version to see how it performs under concurrent load
func BenchmarkHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) {
				e.next();
			});
		}
	`

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	// Create a runtime manager with a pool size matching GOMAXPROCS
	runtimeManager := goja_runtime.NewManager(logger, int32(runtime.GOMAXPROCS(0)))

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		hm,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			title := "Test Anime"
			event := &hook_event.GetBaseAnimeEvent{
				Anime: &anilist.BaseAnime{
					ID: 1234,
					Title: &anilist.BaseAnime_Title{
						English: &title,
					},
				},
			}
			if err := hm.OnGetBaseAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkNoHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnimeError(function(e) {
				e.next();
			});
		}
	`

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	runtimeManager := goja_runtime.NewManager(logger, int32(runtime.GOMAXPROCS(0)))

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		hm,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			title := "Test Anime"
			event := &hook_event.GetBaseAnimeEvent{
				Anime: &anilist.BaseAnime{
					ID: 1234,
					Title: &anilist.BaseAnime_Title{
						English: &title,
					},
				},
			}
			if err := hm.OnGetBaseAnime().Trigger(event); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkBaselineNoHook measures the baseline performance without any hooks
func BenchmarkBaselineNoHook(b *testing.B) {
	b.ReportAllocs()
	title := "Test Anime"
	dummyEvent := &hook_event.GetBaseAnimeEvent{
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

// BenchmarkHookWithWork measures performance with a hook that does some actual work
func BenchmarkHookWithWork(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) {
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

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

	title := "Test Anime"
	dummyEvent := &hook_event.GetBaseAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hm.OnGetBaseAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHookParallel measures parallel performance with a hook that does some work
func BenchmarkHookWithWorkParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) {
				// Do some work
				if (e.anime.id === 1234) {
					e.anime.id = 5678;
					e.anime.title.english = "Modified Title";
					e.anime.idMal = 9012;
				}
				e.next();
			});
		}
		init();
	`

	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	// Create a runtime manager with a pool size matching GOMAXPROCS
	runtimeManager := goja_runtime.NewManager(logger, int32(runtime.GOMAXPROCS(0)))

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
		hm,
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

	title := "Test Anime"
	dummyEvent := &hook_event.GetBaseAnimeEvent{
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
			if err := hm.OnGetBaseAnime().Trigger(dummyEvent); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkNoHookInvocationWithWork(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnimeError(function(e) {
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

	loader := NewGojaPluginLoader(logger, goja_runtime.NewManager(logger, 1), hm)

	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
	if err != nil {
		b.Fatal(err)
	}
	_ = plugin

	title := "Test Anime"
	dummyEvent := &hook_event.GetBaseAnimeEvent{
		Anime: &anilist.BaseAnime{
			ID: 1234,
			Title: &anilist.BaseAnime_Title{
				English: &title,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := hm.OnGetBaseAnime().Trigger(dummyEvent); err != nil {
			b.Fatal(err)
		}
	}
}
