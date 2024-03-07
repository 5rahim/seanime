package handlers

import (
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

	//// Play the video
	//err := c.App.MediaPlayRepository.Play(b.Path)
	//if err != nil {
	//	return c.RespondWithError(err)
	//}

	err := c.App.PlaybackManager.StartPlayingUsingMediaPlayer(b.Path)
	if err != nil {
		return c.RespondWithError(err)
	}
	//c.App.MediaPlayRepository.StartTracking(func() {
	//	// Send a progress update request to the client
	//	// Progress will be automatically updated without having to confirm it when you watch 90% of an episode.
	//	// This is enabled on the settings page.
	//	if settings.Library.AutoUpdateProgress {
	//		c.App.WSEventManager.SendEvent(events.MediaPlayerProgressUpdateRequest, nil)
	//		c.App.Logger.Debug().Msg("media player: Automatic progress update requested")
	//	}
	//})

	return nil
}

// HandleMpvDetectPlayback will detect playback with MPV and start tracking the progress of the video.
func HandleMpvDetectPlayback(c *RouteCtx) error {

	// Detect playback with MPV
	err := c.App.MediaPlayRepository.Mpv.OpenAndPlay("", mpv.StartDetectPlayback)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.MediaPlayRepository.StartTracking()
	//c.App.MediaPlayRepository.StartTracking(func() {
	//	// Send a progress update request to the client
	//	// Progress will be automatically updated without having to confirm it when you watch 90% of an episode.
	//	// This is enabled on the settings page.
	//	if settings.Library.AutoUpdateProgress {
	//		c.App.WSEventManager.SendEvent(events.MediaPlayerProgressUpdateRequest, nil)
	//		c.App.Logger.Debug().Msg("media player: Automatic progress update requested")
	//	}
	//})

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
