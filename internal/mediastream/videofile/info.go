package videofile

import (
	"cmp"
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/text/language"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type MediaInfo struct {
	// closed if the mediainfo is ready for read. open otherwise
	ready <-chan struct{}
	// The sha1 of the video file
	Sha string `json:"sha"`
	// The internal path of the video file
	Path string `json:"path"`
	// The extension currently used to store this video file
	Extension string  `json:"extension"`
	MimeCodec *string `json:"mimeCodec"`
	// The file size of the video file
	Size uint64 `json:"size"`
	// The length of the media in seconds
	Duration float32 `json:"duration"`
	// The container of the video file of this episode
	Container *string `json:"container"`
	// The video codec and information
	Video *Video `json:"video"`
	// The list of videos if there are multiples
	Videos []Video `json:"videos"`
	// The list of audio tracks
	Audios []Audio `json:"audios"`
	// The list of subtitles tracks
	Subtitles []Subtitle `json:"subtitles"`
	// The list of fonts that can be used to display subtitles
	Fonts []string `json:"fonts"`
	// The list of chapters. See Chapter for more information
	Chapters []Chapter `json:"chapters"`
}

type Video struct {
	// The codec of this stream (defined as the RFC 6381)
	Codec string `json:"codec"`
	// RFC 6381 mime codec, e.g., "video/mp4, codecs=avc1.42E01E, mp4a.40.2"
	MimeCodec *string `json:"mimeCodec"`
	// The language of this stream (as a ISO-639-2 language code)
	Language *string `json:"language"`
	// The max quality of this video track
	Quality Quality `json:"quality"`
	// The width of the video stream
	Width uint32 `json:"width"`
	// The height of the video stream
	Height uint32 `json:"height"`
	// The average bitrate of the video in bytes/s
	Bitrate uint32 `json:"bitrate"`
}

type Audio struct {
	// The index of this track on the media
	Index uint32 `json:"index"`
	// The title of the stream
	Title *string `json:"title"`
	// The language of this stream (as a ISO-639-2 language code)
	Language *string `json:"language"`
	// The codec of this stream
	Codec     string  `json:"codec"`
	MimeCodec *string `json:"mimeCodec"`
	// Is this stream the default one of its type?
	IsDefault bool `json:"isDefault"`
	// Is this stream tagged as forced? (useful only for subtitles)
	IsForced bool   `json:"isForced"`
	Channels uint32 `json:"channels"`
}

type Subtitle struct {
	// The index of this track on the media
	Index uint32 `json:"index"`
	// The title of the stream
	Title *string `json:"title"`
	// The language of this stream (as a ISO-639-2 language code)
	Language *string `json:"language"`
	// The codec of this stream
	Codec string `json:"codec"`
	// The extension for the codec
	Extension *string `json:"extension"`
	// Is this stream the default one of its type?
	IsDefault bool `json:"isDefault"`
	// Is this stream tagged as forced? (useful only for subtitles)
	IsForced bool `json:"isForced"`
	// Is this subtitle file external?
	IsExternal bool `json:"isExternal"`
	// The link to access this subtitle
	Link *string `json:"link"`
}

type Chapter struct {
	// The start time of the chapter (in second from the start of the episode)
	StartTime float32 `json:"startTime"`
	// The end time of the chapter (in second from the start of the episode)
	EndTime float32 `json:"endTime"`
	// The name of this chapter. This should be a human-readable name that could be presented to the user
	Name string `json:"name"`
	// TODO: add a type field for Opening, Credits...
}

type MediaInfoExtractor struct {
	fileCacher *filecache.Cacher
	logger     *zerolog.Logger
}

func NewMediaInfoExtractor(fileCacher *filecache.Cacher, logger *zerolog.Logger) *MediaInfoExtractor {
	return &MediaInfoExtractor{
		fileCacher: fileCacher,
		logger:     logger,
	}
}

// GetInfo returns the media information of a file.
// If the information is not in the cache, it will be extracted and saved in the cache.
func (e *MediaInfoExtractor) GetInfo(ffprobePath, path string) (mi *MediaInfo, err error) {
	hash, err := GetHashFromPath(path)
	if err != nil {
		return nil, err
	}

	e.logger.Debug().Str("path", path).Str("hash", hash).Msg("mediastream: Getting media information [MediaInfoExtractor]")

	bucketName := fmt.Sprintf("mediastream_mediainfo_%s", hash)
	bucket := filecache.NewBucket(bucketName, 24*7*52*time.Hour)
	e.logger.Trace().Str("bucketName", bucketName).Msg("mediastream: Using cache bucket [MediaInfoExtractor]")

	e.logger.Trace().Msg("mediastream: Getting media information from cache [MediaInfoExtractor]")

	// Look in the cache
	if found, _ := e.fileCacher.Get(bucket, hash, &mi); found {
		e.logger.Debug().Str("hash", hash).Msg("mediastream: Media information cache HIT [MediaInfoExtractor]")
		return mi, nil
	}

	e.logger.Debug().Str("hash", hash).Msg("mediastream: Extracting media information using FFprobe")

	// Get the media information of the file.
	mi, err = FfprobeGetInfo(ffprobePath, path, hash)
	if err != nil {
		e.logger.Error().Err(err).Str("path", path).Msg("mediastream: Failed to extract media information using FFprobe")
		return nil, err
	}

	// Save in the cache
	_ = e.fileCacher.Set(bucket, hash, mi)

	e.logger.Debug().Str("hash", hash).Msg("mediastream: Extracted media information using FFprobe")

	return mi, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func FfprobeGetInfo(ffprobePath, path, hash string) (*MediaInfo, error) {

	if ffprobePath != "" {
		ffprobe.SetFFProbeBinPath(ffprobePath)
	}

	ffprobeCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	data, err := ffprobe.ProbeURL(ffprobeCtx, path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)[1:]

	sizeUint64, _ := strconv.ParseUint(data.Format.Size, 10, 64)

	mi := &MediaInfo{
		Sha:       hash,
		Path:      path,
		Extension: ext,
		Size:      sizeUint64,
		Duration:  float32(data.Format.DurationSeconds),
		Container: cmp.Or(new(data.Format.FormatName), nil),
	}

	// Get the video streams
	mi.Videos = streamToMap(data.Streams, ffprobe.StreamVideo, func(stream *ffprobe.Stream, i uint32) Video {
		lang, _ := language.Parse(stream.Tags.Language)
		bitrate, _ := strconv.ParseUint(cmp.Or(stream.BitRate, data.Format.BitRate), 10, 32)
		return Video{
			Codec:     stream.CodecName,
			MimeCodec: streamToMimeCodec(stream),
			Language:  nullIfZero(lang.String()),
			Quality:   heightToQuality(uint32(stream.Height)),
			Width:     uint32(stream.Width),
			Height:    uint32(stream.Height),
			// ffmpeg does not report bitrate in mkv files, fallback to bitrate of the whole container
			// (bigger than the result since it contains audio and other videos but better than nothing).
			Bitrate: uint32(bitrate),
		}
	})

	// Get the audio streams
	mi.Audios = streamToMap(data.Streams, ffprobe.StreamAudio, func(stream *ffprobe.Stream, i uint32) Audio {
		lang, _ := language.Parse(stream.Tags.Language)
		return Audio{
			Index:     i,
			Title:     nullIfZero(stream.Tags.Title),
			Language:  nullIfZero(lang.String()),
			Codec:     stream.CodecName,
			MimeCodec: streamToMimeCodec(stream),
			IsDefault: stream.Disposition.Default != 0,
			IsForced:  stream.Disposition.Forced != 0,
		}
	})

	// Get the subtitle streams
	mi.Subtitles = streamToMap(data.Streams, ffprobe.StreamSubtitle, func(stream *ffprobe.Stream, i uint32) Subtitle {
		subExtensions := map[string]string{
			"subrip": "srt",
			"ass":    "ass",
			"vtt":    "vtt",
			"ssa":    "ssa",
		}
		extension, ok := subExtensions[stream.CodecName]
		var link *string
		if ok {
			x := fmt.Sprintf("/%d.%s", i, extension)
			link = &x
		}
		lang, _ := language.Parse(stream.Tags.Language)
		return Subtitle{
			Index:     i,
			Title:     nullIfZero(stream.Tags.Title),
			Language:  nullIfZero(lang.String()),
			Codec:     stream.CodecName,
			Extension: new(extension),
			IsDefault: stream.Disposition.Default != 0,
			IsForced:  stream.Disposition.Forced != 0,
			Link:      link,
		}
	})

	// Remove subtitles without extensions (not supported)
	mi.Subtitles = lo.Filter(mi.Subtitles, func(item Subtitle, _ int) bool {
		if item.Extension == nil || *item.Extension == "" || item.Link == nil {
			return false
		}
		return true
	})

	// Get chapters
	mi.Chapters = lo.Map(data.Chapters, func(chapter *ffprobe.Chapter, _ int) Chapter {
		return Chapter{
			StartTime: float32(chapter.StartTimeSeconds),
			EndTime:   float32(chapter.EndTimeSeconds),
			Name:      chapter.Title(),
		}
	})

	// Get fonts
	mi.Fonts = streamToMap(data.Streams, ffprobe.StreamAttachment, func(stream *ffprobe.Stream, i uint32) string {
		filename, _ := stream.TagList.GetString("filename")
		return filename
	})

	var codecs []string
	if len(mi.Videos) > 0 && mi.Videos[0].MimeCodec != nil {
		codecs = append(codecs, *mi.Videos[0].MimeCodec)
	}
	if len(mi.Audios) > 0 && mi.Audios[0].MimeCodec != nil {
		codecs = append(codecs, *mi.Audios[0].MimeCodec)
	}
	container := mime.TypeByExtension(fmt.Sprintf(".%s", mi.Extension))
	if container != "" {
		if len(codecs) > 0 {
			codecsStr := strings.Join(codecs, ", ")
			mi.MimeCodec = new(fmt.Sprintf("%s; codecs=\"%s\"", container, codecsStr))
		} else {
			mi.MimeCodec = &container
		}
	}

	if len(mi.Videos) > 0 {
		mi.Video = &mi.Videos[0]
	}

	return mi, nil
}

func nullIfZero[T comparable](v T) *T {
	var zero T
	if v != zero {
		return &v
	}
	return nil
}

func streamToMap[T any](streams []*ffprobe.Stream, kind ffprobe.StreamType, mapper func(*ffprobe.Stream, uint32) T) []T {
	count := 0
	for _, stream := range streams {
		if stream.CodecType == string(kind) {
			count++
		}
	}
	ret := make([]T, count)

	i := uint32(0)
	for _, stream := range streams {
		if stream.CodecType == string(kind) {
			ret[i] = mapper(stream, i)
			i++
		}
	}
	return ret
}

func streamToMimeCodec(stream *ffprobe.Stream) *string {
	switch stream.CodecName {
	case "h264":
		ret := "avc1"

		switch strings.ToLower(stream.Profile) {
		case "high":
			ret += ".6400"
		case "main":
			ret += ".4D40"
		case "baseline":
			ret += ".42E0"
		default:
			// Default to constrained baseline if profile is invalid
			ret += ".4240"
		}

		ret += fmt.Sprintf("%02x", stream.Level)
		return &ret

	case "h265", "hevc":
		// The h265 syntax is a bit of a mystery at the time this comment was written.
		// This is what I've found through various sources:
		// FORMAT: [codecTag].[profile].[constraint?].L[level * 30].[UNKNOWN]
		ret := "hvc1"

		if stream.Profile == "main 10" {
			ret += ".2.4"
		} else {
			ret += ".1.4"
		}

		ret += fmt.Sprintf(".L%02X.BO", stream.Level)
		return &ret

	case "av1":
		// https://aomedia.org/av1/specification/annex-a/
		// FORMAT: [codecTag].[profile].[level][tier].[bitDepth]
		ret := "av01"

		switch strings.ToLower(stream.Profile) {
		case "main":
			ret += ".0"
		case "high":
			ret += ".1"
		case "professional":
			ret += ".2"
		default:
		}

		// not sure about this field, we want pixel bit depth
		bitdepth, _ := strconv.ParseUint(stream.BitsPerRawSample, 10, 32)
		if bitdepth != 8 && bitdepth != 10 && bitdepth != 12 {
			// Default to 8 bits
			bitdepth = 8
		}

		tierflag := 'M'
		ret += fmt.Sprintf(".%02X%c.%02d", stream.Level, tierflag, bitdepth)

		return &ret

	case "aac":
		ret := "mp4a"

		switch strings.ToLower(stream.Profile) {
		case "he":
			ret += ".40.5"
		case "lc":
			ret += ".40.2"
		default:
			ret += ".40.2"
		}

		return &ret

	case "opus":
		ret := "Opus"
		return &ret

	case "ac3":
		ret := "mp4a.a5"
		return &ret

	case "eac3":
		ret := "mp4a.a6"
		return &ret

	case "flac":
		ret := "fLaC"
		return &ret

	case "alac":
		ret := "alac"
		return &ret

	default:
		return nil
	}
}

func heightToQuality(height uint32) Quality {
	qualities := Qualities
	for _, quality := range qualities {
		if quality.Height() >= height {
			return quality
		}
	}
	return P240
}
