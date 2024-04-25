package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/util/limiter"
)

// HandleGetLibraryCollection
//
//	@summary returns the main local anime collection.
//	@desc This creates a new LibraryCollection struct and returns it.
//	@desc This is used to get the main anime collection of the user.
//	@desc It uses the cached Anilist anime collection for the GET method.
//	@desc It refreshes the AniList anime collection if the POST method is used.
//	@route /api/v1/library/collection [GET,POST]
//	@returns anime.LibraryCollection
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

// HandleAddUnknownMedia
//
//	@summary adds the given media to the user's AniList planning collections
//	@desc Since media not found in the user's AniList collection are not displayed in the library, this route is used to add them.
//	@desc The response is ignored in the frontend, the client should just refetch the entire library collection.
//	@route /api/v1/library/unknown-media [POST]
//	@returns anilist.AnimeCollection
func HandleAddUnknownMedia(c *RouteCtx) error {

	type body struct {
		MediaIds []int `json:"mediaIds"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Add non-added media entries to AniList collection
	if err := c.App.AnilistClientWrapper.AddMediaToPlanning(b.MediaIds, limiter.NewAnilistLimiter(), c.App.Logger); err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, this is most likely a rate limit issue"))
	}

	// Bypass the cache
	anilistCollection, err := c.App.GetAnilistCollection(true)
	if err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, wait one minute before refreshing"))
	}

	return c.RespondWithData(anilistCollection)

}
