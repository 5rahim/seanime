package handlers

import (
	"errors"
	"seanime/internal/customsource"

	"github.com/labstack/echo/v4"
)

// HandleCustomSourceListAnime
//
//	@summary returns a paginated list of anime from the provider.
//	@desc This will search for torrents and return a list of torrents with previews.
//	@desc If smart search is enabled, it will filter the torrents based on search parameters.
//	@route /api/v1/custom-source/provider/list/anime [POST]
//	@returns hibikecustomsource.ListAnimeResponse
func (h *Handler) HandleCustomSourceListAnime(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		Search   string `json:"search"`
		Page     int    `json:"page"`
		PerPage  int    `json:"perPage"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	provider, ok := h.App.ExtensionRepository.GetCustomSourceExtensionByID(b.Provider)
	if !ok {
		return h.RespondWithError(c, errors.New("provider extension not found"))
	}

	res, err := provider.GetProvider().ListAnime(c.Request().Context(), b.Search, b.Page, b.PerPage)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	for i := range res.Media {
		customsource.NormalizeMedia(provider.GetExtensionIdentifier(), provider.GetID(), res.Media[i])
	}

	return h.RespondWithData(c, res)
}
