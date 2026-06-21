package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleMpvCoreInSightGetCharacterDetails
//
//	@summary returns the character details for MpvCore InSight.
//	@param malId - int - true - "The MAL character ID"
//	@returns mpvcore.InSightCharacterDetails
//	@route /api/v1/mpvcore/insight/character/{malId} [GET]
func (h *Handler) HandleMpvCoreInSightGetCharacterDetails(c echo.Context) error {
	malID, err := strconv.Atoi(c.Param("malId"))
	if err != nil {
		return h.RespondWithError(c, err)
	}
	ret, err := h.App.MpvCore.InSight().GetCharacterInfo(malID)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, ret)
}
