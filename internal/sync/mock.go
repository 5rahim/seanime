package sync

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension_repo"
	"seanime/internal/manga"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func GetMockManager(t *testing.T, db *db.Database) Manager {
	logger := util.NewLogger()
	metadataProvider := metadata.GetMockProvider(t)
	extensionRepository := extension_repo.GetMockExtensionRepository(t)
	mangaRepository := manga.GetMockRepository(t, db)

	mangaRepository.InitExtensionBank(extensionRepository.GetExtensionBank())

	wsEventManager := events.NewMockWSEventManager(logger)
	anilistClient := anilist.NewMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)

	localDir := filepath.Join(test_utils.ConfigData.Path.DataDir, "offline")
	assetsDir := filepath.Join(test_utils.ConfigData.Path.DataDir, "offline", "assets")

	m, err := NewManager(&NewManagerOptions{
		LocalDir:         localDir,
		AssetDir:         assetsDir,
		Logger:           util.NewLogger(),
		MetadataProvider: metadataProvider,
		MangaRepository:  mangaRepository,
		Database:         db,
		WSEventManager:   wsEventManager,
		AnilistPlatform:  anilistPlatform,
		IsOffline:        false,
	})
	require.NoError(t, err)

	return m
}
