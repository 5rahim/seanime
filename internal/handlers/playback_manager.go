package handlers

import (
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/playbackmanager"
)

// HandlePlaybackPlayVideo
//
//	@summary plays the video with the given path using the default media player.
//	@desc This tells the Playback Manager to play the video using the default media player and start tracking progress.
//	@desc This returns 'true' if the video was successfully played.
//	@route /api/v1/playback-manager/play [POST]
//	@returns bool
func HandlePlaybackPlayVideo(c *RouteCtx) error {
	type body struct {
		Path string `json:"path"`
	}
	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.PlaybackManager.StartPlayingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
		Payload:   b.Path,
		UserAgent: c.Fiber.Get("User-Agent"),
		ClientId:  "",
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackPlayRandomVideo
//
//	@summary plays a random, unwatched video using the default media player.
//	@desc This tells the Playback Manager to play a random, unwatched video using the media player and start tracking progress.
//	@desc It respects the user's progress data and will prioritize "current" and "repeating" media if they are many of them.
//	@desc This returns 'true' if the video was successfully played.
//	@route /api/v1/playback-manager/play-random [POST]
//	@returns bool
func HandlePlaybackPlayRandomVideo(c *RouteCtx) error {

	err := c.App.PlaybackManager.StartRandomVideo(&playbackmanager.StartRandomVideoOptions{
		UserAgent: c.Fiber.Get("User-Agent"),
		ClientId:  "",
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackSyncCurrentProgress
//
//	@summary updates the AniList progress of the currently playing media.
//	@desc This is called after 'Update progress' is clicked when watching a media.
//	@desc This route returns the media ID of the currently playing media, so the client can refetch the media entry data.
//	@route /api/v1/playback-manager/sync-current-progress [POST]
//	@returns int
func HandlePlaybackSyncCurrentProgress(c *RouteCtx) error {

	err := c.App.PlaybackManager.SyncCurrentProgress()
	if err != nil {
		return c.RespondWithError(err)
	}

	mId, _ := c.App.PlaybackManager.GetCurrentMediaID()

	return c.RespondWithData(mId)
}

// HandlePlaybackPlayNextEpisode
//
//	@summary plays the next episode of the currently playing media.
//	@desc This will play the next episode of the currently playing media.
//	@desc This is non-blocking so the client should prevent multiple calls until the next status is received.
//	@route /api/v1/playback-manager/next-episode [POST]
//	@returns bool
func HandlePlaybackPlayNextEpisode(c *RouteCtx) error {

	err := c.App.PlaybackManager.PlayNextEpisode()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandlePlaybackStartPlaylist
//
//	@summary starts playing a playlist.
//	@desc The client should refetch playlists.
//	@route /api/v1/playback-manager/start-playlist [POST]
//	@returns bool
func HandlePlaybackStartPlaylist(c *RouteCtx) error {

	type body struct {
		DbId uint `json:"dbId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Get playlist
	playlist, err := db_bridge.GetPlaylist(c.App.Database, b.DbId)
	if err != nil {
		return c.RespondWithError(err)
	}

	err = c.App.PlaybackManager.StartPlaylist(playlist)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackCancelCurrentPlaylist
//
//	@summary ends the current playlist.
//	@desc This will stop the current playlist. This is non-blocking.
//	@route /api/v1/playback-manager/cancel-playlist [POST]
//	@returns bool
func HandlePlaybackCancelCurrentPlaylist(c *RouteCtx) error {

	err := c.App.PlaybackManager.CancelCurrentPlaylist()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackPlaylistNext
//
//	@summary moves to the next item in the current playlist.
//	@desc This is non-blocking so the client should prevent multiple calls until the next status is received.
//	@route /api/v1/playback-manager/playlist-next [POST]
//	@returns bool
func HandlePlaybackPlaylistNext(c *RouteCtx) error {

	err := c.App.PlaybackManager.RequestNextPlaylistFile()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
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
func HandlePlaybackStartManualTracking(c *RouteCtx) error {
	type body struct {
		MediaId       int    `json:"mediaId"`
		EpisodeNumber int    `json:"episodeNumber"`
		ClientId      string `json:"clientId"`
	}
	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.PlaybackManager.StartManualProgressTracking(&playbackmanager.StartManualProgressTrackingOptions{
		ClientId:      b.ClientId,
		MediaId:       b.MediaId,
		EpisodeNumber: b.EpisodeNumber,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackCancelManualTracking
//
//	@summary cancels manual tracking of a media.
//	@desc This will stop the server from expecting progress updates for the media.
//	@route /api/v1/playback-manager/manual-tracking/cancel [POST]
//	@returns bool
func HandlePlaybackCancelManualTracking(c *RouteCtx) error {

	c.App.PlaybackManager.CancelManualProgressTracking()

	return c.RespondWithData(true)
}
