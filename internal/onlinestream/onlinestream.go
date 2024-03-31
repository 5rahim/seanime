package onlinestream

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"time"
)

type (
	OnlineStream struct {
		logger                      *zerolog.Logger
		gogo                        *onlinestream_providers.Gogoanime
		zoro                        *onlinestream_providers.Zoro
		fileCacher                  *filecache.Cacher
		fcEpisodeDataBucket         filecache.Bucket
		fcProviderEpisodeListBucket filecache.Bucket
		anizipCache                 *anizip.Cache
		anilistClientWrapper        anilist.ClientWrapperInterface
		anilistBaseMediaCache       *anilist.BaseMediaCache
	}
)

var (
	ErrNoVideoSourceFound = errors.New("no video source found")
)

type (
	Episode struct {
		Number      int    `json:"number,omitempty"`
		Title       string `json:"title,omitempty"`
		Image       string `json:"image,omitempty"`
		Description string `json:"description,omitempty"`
	}

	EpisodeSource struct {
		Number       int            `json:"number"`
		VideoSources []*VideoSource `json:"videoSources"`
		Subtitles    []*Subtitle    `json:"subtitles,omitempty"`
	}

	VideoSource struct {
		Server  string            `json:"server"`
		Headers map[string]string `json:"headers,omitempty"`
		URL     string            `json:"url"`
		Quality string            `json:"quality"`
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
		logger:                      opts.Logger,
		anizipCache:                 opts.AnizipCache,
		fileCacher:                  opts.FileCacher,
		gogo:                        onlinestream_providers.NewGogoanime(opts.Logger),
		zoro:                        onlinestream_providers.NewZoro(opts.Logger),
		fcEpisodeDataBucket:         filecache.NewBucket("onlinestream-episode-data", 24*time.Hour*7),       // Cache episodes for 7 days
		fcProviderEpisodeListBucket: filecache.NewBucket("onlinestream-provider-episode-list", 1*time.Hour), // Cache provider episodes for 1 hour
		anilistBaseMediaCache:       anilist.NewBaseMediaCache(),
		anilistClientWrapper:        opts.AnilistClientWrapper,
	}
}

func (os *OnlineStream) getMedia(mId int) (*anilist.BaseMedia, error) {
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
	return media, nil
}

func (os *OnlineStream) GetMedia(mId int) (*anilist.BaseMedia, error) {
	return os.getMedia(mId)
}

func (os *OnlineStream) EmptyCache() error {
	_ = os.fileCacher.Empty(os.fcEpisodeDataBucket)
	_ = os.fileCacher.Empty(os.fcProviderEpisodeListBucket)
	return nil
}

func (os *OnlineStream) GetMediaEpisodes(provider string, media *anilist.BaseMedia, dubbed bool) ([]*Episode, error) {

	mId := media.GetID()

	// +---------------------+
	// |       Anizip        |
	// +---------------------+

	anizipMedia, err := anizip.FetchAniZipMediaC("anilist", mId, os.anizipCache)
	foundAnizipMedia := err == nil && anizipMedia != nil

	// +---------------------+
	// |    Episode list     |
	// +---------------------+

	// Only fetch the episode list from the provider without episode servers
	ec, err := os.getEpisodeContainer(onlinestream_providers.Provider(provider), mId, media.GetAllTitles(), 0, 0, dubbed)
	if err != nil {
		return nil, err
	}

	episodes := make([]*Episode, 0)

	for _, episodeDetails := range ec.ProviderEpisodeList {
		if foundAnizipMedia {
			anizipEpisode, found := anizipMedia.Episodes[strconv.Itoa(episodeDetails.Number)]
			if found {
				img := anizipEpisode.Image
				if img == "" {
					img = media.GetCoverImageSafe()
				}
				episodes = append(episodes, &Episode{
					Number:      episodeDetails.Number,
					Title:       anizipEpisode.GetTitle(),
					Image:       img,
					Description: anizipEpisode.Summary,
				})
			} else {
				episodes = append(episodes, &Episode{
					Number: episodeDetails.Number,
					Title:  episodeDetails.Title,
					Image:  media.GetCoverImageSafe(),
				})
			}
		} else {
			episodes = append(episodes, &Episode{
				Number: episodeDetails.Number,
				Title:  episodeDetails.Title,
				Image:  media.GetCoverImageSafe(),
			})
		}
	}

	episodes = lo.Filter(episodes, func(item *Episode, index int) bool {
		return item != nil
	})

	return episodes, nil
}

func (os *OnlineStream) GetEpisodeSources(provider string, mId int, number int, dubbed bool) (*EpisodeSource, error) {

	// +---------------------+
	// |        Media        |
	// +---------------------+

	media, err := os.getMedia(mId)
	if err != nil {
		return nil, err
	}

	// +---------------------+
	// |   Episode servers   |
	// +---------------------+

	ec, err := os.getEpisodeContainer(onlinestream_providers.Provider(provider), mId, media.GetAllTitles(), number, number, dubbed)
	if err != nil {
		return nil, err
	}

	var sources *EpisodeSource
	for _, ep := range ec.Episodes {
		if ep.Number == number {
			s := &EpisodeSource{
				Number:       ep.Number,
				VideoSources: make([]*VideoSource, 0),
			}
			for _, ss := range ep.Servers {

				for _, vs := range ss.VideoSources {
					s.VideoSources = append(s.VideoSources, &VideoSource{
						Server:  string(ss.Server),
						Headers: ss.Headers,
						URL:     vs.URL,
						Quality: vs.Quality,
					})
					// Add subtitles if available
					// Subtitles are stored in each video source, but they are the same, so only add them once.
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
			sources = s
			break
		}
	}

	if sources == nil {
		return nil, ErrNoVideoSourceFound
	}

	return sources, nil
}
