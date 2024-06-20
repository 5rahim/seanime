package handlers

import (
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediastream"
	"os"
	"path/filepath"
)

// HandleGetMediastreamSettings
//
//	@summary get mediastream settings.
//	@desc This returns the mediastream settings.
//	@returns models.MediastreamSettings
//	@route /api/v1/mediastream/settings [GET]
func HandleGetMediastreamSettings(c *RouteCtx) error {
	mediastreamSettings, found := c.App.Database.GetMediastreamSettings()
	if !found {
		return c.RespondWithError(errors.New("media streaming settings not found"))
	}

	return c.RespondWithData(mediastreamSettings)
}

// HandleSaveMediastreamSettings
//
//	@summary save mediastream settings.
//	@desc This saves the mediastream settings.
//	@returns models.MediastreamSettings
//	@route /api/v1/mediastream/settings [PATCH]
func HandleSaveMediastreamSettings(c *RouteCtx) error {

	type body struct {
		Settings models.MediastreamSettings `json:"settings"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// Check JASSUB
	if b.Settings.TranscodeEnabled || b.Settings.PreTranscodeEnabled {
		jassubPath := filepath.Join(c.App.Config.Web.AssetDir, "/jassub/jassub-worker.js")
		if _, err := os.Stat(jassubPath); os.IsNotExist(err) {
			c.App.Logger.Error().Msgf("app: 'Media streaming' cannot be enabled, JASSUB was not located in the asset directory")
			b.Settings.TranscodeEnabled = false
			b.Settings.PreTranscodeEnabled = false
			c.App.WSEventManager.SendEvent(events.ErrorToast, "JASSUB was not located in the asset directory, transcoding has been disabled")
		}
	}

	// Check Transcode directory
	if b.Settings.TranscodeEnabled {
		transcodeDir := filepath.Clean(b.Settings.TranscodeTempDir)
		if transcodeDir == "" {
			c.App.Logger.Error().Msgf("app: 'Media streaming' cannot be enabled, transcode directory is not set")
			b.Settings.TranscodeEnabled = false
			b.Settings.PreTranscodeEnabled = false
			c.App.WSEventManager.SendEvent(events.ErrorToast, "Transcode directory is not set, transcoding has been disabled")
		}
		if !filepath.IsAbs(transcodeDir) {
			c.App.Logger.Error().Msgf("app: 'Media streaming' cannot be enabled, transcode directory is not an absolute path")
			b.Settings.TranscodeEnabled = false
			b.Settings.PreTranscodeEnabled = false
			c.App.WSEventManager.SendEvent(events.ErrorToast, "Transcode directory is not an absolute path, transcoding has been disabled")
		}
		if _, err := os.Stat(transcodeDir); os.IsNotExist(err) {
			c.App.Logger.Error().Msgf("app: 'Media streaming' cannot be enabled, transcode directory cannot be located")
			b.Settings.TranscodeEnabled = false
			b.Settings.PreTranscodeEnabled = false
			c.App.WSEventManager.SendEvent(events.ErrorToast, "Transcode directory cannot be located, transcoding has been disabled")
		}
	}

	settings, err := c.App.Database.UpsertMediastreamSettings(&b.Settings)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.InitOrRefreshMediastreamSettings()

	return c.RespondWithData(settings)
}

// HandleRequestMediastreamMediaContainer
//
//	@summary request media stream.
//	@desc This requests a media stream and returns the media container to start the playback.
//	@returns mediastream.MediaContainer
//	@route /api/v1/mediastream/request [POST]
func HandleRequestMediastreamMediaContainer(c *RouteCtx) error {

	type body struct {
		Path             string                 `json:"path"`             // The path of the file.
		StreamType       mediastream.StreamType `json:"streamType"`       // The type of stream to request.
		AudioStreamIndex int                    `json:"audioStreamIndex"` // The audio stream index to use. (unused)
		ClientId         string                 `json:"clientId"`         // The session id
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var mediaContainer *mediastream.MediaContainer
	var err error

	switch b.StreamType {
	case mediastream.StreamTypeDirect:
		mediaContainer, err = c.App.MediastreamRepository.RequestDirectPlay(b.Path, b.ClientId)
	case mediastream.StreamTypeTranscode:
		mediaContainer, err = c.App.MediastreamRepository.RequestTranscodeStream(b.Path, b.ClientId)
	case mediastream.StreamTypeOptimized:
		err = fmt.Errorf("stream type %s not implemented", b.StreamType)
		//mediaContainer, err = c.App.MediastreamRepository.RequestOptimizedStream(b.Path)
	default:
		err = fmt.Errorf("stream type %s not implemented", b.StreamType)
	}
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mediaContainer)
}

// HandlePreloadMediastreamMediaContainer
//
//	@summary preloads media stream for playback.
//	@desc This preloads a media stream by extracting the media information and attachments.
//	@returns bool
//	@route /api/v1/mediastream/preload [POST]
func HandlePreloadMediastreamMediaContainer(c *RouteCtx) error {

	type body struct {
		Path             string                 `json:"path"`             // The path of the file.
		StreamType       mediastream.StreamType `json:"streamType"`       // The type of stream to request.
		AudioStreamIndex int                    `json:"audioStreamIndex"` // The audio stream index to use.
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var err error

	switch b.StreamType {
	case mediastream.StreamTypeTranscode:
		err = c.App.MediastreamRepository.RequestPreloadTranscodeStream(b.Path)
	default:
		err = fmt.Errorf("stream type %s not implemented", b.StreamType)
	}
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

func HandleMediastreamGetSubtitles(c *RouteCtx) error {
	return c.App.MediastreamRepository.ServeFiberExtractedSubtitles(c.Fiber)
}

func HandleMediastreamGetAttachments(c *RouteCtx) error {
	return c.App.MediastreamRepository.ServeFiberExtractedAttachments(c.Fiber)
}

//
// Direct
//

func HandleMediastreamDirectPlay(c *RouteCtx) error {
	client := "1"
	return c.App.MediastreamRepository.ServeFiberDirectPlay(c.Fiber, client)
}

////
//// Direct Stream
////
//
//func HandleMediastreamDirectStream(c *RouteCtx) error {
//	client := "1"
//	return c.App.MediastreamRepository.ServeFiberDirectStream(c.Fiber, client)
//}

//
// Transcode
//

func HandleMediastreamTranscode(c *RouteCtx) error {
	client := "1"
	return c.App.MediastreamRepository.ServeFiberTranscodeStream(c.Fiber, client)
}

// HandleMediastreamShutdownTranscodeStream
//
//	@summary shuts down the transcode stream
//	@desc This requests the transcoder to shut down. It should be called when unmounting the player (playback is no longer needed).
//	@desc This will also send an events.MediastreamShutdownStream event.
//	@desc It will not return any error and is safe to call multiple times.
//	@returns bool
//	@route /api/v1/mediastream/shutdown-transcode [POST]
func HandleMediastreamShutdownTranscodeStream(c *RouteCtx) error {
	client := "1"
	c.App.MediastreamRepository.ShutdownTranscodeStream(client)
	return c.RespondWithData(true)
}
