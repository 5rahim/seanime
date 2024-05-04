package mediastream

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
)

const (
	StreamTypeFile         StreamType = "file"      // Direct play
	StreamTypeDirectStream StreamType = "direct"    // Direct stream
	StreamTypeTranscode    StreamType = "transcode" // On-the-fly transcoding
	StreamTypeOptimized    StreamType = "optimized" // Pre-transcoded
)

type (
	StreamType string

	PlaybackManager struct {
		logger                *zerolog.Logger
		currentMediaContainer mo.Option[*MediaContainer] // The current media being played.
		transcoderSettings    mo.Option[*transcoder.Settings]
		repository            *Repository
	}

	PlaybackState struct {
		MediaId int `json:"mediaId"` // The media ID
	}

	MediaContainer struct {
		Filepath   string               `json:"filePath"`
		Hash       string               `json:"hash"`
		StreamType StreamType           `json:"streamType"` // Tells the frontend how to play the media.
		StreamUrl  string               `json:"streamUrl"`  // The relative endpoint to stream the media.
		MediaInfo  *videofile.MediaInfo `json:"mediaInfo"`
		//Metadata  *Metadata       `json:"metadata"`
		// todo: add more fields (e.g. metadata)
	}
)

func NewPlaybackManager(repository *Repository) *PlaybackManager {
	return &PlaybackManager{
		logger:     repository.logger,
		repository: repository,
	}
}

func (p *PlaybackManager) KillPlayback() {
	if p.currentMediaContainer.IsPresent() {
		p.currentMediaContainer = mo.None[*MediaContainer]()
	}
}

// RequestPlayback is called by the frontend to stream a media file
func (p *PlaybackManager) RequestPlayback(filepath string, streamType StreamType) (ret *MediaContainer, err error) {

	p.logger.Debug().Str("filepath", filepath).Any("type", streamType).Msg("mediastream: Requesting playback")

	ret, err = p.newMediaContainer(filepath, streamType)

	if err != nil {
		p.logger.Error().Err(err).Msg("mediastream: Failed to create media container")
		return nil, fmt.Errorf("failed to create media container: %v", err)
	}

	// Set the current media container.
	p.currentMediaContainer = mo.Some(ret)

	p.logger.Info().Str("filepath", filepath).Msg("mediastream: Ready to play media")

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Optimize
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Transcode
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *PlaybackManager) SetTranscoderSettings(settings mo.Option[*transcoder.Settings]) {
	if settings.IsPresent() {
		p.transcoderSettings = settings
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *PlaybackManager) newMediaContainer(filepath string, streamType StreamType) (ret *MediaContainer, err error) {
	p.logger.Debug().Str("filepath", filepath).Any("type", streamType).Msg("mediastream: Creating media container")
	// Get the hash of the file.
	hash, err := videofile.GetHashFromPath(filepath)
	if err != nil {
		return nil, err
	}

	// Get the media information of the file.
	ret = &MediaContainer{
		Filepath:   filepath,
		Hash:       hash,
		StreamType: streamType,
	}
	ret.MediaInfo, err = p.repository.mediaInfoExtractor.GetInfo(filepath)
	if err != nil {
		return nil, err
	}

	// Extract the attachments from the file.
	err = videofile.ExtractAttachment(filepath, hash, ret.MediaInfo, p.repository.cacheDir, p.logger)
	if err != nil {
		return nil, err
	}

	streamUrl := ""
	switch streamType {
	case StreamTypeTranscode:
		// Live transcode the file.
		streamUrl = "/api/v1/mediastream/transcode/master.m3u8"
	case StreamTypeFile:
		// TODO
		streamUrl = "/api/v1/mediastream/direct"
	case StreamTypeDirectStream:
		// TODO
		streamUrl = "/api/v1/mediastream/directstream/master.m3u8"
	case StreamTypeOptimized:
		// TODO: Check if the file is already transcoded when the feature is implemented.
		// ...
		streamUrl = "/api/v1/mediastream/hls/master.m3u8"
	}

	// TODO: Add metadata to the media container.
	// ...

	if streamUrl == "" {
		return nil, errors.New("invalid stream type")
	}

	ret.StreamUrl = streamUrl

	return
}
