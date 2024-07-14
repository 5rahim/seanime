package onlinestream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"strings"
	"time"
)

type (
	OnlineStream struct {
		logger                *zerolog.Logger
		gogo                  *onlinestream_providers.Gogoanime
		zoro                  *onlinestream_providers.Zoro
		fileCacher            *filecache.Cacher
		anizipCache           *anizip.Cache
		platform              platform.Platform
		anilistBaseAnimeCache *anilist.BaseAnimeCache
	}
)

var (
	ErrNoVideoSourceFound = errors.New("no video source found")
)

type (
	Episode struct {
		Number      int    `json:"number"`
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

	EpisodeListResponse struct {
		Episodes []*Episode         `json:"episodes"`
		Media    *anilist.BaseAnime `json:"media"`
	}

	Subtitle struct {
		URL      string `json:"url"`
		Language string `json:"language"`
	}
)

type (
	NewOnlineStreamOptions struct {
		Logger      *zerolog.Logger
		FileCacher  *filecache.Cacher
		AnizipCache *anizip.Cache
		Platform    platform.Platform
	}
)

func New(opts *NewOnlineStreamOptions) *OnlineStream {
	return &OnlineStream{
		logger:                opts.Logger,
		anizipCache:           opts.AnizipCache,
		fileCacher:            opts.FileCacher,
		gogo:                  onlinestream_providers.NewGogoanime(opts.Logger),
		zoro:                  onlinestream_providers.NewZoro(opts.Logger),
		anilistBaseAnimeCache: anilist.NewBaseAnimeCache(),
		platform:              opts.Platform,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getFcEpisodeDataBucket returns a episode data bucket for the provider and mediaId.
// "Episode data" refers to the episodeData struct
//
//	e.g., onlinestream_zoro_episode-data_123
func (os *OnlineStream) getFcEpisodeDataBucket(provider onlinestream_providers.Provider, mediaId int) filecache.Bucket {
	return filecache.NewBucket("onlinestream_"+string(provider)+"_episode-data_"+strconv.Itoa(mediaId), time.Hour*24*7)
}

// getFcEpisodeListBucket returns a episode data bucket for the provider and mediaId.
// "Episode list" refers to a slice of onlinestream_providers.EpisodeDetails
//
//	e.g., onlinestream_zoro_episode-list_123
func (os *OnlineStream) getFcEpisodeListBucket(provider onlinestream_providers.Provider, mediaId int) filecache.Bucket {
	return filecache.NewBucket("onlinestream_"+string(provider)+"_episode-data_"+strconv.Itoa(mediaId), time.Hour*24*7)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (os *OnlineStream) getMedia(mId int) (*anilist.BaseAnime, error) {
	media, err := os.anilistBaseAnimeCache.GetOrSet(mId, func() (*anilist.BaseAnime, error) {
		media, err := os.platform.GetAnime(mId)
		if err != nil {
			return nil, err
		}
		return media, nil
	})
	if err != nil {
		return nil, err
	}
	return media, nil
}

func (os *OnlineStream) GetMedia(mId int) (*anilist.BaseAnime, error) {
	return os.getMedia(mId)
}

func (os *OnlineStream) EmptyCache(mediaId int) error {
	_ = os.fileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, "onlinestream_") && strings.Contains(filename, strconv.Itoa(mediaId))
	})
	return nil
}

func (os *OnlineStream) GetMediaEpisodes(provider string, media *anilist.BaseAnime, dubbed bool) ([]*Episode, error) {

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
