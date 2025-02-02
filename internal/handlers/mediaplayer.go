package handlers

import (
	"github.com/labstack/echo/v4"
)

// HandleStartDefaultMediaPlayer
//
//	@summary launches the default media player (vlc or mpc-hc).
//	@route /api/v1/media-player/start [POST]
//	@returns bool
func (h *Handler) HandleStartDefaultMediaPlayer(c echo.Context) error {

	// Retrieve settings
	settings, err := h.App.Database.GetSettings()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	switch settings.MediaPlayer.Default {
	case "vlc":
		err = h.App.MediaPlayer.VLC.Start()
		if err != nil {
			return h.RespondWithError(c, err)
		}
	case "mpc-hc":
		err = h.App.MediaPlayer.MpcHc.Start()
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	return h.RespondWithData(c, true)
}
