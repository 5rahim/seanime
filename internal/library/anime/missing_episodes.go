package anime

import (
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/sourcegraph/conc/pool"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/util/limiter"
	"sort"
	"time"
)

type (
	MissingEpisodes struct {
		Episodes         []*Episode `json:"episodes"`
		SilencedEpisodes []*Episode `json:"silencedEpisodes"`
	}

	NewMissingEpisodesOptions struct {
		AnimeCollection  *anilist.AnimeCollection
		LocalFiles       []*LocalFile
		SilencedMediaIds []int
		MetadataProvider metadata.Provider
	}
)

func NewMissingEpisodes(opts *NewMissingEpisodesOptions) *MissingEpisodes {

	missing := new(MissingEpisodes)
	rateLimiter := limiter.NewLimiter(time.Second, 20)

	groupedLfs := GroupLocalFilesByMediaID(opts.LocalFiles)

	p := pool.NewWithResults[[]*EntryDownloadEpisode]()
	for mId, lfs := range groupedLfs {
		p.Go(func() []*EntryDownloadEpisode {
			entry, found := opts.AnimeCollection.GetListEntryFromAnimeId(mId)
			if !found {
				return nil
			}

			// Skip if the status is nil or dropped
			if entry.Status == nil || *entry.Status == anilist.MediaListStatusDropped {
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
			// Fetch anime metadata
			animeMetadata, err := opts.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, entry.Media.ID)
			if err != nil {
				return nil
			}

			// Get download info
			downloadInfo, err := NewEntryDownloadInfo(&NewEntryDownloadInfoOptions{
				LocalFiles:       lfs,
				AnimeMetadata:    animeMetadata,
				Media:            entry.Media,
				Progress:         entry.Progress,
				Status:           entry.Status,
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
	epsToDownload = lo.Filter(epsToDownload, func(item []*EntryDownloadEpisode, _ int) bool {
		return item != nil
	})

	// Flatten
	flattenedEpsToDownload := lo.Flatten(epsToDownload)
	eps := lop.Map(flattenedEpsToDownload, func(item *EntryDownloadEpisode, _ int) *Episode {
		return item.Episode
	})
	// Sort
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].GetEpisodeNumber() < eps[j].GetEpisodeNumber()
	})
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].BaseAnime.ID < eps[j].BaseAnime.ID
	})

	missing.Episodes = lo.Filter(eps, func(item *Episode, _ int) bool {
		return !lo.Contains(opts.SilencedMediaIds, item.BaseAnime.ID)
	})

	missing.SilencedEpisodes = lo.Filter(eps, func(item *Episode, _ int) bool {
		return lo.Contains(opts.SilencedMediaIds, item.BaseAnime.ID)
	})

	return missing

}
