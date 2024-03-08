package handlers

// HandlePlaybackSyncCurrentProgress will update the current progress of the media player.
// This route returns the media ID of the currently playing media, so the client can refetch the media data.
//
//	POST /v1/playback-manager/sync-current-progress
func HandlePlaybackSyncCurrentProgress(c *RouteCtx) error {

	err := c.App.PlaybackManager.SyncCurrentProgress()
	if err != nil {
		return c.RespondWithError(err)
	}

	mId, _ := c.App.PlaybackManager.GetCurrentMediaID()

	return c.RespondWithData(mId)
}

// HandlePlaybackPlayNextEpisode will play the next episode of the currently playing media.
//
//	POST /v1/playback-manager/play-next
func HandlePlaybackPlayNextEpisode(c *RouteCtx) error {

	err := c.App.PlaybackManager.PlayNextEpisode()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackStartPlaylist will start playing a playlist.
// The client should:
//
//   - Refetch playlists
//
//     POST /v1/playback-manager/start-playlist
func HandlePlaybackStartPlaylist(c *RouteCtx) error {

	type body struct {
		DbId uint `json:"dbId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Get playlist
	playlist, err := c.App.Database.GetPlaylist(b.DbId)
	if err != nil {
		return c.RespondWithError(err)
	}

	err = c.App.PlaybackManager.StartPlaylist(playlist)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackCancelCurrentPlaylist will end the current playlist
//
//	POST /v1/playback-manager/cancel-playlist
func HandlePlaybackCancelCurrentPlaylist(c *RouteCtx) error {

	err := c.App.PlaybackManager.CancelCurrentPlaylist()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandlePlaybackPlaylistNext will end the current playlist
//
//	POST /v1/playback-manager/playlist-next
func HandlePlaybackPlaylistNext(c *RouteCtx) error {

	err := c.App.PlaybackManager.RequestNextPlaylistFile()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
