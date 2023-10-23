package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime-server/internal/scanner"
)

type ScanRequestBody struct {
	Dir      string `json:"dir"`
	Username string `json:"userName"`
}

func ScanLocalFiles(c *RouteCtx) error {

	c.Fiber.Accepts(fiber.MIMEApplicationJSON)

	//token := c.Cookies("anilistToken", "")

	// Body
	body := new(ScanRequestBody)
	// Parse body
	if err := c.Fiber.BodyParser(body); err != nil {
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	// Get local files
	localFiles, err := scanner.GetLocalFilesFromDir(body.Dir)
	if err != nil {
		return c.Fiber.JSON(NewErrorResponse(err))
	}

	//util.Logger.Debug().Msgf("test")

	return c.Fiber.JSON(NewDataResponse(localFiles))

}
