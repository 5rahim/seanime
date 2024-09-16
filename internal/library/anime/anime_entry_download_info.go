package anime

import (
	"errors"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"slices"
	"strconv"
)

type (
	// AnimeEntryDownloadInfo is instantiated by the AnimeEntry
	AnimeEntryDownloadInfo struct {
		EpisodesToDownload    []*AnimeEntryDownloadEpisode `json:"episodesToDownload"`
		CanBatch              bool                         `json:"canBatch"`
		BatchAll              bool                         `json:"batchAll"`
		HasInaccurateSchedule bool                         `json:"hasInaccurateSchedule"`
		Rewatch               bool                         `json:"rewatch"`
		AbsoluteOffset        int                          `json:"absoluteOffset"`
	}

	AnimeEntryDownloadEpisode struct {
		EpisodeNumber int      `json:"episodeNumber"`
		AniDBEpisode  string   `json:"aniDBEpisode"`
		Episode       *Episode `json:"episode"`
	}
)

type (
	NewAnimeEntryDownloadInfoOptions struct {
		// Media's local files
		LocalFiles       []*LocalFile
		AnimeMetadata    *metadata.AnimeMetadata
		Media            *anilist.BaseAnime
		Progress         *int
		Status           *anilist.MediaListStatus
		MetadataProvider metadata.Provider
	}
)

// NewAnimeEntryDownloadInfo creates a new AnimeEntryDownloadInfo
func NewAnimeEntryDownloadInfo(opts *NewAnimeEntryDownloadInfoOptions) (*AnimeEntryDownloadInfo, error) {

	if *opts.Media.Status == anilist.MediaStatusNotYetReleased {
		return &AnimeEntryDownloadInfo{}, nil
	}
	if opts.AnimeMetadata == nil {
		return nil, errors.New("could not get anime metadata")
	}
	if opts.Media.GetCurrentEpisodeCount() == -1 {
		return nil, errors.New("could not get current media episode count")
	}

	// +---------------------+
	// |     Discrepancy     |
	// +---------------------+

	// Whether AniList includes episode 0 as part of main episodes, but Anizip does not, however Anizip has "S1"
	_, hasDiscrepancy := detectDiscrepancy(opts.LocalFiles, opts.Media, opts.AnimeMetadata)

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

	metadataEpSlice := generateEpSlice(opts.AnimeMetadata.GetMainEpisodeCount())                          // e.g, [1,2,3,4]
	unwatchedAnizipEpSlice := lo.Filter(metadataEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, progress = 1: [1,2,3,4] -> [2,3,4]

	// +---------------------+
	// |   Anizip has more   |
	// +---------------------+

	// If Anizip has more episodes
	// e.g, Anizip: 2, Anilist: 1
	if opts.AnimeMetadata.GetMainEpisodeCount() > opts.Media.GetCurrentEpisodeCount() {
		diff := opts.AnimeMetadata.GetMainEpisodeCount() - opts.Media.GetCurrentEpisodeCount()
		// Remove the extra episode number from the Anizip slice
		metadataEpSlice = metadataEpSlice[:len(metadataEpSlice)-diff]                                        // e.g, [1,2] -> [1]
		unwatchedAnizipEpSlice = lo.Filter(metadataEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, [1,2] -> [1]
	}

	// +---------------------+
	// |  Anizip has fewer   |
	// +---------------------+

	// III - Handle discrepancy (inclusion of episode 0 by AniList)
	// If Anilist has more episodes than Anizip
	// e.g, Anilist: 13, Anizip: 12
	if hasDiscrepancy {
		// Add -1 to Anizip slice, -1 is "S1"
		metadataEpSlice = append([]int{-1}, metadataEpSlice...) // e.g, [-1,1,2,...,12]
		unwatchedAnizipEpSlice = metadataEpSlice                // e.g, [-1,1,2,...,12]
		if progress > 0 {
			unwatchedAnizipEpSlice = lo.Filter(metadataEpSlice, func(i int, _ int) bool { return i > progress-1 })
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
			if progress+1 < opts.AnimeMetadata.GetMainEpisodeCount() {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
			} else {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
			}
		} else {
			if progress+1 < opts.AnimeMetadata.GetMainEpisodeCount() {
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
		// If there is a discrepancy, add -1 ("S1") to slice
		if hasDiscrepancy {
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

	p := pool.NewWithResults[*AnimeEntryDownloadEpisode]()
	for _, ep := range toDownloadSlice {
		p.Go(func() *AnimeEntryDownloadEpisode {
			str := new(AnimeEntryDownloadEpisode)
			str.EpisodeNumber = ep
			str.AniDBEpisode = strconv.Itoa(ep)
			if ep == -1 {
				str.EpisodeNumber = 0
				str.AniDBEpisode = "S1"
			}
			str.Episode = NewEpisode(&NewEpisodeOptions{
				LocalFile:            nil,
				OptionalAniDBEpisode: str.AniDBEpisode,
				AnimeMetadata:        opts.AnimeMetadata,
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

	return &AnimeEntryDownloadInfo{
		EpisodesToDownload:    episodesToDownload,
		CanBatch:              canBatch,
		BatchAll:              batchAll,
		Rewatch:               rewatch,
		HasInaccurateSchedule: hasInaccurateSchedule,
		AbsoluteOffset:        opts.AnimeMetadata.GetOffset(),
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
