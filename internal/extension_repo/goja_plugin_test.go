package extension_repo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/platform"
	"seanime/internal/plugin"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
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
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)
	wsEventManager := events.NewMockWSEventManager(logger)
	anilistClientRef := util.NewRef[anilist.AnilistClient](anilist.NewMockAnilistClient())
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClientRef, extensionBankRef, logger, database).(*anilist_platform.AnilistPlatform)
	anilistPlatformRef := util.NewRef[platform.Platform](anilistPlatform)

	// Initialize hook manager if needed
	if opts.SetupHooks {
		hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
		hook.SetGlobalHookManager(hm)
	}

	manager := goja_runtime.NewManager(logger)

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		Database:           database,
		AnilistPlatformRef: anilistPlatformRef,
		WSEventManager:     wsEventManager,
		AnimeLibraryPaths:  &[]string{},
		PlaybackManager:    &playbackmanager.PlaybackManager{},
	})

	p, _, err := NewGojaPlugin(ext, opts.Language, logger, manager, wsEventManager, func(_ string) {})
	return p, logger, manager, anilistPlatform, wsEventManager, err
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginAnime(t *testing.T) {
	payload := `
	function init() {

		$ui.register(async (ctx) => {
			try {
				console.log("Fetching anime entry");
				ctx.anime.getAnimeMetadata("anilist", 21).then((metadata) => {
					console.log("Metadata", metadata)
				}).catch((e) => {
					console.error("Error fetching metadata", e)
				})
			} catch (e) {
				console.error("Error fetching metadata", e)
			}
			try {
				ctx.anime.getAnimeEntry(21).then((anime) => {
					console.log("Anime", anime)
				}).catch((e) => {
					console.error("Error fetching anime entry", e)
				})
			} catch (e) {
				console.error("Error fetching anime entry", e)
			}
		})
	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionAnilist,
			extension.PluginPermissionDatabase,
		},
	}
	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)

	metadataProvider := metadata_provider.NewProvider(&metadata_provider.NewProviderImplOptions{
		Logger:           logger,
		FileCacher:       lo.Must(filecache.NewCacher(t.TempDir())),
		Database:         database,
		ExtensionBankRef: util.NewRef(extension.NewUnifiedBank()),
	})
	metadataProviderRef := util.NewRef(metadataProvider)

	fillerManager := fillermanager.New(&fillermanager.NewFillerManagerOptions{
		Logger: logger,
		DB:     database,
	})

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		Database:            database,
		MetadataProviderRef: metadataProviderRef,
		FillerManager:       fillerManager,
	})

	_, logger, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(3 * time.Second)
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

/////////////////////////////////////////////////////////////////////////////////////////////

// Tests that we can register hooks and the UI handler.
// Tests that the state updates correctly and effects run as expected.
// Tests that we can fetch data from an external source.
func TestGojaPluginUIAndHooks(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		time.Sleep(2000 * time.Millisecond)
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
		_, err := anilistPlatform.GetAnime(t.Context(), 178022)
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

/////////////////////////////////////////////////////////////////////////////////////////////

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

	plugin, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	m, err := anilistPlatform.GetAnime(t.Context(), 178022)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)
	util.Spew(m.Synonyms)

	m, err = anilistPlatform.GetAnime(t.Context(), 177709)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	util.Spew(m.Title)

	value := plugin.store.Get("value")
	require.NotNil(t, value)

	manager.PrintPluginPoolMetrics(opts.ID)
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginJsonFieldNames(t *testing.T) {
	payload := `
	function init() {

		$app.onPreUpdateEntryProgress((e) => {
			console.log("pre update entry progress", e)

			$store.set("mediaId", e.mediaId);

			e.next();
		});

	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	plugin, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	err = anilistPlatform.UpdateEntryProgress(t.Context(), 178022, 1, new(1))
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}

	mediaId := plugin.store.Get("mediaId")
	require.NotNil(t, mediaId)

	manager.PrintPluginPoolMetrics(opts.ID)
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginAnilistCustomQuery(t *testing.T) {
	payload := `
	function init() {
		$ui.register((ctx) => {
		const token = $database.anilist.getToken()
			try {
				const res = $anilist.customQuery({ query:` + "`" + `
					query GetOnePiece {
						Media(id: 21) {
							title {
								romaji
								english
								native
								userPreferred
							}
							airingSchedule(perPage: 1, page: 1) {
								nodes {
									episode
									airingAt
								}
							}
							id
						}
					}
				` + "`" + `, variables: {}}, token);

				console.log("res", res)
			} catch (e) {
				console.error("Error fetching anime list", e);
			}
		});
	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionAnilist,
			extension.PluginPermissionAnilistToken,
			extension.PluginPermissionDatabase,
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	_ = plugin

	manager.PrintPluginPoolMetrics(opts.ID)
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginAnilistListAnime(t *testing.T) {
	payload := `
	function init() {

		$ui.register((ctx) => {
		
		try {
			const res = $anilist.listRecentAnime(1, 15, undefined, undefined, undefined)
			console.log("res", res)
		} catch (e) {
			console.error("Error fetching anime list", e)
		}

        })
	}
	`
	opts := DefaultTestPluginOptions()
	opts.Payload = payload

	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionAnilist,
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	_ = plugin

	manager.PrintPluginPoolMetrics(opts.ID)
}

/////////////////////////////////////////////////////////////////////////////////////////////

func TestGojaPluginStorage(t *testing.T) {
	payload := `
	function init() {

		$app.onGetAnime((e) => {

			if ($storage.get("foo") !== "qux") {
				throw new Error("foo should be qux")
			}

			$storage.set("foo", "anime")
			console.log("foo", $storage.get("foo"))
			$store.set("expectedValue4", "anime")

			e.next();
		});

		$ui.register((ctx) => {
			
			// Test 1: Basic set/get
			$storage.set("foo", "bar")
			console.log("foo", $storage.get("foo"))
			$store.set("expectedValue1", "bar")

			// Test 2: Nested keys, setting parent invalidates children in cache
			$storage.set("user.settings.theme", "light")
			$storage.set("user.settings.lang", "en")
			console.log("user.settings.theme", $storage.get("user.settings.theme"))
			$store.set("nestedTheme1", $storage.get("user.settings.theme"))
			
			// Now set the parent, should invalidate cached children
			$storage.set("user.settings", { theme: "dark", lang: "ja", notifications: true })
			console.log("user.settings", $storage.get("user.settings"))
			
			// Child values should be fresh from DB, not cached
			const themeAfterParentSet = $storage.get("user.settings.theme")
			console.log("user.settings.theme after parent set", themeAfterParentSet)
			$store.set("nestedTheme2", themeAfterParentSet)
			
			// Test 3: Keys and Has methods
			const allKeys = $storage.keys()
			console.log("all keys", allKeys)
			$store.set("hasUserSettings", $storage.has("user.settings"))
			$store.set("hasUserTheme", $storage.has("user.settings.theme"))
			
			// Test 4: Delete nested key
			$storage.remove("user.settings.notifications")
			$store.set("hasNotificationsAfterDelete", $storage.has("user.settings.notifications"))
			
			// Test 5: Sequential updates
			ctx.setTimeout(() => {
				console.log("foo", $storage.get("foo"))
				$storage.set("foo", "baz")
				console.log("foo", $storage.get("foo"))
				$store.set("expectedValue2", "baz")
			}, 1000)

			ctx.setTimeout(() => {
				console.log("foo", $storage.get("foo"))
				$storage.set("foo", "qux")
				console.log("foo", $storage.get("foo"))
				$store.set("expectedValue3", "qux")
				
				// Test that nested access still works
				const finalTheme = $storage.get("user.settings.theme")
				$store.set("finalTheme", finalTheme)
				
				// Test clear operation
				const keysBefore = $storage.keys()
				$store.set("keysBeforeClear", keysBefore.length)
			}, 1500)
			
		})

	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionDatabase,
			extension.PluginPermissionStorage,
		},
	}

	plugin, _, manager, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	_ = plugin

	manager.PrintPluginPoolMetrics(opts.ID)

	time.Sleep(2 * time.Second)

	_, err = anilistPlatform.GetAnime(t.Context(), 178022)
	require.NoError(t, err)

	// Test basic sequential updates
	expectedValue1 := plugin.store.Get("expectedValue1")
	require.Equal(t, "bar", expectedValue1, "Initial value should be 'bar'")

	expectedValue2 := plugin.store.Get("expectedValue2")
	require.Equal(t, "baz", expectedValue2, "Second value should be 'baz'")

	expectedValue3 := plugin.store.Get("expectedValue3")
	require.Equal(t, "qux", expectedValue3, "Third value should be 'qux'")

	expectedValue4 := plugin.store.Get("expectedValue4")
	require.Equal(t, "anime", expectedValue4, "Final value should be 'anime'")

	// Test nested key cache invalidation
	nestedTheme1 := plugin.store.Get("nestedTheme1")
	require.Equal(t, "light", nestedTheme1, "Initial nested theme should be 'light'")

	nestedTheme2 := plugin.store.Get("nestedTheme2")
	require.Equal(t, "dark", nestedTheme2, "Nested theme after parent set should be 'dark' (cache should be invalidated)")

	finalTheme := plugin.store.Get("finalTheme")
	require.Equal(t, "dark", finalTheme, "Final theme should still be 'dark'")

	// Test has/keys methods
	hasUserSettings := plugin.store.Get("hasUserSettings")
	require.Equal(t, true, hasUserSettings, "user.settings should exist")

	hasUserTheme := plugin.store.Get("hasUserTheme")
	require.Equal(t, true, hasUserTheme, "user.settings.theme should exist")

	hasNotificationsAfterDelete := plugin.store.Get("hasNotificationsAfterDelete")
	require.Equal(t, false, hasNotificationsAfterDelete, "user.settings.notifications should not exist after delete")

	keysBeforeClear := plugin.store.Get("keysBeforeClear")
	require.NotNil(t, keysBeforeClear, "Should have keys before clear")
	require.Greater(t, keysBeforeClear.(int64), int64(0), "Should have at least one key")

}

func TestGojaPluginStorage2(t *testing.T) {
	payload := `
	function init() {

		$app.onGetAnime((e) => {

			console.log("hook", $storage.get("object"))
			e.next();
		});

		$ui.register((ctx) => {
			
			try {
				$storage.set("object", { foo: "bar" })
				const object1 = $storage.get("object")
				console.log("object", object1)
				$store.set("object", object1)
				$storage.set("object", { foo: "bar", baz: { "1": { id: 1 } } })
				const object2 = $storage.get("object")
				console.log("object", object2)
				$store.set("object", object2)

				// Runs after first hook trigger
				ctx.setTimeout(() => {
					$storage.set("object", { foo: "bar", baz: { "1": { id: 1 }, "2": { id: 2 } } })
				}, 1000)
			} catch (e) {
				console.error("Test failed", e)
			}
			
		})

	}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionDatabase,
			extension.PluginPermissionStorage,
		},
	}

	p, _, _, anilistPlatform, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	_ = anilistPlatform

	//manager.PrintPluginPoolMetrics(opts.ID)

	//time.Sleep(2 * time.Second)
	//
	//_, err = anilistPlatform.GetAnime(t.Context(), 178022)
	//require.NoError(t, err)

	// Test basic sequential updates
	object1 := p.store.Get("object")
	require.NotNil(t, object1)
	anilistPlatform.GetAnime(t.Context(), 178022)
	time.Sleep(1500 * time.Millisecond)
	anilistPlatform.GetAnime(t.Context(), 178022)

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
	anilistClientRef := util.NewRef(anilistClient)
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClientRef, util.NewRef(extension.NewUnifiedBank()), logger, database)
	animeCollection, err := anilistPlatform.GetAnimeCollection(t.Context(), true)
	metadataProvider := metadata_provider.GetFakeProvider(t, database)
	require.NoError(t, err)
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: filecacher,
		Logger:     logger,
		Database:   database,
	})
	anilistPlatformRef := util.NewRef[platform.Platform](anilistPlatform)
	metadataProviderRef := util.NewRef(metadataProvider)

	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager:      wsEventManager,
		Logger:              logger,
		PlatformRef:         anilistPlatformRef,
		MetadataProviderRef: metadataProviderRef,
		Database:            database,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		DiscordPresence:   nil,
		IsOfflineRef:      util.NewRef(false),
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
