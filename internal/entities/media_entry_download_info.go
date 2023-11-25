package entities

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
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

	NewMediaEntryDownloadInfoOptions struct {
		// Media's local files
		localFiles  []*LocalFile
		anizipMedia *anizip.Media
		media       *anilist.BaseMedia
		progress    *int
		status      *anilist.MediaListStatus
	}
)

// NewMediaEntryDownloadInfo creates a new MediaEntryDownloadInfo
func NewMediaEntryDownloadInfo(opts *NewMediaEntryDownloadInfoOptions) (*MediaEntryDownloadInfo, error) {

	if *opts.media.Status == anilist.MediaStatusNotYetReleased {
		return &MediaEntryDownloadInfo{}, nil
	}
	if opts.anizipMedia == nil {
		return nil, errors.New("could not get anizip media")
	}
	if opts.media.GetCurrentEpisodeCount() == -1 {
		return nil, errors.New("could not get current media episode count")
	}
	possibleSpecialInclusion, hasDiscrepancy := detectDiscrepancy(opts.localFiles, opts.media, opts.anizipMedia)

	// I - Progress
	// Get progress, if the media isn't in the user's list, progress is 0
	// If the media is completed, set progress is 0
	progress := 0
	if opts.progress != nil {
		progress = *opts.progress
	}
	if opts.status != nil {
		if *opts.status == anilist.MediaListStatusCompleted {
			progress = 0
		}
	}

	// II - Create episode number slices for Anilist and Anizip
	// We assume that Episode 0 is 1 if it is included by AniList
	mediaEpSlice := generateEpSlice(opts.media.GetCurrentEpisodeCount())                         // e.g, [1,2,3,4]
	unwatchedEpSlice := lo.Filter(mediaEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, progress = 1: [1,2,3,4] -> [2,3,4]

	anizipEpSlice := generateEpSlice(opts.anizipMedia.GetMainEpisodeCount())                            // e.g, [1,2,3,4]
	unwatchedAnizipEpSlice := lo.Filter(anizipEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, progress = 1: [1,2,3,4] -> [2,3,4]

	// If Anizip has more episodes
	// e.g, Anizip: 2, Anilist: 1
	if opts.anizipMedia.GetMainEpisodeCount() > opts.media.GetCurrentEpisodeCount() {
		diff := opts.anizipMedia.GetMainEpisodeCount() - opts.media.GetCurrentEpisodeCount()
		// Remove the difference from the Anizip slice
		anizipEpSlice = anizipEpSlice[:len(anizipEpSlice)-diff]                                            // e.g, [1,2] -> [1]
		unwatchedAnizipEpSlice = lo.Filter(anizipEpSlice, func(i int, _ int) bool { return i > progress }) // e.g, [1,2] -> [1]
	}

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

	// Filter out unavailable episodes for the slices
	if opts.media.NextAiringEpisode != nil {
		unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i < opts.media.NextAiringEpisode.Episode })
		if hasDiscrepancy {
			unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i < opts.media.NextAiringEpisode.Episode-1 })
		} else {
			unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i < opts.media.NextAiringEpisode.Episode })
		}
	}

	// Inaccurate schedule
	hasInaccurateSchedule := false
	if opts.media.NextAiringEpisode == nil && *opts.media.Status == anilist.MediaStatusReleasing {
		if !hasDiscrepancy {
			if progress+1 < opts.anizipMedia.GetMainEpisodeCount() {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress+1 })
			} else {
				unwatchedEpSlice = lo.Filter(unwatchedEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
				unwatchedAnizipEpSlice = lo.Filter(unwatchedAnizipEpSlice, func(i int, _ int) bool { return i > progress && i <= progress })
			}
		} else {
			if progress+1 < opts.anizipMedia.GetMainEpisodeCount() {
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
	if opts.localFiles != nil {

		// Get all episode numbers of main local files
		for _, lf := range opts.localFiles {
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
		// Filter out downloaed episodes
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

	//---------------------------------

	// Generate `episodesToDownload` based on `toDownloadSlice`
	//episodesToDownload := make([]*MediaEntryDownloadEpisode, 0)
	p := pool.NewWithResults[*MediaEntryDownloadEpisode]()
	for _, ep := range toDownloadSlice {
		ep := ep
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
				AnizipMedia:          opts.anizipMedia,
				Media:                opts.media,
				ProgressOffset:       0,
				IsDownloaded:         false,
			})
			return str
		})
	}
	episodesToDownload := p.Wait()

	//--------------

	canBatch := false
	if *opts.media.GetStatus() == anilist.MediaStatusFinished && opts.media.GetTotalEpisodeCount() > 0 {
		canBatch = true
	}
	batchAll := false
	if canBatch && len(lfsEpSlice) == 0 && progress == 0 {
		batchAll = true
	}
	rewatch := false
	if opts.status != nil && *opts.status == anilist.MediaListStatusCompleted {
		rewatch = true
	}

	return &MediaEntryDownloadInfo{
		EpisodesToDownload:    episodesToDownload,
		CanBatch:              canBatch,
		BatchAll:              batchAll,
		Rewatch:               rewatch,
		HasInaccurateSchedule: hasInaccurateSchedule,
		AbsoluteOffset:        opts.anizipMedia.GetOffset(),
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
