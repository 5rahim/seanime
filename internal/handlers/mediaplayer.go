package handlers

// HandlePlayVideo
//
//	@summary plays the video with the given path using the media player.
//	@desc This tells the Playback Manager to play the video using the media player and start tracking progress.
//	@route /api/v1/media-player/play [POST]
//	@returns nil
func HandlePlayVideo(c *RouteCtx) error {

	type body struct {
		Path string `json:"path"`
	}
	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.PlaybackManager.StartPlayingUsingMediaPlayer(b.Path)
	if err != nil {
		return c.RespondWithError(err)
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------

// HandleStartDefaultMediaPlayer
//
//	@summary launches the default media player (vlc or mpc-hc).
//	@route /api/v1/media-player/start [POST]
//	@returns bool
func HandleStartDefaultMediaPlayer(c *RouteCtx) error {

	// Retrieve settings
	settings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	switch settings.MediaPlayer.Default {
	case "vlc":
		err = c.App.MediaPlayer.VLC.Start()
		if err != nil {
			return c.RespondWithError(err)
		}
	case "mpc-hc":
		err = c.App.MediaPlayer.MpcHc.Start()
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(true)

}
