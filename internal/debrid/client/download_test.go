package debrid_client

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestTorBoxDownload(t *testing.T) {
	test_utils.InitTestProvider(t)

	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)

	repo := GetMockRepository(t, database)

	err = repo.InitializeProvider(&models.DebridSettings{
		Enabled:  true,
		Provider: "torbox",
		ApiKey:   test_utils.ConfigData.Provider.TorBoxApiKey,
	})
	require.NoError(t, err)

	tempDestinationDir := t.TempDir()

	fmt.Println(tempDestinationDir)

	//
	// Test download
	//
	torrentItemId := "116389"

	err = database.InsertDebridTorrentItem(&models.DebridTorrentItem{
		TorrentItemID: torrentItemId,
		Destination:   tempDestinationDir,
		Provider:      "torbox",
		MediaId:       0, // Not yet used
	})
	require.NoError(t, err)

	// Get the provider
	provider, err := repo.GetProvider()
	require.NoError(t, err)

	// Get the torrents from the provider
	torrentItems, err := provider.GetTorrents()
	require.NoError(t, err)

	// Get the torrent item from the database
	dbTorrentItem, err := database.GetDebridTorrentItemByTorrentItemId(torrentItemId)

	// Select the torrent item from the provider
	var torrentItem *debrid.TorrentItem
	for _, item := range torrentItems {
		if item.ID == dbTorrentItem.TorrentItemID {
			torrentItem = item
		}
	}
	require.NotNil(t, torrentItem)

	// Check if the torrent is ready
	require.Truef(t, torrentItem.IsReady, "Torrent is not ready")

	// Remove the item from the database
	err = database.DeleteDebridTorrentItemByDbId(dbTorrentItem.ID)
	require.NoError(t, err)

	// Download the torrent
	err = repo.downloadTorrentItem(dbTorrentItem.TorrentItemID, torrentItem.Name, dbTorrentItem.Destination)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 500)

	// Wait for the download to finish
loop:
	for {
		select {
		case <-time.After(time.Second * 1):
			isEmpty := true
			repo.ctxMap.Range(func(key string, value context.CancelFunc) bool {
				isEmpty = false
				return true
			})
			if isEmpty {
				break loop
			}
		}
	}

	// Check if the file exists
	entries, err := os.ReadDir(tempDestinationDir)
	require.NoError(t, err)

	fmt.Println("=== Downloaded files ===")

	for _, entry := range entries {
		util.Spew(entry.Name())
	}

	require.NotEmpty(t, entries)
}
