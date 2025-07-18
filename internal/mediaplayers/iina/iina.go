package iina

import (
	"context"
	"errors"
	"os/exec"
	"seanime/internal/mediaplayers/mpvipc"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
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

	Iina struct {
		Logger         *zerolog.Logger
		Playback       *Playback
		SocketName     string
		AppPath        string
		Args           string
		mu             sync.Mutex
		playbackMu     sync.RWMutex
		cancel         context.CancelFunc               // Cancel function for the context
		subscribers    *result.Map[string, *Subscriber] // Subscribers to the iina events
		conn           *mpvipc.Connection               // Reference to the mpv connection (iina uses mpv IPC)
		cmd            *exec.Cmd
		prevSocketName string
		exitedCh       chan struct{}
	}

	// Subscriber is a subscriber to the iina events.
	// Make sure the subscriber listens to both channels, otherwise it will deadlock.
	Subscriber struct {
		eventCh  chan *mpvipc.Event
		closedCh chan struct{}
	}
)

var cmdCtx, cmdCancel = context.WithCancel(context.Background())

func New(logger *zerolog.Logger, socketName string, appPath string, optionalArgs ...string) *Iina {
	if cmdCancel != nil {
		cmdCancel()
	}

	sn := socketName
	if socketName == "" {
		sn = getDefaultSocketName()
	}

	additionalArgs := ""
	if len(optionalArgs) > 0 {
		additionalArgs = optionalArgs[0]
	}

	return &Iina{
		Logger:      logger,
		Playback:    &Playback{},
		mu:          sync.Mutex{},
		playbackMu:  sync.RWMutex{},
		SocketName:  sn,
		AppPath:     appPath,
		Args:        additionalArgs,
		subscribers: result.NewResultMap[string, *Subscriber](),
		exitedCh:    make(chan struct{}),
	}
}

func (i *Iina) GetExecutablePath() string {
	if i.AppPath != "" {
		return i.AppPath
	}
	return "iina-cli"
}

// launchPlayer starts the iina player and plays the file.
// If the player is already running, it just loads the new file.
func (i *Iina) launchPlayer(idle bool, filePath string, args ...string) error {
	var err error

	i.Logger.Trace().Msgf("iina: Launching player with args: %+v", args)

	// Cancel previous goroutine context
	if i.cancel != nil {
		i.Logger.Trace().Msg("iina: Cancelling previous context")
		i.cancel()
	}
	// Cancel previous command context
	if cmdCancel != nil {
		i.Logger.Trace().Msg("iina: Cancelling previous command context")
		cmdCancel()
	}
	cmdCtx, cmdCancel = context.WithCancel(context.Background())

	i.Logger.Debug().Msg("iina: Starting player")

	iinaArgs := []string{
		"--mpv-input-ipc-server=" + i.SocketName,
		"--no-stdin",
	}

	if idle {
		iinaArgs = append(iinaArgs, "--mpv-idle")
		iinaArgs = append(iinaArgs, args...)
		i.cmd, err = i.createCmd("", iinaArgs...)
	} else {
		iinaArgs = append(iinaArgs, args...)
		i.cmd, err = i.createCmd(filePath, iinaArgs...)
	}

	if err != nil {
		return err
	}
	i.prevSocketName = i.SocketName

	err = i.cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		err := i.cmd.Wait()
		if err != nil {
			i.Logger.Warn().Err(err).Msg("iina: Player has exited")
		}
	}()

	time.Sleep(2 * time.Second)

	i.Logger.Debug().Msg("iina: Player started")

	return nil
}

func (i *Iina) replaceFile(filePath string) error {
	i.Logger.Debug().Msg("iina: Replacing file")

	if i.conn != nil && !i.conn.IsClosed() {
		_, err := i.conn.Call("loadfile", filePath, "replace")
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Iina) Exited() chan struct{} {
	return i.exitedCh
}

func (i *Iina) OpenAndPlay(filePath string, args ...string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.Playback = &Playback{}

	// If the player is already running, just load the new file
	var err error
	if i.conn != nil && !i.conn.IsClosed() {
		// Launch player or replace file
		err = i.replaceFile(filePath)
	} else {
		// Launch player
		err = i.launchPlayer(false, filePath, args...)
	}
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, i.cancel = context.WithCancel(context.Background())

	// Establish new connection, only if it doesn't exist
	if i.conn != nil && !i.conn.IsClosed() {
		return nil
	}

	err = i.establishConnection()
	if err != nil {
		return err
	}

	i.Playback.IsRunning = false

	// Listen for events in a goroutine
	go i.listenForEvents(ctx)

	return nil
}

func (i *Iina) Pause() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.conn == nil || i.conn.IsClosed() {
		return errors.New("iina is not running")
	}

	_, err := i.conn.Call("set_property", "pause", true)
	if err != nil {
		return err
	}

	return nil
}

func (i *Iina) Resume() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.conn == nil || i.conn.IsClosed() {
		return errors.New("iina is not running")
	}

	_, err := i.conn.Call("set_property", "pause", false)
	if err != nil {
		return err
	}

	return nil
}

// SeekTo seeks to the given position in the file.
func (i *Iina) SeekTo(position float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.conn == nil || i.conn.IsClosed() {
		return errors.New("iina is not running")
	}

	_, err := i.conn.Call("set_property", "time-pos", position)
	if err != nil {
		return err
	}

	return nil
}

// Seek seeks to the given position in the file.
func (i *Iina) Seek(position float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.conn == nil || i.conn.IsClosed() {
		return errors.New("iina is not running")
	}

	_, err := i.conn.Call("set_property", "time-pos", position)
	if err != nil {
		return err
	}

	return nil
}

func (i *Iina) GetOpenConnection() (*mpvipc.Connection, error) {
	if i.conn == nil || i.conn.IsClosed() {
		return nil, errors.New("iina is not running")
	}
	return i.conn, nil
}

func (i *Iina) establishConnection() error {
	tries := 1
	for {
		i.conn = mpvipc.NewConnection(i.SocketName)
		err := i.conn.Open()
		if err != nil {
			if tries >= 3 {
				i.Logger.Error().Err(err).Msg("iina: Failed to establish connection")
				return err
			}
			i.Logger.Error().Err(err).Msgf("iina: Failed to establish connection (%d/8), retrying...", tries)
			tries++
			time.Sleep(1500 * time.Millisecond)
			continue
		}
		i.Logger.Debug().Msg("iina: Connection established")
		break
	}

	return nil
}

func (i *Iina) listenForEvents(ctx context.Context) {
	// Close the connection when the goroutine ends
	defer func() {
		i.Logger.Debug().Msg("iina: Closing socket connection")
		i.conn.Close()
		i.terminate()
		i.Logger.Debug().Msg("iina: Instance closed")
	}()

	events, stopListening := i.conn.NewEventListener()
	i.Logger.Debug().Msg("iina: Listening for events")

	_, err := i.conn.Get("path")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to get path")
		return
	}

	_, err = i.conn.Call("observe_property", 42, "time-pos")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to observe time-pos")
		return
	}
	_, err = i.conn.Call("observe_property", 43, "pause")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to observe pause")
		return
	}
	_, err = i.conn.Call("observe_property", 44, "duration")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to observe duration")
		return
	}
	_, err = i.conn.Call("observe_property", 45, "filename")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to observe filename")
		return
	}
	_, err = i.conn.Call("observe_property", 46, "path")
	if err != nil {
		i.Logger.Error().Err(err).Msg("iina: Failed to observe path")
		return
	}

	// Listen for close event
	go func() {
		i.conn.WaitUntilClosed()
		i.Logger.Debug().Msg("iina: Connection has been closed")
		stopListening <- struct{}{}
	}()

	go func() {
		// When the context is cancelled, close the connection
		<-ctx.Done()
		i.Logger.Debug().Msg("iina: Context cancelled")
		i.Playback.IsRunning = false
		err := i.conn.Close()
		if err != nil {
			i.Logger.Error().Err(err).Msg("iina: Failed to close connection")
		}
		stopListening <- struct{}{}
		return
	}()

	// Listen for events
	for event := range events {
		if event.Data != nil {
			i.Playback.IsRunning = true
			switch event.ID {
			case 43:
				i.Playback.Paused = event.Data.(bool)
			case 42:
				i.Playback.Position = event.Data.(float64)
			case 44:
				i.Playback.Duration = event.Data.(float64)
			case 45:
				i.Playback.Filename = event.Data.(string)
			case 46:
				i.Playback.Filepath = event.Data.(string)
			}
			i.subscribers.Range(func(key string, sub *Subscriber) bool {
				go func() {
					sub.eventCh <- event
				}()
				return true
			})
		}
	}
}

func (i *Iina) GetPlaybackStatus() (*Playback, error) {
	i.playbackMu.RLock()
	defer i.playbackMu.RUnlock()
	if !i.Playback.IsRunning {
		return nil, errors.New("iina is not running")
	}
	if i.Playback == nil {
		return nil, errors.New("no playback status")
	}
	if i.Playback.Filename == "" {
		return nil, errors.New("no media found")
	}
	if i.Playback.Duration == 0 {
		return nil, errors.New("no duration found")
	}
	return i.Playback, nil
}

func (i *Iina) CloseAll() {
	i.Logger.Debug().Msg("iina: Received close request")
	if i.conn != nil && !i.conn.IsClosed() {
		// Send quit command to IINA before closing connection
		i.Logger.Debug().Msg("iina: Sending quit command")
		_, err := i.conn.Call("quit")
		if err != nil {
			i.Logger.Warn().Err(err).Msg("iina: Failed to send quit command")
		}
		time.Sleep(500 * time.Millisecond)

		err = i.conn.Close()
		if err != nil {
			i.Logger.Error().Err(err).Msg("iina: Failed to close connection")
		}
	}
	i.terminate()
}

func (i *Iina) terminate() {
	defer func() {
		if r := recover(); r != nil {
			i.Logger.Warn().Msgf("iina: Termination panic")
		}
	}()
	i.Logger.Trace().Msg("iina: Terminating")
	i.resetPlaybackStatus()
	i.publishDone()
	if i.cancel != nil {
		i.cancel()
	}
	if cmdCancel != nil {
		cmdCancel()
	}
	i.Logger.Trace().Msg("iina: Terminated")
}

func (i *Iina) Subscribe(id string) *Subscriber {
	sub := &Subscriber{
		eventCh:  make(chan *mpvipc.Event, 100),
		closedCh: make(chan struct{}),
	}
	i.subscribers.Set(id, sub)
	return sub
}

func (i *Iina) Unsubscribe(id string) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	sub, ok := i.subscribers.Get(id)
	if !ok {
		return
	}
	close(sub.eventCh)
	close(sub.closedCh)
	i.subscribers.Delete(id)
}

func (s *Subscriber) Events() <-chan *mpvipc.Event {
	return s.eventCh
}

func (s *Subscriber) Closed() <-chan struct{} {
	return s.closedCh
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// parseArgs parses a command line string into individual arguments, respecting quotes
func parseArgs(s string) ([]string, error) {
	args := make([]string, 0)
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
				// Add the current string even if it's empty (for empty quoted strings)
				args = append(args, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		case char == ' ' || char == '\t':
			if inQuotes {
				current.WriteRune(char)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case char == '\\' && i+1 < len(runes):
			// Handle escaped characters
			if inQuotes && (runes[i+1] == '"' || runes[i+1] == '\'') {
				i++
				current.WriteRune(runes[i])
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if inQuotes {
		return nil, errors.New("unclosed quote in arguments")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func getDefaultSocketName() string {
	return "/tmp/iina_socket"
}

// createCmd returns a new exec.Cmd instance for iina-cli.
func (i *Iina) createCmd(filePath string, args ...string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	// Add user-defined arguments
	if i.Args != "" {
		userArgs, err := parseArgs(i.Args)
		if err != nil {
			i.Logger.Warn().Err(err).Msg("iina: Failed to parse user arguments, using simple split")
			userArgs = strings.Fields(i.Args)
		}
		args = append(args, userArgs...)
	}

	if filePath != "" {
		args = append(args, filePath)
	}

	binaryPath := i.GetExecutablePath()

	cmd = util.NewCmdCtx(cmdCtx, binaryPath, args...)

	i.Logger.Trace().Msgf("iina: Command: %s", strings.Join(cmd.Args, " "))

	return cmd, nil
}

func (i *Iina) resetPlaybackStatus() {
	i.playbackMu.Lock()
	i.Logger.Trace().Msg("iina: Resetting playback status")
	i.Playback.Filename = ""
	i.Playback.Filepath = ""
	i.Playback.Paused = false
	i.Playback.Position = 0
	i.Playback.Duration = 0
	i.Playback.IsRunning = false
	i.playbackMu.Unlock()
	return
}

func (i *Iina) publishDone() {
	defer func() {
		if r := recover(); r != nil {
			i.Logger.Warn().Msgf("iina: Connection already closed")
		}
	}()
	i.subscribers.Range(func(key string, sub *Subscriber) bool {
		go func() {
			sub.closedCh <- struct{}{}
		}()
		return true
	})
}
