package anime

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/samber/lo"
)

type (
	UpcomingEpisodes struct {
		Episodes []*UpcomingEpisode `json:"episodes"`
	}

	UpcomingEpisode struct {
		MediaId         int                `json:"mediaId"`
		EpisodeNumber   int                `json:"episodeNumber"`
		AiringAt        int64              `json:"airingAt"`
		TimeUntilAiring int                `json:"timeUntilAiring"`
		BaseAnime       *anilist.BaseAnime `json:"baseAnime"`
		EpisodeMetadata *EpisodeMetadata   `json:"episodeMetadata,omitempty"`
	}

	NewUpcomingEpisodesOptions struct {
		AnimeCollection     *anilist.AnimeCollection
		MetadataProviderRef *util.Ref[metadata_provider.Provider]
	}
)

func NewUpcomingEpisodes(opts *NewUpcomingEpisodesOptions) *UpcomingEpisodes {
	upcoming := new(UpcomingEpisodes)

	reqEvent := new(UpcomingEpisodesRequestedEvent)
	reqEvent.AnimeCollection = opts.AnimeCollection
	reqEvent.UpcomingEpisodes = upcoming
	err := hook.GlobalHookManager.OnUpcomingEpisodesRequested().Trigger(reqEvent)
	if err != nil {
		return nil
	}
	opts.AnimeCollection = reqEvent.AnimeCollection
	upcoming = reqEvent.UpcomingEpisodes

	// Default prevented by hook, return the upcoming episodes
	if reqEvent.DefaultPrevented {
		event := new(UpcomingEpisodesEvent)
		event.UpcomingEpisodes = upcoming
		err = hook.GlobalHookManager.OnUpcomingEpisodes().Trigger(event)
		if err != nil {
			return nil
		}
		return event.UpcomingEpisodes
	}

	// Get all media with next airing episodes
	allMedia := opts.AnimeCollection.GetAllAnime()
	mediaWithNextAiring := lo.Filter(allMedia, func(item *anilist.BaseAnime, _ int) bool {
		return item.NextAiringEpisode != nil && item.NextAiringEpisode.Episode > 0
	})

	// Sort by time until airing
	sort.Slice(mediaWithNextAiring, func(i, j int) bool {
		return mediaWithNextAiring[i].NextAiringEpisode.TimeUntilAiring < mediaWithNextAiring[j].NextAiringEpisode.TimeUntilAiring
	})

	rateLimiter := limiter.NewLimiter(time.Second, 20)
	upcomingEps := make([]*UpcomingEpisode, 0, len(mediaWithNextAiring))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, media := range mediaWithNextAiring {
		wg.Add(1)
		go func(media *anilist.BaseAnime) {
			defer wg.Done()

			entry, found := opts.AnimeCollection.GetListEntryFromAnimeId(media.ID)
			if !found {
				return
			}

			if entry.Status == nil || *entry.Status == anilist.MediaListStatusDropped {
				return
			}

			if media.NextAiringEpisode.Episode <= 0 {
				return
			}

			upcomingEp := &UpcomingEpisode{
				MediaId:         media.ID,
				EpisodeNumber:   media.NextAiringEpisode.Episode,
				AiringAt:        int64(media.NextAiringEpisode.AiringAt),
				TimeUntilAiring: media.NextAiringEpisode.TimeUntilAiring,
				BaseAnime:       media,
			}

			// Fetch episode metadata
			rateLimiter.Wait()
			animeMetadata, err := opts.MetadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, media.ID)
			if err == nil && animeMetadata != nil {
				// Get episode metadata
				metadataWrapper := opts.MetadataProviderRef.Get().GetAnimeMetadataWrapper(media, animeMetadata)
				episodeStr := strconv.Itoa(media.NextAiringEpisode.Episode)
				epMetadata := metadataWrapper.GetEpisodeMetadata(episodeStr)

				upcomingEp.EpisodeMetadata = &EpisodeMetadata{
					AnidbId:  epMetadata.AnidbId,
					Image:    epMetadata.Image,
					AirDate:  epMetadata.AirDate,
					Length:   epMetadata.Length,
					Summary:  epMetadata.Summary,
					Overview: epMetadata.Overview,
					Title:    epMetadata.Title,
				}
			}

			mu.Lock()
			upcomingEps = append(upcomingEps, upcomingEp)
			mu.Unlock()
		}(media)
	}
	wg.Wait()

	upcomingEps = lo.Filter(upcomingEps, func(item *UpcomingEpisode, _ int) bool {
		return item != nil
	})

	// Sort by time until airing
	sort.Slice(upcomingEps, func(i, j int) bool {
		return upcomingEps[i].TimeUntilAiring < upcomingEps[j].TimeUntilAiring
	})

	upcoming.Episodes = upcomingEps

	// Event
	event := new(UpcomingEpisodesEvent)
	event.UpcomingEpisodes = upcoming
	err = hook.GlobalHookManager.OnUpcomingEpisodes().Trigger(event)
	if err != nil {
		return nil
	}

	return event.UpcomingEpisodes
}
