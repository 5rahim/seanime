package anime

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"sort"
	"sync"
	"time"

	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

type (
	MissingEpisodes struct {
		Episodes         []*Episode `json:"episodes"`
		SilencedEpisodes []*Episode `json:"silencedEpisodes"`
	}

	NewMissingEpisodesOptions struct {
		AnimeCollection     *anilist.AnimeCollection
		LocalFiles          []*LocalFile
		SilencedMediaIds    []int
		MetadataProviderRef *util.Ref[metadata_provider.Provider]
	}
)

func NewMissingEpisodes(opts *NewMissingEpisodesOptions) *MissingEpisodes {
	missing := new(MissingEpisodes)

	reqEvent := new(MissingEpisodesRequestedEvent)
	reqEvent.AnimeCollection = opts.AnimeCollection
	reqEvent.LocalFiles = opts.LocalFiles
	reqEvent.SilencedMediaIds = opts.SilencedMediaIds
	reqEvent.MissingEpisodes = missing
	err := hook.GlobalHookManager.OnMissingEpisodesRequested().Trigger(reqEvent)
	if err != nil {
		return nil
	}
	opts.AnimeCollection = reqEvent.AnimeCollection   // Override the anime collection
	opts.LocalFiles = reqEvent.LocalFiles             // Override the local files
	opts.SilencedMediaIds = reqEvent.SilencedMediaIds // Override the silenced media IDs
	missing = reqEvent.MissingEpisodes

	// Default prevented by hook, return the missing episodes
	if reqEvent.DefaultPrevented {
		event := new(MissingEpisodesEvent)
		event.MissingEpisodes = missing
		err = hook.GlobalHookManager.OnMissingEpisodes().Trigger(event)
		if err != nil {
			return nil
		}
		return event.MissingEpisodes
	}

	groupedLfs := GroupLocalFilesByMediaID(opts.LocalFiles)

	rateLimiter := limiter.NewLimiter(time.Second, 20)
	var epsToDownload [][]*EntryDownloadEpisode
	var mu sync.Mutex
	var wg sync.WaitGroup

	for mId, lfs := range groupedLfs {
		wg.Add(1)
		go func(mId int, lfs []*LocalFile) {
			defer wg.Done()

			entry, found := opts.AnimeCollection.GetListEntryFromAnimeId(mId)
			if !found {
				return
			}

			// Skip if the status is nil, dropped or completed
			if entry.Status == nil || *entry.Status == anilist.MediaListStatusDropped || *entry.Status == anilist.MediaListStatusCompleted {
				return
			}

			latestLf, found := FindLatestLocalFileFromGroup(lfs)
			if !found {
				return
			}
			//If the latest local file is the same or higher than the current episode count, skip
			if entry.Media.GetCurrentEpisodeCount() == -1 || entry.Media.GetCurrentEpisodeCount() <= latestLf.GetEpisodeNumber() {
				return
			}
			rateLimiter.Wait()
			// Fetch anime metadata
			animeMetadata, err := opts.MetadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, entry.Media.ID)
			if err != nil {
				return
			}

			// Get download info
			downloadInfo, err := NewEntryDownloadInfo(&NewEntryDownloadInfoOptions{
				LocalFiles:          lfs,
				AnimeMetadata:       animeMetadata,
				Media:               entry.Media,
				Progress:            entry.Progress,
				Status:              entry.Status,
				MetadataProviderRef: opts.MetadataProviderRef,
			})
			if err != nil {
				return
			}

			episodes := downloadInfo.EpisodesToDownload

			sort.Slice(episodes, func(i, j int) bool {
				return episodes[i].Episode.GetEpisodeNumber() < episodes[j].Episode.GetEpisodeNumber()
			})

			// If there are more than 1 episode to download, modify the name of the first episode
			if len(episodes) > 1 {
				episodes = episodes[:1] // keep the first episode
				if episodes[0].Episode != nil {
					episodes[0].Episode.DisplayTitle = episodes[0].Episode.DisplayTitle + fmt.Sprintf(" & %d more", len(downloadInfo.EpisodesToDownload)-1)
				}
			}

			mu.Lock()
			epsToDownload = append(epsToDownload, episodes)
			mu.Unlock()
		}(mId, lfs)
	}
	wg.Wait()

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

	// Event
	event := new(MissingEpisodesEvent)
	event.MissingEpisodes = missing
	err = hook.GlobalHookManager.OnMissingEpisodes().Trigger(event)
	if err != nil {
		return nil
	}

	return event.MissingEpisodes
}
