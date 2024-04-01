package handlers

import (
	"github.com/seanime-app/seanime/internal/util"
	"strings"
)

// HandleGetFileCacheTotalSize will return the total size of the file cache.
//
//	POST /api/v1/filecache/total-size
func HandleGetFileCacheTotalSize(c *RouteCtx) error {
	// Get the cache size
	size, err := c.App.FileCacher.GetTotalSize(func(filename string) bool {
		return true
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	// Return the cache size
	return c.RespondWithData(util.ToHumanReadableSize(size))
}

// HandleRemoveFileCacheBucket will remove all cache files associated with the given bucket.
//
//	DELETE /api/v1/filecache/bucket
func HandleRemoveFileCacheBucket(c *RouteCtx) error {

	type body struct {
		Bucket string `json:"bucket"`
	}

	// Parse the request body
	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Remove all files in the cache directory that match the given filter
	err := c.App.FileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, b.Bucket)
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	// Return a success response
	return c.RespondWithData(true)
}
