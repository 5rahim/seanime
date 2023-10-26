package handlers

import (
	"github.com/seanime-app/seanime-server/internal/scanner"
)

type ScanRequestBody struct {
	DirPath  string `json:"dirPath"`
	Username string `json:"username"`
	Enhanced bool   `json:"enhanced"`
}

func HandleScanLocalFiles(c *RouteCtx) error {

	c.AcceptJSON()

	token := c.GetAnilistToken()

	// Body
	body := new(ScanRequestBody)
	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	sc := scanner.Scanner{
		Token:         token,
		DirPath:       body.DirPath,
		Username:      body.Username,
		Enhanced:      body.Enhanced,
		AnilistClient: c.App.AnilistClient,
		Logger:        c.App.Logger,
		DB:            c.App.Database,
	}

	_, err := sc.Scan()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData("")

}
