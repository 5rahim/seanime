package metadata

import (
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/tvdb"
	"seanime/internal/util/filecache"
)

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

// GetAnimeMetadataWrapper creates a new anime wrapper.
//
//	Example:
//
//	metadataProvider.GetAnimeMetadataWrapper(media, anizipMedia)
//	metadataProvider.GetAnimeMetadataWrapper(media, nil)
func (p *ProviderImpl) GetAnimeMetadataWrapper(media *anilist.BaseAnime, anizipMedia *anizip.Media) AnimeMetadataWrapper {
	aw := &AnimeWrapperImpl{
		anizipMedia:  mo.None[*anizip.Media](),
		baseAnime:    media,
		fileCacher:   p.fileCacher,
		logger:       p.logger,
		tvdbEpisodes: make([]*tvdb.Episode, 0),
	}

	if anizipMedia != nil {
		aw.anizipMedia = mo.Some(anizipMedia)
	}

	episodes, err := aw.GetTVDBEpisodes(false)
	if err == nil {
		aw.tvdbEpisodes = episodes
	}

	return aw
}
