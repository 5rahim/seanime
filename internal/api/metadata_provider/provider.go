package metadata_provider

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/animap"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/customsource"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"golang.org/x/sync/singleflight"
)

type (
	ProviderImpl struct {
		logger              *zerolog.Logger
		fileCacher          *filecache.Cacher
		animeMetadataCache  *result.BoundedCache[string, *metadata.AnimeMetadata]
		singleflight        *singleflight.Group
		extensionBankRef    *util.Ref[*extension.UnifiedBank]
		customSourceManager *customsource.Manager
		db                  *db.Database
		useFallbackProvider atomic.Bool
	}

	NewProviderImplOptions struct {
		Logger           *zerolog.Logger
		FileCacher       *filecache.Cacher
		Database         *db.Database
		ExtensionBankRef *util.Ref[*extension.UnifiedBank]
	}

	Provider interface {
		// GetAnimeMetadata fetches anime metadata for the given platform from a source.
		// In this case, the source is api.ani.zip.
		GetAnimeMetadata(platform metadata.Platform, mId int) (*metadata.AnimeMetadata, error)
		// GetAnimeMetadataWrapper creates a wrapper for anime metadata.
		GetAnimeMetadataWrapper(anime *anilist.BaseAnime, metadata *metadata.AnimeMetadata) AnimeMetadataWrapper
		GetCache() *result.BoundedCache[string, *metadata.AnimeMetadata]
		SetUseFallbackProvider(bool)
		ClearCache()
		Close()
	}

	// AnimeMetadataWrapper is a container for anime metadata.
	// This wrapper is used to get a more complete metadata object by getting data from multiple sources in the Provider.
	// In previous versions: The user can request metadata to be fetched from TVDB as well, which will be stored in the cache.
	// Now: It sets default values for missing metadata based on the media.
	AnimeMetadataWrapper interface {
		// GetEpisodeMetadata combines metadata from multiple sources to create a single EpisodeMetadata object.
		GetEpisodeMetadata(episode string) metadata.EpisodeMetadata
	}
)

func GetAnimeMetadataCacheKey(platform metadata.Platform, mId int) string {
	return fmt.Sprintf("%s$%d", platform, mId)
}

// NewProvider creates a new metadata provider.
func NewProvider(options *NewProviderImplOptions) Provider {
	ret := &ProviderImpl{
		logger:              options.Logger,
		fileCacher:          options.FileCacher,
		animeMetadataCache:  result.NewBoundedCache[string, *metadata.AnimeMetadata](100),
		singleflight:        &singleflight.Group{},
		db:                  options.Database,
		extensionBankRef:    options.ExtensionBankRef,
		customSourceManager: customsource.NewManager(options.ExtensionBankRef, options.Database, options.Logger),
	}

	return ret
}

func (p *ProviderImpl) Close() {
	p.customSourceManager.Close()
	go p.animeMetadataCache.Clear()
}

func (p *ProviderImpl) ClearCache() {
	p.animeMetadataCache.Clear()
}

// GetCache returns the anime metadata cache.
func (p *ProviderImpl) GetCache() *result.BoundedCache[string, *metadata.AnimeMetadata] {
	return p.animeMetadataCache
}

func (p *ProviderImpl) SetUseFallbackProvider(useFallback bool) {
	if useFallback != p.useFallbackProvider.Load() {
		go p.animeMetadataCache.Clear()
	}
	p.useFallbackProvider.Store(useFallback)
}

func (p *ProviderImpl) GetAnimeMetadata(platform metadata.Platform, mId int) (ret *metadata.AnimeMetadata, err error) {
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

	return res.(*metadata.AnimeMetadata), nil
}

func (p *ProviderImpl) fetchAnimeMetadata(platform metadata.Platform, mId int) (*metadata.AnimeMetadata, error) {
	ret := &metadata.AnimeMetadata{
		Titles:       make(map[string]string),
		Episodes:     make(map[string]*metadata.EpisodeMetadata),
		EpisodeCount: 0,
		SpecialCount: 0,
		Mappings:     &metadata.AnimeMappings{},
	}

	// Invoke AnimeMetadataRequested hook
	reqEvent := &metadata.AnimeMetadataRequestedEvent{
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
		event := &metadata.AnimeMetadataEvent{
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

	if customSource, localId, isCustom, hasExtension := p.customSourceManager.GetProviderFromId(mId); isCustom {
		if !hasExtension {
			return nil, errors.New("custom source does not exist or identifier has changed")
		}
		m, err := customSource.GetProvider().GetAnimeMetadata(context.Background(), localId)
		if err != nil {
			return nil, err
		}
		ret = m
	} else if p.useFallbackProvider.Load() {
		return p.AnizipFallback(platform, mId)
	} else {
		p.logger.Debug().Msgf("animap: Fetching metadata for %d", mId)

		m, err := animap.FetchAnimapMedia(string(platform), mId)
		if err != nil || m == nil {
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
			em := &metadata.EpisodeMetadata{
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
	}

	// Event
	event := &metadata.AnimeMetadataEvent{
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
func (p *ProviderImpl) GetAnimeMetadataWrapper(media *anilist.BaseAnime, m *metadata.AnimeMetadata) AnimeMetadataWrapper {
	aw := &AnimeWrapperImpl{
		metadata:   mo.None[*metadata.AnimeMetadata](),
		baseAnime:  media,
		fileCacher: p.fileCacher,
		logger:     p.logger,
		db:         p.db,
	}
	aw.provider = p
	if m != nil {
		aw.metadata = mo.Some(m)
	}

	if aw.db != nil {
		parent, err := aw.db.GetMediaMetadataParent(aw.baseAnime.GetID())
		if err == nil {
			aw.parentEntry, err = p.GetAnimeMetadata(metadata.AnilistPlatform, parent.ParentId)
			if err == nil {
				p.logger.Debug().Msgf("metadata provoder: Media %d is child of %d", aw.baseAnime.GetID(), parent.ParentId)
				aw.parentSpecialOffset = parent.SpecialOffset
				aw.metadata = mo.Some(aw.parentEntry)
			}
		}
	}

	return aw
}

func (p *ProviderImpl) AnizipFallback(platform metadata.Platform, mId int) (ret *metadata.AnimeMetadata, err error) {
	ret, ok := p.animeMetadataCache.Get(GetAnimeMetadataCacheKey(platform, mId))
	if ok {
		return ret, nil
	}

	ret = &metadata.AnimeMetadata{
		Titles:       make(map[string]string),
		Episodes:     make(map[string]*metadata.EpisodeMetadata),
		EpisodeCount: 0,
		SpecialCount: 0,
		Mappings:     &metadata.AnimeMappings{},
	}

	// Invoke AnimeMetadataRequested hook
	reqEvent := &metadata.AnimeMetadataRequestedEvent{
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
		event := &metadata.AnimeMetadataEvent{
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

	p.logger.Debug().Msgf("anizip: Fetching metadata for %d", mId)

	anizipMedia, err := anizip.FetchAniZipMedia(string(platform), mId)
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

	for key, ep := range anizipMedia.Episodes {
		em := &metadata.EpisodeMetadata{
			AnidbId:               ep.AnidbEid,
			TvdbId:                ep.TvdbEid,
			Title:                 ep.GetTitle(),
			Image:                 ep.Image,
			AirDate:               ep.AirDate,
			Length:                ep.Runtime,
			Summary:               strings.ReplaceAll(ep.Summary, "`", "'"),
			Overview:              strings.ReplaceAll(ep.Overview, "`", "'"),
			EpisodeNumber:         ep.EpisodeNumber,
			Episode:               ep.Episode,
			SeasonNumber:          ep.SeasonNumber,
			AbsoluteEpisodeNumber: ep.AbsoluteEpisodeNumber,
			AnidbEid:              ep.AnidbEid,
			HasImage:              ep.Image != "",
		}
		if em.Length == 0 && ep.Length > 0 {
			em.Length = ep.Length
		}
		if em.Summary == "" && ep.Overview != "" {
			em.Summary = ep.Overview
		}
		if em.Overview == "" && ep.Summary != "" {
			em.Overview = ep.Summary
		}
		ret.Episodes[key] = em
	}

	// Event
	event := &metadata.AnimeMetadataEvent{
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
