package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/sourcegraph/conc/pool"
)

type MediaTreeAnalysisOptions struct {
	tree        *anilist.BaseMediaRelationTree
	anizipCache *anizip.Cache
	rateLimiter *limiter.Limiter
}

type MediaTreeAnalysis struct {
	branches []*MediaTreeAnalysisBranch
}

type MediaTreeAnalysisBranch struct {
	media              *anilist.BaseMedia
	anizipMedia        *anizip.Media
	minAbsoluteEpisode int
	maxAbsoluteEpisode int
}

// NewMediaTreeAnalysis will analyze the media tree and create and store a MediaTreeAnalysisBranch for each media in the tree.
// Each MediaTreeAnalysisBranch will contain the min and max absolute episode number for the media.
// The min and max absolute episode numbers are used to get the relative episode number from an absolute episode number.
func NewMediaTreeAnalysis(opts *MediaTreeAnalysisOptions) *MediaTreeAnalysis {

	relations := make([]*anilist.BaseMedia, 0)
	opts.tree.Range(func(key int, value *anilist.BaseMedia) bool {
		relations = append(relations, value)
		return true
	})

	// Get Anizip data for all related media in the tree
	// With each Anizip media, get the min and max absolute episode number
	// Create new MediaTreeAnalysisBranch for each Anizip media
	p := pool.NewWithResults[*MediaTreeAnalysisBranch]()
	for _, rel := range relations {
		rel := rel
		p.Go(func() *MediaTreeAnalysisBranch {
			opts.rateLimiter.Wait()
			if azm, err := anizip.FetchAniZipMedia("anilist", rel.ID); err == nil {

				// Get the first episode
				firstEp, ok := azm.Episodes["1"]
				// If the first episode exists and has a valid absolute episode number, create a new MediaTreeAnalysisBranch
				if azm.Episodes != nil && ok && firstEp.AbsoluteEpisodeNumber > 0 {
					return &MediaTreeAnalysisBranch{
						media:              rel,
						anizipMedia:        azm,
						minAbsoluteEpisode: firstEp.AbsoluteEpisodeNumber,
						// The max absolute episode number is the first episode's absolute episode number plus the total episode count minus 1
						// We subtract 1 because the first episode's absolute episode number is already included in the total episode count
						// e.g, if the first episode's absolute episode number is 13 and the total episode count is 12, the max absolute episode number is 24
						maxAbsoluteEpisode: (firstEp.AbsoluteEpisodeNumber - 1) + rel.GetTotalEpisodeCount(),
					}
				}

			} else if err != nil {
				println("anizip.FetchAniZipMedia error:", err.Error())
			}

			return &MediaTreeAnalysisBranch{}

		})
	}
	branches := p.Wait()

	return &MediaTreeAnalysis{branches: branches}

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
