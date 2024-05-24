package mediaplayer

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	mpchc2 "github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	vlc2 "github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/util/result"
	"sync"
	"time"
)

type (
	// Repository provides a common interface to interact with media players
	Repository struct {
		Logger                *zerolog.Logger
		Default               string
		VLC                   *vlc2.VLC
		MpcHc                 *mpchc2.MpcHc
		Mpv                   *mpv.Mpv
		WSEventManager        events.WSEventManagerInterface
		completionThreshold   float64
		mu                    sync.Mutex
		isRunning             bool
		currentPlaybackStatus *PlaybackStatus
		subscribers           *result.Map[string, *RepositorySubscriber]
		cancel                context.CancelFunc
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		Default        string
		VLC            *vlc2.VLC
		MpcHc          *mpchc2.MpcHc
		Mpv            *mpv.Mpv
		WSEventManager events.WSEventManagerInterface
	}

	RepositorySubscriber struct {
		TrackingStartedCh chan *PlaybackStatus
		TrackingRetryCh   chan string
		VideoCompletedCh  chan *PlaybackStatus
		TrackingStoppedCh chan string
		PlaybackStatusCh  chan *PlaybackStatus

		StreamingTrackingStartedCh chan *PlaybackStatus
		StreamingTrackingRetryCh   chan string
		StreamingVideoCompletedCh  chan *PlaybackStatus
		StreamingTrackingStoppedCh chan string
		StreamingPlaybackStatusCh  chan *PlaybackStatus
	}

	PlaybackStatus struct {
		CompletionPercentage float64 `json:"completionPercentage"`
		Playing              bool    `json:"playing"`
		Filename             string  `json:"filename"`
		Path                 string  `json:"path"`
		Duration             int     `json:"duration"` // in ms
		Filepath             string  `json:"filepath"`
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	return &Repository{
		Logger:              opts.Logger,
		Default:             opts.Default,
		VLC:                 opts.VLC,
		MpcHc:               opts.MpcHc,
		Mpv:                 opts.Mpv,
		WSEventManager:      opts.WSEventManager,
		completionThreshold: 0.8,
		subscribers:         result.NewResultMap[string, *RepositorySubscriber](),
	}
}

func (m *Repository) Subscribe(id string) *RepositorySubscriber {
	sub := &RepositorySubscriber{
		TrackingStartedCh:          make(chan *PlaybackStatus, 1),
		TrackingRetryCh:            make(chan string, 1),
		VideoCompletedCh:           make(chan *PlaybackStatus, 1),
		TrackingStoppedCh:          make(chan string, 1),
		PlaybackStatusCh:           make(chan *PlaybackStatus, 1),
		StreamingTrackingStartedCh: make(chan *PlaybackStatus, 1),
		StreamingTrackingRetryCh:   make(chan string, 1),
		StreamingVideoCompletedCh:  make(chan *PlaybackStatus, 1),
		StreamingTrackingStoppedCh: make(chan string, 1),
		StreamingPlaybackStatusCh:  make(chan *PlaybackStatus, 1),
	}
	m.subscribers.Set(id, sub)
	return sub
}

func (m *Repository) GetStatus() *PlaybackStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentPlaybackStatus
}

func (m *Repository) IsRunning() bool {
	return m.isRunning
}

// Play will start the media player and load the video at the given path.
// The implementation of the specific media player is handled by the respective media player package.
// Calling it multiple *should* not open multiple instances of the media player -- subsequent calls should just load a new video if the media player is already open.
func (m *Repository) Play(path string) error {

	m.Logger.Debug().Str("path", path).Msg("media player: Media requested")

	switch m.Default {
	case "vlc":
		err := m.VLC.Start()
		if err != nil {
			return errors.New("could not start media player")
		}
		err = m.VLC.AddAndPlay(path)
		if err != nil {
			return errors.New("could not open and play video, verify your settings")
		}
		return nil
	case "mpc-hc":
		err := m.MpcHc.Start()
		if err != nil {
			return errors.New("could not start media player")
		}
		_, err = m.MpcHc.OpenAndPlay(path)
		if err != nil {
			return errors.New("could not open and play video, verify your settings")
		}
		return nil
	case "mpv":
		err := m.Mpv.OpenAndPlay(path)
		if err != nil {
			return fmt.Errorf("could not open and play video, %s", err.Error())
		}
		return nil
	default:
		return errors.New("no default media player set")
	}

}

func (m *Repository) Stream(streamUrl string) error {

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
	default:
		return errors.New("no default media player set")
	}

	if err != nil {
		return fmt.Errorf("could not open media player, %s", err.Error())
	}

	switch m.Default {
	case "vlc":
		err = m.VLC.AddAndPlay(streamUrl)
	case "mpc-hc":
		_, err = m.MpcHc.OpenAndPlay(streamUrl)
	case "mpv":
		err = m.Mpv.OpenAndStream(streamUrl, "--no-cache", "--force-window")
	}

	if err != nil {
		return fmt.Errorf("could not open and play video, %s", err.Error())
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
		m.trackingStopped("Tracking stopped")
	}
	// Close MPV if it's the default player
	if m.Default == "mpv" {
		m.Mpv.CloseAll()
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

	// Unlike normal tracking when the file is downloaded, we may need to wait a bit before we can get the status
	// So we need to keep track of whether we have started tracking
	// Unlike normal tracking we won't count retries until we have started tracking
	var trackingStarted bool
	var waitInSeconds int

	m.isRunning = true

	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			m.isRunning = false
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
			default:
				time.Sleep(3 * time.Second)
				status, err := m.getStatus()
				//fmt.Printf("status: %v\n", status)
				if err != nil {
					if !trackingStarted {
						if waitInSeconds > 60 {
							m.Logger.Warn().Msg("media player: Ending goroutine, waited too long")
							return
						}
						m.Logger.Trace().Msgf("media player: Waiting for torrent file, %d seconds", waitInSeconds)
						waitInSeconds += 3
						continue
					} else {
						m.trackingRetry("Failed to get player status")
						m.Logger.Error().Msgf("media player: Failed to get player status, retrying (%d/%d)", retries+1, 3)
						// Video is completed, and we are unable to get the status
						// We can safely assume that the player has been closed
						if retries == 1 && completed {
							m.Logger.Debug().Msg("media player: Sending player closed event")
							m.streamingTrackingStopped("Player closed")
							close(done)
							break
						}

						if retries >= 2 {
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
				playback, ok := m.processStreamStatus(m.Default, status)

				if !ok {
					m.streamingTrackingRetry("Failed to get player status")
					m.Logger.Error().Msgf("media player: Failed to process status, retrying (%d/%d)", retries+1, 3)
					if retries >= 2 {
						m.Logger.Debug().Msg("media player: Sending failed status query event")
						m.streamingTrackingStopped("Failed to process status")
						close(done)
						break
					}
					retries++
					continue
				}

				m.currentPlaybackStatus = playback

				// New video has started playing \/
				if filename == "" || filename != playback.Filename {
					m.Logger.Debug().Msg("media player: Video loaded")
					m.streamingTrackingStarted(playback)
					filename = playback.Filename
					completed = false
				}

				// Video completed \/
				if playback.CompletionPercentage > m.completionThreshold && !completed {
					m.Logger.Debug().Msg("media player: Video completed")
					m.streamingVideoCompleted(playback)
					completed = true
				}

				m.streamingPlaybackStatus(playback)
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

	m.isRunning = true

	m.mu.Unlock()

	go func() {
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
			default:
				time.Sleep(3 * time.Second)
				status, err := m.getStatus()

				if err != nil {
					m.trackingRetry("Failed to get player status")
					m.Logger.Error().Msgf("media player: Failed to get player status, retrying (%d/%d)", retries+1, 3)

					// Video is completed, and we are unable to get the status
					// We can safely assume that the player has been closed
					if retries == 1 && completed {
						m.trackingStopped("Player closed")
						close(done)
						break
					}

					if retries >= 2 {
						m.trackingStopped("Failed to get player status")
						close(done)
						break
					}
					retries++
					continue
				}

				playback, ok := m.processStatus(m.Default, status)

				if !ok {
					m.trackingRetry("Failed to get player status")
					m.Logger.Error().Msgf("media player: Failed to process status, retrying (%d/%d)", retries+1, 3)
					if retries >= 2 {
						m.trackingStopped("Failed to process status")
						close(done)
						break
					}
					retries++
					continue
				}

				m.currentPlaybackStatus = playback

				// New video has started playing \/
				if filename == "" || filename != playback.Filename {
					m.Logger.Debug().Msg("media player: Video started playing")
					m.trackingStarted(playback)
					filename = playback.Filename
					completed = false
				}

				// Video completed \/
				if playback.CompletionPercentage > m.completionThreshold && !completed {
					m.Logger.Debug().Msg("media player: Video completed")
					m.videoCompleted(playback)
					completed = true
				}

				m.playbackStatus(playback)

			}
		}
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Repository) trackingStopped(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.TrackingStoppedCh <- reason
		return true
	})
}

func (m *Repository) trackingStarted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.TrackingStartedCh <- status
		return true
	})
}

func (m *Repository) trackingRetry(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.TrackingRetryCh <- reason
		return true
	})
}

func (m *Repository) videoCompleted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.VideoCompletedCh <- status
		return true
	})
}

func (m *Repository) playbackStatus(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.PlaybackStatusCh <- status
		return true
	})
}

func (m *Repository) streamingTrackingStopped(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.StreamingTrackingStoppedCh <- reason
		return true
	})
}

func (m *Repository) streamingTrackingStarted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.StreamingTrackingStartedCh <- status
		return true
	})
}

func (m *Repository) streamingTrackingRetry(reason string) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.StreamingTrackingRetryCh <- reason
		return true
	})
}

func (m *Repository) streamingVideoCompleted(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.StreamingVideoCompletedCh <- status
		return true
	})
}

func (m *Repository) streamingPlaybackStatus(status *PlaybackStatus) {
	m.subscribers.Range(func(key string, value *RepositorySubscriber) bool {
		value.StreamingPlaybackStatusCh <- status
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
	}
	return nil, errors.New("unsupported media player")
}

func (m *Repository) processStatus(player string, status interface{}) (*PlaybackStatus, bool) {
	switch player {
	case "vlc":
		// Process VLC status
		st := status.(*vlc2.Status)
		if st == nil {
			return nil, false
		}

		ret := &PlaybackStatus{
			CompletionPercentage: st.Position,
			Playing:              st.State == "playing",
			Filename:             st.Information.Category["meta"].Filename,
			Duration:             int(st.Length * 1000),
			Filepath:             "", // VLC does not provide the filepath
		}

		return ret, true
	case "mpc-hc":
		// Process MPC-HC status
		st := status.(*mpchc2.Variables)
		if st == nil || st.Duration == 0 {
			return nil, false
		}
		ret := &PlaybackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              st.State == 2,
			Filename:             st.File,
			Duration:             int(st.Duration),
			Filepath:             st.FilePath,
		}

		return ret, true
	case "mpv":
		// Process MPV status
		st := status.(*mpv.Playback)
		if st == nil || st.Duration == 0 || st.IsRunning == false {
			return nil, false
		}
		ret := &PlaybackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              !st.Paused,
			Filename:             st.Filename,
			Duration:             int(st.Duration),
			Filepath:             st.Filepath,
		}

		return ret, true
	default:
		return nil, false
	}
}

func (m *Repository) processStreamStatus(player string, status interface{}) (*PlaybackStatus, bool) {
	switch player {
	case "vlc":
		// Process VLC status
		st := status.(*vlc2.Status)
		if st == nil {
			return nil, false
		}

		ret := &PlaybackStatus{
			CompletionPercentage: st.Position,
			Playing:              st.State == "playing",
			Filename:             st.Information.Category["meta"].Filename,
			Duration:             int(st.Length * 1000),
			Filepath:             "", // VLC does not provide the filepath
		}

		return ret, true
	case "mpc-hc":
		// Process MPC-HC status
		st := status.(*mpchc2.Variables)
		if st == nil {
			return nil, false
		}
		ret := &PlaybackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              st.State == 2,
			Filename:             st.File,
			Duration:             int(st.Duration),
			Filepath:             st.FilePath,
		}

		return ret, true
	case "mpv":
		// Process MPV status
		st := status.(*mpv.Playback)
		if st == nil || st.IsRunning == false {
			return nil, false
		}
		ret := &PlaybackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              !st.Paused,
			Filename:             st.Filename,
			Duration:             int(st.Duration),
			Filepath:             st.Filepath,
		}

		return ret, true
	default:
		return nil, false
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
