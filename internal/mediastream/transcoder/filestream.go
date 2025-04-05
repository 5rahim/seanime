package transcoder

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"seanime/internal/mediastream/videofile"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// FileStream represents a stream of file data.
// It holds the keyframes, media information, video streams, and audio streams.
type FileStream struct {
	ready     sync.WaitGroup                     // A WaitGroup to synchronize go routines.
	err       error                              // An error that might occur during processing.
	Path      string                             // The path of the file.
	Out       string                             // The output path.
	Keyframes *Keyframe                          // The keyframes of the video.
	Info      *videofile.MediaInfo               // The media information of the file.
	videos    *result.Map[Quality, *VideoStream] // A map of video streams.
	audios    *result.Map[int32, *AudioStream]   // A map of audio streams.
	logger    *zerolog.Logger
	settings  *Settings
}

// NewFileStream creates a new FileStream.
func NewFileStream(
	path string,
	sha string,
	mediaInfo *videofile.MediaInfo,
	settings *Settings,
	logger *zerolog.Logger,
) *FileStream {
	ret := &FileStream{
		Path:     path,
		Out:      filepath.Join(settings.StreamDir, sha),
		videos:   result.NewResultMap[Quality, *VideoStream](),
		audios:   result.NewResultMap[int32, *AudioStream](),
		logger:   logger,
		settings: settings,
		Info:     mediaInfo,
	}

	ret.ready.Add(1)
	go func() {
		defer ret.ready.Done()
		ret.Keyframes = GetKeyframes(path, sha, logger, settings)
	}()

	return ret
}

// Kill stops all streams.
func (fs *FileStream) Kill() {
	fs.videos.Range(func(_ Quality, s *VideoStream) bool {
		s.Kill()
		return true
	})
	fs.audios.Range(func(_ int32, s *AudioStream) bool {
		s.Kill()
		return true
	})
}

// Destroy stops all streams and removes the output directory.
func (fs *FileStream) Destroy() {
	fs.logger.Debug().Msg("filestream: Destroying streams")
	fs.Kill()
	_ = os.RemoveAll(fs.Out)
}

// GetMaster generates the master playlist.
func (fs *FileStream) GetMaster() string {
	master := "#EXTM3U\n"
	if fs.Info.Video != nil {
		var transmuxQuality Quality
		for _, quality := range Qualities {
			if quality.Height() >= fs.Info.Video.Quality.Height() || quality.AverageBitrate() >= fs.Info.Video.Bitrate {
				transmuxQuality = quality
				break
			}
		}
		{
			bitrate := float64(fs.Info.Video.Bitrate)
			master += "#EXT-X-STREAM-INF:"
			master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", int(math.Min(bitrate*0.8, float64(transmuxQuality.AverageBitrate()))))
			master += fmt.Sprintf("BANDWIDTH=%d,", int(math.Min(bitrate, float64(transmuxQuality.MaxBitrate()))))
			master += fmt.Sprintf("RESOLUTION=%dx%d,", fs.Info.Video.Width, fs.Info.Video.Height)
			if fs.Info.Video.MimeCodec != nil {
				master += fmt.Sprintf("CODECS=\"%s\",", *fs.Info.Video.MimeCodec)
			}
			master += "AUDIO=\"audio\","
			master += "CLOSED-CAPTIONS=NONE\n"
			master += fmt.Sprintf("./%s/index.m3u8\n", Original)
		}
		aspectRatio := float32(fs.Info.Video.Width) / float32(fs.Info.Video.Height)
		// codec is the prefix + the level, the level is not part of the codec we want to compare for the same_codec check bellow
		transmuxPrefix := "avc1.6400"
		transmuxCodec := transmuxPrefix + "28"

		for _, quality := range Qualities {
			sameCodec := fs.Info.Video.MimeCodec != nil && strings.HasPrefix(*fs.Info.Video.MimeCodec, transmuxPrefix)
			includeLvl := quality.Height() < fs.Info.Video.Quality.Height() || (quality.Height() == fs.Info.Video.Quality.Height() && !sameCodec)

			if includeLvl {
				master += "#EXT-X-STREAM-INF:"
				master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", quality.AverageBitrate())
				master += fmt.Sprintf("BANDWIDTH=%d,", quality.MaxBitrate())
				master += fmt.Sprintf("RESOLUTION=%dx%d,", int(aspectRatio*float32(quality.Height())+0.5), quality.Height())
				master += fmt.Sprintf("CODECS=\"%s\",", transmuxCodec)
				master += "AUDIO=\"audio\","
				master += "CLOSED-CAPTIONS=NONE\n"
				master += fmt.Sprintf("./%s/index.m3u8\n", quality)
			}
		}

		//for _, quality := range Qualities {
		//	if quality.Height() < fs.Info.Video.Quality.Height() && quality.AverageBitrate() < fs.Info.Video.Bitrate {
		//		master += "#EXT-X-STREAM-INF:"
		//		master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", quality.AverageBitrate())
		//		master += fmt.Sprintf("BANDWIDTH=%d,", quality.MaxBitrate())
		//		master += fmt.Sprintf("RESOLUTION=%dx%d,", int(aspectRatio*float32(quality.Height())+0.5), quality.Height())
		//		master += "CODECS=\"avc1.640028\","
		//		master += "AUDIO=\"audio\","
		//		master += "CLOSED-CAPTIONS=NONE\n"
		//		master += fmt.Sprintf("./%s/index.m3u8\n", quality)
		//	}
		//}
	}
	for _, audio := range fs.Info.Audios {
		master += "#EXT-X-MEDIA:TYPE=AUDIO,"
		master += "GROUP-ID=\"audio\","
		if audio.Language != nil {
			master += fmt.Sprintf("LANGUAGE=\"%s\",", *audio.Language)
		}
		if audio.Title != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Title)
		} else if audio.Language != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Language)
		} else {
			master += fmt.Sprintf("NAME=\"Audio %d\",", audio.Index)
		}
		if audio.IsDefault {
			master += "DEFAULT=YES,"
		}
		master += "CHANNELS=\"2\","
		master += fmt.Sprintf("URI=\"./audio/%d/index.m3u8\"\n", audio.Index)
	}
	return master
}

// GetVideoIndex gets the index of a video stream of a specific quality.
func (fs *FileStream) GetVideoIndex(quality Quality) (string, error) {
	stream := fs.getVideoStream(quality)
	return stream.GetIndex()
}

// getVideoStream gets a video stream of a specific quality.
// It creates a new stream if it does not exist.
func (fs *FileStream) getVideoStream(quality Quality) *VideoStream {
	stream, _ := fs.videos.GetOrSet(quality, func() (*VideoStream, error) {
		return NewVideoStream(fs, quality, fs.logger, fs.settings), nil
	})
	return stream
}

// GetVideoSegment gets a segment of a video stream of a specific quality.
//func (fs *FileStream) GetVideoSegment(quality Quality, segment int32) (string, error) {
//	stream := fs.getVideoStream(quality)
//	return stream.GetSegment(segment)
//}

// GetVideoSegment gets a segment of a video stream of a specific quality.
func (fs *FileStream) GetVideoSegment(quality Quality, segment int32) (string, error) {
	streamLogger.Debug().Msgf("filestream: Retrieving video segment %d (%s)", segment, quality)
	// Debug
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	debugStreamRequest(fmt.Sprintf("video %s, segment %d", quality, segment), ctx)

	//stream := fs.getVideoStream(quality)
	//return stream.GetSegment(segment)

	// Channel to signal completion
	done := make(chan struct{})

	var ret string
	var err error

	// Execute the retrieval operation in a goroutine
	go func() {
		defer close(done)
		stream := fs.getVideoStream(quality)
		ret, err = stream.GetSegment(segment)
	}()

	// Wait for either the operation to complete or the timeout to occur
	select {
	case <-done:
		return ret, err
	case <-ctx.Done():
		return "", fmt.Errorf("filestream: timeout while retrieving video segment %d (%s)", segment, quality)
	}
}

// GetAudioIndex gets the index of an audio stream of a specific index.
func (fs *FileStream) GetAudioIndex(audio int32) (string, error) {
	stream := fs.getAudioStream(audio)
	return stream.GetIndex()
}

// GetAudioSegment gets a segment of an audio stream of a specific index.
func (fs *FileStream) GetAudioSegment(audio int32, segment int32) (string, error) {
	streamLogger.Debug().Msgf("filestream: Retrieving audio %d segment %d", audio, segment)
	// Debug
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	debugStreamRequest(fmt.Sprintf("audio %d, segment %d", audio, segment), ctx)

	stream := fs.getAudioStream(audio)
	return stream.GetSegment(segment)
}

// getAudioStream gets an audio stream of a specific index.
// It creates a new stream if it does not exist.
func (fs *FileStream) getAudioStream(audio int32) *AudioStream {
	stream, _ := fs.audios.GetOrSet(audio, func() (*AudioStream, error) {
		return NewAudioStream(fs, audio, fs.logger, fs.settings), nil
	})
	return stream
}

func debugStreamRequest(text string, ctx context.Context) {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	if debugStream {
		start := time.Now()
		ticker := time.NewTicker(2 * time.Second)
		go func() {
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					if debugStream {
						time.Sleep(2 * time.Second)
						streamLogger.Debug().Msgf("t: %s has been running for %.2f", text, time.Since(start).Seconds())
					}
				}
			}
		}()
	}
}
