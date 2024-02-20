package handlers

// HandleGetLatestScanSummaries will return the latest scan summaries.
//
//	GET /v1/library/scan-summaries
func HandleGetLatestScanSummaries(c *RouteCtx) error {

	sm, err := c.App.Database.GetScanSummaries()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sm)
}
