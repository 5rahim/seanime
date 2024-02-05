package summary

import (
	"fmt"
	"github.com/google/uuid"
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
	LogMetadataEpisodeZero
	LogMetadataNC
	LogMetadataSpecial
	LogMetadataMain
	LogMetadataHydrated
)

type (
	LogType int

	ScanSummaryLogger struct {
		Logs              []*ScanSummaryLog
		LocalFiles        []*entities.LocalFile
		AllMedia          []*entities.NormalizedMedia
		AnilistCollection *anilist.AnimeCollection
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
		ID        string              `json:"id"`
		LocalFile *entities.LocalFile `json:"localFile"`
		Logs      []*ScanSummaryLog   `json:"logs"`
	}

	ScanSummaryGroup struct {
		ID                  string             `json:"id"`
		Files               []*ScanSummaryFile `json:"files"`
		MediaId             int                `json:"mediaId"`
		MediaTitle          string             `json:"mediaTitle"`
		MediaImage          string             `json:"mediaImage"`
		MediaIsInCollection bool               `json:"mediaIsInCollection"` // Whether the media is in the user's AniList collection
	}
)

func NewScanSummaryLogger() *ScanSummaryLogger {
	return &ScanSummaryLogger{
		Logs: make([]*ScanSummaryLog, 0),
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
		ID:     uuid.NewString(),
		Groups: make([]*ScanSummaryGroup, 0),
		//Files:          make([]*ScanSummaryFile, 0),
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

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetched(lf *entities.LocalFile, ms int64, branches int) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Media tree fetched in %dms. Branches: %d", ms, branches)
	l.logType(LogMetadataMediaTreeFetched, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataMediaTreeFetchFailed(lf *entities.LocalFile, err error, ms int64) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Could not fetch media tree: %s. Took %dms", err.Error(), ms)
	l.logType(LogMetadataMediaTreeFetchFailed, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalized(lf *entities.LocalFile, mediaId int, episode int, newEpisode int, newMediaId int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode %d normalized to %d. New media ID: %d. AniDB episode: %s", episode, newEpisode, newMediaId, aniDBEpisode)
	l.logType(LogMetadataEpisodeNormalized, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeNormalizationFailed(lf *entities.LocalFile, err error, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Episode normalization failed: Reason \"%s\". Episode %d. AniDB episode %s", err.Error(), episode, aniDBEpisode)
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

func (l *ScanSummaryLogger) LogMetadataMain(lf *entities.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as main episode. Episode %d. AniDB episode: %s", episode, aniDBEpisode)
	l.logType(LogMetadataMain, lf, msg)
}

func (l *ScanSummaryLogger) LogMetadataEpisodeZero(lf *entities.LocalFile, episode int, aniDBEpisode string) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf("Marked as main episode. Episode %d. AniDB episode set to %s assuming AniDB does not include episode 0 in the episode count.", episode, aniDBEpisode)
	l.logType(LogMetadataEpisodeZero, lf, msg)
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
	case LogMetadataEpisodeZero:
		l.log(lf, "warning", message)
	case LogFileNotMatched:
		l.log(lf, "warning", message)
	}
}

func (l *ScanSummaryLogger) log(lf *entities.LocalFile, level string, message string) {
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

func (l *ScanSummaryLogger) getFileLogs(lf *entities.LocalFile) []*ScanSummaryLog {
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
