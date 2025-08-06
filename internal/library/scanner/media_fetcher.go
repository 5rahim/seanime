package scanner

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/api/mal"
	"seanime/internal/api/metadata"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"seanime/internal/util/parallel"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

// MediaFetcher holds all anilist.BaseAnime that will be used for the comparison process
type MediaFetcher struct {
	AllMedia                     []*anilist.CompleteAnime
	CollectionMediaIds           []int
	UnknownMediaIds              []int // Media IDs that are not in the user's collection
	AnimeCollectionWithRelations *anilist.AnimeCollectionWithRelations
	ScanLogger                   *ScanLogger
}

type MediaFetcherOptions struct {
	Enhanced               bool
	Platform               platform.Platform
	MetadataProvider       metadata.Provider
	LocalFiles             []*anime.LocalFile
	CompleteAnimeCache     *anilist.CompleteAnimeCache
	Logger                 *zerolog.Logger
	AnilistRateLimiter     *limiter.Limiter
	DisableAnimeCollection bool
	ScanLogger             *ScanLogger
}

// NewMediaFetcher
// Calling this method will kickstart the fetch process
// When enhancing is false, MediaFetcher.AllMedia will be all anilist.BaseAnime from the user's AniList collection.
// When enhancing is true, MediaFetcher.AllMedia will be anilist.BaseAnime for each unique, parsed anime title and their relations.
func NewMediaFetcher(ctx context.Context, opts *MediaFetcherOptions) (ret *MediaFetcher, retErr error) {
	defer util.HandlePanicInModuleWithError("library/scanner/NewMediaFetcher", &retErr)

	if opts.Platform == nil ||
		opts.LocalFiles == nil ||
		opts.CompleteAnimeCache == nil ||
		opts.MetadataProvider == nil ||
		opts.Logger == nil ||
		opts.AnilistRateLimiter == nil {
		return nil, errors.New("missing options")
	}

	mf := new(MediaFetcher)
	mf.ScanLogger = opts.ScanLogger

	opts.Logger.Debug().
		Any("enhanced", opts.Enhanced).
		Msg("media fetcher: Creating media fetcher")

	if mf.ScanLogger != nil {
		mf.ScanLogger.LogMediaFetcher(zerolog.InfoLevel).
			Msg("Creating media fetcher")
	}

	// Invoke ScanMediaFetcherStarted hook
	event := &ScanMediaFetcherStartedEvent{
		Enhanced: opts.Enhanced,
	}
	hook.GlobalHookManager.OnScanMediaFetcherStarted().Trigger(event)
	opts.Enhanced = event.Enhanced

	// +---------------------+
	// |     All media       |
	// +---------------------+

	// Fetch latest user's AniList collection
	animeCollectionWithRelations, err := opts.Platform.GetAnimeCollectionWithRelations(ctx)
	if err != nil {
		return nil, err
	}

	mf.AnimeCollectionWithRelations = animeCollectionWithRelations

	mf.AllMedia = make([]*anilist.CompleteAnime, 0)

	if !opts.DisableAnimeCollection {
		// For each collection entry, append the media to AllMedia
		for _, list := range animeCollectionWithRelations.GetMediaListCollection().GetLists() {
			for _, entry := range list.GetEntries() {
				mf.AllMedia = append(mf.AllMedia, entry.GetMedia())

				// +---------------------+
				// |        Cache        |
				// +---------------------+
				// We assume the CompleteAnimeCache is empty. Add media to cache.
				opts.CompleteAnimeCache.Set(entry.GetMedia().ID, entry.GetMedia())
			}
		}
	}

	if mf.ScanLogger != nil {
		mf.ScanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Int("count", len(mf.AllMedia)).
			Msg("Fetched media from AniList collection")
	}

	//--------------------------------------------

	// Get the media IDs from the collection
	mf.CollectionMediaIds = lop.Map(mf.AllMedia, func(m *anilist.CompleteAnime, index int) int {
		return m.ID
	})

	//--------------------------------------------

	// +---------------------+
	// |      Enhanced       |
	// +---------------------+

	// If enhancing is on, scan media from local files and get their relations
	if opts.Enhanced {

		_, ok := FetchMediaFromLocalFiles(
			ctx,
			opts.Platform,
			opts.LocalFiles,
			opts.CompleteAnimeCache, // CompleteAnimeCache will be populated on success
			opts.MetadataProvider,
			opts.AnilistRateLimiter,
			mf.ScanLogger,
		)
		if ok {
			// We assume the CompleteAnimeCache is populated. We overwrite AllMedia with the cache content.
			// This is because the cache will contain all media from the user's collection AND scanned ones
			mf.AllMedia = make([]*anilist.CompleteAnime, 0)
			opts.CompleteAnimeCache.Range(func(key int, value *anilist.CompleteAnime) bool {
				mf.AllMedia = append(mf.AllMedia, value)
				return true
			})
		}
	}

	// +---------------------+
	// |   Unknown media     |
	// +---------------------+
	// Media that are not in the user's collection

	// Get the media that are not in the user's collection
	unknownMedia := lo.Filter(mf.AllMedia, func(m *anilist.CompleteAnime, _ int) bool {
		return !lo.Contains(mf.CollectionMediaIds, m.ID)
	})
	// Get the media IDs that are not in the user's collection
	mf.UnknownMediaIds = lop.Map(unknownMedia, func(m *anilist.CompleteAnime, _ int) int {
		return m.ID
	})

	if mf.ScanLogger != nil {
		mf.ScanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Int("unknownMediaCount", len(mf.UnknownMediaIds)).
			Int("allMediaCount", len(mf.AllMedia)).
			Msg("Finished creating media fetcher")
	}

	// Invoke ScanMediaFetcherCompleted hook
	completedEvent := &ScanMediaFetcherCompletedEvent{
		AllMedia:        mf.AllMedia,
		UnknownMediaIds: mf.UnknownMediaIds,
	}
	_ = hook.GlobalHookManager.OnScanMediaFetcherCompleted().Trigger(completedEvent)
	mf.AllMedia = completedEvent.AllMedia
	mf.UnknownMediaIds = completedEvent.UnknownMediaIds

	return mf, nil
}

//----------------------------------------------------------------------------------------------------------------------

// FetchMediaFromLocalFiles gets media and their relations from local file titles.
// It retrieves unique titles from local files,
// fetches mal.SearchResultAnime from MAL,
// uses these search results to get AniList IDs using metadata.AnimeMetadata mappings,
// queries AniList to retrieve all anilist.BaseAnime using anilist.GetBaseAnimeById and their relations using anilist.FetchMediaTree.
// It does not return an error if one of the steps fails.
// It returns the scanned media and a boolean indicating whether the process was successful.
func FetchMediaFromLocalFiles(
	ctx context.Context,
	platform platform.Platform,
	localFiles []*anime.LocalFile,
	completeAnime *anilist.CompleteAnimeCache,
	metadataProvider metadata.Provider,
	anilistRateLimiter *limiter.Limiter,
	scanLogger *ScanLogger,
) (ret []*anilist.CompleteAnime, ok bool) {
	defer util.HandlePanicInModuleThen("library/scanner/FetchMediaFromLocalFiles", func() {
		ok = false
	})

	if scanLogger != nil {
		scanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Str("module", "Enhanced").
			Msg("Fetching media from local files")
	}

	rateLimiter := limiter.NewLimiter(time.Second, 20)
	rateLimiter2 := limiter.NewLimiter(time.Second, 20)

	// Get titles
	titles := anime.GetUniqueAnimeTitlesFromLocalFiles(localFiles)

	if scanLogger != nil {
		scanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Str("module", "Enhanced").
			Str("context", spew.Sprint(titles)).
			Msg("Parsed titles from local files")
	}

	// +---------------------+
	// |     MyAnimeList     |
	// +---------------------+

	// Get MAL media from titles
	malSR := parallel.NewSettledResults[string, *mal.SearchResultAnime](titles)
	malSR.AllSettled(func(title string, index int) (*mal.SearchResultAnime, error) {
		rateLimiter.Wait()
		return mal.AdvancedSearchWithMAL(title)
	})
	malRes, ok := malSR.GetFulfilledResults()
	if !ok {
		return nil, false
	}

	// Get duplicate-free version of MAL media
	malMedia := lo.UniqBy(*malRes, func(res *mal.SearchResultAnime) int { return res.ID })
	// Get the MAL media IDs
	malIds := lop.Map(malMedia, func(n *mal.SearchResultAnime, index int) int { return n.ID })

	if scanLogger != nil {
		scanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Str("module", "Enhanced").
			Str("context", spew.Sprint(lo.Map(malMedia, func(n *mal.SearchResultAnime, _ int) string {
				return n.Name
			}))).
			Msg("Fetched MAL media from titles")
	}

	// +---------------------+
	// |       Animap        |
	// +---------------------+

	// Get Animap mappings for each MAL ID and store them in `metadataProvider`
	// This step is necessary because MAL doesn't provide AniList IDs and some MAL media don't exist on AniList
	lop.ForEach(malIds, func(id int, index int) {
		rateLimiter2.Wait()
		//_, _ = metadataProvider.GetAnimeMetadata(metadata.MalPlatform, id)
		_, _ = metadataProvider.GetCache().GetOrSet(metadata.GetAnimeMetadataCacheKey(metadata.MalPlatform, id), func() (*metadata.AnimeMetadata, error) {
			res, err := metadataProvider.GetAnimeMetadata(metadata.MalPlatform, id)
			return res, err
		})
	})

	// +---------------------+
	// |       AniList       |
	// +---------------------+

	// Retrieve the AniList IDs from the Animap mappings stored in the cache
	anilistIds := make([]int, 0)
	metadataProvider.GetCache().Range(func(key string, value *metadata.AnimeMetadata) bool {
		if value != nil {
			anilistIds = append(anilistIds, value.GetMappings().AnilistId)
		}
		return true
	})

	// Fetch all media from the AniList IDs
	anilistMedia := make([]*anilist.CompleteAnime, 0)
	lop.ForEach(anilistIds, func(id int, index int) {
		anilistRateLimiter.Wait()
		media, err := platform.GetAnimeWithRelations(ctx, id)
		if err == nil {
			anilistMedia = append(anilistMedia, media)
			if scanLogger != nil {
				scanLogger.LogMediaFetcher(zerolog.DebugLevel).
					Str("module", "Enhanced").
					Str("title", media.GetTitleSafe()).
					Msg("Fetched Anilist media from MAL id")
			}
		} else {
			if scanLogger != nil {
				scanLogger.LogMediaFetcher(zerolog.WarnLevel).
					Str("module", "Enhanced").
					Int("id", id).
					Msg("Failed to fetch Anilist media from MAL id")
			}
		}
	})

	if scanLogger != nil {
		scanLogger.LogMediaFetcher(zerolog.DebugLevel).
			Str("module", "Enhanced").
			Str("context", spew.Sprint(lo.Map(anilistMedia, func(n *anilist.CompleteAnime, _ int) string {
				return n.GetTitleSafe()
			}))).
			Msg("Fetched Anilist media from MAL ids")
	}

	// +---------------------+
	// |     MediaTree       |
	// +---------------------+

	// Create a new tree that will hold the fetched relations
	// /!\ This is redundant because we already have a cache, but `FetchMediaTree` needs its
	tree := anilist.NewCompleteAnimeRelationTree()

	start := time.Now()
	// For each media, fetch its relations
	// The relations are fetched in parallel and added to `completeAnime`
	lop.ForEach(anilistMedia, func(m *anilist.CompleteAnime, index int) {
		// We ignore errors because we want to continue even if one of the media fails
		_ = m.FetchMediaTree(anilist.FetchMediaTreeAll, platform.GetAnilistClient(), anilistRateLimiter, tree, completeAnime)
	})

	// +---------------------+
	// |        Cache        |
	// +---------------------+

	// Retrieve all media from the cache
	scanned := make([]*anilist.CompleteAnime, 0)
	completeAnime.Range(func(key int, value *anilist.CompleteAnime) bool {
		scanned = append(scanned, value)
		return true
	})

	if scanLogger != nil {
		scanLogger.LogMediaFetcher(zerolog.InfoLevel).
			Str("module", "Enhanced").
			Int("ms", int(time.Since(start).Milliseconds())).
			Int("count", len(scanned)).
			Str("context", spew.Sprint(lo.Map(scanned, func(n *anilist.CompleteAnime, _ int) string {
				return n.GetTitleSafe()
			}))).
			Msg("Finished fetching media from local files")
	}

	return scanned, true
}
