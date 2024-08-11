package extension_repo_test

import (
	"seanime/internal/events"
	"seanime/internal/extension_repo"
	"seanime/internal/util"
	"testing"
)

func getRepo(t *testing.T) *extension_repo.Repository {
	logger := util.NewLogger()
	wsEventManager := events.NewMockWSEventManager(logger)

	return extension_repo.NewRepository(&extension_repo.NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   "testdir",
		WSEventManager: wsEventManager,
	})
}
