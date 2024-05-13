package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
)

// HandleGetTorrentstreamSettings
//
//	@summary get torrentstream settings.
//	@desc This returns the torrentstream settings.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [GET]
func HandleGetTorrentstreamSettings(c *RouteCtx) error {
	torrentstreamSettings, found := c.App.Database.GetTorrentstreamSettings()
	if !found {
		return c.RespondWithError(errors.New("torrent streaming settings not found"))
	}

	return c.RespondWithData(torrentstreamSettings)
}

// HandleSaveTorrentstreamSettings
//
//	@summary save torrentstream settings.
//	@desc This saves the torrentstream settings.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [PATCH]
func HandleSaveTorrentstreamSettings(c *RouteCtx) error {

	type body struct {
		Settings models.TorrentstreamSettings `json:"settings"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertTorrentstreamSettings(&b.Settings)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.InitOrRefreshTorrentstreamSettings()

	return c.RespondWithData(settings)
}

func HandleTorrentstreamServeStream(c *RouteCtx) error {

	return nil
}
