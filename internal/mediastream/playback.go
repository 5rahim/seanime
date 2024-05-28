package mediastream

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"github.com/seanime-app/seanime/internal/util/result"
)

const (
	StreamTypeTranscode StreamType = "transcode" // On-the-fly transcoding
	StreamTypeOptimized StreamType = "optimized" // Pre-transcoded
)

type (
	StreamType string

	PlaybackManager struct {
		logger                *zerolog.Logger
		currentMediaContainer mo.Option[*MediaContainer] // The current media being played.
		repository            *Repository
		mediaContainers       *result.Map[string, *MediaContainer] // Temporary cache for the media containers.
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
		logger:          repository.logger,
		repository:      repository,
		mediaContainers: result.NewResultMap[string, *MediaContainer](),
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

	// Create a new media container
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

// PreloadPlayback is called by the frontend to preload a media container so that the data is stored in advanced
func (p *PlaybackManager) PreloadPlayback(filepath string, streamType StreamType) (ret *MediaContainer, err error) {

	p.logger.Debug().Str("filepath", filepath).Any("type", streamType).Msg("mediastream: Preloading playback")

	// Create a new media container
	ret, err = p.newMediaContainer(filepath, streamType)

	if err != nil {
		p.logger.Error().Err(err).Msg("mediastream: Failed to create media container")
		return nil, fmt.Errorf("failed to create media container: %v", err)
	}

	p.logger.Info().Str("filepath", filepath).Msg("mediastream: Ready to play media")

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Optimize
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *PlaybackManager) newMediaContainer(filepath string, streamType StreamType) (ret *MediaContainer, err error) {
	p.logger.Debug().Str("filepath", filepath).Any("type", streamType).Msg("mediastream: Creating media container")
	// Get the hash of the file.
	hash, err := videofile.GetHashFromPath(filepath)
	if err != nil {
		return nil, err
	}

	// Check the cache
	if mc, ok := p.mediaContainers.Get(hash); ok {
		return mc, nil
	}

	// Get the media information of the file.
	ret = &MediaContainer{
		Filepath:   filepath,
		Hash:       hash,
		StreamType: streamType,
	}

	p.logger.Debug().Msg("mediastream: Extracting media info")

	ret.MediaInfo, err = p.repository.mediaInfoExtractor.GetInfo(p.repository.settings.MustGet().FfprobePath, filepath)
	if err != nil {
		return nil, err
	}

	// Extract the attachments from the file.
	err = videofile.ExtractAttachment(p.repository.settings.MustGet().FfmpegPath, filepath, hash, ret.MediaInfo, p.repository.cacheDir, p.logger)
	if err != nil {
		return nil, err
	}

	streamUrl := ""
	switch streamType {
	case StreamTypeTranscode:
		// Live transcode the file.
		streamUrl = "/api/v1/mediastream/transcode/master.m3u8"
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

	// Set the stream URL.
	ret.StreamUrl = streamUrl

	// Store the media container in the map.
	p.mediaContainers.Set(hash, ret)

	return
}
