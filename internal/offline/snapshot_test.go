package offline

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/metadata"
	db2 "github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"path/filepath"
	"testing"
)

func getHub(t *testing.T) *Hub {
	logger := util.NewLogger()
	fileCacher, err := filecache.NewCacher(filepath.Join(test_utils.ConfigData.Path.DataDir, "cache"))
	if err != nil {
		t.Fatal(err)
	}

	db, err := db2.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatal(err)
	}

	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := platform.NewAnilistPlatform(anilistClient, logger)

	metadataProvider := metadata.NewProvider(&metadata.NewProviderOptions{
		Logger:     logger,
		FileCacher: fileCacher,
	})

	// Manga Repository
	mangaRepository := manga.NewRepository(&manga.NewRepositoryOptions{
		Logger:         logger,
		FileCacher:     fileCacher,
		BackupDir:      "",
		ServerURI:      "",
		WsEventManager: events.NewMockWSEventManager(logger),
		DownloadDir:    filepath.Join(test_utils.ConfigData.Path.DataDir, "manga"),
	})

	offlineHub := NewHub(&NewHubOptions{
		Platform:         anilistPlatform,
		WSEventManager:   events.NewMockWSEventManager(logger),
		MetadataProvider: metadataProvider,
		MangaRepository:  mangaRepository,
		Database:         db,
		FileCacher:       fileCacher,
		Logger:           logger,
		OfflineDir:       filepath.Join(test_utils.ConfigData.Path.DataDir, "offline"),
		AssetDir:         filepath.Join(test_utils.ConfigData.Path.DataDir, "offline", "assets"),
		IsOffline:        false,
	})

	return offlineHub
}

func TestSnapshot(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	offlineHub := getHub(t)

	// Test
	err := offlineHub.CreateSnapshot(&NewSnapshotOptions{
		AnimeToDownload:  []int{153518},
		DownloadAssetsOf: []int{153518, 101517, 144946},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get snapshot
	snapshot, err := offlineHub.GetLatestSnapshot(false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Snapshot ID: %+v", snapshot.DbId)
	t.Logf(" User: %s", snapshot.User.Viewer.Name)
	t.Logf(" Anime Entries: %d", len(snapshot.Entries.AnimeEntries))
	for _, entry := range snapshot.Entries.AnimeEntries {
		t.Logf("  %s", entry.Media.GetPreferredTitle())
		t.Logf("    %d episodes", len(entry.Episodes))
		t.Logf("    hasDownloadedAssets %t", entry.DownloadedAssets)
		t.Logf("")
	}
	t.Logf(" Manga Entries: %d", len(snapshot.Entries.MangaEntries))
	for _, entry := range snapshot.Entries.MangaEntries {
		t.Logf("    %s", entry.Media.GetPreferredTitle())
		t.Logf("       %d chapter containers", len(entry.ChapterContainers))
		t.Logf("       hasDownloadedAssets %t", entry.DownloadedAssets)
		t.Logf("")
	}

}

func TestSnapshot_GetLatestSnapshot(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	offlineHub := getHub(t)

	snapshot, err := offlineHub.GetLatestSnapshot(false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Snapshot ID: %+v", snapshot.DbId)
	t.Logf(" User: %s", snapshot.User.Viewer.Name)
	t.Logf(" Anime Entries: %d", len(snapshot.Entries.AnimeEntries))
	for _, entry := range snapshot.Entries.AnimeEntries {
		t.Logf("    %s", entry.Media.GetPreferredTitle())
		t.Logf("       %d episodes", len(entry.Episodes))
		t.Logf("       hasDownloadedAssets %t", entry.DownloadedAssets)
		t.Logf("")
	}
	t.Logf(" Manga Entries: %d", len(snapshot.Entries.MangaEntries))
	for _, entry := range snapshot.Entries.MangaEntries {
		t.Logf("    %s", entry.Media.GetPreferredTitle())
		t.Logf("       %d chapter containers", len(entry.ChapterContainers))
		t.Logf("       hasDownloadedAssets %t", entry.DownloadedAssets)
		t.Logf("")
	}

}
