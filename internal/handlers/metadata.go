package handlers

import (
	"seanime/internal/api/anizip"
)

// HandlePopulateTVDBEpisodes
//
//	@summary populate cache with TVDB episode metadata.
//	@desc This will populate the cache with TVDB episode metadata for the given media.
//	@returns []tvdb.Episode
//	@route /api/v1/metadata-provider/tvdb-episodes [POST]
func HandlePopulateTVDBEpisodes(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	anizipMedia, err := anizip.FetchAniZipMedia("anilist", b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	media, err := c.App.AnilistPlatform.GetAnime(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create media wrapper
	mw := c.App.MetadataProvider.NewMediaWrapper(media, anizipMedia)

	// Fetch episodes
	episodes, err := mw.GetTVDBEpisodes(true)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Respond
	return c.RespondWithData(episodes)
}

// HandleEmptyTVDBEpisodes
//
//	@summary empties TVDB episode metadata cache.
//	@desc This will empty the TVDB episode metadata cache for the given media.
//	@returns bool
//	@route /api/v1/metadata-provider/tvdb-episodes [DELETE]
func HandleEmptyTVDBEpisodes(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	anizipMedia, err := anizip.FetchAniZipMedia("anilist", b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	media, err := c.App.AnilistPlatform.GetAnime(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create media wrapper
	mw := c.App.MetadataProvider.NewMediaWrapper(media, anizipMedia)

	// Empty TVDB episodes bucket
	err = mw.EmptyTVDBEpisodesBucket(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Respond
	return c.RespondWithData(true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandlePopulateFillerData
//
//	@summary fetches and caches filler data for the given media.
//	@desc This will fetch and cache filler data for the given media.
//	@returns true
//	@route /api/v1/metadata-provider/filler [POST]
func HandlePopulateFillerData(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	animeCollection, err := c.App.GetAnimeCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	media, found := animeCollection.FindAnime(b.MediaId)
	if !found {
		// Fetch media
		media, err = c.App.AnilistPlatform.GetAnime(b.MediaId)
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	// Fetch filler data
	err = c.App.FillerManager.FetchAndStoreFillerData(b.MediaId, media.GetAllTitlesDeref())
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleRemoveFillerData
//
//	@summary removes filler data cache.
//	@desc This will remove the filler data cache for the given media.
//	@returns bool
//	@route /api/v1/metadata-provider/filler [DELETE]
func HandleRemoveFillerData(c *RouteCtx) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.FillerManager.RemoveFillerData(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
