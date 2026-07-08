package debrid_client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	databasepkg "seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/hook_resolver"
	"seanime/internal/testutil"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"testing"
	"time"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type fakeDebridProvider struct {
	addTorrent            func(opts debrid.AddTorrentOptions) (string, error)
	getTorrent            func(id string) (*debrid.TorrentItem, error)
	getTorrentDownloadUrl func(opts debrid.DownloadTorrentOptions) (string, error)
}

func (f *fakeDebridProvider) GetSettings() debrid.Settings {
	return debrid.Settings{ID: "fake-provider", Name: "Fake Provider"}
}

func (f *fakeDebridProvider) Authenticate(apiKey string) error {
	return nil
}

func (f *fakeDebridProvider) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {
	if f.addTorrent != nil {
		return f.addTorrent(opts)
	}
	return "fake-torrent-id", nil
}

func (f *fakeDebridProvider) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (string, error) {
	return "", nil
}

func (f *fakeDebridProvider) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (string, error) {
	if f.getTorrentDownloadUrl != nil {
		return f.getTorrentDownloadUrl(opts)
	}
	return "", nil
}

func (f *fakeDebridProvider) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {
	return map[string]debrid.TorrentItemInstantAvailability{}
}

func (f *fakeDebridProvider) GetTorrent(id string) (*debrid.TorrentItem, error) {
	if f.getTorrent != nil {
		return f.getTorrent(id)
	}
	return nil, nil
}

func (f *fakeDebridProvider) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error) {
	return nil, nil
}

func (f *fakeDebridProvider) GetTorrents() ([]*debrid.TorrentItem, error) {
	return nil, nil
}

func (f *fakeDebridProvider) DeleteTorrent(id string) error {
	return nil
}

func TestInitializeProviderRequiresDummyFeatureFlag(t *testing.T) {
	databasepkg.CurrentDummyDebridSettings = nil
	t.Cleanup(func() {
		databasepkg.CurrentDummyDebridSettings = nil
	})

	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	fixture := env.MustWriteFixtureFile("fixture.mkv", []byte("fixture"))
	_, err := database.UpsertDummyDebridSettings(models.NewDefaultDummyDebridSettings(fixture))
	require.NoError(t, err)

	disabledRepo := &Repository{
		provider:           mo.None[debrid.Provider](),
		logger:             logger,
		db:                 database,
		settings:           &models.DebridSettings{},
		ctxMap:             result.NewMap[string, context.CancelFunc](),
		dummyDebridEnabled: false,
	}
	err = disabledRepo.InitializeProvider(&models.DebridSettings{Enabled: true, Provider: "dummy"})
	require.NoError(t, err)
	require.False(t, disabledRepo.HasProvider())

	enabledRepo := &Repository{
		provider:           mo.None[debrid.Provider](),
		logger:             logger,
		db:                 database,
		settings:           &models.DebridSettings{},
		ctxMap:             result.NewMap[string, context.CancelFunc](),
		dummyDebridEnabled: true,
	}
	t.Cleanup(func() {
		if enabledRepo.downloadLoopCancelFunc != nil {
			enabledRepo.downloadLoopCancelFunc()
		}
		enabledRepo.closeProvider()
	})

	err = enabledRepo.InitializeProvider(&models.DebridSettings{Enabled: true, Provider: "dummy"})
	require.NoError(t, err)
	require.True(t, enabledRepo.HasProvider())
	provider, err := enabledRepo.GetProvider()
	require.NoError(t, err)
	require.Equal(t, "dummy", provider.GetSettings().ID)
}

func TestAddAndQueueTorrentUsesHookOverrideAndQueuesItem(t *testing.T) {
	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	ws := events.NewMockWSEventManager(logger)

	addCalls := 0
	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{addTorrent: func(opts debrid.AddTorrentOptions) (string, error) {
			addCalls++
			return "provider-id", nil
		}}),
		logger:         logger,
		db:             database,
		wsEventManager: ws,
	}

	oldManager := hook.GlobalHookManager
	manager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(manager)
	t.Cleanup(func() {
		hook.SetGlobalHookManager(oldManager)
	})

	postCalled := false
	manager.OnDebridAddTorrentRequested().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*DebridAddTorrentRequestedEvent)
		event.PreventDefault()
		event.TorrentItemID = "hook-id"
		return event.Next()
	})
	manager.OnDebridAddTorrent().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*DebridAddTorrentEvent)
		postCalled = true
		require.Equal(t, "hook-id", event.TorrentItemID)
		require.Equal(t, 21, event.MediaID)
		return event.Next()
	})

	destination := t.TempDir()
	torrentItemID, err := repo.AddAndQueueTorrent(debrid.AddTorrentOptions{MagnetLink: "magnet:?xt=urn:btih:abc"}, destination, 21)
	require.NoError(t, err)
	require.Equal(t, "hook-id", torrentItemID)
	require.Equal(t, 0, addCalls)
	require.True(t, postCalled)

	items, err := database.GetDebridTorrentItems()
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, &models.DebridTorrentItem{
		BaseModel:     items[0].BaseModel,
		TorrentItemID: "hook-id",
		Destination:   destination,
		Provider:      "fake-provider",
		MediaId:       21,
	}, items[0])
}

func TestDownloadLifecycleHooksTrigger(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	repo := &Repository{
		logger:         logger,
		wsEventManager: ws,
	}

	oldManager := hook.GlobalHookManager
	manager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(manager)
	t.Cleanup(func() {
		hook.SetGlobalHookManager(oldManager)
	})

	started := false
	completed := false
	manager.OnDebridLocalDownloadStarted().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*DebridLocalDownloadStartedEvent)
		started = true
		require.Equal(t, "torrent-1", event.TorrentItemID)
		require.Equal(t, "test-torrent", event.TorrentName)
		require.Equal(t, "/library/anime", event.Destination)
		require.Equal(t, "https://example.com/file", event.DownloadUrl)
		return event.Next()
	})
	manager.OnDebridLocalDownloadCompleted().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*DebridLocalDownloadCompletedEvent)
		completed = true
		require.Equal(t, "torrent-1", event.TorrentItemID)
		require.Equal(t, "test-torrent", event.TorrentName)
		require.Equal(t, "/library/anime", event.Destination)
		return event.Next()
	})

	err := repo.sendDownloadStartedEvent("torrent-1", "test-torrent", "/library/anime", "https://example.com/file")
	require.NoError(t, err)
	repo.sendDownloadCompletedEvent("torrent-1", "test-torrent", "/library/anime")

	require.True(t, started)
	require.True(t, completed)
}

func TestDownlaoded_KeepItemOnDownloadUrlFailure(t *testing.T) {
	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()

	// scenario: the torrent looks ready before requestdl works
	require.NoError(t, database.InsertDebridTorrentItem(&models.DebridTorrentItem{
		TorrentItemID: "torrent-1",
		Destination:   destination,
		Provider:      "fake-provider",
		MediaId:       21,
	}))

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{
			getTorrent: func(id string) (*debrid.TorrentItem, error) {
				return &debrid.TorrentItem{ID: id, Name: "test", Hash: "ABC", IsReady: true}, nil
			},
			getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
				return "", errors.New("not ready yet")
			},
		}),
		logger:         logger,
		db:             database,
		wsEventManager: ws,
	}

	repo.processQueuedDownloads(repo.provider.MustGet())

	items, err := database.GetDebridTorrentItems()
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "torrent-1", items[0].TorrentItemID)
}

func TestDownload_removeItemAfterDownloadCompletes(t *testing.T) {
	initTestDownload(t, 1, func(int) time.Duration { return 0 })
	initTestDownloadManager(t)

	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write([]byte("not a zip"))
	}))
	t.Cleanup(server.Close)

	// failed local downloads should stay queued for a fresh URL later
	require.NoError(t, database.InsertDebridTorrentItem(&models.DebridTorrentItem{
		TorrentItemID: "torrent-1",
		Destination:   destination,
		Provider:      "fake-provider",
		MediaId:       21,
	}))
	require.NoError(t, database.InsertAutoDownloaderItem(&models.AutoDownloaderItem{
		RuleID:      1,
		MediaID:     21,
		Episode:     4,
		Hash:        "abc",
		TorrentName: "test",
		Downloaded:  false,
	}))

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{
			getTorrent: func(id string) (*debrid.TorrentItem, error) {
				return &debrid.TorrentItem{ID: id, Name: "test", Hash: "ABC", IsReady: true}, nil
			},
			getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
				return server.URL + "/bad.zip", nil
			},
		}),
		logger:         logger,
		db:             database,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}

	repo.processQueuedDownloads(repo.provider.MustGet())
	require.Eventually(t, func() bool {
		return hasDebridDownloadStatus(ws, "cancelled")
	}, time.Second, 10*time.Millisecond)

	items, err := database.GetDebridTorrentItems()
	require.NoError(t, err)
	require.Len(t, items, 1)
	queued, err := database.GetAutoDownloaderItem(1)
	require.NoError(t, err)
	require.False(t, queued.Downloaded)
}

func TestDownloadedItemsAreRemovedFromQueue(t *testing.T) {
	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()

	mux := http.NewServeMux()
	mux.HandleFunc("/episode.mkv", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/x-matroska")
		_, _ = w.Write([]byte("episode"))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	downloadURL := server.URL + "/episode.mkv"

	// once the file is moved, the debrid retry row is done
	require.NoError(t, database.InsertDebridTorrentItem(&models.DebridTorrentItem{
		TorrentItemID: "torrent-1",
		Destination:   destination,
		Provider:      "fake-provider",
		MediaId:       21,
	}))
	require.NoError(t, database.InsertAutoDownloaderItem(&models.AutoDownloaderItem{
		RuleID:      1,
		MediaID:     21,
		Episode:     4,
		Hash:        "abc",
		TorrentName: "test",
		Downloaded:  false,
	}))

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{
			getTorrent: func(id string) (*debrid.TorrentItem, error) {
				return &debrid.TorrentItem{ID: id, Name: "test", Hash: "ABC", IsReady: true}, nil
			},
			getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
				return downloadURL, nil
			},
		}),
		logger:         logger,
		db:             database,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}

	repo.processQueuedDownloads(repo.provider.MustGet())

	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(destination, "episode.mkv"))
		return err == nil
	}, time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool {
		items, err := database.GetDebridTorrentItems()
		return err == nil && len(items) == 0
	}, time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool {
		queued, err := database.GetAutoDownloaderItem(1)
		return err == nil && queued.Downloaded
	}, time.Second, 10*time.Millisecond)
}
