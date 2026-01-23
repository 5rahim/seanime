package debrid_client

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"testing"
)

func GetMockRepository(t *testing.T, db *db.Database) *Repository {
	logger := util.NewLogger()
	wsEventManager := events.NewWSEventManager(logger)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistClientRef := util.NewRef(anilistClient)
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	platform := anilist_platform.NewAnilistPlatform(anilistClientRef, extensionBankRef, logger, db)
	metadataProvider := metadata_provider.GetFakeProvider(t, db)
	platformRef := util.NewRef(platform)
	metadataProviderRef := util.NewRef(metadataProvider)
	playbackManager := playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		WSEventManager:      wsEventManager,
		Logger:              logger,
		PlatformRef:         platformRef,
		MetadataProviderRef: metadataProviderRef,
		Database:            db,
		RefreshAnimeCollectionFunc: func() {
			// Do nothing
		},
		DiscordPresence:   nil,
		IsOfflineRef:      util.NewRef(false),
		ContinuityManager: continuity.GetMockManager(t, db),
	})

	r := NewRepository(&NewRepositoryOptions{
		Logger:              logger,
		WSEventManager:      wsEventManager,
		Database:            db,
		MetadataProviderRef: metadataProviderRef,
		PlatformRef:         platformRef,
		PlaybackManager:     playbackManager,
	})

	return r
}
