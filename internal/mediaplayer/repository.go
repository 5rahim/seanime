package mediaplayer

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/mpchc"
	"github.com/seanime-app/seanime-server/internal/vlc"
	"time"
)

type (
	Repository struct {
		Logger         *zerolog.Logger
		Default        string
		VLC            *vlc.VLC
		MpcHc          *mpchc.MpcHc
		WSEventManager events.IWSEventManager
	}

	playbackStatus struct {
		completionPercentage float64
		playing              bool
		filename             string
		duration             int // in ms
	}
)

func (m *Repository) Play(path string) error {
	switch m.Default {
	case "vlc":
		m.VLC.Start()
		err := m.VLC.AddAndPlay(path)
		if err != nil {
			return err
		}
		return nil
	case "mpc-hc":
		m.MpcHc.Start()
		_, err := m.MpcHc.OpenAndPlay(path)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("no default media player set")
	}
}

// StartTracking will start a goroutine that will monitor the status of the media player
func (m *Repository) StartTracking() {
	// Create a channel to signal the goroutine to exit
	done := make(chan struct{})
	var filename string
	var completed bool

	go func() {
		for {
			select {
			case <-done:
				m.Logger.Printf("Exiting media player status monitoring goroutine")
				return
			case <-time.After(3 * time.Second):
				var status interface{}
				var err error

				// Check the status based on the default player
				switch m.Default {
				case "vlc":
					status, err = m.VLC.GetStatus()
				case "mpc-hc":
					status, err = m.MpcHc.GetVariables()
				}

				if err != nil {
					m.Logger.Debug().Msg("mediaplayer: Tracking stopped")
					close(done) // Signal to exit the goroutine
					return
				}

				// Process the status
				playback, ok := m.processStatus(m.Default, status)
				if !ok {
					m.Logger.Error().Msg("mediaplayer: Failed to process status")
					close(done) // Signal to exit the goroutine
					return
				}

				if filename == "" {
					filename = playback.filename
					completed = false
				}
				// reset completed status if filename changes
				if filename != "" && filename != playback.filename {
					m.WSEventManager.SendEvent("mediaplayer-fresh-start", playback)
					filename = playback.filename
					completed = false
				}

				if playback.completionPercentage > 0.9 && playback.filename == filename && !completed {
					m.Logger.Debug().Msg("mediaplayer: Video completed")
					m.WSEventManager.SendEvent("mediaplayer-video-completed", playback)
					completed = true
				}

				m.WSEventManager.SendEvent("mediaplayer-status", playback)
			}
		}
	}()
}

func (m *Repository) processStatus(player string, status interface{}) (*playbackStatus, bool) {
	switch player {
	case "vlc":
		// Process VLC status
		st := status.(*vlc.Status)

		ret := &playbackStatus{
			completionPercentage: st.Position,
			playing:              st.State == "playing",
			filename:             st.Information.Category["meta"].Filename,
			duration:             int(st.Length * 1000),
		}

		return ret, true
	case "mpc-hc":
		// Process MPC-HC status
		st := status.(*mpchc.Variables)

		ret := &playbackStatus{
			completionPercentage: st.Position / st.Duration,
			playing:              st.State == 2,
			filename:             st.File,
			duration:             int(st.Duration),
		}

		return ret, true
	default:
		return nil, false
	}
}
