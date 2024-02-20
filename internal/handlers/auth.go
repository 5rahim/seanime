package handlers

import (
	"context"
	"errors"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/models"
	"time"
)

type AuthRequestBody struct {
	Token string
}

// HandleLogin is called when a new JWT is obtained (after login from AniList)
// It saves the JWT in the database, fetches the user data from AniList, and returns a new Status.
//
//	POST /v1/auth/login
func HandleLogin(c *RouteCtx) error {

	c.Fiber.Accepts("application/json")

	// Body data
	body := new(AuthRequestBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	// Set a new AniList client by passing to JWT token
	c.App.UpdateAnilistClientToken(body.Token)

	// Get viewer data from AniList
	getViewer, err := c.App.AnilistClientWrapper.Client.GetViewer(context.Background())
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
		Token:    body.Token,
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

// HandleLogout logs out the user by removing the user data from the database.
// It returns a new Status that will be used to update the client.
//
//	POST /auth/logout
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
