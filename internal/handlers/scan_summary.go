package handlers

func HandleGetLatestScanSummaries(c *RouteCtx) error {

	sm, err := c.App.Database.GetScanSummaries()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sm)
}
