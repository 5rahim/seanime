package extension_repo

import (
	"seanime/internal/events"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func GetMockExtensionRepository(t *testing.T) *Repository {
	logger := util.NewLogger()
	filecacher, _ := filecache.NewCacher(t.TempDir())
	extensionRepository := NewRepository(&NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   t.TempDir(),
		WSEventManager: events.NewMockWSEventManager(logger),
		FileCacher:     filecacher,
	})

	return extensionRepository
}
