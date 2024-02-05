package handlers

import (
	"github.com/seanime-app/seanime/internal/scanner"
	"github.com/seanime-app/seanime/internal/summary"
)

type scanRequestBody struct {
	Enhanced         bool `json:"enhanced"`
	SkipLockedFiles  bool `json:"skipLockedFiles"`
	SkipIgnoredFiles bool `json:"skipIgnoredFiles"`
}

// HandleScanLocalFiles will scan the user's library.
// POST /library/scan
// Response: entities.LocalFile[]
func HandleScanLocalFiles(c *RouteCtx) error {

	c.AcceptJSON()

	// Retrieve the user's library path
	libraryPath, err := c.App.Database.GetLibraryPathFromSettings()
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

	// Create scan summary logger
	scanSummaryLogger := summary.NewScanSummaryLogger()

	// Create a new scanner
	sc := scanner.Scanner{
		DirPath:              libraryPath,
		Username:             acc.Username,
		Enhanced:             body.Enhanced,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		Logger:               c.App.Logger,
		WSEventManager:       c.App.WSEventManager,
		ExistingLocalFiles:   existingLfs,
		SkipLockedFiles:      body.SkipLockedFiles,
		SkipIgnoredFiles:     body.SkipIgnoredFiles,
		ScanSummaryLogger:    scanSummaryLogger,
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

	// Save the scan summary
	err = c.App.Database.InsertScanSummary(scanSummaryLogger.GenerateSummary())

	return c.RespondWithData(lfs)

}
