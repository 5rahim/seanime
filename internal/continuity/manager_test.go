package continuity

import (
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/testutil"
	"seanime/internal/util/filecache"
	"strconv"
	"testing"
	"time"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func TestTrimWatchHistoryItemsRemovesOldestItem(t *testing.T) {
	manager, cacher := newHistoryTestManager(t)
	baseTime := time.Now().Add(-time.Hour)

	for mediaID := 1; mediaID <= MaxWatchHistoryItems+1; mediaID++ {
		err := cacher.Set(*manager.watchHistoryFileCacheBucket, strconv.Itoa(mediaID), &WatchHistoryItem{
			MediaId:       mediaID,
			EpisodeNumber: 1,
			CurrentTime:   10,
			Duration:      100,
			TimeAdded:     baseTime.Add(time.Duration(mediaID) * time.Minute),
			TimeUpdated:   baseTime.Add(time.Duration(mediaID) * time.Minute),
		})
		require.NoError(t, err)
	}

	require.NoError(t, manager.trimWatchHistoryItems())

	items := getAllHistoryItems(t, cacher, manager)
	require.Len(t, items, MaxWatchHistoryItems)
	require.NotContains(t, items, "1")
	require.Contains(t, items, strconv.Itoa(MaxWatchHistoryItems+1))
}

func TestUpdateWatchHistoryItemCreatesAndUpdatesExistingItem(t *testing.T) {
	manager, cacher := newHistoryTestManager(t)
	originalTime := time.Now().Add(-2 * time.Hour)

	err := cacher.Set(*manager.watchHistoryFileCacheBucket, "42", &WatchHistoryItem{
		Kind:          MediastreamKind,
		Filepath:      "/tmp/original.mkv",
		MediaId:       42,
		EpisodeNumber: 1,
		CurrentTime:   20,
		Duration:      100,
		TimeAdded:     originalTime,
		TimeUpdated:   originalTime,
	})
	require.NoError(t, err)

	err = manager.UpdateWatchHistoryItem(&UpdateWatchHistoryItemOptions{
		Kind:          OnlinestreamKind,
		MediaId:       42,
		EpisodeNumber: 2,
		CurrentTime:   30,
		Duration:      120,
		Filepath:      "/tmp/updated.mkv",
	})
	require.NoError(t, err)

	response := manager.GetWatchHistoryItem(42)
	require.True(t, response.Found)
	require.NotNil(t, response.Item)
	require.Equal(t, OnlinestreamKind, response.Item.Kind)
	require.Equal(t, 2, response.Item.EpisodeNumber)
	require.Equal(t, 30.0, response.Item.CurrentTime)
	require.Equal(t, 120.0, response.Item.Duration)
	require.True(t, response.Item.TimeAdded.Equal(originalTime))
	require.True(t, response.Item.TimeUpdated.After(originalTime))
	require.Equal(t, "/tmp/original.mkv", response.Item.Filepath)
}

func TestGetWatchHistoryItemAppliesCompletionThresholds(t *testing.T) {
	t.Run("returns item within resumable range", func(t *testing.T) {
		manager, cacher := newHistoryTestManager(t)
		seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
			MediaId:       10,
			EpisodeNumber: 1,
			CurrentTime:   50,
			Duration:      100,
		})

		response := manager.GetWatchHistoryItem(10)
		require.True(t, response.Found)
		require.NotNil(t, response.Item)
	})

	t.Run("hides nearly finished item without deleting it", func(t *testing.T) {
		manager, cacher := newHistoryTestManager(t)
		seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
			MediaId:       11,
			EpisodeNumber: 1,
			CurrentTime:   90,
			Duration:      100,
		})

		response := manager.GetWatchHistoryItem(11)
		require.False(t, response.Found)
		require.Nil(t, response.Item)
		require.Contains(t, manager.GetWatchHistory(), 11)
	})

	t.Run("hides barely started item without deleting it", func(t *testing.T) {
		manager, cacher := newHistoryTestManager(t)
		seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
			MediaId:       12,
			EpisodeNumber: 1,
			CurrentTime:   4,
			Duration:      100,
		})

		response := manager.GetWatchHistoryItem(12)
		require.False(t, response.Found)
		require.Nil(t, response.Item)

		items := getAllHistoryItems(t, cacher, manager)
		item, found := items["12"]
		require.True(t, found)
		require.Equal(t, 4.0, item.CurrentTime)
	})
}

func TestDeleteWatchHistoryItemRemovesStoredEntry(t *testing.T) {
	manager, cacher := newHistoryTestManager(t)
	seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
		MediaId:       20,
		EpisodeNumber: 1,
		CurrentTime:   20,
		Duration:      100,
	})

	require.NoError(t, manager.DeleteWatchHistoryItem(20))

	response := manager.GetWatchHistoryItem(20)
	require.False(t, response.Found)
	require.Nil(t, response.Item)
	require.NotContains(t, getAllHistoryItems(t, cacher, manager), "20")
}

func TestUpdateExternalPlayerEpisodeWatchHistoryItem(t *testing.T) {
	t.Run("does nothing when continuity is disabled", func(t *testing.T) {
		manager, cacher := newHistoryTestManager(t)
		manager.SetExternalPlayerEpisodeDetails(&ExternalPlayerEpisodeDetails{
			MediaId:       30,
			EpisodeNumber: 5,
			Filepath:      "/tmp/external-disabled.mkv",
		})

		manager.UpdateExternalPlayerEpisodeWatchHistoryItem(40, 100)
		require.Empty(t, getAllHistoryItems(t, cacher, manager))
	})

	t.Run("creates and updates item when enabled", func(t *testing.T) {
		manager, _ := newHistoryTestManager(t)
		manager.SetSettings(&Settings{WatchContinuityEnabled: true})
		manager.SetExternalPlayerEpisodeDetails(&ExternalPlayerEpisodeDetails{
			MediaId:       31,
			EpisodeNumber: 5,
			Filepath:      "/tmp/external.mkv",
		})

		manager.UpdateExternalPlayerEpisodeWatchHistoryItem(40, 100)

		response := manager.GetWatchHistoryItem(31)
		require.True(t, response.Found)
		require.NotNil(t, response.Item)
		require.Equal(t, ExternalPlayerKind, response.Item.Kind)
		require.Equal(t, "/tmp/external.mkv", response.Item.Filepath)
		require.Equal(t, 5, response.Item.EpisodeNumber)
		require.Equal(t, 40.0, response.Item.CurrentTime)

		manager.SetExternalPlayerEpisodeDetails(&ExternalPlayerEpisodeDetails{
			MediaId:       31,
			EpisodeNumber: 6,
			Filepath:      "/tmp/external.mkv",
		})
		manager.UpdateExternalPlayerEpisodeWatchHistoryItem(55, 120)

		updated := manager.GetWatchHistoryItem(31)
		require.True(t, updated.Found)
		require.Equal(t, 6, updated.Item.EpisodeNumber)
		require.Equal(t, 55.0, updated.Item.CurrentTime)
		require.Equal(t, 120.0, updated.Item.Duration)
	})
}

func TestGetExternalPlayerEpisodeWatchHistoryItemStream(t *testing.T) {
	manager, cacher := newHistoryTestManager(t)
	seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
		MediaId:       40,
		EpisodeNumber: 7,
		CurrentTime:   45,
		Duration:      100,
	})

	response := manager.GetExternalPlayerEpisodeWatchHistoryItem("ignored", true, 7, 40)
	require.False(t, response.Found)

	manager.SetSettings(&Settings{WatchContinuityEnabled: true})

	response = manager.GetExternalPlayerEpisodeWatchHistoryItem("ignored", true, 7, 40)
	require.True(t, response.Found)
	require.NotNil(t, response.Item)
	require.Equal(t, 40, response.Item.MediaId)

	mismatch := manager.GetExternalPlayerEpisodeWatchHistoryItem("ignored", true, 8, 40)
	require.False(t, mismatch.Found)
	require.Nil(t, mismatch.Item)

	missingIDs := manager.GetExternalPlayerEpisodeWatchHistoryItem("ignored", true, 0, 40)
	require.False(t, missingIDs.Found)
	require.Nil(t, missingIDs.Item)
}

func TestGetExternalPlayerEpisodeWatchHistoryItemLocalFile(t *testing.T) {
	manager, cacher := newHistoryTestManager(t)
	manager.SetSettings(&Settings{WatchContinuityEnabled: true})
	resetLocalFilesCache(t)

	localFiles := anime.NewTestLocalFiles(anime.TestLocalFileGroup{
		LibraryPath:      "/library",
		FilePathTemplate: "/library/show/episode-%ep.mkv",
		MediaID:          50,
		Episodes: []anime.TestLocalFileEpisode{{
			Episode:      3,
			AniDBEpisode: "3",
			Type:         anime.LocalFileTypeMain,
		}},
	})

	_, err := db_bridge.InsertLocalFiles(manager.db, localFiles)
	require.NoError(t, err)

	seedWatchHistoryItem(t, cacher, manager, &WatchHistoryItem{
		MediaId:       50,
		EpisodeNumber: 3,
		CurrentTime:   60,
		Duration:      120,
	})

	byPath := manager.GetExternalPlayerEpisodeWatchHistoryItem(localFiles[0].Path, false, 0, 0)
	require.True(t, byPath.Found)
	require.NotNil(t, byPath.Item)
	require.Equal(t, 50, byPath.Item.MediaId)

	byFilename := manager.GetExternalPlayerEpisodeWatchHistoryItem(localFiles[0].Name, false, 0, 0)
	require.True(t, byFilename.Found)
	require.NotNil(t, byFilename.Item)
}

func newHistoryTestManager(t *testing.T) (*Manager, *filecache.Cacher) {
	t.Helper()

	env := testutil.NewTestEnv(t)
	manager := NewManager(&NewManagerOptions{
		FileCacher: env.NewCacher("continuity"),
		Logger:     env.Logger(),
		Database:   env.NewDatabase(""),
	})

	require.NotNil(t, manager)
	return manager, manager.fileCacher
}

func seedWatchHistoryItem(t *testing.T, cacher *filecache.Cacher, manager *Manager, item *WatchHistoryItem) {
	t.Helper()

	if item.TimeAdded.IsZero() {
		item.TimeAdded = time.Now().Add(-time.Minute)
	}
	if item.TimeUpdated.IsZero() {
		item.TimeUpdated = item.TimeAdded
	}

	err := cacher.Set(*manager.watchHistoryFileCacheBucket, strconv.Itoa(item.MediaId), item)
	require.NoError(t, err)
}

func getAllHistoryItems(t *testing.T, cacher *filecache.Cacher, manager *Manager) map[string]*WatchHistoryItem {
	t.Helper()

	items, err := filecache.GetAll[*WatchHistoryItem](cacher, *manager.watchHistoryFileCacheBucket)
	require.NoError(t, err)
	return items
}

func resetLocalFilesCache(t *testing.T) {
	t.Helper()

	originalFiles := db_bridge.CurrLocalFiles
	originalID := db_bridge.CurrLocalFilesDbId
	db_bridge.CurrLocalFiles = mo.None[[]*anime.LocalFile]()
	db_bridge.CurrLocalFilesDbId = 0

	t.Cleanup(func() {
		db_bridge.CurrLocalFiles = originalFiles
		db_bridge.CurrLocalFilesDbId = originalID
	})
}
