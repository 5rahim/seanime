package transcoder

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog"
)

type VideoStream struct {
	Stream
	quality  Quality
	logger   *zerolog.Logger
	settings *Settings
}

func NewVideoStream(file *FileStream, quality Quality, logger *zerolog.Logger, settings *Settings) *VideoStream {
	logger.Trace().Str("file", filepath.Base(file.Path)).Any("quality", quality).Msgf("transcoder: Creating video stream")
	ret := new(VideoStream)
	ret.quality = quality
	ret.logger = logger
	ret.settings = settings
	NewStream(fmt.Sprintf("video (%s)", quality), file, ret, &ret.Stream, settings, logger)
	return ret
}

func (vs *VideoStream) getFlags() Flags {
	if vs.quality == Original {
		return VideoF | Transmux
	}
	return VideoF
}

func (vs *VideoStream) getOutPath(encoderId int) string {
	return filepath.Join(vs.file.Out, fmt.Sprintf("segment-%s-%d-%%d.ts", vs.quality, encoderId))
}

func closestMultiple(n int32, x int32) int32 {
	if x > n {
		return x
	}

	n = n + x/2
	n = n - (n % x)
	return n
}

func (vs *VideoStream) getTranscodeArgs(segments string) []string {
	args := []string{
		"-map", "0:V:0",
	}

	if vs.quality == Original {
		args = append(args,
			"-c:v", "copy",
		)
		vs.logger.Debug().Msg("videostream: Transcoding to original quality")
		return args
	}

	vs.logger.Debug().Interface("hwaccelArgs", vs.settings.HwAccel).Msg("videostream: Hardware Acceleration")

	args = append(args, vs.settings.HwAccel.EncodeFlags...)
	width := int32(float64(vs.quality.Height()) / float64(vs.file.Info.Video.Height) * float64(vs.file.Info.Video.Width))
	// force a width that is a multiple of two else some apps behave badly.
	width = closestMultiple(width, 2)
	args = append(args,
		"-vf", fmt.Sprintf(vs.settings.HwAccel.ScaleFilter, width, vs.quality.Height()),
		// Even less sure but buf size are 5x the average bitrate since the average bitrate is only
		// useful for hls segments.
		"-bufsize", fmt.Sprint(vs.quality.MaxBitrate()*5),
		"-b:v", fmt.Sprint(vs.quality.AverageBitrate()),
		"-maxrate", fmt.Sprint(vs.quality.MaxBitrate()),
	)
	if vs.settings.HwAccel.WithForcedIdr {
		// Force segments to be split exactly on keyframes (only works when transcoding)
		// forced-idr is needed to force keyframes to be an idr-frame (by default it can be any i frames)
		// without this option, some hardware encoders uses others i-frames and the -f segment can't cut at them.
		args = append(args, "-forced-idr", "1")
	}

	args = append(args,
		"-force_key_frames", segments,
		// make ffmpeg globally less buggy
		"-strict", "-2",
	)

	vs.logger.Debug().Interface("args", args).Msgf("videostream: Transcoding to %s quality", vs.quality)

	return args
}
