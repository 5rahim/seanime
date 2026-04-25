package debrid_client

import (
	"context"
	"testing"

	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/hook_resolver"
	"seanime/internal/testutil"
	"seanime/internal/util"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type fakeDebridProvider struct {
	addTorrent  func(opts debrid.AddTorrentOptions) (string, error)
	downloadURL func(opts debrid.DownloadTorrentOptions) (string, error)
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
	if f.downloadURL != nil {
		return f.downloadURL(opts)
	}
	return "", nil
}

func (f *fakeDebridProvider) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {
	return map[string]debrid.TorrentItemInstantAvailability{}
}

func (f *fakeDebridProvider) GetTorrent(id string) (*debrid.TorrentItem, error) {
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
		require.Equal(t, 77, event.MediaID)
		return event.Next()
	})

	destination := t.TempDir()
	torrentItemID, err := repo.AddAndQueueTorrent(debrid.AddTorrentOptions{MagnetLink: "magnet:?xt=urn:btih:abc"}, destination, 77)
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
		MediaId:       77,
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
