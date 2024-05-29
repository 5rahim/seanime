package handlers

import (
	"github.com/seanime-app/seanime/internal/updater"
	"time"
)

// HandleInstallLatestUpdate
//
//	@summary installs the latest update.
//	@desc This will install the latest update and launch the new version.
//	@route /api/v1/install-update [POST]
//	@returns handler.Status
func HandleInstallLatestUpdate(c *RouteCtx) error {

	go func() {
		time.Sleep(2 * time.Second)
		c.App.SelfUpdater.StartSelfUpdate()
	}()

	status := NewStatus(c)
	status.Updating = true

	time.Sleep(1 * time.Second)

	return c.RespondWithData(status)
}

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
