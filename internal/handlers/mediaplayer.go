package handlers

// HandlePlayVideo will play the video with the given path with the default media player.
// It returns nil.
//
// It also starts tracking the progress of the video by launching a goroutine.
//
//	POST /v1/media-player/play
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

// HandleStartDefaultMediaPlayer will launch the default media player.
//
//	POST /v1/media-player/start
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
