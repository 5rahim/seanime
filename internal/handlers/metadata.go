package handlers

import (
	"seanime/internal/api/metadata"

	"github.com/labstack/echo/v4"
)

// HandlePopulateTVDBEpisodes
//
//	@summary populate cache with TVDB episode metadata.
//	@desc This will populate the cache with TVDB episode metadata for the given media.
//	@returns []tvdb.Episode
//	@route /api/v1/metadata-provider/tvdb-episodes [POST]
func (h *Handler) HandlePopulateTVDBEpisodes(c echo.Context) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	animeMetadata, err := h.App.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	media, err := h.App.AnilistPlatform.GetAnime(c.Request().Context(), b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Create media wrapper
	aw := h.App.MetadataProvider.GetAnimeMetadataWrapper(media, animeMetadata)

	// Fetch episodes
	episodes, err := aw.GetTVDBEpisodes(true)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Respond
	return h.RespondWithData(c, episodes)
}

// HandleEmptyTVDBEpisodes
//
//	@summary empties TVDB episode metadata cache.
//	@desc This will empty the TVDB episode metadata cache for the given media.
//	@returns bool
//	@route /api/v1/metadata-provider/tvdb-episodes [DELETE]
func (h *Handler) HandleEmptyTVDBEpisodes(c echo.Context) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	animeMetadata, err := h.App.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	media, err := h.App.AnilistPlatform.GetAnime(c.Request().Context(), b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Create media wrapper
	aw := h.App.MetadataProvider.GetAnimeMetadataWrapper(media, animeMetadata)

	// Empty TVDB episodes bucket
	err = aw.EmptyTVDBEpisodesBucket(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Respond
	return h.RespondWithData(c, true)
}

// HandlePopulateFillerData
//
//	@summary fetches and caches filler data for the given media.
//	@desc This will fetch and cache filler data for the given media.
//	@returns true
//	@route /api/v1/metadata-provider/filler [POST]
func (h *Handler) HandlePopulateFillerData(c echo.Context) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	animeCollection, err := h.App.GetAnimeCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	media, found := animeCollection.FindAnime(b.MediaId)
	if !found {
		// Fetch media
		media, err = h.App.AnilistPlatform.GetAnime(c.Request().Context(), b.MediaId)
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	// Fetch filler data
	err = h.App.FillerManager.FetchAndStoreFillerData(b.MediaId, media.GetAllTitlesDeref())
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleRemoveFillerData
//
//	@summary removes filler data cache.
//	@desc This will remove the filler data cache for the given media.
//	@returns bool
//	@route /api/v1/metadata-provider/filler [DELETE]
func (h *Handler) HandleRemoveFillerData(c echo.Context) error {
	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.FillerManager.RemoveFillerData(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
