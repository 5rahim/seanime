package handlers

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/torrentstream"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

// HandleGetLibraryCollection
//
//	@summary returns the main local anime collection.
//	@desc This creates a new LibraryCollection struct and returns it.
//	@desc This is used to get the main anime collection of the user.
//	@desc It uses the cached Anilist anime collection for the GET method.
//	@desc It refreshes the AniList anime collection if the POST method is used.
//	@route /api/v1/library/collection [GET,POST]
//	@returns anime.LibraryCollection
func (h *Handler) HandleGetLibraryCollection(c echo.Context) error {

	animeCollection, err := h.App.GetAnimeCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if animeCollection == nil {
		return h.RespondWithData(c, &anime.LibraryCollection{})
	}

	originalAnimeCollection := animeCollection

	var lfs []*anime.LocalFile
	nakamaLibrary, fromNakama := h.App.NakamaManager.GetHostAnimeLibrary()
	if fromNakama {
		// Save the original anime collection to restore it later
		originalAnimeCollection = animeCollection.Copy()
		lfs = nakamaLibrary.LocalFiles
		// Merge missing media entries into the collection
		currentMediaIds := make(map[int]struct{})
		for _, list := range animeCollection.MediaListCollection.GetLists() {
			for _, entry := range list.GetEntries() {
				currentMediaIds[entry.GetMedia().GetID()] = struct{}{}
			}
		}

		nakamaMediaIds := make(map[int]struct{})
		for _, lf := range lfs {
			if lf.MediaId > 0 {
				nakamaMediaIds[lf.MediaId] = struct{}{}
			}
		}

		missingMediaIds := make(map[int]struct{})
		for _, lf := range lfs {
			if lf.MediaId > 0 {
				if _, ok := currentMediaIds[lf.MediaId]; !ok {
					missingMediaIds[lf.MediaId] = struct{}{}
				}
			}
		}

		for _, list := range nakamaLibrary.AnimeCollection.MediaListCollection.GetLists() {
			for _, entry := range list.GetEntries() {
				if _, ok := missingMediaIds[entry.GetMedia().GetID()]; ok {
					// create a new entry with blank list data
					newEntry := &anilist.AnimeListEntry{
						ID:     entry.GetID(),
						Media:  entry.GetMedia(),
						Status: &[]anilist.MediaListStatus{anilist.MediaListStatusPlanning}[0],
					}
					animeCollection.MediaListCollection.AddEntryToList(newEntry, anilist.MediaListStatusPlanning)
				}
			}
		}

	} else {
		lfs, _, err = db_bridge.GetLocalFiles(h.App.Database)
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	libraryCollection, err := anime.NewLibraryCollection(c.Request().Context(), &anime.NewLibraryCollectionOptions{
		AnimeCollection:  animeCollection,
		Platform:         h.App.AnilistPlatform,
		LocalFiles:       lfs,
		MetadataProvider: h.App.MetadataProvider,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Restore the original anime collection if it was modified
	if fromNakama {
		*animeCollection = *originalAnimeCollection
	}

	if !fromNakama {
		if (h.App.SecondarySettings.Torrentstream != nil && h.App.SecondarySettings.Torrentstream.Enabled && h.App.SecondarySettings.Torrentstream.IncludeInLibrary) ||
			(h.App.Settings.GetLibrary() != nil && h.App.Settings.GetLibrary().EnableOnlinestream && h.App.Settings.GetLibrary().IncludeOnlineStreamingInLibrary) ||
			(h.App.SecondarySettings.Debrid != nil && h.App.SecondarySettings.Debrid.Enabled && h.App.SecondarySettings.Debrid.IncludeDebridStreamInLibrary) {
			h.App.TorrentstreamRepository.HydrateStreamCollection(&torrentstream.HydrateStreamCollectionOptions{
				AnimeCollection:   animeCollection,
				LibraryCollection: libraryCollection,
				MetadataProvider:  h.App.MetadataProvider,
			})
		}
	}

	// Add and remove necessary metadata when hydrating from Nakama
	if fromNakama {
		for _, ep := range libraryCollection.ContinueWatchingList {
			ep.IsNakamaEpisode = true
		}
		for _, list := range libraryCollection.Lists {
			for _, entry := range list.Entries {
				if entry.EntryLibraryData == nil {
					continue
				}
				entry.NakamaEntryLibraryData = &anime.NakamaEntryLibraryData{
					UnwatchedCount: entry.EntryLibraryData.UnwatchedCount,
					MainFileCount:  entry.EntryLibraryData.MainFileCount,
				}
				entry.EntryLibraryData = nil
			}
		}
	}

	// Hydrate total library size
	if libraryCollection != nil && libraryCollection.Stats != nil {
		libraryCollection.Stats.TotalSize = util.Bytes(h.App.TotalLibrarySize)
	}

	return h.RespondWithData(c, libraryCollection)
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

type AnimeCollectionScheduleItem struct {
	MediaId        int       `json:"mediaId"`
	Title          string    `json:"title"`
	Time           string    `json:"time"`
	DateTime       time.Time `json:"dateTime"`
	Image          string    `json:"image"`
	EpisodeNumber  int       `json:"episodeNumber"`
	IsMovie        bool      `json:"isMovie"`
	IsSeasonFinale bool      `json:"isSeasonFinale"`
}

var animeScheduleCache = result.NewCache[int, []*AnimeCollectionScheduleItem]()

// HandleGetAnimeCollectionSchedule
//
//	@summary returns anime collection schedule
//	@desc This is used by the "Schedule" page to display the anime schedule.
//	@route /api/v1/library/schedule [GET]
//	@returns []handlers.AnimeCollectionScheduleItem
func (h *Handler) HandleGetAnimeCollectionSchedule(c echo.Context) error {

	// Invalidate the cache when the Anilist collection is refreshed
	h.App.AddOnRefreshAnilistCollectionFunc("HandleGetAnimeCollectionSchedule", func() {
		animeScheduleCache.Delete(1)
	})

	if ret, ok := animeScheduleCache.Get(1); ok {
		return h.RespondWithData(c, ret)
	}

	animeSchedule, err := h.App.AnilistPlatform.GetAnimeAiringSchedule(c.Request().Context())
	if err != nil {
		return h.RespondWithError(c, err)
	}

	animeCollection, err := h.App.GetAnimeCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	animeEntryMap := make(map[int]*anilist.AnimeListEntry)
	for _, list := range animeCollection.MediaListCollection.GetLists() {
		for _, entry := range list.GetEntries() {
			animeEntryMap[entry.GetMedia().GetID()] = entry
		}
	}

	type animeScheduleNode interface {
		GetAiringAt() int
		GetTimeUntilAiring() int
		GetEpisode() int
	}

	type animeScheduleMedia interface {
		GetMedia() []*anilist.AnimeSchedule
	}

	formatNodeItem := func(node animeScheduleNode, entry *anilist.AnimeListEntry) *AnimeCollectionScheduleItem {
		t := time.Unix(int64(node.GetAiringAt()), 0)
		item := &AnimeCollectionScheduleItem{
			MediaId:        entry.GetMedia().GetID(),
			Title:          *entry.GetMedia().GetTitle().GetUserPreferred(),
			Time:           t.UTC().Format("15:04"),
			DateTime:       t.UTC(),
			Image:          entry.GetMedia().GetCoverImageSafe(),
			EpisodeNumber:  node.GetEpisode(),
			IsMovie:        entry.GetMedia().IsMovie(),
			IsSeasonFinale: false,
		}
		if entry.GetMedia().GetTotalEpisodeCount() > 0 && node.GetEpisode() == entry.GetMedia().GetTotalEpisodeCount() {
			item.IsSeasonFinale = true
		}
		return item
	}

	formatPart := func(m animeScheduleMedia) ([]*AnimeCollectionScheduleItem, bool) {
		if m == nil {
			return nil, false
		}
		ret := make([]*AnimeCollectionScheduleItem, 0)
		for _, m := range m.GetMedia() {
			entry, ok := animeEntryMap[m.GetID()]
			if !ok || entry.Status == nil || *entry.Status == anilist.MediaListStatusDropped {
				continue
			}
			for _, n := range m.GetPrevious().GetNodes() {
				ret = append(ret, formatNodeItem(n, entry))
			}
			for _, n := range m.GetUpcoming().GetNodes() {
				ret = append(ret, formatNodeItem(n, entry))
			}
		}
		return ret, true
	}

	ongoingItems, _ := formatPart(animeSchedule.GetOngoing())
	ongoingNextItems, _ := formatPart(animeSchedule.GetOngoingNext())
	precedingItems, _ := formatPart(animeSchedule.GetPreceding())
	upcomingItems, _ := formatPart(animeSchedule.GetUpcoming())
	upcomingNextItems, _ := formatPart(animeSchedule.GetUpcomingNext())

	allItems := make([]*AnimeCollectionScheduleItem, 0)
	allItems = append(allItems, ongoingItems...)
	allItems = append(allItems, ongoingNextItems...)
	allItems = append(allItems, precedingItems...)
	allItems = append(allItems, upcomingItems...)
	allItems = append(allItems, upcomingNextItems...)

	ret := lo.UniqBy(allItems, func(item *AnimeCollectionScheduleItem) string {
		if item == nil {
			return ""
		}
		return fmt.Sprintf("%d-%d-%d", item.MediaId, item.EpisodeNumber, item.DateTime.Unix())
	})

	animeScheduleCache.SetT(1, ret, 1*time.Hour)

	return h.RespondWithData(c, ret)
}

// HandleAddUnknownMedia
//
//	@summary adds the given media to the user's AniList planning collections
//	@desc Since media not found in the user's AniList collection are not displayed in the library, this route is used to add them.
//	@desc The response is ignored in the frontend, the client should just refetch the entire library collection.
//	@route /api/v1/library/unknown-media [POST]
//	@returns anilist.AnimeCollection
func (h *Handler) HandleAddUnknownMedia(c echo.Context) error {

	type body struct {
		MediaIds []int `json:"mediaIds"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Add non-added media entries to AniList collection
	if err := h.App.AnilistPlatform.AddMediaToCollection(c.Request().Context(), b.MediaIds); err != nil {
		return h.RespondWithError(c, errors.New("error: Anilist responded with an error, this is most likely a rate limit issue"))
	}

	// Bypass the cache
	animeCollection, err := h.App.GetAnimeCollection(true)
	if err != nil {
		return h.RespondWithError(c, errors.New("error: Anilist responded with an error, wait one minute before refreshing"))
	}

	return h.RespondWithData(c, animeCollection)

}
