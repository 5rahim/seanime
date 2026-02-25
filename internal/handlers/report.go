package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
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
		Description         string                  `json:"description"`
		ClickLogs           []*report.ClickLog      `json:"clickLogs"`
		NetworkLogs         []*report.NetworkLog    `json:"networkLogs"`
		ReactQueryLogs      []*report.ReactQueryLog `json:"reactQueryLogs"`
		ConsoleLogs         []*report.ConsoleLog    `json:"consoleLogs"`
		NavigationLogs      []*report.NavigationLog `json:"navigationLogs"`
		Screenshots         []*report.Screenshot    `json:"screenshots"`
		WebSocketLogs       []*report.WebSocketLog  `json:"websocketLogs"`
		RRWebEvents         []json.RawMessage       `json:"rrwebEvents"`
		IsAnimeLibraryIssue bool                    `json:"isAnimeLibraryIssue"`
		ViewportWidth       int                     `json:"viewportWidth"`
		ViewportHeight      int                     `json:"viewportHeight"`
		RecordingDurationMs int64                   `json:"recordingDurationMs"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Flush logs to ensure server logs are up-to-date
	if h.App.OnFlushLogs != nil {
		h.App.OnFlushLogs()
		time.Sleep(100 * time.Millisecond)
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
		Description:         b.Description,
		ClickLogs:           b.ClickLogs,
		NetworkLogs:         b.NetworkLogs,
		ReactQueryLogs:      b.ReactQueryLogs,
		ConsoleLogs:         b.ConsoleLogs,
		NavigationLogs:      b.NavigationLogs,
		Screenshots:         b.Screenshots,
		WebSocketLogs:       b.WebSocketLogs,
		RRWebEvents:         b.RRWebEvents,
		Settings:            h.App.Settings,
		DebridSettings:      h.App.SecondarySettings.Debrid,
		IsAnimeLibraryIssue: b.IsAnimeLibraryIssue,
		LocalFiles:          localFiles,
		ServerStatus:        status,
		ViewportWidth:       b.ViewportWidth,
		ViewportHeight:      b.ViewportHeight,
		RecordingDurationMs: b.RecordingDurationMs,
		Username:            h.App.GetUsername(),
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

	// Create a zip archive
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	// Create file in zip
	f, err := zipWriter.Create("issue_report.json")
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create zip entry: %w", err))
	}

	// Write JSON to file in zip
	_, err = f.Write(marshaledIssueReport)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to write to zip entry: %w", err))
	}

	// Close zip writer
	if err := zipWriter.Close(); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to close zip archive: %w", err))
	}

	// Generate filename with current timestamp
	filename := fmt.Sprintf("issue_report_%s.zip", time.Now().Format("2006-01-02_15-04-05"))

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "application/zip")

	return c.Stream(200, "application/zip", buffer)
}

// HandleDecompressIssueReport
//
//	@summary accepts a zip file, decompresses it and returns the issue report JSON.
//	@route /api/v1/report/issue/decompress [POST]
//	@returns report.IssueReport
func (h *Handler) HandleDecompressIssueReport(c echo.Context) error {

	// Read form file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return h.RespondWithError(c, err)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer src.Close()

	// Read file content into a buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return h.RespondWithError(c, err)
	}

	// Open zip archive from the buffer
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to open zip archive: %w", err))
	}

	// Find issue report json
	var reportFile *zip.File
	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".json" {
			reportFile = f
			break
		}
	}

	if reportFile == nil {
		return h.RespondWithError(c, fmt.Errorf("no json file found in zip archive"))
	}

	// Open report file
	rc, err := reportFile.Open()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer rc.Close()

	// Decode json
	var issueReport report.IssueReport
	if err := json.NewDecoder(rc).Decode(&issueReport); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to decode issue report: %w", err))
	}

	return h.RespondWithData(c, issueReport)
}
