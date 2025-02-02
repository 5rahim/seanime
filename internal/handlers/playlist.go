package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"strconv"
)

// HandleCreatePlaylist
//
//	@summary creates a new playlist.
//	@desc This will create a new playlist with the given name and local file paths.
//	@desc The response is ignored, the client should re-fetch the playlists after this.
//	@route /api/v1/playlist [POST]
//	@returns anime.Playlist
func (h *Handler) HandleCreatePlaylist(c echo.Context) error {

	type body struct {
		Name  string   `json:"name"`
		Paths []string `json:"paths"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get the local files
	dbLfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Filter the local files
	lfs := make([]*anime.LocalFile, 0)
	for _, path := range b.Paths {
		for _, lf := range dbLfs {
			if lf.GetNormalizedPath() == util.NormalizePath(path) {
				lfs = append(lfs, lf)
				break
			}
		}
	}

	// Create the playlist
	playlist := anime.NewPlaylist(b.Name)
	playlist.SetLocalFiles(lfs)

	// Save the playlist
	if err := db_bridge.SavePlaylist(h.App.Database, playlist); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, playlist)
}

// HandleGetPlaylists
//
//	@summary returns all playlists.
//	@route /api/v1/playlists [GET]
//	@returns []anime.Playlist
func (h *Handler) HandleGetPlaylists(c echo.Context) error {

	playlists, err := db_bridge.GetPlaylists(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, playlists)
}

// HandleUpdatePlaylist
//
//	@summary updates a playlist.
//	@returns the updated playlist
//	@desc The response is ignored, the client should re-fetch the playlists after this.
//	@route /api/v1/playlist [PATCH]
//	@param id - int - true - "The ID of the playlist to update."
//	@returns anime.Playlist
func (h *Handler) HandleUpdatePlaylist(c echo.Context) error {

	type body struct {
		DbId  uint     `json:"dbId"`
		Name  string   `json:"name"`
		Paths []string `json:"paths"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get the local files
	dbLfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Filter the local files
	lfs := make([]*anime.LocalFile, 0)
	for _, path := range b.Paths {
		for _, lf := range dbLfs {
			if lf.GetNormalizedPath() == util.NormalizePath(path) {
				lfs = append(lfs, lf)
				break
			}
		}
	}

	// Recreate playlist
	playlist := anime.NewPlaylist(b.Name)
	playlist.DbId = b.DbId
	playlist.Name = b.Name
	playlist.SetLocalFiles(lfs)

	// Save the playlist
	if err := db_bridge.UpdatePlaylist(h.App.Database, playlist); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, playlist)
}

// HandleDeletePlaylist
//
//	@summary deletes a playlist.
//	@route /api/v1/playlist [DELETE]
//	@returns bool
func (h *Handler) HandleDeletePlaylist(c echo.Context) error {

	type body struct {
		DbId uint `json:"dbId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)

	}

	if err := db_bridge.DeletePlaylist(h.App.Database, b.DbId); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleGetPlaylistEpisodes
//
//	@summary returns all the local files of a playlist media entry that have not been watched.
//	@route /api/v1/playlist/episodes/{id}/{progress} [GET]
//	@param id - int - true - "The ID of the media entry."
//	@param progress - int - true - "The progress of the media entry."
//	@returns []anime.LocalFile
func (h *Handler) HandleGetPlaylistEpisodes(c echo.Context) error {

	lfs, _, err := db_bridge.GetLocalFiles(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	lfw := anime.NewLocalFileWrapper(lfs)

	// Params
	mId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}
	progress, err := strconv.Atoi(c.Param("progress"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	group, found := lfw.GetLocalEntryById(mId)
	if !found {
		return h.RespondWithError(c, errors.New("media entry not found"))
	}

	toWatch := group.GetUnwatchedLocalFiles(progress)

	return h.RespondWithData(c, toWatch)
}
