package handlers

import (
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/playbackmanager"

	"github.com/labstack/echo/v4"
)

// HandlePlaybackPlayVideo
//
//	@summary plays the video with the given path using the default media player.
//	@desc This tells the Playback Manager to play the video using the default media player and start tracking progress.
//	@desc This returns 'true' if the video was successfully played.
//	@route /api/v1/playback-manager/play [POST]
//	@returns bool
func (h *Handler) HandlePlaybackPlayVideo(c echo.Context) error {
	type body struct {
		Path string `json:"path"`
	}
	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.PlaybackManager.StartPlayingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
		Payload:   b.Path,
		UserAgent: c.Request().Header.Get("User-Agent"),
		ClientId:  "",
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackPlayRandomVideo
//
//	@summary plays a random, unwatched video using the default media player.
//	@desc This tells the Playback Manager to play a random, unwatched video using the media player and start tracking progress.
//	@desc It respects the user's progress data and will prioritize "current" and "repeating" media if they are many of them.
//	@desc This returns 'true' if the video was successfully played.
//	@route /api/v1/playback-manager/play-random [POST]
//	@returns bool
func (h *Handler) HandlePlaybackPlayRandomVideo(c echo.Context) error {

	err := h.App.PlaybackManager.StartRandomVideo(&playbackmanager.StartRandomVideoOptions{
		UserAgent: c.Request().Header.Get("User-Agent"),
		ClientId:  "",
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackSyncCurrentProgress
//
//	@summary updates the AniList progress of the currently playing media.
//	@desc This is called after 'Update progress' is clicked when watching a media.
//	@desc This route returns the media ID of the currently playing media, so the client can refetch the media entry data.
//	@route /api/v1/playback-manager/sync-current-progress [POST]
//	@returns int
func (h *Handler) HandlePlaybackSyncCurrentProgress(c echo.Context) error {

	err := h.App.PlaybackManager.SyncCurrentProgress()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	mId, _ := h.App.PlaybackManager.GetCurrentMediaID()

	return h.RespondWithData(c, mId)
}

// HandlePlaybackPlayNextEpisode
//
//	@summary plays the next episode of the currently playing media.
//	@desc This will play the next episode of the currently playing media.
//	@desc This is non-blocking so the client should prevent multiple calls until the next status is received.
//	@route /api/v1/playback-manager/next-episode [POST]
//	@returns bool
func (h *Handler) HandlePlaybackPlayNextEpisode(c echo.Context) error {

	err := h.App.PlaybackManager.PlayNextEpisode()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackGetNextEpisode
//
//	@summary gets the next episode of the currently playing media.
//	@desc This is used by the client's autoplay feature
//	@route /api/v1/playback-manager/next-episode [GET]
//	@returns *anime.LocalFile
func (h *Handler) HandlePlaybackGetNextEpisode(c echo.Context) error {

	lf := h.App.PlaybackManager.GetNextEpisode()
	return h.RespondWithData(c, lf)
}

// HandlePlaybackAutoPlayNextEpisode
//
//	@summary plays the next episode of the currently playing media.
//	@desc This will play the next episode of the currently playing media.
//	@route /api/v1/playback-manager/autoplay-next-episode [POST]
//	@returns bool
func (h *Handler) HandlePlaybackAutoPlayNextEpisode(c echo.Context) error {

	err := h.App.PlaybackManager.AutoPlayNextEpisode()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandlePlaybackStartPlaylist
//
//	@summary starts playing a playlist.
//	@desc The client should refetch playlists.
//	@route /api/v1/playback-manager/start-playlist [POST]
//	@returns bool
func (h *Handler) HandlePlaybackStartPlaylist(c echo.Context) error {

	type body struct {
		DbId uint `json:"dbId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get playlist
	playlist, err := db_bridge.GetPlaylist(h.App.Database, b.DbId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	err = h.App.PlaybackManager.StartPlaylist(playlist)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackCancelCurrentPlaylist
//
//	@summary ends the current playlist.
//	@desc This will stop the current playlist. This is non-blocking.
//	@route /api/v1/playback-manager/cancel-playlist [POST]
//	@returns bool
func (h *Handler) HandlePlaybackCancelCurrentPlaylist(c echo.Context) error {

	err := h.App.PlaybackManager.CancelCurrentPlaylist()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackPlaylistNext
//
//	@summary moves to the next item in the current playlist.
//	@desc This is non-blocking so the client should prevent multiple calls until the next status is received.
//	@route /api/v1/playback-manager/playlist-next [POST]
//	@returns bool
func (h *Handler) HandlePlaybackPlaylistNext(c echo.Context) error {

	err := h.App.PlaybackManager.RequestNextPlaylistFile()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandlePlaybackStartManualTracking
//
//	@summary starts manual tracking of a media.
//	@desc Used for tracking progress of media that is not played through any integrated media player.
//	@desc This should only be used for trackable episodes (episodes that count towards progress).
//	@desc This returns 'true' if the tracking was successfully started.
//	@route /api/v1/playback-manager/manual-tracking/start [POST]
//	@returns bool
func (h *Handler) HandlePlaybackStartManualTracking(c echo.Context) error {
	type body struct {
		MediaId       int    `json:"mediaId"`
		EpisodeNumber int    `json:"episodeNumber"`
		ClientId      string `json:"clientId"`
	}
	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.PlaybackManager.StartManualProgressTracking(&playbackmanager.StartManualProgressTrackingOptions{
		ClientId:      b.ClientId,
		MediaId:       b.MediaId,
		EpisodeNumber: b.EpisodeNumber,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandlePlaybackCancelManualTracking
//
//	@summary cancels manual tracking of a media.
//	@desc This will stop the server from expecting progress updates for the media.
//	@route /api/v1/playback-manager/manual-tracking/cancel [POST]
//	@returns bool
func (h *Handler) HandlePlaybackCancelManualTracking(c echo.Context) error {

	h.App.PlaybackManager.CancelManualProgressTracking()

	return h.RespondWithData(c, true)
}
