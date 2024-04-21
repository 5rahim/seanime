package anime

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/sourcegraph/conc/pool"
	"slices"
	"strconv"
)

type (
	// MediaEntryDownloadInfo is instantiated by the MediaEntry
	MediaEntryDownloadInfo struct {
		EpisodesToDownload    []*MediaEntryDownloadEpisode `json:"episodesToDownload"`
		CanBatch              bool                         `json:"canBatch"`
		BatchAll              bool                         `json:"batchAll"`
		HasInaccurateSchedule bool                         `json:"hasInaccurateSchedule"`
		Rewatch               bool                         `json:"rewatch"`
		AbsoluteOffset        int                          `json:"absoluteOffset"`
	}

	MediaEntryDownloadEpisode struct {
		EpisodeNumber int                `json:"episodeNumber"`
		AniDBEpisode  string             `json:"aniDBEpisode"`
		Episode       *MediaEntryEpisode `json:"episode"`
	}
)

type (
	NewMediaEntryDownloadInfoOptions struct {
		// Media's local files
		LocalFiles       []*LocalFile
		AnizipMedia      *anizip.Media
		Media            *anilist.BaseMedia
		Progress         *int
		Status           *anilist.MediaListStatus
		MetadataProvider *metadata.Provider
	}
)

// NewMediaEntryDownloadInfo creates a new MediaEntryDownloadInfo
func NewMediaEntryDownloadInfo(opts *NewMediaEntryDownloadInfoOptions) (*MediaEntryDownloadInfo, error) {

	if *opts.Media.Status == anilist.MediaStatusNotYetReleased {
		return &MediaEntryDownloadInfo{}, nil
	}
	if opts.AnizipMedia == nil {
		return nil, errors.New("could not get anizip media")
	}
	if opts.Media.GetCurrentEpisodeCount() == -1 {
		return nil, errors.New("could not get current media episode count")
	}

	// +---------------------+
	// |     Discrepancy     |
	// +---------------------+

	// Whether AniList includes episode 0 as part of main episodes, but Anizip does not, however Anizip has "S1"
	possibleSpecialInclusion, hasDiscrepancy := detectDiscrepancy(opts.LocalFiles, opts.Media, opts.AnizipMedia)

	// I - Progress
	// Get progress, if the media isn't in the user's list, progress is 0
	// If the media is completed, set progress is 0
	progress := 0
	if opts.Progress != nil {
		progress = *opts.Progress
	}
	if opts.Status != nil {
		if *opts.Status == anilist.MediaListStatusCompleted {
			progress = 0
		}
	}

	// II - Create episode number slices for Anilist and Anizip
	// We assume that Episode 0 is 1 if it is included by AniList
	mediaEpSlice := generateEpSlice(opts.Media.GetCurrentEpisodeCount())                         // e.g, [1,2,3,4]
	unwatchedEpSlice := lo.Filter(mediaEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, progress = 1: [1,2,3,4] -> [2,3,4]

	anizipEpSlice := generateEpSlice(opts.AnizipMedia.GetMainEpisodeCount())                            // e.g, [1,2,3,4]
	unwatchedAnizipEpSlice := lo.Filter(anizipEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, progress = 1: [1,2,3,4] -> [2,3,4]

	// +---------------------+
	// |   Anizip has more   |
	// +---------------------+

	// If Anizip has more episodes
	// e.g, Anizip: 2, Anilist: 1
	if opts.AnizipMedia.GetMainEpisodeCount() > opts.Media.GetCurrentEpisodeCount() {
		diff := opts.AnizipMedia.GetMainEpisodeCount() - opts.Media.GetCurrentEpisodeCount()
		// Remove the extra episode number from the Anizip slice
		anizipEpSlice = anizipEpSlice[:len(anizipEpSlice)-diff]                                            // e.g, [1,2] -> [1]
		unwatchedAnizipEpSlice = lo.Filter(anizipEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, [1,2] -> [1]
	}

	// +---------------------+
	// |  Anizip has fewer   |
	// +---------------------+

	// III - Handle discrepancy (inclusion of episode 0 by AniList)
	// If there Anilist has more episodes than Anizip
	// e.g, Anilist: 13, Anizip: 12
	if hasDiscrepancy {
		// Add -1 to Anizip slice, -1 is "S1"
		anizipEpSlice = append([]int{-1}, anizipEpSlice...) // e.g, [-1,1,2,...,12]
		unwatchedAnizipEpSlice = anizipEpSlice              // e.g, [-1,1,2,...,12]
		if progress > 0 {
			unwatchedAnizipEpSlice = lo.Filter(anizipEpSlice, func(i int, _ int) bool { return i > progress-1 })
		}
	}

	// Filter out episodes not aired from the slices
	if opts.Media.NextAiringEpisode != nil {
		unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i < opts.Media.NextAiringEpisode.Episode })
		if hasDiscrepancy {
			unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i < opts.Media.NextAiringEpisode.Episode-1 })
		} else {
			unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i < opts.Media.NextAiringEpisode.Episode })
		}
	}

	// Inaccurate schedule (hacky fix)
	hasInaccurateSchedule := false
	if opts.Media.NextAiringEpisode == nil && *opts.Media.Status == anilist.MediaStatusReleasing {
		if !hasDiscrepancy {
			if progress+1 < opts.AnizipMedia.GetMainEpisodeCount() {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
			} else {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
			}
		} else {
			if progress+1 < opts.AnizipMedia.GetMainEpisodeCount() {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
			} else {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress-1 })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress-1 })
			}
		}
		hasInaccurateSchedule = true
	}

	// This slice contains episode numbers that are not downloaded
	// The source of truth is AniZip, but we will handle discrepancies
	toDownloadSlice := make([]int, 0)
	lfsEpSlice := make([]int, 0)
	if opts.LocalFiles != nil {

		// Get all episode numbers of main local files
		for _, lf := range opts.LocalFiles {
			if lf.Metadata.Type == LocalFileTypeMain {
				if !slices.Contains(lfsEpSlice, lf.GetEpisodeNumber()) {
					lfsEpSlice = append(lfsEpSlice, lf.GetEpisodeNumber())
				}
			}
		}
		// If there is a discrepancy and local files include episode 0, add -1 ("S1") to slice
		if hasDiscrepancy && possibleSpecialInclusion {
			lfsEpSlice = lo.Filter(lfsEpSlice, func(i int, _ int) bool { return i != 0 })
			lfsEpSlice = append([]int{-1}, lfsEpSlice...) // e.g, [-1,1,2,...,12]
		}
		// Filter out downloaded episodes
		if len(lfsEpSlice) > 0 {
			toDownloadSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool {
				return !lo.Contains(lfsEpSlice, i)
			})
		} else {
			toDownloadSlice = unwatchedAnizipEpSlice
		}
	} else {
		toDownloadSlice = unwatchedAnizipEpSlice
	}

	// +---------------------+
	// |   EntryEpisode      |
	// +---------------------+

	// Generate `episodesToDownload` based on `toDownloadSlice`

	// DEVNOTE: The EntryEpisode generated has inaccurate progress numbers since not local files are passed in

	p := pool.NewWithResults[*MediaEntryDownloadEpisode]()
	for _, ep := range toDownloadSlice {
		p.Go(func() *MediaEntryDownloadEpisode {
			str := new(MediaEntryDownloadEpisode)
			str.EpisodeNumber = ep
			str.AniDBEpisode = strconv.Itoa(ep)
			if ep == -1 {
				str.EpisodeNumber = 0
				str.AniDBEpisode = "S1"
			}
			str.Episode = NewMediaEntryEpisode(&NewMediaEntryEpisodeOptions{
				LocalFile:            nil,
				OptionalAniDBEpisode: str.AniDBEpisode,
				AnizipMedia:          opts.AnizipMedia,
				Media:                opts.Media,
				ProgressOffset:       0,
				IsDownloaded:         false,
				MetadataProvider:     opts.MetadataProvider,
			})
			return str
		})
	}
	episodesToDownload := p.Wait()

	//--------------

	canBatch := false
	if *opts.Media.GetStatus() == anilist.MediaStatusFinished && opts.Media.GetTotalEpisodeCount() > 0 {
		canBatch = true
	}
	batchAll := false
	if canBatch && len(lfsEpSlice) == 0 && progress == 0 {
		batchAll = true
	}
	rewatch := false
	if opts.Status != nil && *opts.Status == anilist.MediaListStatusCompleted {
		rewatch = true
	}

	return &MediaEntryDownloadInfo{
		EpisodesToDownload:    episodesToDownload,
		CanBatch:              canBatch,
		BatchAll:              batchAll,
		Rewatch:               rewatch,
		HasInaccurateSchedule: hasInaccurateSchedule,
		AbsoluteOffset:        opts.AnizipMedia.GetOffset(),
	}, nil
}

// generateEpSlice
// e.g, 4 -> [1,2,3,4], 3 -> [1,2,3]
func generateEpSlice(n int) []int {
	if n < 1 {
		return nil
	}
	result := make([]int, n)
	for i := 1; i <= n; i++ {
		result[i-1] = i
	}
	return result
}
