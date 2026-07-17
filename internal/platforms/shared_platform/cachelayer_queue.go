package shared_platform

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/util/filecache"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gqlgo/gqlgenc/clientv2"
)

const (
	queueSyncInterval  = 10 * time.Second
	queueSyncTimeout   = 20 * time.Second
	queueRetryDelay    = 15 * time.Second
	queueRetryDelayMax = 5 * time.Minute
)

type queuedMediaListUpdate struct {
	MediaID       int                      `json:"mediaId"`
	Status        *anilist.MediaListStatus `json:"status,omitempty"`
	ScoreRaw      *int                     `json:"scoreRaw,omitempty"`
	Progress      *int                     `json:"progress,omitempty"`
	StartedAt     *anilist.FuzzyDateInput  `json:"startedAt,omitempty"`
	CompletedAt   *anilist.FuzzyDateInput  `json:"completedAt,omitempty"`
	FullUpdate    bool                     `json:"fullUpdate,omitempty"`
	Attempts      int                      `json:"attempts,omitempty"`
	UpdatedAt     time.Time                `json:"updatedAt"`
	NextAttemptAt *time.Time               `json:"nextAttemptAt,omitempty"`
}

func (c *CacheLayer) startQueuedUpdateSync() {
	go func() {
		c.syncQueuedUpdates(context.Background())

		ticker := time.NewTicker(queueSyncInterval)
		defer ticker.Stop()

		for range ticker.C {
			c.syncQueuedUpdates(context.Background())
		}
	}()
}

func shouldQueueMediaListUpdate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, anilist.ErrNotAuthenticated) {
		return false
	}

	errStr := strings.ToLower(err.Error())
	if isAnilistAuthError(err) || strings.Contains(errStr, "not authenticated") {
		return false
	}
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") || strings.Contains(errStr, "404") {
		return false
	}

	if !IsWorking.Load() {
		return true
	}

	queueableParts := []string{
		"429",
		"500",
		"502",
		"503",
		"504",
		"connection",
		"deadline exceeded",
		"eof",
		"failed to decode",
		"failed to read response",
		"request failed",
		"server error",
		"timeout",
		"unexpected end",
	}
	return slices.ContainsFunc(queueableParts, func(part string) bool {
		return strings.Contains(errStr, part)
	})
}

func (c *CacheLayer) queueMediaListEntryUpdate(mediaID *int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) (int, error) {
	if mediaID == nil {
		return 0, errors.New("anilist cache: media ID is required to queue list update")
	}

	update := queuedMediaListUpdate{
		MediaID:     *mediaID,
		Status:      newCloned(status),
		ScoreRaw:    newCloned(scoreRaw),
		Progress:    newCloned(progress),
		StartedAt:   cloneFuzzyDateInput(startedAt),
		CompletedAt: cloneFuzzyDateInput(completedAt),
		FullUpdate:  true,
	}

	return c.queueMediaListUpdate(update)
}

func (c *CacheLayer) queueMediaListEntryProgressUpdate(mediaID *int, progress *int, status *anilist.MediaListStatus) (int, error) {
	if mediaID == nil {
		return 0, errors.New("anilist cache: media ID is required to queue progress update")
	}

	update := queuedMediaListUpdate{
		MediaID:  *mediaID,
		Status:   newCloned(status),
		Progress: newCloned(progress),
	}

	return c.queueMediaListUpdate(update)
}

func (c *CacheLayer) queueMediaListUpdate(update queuedMediaListUpdate) (int, error) {
	if !ShouldCache.Load() {
		return 0, errors.New("anilist cache: cache layer is disabled, list update cannot be queued")
	}

	c.pendingUpdateSyncMutex.Lock()
	defer c.pendingUpdateSyncMutex.Unlock()

	queued, err := c.saveQueuedMediaListUpdate(update)
	if err != nil {
		return 0, err
	}

	entryID, patched, err := c.applyQueuedUpdateToCache(queued)
	if err != nil {
		c.logger.Warn().Err(err).Int("mediaId", queued.MediaID).Msg("anilist cache: Failed to apply queued list update to cache")
	}
	if !patched {
		c.logger.Debug().Int("mediaId", queued.MediaID).Msg("anilist cache: Queued list update without a cached collection entry")
	}

	c.logger.Info().Int("mediaId", queued.MediaID).Msg("anilist cache: Queued list update for retry")
	return entryID, nil
}

func (c *CacheLayer) sendMediaListEntryUpdate(ctx context.Context, mediaID *int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput, interceptors ...clientv2.RequestInterceptor) (*anilist.UpdateMediaListEntry, error) {
	if mediaID == nil {
		return c.anilistClientRef.Get().UpdateMediaListEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, interceptors...)
	}

	c.pendingUpdateSyncMutex.Lock()
	defer c.pendingUpdateSyncMutex.Unlock()

	res, err := c.anilistClientRef.Get().UpdateMediaListEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, interceptors...)
	if err == nil {
		c.deleteQueuedUpdate(*mediaID)
	}
	return res, err
}

func (c *CacheLayer) sendMediaListEntryProgressUpdate(ctx context.Context, mediaID *int, progress *int, status *anilist.MediaListStatus, interceptors ...clientv2.RequestInterceptor) (*anilist.UpdateMediaListEntryProgress, error) {
	if mediaID == nil {
		return c.anilistClientRef.Get().UpdateMediaListEntryProgress(ctx, mediaID, progress, status, interceptors...)
	}

	c.pendingUpdateSyncMutex.Lock()
	defer c.pendingUpdateSyncMutex.Unlock()

	res, err := c.anilistClientRef.Get().UpdateMediaListEntryProgress(ctx, mediaID, progress, status, interceptors...)
	if err == nil {
		c.deleteQueuedUpdate(*mediaID)
	}
	return res, err
}

func (c *CacheLayer) deleteQueuedUpdate(mediaID int) {
	bucket := c.buckets[PendingMediaListUpdatesBucket]
	if err := c.fileCacher.DeletePerm(bucket, strconv.Itoa(mediaID)); err != nil {
		c.logger.Warn().Err(err).Int("mediaId", mediaID).Msg("anilist cache: Failed to delete queued list update")
	}
}

func (c *CacheLayer) saveQueuedMediaListUpdate(update queuedMediaListUpdate) (queuedMediaListUpdate, error) {
	bucket := c.buckets[PendingMediaListUpdatesBucket]
	key := strconv.Itoa(update.MediaID)
	now := time.Now()

	var current queuedMediaListUpdate
	found, err := c.fileCacher.GetPerm(bucket, key, &current)
	if err != nil {
		return queuedMediaListUpdate{}, err
	}

	if !found {
		current = queuedMediaListUpdate{
			MediaID: update.MediaID,
		}
	}

	if update.Status != nil {
		current.Status = update.Status
	}
	if update.ScoreRaw != nil {
		current.ScoreRaw = update.ScoreRaw
	}
	if update.Progress != nil {
		current.Progress = update.Progress
	}
	if update.StartedAt != nil {
		current.StartedAt = update.StartedAt
	}
	if update.CompletedAt != nil {
		current.CompletedAt = update.CompletedAt
	}
	if update.FullUpdate {
		current.FullUpdate = true
	}

	current.MediaID = update.MediaID
	current.Attempts = 0
	current.NextAttemptAt = nil
	current.UpdatedAt = now

	if err := c.fileCacher.SetPerm(bucket, key, current); err != nil {
		return queuedMediaListUpdate{}, err
	}

	return current, nil
}

func (c *CacheLayer) getQueuedMediaListUpdates() ([]queuedMediaListUpdate, error) {
	bucket := c.buckets[PendingMediaListUpdatesBucket]
	data, err := filecache.GetAll[queuedMediaListUpdate](c.fileCacher, filecache.NewBucket(bucket.Name(), 0))
	if err != nil {
		return nil, err
	}

	updates := make([]queuedMediaListUpdate, 0, len(data))
	for _, update := range data {
		updates = append(updates, update)
	}
	slices.SortFunc(updates, func(a, b queuedMediaListUpdate) int {
		return a.UpdatedAt.Compare(b.UpdatedAt)
	})

	return updates, nil
}

func (c *CacheLayer) syncQueuedUpdates(ctx context.Context) {
	if !ShouldCache.Load() || !IsWorking.Load() {
		return
	}

	c.pendingUpdateSyncMutex.Lock()
	defer c.pendingUpdateSyncMutex.Unlock()

	updates, err := c.getQueuedMediaListUpdates()
	if err != nil {
		c.logger.Warn().Err(err).Msg("anilist cache: Failed to load queued list updates")
		return
	}
	if len(updates) == 0 {
		return
	}

	now := time.Now()
	synced := 0
	for _, update := range updates {
		if update.NextAttemptAt != nil && update.NextAttemptAt.After(now) {
			continue
		}

		updateCtx, cancel := context.WithTimeout(ctx, queueSyncTimeout)
		err := c.syncQueuedUpdate(updateCtx, update)
		cancel()

		c.checkAndUpdateWorkingState(err)
		if err != nil {
			c.setQueuedUpdateSyncFailed(update, err)
			if !IsWorking.Load() {
				return
			}
			continue
		}

		if c.deleteQueuedUpdateIfCurrent(update) {
			synced++
		}
	}

	if synced > 0 {
		c.logger.Info().Int("count", synced).Msg("anilist cache: Synced queued list updates")
	}
}

func (c *CacheLayer) syncQueuedUpdate(ctx context.Context, update queuedMediaListUpdate) error {
	mediaID := update.MediaID
	if update.FullUpdate || update.ScoreRaw != nil || update.StartedAt != nil || update.CompletedAt != nil {
		_, err := c.anilistClientRef.Get().UpdateMediaListEntry(ctx, &mediaID, update.Status, update.ScoreRaw, update.Progress, update.StartedAt, update.CompletedAt)
		return err
	}

	if update.Progress == nil && update.Status == nil {
		return fmt.Errorf("queued list update for media %d has no fields", update.MediaID)
	}

	_, err := c.anilistClientRef.Get().UpdateMediaListEntryProgress(ctx, &mediaID, update.Progress, update.Status)
	return err
}

func (c *CacheLayer) setQueuedUpdateSyncFailed(update queuedMediaListUpdate, syncErr error) {
	bucket := c.buckets[PendingMediaListUpdatesBucket]
	key := strconv.Itoa(update.MediaID)

	var current queuedMediaListUpdate
	found, err := c.fileCacher.GetPerm(bucket, key, &current)
	if err != nil || !found || !sameQueuedUpdate(current, update) {
		return
	}

	current.Attempts++
	current.NextAttemptAt = new(time.Now().Add(queuedUpdateRetryDelay(current.Attempts)))
	if err := c.fileCacher.SetPerm(bucket, key, current); err != nil {
		c.logger.Warn().Err(err).Int("mediaId", update.MediaID).Msg("anilist cache: Failed to update queued list retry state")
	}
}

func (c *CacheLayer) deleteQueuedUpdateIfCurrent(update queuedMediaListUpdate) bool {
	bucket := c.buckets[PendingMediaListUpdatesBucket]
	key := strconv.Itoa(update.MediaID)

	var current queuedMediaListUpdate
	found, err := c.fileCacher.GetPerm(bucket, key, &current)
	if err != nil || !found || !sameQueuedUpdate(current, update) {
		return false
	}

	if err := c.fileCacher.DeletePerm(bucket, key); err != nil {
		c.logger.Warn().Err(err).Int("mediaId", update.MediaID).Msg("anilist cache: Failed to delete synced queued list update")
		return false
	}

	return true
}

func queuedUpdateRetryDelay(attempts int) time.Duration {
	if attempts <= 0 {
		return queueRetryDelay
	}

	delay := queueRetryDelay
	for i := 1; i < attempts; i++ {
		delay *= 2
		if delay >= queueRetryDelayMax {
			return queueRetryDelayMax
		}
	}
	return delay
}

func sameQueuedUpdate(a, b queuedMediaListUpdate) bool {
	return a.MediaID == b.MediaID && a.UpdatedAt.Equal(b.UpdatedAt)
}

func (c *CacheLayer) applyQueuedUpdateToCache(update queuedMediaListUpdate) (int, bool, error) {
	cacheKey := c.generateCacheKey("collection", nil)
	entryID := 0
	patched := false

	animeBucket := c.buckets[AnimeCollectionBucket]
	var animeCollection anilist.AnimeCollection
	found, err := c.fileCacher.GetPerm(animeBucket, cacheKey, &animeCollection)
	if err != nil {
		return 0, false, err
	}
	if found {
		if id, ok := applyQueuedUpdateToAnimeCollection(&animeCollection, update); ok {
			entryID = cmp.Or(entryID, id)
			patched = true
			if err := c.fileCacher.SetPerm(animeBucket, cacheKey, &animeCollection); err != nil {
				return entryID, patched, err
			}
		}
	}

	mangaBucket := c.buckets[MangaCollectionBucket]
	var mangaCollection anilist.MangaCollection
	found, err = c.fileCacher.GetPerm(mangaBucket, cacheKey, &mangaCollection)
	if err != nil {
		return entryID, patched, err
	}
	if found {
		if id, ok := applyQueuedUpdateToMangaCollection(&mangaCollection, update); ok {
			entryID = cmp.Or(entryID, id)
			patched = true
			if err := c.fileCacher.SetPerm(mangaBucket, cacheKey, &mangaCollection); err != nil {
				return entryID, patched, err
			}
		}
	}

	return entryID, patched, nil
}

func (c *CacheLayer) applyQueuedUpdatesToAnimeCollection(collection *anilist.AnimeCollection) bool {
	updates, err := c.getQueuedMediaListUpdates()
	if err != nil {
		c.logger.Warn().Err(err).Msg("anilist cache: Failed to overlay queued anime updates")
		return false
	}

	updated := false
	for _, update := range updates {
		if _, ok := applyQueuedUpdateToAnimeCollection(collection, update); ok {
			updated = true
		}
	}
	return updated
}

func (c *CacheLayer) applyQueuedUpdatesToMangaCollection(collection *anilist.MangaCollection) bool {
	updates, err := c.getQueuedMediaListUpdates()
	if err != nil {
		c.logger.Warn().Err(err).Msg("anilist cache: Failed to overlay queued manga updates")
		return false
	}

	updated := false
	for _, update := range updates {
		if _, ok := applyQueuedUpdateToMangaCollection(collection, update); ok {
			updated = true
		}
	}
	return updated
}

func applyQueuedUpdateToAnimeCollection(collection *anilist.AnimeCollection, update queuedMediaListUpdate) (int, bool) {
	if collection == nil || collection.MediaListCollection == nil {
		return 0, false
	}

	entryID := 0
	updated := false
	for _, list := range collection.MediaListCollection.Lists {
		if list == nil {
			continue
		}
		for _, entry := range list.GetEntries() {
			if entry == nil || entry.GetMedia() == nil || entry.GetMedia().GetID() != update.MediaID {
				continue
			}
			entryID = cmp.Or(entryID, entry.GetID())
			applyUpdateToAnimeEntry(entry, update)
			updated = true
		}
	}

	if updated {
		rearrangeCachedAnimeCollectionLists(collection)
	}
	return entryID, updated
}

func applyQueuedUpdateToMangaCollection(collection *anilist.MangaCollection, update queuedMediaListUpdate) (int, bool) {
	if collection == nil || collection.MediaListCollection == nil {
		return 0, false
	}

	entryID := 0
	updated := false
	for _, list := range collection.MediaListCollection.Lists {
		if list == nil {
			continue
		}
		for _, entry := range list.GetEntries() {
			if entry == nil || entry.GetMedia() == nil || entry.GetMedia().GetID() != update.MediaID {
				continue
			}
			entryID = cmp.Or(entryID, entry.GetID())
			applyUpdateToMangaEntry(entry, update)
			updated = true
		}
	}

	if updated {
		rearrangeCachedMangaCollectionLists(collection)
	}
	return entryID, updated
}

func applyUpdateToAnimeEntry(entry *anilist.AnimeCollection_MediaListCollection_Lists_Entries, update queuedMediaListUpdate) {
	if update.Status != nil {
		entry.Status = newCloned(update.Status)
	}
	if update.ScoreRaw != nil {
		entry.Score = new(float64(*update.ScoreRaw))
	}
	if update.Progress != nil {
		entry.Progress = newCloned(update.Progress)
	}
	if update.StartedAt != nil {
		entry.StartedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{
			Year:  newCloned(update.StartedAt.Year),
			Month: newCloned(update.StartedAt.Month),
			Day:   newCloned(update.StartedAt.Day),
		}
	}
	if update.CompletedAt != nil {
		entry.CompletedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{
			Year:  newCloned(update.CompletedAt.Year),
			Month: newCloned(update.CompletedAt.Month),
			Day:   newCloned(update.CompletedAt.Day),
		}
	}
}

func applyUpdateToMangaEntry(entry *anilist.MangaCollection_MediaListCollection_Lists_Entries, update queuedMediaListUpdate) {
	if update.Status != nil {
		entry.Status = newCloned(update.Status)
	}
	if update.ScoreRaw != nil {
		entry.Score = new(float64(*update.ScoreRaw))
	}
	if update.Progress != nil {
		entry.Progress = newCloned(update.Progress)
	}
	if update.StartedAt != nil {
		entry.StartedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{
			Year:  newCloned(update.StartedAt.Year),
			Month: newCloned(update.StartedAt.Month),
			Day:   newCloned(update.StartedAt.Day),
		}
	}
	if update.CompletedAt != nil {
		entry.CompletedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{
			Year:  newCloned(update.CompletedAt.Year),
			Month: newCloned(update.CompletedAt.Month),
			Day:   newCloned(update.CompletedAt.Day),
		}
	}
}

func rearrangeCachedAnimeCollectionLists(collection *anilist.AnimeCollection) {
	removedEntries := make([]*anilist.AnimeCollection_MediaListCollection_Lists_Entries, 0)
	for _, list := range collection.MediaListCollection.Lists {
		if list == nil || list.GetStatus() == nil || list.GetEntries() == nil {
			continue
		}

		entries := list.Entries[:0]
		for _, entry := range list.GetEntries() {
			if entry == nil || entry.GetStatus() == nil || *list.GetStatus() == *entry.GetStatus() {
				entries = append(entries, entry)
				continue
			}
			removedEntries = append(removedEntries, entry)
		}
		list.Entries = entries
	}

	for _, entry := range removedEntries {
		if entry.GetStatus() == nil {
			continue
		}
		list := getOrCreateCachedAnimeList(collection, *entry.GetStatus())
		list.Entries = append(list.Entries, entry)
	}
}

func rearrangeCachedMangaCollectionLists(collection *anilist.MangaCollection) {
	removedEntries := make([]*anilist.MangaCollection_MediaListCollection_Lists_Entries, 0)
	for _, list := range collection.MediaListCollection.Lists {
		if list == nil || list.GetStatus() == nil || list.GetEntries() == nil {
			continue
		}

		entries := list.Entries[:0]
		for _, entry := range list.GetEntries() {
			if entry == nil || entry.GetStatus() == nil || *list.GetStatus() == *entry.GetStatus() {
				entries = append(entries, entry)
				continue
			}
			removedEntries = append(removedEntries, entry)
		}
		list.Entries = entries
	}

	for _, entry := range removedEntries {
		if entry.GetStatus() == nil {
			continue
		}
		list := getOrCreateCachedMangaList(collection, *entry.GetStatus())
		list.Entries = append(list.Entries, entry)
	}
}

func getOrCreateCachedAnimeList(collection *anilist.AnimeCollection, status anilist.MediaListStatus) *anilist.AnimeCollection_MediaListCollection_Lists {
	for _, list := range collection.MediaListCollection.Lists {
		if list != nil && list.GetStatus() != nil && *list.GetStatus() == status {
			return list
		}
	}

	list := &anilist.AnimeCollection_MediaListCollection_Lists{
		Status:       newCloned(&status),
		Name:         new(string(status)),
		IsCustomList: new(false),
		Entries:      []*anilist.AnimeCollection_MediaListCollection_Lists_Entries{},
	}
	collection.MediaListCollection.Lists = append(collection.MediaListCollection.Lists, list)
	return list
}

func getOrCreateCachedMangaList(collection *anilist.MangaCollection, status anilist.MediaListStatus) *anilist.MangaCollection_MediaListCollection_Lists {
	for _, list := range collection.MediaListCollection.Lists {
		if list != nil && list.GetStatus() != nil && *list.GetStatus() == status {
			return list
		}
	}

	list := &anilist.MangaCollection_MediaListCollection_Lists{
		Status:       newCloned(&status),
		Name:         new(string(status)),
		IsCustomList: new(false),
		Entries:      []*anilist.MangaCollection_MediaListCollection_Lists_Entries{},
	}
	collection.MediaListCollection.Lists = append(collection.MediaListCollection.Lists, list)
	return list
}

func newCloned[T any](value *T) *T {
	if value == nil {
		return nil
	}
	return new(*value)
}

func cloneFuzzyDateInput(value *anilist.FuzzyDateInput) *anilist.FuzzyDateInput {
	if value == nil {
		return nil
	}
	return &anilist.FuzzyDateInput{
		Year:  newCloned(value.Year),
		Month: newCloned(value.Month),
		Day:   newCloned(value.Day),
	}
}
