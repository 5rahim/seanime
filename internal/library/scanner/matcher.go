package scanner

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/sourcegraph/conc/pool"
	"math"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/library/summary"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"time"
)

type Matcher struct {
	LocalFiles         []*anime.LocalFile
	MediaContainer     *MediaContainer
	CompleteAnimeCache *anilist.CompleteAnimeCache
	Logger             *zerolog.Logger
	ScanLogger         *ScanLogger
	ScanSummaryLogger  *summary.ScanSummaryLogger // optional
}

var (
	ErrNoLocalFiles = errors.New("[matcher] no local files")
)

// MatchLocalFilesWithMedia will match each anime.LocalFile with a specific anilist.BaseAnime and modify the LocalFile's `mediaId`
func (m *Matcher) MatchLocalFilesWithMedia() error {

	start := time.Now()

	if len(m.LocalFiles) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).Msg("No local files")
		}
		return ErrNoLocalFiles
	}
	if len(m.MediaContainer.allMedia) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.WarnLevel).Msg("No media fed into the matcher")
		}
		return errors.New("[matcher] no media fed into the matcher")
	}

	m.Logger.Debug().Msg("matcher: Starting matching process")

	// Parallelize the matching process
	lop.ForEach(m.LocalFiles, func(localFile *anime.LocalFile, _ int) {
		m.matchLocalFileWithMedia(localFile)
	})

	m.validateMatches()

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.InfoLevel).
			Any("ms", time.Since(start).Milliseconds()).
			Any("files", len(m.LocalFiles)).
			Any("unmatched", lo.CountBy(m.LocalFiles, func(localFile *anime.LocalFile) bool {
				return localFile.MediaId == 0
			})).
			Msg("matcher: Finished matching process")
	}

	return nil
}

// matchLocalFileWithMedia finds the best match for the local file
// If the best match is above a certain threshold, set the local file's mediaId to the best match's id
// If the best match is below a certain threshold, leave the local file's mediaId to 0
func (m *Matcher) matchLocalFileWithMedia(lf *anime.LocalFile) {

	defer util.HandlePanicInModuleThenS("scanner/matcher/matchLocalFileWithMedia", func(stackTrace string) {
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

	//m.ScanLogger.LogMatcher(zerolog.DebugLevel).
	//	Str("filename", lf.Name).
	//	Any("titleVariations", len(titleVariations)).
	//	Msg("Matching local file")

	// Using Sorensen-Dice
	// Get the best results for each title variation
	sdVariationRes := lop.Map(titleVariations, func(title *string, _ int) *comparison.SorensenDiceResult {
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

	// Retrieve the best result from all the title variations results
	sdRes := lo.Reduce(sdVariationRes, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
		if prev.Rating > curr.Rating {
			return prev
		} else {
			return curr
		}
	}, sdVariationRes[0])

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Str("sdRes", spew.Sprint(sdRes)).
			Msg("Sorensen-Dice best result")
	}
	m.ScanSummaryLogger.LogComparison(lf, "Sorensen-Dice", *sdRes.Value, "Rating", spew.Sprint(sdRes.Rating))

	//------------------

	// Using Levenshtein
	// Get the best results for each title variation
	levVariationRes := lop.Map(titleVariations, func(title *string, _ int) *comparison.LevenshteinResult {
		comps := make([]*comparison.LevenshteinResult, 0)
		if len(m.MediaContainer.engTitles) > 0 {
			if eng, found := comparison.FindBestMatchWithLevenstein(title, m.MediaContainer.engTitles); found {
				comps = append(comps, eng)
			}
		}
		if len(m.MediaContainer.romTitles) > 0 {
			if rom, found := comparison.FindBestMatchWithLevenstein(title, m.MediaContainer.romTitles); found {
				comps = append(comps, rom)
			}
		}
		if len(m.MediaContainer.synonyms) > 0 {
			if syn, found := comparison.FindBestMatchWithLevenstein(title, m.MediaContainer.synonyms); found {
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

	levRes := lo.Reduce(levVariationRes, func(prev *comparison.LevenshteinResult, curr *comparison.LevenshteinResult, _ int) *comparison.LevenshteinResult {
		if prev.Distance < curr.Distance {
			return prev
		} else {
			return curr
		}
	}, levVariationRes[0])

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Str("levRes", spew.Sprint(levRes)).
			Msg("Levenshtein best result")
	}
	m.ScanSummaryLogger.LogComparison(lf, "Levenshtein", *levRes.Value, "Distance", spew.Sprint(levRes.Distance))

	//------------------

	comparisonResValues := make([]*string, 0)
	if sdRes.Value != nil {
		comparisonResValues = append(comparisonResValues, sdRes.Value)
	}
	if levRes.Value != nil {
		comparisonResValues = append(comparisonResValues, levRes.Value)
	}
	if len(comparisonResValues) == 0 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No comparison results")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No comparison results")
		return
	}

	// Using Sorensen-Dice
	// Compare the title variations with the results from Sorensen-Dice and Levenshtein
	bestTitleVariationRes := lop.Map(titleVariations, func(title *string, _ int) *comparison.SorensenDiceResult {
		if v, found := comparison.FindBestMatchWithSorensenDice(title, comparisonResValues); found {
			return v
		}
		return nil
	})

	//m.ScanLogger.LogMatcher(zerolog.DebugLevel).
	//	Str("filename", lf.Name).
	//	Any("bestTitleVariationRes", spew.Sprint(bestTitleVariationRes)).
	//	Msg("Compared best title variations from Sorensen-Dice and Levenshtein with file title variations")

	// Retrieve the best result from all the title variations results
	bestTitleRes := lo.Reduce(bestTitleVariationRes, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
		if prev.Rating > curr.Rating {
			return prev
		} else {
			return curr
		}
	}, &comparison.SorensenDiceResult{})

	if bestTitleRes == nil {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Str("filename", lf.Name).
				Msg("No best title found")
		}
		m.ScanSummaryLogger.LogFileNotMatched(lf, "No best title found")
		return
	}
	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Any("bestTitleRes", spew.Sprint(bestTitleRes)).
			Msg("Best title found")
	}
	m.ScanSummaryLogger.LogComparison(lf, "Sorensen-Dice (Final)", *bestTitleRes.Value, "Rating", spew.Sprint(bestTitleRes.Rating))

	bestMedia, found := m.MediaContainer.GetMediaFromTitleOrSynonym(bestTitleRes.Value)

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
			Any("title", bestMedia.GetTitleSafe()).
			Any("id", bestMedia.ID).
			Msg("Best media")
	}
	if bestTitleRes.Rating < 0.5 {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.DebugLevel).
				Str("filename", lf.Name).
				Any("rating", bestTitleRes.Rating).
				Msg("Best title rating too low, un-matching file")
		}
		m.ScanSummaryLogger.LogFailedMatch(lf, "Rating too low")
		return
	}

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Str("filename", lf.Name).
			Any("rating", bestTitleRes.Rating).
			Msg("Best title rating high enough, matching file")
	}
	m.ScanSummaryLogger.LogSuccessfullyMatched(lf, bestMedia.ID)

	lf.MediaId = bestMedia.ID
	//println(fmt.Sprintf("Local file title: %s,\nbestMedia: %s,\nrating: %f,\nlfMediaId: %d\n", lf.Name, bestMedia.GetTitleSafe(), bestTitleRes.Rating, lf.MediaId))

}

//----------------------------------------------------------------------------------------------------------------------

// validateMatches compares groups of local files' titles with the media titles and un-matches the local files that have a lower rating than the highest rating.
func (m *Matcher) validateMatches() {

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.InfoLevel).Msg("Validating matches")
	}

	// Group local files by media ID
	groups := lop.GroupBy(m.LocalFiles, func(localFile *anime.LocalFile) int {
		return localFile.MediaId
	})

	// Remove the group with unmatched media
	delete(groups, 0)

	// Un-match files with lower ratings
	p := pool.New()
	for mId, files := range groups {
		p.Go(func() {
			if len(files) > 0 {
				m.validateMatchGroup(mId, files)
			}
		})
	}
	p.Wait()

}

// validateMatchGroup compares the local files' titles under the same media
// with the media titles and un-matches the local files that have a lower rating.
// This is done to try and filter out wrong matches.
func (m *Matcher) validateMatchGroup(mediaId int, lfs []*anime.LocalFile) {

	media, found := m.MediaContainer.GetMediaFromId(mediaId)
	if !found {
		if m.ScanLogger != nil {
			m.ScanLogger.LogMatcher(zerolog.ErrorLevel).
				Int("mediaId", mediaId).
				Msg("Media not found in media container")
		}
		return
	}

	titles := media.GetAllTitles()

	// Compare all files' parsed title with the media title
	// Get the highest rating that will be used to un-match lower rated files
	p := pool.NewWithResults[float64]()
	for _, lf := range lfs {
		p.Go(func() float64 {
			t := lf.GetParsedTitle()
			if comparison.ValueContainsSpecial(lf.Name) || comparison.ValueContainsNC(lf.Name) {
				return 0
			}
			compRes, ok := comparison.FindBestMatchWithSorensenDice(&t, titles)
			if ok {
				return compRes.Rating
			}
			return 0
		})
	}
	fileRatings := p.Wait()

	if m.ScanLogger != nil {
		m.ScanLogger.LogMatcher(zerolog.DebugLevel).
			Int("mediaId", mediaId).
			Any("fileRatings", fileRatings).
			Msg("File ratings")
	}

	highestRating := lo.Reduce(fileRatings, func(prev float64, curr float64, _ int) float64 {
		if prev > curr {
			return prev
		} else {
			return curr
		}
	}, 0.0)

	// Un-match files that have a lower rating than the ceiling
	// UNLESS they are Special or NC
	lop.ForEach(lfs, func(lf *anime.LocalFile, _ int) {
		if !comparison.ValueContainsSpecial(lf.Name) && !comparison.ValueContainsNC(lf.Name) {
			t := lf.GetParsedTitle()
			if compRes, ok := comparison.FindBestMatchWithSorensenDice(&t, titles); ok {
				// If the local file's rating is lower, un-match it
				// Unless the difference is less than 0.7 (very lax since a lot of anime have very long names that can be truncated)
				if compRes.Rating < highestRating && math.Abs(compRes.Rating-highestRating) > 0.7 {
					lf.MediaId = 0

					if m.ScanLogger != nil {
						m.ScanLogger.LogMatcher(zerolog.WarnLevel).
							Int("mediaId", mediaId).
							Str("filename", lf.Name).
							Any("rating", compRes.Rating).
							Any("highestRating", highestRating).
							Msg("Rating does not match parameters, un-matching file")
					}
					m.ScanSummaryLogger.LogUnmatched(lf, spew.Sprintf("Rating does not match parameters. File rating: %f, highest rating: %f", compRes.Rating, highestRating))

				} else {

					if m.ScanLogger != nil {
						m.ScanLogger.LogMatcher(zerolog.DebugLevel).
							Int("mediaId", mediaId).
							Str("filename", lf.Name).
							Any("rating", compRes.Rating).
							Any("highestRating", highestRating).
							Msg("Rating matches parameters, keeping file matched")
					}
					m.ScanSummaryLogger.LogMatchValidated(lf, mediaId)

				}
			}
		}
	})

}
