package handlers

import (
	"errors"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"os"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
)

// HandleGetLocalFiles
//
//	@summary returns all local files.
//	@desc Reminder that local files are scanned from the library path.
//	@route /api/v1/library/local-files [GET]
//	@returns []anime.LocalFile
func HandleGetLocalFiles(c *RouteCtx) error {

	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleLocalFileBulkAction
//
//	@summary performs an action on all local files.
//	@desc This will perform the given action on all local files.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [POST]
//	@returns []anime.LocalFile
func HandleLocalFileBulkAction(c *RouteCtx) error {

	type body struct {
		Action string `json:"action"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	switch b.Action {
	case "lock":
		for _, lf := range lfs {
			// Note: Don't lock local files that are not associated with a media.
			// Else refreshing the library will ignore them.
			if lf.MediaId != 0 {
				lf.Locked = true
			}
		}
	case "unlock":
		for _, lf := range lfs {
			lf.Locked = false
		}
	}

	// Save the local files
	retLfs, err := db_bridge.SaveLocalFiles(c.App.Database, lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleUpdateLocalFileData
//
//	@summary updates the local file with the given path.
//	@desc This will update the local file with the given path.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-file [PATCH]
//	@returns []anime.LocalFile
func HandleUpdateLocalFileData(c *RouteCtx) error {

	type body struct {
		Path     string                   `json:"path"`
		Metadata *anime.LocalFileMetadata `json:"metadata"`
		Locked   bool                     `json:"locked"`
		Ignored  bool                     `json:"ignored"`
		MediaId  int                      `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	lf, found := lo.Find(lfs, func(i *anime.LocalFile) bool {
		return i.HasSamePath(b.Path)
	})
	if !found {
		return c.RespondWithError(errors.New("local file not found"))
	}
	lf.Metadata = b.Metadata
	lf.Locked = b.Locked
	lf.Ignored = b.Ignored
	lf.MediaId = b.MediaId

	// Save the local files
	retLfs, err := db_bridge.SaveLocalFiles(c.App.Database, lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleUpdateLocalFiles
//
//	@summary updates local files with the given paths.
//	@desc The client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [PATCH]
//	@returns bool
func HandleUpdateLocalFiles(c *RouteCtx) error {

	type body struct {
		Paths   []string `json:"paths"`
		Action  string   `json:"action"`
		MediaId int      `json:"mediaId,omitempty"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Update the files
	for _, path := range b.Paths {
		lf, found := lo.Find(lfs, func(i *anime.LocalFile) bool {
			return i.HasSamePath(path)
		})
		if !found {
			continue
		}
		switch b.Action {
		case "lock":
			lf.Locked = true
		case "unlock":
			lf.Locked = false
		case "ignore":
			lf.MediaId = 0
			lf.Ignored = true
			lf.Locked = false
		case "unignore":
			lf.Ignored = false
			lf.Locked = false
		case "unmatch":
			lf.MediaId = 0
			lf.Locked = false
			lf.Ignored = false
		case "match":
			lf.MediaId = b.MediaId
			lf.Locked = true
			lf.Ignored = false
		}
	}

	// Save the local files
	_, err = db_bridge.SaveLocalFiles(c.App.Database, lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleDeleteLocalFiles
//
//	@summary deletes the local file with the given paths.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [DELETE]
//	@returns []anime.LocalFile
func HandleDeleteLocalFiles(c *RouteCtx) error {

	type body struct {
		Paths []string `json:"paths"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Delete the files
	p := pool.NewWithResults[string]()
	for _, path := range b.Paths {
		p.Go(func() string {
			err = os.Remove(path)
			if err != nil {
				return ""
			}
			return path
		})
	}
	deletedPaths := p.Wait()
	deletedPaths = lo.Filter(deletedPaths, func(i string, _ int) bool {
		return i != ""
	})

	// Remove the deleted files from the local files
	newLfs := lo.Filter(lfs, func(i *anime.LocalFile, _ int) bool {
		return !lo.Contains(deletedPaths, i.Path)
	})

	// Save the local files
	retLfs, err := db_bridge.SaveLocalFiles(c.App.Database, lfsId, newLfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleRemoveEmptyDirectories
//
//	@summary deletes the empty directories from the library path.
//	@route /api/v1/library/empty-directories [DELETE]
//	@returns bool
func HandleRemoveEmptyDirectories(c *RouteCtx) error {

	libraryPath, err := c.App.Database.GetLibraryPathFromSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	filesystem.RemoveEmptyDirectories(libraryPath, c.App.Logger)

	return c.RespondWithData(true)

}
