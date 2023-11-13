package mediaplayer

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/mpchc"
	"github.com/seanime-app/seanime-server/internal/vlc"
	"time"
)

type (
	Repository struct {
		Logger  *zerolog.Logger
		Default string
		VLC     *vlc.VLC
		MpcHc   *mpchc.MpcHc
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
					m.Logger.Printf("Error in monitoring media player status: %v", err)
					close(done) // Signal to exit the goroutine
					return
				}

				// Process the status
				m.Logger.Printf("Media Player Status: %v\n", status)
			}
		}
	}()
}
