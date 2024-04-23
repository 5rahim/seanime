package handlers

import (
	"context"
	"errors"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/database/models"
	"time"
)

// HandleLogin
//
//	@summary logs in the user by saving the JWT token in the database.
//	@desc This is called when the JWT token is obtained from AniList after logging in with redirection on the client.
//	@desc It also fetches the Viewer data from AniList and saves it in the database.
//	@desc It creates a new handlers.Status and refreshes App modules.
//	@route /api/v1/auth/login [POST]
//	@returns handlers.Status
func HandleLogin(c *RouteCtx) error {

	type body struct {
		Token string `json:"token"`
	}

	var b body

	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	// Set a new AniList client by passing to JWT token
	c.App.UpdateAnilistClientToken(b.Token)

	// Get viewer data from AniList
	getViewer, err := c.App.AnilistClientWrapper.GetViewer(context.Background())
	if err != nil {
		c.App.Logger.Error().Msg("Could not authenticate to AniList")
		return c.RespondWithError(err)
	}

	if len(getViewer.Viewer.Name) == 0 {
		return c.RespondWithError(errors.New("could not find user"))
	}

	// Marshal viewer data
	bytes, err := json.Marshal(getViewer.Viewer)
	if err != nil {
		c.App.Logger.Err(err).Msg("scan: could not save local files")
	}

	// Save account data in database
	_, err = c.App.Database.UpsertAccount(&models.Account{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Username: getViewer.Viewer.Name,
		Token:    b.Token,
		Viewer:   bytes,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.Logger.Info().Msg("Authenticated to AniList as " + getViewer.Viewer.Name)

	// Create a new status
	status := NewStatus(c)

	c.App.InitOrRefreshModules()

	// Return new status
	return c.RespondWithData(status)

}

// HandleLogout
//
//	@summary logs out the user by removing JWT token from the database.
//	@desc It removes JWT token and Viewer data from the database.
//	@desc It creates a new handlers.Status and refreshes App modules.
//	@route /api/v1/auth/logout [POST]
//	@returns handlers.Status
func HandleLogout(c *RouteCtx) error {

	_, err := c.App.Database.UpsertAccount(&models.Account{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Username: "",
		Token:    "",
		Viewer:   nil,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.Logger.Info().Msg("Logged out of AniList")

	status := NewStatus(c)

	c.App.InitOrRefreshModules()

	return c.RespondWithData(status)
}
