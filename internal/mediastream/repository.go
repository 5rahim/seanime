package mediastream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
)

type (
	Repository struct {
		transcoder      mo.Option[*transcoder.Transcoder]
		settings        mo.Option[*models.MediastreamSettings]
		playbackManager *PlaybackManager
		logger          *zerolog.Logger
		wsEventManager  events.IWSEventManager
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.IWSEventManager
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:          opts.Logger,
		playbackManager: NewPlaybackManager(opts.Logger),
		settings:        mo.None[*models.MediastreamSettings](),
		transcoder:      mo.None[*transcoder.Transcoder](),
		wsEventManager:  opts.WSEventManager,
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
	if ok := r.initializeTranscoder(r.settings); ok {
		r.playbackManager.SetTranscoderSettings(mo.Some(r.transcoder.MustGet().GetSettings()))
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) RequestTranscodeStream(filepath string) (ret *MediaContainer, err error) {
	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Transcode stream requested")

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

func (r *Repository) initializeTranscoder(settings mo.Option[*models.MediastreamSettings]) bool {
	// Destroy the old transcoder if it exists
	if r.transcoder.IsPresent() {
		tc, _ := r.transcoder.Get()
		tc.Destroy()
	}

	r.transcoder = mo.None[*transcoder.Transcoder]()

	// If the temp directory is not set, don't initialize the transcoder
	if settings.MustGet().TranscodeTempDir == "" {
		return false
	}

	// If the transcoder is not enabled, don't initialize the transcoder
	if !settings.MustGet().TranscodeEnabled {
		return false
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
		return false
	}

	r.logger.Info().Msg("mediastream: Transcoder initialized")
	r.transcoder = mo.Some[*transcoder.Transcoder](tc)

	return true
}
