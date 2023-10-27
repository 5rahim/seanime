package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/sourcegraph/conc/pool"
)

type MediaTreeAnalysisOptions struct {
	tree        *anilist.BaseMediaRelationTree
	anizipCache *anizip.Cache
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

func NewMediaTreeAnalysis(opts *MediaTreeAnalysisOptions) *MediaTreeAnalysis {

	relations := make([]*anilist.BaseMedia, 0)
	opts.tree.Range(func(key int, value *anilist.BaseMedia) bool {
		relations = append(relations, value)
		return true
	})

	// Get Anizip data for all related media in the tree
	p := pool.NewWithResults[*MediaTreeAnalysisBranch]()
	for _, rel := range relations {
		rel := rel
		p.Go(func() *MediaTreeAnalysisBranch {

			if azm, err := anizip.FetchAniZipMedia("anilist", rel.ID); err == nil {

				firstEp, ok := azm.Episodes["1"]

				if azm.Episodes != nil && ok && firstEp.AbsoluteEpisodeNumber > 0 {
					return &MediaTreeAnalysisBranch{
						media:              rel,
						anizipMedia:        azm,
						minAbsoluteEpisode: firstEp.AbsoluteEpisodeNumber,
						maxAbsoluteEpisode: (firstEp.AbsoluteEpisodeNumber - 1) + rel.GetTotalEpisodeCount(),
					}
				}

			}

			return &MediaTreeAnalysisBranch{}

		})
	}
	branches := p.Wait()

	return &MediaTreeAnalysis{branches: branches}

}

func (o *MediaTreeAnalysis) getRelativeEpisodeNumber(abs int) (int, bool) {
	branch, ok := lo.Find(o.branches, func(n *MediaTreeAnalysisBranch) bool {
		if n.minAbsoluteEpisode <= abs && n.maxAbsoluteEpisode >= abs {
			return true
		}
		return false
	})
	if !ok {
		return 0, false
	}

	relativeEp := abs - (branch.minAbsoluteEpisode - 1)

	return relativeEp, true
}
