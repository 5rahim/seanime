package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/report"
	"time"

	"github.com/labstack/echo/v4"
)

// HandleSaveIssueReport
//
//	@summary saves the issue report in memory.
//	@route /api/v1/report/issue [POST]
//	@returns bool
func (h *Handler) HandleSaveIssueReport(c echo.Context) error {

	type body struct {
		ClickLogs           []*report.ClickLog      `json:"clickLogs"`
		NetworkLogs         []*report.NetworkLog    `json:"networkLogs"`
		ReactQueryLogs      []*report.ReactQueryLog `json:"reactQueryLogs"`
		ConsoleLogs         []*report.ConsoleLog    `json:"consoleLogs"`
		IsAnimeLibraryIssue bool                    `json:"isAnimeLibraryIssue"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	var localFiles []*anime.LocalFile
	if b.IsAnimeLibraryIssue {
		// Get local files
		var err error
		localFiles, _, err = db_bridge.GetLocalFiles(h.App.Database)
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	status := h.NewStatus(c)

	if err := h.App.ReportRepository.SaveIssueReport(report.SaveIssueReportOptions{
		LogsDir:             h.App.Config.Logs.Dir,
		UserAgent:           c.Request().Header.Get("User-Agent"),
		ClickLogs:           b.ClickLogs,
		NetworkLogs:         b.NetworkLogs,
		ReactQueryLogs:      b.ReactQueryLogs,
		ConsoleLogs:         b.ConsoleLogs,
		Settings:            h.App.Settings,
		DebridSettings:      h.App.SecondarySettings.Debrid,
		IsAnimeLibraryIssue: b.IsAnimeLibraryIssue,
		LocalFiles:          localFiles,
		ServerStatus:        status,
	}); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDownloadIssueReport
//
//	@summary generates and downloads the issue report file.
//	@route /api/v1/report/issue/download [GET]
//	@returns report.IssueReport
func (h *Handler) HandleDownloadIssueReport(c echo.Context) error {

	issueReport, ok := h.App.ReportRepository.GetSavedIssueReport()
	if !ok {
		return h.RespondWithError(c, fmt.Errorf("no issue report found"))
	}

	marshaledIssueReport, err := json.Marshal(issueReport)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to marshal issue report: %w", err))
	}

	buffer := bytes.Buffer{}
	buffer.Write(marshaledIssueReport)

	// Generate filename with current timestamp
	filename := fmt.Sprintf("issue_report_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "application/json")

	return c.Stream(200, "application/json", &buffer)
}
