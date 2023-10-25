package handlers

import (
	"github.com/seanime-app/seanime-server/internal/scanner"
)

type ScanRequestBody struct {
	Dir      string `json:"dir"`
	Username string `json:"userName"`
}

func HandleScanLocalFiles(c *RouteCtx) error {

	c.AcceptJSON()

	//token := c.Cookies("anilistToken", "")

	// Body
	body := new(ScanRequestBody)
	// Parse body
	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	// Get local files
	localFiles, err := scanner.GetLocalFilesFromDir(body.Dir, c.App.Logger)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(localFiles)

}
