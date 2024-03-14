package handlers

import "github.com/seanime-app/seanime/internal/library/entities"

// HandleGetLibraryCollection generates and returns the library collection.
//
//	GET /v1/library/collection -> Uses cached Anilist collection
//
//	POST /v1/library/collection -> Bypasses cache and fetches Anilist collection
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

	libraryCollection, err := entities.NewLibraryCollection(&entities.NewLibraryCollectionOptions{
		AnilistCollection:    anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		AnizipCache:          c.App.AnizipCache,
		LocalFiles:           lfs,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(libraryCollection)
}
