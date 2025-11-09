package mediastream

import (
	"os"
	"sync"
	"testing"

	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/util/filecache"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWSEventManager struct{}

func (m *mockWSEventManager) SendEvent(event string, data interface{}) {}
func (m *mockWSEventManager) SendEventTo(clientId string, event string, data interface{}, skipSelf ...bool) {}
func (m *mockWSEventManager) SubscribeToClientEvents(id string) *events.ClientEventSubscriber { return nil }
func (m *mockWSEventManager) SubscribeToClientNativePlayerEvents(id string) *events.ClientEventSubscriber { return nil }
func (m *mockWSEventManager) SubscribeToClientNakamaEvents(id string) *events.ClientEventSubscriber { return nil }
func (m *mockWSEventManager) SubscribeToClientPlaylistEvents(id string) *events.ClientEventSubscriber { return nil }
func (m *mockWSEventManager) UnsubscribeFromClientEvents(id string) {}

func TestTranscoderLifecycle(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	cacheDir, err := os.MkdirTemp("", "seanime-test-cache-*")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	transcodeDir, err := os.MkdirTemp("", "seanime-test-transcode-*")
	require.NoError(t, err)
	defer os.RemoveAll(transcodeDir)

	fileCacher, err := filecache.NewCacher(cacheDir)
	require.NoError(t, err)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:         &logger,
		WSEventManager: &mockWSEventManager{},
		FileCacher:     fileCacher,
	})

	t.Run("Transcoder initialization", func(t *testing.T) {
		settings := &models.MediastreamSettings{
			BaseModel: models.BaseModel{ID: 1},
			TranscodeEnabled: true,
			TranscodeHwAccel: "cpu",
			TranscodePreset:  "fast",
			FfmpegPath:      "ffmpeg",
			FfprobePath:     "ffprobe",
		}

		repo.InitializeModules(settings, cacheDir, transcodeDir)

		assert.True(t, repo.IsInitialized())
		assert.True(t, repo.transcoder.IsPresent())
	})

	t.Run("Transcoder persistence across requests", func(t *testing.T) {
		assert.True(t, repo.transcoder.IsPresent())

		// Simulate multiple transcode requests
		// The transcoder should not be recreated
		for i := 0; i < 3; i++ {
			// This would normally be called by RequestTranscodeStream
			result := repo.initializeTranscoder(repo.settings)
			assert.True(t, result)
		}

		assert.True(t, repo.transcoder.IsPresent())
	})

	t.Run("Settings change triggers transcoder refresh", func(t *testing.T) {
		assert.True(t, repo.transcoder.IsPresent())

		newSettings := &models.MediastreamSettings{
			BaseModel: models.BaseModel{ID: 1},
			TranscodeEnabled: true,
			TranscodeHwAccel: "vaapi", // Changed from "cpu"
			TranscodePreset:  "fast",
			FfmpegPath:      "ffmpeg",
			FfprobePath:     "ffprobe",
		}

		repo.InitializeModules(newSettings, cacheDir, transcodeDir)

		assert.True(t, repo.transcoder.IsPresent())
		assert.Equal(t, "vaapi", repo.settings.MustGet().TranscodeHwAccel)
	})

	t.Run("Transcoder disabled in settings", func(t *testing.T) {
		disabledSettings := &models.MediastreamSettings{
			BaseModel: models.BaseModel{ID: 1},
			TranscodeEnabled: false, // Disabled
			TranscodeHwAccel: "cpu",
			TranscodePreset:  "fast",
			FfmpegPath:      "ffmpeg",
			FfprobePath:     "ffprobe",
		}

		repo.InitializeModules(disabledSettings, cacheDir, transcodeDir)

		assert.False(t, repo.transcoder.IsPresent())
	})

	t.Run("Explicit transcoder destruction", func(t *testing.T) {
		settings := &models.MediastreamSettings{
			BaseModel: models.BaseModel{ID: 1},
			TranscodeEnabled: true,
			TranscodeHwAccel: "cpu",
			TranscodePreset:  "fast",
			FfmpegPath:      "ffmpeg",
			FfprobePath:     "ffprobe",
		}
		repo.InitializeModules(settings, cacheDir, transcodeDir)
		assert.True(t, repo.transcoder.IsPresent())

		repo.DestroyTranscoder()
		assert.False(t, repo.transcoder.IsPresent())

		result := repo.initializeTranscoder(repo.settings)
		assert.True(t, result)
		assert.True(t, repo.transcoder.IsPresent())
	})
}

func TestTranscoderThreadSafety(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	cacheDir, err := os.MkdirTemp("", "seanime-test-cache-*")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	transcodeDir, err := os.MkdirTemp("", "seanime-test-transcode-*")
	require.NoError(t, err)
	defer os.RemoveAll(transcodeDir)

	fileCacher, err := filecache.NewCacher(cacheDir)
	require.NoError(t, err)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:         &logger,
		WSEventManager: &mockWSEventManager{},
		FileCacher:     fileCacher,
	})

	settings := &models.MediastreamSettings{
		BaseModel: models.BaseModel{ID: 1},
		TranscodeEnabled: true,
		TranscodeHwAccel: "cpu",
		TranscodePreset:  "fast",
		FfmpegPath:      "ffmpeg",
		FfprobePath:     "ffprobe",
	}

	repo.InitializeModules(settings, cacheDir, transcodeDir)

	t.Run("Concurrent initialization attempts", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make([]error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				result := repo.initializeTranscoder(repo.settings)
				if !result {
					errors[index] = assert.AnError
				}
			}(i)
		}

		wg.Wait()

		for _, err := range errors {
			assert.NoError(t, err)
		}

		assert.True(t, repo.transcoder.IsPresent())
	})

	t.Run("Concurrent settings changes", func(t *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				newSettings := &models.MediastreamSettings{
					BaseModel: models.BaseModel{ID: 1},
					TranscodeEnabled: true,
					TranscodeHwAccel: "cpu",
					TranscodePreset:  []string{"fast", "medium", "slow"}[index%3],
					FfmpegPath:      "ffmpeg",
					FfprobePath:     "ffprobe",
				}

				repo.InitializeModules(newSettings, cacheDir, transcodeDir)
			}(i)
		}

		wg.Wait()

		assert.True(t, repo.IsInitialized())
		if repo.settings.MustGet().TranscodeEnabled {
			assert.True(t, repo.transcoder.IsPresent())
		}
	})
}

func TestRefreshTranscoderOnSettingsChange(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	cacheDir, err := os.MkdirTemp("", "seanime-test-cache-*")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	transcodeDir, err := os.MkdirTemp("", "seanime-test-transcode-*")
	require.NoError(t, err)
	defer os.RemoveAll(transcodeDir)

	fileCacher, err := filecache.NewCacher(cacheDir)
	require.NoError(t, err)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:         &logger,
		WSEventManager: &mockWSEventManager{},
		FileCacher:     fileCacher,
	})

	t.Run("RefreshTranscoderOnSettingsChange method", func(t *testing.T) {
		settings := &models.MediastreamSettings{
			BaseModel: models.BaseModel{ID: 1},
			TranscodeEnabled: true,
			TranscodeHwAccel: "cpu",
			TranscodePreset:  "fast",
			FfmpegPath:      "ffmpeg",
			FfprobePath:     "ffprobe",
		}

		repo.InitializeModules(settings, cacheDir, transcodeDir)
		assert.True(t, repo.transcoder.IsPresent())

		repo.RefreshTranscoderOnSettingsChange()

		assert.True(t, repo.transcoder.IsPresent())
	})

	t.Run("No refresh when uninitialized", func(t *testing.T) {
		newRepo := NewRepository(&NewRepositoryOptions{
			Logger:         &logger,
			WSEventManager: &mockWSEventManager{},
			FileCacher:     fileCacher,
		})

		assert.NotPanics(t, func() {
			newRepo.RefreshTranscoderOnSettingsChange()
		})

		assert.False(t, newRepo.transcoder.IsPresent())
	})
}