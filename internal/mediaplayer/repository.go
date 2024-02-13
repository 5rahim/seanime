package mediaplayer

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/vlc"
	"time"
)

type (
	Repository struct {
		Logger         *zerolog.Logger
		Default        string
		VLC            *vlc.VLC
		MpcHc          *mpchc.MpcHc
		Mpv            *mpv.Mpv
		WSEventManager events.IWSEventManager
	}

	playbackStatus struct {
		CompletionPercentage float64 `json:"completionPercentage"`
		Playing              bool    `json:"playing"`
		Filename             string  `json:"filename"`
		Duration             int     `json:"duration"` // in ms
	}
)

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
		err := m.Mpv.OpenAndPlay(path)
		if err != nil {
			return fmt.Errorf("could not open and play video, %s", err.Error())
		}
		return nil
	default:
		return errors.New("no default media player set")
	}
}

// StartTracking will start a goroutine that will monitor the status of the media player
func (m *Repository) StartTracking(onVideoCompleted func()) {
	// Create a channel to signal the goroutine to exit
	done := make(chan struct{})
	var filename string
	var completed bool
	var retries int
	var sRetries int
	var started bool

	m.WSEventManager.SendEvent(events.MediaPlayerTrackingStarted, nil)

	go func() {
		for {
			select {
			case <-done:
				m.Logger.Printf("Exiting media player status monitoring goroutine")
				return
			case <-time.After(3 * time.Second):
				var status interface{}
				var err error

				// Get the status based on the default player
				switch m.Default {
				case "vlc":
					status, err = m.VLC.GetStatus()
				case "mpc-hc":
					status, err = m.MpcHc.GetVariables()
				case "mpv":
					status, err = m.Mpv.GetPlaybackStatus()
				}

				if err != nil {
					// Retry 2 times before exiting, only if the tracking has started
					if started || (!started && retries >= 2) {
						if !started {
							m.WSEventManager.SendEvent(events.MediaPlayerTrackingStopped, "Failed to get status")
						} else {
							m.WSEventManager.SendEvent(events.MediaPlayerTrackingStopped, "Closed")
						}
						m.Logger.Error().Msg("mediaplayer: Failed to get status")
						m.Logger.Debug().Msg("mediaplayer: Tracking stopped")
						switch m.Default {
						case "vlc":
							m.VLC.Stop()
						case "mpc-hc":
							m.MpcHc.Stop()
						case "mpv":
							m.Mpv.Close()
						}
						close(done) // Signal to exit the goroutine
						return
					}
					if !started {
						retries++
						m.Logger.Error().Msgf("mediaplayer: Failed to get status, retrying (%d/%d)", retries, 2)
						continue
					}
				}

				// Process the status
				playback, ok := m.processStatus(m.Default, status)

				// Signal that the tracking has started
				if !started && err == nil && ok {
					started = true
				}

				if !ok {
					if started || (!started && sRetries >= 2) {
						if !started {
							m.WSEventManager.SendEvent(events.MediaPlayerTrackingStopped, "Failed to process status")
						} else {
							m.WSEventManager.SendEvent(events.MediaPlayerTrackingStopped, "Closed")
						}
						m.Logger.Error().Msg("mediaplayer: Failed to process status")
						m.Logger.Debug().Msg("mediaplayer: Tracking stopped")
						switch m.Default {
						case "vlc":
							m.VLC.Stop()
						case "mpc-hc":
							m.MpcHc.Stop()
						case "mpv":
							m.Mpv.Close()
						}
						close(done) // Signal to exit the goroutine
						return
					}
					if !started {
						sRetries++
						m.Logger.Error().Msgf("mediaplayer: Failed to process status, retrying (%d/%d)", sRetries, 2)
						continue
					}
				}

				if filename == "" {
					m.WSEventManager.SendEvent(events.MediaPlayerTrackingStarted, playback)
					filename = playback.Filename
					completed = false
				}

				// reset completed status if filename changes
				if filename != "" && filename != playback.Filename {
					m.WSEventManager.SendEvent(events.MediaPlayerTrackingStarted, playback)
					filename = playback.Filename
					completed = false
				}

				if playback.CompletionPercentage > 0.9 && playback.Filename == filename && !completed {
					m.WSEventManager.SendEvent(events.MediaPlayerVideoCompleted, playback)
					m.Logger.Debug().Msg("mediaplayer: Video completed")
					completed = true
					onVideoCompleted()
				}

				//m.WSEventManager.SendEvent(events.MediaPlayerPlaybackStatus, playback)
			}
		}
	}()
}

func (m *Repository) processStatus(player string, status interface{}) (*playbackStatus, bool) {
	switch player {
	case "vlc":
		// Process VLC status
		st := status.(*vlc.Status)
		if st == nil {
			return nil, false
		}

		ret := &playbackStatus{
			CompletionPercentage: st.Position,
			Playing:              st.State == "playing",
			Filename:             st.Information.Category["meta"].Filename,
			Duration:             int(st.Length * 1000),
		}

		return ret, true
	case "mpc-hc":
		// Process MPC-HC status
		st := status.(*mpchc.Variables)
		if st == nil || st.Duration == 0 {
			return nil, false
		}

		ret := &playbackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              st.State == 2,
			Filename:             st.File,
			Duration:             int(st.Duration),
		}

		return ret, true
	case "mpv":
		// Process MPV status
		st := status.(*mpv.Playback)
		spew.Dump(st)
		if st == nil || st.Duration == 0 || st.IsRunning == false {
			return nil, false
		}
		ret := &playbackStatus{
			CompletionPercentage: st.Position / st.Duration,
			Playing:              st.Paused,
			Filename:             st.Filename,
			Duration:             int(st.Duration),
		}

		return ret, true
	default:
		return nil, false
	}
}
