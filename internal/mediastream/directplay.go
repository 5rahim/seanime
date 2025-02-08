package mediastream

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"seanime/internal/events"

	"github.com/labstack/echo/v4"
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

	if c.Request().Method == http.MethodHead {
		r.logger.Trace().Msg("mediastream: Received HEAD request for direct play")

		// Get the file size
		fileInfo, err := os.Stat(mediaContainer.Filepath)
		if err != nil {
			r.logger.Error().Msg("mediastream: Failed to get file info")
			return c.NoContent(http.StatusInternalServerError)
		}

		// Set the content length
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		c.Response().Header().Set("Content-Type", "video/mp4")
		c.Response().Header().Set("Accept-Ranges", "bytes")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", mediaContainer.Filepath))
		return c.NoContent(http.StatusOK)
	}

	return c.File(mediaContainer.Filepath)
}
