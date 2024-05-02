package handlers

import (
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"strconv"
	"strings"
)

var path = "E:/COLLECTION/One Piece/[Erai-raws] One Piece - 1072 [1080p][Multiple Subtitle][51CB925F].mkv"

func HandleStream(c *RouteCtx) error {
	client, err := transcoder.GetClientId(c.Fiber)
	if err != nil {
		return err
	}
	//path, err := GetPath(c)
	//if err != nil {
	//	return err
	//}

	params := c.Fiber.AllParams()
	if len(params) == 0 {
		return errors.New("no params")
	}

	firstParam := params["*1"]

	// /master.m3u8
	if firstParam == "master.m3u8" {
		ret, err := c.App.Transcoder.GetMaster(path, client)
		if err != nil {
			return err
		}
		return c.Fiber.SendString(ret)
	}
	// /info
	if firstParam == "info" {
		return GetInfo(c)
	}
	c.App.Logger.Trace().Any("firstParam: ", firstParam).Msg("")
	// /:quality/index.m3u8
	if strings.HasSuffix(firstParam, "index.m3u8") && !strings.Contains(firstParam, "audio") {
		split := strings.Split(firstParam, "/")
		if len(split) != 2 {
			return errors.New("invalid index.m3u8 path")
		}
		if split[0] == "original" {
			split[0] = "480p"
		}
		quality, err := transcoder.QualityFromString(split[0])
		ret, err := c.App.Transcoder.GetVideoIndex(path, quality, client)
		if err != nil {
			return err
		}
		return c.Fiber.SendString(ret)
	}
	// /audio/:audio/index.m3u8
	if strings.HasSuffix(firstParam, "index.m3u8") && strings.Contains(firstParam, "audio") {
		split := strings.Split(firstParam, "/")
		if len(split) != 3 {
			return errors.New("invalid index.m3u8 path")
		}
		audio, err := strconv.ParseInt(split[1], 10, 32)
		ret, err := c.App.Transcoder.GetAudioIndex(path, int32(audio), client)
		if err != nil {
			return err
		}
		return c.Fiber.SendString(ret)
	}
	// /:quality/segments-:chunk.ts
	if strings.HasSuffix(firstParam, ".ts") && !strings.Contains(firstParam, "audio") {
		split := strings.Split(firstParam, "/")
		if len(split) != 2 {
			return errors.New("invalid segments-:chunk.ts path")
		}
		quality, err := transcoder.QualityFromString(split[0])
		segment, err := transcoder.ParseSegment(split[1])

		ret, err := c.App.Transcoder.GetVideoSegment(path, quality, segment, client)
		if err != nil {
			return err
		}
		return c.Fiber.SendFile(ret)
	}
	// /audio/:audio/segments-:chunk.ts
	if strings.HasSuffix(firstParam, ".ts") && strings.Contains(firstParam, "audio") {
		split := strings.Split(firstParam, "/")
		if len(split) != 3 {
			return errors.New("invalid segments-:chunk.ts path")
		}
		audio, err := strconv.ParseInt(split[1], 10, 32)
		segment, err := transcoder.ParseSegment(split[2])

		ret, err := c.App.Transcoder.GetAudioSegment(path, int32(audio), segment, client)
		if err != nil {
			return err
		}
		return c.Fiber.SendFile(ret)
	}

	return errors.New("not implemented")
}

// GetInfo Identify
//
// Identify metadata about a file.
//
// Path: /info
func GetInfo(c *RouteCtx) error {
	//path, err := GetPath(c)
	//if err != nil {
	//	return nil, err
	//}

	//route :
	sha, err := transcoder.GetHash(path)
	if err != nil {
		return err
	}
	ret, err := transcoder.GetInfo(path, c.App.Logger)
	if err != nil {
		return err
	}
	// Run extractors to have them in cache
	transcoder.Extract(ret.Path, sha, c.App.Logger)
	//go ExtractThumbnail(
	//	ret.Path,
	//	route,
	//	sha,
	//)
	return c.Fiber.JSON(ret)
}

// GetAttachment Get attachments
//
// Get a specific attachment.
//
// Path: /attachment/:name
func GetAttachment(c *RouteCtx) error {
	//path, err := GetPath(c)
	//if err != nil {
	//	return err
	//}
	name := c.Fiber.Params("name")
	if err := transcoder.SanitizePath(name); err != nil {
		return err
	}

	//route :
	sha, err := transcoder.GetHash(path)
	if err != nil {
		return err
	}
	wait, err := transcoder.Extract(path, sha, c.App.Logger)
	if err != nil {
		return err
	}
	<-wait

	ret := fmt.Sprintf("%s/%s/att/%s", transcoder.Settings.Metadata, sha, name)
	return c.Fiber.SendFile(ret)
}

// GetSubtitle Get subtitle
//
// Get a specific subtitle.
//
// Path: /subtitle/:name
func GetSubtitle(c *RouteCtx) error {
	//path, err := streamer.GetPath(c.Fiber)
	//if err != nil {
	//	return err
	//}
	name := c.Fiber.Params("name")
	if err := transcoder.SanitizePath(name); err != nil {
		return err
	}

	//route := transcoder.GetRoute(c.Fiber)
	sha, err := transcoder.GetHash(path)
	if err != nil {
		return err
	}
	wait, err := transcoder.Extract(path, sha, c.App.Logger)
	if err != nil {
		return err
	}
	<-wait

	ret := fmt.Sprintf("%s/%s/sub/%s", transcoder.Settings.Metadata, sha, name)
	return c.Fiber.SendFile(ret)
}
