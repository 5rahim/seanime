package handlers

import "github.com/seanime-app/seanime/internal/library/anime"

// HandleGetLibraryCollection
//
//	@summary returns the main local anime collection.
//	@desc This creates a new LibraryCollection struct and returns it.
//	@desc This is used to get the main anime collection of the user.
//	@desc It uses the cached Anilist anime collection for the GET method.
//	@desc It refreshes the AniList anime collection if the POST method is used.
//	@route /api/v1/library/collection [GET,POST]
//	@returns entities.LibraryCollection
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

	libraryCollection, err := anime.NewLibraryCollection(&anime.NewLibraryCollectionOptions{
		AnilistCollection:    anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		AnizipCache:          c.App.AnizipCache,
		LocalFiles:           lfs,
		MetadataProvider:     c.App.MetadataProvider,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(libraryCollection)
}
