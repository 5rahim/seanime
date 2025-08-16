package metadata

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/animap"
	"seanime/internal/hook"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"golang.org/x/sync/singleflight"
)

type (
	ProviderImpl struct {
		logger             *zerolog.Logger
		fileCacher         *filecache.Cacher
		animeMetadataCache *result.BoundedCache[string, *AnimeMetadata]
		singleflight       *singleflight.Group
	}

	NewProviderImplOptions struct {
		Logger     *zerolog.Logger
		FileCacher *filecache.Cacher
	}
)

func GetAnimeMetadataCacheKey(platform Platform, mId int) string {
	return fmt.Sprintf("%s$%d", platform, mId)
}

// NewProvider creates a new metadata provider.
func NewProvider(options *NewProviderImplOptions) Provider {
	return &ProviderImpl{
		logger:             options.Logger,
		fileCacher:         options.FileCacher,
		animeMetadataCache: result.NewBoundedCache[string, *AnimeMetadata](100),
		singleflight:       &singleflight.Group{},
	}
}

// GetCache returns the anime metadata cache.
func (p *ProviderImpl) GetCache() *result.BoundedCache[string, *AnimeMetadata] {
	return p.animeMetadataCache
}

// GetAnimeMetadata fetches anime metadata from api.ani.zip.
func (p *ProviderImpl) GetAnimeMetadata(platform Platform, mId int) (ret *AnimeMetadata, err error) {
	cacheKey := GetAnimeMetadataCacheKey(platform, mId)
	if cached, ok := p.animeMetadataCache.Get(cacheKey); ok {
		return cached, nil
	}

	res, err, _ := p.singleflight.Do(cacheKey, func() (interface{}, error) {
		return p.fetchAnimeMetadata(platform, mId)
	})
	if err != nil {
		return nil, err
	}

	return res.(*AnimeMetadata), nil
}

func (p *ProviderImpl) fetchAnimeMetadata(platform Platform, mId int) (*AnimeMetadata, error) {
	ret := &AnimeMetadata{
		Titles:       make(map[string]string),
		Episodes:     make(map[string]*EpisodeMetadata),
		EpisodeCount: 0,
		SpecialCount: 0,
		Mappings:     &AnimeMappings{},
	}

	// Invoke AnimeMetadataRequested hook
	reqEvent := &AnimeMetadataRequestedEvent{
		MediaId:       mId,
		AnimeMetadata: ret,
	}
	err := hook.GlobalHookManager.OnAnimeMetadataRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}
	mId = reqEvent.MediaId

	// Default prevented by hook, return the metadata
	if reqEvent.DefaultPrevented {
		// Override the metadata
		ret = reqEvent.AnimeMetadata

		// Trigger the event
		event := &AnimeMetadataEvent{
			MediaId:       mId,
			AnimeMetadata: ret,
		}
		err = hook.GlobalHookManager.OnAnimeMetadata().Trigger(event)
		if err != nil {
			return nil, err
		}
		ret = event.AnimeMetadata
		mId = event.MediaId

		if ret == nil {
			return nil, errors.New("no metadata was returned")
		}
		p.animeMetadataCache.SetT(GetAnimeMetadataCacheKey(platform, mId), ret, 1*time.Hour)
		return ret, nil
	}

	m, err := animap.FetchAnimapMedia(string(platform), mId)
	if err != nil || m == nil {
		//return p.AnizipFallback(platform, mId)
		return nil, err
	}

	ret.Titles = m.Titles
	ret.EpisodeCount = 0
	ret.SpecialCount = 0
	ret.Mappings.AnimeplanetId = m.Mappings.AnimePlanetID
	ret.Mappings.KitsuId = m.Mappings.KitsuID
	ret.Mappings.MalId = m.Mappings.MalID
	ret.Mappings.Type = m.Mappings.Type
	ret.Mappings.AnilistId = m.Mappings.AnilistID
	ret.Mappings.AnisearchId = m.Mappings.AnisearchID
	ret.Mappings.AnidbId = m.Mappings.AnidbID
	ret.Mappings.NotifymoeId = m.Mappings.NotifyMoeID
	ret.Mappings.LivechartId = m.Mappings.LivechartID
	ret.Mappings.ThetvdbId = m.Mappings.TheTvdbID
	ret.Mappings.ImdbId = ""
	ret.Mappings.ThemoviedbId = m.Mappings.TheMovieDbID

	for key, ep := range m.Episodes {
		firstChar := key[0]
		if firstChar == 'S' {
			ret.SpecialCount++
		} else {
			if firstChar >= '0' && firstChar <= '9' {
				ret.EpisodeCount++
			}
		}
		em := &EpisodeMetadata{
			AnidbId:               ep.AnidbId,
			TvdbId:                ep.TvdbId,
			Title:                 ep.AnidbTitle,
			Image:                 ep.Image,
			AirDate:               ep.AirDate,
			Length:                ep.Runtime,
			Summary:               strings.ReplaceAll(ep.Overview, "`", "'"),
			Overview:              strings.ReplaceAll(ep.Overview, "`", "'"),
			EpisodeNumber:         ep.Number,
			Episode:               key,
			SeasonNumber:          ep.SeasonNumber,
			AbsoluteEpisodeNumber: ep.AbsoluteNumber,
			AnidbEid:              ep.AnidbId,
			HasImage:              ep.Image != "",
		}
		if em.Length == 0 && ep.Runtime > 0 {
			em.Length = ep.Runtime
		}
		if em.Summary == "" && ep.Overview != "" {
			em.Summary = ep.Overview
		}
		if em.Overview == "" && ep.Overview != "" {
			em.Overview = ep.Overview
		}
		if ep.TvdbTitle != "" && ep.AnidbTitle == "Episode "+ep.AnidbEpisode {
			em.Title = ep.TvdbTitle

		}
		ret.Episodes[key] = em
	}

	// Event
	event := &AnimeMetadataEvent{
		MediaId:       mId,
		AnimeMetadata: ret,
	}
	err = hook.GlobalHookManager.OnAnimeMetadata().Trigger(event)
	if err != nil {
		return nil, err
	}
	ret = event.AnimeMetadata
	mId = event.MediaId

	p.animeMetadataCache.SetT(GetAnimeMetadataCacheKey(platform, mId), ret, 1*time.Hour)

	return ret, nil
}

// GetAnimeMetadataWrapper creates a new anime wrapper.
//
//	Example:
//
//	metadataProvider.GetAnimeMetadataWrapper(media, metadata)
//	metadataProvider.GetAnimeMetadataWrapper(media, nil)
func (p *ProviderImpl) GetAnimeMetadataWrapper(media *anilist.BaseAnime, metadata *AnimeMetadata) AnimeMetadataWrapper {
	aw := &AnimeWrapperImpl{
		metadata:   mo.None[*AnimeMetadata](),
		baseAnime:  media,
		fileCacher: p.fileCacher,
		logger:     p.logger,
	}

	if metadata != nil {
		aw.metadata = mo.Some(metadata)
	}

	return aw
}
