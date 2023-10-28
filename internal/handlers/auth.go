package handlers

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime-server/internal/models"
	"time"
)

type AuthRequestBody struct {
	Token string
}

func HandleEnforceAnilistToken(c *RouteCtx) error {

	token := c.Fiber.Cookies("anilistToken", "")

	if len(token) == 0 {
		return c.Fiber.Status(fiber.StatusMethodNotAllowed).JSON(NewErrorResponse(errors.New("missing AniList token")))
	}

	return c.Fiber.Next()

}

func HandleAuth(c *RouteCtx) error {

	c.Fiber.Accepts("application/json")

	// Body data
	body := new(AuthRequestBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	// Re-init the client, this time by passing the JWT token
	c.App.UpdateAnilistClientToken(body.Token)

	// Get viewer data from AniList
	getViewer, err := c.App.AnilistClient.GetViewer(context.Background())
	if err != nil {
		c.App.Logger.Error().Msg("Could not authenticate to AniList")
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	// Success
	if len(getViewer.Viewer.Name) > 0 {

		_, err = c.App.Database.UpsertToken(&models.Token{
			BaseModel: models.BaseModel{
				ID:        1,
				UpdatedAt: time.Now(),
			},
			Value: body.Token,
		})

		if err != nil {
			return c.Fiber.JSON(NewErrorResponse(err))
		}

		c.App.Logger.Info().Msg("Authenticated to AniList as " + getViewer.Viewer.Name)

		return c.Fiber.JSON(NewDataResponse(getViewer.Viewer.Name))
	}

	return c.Fiber.JSON(NewErrorResponse(errors.New("could not authenticate to AniList")))

}
