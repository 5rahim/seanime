package handlers

// HandleGetLatestScanSummaries
//
//	@summary returns the latest scan summaries.
//	@route /v1/library/scan-summaries [GET]
//	@returns []db.ScanSummaryItem
func HandleGetLatestScanSummaries(c *RouteCtx) error {

	sm, err := c.App.Database.GetScanSummaries()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sm)
}
