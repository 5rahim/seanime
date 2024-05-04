package directstream

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/xfrr/goffmpeg/transcoder"
	"os"
	"path/filepath"
)

type (
	DirectStream struct {
		logger *zerolog.Logger
	}
)

func NewDirectStream(logger *zerolog.Logger) *DirectStream {
	return &DirectStream{
		logger: logger,
	}
}

type CopyToHLSOptions struct {
	Filepath         string
	Hash             string
	OutDir           string // The main temp directory. A subdirectory will be created for the file.
	AudioStreamIndex int
}

func (ds *DirectStream) GetFileOutDir(outDir string, hash string) string {
	return filepath.Join(outDir, "directstreams", hash)
}

func (ds *DirectStream) CopyToHLS(opts *CopyToHLSOptions) {

	// Create a temp directory for the stream output
	// e.g. /Users/username/temp_collection/{hash}
	fileOutDir := ds.GetFileOutDir(opts.OutDir, opts.Hash)
	_ = os.MkdirAll(fileOutDir, 0755)

	transc := new(transcoder.Transcoder)

	masterFilePath := filepath.Join(fileOutDir, "master.m3u8")

	ds.logger.Trace().Msgf(fmt.Sprintf("directstream: Master playlist path: %v", masterFilePath))

	// Check if the master playlist already exists
	_, err := os.Stat(masterFilePath)
	if err == nil {
		ds.logger.Trace().Msgf(fmt.Sprintf("directstream: Master playlist already exists"))
		return
	}

	err = transc.Initialize(opts.Filepath, masterFilePath)
	if err != nil {
		panic(err)
	}

	transc.MediaFile().SetHardwareAcceleration("auto")
	transc.MediaFile().SetRawOutputArgs([]string{"-c", "copy", "-map", "0:v", "-map", "0:a:1"})
	transc.MediaFile().SetHlsMasterPlaylistName("master.m3u8")
	transc.MediaFile().SetHlsSegmentDuration(4)
	//transc.MediaFile().SetHlsListSize(0)

	ds.logger.Trace().Msgf(fmt.Sprintf("directstream: Starting copy"))

	done := transc.Run(true)
	progress := transc.Output()
	for p := range progress {
		ds.logger.Trace().Msgf(fmt.Sprintf("directstream: Progress: %v", p))
	}
	<-done

	ds.logger.Info().Msgf(fmt.Sprintf("directstream: Copy done"))
}
