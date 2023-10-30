package handlers

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime-server/internal/models"
	"github.com/seanime-app/seanime-server/internal/scanner"
)

type scanRequestBody struct {
	Enhanced bool `json:"enhanced"`
}

func HandleScanLocalFiles(c *RouteCtx) error {

	c.AcceptJSON()

	token := c.GetAnilistToken()

	// Retrieve the user's library path
	libraryPath, err := c.App.Database.GetLibraryPath()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Body
	body := new(scanRequestBody)
	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	acc, err := c.App.GetAccount()
	if err != nil {
		return c.RespondWithError(err)
	}

	sc := scanner.Scanner{
		Token:         token,
		DirPath:       libraryPath,
		Username:      acc.Username,
		Enhanced:      body.Enhanced,
		AnilistClient: c.App.AnilistClient,
		Logger:        c.App.Logger,
	}

	localFiles, err := sc.Scan()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Marshal the local files
	bytes, err := json.Marshal(localFiles)
	if err != nil {
		c.App.Logger.Err(err).Msg("scan: could not save local files")
	}
	// Save the local files to the database
	if _, err := c.App.Database.InsertLocalFiles(&models.LocalFiles{
		Value: bytes,
	}); err != nil {
		c.App.Logger.Err(err).Msg("scan: could not save local files")
	}

	return c.RespondWithData(localFiles)

}
