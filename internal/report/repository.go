package report

import (
	"bytes"
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
	ClickLogs           []*ClickLog            `json:"clickLogs"`
	NetworkLogs         []*NetworkLog          `json:"networkLogs"`
	ReactQueryLogs      []*ReactQueryLog       `json:"reactQueryLogs"`
	ConsoleLogs         []*ConsoleLog          `json:"consoleLogs"`
	LocalFiles          []*anime.LocalFile     `json:"localFiles"`
	Settings            *models.Settings       `json:"settings"`
	DebridSettings      *models.DebridSettings `json:"debridSettings"`
	IsAnimeLibraryIssue bool                   `json:"isAnimeLibraryIssue"`
	ServerStatus        interface{}            `json:"serverStatus"`
}

func (r *Repository) SaveIssueReport(opts SaveIssueReportOptions) error {

	var toRedact []string
	if opts.Settings != nil {
		toRedact = opts.Settings.GetSensitiveValues()
	}
	if opts.DebridSettings != nil {
		toRedact = append(toRedact, opts.DebridSettings.GetSensitiveValues()...)
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

	issueReport.ClickLogs = opts.ClickLogs
	issueReport.NetworkLogs = opts.NetworkLogs
	issueReport.ReactQueryLogs = opts.ReactQueryLogs
	issueReport.ConsoleLogs = opts.ConsoleLogs
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
	toRedact = lo.Filter(toRedact, func(s string, _ int) bool {
		return s != ""
	})

	content := opts.Content

	content = userPathPattern.ReplaceAll(content, []byte("${1}[REDACTED]"))
	content = urlSensitivePattern.ReplaceAll(content, []byte("${1}[REDACTED]"))

	replacement := []byte("[REDACTED]")
	for _, redact := range toRedact {
		content = bytes.ReplaceAll(content, []byte(redact), replacement)
	}

	return string(content)
}
