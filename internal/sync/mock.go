package sync

import (
	"github.com/stretchr/testify/require"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/extension_repo"
	"seanime/internal/manga"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func GetMockManager(t *testing.T, db *db.Database) Manager {
	metadataProvider := metadata.GetMockProvider(t)
	extensionRepository := extension_repo.GetMockExtensionRepository(t)
	mangaRepository := manga.GetMockRepository(t, db)

	mangaRepository.InitExtensionBank(extensionRepository.GetExtensionBank())

	m, err := NewManager(&NewManagerOptions{
		DataDir:          test_utils.ConfigData.Path.DataDir,
		Logger:           util.NewLogger(),
		MetadataProvider: metadataProvider,
		MangaRepository:  mangaRepository,
		Database:         db,
	})
	require.NoError(t, err)

	return m
}
