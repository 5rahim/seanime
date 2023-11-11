package handlers

import "github.com/seanime-app/seanime-server/internal/entities"

// HandleGetLibraryCollection returns the library collection
func HandleGetLibraryCollection(c *RouteCtx) error {

	bypassCache := c.Fiber.Method() == "POST"

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistCollection, err := c.App.GetAnilistCollection(bypassCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	libraryCollection := entities.NewLibraryCollection(&entities.NewLibraryCollectionOptions{
		AnilistCollection: anilistCollection,
		AnilistClient:     c.App.AnilistClient,
		AnizipCache:       c.App.AnizipCache,
		LocalFiles:        lfs,
	})

	return c.RespondWithData(libraryCollection)
}

//----------------------------------------------------------------------------------------------------------------------

//func HandleGetContinueWatching(c *RouteCtx) error {
//
//	type ContinueWatching struct {
//		Entry *entities.MediaEntry `json:"entry"`
//	}
//
//	return nil
//}
