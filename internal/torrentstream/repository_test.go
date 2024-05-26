package torrentstream

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/playbackmanager"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/offline"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"testing"
)

func TestTorrentstream(t *testing.T) {
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
	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	animeCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)
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
		WSEventManager:       wsEventManager,
		Logger:               logger,
		AnilistClientWrapper: anilistClientWrapper,
		AnilistCollection:    animeCollection,
		Database:             database,
		RefreshAnilistCollectionFunc: func() {

		},
		DiscordPresence: nil,
		IsOffline:       false,
		OfflineHub:      offline.NewMockHub(),
	})

	playbackManager.SetAnilistCollection(animeCollection)
	playbackManager.SetMediaPlayerRepository(mediaPlayerRepo)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:                logger,
		AnizipCache:           anizip.NewCache(),
		BaseMediaCache:        anilist.NewBaseMediaCache(),
		AnimeCollection:       animeCollection,
		AnilistClientWrapper:  anilistClientWrapper,
		AnimeToshoSearchCache: animetosho.NewSearchCache(),
		NyaaSearchCache:       nyaa.NewSearchCache(),
		MetadataProvider: metadata.NewProvider(&metadata.NewProviderOptions{
			Logger:     logger,
			FileCacher: filecacher,
		}),
		PlaybackManager: playbackManager,
		WSEventManager:  wsEventManager,
	})
	repo.SetMediaPlayerRepository(mediaPlayerRepo)
	repo.SetAnimeCollection(animeCollection)
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
