package handlers

import "github.com/seanime-app/seanime-server/internal/entities"

// HandleGetLibraryCollection returns the library collection (groups of entries)
func HandleGetLibraryCollection(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	collec, err := c.App.GetAnilistCollection()
	if err != nil {
		return c.RespondWithError(err)
	}

	entries := entities.NewLibraryCollection(&entities.NewLibraryCollectionOptions{
		Collection: collec,
		LocalFiles: lfs,
	})

	return c.RespondWithData(entries)
}
