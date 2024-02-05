package summary

import (
	"fmt"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
)

const (
	LogFileNotMatched LogType = iota
	LogComparison
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
		Logs              []*ScanSummaryLoggerLog
		LocalFiles        []*entities.LocalFile
		AllMedia          []*entities.NormalizedMedia
		AnilistCollection *anilist.AnimeCollection
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

// HydrateData will hydrate the data needed to generate the summary.
func (l *ScanSummaryLogger) HydrateData(lfs []*entities.LocalFile, media []*entities.NormalizedMedia, anilistCollection *anilist.AnimeCollection) {
	l.LocalFiles = lfs
	l.AllMedia = media
	l.AnilistCollection = anilistCollection
}

func (l *ScanSummaryLogger) GenerateSummary() *ScanSummary {
	if l == nil || l.LocalFiles == nil || l.AllMedia == nil || l.AnilistCollection == nil {
		return nil
	}
	summary := &ScanSummary{
		Groups:         make([]*ScanSummaryGroup, 0),
		Files:          make([]*ScanSummaryFile, 0),
		UnmatchedFiles: make([]*ScanSummaryFile, 0),
	}

	groupsMap := make(map[int][]*ScanSummaryFile)

	// Generate summary files
	for _, lf := range l.LocalFiles {

		if lf.MediaId == 0 {
			summary.UnmatchedFiles = append(summary.UnmatchedFiles, &ScanSummaryFile{
				LocalFile: lf,
				Logs:      l.getFileLogs(lf),
			})
			continue
		}

		summaryFile := &ScanSummaryFile{
			LocalFile: lf,
			Logs:      l.getFileLogs(lf),
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
		for _, m := range l.AllMedia {
			if m.ID == mediaId {
				mediaTitle = m.GetTitleSafe()
				mediaImage = ""
				if m.GetCoverImage() != nil && m.GetCoverImage().GetLarge() != nil {
					mediaImage = *m.GetCoverImage().GetLarge()
				}
				break
			}
		}
		if _, found := l.AnilistCollection.GetListEntryFromMediaId(mediaId); found {
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

func (l *ScanSummaryLogger) LogComparison(lf *entities.LocalFile, algo string, bestTitle string, rating string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Comparison using %s. Best title: \"%s\". Rating: %s", algo, bestTitle, rating)
	l.logType(LogComparison, lf, msg)
}

func (l *ScanSummaryLogger) LogSuccessfullyMatched(lf *entities.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Successfully matched to media %d", mediaId)
	l.logType(LogSuccessfullyMatched, lf, msg)
}

func (l *ScanSummaryLogger) LogFailedMatch(lf *entities.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Failed to match: %s", reason)
	l.logType(LogFailedMatch, lf, msg)
}

func (l *ScanSummaryLogger) LogMatchValidated(lf *entities.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Match validated for media %d", mediaId)
	l.logType(LogMatchValidated, lf, msg)
}

func (l *ScanSummaryLogger) LogUnmatched(lf *entities.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Unmatched: %s", reason)
	l.logType(LogUnmatched, lf, msg)
}

func (l *ScanSummaryLogger) LogFileNotMatched(lf *entities.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Not matched: %s", reason)
	l.logType(LogFileNotMatched, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetched(lf *entities.LocalFile, ms int, requests int, branches string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Media tree fetched in %dms. Requests: %d. Branches: %s", ms, requests, branches)
	l.logType(LogMetadataMediaTreeFetched, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetchFailed(lf *entities.LocalFile, err error, ms int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Could not fetch media tree: %s. Took %dms", err.Error(), ms)
	l.logType(LogMetadataMediaTreeFetchFailed, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalized(lf *entities.LocalFile, mediaId int, episode int, newEpisode int, newMediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode %d normalized to %d. New media ID: %d", episode, newEpisode, newMediaId)
	l.logType(LogMetadataEpisodeNormalized, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalizationFailed(lf *entities.LocalFile, err error) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode normalization failed: %s", err.Error())
	l.logType(LogMetadataEpisodeNormalizationFailed, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataNC(lf *entities.LocalFile) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as NC file")
	l.logType(LogMetadataNC, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataSpecial(lf *entities.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as special episode. Episode %d. AniDB episode: %s", episode, aniDBEpisode)
	l.logType(LogMetadataSpecial, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataHydrated(lf *entities.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Metadata hydrated for media %d", mediaId)
	l.logType(LogMetadataHydrated, lf, msg)
}

func (l *ScanSummaryLogger) logType(logType LogType, lf *entities.LocalFile, message string) {
	if l == nil {
		return
	}
	switch logType {
	case LogComparison:
		l.log(lf, "info", message)
	case LogSuccessfullyMatched:
		l.log(lf, "info", message)
	case LogFailedMatch:
		l.log(lf, "warning", message)
	case LogMatchValidated:
		l.log(lf, "info", message)
	case LogUnmatched:
		l.log(lf, "warning", message)
	case LogMetadataMediaTreeFetched:
		l.log(lf, "info", message)
	case LogMetadataMediaTreeFetchFailed:
		l.log(lf, "error", message)
	case LogMetadataEpisodeNormalized:
		l.log(lf, "info", message)
	case LogMetadataEpisodeNormalizationFailed:
		l.log(lf, "error", message)
	case LogMetadataHydrated:
		l.log(lf, "info", message)
	case LogMetadataNC:
		l.log(lf, "info", message)
	case LogMetadataSpecial:
		l.log(lf, "info", message)
	case LogMetadataMain:
		l.log(lf, "info", message)
	case LogFileNotMatched:
		l.log(lf, "warning", message)
	}
}

func (l *ScanSummaryLogger) log(lf *entities.LocalFile, level string, message string) {
	if l == nil {
		return
	}
	l.Logs = append(l.Logs, &ScanSummaryLoggerLog{
		FilePath: lf.Path,
		Level:    level,
		Message:  message,
	})
}

func (l *ScanSummaryLogger) getFileLogs(lf *entities.LocalFile) []*ScanSummaryLoggerLog {
	logs := make([]*ScanSummaryLoggerLog, 0)
	if l == nil {
		return logs
	}
	for _, log := range l.Logs {
		if lf.HasSamePath(log.FilePath) {
			logs = append(logs, log)
		}
	}
	return logs
}
