package handlers

import (
	"github.com/seanime-app/seanime/internal/listsync"
	"time"
)

// HandleDeleteListSyncCache
// Will delete the listsync.ListSync cached instance.
// This will allow the client to get a fresh and up-to-date instance.
// (e.g. The diffs should be re-fetch after calling this endpoint)
func HandleDeleteListSyncCache(c *RouteCtx) error {

	c.App.ListSyncCache.Delete(0)

	return c.RespondWithData(true)
}

func HandleGetListSyncAnimeDiffs(c *RouteCtx) error {
	// Fetch the list sync instance from the cache
	cachedLs, found := c.App.ListSyncCache.Get(0)
	if found {
		return c.RespondWithData(cachedLs.AnimeDiffs)
	}

	ls, err := listsync.BuildListSync(c.App.Database, c.App.Logger)
	if err != nil {
		return c.RespondWithData(err.Error())
	}

	// Cache the list sync for 10 minutes
	c.App.ListSyncCache.SetT(0, ls, time.Minute*10)

	return c.RespondWithData(ls.AnimeDiffs)
}

// HandleSyncAnime
// POST /v1/list-sync/anime
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
		ls, err := listsync.BuildListSync(c.App.Database, c.App.Logger)
		if err != nil {
			return c.RespondWithError(err)
		}
		c.App.ListSyncCache.SetT(0, ls, time.Minute*10)
		cachedLs = ls
	}

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
