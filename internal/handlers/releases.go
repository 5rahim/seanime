package handlers

import (
	"seanime/internal/updater"
	"time"

	"github.com/labstack/echo/v4"
)

// HandleInstallLatestUpdate
//
//	@summary installs the latest update.
//	@desc This will install the latest update and launch the new version.
//	@route /api/v1/install-update [POST]
//	@returns handler.Status
func (h *Handler) HandleInstallLatestUpdate(c echo.Context) error {
	type body struct {
		FallbackDestination string `json:"fallback_destination"`
	}
	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		h.App.SelfUpdater.StartSelfUpdate(b.FallbackDestination)
	}()

	status := h.NewStatus(c)
	status.Updating = true

	time.Sleep(1 * time.Second)

	return h.RespondWithData(c, status)
}

// HandleGetLatestUpdate
//
//	@summary returns the latest update.
//	@desc This will return the latest update.
//	@desc If an error occurs, it will return an empty update.
//	@route /api/v1/latest-update [GET]
//	@returns updater.Update
func (h *Handler) HandleGetLatestUpdate(c echo.Context) error {
	update, err := h.App.Updater.GetLatestUpdate()
	if err != nil {
		return h.RespondWithData(c, &updater.Update{})
	}

	return h.RespondWithData(c, update)
}
