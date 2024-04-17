package handlers

import "github.com/seanime-app/seanime/internal/database/models"

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
	var theme models.Theme
	if err := c.Fiber.BodyParser(&theme); err != nil {
		return c.RespondWithError(err)
	}

	// Set the theme ID to 1, so we overwrite the previous settings
	theme.BaseModel = models.BaseModel{
		ID: 1,
	}

	// Update the theme settings
	if _, err := c.App.Database.UpsertTheme(&theme); err != nil {
		return c.RespondWithError(err)
	}

	// Send the new theme to the client
	return c.RespondWithData(theme)
}
