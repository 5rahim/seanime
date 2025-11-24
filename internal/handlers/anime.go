package handlers

import (
	"seanime/internal/library/anime"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getAnimeEpisodeCollection(c echo.Context, mId int) (*anime.EpisodeCollection, error) {

	h.App.AddOnRefreshAnilistCollectionFunc("HandleGetAnimeEpisodeCollection", func() {
		anime.ClearEpisodeCollectionCache()
	})

	completeAnime, animeMetadata, err := h.App.TorrentstreamRepository.GetMediaInfo(c.Request().Context(), mId)
	if err != nil {
		return nil, err
	}

	ec, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:       animeMetadata,
		Media:               completeAnime.ToBaseAnime(),
		MetadataProviderRef: h.App.MetadataProviderRef,
		Logger:              h.App.Logger,
	})
	if err != nil {
		return nil, err
	}

	h.App.FillerManager.HydrateEpisodeFillerData(mId, ec.Episodes)

	return ec, nil
}

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

	ec, err := h.getAnimeEpisodeCollection(c, mId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ec)
}
