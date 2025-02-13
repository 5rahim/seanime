package extension_repo

import (
	"runtime"
	"seanime/internal/hook_event"
	"testing"

	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
)

func TestNewGojaPlugin(t *testing.T) {

	payload := `
	function init() {
		console.log("init called");
		$app.onGetBaseAnime((e) => {
			console.log("onGetBaseAnime fired")
			console.log(e.anime)
			if(e.anime.id === 178022) {
				e.anime.id = 22;
				e.anime.idMal = 22;
				e.anime.title.english = "The One Piece is Real";
			}
			e.next();
		});

		$app.onGetBaseAnime((e) => {
			console.log("onGetBaseAnime(2) fired")
			console.log(e.anime.id)
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

	// Here the plugin is being initialized
	_, err := NewGojaPlugin(ext, lang, logger, goja_runtime.NewManager(logger, 1), hm)
	if err != nil {
		t.Fatalf("NewGojaPlugin returned error: %v", err)
	}

	m, err := anilistPlatform.GetAnime(178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m)

	m, err = anilistPlatform.GetAnime(177709)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m)

}

// func TestNewGojaPlugin(t *testing.T) {

// 	// create a dummy extension with a simple JS payload
// 	ext := &extension.Extension{
// 		ID:      "dummy-plugin",
// 		Payload: "onGetBaseAnime((e) => { console.log(e.Anime); e.next(); })", // log the anime object
// 	}

// 	// use 'javascript' as the language
// 	lang := extension.Language("javascript")
// 	// Create a pointer logger
// 	nLogger := zerolog.Nop()
// 	logger := &nLogger

// 	// Create a runtime manager with a pool size of 1
// 	runtimeManager := goja_runtime.NewManager(logger, 1)

// 	// Create a hook manager with the no-op logger
// 	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

// 	// Create the GojaPlugin first so hooks are registered
// 	plugin, err := NewGojaPlugin(ext, lang, logger, runtimeManager, hm)
// 	if err != nil {
// 		t.Fatalf("NewGojaPlugin returned error: %v", err)
// 	}
// 	if plugin == nil {
// 		t.Fatal("NewGojaPlugin returned nil plugin")
// 	}

// 	// Create AniList platform after plugin is setup
// 	anilistPlatform := anilist_platform.NewAnilistPlatform(anilist.NewAnilistClient(""), logger, hm)

// 	// Retrieve a runtime from the plugin's pool
// 	vm, err := plugin.pool.Get(context.Background())
// 	if err != nil {
// 		t.Fatalf("Failed to get runtime from pool: %v", err)
// 	}
// 	defer plugin.pool.Put(vm)

// 	// Create a channel to wait for the hook to be triggered
// 	done := make(chan struct{})

// 	// Add a hook handler that will close the channel when called
// 	hm.OnGetBaseAnime().BindFunc(func(e hook.Resolver) error {
// 		close(done)
// 		return e.Next()
// 	})

// 	// Call GetAnime which should trigger the hook
// 	go anilistPlatform.GetAnime(21)

// 	// Wait for the hook to be triggered or timeout
// 	select {
// 	case <-done:
// 		// Hook was triggered successfully
// 	case <-time.After(2 * time.Second):
// 		t.Fatal("Hook was not triggered within timeout")
// 	}

// 	// Check that hook functions were bound.
// 	// The hooksBinds function registers methods from the hook manager (e.g. OnGetBaseAnime).
// 	// Depending on FieldMapper implementation, the property name might be 'OnGetBaseAnime' or 'onGetBaseAnime'.
// 	var hookFn goja.Value
// 	hookFn = vm.Get("OnGetBaseAnime")
// 	if goja.IsUndefined(hookFn) || hookFn == nil {
// 		hookFn = vm.Get("onGetBaseAnime")
// 	}
// 	if goja.IsUndefined(hookFn) || hookFn == nil {
// 		t.Error("Expected hook function for OnGetBaseAnime to be defined in the runtime")
// 	}
// }

func BenchmarkHookInvocation(b *testing.B) {
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	// Dummy extension payload that registers a hook
	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) { 
				e.next();
			});
		}
		init();  // Call init immediately
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	// Initialize the plugin, which will bind the hook
	plugin, err := NewGojaPlugin(ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
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
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) { 
				e.next();
			});
		}
		init();
	`

	// Create a runtime manager with a pool size matching GOMAXPROCS
	runtimeManager := goja_runtime.NewManager(logger, int32(runtime.GOMAXPROCS(0)))

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		&extension.Extension{
			ID:      "dummy-hook-benchmark",
			Payload: payload,
		},
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

// BenchmarkHookEmpty measures performance with an empty hook that just calls next()
func BenchmarkHookEmpty(b *testing.B) {
	logger := util.NewLogger()
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})

	payload := `
		function init() {
			$app.onGetBaseAnime(function(e) { 
				e.next();
			});
		}
		init();
	`
	ext := &extension.Extension{
		ID:      "dummy-hook-benchmark",
		Payload: payload,
	}

	plugin, err := NewGojaPlugin(ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
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

// BenchmarkHookWithWork measures performance with a hook that does some actual work
func BenchmarkHookWithWork(b *testing.B) {
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

	plugin, err := NewGojaPlugin(ext, extension.LanguageJavascript, logger, goja_runtime.NewManager(logger, 1), hm)
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
func BenchmarkHookParallel(b *testing.B) {
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

	// Create a runtime manager with a pool size matching GOMAXPROCS
	runtimeManager := goja_runtime.NewManager(logger, int32(runtime.GOMAXPROCS(0)))

	// Initialize the plugin with the runtime manager
	plugin, err := NewGojaPlugin(
		&extension.Extension{
			ID:      "dummy-hook-benchmark",
			Payload: payload,
		},
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
