package mediastream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediastream/directstream"
	"github.com/seanime-app/seanime/internal/mediastream/optimizer"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"os"
)

type (
	Repository struct {
		transcoder         mo.Option[*transcoder.Transcoder]
		optimizer          *optimizer.Optimizer
		settings           mo.Option[*models.MediastreamSettings]
		playbackManager    *PlaybackManager
		directStream       *directstream.DirectStream
		mediaInfoExtractor *videofile.MediaInfoExtractor
		logger             *zerolog.Logger
		wsEventManager     events.IWSEventManager
		fileCacher         *filecache.Cacher
		cacheDir           string
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.IWSEventManager
		FileCacher     *filecache.Cacher
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger: opts.Logger,
		optimizer: optimizer.NewOptimizer(&optimizer.NewOptimizerOptions{
			Logger:         opts.Logger,
			WSEventManager: opts.WSEventManager,
		}),
		settings:           mo.None[*models.MediastreamSettings](),
		transcoder:         mo.None[*transcoder.Transcoder](),
		directStream:       directstream.NewDirectStream(opts.Logger),
		wsEventManager:     opts.WSEventManager,
		fileCacher:         opts.FileCacher,
		mediaInfoExtractor: videofile.NewMediaInfoExtractor(opts.FileCacher),
	}
	ret.playbackManager = NewPlaybackManager(ret)

	return ret
}

func (r *Repository) IsInitialized() bool {
	return r.settings.IsPresent()
}

func (r *Repository) InitializeModules(settings *models.MediastreamSettings, cacheDir string) {
	if settings == nil {
		r.logger.Error().Msg("mediastream: Settings not present")
		return
	}
	// Create the temp directory
	_ = os.MkdirAll(settings.TranscodeTempDir, 0755)
	// Set the settings
	r.settings = mo.Some[*models.MediastreamSettings](settings)
	r.cacheDir = cacheDir
	// Set the optimizer settings
	r.optimizer.SetLibraryDir(settings.PreTranscodeLibraryDir)
	// Initialize the transcoder
	if ok := r.initializeTranscoder(r.settings); ok {
		r.playbackManager.SetTranscoderSettings(mo.Some(r.transcoder.MustGet().GetSettings()))
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Direct Play
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) RequestDirectPlay(fp string) (ret *MediaContainer, err error) {
	r.logger.Debug().Str("filepath", fp).Msg("mediastream: Direct play requested")

	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	ret, err = r.playbackManager.RequestPlayback(fp, StreamTypeFile)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Direct Stream
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) RequestDirectStream(fp string, audioStreamIndex int) (ret *MediaContainer, err error) {
	r.logger.Debug().Str("fp", fp).Msg("mediastream: Direct play requested")

	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	ret, err = r.playbackManager.RequestPlayback(fp, StreamTypeDirectStream)

	// Copy video and audio streams to HLS
	r.directStream.CopyToHLS(&directstream.CopyToHLSOptions{
		Filepath:         ret.Filepath,
		Hash:             ret.Hash,
		OutDir:           r.settings.MustGet().TranscodeTempDir,
		AudioStreamIndex: audioStreamIndex,
	})

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Optimize
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StartMediaOptimizationOptions struct {
	Filepath          string
	Quality           optimizer.Quality
	AudioChannelIndex int
}

func (r *Repository) StartMediaOptimization(opts *StartMediaOptimizationOptions) (err error) {
	if !r.IsInitialized() {
		return errors.New("module not initialized")
	}

	mediaInfo, err := r.mediaInfoExtractor.GetInfo(opts.Filepath)
	if err != nil {
		return
	}

	err = r.optimizer.StartMediaOptimization(&optimizer.StartMediaOptimizationOptions{
		Filepath:  opts.Filepath,
		Quality:   opts.Quality,
		MediaInfo: mediaInfo,
	})
	return
}

func (r *Repository) RequestOptimizedStream(filepath string) (ret *MediaContainer, err error) {
	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	ret, err = r.playbackManager.RequestPlayback(filepath, StreamTypeOptimized)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Transcode
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) TranscoderIsInitialized() bool {
	return r.IsInitialized() && r.transcoder.IsPresent()
}

func (r *Repository) RequestTranscodeStream(filepath string) (ret *MediaContainer, err error) {
	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Transcode stream requested")

	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	// Reinitialize the transcoder for each new transcode request
	if ok := r.initializeTranscoder(r.settings); !ok {
		return nil, errors.New("transcoder not initialized")
	}

	r.playbackManager.SetTranscoderSettings(mo.Some(r.transcoder.MustGet().GetSettings()))

	ret, err = r.playbackManager.RequestPlayback(filepath, StreamTypeTranscode)

	return
}

///////////////////////////////////////////////////////////////////////////////////////////////

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
