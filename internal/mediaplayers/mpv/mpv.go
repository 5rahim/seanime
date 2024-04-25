package mpv

import (
	"context"
	"errors"
	"github.com/jannson/mpvipc"
	"github.com/rs/zerolog"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

const (
	StartExecCommand    = iota // Start mpv using the "mpv" command
	StartDetectPlayback        // Skip starting mpv, just detect if it's already running
	StartExecPath              // Start mpv using the path provided
	StartExec                  // Start mpv using the path provided, if not provided, use the "mpv" command
)

type (
	Playback struct {
		Filename  string
		Paused    bool
		Position  float64
		Duration  float64
		IsRunning bool
		Filepath  string
	}

	Mpv struct {
		Logger      *zerolog.Logger
		Playback    *Playback
		SocketName  string
		AppPath     string
		isRunning   bool
		mu          sync.Mutex
		playbackMu  sync.RWMutex
		cancel      context.CancelFunc     // Cancel function for the context
		subscribers map[string]*Subscriber // Subscribers to the mpv events
		conn        *mpvipc.Connection     // Reference to the mpv connection
	}

	Subscriber struct {
		ClosedCh chan struct{}
	}
)

func New(logger *zerolog.Logger, socketName string, appPath string) *Mpv {
	if socketName == "" {
		socketName = getSocketName()
	}
	return &Mpv{
		Logger:      logger,
		Playback:    &Playback{},
		mu:          sync.Mutex{},
		playbackMu:  sync.RWMutex{},
		SocketName:  socketName,
		AppPath:     appPath,
		subscribers: make(map[string]*Subscriber),
	}
}

// getSocketName returns the default name of the socket/pipe.
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

// execCmd returns a new exec.Cmd instance based on the provided mode and arguments.
// The mode is determined by user settings.
func (m *Mpv) execCmd(mode int, args ...string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch mode {
	case StartExecPath:
		if m.AppPath == "" {
			return nil, errors.New("mpv path is not set")
		}
		cmd = exec.Command(m.AppPath, args...)

	case StartExecCommand:
		cmd = exec.Command("mpv", args...)

	case StartExec:
		if m.AppPath > "" {
			cmd = exec.Command(m.AppPath, args...)
		} else {
			cmd = exec.Command("mpv", args...)
		}

	default:
		panic("invalid execution mode")
	}
	return cmd, nil
}

// launchPlayer starts the mpv player and plays the file.
// If the player is already running, it just loads the new file.
func (m *Mpv) launchPlayer(start int, filePath string) error {
	// Cancel previous context
	// This is done so that we only have one connection open at a time
	//if m.cancel != nil {
	//	m.Logger.Debug().Msg("mpv: Cancelling previous context")
	//	m.cancel()
	//}

	switch start {
	case StartExecPath, StartExecCommand, StartExec:
		// If no connection exists, start the player and play the file
		if m.conn == nil || m.conn.IsClosed() {
			m.Logger.Debug().Msg("mpv: Starting player")
			cmd, err := m.execCmd(start, "--input-ipc-server="+m.SocketName, filePath)
			if err != nil {
				return err
			}

			err = cmd.Start()
			if err != nil {
				return err
			}
		} else {
			m.Logger.Debug().Msg("mpv: Replacing file")
			// If the connection is still open, just play the file
			_, err := m.conn.Call("loadfile", filePath, "replace")
			if err != nil {
				return err
			}
		}

		// Wait 1 second for the player to start
		time.Sleep(1 * time.Second)

	case StartDetectPlayback:
		// Do nothing
	}

	return nil
}

func (m *Mpv) OpenAndPlay(filePath string, start int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Playback = &Playback{}

	// Launch player
	err := m.launchPlayer(start, filePath)
	if err != nil {
		return err
	}

	// Create context for the connection
	// When the cancel method is called (by launchPlayer), the previous connection will be closed
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// Establish new connection, only if it doesn't exist
	if m.conn != nil {
		return nil
	}

	m.conn = mpvipc.NewConnection(m.SocketName)
	err = m.conn.Open()
	if err != nil {
		return err
	}

	m.isRunning = true
	m.Logger.Debug().Msg("mpv: Connection established")

	// Reset subscriber's done channel in case it was closed
	for _, sub := range m.subscribers {
		sub.ClosedCh = make(chan struct{})
	}

	// Listen for events in a goroutine
	go func() {
		// Close the connection when the goroutine ends
		defer func() {
			m.Logger.Debug().Msg("mpv: Closing socket connection")
			m.ResetPlaybackStatus()
			m.isRunning = false
			m.conn.Close()
			m.publishDone()
			m.Logger.Debug().Msg("mpv: Instance closed")
		}()

		events, stopListening := m.conn.NewEventListener()
		m.Logger.Debug().Msg("mpv: Listening for events")

		_, err = m.conn.Get("path")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to get path")
			return
		}

		_, err = m.conn.Call("observe_property", 42, "time-pos")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to observe time-pos")
			return
		}
		_, err = m.conn.Call("observe_property", 43, "pause")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to observe pause")
			return
		}
		_, err = m.conn.Call("observe_property", 44, "duration")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to observe duration")
			return
		}
		_, err = m.conn.Call("observe_property", 45, "filename")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to observe filename")
			return
		}
		_, err = m.conn.Call("observe_property", 46, "path")
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to observe path")
			return
		}

		// Listen for close event
		go func() {
			m.conn.WaitUntilClosed()
			m.Logger.Debug().Msg("mpv: Connection has been closed")
			stopListening <- struct{}{}
		}()

		go func() {
			// When the context is cancelled, close the connection
			<-ctx.Done()
			m.Logger.Debug().Msg("mpv: Context cancelled")
			err := m.conn.Close()
			if err != nil {
				m.Logger.Error().Err(err).Msg("mpv: Failed to close connection")
			}
			stopListening <- struct{}{}
			return
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
				case 46:
					m.Playback.Filepath = event.Data.(string)
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
	m.Logger.Debug().Msg("mpv: Resetting playback status")
	m.Playback.Filename = ""
	m.Playback.Filepath = ""
	m.Playback.Paused = false
	m.Playback.Position = 0
	m.Playback.Duration = 0
	m.playbackMu.Unlock()
	return
}

func (m *Mpv) CloseAll() {
	err := m.conn.Close()
	if err != nil {
		m.Logger.Error().Err(err).Msg("mpv: Failed to close connection")
	}
	m.ResetPlaybackStatus()
	m.isRunning = false
}

func (m *Mpv) Subscribe(id string) *Subscriber {
	sub := &Subscriber{
		ClosedCh: make(chan struct{}),
	}
	m.subscribers[id] = sub
	return sub
}

func (m *Mpv) Unsubscribe(id string) {
	delete(m.subscribers, id)
}

func (m *Mpv) publishDone() {
	defer func() {
		if r := recover(); r != nil {
			m.Logger.Warn().Msgf("mpv: Connection already closed")
		}
	}()
	for _, sub := range m.subscribers {
		close(sub.ClosedCh)
	}
}

func (s *Subscriber) Done() chan struct{} {
	return s.ClosedCh
}
