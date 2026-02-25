package shared_platform

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gqlgo/gqlgenc/clientv2"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

// devnote: I got lazy and used global variables

var ShouldCache = atomic.Bool{}
var IsWorking = atomic.Bool{}
var AnilistClient = atomic.Value{}

type failureRecord struct {
	timestamp time.Time
	err       error
}

var (
	failureTracking      = make([]failureRecord, 0)
	failureTrackingMutex sync.RWMutex
)

const (
	failureWindow     = 30 * time.Second // time window to consider failures
	failureThreshold  = 4                // number of failures needed to mark as down
	cleanupInterval   = 5 * time.Minute  // how often to clean up old failure records
	maxFailureRecords = 50               // maximum number of failure records to keep
)

func init() {
	ShouldCache.Store(true)
	IsWorking.Store(true)

	go func() {
		// Every 10 seconds, check if the AniList client is working
		for {
			time.Sleep(time.Second * 10)
			if !ShouldCache.Load() {
				IsWorking.Store(true)
				continue
			}
			if IsWorking.Load() {
				continue
			}
			if AnilistClient.Load() == nil {
				IsWorking.Store(true)
				continue
			}
			anilistClient, ok := AnilistClient.Load().(anilist.AnilistClient)
			if !ok {
				IsWorking.Store(true)
				continue
			}
			_, err := anilistClient.BaseAnimeByID(context.Background(), new(1))
			if err != nil {
				IsWorking.Store(false)
			} else {
				events.GlobalWSEventManager.SendEvent(events.InfoToast, "The AniList API is back online")
				IsWorking.Store(true)
			}
		}
	}()

	// periodic cleanup of old failure records
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			cleanupOldFailures()
		}
	}()
}

type (
	// CacheLayer is a "network-first" wrapper around an AniList client that caches fetched data in cache files.
	// It detects when the API client is not working and falls back to the cached data instead.
	// When the API client not working, it will still send the requests in the background and transition back to working state when the API client is working again.
	// Mutations will always return an error if the API client is not working.
	// Caching strategy:
	// - All queries to a specific media that IS in the anime collection or manga collection will be always cached/updated without limit
	// - Media that are NOT in the anime or manga collection will be bounded to a maximum of 10 entries
	CacheLayer struct {
		anilistClientRef     *util.Ref[anilist.AnilistClient]
		fileCacher           *filecache.Cacher
		buckets              map[string]filecache.PermanentBucket
		logger               *zerolog.Logger
		collectionMediaIDs   *result.Map[int, struct{}] // Track which media IDs are in collections
		lastCollectionUpdate time.Time                  // When collections were last fetched
		logoutFunc           func()                     // called when an invalid token is detected
	}
)

const (
	AnimeCollectionBucket          = "anime-collection"
	AnimeCollectionRelationsBucket = "anime-collection-relations"
	MangaCollectionBucket          = "manga-collection"
	BaseAnimeBucket                = "base-anime"
	BaseAnimeMalBucket             = "base-anime-mal"
	CompleteAnimeBucket            = "complete-anime"
	AnimeDetailsBucket             = "anime-details"
	BaseMangaBucket                = "base-manga"
	MangaDetailsBucket             = "manga-details"
	ViewerBucket                   = "viewer"
	ViewerStatsBucket              = "viewer-stats"
	StudioDetailsBucket            = "studio-details"
	AnimeAiringScheduleBucket      = "anime-airing-schedule"
	AnimeAiringScheduleRawBucket   = "anime-airing-schedule-raw"
	ListAnimeBucket                = "list-anime"
	ListRecentAnimeBucket          = "list-recent-anime"
	SearchBaseMangaBucket          = "search-base-manga"
	ListMangaBucket                = "list-manga"
	SearchBaseAnimeByIdsBucket     = "search-base-anime-by-ids"
	CustomQueryBucket              = "custom-query"

	maxNonCollectionCacheEntries      = 10
	maxNonCollectionMediaCacheEntries = 50
	// Collection update interval (refresh collection tracking every 30 minutes)
	collectionUpdateInterval = 30 * time.Minute
)

// addFailureRecord adds a new failure record to the tracking
func addFailureRecord(err error) {
	failureTrackingMutex.Lock()
	defer failureTrackingMutex.Unlock()

	now := time.Now()
	failureTracking = append(failureTracking, failureRecord{
		timestamp: now,
		err:       err,
	})

	// keep only the most recent records
	if len(failureTracking) > maxFailureRecords {
		failureTracking = failureTracking[len(failureTracking)-maxFailureRecords:]
	}
}

// getRecentFailureCount returns the number of failures within the failure window
func getRecentFailureCount() int {
	failureTrackingMutex.RLock()
	defer failureTrackingMutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-failureWindow)
	count := 0

	for _, record := range failureTracking {
		if record.timestamp.After(cutoff) {
			count++
		}
	}

	return count
}

// cleanupOldFailures removes failure records older than the failure window
func cleanupOldFailures() {
	failureTrackingMutex.Lock()
	defer failureTrackingMutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-failureWindow)
	validRecords := make([]failureRecord, 0, len(failureTracking))

	for _, record := range failureTracking {
		if record.timestamp.After(cutoff) {
			validRecords = append(validRecords, record)
		}
	}

	failureTracking = validRecords
}

// clearFailureTracking clears all failure records (called when API comes back online)
func clearFailureTracking() {
	failureTrackingMutex.Lock()
	defer failureTrackingMutex.Unlock()
	failureTracking = failureTracking[:0]
}

// NewCacheLayer returns a new instance of the global cache layer.
// An optional logoutFunc can be passed to perform server-side cleanup when an invalid token is detected.
func NewCacheLayer(anilistClientRef *util.Ref[anilist.AnilistClient], logoutFunc ...func()) anilist.AnilistClient {
	fileCacher, err := filecache.NewCacher(anilistClientRef.Get().GetCacheDir())
	if err != nil {
		return anilistClientRef.Get()
	}

	buckets := make(map[string]filecache.PermanentBucket)
	buckets[AnimeCollectionBucket] = filecache.NewPermanentBucket(AnimeCollectionBucket)
	buckets[AnimeCollectionRelationsBucket] = filecache.NewPermanentBucket(AnimeCollectionRelationsBucket)
	buckets[MangaCollectionBucket] = filecache.NewPermanentBucket(MangaCollectionBucket)
	buckets[BaseAnimeBucket] = filecache.NewPermanentBucket(BaseAnimeBucket)
	buckets[BaseAnimeMalBucket] = filecache.NewPermanentBucket(BaseAnimeMalBucket)
	buckets[CompleteAnimeBucket] = filecache.NewPermanentBucket(CompleteAnimeBucket)
	buckets[AnimeDetailsBucket] = filecache.NewPermanentBucket(AnimeDetailsBucket)
	buckets[BaseMangaBucket] = filecache.NewPermanentBucket(BaseMangaBucket)
	buckets[MangaDetailsBucket] = filecache.NewPermanentBucket(MangaDetailsBucket)
	buckets[ViewerBucket] = filecache.NewPermanentBucket(ViewerBucket)
	buckets[ViewerStatsBucket] = filecache.NewPermanentBucket(ViewerStatsBucket)
	buckets[StudioDetailsBucket] = filecache.NewPermanentBucket(StudioDetailsBucket)
	buckets[AnimeAiringScheduleBucket] = filecache.NewPermanentBucket(AnimeAiringScheduleBucket)
	buckets[AnimeAiringScheduleRawBucket] = filecache.NewPermanentBucket(AnimeAiringScheduleRawBucket)
	buckets[ListAnimeBucket] = filecache.NewPermanentBucket(ListAnimeBucket)
	buckets[ListRecentAnimeBucket] = filecache.NewPermanentBucket(ListRecentAnimeBucket)
	buckets[SearchBaseMangaBucket] = filecache.NewPermanentBucket(SearchBaseMangaBucket)
	buckets[ListMangaBucket] = filecache.NewPermanentBucket(ListMangaBucket)
	buckets[SearchBaseAnimeByIdsBucket] = filecache.NewPermanentBucket(SearchBaseAnimeByIdsBucket)
	buckets[CustomQueryBucket] = filecache.NewPermanentBucket(CustomQueryBucket)

	logger := util.NewLogger()

	var logout func()
	if len(logoutFunc) > 0 {
		logout = logoutFunc[0]
	}

	cl := &CacheLayer{
		anilistClientRef:   anilistClientRef,
		fileCacher:         fileCacher,
		buckets:            buckets,
		logger:             logger,
		collectionMediaIDs: result.NewMap[int, struct{}](),
		logoutFunc:         logout,
	}

	AnilistClient.Store(anilistClientRef.Get())

	return cl
}

var _ anilist.AnilistClient = (*CacheLayer)(nil)

func (c *CacheLayer) IsAuthenticated() bool {
	return c.anilistClientRef.Get().IsAuthenticated()
}

func (c *CacheLayer) GetCacheDir() string {
	return c.anilistClientRef.Get().GetCacheDir()
}

func (c *CacheLayer) CustomQuery(body []byte, logger *zerolog.Logger, token ...string) (interface{}, error) {
	// Use the stringified body as cache key
	cacheKey := string(body)
	bucket := c.buckets[CustomQueryBucket]

	// Try network first if API is working
	if IsWorking.Load() {
		res, err := c.anilistClientRef.Get().CustomQuery(body, logger, token...)
		c.checkAndUpdateWorkingState(err)

		if err == nil {
			go func() {
				if !ShouldCache.Load() {
					return
				}
				allData, err := filecache.GetAll[interface{}](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
				if err == nil && len(allData) >= maxNonCollectionCacheEntries {
					_ = c.fileCacher.DeletePermOldest(bucket)
				}

				if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
					c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache custom query result")
				}
			}()
			return res, nil
		}
	} else {
		// If API is not working, try it in the background to check if it's back
		go func() {
			res, err := c.anilistClientRef.Get().CustomQuery(body, logger, token...)
			c.checkAndUpdateWorkingState(err)
			if err == nil {
				// Cache the result for future use with bounded size
				allData, err := filecache.GetAll[interface{}](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
				if err == nil && len(allData) >= maxNonCollectionCacheEntries {
					_ = c.fileCacher.DeletePermOldest(bucket)
				}

				if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
					c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache background custom query result")
				}
			}
		}()
	}

	// Fall back to cache
	var cached interface{}
	found, err := c.fileCacher.GetPerm(bucket, cacheKey, &cached)
	if err != nil {
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no cached data available")
	}

	c.logger.Debug().Str("bucket", CustomQueryBucket).Str("key", cacheKey).Msg("anilist cache: Serving custom query from cache")
	return cached, nil
}

// checkAndUpdateWorkingState checks if the API client is working and updates the state
func (c *CacheLayer) checkAndUpdateWorkingState(err error) {
	if err != nil {
		// Skip context.Canceled errors, not indicative of API issues
		if errors.Is(err, context.Canceled) {
			return
		}

		// skip 404 errors
		if strings.Contains(err.Error(), "404") {
			return
		}

		// handle invalid token
		if strings.Contains(err.Error(), "Invalid token") {
			events.GlobalWSEventManager.SendEvent(events.ServerLoggedOutAnilist, "Your AniList session has expired. Please log in again.")
			if c.logoutFunc != nil {
				go c.logoutFunc()
			}
			return
		}

		// Add failure to tracking
		addFailureRecord(err)

		// Only mark as down if we have enough recent failures and are currently marked as working
		if IsWorking.Load() {
			recentFailures := getRecentFailureCount()
			if recentFailures >= failureThreshold {
				c.logger.Warn().
					Err(err).
					Int("recent_failures", recentFailures).
					Dur("within_window", failureWindow).
					Msg("anilist cache: Multiple API failures detected, switching to cache-only mode.")
				events.GlobalWSEventManager.SendEvent(events.WarningToast,
					fmt.Sprintf("The AniList API is experiencing issues (%d failures in %v), switching to cache-only mode.",
						recentFailures, failureWindow))
				IsWorking.Store(false)
			} else {
				c.logger.Debug().
					Err(err).
					Int("recent_failures", recentFailures).
					Int("threshold", failureThreshold).
					Msg("anilist cache: API failure recorded, monitoring for more failures")
			}
		}
	} else {
		// clear failure tracking and mark as working if not already
		if !IsWorking.Load() {
			c.logger.Info().Msg("anilist cache: API client is working again, switching back to network-first mode.")
			events.GlobalWSEventManager.SendEvent(events.InfoToast, "The AniList API is back online")
			IsWorking.Store(true)
		}
		clearFailureTracking()
	}
}

// generateCacheKey generates a cache key from the given parameters
func (c *CacheLayer) generateCacheKey(params ...interface{}) string {
	var keyParts []string
	for _, param := range params {
		if param == nil {
			keyParts = append(keyParts, "nil")
			continue
		}
		switch v := param.(type) {
		case *int:
			if v != nil {
				keyParts = append(keyParts, strconv.Itoa(*v))
			} else {
				keyParts = append(keyParts, "nil")
			}
		case *string:
			if v != nil {
				keyParts = append(keyParts, *v)
			} else {
				keyParts = append(keyParts, "nil")
			}
		case *bool:
			if v != nil {
				keyParts = append(keyParts, strconv.FormatBool(*v))
			} else {
				keyParts = append(keyParts, "nil")
			}
		case []*int:
			for _, id := range v {
				if id != nil {
					keyParts = append(keyParts, strconv.Itoa(*id))
				}
			}
		case []*string:
			for _, s := range v {
				if s != nil {
					keyParts = append(keyParts, *s)
				}
			}
		default:
			keyParts = append(keyParts, fmt.Sprintf("%v", param))
		}
	}
	return lo.Reduce(keyParts, func(acc, item string, _ int) string {
		if acc == "" {
			return item
		}
		return acc + "-" + item
	}, "")
}

// isInCollection checks if a media ID is in the user's collection
func (c *CacheLayer) isInCollection(mediaID int) bool {
	// Update collection tracking if needed
	c.updateCollectionTracking()
	_, ok := c.collectionMediaIDs.Get(mediaID)
	return ok
}

// updateCollectionTracking updates the collection media IDs tracking
func (c *CacheLayer) updateCollectionTracking() {
	if time.Since(c.lastCollectionUpdate) < collectionUpdateInterval {
		return
	}

	go func() {
		defer func() {
			c.lastCollectionUpdate = time.Now()
		}()

		// Try to fetch anime collection
		if animeCollection, err := c.anilistClientRef.Get().AnimeCollection(context.Background(), nil); err == nil && animeCollection != nil {
			for _, list := range animeCollection.MediaListCollection.Lists {
				if list != nil {
					for _, entry := range list.Entries {
						if entry != nil && entry.Media != nil {
							c.collectionMediaIDs.Set(entry.Media.ID, struct{}{})
						}
					}
				}
			}
		}

		// Try to fetch manga collection
		if mangaCollection, err := c.anilistClientRef.Get().MangaCollection(context.Background(), nil); err == nil && mangaCollection != nil {
			for _, list := range mangaCollection.MediaListCollection.Lists {
				if list != nil {
					for _, entry := range list.Entries {
						if entry != nil && entry.Media != nil {
							c.collectionMediaIDs.Set(entry.Media.ID, struct{}{})
						}
					}
				}
			}
		}
	}()
}

// networkFirstGet performs a network-first get operation with caching
func networkFirstGet[T any](c *CacheLayer, bucketName string, cacheKey string, networkFn func() (*T, error)) (*T, error) {
	if !ShouldCache.Load() {
		return networkFn()
	}

	bucket := c.buckets[bucketName]

	// Try network first if API is working
	if IsWorking.Load() {
		res, err := networkFn()
		c.checkAndUpdateWorkingState(err)

		if err == nil && res != nil {
			// Cache the successful result
			if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache result")
			}
			return res, nil
		}
	} else {
		// If API is not working, try it in the background to check if it's back
		go func() {
			res, err := networkFn()
			c.checkAndUpdateWorkingState(err)
			if err == nil && res != nil {
				// Cache the result for future use
				if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
					c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache background result")
				}
			}
		}()
	}

	// Fall back to cache
	var cached T
	found, err := c.fileCacher.GetPerm(bucket, cacheKey, &cached)
	if err != nil {
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no cached data available")
	}

	c.logger.Debug().Str("bucket", bucketName).Str("key", cacheKey).Msg("anilist cache: Serving from cache")
	return &cached, nil
}

// boundedCacheSet caches data with a limit on non-collection entries
func (c *CacheLayer) boundedCacheSet(bucketName string, cacheKey string, data interface{}, mediaID int) error {
	if !ShouldCache.Load() {
		return nil
	}

	bucket := c.buckets[bucketName]

	// Always cache collection media
	if c.isInCollection(mediaID) {
		return c.fileCacher.SetPerm(bucket, cacheKey, data)
	}

	// For non-collection media, enforce the limit
	allData, err := filecache.GetAll[interface{}](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
	if err != nil {
		return err
	}

	// If we're at the limit, remove the oldest entry (simple FIFO for now)
	if len(allData) >= maxNonCollectionMediaCacheEntries {
		// Remove the first key we find (this is a simple implementation)
		for key := range allData {
			if err := c.fileCacher.DeletePerm(bucket, key); err == nil {
				break
			}
		}
	}

	return c.fileCacher.SetPerm(bucket, cacheKey, data)
}

// updateCollectionTrackingFromAnimeCollection updates collection tracking from anime collection
func (c *CacheLayer) updateCollectionTrackingFromAnimeCollection(collection *anilist.AnimeCollection) {
	if !ShouldCache.Load() {
		return
	}

	if !ShouldCache.Load() || collection == nil || collection.MediaListCollection == nil {
		return
	}

	for _, list := range collection.MediaListCollection.Lists {
		if list != nil {
			for _, entry := range list.Entries {
				if entry != nil && entry.Media != nil {
					c.collectionMediaIDs.Set(entry.Media.ID, struct{}{})
				}
			}
		}
	}
	c.lastCollectionUpdate = time.Now()
}

func (c *CacheLayer) updateCollectionTrackingFromAnimeCollectionWithRelations(collection *anilist.AnimeCollectionWithRelations) {
	if !ShouldCache.Load() {
		return
	}

	if !ShouldCache.Load() || collection == nil || collection.MediaListCollection == nil {
		return
	}

	for _, list := range collection.MediaListCollection.Lists {
		if list != nil {
			for _, entry := range list.Entries {
				if entry != nil && entry.Media != nil {
					c.collectionMediaIDs.Set(entry.Media.ID, struct{}{})
				}
			}
		}
	}
	c.lastCollectionUpdate = time.Now()
}

func (c *CacheLayer) updateCollectionTrackingFromMangaCollection(collection *anilist.MangaCollection) {
	if !ShouldCache.Load() {
		return
	}

	if !ShouldCache.Load() || collection == nil || collection.MediaListCollection == nil {
		return
	}

	for _, list := range collection.MediaListCollection.Lists {
		if list != nil {
			for _, entry := range list.Entries {
				if entry != nil && entry.Media != nil {
					c.collectionMediaIDs.Set(entry.Media.ID, struct{}{})
				}
			}
		}
	}
	c.lastCollectionUpdate = time.Now()
}

// invalidateMediaCaches invalidates caches for a specific media ID
func (c *CacheLayer) invalidateMediaCaches(mediaID int) {
	if !ShouldCache.Load() {
		return
	}

	mediaIDStr := strconv.Itoa(mediaID)

	// Delete from all media-specific buckets
	buckets := []string{
		BaseAnimeBucket,
		CompleteAnimeBucket,
		AnimeDetailsBucket,
		BaseMangaBucket,
		MangaDetailsBucket,
	}

	for _, bucketName := range buckets {
		bucket := c.buckets[bucketName]
		if err := c.fileCacher.DeletePerm(bucket, mediaIDStr); err != nil {
			c.logger.Debug().Err(err).Str("bucket", bucketName).Int("mediaID", mediaID).Msg("anilist cache: Failed to invalidate cache entry")
		}
	}
}

// invalidateCollectionCaches invalidates all collection caches and custom queries
func (c *CacheLayer) invalidateCollectionCaches() {
	if !ShouldCache.Load() {
		return
	}

	collectionBuckets := []string{
		AnimeCollectionBucket,
		AnimeCollectionRelationsBucket,
		MangaCollectionBucket,
		CustomQueryBucket,
	}

	for _, bucketName := range collectionBuckets {
		bucket := c.buckets[bucketName]
		if err := c.fileCacher.EmptyPerm(bucket); err != nil {
			c.logger.Warn().Err(err).Str("bucket", bucketName).Msg("anilist cache: Failed to invalidate collection cache")
		}
	}

	// Reset collection tracking
	c.collectionMediaIDs.Clear()
	c.lastCollectionUpdate = time.Time{}
}

// extractBaseAnimeFromCollection attempts to extract BaseAnime data from cached anime collection
func (c *CacheLayer) extractBaseAnimeFromCollection(mediaID int) *anilist.BaseAnimeByID {
	// Try anime collection
	bucket := c.buckets[AnimeCollectionBucket]
	cacheKey := c.generateCacheKey("collection", nil)
	var animeCollection anilist.AnimeCollection
	found, err := c.fileCacher.GetPerm(bucket, cacheKey, &animeCollection)
	if err == nil && found && animeCollection.MediaListCollection != nil {
		for _, list := range animeCollection.MediaListCollection.Lists {
			if list != nil {
				for _, entry := range list.Entries {
					if entry != nil && entry.Media != nil && entry.Media.ID == mediaID {
						return &anilist.BaseAnimeByID{
							Media: entry.Media,
						}
					}
				}
			}
		}
	}

	// Try anime collection with relations
	relBucket := c.buckets[AnimeCollectionRelationsBucket]
	var animeCollectionRel anilist.AnimeCollectionWithRelations
	found, err = c.fileCacher.GetPerm(relBucket, cacheKey, &animeCollectionRel)
	if err == nil && found && animeCollectionRel.MediaListCollection != nil {
		for _, list := range animeCollectionRel.MediaListCollection.Lists {
			if list != nil {
				for _, entry := range list.Entries {
					if entry != nil && entry.Media != nil && entry.Media.ID == mediaID {
						return &anilist.BaseAnimeByID{
							Media: entry.Media.ToBaseAnime(),
						}
					}
				}
			}
		}
	}

	return nil
}

// extractBaseMangaFromCollection attempts to extract BaseManga data from cached manga collection
func (c *CacheLayer) extractBaseMangaFromCollection(mediaID int) *anilist.BaseMangaByID {
	if !ShouldCache.Load() {
		return nil
	}

	bucket := c.buckets[MangaCollectionBucket]
	cacheKey := c.generateCacheKey("collection", nil)
	var mangaCollection anilist.MangaCollection
	found, err := c.fileCacher.GetPerm(bucket, cacheKey, &mangaCollection)
	if err == nil && found && mangaCollection.MediaListCollection != nil {
		for _, list := range mangaCollection.MediaListCollection.Lists {
			if list != nil {
				for _, entry := range list.Entries {
					if entry != nil && entry.Media != nil && entry.Media.ID == mediaID {
						return &anilist.BaseMangaByID{
							Media: entry.Media,
						}
					}
				}
			}
		}
	}

	return nil
}

// networkFirstGetWithBoundedCache performs a network-first get operation with bounded caching for list/search results
func networkFirstGetWithBoundedCache[T any](c *CacheLayer, bucketName string, cacheKey string, networkFn func() (*T, error)) (*T, error) {
	bucket := c.buckets[bucketName]

	// Try network first if API is working
	if IsWorking.Load() {
		res, err := networkFn()
		c.checkAndUpdateWorkingState(err)

		if err == nil && res != nil {
			// Cache the successful result with bounded size
			go func() {
				// For list/search results, always apply bounded caching
				allData, err := filecache.GetAll[interface{}](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
				if err == nil && len(allData) >= maxNonCollectionCacheEntries {
					_ = c.fileCacher.DeletePermOldest(bucket)
				}

				if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
					c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache bounded result")
				}
			}()
			return res, nil
		}
	} else {
		// If API is not working, try it in the background to check if it's back
		go func() {
			res, err := networkFn()
			c.checkAndUpdateWorkingState(err)
			if err == nil && res != nil {
				// Cache the result for future use with bounded size
				allData, err := filecache.GetAll[interface{}](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
				if err == nil && len(allData) >= maxNonCollectionCacheEntries {
					_ = c.fileCacher.DeletePermOldest(bucket)
				}

				if err := c.fileCacher.SetPerm(bucket, cacheKey, res); err != nil {
					c.logger.Warn().Err(err).Msg("anilist cache: Failed to cache background bounded result")
				}
			}
		}()
	}

	// Fall back to cache
	var cached T
	found, err := c.fileCacher.GetPerm(bucket, cacheKey, &cached)
	if err != nil {
		return nil, fmt.Errorf("cache lookup failed: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no cached data available")
	}

	c.logger.Debug().Str("bucket", bucketName).Str("key", cacheKey).Msg("anilist cache: Serving bounded result from cache")
	return &cached, nil
}

func (c *CacheLayer) AnimeCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*anilist.AnimeCollection, error) {
	cacheKey := c.generateCacheKey("collection", nil)
	res, err := networkFirstGet(c, AnimeCollectionBucket, cacheKey, func() (*anilist.AnimeCollection, error) {
		return c.anilistClientRef.Get().AnimeCollection(ctx, userName, interceptors...)
	})

	// Update collection tracking with the fetched data
	if err == nil && res != nil {
		go c.updateCollectionTrackingFromAnimeCollection(res)
	}

	return res, err
}

func (c *CacheLayer) AnimeCollectionWithRelations(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*anilist.AnimeCollectionWithRelations, error) {
	cacheKey := c.generateCacheKey("collection-relations", nil)
	res, err := networkFirstGet(c, AnimeCollectionRelationsBucket, cacheKey, func() (*anilist.AnimeCollectionWithRelations, error) {
		return c.anilistClientRef.Get().AnimeCollectionWithRelations(ctx, userName, interceptors...)
	})

	// Update collection tracking with the fetched data
	if err == nil && res != nil {
		go c.updateCollectionTrackingFromAnimeCollectionWithRelations(res)
	}

	return res, err
}

func (c *CacheLayer) BaseAnimeByMalID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.BaseAnimeByMalID, error) {
	if id == nil {
		return c.anilistClientRef.Get().BaseAnimeByMalID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey("mal", id)
	return networkFirstGet(c, BaseAnimeMalBucket, cacheKey, func() (*anilist.BaseAnimeByMalID, error) {
		return c.anilistClientRef.Get().BaseAnimeByMalID(ctx, id, interceptors...)
	})
}

func (c *CacheLayer) BaseAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.BaseAnimeByID, error) {
	if id == nil {
		return c.anilistClientRef.Get().BaseAnimeByID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	res, err := networkFirstGet(c, BaseAnimeBucket, cacheKey, func() (*anilist.BaseAnimeByID, error) {
		return c.anilistClientRef.Get().BaseAnimeByID(ctx, id, interceptors...)
	})

	// If network and direct cache failed, try to extract from collection cache
	if err != nil {
		if collectionResult := c.extractBaseAnimeFromCollection(*id); collectionResult != nil {
			c.logger.Debug().Int("mediaID", *id).Msg("anilist cache: Extracted BaseAnime from collection cache")
			return collectionResult, nil
		}
	}

	// If successful, update bounded cache for non-collection media
	if err == nil && res != nil {
		go func() {
			if err := c.boundedCacheSet(BaseAnimeBucket, cacheKey, res, *id); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: Failed to update bounded cache")
			}
		}()
	}

	return res, err
}

func (c *CacheLayer) SearchBaseAnimeByIds(ctx context.Context, ids []*int, page *int, perPage *int, status []*anilist.MediaStatus, inCollection *bool, sort []*anilist.MediaSort, season *anilist.MediaSeason, year *int, genre *string, format *anilist.MediaFormat, interceptors ...clientv2.RequestInterceptor) (*anilist.SearchBaseAnimeByIds, error) {
	cacheKey := c.generateCacheKey(ids, page, perPage, status, inCollection, sort, season, year, genre, format)
	return networkFirstGetWithBoundedCache(c, SearchBaseAnimeByIdsBucket, cacheKey, func() (*anilist.SearchBaseAnimeByIds, error) {
		return c.anilistClientRef.Get().SearchBaseAnimeByIds(ctx, ids, page, perPage, status, inCollection, sort, season, year, genre, format, interceptors...)
	})
}

func (c *CacheLayer) CompleteAnimeByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.CompleteAnimeByID, error) {
	if id == nil {
		return c.anilistClientRef.Get().CompleteAnimeByID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	res, err := networkFirstGet(c, CompleteAnimeBucket, cacheKey, func() (*anilist.CompleteAnimeByID, error) {
		return c.anilistClientRef.Get().CompleteAnimeByID(ctx, id, interceptors...)
	})

	// If successful, update bounded cache for non-collection media
	if err == nil && res != nil {
		go func() {
			if err := c.boundedCacheSet(CompleteAnimeBucket, cacheKey, res, *id); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: failed to update bounded cache")
			}
		}()
	}

	return res, err
}

func (c *CacheLayer) AnimeDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.AnimeDetailsByID, error) {
	if id == nil {
		return c.anilistClientRef.Get().AnimeDetailsByID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	res, err := networkFirstGet(c, AnimeDetailsBucket, cacheKey, func() (*anilist.AnimeDetailsByID, error) {
		return c.anilistClientRef.Get().AnimeDetailsByID(ctx, id, interceptors...)
	})

	// If successful, update bounded cache for non-collection media
	if err == nil && res != nil {
		go func() {
			if err := c.boundedCacheSet(AnimeDetailsBucket, cacheKey, res, *id); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: failed to update bounded cache")
			}
		}()
	}

	return res, err
}

func (c *CacheLayer) ListAnime(ctx context.Context, page *int, search *string, perPage *int, sort []*anilist.MediaSort, status []*anilist.MediaStatus, genres []*string, averageScoreGreater *int, season *anilist.MediaSeason, seasonYear *int, format *anilist.MediaFormat, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*anilist.ListAnime, error) {
	cacheKey := c.generateCacheKey(page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, isAdult)
	return networkFirstGetWithBoundedCache(c, ListAnimeBucket, cacheKey, func() (*anilist.ListAnime, error) {
		return c.anilistClientRef.Get().ListAnime(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, isAdult, interceptors...)
	})
}

func (c *CacheLayer) ListRecentAnime(ctx context.Context, page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool, interceptors ...clientv2.RequestInterceptor) (*anilist.ListRecentAnime, error) {
	cacheKey := c.generateCacheKey(page, perPage, airingAtGreater, airingAtLesser, notYetAired)
	return networkFirstGetWithBoundedCache(c, ListRecentAnimeBucket, cacheKey, func() (*anilist.ListRecentAnime, error) {
		return c.anilistClientRef.Get().ListRecentAnime(ctx, page, perPage, airingAtGreater, airingAtLesser, notYetAired, interceptors...)
	})
}

func (c *CacheLayer) UpdateMediaListEntry(ctx context.Context, mediaID *int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*anilist.UpdateMediaListEntry, error) {
	// Mutations require the API to be working
	if !IsWorking.Load() {
		return nil, fmt.Errorf("anilist cache: API client is not working, mutation operations are not available")
	}

	res, err := c.anilistClientRef.Get().UpdateMediaListEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, interceptors...)
	c.checkAndUpdateWorkingState(err)

	// Invalidate relevant caches on successful mutation
	if err == nil && mediaID != nil {
		c.invalidateMediaCaches(*mediaID)
		c.invalidateCollectionCaches()
	}

	return res, err
}

func (c *CacheLayer) UpdateMediaListEntryProgress(ctx context.Context, mediaID *int, progress *int, status *anilist.MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*anilist.UpdateMediaListEntryProgress, error) {
	// Mutations require the API to be working
	if !IsWorking.Load() {
		return nil, fmt.Errorf("anilist cache: API client is not working, mutation operations are not available")
	}

	res, err := c.anilistClientRef.Get().UpdateMediaListEntryProgress(ctx, mediaID, progress, status, interceptors...)
	c.checkAndUpdateWorkingState(err)

	// Invalidate relevant caches on successful mutation
	if err == nil && mediaID != nil {
		c.invalidateMediaCaches(*mediaID)
		c.invalidateCollectionCaches()
	}

	return res, err
}

func (c *CacheLayer) UpdateMediaListEntryRepeat(ctx context.Context, mediaID *int, repeat *int, interceptors ...clientv2.RequestInterceptor) (*anilist.UpdateMediaListEntryRepeat, error) {
	// Mutations require the API to be working
	if !IsWorking.Load() {
		return nil, fmt.Errorf("anilist cache: API client is not working, mutation operations are not available")
	}

	res, err := c.anilistClientRef.Get().UpdateMediaListEntryRepeat(ctx, mediaID, repeat, interceptors...)
	c.checkAndUpdateWorkingState(err)

	// Invalidate relevant caches on successful mutation
	if err == nil && mediaID != nil {
		c.invalidateMediaCaches(*mediaID)
		c.invalidateCollectionCaches()
	}

	return res, err
}

func (c *CacheLayer) DeleteEntry(ctx context.Context, mediaListEntryID *int, interceptors ...clientv2.RequestInterceptor) (*anilist.DeleteEntry, error) {
	// Mutations require the API to be working
	if !IsWorking.Load() {
		return nil, fmt.Errorf("anilist cache: API client is not working, mutation operations are not available")
	}

	res, err := c.anilistClientRef.Get().DeleteEntry(ctx, mediaListEntryID, interceptors...)
	c.checkAndUpdateWorkingState(err)

	// Invalidate collection caches on successful deletion
	if err == nil {
		c.invalidateCollectionCaches()
	}

	return res, err
}

func (c *CacheLayer) MangaCollection(ctx context.Context, userName *string, interceptors ...clientv2.RequestInterceptor) (*anilist.MangaCollection, error) {
	cacheKey := c.generateCacheKey("collection", nil)
	res, err := networkFirstGet(c, MangaCollectionBucket, cacheKey, func() (*anilist.MangaCollection, error) {
		return c.anilistClientRef.Get().MangaCollection(ctx, userName, interceptors...)
	})

	// Update collection tracking with the fetched data
	if err == nil && res != nil {
		go c.updateCollectionTrackingFromMangaCollection(res)
	}

	return res, err
}

func (c *CacheLayer) SearchBaseManga(ctx context.Context, page *int, perPage *int, sort []*anilist.MediaSort, search *string, status []*anilist.MediaStatus, interceptors ...clientv2.RequestInterceptor) (*anilist.SearchBaseManga, error) {
	cacheKey := c.generateCacheKey(page, perPage, sort, search, status)
	return networkFirstGetWithBoundedCache(c, SearchBaseMangaBucket, cacheKey, func() (*anilist.SearchBaseManga, error) {
		return c.anilistClientRef.Get().SearchBaseManga(ctx, page, perPage, sort, search, status, interceptors...)
	})
}

func (c *CacheLayer) BaseMangaByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.BaseMangaByID, error) {
	if id == nil {
		return c.anilistClientRef.Get().BaseMangaByID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	res, err := networkFirstGet(c, BaseMangaBucket, cacheKey, func() (*anilist.BaseMangaByID, error) {
		return c.anilistClientRef.Get().BaseMangaByID(ctx, id, interceptors...)
	})

	// If network and direct cache failed, try to extract from collection cache
	if err != nil {
		if collectionResult := c.extractBaseMangaFromCollection(*id); collectionResult != nil {
			c.logger.Debug().Int("mediaID", *id).Msg("anilist cache: Extracted BaseManga from collection cache")
			return collectionResult, nil
		}
	}

	// If successful, update bounded cache for non-collection media
	if err == nil && res != nil {
		go func() {
			if err := c.boundedCacheSet(BaseMangaBucket, cacheKey, res, *id); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: Failed to update bounded cache")
			}
		}()
	}

	return res, err
}

func (c *CacheLayer) MangaDetailsByID(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.MangaDetailsByID, error) {
	if id == nil {
		return c.anilistClientRef.Get().MangaDetailsByID(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	res, err := networkFirstGet(c, MangaDetailsBucket, cacheKey, func() (*anilist.MangaDetailsByID, error) {
		return c.anilistClientRef.Get().MangaDetailsByID(ctx, id, interceptors...)
	})

	// If successful, update bounded cache for non-collection media
	if err == nil && res != nil {
		go func() {
			if err := c.boundedCacheSet(MangaDetailsBucket, cacheKey, res, *id); err != nil {
				c.logger.Warn().Err(err).Msg("anilist cache: failed to update bounded cache")
			}
		}()
	}

	return res, err
}

func (c *CacheLayer) ListManga(ctx context.Context, page *int, search *string, perPage *int, sort []*anilist.MediaSort, status []*anilist.MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *anilist.MediaFormat, countryOfOrigin *string, isAdult *bool, interceptors ...clientv2.RequestInterceptor) (*anilist.ListManga, error) {
	cacheKey := c.generateCacheKey(page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult)
	return networkFirstGetWithBoundedCache(c, ListMangaBucket, cacheKey, func() (*anilist.ListManga, error) {
		return c.anilistClientRef.Get().ListManga(ctx, page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult, interceptors...)
	})
}

func (c *CacheLayer) ViewerStats(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*anilist.ViewerStats, error) {
	cacheKey := "stats"
	return networkFirstGet(c, ViewerStatsBucket, cacheKey, func() (*anilist.ViewerStats, error) {
		return c.anilistClientRef.Get().ViewerStats(ctx, interceptors...)
	})
}

func (c *CacheLayer) StudioDetails(ctx context.Context, id *int, interceptors ...clientv2.RequestInterceptor) (*anilist.StudioDetails, error) {
	if id == nil {
		return c.anilistClientRef.Get().StudioDetails(ctx, id, interceptors...)
	}

	cacheKey := c.generateCacheKey(id)
	return networkFirstGet(c, StudioDetailsBucket, cacheKey, func() (*anilist.StudioDetails, error) {
		return c.anilistClientRef.Get().StudioDetails(ctx, id, interceptors...)
	})
}

func (c *CacheLayer) GetViewer(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*anilist.GetViewer, error) {
	cacheKey := "viewer"
	return networkFirstGet(c, ViewerBucket, cacheKey, func() (*anilist.GetViewer, error) {
		return c.anilistClientRef.Get().GetViewer(ctx, interceptors...)
	})
}

func (c *CacheLayer) AnimeAiringSchedule(ctx context.Context, ids []*int, season *anilist.MediaSeason, seasonYear *int, previousSeason *anilist.MediaSeason, previousSeasonYear *int, nextSeason *anilist.MediaSeason, nextSeasonYear *int, interceptors ...clientv2.RequestInterceptor) (*anilist.AnimeAiringSchedule, error) {
	cacheKey := c.generateCacheKey(ids, season, seasonYear, previousSeason, previousSeasonYear, nextSeason, nextSeasonYear)
	return networkFirstGet(c, AnimeAiringScheduleBucket, cacheKey, func() (*anilist.AnimeAiringSchedule, error) {
		return c.anilistClientRef.Get().AnimeAiringSchedule(ctx, ids, season, seasonYear, previousSeason, previousSeasonYear, nextSeason, nextSeasonYear, interceptors...)
	})
}

func (c *CacheLayer) AnimeAiringScheduleRaw(ctx context.Context, ids []*int, interceptors ...clientv2.RequestInterceptor) (*anilist.AnimeAiringScheduleRaw, error) {
	cacheKey := c.generateCacheKey(ids)
	return networkFirstGet(c, AnimeAiringScheduleRawBucket, cacheKey, func() (*anilist.AnimeAiringScheduleRaw, error) {
		return c.anilistClientRef.Get().AnimeAiringScheduleRaw(ctx, ids, interceptors...)
	})
}
