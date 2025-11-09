package mediastream

import (
	"errors"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/mediastream/optimizer"
	"seanime/internal/mediastream/transcoder"
	"seanime/internal/mediastream/videofile"
	"seanime/internal/util/filecache"
	"sync"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	Repository struct {
		transcoder         mo.Option[*transcoder.Transcoder]
		optimizer          *optimizer.Optimizer
		settings           mo.Option[*models.MediastreamSettings]
		playbackManager    *PlaybackManager
		mediaInfoExtractor *videofile.MediaInfoExtractor
		logger             *zerolog.Logger
		wsEventManager     events.WSEventManagerInterface
		fileCacher         *filecache.Cacher
		reqMu              sync.Mutex
		cacheDir           string // where attachments are stored
		transcodeDir       string // where stream segments are stored
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
		settings:           mo.None[*models.MediastreamSettings](),
		transcoder:         mo.None[*transcoder.Transcoder](),
		wsEventManager:     opts.WSEventManager,
		fileCacher:         opts.FileCacher,
		mediaInfoExtractor: videofile.NewMediaInfoExtractor(opts.FileCacher, opts.Logger),
	}
	ret.playbackManager = NewPlaybackManager(ret)

	return ret
}

func (r *Repository) IsInitialized() bool {
	return r.settings.IsPresent()
}

func (r *Repository) OnCleanup() {

}

func (r *Repository) InitializeModules(settings *models.MediastreamSettings, cacheDir string, transcodeDir string) {
	if settings == nil {
		r.logger.Error().Msg("mediastream: Settings not present")
		return
	}
	// Create the temp directory
	err := os.MkdirAll(transcodeDir, 0755)
	if err != nil {
		r.logger.Error().Err(err).Msg("mediastream: Failed to create transcode directory")
	}

	if settings.FfmpegPath == "" {
		settings.FfmpegPath = "ffmpeg"
	}

	if settings.FfprobePath == "" {
		settings.FfprobePath = "ffprobe"
	}

	// Check if we need to refresh the transcoder due to settings change
	oldSettings := r.settings
	needsTranscoderRefresh := false
	if oldSettings.IsPresent() {
		old := oldSettings.MustGet()
		if old.TranscodeEnabled != settings.TranscodeEnabled ||
			old.TranscodeHwAccel != settings.TranscodeHwAccel ||
			old.TranscodePreset != settings.TranscodePreset ||
			old.FfmpegPath != settings.FfmpegPath ||
			old.FfprobePath != settings.FfprobePath ||
			old.TranscodeHwAccelCustomSettings != settings.TranscodeHwAccelCustomSettings {
			needsTranscoderRefresh = true
			r.logger.Info().Msg("mediastream: Transcoder settings changed, will refresh transcoder")
		}
	}

	// Set the settings
	r.settings = mo.Some[*models.MediastreamSettings](settings)

	r.cacheDir = cacheDir
	r.transcodeDir = transcodeDir

	// Set the optimizer settings
	r.optimizer.SetLibraryDir(settings.PreTranscodeLibraryDir)

	// Handle transcoder initialization or refresh
	if needsTranscoderRefresh {
		r.RefreshTranscoderOnSettingsChange()
	} else if !r.initializeTranscoder(r.settings) {
		r.logger.Warn().Msg("mediastream: Transcoder was not initialized (disabled or failed)")
	}

	r.logger.Info().Msg("mediastream: Module initialized")
}

// CacheWasCleared should be called when the cache directory is manually cleared.
func (r *Repository) CacheWasCleared() {
	r.playbackManager.mediaContainers.Clear()
}

func (r *Repository) ClearTranscodeDir() {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	r.logger.Trace().Msg("mediastream: Clearing transcode directory")

	// Empty the transcode directory
	if r.transcodeDir != "" {
		files, err := os.ReadDir(r.transcodeDir)
		if err != nil {
			r.logger.Error().Err(err).Msg("mediastream: Failed to read transcode directory")
			return
		}

		for _, file := range files {
			err = os.RemoveAll(filepath.Join(r.transcodeDir, file.Name()))
			if err != nil {
				r.logger.Error().Err(err).Msg("mediastream: Failed to remove file from transcode directory")
			}
		}
	}

	r.logger.Debug().Msg("mediastream: Transcode directory cleared")

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

	// Only initialize the transcoder if it doesn't exist to prevent destroying active sessions
	if !r.transcoder.IsPresent() {
		r.logger.Debug().Msg("mediastream: Transcoder not present, initializing")
		if ok := r.initializeTranscoder(r.settings); !ok {
			return nil, errors.New("real-time transcoder not initialized, check your settings")
		}
	}
	r.logger.Debug().Msg("mediastream: Using existing transcoder")

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
	// Ensure thread-safe transcoder initialization
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	if r.transcoder.IsPresent() {
		r.logger.Debug().Msg("mediastream: Transcoder already initialized, skipping re-initialization")
		return true
	}

	// Note: Transcoder is not destroyed here to prevent HLS segment interruption.
	// Destruction only occurs on settings change, shutdown, or explicit cleanup.

	r.transcoder = mo.None[*transcoder.Transcoder]()

	// If the transcoder is not enabled, don't initialize the transcoder
	if !settings.MustGet().TranscodeEnabled {
		return false
	}

	// If the temp directory is not set, don't initialize the transcoder
	if r.transcodeDir == "" {
		r.logger.Error().Msg("mediastream: Transcode directory not set, could not initialize transcoder")
		return false
	}

	opts := &transcoder.NewTranscoderOptions{
		Logger:                r.logger,
		HwAccelKind:           settings.MustGet().TranscodeHwAccel,
		Preset:                settings.MustGet().TranscodePreset,
		FfmpegPath:            settings.MustGet().FfmpegPath,
		FfprobePath:           settings.MustGet().FfprobePath,
		HwAccelCustomSettings: settings.MustGet().TranscodeHwAccelCustomSettings,
		TempOutDir:            r.transcodeDir,
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

// DestroyTranscoder explicitly destroys the transcoder and releases all associated resources.
// Thread-safe. Only call when settings change, during shutdown, or for manual cleanup.
// Transcoder must be re-initialized after calling this method.
func (r *Repository) DestroyTranscoder() {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	if r.transcoder.IsPresent() {
		r.logger.Debug().Msg("mediastream: Destroying transcoder explicitly")
		tc, _ := r.transcoder.Get()
		tc.Destroy()
		r.transcoder = mo.None[*transcoder.Transcoder]()
	}
}

// RefreshTranscoderOnSettingsChange safely refreshes the transcoder when settings are updated.
// Destroys existing transcoder and re-initializes with current settings.
// Thread-safe. Called automatically by InitializeModules on settings change.
func (r *Repository) RefreshTranscoderOnSettingsChange() {
	r.DestroyTranscoder()
	if r.IsInitialized() {
		r.initializeTranscoder(r.settings)
		r.logger.Info().Msg("mediastream: Transcoder refreshed due to settings change")
	}
}
