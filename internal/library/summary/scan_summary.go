package summary

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"time"

	"github.com/google/uuid"
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
	LogMetadataEpisodeZero
	LogMetadataNC
	LogMetadataSpecial
	LogMetadataMain
	LogMetadataHydrated
	LogPanic
	LogDebug
)

type (
	LogType int

	ScanSummaryLogger struct {
		Logs            []*ScanSummaryLog
		LocalFiles      []*anime.LocalFile
		AllMedia        []*anime.NormalizedMedia
		AnimeCollection *anilist.AnimeCollectionWithRelations
	}

	ScanSummaryLog struct { // Holds a log entry. The log entry will then be used to generate a ScanSummary.
		ID       string `json:"id"`
		FilePath string `json:"filePath"`
		Level    string `json:"level"`
		Message  string `json:"message"`
	}

	ScanSummary struct {
		ID             string              `json:"id"`
		Groups         []*ScanSummaryGroup `json:"groups"`
		UnmatchedFiles []*ScanSummaryFile  `json:"unmatchedFiles"`
	}

	ScanSummaryFile struct {
		ID        string            `json:"id"`
		LocalFile *anime.LocalFile  `json:"localFile"`
		Logs      []*ScanSummaryLog `json:"logs"`
	}

	ScanSummaryGroup struct {
		ID                  string             `json:"id"`
		Files               []*ScanSummaryFile `json:"files"`
		MediaId             int                `json:"mediaId"`
		MediaTitle          string             `json:"mediaTitle"`
		MediaImage          string             `json:"mediaImage"`
		MediaIsInCollection bool               `json:"mediaIsInCollection"` // Whether the media is in the user's AniList collection
	}

	ScanSummaryItem struct { // Database item
		CreatedAt   time.Time    `json:"createdAt"`
		ScanSummary *ScanSummary `json:"scanSummary"`
	}
)

func NewScanSummaryLogger() *ScanSummaryLogger {
	return &ScanSummaryLogger{
		Logs: make([]*ScanSummaryLog, 0),
	}
}

// HydrateData will hydrate the data needed to generate the summary.
func (l *ScanSummaryLogger) HydrateData(lfs []*anime.LocalFile, media []*anime.NormalizedMedia, animeCollection *anilist.AnimeCollectionWithRelations) {
	l.LocalFiles = lfs
	l.AllMedia = media
	l.AnimeCollection = animeCollection
}

func (l *ScanSummaryLogger) GenerateSummary() *ScanSummary {
	if l == nil || l.LocalFiles == nil || l.AllMedia == nil || l.AnimeCollection == nil {
		return nil
	}
	summary := &ScanSummary{
		ID:             uuid.NewString(),
		Groups:         make([]*ScanSummaryGroup, 0),
		UnmatchedFiles: make([]*ScanSummaryFile, 0),
	}

	groupsMap := make(map[int][]*ScanSummaryFile)

	// Generate summary files
	for _, lf := range l.LocalFiles {

		if lf.MediaId == 0 {
			summary.UnmatchedFiles = append(summary.UnmatchedFiles, &ScanSummaryFile{
				ID:        uuid.NewString(),
				LocalFile: lf,
				Logs:      l.getFileLogs(lf),
			})
			continue
		}

		summaryFile := &ScanSummaryFile{
			ID:        uuid.NewString(),
			LocalFile: lf,
			Logs:      l.getFileLogs(lf),
		}

		//summary.Files = append(summary.Files, summaryFile)

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
				mediaTitle = m.GetPreferredTitle()
				mediaImage = ""
				if m.GetCoverImage() != nil && m.GetCoverImage().GetLarge() != nil {
					mediaImage = *m.GetCoverImage().GetLarge()
				}
				break
			}
		}
		if _, found := l.AnimeCollection.GetListEntryFromMediaId(mediaId); found {
			mediaIsInCollection = true
		}

		summary.Groups = append(summary.Groups, &ScanSummaryGroup{
			ID:                  uuid.NewString(),
			Files:               files,
			MediaId:             mediaId,
			MediaTitle:          mediaTitle,
			MediaImage:          mediaImage,
			MediaIsInCollection: mediaIsInCollection,
		})
	}

	return summary
}

func (l *ScanSummaryLogger) LogComparison(lf *anime.LocalFile, algo string, bestTitle string, ratingType string, rating string) {
	if l == nil {
		return
	}

	msg := fmt.Sprintf("Comparison using %s. Best title: \"%s\". %s: %s", algo, bestTitle, ratingType, rating)
	l.logType(LogComparison, lf, msg)
}

func (l *ScanSummaryLogger) LogSuccessfullyMatched(lf *anime.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Successfully matched to media %d", mediaId)
	l.logType(LogSuccessfullyMatched, lf, msg)
}

func (l *ScanSummaryLogger) LogPanic(lf *anime.LocalFile, stackTrace string) {
	if l == nil {
		return
	}
	//msg := fmt.Sprintf("Panic occurred, please report this issue on the GitHub repository with the stack trace printed in the terminal")
	l.logType(LogPanic, lf, "PANIC! "+stackTrace)
}

func (l *ScanSummaryLogger) LogFailedMatch(lf *anime.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Failed to match: %s", reason)
	l.logType(LogFailedMatch, lf, msg)
}

func (l *ScanSummaryLogger) LogMatchValidated(lf *anime.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Match validated for media %d", mediaId)
	l.logType(LogMatchValidated, lf, msg)
}

func (l *ScanSummaryLogger) LogUnmatched(lf *anime.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Unmatched: %s", reason)
	l.logType(LogUnmatched, lf, msg)
}

func (l *ScanSummaryLogger) LogFileNotMatched(lf *anime.LocalFile, reason string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Not matched: %s", reason)
	l.logType(LogFileNotMatched, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetched(lf *anime.LocalFile, ms int64, branches int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Media tree fetched in %dms. Branches: %d", ms, branches)
	l.logType(LogMetadataMediaTreeFetched, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetchFailed(lf *anime.LocalFile, err error, ms int64) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Could not fetch media tree: %s. Took %dms", err.Error(), ms)
	l.logType(LogMetadataMediaTreeFetchFailed, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalized(lf *anime.LocalFile, mediaId int, episode int, newEpisode int, newMediaId int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode %d normalized to %d. New media ID: %d. AniDB episode: %s", episode, newEpisode, newMediaId, aniDBEpisode)
	l.logType(LogMetadataEpisodeNormalized, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalizationFailed(lf *anime.LocalFile, err error, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode normalization failed. Reason \"%s\". Episode %d. AniDB episode %s", err.Error(), episode, aniDBEpisode)
	l.logType(LogMetadataEpisodeNormalizationFailed, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataNC(lf *anime.LocalFile) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as NC file")
	l.logType(LogMetadataNC, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataSpecial(lf *anime.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as Special episode. Episode %d. AniDB episode: %s", episode, aniDBEpisode)
	l.logType(LogMetadataSpecial, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMain(lf *anime.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as main episode. Episode %d. AniDB episode: %s", episode, aniDBEpisode)
	l.logType(LogMetadataMain, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeZero(lf *anime.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as main episode. Episode %d. AniDB episode set to %s assuming AniDB does not include episode 0 in the episode count.", episode, aniDBEpisode)
	l.logType(LogMetadataEpisodeZero, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataHydrated(lf *anime.LocalFile, mediaId int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Metadata hydrated for media %d", mediaId)
	l.logType(LogMetadataHydrated, lf, msg)
}

func (l *ScanSummaryLogger) LogDebug(lf *anime.LocalFile, message string) {
	if l == nil {
		return
	}
	l.log(lf, "info", message)
}

func (l *ScanSummaryLogger) logType(logType LogType, lf *anime.LocalFile, message string) {
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
	case LogMetadataEpisodeZero:
		l.log(lf, "warning", message)
	case LogFileNotMatched:
		l.log(lf, "warning", message)
	case LogPanic:
		l.log(lf, "error", message)
	case LogDebug:
		l.log(lf, "info", message)
	}
}

func (l *ScanSummaryLogger) log(lf *anime.LocalFile, level string, message string) {
	if l == nil {
		return
	}
	l.Logs = append(l.Logs, &ScanSummaryLog{
		ID:       uuid.NewString(),
		FilePath: lf.Path,
		Level:    level,
		Message:  message,
	})
}

func (l *ScanSummaryLogger) getFileLogs(lf *anime.LocalFile) []*ScanSummaryLog {
	logs := make([]*ScanSummaryLog, 0)
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
