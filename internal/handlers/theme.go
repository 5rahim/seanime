package handlers

import (
	"seanime/internal/database/models"

	"github.com/labstack/echo/v4"
)

// HandleGetTheme
//
//	@summary returns the theme settings.
//	@route /api/v1/theme [GET]
//	@returns models.Theme
func (h *Handler) HandleGetTheme(c echo.Context) error {
	theme, err := h.App.Database.GetTheme()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, theme)
}

// HandleUpdateTheme
//
//	@summary updates the theme settings.
//	@desc The server status should be re-fetched after this on the client.
//	@route /api/v1/theme [PATCH]
//	@returns models.Theme
func (h *Handler) HandleUpdateTheme(c echo.Context) error {
	type body struct {
		Theme models.Theme `json:"theme"`
	}

	var b body

	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Set the theme ID to 1, so we overwrite the previous settings
	b.Theme.BaseModel = models.BaseModel{
		ID: 1,
	}

	currentTheme, err := h.App.Database.GetTheme()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	b.Theme.HomeItems = currentTheme.HomeItems

	// Update the theme settings
	if _, err := h.App.Database.UpsertTheme(&b.Theme); err != nil {
		return h.RespondWithError(c, err)
	}

	// Send the new theme to the client
	return h.RespondWithData(c, b.Theme)
}
