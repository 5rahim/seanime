package handlers

import "github.com/seanime-app/seanime/internal/entities"

// HandleGetLibraryCollection returns the library collection
// GET /library/collection
func HandleGetLibraryCollection(c *RouteCtx) error {

	bypassCache := c.Fiber.Method() == "POST"

	lfs, _, err := c.App.Database.GetLocalFiles()
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
