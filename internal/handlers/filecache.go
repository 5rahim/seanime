package handlers

import (
	"github.com/dustin/go-humanize"
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
	size, err := c.App.FileCacher.GetTotalSize()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Return the cache size
	return c.RespondWithData(humanize.Bytes(uint64(size)))
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
		Bucket string `json:"bucket"` // e.g. "onlinestream_"
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

// HandleGetFileCacheMediastreamVideoFilesTotalSize
//
//	@summary returns the total size of cached video file data.
//	@desc The total size of the cache video file data is returned in human-readable format.
//	@route /api/v1/filecache/mediastream/videofiles/total-size [GET]
//	@returns bool
func HandleGetFileCacheMediastreamVideoFilesTotalSize(c *RouteCtx) error {
	// Get the cache size
	size, err := c.App.FileCacher.GetMediastreamVideoFilesTotalSize()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Return the cache size
	return c.RespondWithData(humanize.Bytes(uint64(size)))
}

// HandleClearFileCacheMediastreamVideoFiles
//
//	@summary deletes the contents of the mediastream video file cache directory.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/filecache/mediastream/videofiles [DELETE]
//	@returns bool
func HandleClearFileCacheMediastreamVideoFiles(c *RouteCtx) error {

	err := c.App.FileCacher.ClearMediastreamVideoFiles()

	if err != nil {
		return c.RespondWithError(err)
	}

	if c.App.MediastreamRepository != nil {
		go c.App.MediastreamRepository.CacheWasCleared()
	}

	// Return a success response
	return c.RespondWithData(true)
}
