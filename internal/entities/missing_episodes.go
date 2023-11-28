package entities

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"time"
)

type (
	MissingEpisodes struct {
		Episodes []*MediaEntryEpisode `json:"episodes"`
	}

	NewMissingEpisodesOptions struct {
		AnilistCollection *anilist.AnimeCollection
		LocalFiles        []*LocalFile
		AnizipCache       *anizip.Cache
	}
)

func NewMissingEpisodes(opts *NewMissingEpisodesOptions) *MissingEpisodes {

	missing := new(MissingEpisodes)
	rateLimiter := limiter.NewLimiter(time.Second, 20)

	groupedLfs := GroupLocalFilesByMediaID(opts.LocalFiles)

	p := pool.NewWithResults[[]*MediaEntryDownloadEpisode]()
	for mId, lfs := range groupedLfs {
		mId := mId
		lfs := lfs
		p.Go(func() []*MediaEntryDownloadEpisode {
			entry, found := opts.AnilistCollection.GetListEntryFromMediaId(mId)
			if !found {
				return nil
			}

			latestLf, found := FindLatestLocalFileFromGroup(lfs)
			if !found {
				return nil
			}
			//If the latest local file is the same or higher than the current episode count, skip
			if entry.Media.GetCurrentEpisodeCount() <= latestLf.GetEpisodeNumber() {
				return nil
			}
			rateLimiter.Wait()
			// Fetch anizip media
			anizipMedia, err := anizip.FetchAniZipMediaC("anilist", entry.Media.ID, opts.AnizipCache)
			if err != nil {
				return nil
			}

			// Get download info
			downloadInfo, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
				localFiles:  lfs,
				anizipMedia: anizipMedia,
				progress:    entry.Progress,
				status:      entry.Status,
				media:       entry.Media,
			})
			if err != nil {
				return nil
			}

			episodes := downloadInfo.EpisodesToDownload
			// Truncate to 5 max
			if len(episodes) > 5 {
				episodes = episodes[:5]
			}
			return episodes
		})
	}
	epsToDownload := p.Wait()
	epsToDownload = lo.Filter(epsToDownload, func(item []*MediaEntryDownloadEpisode, _ int) bool {
		return item != nil
	})

	// Flatten
	flattenedEpsToDownload := lo.Flatten(epsToDownload)
	eps := lop.Map(flattenedEpsToDownload, func(item *MediaEntryDownloadEpisode, _ int) *MediaEntryEpisode {
		return item.Episode
	})
	// Sort
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].GetEpisodeNumber() < eps[j].GetEpisodeNumber()
	})
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].BasicMedia.ID < eps[j].BasicMedia.ID
	})

	missing.Episodes = eps

	return missing

}
