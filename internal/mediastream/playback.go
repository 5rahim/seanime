package mediastream

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"os"
)

const (
	StreamTypeFile          StreamType = "file"
	StreamTypeTranscode     StreamType = "transcode"
	StreamTypePreTranscoded StreamType = "pre_transcoded"
)

type (
	StreamType string

	PlaybackManager struct {
		logger                *zerolog.Logger
		currentMediaContainer mo.Option[*MediaContainer] // The current media being played.
		transcoderSettings    mo.Option[*transcoder.Settings]
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

func NewPlaybackManager(logger *zerolog.Logger) *PlaybackManager {
	return &PlaybackManager{
		logger: logger,
	}
}

func (p *PlaybackManager) SetTranscoderSettings(settings mo.Option[*transcoder.Settings]) {
	if settings.IsPresent() {
		p.transcoderSettings = settings
	}
}

// RequestTranscodePlayback is called by the frontend to stream a media file with HLS (Transcoding).
func (p *PlaybackManager) RequestTranscodePlayback(filepath string) (ret *MediaContainer, err error) {

	p.logger.Debug().Str("filepath", filepath).Msg("mediastream: Playback request received")

	ret, err = p.newMediaContainer(filepath, StreamTypeTranscode)

	if err != nil {
		p.logger.Error().Err(err).Msg("mediastream: Failed to create media container")
		return nil, fmt.Errorf("failed to create media container: %v", err)
	}

	// Set the current media container.
	p.currentMediaContainer = mo.Some(ret)

	p.logger.Info().Msg("mediastream: Ready to stream media")

	return
}

func (p *PlaybackManager) newMediaContainer(filepath string, streamType StreamType) (ret *MediaContainer, err error) {
	// Get the hash of the file.
	hash, err := getHash(filepath)
	if err != nil {
		return nil, err
	}
	ret = &MediaContainer{
		Filepath:   filepath,
		Hash:       hash,
		StreamType: streamType,
	}

	// Get the media information of the file.
	mediaInfoExtractor, err := videofile.NewMediaInfoExtractor(filepath, hash, p.logger)
	if err != nil {
		return nil, err
	}

	if !p.transcoderSettings.IsPresent() {
		return nil, errors.New("transcoder settings not set")
	}

	ret.MediaInfo, err = mediaInfoExtractor.GetInfo(p.transcoderSettings.MustGet().MetadataDir)
	if err != nil {
		return nil, err
	}

	ch, err := transcoder.Extract(filepath, hash, ret.MediaInfo, p.transcoderSettings.MustGet(), p.logger)
	if err != nil {
		return nil, err
	}
	<-ch

	streamUrl := ""
	switch streamType {
	case StreamTypeTranscode:
		// Live transcode the file.
		streamUrl = "/api/v1/mediastream/transcode/master.m3u8"
	case StreamTypeFile:
		// TODO
		streamUrl = "/api/v1/mediastream/direct"
	case StreamTypePreTranscoded:
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

func getHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write([]byte(path))
	h.Write([]byte(info.ModTime().String()))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha, nil
}
