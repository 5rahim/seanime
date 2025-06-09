package handlers

import (
	"seanime/internal/library/anime"
	"strconv"

	"github.com/labstack/echo/v4"
	lop "github.com/samber/lo/parallel"
)

// HandleGetAnimeEpisodeCollection
//
//	@summary gets list of main episodes
//	@desc This returns a list of main episodes for the given AniList anime media id.
//	@desc It also loads the episode list into the different modules.
//	@returns anime.EpisodeCollection
//	@param id - int - true - "AniList anime media ID"
//	@route /api/v1/anime/episode-collection/{id} [GET]
func (h *Handler) HandleGetAnimeEpisodeCollection(c echo.Context) error {
	mId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	completeAnime, animeMetadata, err := h.App.TorrentstreamRepository.GetMediaInfo(c.Request().Context(), mId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	ec, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:    animeMetadata,
		Media:            completeAnime.ToBaseAnime(),
		MetadataProvider: h.App.MetadataProvider,
		Logger:           h.App.Logger,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	lop.ForEach(ec.Episodes, func(e *anime.Episode, _ int) {
		h.App.FillerManager.HydrateEpisodeFillerData(mId, e)
	})

	return h.RespondWithData(c, ec)
}
