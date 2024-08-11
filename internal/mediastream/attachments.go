package mediastream

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/url"
	"os"
	"path/filepath"
	"seanime/internal/events"
	"seanime/internal/mediastream/videofile"
)

// ServeFiberExtractedSubtitles serves the extracted subtitles
func (r *Repository) ServeFiberExtractedSubtitles(fiberCtx *fiber.Ctx) error {

	if !r.IsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Module not initialized")
		return errors.New("module not initialized")
	}

	if !r.TranscoderIsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Transcoder not initialized")
		return errors.New("transcoder not initialized")
	}

	// Get the route parameters
	params := fiberCtx.AllParams()
	if len(params) == 0 {
		return errors.New("no params")
	}

	// Get the parameter group
	subFilePath := params["*1"]

	// Get current media
	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
	if !found {
		return errors.New("no file has been loaded")
	}

	retPath := videofile.GetFileSubsCacheDir(r.cacheDir, mediaContainer.Hash)

	if retPath == "" {
		return errors.New("could not find subtitles")
	}

	contentB, err := os.ReadFile(filepath.Join(retPath, subFilePath))
	if err != nil {
		return err
	}

	r.logger.Trace().Msgf("mediastream: Serving subtitles from %s", retPath)

	return fiberCtx.SendString(string(contentB))
}

// ServeFiberExtractedAttachments serves the extracted attachments
func (r *Repository) ServeFiberExtractedAttachments(fiberCtx *fiber.Ctx) error {

	if !r.IsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Module not initialized")
		return errors.New("module not initialized")
	}

	if !r.TranscoderIsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Transcoder not initialized")
		return errors.New("transcoder not initialized")
	}

	// Get the route parameters
	params := fiberCtx.AllParams()
	if len(params) == 0 {
		return errors.New("no params")
	}

	// Get the parameter group
	subFilePath := params["*1"]

	// Get current media
	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
	if !found {
		return errors.New("no file has been loaded")
	}

	retPath := videofile.GetFileAttCacheDir(r.cacheDir, mediaContainer.Hash)

	if retPath == "" {
		return errors.New("could not find subtitles")
	}

	subFilePath, _ = url.QueryUnescape(subFilePath)

	contentB, err := os.ReadFile(filepath.Join(retPath, subFilePath))
	if err != nil {
		return err
	}

	r.logger.Trace().Msgf("mediastream: Serving subtitles from %s", retPath)

	return fiberCtx.SendString(string(contentB))
}
