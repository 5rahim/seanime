package onlinestream

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"sort"
	"strconv"
	"sync"
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
		anizipCache                  *anizip.Cache
		anilistClientWrapper         anilist.ClientWrapperInterface
		anilistBaseMediaCache        *anilist.BaseMediaCache
	}
)

type (
	Episode struct {
		Number      int    `json:"number"`
		Title       string `json:"title,omitempty"`
		Image       string `json:"image,omitempty"`
		Description string `json:"description,omitempty"`
	}

	EpisodeSource struct {
		Number    int         `json:"number"`
		Sources   []*Source   `json:"sources"`
		Subtitles []*Subtitle `json:"subtitles,omitempty"`
	}

	Source struct {
		URL     string `json:"url"`
		Quality string `json:"quality"`
	}

	Subtitle struct {
		URL      string `json:"url"`
		Language string `json:"language"`
	}
)

type (
	NewOnlineStreamOptions struct {
		Logger               *zerolog.Logger
		FileCacher           *filecache.Cacher
		AnizipCache          *anizip.Cache
		AnilistClientWrapper anilist.ClientWrapperInterface
	}
)

func New(opts *NewOnlineStreamOptions) *OnlineStream {
	return &OnlineStream{
		logger:                       opts.Logger,
		anizipCache:                  opts.AnizipCache,
		fileCacher:                   opts.FileCacher,
		gogo:                         onlinestream_providers.NewGogoanime(opts.Logger),
		zoro:                         onlinestream_providers.NewZoro(opts.Logger),
		fcEpisodeBucket:              filecache.NewBucket("onlinestream-episodes", 24*time.Hour*7),            // Cache episodes for 7 days
		fcProviderEpisodesInfoBucket: filecache.NewBucket("onlinestream-provider-episodes-info", 1*time.Hour), // Cache provider episodes for 1 hour
		anilistBaseMediaCache:        anilist.NewBaseMediaCache(),
		anilistClientWrapper:         opts.AnilistClientWrapper,
	}
}

func (os *OnlineStream) GetMediaEpisodes(provider string, mId int, dubbed bool) ([]*Episode, error) {

	media, err := os.anilistBaseMediaCache.GetOrSet(mId, func() (*anilist.BaseMedia, error) {
		mediaF, err := os.anilistClientWrapper.BaseMediaByID(context.Background(), &mId)
		if err != nil {
			return nil, err
		}
		media := mediaF.GetMedia()
		return media, nil
	})
	if err != nil {
		return nil, err
	}

	anizipMedia, err := anizip.FetchAniZipMediaC("anilist", mId, os.anizipCache)
	foundAnizipMedia := err == nil && anizipMedia != nil

	var providerEpisodesInfo []*onlinestream_providers.ProviderEpisodeInfo

	providerEpisodesInfoKey := strconv.Itoa(mId) + "$" + provider

	if found, _ := os.fileCacher.Get(os.fcProviderEpisodesInfoBucket, providerEpisodesInfoKey, &providerEpisodesInfo); !found {
		providerEpisodesInfo, err = os.getProviderEpisodes(Provider(provider), media.GetAllTitles(), dubbed)
		if err != nil {
			os.logger.Error().Err(err).Str("provider", provider).Msg("onlinestream: failed to get provider episodes")
			return nil, err
		}
		_ = os.fileCacher.Set(os.fcProviderEpisodesInfoBucket, providerEpisodesInfoKey, providerEpisodesInfo)
	}

	if providerEpisodesInfo == nil {
		return nil, ErrNoAnimeFound
	}

	episodes := make([]*Episode, 0)

	wg := sync.WaitGroup{}

	for _, _providerEpisodeInfo := range providerEpisodesInfo {
		wg.Add(1)
		go func(providerEpisodeInfo *onlinestream_providers.ProviderEpisodeInfo) {
			defer wg.Done()
			if foundAnizipMedia {
				anizipEpisode, found := anizipMedia.Episodes[strconv.Itoa(providerEpisodeInfo.Number)]
				if found {
					img := anizipEpisode.Image
					if img == "" {
						img = media.GetCoverImageSafe()
					}
					episodes = append(episodes, &Episode{
						Number:      providerEpisodeInfo.Number,
						Title:       anizipEpisode.GetTitle(),
						Image:       img,
						Description: anizipEpisode.Summary,
					})
				}
			} else {
				episodes = append(episodes, &Episode{
					Number: providerEpisodeInfo.Number,
					Title:  providerEpisodeInfo.Title,
					Image:  media.GetCoverImageSafe(),
				})
			}
		}(_providerEpisodeInfo)
	}

	wg.Wait()

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})

	return episodes, nil
}

func (os *OnlineStream) GetEpisodeSources(provider string, mId int, number int, dubbed bool) ([]*EpisodeSource, error) {

	media, err := os.anilistBaseMediaCache.GetOrSet(mId, func() (*anilist.BaseMedia, error) {
		mediaF, err := os.anilistClientWrapper.BaseMediaByID(context.Background(), &mId)
		if err != nil {
			return nil, err
		}
		media := mediaF.GetMedia()
		return media, nil
	})
	if err != nil {
		return nil, err
	}

	res, found := os.getEpisodeContainer(Provider(provider), mId, media.GetAllTitles(), number, number, dubbed)
	if !found {
		return nil, ErrNoAnimeFound
	}

	sources := make([]*EpisodeSource, 0)

	for _, e := range res.ProviderEpisodes {
		for _, ep := range e.ExtractedEpisodes {
			if ep.Number == number {
				s := &EpisodeSource{
					Number: ep.Number,
				}
				for _, ss := range ep.ServerSources {
					for _, vs := range ss.VideoSources {
						s.Sources = append(s.Sources, &Source{
							URL:     vs.URL,
							Quality: vs.Quality,
						})
						if len(vs.Subtitles) > 0 && s.Subtitles == nil {
							s.Subtitles = make([]*Subtitle, 0, len(vs.Subtitles))
							for _, sub := range vs.Subtitles {
								s.Subtitles = append(s.Subtitles, &Subtitle{
									URL:      sub.URL,
									Language: sub.Language,
								})
							}
						}
					}
				}
				sources = append(sources, s)
			}
		}
	}

	return sources, nil
}
