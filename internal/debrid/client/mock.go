package debrid_client

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"testing"
)

func GetMockRepository(t *testing.T, db *db.Database) *Repository {
	logger := util.NewLogger()
	wsEventManager := events.NewWSEventManager(logger)
	anilistClient := anilist.TestGetMockAnilistClient()
	platform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	metadataProvider := metadata.GetMockProvider(t)
	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager:   wsEventManager,
		Logger:           logger,
		Platform:         platform,
		MetadataProvider: metadataProvider,
		Database:         db,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		DiscordPresence:   nil,
		IsOffline:         &[]bool{false}[0],
		ContinuityManager: continuity.GetMockManager(t, db),
	})

	r := NewRepository(&NewRepositoryOptions{
		Logger:           logger,
		WSEventManager:   wsEventManager,
		Database:         db,
		MetadataProvider: metadataProvider,
		Platform:         platform,
		PlaybackManager:  playbackManager,
	})

	return r
}
