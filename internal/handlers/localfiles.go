package handlers

import (
	"errors"
	"fmt"
	"os"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
	"seanime/internal/library_explorer"
	"time"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
)

// HandleGetLocalFiles
//
//	@summary returns all local files.
//	@desc Reminder that local files are scanned from the library path.
//	@route /api/v1/library/local-files [GET]
//	@returns []anime.LocalFile
func (h *Handler) HandleGetLocalFiles(c echo.Context) error {

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, lfs)
}

func (h *Handler) HandleDumpLocalFilesToFile(c echo.Context) error {

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	filename := fmt.Sprintf("seanime-localfiles-%s.json", time.Now().Format("2006-01-02_15-04-05"))

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "application/json")

	jsonData, err := json.MarshalIndent(lfs, "", "  ")
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return c.Blob(200, "application/json", jsonData)
}

// HandleImportLocalFiles
//
//	@summary imports local files from the given path.
//	@desc This will import local files from the given path.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files/import [POST]
func (h *Handler) HandleImportLocalFiles(c echo.Context) error {
	type body struct {
		DataFilePath string `json:"dataFilePath"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	contentB, err := os.ReadFile(b.DataFilePath)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	var lfs []*anime.LocalFile
	if err := json.Unmarshal(contentB, &lfs); err != nil {
		return h.RespondWithError(c, err)
	}

	if len(lfs) == 0 {
		return h.RespondWithError(c, errors.New("no local files found"))
	}

	_, err = db_bridge.InsertLocalFiles(h.App.Database, lfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.Database.TrimLocalFileEntries()

	return h.RespondWithData(c, true)
}

// HandleLocalFileBulkAction
//
//	@summary performs an action on all local files.
//	@desc This will perform the given action on all local files.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [POST]
//	@returns []anime.LocalFile
func (h *Handler) HandleLocalFileBulkAction(c echo.Context) error {

	type body struct {
		Action string `json:"action"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
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
	retLfs, err := db_bridge.SaveLocalFiles(h.App.Database, lfsId, lfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, retLfs)
}

// HandleUpdateLocalFileData
//
//	@summary updates the local file with the given path.
//	@desc This will update the local file with the given path.
//	@desc The response is ignored, the client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-file [PATCH]
//	@returns []anime.LocalFile
func (h *Handler) HandleUpdateLocalFileData(c echo.Context) error {

	type body struct {
		Path     string                   `json:"path"`
		Metadata *anime.LocalFileMetadata `json:"metadata"`
		Locked   bool                     `json:"locked"`
		Ignored  bool                     `json:"ignored"`
		MediaId  int                      `json:"mediaId"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	lf, found := lo.Find(lfs, func(i *anime.LocalFile) bool {
		return i.HasSamePath(b.Path)
	})
	if !found {
		return h.RespondWithError(c, errors.New("local file not found"))
	}
	lf.Metadata = b.Metadata
	lf.Locked = b.Locked
	lf.Ignored = b.Ignored
	lf.MediaId = b.MediaId

	// Save the local files
	retLfs, err := db_bridge.SaveLocalFiles(h.App.Database, lfsId, lfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, retLfs)
}

// HandleSuperUpdateLocalFiles
//
//	@summary updates local files with the given paths.
//	@desc The client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files/super-update [PATCH]
//	@returns bool
func (h *Handler) HandleSuperUpdateLocalFiles(c echo.Context) error {

	type body struct {
		Files []*library_explorer.SuperUpdateFileOptions `json:"files"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}
	// Super update the files
	err := h.App.LibraryExplorer.SuperUpdateFiles(b.Files)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleUpdateLocalFiles
//
//	@summary updates local files with the given paths.
//	@desc The client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [PATCH]
//	@returns bool
func (h *Handler) HandleUpdateLocalFiles(c echo.Context) error {

	type body struct {
		Paths   []string `json:"paths"`
		Action  string   `json:"action"`
		MediaId int      `json:"mediaId,omitempty"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
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
	_, err = db_bridge.SaveLocalFiles(h.App.Database, lfsId, lfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDeleteLocalFiles
//
//	@summary deletes local files with the given paths.
//	@desc This will delete the local files with the given paths.
//	@desc The client should refetch the entire library collection and media entry.
//	@route /api/v1/library/local-files [DELETE]
//	@returns bool
func (h *Handler) HandleDeleteLocalFiles(c echo.Context) error {

	type body struct {
		Paths []string `json:"paths"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Delete the files
	p := pool.New().WithErrors()
	for _, path := range b.Paths {
		path := path
		p.Go(func() error {
			err := os.Remove(path)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := p.Wait(); err != nil {
		return h.RespondWithError(c, err)
	}

	// Remove the files from the list
	lfs = lo.Filter(lfs, func(i *anime.LocalFile, _ int) bool {
		return !lo.Contains(b.Paths, i.Path)
	})

	// Save the local files
	_, err = db_bridge.SaveLocalFiles(h.App.Database, lfsId, lfs)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleRemoveEmptyDirectories
//
//	@summary removes empty directories.
//	@desc This will remove empty directories in the library path.
//	@route /api/v1/library/empty-directories [DELETE]
//	@returns bool
func (h *Handler) HandleRemoveEmptyDirectories(c echo.Context) error {

	libraryPaths, err := h.App.Database.GetAllLibraryPathsFromSettings()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	for _, path := range libraryPaths {
		filesystem.RemoveEmptyDirectories(path, h.App.Logger)
	}

	return h.RespondWithData(c, true)
}
