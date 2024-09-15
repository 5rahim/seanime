package metadata

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/tvdb"
)

type (
	Provider interface {
		GetAnimeMetadataWrapper(anime *anilist.BaseAnime, anizipMedia *anizip.Media) AnimeMetadataWrapper
	}

	// AnimeMetadataWrapper is a container for anime metadata.
	// The user can request metadata to be fetched from TVDB as well, which will be stored in the cache.
	AnimeMetadataWrapper interface {
		GetEpisodeMetadata(episodeNumber int) EpisodeMetadata
		EmptyTVDBEpisodesBucket(mediaId int) error
		GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error)
		GetTVDBEpisodeByNumber(episodeNumber int) (*tvdb.Episode, bool)
	}
)

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
		TvdbId                int64  `json:"tvdbId"`
		Title                 string `json:"title"`
		Image                 string `json:"image"`
		AirDate               string `json:"airDate"`
		Length                int    `json:"length"`
		Summary               string `json:"summary"`
		Overview              string `json:"overview"`
		EpisodeNumber         int    `json:"episodeNumber"`
		SeasonNumber          int    `json:"seasonNumber"`
		AbsoluteEpisodeNumber int    `json:"absoluteEpisodeNumber"`
		AnidbEid              int    `json:"anidbEid"`
	}
)
