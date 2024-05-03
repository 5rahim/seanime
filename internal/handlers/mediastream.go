package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
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
		return c.RespondWithError(errors.New("mediastream settings not found"))
	}

	return c.RespondWithData(mediastreamSettings)
}

// HandleSaveMediastreamSettings
//
//	@summary save mediastream settings.
//	@desc This saves the mediastream settings.
//	@returns models.MediastreamSettings
//	@route /api/v1/mediastream/settings [POST]
func HandleSaveMediastreamSettings(c *RouteCtx) error {

	type body struct {
		Settings models.MediastreamSettings `json:"settings"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertMediastreamSettings(&b.Settings)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.InitOrRefreshMediastreamSettings()

	return c.RespondWithData(settings)
}

// HandleMediastreamRequestTranscodeStream
//
//	@summary request on-the-fly transcoding of a media.
//	@desc This requests on-the-fly transcoding of a media and returns the media container to start the playback.
//	@returns mediastream.MediaContainer
//	@route /api/v1/mediastream/transcode [POST]
func HandleMediastreamRequestTranscodeStream(c *RouteCtx) error {

	type body struct {
		Path string `json:"path"` // The path of the file.
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	mediaContainer, err := c.App.MediastreamRepository.RequestTranscodeStream(b.Path)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mediaContainer)
}

func HandleMediastreamTranscode(c *RouteCtx) error {
	client := "1"

	return c.App.MediastreamRepository.ServeFiberTranscodeStream(c.Fiber, client)
}

func HandleMediastreamGetTranscodeSubtitles(c *RouteCtx) error {

	return c.App.MediastreamRepository.ServeFiberTranscodeSubtitles(c.Fiber)
}

//// GetInfo Identify
////
//// Identify metadata about a file.
////
//// Path: /info
//func GetInfo(c *RouteCtx) error {
//	//path, err := GetPath(c)
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	//route :
//	sha, err := transcoder.GetHash(path)
//	if err != nil {
//		return err
//	}
//	ret, err := transcoder.GetInfo(path, c.App.Logger)
//	if err != nil {
//		return err
//	}
//	// Run extractors to have them in cache
//	transcoder.Extract(ret.Path, sha, c.App.Logger)
//	//go ExtractThumbnail(
//	//	ret.Path,
//	//	route,
//	//	sha,
//	//)
//	return c.Fiber.JSON(ret)
//}
//
//// GetAttachment Get attachments
////
//// Get a specific attachment.
////
//// Path: /attachment/:name
//func GetAttachment(c *RouteCtx) error {
//	//path, err := GetPath(c)
//	//if err != nil {
//	//	return err
//	//}
//	name := c.Fiber.Params("name")
//	if err := transcoder.SanitizePath(name); err != nil {
//		return err
//	}
//
//	//route :
//	sha, err := transcoder.GetHash(path)
//	if err != nil {
//		return err
//	}
//	wait, err := transcoder.Extract(path, sha, c.App.Logger)
//	if err != nil {
//		return err
//	}
//	<-wait
//
//	ret := fmt.Sprintf("%s/%s/att/%s", transcoder.Settings.Metadata, sha, name)
//	return c.Fiber.SendFile(ret)
//}
//
//// GetSubtitle Get subtitle
////
//// Get a specific subtitle.
////
