package metadata

import (
	"github.com/rs/zerolog"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/tvdb"
	"seanime/internal/util/filecache"
)

type Provider interface {
	GetAnimeMetadata(anime *anilist.BaseAnime, anizipMedia *anizip.Media) AnimeMetadata
}

// AnimeMetadata is a wrapper for anime metadata.
// The user can request metadata to be fetched from TVDB as well, which will be stored in the cache.
type AnimeMetadata interface {
	GetEpisodeMetadata(episodeNumber int) EpisodeMetadata
	EmptyTVDBEpisodesBucket(mediaId int) error
	GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error)
	GetTVDBEpisodeByNumber(episodeNumber int) (*tvdb.Episode, bool)
}

type (
	ProviderImpl struct {
		logger     *zerolog.Logger
		fileCacher *filecache.Cacher
	}

	NewProviderImplOptions struct {
		Logger     *zerolog.Logger
		FileCacher *filecache.Cacher
	}
)

// NewProvider creates a new metadata provider.
func NewProvider(options *NewProviderImplOptions) Provider {
	return &ProviderImpl{
		logger:     options.Logger,
		fileCacher: options.FileCacher,
	}
}
