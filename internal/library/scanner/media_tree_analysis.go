package scanner

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/util/limiter"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
)

type (
	MediaTreeAnalysisOptions struct {
		tree             *anilist.CompleteAnimeRelationTree
		metadataProvider metadata.Provider
		rateLimiter      *limiter.Limiter
	}

	MediaTreeAnalysis struct {
		branches []*MediaTreeAnalysisBranch
	}

	MediaTreeAnalysisBranch struct {
		media         *anilist.CompleteAnime
		animeMetadata *metadata.AnimeMetadata
		// The second absolute episode number of the first episode
		// Sometimes, the metadata provider may have a 'true' absolute episode number and a 'part' absolute episode number
		// 'part' absolute episode numbers might be used for "Part 2s" of a season
		minPartAbsoluteEpisodeNumber int
		maxPartAbsoluteEpisodeNumber int
		minAbsoluteEpisode           int
		maxAbsoluteEpisode           int
		totalEpisodeCount            int
		noAbsoluteEpisodesFound      bool
	}
)

// NewMediaTreeAnalysis will analyze the media tree and create and store a MediaTreeAnalysisBranch for each media in the tree.
// Each MediaTreeAnalysisBranch will contain the min and max absolute episode number for the media.
// The min and max absolute episode numbers are used to get the relative episode number from an absolute episode number.
func NewMediaTreeAnalysis(opts *MediaTreeAnalysisOptions) (*MediaTreeAnalysis, error) {

	relations := make([]*anilist.CompleteAnime, 0)
	opts.tree.Range(func(key int, value *anilist.CompleteAnime) bool {
		relations = append(relations, value)
		return true
	})

	// Get Animap data for all related media in the tree
	// With each Animap media, get the min and max absolute episode number
	// Create new MediaTreeAnalysisBranch for each Animap media
	p := pool.NewWithResults[*MediaTreeAnalysisBranch]().WithErrors()
	for _, rel := range relations {
		p.Go(func() (*MediaTreeAnalysisBranch, error) {
			opts.rateLimiter.Wait()

			animeMetadata, err := opts.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, rel.ID)
			if err != nil {
				return nil, err
			}

			// Get the first episode
			firstEp, ok := animeMetadata.Episodes["1"]
			if !ok {
				return nil, errors.New("no first episode")
			}

			// discrepancy: "seasonNumber":1,"episodeNumber":12,"absoluteEpisodeNumber":13,
			// this happens when the media has a separate entry but is technically the same season
			// when we detect this, we should use the "episodeNumber" as the absoluteEpisodeNumber
			// this is a hacky fix, but it works for the cases I've seen so far
			usePartEpisodeNumber := firstEp.EpisodeNumber > 1 && firstEp.AbsoluteEpisodeNumber-firstEp.EpisodeNumber > 1
			partAbsoluteEpisodeNumber := 0
			maxPartAbsoluteEpisodeNumber := 0
			if usePartEpisodeNumber {
				partAbsoluteEpisodeNumber = firstEp.EpisodeNumber
				maxPartAbsoluteEpisodeNumber = partAbsoluteEpisodeNumber + animeMetadata.GetMainEpisodeCount() - 1
			}

			// If the first episode exists and has a valid absolute episode number, create a new MediaTreeAnalysisBranch
			if animeMetadata.Episodes != nil {
				return &MediaTreeAnalysisBranch{
					media:                        rel,
					animeMetadata:                animeMetadata,
					minPartAbsoluteEpisodeNumber: partAbsoluteEpisodeNumber,
					maxPartAbsoluteEpisodeNumber: maxPartAbsoluteEpisodeNumber,
					minAbsoluteEpisode:           firstEp.AbsoluteEpisodeNumber,
					// The max absolute episode number is the first episode's absolute episode number plus the total episode count minus 1
					// We subtract 1 because the first episode's absolute episode number is already included in the total episode count
					// e.g, if the first episode's absolute episode number is 13 and the total episode count is 12, the max absolute episode number is 24
					maxAbsoluteEpisode:      firstEp.AbsoluteEpisodeNumber + (animeMetadata.GetMainEpisodeCount() - 1),
					totalEpisodeCount:       animeMetadata.GetMainEpisodeCount(),
					noAbsoluteEpisodesFound: firstEp.AbsoluteEpisodeNumber == 0,
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

	isPartAbsolute := false

	// Find the MediaTreeAnalysisBranch that contains the absolute episode number
	branch, ok := lo.Find(o.branches, func(n *MediaTreeAnalysisBranch) bool {
		// First check if the partAbsoluteEpisodeNumber is set
		if n.minPartAbsoluteEpisodeNumber > 0 && n.maxPartAbsoluteEpisodeNumber > 0 {
			// If it is, check if the absolute episode number given is the same as the partAbsoluteEpisodeNumber
			// If it is, return true
			if n.minPartAbsoluteEpisodeNumber <= abs && n.maxPartAbsoluteEpisodeNumber >= abs {
				isPartAbsolute = true
				return true
			}
		}

		// Else, check if the absolute episode number given is within the min and max absolute episode numbers of the branch
		if n.minAbsoluteEpisode <= abs && n.maxAbsoluteEpisode >= abs {
			return true
		}
		return false
	})
	if !ok {
		// Sort branches manually
		type branchByFirstEpDate struct {
			branch             *MediaTreeAnalysisBranch
			firstEpDate        time.Time
			minAbsoluteEpisode int
			maxAbsoluteEpisode int
		}
		branches := make([]*branchByFirstEpDate, 0)
		for _, b := range o.branches {
			// Get the first episode date
			firstEp, ok := b.animeMetadata.Episodes["1"]
			if !ok {
				continue
			}
			// parse date
			t, err := time.Parse(time.DateOnly, firstEp.AirDate)
			if err != nil {
				continue
			}
			branches = append(branches, &branchByFirstEpDate{
				branch:      b,
				firstEpDate: t,
			})
		}

		// Sort branches by first episode date
		// If the first episode date is not available, the branch will be placed at the end
		sort.Slice(branches, func(i, j int) bool {
			return branches[i].firstEpDate.Before(branches[j].firstEpDate)
		})

		// Hydrate branches with min and max absolute episode numbers
		visited := make(map[int]*branchByFirstEpDate)
		for idx, b := range branches {
			visited[idx] = b
			if v, ok := visited[idx-1]; ok {
				b.minAbsoluteEpisode = v.maxAbsoluteEpisode + 1
				b.maxAbsoluteEpisode = b.minAbsoluteEpisode + b.branch.totalEpisodeCount - 1
				continue
			}
			b.minAbsoluteEpisode = 1
			b.maxAbsoluteEpisode = b.minAbsoluteEpisode + b.branch.totalEpisodeCount - 1
		}

		for _, b := range branches {
			if b.minAbsoluteEpisode <= abs && b.maxAbsoluteEpisode >= abs {
				b.branch.minAbsoluteEpisode = b.minAbsoluteEpisode
				b.branch.maxAbsoluteEpisode = b.maxAbsoluteEpisode
				branch = b.branch
				relativeEp = abs - (branch.minAbsoluteEpisode - 1)
				mediaId = branch.media.ID
				ok = true
				return
			}
		}

		return 0, 0, false
	}

	if isPartAbsolute {
		// Let's say the media has 12 episodes and the file is "episode 13"
		// If the [partAbsoluteEpisodeNumber] is 13, then the [relativeEp] will be 1, we can safely ignore the [absoluteEpisodeNumber]
		// e.g. 13 - (13-1) = 1
		relativeEp = abs - (branch.minPartAbsoluteEpisodeNumber - 1)
	} else {
		// Let's say the media has 12 episodes and the file is "episode 38"
		// The [minAbsoluteEpisode] will be 38 and the [relativeEp] will be 1
		// e.g. 38 - (38-1) = 1
		relativeEp = abs - (branch.minAbsoluteEpisode - 1)
	}

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
