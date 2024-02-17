package handlers

func HandleGetLatestUpdate(c *RouteCtx) error {
	update, err := c.App.Updater.GetLatestUpdate()
	if err != nil {
		return c.RespondWithData(nil)
	}

	return c.RespondWithData(update)
}
