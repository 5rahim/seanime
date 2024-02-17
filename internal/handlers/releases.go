package handlers

import "github.com/seanime-app/seanime/internal/updater"

func HandleGetLatestUpdate(c *RouteCtx) error {
	update, err := c.App.Updater.GetLatestUpdate()
	if err != nil {
		return c.RespondWithData(&updater.Update{})
	}

	return c.RespondWithData(update)
}
