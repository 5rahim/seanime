package metadata

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/tvdb"
	"seanime/internal/util/result"
	"strings"
	"time"
)

const (
	AnilistPlatform Platform = "anilist"
	MalPlatform     Platform = "mal"
)

type (
	Platform string

	Provider interface {
		// GetAnimeMetadata fetches anime metadata for the given platform from a source.
		// In this case, the source is api.ani.zip.
		GetAnimeMetadata(platform Platform, mId int) (*AnimeMetadata, error)
		GetCache() *result.Cache[string, *AnimeMetadata]
		// GetAnimeMetadataWrapper creates a wrapper for anime metadata.
		GetAnimeMetadataWrapper(anime *anilist.BaseAnime, metadata *AnimeMetadata) AnimeMetadataWrapper
	}

	// AnimeMetadataWrapper is a container for anime metadata.
	// This wrapper is used to get a more complete metadata object by getting data from multiple sources in the Provider.
	// The user can request metadata to be fetched from TVDB as well, which will be stored in the cache.
	AnimeMetadataWrapper interface {
		// GetEpisodeMetadata combines metadata from multiple sources to create a single EpisodeMetadata object.
		GetEpisodeMetadata(episodeNumber int) EpisodeMetadata

		EmptyTVDBEpisodesBucket(mediaId int) error
		GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error)
		GetTVDBEpisodeByNumber(episodeNumber int) (*tvdb.Episode, bool)
	}
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	AnimeMetadata struct {
		Titles       map[string]string           `json:"titles"`
		Episodes     map[string]*EpisodeMetadata `json:"episodes"`
		EpisodeCount int                         `json:"episodeCount"`
		SpecialCount int                         `json:"specialCount"`
		Mappings     *AnimeMappings              `json:"mappings"`
	}

	AnimeMappings struct {
		AnimeplanetId string `json:"animeplanetId"`
		KitsuId       int    `json:"kitsuId"`
		MalId         int    `json:"malId"`
		Type          string `json:"type"`
		AnilistId     int    `json:"anilistId"`
		AnisearchId   int    `json:"anisearchId"`
		AnidbId       int    `json:"anidbId"`
		NotifymoeId   string `json:"notifymoeId"`
		LivechartId   int    `json:"livechartId"`
		ThetvdbId     int    `json:"thetvdbId"`
		ImdbId        string `json:"imdbId"`
		ThemoviedbId  string `json:"themoviedbId"`
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
	if m == nil {
		return &AnimeMappings{}
	}
	return m.Mappings
}

func (m *AnimeMetadata) FindEpisode(ep string) (*EpisodeMetadata, bool) {
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
