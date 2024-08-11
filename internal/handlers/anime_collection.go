package handlers

import (
	"errors"
	"github.com/dustin/go-humanize"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/torrentstream"
	"seanime/internal/util/result"
)

var libraryCollectionCache = result.NewResultMap[int, *anime.LibraryCollection]()

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

	animeCollection, err := c.App.GetAnimeCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	if animeCollection == nil {
		return c.RespondWithData(&anime.LibraryCollection{})
	}

	//if lc, found := libraryCollectionCache.Get(core.AnimeCollectionCacheId); found {
	//	return c.RespondWithData(lc)
	//}

	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	libraryCollection, err := anime.NewLibraryCollection(&anime.NewLibraryCollectionOptions{
		AnimeCollection:  animeCollection,
		Platform:         c.App.AnilistPlatform,
		AnizipCache:      c.App.AnizipCache,
		LocalFiles:       lfs,
		MetadataProvider: c.App.MetadataProvider,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	if c.App.SecondarySettings.Torrentstream != nil && c.App.SecondarySettings.Torrentstream.IncludeInLibrary {
		c.App.TorrentstreamRepository.HydrateStreamCollection(&torrentstream.HydrateStreamCollectionOptions{
			AnimeCollection:   animeCollection,
			LibraryCollection: libraryCollection,
			AnizipCache:       c.App.AnizipCache,
		})
	}

	// Hydrate total library size
	libraryCollection.Stats.TotalSize = humanize.Bytes(c.App.TotalLibrarySize)

	//libraryCollectionCache.Clear()
	//libraryCollectionCache.Set(core.AnimeCollectionCacheId, libraryCollection)

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
	if err := c.App.AnilistPlatform.AddMediaToCollection(b.MediaIds); err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, this is most likely a rate limit issue"))
	}

	// Bypass the cache
	animeCollection, err := c.App.GetAnimeCollection(true)
	if err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, wait one minute before refreshing"))
	}

	return c.RespondWithData(animeCollection)

}
