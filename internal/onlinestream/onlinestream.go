package onlinestream

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"github.com/seanime-app/seanime/internal/util/result"
	"time"
)

type (
	OnlineStream struct {
		logger                *zerolog.Logger
		episodeCache          *EpisodeCache
		providerEpisodesCache *ProviderEpisodesCache
		gogo                  *onlinestream_providers.Gogoanime
		zoro                  *onlinestream_providers.Zoro
		fileCacher            *filecache.Cacher
		fcEpisodeBucket       filecache.Bucket
	}
)

type (
	NewOnlineStreamOptions struct {
		Logger     *zerolog.Logger
		FileCacher *filecache.Cacher
	}
)

func New(opts *NewOnlineStreamOptions) *OnlineStream {
	return &OnlineStream{
		logger: opts.Logger,
		episodeCache: &EpisodeCache{
			Cache: result.NewCache[string, *Episode](),
		},
		providerEpisodesCache: &ProviderEpisodesCache{
			Cache: result.NewCache[int, []*onlinestream_providers.ProviderEpisode](),
		},
		gogo:            onlinestream_providers.NewGogoanime(opts.Logger),
		zoro:            onlinestream_providers.NewZoro(opts.Logger),
		fileCacher:      opts.FileCacher,
		fcEpisodeBucket: filecache.NewBucket("onlinestream-episodes", 24*time.Hour*7), // Cache episodes for 7 days
	}
}

func (os *OnlineStream) Start() {

}
