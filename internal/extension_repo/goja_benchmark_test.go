package extension_repo

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGojaPlugin(t *testing.T) {
	payload := `
	function init() {

		$app.onGetAnime((e) => {

			if(e.anime.id === 178022) {
				e.anime.id = 21;
				e.anime.idMal = 21;
				$replace(e.anime.id, 22)
				$replace(e.anime.title, { "english": "The One Piece is Real" })
				// e.anime.title = { "english": "The One Piece is Real" }
				// $replace(e.anime.synonyms, ["The One Piece is Real"])
				e.anime.synonyms = ["The One Piece is Real"]
				// e.anime.synonyms[0] = "The One Piece is Real"
				// $replace(e.anime.synonyms[0], "The One Piece is Real")
			}

			e.next();
		});

		$app.onGetAnime((e) => {
			console.log("$app.onGetAnime(2) fired")
			console.log(e.anime.id)
			console.log(e.anime.idMal)
			console.log(e.anime.synonyms[0])
			console.log(e.anime.title)
		});
	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	_, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	m, err := anilistPlatform.GetAnime(178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)
	util.Spew(m.Synonyms)

	// m, err = anilistPlatform.GetAnime(177709)
	// if err != nil {
	// 	t.Fatalf("GetAnime returned error: %v", err)
	// }

	// util.Spew(m.Title)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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

	_, _, runtimeManager, _, _, err := InitTestPlugin(b, opts)
	require.NoError(b, err)

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
