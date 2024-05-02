package transcoder

import (
	"fmt"
	"github.com/rs/zerolog"
)

type AudioStream struct {
	Stream
	index  int32
	logger *zerolog.Logger
}

// NewAudioStream creates a new AudioStream for a file, at a given audio index.
func NewAudioStream(file *FileStream, idx int32, logger *zerolog.Logger) *AudioStream {
	logger.Trace().Str("path", file.Path).Msgf("Creating audio stream %d", idx)
	ret := new(AudioStream)
	ret.index = idx
	ret.logger = logger
	ret.Stream.logger = logger
	NewStream(file, ret, &ret.Stream)
	return ret
}

func (as *AudioStream) getOutPath(encoderId int) string {
	return fmt.Sprintf("%s/segment-a%d-%d-%%d.ts", as.file.Out, as.index, encoderId)
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
