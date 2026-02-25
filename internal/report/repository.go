package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"seanime/internal/constants"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
	"seanime/internal/util"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type Repository struct {
	logger *zerolog.Logger

	savedIssueReport mo.Option[*IssueReport]
}

func NewRepository(logger *zerolog.Logger) *Repository {
	return &Repository{
		logger:           logger,
		savedIssueReport: mo.None[*IssueReport](),
	}
}

type SaveIssueReportOptions struct {
	LogsDir             string                 `json:"logsDir"`
	UserAgent           string                 `json:"userAgent"`
	Description         string                 `json:"description"`
	ClickLogs           []*ClickLog            `json:"clickLogs"`
	NetworkLogs         []*NetworkLog          `json:"networkLogs"`
	ReactQueryLogs      []*ReactQueryLog       `json:"reactQueryLogs"`
	ConsoleLogs         []*ConsoleLog          `json:"consoleLogs"`
	NavigationLogs      []*NavigationLog       `json:"navigationLogs"`
	Screenshots         []*Screenshot          `json:"screenshots"`
	WebSocketLogs       []*WebSocketLog        `json:"websocketLogs"`
	RRWebEvents         []json.RawMessage      `json:"rrwebEvents"`
	LocalFiles          []*anime.LocalFile     `json:"localFiles"`
	Settings            *models.Settings       `json:"settings"`
	DebridSettings      *models.DebridSettings `json:"debridSettings"`
	IsAnimeLibraryIssue bool                   `json:"isAnimeLibraryIssue"`
	ServerStatus        interface{}            `json:"serverStatus"`
	ViewportWidth       int                    `json:"viewportWidth"`
	ViewportHeight      int                    `json:"viewportHeight"`
	RecordingDurationMs int64                  `json:"recordingDurationMs"`
	Username            string                 `json:"username"`
}

func (r *Repository) SaveIssueReport(opts SaveIssueReportOptions) error {

	var toRedact []string
	if opts.Settings != nil {
		toRedact = opts.Settings.GetSensitiveValues()
	}
	if opts.DebridSettings != nil {
		toRedact = append(toRedact, opts.DebridSettings.GetSensitiveValues()...)
	}
	if opts.Username != "" {
		toRedact = append(toRedact, opts.Username)
	}
	toRedact = lo.Filter(toRedact, func(s string, _ int) bool {
		return s != ""
	})

	issueReport, err := NewIssueReport(
		opts.UserAgent,
		constants.Version,
		runtime.GOOS,
		runtime.GOARCH,
		opts.LogsDir,
		opts.IsAnimeLibraryIssue,
		opts.ServerStatus,
		toRedact,
	)
	if err != nil {
		return fmt.Errorf("failed to create issue report: %w", err)
	}

	for _, log := range opts.ConsoleLogs {
		log.Content = util.StripAnsi(log.Content)
	}

	issueReport.Description = opts.Description
	issueReport.ClickLogs = opts.ClickLogs
	issueReport.NetworkLogs = opts.NetworkLogs
	issueReport.ReactQueryLogs = opts.ReactQueryLogs
	issueReport.ConsoleLogs = opts.ConsoleLogs
	issueReport.NavigationLogs = opts.NavigationLogs
	issueReport.Screenshots = opts.Screenshots
	issueReport.WebSocketLogs = opts.WebSocketLogs
	issueReport.RRWebEvents = opts.RRWebEvents
	issueReport.ViewportWidth = opts.ViewportWidth
	issueReport.ViewportHeight = opts.ViewportHeight
	issueReport.RecordingDurationMs = opts.RecordingDurationMs
	if opts.IsAnimeLibraryIssue {
		for _, localFile := range opts.LocalFiles {
			if localFile.Locked {
				continue
			}
			issueReport.UnlockedLocalFiles = append(issueReport.UnlockedLocalFiles, &UnlockedLocalFile{
				Path:    localFile.Path,
				MediaId: localFile.MediaId,
			})
		}
	}

	r.savedIssueReport = mo.Some(issueReport)

	return nil
}

func (r *Repository) GetSavedIssueReport() (*IssueReport, bool) {
	if r.savedIssueReport.IsAbsent() {
		return nil, false
	}

	return r.savedIssueReport.MustGet(), true
}

type AnonymizeOptions struct {
	Content        []byte `json:"content"`
	Settings       *models.Settings
	DebridSettings *models.DebridSettings
	Username       string
}

func (r *Repository) Anonymize(opts AnonymizeOptions) string {
	userPathPattern := regexp.MustCompile(`(?i)(/home/|/Users/|C:\\Users\\)([^/\\]+)`)

	urlSensitivePattern := regexp.MustCompile(`(?i)(\b(?:client_id|token|secret|password)=)([^&\s"']+)`)

	var toRedact []string
	if opts.Settings != nil {
		toRedact = opts.Settings.GetSensitiveValues()
	}
	if opts.DebridSettings != nil {
		toRedact = append(toRedact, opts.DebridSettings.GetSensitiveValues()...)
	}

	// Remove empty strings to avoid infinite replacements
	// don't redact "seanime"
	toRedact = lo.Filter(toRedact, func(s string, _ int) bool {
		return s != "" && s != "seanime" && s != "Seanime"
	})

	content := opts.Content

	if opts.Username != "" {
		content = bytes.ReplaceAll(content, []byte(opts.Username), []byte("[REDACTED]"))
	}
	content = userPathPattern.ReplaceAll(content, []byte("${1}[REDACTED]"))
	content = urlSensitivePattern.ReplaceAll(content, []byte("${1}[REDACTED]"))

	replacement := []byte("[REDACTED]")
	for _, redact := range toRedact {
		content = bytes.ReplaceAll(content, []byte(redact), replacement)
	}

	return string(content)
}
