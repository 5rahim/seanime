package handlers

import (
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediaplayer"
	"github.com/seanime-app/seanime/internal/mpv"
)

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
		Mpv:            c.App.MediaPlayer.Mpv,
		WSEventManager: c.App.WSEventManager,
	}

	// Play the video
	err = mediaPlayerRepo.Play(b.Path)
	if err != nil {
		return c.RespondWithError(err)
	}

	mediaPlayerRepo.StartTracking(func() {
		// Send a progress update request to the client
		// Progress will be automatically updated without having to confirm it when you watch 90% of an episode.
		// This is enabled on the settings page.
		if settings.Library.AutoUpdateProgress {
			c.App.WSEventManager.SendEvent(events.MediaPlayerProgressUpdateRequest, nil)
			c.App.Logger.Debug().Msg("mediaplayer: Automatic progress update requested")
		}
	})

	return nil
}

// HandleMpvDetectPlayback will detect playback with MPV and start tracking the progress of the video.
func HandleMpvDetectPlayback(c *RouteCtx) error {

	// Retrieve settings
	settings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media player repository
	mediaPlayerRepo := mediaplayer.Repository{
		Logger:         c.App.Logger,
		Default:        "mpv",
		VLC:            c.App.MediaPlayer.VLC,
		MpcHc:          c.App.MediaPlayer.MpcHc,
		Mpv:            c.App.MediaPlayer.Mpv,
		WSEventManager: c.App.WSEventManager,
	}

	// Detect playback with MPV
	err = mediaPlayerRepo.Mpv.OpenAndPlay("", mpv.StartDetectPlayback)
	if err != nil {
		return c.RespondWithError(err)
	}

	mediaPlayerRepo.StartTracking(func() {
		// Send a progress update request to the client
		// Progress will be automatically updated without having to confirm it when you watch 90% of an episode.
		// This is enabled on the settings page.
		if settings.Library.AutoUpdateProgress {
			c.App.WSEventManager.SendEvent(events.MediaPlayerProgressUpdateRequest, nil)
			c.App.Logger.Debug().Msg("mediaplayer: Automatic progress update requested")
		}
	})

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
