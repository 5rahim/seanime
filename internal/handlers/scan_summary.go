package handlers

import "seanime/internal/database/db_bridge"

// HandleGetScanSummaries
//
//	@summary returns the latest scan summaries.
//	@route /api/v1/library/scan-summaries [GET]
//	@returns []db.ScanSummaryItem
func HandleGetScanSummaries(c *RouteCtx) error {

	sm, err := db_bridge.GetScanSummaries(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sm)
}
