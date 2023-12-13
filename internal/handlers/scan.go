package handlers

import (
	"github.com/seanime-app/seanime/internal/scanner"
)

type scanRequestBody struct {
	Enhanced         bool `json:"enhanced"`
	SkipLockedFiles  bool `json:"skipLockedFiles"`
	SkipIgnoredFiles bool `json:"skipIgnoredFiles"`
}

func HandleScanLocalFiles(c *RouteCtx) error {

	c.AcceptJSON()

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

	// +---------------------+
	// |      Account        |
	// +---------------------+

	// Get the user's account
	// If the account is not defined, return an error
	acc, err := c.App.GetAccount()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the latest local files
	existingLfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// +---------------------+
	// |       Scanner       |
	// +---------------------+

	// Create a new scanner
	sc := scanner.Scanner{
		DirPath:            libraryPath,
		Username:           acc.Username,
		Enhanced:           body.Enhanced,
		AnilistClient:      c.App.AnilistClient,
		Logger:             c.App.Logger,
		WSEventManager:     c.App.WSEventManager,
		ExistingLocalFiles: existingLfs,
		SkipLockedFiles:    body.SkipLockedFiles,
		SkipIgnoredFiles:   body.SkipIgnoredFiles,
	}

	// Scan the library
	allLfs, err := sc.Scan()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Insert the local files
	lfs, err := c.App.Database.InsertLocalFiles(allLfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}
