package extension_repo

import (
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

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

	lang := extension.Language("typescript")

	logger := util.NewLogger()

	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewAnilistClient(""), logger)

	manager := goja_runtime.NewManager(logger, 15)
	loader := NewGojaPluginLoader(logger, manager)

	_, err := NewGojaPlugin(loader, ext, lang, logger, manager)
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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager)
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

}

func BenchmarkNoHookInvocation(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager)
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
}

// Add a parallel version to see how it performs under concurrent load
func BenchmarkHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
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

func BenchmarkNoHookInvocationParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
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

	loader := NewGojaPluginLoader(logger, runtimeManager)
	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager)
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
}

// BenchmarkHookParallel measures parallel performance with a hook that does some work
func BenchmarkHookInvocationWithWorkParallel(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()

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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		loader,
		ext,
		extension.LanguageJavascript,
		logger,
		runtimeManager,
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
}

func BenchmarkNoHookInvocationWithWork(b *testing.B) {
	b.ReportAllocs()
	logger := util.NewLogger()

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
	loader := NewGojaPluginLoader(logger, runtimeManager)

	plugin, err := NewGojaPlugin(loader, ext, extension.LanguageJavascript, logger, runtimeManager)
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
}
