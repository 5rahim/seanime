package anizip

func (m *Media) GetTitle() string {
	if m == nil {
		return ""
	}
	if len(m.Titles["en"]) > 0 {
		return m.Titles["en"]
	}
	return m.Titles["ro"]
}

func (m *Media) GetMappings() *Mappings {
	if m == nil {
		return &Mappings{}
	}
	return m.Mappings
}

func (m *Media) FindEpisode(ep string) (*Episode, bool) {
	if m.Episodes == nil {
		return nil, false
	}
	episode, found := m.Episodes[ep]
	if !found {
		return nil, false
	}

	return &episode, true
}

func (m *Media) GetMainEpisodeCount() int {
	if m == nil {
		return 0
	}
	return m.EpisodeCount
}

// GetOffset returns the offset of the first episode relative to the absolute episode number.
// e.g, if the first episode's absolute number is 13, then the offset is 12.
func (m *Media) GetOffset() int {
	if m == nil {
		return 0
	}
	firstEp, found := m.FindEpisode("1")
	if !found {
		return 0
	}
	if firstEp.AbsoluteEpisodeNumber == 0 {
		return 0
	}
	return firstEp.AbsoluteEpisodeNumber - 1
}

func (e *Episode) GetTitle() string {
	eng, ok := e.Title["en"]
	if ok {
		return eng
	}
	rom, ok := e.Title["x-jat"]
	if ok {
		return rom
	}
	return ""
}
