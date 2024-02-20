package handlers

import "github.com/seanime-app/seanime/internal/updater"

// HandleGetLatestUpdate will return the latest update.
//
//	GET /v1/latest-update
func HandleGetLatestUpdate(c *RouteCtx) error {
	update, err := c.App.Updater.GetLatestUpdate()
	if err != nil {
		return c.RespondWithData(&updater.Update{})
	}

	return c.RespondWithData(update)
}
