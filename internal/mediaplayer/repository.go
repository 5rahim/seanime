package mediaplayer

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/vlc"
	"sync"
	"time"
)

type (
	// Repository provides a common interface to interact with media players
	Repository struct {
		Logger                *zerolog.Logger
		Default               string
		VLC                   *vlc.VLC
		MpcHc                 *mpchc.MpcHc
		Mpv                   *mpv.Mpv
		WSEventManager        events.IWSEventManager
		completionThreshold   float64
		mu                    sync.Mutex
		isRunning             bool
		currentPlaybackStatus *PlaybackStatus
		subscribers           map[string]*RepositorySubscriber
		cancel                context.CancelFunc
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		Default        string
		VLC            *vlc.VLC
		MpcHc          *mpchc.MpcHc
		Mpv            *mpv.Mpv
		WSEventManager events.IWSEventManager
	}

	RepositorySubscriber struct {
		TrackingStartedCh chan *PlaybackStatus
		TrackingRetryCh   chan string
		VideoCompletedCh  chan *PlaybackStatus
		TrackingStoppedCh chan string
		PlaybackStatusCh  chan *PlaybackStatus
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
		completionThreshold: 0.9,
		subscribers:         make(map[string]*RepositorySubscriber),
	}
}

func (m *Repository) Subscribe(id string) *RepositorySubscriber {
	sub := &RepositorySubscriber{
		TrackingStartedCh: make(chan *PlaybackStatus, 1),
		TrackingRetryCh:   make(chan string, 1),
		VideoCompletedCh:  make(chan *PlaybackStatus, 1),
		TrackingStoppedCh: make(chan string, 1),
		PlaybackStatusCh:  make(chan *PlaybackStatus, 1),
	}
	m.subscribers[id] = sub
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

func (m *Repository) trackingStopped(reason string) {
	for _, sub := range m.subscribers {
		sub.TrackingStoppedCh <- reason
	}
}

func (m *Repository) trackingStarted(status *PlaybackStatus) {
	for _, sub := range m.subscribers {
		sub.TrackingStartedCh <- status
	}
}

func (m *Repository) trackingRetry(reason string) {
	for _, sub := range m.subscribers {
		sub.TrackingRetryCh <- reason
	}
}

func (m *Repository) videoCompleted(status *PlaybackStatus) {
	for _, sub := range m.subscribers {
		sub.VideoCompletedCh <- status
	}
}

func (m *Repository) playbackStatus(status *PlaybackStatus) {
	for _, sub := range m.subscribers {
		sub.PlaybackStatusCh <- status
	}
}

func (m *Repository) Play(path string) error {

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
		err := m.Mpv.OpenAndPlay(path, mpv.StartExec)
		if err != nil {
			return fmt.Errorf("could not open and play video, %s", err.Error())
		}
		return nil
	default:
		return errors.New("no default media player set")
	}
}

func (m *Repository) StartTracking() {
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

	go func() {
		for {
			select {
			case <-done:
				m.Logger.Debug().Msg("media player: Connection lost")
				return
			case <-trackingCtx.Done():
				m.Logger.Debug().Msg("media player: Context cancelled")
				return
			default:
				time.Sleep(3 * time.Second)
				status, err := m.getStatus()

				if err != nil {
					m.Logger.Error().Msgf("media player: Failed to get status, retrying (%d/%d)", retries+1, 3)

					// Video is completed, and we are unable to get the status
					// We can safely assume that the player has been closed
					if retries == 1 && completed {
						m.trackingStopped("Player closed")
						close(done)
						break
					}

					if retries >= 2 {
						m.trackingStopped("Failed to get status")
						close(done)
						break
					}
					retries++
					continue
				}
				retries = 0

				playback, ok := m.processStatus(m.Default, status)

				if !ok {
					m.Logger.Error().Msgf("media player: Failed to get status, retrying (%d/%d)", retries+1, 3)
					if retries >= 2 {
						m.trackingStopped("Failed to process status")
						close(done)
						break
					}
					retries++
					continue
				}

				m.mu.Lock()
				m.currentPlaybackStatus = playback
				m.mu.Unlock()

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
		st := status.(*vlc.Status)
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
		st := status.(*mpchc.Variables)
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
