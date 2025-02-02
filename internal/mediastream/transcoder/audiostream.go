package transcoder

import (
	"fmt"
	"github.com/rs/zerolog"
	"path/filepath"
)

type AudioStream struct {
	Stream
	index    int32
	logger   *zerolog.Logger
	settings *Settings
}

// NewAudioStream creates a new AudioStream for a file, at a given audio index.
func NewAudioStream(file *FileStream, idx int32, logger *zerolog.Logger, settings *Settings) *AudioStream {
	logger.Trace().Str("file", filepath.Base(file.Path)).Int32("idx", idx).Msgf("trancoder: Creating audio stream")
	ret := new(AudioStream)
	ret.index = idx
	ret.logger = logger
	ret.settings = settings
	NewStream(fmt.Sprintf("audio %d", idx), file, ret, &ret.Stream, settings, logger)
	return ret
}

func (as *AudioStream) getOutPath(encoderId int) string {
	return filepath.Join(as.file.Out, fmt.Sprintf("segment-a%d-%d-%%d.ts", as.index, encoderId))
}

func (as *AudioStream) getFlags() Flags {
	return AudioF
}

func (as *AudioStream) getTranscodeArgs(segments string) []string {
	return []string{
		"-map", fmt.Sprintf("0:a:%d", as.index),
		"-c:a", "aac",
		// TODO: Support 5.1 audio streams.
		"-ac", "2",
		// TODO: Support multi audio qualities.
		"-b:a", "128k",
	}
}
