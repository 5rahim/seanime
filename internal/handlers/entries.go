package handlers

import (
	"github.com/seanime-app/seanime-server/internal/entities"
)

// HandleGetLibraryEntries returns all library entries
func HandleGetLibraryEntries(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	collec, err := c.App.GetAnilistCollection()
	if err != nil {
		return c.RespondWithError(err)
	}

	entries := entities.NewLibraryEntries(&entities.NewLibraryEntriesOptions{
		Collection: collec,
		LocalFiles: lfs,
	})

	return c.RespondWithData(entries)
}

func HandleGetLibraryEntry(c *RouteCtx) error {
	return nil
}
