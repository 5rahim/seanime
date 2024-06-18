package videofile

import (
	"cmp"
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"golang.org/x/text/language"
	"gopkg.in/vansante/go-ffprobe.v2"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
}

func NewMediaInfoExtractor(fileCacher *filecache.Cacher) *MediaInfoExtractor {
	return &MediaInfoExtractor{
		fileCacher: fileCacher,
	}
}

// GetInfo returns the media information of a file.
// If the information is not in the cache, it will be extracted and saved in the cache.
func (e *MediaInfoExtractor) GetInfo(ffprobePath, path string) (mi *MediaInfo, err error) {
	hash, err := GetHashFromPath(path)
	if err != nil {
		return nil, err
	}

	bucket := filecache.NewBucket(fmt.Sprintf("mediastream_mediainfo_%s", hash), 24*7*52*time.Hour)

	// Look in the cache
	if found, _ := e.fileCacher.Get(bucket, hash, &mi); found {
		return mi, nil
	}

	// Get the media information of the file.
	mi, err = FfprobeGetInfo(ffprobePath, path, hash)
	if err != nil {
		return nil, err
	}

	// Save in the cache
	_ = e.fileCacher.Set(bucket, hash, mi)

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
		Container: cmp.Or(lo.ToPtr(data.Format.FormatName), nil),
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
			Extension: lo.ToPtr(extension),
			IsDefault: stream.Disposition.Default != 0,
			IsForced:  stream.Disposition.Forced != 0,
			Link:      link,
		}
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
			mi.MimeCodec = lo.ToPtr(fmt.Sprintf("%s; codecs=\"%s\"", container, codecsStr))
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//func getInfo(ffprobePath, path, hash string) (*MediaInfo, error) {
//
//	// Open file
//	file, err := os.Open(path)
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//	// Get file info
//	fInfo, err := file.Stat()
//	if err != nil {
//		return nil, err
//	}
//
//	// Matroska scanner
//	scanner, err := matroska.NewScanner(file)
//	if err != nil {
//		return nil, err
//	}
//	info := scanner.Info()
//	tracks := scanner.Tracks()
//
//	// Get extension (e.g., ".mkv")
//	ext := filepath.Ext(path)[1:]
//	// Get size (in bytes)
//	size := uint64(fInfo.Size())
//	// Get duration (in seconds)
//	duration := 0.0
//	if info.Duration != nil {
//		duration = *info.Duration / 1000
//	}
//	// Estimate bitrate
//	bitrate := 0.0
//	if duration > 0 {
//		bitrate = float64(size) / duration
//	}
//
//	mi := &MediaInfo{
//		Sha:       hash,
//		Path:      path,
//		Extension: ext,
//		Size:      size,
//		Duration:  float32(duration),
//		Container: lo.ToPtr("matroska"),
//	}
//
//	//
//	// Get advanced info \/
//	//
//
//	videos := make([]Video, 0)
//	audios := make([]Audio, 0)
//	subtitles := make([]Subtitle, 0)
//	// Go through track entries
//	audioIndex := 0
//	subtitleIndex := 0
//
//	profile, level, bitrate2, _ := getProfileAndLevel(ffprobePath, path)
//	if bitrate2 != "" {
//		bitrate, _ = strconv.ParseFloat(bitrate2, 64)
//	}
//
//	for _, entry := range tracks.TrackEntry {
//		//
//		// Video
//		//
//
//		if entry.TrackType == matroska.TrackTypeVideo {
//			v := &Video{
//				Codec:     entry.CodecID,
//				MimeCodec: matroskaToRFC6381(entry.CodecID, profile, level, nil),
//				Width:     uint32(entry.Video.PixelWidth),
//				Height:    uint32(entry.Video.PixelHeight),
//				Bitrate:   uint32(bitrate),
//				Language:  &entry.Language,
//				Quality:   GetQualityFromHeight(uint32(entry.Video.PixelHeight)),
//			}
//			videos = append(videos, *v)
//		}
//		//
//		// Audio
//		//
//		if entry.TrackType == matroska.TrackTypeAudio {
//			a := &Audio{
//				Title:     entry.Name,
//				Index:     uint32(audioIndex),
//				Codec:     entry.CodecID,
//				MimeCodec: matroskaToRFC6381(entry.CodecID, profile, level, entry.Audio.BitDepth),
//				IsDefault: entry.FlagDefault == 1,
//				IsForced:  entry.FlagForced == 1,
//				Language:  &entry.Language,
//				Channels:  uint32(entry.Audio.Channels),
//			}
//			audios = append(audios, *a)
//			audioIndex += 1
//		}
//		//
//		// Subtitles
//		//
//		if entry.TrackType == matroska.TrackTypeSubtitle {
//			subExt := guessSubtitleExt(entry.CodecID)
//			var link *string
//			if subExt != "" {
//				subExt = subExt[1:] // remove the dot
//				x := fmt.Sprintf("/%d.%s", subtitleIndex, subExt)
//				link = &x
//			}
//			s := &Subtitle{
//				Index:     uint32(subtitleIndex),
//				Title:     entry.Name,
//				Codec:     entry.CodecID,
//				IsDefault: entry.FlagDefault == 1,
//				IsForced:  entry.FlagForced == 1,
//				Extension: &subExt,
//				Language:  &entry.Language,
//				Link:      link,
//			}
//			subtitles = append(subtitles, *s)
//			subtitleIndex += 1
//		}
//	}
//
//	mi.Videos = videos
//	if len(videos) > 0 {
//		mi.Video = &videos[0]
//	}
//	mi.Audios = audios
//	mi.Subtitles = subtitles
//
//	// Close file
//	file.Close()
//
//	//
//	// Get chapters & fonts \/
//	//
//
//	chapters := make([]Chapter, 0)
//	fonts := make([]string, 0)
//
//	// Reopen file
//	file, err = os.Open(path)
//	if err != nil {
//		return nil, err
//	}
//
//	d := ebml.NewDecoder(file) // Use ebml decoder since scanner doesn't have the fields
//
//	_, err = d.DecodeHeader()
//	if err != nil {
//		return nil, err
//	}
//
//	var b matroska.Segment
//	if err = d.DecodeBody(&b); err != nil && err != io.EOF {
//		return nil, err
//	}
//	if b.Attachments != nil {
//		for _, a := range b.Attachments.AttachedFile {
//			if strings.Contains(a.FileMediaType, "font") {
//				fonts = append(fonts, a.FileName)
//			}
//		}
//	}
//	if b.Chapters != nil {
//		for _, c := range b.Chapters.EditionEntry {
//			chs := c.ChapterAtom
//
//			for _, ch := range chs { // Go through chapter atoms
//				startTime := float32(ch.ChapterTimeStart) / 1e9
//				chapter := Chapter{
//					StartTime: startTime,
//					EndTime:   0, // We don't have the end time, this will be set by the next chapter
//				}
//				// Get chapter name
//				if ch.ChapterDisplay != nil {
//					for _, d := range ch.ChapterDisplay {
//						if d.ChapString != "" {
//							chapter.Name = d.ChapString
//						}
//					}
//				}
//				chapters = append(chapters, chapter)
//			}
//		}
//	}
//	// Set end time for chapters
//	for i := 0; i <= len(chapters)-1; i++ {
//		if i == len(chapters)-1 { // Last chapter
//			chapters[i].EndTime = float32(duration)
//			break
//		}
//		chapters[i].EndTime = chapters[i+1].StartTime - 1 // Set end time to the start time of the next chapter - 1
//	}
//
//	mi.Chapters = chapters
//	mi.Fonts = fonts
//
//	return mi, nil
//}
//
//func guessSubtitleExt(codecID string) string {
//	switch codecID {
//	// Audio
//	case matroska.AudioCodecAAC:
//		return ".aac"
//	case matroska.AudioCodecAC3:
//		return ".ac3"
//	case matroska.AudioCodecMP3:
//		return ".mp3"
//	// Video
//	case matroska.VideoCodecMSCOMP:
//		return ".avi"
//	// Subtitle
//	case matroska.SubtitleCodecTEXTASS:
//		return ".ass"
//	case matroska.SubtitleCodecTEXTSSA:
//		return ".ssa"
//	case matroska.SubtitleCodecTEXTUTF8, matroska.SubtitleCodecTEXTASCII:
//		return ".srt"
//	case matroska.SubtitleCodecVOBSUB, matroska.SubtitleCodecVOBSUBZLIB:
//		return ".idx"
//	case matroska.SubtitleCodecTEXTWEBVTT:
//		return ".vtt"
//	default:
//		return ""
//	}
//}
//
//// getProfileAndLevel
//// ref: https://stackoverflow.com/a/36317694
//func getProfileAndLevel(ffprobePath, fp string) (profile string, level string, bitrate string, err error) {
//	defer func() {
//		if r := recover(); r != nil {
//			err = fmt.Errorf("panic: %v", r)
//			return
//		}
//	}()
//
//	args := []string{"-v", "error", "-select_streams", "v:0", "-show_entries", "stream=profile,level,bit_rate ", "-of", "default=noprint_wrappers=1", fp}
//
//	// Execute ffprobe
//	cmd := exec.Command(ffprobePath, args...)
//	out, err := cmd.CombinedOutput()
//	if err != nil {
//		return "", "", "", err
//	}
//
//	// Parse the output.
//	for _, line := range strings.Split(string(out), "\n") {
//		if strings.HasPrefix(line, "profile=") && len(strings.Split(line, "=")) > 1 {
//			profile = strings.Split(line, "=")[1]
//		}
//		if strings.HasPrefix(line, "level=") && len(strings.Split(line, "=")) > 1 {
//			level = strings.Split(line, "=")[1]
//		}
//		if strings.HasPrefix(line, "bit_rate=") && len(strings.Split(line, "=")) > 1 {
//			bitrateStr := strings.Split(line, "=")[1]
//			_, errB := strconv.Atoi(bitrateStr)
//			if errB == nil {
//				bitrate = bitrateStr
//			}
//		}
//	}
//
//	return
//}
//
//func matroskaToRFC6381(matroskaCodecID string, profile string, level string, bitDepth *uint) *string {
//	ret := ""
//	switch matroskaCodecID {
//	case "V_MPEG4/ISO/AVC":
//		ret = "avc1"
//		switch strings.ToLower(profile) {
//		case "baseline":
//			ret += ".42E0"
//		case "main":
//			ret += ".4D40"
//		case "high":
//			ret += ".6400"
//		default:
//			ret += ".4240"
//		}
//
//		lvl := fmt.Sprintf("%02x", level)
//		ret += lvl
//	case "V_MPEG4/ISO/HEVC":
//		ret = "hvc1"
//		switch strings.ToLower(profile) {
//		case "main", "main 10":
//			ret += ".2.4"
//		default:
//			ret += ".1.4"
//		}
//		lvlI, _ := strconv.Atoi(level)
//		lvl := fmt.Sprintf(".L%02x.BO", lvlI*3)
//		ret += lvl
//	case "V_AV1":
//		ret = "av01"
//		switch strings.ToLower(profile) {
//		case "main":
//			ret += ".0"
//		case "high":
//			ret += ".1"
//		case "professional":
//			ret += ".2"
//		default:
//			ret += ".0"
//		}
//
//		lvlI := 19
//		if level != "" {
//			lvlI, _ = strconv.Atoi(level)
//			if lvlI <= 0 || lvlI > 31 {
//				lvlI = 19
//			}
//		}
//
//		bd := 8
//		if bitDepth != nil {
//			if *bitDepth == 8 || *bitDepth == 10 || *bitDepth == 12 {
//				bd = int(*bitDepth)
//			}
//		}
//
//		lvl := fmt.Sprintf(".%02X%c.%02d", level, 'M', bd)
//		ret += lvl
//	case "A_AAC":
//		ret = "mp4a"
//
//		switch strings.ToLower(profile) {
//		case "lc":
//			ret += ".40.2"
//		case "he":
//			ret += ".40.5"
//		default:
//			ret += ".40.2"
//		}
//
//	case "A_VORBIS":
//		ret = "vorbis"
//	case "A_OPUS":
//		ret = "Opus"
//	case "A_AC3":
//		ret = "mp4a.a5"
//	case "A_TRUEHD":
//		ret = "mlp"
//	case "A_EAC3":
//		ret = "ec-3"
//	case "A_FLAC":
//		ret = "fLaC"
//	}
//
//	if ret == "" {
//		return nil
//	}
//
//	return &ret
//}
