package torrent

import (
	"seanime/internal/api/metadata_provider"
	"seanime/internal/extension"
	"seanime/internal/util"
	"testing"
)

func getTestRepo(t *testing.T) *Repository {
	logger := util.NewLogger()
	metadataProvider := metadata_provider.GetMockProvider(t, nil)

	extensionBank := extension.NewUnifiedBank()

	repo := NewRepository(&NewRepositoryOptions{
		Logger:           logger,
		MetadataProvider: metadataProvider,
	})

	repo.InitExtensionBank(extensionBank)

	repo.SetSettings(&RepositorySettings{
		DefaultAnimeProvider: "",
	})

	return repo
}
