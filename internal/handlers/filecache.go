package handlers

import (
	"seanime/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

// HandleGetFileCacheTotalSize
//
//	@summary returns the total size of cache files.
//	@desc The total size of the cache files is returned in human-readable format.
//	@route /api/v1/filecache/total-size [GET]
//	@returns string
func (h *Handler) HandleGetFileCacheTotalSize(c echo.Context) error {
	// Get the cache size
	size, err := h.App.FileCacher.GetTotalSize()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Return the cache size
	return h.RespondWithData(c, util.Bytes(uint64(size)))
}

// HandleRemoveFileCacheBucket
//
//	@summary deletes all buckets with the given prefix.
//	@desc The bucket value is the prefix of the cache files that should be deleted.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/filecache/bucket [DELETE]
//	@returns bool
func (h *Handler) HandleRemoveFileCacheBucket(c echo.Context) error {

	type body struct {
		Bucket string `json:"bucket"` // e.g. "onlinestream_"
	}

	// Parse the request body
	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Remove all files in the cache directory that match the given filter
	err := h.App.FileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, b.Bucket)
	})

	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Return a success response
	return h.RespondWithData(c, true)
}

// HandleGetFileCacheMediastreamVideoFilesTotalSize
//
//	@summary returns the total size of cached video file data.
//	@desc The total size of the cache video file data is returned in human-readable format.
//	@route /api/v1/filecache/mediastream/videofiles/total-size [GET]
//	@returns string
func (h *Handler) HandleGetFileCacheMediastreamVideoFilesTotalSize(c echo.Context) error {
	// Get the cache size
	size, err := h.App.FileCacher.GetMediastreamVideoFilesTotalSize()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Return the cache size
	return h.RespondWithData(c, util.Bytes(uint64(size)))
}

// HandleClearFileCacheMediastreamVideoFiles
//
//	@summary deletes the contents of the mediastream video file cache directory.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/filecache/mediastream/videofiles [DELETE]
//	@returns bool
func (h *Handler) HandleClearFileCacheMediastreamVideoFiles(c echo.Context) error {

	// Clear the attachments
	err := h.App.FileCacher.ClearMediastreamVideoFiles()

	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Clear the transcode dir
	h.App.MediastreamRepository.ClearTranscodeDir()

	if h.App.MediastreamRepository != nil {
		go h.App.MediastreamRepository.CacheWasCleared()
	}

	// Return a success response
	return h.RespondWithData(c, true)
}
