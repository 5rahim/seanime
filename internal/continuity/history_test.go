package continuity

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"seanime/internal/database/db"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func TestHistoryItems(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	logger := util.NewLogger()

	tempDir := t.TempDir()
	t.Log(tempDir)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)

	cacher, err := filecache.NewCacher(filepath.Join(tempDir, "cache"))
	require.NoError(t, err)

	manager := NewManager(&NewManagerOptions{
		FileCacher: cacher,
		Logger:     logger,
		Database:   database,
	})
	require.NotNil(t, manager)

	var mediaIds = make([]int, MaxWatchHistoryItems+1)
	for i := 0; i < MaxWatchHistoryItems+1; i++ {
		mediaIds[i] = i + 1
	}

	// Add items to the history
	for _, mediaId := range mediaIds {
		err = manager.UpdateWatchHistoryItem(&UpdateWatchHistoryItemOptions{
			MediaId:       mediaId,
			EpisodeNumber: 1,
			CurrentTime:   10,
			Duration:      100,
		})
		require.NoError(t, err)
	}

	// Check if the oldest item was removed
	items, err := filecache.GetAll[WatchHistoryItem](cacher, *manager.watchHistoryFileCacheBucket)
	require.NoError(t, err)

	require.Len(t, items, MaxWatchHistoryItems)

	// Update an item
	err = manager.UpdateWatchHistoryItem(&UpdateWatchHistoryItemOptions{
		MediaId:       mediaIds[0], // 1
		EpisodeNumber: 2,
		CurrentTime:   30,
		Duration:      100,
	})
	require.NoError(t, err)

	// Check if the item was updated
	items, err = filecache.GetAll[WatchHistoryItem](cacher, *manager.watchHistoryFileCacheBucket)
	require.NoError(t, err)

	require.Len(t, items, MaxWatchHistoryItems)

	item, found := items["1"]
	require.True(t, found)

	require.Equal(t, 2, item.EpisodeNumber)
	require.Equal(t, 30., item.CurrentTime)
	require.Equal(t, 100., item.Duration)

}
