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
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/text/language"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type MediaInfo struct {
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
	// The pixel format
	PixFmt string `json:"pixFmt"`
	// The color space
	ColorSpace string `json:"colorSpace"`
	// The color transfer
	ColorTransfer string `json:"colorTransfer"`
	// The color primaries
	ColorPrimaries string `json:"colorPrimaries"`
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

// ffprobeOnce ensures the binary path is set exactly once. The go-ffprobe
// library uses a package-global variable for the path, which is racy under
// concurrent use. Setting it via sync.Once eliminates the race.
var ffprobeOnce sync.Once

func FfprobeGetInfo(ffprobePath, path, hash string) (*MediaInfo, error) {

	if ffprobePath != "" {
		ffprobeOnce.Do(func() {
			ffprobe.SetFFProbeBinPath(ffprobePath)
		})
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
			Bitrate:        uint32(bitrate),
			PixFmt:         stream.PixFmt,
			ColorSpace:     stream.ColorSpace,
			ColorTransfer:  stream.ColorTransfer,
			ColorPrimaries: stream.ColorPrimaries,
		}
	})

	// Get the audio streams
	mi.Audios = streamToMap(data.Streams, ffprobe.StreamAudio, func(stream *ffprobe.Stream, i uint32) Audio {
		lang, _ := language.Parse(stream.Tags.Language)
		// Parse channel count from the stream. This is critical for the
		// cassette package's surround audio passthrough and for advertising
		// the correct channel count in the HLS master playlist.
		channels := uint32(stream.Channels)
		if channels == 0 {
			// Fallback: parse ChannelLayout for channel count estimation.
			// Common layouts: "stereo" (2), "5.1" (6), "5.1(side)" (6), "7.1" (8)
			switch {
			case strings.Contains(stream.ChannelLayout, "7.1"):
				channels = 8
			case strings.Contains(stream.ChannelLayout, "5.1"):
				channels = 6
			case strings.Contains(stream.ChannelLayout, "stereo"):
				channels = 2
			case strings.Contains(stream.ChannelLayout, "mono"):
				channels = 1
			default:
				channels = 2 // Safe default
			}
		}
		return Audio{
			Index:     i,
			Title:     nullIfZero(stream.Tags.Title),
			Language:  nullIfZero(lang.String()),
			Codec:     stream.CodecName,
			MimeCodec: streamToMimeCodec(stream),
			IsDefault: stream.Disposition.Default != 0,
			IsForced:  stream.Disposition.Forced != 0,
			Channels:  channels,
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

	mi.MimeCodec = mediaMimeCodec(mi.Extension, mi.Videos, mi.Audios)

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
	switch strings.ToLower(stream.CodecName) {
	case "h264":
		var profile string
		switch strings.ToLower(stream.Profile) {
		case "high":
			profile = "6400"
		case "main":
			profile = "4D40"
		case "baseline":
			profile = "4200"
		case "constrained baseline":
			profile = "42E0"
		default:
			return nil
		}

		if stream.Level <= 0 || stream.Level > 255 {
			return nil
		}

		return new(fmt.Sprintf("avc1.%s%02X", profile, stream.Level))

	case "h265", "hevc":
		var profile string
		switch strings.ToLower(stream.Profile) {
		case "main":
			profile = "1.6"
		case "main 10":
			profile = "2.4"
		default:
			return nil
		}

		if stream.Level <= 0 || stream.Level > 255 {
			return nil
		}

		return new(fmt.Sprintf("hvc1.%s.L%d.B0", profile, stream.Level))

	case "av1":
		var profile string
		switch strings.ToLower(stream.Profile) {
		case "main":
			profile = "0"
		case "high":
			profile = "1"
		case "professional":
			profile = "2"
		default:
			return nil
		}

		if stream.Level < 0 || stream.Level > 23 {
			return nil
		}

		bitDepth, ok := videoBitDepth(stream)
		if !ok {
			return nil
		}

		return new(fmt.Sprintf("av01.%s.%02dM.%02d", profile, stream.Level, bitDepth))

	case "vp9":
		var profile string
		var bitDepth uint64
		switch strings.ToLower(stream.Profile) {
		case "profile 0":
			profile = "00"
			bitDepth = 8
		case "profile 1":
			profile = "01"
			bitDepth = 8
		case "profile 2":
			profile = "02"
		case "profile 3":
			profile = "03"
		default:
			return nil
		}

		if !isVP9Level(stream.Level) {
			return nil
		}

		if profile == "02" || profile == "03" {
			var ok bool
			bitDepth, ok = videoBitDepth(stream)
			if !ok || (bitDepth != 10 && bitDepth != 12) {
				return nil
			}
		}

		return new(fmt.Sprintf("vp09.%s.%02d.%02d", profile, stream.Level, bitDepth))

	case "vp8":
		return new("vp8")

	case "aac":
		switch strings.ToLower(stream.Profile) {
		case "lc":
			return new("mp4a.40.2")
		case "he", "he-aac":
			return new("mp4a.40.5")
		case "he-aacv2", "he-aac v2", "he v2":
			return new("mp4a.40.29")
		default:
			return nil
		}

	case "opus":
		return new("opus")

	case "mp3":
		return new("mp3")

	case "vorbis":
		return new("vorbis")

	case "ac3":
		return new("ac-3")

	case "eac3":
		return new("ec-3")

	case "flac":
		return new("fLaC")

	case "alac":
		return new("alac")
	default:
		return nil
	}
}

func mediaMimeCodec(extension string, videos []Video, audios []Audio) *string {
	container := containerMimeType(extension)
	if container == "" {
		return nil
	}

	codecs := make([]string, 0, 2)
	if len(videos) > 0 {
		if videos[0].MimeCodec == nil || *videos[0].MimeCodec == "" {
			return nil
		}
		codecs = append(codecs, *videos[0].MimeCodec)
	}

	if len(audios) > 0 {
		audio := &audios[0]
		for i := range audios {
			if audios[i].IsDefault {
				audio = &audios[i]
				break
			}
		}
		if audio.MimeCodec == nil || *audio.MimeCodec == "" {
			return nil
		}
		codecs = append(codecs, *audio.MimeCodec)
	}

	if len(codecs) == 0 {
		return nil
	}

	return new(fmt.Sprintf("%s; codecs=\"%s\"", container, strings.Join(codecs, ", ")))
}

func containerMimeType(extension string) string {
	extension = strings.ToLower(strings.TrimPrefix(extension, "."))
	switch extension {
	case "mkv":
		return "video/matroska"
	case "mka":
		return "audio/matroska"
	case "mk3d":
		return "video/matroska-3d"
	case "mp4", "m4v":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "mov":
		return "video/quicktime"
	case "avi":
		return "video/x-msvideo"
	case "":
		return ""
	default:
		return mime.TypeByExtension("." + extension)
	}
}

func videoBitDepth(stream *ffprobe.Stream) (uint64, bool) {
	bitDepth, err := strconv.ParseUint(stream.BitsPerRawSample, 10, 32)
	if err == nil && (bitDepth == 8 || bitDepth == 10 || bitDepth == 12) {
		return bitDepth, true
	}

	pixFmt := strings.ToLower(stream.PixFmt)
	switch {
	case strings.Contains(pixFmt, "12"):
		return 12, true
	case strings.Contains(pixFmt, "10"):
		return 10, true
	case pixFmt != "" && !strings.Contains(pixFmt, "16"):
		return 8, true
	default:
		return 0, false
	}
}

func isVP9Level(level int) bool {
	switch level {
	case 10, 11, 20, 21, 30, 31, 40, 41, 50, 51, 52, 60, 61, 62:
		return true
	default:
		return false
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
