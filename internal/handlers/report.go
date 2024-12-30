package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/report"
	"time"
)

// HandleSaveIssueReport
//
//	@summary saves the issue report in memory.
//	@route /api/v1/report/issue [POST]
//	@returns bool
func HandleSaveIssueReport(c *RouteCtx) error {

	type body struct {
		ClickLogs           []*report.ClickLog      `json:"clickLogs"`
		NetworkLogs         []*report.NetworkLog    `json:"networkLogs"`
		ReactQueryLogs      []*report.ReactQueryLog `json:"reactQueryLogs"`
		ConsoleLogs         []*report.ConsoleLog    `json:"consoleLogs"`
		IsAnimeLibraryIssue bool                    `json:"isAnimeLibraryIssue"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var localFiles []*anime.LocalFile
	if b.IsAnimeLibraryIssue {
		// Get local files
		var err error
		localFiles, _, err = db_bridge.GetLocalFiles(c.App.Database)
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	if err := c.App.ReportRepository.SaveIssueReport(report.SaveIssueReportOptions{
		LogsDir:             c.App.Config.Logs.Dir,
		UserAgent:           c.Fiber.Get("User-Agent"),
		ClickLogs:           b.ClickLogs,
		NetworkLogs:         b.NetworkLogs,
		ReactQueryLogs:      b.ReactQueryLogs,
		ConsoleLogs:         b.ConsoleLogs,
		Settings:            c.App.Settings,
		DebridSettings:      c.App.SecondarySettings.Debrid,
		IsAnimeLibraryIssue: b.IsAnimeLibraryIssue,
		LocalFiles:          localFiles,
	}); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleDownloadIssueReport
//
//	@summary generates and downloads the issue report file.
//	@route /api/v1/report/issue/download [GET]
//	@returns report.IssueReport
func HandleDownloadIssueReport(c *RouteCtx) error {

	issueReport, ok := c.App.ReportRepository.GetSavedIssueReport()
	if !ok {
		return c.RespondWithError(fmt.Errorf("no issue report found"))
	}

	marshaledIssueReport, err := json.Marshal(issueReport)
	if err != nil {
		return c.RespondWithError(fmt.Errorf("failed to marshal issue report: %w", err))
	}

	buffer := bytes.Buffer{}
	buffer.Write(marshaledIssueReport)

	// Generate filename with current timestamp
	filename := fmt.Sprintf("issue_report_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	// Set content disposition header for file download
	c.Fiber.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Fiber.Set("Content-Type", "application/json")

	return c.Fiber.SendStream(&buffer)
}
