package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleVideoCoreInSightGetCharacterDetails
//
//	@summary returns the character details.
//	@param malId - int - true - "The MAL character ID"
//	@returns videocore.InSightCharacterDetails
//	@route /api/v1/videocore/insight/character/{malId} [GET]
func (h *Handler) HandleVideoCoreInSightGetCharacterDetails(c echo.Context) error {
	malId, err := strconv.Atoi(c.Param("malId"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	ret, err := h.App.VideoCore.InSight().GetCharacterInfo(malId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}
