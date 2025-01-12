package mediastream

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/url"
	"seanime/internal/events"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Direct
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ServeEchoFile(c echo.Context, filePath string, clientId string) error {
	filePath, _ = url.QueryUnescape(filePath)
	return c.File(filePath)
}

func (r *Repository) ServeEchoDirectPlay(c echo.Context, clientId string) error {

	if !r.IsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Module not initialized")
		return errors.New("module not initialized")
	}

	// Get current media
	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
	if !found {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "no file has been loaded")
		return errors.New("no file has been loaded")
	}

	return c.File(mediaContainer.Filepath)
}
