package onlinestream

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"time"
)

type (
	OnlineStream struct {
		logger                       *zerolog.Logger
		gogo                         *onlinestream_providers.Gogoanime
		zoro                         *onlinestream_providers.Zoro
		fileCacher                   *filecache.Cacher
		fcEpisodeBucket              filecache.Bucket
		fcProviderEpisodesInfoBucket filecache.Bucket
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
		logger:                       opts.Logger,
		gogo:                         onlinestream_providers.NewGogoanime(opts.Logger),
		zoro:                         onlinestream_providers.NewZoro(opts.Logger),
		fileCacher:                   opts.FileCacher,
		fcEpisodeBucket:              filecache.NewBucket("onlinestream-episodes", 24*time.Hour*7),            // Cache episodes for 7 days
		fcProviderEpisodesInfoBucket: filecache.NewBucket("onlinestream-provider-episodes-info", 1*time.Hour), // Cache provider episodes for 1 hour
	}
}

func (os *OnlineStream) Start() {

}
