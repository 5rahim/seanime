package mediastream

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"strconv"
	"strings"
)

// ServeFiberTranscodeStream serves the transcoded segments
func (r *Repository) ServeFiberTranscodeStream(fiberCtx *fiber.Ctx, clientId string) error {

	if !r.IsInitialized() {
		return errors.New("transcoding module not initialized")
	}

	if !r.TranscoderIsInitialized() {
		return errors.New("transcoder not initialized")
	}

	// Get the route parameters
	params := fiberCtx.AllParams()
	if len(params) == 0 {
		return errors.New("no params")
	}

	// Get the parameter group
	path := params["*1"]

	// Get current media
	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
	if !found {
		return errors.New("no media has been requested")
	}

	r.logger.Trace().Any("path", path).Msg("mediastream: Req")

	// /master.m3u8
	if path == "master.m3u8" {
		ret, err := r.transcoder.MustGet().GetMaster(mediaContainer.Filepath, mediaContainer.Hash, clientId)
		if err != nil {
			return err
		}
		return fiberCtx.SendString(ret)
	}

	// Video stream
	// /:quality/index.m3u8
	if strings.HasSuffix(path, "index.m3u8") && !strings.Contains(path, "audio") {
		split := strings.Split(path, "/")
		if len(split) != 2 {
			return errors.New("invalid index.m3u8 path")
		}
		quality, err := transcoder.QualityFromString(split[0])
		ret, err := r.transcoder.MustGet().GetVideoIndex(mediaContainer.Filepath, mediaContainer.Hash, quality, clientId)
		if err != nil {
			return err
		}
		return fiberCtx.SendString(ret)
	}

	// Audio stream
	// /audio/:audio/index.m3u8
	if strings.HasSuffix(path, "index.m3u8") && strings.Contains(path, "audio") {
		split := strings.Split(path, "/")
		if len(split) != 3 {
			return errors.New("invalid index.m3u8 path")
		}
		audio, err := strconv.ParseInt(split[1], 10, 32)
		ret, err := r.transcoder.MustGet().GetAudioIndex(mediaContainer.Filepath, mediaContainer.Hash, int32(audio), clientId)
		if err != nil {
			return err
		}
		return fiberCtx.SendString(ret)
	}

	// Video segment
	// /:quality/segments-:chunk.ts
	if strings.HasSuffix(path, ".ts") && !strings.Contains(path, "audio") {
		split := strings.Split(path, "/")
		if len(split) != 2 {
			return errors.New("invalid segments-:chunk.ts path")
		}
		quality, err := transcoder.QualityFromString(split[0])
		segment, err := transcoder.ParseSegment(split[1])

		ret, err := r.transcoder.MustGet().GetVideoSegment(mediaContainer.Filepath, mediaContainer.Hash, quality, segment, clientId)
		if err != nil {
			return err
		}
		return fiberCtx.SendFile(ret)
	}

	// Audio segment
	// /audio/:audio/segments-:chunk.ts
	if strings.HasSuffix(path, ".ts") && strings.Contains(path, "audio") {
		split := strings.Split(path, "/")
		if len(split) != 3 {
			return errors.New("invalid segments-:chunk.ts path")
		}
		audio, err := strconv.ParseInt(split[1], 10, 32)
		segment, err := transcoder.ParseSegment(split[2])

		ret, err := r.transcoder.MustGet().GetAudioSegment(mediaContainer.Filepath, mediaContainer.Hash, int32(audio), segment, clientId)
		if err != nil {
			return err
		}
		return fiberCtx.SendFile(ret)
	}

	return errors.New("invalid path")
}
