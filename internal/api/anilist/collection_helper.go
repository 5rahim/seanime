package anilist

import (
	"time"

	"github.com/goccy/go-json"
)

type (
	AnimeListEntry = AnimeCollection_MediaListCollection_Lists_Entries
	AnimeList      = AnimeCollection_MediaListCollection_Lists

	EntryDate struct {
		Year  *int `json:"year,omitempty"`
		Month *int `json:"month,omitempty"`
		Day   *int `json:"day,omitempty"`
	}
)

func (ac *AnimeCollection) GetListEntryFromAnimeId(id int) (*AnimeListEntry, bool) {
	if ac == nil || ac.MediaListCollection == nil {
		return nil, false
	}

	var entry *AnimeCollection_MediaListCollection_Lists_Entries
	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if e.Media.ID == id {
				entry = e
				break
			}
		}
	}
	if entry == nil {
		return nil, false
	}

	return entry, true
}

func (ac *AnimeCollection) GetAllAnime() []*BaseAnime {
	if ac == nil {
		return make([]*BaseAnime, 0)
	}

	var ret []*BaseAnime
	addedId := make(map[int]bool)
	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if _, ok := addedId[e.Media.ID]; !ok {
				ret = append(ret, e.Media)
				addedId[e.Media.ID] = true
			}
		}
	}
	return ret
}

func (ac *AnimeCollection) FindAnime(mediaId int) (*BaseAnime, bool) {
	if ac == nil {
		return nil, false
	}

	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if e.Media.ID == mediaId {
				return e.Media, true
			}
		}
	}
	return nil, false
}

func (ac *AnimeCollectionWithRelations) GetListEntryFromMediaId(id int) (*AnimeCollectionWithRelations_MediaListCollection_Lists_Entries, bool) {

	if ac == nil || ac.MediaListCollection == nil {
		return nil, false
	}

	var entry *AnimeCollectionWithRelations_MediaListCollection_Lists_Entries
	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if e.Media.ID == id {
				entry = e
				break
			}
		}
	}
	if entry == nil {
		return nil, false
	}

	return entry, true
}

func (ac *AnimeCollectionWithRelations) GetAllAnime() []*CompleteAnime {

	var ret []*CompleteAnime
	addedId := make(map[int]bool)
	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if _, ok := addedId[e.Media.ID]; !ok {
				ret = append(ret, e.Media)
				addedId[e.Media.ID] = true
			}
		}
	}
	return ret
}

func (ac *AnimeCollectionWithRelations) FindAnime(mediaId int) (*CompleteAnime, bool) {
	for _, l := range ac.MediaListCollection.Lists {
		if l.Entries == nil || len(l.Entries) == 0 {
			continue
		}
		for _, e := range l.Entries {
			if e.Media.ID == mediaId {
				return e.Media, true
			}
		}
	}
	return nil, false
}

type IFuzzyDate interface {
	GetYear() *int
	GetMonth() *int
	GetDay() *int
}

func FuzzyDateToString(d IFuzzyDate) string {
	if d == nil {
		return ""
	}
	return fuzzyDateToString(d.GetYear(), d.GetMonth(), d.GetDay())
}

func ToEntryStartDate(d *AnimeCollection_MediaListCollection_Lists_Entries_StartedAt) string {
	if d == nil {
		return ""
	}
	return fuzzyDateToString(d.GetYear(), d.GetMonth(), d.GetDay())
}

func ToEntryCompletionDate(d *AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt) string {
	if d == nil {
		return ""
	}
	return fuzzyDateToString(d.GetYear(), d.GetMonth(), d.GetDay())
}

func fuzzyDateToString(year *int, month *int, day *int) string {
	_year := 0
	if year != nil {
		_year = *year
	}
	if _year == 0 {
		return ""
	}
	_month := 0
	if month != nil {
		_month = *month
	}
	_day := 0
	if day != nil {
		_day = *day
	}
	return time.Date(_year, time.Month(_month), _day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

// AddEntryToList adds an entry to the appropriate list based on the provided status.
// If no list exists with the given status, a new list is created.
func (mc *AnimeCollection_MediaListCollection) AddEntryToList(entry *AnimeCollection_MediaListCollection_Lists_Entries, status MediaListStatus) {
	if mc == nil || entry == nil {
		return
	}

	// Initialize Lists slice if nil
	if mc.Lists == nil {
		mc.Lists = make([]*AnimeCollection_MediaListCollection_Lists, 0)
	}

	// Find existing list with the target status
	for _, list := range mc.Lists {
		if list.Status != nil && *list.Status == status {
			// Found the list, add the entry
			if list.Entries == nil {
				list.Entries = make([]*AnimeCollection_MediaListCollection_Lists_Entries, 0)
			}
			list.Entries = append(list.Entries, entry)
			return
		}
	}

	// No list found with the target status, create a new one
	newList := &AnimeCollection_MediaListCollection_Lists{
		Status:  &status,
		Entries: []*AnimeCollection_MediaListCollection_Lists_Entries{entry},
	}
	mc.Lists = append(mc.Lists, newList)
}

func (ac *AnimeCollection) Copy() *AnimeCollection {
	if ac == nil {
		return nil
	}
	marshaled, err := json.Marshal(ac)
	if err != nil {
		return nil
	}
	var copy AnimeCollection
	err = json.Unmarshal(marshaled, &copy)
	if err != nil {
		return nil
	}
	return &copy
}

func (ac *AnimeList) CopyT() *AnimeCollection_MediaListCollection_Lists {
	if ac == nil {
		return nil
	}
	marshaled, err := json.Marshal(ac)
	if err != nil {
		return nil
	}
	var copy AnimeCollection_MediaListCollection_Lists
	err = json.Unmarshal(marshaled, &copy)
	if err != nil {
		return nil
	}
	return &copy
}
