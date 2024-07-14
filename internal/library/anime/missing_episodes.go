package anime

import (
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"time"
)

type (
	MissingEpisodes struct {
		Episodes         []*MediaEntryEpisode `json:"episodes"`
		SilencedEpisodes []*MediaEntryEpisode `json:"silencedEpisodes"`
	}

	NewMissingEpisodesOptions struct {
		AnimeCollection  *anilist.AnimeCollection
		LocalFiles       []*LocalFile
		AnizipCache      *anizip.Cache
		SilencedMediaIds []int
		MetadataProvider *metadata.Provider
	}
)

func NewMissingEpisodes(opts *NewMissingEpisodesOptions) *MissingEpisodes {

	missing := new(MissingEpisodes)
	rateLimiter := limiter.NewLimiter(time.Second, 20)

	groupedLfs := GroupLocalFilesByMediaID(opts.LocalFiles)

	p := pool.NewWithResults[[]*MediaEntryDownloadEpisode]()
	for mId, lfs := range groupedLfs {
		p.Go(func() []*MediaEntryDownloadEpisode {
			entry, found := opts.AnimeCollection.GetListEntryFromMediaId(mId)
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
				LocalFiles:       lfs,
				AnizipMedia:      anizipMedia,
				Progress:         entry.Progress,
				Status:           entry.Status,
				Media:            entry.Media,
				MetadataProvider: opts.MetadataProvider,
			})
			if err != nil {
				return nil
			}

			episodes := downloadInfo.EpisodesToDownload
			// If there are more than 1 episode to download, modify the name of the first episode
			if len(episodes) > 1 {
				episodes = episodes[:1] // keep the first episode
				if episodes[0].Episode != nil {
					episodes[0].Episode.DisplayTitle = episodes[0].Episode.DisplayTitle + fmt.Sprintf(" & %d more", len(downloadInfo.EpisodesToDownload)-1)
				}
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
		return eps[i].BaseAnime.ID < eps[j].BaseAnime.ID
	})

	missing.Episodes = lo.Filter(eps, func(item *MediaEntryEpisode, _ int) bool {
		return !lo.Contains(opts.SilencedMediaIds, item.BaseAnime.ID)
	})

	missing.SilencedEpisodes = lo.Filter(eps, func(item *MediaEntryEpisode, _ int) bool {
		return lo.Contains(opts.SilencedMediaIds, item.BaseAnime.ID)
	})

	return missing

}
