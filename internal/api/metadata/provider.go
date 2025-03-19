package metadata

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/tvdb"
	"seanime/internal/hook"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	ProviderImpl struct {
		logger             *zerolog.Logger
		fileCacher         *filecache.Cacher
		animeMetadataCache *result.Cache[string, *AnimeMetadata]
		anizipCache        *anizip.Cache
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
		animeMetadataCache: result.NewCache[string, *AnimeMetadata](),
		anizipCache:        anizip.NewCache(),
	}
}

// GetCache returns the anime metadata cache.
func (p *ProviderImpl) GetCache() *result.Cache[string, *AnimeMetadata] {
	return p.animeMetadataCache
}

// GetAnimeMetadata fetches anime metadata from api.ani.zip.
func (p *ProviderImpl) GetAnimeMetadata(platform Platform, mId int) (ret *AnimeMetadata, err error) {
	ret, ok := p.animeMetadataCache.Get(GetAnimeMetadataCacheKey(platform, mId))
	if ok {
		return ret, nil
	}

	ret = &AnimeMetadata{
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
	err = hook.GlobalHookManager.OnAnimeMetadataRequested().Trigger(reqEvent)
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
		err = hook.GlobalHookManager.OnAnimeMetadataEvent().Trigger(event)
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

	anizipMedia, err := anizip.FetchAniZipMediaC(string(platform), mId, p.anizipCache)
	if err != nil || anizipMedia == nil {
		return nil, err
	}

	ret.Titles = anizipMedia.Titles
	ret.EpisodeCount = anizipMedia.EpisodeCount
	ret.SpecialCount = anizipMedia.SpecialCount
	ret.Mappings.AnimeplanetId = anizipMedia.Mappings.AnimeplanetID
	ret.Mappings.KitsuId = anizipMedia.Mappings.KitsuID
	ret.Mappings.MalId = anizipMedia.Mappings.MalID
	ret.Mappings.Type = anizipMedia.Mappings.Type
	ret.Mappings.AnilistId = anizipMedia.Mappings.AnilistID
	ret.Mappings.AnisearchId = anizipMedia.Mappings.AnisearchID
	ret.Mappings.AnidbId = anizipMedia.Mappings.AnidbID
	ret.Mappings.NotifymoeId = anizipMedia.Mappings.NotifymoeID
	ret.Mappings.LivechartId = anizipMedia.Mappings.LivechartID
	ret.Mappings.ThetvdbId = anizipMedia.Mappings.ThetvdbID
	ret.Mappings.ImdbId = anizipMedia.Mappings.ImdbID
	ret.Mappings.ThemoviedbId = anizipMedia.Mappings.ThemoviedbID

	for key, anizipEp := range anizipMedia.Episodes {
		em := &EpisodeMetadata{
			AnidbId:               anizipEp.AnidbEid,
			TvdbId:                anizipEp.TvdbEid,
			Title:                 anizipEp.GetTitle(),
			Image:                 anizipEp.Image,
			AirDate:               anizipEp.AirDate,
			Length:                anizipEp.Runtime,
			Summary:               strings.ReplaceAll(anizipEp.Summary, "`", "'"),
			Overview:              strings.ReplaceAll(anizipEp.Overview, "`", "'"),
			EpisodeNumber:         anizipEp.EpisodeNumber,
			Episode:               anizipEp.Episode,
			SeasonNumber:          anizipEp.SeasonNumber,
			AbsoluteEpisodeNumber: anizipEp.AbsoluteEpisodeNumber,
			AnidbEid:              anizipEp.AnidbEid,
		}
		if em.Length == 0 && anizipEp.Length > 0 {
			em.Length = anizipEp.Length
		}
		if em.Summary == "" && anizipEp.Overview != "" {
			em.Summary = anizipEp.Overview
		}
		if em.Overview == "" && anizipEp.Summary != "" {
			em.Overview = anizipEp.Summary
		}
		ret.Episodes[key] = em
	}

	// Event
	event := &AnimeMetadataEvent{
		MediaId:       mId,
		AnimeMetadata: ret,
	}
	err = hook.GlobalHookManager.OnAnimeMetadataEvent().Trigger(event)
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
		metadata:     mo.None[*AnimeMetadata](),
		baseAnime:    media,
		fileCacher:   p.fileCacher,
		logger:       p.logger,
		tvdbEpisodes: make([]*tvdb.Episode, 0),
	}

	if metadata != nil {
		aw.metadata = mo.Some(metadata)
	}

	episodes, err := aw.GetTVDBEpisodes(false)
	if err == nil {
		aw.tvdbEpisodes = episodes
	}

	return aw
}
