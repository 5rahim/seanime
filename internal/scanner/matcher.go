package scanner

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/comparison"
	"github.com/seanime-app/seanime-server/internal/result"
)

type Matcher struct {
	localFiles     []*LocalFile
	mediaContainer *MediaContainer
	baseMediaCache *anilist.BaseMediaCache
	matchingCache  *MatchingCache
}

type MatcherOptions struct {
	localFiles     []*LocalFile
	mediaContainer *MediaContainer
	baseMediaCache *anilist.BaseMediaCache
}

// MatchingCache holds the previous results of the matching process.
// The key is a slice of strings representing the title variations of a local file.
// The value is the media ID of the best match.
type MatchingCache struct {
	*result.Cache[[]string, int]
}

func NewMatcher(opts *MatcherOptions) *Matcher {
	m := new(Matcher)
	m.localFiles = opts.localFiles
	m.mediaContainer = opts.mediaContainer
	m.baseMediaCache = opts.baseMediaCache
	m.matchingCache = &MatchingCache{result.NewCache[[]string, int]()}
	return m
}

// MatchLocalFilesWithMedia will match a LocalFile with a specific anilist.BaseMedia and modify the LocalFile's `mediaId`
func (m *Matcher) MatchLocalFilesWithMedia() error {

	if len(m.localFiles) == 0 {
		return errors.New("[matcher] no local files")
	}
	if len(m.mediaContainer.allMedia) == 0 {
		return errors.New("[matcher] no media fed into the matcher")
	}

	// Parallelize the matching process
	lop.ForEach(m.localFiles, func(localFile *LocalFile, index int) {
		m.MatchLocalFileWithMedia(localFile)
	})

	return nil
}

// MatchLocalFileWithMedia finds the best match for the local file
// If the best match is above a certain threshold, set the local file's mediaId to the best match's id
// If the best match is below a certain threshold, leave the local file's mediaId to 0
func (m *Matcher) MatchLocalFileWithMedia(lf *LocalFile) {
	// Check if the local file has already been matched
	if lf.MediaId != 0 {
		return
	}
	// Check if the local file has a title
	if lf.GetParsedTitle() == "" {
		return
	}

	// Create title variations
	// Check cache for title variation

	titleVariations := lf.GetTitleVariations()

	// Using Sorensen-Dice
	// Get the best results for each title variation
	sdVariationRes := lop.Map(titleVariations, func(title *string, _ int) *comparison.SorensenDiceResult {
		cats := make([]*comparison.SorensenDiceResult, 0)
		if eng, found := comparison.FindBestMatchWithSorensenDice(title, m.mediaContainer.engTitles); found {
			cats = append(cats, eng)
		}
		if rom, found := comparison.FindBestMatchWithSorensenDice(title, m.mediaContainer.romTitles); found {
			cats = append(cats, rom)
		}
		if syn, found := comparison.FindBestMatchWithSorensenDice(title, m.mediaContainer.synonyms); found {
			cats = append(cats, syn)
		}
		var res *comparison.SorensenDiceResult
		if len(cats) > 1 {
			res = lo.Reduce(cats, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
				if prev.Rating > curr.Rating {
					return prev
				} else {
					return curr
				}
			}, cats[0])
		} else if len(cats) == 1 {
			return cats[0]
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

	//------------------

	// Using Sorensen-Dice
	// Get the best results for each title variation
	levVariationRes := lop.Map(titleVariations, func(title *string, _ int) *comparison.LevenshteinResult {
		cats := make([]*comparison.LevenshteinResult, 0)
		if eng, found := comparison.FindBestMatchWithLevenstein(title, m.mediaContainer.engTitles); found {
			cats = append(cats, eng)
		}
		if rom, found := comparison.FindBestMatchWithLevenstein(title, m.mediaContainer.romTitles); found {
			cats = append(cats, rom)
		}
		if syn, found := comparison.FindBestMatchWithLevenstein(title, m.mediaContainer.synonyms); found {
			cats = append(cats, syn)
		}
		var res *comparison.LevenshteinResult
		if len(cats) > 1 {
			res = lo.Reduce(cats, func(prev *comparison.LevenshteinResult, curr *comparison.LevenshteinResult, _ int) *comparison.LevenshteinResult {
				if prev.Distance < curr.Distance {
					return prev
				} else {
					return curr
				}
			}, cats[0])
		} else if len(cats) == 1 {
			return cats[0]
		}
		return res
	})

	// Retrieve the best result from all the title variations results
	levRes := lo.Reduce(levVariationRes, func(prev *comparison.LevenshteinResult, curr *comparison.LevenshteinResult, _ int) *comparison.LevenshteinResult {
		if prev.Value == nil {
			return curr
		}
		if prev.Distance < curr.Distance {
			return prev
		} else {
			return curr
		}
	}, &comparison.LevenshteinResult{})

	//------------------

	comparisonResValues := make([]*string, 0)
	if sdRes.Value != nil {
		comparisonResValues = append(comparisonResValues, sdRes.Value)
	}
	if levRes.Value != nil {
		comparisonResValues = append(comparisonResValues, levRes.Value)
	}
	if len(comparisonResValues) == 0 {
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
	// Retrieve the best result from all the title variations results
	bestTitleRes := lo.Reduce(bestTitleVariationRes, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
		if prev.Rating > curr.Rating {
			return prev
		} else {
			return curr
		}
	}, &comparison.SorensenDiceResult{})

	if bestTitleRes == nil {
		return
	}

	bestMedia, found := m.mediaContainer.GetMediaFromTitleOrSynonym(bestTitleRes.Value)

	if !found {
		return
	}

	if bestTitleRes.Rating < 0.5 {
		return
	}

	println(fmt.Sprintf("Local file title: %s,\nbestMedia: %s,\nrating: %f\n", lf.Name, bestMedia.GetTitleSafe(), bestTitleRes.Rating))

	lf.MediaId = bestMedia.ID

	// Compare the local file's title with all the media titles
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Matcher) ValideMatches() {

}
