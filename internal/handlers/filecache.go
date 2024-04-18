package handlers

import (
	"github.com/seanime-app/seanime/internal/util"
	"strings"
)

// HandleGetFileCacheTotalSize
//
//	@summary returns the total size of cache files.
//	@desc The total size of the cache files is returned in human-readable format.
//	@route /api/v1/filecache/total-size [GET]
//	@returns bool
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

// HandleRemoveFileCacheBucket
//
//	@summary deletes all buckets with the given prefix.
//	@desc The bucket value is the prefix of the cache files that should be deleted.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/filecache/bucket [DELETE]
//	@returns bool
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
