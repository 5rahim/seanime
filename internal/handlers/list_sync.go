package handlers

import (
	"github.com/seanime-app/seanime/internal/api/listsync"
	"time"
)

// HandleDeleteListSyncCache will delete the listsync.ListSync cached instance.
// This will allow the client to fetch an up-to-date instance the next time it requests it.
//
//	POST /v1/list-sync/cache
func HandleDeleteListSyncCache(c *RouteCtx) error {

	c.App.ListSyncCache.Delete(0)

	return c.RespondWithData(true)
}

// HandleGetListSyncAnimeDiffs will return []*listsync.AnimeDiff from the cached listsync.ListSync instance.
// If the instance is not found, it will generate a new listsync.ListSync instance and cache them for 10 minutes in App.ListSyncCache.
//
//	GET /v1/list-sync/anime
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

// HandleSyncAnime will sync the anime based on the provided diff kind.
//
//	POST /v1/list-sync/anime
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
