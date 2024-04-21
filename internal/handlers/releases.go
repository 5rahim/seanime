package handlers

import "github.com/seanime-app/seanime/internal/updater"

// HandleGetLatestUpdate
//
//	@summary returns the latest update.
//	@desc This will return the latest update.
//	@desc If an error occurs, it will return an empty update.
//	@route /api/v1/latest-update [GET]
//	@returns updater.Update
func HandleGetLatestUpdate(c *RouteCtx) error {
	update, err := c.App.Updater.GetLatestUpdate()
	if err != nil {
		return c.RespondWithData(&updater.Update{})
	}

	return c.RespondWithData(update)
}
