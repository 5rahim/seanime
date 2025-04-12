package handlers

import (
	"seanime/internal/database/db_bridge"

	"github.com/labstack/echo/v4"
)

// HandleGetScanSummaries
//
//	@summary returns the latest scan summaries.
//	@route /api/v1/library/scan-summaries [GET]
//	@returns []summary.ScanSummaryItem
func (h *Handler) HandleGetScanSummaries(c echo.Context) error {

	sm, err := db_bridge.GetScanSummaries(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, sm)
}
