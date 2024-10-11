package debrid_client

import (
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/util"
	"testing"
)

func GetMockRepository(t *testing.T, db *db.Database) *Repository {
	logger := util.NewLogger()

	r := NewRepository(&NewRepositoryOptions{
		Logger:         logger,
		WSEventManager: events.NewMockWSEventManager(logger),
		Database:       db,
	})

	return r
}
