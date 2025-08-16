package mediaplayer

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/continuity"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/mediaplayers/iina"
	mpchc2 "seanime/internal/mediaplayers/mpchc"
	"seanime/internal/mediaplayers/mpv"
	vlc2 "seanime/internal/mediaplayers/vlc"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	PlayerClosedEvent = "Player closed"
)

type PlaybackType string

const (
	PlaybackTypeFile   PlaybackType = "file"
	PlaybackTypeStream PlaybackType = "stream"
)

type (
	// Repository provides a common interface to interact with media players
	Repository struct {
		Logger                *zerolog.Logger
		Default               string
		VLC                   *vlc2.VLC
		MpcHc                 *mpchc2.MpcHc
		Mpv                   *mpv.Mpv
		Iina                  *iina.Iina
		wsEventManager        events.WSEventManagerInterface
		continuityManager     *continuity.Manager
		playerInUse           string
		completionThreshold   float64
		mu                    sync.RWMutex
		isRunning             bool
		currentPlaybackStatus *PlaybackStatus
		subscribers           *result.Map[string, *RepositorySubscriber]
		cancel                context.CancelFunc
		exitedCh              chan struct{} // Closed when the media player exits
	}

	NewRepositoryOptions struct {
		Logger            *zerolog.Logger
		Default           string
		VLC               *vlc2.VLC
		MpcHc             *mpchc2.MpcHc
		Mpv               *mpv.Mpv
		Iina              *iina.Iina
		WSEventManager    events.WSEventManagerInterface
		ContinuityManager *continuity.Manager
	}

	// RepositorySubscriber provides a single event channel for all media player events
	RepositorySubscriber struct {
		EventCh chan MediaPlayerEvent
	}

	// MediaPlayerEvent is the base interface for all media player events
	MediaPlayerEvent interface {
		Type() string
	}

	// Local file playback events
	TrackingStartedEvent struct {
		Status *PlaybackStatus
	}

	TrackingRetryEvent struct {
		Reason string
	}

	VideoCompletedEvent struct {
		Status *PlaybackStatus
	}

	TrackingStoppedEvent struct {
		Reason string
	}

	PlaybackStatusEvent struct {
		Status *PlaybackStatus
	}

	// Streaming playback events
	StreamingTrackingStartedEvent struct {
		Status *PlaybackStatus
	}

	StreamingTrackingRetryEvent struct {
		Reason string
	}

	StreamingVideoCompletedEvent struct {
		Status *PlaybackStatus
	}

	StreamingTrackingStoppedEvent struct {
		Reason string
	}

	StreamingPlaybackStatusEvent struct {
		Status *PlaybackStatus
	}

	PlaybackStatus struct {
		CompletionPercentage float64 `json:"completionPercentage"`
		Playing              bool    `json:"playing"`
		Filename             string  `json:"filename"`
		Path                 string  `json:"path"`
		Duration             int     `json:"duration"` // in ms
		Filepath             string  `json:"filepath"`

		CurrentTimeInSeconds float64 `json:"currentTimeInSeconds"` // in seconds
		DurationInSeconds    float64 `json:"durationInSeconds"`    // in seconds

		PlaybackType PlaybackType `json:"playbackType"` // "file", "stream"
	}
)

func (e TrackingStartedEvent) Type() string          { return "tracking_started" }
func (e TrackingRetryEvent) Type() string            { return "tracking_retry" }
func (e VideoCompletedEvent) Type() string           { return "video_completed" }
func (e TrackingStoppedEvent) Type() string          { return "tracking_stopped" }
func (e PlaybackStatusEvent) Type() string           { return "playback_status" }
func (e StreamingTrackingStartedEvent) Type() string { return "streaming_tracking_started" }
func (e StreamingTrackingRetryEvent) Type() string   { return "streaming_tracking_retry" }
func (e StreamingVideoCompletedEvent) Type() string  { return "streaming_video_completed" }
func (e StreamingTrackingStoppedEvent) Type() string { return "streaming_tracking_stopped" }
func (e StreamingPlaybackStatusEvent) Type() string  { return "streaming_playback_status" }

func NewRepository(opts *NewRepositoryOptions) *Repository {

	return &Repository{
		Logger:                opts.Logger,
		Default:               opts.Default,
		VLC:                   opts.VLC,
		MpcHc:                 opts.MpcHc,
		Mpv:                   opts.Mpv,
		Iina:                  opts.Iina,
		wsEventManager:        opts.WSEventManager,
		continuityManager:     opts.ContinuityManager,
		completionThreshold:   0.8,
		subscribers:           result.NewResultMap[string, *RepositorySubscriber](),
		currentPlaybackStatus: &PlaybackStatus{},
		exitedCh:              make(chan struct{}),
	}
}

func (m *Repository) Subscribe(id string) *RepositorySubscriber {
	sub := &RepositorySubscriber{
		EventCh: make(chan MediaPlayerEvent, 10), // Buffered channel to avoid blocking
	}
	m.subscribers.Set(id, sub)
	return sub
}

func (m *Repository) Unsubscribe(id string) {
	m.subscribers.Delete(id)
}

func (m *Repository) GetStatus() *PlaybackStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentPlaybackStatus
}

// PullStatus returns the current playback status directly from the media player.
func (m *Repository) PullStatus() (*PlaybackStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, err := m.getStatus()
	if err != nil {
		return nil, false
	}

	var ok bool
	if m.currentPlaybackStatus == nil {
		return nil, false
	}

	if m.currentPlaybackStatus.PlaybackType == PlaybackTypeFile {
		ok = m.processStatus(m.Default, status)
	} else {
		ok = m.processStreamStatus(m.Default, status)
	}
	return m.currentPlaybackStatus, ok
}

func (m *Repository) IsRunning() bool {
	return m.isRunning
}

func (m *Repository) GetExecutablePath() string {
	switch m.Default {
	case "vlc":
		return m.VLC.GetExecutablePath()
	case "mpc-hc":
		return m.MpcHc.GetExecutablePath()
	case "mpv":
		return m.Mpv.GetExecutablePath()
	case "iina":
		return m.Iina.GetExecutablePath()
	}
	return ""
}

func (m *Repository) GetDefault() string {
	return m.Default
}

// Play will start the media player and load the video at the given path.
// The implementation of the specific media player is handled by the respective media player package.
// Calling it multiple *should* not open multiple instances of the media player -- subsequent calls should just load a new video if the media player is already open.
func (m *Repository) Play(path string) error {

	m.Logger.Debug().Str("path", path).Msg("media player: Media requested")

	lastWatched := m.continuityManager.GetExternalPlayerEpisodeWatchHistoryItem(path, false, 0, 0)

	switch m.Default {
	case "vlc":
		err := m.VLC.Start()
		if err != nil {
			m.Logger.Error().Err(err).Msg("media player: Could not start media player using VLC")
			return fmt.Errorf("could not start VLC, %w", err)
		}

		err = m.VLC.AddAndPlay(path)
		if err != nil {
			m.Logger.Error().Err(err).Msg("media player: Could not open and play video using VLC")
			if m.VLC.Path != "" {
				return fmt.Errorf("could not open and play video, %w", err)
			} else {
				return fmt.Errorf("could not open and play video, %w", err)
			}
		}

		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			if lastWatched.Found {
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.ForcePause()
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.Seek(fmt.Sprintf("%d", int(lastWatched.Item.CurrentTime)))
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.Resume()
			}
		}

		return nil
	case "mpc-hc":
		err := m.MpcHc.Start()
		if err != nil {
			m.Logger.Error().Err(err).Msg("media player: Could not start media player using MPC-HC")
			return fmt.Errorf("could not start MPC-HC, %w", err)
		}
		_, err = m.MpcHc.OpenAndPlay(path)
		if err != nil {
			m.Logger.Error().Err(err).Msg("media player: Could not open and play video using MPC-HC")
			return fmt.Errorf("could not open and play video, %w", err)
		}

		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			if lastWatched.Found {
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Pause()
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Seek(int(lastWatched.Item.CurrentTime))
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Play()
			}
		}

		return nil
	case "mpv":
		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			var args []string
			if lastWatched.Found {
				//args = append(args, "--no-resume-playback", fmt.Sprintf("--start=+%d", int(lastWatched.Item.CurrentTime)))
				args = append(args, "--no-resume-playback")
			}
			err := m.Mpv.OpenAndPlay(path, args...)
			if err != nil {
				m.Logger.Error().Err(err).Msg("media player: Could not open and play video using MPV")
				return fmt.Errorf("could not open and play video, %w", err)
			}
			if lastWatched.Found {
				_ = m.Mpv.SeekTo(lastWatched.Item.CurrentTime)
			}
		} else {
			err := m.Mpv.OpenAndPlay(path)
			if err != nil {
				m.Logger.Error().Err(err).Msg("media player: Could not open and play video using MPV")
				return fmt.Errorf("could not open and play video, %w", err)
			}
		}

		return nil
	case "iina":
		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			var args []string
			if lastWatched.Found {
				//args = append(args, "--mpv-no-resume-playback", fmt.Sprintf("--mpv-start=+%d", int(lastWatched.Item.CurrentTime)))
				args = append(args, "--mpv-no-resume-playback")
			}
			err := m.Iina.OpenAndPlay(path, args...)
			if err != nil {
				m.Logger.Error().Err(err).Msg("media player: Could not open and play video using IINA")
				return fmt.Errorf("could not open and play video, %w", err)
			}
			if lastWatched.Found {
				_ = m.Iina.SeekTo(lastWatched.Item.CurrentTime)
			}
		} else {
			err := m.Iina.OpenAndPlay(path)
			if err != nil {
				m.Logger.Error().Err(err).Msg("media player: Could not open and play video using IINA")
				return fmt.Errorf("could not open and play video, %w", err)
			}
		}

		return nil
	default:
		return errors.New("no default media player set")
	}

}

func (m *Repository) Pause() error {
	switch m.Default {
	case "vlc":
		return m.VLC.Pause()
	case "mpc-hc":
		return m.MpcHc.Pause()
	case "mpv":
		return m.Mpv.Pause()
	case "iina":
		return m.Iina.Pause()
	default:
		return errors.New("no default media player set")
	}
}

func (m *Repository) Resume() error {
	switch m.Default {
	case "vlc":
		return m.VLC.Resume()
	case "mpc-hc":
		return m.MpcHc.Play()
	case "mpv":
		return m.Mpv.Resume()
	case "iina":
		return m.Iina.Resume()
	default:
		return errors.New("no default media player set")
	}
}

func (m *Repository) Seek(seconds float64) error {
	switch m.Default {
	case "vlc":
		return m.VLC.Seek(fmt.Sprintf("%d", int(seconds)))
	case "mpc-hc":
		return m.MpcHc.Seek(int(seconds))
	case "mpv":
		return m.Mpv.Seek(seconds)
	case "iina":
		return m.Iina.Seek(seconds)
	default:
		return errors.New("no default media player set")
	}
}

func (m *Repository) Stream(streamUrl string, episode int, mediaId int, windowTitle string) error {

	m.Logger.Debug().Str("streamUrl", streamUrl).Msg("media player: Stream requested")
	var err error

	switch m.Default {
	case "vlc":
		err = m.VLC.Start()
	case "mpc-hc":
		err = m.MpcHc.Start()
		_, err = m.MpcHc.OpenAndPlay(streamUrl)
	case "mpv":
		// MPV does not need to be started
	case "iina":
		// IINA does not need to be started
	default:
		return errors.New("no default media player set")
	}

	if err != nil {
		m.Logger.Error().Err(err).Msg("media player: Could not start media player for stream")
		return fmt.Errorf("could not open media player, %w", err)
	}

	lastWatched := m.continuityManager.GetExternalPlayerEpisodeWatchHistoryItem("", true, episode, mediaId)

	switch m.Default {
	case "vlc":
		err = m.VLC.AddAndPlay(streamUrl)

		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			if lastWatched.Found {
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.ForcePause()
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.Seek(fmt.Sprintf("%d", int(lastWatched.Item.CurrentTime)))
				time.Sleep(400 * time.Millisecond)
				_ = m.VLC.Resume()
			}
		}

	case "mpc-hc":
		_, err = m.MpcHc.OpenAndPlay(streamUrl)

		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			if lastWatched.Found {
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Pause()
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Seek(int(lastWatched.Item.CurrentTime))
				time.Sleep(400 * time.Millisecond)
				_ = m.MpcHc.Play()
			}
		}

	case "mpv":
		args := []string{}
		if windowTitle != "" {
			args = append(args, fmt.Sprintf("--title=%q", windowTitle))
		}
		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			err = m.Mpv.OpenAndPlay(streamUrl, args...)
			if lastWatched.Found {
				_ = m.Mpv.SeekTo(lastWatched.Item.CurrentTime)
			}
		} else {
			err = m.Mpv.OpenAndPlay(streamUrl, args...)
		}

	case "iina":
		args := []string{}
		if windowTitle != "" {
			args = append(args, fmt.Sprintf("--mpv-title=%q", windowTitle))
		}
		if m.continuityManager.GetSettings().WatchContinuityEnabled {
			err = m.Iina.OpenAndPlay(streamUrl, args...)
			if lastWatched.Found {
				_ = m.Iina.SeekTo(lastWatched.Item.CurrentTime)
			}
		} else {
			err = m.Iina.OpenAndPlay(streamUrl, args...)
		}

	}

	if err != nil {
		m.Logger.Error().Err(err).Msg("media player: Could not open and play stream")
		return fmt.Errorf("could not open and play stream, %w", err)
	}

	return nil
}

// Cancel will stop the tracking process and publish an "abnormal" event
func (m *Repository) Cancel() {
	m.mu.Lock()
	if m.cancel != nil {
		m.Logger.Debug().Msg("media player: Cancel request received")
		m.cancel()
		m.trackingStopped("Something went wrong, tracking cancelled")
	} else {
		m.Logger.Debug().Msg("media player: Cancel request received, but no context found")
	}
	// Close MPV if it's the default player
	if m.Default == "mpv" {
		m.Mpv.CloseAll()
	}
	m.mu.Unlock()
}

// Stop will stop the tracking process and publish a "normal" event
func (m *Repository) Stop() {
	m.mu.Lock()
	if m.cancel != nil {
		m.Logger.Debug().Msg("media player: Stop request received")
		m.cancel()
		m.cancel = nil
		m.trackingStopped("Tracking stopped")
		// Close MPV if it's the default player
		if m.Default == "mpv" {
			go m.Mpv.CloseAll()
		}
	}
	m.mu.Unlock()
}

// StartTrackingTorrentStream will start tracking media player status for torrent streaming
func (m *Repository) StartTrackingTorrentStream() {
	m.mu.Lock()
	// If a previous context exists, cancel it
	if m.cancel != nil {
		m.Logger.Debug().Msg("media player: Cancelling previous context")
		m.cancel()
	}

	// Create a new context
	var trackingCtx context.Context
	trackingCtx, m.cancel = context.WithCancel(context.Background())

	done := make(chan struct{})
	var filename string
	var completed bool
	var retries int

	hookEvent := &MediaPlayerStreamTrackingRequestedEvent{
		StartRefreshDelay:    3,
		RefreshDelay:         1,
		MaxRetries:           5,
		MaxRetriesAfterStart: 5,
	}
	_ = hook.GlobalHookManager.OnMediaPlayerStreamTrackingRequested().Trigger(hookEvent)
	startRefreshDelay := hookEvent.StartRefreshDelay
	maxTries := hookEvent.MaxRetries
	refreshDelay := hookEvent.RefreshDelay
	maxRetriesAfterStart := hookEvent.MaxRetriesAfterStart

	// Default prevented, do not track
	if hookEvent.DefaultPrevented {
		m.Logger.Debug().Msg("media player: Tracking cancelled by hook")
		return
	}

	// Unlike normal tracking when the file is downloaded, we may need to wait a bit before we can get the status,
	// so we won't count retries until it's confirmed that the file has started playing.
	var trackingStarted bool
	var waitInSeconds int

	m.isRunning = true
	gotFirstStatus := false

	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			m.isRunning = false
			if m.cancel != nil {
				m.cancel()
			}
			m.mu.Unlock()
		}()
		for {
			select {
			case <-done:
				m.mu.Lock()
				m.Logger.Debug().Msg("media player: Connection lost")
				m.isRunning = false
				m.mu.Unlock()
				return
			case <-trackingCtx.Done():
				m.mu.Lock()
				m.Logger.Debug().Msg("media player: Context cancelled")
				m.isRunning = false
				m.mu.Unlock()
				return
			//case <-m.exitedCh:
			//	m.mu.Lock()
			//	m.Logger.Debug().Msg("media player: Player exited")
			//	m.isRunning = false
			//	m.streamingTrackingStopped(PlayerClosedEvent)
			//	m.mu.Unlock()
			//	return
			default:
				// Wait at least 3 seconds before we start checking the status
				if !gotFirstStatus {
					time.Sleep(time.Duration(startRefreshDelay) * time.Second)
				} else {
					time.Sleep(time.Duration(refreshDelay) * time.Second)
				}
				status, err := m.getStatus()
				if err != nil {
					if !trackingStarted {
						if waitInSeconds > 60 {
							m.Logger.Warn().Msg("media player: Ending goroutine, waited too long")
							return
						}
						m.Logger.Trace().Msgf("media player: Waiting for stream, %d seconds", waitInSeconds)
						waitInSeconds += refreshDelay
						continue
					} else {
						m.streamingTrackingRetry("Failed to get player status")
						m.Logger.Error().Msgf("media player: Failed to get player status, retrying (%d/%d)", retries+1, maxTries)

						// Video is completed, and we are unable to get the status
						// We can safely assume that the player has been closed
						if retries == 1 && (completed || m.continuityManager.GetSettings().WatchContinuityEnabled) {
							m.Logger.Debug().Msg("media player: Sending player closed event")
							m.streamingTrackingStopped(PlayerClosedEvent)
							close(done)
							break
						}

						if retries >= maxTries-1 {
							m.Logger.Debug().Msg("media player: Sending failed status query event")
							m.streamingTrackingStopped("Failed to get player status")
							close(done)
							break
						}
						retries++
						continue
					}
				}

				trackingStarted = true
				ok := m.processStreamStatus(m.Default, status)

				if !ok {
					m.streamingTrackingRetry("Failed to get player status")
					m.Logger.Error().Interface("status", status).Msgf("media player: Failed to process status, retrying (%d/%d)", retries+1, maxRetriesAfterStart)
					if retries >= maxRetriesAfterStart-1 {
						m.Logger.Debug().Msg("media player: Sending failed status query event")
						m.streamingTrackingStopped("Failed to process status")
						close(done)
						break
					}
					retries++
					continue
				}

				// New video has started playing \/
				if filename == "" || filename != m.currentPlaybackStatus.Filename {
					m.Logger.Debug().Str("previousFilename", filename).Str("newFilename", m.currentPlaybackStatus.Filename).Msg("media player: Video loaded")
					m.streamingTrackingStarted(m.currentPlaybackStatus)
					filename = m.currentPlaybackStatus.Filename
					completed = false
				}

				// Video completed \/
				if m.currentPlaybackStatus.CompletionPercentage > m.completionThreshold && !completed {
					m.Logger.Debug().Msg("media player: Video completed")
					m.streamingVideoCompleted(m.currentPlaybackStatus)
					completed = true
				}

				m.streamingPlaybackStatus(m.currentPlaybackStatus)
			}
		}
	}()
}

// StartTracking will start tracking media player status.
// This method is safe to call multiple times -- it will cancel the previous context and start a new one.
func (m *Repository) StartTracking() {
	m.mu.Lock()
	// If a previous context exists, cancel it
	if m.cancel != nil {
		m.Logger.Debug().Msg("media player: Cancelling previous context")
		m.cancel()
	}

	// Create a new context
	var trackingCtx context.Context
	trackingCtx, m.cancel = context.WithCancel(context.Background())

	done := make(chan struct{})
	var filename string
	var completed bool
	var retries int

	hookEvent := &MediaPlayerLocalFileTrackingRequestedEvent{
		StartRefreshDelay: 3,
		RefreshDelay:      1,
		MaxRetries:        5,
	}
	_ = hook.GlobalHookManager.OnMediaPlayerLocalFileTrackingRequested().Trigger(hookEvent)
	startRefreshDelay := hookEvent.StartRefreshDelay
	maxTries := hookEvent.MaxRetries
	refreshDelay := hookEvent.RefreshDelay

	// Default prevented, do not track
	if hookEvent.DefaultPrevented {
		m.Logger.Debug().Msg("media player: Tracking cancelled by hook")
		return
	}

	m.isRunning = true
	gotFirstStatus := false

	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-done:
				m.mu.Lock()
				m.Logger.Debug().Msg("media player: Connection lost")
				m.isRunning = false
				m.mu.Unlock()
				if m.cancel != nil {
					m.cancel()
					m.cancel = nil
				}
				return
			case <-trackingCtx.Done():
				m.mu.Lock()
				m.Logger.Debug().Msg("media player: Context cancelled")
				m.isRunning = false
				m.cancel = nil
				m.mu.Unlock()
				return
			//case <-m.exitedCh:
			//	m.mu.Lock()
			//	m.Logger.Debug().Msg("media player: Player exited")
			//	m.isRunning = false
			//	m.trackingStopped(PlayerClosedEvent)
			//	m.mu.Unlock()
			//	return
			default:
				// Wait at least X seconds before we start checking the status
				if !gotFirstStatus {
					time.Sleep(time.Duration(startRefreshDelay) * time.Second)
				} else {
					time.Sleep(time.Duration(refreshDelay) * time.Second)
				}
				status, err := m.getStatus()
				if err != nil {
					m.trackingRetry("Failed to get player status")
					m.Logger.Error().Msgf("media player: Failed to get player status, retrying (%d/%d)", retries+1, maxTries)

					// Video is completed, and we are unable to get the status
					// We can safely assume that the player has been closed
					if retries == 1 && (completed || m.continuityManager.GetSettings().WatchContinuityEnabled) {
						m.trackingStopped(PlayerClosedEvent)
						close(done)
						break
					}

					if retries >= maxTries-1 {
						m.trackingStopped("Failed to get player status")
						close(done)
						break
					}
					retries++
					continue
				}

				gotFirstStatus = true

				ok := m.processStatus(m.Default, status)

				if !ok {
					m.trackingRetry("Failed to get player status")
					m.Logger.Error().Interface("status", status).Msgf("media player: Failed to process status, retrying (%d/%d)", retries+1, maxTries)
					if retries >= maxTries-1 {
						m.trackingStopped("Failed to process status")
						close(done)
						break
					}
					retries++
					continue
				}

				// New video has started playing \/
				if filename == "" || filename != m.currentPlaybackStatus.Filename {
					m.Logger.Debug().Str("previousFilename", filename).Str("newFilename", m.currentPlaybackStatus.Filename).Msg("media player: Video started playing")
					m.Logger.Debug().Interface("currentPlaybackStatus", m.currentPlaybackStatus).Msg("media player: Playback status")
					m.trackingStarted(m.currentPlaybackStatus)
					filename = m.currentPlaybackStatus.Filename
					completed = false
				}

				// Video completed \/
				if m.currentPlaybackStatus.CompletionPercentage > m.completionThreshold && !completed {
					m.Logger.Debug().Msg("media player: Video completed")
					m.Logger.Debug().Interface("currentPlaybackStatus", m.currentPlaybackStatus).Msg("media player: Playback status")
					m.videoCompleted(m.currentPlaybackStatus)
					completed = true
				}

				m.playbackStatus(m.currentPlaybackStatus)
			}
		}
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Repository) trackingStopped(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- TrackingStoppedEvent{Reason: reason}
		return true
	})
}

func (m *Repository) trackingStarted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- TrackingStartedEvent{Status: status}
		return true
	})
}

func (m *Repository) trackingRetry(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- TrackingRetryEvent{Reason: reason}
		return true
	})
}

func (m *Repository) videoCompleted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- VideoCompletedEvent{Status: status}
		return true
	})
}

func (m *Repository) playbackStatus(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- PlaybackStatusEvent{Status: status}
		return true
	})
}

func (m *Repository) streamingTrackingStopped(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- StreamingTrackingStoppedEvent{Reason: reason}
		return true
	})
}

func (m *Repository) streamingTrackingStarted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- StreamingTrackingStartedEvent{Status: status}
		return true
	})
}

func (m *Repository) streamingTrackingRetry(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- StreamingTrackingRetryEvent{Reason: reason}
		return true
	})
}

func (m *Repository) streamingVideoCompleted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- StreamingVideoCompletedEvent{Status: status}
		return true
	})
}

func (m *Repository) streamingPlaybackStatus(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.EventCh <- StreamingPlaybackStatusEvent{Status: status}
		return true
	})
}

func (m *Repository) getStatus() (interface{}, error) {
	switch m.Default {
	case "vlc":
		return m.VLC.GetStatus()
	case "mpc-hc":
		return m.MpcHc.GetVariables()
	case "mpv":
		return m.Mpv.GetPlaybackStatus()
	case "iina":
		return m.Iina.GetPlaybackStatus()
	}
	return nil, errors.New("unsupported media player")
}

func (m *Repository) processStatus(player string, status interface{}) bool {
	m.currentPlaybackStatus.PlaybackType = PlaybackTypeFile
	switch player {
	case "vlc":
		// Process VLC status
		st, ok := status.(*vlc2.Status)
		if !ok || st == nil {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position
		m.currentPlaybackStatus.Playing = st.State == "playing"
		m.currentPlaybackStatus.Filename = st.Information.Category["meta"].Filename
		m.currentPlaybackStatus.Duration = int(st.Length * 1000)
		m.currentPlaybackStatus.Filepath = "" // VLC does not provide the filepath

		m.currentPlaybackStatus.CurrentTimeInSeconds = float64(st.Time)
		m.currentPlaybackStatus.DurationInSeconds = float64(st.Length)
		return true
	case "mpc-hc":
		// Process MPC-HC status
		st, ok := status.(*mpchc2.Variables)
		if !ok || st == nil || st.Duration == 0 {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = st.State == 2
		m.currentPlaybackStatus.Filename = st.File
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.FilePath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position / 1000
		m.currentPlaybackStatus.DurationInSeconds = st.Duration / 1000

		return true
	case "mpv":
		// Process MPV status
		st, ok := status.(*mpv.Playback)
		if !ok || st == nil || st.Duration == 0 || st.IsRunning == false {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = !st.Paused
		m.currentPlaybackStatus.Filename = st.Filename
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.Filepath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position
		m.currentPlaybackStatus.DurationInSeconds = st.Duration

		return true
	case "iina":
		// Process IINA status
		st, ok := status.(*iina.Playback)
		if !ok || st == nil || st.Duration == 0 || st.IsRunning == false {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = !st.Paused
		m.currentPlaybackStatus.Filename = st.Filename
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.Filepath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position
		m.currentPlaybackStatus.DurationInSeconds = st.Duration

		return true
	default:
		return false
	}
}

func (m *Repository) processStreamStatus(player string, status interface{}) bool {
	m.currentPlaybackStatus.PlaybackType = PlaybackTypeStream
	switch player {
	case "vlc":
		// Process VLC status
		st, ok := status.(*vlc2.Status)
		if !ok || st == nil {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position
		m.currentPlaybackStatus.Playing = st.State == "playing"
		m.currentPlaybackStatus.Filename = st.Information.Category["meta"].Filename
		m.currentPlaybackStatus.Duration = int(st.Length * 1000)
		m.currentPlaybackStatus.Filepath = st.Information.Category["meta"].Filename // VLC does not provide the filepath, use filename

		m.currentPlaybackStatus.CurrentTimeInSeconds = float64(st.Time)
		m.currentPlaybackStatus.DurationInSeconds = float64(st.Length)

		return true
	case "mpc-hc":
		// Process MPC-HC status
		st, ok := status.(*mpchc2.Variables)
		if !ok || st == nil {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = st.State == 2
		m.currentPlaybackStatus.Filename = st.File
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.FilePath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position / 1000
		m.currentPlaybackStatus.DurationInSeconds = st.Duration / 1000

		return true
	case "mpv":
		// Process MPV status
		st, ok := status.(*mpv.Playback)
		if !ok || st == nil || st.Duration == 0 || st.IsRunning == false {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = !st.Paused
		m.currentPlaybackStatus.Filename = st.Filename
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.Filepath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position
		m.currentPlaybackStatus.DurationInSeconds = st.Duration

		return true
	case "iina":
		// Process IINA status
		st, ok := status.(*iina.Playback)
		if !ok || st == nil || st.Duration == 0 || st.IsRunning == false {
			return false
		}

		m.currentPlaybackStatus.CompletionPercentage = st.Position / st.Duration
		m.currentPlaybackStatus.Playing = !st.Paused
		m.currentPlaybackStatus.Filename = st.Filename
		m.currentPlaybackStatus.Duration = int(st.Duration)
		m.currentPlaybackStatus.Filepath = st.Filepath

		m.currentPlaybackStatus.CurrentTimeInSeconds = st.Position
		m.currentPlaybackStatus.DurationInSeconds = st.Duration

		return true
	default:
		return false
	}
}
