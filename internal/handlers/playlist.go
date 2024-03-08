package handlers

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/entities"
	"path/filepath"
	"strings"
)

// HandleCreatePlaylist will create a new playlist.
// It returns the playlist
//
//	POST /v1/playlist
func HandleCreatePlaylist(c *RouteCtx) error {

	type body struct {
		Name  string   `json:"name"`
		Paths []string `json:"paths"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Get the local files
	dbLfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Filter the local files
	lfs := make([]*entities.LocalFile, 0)
	for _, path := range b.Paths {
		for _, lf := range dbLfs {
			if lf.GetNormalizedPath() == strings.ToLower(filepath.ToSlash(path)) {
				lfs = append(lfs, lf)
				break
			}
		}
	}

	// Create the playlist
	playlist := entities.NewPlaylist(b.Name)
	playlist.SetLocalFiles(lfs)

	// Save the playlist
	if err := c.App.Database.SavePlaylist(playlist); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(playlist)
}

// HandleGetPlaylists will return all playlists.
//
//	GET /v1/playlists
func HandleGetPlaylists(c *RouteCtx) error {

	playlists, err := c.App.Database.GetPlaylists()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(playlists)
}

// HandleUpdatePlaylist will update a playlist.
// It returns the updated playlist
//
//	PATCH /v1/playlist/:id
func HandleUpdatePlaylist(c *RouteCtx) error {

	type body struct {
		DbId  uint     `json:"dbId"`
		Name  string   `json:"name"`
		Paths []string `json:"paths"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Get the local files
	dbLfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Filter the local files
	lfs := make([]*entities.LocalFile, 0)
	for _, path := range b.Paths {
		for _, lf := range dbLfs {
			if lf.GetNormalizedPath() == strings.ToLower(filepath.ToSlash(path)) {
				lfs = append(lfs, lf)
				break
			}
		}
	}

	// Recreate playlist
	playlist := entities.NewPlaylist(b.Name)
	playlist.DbId = b.DbId
	playlist.Name = b.Name
	playlist.SetLocalFiles(lfs)

	// Save the playlist
	if err := c.App.Database.UpdatePlaylist(playlist); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(playlist)
}

// HandleDeletePlaylist will delete a playlist.
//
//	DELETE /v1/playlist
func HandleDeletePlaylist(c *RouteCtx) error {

	type body struct {
		DbId uint `json:"dbId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)

	}

	if err := c.App.Database.DeletePlaylist(b.DbId); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleGetPlaylistEpisodes will return all the playable local files of a playlist media entry
//
//	GET /v1/playlist/episodes/:id/:progress
func HandleGetPlaylistEpisodes(c *RouteCtx) error {

	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	lfw := entities.NewLocalFileWrapper(lfs)

	// Params
	mId, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}
	progress, err := c.Fiber.ParamsInt("progress")
	if err != nil {
		return c.RespondWithError(err)
	}

	group, found := lfw.GetLocalEntryById(mId)
	if !found {
		return c.RespondWithError(errors.New("media entry not found"))
	}

	toWatch, found := group.GetMainLocalFiles()
	if !found {
		return c.RespondWithError(errors.New("no local files found"))
	}

	toWatch = lo.Filter(toWatch, func(lf *entities.LocalFile, i int) bool {
		return lf.GetEpisodeNumber() > progress
	})

	return c.RespondWithData(toWatch)
}
