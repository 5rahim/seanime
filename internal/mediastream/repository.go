package mediastream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mediastream/optimizer"
	"github.com/seanime-app/seanime/internal/mediastream/transcoder"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"github.com/seanime-app/seanime/internal/util/result"
	"os"
	"sync"
)

type (
	Repository struct {
		transcoder                 mo.Option[*transcoder.Transcoder]
		directPlayVideoStreamCache *result.Map[string, *VideoStream]
		optimizer                  *optimizer.Optimizer
		settings                   mo.Option[*models.MediastreamSettings]
		playbackManager            *PlaybackManager
		mediaInfoExtractor         *videofile.MediaInfoExtractor
		logger                     *zerolog.Logger
		wsEventManager             events.WSEventManagerInterface
		fileCacher                 *filecache.Cacher
		reqMu                      sync.Mutex
		cacheDir                   string
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.WSEventManagerInterface
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
		directPlayVideoStreamCache: result.NewResultMap[string, *VideoStream](),
		settings:                   mo.None[*models.MediastreamSettings](),
		transcoder:                 mo.None[*transcoder.Transcoder](),
		wsEventManager:             opts.WSEventManager,
		fileCacher:                 opts.FileCacher,
		mediaInfoExtractor:         videofile.NewMediaInfoExtractor(opts.FileCacher, opts.Logger),
	}
	ret.playbackManager = NewPlaybackManager(ret)

	return ret
}

func (r *Repository) IsInitialized() bool {
	return r.settings.IsPresent()
}

func (r *Repository) OnCleanup() {
	r.directPlayVideoStreamCache.Range(func(key string, value *VideoStream) bool {
		_ = value.File.Close()
		return true
	})
}

func (r *Repository) InitializeModules(settings *models.MediastreamSettings, cacheDir string) {
	if settings == nil {
		r.logger.Error().Msg("mediastream: Settings not present")
		return
	}
	// Create the temp directory
	_ = os.MkdirAll(settings.TranscodeTempDir, 0755)

	if settings.FfmpegPath == "" {
		settings.FfmpegPath = "ffmpeg"
	}

	if settings.FfprobePath == "" {
		settings.FfprobePath = "ffprobe"
	}

	// Set the settings
	r.settings = mo.Some[*models.MediastreamSettings](settings)

	r.cacheDir = cacheDir

	// Set the optimizer settings
	r.optimizer.SetLibraryDir(settings.PreTranscodeLibraryDir)

	// Initialize the transcoder
	if ok := r.initializeTranscoder(r.settings); ok {
	}

	r.logger.Info().Msg("mediastream: Module initialized")
}

// CacheWasCleared should be called when the cache directory is manually cleared.
func (r *Repository) CacheWasCleared() {
	r.playbackManager.mediaContainers.Clear()
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

	mediaInfo, err := r.mediaInfoExtractor.GetInfo(r.settings.MustGet().FfmpegPath, opts.Filepath)
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

func (r *Repository) RequestTranscodeStream(filepath string, clientId string) (ret *MediaContainer, err error) {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Transcode stream requested")

	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	// Reinitialize the transcoder for each new transcode request
	if ok := r.initializeTranscoder(r.settings); !ok {
		return nil, errors.New("real-time transcoder not initialized, check your settings")
	}

	ret, err = r.playbackManager.RequestPlayback(filepath, StreamTypeTranscode)

	return
}

func (r *Repository) RequestPreloadTranscodeStream(filepath string) (err error) {
	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Transcode stream preloading requested")

	if !r.IsInitialized() {
		return errors.New("module not initialized")
	}

	_, err = r.playbackManager.PreloadPlayback(filepath, StreamTypeTranscode)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Direct Play
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) RequestDirectPlay(filepath string, clientId string) (ret *MediaContainer, err error) {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Direct play requested")

	if !r.IsInitialized() {
		return nil, errors.New("module not initialized")
	}

	ret, err = r.playbackManager.RequestPlayback(filepath, StreamTypeDirect)

	return
}

func (r *Repository) RequestPreloadDirectPlay(filepath string) (err error) {
	r.logger.Debug().Str("filepath", filepath).Msg("mediastream: Direct stream preloading requested")

	if !r.IsInitialized() {
		return errors.New("module not initialized")
	}

	_, err = r.playbackManager.PreloadPlayback(filepath, StreamTypeDirect)

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

	// If the transcoder is not enabled, don't initialize the transcoder
	if !settings.MustGet().TranscodeEnabled {
		return false
	}

	// If the temp directory is not set, don't initialize the transcoder
	if settings.MustGet().TranscodeTempDir == "" {
		r.logger.Error().Msg("mediastream: Transcode directory not set, could not initialize transcoder")
		return false
	}

	opts := &transcoder.NewTranscoderOptions{
		Logger:      r.logger,
		HwAccelKind: settings.MustGet().TranscodeHwAccel,
		Preset:      settings.MustGet().TranscodePreset,
		TempOutDir:  settings.MustGet().TranscodeTempDir,
		FfmpegPath:  settings.MustGet().FfmpegPath,
		FfprobePath: settings.MustGet().FfprobePath,
	}

	tc, err := transcoder.NewTranscoder(opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("mediastream: Failed to initialize transcoder")
		return false
	}

	r.logger.Info().Msg("mediastream: Transcoder module initialized")
	r.transcoder = mo.Some[*transcoder.Transcoder](tc)

	return true
}
