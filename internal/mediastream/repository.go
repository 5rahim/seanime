package mediastream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
)

type (
	Repository struct {
		transcoder      mo.Option[*transcoder.Transcoder]
		settings        mo.Option[*models.MediastreamSettings]
		playbackManager *PlaybackManager
		logger          *zerolog.Logger
	}

	NewRepositoryOptions struct {
		Logger *zerolog.Logger
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:          opts.Logger,
		playbackManager: NewPlaybackManager(opts.Logger),
		settings:        mo.None[*models.MediastreamSettings](),
		transcoder:      mo.None[*transcoder.Transcoder](),
	}

	return ret
}

func (r *Repository) IsInitialized() bool {
	return r.settings.IsPresent()
}

func (r *Repository) TranscoderIsInitialized() bool {
	return r.IsInitialized() && r.transcoder.IsPresent()
}

func (r *Repository) InitializeModules(settings *models.MediastreamSettings) {
	if settings == nil {
		r.logger.Error().Msg("mediastream: Settings not present")
		return
	}
	// Set the settings
	r.settings = mo.Some[*models.MediastreamSettings](settings)
	// Initialize the transcoder
	r.initializeTranscoder(r.settings)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) RequestTranscodeStream(filepath string) (ret *MediaContainer, err error) {
	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Playback request received")

	if !r.IsInitialized() {
		return nil, errors.New("transcoding module not initialized")
	}

	if !r.TranscoderIsInitialized() {
		return nil, errors.New("transcoder not initialized")
	}

	ret, err = r.playbackManager.RequestTranscodePlayback(filepath)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) initializeTranscoder(settings mo.Option[*models.MediastreamSettings]) {
	// Destroy the old transcoder if it exists
	if r.transcoder.IsPresent() {
		tc, _ := r.transcoder.Get()
		tc.Destroy()
	}

	r.transcoder = mo.None[*transcoder.Transcoder]()

	// If the temp directory is not set, don't initialize the transcoder
	if settings.MustGet().TranscodeTempDir == "" {
		return
	}

	// If the transcoder is not enabled, don't initialize the transcoder
	if !settings.MustGet().TranscodeEnabled {
		return
	}

	opts := &transcoder.NewTranscoderOptions{
		Logger:      r.logger,
		HwAccelKind: settings.MustGet().TranscodeHwAccel,
		Preset:      settings.MustGet().TranscodePreset,
		TempOutDir:  settings.MustGet().TranscodeTempDir,
	}

	tc, err := transcoder.NewTranscoder(opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("mediastream: Failed to initialize transcoder")
		return
	}

	r.logger.Info().Msg("mediastream: Transcoder initialized")
	r.transcoder = mo.Some[*transcoder.Transcoder](tc)
}
