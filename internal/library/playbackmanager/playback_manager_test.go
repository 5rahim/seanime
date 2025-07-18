package playbackmanager_test

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"

	"github.com/stretchr/testify/require"
)

func getPlaybackManager(t *testing.T) (*playbackmanager.PlaybackManager, *anilist.AnimeCollection, error) {

	logger := util.NewLogger()

	wsEventManager := events.NewMockWSEventManager(logger)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)

	if err != nil {
		t.Fatalf("error while creating database, %v", err)
	}

	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollection(t.Context(), true)
	metadataProvider := metadata.GetMockProvider(t)
	require.NoError(t, err)
	continuityManager := continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: filecacher,
		Logger:     logger,
		Database:   database,
	})

	return playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager:   wsEventManager,
		Logger:           logger,
		Platform:         anilistPlatform,
		MetadataProvider: metadataProvider,
		Database:         database,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		DiscordPresence:   nil,
		IsOffline:         &[]bool{false}[0],
		ContinuityManager: continuityManager,
	}), animeCollection, nil
}
