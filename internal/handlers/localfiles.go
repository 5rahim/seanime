package handlers

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/library/entities"
	"github.com/seanime-app/seanime/internal/library/filesystem"
	"github.com/sourcegraph/conc/pool"
	"os"
)

// HandleGetLocalFiles will return all local files.
//
//	GET /v1/local-files
func HandleGetLocalFiles(c *RouteCtx) error {

	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleLocalFileBulkAction will perform an action on all local files.
// It is used by the [Bulk Action Modal] feature.
// It returns the updated local files.
//
//	POST /v1/library/local-files
func HandleLocalFileBulkAction(c *RouteCtx) error {

	type body struct {
		Action string `json:"action"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := c.App.Database.GetLocalFiles()
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
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleUpdateLocalFileData will update the metadata of the local file with the given path.
// It is used by the [Local File Metadata Editor] feature.
// It returns the updated local files.
//
//	PATCH /v1/library/local-files
func HandleUpdateLocalFileData(c *RouteCtx) error {

	type body struct {
		Path     string                      `json:"path"`
		Metadata *entities.LocalFileMetadata `json:"metadata"`
		Locked   bool                        `json:"locked"`
		Ignored  bool                        `json:"ignored"`
		MediaId  int                         `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	lf, found := lo.Find(lfs, func(i *entities.LocalFile) bool {
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
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleDeleteLocalFiles will delete local files with the given paths.
//
//	DELETE /v1/library/local-files
func HandleDeleteLocalFiles(c *RouteCtx) error {

	type body struct {
		Paths []string `json:"paths"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := c.App.Database.GetLocalFiles()
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
	newLfs := lo.Filter(lfs, func(i *entities.LocalFile, _ int) bool {
		return !lo.Contains(deletedPaths, i.Path)
	})

	// Save the local files
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, newLfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleRemoveEmptyDirectories will remove empty directories from the library path.
//
//	DELETE /v1/library/empty-directories
func HandleRemoveEmptyDirectories(c *RouteCtx) error {

	libraryPath, err := c.App.Database.GetLibraryPathFromSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	filesystem.RemoveEmptyDirectories(libraryPath, c.App.Logger)

	return c.RespondWithData(nil)

}
