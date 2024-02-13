package mpv

import (
	"errors"
	"github.com/jannson/mpvipc"
	"github.com/rs/zerolog"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

var (
	ErrConnClosed = errors.New("connection closed")
)

type (
	Playback struct {
		Filename  string
		Paused    bool
		Position  float64
		Duration  float64
		IsRunning bool
	}

	Mpv struct {
		Logger     *zerolog.Logger
		ExitCh     chan error
		CloseCh    chan struct{}
		Playback   *Playback
		SocketName string
		isRunning  bool
		mu         sync.Mutex
		playbackMu sync.RWMutex
	}
)

func New(logger *zerolog.Logger, socketName string) *Mpv {
	if socketName == "" {
		socketName = getSocketName()
	}
	return &Mpv{
		Logger:     logger,
		ExitCh:     make(chan error),
		CloseCh:    make(chan struct{}),
		Playback:   &Playback{},
		mu:         sync.Mutex{},
		playbackMu: sync.RWMutex{},
		SocketName: socketName,
	}
}

func getSocketName() string {
	switch runtime.GOOS {
	case "windows":
		return "\\\\.\\pipe\\mpv_ipc"
	case "linux":
		return "/tmp/mpv_socket"
	case "darwin":
		return "/tmp/mpv_socket"
	default:
		return "/tmp/mpv_socket"
	}
}

func (m *Mpv) OpenAndPlay(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return errors.New("an instance of mpv is already running")
	}

	sn := m.SocketName

	// Launch player
	cmd := exec.Command("mpv", "--input-ipc-server="+sn, filePath)
	err := cmd.Start()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// Establish connection
	conn := mpvipc.NewConnection(sn)
	err = conn.Open()
	if err != nil {
		return err
	}

	m.isRunning = true
	m.Logger.Debug().Msg("mpv: connection established")

	// Listen for events in a goroutine
	go func() {
		// Close the connection when the goroutine ends
		defer func() {
			m.Logger.Debug().Msg("mpv: closing socket connection")
			m.ResetPlaybackStatus()
			m.isRunning = false
			conn.Close()
			m.ExitCh <- ErrConnClosed
			m.Logger.Debug().Msg("mpv: connection closed")
		}()

		events, stopListening := conn.NewEventListener()

		_, err = conn.Get("path")
		if err != nil {
			m.ExitCh <- err
			return
		}

		//err = conn.Set("pause", true)
		//if err != nil {
		//	m.ExitCh <- err
		//	return
		//}

		_, err = conn.Call("observe_property", 42, "time-pos")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 43, "pause")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 44, "duration")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 45, "filename")
		if err != nil {
			m.ExitCh <- err
			return
		}

		// Listen for close event
		go func() {
			conn.WaitUntilClosed()
			stopListening <- struct{}{}
		}()

		// Close the connection when external signal is received
		go func() {
			select {
			case <-m.CloseCh:
				stopListening <- struct{}{}
				break
			}
		}()

		// Listen for events
		for event := range events {
			m.Playback.IsRunning = true
			if event.Data != nil {
				//m.Logger.Trace().Msgf("received event: %s, %v, %+v", event.Name, event.ID, event.Data)
				switch event.ID {
				case 43:
					m.Playback.Paused = event.Data.(bool)
				case 42:
					m.Playback.Position = event.Data.(float64)
				case 44:
					m.Playback.Duration = event.Data.(float64)
				case 45:
					m.Playback.Filename = event.Data.(string)
				}
			}
		}
	}()

	return nil
}

func (m *Mpv) DetectPlayback() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return errors.New("playback detection already running")
	}

	sn := getSocketName()

	// Establish connection
	conn := mpvipc.NewConnection(sn)
	err := conn.Open()
	if err != nil {
		return err
	}

	m.isRunning = true
	m.Logger.Debug().Msg("mpv: connection established")

	// Listen for events in a goroutine
	go func() {
		// Close the connection when the goroutine ends
		defer func() {
			m.Logger.Debug().Msg("mpv: closing socket connection")
			m.ResetPlaybackStatus()
			m.isRunning = false
			conn.Close()
			m.ExitCh <- ErrConnClosed
			m.Logger.Debug().Msg("mpv: connection closed")
		}()

		events, stopListening := conn.NewEventListener()

		_, err = conn.Get("path")
		if err != nil {
			m.ExitCh <- err
			return
		}

		//err = conn.Set("pause", true)
		//if err != nil {
		//	m.ExitCh <- err
		//	return
		//}

		_, err = conn.Call("observe_property", 42, "time-pos")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 43, "pause")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 44, "duration")
		if err != nil {
			m.ExitCh <- err
			return
		}
		_, err = conn.Call("observe_property", 45, "filename")
		if err != nil {
			m.ExitCh <- err
			return
		}

		// Listen for close event
		go func() {
			conn.WaitUntilClosed()
			stopListening <- struct{}{}
		}()

		// Close the connection when external signal is received
		go func() {
			select {
			case <-m.CloseCh:
				stopListening <- struct{}{}
				break
			}
		}()

		// Listen for events
		for event := range events {
			m.Playback.IsRunning = true
			if event.Data != nil {
				switch event.ID {
				case 43:
					m.Playback.Paused = event.Data.(bool)
				case 42:
					m.Playback.Position = event.Data.(float64)
				case 44:
					m.Playback.Duration = event.Data.(float64)
				case 45:
					m.Playback.Filename = event.Data.(string)
				}
			}
		}
	}()

	return nil
}

func (m *Mpv) GetPlaybackStatus() (*Playback, error) {
	m.playbackMu.RLock()
	defer m.playbackMu.RUnlock()
	if m.Playback.IsRunning == false {
		return nil, errors.New("mpv is not running")
	}
	if m.Playback == nil {
		return nil, errors.New("no playback status")
	}
	if m.Playback.Filename == "" {
		return nil, errors.New("no media found")
	}
	return m.Playback, nil
}

func (m *Mpv) ResetPlaybackStatus() {
	m.playbackMu.Lock()
	m.Logger.Debug().Msg("mpv: resetting playback status")
	m.Playback.Filename = ""
	m.Playback.Paused = false
	m.Playback.Position = 0
	m.Playback.Duration = 0
	m.playbackMu.Unlock()
	return
}

func (m *Mpv) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseCh <- struct{}{}
}
