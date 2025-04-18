package handlers

import (
	"errors"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/scanner"
	"seanime/internal/library/summary"

	"github.com/labstack/echo/v4"
)

// HandleScanLocalFiles
//
//	@summary scans the user's library.
//	@desc This will scan the user's library.
//	@desc The response is ignored, the client should re-fetch the library after this.
//	@route /api/v1/library/scan [POST]
//	@returns []anime.LocalFile
func (h *Handler) HandleScanLocalFiles(c echo.Context) error {

	type body struct {
		Enhanced         bool `json:"enhanced"`
		SkipLockedFiles  bool `json:"skipLockedFiles"`
		SkipIgnoredFiles bool `json:"skipIgnoredFiles"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Retrieve the user's library path
	libraryPath, err := h.App.Database.GetLibraryPathFromSettings()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	additionalLibraryPaths, err := h.App.Database.GetAdditionalLibraryPathsFromSettings()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Get the latest local files
	existingLfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// +---------------------+
	// |       Scanner       |
	// +---------------------+

	// Create scan summary logger
	scanSummaryLogger := summary.NewScanSummaryLogger()

	// Create a new scan logger
	scanLogger, err := scanner.NewScanLogger(h.App.Config.Logs.Dir)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer scanLogger.Done()

	// Create a new scanner
	sc := scanner.Scanner{
		DirPath:            libraryPath,
		OtherDirPaths:      additionalLibraryPaths,
		Enhanced:           b.Enhanced,
		Platform:           h.App.AnilistPlatform,
		Logger:             h.App.Logger,
		WSEventManager:     h.App.WSEventManager,
		ExistingLocalFiles: existingLfs,
		SkipLockedFiles:    b.SkipLockedFiles,
		SkipIgnoredFiles:   b.SkipIgnoredFiles,
		ScanSummaryLogger:  scanSummaryLogger,
		ScanLogger:         scanLogger,
		MetadataProvider:   h.App.MetadataProvider,
		MatchingAlgorithm:  h.App.Settings.GetLibrary().ScannerMatchingAlgorithm,
		MatchingThreshold:  h.App.Settings.GetLibrary().ScannerMatchingThreshold,
	}

	// Scan the library
	allLfs, err := sc.Scan()
	if err != nil {
		if errors.Is(err, scanner.ErrNoLocalFiles) {
			return h.RespondWithData(c, []interface{}{})
		} else {
			return h.RespondWithError(c, err)
		}
	}

	// Insert the local files
	lfs, err := db_bridge.InsertLocalFiles(h.App.Database, allLfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Save the scan summary
	_ = db_bridge.InsertScanSummary(h.App.Database, scanSummaryLogger.GenerateSummary())

	go h.App.AutoDownloader.CleanUpDownloadedItems()

	return h.RespondWithData(c, lfs)

}
