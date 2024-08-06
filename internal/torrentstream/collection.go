package torrentstream

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/library/anime"
	"strconv"
	"sync"
)

type (
	// StreamCollection is used to "complete" the anime.LibraryCollection if the user chooses
	// to include torrent streams in the library view.
	StreamCollection struct {
		ContinueWatchingList []*anime.AnimeEntryEpisode        `json:"continueWatchingList"`
		Anime                []*anilist.BaseAnime              `json:"anime"`
		ListData             map[int]*anime.AnimeEntryListData `json:"listData"`
	}

	HydrateStreamCollectionOptions struct {
		LibraryCollection *anime.LibraryCollection
		AnizipCache       *anizip.Cache
	}
)

func (r *Repository) HydrateStreamCollection(opts *HydrateStreamCollectionOptions) {
	if r.settings.IsAbsent() || !r.settings.MustGet().Enabled {
		return
	}

	animeCollection, err := r.platform.GetAnimeCollection(false)
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: could not get anime collection")
		return
	}

	lists := animeCollection.MediaListCollection.GetLists()
	// Get the anime that are currently being watched
	var currentlyWatching *anilist.AnimeCollection_MediaListCollection_Lists
	for _, list := range lists {
		if list.Status == nil {
			continue
		}
		if *list.Status == anilist.MediaListStatusCurrent {
			currentlyWatching = list
			break
		}
	}

	if currentlyWatching == nil {
		return
	}

	ret := &StreamCollection{
		ContinueWatchingList: make([]*anime.AnimeEntryEpisode, 0),
		Anime:                make([]*anilist.BaseAnime, 0),
		ListData:             make(map[int]*anime.AnimeEntryListData),
	}

	visitedMediaIds := make(map[int]struct{})

	animeAdded := make(map[int]*anilist.MediaListEntry)

	// Go through each entry in the currently watching list
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(currentlyWatching.Entries))
	for _, entry := range currentlyWatching.Entries {
		go func(entry *anilist.MediaListEntry) {
			defer wg.Done()

			mu.Lock()
			if _, found := visitedMediaIds[entry.GetMedia().GetID()]; found {
				return
			}
			// Get the next episode to watch
			// i.e. if the user has watched episode 1, the next episode to watch is 2
			nextEpisodeToWatch := entry.GetProgressSafe() + 1
			if nextEpisodeToWatch > entry.GetMedia().GetCurrentEpisodeCount() {
				return // Skip this entry if the user has watched all episodes
			}
			mediaId := entry.GetMedia().GetID()
			visitedMediaIds[mediaId] = struct{}{}
			mu.Unlock()
			// Check if the anime's "next episode to watch" is already in the library collection
			// If it is, we don't need to add it to the stream collection
			for _, libraryEp := range opts.LibraryCollection.ContinueWatchingList {
				if libraryEp.BaseAnime.ID == mediaId && libraryEp.GetProgressNumber() == nextEpisodeToWatch {
					return
				}
			}

			if *entry.GetMedia().GetStatus() == anilist.MediaStatusNotYetReleased {
				return
			}

			// Get the media info
			anizipMedia, err := anizip.FetchAniZipMediaC("anilist", mediaId, r.anizipCache)
			if err != nil {
				r.logger.Error().Err(err).Msg("torrentstream: could not fetch AniDB media")
				return
			}

			_, found := anizipMedia.FindEpisode(strconv.Itoa(nextEpisodeToWatch))
			//if !found {
			//	r.logger.Error().Msg("torrentstream: could not find episode in AniDB")
			//	return
			//}

			progressOffset := 0
			anidbEpisode := strconv.Itoa(nextEpisodeToWatch)
			if anime.HasDiscrepancy(entry.GetMedia(), anizipMedia) {
				progressOffset = 1
				if nextEpisodeToWatch == 1 {
					anidbEpisode = "S1"
				}
			}

			// Add the anime & episode
			episode := anime.NewAnimeEntryEpisode(&anime.NewAnimeEntryEpisodeOptions{
				LocalFile:            nil,
				OptionalAniDBEpisode: anidbEpisode,
				AnizipMedia:          anizipMedia,
				Media:                entry.GetMedia(),
				ProgressOffset:       progressOffset,
				IsDownloaded:         false,
				MetadataProvider:     r.metadataProvider,
			})
			if !found {
				episode.EpisodeTitle = entry.GetMedia().GetPreferredTitle()
				episode.DisplayTitle = fmt.Sprintf("Episode %d", nextEpisodeToWatch)
				episode.ProgressNumber = nextEpisodeToWatch
				episode.EpisodeNumber = nextEpisodeToWatch
				episode.EpisodeMetadata = &anime.AnimeEntryEpisodeMetadata{
					Image: entry.GetMedia().GetBannerImageSafe(),
				}
			}

			if episode == nil {
				r.logger.Error().Msg("torrentstream: could not get anime entry episode")
				return
			}

			mu.Lock()
			ret.ContinueWatchingList = append(ret.ContinueWatchingList, episode)
			animeAdded[mediaId] = entry
			mu.Unlock()
		}(entry)
	}
	wg.Wait()

	// Remove anime that are already in the library collection
	for _, list := range opts.LibraryCollection.Lists {
		if list.Status == anilist.MediaListStatusCurrent {
			for _, entry := range list.Entries {
				if _, found := animeAdded[entry.MediaId]; found {
					delete(animeAdded, entry.MediaId)
				}
			}
		}
	}

	for _, a := range animeAdded {
		ret.Anime = append(ret.Anime, a.GetMedia())
		ret.ListData[a.GetMedia().GetID()] = &anime.AnimeEntryListData{
			Progress:    a.GetProgressSafe(),
			Score:       a.GetScoreSafe(),
			Status:      a.GetStatus(),
			StartedAt:   anilist.FuzzyDateToString(a.StartedAt),
			CompletedAt: anilist.FuzzyDateToString(a.CompletedAt),
		}

	}

	if len(ret.ContinueWatchingList) == 0 {
		return
	}

	opts.LibraryCollection.Stream = &anime.StreamCollection{
		ContinueWatchingList: ret.ContinueWatchingList,
		Anime:                ret.Anime,
		ListData:             ret.ListData,
	}
}
