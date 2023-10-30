package handlers

import (
	"context"
	"github.com/seanime-app/seanime-server/internal/entities"
)

func HandleGetLibraryEntries(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	username := "5unwired"
	// TODO Hoist it up to global scope
	// TODO Make auth table in database contaning username and token
	// TODO Create MediaContainer on startup
	collection, err := c.App.AnilistClient.AnimeCollection(context.Background(), &username)
	if err != nil {
		return c.RespondWithError(err)
	}

	le := entities.NewLibraryEntries(&entities.NewLibraryEntriesOptions{
		Collection: collection,
		LocalFiles: lfs,
	})

	return c.RespondWithData(le)
}
