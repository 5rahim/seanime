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

func (m *Media) GetEpisode(id string) (*Episode, bool) {
	if m.Episodes == nil {
		return nil, false
	}
	episode, found := m.Episodes[id]
	if !found {
		return nil, false
	}

	return &episode, true
}
