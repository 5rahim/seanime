package handlers

import (
	"github.com/seanime-app/seanime/internal/api/listsync"
	"time"
)

// HandleDeleteListSyncCache
//
//	@summary deletes the list sync cache.
//	@desc This will delete the list sync cache and allows the client to fetch an up-to-date list sync instance in the next request.
//	@route /api/v1/filecache/cache [POST]
//	@returns bool
func HandleDeleteListSyncCache(c *RouteCtx) error {

	c.App.ListSyncCache.Delete(0)

	return c.RespondWithData(true)
}

// HandleGetListSyncAnimeDiffs
//
//	@summary returns the anime diffs from the list sync instance.
//	@desc If the instance is not cached, it will generate a new listsync.ListSync and cache them for 10 minutes
//	@route /api/v1/filecache/anime [GET]
//	@returns []listsync.AnimeDiff
func HandleGetListSyncAnimeDiffs(c *RouteCtx) error {
	// Fetch the list sync instance from the cache
	cachedLs, found := c.App.ListSyncCache.Get(0)
	if found {
		return c.RespondWithData(cachedLs.AnimeDiffs)
	}

	// Build a new list sync instance
	ls, err := listsync.BuildListSync(c.App.Database, c.App.Logger)
	if err != nil {
		return c.RespondWithData(err.Error())
	}

	// Cache the list sync for 10 minutes
	c.App.ListSyncCache.SetT(0, ls, time.Minute*10)

	return c.RespondWithData(ls.AnimeDiffs)
}

// HandleSyncAnime
//
//	@summary syncs the anime based on the provided diff kind
//	@route /api/v1/filecache/anime [POST]
//	@returns []listsync.AnimeDiff
func HandleSyncAnime(c *RouteCtx) error {

	type body struct {
		Kind listsync.AnimeDiffKind `json:"kind"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Fetch the list sync instance from the cache
	cachedLs, found := c.App.ListSyncCache.Get(0)
	if !found {
		// If nothing is cached, build a new list sync instance like HandleGetListSyncAnimeDiffs
		ls, err := listsync.BuildListSync(c.App.Database, c.App.Logger)
		if err != nil {
			return c.RespondWithError(err)
		}
		c.App.ListSyncCache.SetT(0, ls, time.Minute*10) // Cache the new instance for 10 minutes
		cachedLs = ls
	}

	// Go through the diffs and sync the anime between providers
	for _, diff := range cachedLs.AnimeDiffs {
		if diff.Kind == b.Kind {
			// Sync the anime
			if err := cachedLs.SyncAnime(diff); err != nil {
				return c.RespondWithError(err)
			}
			break
		}
	}
	c.App.ListSyncCache.SetT(0, cachedLs, time.Minute*10)

	return c.RespondWithData(cachedLs.AnimeDiffs)
}
