package anilist

type MangaList = MangaCollection_MediaListCollection_Lists
type MangaListEntry = MangaCollection_MediaListCollection_Lists_Entries

func (ac *MangaCollection) GetListEntryFromMangaId(id int) (*MangaListEntry, bool) {

	if ac == nil || ac.MediaListCollection == nil {
		return nil, false
	}

	var entry *MangaCollection_MediaListCollection_Lists_Entries
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

func (m *BaseManga) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}
func (m *BaseManga) GetRomajiTitleSafe() string {
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	return "N/A"
}

func (m *BaseManga) GetPreferredTitle() string {
	if m.GetTitle().GetUserPreferred() != nil {
		return *m.GetTitle().GetUserPreferred()
	}
	return m.GetTitleSafe()
}

func (m *BaseManga) GetCoverImageSafe() string {
	if m.GetCoverImage().GetExtraLarge() != nil {
		return *m.GetCoverImage().GetExtraLarge()
	}
	if m.GetCoverImage().GetLarge() != nil {
		return *m.GetCoverImage().GetLarge()
	}
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return ""
}
func (m *BaseManga) GetBannerImageSafe() string {
	if m.GetBannerImage() != nil {
		return *m.GetBannerImage()
	}
	return m.GetCoverImageSafe()
}

func (m *BaseManga) GetAllTitles() []*string {
	titles := make([]*string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, m.Title.English)
	}
	if m.HasSynonyms() && len(m.Synonyms) > 1 {
		titles = append(titles, m.Synonyms...)
	}
	return titles
}

func (m *BaseManga) GetMainTitlesDeref() []string {
	titles := make([]string, 0)
	if m.HasRomajiTitle() {
		titles = append(titles, *m.Title.Romaji)
	}
	if m.HasEnglishTitle() {
		titles = append(titles, *m.Title.English)
	}
	return titles
}

func (m *BaseManga) HasEnglishTitle() bool {
	return m.Title.English != nil
}
func (m *BaseManga) HasRomajiTitle() bool {
	return m.Title.Romaji != nil
}
func (m *BaseManga) HasSynonyms() bool {
	return m.Synonyms != nil
}

func (m *BaseManga) GetStartYearSafe() int {
	if m.GetStartDate() != nil && m.GetStartDate().GetYear() != nil {
		return *m.GetStartDate().GetYear()
	}
	return 0
}

func (m *MangaListEntry) GetRepeatSafe() int {
	if m.Repeat == nil {
		return 0
	}
	return *m.Repeat
}
