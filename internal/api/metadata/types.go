package metadata

import (
	"strings"
	"time"
)

const (
	AnilistPlatform Platform = "anilist"
	MalPlatform     Platform = "mal"
)

type (
	Platform string
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	AnimeMetadata struct {
		Titles       map[string]string           `json:"titles"`
		Episodes     map[string]*EpisodeMetadata `json:"episodes"`
		EpisodeCount int                         `json:"episodeCount"`
		SpecialCount int                         `json:"specialCount"`
		Mappings     *AnimeMappings              `json:"mappings"`

		currentEpisodeCount int `json:"-"`
	}

	AnimeMappings struct {
		AnimeplanetId string `json:"animeplanetId,omitempty"`
		KitsuId       int    `json:"kitsuId,omitempty"`
		MalId         int    `json:"malId,omitempty"`
		Type          string `json:"type,omitempty"`
		AnilistId     int    `json:"anilistId,omitempty"`
		AnisearchId   int    `json:"anisearchId,omitempty"`
		AnidbId       int    `json:"anidbId,omitempty"`
		NotifymoeId   string `json:"notifymoeId,omitempty"`
		LivechartId   int    `json:"livechartId,omitempty"`
		ThetvdbId     int    `json:"thetvdbId,omitempty"`
		ImdbId        string `json:"imdbId,omitempty"`
		ThemoviedbId  string `json:"themoviedbId,omitempty"`
	}

	EpisodeMetadata struct {
		AnidbId               int    `json:"anidbId"`
		TvdbId                int    `json:"tvdbId"`
		Title                 string `json:"title"`
		Image                 string `json:"image"`
		AirDate               string `json:"airDate"`
		Length                int    `json:"length"`
		Summary               string `json:"summary"`
		Overview              string `json:"overview"`
		EpisodeNumber         int    `json:"episodeNumber"`
		Episode               string `json:"episode"`
		SeasonNumber          int    `json:"seasonNumber"`
		AbsoluteEpisodeNumber int    `json:"absoluteEpisodeNumber"`
		AnidbEid              int    `json:"anidbEid"`
		HasImage              bool   `json:"hasImage"` // Indicates if the episode has a real image
	}
)

func (m *AnimeMetadata) GetTitle() string {
	if m == nil {
		return ""
	}
	if len(m.Titles["en"]) > 0 {
		return m.Titles["en"]
	}
	return m.Titles["ro"]
}

func (m *AnimeMetadata) GetMappings() *AnimeMappings {
	if m == nil || m.Mappings == nil {
		return &AnimeMappings{}
	}
	return m.Mappings
}

func (m *AnimeMetadata) FindEpisode(ep string) (*EpisodeMetadata, bool) {
	if m == nil {
		return nil, false
	}
	if m.Episodes == nil {
		return nil, false
	}
	episode, found := m.Episodes[ep]
	if !found {
		return nil, false
	}

	return episode, true
}

func (m *AnimeMetadata) GetMainEpisodeCount() int {
	if m == nil {
		return 0
	}
	return m.EpisodeCount
}

func (m *AnimeMetadata) GetCurrentEpisodeCount() int {
	if m == nil {
		return 0
	}
	if m.currentEpisodeCount > 0 {
		return m.currentEpisodeCount
	}
	count := 0
	for _, ep := range m.Episodes {
		firstChar := ep.Episode[0]
		if firstChar >= '0' && firstChar <= '9' {
			// Check if aired
			if ep.AirDate != "" {
				date, err := time.Parse("2006-01-02", ep.AirDate)
				if err == nil {
					if date.Before(time.Now()) || date.Equal(time.Now()) {
						count++
					}
				}
			}
		}
	}
	m.currentEpisodeCount = count
	return count
}

// GetOffset returns the offset of the first episode relative to the absolute episode number.
// e.g, if the first episode's absolute number is 13, then the offset is 12.
func (m *AnimeMetadata) GetOffset() int {
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

func (e *EpisodeMetadata) GetTitle() string {
	if e == nil {
		return ""
	}
	return strings.ReplaceAll(e.Title, "`", "'")
}
