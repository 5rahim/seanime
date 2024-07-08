package anilist

import "time"

type (
	MediaListEntry = AnimeCollection_MediaListCollection_Lists_Entries

	EntryDate struct {
		Year  *int `json:"year,omitempty"`
		Month *int `json:"month,omitempty"`
		Day   *int `json:"day,omitempty"`
	}
)

func (ac *AnimeCollection) GetListEntryFromMediaId(id int) (*MediaListEntry, bool) {

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

func (ac *AnimeCollection) GetAllMedia() []*BaseMedia {

	var ret []*BaseMedia
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

func (ac *AnimeCollection) FindMedia(mediaId int) (*BaseMedia, bool) {
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

func ToEntryDate(d IFuzzyDate) string {
	if d == nil {
		return ""
	}
	return toEntryDate(d.GetYear(), d.GetMonth(), d.GetDay())
}

func ToEntryStartDate(d *AnimeCollection_MediaListCollection_Lists_Entries_StartedAt) string {
	if d == nil {
		return ""
	}
	return toEntryDate(d.GetYear(), d.GetMonth(), d.GetDay())
}

func ToEntryCompletionDate(d *AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt) string {
	if d == nil {
		return ""
	}
	return toEntryDate(d.GetYear(), d.GetMonth(), d.GetDay())
}

func toEntryDate(year *int, month *int, day *int) string {
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
	return time.Date(_year, time.Month(_month), _day, 0, 0, 0, 0, time.Local).UTC().Format(time.RFC3339)
}
