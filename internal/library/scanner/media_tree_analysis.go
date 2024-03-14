package scanner

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/sourcegraph/conc/pool"
)

type (
	MediaTreeAnalysisOptions struct {
		tree        *anilist.BaseMediaRelationTree
		anizipCache *anizip.Cache
		rateLimiter *limiter.Limiter
	}

	MediaTreeAnalysis struct {
		branches []*MediaTreeAnalysisBranch
	}

	MediaTreeAnalysisBranch struct {
		media              *anilist.BaseMedia
		anizipMedia        *anizip.Media
		minAbsoluteEpisode int
		maxAbsoluteEpisode int
		totalEpisodeCount  int
	}
)

// NewMediaTreeAnalysis will analyze the media tree and create and store a MediaTreeAnalysisBranch for each media in the tree.
// Each MediaTreeAnalysisBranch will contain the min and max absolute episode number for the media.
// The min and max absolute episode numbers are used to get the relative episode number from an absolute episode number.
func NewMediaTreeAnalysis(opts *MediaTreeAnalysisOptions) (*MediaTreeAnalysis, error) {

	relations := make([]*anilist.BaseMedia, 0)
	opts.tree.Range(func(key int, value *anilist.BaseMedia) bool {
		relations = append(relations, value)
		return true
	})

	// Get Anizip data for all related media in the tree
	// With each Anizip media, get the min and max absolute episode number
	// Create new MediaTreeAnalysisBranch for each Anizip media
	p := pool.NewWithResults[*MediaTreeAnalysisBranch]().WithErrors()
	for _, rel := range relations {
		p.Go(func() (*MediaTreeAnalysisBranch, error) {
			opts.rateLimiter.Wait()
			azm, err := anizip.FetchAniZipMedia("anilist", rel.ID)
			if err != nil {
				return nil, err
			}
			// Get the first episode
			firstEp, ok := azm.Episodes["1"]
			if !ok {
				return nil, errors.New("no first episode")
			}

			// discrepancy: "seasonNumber":1,"episodeNumber":12,"absoluteEpisodeNumber":13,
			// this happens when the media has a separate entry but is technically the same season
			// when we detect this, we should use the "episodeNumber" as the absoluteEpisodeNumber
			// this is a hacky fix, but it works for the cases I've seen so far
			shouldUseEpisodeNumber := firstEp.EpisodeNumber > 1 && firstEp.AbsoluteEpisodeNumber-firstEp.EpisodeNumber == 1

			absoluteEpisodeNumber := firstEp.AbsoluteEpisodeNumber
			if shouldUseEpisodeNumber {
				absoluteEpisodeNumber = firstEp.AbsoluteEpisodeNumber - 1 // we offset by one
			}

			// If the first episode exists and has a valid absolute episode number, create a new MediaTreeAnalysisBranch
			if azm.Episodes != nil && firstEp.AbsoluteEpisodeNumber > 0 {
				return &MediaTreeAnalysisBranch{
					media:              rel,
					anizipMedia:        azm,
					minAbsoluteEpisode: absoluteEpisodeNumber,
					// The max absolute episode number is the first episode's absolute episode number plus the total episode count minus 1
					// We subtract 1 because the first episode's absolute episode number is already included in the total episode count
					// e.g, if the first episode's absolute episode number is 13 and the total episode count is 12, the max absolute episode number is 24
					maxAbsoluteEpisode: (absoluteEpisodeNumber - 1) + rel.GetTotalEpisodeCount(),
					totalEpisodeCount:  rel.GetTotalEpisodeCount(),
				}, nil
			}

			return nil, errors.New("could not analyze media tree branch")

		})
	}
	branches, _ := p.Wait()

	if branches == nil || len(branches) == 0 {
		return nil, errors.New("no branches found")
	}

	return &MediaTreeAnalysis{branches: branches}, nil

}

// getRelativeEpisodeNumber uses the MediaTreeAnalysis to get the relative episode number for an absolute episode number
func (o *MediaTreeAnalysis) getRelativeEpisodeNumber(abs int) (relativeEp int, mediaId int, ok bool) {
	// Find the MediaTreeAnalysisBranch that contains the absolute episode number
	branch, ok := lo.Find(o.branches, func(n *MediaTreeAnalysisBranch) bool {
		if n.minAbsoluteEpisode <= abs && n.maxAbsoluteEpisode >= abs {
			return true
		}
		return false
	})
	if !ok {
		return 0, 0, false
	}

	relativeEp = abs - (branch.minAbsoluteEpisode - 1)
	mediaId = branch.media.ID

	return
}

func (o *MediaTreeAnalysis) printBranches() (str string) {
	str = "["
	for _, branch := range o.branches {
		str += fmt.Sprintf("media: '%s', minAbsoluteEpisode: %d, maxAbsoluteEpisode: %d, totalEpisodeCount: %d; ", branch.media.GetTitleSafe(), branch.minAbsoluteEpisode, branch.maxAbsoluteEpisode, branch.totalEpisodeCount)
	}
	if len(o.branches) > 0 {
		str = str[:len(str)-2]
	}
	str += "]"
	return str

}
