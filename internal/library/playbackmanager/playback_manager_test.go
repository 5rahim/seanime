package playbackmanager_test

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/playbackmanager"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func getPlaybackManager(t *testing.T) (*playbackmanager.PlaybackManager, anilist.AnilistClient, *anilist.AnimeCollection, error) {

	logger := util.NewLogger()

	wsEventManager := events.NewMockWSEventManager(logger)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)

	if err != nil {
		t.Fatalf("error while creating database, %v", err)
	}

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anilistPlatform := platform.NewAnilistPlatform(anilistClientWrapper, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollection(false)
	if err != nil {
		return nil, nil, nil, err
	}

	return playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:         logger,
		WSEventManager: wsEventManager,
		Platform:       anilistPlatform,
		Database:       database,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		IsOffline: false,
	}), anilistClientWrapper, animeCollection, nil
}
