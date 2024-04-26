package handlers

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
