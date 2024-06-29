package mpv

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpvipc"
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
		Logger         *zerolog.Logger
		Playback       *Playback
		SocketName     string
		AppPath        string
		isRunning      bool
		mu             sync.Mutex
		playbackMu     sync.RWMutex
		cancel         context.CancelFunc     // Cancel function for the context
		subscribers    map[string]*Subscriber // Subscribers to the mpv events
		conn           *mpvipc.Connection     // Reference to the mpv connection
		cmd            *exec.Cmd
		prevSocketName string
	}

	Subscriber struct {
		ClosedCh chan struct{}
	}
)

var cmdCtx, cmdCancel = context.WithCancel(context.Background())

func New(logger *zerolog.Logger, socketName string, appPath string) *Mpv {
	if cmdCancel != nil {
		cmdCancel()
	}

	sn := socketName
	if socketName == "" {
		sn = getDefaultSocketName()
	}

	return &Mpv{
		Logger:      logger,
		Playback:    &Playback{},
		mu:          sync.Mutex{},
		playbackMu:  sync.RWMutex{},
		SocketName:  sn,
		AppPath:     appPath,
		subscribers: make(map[string]*Subscriber),
	}
}

// launchPlayer starts the mpv player and plays the file.
// If the player is already running, it just loads the new file.
func (m *Mpv) launchPlayer(idle bool, filePath string, args ...string) error {
	// Cancel previous context
	// This is done so that we only have one connection open at a time
	// DEVNOTE: Commented in favor of replacing the file
	//if m.cancel != nil {
	//	m.Logger.Debug().Msg("mpv: Cancelling previous context")
	//	m.cancel()
	//}

	var err error

	if m.conn == nil || m.conn.IsClosed() {

		cmdCtx, cmdCancel = context.WithCancel(context.Background())

		m.Logger.Debug().Msg("mpv: Starting player")
		if idle {
			args = append(args, "--input-ipc-server="+m.SocketName, "--idle")
			m.cmd, err = m.createCmd("", args...)
		} else {
			args = append(args, "--input-ipc-server="+m.SocketName)
			m.cmd, err = m.createCmd(filePath, args...)
		}
		if err != nil {
			return err
		}
		m.prevSocketName = m.SocketName

		err = m.cmd.Start()
		if err != nil {
			return err
		}
	} else {
		m.Logger.Debug().Msg("mpv: Replacing file")
		// If the connection is still open, just play the file
		_, err = m.conn.Call("loadfile", filePath, "replace")
		if err != nil {
			return err
		}
	}

	// Wait 1 second for the player to start
	time.Sleep(1 * time.Second)

	return nil
}

func (m *Mpv) OpenAndPlay(filePath string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Playback = &Playback{}

	// Launch player or replace file
	err := m.launchPlayer(false, filePath, args...)
	if err != nil {
		return err
	}

	// Create context for the connection
	// When the cancel method is called (by launchPlayer), the previous connection will be closed
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// Establish new connection, only if it doesn't exist
	// We don't continue past this point if the connection is already open, because it means the goroutine is already running
	if m.conn != nil && !m.conn.IsClosed() {
		return nil
	}

	err = m.establishConnection()
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
	go m.listenForEvents(ctx)

	return nil
}

func (m *Mpv) establishConnection() error {
	tries := 1
	for {
		m.conn = mpvipc.NewConnection(m.SocketName)
		err := m.conn.Open()
		if err != nil {
			if tries >= 4 {
				m.Logger.Error().Err(err).Msg("mpv: Failed to establish connection")
				return err
			}
			m.Logger.Error().Err(err).Msgf("mpv: Failed to establish connection (%d/3), retrying...", tries)
			tries++
			time.Sleep(1 * time.Second)
			continue
		}
		m.Logger.Debug().Msg("mpv: Connection established")
		break
	}

	return nil
}

func (m *Mpv) listenForEvents(ctx context.Context) {
	// Close the connection when the goroutine ends
	defer func() {
		m.Logger.Debug().Msg("mpv: Closing socket connection")
		m.resetPlaybackStatus()
		m.isRunning = false
		m.conn.Close()
		m.publishDone()
		m.cancel()
		cmdCancel()
		m.Logger.Debug().Msg("mpv: Instance closed")
	}()

	events, stopListening := m.conn.NewEventListener()
	m.Logger.Debug().Msg("mpv: Listening for events")

	_, err := m.conn.Get("path")
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

func (m *Mpv) CloseAll() {
	if m.conn != nil {
		err := m.conn.Close()
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to close connection")
		}
	}
	m.resetPlaybackStatus()
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

func (s *Subscriber) Done() chan struct{} {
	return s.ClosedCh
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getDefaultSocketName returns the default name of the socket/pipe.
func getDefaultSocketName() string {
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

//func getDefaultSocketName() string {
//	switch runtime.GOOS {
//	case "windows":
//		return fmt.Sprintf("\\\\.\\pipe\\mpv_ipc_%d", time.Now().UnixNano())
//	case "linux":
//		return fmt.Sprintf("/tmp/mpv_socket_%d", time.Now().UnixNano())
//	case "darwin":
//		return fmt.Sprintf("/tmp/mpv_socket_%d", time.Now().UnixNano())
//	default:
//		return fmt.Sprintf("/tmp/mpv_socket_%d", time.Now().UnixNano())
//	}
//}

// createCmd returns a new exec.Cmd instance.
func (m *Mpv) createCmd(filePath string, args ...string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	if filePath != "" {
		args = append(args, filePath)
	}

	switch m.AppPath {
	case "":
		cmd = exec.CommandContext(cmdCtx, "mpv", args...)
	default:
		cmd = exec.CommandContext(cmdCtx, m.AppPath, args...)
	}

	//cmdName := "mpv"
	//if runtime.GOOS == "windows" {
	//	cmdName = "mpv.exe"
	//}

	return cmd, nil
}

func (m *Mpv) resetPlaybackStatus() {
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
