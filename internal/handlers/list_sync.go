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

func HandleGetListSyncDiffs(c *RouteCtx) error {
	// Fetch the list sync instance from the cache
	cachedLs, found := c.App.ListSyncCache.Get(0)
	if found {
		return c.RespondWithData(cachedLs.CheckDiffs())
	}

	ls, err := listsync.BuildListSync(c.App.Database)
	if err != nil {
		return c.RespondWithData(err.Error())
	}

	// Cache the list sync for 10 minutes
	c.App.ListSyncCache.SetT(0, ls, time.Minute*10)

	return c.RespondWithData(ls.CheckDiffs())
}
