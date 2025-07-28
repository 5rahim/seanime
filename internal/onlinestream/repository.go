package onlinestream

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type (
	Repository struct {
		logger                *zerolog.Logger
		providerExtensionBank *extension.UnifiedBank
		fileCacher            *filecache.Cacher
		metadataProvider      metadata.Provider
		platform              platform.Platform
		anilistBaseAnimeCache *anilist.BaseAnimeCache
		db                    *db.Database
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
		IsFiller    bool   `json:"isFiller,omitempty"`
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
	NewRepositoryOptions struct {
		Logger           *zerolog.Logger
		FileCacher       *filecache.Cacher
		MetadataProvider metadata.Provider
		Platform         platform.Platform
		Database         *db.Database
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	return &Repository{
		logger:                opts.Logger,
		metadataProvider:      opts.MetadataProvider,
		fileCacher:            opts.FileCacher,
		providerExtensionBank: extension.NewUnifiedBank(),
		anilistBaseAnimeCache: anilist.NewBaseAnimeCache(),
		platform:              opts.Platform,
		db:                    opts.Database,
	}
}

func (r *Repository) InitExtensionBank(bank *extension.UnifiedBank) {
	r.providerExtensionBank = bank

	r.logger.Debug().Msg("onlinestream: Initialized provider extension bank")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getFcEpisodeDataBucket returns a episode data bucket for the provider and mediaId.
// "Episode data" refers to the episodeData struct
//
//	e.g., onlinestream_zoro_episode-data_123
func (r *Repository) getFcEpisodeDataBucket(provider string, mediaId int) filecache.Bucket {
	return filecache.NewBucket("onlinestream_"+provider+"_episode-data_"+strconv.Itoa(mediaId), time.Hour*24*2)
}

// getFcEpisodeListBucket returns a episode data bucket for the provider and mediaId.
// "Episode list" refers to a slice of onlinestream_providers.EpisodeDetails
//
//	e.g., onlinestream_zoro_episode-list_123
func (r *Repository) getFcEpisodeListBucket(provider string, mediaId int) filecache.Bucket {
	return filecache.NewBucket("onlinestream_"+provider+"_episode-data_"+strconv.Itoa(mediaId), time.Hour*24*1)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) getMedia(ctx context.Context, mId int) (*anilist.BaseAnime, error) {
	media, err := r.anilistBaseAnimeCache.GetOrSet(mId, func() (*anilist.BaseAnime, error) {
		media, err := r.platform.GetAnime(ctx, mId)
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

func (r *Repository) GetMedia(ctx context.Context, mId int) (*anilist.BaseAnime, error) {
	return r.getMedia(ctx, mId)
}

func (r *Repository) EmptyCache(mediaId int) error {
	_ = r.fileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, "onlinestream_") && strings.Contains(filename, strconv.Itoa(mediaId))
	})
	return nil
}

func (r *Repository) GetMediaEpisodes(provider string, media *anilist.BaseAnime, dubbed bool) ([]*Episode, error) {
	episodes := make([]*Episode, 0)

	if provider == "" {
		return episodes, nil
	}

	// +---------------------+
	// |       Anizip        |
	// +---------------------+

	//animeMetadata, err := r.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, mId)
	//	//foundAnimeMetadata := err == nil && animeMetadata != nil
	//aw := r.metadataProvider.GetAnimeMetadataWrapper(media, animeMetadata)

	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:    nil,
		Media:            media,
		MetadataProvider: r.metadataProvider,
		Logger:           r.logger,
	})
	foundEpisodeCollection := err == nil && episodeCollection != nil

	// +---------------------+
	// |    Episode list     |
	// +---------------------+

	// Fetch the episode list from the provider
	// "from" and "to" are set to 0 in order not to fetch episode servers
	ec, err := r.getEpisodeContainer(provider, media, 0, 0, dubbed, media.GetStartYearSafe())
	if err != nil {
		return nil, err
	}

	for _, episodeDetails := range ec.ProviderEpisodeList {

		// If the title contains "[{", it means it's an episode part (e.g. "Episode 6 [{6.5}]", the episode number should be 6)
		if strings.Contains(episodeDetails.Title, "[{") {
			ep := strings.Split(episodeDetails.Title, "[{")[1]
			ep = strings.Split(ep, "}]")[0]
			episodes = append(episodes, &Episode{
				Number:      episodeDetails.Number,
				Title:       fmt.Sprintf("Episode %s", ep),
				Image:       media.GetBannerImageSafe(),
				Description: "",
				IsFiller:    false,
			})

		} else {

			if foundEpisodeCollection {
				episode, found := episodeCollection.FindEpisodeByNumber(episodeDetails.Number)
				if found {
					episodes = append(episodes, &Episode{
						Number:      episodeDetails.Number,
						Title:       episode.EpisodeTitle,
						Image:       episode.EpisodeMetadata.Image,
						Description: episode.EpisodeMetadata.Summary,
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
	}

	episodes = lo.Filter(episodes, func(item *Episode, index int) bool {
		return item != nil
	})

	return episodes, nil
}

func (r *Repository) GetEpisodeSources(ctx context.Context, provider string, mId int, number int, dubbed bool, year int) (*EpisodeSource, error) {

	// +---------------------+
	// |        Media        |
	// +---------------------+

	media, err := r.getMedia(ctx, mId)
	if err != nil {
		return nil, err
	}

	// +---------------------+
	// |   Episode servers   |
	// +---------------------+

	ec, err := r.getEpisodeContainer(provider, media, number, number, dubbed, year)
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
			for _, es := range ep.Servers {

				for _, vs := range es.VideoSources {
					s.VideoSources = append(s.VideoSources, &VideoSource{
						Server:  es.Server,
						Headers: es.Headers,
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
