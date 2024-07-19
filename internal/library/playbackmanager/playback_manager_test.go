package playbackmanager_test

import (
	"github.com/stretchr/testify/require"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func getPlaybackManager(t *testing.T) (*playbackmanager.PlaybackManager, *anilist.AnimeCollection, error) {

	logger := util.NewLogger()

	wsEventManager := events.NewMockWSEventManager(logger)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)

	if err != nil {
		t.Fatalf("error while creating database, %v", err)
	}

	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := platform.NewAnilistPlatform(anilistClient, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollection(true)
	require.NoError(t, err)

	return playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:         logger,
		WSEventManager: wsEventManager,
		Platform:       anilistPlatform,
		Database:       database,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		IsOffline: false,
	}), animeCollection, nil
}
