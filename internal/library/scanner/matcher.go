package scanner

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/library/summary"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

var candidatesPool = sync.Pool{
	New: func() interface{} {
		return make([]*anime.NormalizedMedia, 0, 500)
	},
}

var seenCandidatesPool = sync.Pool{
	New: func() interface{} {
		return make(map[int]struct{}, 500)
	},
}

var synonymOwnersPool = sync.Pool{
	New: func() interface{} {
		return make(map[string][]*anime.NormalizedMedia, 500)
	},
}

var ignoredSynonymsPool = sync.Pool{
	New: func() interface{} {
		return make(map[int]map[string]struct{}, 500)
	},
}

type Matcher struct {
	LocalFiles        []*anime.LocalFile
	MediaContainer    *MediaContainer
	Logger            *zerolog.Logger
	ScanLogger        *ScanLogger
	ScanSummaryLogger *summary.ScanSummaryLogger // optional
	Algorithm         string
	Threshold         float64
	Debug             bool
	UseLegacyMatching bool
	Config            *Config
	matchingRules     map[string]*compiledMatchingRule
}

type compiledMatchingRule struct {
	regex *regexp.Regexp
	rule  *MatchingRule
}

var (
	ErrNoLocalFiles = errors.New("[matcher] no local files")
)

const (
	// scoring for title matching (highest prio)
	scoreTitleExact         = 15.0 // exact match (case-insensitive)
	scoreTitleDecay         = 5.0  // exact match penalty for short titles
	scoreTitleNormalizedEq  = 14.0 // normalized strings are identical
	scoreTitleTokenPerfect  = 12.0 // all significant tokens match, same count
	scoreTitleTokenComplete = 10.0 // all tokens of one are in the other (subset)
	scoreTitleTokenHigh     = 8.0  // 90%+ token match ratio
	scoreTitleTokenMedium   = 5.0  // 70%+ token match ratio
	scoreTitleFuzzyHigh     = 6.0  // 90%+ fuzzy similarity
	scoreTitleFuzzyMedium   = 3.0  // 80%+ fuzzy similarity
	scoreTitleBaseMatch     = 4.0  // base title (without season/part) matches
	scoreTitleMainBonus     = 2.0  // bonus for matching a main title (romaji/english)

	// season/part scoring
	scoreSeasonExactMatch      = 5.0  // season numbers match exactly
	scoreSeasonMismatch        = -8.0 // season numbers explicitly don't match
	scoreSeasonImplicitPenalty = -3.0 // file has season > 1 but media doesn't indicate season
	scorePartExactMatch        = 3.0  // part numbers match exactly
	scorePartMismatch          = -5.0 // part numbers explicitly don't match

	// format type scoring (OVA/Special/Movie detection)
	scoreFormatMatch    = 5.0  // file indicates OVA/Special and media format matches
	scoreFormatMismatch = -5.0 // file indicates OVA/Special but media is regular TV

	// year matching
	scoreYearExactMatch = 4.0   // years match exactly
	scoreYearCloseMatch = 2.0   // years within 1 year
	scoreYearMismatch   = -10.0 // years differ by more than 1

	// threashold
	thresholdMatch             = 6.0  // minimum score to consider a match
	thresholdTokenHigh         = 0.9  // 90% token overlap
	thresholdTokenMedium       = 0.7  // 70% token overlap
	thresholdFuzzyHigh         = 0.90 // 90% fuzzy similarity
	thresholdFuzzyMedium       = 0.80 // 80% fuzzy similarity
	thresholdBaseTitleFuzzy    = 0.85 // for base title comparison
	thresholdNormalTitleLength = 4
)

// MatchLocalFilesWithMedia will match each anime.LocalFile with a specific anilist.BaseAnime and modify the LocalFile's `mediaId`
func (m *Matcher) MatchLocalFilesWithMedia() error {

	if m.Threshold == 0 {
		m.Threshold = 0.5
	}

	start := time.Now()

	if len(m.LocalFiles) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).Msg("No local files")
		}
		return ErrNoLocalFiles
	}
	if len(m.MediaContainer.NormalizedMedia) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).Msg("No media fed into the matcher")
		}
		return errors.New("[matcher] no media fed into the matcher")
	}

	m.Logger.Debug().Msg("matcher: Starting matching process")

	m.precompileRules()

	// Invoke ScanMatchingStarted hook
	event := &ScanMatchingStartedEvent{
		LocalFiles:      m.LocalFiles,
		NormalizedMedia: m.MediaContainer.NormalizedMedia,
		Algorithm:       m.Algorithm,
		Threshold:       m.Threshold,
	}
	_ = hook.GlobalHookManager.OnScanMatchingStarted().Trigger(event)
	m.LocalFiles = event.LocalFiles
	m.MediaContainer.NormalizedMedia = event.NormalizedMedia
	m.Algorithm = event.Algorithm
	m.Threshold = event.Threshold

	if event.DefaultPrevented {
		m.Logger.Debug().Msg("matcher: Match stopped by hook")
		return nil
	}

	// process files concurrently using a worker pool
	numWorkers := runtime.NumCPU()
	if numWorkers > len(m.LocalFiles) {
		numWorkers = len(m.LocalFiles)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	fileChan := make(chan *anime.LocalFile, len(m.LocalFiles))
	var wg sync.WaitGroup

	// start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for localFile := range fileChan {
				if m.UseLegacyMatching {
					m.matchLocalFileLegacy(localFile)
				} else {
					m.matchLocalFile(localFile)
				}
			}
		}()
	}

	// send work to workers
	for _, lf := range m.LocalFiles {
		fileChan <- lf
	}
	close(fileChan)

	wg.Wait()

	m.Logger.Debug().Msg("matcher: Finished matching process")

	// unused for now
	//m.validateMatches()

	// Invoke ScanMatchingCompleted hook
	completedEvent := &ScanMatchingCompletedEvent{
		LocalFiles: m.LocalFiles,
	}
	_ = hook.GlobalHookManager.OnScanMatchingCompleted().Trigger(completedEvent)
	m.LocalFiles = completedEvent.LocalFiles

	if m.ScanLogger != nil {
		// count unmatched files
		unmatchedCount := 0
		for _, lf := range m.LocalFiles {
			if lf.MediaId == 0 {
				unmatchedCount++
			}
		}
		m.ScanLogger.LogMatcher(zerolog.InfoLevel).
			Int64("ms", time.Since(start).Milliseconds()).
			Int("files", len(m.LocalFiles)).
			Int("unmatched", unmatchedCount).
			Msg("Finished matching process")
	}

	return nil
}

// matchLocalFile uses a matching algorithm to match a local file to the best fitting media.
// It uses multi-layer approach: exact, normalized, token-based, and fuzzy.
// It also considers year, season, and part information for accurate differentiation.
func (m *Matcher) matchLocalFile(lf *anime.LocalFile) {
	defer util.HandlePanicInModuleThenS("library/scanner/matchLocalFile", func(s string) {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msgf("Panic occurred in matchLocalFile: %v", s)
		}
		m.ScanSummaryLogger.LogPanic(lf, s)
	})

	// Check if the local file has already been matched
	if lf.MediaId != 0 {
		return
	}

	if m.applyMatcingRule(lf) {
		return
	}

	// Check if the local file or any of its folders have a parsed title.
	// If not, we skip it,
	if lf.GetParsedTitle() == "" {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No parsed title found")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No parsed title found")
		return
	}

	// Get title variations for the file
	titleVariations := lf.GetTitleVariations()

	if len(titleVariations) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No title variations found")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No title variations found")
		return
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Interface("titleVariations", titleVariations).
			Msg("Found title variations")
	}
	m.ScanSummaryLogger.LogDebug(lf, fmt.Sprintf("Title variations: %s", m.getLogVariations(titleVariations)))

	// identify extra titles (parent folder titles)
	// they'll be penalized later so they don't outcompete more specific files/folders
	// this is useful for cases like: /Monogatari Series/Kizumonogatari/Kizumonogatari I - Tekketsu-hen (1920x1080 Blu-Ray FLAC).mkv
	// where "Monogatari Series" can cause mismatches if it has the same weight as "Kizumonogatari"
	primaryTitlesMap := make(map[string]struct{})
	extraTitlesMap := make(map[string]struct{})
	if len(lf.ParsedData.Title) > 0 {
		primaryTitlesMap[lf.ParsedData.Title] = struct{}{}
	}
	if ft := lf.GetFolderTitle(); len(ft) > 0 && ft != lf.ParsedData.Title {
		primaryTitlesMap[ft] = struct{}{}
	}
	for _, ft := range lf.GetAllFolderTitles() {
		if _, ok := primaryTitlesMap[ft]; !ok {
			extraTitlesMap[ft] = struct{}{}
		}
	}

	// Normalize all title variations for matching
	normalizedVariations := make([]*NormalizedTitle, 0, len(titleVariations))
	for _, t := range titleVariations {
		if t != nil && *t != "" {
			nt := NormalizeTitle(*t)
			_, nt.IsExtra = extraTitlesMap[*t]
			normalizedVariations = append(normalizedVariations, nt)
		}
	}

	if len(normalizedVariations) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No valid title variations")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No valid title variations")
		return
	}

	// Extract file's season, part, and year info (from filename/folder and parsed data)
	fileSeason := getFileSeason(lf)
	filePart := getFilePart(lf)
	fileYear := getFileYear(lf)
	fileFormatType := getFileFormatType(lf)

	// Also try to extract from title variations
	for _, nv := range normalizedVariations {
		if fileSeason == -1 && nv.Season > 0 {
			fileSeason = nv.Season
		}
		if filePart == -1 && nv.Part > 0 {
			filePart = nv.Part
		}
		if fileYear == -1 && nv.Year > 0 {
			fileYear = nv.Year
		}
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Int("season", fileSeason).
			Int("part", filePart).
			Int("year", fileYear).
			Msg("Extracted metadata")
	}
	m.ScanSummaryLogger.LogDebug(lf, fmt.Sprintf("Extracted metadata: season=%d, part=%d, year=%d", fileSeason, filePart, fileYear))

	bestScore := 0.0
	var bestMedia *anime.NormalizedMedia

	// We filter candidates using token index
	// Instead of iterating over all media, we only check media that share at least one significant token
	var candidates []*anime.NormalizedMedia

	// Get pooled resources
	pooledCandidates := candidatesPool.Get().([]*anime.NormalizedMedia)
	pooledSeen := seenCandidatesPool.Get().(map[int]struct{})

	// Reset pooled resources
	pooledCandidates = pooledCandidates[:0]

	defer func() {
		// Clear map
		for k := range pooledSeen {
			delete(pooledSeen, k)
		}
		seenCandidatesPool.Put(pooledSeen)

		// Return slice
		candidatesPool.Put(pooledCandidates)
	}()

	if len(m.MediaContainer.TokenIndex) > 0 {
		for _, nv := range normalizedVariations {
			// Get significant tokens from the normalized title variation
			tokens := GetSignificantTokens(nv.Tokens)
			for _, token := range tokens {
				if nMedia, ok := m.MediaContainer.TokenIndex[token]; ok {
					for _, media := range nMedia {
						if _, seen := pooledSeen[media.ID]; !seen {
							pooledCandidates = append(pooledCandidates, media)
							pooledSeen[media.ID] = struct{}{}
						}
					}
				}
			}
		}
		candidates = pooledCandidates
	}

	// Fallback to all media if no candidates found (e.g. no significant tokens or no matches)
	if len(candidates) == 0 {
		candidates = m.MediaContainer.NormalizedMedia
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Int("candidates", len(candidates)).
			Msg("Found candidates")
	}
	m.ScanSummaryLogger.LogDebug(lf, fmt.Sprintf("Found %d candidates", len(candidates)))

	sd := GetEfficientDice()
	defer PutEfficientDice(sd)

	// devnote: causes lower title scoring on some seasons with long titles
	//ignoredSynonyms := m.getIgnoredSynonyms(candidates)
	//defer func() {
	//	// Clear map
	//	for k := range ignoredSynonyms {
	//		delete(ignoredSynonyms, k)
	//	}
	//	ignoredSynonymsPool.Put(ignoredSynonyms)
	//}()
	ignoredSynonyms := map[int]map[string]struct{}{}

	// Process candidates serially
	// devnote: slower than doing it concurrently but we won't abuse goroutines
	for _, media := range candidates {
		currentScore := 0.0

		// use cached normalized titles
		originalMediaTitles, ok := m.MediaContainer.NormalizedTitlesCache[media.ID]
		if !ok || len(originalMediaTitles) == 0 {
			continue
		}

		// Filter out ignored synonyms
		var normalizedMediaTitles []*NormalizedTitle
		if ignored, ok := ignoredSynonyms[media.ID]; ok {
			normalizedMediaTitles = make([]*NormalizedTitle, 0, len(originalMediaTitles))
			for _, t := range originalMediaTitles {
				if !t.IsMain {
					if _, isIgnored := ignored[t.Normalized]; isIgnored {
						continue
					}
				}
				normalizedMediaTitles = append(normalizedMediaTitles, t)
			}
		} else {
			normalizedMediaTitles = originalMediaTitles
		}

		// 1. Title matching (highest prio)
		titleScore := calculateTitleScore(normalizedVariations, normalizedMediaTitles, sd)

		// skip if title score is too low
		if titleScore < 2.0 {
			continue
		}

		currentScore += titleScore

		// 2. Season/Part matching
		mediaSeason, mediaSeasonExplicit, mediaSeasonConfidence := getMediaSeason(media, normalizedMediaTitles)
		mediaPart, mediaPartExplicit := getMediaPart(normalizedMediaTitles)
		seasonPartScore := calculateSeasonPartScore(fileSeason, filePart, mediaSeason, mediaSeasonExplicit, mediaSeasonConfidence, mediaPart, mediaPartExplicit)
		currentScore += seasonPartScore

		// 3. Year comparison
		yearScore := calculateYearScore(fileYear, media, titleScore)
		currentScore += yearScore

		// 4. Base title matching bonus
		baseTitleScore := calculateBaseTitleScore(normalizedVariations, normalizedMediaTitles, sd)
		if baseTitleScore > 0 && seasonPartScore >= 0 {
			currentScore += baseTitleScore
		}

		// 5. Format type matching (OVA/Special/Movie detection)
		formatScore := calculateFormatScore(fileFormatType, media)
		currentScore += formatScore

		if m.Debug {
			m.Logger.Debug().
				Str("filename", lf.Name).
				Int("id", media.ID).
				Str("match", media.GetTitleSafe()).
				Float64("score", currentScore).
				Float64("titleScore", titleScore).
				Float64("baseTitleScore", baseTitleScore).
				Float64("seasonPartScore", seasonPartScore).
				Float64("yearScore", yearScore).
				Float64("formatScore", formatScore).
				Int("season", mediaSeason).
				Int("part", mediaPart).
				Interface("titles", normalizedMediaTitles).
				Msg("matcher: debug")
		}
		if titleScore > 2.0 {
			if m.Config != nil && m.Config.Logs.Verbose {
				if m.ScanLogger != nil {
					m.ScanLogger.LogMatcher(zerolog.DebugLevel).
						Str("filename", lf.Name).
						Int("id", media.ID).
						Str("match", media.GetTitleSafe()).
						Float64("score", currentScore).
						Float64("titleScore", titleScore).
						Float64("baseTitleScore", baseTitleScore).
						Float64("seasonPartScore", seasonPartScore).
						Float64("yearScore", yearScore).
						Float64("formatScore", formatScore).
						Int("season", mediaSeason).
						Int("part", mediaPart).
						Interface("titles", normalizedMediaTitles).
						Msg("Comparison")
				}
			}
		}

		if currentScore > bestScore {
			bestScore = currentScore
			bestMedia = media
		}
	}

	// Invoke ScanLocalFileMatched hook
	event := &ScanLocalFileMatchedEvent{
		LocalFile: lf,
		Score:     bestScore,
		Match:     bestMedia,
		Found:     bestMedia != nil && bestScore >= thresholdMatch,
	}
	_ = hook.GlobalHookManager.OnScanLocalFileMatched().Trigger(event)
	lf = event.LocalFile
	bestMedia = event.Match
	bestScore = event.Score

	// Check if the hook overrode the match
	if event.DefaultPrevented {
		if m.ScanLogger != nil {
			if bestMedia != nil {
				m.ScanLogger.LogMatcher(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Int("id", bestMedia.ID).
					Str("match", bestMedia.GetTitleSafe()).
					Float64("score", bestScore).
					Msg("Hook overrode match")
			} else {
				m.ScanLogger.LogMatcher(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Msg("Hook overrode match, no match found")
			}
		}
		if bestMedia != nil {
			lf.MediaId = bestMedia.ID
			if m.ScanSummaryLogger != nil {
				m.ScanSummaryLogger.LogSuccessfullyMatched(lf, bestMedia.ID)
			}
		}
		return
	}

	// Threshold check
	if bestScore >= thresholdMatch && bestMedia != nil {
		lf.MediaId = bestMedia.ID
		if m.ScanSummaryLogger != nil {
			m.ScanSummaryLogger.LogSuccessfullyMatched(lf, bestMedia.ID)
		}
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Str("match", bestMedia.GetTitleSafe()).
				Int("id", bestMedia.ID).
				Float64("score", bestScore).
				Msg("Best match found")
		}
		m.ScanSummaryLogger.LogDebug(lf, fmt.Sprintf("Candidate %d: %q | score=%.2f titles=%s",
			bestMedia.ID, bestMedia.GetTitleSafe(), bestScore, strings.Join(getMediaTitlesExpanded(bestMedia), ", ")))

	} else {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Float64("score", bestScore).
				Msg("No match found")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, fmt.Sprintf("Score too low: %.2f", bestScore))
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getFileSeason(lf *anime.LocalFile) int {
	// Check parsed data first (S01, S02, etc.)
	if lf.ParsedData.Season != "" {
		if val, ok := util.StringToInt(lf.ParsedData.Season); ok {
			return val
		}
	}

	// Then check folder data for explicit season
	if lf.ParsedFolderData != nil {
		for _, fpd := range lf.ParsedFolderData {
			if fpd.Season != "" {
				if val, ok := util.StringToInt(fpd.Season); ok {
					return val
				}
			}
		}
	}

	// Fallback: Extract season from parsed title (will handle roman numerals like "Title II")
	if lf.ParsedData.Title != "" {
		if season := comparison.ExtractSeasonNumber(lf.ParsedData.Title); season > 0 {
			return season
		}
	}

	// Fallback: Try folder titles
	if lf.ParsedFolderData != nil {
		for _, fpd := range lf.ParsedFolderData {
			if fpd.Title != "" {
				if season := comparison.ExtractSeasonNumber(fpd.Title); season > 0 {
					return season
				}
			}
		}
	}

	return -1
}

func getFilePart(lf *anime.LocalFile) int {
	// Check parsed data first
	if lf.ParsedData.Part != "" {
		if val, ok := util.StringToInt(lf.ParsedData.Part); ok {
			return val
		}
	}

	// Then check folder data
	if lf.ParsedFolderData != nil {
		for _, fpd := range lf.ParsedFolderData {
			if fpd.Part != "" {
				if val, ok := util.StringToInt(fpd.Part); ok {
					return val
				}
			}
		}
	}

	// Try to extract part from parsed title (handles "Part 2", "Cour 2")
	if lf.ParsedData.Title != "" {
		if part := ExtractPartNumber(lf.ParsedData.Title); part > 0 {
			return part
		}
	}

	// Try folder titles
	if lf.ParsedFolderData != nil {
		for _, fpd := range lf.ParsedFolderData {
			if fpd.Title != "" {
				if part := ExtractPartNumber(fpd.Title); part > 0 {
					return part
				}
			}
		}
	}

	return -1
}

func getFileYear(lf *anime.LocalFile) int {
	if lf.ParsedData.Year != "" {
		if val, ok := util.StringToInt(lf.ParsedData.Year); ok {
			return val
		}
	}
	return -1
}

// calculateFormatScore returns a score adjustment based on whether the file's detected format type
// matches the media's format.
// i.e. differentiate main series from specials when the file clearly indicates it's a special.
func calculateFormatScore(fileFormat fileFormatType, media *anime.NormalizedMedia) float64 {
	if fileFormat == fileFormatUnknown || media.Format == nil {
		return 0
	}

	mediaFormat := *media.Format

	switch fileFormat {
	case fileFormatOVA:
		if mediaFormat == anilist.MediaFormatOva || mediaFormat == anilist.MediaFormatSpecial {
			return scoreFormatMatch
		}
		if mediaFormat == anilist.MediaFormatTv || mediaFormat == anilist.MediaFormatTvShort {
			return scoreFormatMismatch
		}

	case fileFormatSpecial:
		if mediaFormat == anilist.MediaFormatSpecial || mediaFormat == anilist.MediaFormatOva {
			return scoreFormatMatch
		}
		if mediaFormat == anilist.MediaFormatTv || mediaFormat == anilist.MediaFormatTvShort {
			return scoreFormatMismatch
		}

	case fileFormatMovie:
		if mediaFormat == anilist.MediaFormatMovie {
			return scoreFormatMatch
		}
		if mediaFormat == anilist.MediaFormatTv || mediaFormat == anilist.MediaFormatTvShort {
			return scoreFormatMismatch
		}

	case fileFormatNC:
		// NC content typically belongs to the main series, don't penalize
		return 0
	default:
	}

	return 0
}

// getMediaTitlesExpanded returns all titles for a media, including synonyms
func getMediaTitlesExpanded(media *anime.NormalizedMedia) []string {
	titles := make([]string, 0, 10)
	for _, t := range media.GetAllTitles() {
		if t != nil && *t != "" {
			titles = append(titles, *t)
		}
	}
	return titles
}

// getMediaSeason returns the media's season number, whether it's explicit (from main title),
// and a confidence level 0-1 indicating how strongly the synonyms agree on the season.
// 1 = main/explicit seasons, <1 = ratio of agreeing synonyms.
func getMediaSeason(media *anime.NormalizedMedia, normalizedTitles []*NormalizedTitle) (int, bool, float64) {
	// Check from media
	// This check is stricter because it filters out synonyms that don't explicitly mention a season keyword
	// e.g. "KonoSuba 3" will be ignored but other checks below can detect it
	if season := media.GetPossibleSeasonNumber(); season > 0 {
		return season, true, 1.0
	}

	// Check from normalized titles (main titles first)
	for _, nt := range normalizedTitles {
		if nt.IsMain && nt.Season > 0 {
			return nt.Season, true, 1.0
		}
	}

	// Check non-main synonyms with confidence based on agreement
	// e.g. Brotherhood has "FMA 2" (season 2) but 35+ other synonyms without season → low confidence
	// e.g. KnY S3 "Katanakaji no Sato-hen" has "KnY 3" → similar low count but still useful for matching
	nonMainTotal := 0
	nonMainWithSeason := 0
	nonMainSeason := -1
	for _, nt := range normalizedTitles {
		if !nt.IsMain {
			nonMainTotal++
			if nt.Season > 0 {
				nonMainWithSeason++
				if nonMainSeason == -1 {
					nonMainSeason = nt.Season
				}
			}
		}
	}
	if nonMainWithSeason > 0 && nonMainTotal > 0 {
		confidence := float64(nonMainWithSeason) / float64(nonMainTotal)
		return nonMainSeason, false, confidence
	}

	// Check from all media titles (fallback using ExtractSeasonNumber)
	allTitles := getMediaTitlesExpanded(media)
	if len(allTitles) > 0 {
		titlesWithSeason := 0
		firstSeason := -1
		for _, title := range allTitles {
			if season := comparison.ExtractSeasonNumber(title); season > 0 {
				titlesWithSeason++
				if firstSeason == -1 {
					firstSeason = season
				}
			}
		}
		if titlesWithSeason > 0 {
			confidence := float64(titlesWithSeason) / float64(len(allTitles))
			return firstSeason, false, confidence
		}
	}

	return -1, false, 0.0
}

func getMediaPart(normalizedTitles []*NormalizedTitle) (int, bool) {
	for _, nt := range normalizedTitles {
		if nt.IsMain && nt.Part > 0 {
			return nt.Part, true
		}
	}
	for _, nt := range normalizedTitles {
		if !nt.IsMain && nt.Part > 0 {
			return nt.Part, false
		}
	}
	return -1, false
}

func calculateTitleScore(
	fileVariations []*NormalizedTitle,
	mediaTitles []*NormalizedTitle,
	sd *EfficientDice,
) float64 {
	bestScore := 0.0

	for _, fv := range fileVariations {
		for _, mt := range mediaTitles {
			score := compareTitles(fv, mt, sd)
			// apply a small penalty to far away folder titles (extra titles)
			// so they don't outcompete more specific files/folders
			if fv.IsExtra {
				score -= 2.0
			}
			if score > bestScore {
				bestScore = score
			}
		}
	}

	return bestScore
}

func compareTitles(file, media *NormalizedTitle, sd *EfficientDice) float64 {
	// Exact match
	if strings.EqualFold(file.Original, media.Original) {
		score := scoreTitleExact
		// if both titles match but they are short, penalize
		if len(file.Original) < thresholdNormalTitleLength {
			score -= scoreTitleDecay
		} else {
			// If the titles aren't short, add a complexity bonus the more complex the file title is
			// i.e. the more significant tokens it has
			score += getComplexityBonus(file.Tokens)
		}
		if media.IsMain {
			score += scoreTitleMainBonus
		}
		return score
	}

	// Normalized string match
	if file.Normalized == media.Normalized {
		score := scoreTitleNormalizedEq
		// penalize if the titles are short
		if len(file.Normalized) < thresholdNormalTitleLength {
			score -= scoreTitleDecay // 14.0 -> 9.0
		} else {
			// add complexity bonus the more complex the file title is
			score += getComplexityBonus(file.Tokens)
		}
		if media.IsMain {
			score += scoreTitleMainBonus
		}
		return score
	}

	// Token-based matching
	tokenScore := compareTokens(file.Tokens, media.Tokens)
	if tokenScore >= scoreTitleTokenPerfect {
		if media.IsMain {
			tokenScore += scoreTitleMainBonus
		}
		return tokenScore
	}

	// Fuzzy matching using Sorensen Dice
	fuzzyScore := sd.Compare(file.Normalized, media.Normalized)
	if fuzzyScore >= thresholdFuzzyHigh {
		score := scoreTitleFuzzyHigh + (fuzzyScore-thresholdFuzzyHigh)*5 // Bonus for higher scores
		if media.IsMain {
			score += scoreTitleMainBonus
		}
		return score
	}
	if fuzzyScore >= thresholdFuzzyMedium {
		score := scoreTitleFuzzyMedium + (fuzzyScore-thresholdFuzzyMedium)*5
		if media.IsMain {
			score += scoreTitleMainBonus
		}
		return score
	}

	scaledFuzzy := fuzzyScore * 5
	if tokenScore > scaledFuzzy {
		if media.IsMain {
			return tokenScore + scoreTitleMainBonus
		}
		return tokenScore
	}
	// Fallback to fuzzy match
	// Only add the main title bonus if the fuzzy match is decent or if we had a strong token match
	// This prevents cases where a very low fuzzy score (random words) + bonus results in a misleading score
	if media.IsMain {
		if scaledFuzzy > 2.5 || HasStrongMatch(file.Tokens, media.Tokens) {
			return scaledFuzzy + scoreTitleMainBonus
		}
	}
	return scaledFuzzy
}

func compareTokens(fileTokens, mediaTokens []string) float64 {
	if len(fileTokens) == 0 || len(mediaTokens) == 0 {
		return 0.0
	}

	// Get pooled slices for significant tokens
	fileSigPtr := getStringSlice()
	mediaSigPtr := getStringSlice()
	defer func() {
		putStringSlice(fileSigPtr)
		putStringSlice(mediaSigPtr)
	}()

	// Filter out noise words using pooled slices
	fileSig := getSignificantTokensInto(fileTokens, *fileSigPtr)
	mediaSig := getSignificantTokensInto(mediaTokens, *mediaSigPtr)

	// Fallback to all tokens if no significant tokens
	if len(fileSig) == 0 {
		fileSig = fileTokens
	}
	if len(mediaSig) == 0 {
		mediaSig = mediaTokens
	}

	// Perfect match
	if ContainsAllTokens(fileSig, mediaSig) && ContainsAllTokens(mediaSig, fileSig) {
		score := scoreTitleTokenPerfect
		// penalize if matched on a single short token
		if len(fileSig) == 1 && len(fileSig[0]) < 4 {
			score -= scoreTitleDecay // 12.0 -> 7.0
		} else {
			score += getComplexityBonusFromSlice(fileSig)
		}
		return score
	}

	// Subset match
	// require at least 2 tokens in the smaller set to avoid false positives
	minTokens := len(fileSig)
	if len(mediaSig) < minTokens {
		minTokens = len(mediaSig)
	}
	if minTokens >= 2 {
		if ContainsAllTokens(fileSig, mediaSig) || ContainsAllTokens(mediaSig, fileSig) {
			return scoreTitleTokenComplete
		}
	}

	// Token overlap
	ratio := WeightedTokenMatchRatio(fileSig, mediaSig)

	// If the ratio is non-zero, check if we have at least one strong match (non-year, non-noise)
	// This prevents matches based solely on year, which can lead to false positives
	if ratio > 0 && !HasStrongMatch(fileSig, mediaSig) {
		return 0.0
	}

	if ratio >= thresholdTokenHigh {
		return scoreTitleTokenHigh
	}
	if ratio >= thresholdTokenMedium {
		return scoreTitleTokenMedium
	}

	return ratio * scoreTitleTokenMedium
}

func getComplexityBonus(tokens []string) float64 {
	sig := GetSignificantTokens(tokens)
	return getComplexityBonusFromSlice(sig)
}

// getComplexityBonusFromSlice calculates complexity bonus from already filtered significant tokens
func getComplexityBonusFromSlice(sigTokens []string) float64 {
	count := len(sigTokens)
	if count >= 3 {
		return 3.0 // Cap bonus
	}
	if count == 2 {
		return 1.0
	}
	return 0.0
}

func calculateSeasonPartScore(fileSeason, filePart, mediaSeason int, mediaSeasonExplicit bool, mediaSeasonConfidence float64, mediaPart int, mediaPartExplicit bool) float64 {
	score := 0.0

	// Season scoring
	if fileSeason > 0 && mediaSeason > 0 {
		if fileSeason == mediaSeason {
			score += scoreSeasonExactMatch
		} else {
			if mediaSeasonExplicit {
				score += scoreSeasonMismatch // Heavy penalty for season mismatch
			} else {
				// scale the mismatch penalty by confidence
				// e.g. Brotherhood "FMA 2" is 5/40 synonyms (confidence=0.125) → penalty = -8 * 0.125 = -1
				// vs a media where most synonyms agree on a season → higher penalty
				score += scoreSeasonMismatch * mediaSeasonConfidence
			}
		}
	} else if fileSeason > 1 && mediaSeason <= 0 {
		// File has explicit season > 1 but media doesn't indicate a season
		// This is sus, possibly wrong match
		score += scoreSeasonImplicitPenalty
	} else if fileSeason == 1 && mediaSeason <= 0 {
		// File is season 1, media doesn't specify, this is fine
		score += 1.0
	} else if fileSeason <= 0 && mediaSeason > 1 {
		// File has no season indicator but media is explicitly season 2+

		if mediaSeasonExplicit {
			// Main title implies season, so file should have it
			score += scoreSeasonImplicitPenalty
		}
	}

	// Part scoring
	if filePart > 0 && mediaPart > 0 {
		if filePart == mediaPart {
			score += scorePartExactMatch
		} else {
			if mediaPartExplicit {
				score += scorePartMismatch
			} else {
				score += scorePartMismatch / 2.0
			}
		}
	} else if filePart > 0 && mediaPart <= 0 {
		// File has part but media doesn't, slight penalty
		score -= 1.0
	} else if filePart <= 0 && mediaPart > 0 {
		// File has no part but media does, penalty since we prefer matching
		// to media without explicit part when file doesn't specify one.
		// This prevents "Title S2" from matching "Title II Part 2" over "Title II"
		if mediaPartExplicit {
			score -= 2.0
		} else {
			score -= 1.0
		}
	}

	return score
}

func calculateYearScore(fileYear int, media *anime.NormalizedMedia, titleScore float64) float64 {
	if fileYear <= 0 {
		return 0.0
	}

	if media.StartDate == nil || media.StartDate.Year == nil {
		return 0.0
	}

	mediaYear := *media.StartDate.Year

	if fileYear == mediaYear {
		// Scale the bonus based on title score confidence
		// Max bonus is scoreYearExactMatch
		scale := math.Min(1.0, titleScore/thresholdMatch)
		return scoreYearExactMatch * scale
	}
	if math.Abs(float64(fileYear-mediaYear)) <= 1 {
		// Scale the bonus based on title score confidence
		scale := math.Min(1.0, titleScore/thresholdMatch)
		return scoreYearCloseMatch * scale
	}
	return scoreYearMismatch
}

// calculateBaseTitleScore compares base titles (without season/part markers).
func calculateBaseTitleScore(
	fileVariations []*NormalizedTitle,
	mediaTitles []*NormalizedTitle,
	sd *EfficientDice,
) float64 {
	bestScore := 0.0

	for _, fv := range fileVariations {
		if fv.Normalized == "" {
			continue
		}
		for _, mt := range mediaTitles {
			if mt.Normalized == "" {
				continue
			}

			// Compare clean base titles
			if fv.Normalized == mt.Normalized {
				return scoreTitleBaseMatch
			}

			fuzzyScore := sd.Compare(fv.Normalized, mt.Normalized)
			if fuzzyScore >= thresholdBaseTitleFuzzy {
				score := scoreTitleBaseMatch * fuzzyScore
				if score > bestScore {
					bestScore = score
				}
			}
		}
	}

	return bestScore
}

var builderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

func (m *Matcher) getLogVariations(titleVariations []*string) string {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.WriteByte('[')
	for i, t := range titleVariations {
		if i > 0 {
			buf.WriteString(", ")
		}
		_, _ = fmt.Fprintf(buf, "%q", *t)
	}
	buf.WriteByte(']')

	return buf.String()
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Matcher) precompileRules() {
	defer util.HandlePanicInModuleThenS("scanner/matcher/precompileRules", func(stackTrace string) {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Msg("Panic occurred, when compiling matching rules")
		}
	})

	if m.Config == nil || len(m.Config.Matching.Rules) == 0 {
		return
	}

	for _, rule := range m.Config.Matching.Rules {
		if rule.Pattern == "" || rule.MediaID == 0 {
			continue
		}

		rgx, err := regexp.Compile(rule.Pattern)
		if err != nil {
			if m.ScanLogger != nil {
				m.ScanLogger.LogMatcher(zerolog.WarnLevel).
					Str("pattern", rule.Pattern).
					Msg("Config: Invalid regex pattern")
			}
			continue
		}
		rgx.Longest()

		if m.matchingRules == nil {
			m.matchingRules = make(map[string]*compiledMatchingRule)
		}

		m.matchingRules[rule.Pattern] = &compiledMatchingRule{
			regex: rgx,
			rule:  rule,
		}
	}
}

func (m *Matcher) applyMatcingRule(lf *anime.LocalFile) bool {
	defer util.HandlePanicInModuleThenS("scanner/matcher/applyMatcingRule", func(stackTrace string) {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("Panic occurred when applying matching rule")
		}
		m.ScanSummaryLogger.LogPanic(lf, stackTrace)
	})

	for _, rule := range m.matchingRules {
		if rule.regex.MatchString(lf.Path) {
			lf.MediaId = rule.rule.MediaID

			if m.ScanLogger != nil {
				var title string
				for _, media := range m.MediaContainer.NormalizedMedia {
					if media.ID == rule.rule.MediaID {
						title = media.GetTitleSafe()
						break
					}
				}
				m.ScanLogger.LogMatcher(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Int("id", rule.rule.MediaID).
					Str("match", title).
					Str("rule", rule.rule.Pattern).
					Msg("Matched by rule")
			}
			if m.ScanSummaryLogger != nil {
				m.ScanSummaryLogger.LogSuccessfullyMatched(lf, rule.rule.MediaID)
			}
			return true
		}
	}

	return false
}

func (m *Matcher) getIgnoredSynonyms(candidates []*anime.NormalizedMedia) map[int]map[string]struct{} {
	// Filter out synonyms that are shared between multiple candidates
	// We keep the synonym only for the candidate with the shortest main title
	// This helps avoiding false positives when a synonym is overly generic or shared
	synonymOwners := synonymOwnersPool.Get().(map[string][]*anime.NormalizedMedia)
	defer func() {
		// Clear map
		for k := range synonymOwners {
			delete(synonymOwners, k)
		}
		synonymOwnersPool.Put(synonymOwners)
	}()
	for _, media := range candidates {
		if titles, ok := m.MediaContainer.NormalizedTitlesCache[media.ID]; ok {
			for _, t := range titles {
				if !t.IsMain && len(t.Normalized) > 0 {
					synonymOwners[t.Normalized] = append(synonymOwners[t.Normalized], media)
				}
			}
		}
	}

	ignoredSynonyms := ignoredSynonymsPool.Get().(map[int]map[string]struct{})

	for synonym, owners := range synonymOwners {
		if len(owners) > 1 {
			var bestCandidate *anime.NormalizedMedia
			shortestTitleLen := 99999

			for _, owner := range owners {
				titleLen := len(owner.GetTitleSafe())
				if titleLen < shortestTitleLen {
					shortestTitleLen = titleLen
					bestCandidate = owner
				} else if titleLen == shortestTitleLen {
					if bestCandidate == nil || owner.ID < bestCandidate.ID {
						bestCandidate = owner
					}
				}
			}

			if bestCandidate == nil {
				bestCandidate = owners[0]
			}

			for _, owner := range owners {
				if owner.ID != bestCandidate.ID {
					if _, ok := ignoredSynonyms[owner.ID]; !ok {
						ignoredSynonyms[owner.ID] = make(map[string]struct{})
					}
					ignoredSynonyms[owner.ID][synonym] = struct{}{}
				}
			}
		}
	}

	return ignoredSynonyms
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// matchLocalFileLegacy finds the best match for the local file (legacy implementation)
// If the best match is above a certain threshold, set the local file's mediaId to the best match's id
// If the best match is below a certain threshold, leave the local file's mediaId to 0
func (m *Matcher) matchLocalFileLegacy(lf *anime.LocalFile) {
	defer util.HandlePanicInModuleThenS("scanner/matcher/matchLocalFileLegacy", func(stackTrace string) {
		lf.MediaId = 0
		/*Log*/
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("Panic occurred, file un-matched")
		}
		m.ScanSummaryLogger.LogPanic(lf, stackTrace)
	})

	// Check if the local file has already been matched
	if lf.MediaId != 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Msg("File already matched")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "Already matched")
		return
	}
	// Check if the local file has a title
	if lf.GetParsedTitle() == "" {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).
				Str("filename", lf.Name).
				Msg("File has no parsed title")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No parsed title found")
		return
	}

	// Create title variations
	// Check cache for title variation

	titleVariations := lf.GetTitleVariations()

	if len(titleVariations) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).
				Str("filename", lf.Name).
				Msg("No titles found")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No title variations found")
		return
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Interface("titleVariations", titleVariations).
			Msg("Matching local file")
	}
	m.ScanSummaryLogger.LogDebug(lf, util.InlineSpewT(titleVariations))

	//------------------

	var levMatch *comparison.LevenshteinResult
	var sdMatch *comparison.SorensenDiceResult
	var jaccardMatch *comparison.JaccardResult

	if m.Algorithm == "jaccard" {
		// Using Jaccard
		// Get the matchs for each title variation
		compResults := lop.Map(titleVariations, func(title *string, _ int) *comparison.JaccardResult {
			comps := make([]*comparison.JaccardResult, 0)
			if len(m.MediaContainer.engTitles) > 0 {
				if eng, found := comparison.FindBestMatchWithJaccard(title, m.MediaContainer.engTitles); found {
					comps = append(comps, eng)
				}
			}
			if len(m.MediaContainer.romTitles) > 0 {
				if rom, found := comparison.FindBestMatchWithJaccard(title, m.MediaContainer.romTitles); found {
					comps = append(comps, rom)
				}
			}
			if len(m.MediaContainer.synonyms) > 0 {
				if syn, found := comparison.FindBestMatchWithJaccard(title, m.MediaContainer.synonyms); found {
					comps = append(comps, syn)
				}
			}
			var res *comparison.JaccardResult
			if len(comps) > 1 {
				res = lo.Reduce(comps, func(prev *comparison.JaccardResult, curr *comparison.JaccardResult, _ int) *comparison.JaccardResult {
					if prev.Rating > curr.Rating {
						return prev
					} else {
						return curr
					}
				}, comps[0])
			} else if len(comps) == 1 {
				return comps[0]
			}
			return res
		})

		// Retrieve the match from all the title variations results
		jaccardMatch = lo.Reduce(compResults, func(prev *comparison.JaccardResult, curr *comparison.JaccardResult, _ int) *comparison.JaccardResult {
			if prev.Rating > curr.Rating {
				return prev
			} else {
				return curr
			}
		}, compResults[0])

		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Interface("match", jaccardMatch).
				Interface("results", compResults).
				Msg("Jaccard match")
		}
		m.ScanSummaryLogger.LogComparison(lf, "Jaccard", *jaccardMatch.Value, "Rating", util.InlineSpewT(jaccardMatch.Rating))

	} else if m.Algorithm == "sorensen-dice" {
		// Using Sorensen-Dice
		// Get the matchs for each title variation
		compResults := lop.Map(titleVariations, func(title *string, _ int) *comparison.SorensenDiceResult {
			comps := make([]*comparison.SorensenDiceResult, 0)
			if len(m.MediaContainer.engTitles) > 0 {
				if eng, found := comparison.FindBestMatchWithSorensenDice(title, m.MediaContainer.engTitles); found {
					comps = append(comps, eng)
				}
			}
			if len(m.MediaContainer.romTitles) > 0 {
				if rom, found := comparison.FindBestMatchWithSorensenDice(title, m.MediaContainer.romTitles); found {
					comps = append(comps, rom)
				}
			}
			if len(m.MediaContainer.synonyms) > 0 {
				if syn, found := comparison.FindBestMatchWithSorensenDice(title, m.MediaContainer.synonyms); found {
					comps = append(comps, syn)
				}
			}
			var res *comparison.SorensenDiceResult
			if len(comps) > 1 {
				res = lo.Reduce(comps, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
					if prev.Rating > curr.Rating {
						return prev
					} else {
						return curr
					}
				}, comps[0])
			} else if len(comps) == 1 {
				return comps[0]
			}
			return res
		})

		// Retrieve the match from all the title variations results
		sdMatch = lo.Reduce(compResults, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
			if prev.Rating > curr.Rating {
				return prev
			} else {
				return curr
			}
		}, compResults[0])

		//util.Spew(compResults)

		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Interface("match", sdMatch).
				Interface("results", compResults).
				Msg("Sorensen-Dice match")
		}
		m.ScanSummaryLogger.LogComparison(lf, "Sorensen-Dice", *sdMatch.Value, "Rating", util.InlineSpewT(sdMatch.Rating))

	} else {
		// Using Levenshtein
		// Get the matches for each title variation
		levCompResults := lop.Map(titleVariations, func(title *string, _ int) *comparison.LevenshteinResult {
			comps := make([]*comparison.LevenshteinResult, 0)
			if len(m.MediaContainer.engTitles) > 0 {
				if eng, found := comparison.FindBestMatchWithLevenshtein(title, m.MediaContainer.engTitles); found {
					comps = append(comps, eng)
				}
			}
			if len(m.MediaContainer.romTitles) > 0 {
				if rom, found := comparison.FindBestMatchWithLevenshtein(title, m.MediaContainer.romTitles); found {
					comps = append(comps, rom)
				}
			}
			if len(m.MediaContainer.synonyms) > 0 {
				if syn, found := comparison.FindBestMatchWithLevenshtein(title, m.MediaContainer.synonyms); found {
					comps = append(comps, syn)
				}
			}
			var res *comparison.LevenshteinResult
			if len(comps) > 1 {
				res = lo.Reduce(comps, func(prev *comparison.LevenshteinResult, curr *comparison.LevenshteinResult, _ int) *comparison.LevenshteinResult {
					if prev.Distance < curr.Distance {
						return prev
					} else {
						return curr
					}
				}, comps[0])
			} else if len(comps) == 1 {
				return comps[0]
			}
			return res
		})

		levMatch = lo.Reduce(levCompResults, func(prev *comparison.LevenshteinResult, curr *comparison.LevenshteinResult, _ int) *comparison.LevenshteinResult {
			if prev.Distance < curr.Distance {
				return prev
			} else {
				return curr
			}
		}, levCompResults[0])

		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Interface("match", levMatch).
				Interface("results", levCompResults).
				Int("distance", levMatch.Distance).
				Msg("Levenshtein match")
		}
		m.ScanSummaryLogger.LogComparison(lf, "Levenshtein", *levMatch.Value, "Distance", util.InlineSpewT(levMatch.Distance))
	}

	//------------------

	var mediaMatch *anime.NormalizedMedia
	var found bool
	finalRating := 0.0

	if sdMatch != nil {
		finalRating = sdMatch.Rating
		mediaMatch, found = m.MediaContainer.GetMediaFromTitleOrSynonym(sdMatch.Value)

	} else if jaccardMatch != nil {
		finalRating = jaccardMatch.Rating
		mediaMatch, found = m.MediaContainer.GetMediaFromTitleOrSynonym(jaccardMatch.Value)

	} else {
		dice := GetEfficientDice()
		finalRating = dice.Compare(*levMatch.OriginalValue, *levMatch.Value)
		PutEfficientDice(dice)
		m.ScanSummaryLogger.LogComparison(lf, "Sorensen-Dice", *levMatch.Value, "Final rating", util.InlineSpewT(finalRating))
		mediaMatch, found = m.MediaContainer.GetMediaFromTitleOrSynonym(levMatch.Value)
	}

	// After setting the mediaId, add the hook invocation
	// Invoke ScanLocalFileMatched hook
	event := &ScanLocalFileMatchedEvent{
		LocalFile: lf,
		Score:     finalRating,
		Match:     mediaMatch,
		Found:     found,
	}
	hook.GlobalHookManager.OnScanLocalFileMatched().Trigger(event)
	lf = event.LocalFile
	mediaMatch = event.Match
	found = event.Found
	finalRating = event.Score

	// Check if the hook overrode the match
	if event.DefaultPrevented {
		if m.ScanLogger != nil {
			if mediaMatch != nil {
				m.ScanLogger.LogMatcher(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Int("id", mediaMatch.ID).
					Msg("Hook overrode match")
			} else {
				m.ScanLogger.LogMatcher(zerolog.DebugLevel).
					Str("filename", lf.Name).
					Msg("Hook overrode match, no match found")
			}
		}
		if mediaMatch != nil {
			lf.MediaId = mediaMatch.ID
		} else {
			lf.MediaId = 0
		}
		return
	}

	if !found {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No media found from comparison result")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No media found from comparison result")
		return
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Str("title", mediaMatch.GetTitleSafe()).
			Int("id", mediaMatch.ID).
			Msg("Best match found")
	}

	if finalRating < m.Threshold {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Float64("rating", finalRating).
				Float64("threshold", m.Threshold).
				Msg("Best match Sorensen-Dice rating too low, un-matching file")
		}
		m.ScanSummaryLogger.LogFailedMatch(lf, "Rating too low, threshold is "+fmt.Sprintf("%f", m.Threshold))
		return
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Float64("rating", finalRating).
			Float64("threshold", m.Threshold).
			Msg("Best match rating high enough, matching file")
	}
	m.ScanSummaryLogger.LogSuccessfullyMatched(lf, mediaMatch.ID)

	lf.MediaId = mediaMatch.ID
}
