package handlers

import "seanime/internal/database/models"

// HandleGetTheme
//
//	@summary returns the theme settings.
//	@route /api/v1/theme [GET]
//	@returns models.Theme
func HandleGetTheme(c *RouteCtx) error {
	theme, err := c.App.Database.GetTheme()
	if err != nil {
		return c.RespondWithError(err)
	}
	return c.RespondWithData(theme)
}

// HandleUpdateTheme
//
//	@summary updates the theme settings.
//	@desc The server status should be re-fetched after this on the client.
//	@route /api/v1/theme [PATCH]
//	@returns models.Theme
func HandleUpdateTheme(c *RouteCtx) error {
	type body struct {
		Theme models.Theme `json:"theme"`
	}

	var b body

	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Set the theme ID to 1, so we overwrite the previous settings
	b.Theme.BaseModel = models.BaseModel{
		ID: 1,
	}

	// Update the theme settings
	if _, err := c.App.Database.UpsertTheme(&b.Theme); err != nil {
		return c.RespondWithError(err)
	}

	// Send the new theme to the client
	return c.RespondWithData(b.Theme)
}
