package onlinestream

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/result"
)

type (
	OnlineStream struct {
		logger                *zerolog.Logger
		cache                 *AnimeEpisodesCache
		episodeCache          *EpisodeCache
		providerEpisodesCache *ProviderEpisodesCache
		gogo                  *onlinestream_providers.Gogoanime
		zoro                  *onlinestream_providers.Zoro
	}
)

type (
	NewOnlineStreamOptions struct {
		Logger *zerolog.Logger
	}
)

func New(opts *NewOnlineStreamOptions) *OnlineStream {
	return &OnlineStream{
		logger: opts.Logger,
		cache: &AnimeEpisodesCache{
			Cache: result.NewCache[int, *AnimeEpisodes](),
		},
		episodeCache: &EpisodeCache{
			Cache: result.NewCache[string, *Episode](),
		},
		providerEpisodesCache: &ProviderEpisodesCache{
			Cache: result.NewCache[int, []*onlinestream_providers.ProviderEpisode](),
		},
		gogo: onlinestream_providers.NewGogoanime(opts.Logger),
		zoro: onlinestream_providers.NewZoro(opts.Logger),
	}
}

func (os *OnlineStream) Start() {

}
