package torrentstream

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"strconv"
	"sync"
)

type (
	// StreamCollection is used to "complete" the anime.LibraryCollection if the user chooses
	// to include torrent streams in the library view.
	StreamCollection struct {
		ContinueWatchingList []*anime.Episode             `json:"continueWatchingList"`
		Anime                []*anilist.BaseAnime         `json:"anime"`
		ListData             map[int]*anime.EntryListData `json:"listData"`
	}

	HydrateStreamCollectionOptions struct {
		AnimeCollection   *anilist.AnimeCollection
		LibraryCollection *anime.LibraryCollection
		MetadataProvider  metadata.Provider
	}
)

func (r *Repository) HydrateStreamCollection(opts *HydrateStreamCollectionOptions) {

	reqEvent := new(anime.AnimeLibraryStreamCollectionRequestedEvent)
	reqEvent.AnimeCollection = opts.AnimeCollection
	reqEvent.LibraryCollection = opts.LibraryCollection
	err := hook.GlobalHookManager.OnAnimeLibraryStreamCollectionRequested().Trigger(reqEvent)
	if err != nil {
		return
	}
	opts.AnimeCollection = reqEvent.AnimeCollection
	opts.LibraryCollection = reqEvent.LibraryCollection

	lists := opts.AnimeCollection.MediaListCollection.GetLists()
	// Get the anime that are currently being watched
	var currentlyWatching *anilist.AnimeCollection_MediaListCollection_Lists
	for _, list := range lists {
		if list.Status == nil {
			continue
		}
		if *list.Status == anilist.MediaListStatusCurrent {
			currentlyWatching = list.CopyT()
			continue
		}
	}
	for _, list := range lists {
		if list.Status == nil {
			continue
		}
		if *list.Status == anilist.MediaListStatusRepeating {
			if currentlyWatching == nil {
				currentlyWatching = list
			} else {
				for _, entry := range list.Entries {
					currentlyWatching.Entries = append(currentlyWatching.Entries, entry)
				}
			}
			break
		}
	}

	if currentlyWatching == nil {
		return
	}

	ret := &StreamCollection{
		ContinueWatchingList: make([]*anime.Episode, 0),
		Anime:                make([]*anilist.BaseAnime, 0),
		ListData:             make(map[int]*anime.EntryListData),
	}

	visitedMediaIds := make(map[int]struct{})

	animeAdded := make(map[int]*anilist.AnimeListEntry)

	// Go through each entry in the currently watching list
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(currentlyWatching.Entries))
	for _, entry := range currentlyWatching.Entries {
		go func(entry *anilist.AnimeListEntry) {
			defer wg.Done()

			mu.Lock()
			if _, found := visitedMediaIds[entry.GetMedia().GetID()]; found {
				mu.Unlock()
				return
			}
			// Get the next episode to watch
			// i.e. if the user has watched episode 1, the next episode to watch is 2
			nextEpisodeToWatch := entry.GetProgressSafe() + 1
			if nextEpisodeToWatch > entry.GetMedia().GetCurrentEpisodeCount() {
				mu.Unlock()
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
			animeMetadata, err := opts.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
			if err != nil {
				r.logger.Error().Err(err).Msg("torrentstream: could not fetch AniDB media")
				return
			}

			_, found := animeMetadata.FindEpisode(strconv.Itoa(nextEpisodeToWatch))
			//if !found {
			//	r.logger.Error().Msg("torrentstream: could not find episode in AniDB")
			//	return
			//}

			progressOffset := 0
			anidbEpisode := strconv.Itoa(nextEpisodeToWatch)
			if anime.FindDiscrepancy(entry.GetMedia(), animeMetadata) == anime.DiscrepancyAniListCountsEpisodeZero {
				progressOffset = 1
				if nextEpisodeToWatch == 1 {
					anidbEpisode = "S1"
				}
			}

			// Add the anime & episode
			episode := anime.NewEpisode(&anime.NewEpisodeOptions{
				LocalFile:            nil,
				OptionalAniDBEpisode: anidbEpisode,
				AnimeMetadata:        animeMetadata,
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
				episode.EpisodeMetadata = &anime.EpisodeMetadata{
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

	libraryAnimeMap := make(map[int]struct{})

	// Remove anime that are already in the library collection
	for _, list := range opts.LibraryCollection.Lists {
		if list.Status == anilist.MediaListStatusCurrent {
			for _, entry := range list.Entries {
				libraryAnimeMap[entry.MediaId] = struct{}{}
				if _, found := animeAdded[entry.MediaId]; found {
					delete(animeAdded, entry.MediaId)
				}
			}
		}
	}

	for _, entry := range currentlyWatching.Entries {
		if _, found := libraryAnimeMap[entry.GetMedia().GetID()]; found {
			continue
		}
		if *entry.GetMedia().GetStatus() == anilist.MediaStatusNotYetReleased {
			continue
		}
		animeAdded[entry.GetMedia().GetID()] = entry
	}

	for _, a := range animeAdded {
		ret.Anime = append(ret.Anime, a.GetMedia())
		ret.ListData[a.GetMedia().GetID()] = &anime.EntryListData{
			Progress:    a.GetProgressSafe(),
			Score:       a.GetScoreSafe(),
			Status:      a.GetStatus(),
			Repeat:      a.GetRepeatSafe(),
			StartedAt:   anilist.FuzzyDateToString(a.StartedAt),
			CompletedAt: anilist.FuzzyDateToString(a.CompletedAt),
		}
	}

	if len(ret.ContinueWatchingList) == 0 && len(ret.Anime) == 0 {
		return
	}

	sc := &anime.StreamCollection{
		ContinueWatchingList: ret.ContinueWatchingList,
		Anime:                ret.Anime,
		ListData:             ret.ListData,
	}

	event := new(anime.AnimeLibraryStreamCollectionEvent)
	event.StreamCollection = sc
	err = hook.GlobalHookManager.OnAnimeLibraryStreamCollection().Trigger(event)
	if err != nil {
		return
	}

	opts.LibraryCollection.Stream = event.StreamCollection
}
