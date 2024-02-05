package summary

import (
	"fmt"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
)

const (
	LogComparison LogType = iota
	LogSuccessfullyMatched
	LogFailedMatch
	LogMatchValidated
	LogUnmatched
	LogMetadataMediaTreeFetched
	LogMetadataMediaTreeFetchFailed
	LogMetadataEpisodeNormalized
	LogMetadataEpisodeNormalizationFailed
	LogMetadataNC
	LogMetadataSpecial
	LogMetadataMain
	LogMetadataHydrated
)

type (
	LogType int

	ScanSummaryLogger struct {
		Logs []*ScanSummaryLoggerLog
	}

	ScanSummaryLoggerLog struct { // Holds a log entry. The log entry will then be used to generate a ScanSummary.
		FilePath string `json:"filePath"`
		Level    string `json:"level"`
		Message  string `json:"message"`
	}

	ScanSummary struct {
		Groups         []*ScanSummaryGroup `json:"groups"`
		UnmatchedFiles []*ScanSummaryFile  `json:"unmatchedFiles"`
		Files          []*ScanSummaryFile  `json:"files"`
	}

	ScanSummaryFile struct {
		LocalFile *entities.LocalFile     `json:"localFile"`
		Logs      []*ScanSummaryLoggerLog `json:"logs"`
	}

	ScanSummaryGroup struct {
		LocalFiles          []*ScanSummaryFile `json:"files"`
		MediaId             int                `json:"mediaId"`
		MediaTitle          string             `json:"mediaTitle"`
		MediaImage          string             `json:"mediaImage"`
		MediaIsInCollection bool               `json:"mediaIsInCollection"` // Whether the media is in the user's AniList collection
	}
)

func NewScanSummaryLogger() *ScanSummaryLogger {
	return &ScanSummaryLogger{
		Logs: make([]*ScanSummaryLoggerLog, 0),
	}
}

func (l *ScanSummaryLogger) GenerateSummary(lfs []*entities.LocalFile, media []*anilist.BasicMedia, anilistCollection *anilist.AnimeCollection) *ScanSummary {
	if l == nil {
		return nil
	}
	summary := &ScanSummary{
		Groups:         make([]*ScanSummaryGroup, 0),
		Files:          make([]*ScanSummaryFile, 0),
		UnmatchedFiles: make([]*ScanSummaryFile, 0),
	}

	groupsMap := make(map[int][]*ScanSummaryFile)

	// Generate summary files
	for _, lf := range lfs {

		if lf.MediaId == 0 {
			summary.UnmatchedFiles = append(summary.UnmatchedFiles, &ScanSummaryFile{
				LocalFile: lf,
				Logs:      l.getFileLogs(lf.Path),
			})
			continue
		}

		summaryFile := &ScanSummaryFile{
			LocalFile: lf,
			Logs:      l.getFileLogs(lf.Path),
		}

		summary.Files = append(summary.Files, summaryFile)

		// Add to group
		if _, ok := groupsMap[lf.MediaId]; !ok {
			groupsMap[lf.MediaId] = make([]*ScanSummaryFile, 0)
			groupsMap[lf.MediaId] = append(groupsMap[lf.MediaId], summaryFile)
		} else {
			groupsMap[lf.MediaId] = append(groupsMap[lf.MediaId], summaryFile)
		}
	}

	// Generate summary groups
	for mediaId, files := range groupsMap {
		mediaTitle := ""
		mediaImage := ""
		mediaIsInCollection := false
		for _, m := range media {
			if m.ID == mediaId {
				mediaTitle = m.GetTitleSafe()
				mediaImage = ""
				if m.GetCoverImage() != nil && m.GetCoverImage().GetLarge() != nil {
					mediaImage = *m.GetCoverImage().GetLarge()
				}
				break
			}
		}
		if _, found := anilistCollection.GetListEntryFromMediaId(mediaId); found {
			mediaIsInCollection = true
		}

		summary.Groups = append(summary.Groups, &ScanSummaryGroup{
			LocalFiles:          files,
			MediaId:             mediaId,
			MediaTitle:          mediaTitle,
			MediaImage:          mediaImage,
			MediaIsInCollection: mediaIsInCollection,
		})
	}

	return summary
}

func (l *ScanSummaryLogger) LogComparison(filePath string, algo string, bestTitle string, rating string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Comparison using %s. Best title: \"%s\". Rating: %s", algo, bestTitle, rating)
	l.logType(LogComparison, filePath, msg)
}

func (l *ScanSummaryLogger) LogSuccessfullyMatched(filePath string, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Successfully matched to media %d", mediaId)
	l.logType(LogSuccessfullyMatched, filePath, msg)
}

func (l *ScanSummaryLogger) LogFailedMatch(filePath string, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Failed to match: %s", reason)
	l.logType(LogFailedMatch, filePath, msg)
}

func (l *ScanSummaryLogger) LogMatchValidated(filePath string, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Match validated for media %d", mediaId)
	l.logType(LogMatchValidated, filePath, msg)
}

func (l *ScanSummaryLogger) LogUnmatched(filePath string, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Unmatched: %s", reason)
	l.logType(LogUnmatched, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetched(filePath string, ms int, requests int, branches string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Media tree fetched in %dms. Requests: %d. Branches: %s", ms, requests, branches)
	l.logType(LogMetadataMediaTreeFetched, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetchFailed(filePath string, err error, ms int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Could not fetch media tree: %s. Took %dms", err.Error(), ms)
	l.logType(LogMetadataMediaTreeFetchFailed, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalized(filePath string, mediaId int, episode int, newEpisode int, newMediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode %d normalized to %d. New media ID: %d", episode, newEpisode, newMediaId)
	l.logType(LogMetadataEpisodeNormalized, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalizationFailed(filePath string, err error) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode normalization failed: %s", err.Error())
	l.logType(LogMetadataEpisodeNormalizationFailed, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataNC(filePath string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as NC file")
	l.logType(LogMetadataNC, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataSpecial(filePath string, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as special episode. Episode %d. AniDB episode: %s", episode, aniDBEpisode)
	l.logType(LogMetadataSpecial, filePath, msg)
}

func (l *ScanSummaryLogger) LogMetadataHydrated(filePath string, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Metadata hydrated for media %d", mediaId)
	l.logType(LogMetadataHydrated, filePath, msg)
}

func (l *ScanSummaryLogger) logType(logType LogType, filePath string, message string) {
	if l == nil {
		return
	}
	switch logType {
	case LogComparison:
		l.log(filePath, "info", message)
	case LogSuccessfullyMatched:
		l.log(filePath, "info", message)
	case LogFailedMatch:
		l.log(filePath, "warning", message)
	case LogMatchValidated:
		l.log(filePath, "info", message)
	case LogUnmatched:
		l.log(filePath, "warning", message)
	case LogMetadataMediaTreeFetched:
		l.log(filePath, "info", message)
	case LogMetadataMediaTreeFetchFailed:
		l.log(filePath, "error", message)
	case LogMetadataEpisodeNormalized:
		l.log(filePath, "info", message)
	case LogMetadataEpisodeNormalizationFailed:
		l.log(filePath, "error", message)
	case LogMetadataHydrated:
		l.log(filePath, "info", message)
	case LogMetadataNC:
		l.log(filePath, "info", message)
	case LogMetadataSpecial:
		l.log(filePath, "info", message)
	case LogMetadataMain:
		l.log(filePath, "info", message)
	}
}

func (l *ScanSummaryLogger) log(filePath string, level string, message string) {
	if l == nil {
		return
	}
	l.Logs = append(l.Logs, &ScanSummaryLoggerLog{
		FilePath: filePath,
		Level:    level,
		Message:  message,
	})
}

func (l *ScanSummaryLogger) getFileLogs(filePath string) []*ScanSummaryLoggerLog {
	logs := make([]*ScanSummaryLoggerLog, 0)
	if l == nil {
		return logs
	}
	for _, log := range l.Logs {
		if log.FilePath == filePath {
			logs = append(logs, log)
		}
	}
	return logs
}
