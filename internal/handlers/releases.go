package handlers

func HandleGetLatestUpdate(c *RouteCtx) error {
	update, err := c.App.Updater.GetLatestUpdate()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(update)
}
