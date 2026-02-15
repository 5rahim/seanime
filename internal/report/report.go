package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ClickLog struct {
	Timestamp time.Time `json:"timestamp"`
	Element   string    `json:"element"`
	PageURL   string    `json:"pageUrl"`
	Text      *string   `json:"text"`
	ClassName *string   `json:"className"`
}

type NetworkLog struct {
	Type        string    `json:"type"`
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	PageURL     string    `json:"pageUrl"`
	Status      int       `json:"status"`
	Duration    int       `json:"duration"`
	DataPreview string    `json:"dataPreview"`
	Body        string    `json:"body"`
	Timestamp   time.Time `json:"timestamp"`
}

type ReactQueryLog struct {
	Type        string      `json:"type"`
	PageURL     string      `json:"pageUrl"`
	Status      string      `json:"status"`
	Hash        string      `json:"hash"`
	Error       interface{} `json:"error"`
	DataPreview string      `json:"dataPreview"`
	DataType    string      `json:"dataType"`
	Timestamp   time.Time   `json:"timestamp"`
}

type ConsoleLog struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	PageURL   string    `json:"pageUrl"`
	Timestamp time.Time `json:"timestamp"`
}

type UnlockedLocalFile struct {
	Path    string `json:"path"`
	MediaId int    `json:"mediaId"`
}

type NavigationLog struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Timestamp time.Time `json:"timestamp"`
}

type Screenshot struct {
	Data      string    `json:"data"` // base64 encoded image
	Caption   string    `json:"caption,omitempty"`
	PageURL   string    `json:"pageUrl"`
	Timestamp time.Time `json:"timestamp"`
}

// WebSocketLog represents a captured WebSocket message during recording
type WebSocketLog struct {
	Direction string          `json:"direction"` // "incoming" or "outgoing"
	EventType string          `json:"eventType"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type IssueReport struct {
	CreatedAt           time.Time            `json:"createdAt"`
	UserAgent           string               `json:"userAgent"`
	AppVersion          string               `json:"appVersion"`
	OS                  string               `json:"os"`
	Arch                string               `json:"arch"`
	Description         string               `json:"description,omitempty"`
	ClickLogs           []*ClickLog          `json:"clickLogs,omitempty"`
	NetworkLogs         []*NetworkLog        `json:"networkLogs,omitempty"`
	ReactQueryLogs      []*ReactQueryLog     `json:"reactQueryLogs,omitempty"`
	ConsoleLogs         []*ConsoleLog        `json:"consoleLogs,omitempty"`
	NavigationLogs      []*NavigationLog     `json:"navigationLogs,omitempty"`
	Screenshots         []*Screenshot        `json:"screenshots,omitempty"`
	WebSocketLogs       []*WebSocketLog      `json:"websocketLogs,omitempty"`
	RRWebEvents         []json.RawMessage    `json:"rrwebEvents,omitempty"`
	UnlockedLocalFiles  []*UnlockedLocalFile `json:"unlockedLocalFiles,omitempty"`
	ScanLogs            []string             `json:"scanLogs,omitempty"`
	ServerLogs          string               `json:"serverLogs,omitempty"`
	ServerStatus        string               `json:"status,omitempty"`
	ViewportWidth       int                  `json:"viewportWidth,omitempty"`
	ViewportHeight      int                  `json:"viewportHeight,omitempty"`
	RecordingDurationMs int64                `json:"recordingDurationMs,omitempty"`
}

func NewIssueReport(userAgent, appVersion, _os, arch string, logsDir string, isAnimeLibraryIssue bool, serverStatus interface{}, toRedact []string) (ret *IssueReport, err error) {
	ret = &IssueReport{
		CreatedAt:          time.Now(),
		UserAgent:          userAgent,
		AppVersion:         appVersion,
		OS:                 _os,
		Arch:               arch,
		ClickLogs:          make([]*ClickLog, 0),
		NetworkLogs:        make([]*NetworkLog, 0),
		ReactQueryLogs:     make([]*ReactQueryLog, 0),
		ConsoleLogs:        make([]*ConsoleLog, 0),
		NavigationLogs:     make([]*NavigationLog, 0),
		Screenshots:        make([]*Screenshot, 0),
		WebSocketLogs:      make([]*WebSocketLog, 0),
		RRWebEvents:        make([]json.RawMessage, 0),
		UnlockedLocalFiles: make([]*UnlockedLocalFile, 0),
		ScanLogs:           make([]string, 0),
		ServerLogs:         "",
		ServerStatus:       "",
	}

	// Get all log files in the directory
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}
	var serverLogFiles []os.FileInfo
	var scanLogFiles []os.FileInfo

	for _, file := range entries {
		if strings.HasPrefix(file.Name(), "seanime-") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			serverLogFiles = append(serverLogFiles, info)
		}
		if strings.Contains(file.Name(), "-scan") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			// Check if file is newer than 1 day
			if time.Since(info.ModTime()).Hours() < 24 {
				scanLogFiles = append(scanLogFiles, info)
			}
		}
	}

	userPathPattern := regexp.MustCompile(`(?i)(/home/|/Users/|C:\\Users\\)([^/\\]+)`)

	if serverStatus != nil {
		serverStatusMarshaled, err := json.Marshal(serverStatus)
		if err == nil {
			// pretty print the json
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, serverStatusMarshaled, "", "  ")
			if err == nil {
				ret.ServerStatus = prettyJSON.String()

				for _, redact := range toRedact {
					ret.ServerStatus = strings.ReplaceAll(ret.ServerStatus, redact, "[REDACTED]")
				}

				ret.ServerStatus = userPathPattern.ReplaceAllString(ret.ServerStatus, "${1}[REDACTED]")
			}
		}
	}

	if len(serverLogFiles) > 0 {
		sort.Slice(serverLogFiles, func(i, j int) bool {
			return serverLogFiles[i].ModTime().After(serverLogFiles[j].ModTime())
		})
		// Get the most recent log file
		latestLog := serverLogFiles[0]
		latestLogPath := filepath.Join(logsDir, latestLog.Name())
		latestLogContent, err := os.ReadFile(latestLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read log file: %w", err)
		}

		latestLogContent = userPathPattern.ReplaceAll(latestLogContent, []byte("${1}[REDACTED]"))

		for _, redact := range toRedact {
			latestLogContent = bytes.ReplaceAll(latestLogContent, []byte(redact), []byte("[REDACTED]"))
		}
		ret.ServerLogs = string(latestLogContent)
	}

	if isAnimeLibraryIssue {
		if len(scanLogFiles) > 0 {
			for _, file := range scanLogFiles {
				scanLogPath := filepath.Join(logsDir, file.Name())
				scanLogContent, err := os.ReadFile(scanLogPath)
				if err != nil {
					continue
				}

				scanLogContent = userPathPattern.ReplaceAll(scanLogContent, []byte("${1}[REDACTED]"))

				if len(scanLogContent) == 0 {
					ret.ScanLogs = append(ret.ScanLogs, "EMPTY")
				} else {
					ret.ScanLogs = append(ret.ScanLogs, string(scanLogContent))

				}
			}
		}
	}

	return
}

func (ir *IssueReport) AddClickLogs(clickLogs []*ClickLog) {
	ir.ClickLogs = append(ir.ClickLogs, clickLogs...)
}

func (ir *IssueReport) AddNetworkLogs(networkLogs []*NetworkLog) {
	ir.NetworkLogs = append(ir.NetworkLogs, networkLogs...)
}

func (ir *IssueReport) AddReactQueryLogs(reactQueryLogs []*ReactQueryLog) {
	ir.ReactQueryLogs = append(ir.ReactQueryLogs, reactQueryLogs...)
}

func (ir *IssueReport) AddConsoleLogs(consoleLogs []*ConsoleLog) {
	ir.ConsoleLogs = append(ir.ConsoleLogs, consoleLogs...)
}
