package torrentstream

import (
	"fmt"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/offline"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func TestTorrentstream(t *testing.T) {
	//t.Skip()
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist(), test_utils.MediaPlayer(), test_utils.Torrentstream())

	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatalf("error while creating database, %v", err)
	}
	filecacher, err := filecache.NewCacher(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wsEventManager := events.NewMockWSEventManager(logger)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollection(false)
	if err != nil {
		t.Fatal(err)
	}

	mediaId := 163132 // Horimiya: piece

	anilist.TestModifyAnimeCollectionEntry(animeCollection, mediaId, anilist.TestModifyAnimeCollectionEntryInput{
		Status:   lo.ToPtr(anilist.MediaListStatusCurrent),
		Progress: lo.ToPtr(5), // Mock progress
	})

	mediaPlayerRepo := mediaplayer.NewTestRepository(t, "mpv")

	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager: wsEventManager,
		Logger:         logger,
		Platform:       anilistPlatform,
		Database:       database,
		RefreshAnimeCollectionFunc: func() {

		},
		DiscordPresence: nil,
		IsOffline:       false,
		OfflineHub:      offline.NewMockHub(),
	})

	playbackManager.SetMediaPlayerRepository(mediaPlayerRepo)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:             logger,
		AnizipCache:        anizip.NewCache(),
		BaseAnimeCache:     anilist.NewBaseAnimeCache(),
		CompleteAnimeCache: anilist.NewCompleteAnimeCache(),
		Platform:           anilistPlatform,
		MetadataProvider: metadata.NewProvider(&metadata.NewProviderOptions{
			Logger:     logger,
			FileCacher: filecacher,
		}),
		PlaybackManager: playbackManager,
		WSEventManager:  wsEventManager,
	})
	repo.SetMediaPlayerRepository(mediaPlayerRepo)
	defer repo.Shutdown()

	fmt.Println(repo.GetDownloadDir())

	err = repo.InitModules(&models.TorrentstreamSettings{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Enabled:             true,
		AutoSelect:          true,
		PreferredResolution: "1080",
		DisableIPV6:         false,
		DownloadDir:         "",
		AddToLibrary:        false,
		TorrentClientPort:   42069,
		StreamingServerHost: "0.0.0.0",
		StreamingServerPort: 43214,
	}, "0.0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = repo.getMediaInfo(mediaId)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.StartStream(&StartStreamOptions{
		MediaId:       mediaId,
		EpisodeNumber: 8,
		AniDBEpisode:  "8",
		AutoSelect:    true,
		Torrent:       nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	select {}

}
