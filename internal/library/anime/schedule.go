package anime

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/customsource"
	"seanime/internal/hook"
	"time"

	"github.com/samber/lo"
)

type ScheduleItem struct {
	MediaId int    `json:"mediaId"`
	Title   string `json:"title"`
	// Time is in 15:04 format
	Time string `json:"time"`
	// DateTime is in UTC
	DateTime       time.Time `json:"dateTime"`
	Image          string    `json:"image"`
	EpisodeNumber  int       `json:"episodeNumber"`
	IsMovie        bool      `json:"isMovie"`
	IsSeasonFinale bool      `json:"isSeasonFinale"`
}

func GetScheduleItems(animeSchedule *anilist.AnimeAiringSchedule, animeCollection *anilist.AnimeCollection) []*ScheduleItem {
	animeEntryMap := make(map[int]*anilist.AnimeListEntry)
	for _, list := range animeCollection.MediaListCollection.GetLists() {
		for _, entry := range list.GetEntries() {
			if customsource.IsExtensionId(entry.Media.GetID()) {
				continue
			}
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

	formatNodeItem := func(node animeScheduleNode, entry *anilist.AnimeListEntry) *ScheduleItem {
		t := time.Unix(int64(node.GetAiringAt()), 0)
		item := &ScheduleItem{
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

	formatPart := func(m animeScheduleMedia) ([]*ScheduleItem, bool) {
		if m == nil {
			return nil, false
		}
		ret := make([]*ScheduleItem, 0)
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

	allItems := make([]*ScheduleItem, 0)
	allItems = append(allItems, ongoingItems...)
	allItems = append(allItems, ongoingNextItems...)
	allItems = append(allItems, precedingItems...)
	allItems = append(allItems, upcomingItems...)
	allItems = append(allItems, upcomingNextItems...)

	ret := lo.UniqBy(allItems, func(item *ScheduleItem) string {
		if item == nil {
			return ""
		}
		return fmt.Sprintf("%d-%d-%d", item.MediaId, item.EpisodeNumber, item.DateTime.Unix())
	})

	event := &AnimeScheduleItemsEvent{
		AnimeCollection: animeCollection,
		Items:           ret,
	}
	err := hook.GlobalHookManager.OnAnimeScheduleItems().Trigger(event)
	if err != nil {
		return ret
	}

	return event.Items
}
