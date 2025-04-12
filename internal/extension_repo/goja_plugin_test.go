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
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpv"
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

var (
	testDocumentsDir          = "/Users/rahim/Documents"
	testDocumentCollectionDir = "/Users/rahim/Documents/collection"
	testVideoPath             = "/Users/rahim/Documents/collection/Bocchi the Rock/[ASW] Bocchi the Rock! - 01 [1080p HEVC][EDC91675].mkv"

	tempTestDir = "$TEMP/test"
)

// TestPluginOptions contains options for initializing a test plugin
type TestPluginOptions struct {
	ID          string
	Payload     string
	Language    extension.Language
	Permissions extension.PluginPermissions
	PoolSize    int
	SetupHooks  bool
}

// DefaultTestPluginOptions returns default options for a test plugin
func DefaultTestPluginOptions() TestPluginOptions {
	return TestPluginOptions{
		ID:          "dummy-plugin",
		Payload:     "",
		Language:    extension.LanguageJavascript,
		Permissions: extension.PluginPermissions{},
		PoolSize:    15,
		SetupHooks:  true,
	}
}

// InitTestPlugin initializes a test plugin with the given options
func InitTestPlugin(t testing.TB, opts TestPluginOptions) (*GojaPlugin, *zerolog.Logger, *goja_runtime.Manager, *anilist_platform.AnilistPlatform, events.WSEventManagerInterface, error) {
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
		Plugin:   &extension.PluginManifest{},
	}

	if len(opts.Permissions.Scopes) > 0 {
		ext.Plugin = &extension.PluginManifest{
			Permissions: opts.Permissions,
		}
	}

	ext.Plugin.Permissions.Allow = opts.Permissions.Allow

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

	plugin, _, err := NewGojaPlugin(ext, opts.Language, logger, manager, wsEventManager)
	return plugin, logger, manager, anilistPlatform, wsEventManager, err
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginMpv(t *testing.T) {
	payload := fmt.Sprintf(`
function init() {

	$ui.register(async (ctx) => {

		console.log("Testing MPV");

		await ctx.mpv.openAndPlay("%s")

		const cancel = ctx.mpv.onEvent((event) => {
			console.log("Event received", event)
		})

		ctx.setTimeout(() => {
			const conn = ctx.mpv.getConnection()
			if (conn) {
				conn.call("set_property", "pause", true)
			}
		}, 3000)

		ctx.setTimeout(async () => {
			console.log("Cancelling event listener")
			cancel()
			await ctx.mpv.stop()
		}, 5000)
	});

}
	`, testVideoPath)

	playbackManager, _, err := getPlaybackManager(t)
	require.NoError(t, err)

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		PlaybackManager: playbackManager,
	})

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionPlayback,
		},
	}

	_, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(8 * time.Second)
}

/////////////////////////////////////////////////////////////////////////////////////////////

// Test that the plugin cannot access paths that are not allowed
// $os.readDir should throw an error
func TestGojaPluginPathNotAllowed(t *testing.T) {
	payload := fmt.Sprintf(`
function init() {
	$ui.register((ctx) => {

		const tempDir = $os.tempDir();
		console.log("Temp dir", tempDir);

		const dirPath = "%s";
		const entries = $os.readDir(dirPath);
	});
}
	`, testDocumentCollectionDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{"$TEMP/*", testDocumentsDir},
			WritePaths: []string{"$TEMP/*"},
		},
	}

	_, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.Error(t, err)

	manager.PrintPluginPoolMetrics(opts.ID)

}

/////////////////////////////////////////////////////////////////////////////////////////////

// Test that the plugin can play a video and listen to events
func TestGojaPluginPlaybackEvents(t *testing.T) {
	payload := fmt.Sprintf(`
function init() {

	$ui.register((ctx) => {
		console.log("Testing Playback");

		const cancel = ctx.playback.registerEventListener("mySubscriber", (event) => {
			console.log("Event received", event)
		})

		ctx.playback.playUsingMediaPlayer("%s")

		ctx.setTimeout(() => {
			console.log("Cancelling event listener")
			cancel()
		}, 15000)
	});

}
	`, testVideoPath)

	playbackManager, _, err := getPlaybackManager(t)
	require.NoError(t, err)

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		PlaybackManager: playbackManager,
	})

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionPlayback,
		},
	}

	_, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(16 * time.Second)
}

// Tests that we can register hooks and the UI handler.
// Tests that the state updates correctly and effects run as expected.
// Tests that we can fetch data from an external source.
func TestGojaPluginUIAndHooks(t *testing.T) {
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
			const url = "%s"
			console.log("this is the start");

			const count = ctx.state(0)

			ctx.effect(async () => {
				console.log("running effect that takes 1s")
				ctx.setTimeout(() => {
					console.log("1s elapsed since first effect called")
				}, 1000)
				const [a, b, c, d, e, f] = await Promise.all([
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/1"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/2"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch("https://jsonplaceholder.typicode.com/todos/3"),
					ctx.fetch(url),
				])
				console.log("fetch results", a.json(), b.json(), c.json(), d.json(), e.json(), f.json())
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

	_, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

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

	time.Sleep(3 * time.Second)
}

func TestGojaPluginStore(t *testing.T) {
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

	_, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

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

	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
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
	})

	playbackManager.SetAnimeCollection(animeCollection)
	playbackManager.SetSettings(&playbackmanager.Settings{
		AutoPlayNextEpisode: false,
	})

	playbackManager.SetMediaPlayerRepository(mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
		Mpv:               mpv.New(logger, "", ""),
		Logger:            logger,
		Default:           "mpv",
		ContinuityManager: continuityManager,
	}))

	return playbackManager, animeCollection, nil
}
