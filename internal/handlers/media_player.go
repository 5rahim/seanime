package handlers

import (
	"github.com/seanime-app/seanime-server/internal/mediaplayer"
)

func HandlePlayVideo(c *RouteCtx) error {

	type body struct {
		Path string `json:"path"`
	}
	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Retrieve settings
	settings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media player repository
	mediaPlayerRepo := mediaplayer.Repository{
		Logger:         c.App.Logger,
		Default:        settings.MediaPlayer.Default,
		VLC:            c.App.MediaPlayer.VLC,
		MpcHc:          c.App.MediaPlayer.MpcHc,
		WSEventManager: c.App.WSEventManager,
	}

	// Play the video
	err = mediaPlayerRepo.Play(b.Path)
	if err != nil {
		return c.RespondWithError(err)
	}

	mediaPlayerRepo.StartTracking()

	return nil
}

//----------------------------------------------------------------------------------------------------------------------

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

//----------------------------------------------------------------------------------------------------------------------
