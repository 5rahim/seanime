package mpv

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"seanime/internal/mediaplayers/mpvipc"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"sync/atomic"
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

	Mpv struct {
		Logger         *zerolog.Logger
		Playback       *Playback
		SocketName     string
		AppPath        string
		Args           string
		mu             sync.Mutex
		playbackMu     sync.RWMutex
		cancel         context.CancelFunc               // Cancel function for the context
		subscribers    *result.Map[string, *Subscriber] // Subscribers to the mpv events
		conn           *mpvipc.Connection               // Reference to the mpv connection
		cmd            *exec.Cmd
		prevSocketName string
		exitedCh       chan struct{}
		playbackSwitch bool
		isFileLoaded   bool
		freshPosition  bool
		freshDuration  bool
		autoSocket     bool
		launchErrCh    chan error
		launchLogPath  string
	}

	// Subscriber is a subscriber to the mpv events.
	// Make sure the subscriber listens to both channels, otherwise it will deadlock.
	Subscriber struct {
		eventCh  chan *mpvipc.Event
		closedCh chan struct{}
	}
)

var cmdCtx, cmdCancel = context.WithCancel(context.Background())
var socketCounter uint64

func New(logger *zerolog.Logger, socketName string, appPath string, optionalArgs ...string) *Mpv {
	if cmdCancel != nil {
		cmdCancel()
	}

	autoSocket := shouldAutoSocket(socketName)
	sn := socketName
	if autoSocket {
		sn = getDefaultSocketName()
	}

	additionalArgs := ""
	if len(optionalArgs) > 0 {
		additionalArgs = optionalArgs[0]
	}

	return &Mpv{
		Logger:      logger,
		Playback:    &Playback{},
		mu:          sync.Mutex{},
		playbackMu:  sync.RWMutex{},
		SocketName:  sn,
		AppPath:     appPath,
		Args:        additionalArgs,
		subscribers: result.NewMap[string, *Subscriber](),
		exitedCh:    make(chan struct{}),
		autoSocket:  autoSocket,
	}
}

func (m *Mpv) GetExecutablePath() string {
	if m.AppPath != "" {
		return m.AppPath
	}
	return "mpv"
}

// launchPlayer starts the mpv player and plays the file.
// If the player is already running, it just loads the new file.
func (m *Mpv) launchPlayer(idle bool, filePath string, args ...string) error {
	var err error

	m.Logger.Trace().Msgf("mpv: Launching player with args: %+v", args)

	// Cancel previous goroutine context
	if m.cancel != nil {
		m.Logger.Trace().Msg("mpv: Cancelling previous context")
		m.cancel()
	}
	// Cancel previous command context
	if cmdCancel != nil {
		m.Logger.Trace().Msg("mpv: Cancelling previous command context")
		cmdCancel()
	}
	cmdCtx, cmdCancel = context.WithCancel(context.Background())
	if m.autoSocket {
		m.SocketName = getDefaultSocketName()
	}

	launchLogPath, err := createLaunchLogPath()
	if err != nil {
		return err
	}
	m.launchLogPath = launchLogPath
	launchErrCh := make(chan error, 1)
	m.launchErrCh = launchErrCh

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
	m.launchLogPath = ""

	// Create a pipe for stdout
	stdoutPipe, err := m.cmd.StdoutPipe()
	if err != nil {
		m.Logger.Error().Err(err).Msg("mpv: Failed to create stdout pipe")
		return err
	}

	stderrPipe, err := m.cmd.StderrPipe()
	if err != nil {
		m.Logger.Error().Err(err).Msg("mpv: Failed to create stderr pipe")
		return err
	}

	err = m.cmd.Start()
	if err != nil {
		return err
	}

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	receivedLogCh := make(chan struct{})
	var receivedLogOnce sync.Once
	notifyLogReceived := func() {
		receivedLogOnce.Do(func() {
			close(receivedLogCh)
		})
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			// Skip AV messages
			if bytes.Contains(scanner.Bytes(), []byte("AV:")) {
				continue
			}
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				notifyLogReceived()
				m.Logger.Trace().Msg("mpv cmd: " + line) // Print to logger
			}
		}
		if err := scanner.Err(); err != nil {
			if strings.Contains(err.Error(), "file already closed") {
				m.Logger.Debug().Msg("mpv: File closed")
			} else {
				m.Logger.Error().Err(err).Msg("mpv: Error reading from stdout")
			}
		}
		// unblock startup when stdout closes before logging.
		notifyLogReceived()
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				notifyLogReceived()
				m.Logger.Trace().Msg("mpv stderr: " + line)
			}
		}
		if err := scanner.Err(); err != nil {
			if strings.Contains(err.Error(), "file already closed") {
				m.Logger.Debug().Msg("mpv: Stderr pipe closed")
			} else {
				m.Logger.Error().Err(err).Msg("mpv: Error reading from stderr")
			}
		}
		// unblock startup when stderr closes before logging
		notifyLogReceived()
	}()

	go func() {
		err := m.cmd.Wait()
		if err != nil {
			launchErr := formatLaunchExitError(err, launchLogPath)
			m.Logger.Warn().Err(launchErr).Msg("mpv: Player has exited")
			launchErrCh <- launchErr
		}
		_ = os.Remove(launchLogPath)
	}()

	// Wait until either an initial log is received or 2 seconds have passed
	select {
	case <-receivedLogCh:
	case <-timer.C:
		m.Logger.Trace().Msg("mpv: Proceeding without initial log (timeout reached)")
	}

	time.Sleep(1 * time.Second)

	m.Logger.Debug().Msg("mpv: Player started")

	return nil
}

func (m *Mpv) replaceFile(filePath string) error {
	m.Logger.Debug().Msg("mpv: Replacing file")

	if m.conn != nil && !m.conn.IsClosed() {
		_, err := m.conn.Call("loadfile", filePath, "replace")
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Mpv) Exited() chan struct{} {
	return m.exitedCh
}

func (m *Mpv) OpenAndPlay(filePath string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.playbackMu.Lock()
	m.Playback = &Playback{}
	m.playbackSwitch = false
	m.isFileLoaded = false
	m.freshPosition = false
	m.freshDuration = false
	m.playbackMu.Unlock()

	// If the player is already running, just load the new file
	var err error
	if m.conn != nil && !m.conn.IsClosed() {
		// Launch player or replace file
		err = m.replaceFile(filePath)
	} else {
		// Launch player
		err = m.launchPlayer(false, filePath, args...)
	}
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

	// // Reset subscriber's done channel in case it was closed
	// m.subscribers.Range(func(key string, sub *Subscriber) bool {
	// 	sub.eventCh = make(chan *mpvipc.Event)
	// 	return true
	// })

	m.Playback.IsRunning = false

	// Listen for events in a goroutine
	go m.listenForEvents(ctx)

	return nil
}

func (m *Mpv) Append(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil || m.conn.IsClosed() {
		return errors.New("mpv is not running")
	}

	// Clear playlist if any
	_, _ = m.conn.Call("playlist-clear")

	_, err := m.conn.Call("loadfile", path, "append")
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpv) Pause() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil || m.conn.IsClosed() {
		return errors.New("mpv is not running")
	}

	_, err := m.conn.Call("set_property", "pause", true)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpv) Resume() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil || m.conn.IsClosed() {
		return errors.New("mpv is not running")
	}

	_, err := m.conn.Call("set_property", "pause", false)
	if err != nil {
		return err
	}

	return nil
}

// waitForFileLoad waits until the MPV track is ready to seek.
func (m *Mpv) waitForFileLoad(timeoutDuration time.Duration) error {
	timeout := time.Now().Add(timeoutDuration)
	for {
		if time.Now().After(timeout) {
			return errors.New("timed out waiting for file to load")
		}

		m.mu.Lock()
		connClosed := m.conn == nil || m.conn.IsClosed()
		m.mu.Unlock()

		if connClosed {
			return errors.New("mpv is not running")
		}

		if m.isFileReady() {
			return nil
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (m *Mpv) isFileReady() bool {
	m.playbackMu.RLock()
	defer m.playbackMu.RUnlock()
	return m.isFileLoaded || (m.freshDuration && m.Playback.Duration > 0)
}

// SeekToSlow seeks to the given position in the file by first pausing the player and unpausing it after seeking.
func (m *Mpv) SeekToSlow(position float64) error {
	// Wait for file to load before seeking
	err := m.waitForFileLoad(30 * time.Second)
	if err != nil {
		m.Logger.Warn().Err(err).Msg("file didn't load in time, attempting seek anyway")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil || m.conn.IsClosed() {
		return errors.New("mpv is not running")
	}

	// pause the player
	_, err = m.conn.Call("set_property", "pause", true)
	if err != nil {
		return err
	}

	// unpause the player
	defer func() {
		_, _ = m.conn.Call("set_property", "pause", false)
	}()

	time.Sleep(100 * time.Millisecond)

	_, err = m.conn.Call("set_property", "time-pos", position)
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}

// SeekTo seeks to the given position in the file.
func (m *Mpv) SeekTo(position float64) error {
	// Wait for file to load before seeking
	err := m.waitForFileLoad(30 * time.Second)
	if err != nil {
		m.Logger.Warn().Err(err).Msg("file didn't load in time, attempting seek anyway")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn == nil || m.conn.IsClosed() {
		return errors.New("mpv is not running")
	}

	_, err = m.conn.Call("set_property", "time-pos", position)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpv) GetOpenConnection() (*mpvipc.Connection, error) {
	if m.conn == nil || m.conn.IsClosed() {
		return nil, errors.New("mpv is not running")
	}
	return m.conn, nil
}

func (m *Mpv) establishConnection() error {
	tries := 1
	for {
		if launchErr := m.takeLaunchError(); launchErr != nil {
			m.Logger.Error().Err(launchErr).Msg("mpv: Failed to establish connection")
			return launchErr
		}
		m.conn = mpvipc.NewConnection(m.SocketName)
		err := m.conn.Open()
		if err != nil {
			if launchErr := m.takeLaunchError(); launchErr != nil {
				m.Logger.Error().Err(launchErr).Msg("mpv: Failed to establish connection")
				return launchErr
			}
			if tries >= 5 {
				m.Logger.Error().Err(err).Msg("mpv: Failed to establish connection")
				return err
			}
			m.Logger.Error().Err(err).Msgf("mpv: Failed to establish connection (%d/4), retrying...", tries)
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
		m.conn.Close()
		m.terminate()
		m.Logger.Debug().Msg("mpv: Instance closed")
	}()

	events, stopListening := m.conn.NewEventListener()
	m.Logger.Debug().Msg("mpv: Listening for events")

	_, err := m.conn.Call("observe_property", 42, "time-pos")
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
		m.playbackMu.Lock()
		if m.Playback != nil {
			m.Playback.IsRunning = false
		}
		m.playbackMu.Unlock()
		err := m.conn.Close()
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to close connection")
		}
		stopListening <- struct{}{}
	}()

	// Listen for events
	for event := range events {
		m.applyPlaybackEvent(event)
		if event.Data != nil {
			m.subscribers.Range(func(key string, sub *Subscriber) bool {
				go func() {
					sub.eventCh <- event
				}()
				return true
			})
		}
	}
}

func (m *Mpv) GetPlaybackStatus() (*Playback, error) {
	m.playbackMu.RLock()
	defer m.playbackMu.RUnlock()
	if m.Playback == nil {
		return nil, errors.New("no playback status")
	}
	playback := m.snapshotPlaybackLocked()
	if !playback.IsRunning {
		return nil, errors.New("mpv is not running")
	}
	if playback.Filename == "" {
		return nil, errors.New("no media found")
	}
	if playback.Duration == 0 {
		return nil, errors.New("no duration found")
	}
	return playback, nil
}

func (m *Mpv) CloseAll() {
	m.Logger.Debug().Msg("mpv: Received close request")
	if m.conn != nil {
		err := m.conn.Close()
		if err != nil {
			m.Logger.Error().Err(err).Msg("mpv: Failed to close connection")
		}
	}
	m.terminate()
}

func (m *Mpv) terminate() {
	defer func() {
		if r := recover(); r != nil {
			m.Logger.Warn().Msgf("mpv: Termination panic")
		}
	}()
	m.Logger.Trace().Msg("mpv: Terminating")
	m.resetPlaybackStatus()
	m.publishDone()
	if m.cancel != nil {
		m.cancel()
	}
	if cmdCancel != nil {
		cmdCancel()
	}
	m.Logger.Trace().Msg("mpv: Terminated")
}

func (m *Mpv) Quit() {
	if cmdCancel != nil {
		cmdCancel()
	}
}

func (m *Mpv) Subscribe(id string) *Subscriber {
	sub := &Subscriber{
		eventCh:  make(chan *mpvipc.Event, 100),
		closedCh: make(chan struct{}),
	}
	m.subscribers.Set(id, sub)
	return sub
}

func (m *Mpv) Unsubscribe(id string) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	sub, ok := m.subscribers.Get(id)
	if !ok {
		return
	}
	close(sub.eventCh)
	close(sub.closedCh)
	m.subscribers.Delete(id)
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

func shouldAutoSocket(socketName string) bool {
	return socketName == "" || socketName == getLegacySocketName()
}

func getLegacySocketName() string {
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

// getDefaultSocketName returns a fresh default name for the socket/pipe
func getDefaultSocketName() string {
	suffix := fmt.Sprintf("%d_%d_%d", os.Getpid(), time.Now().UnixNano(), atomic.AddUint64(&socketCounter, 1))
	switch runtime.GOOS {
	case "windows":
		return "\\\\.\\pipe\\mpv_ipc_" + suffix
	case "linux":
		return filepath.Join(os.TempDir(), "mpv_socket_"+suffix)
	case "darwin":
		return filepath.Join(os.TempDir(), "mpv_socket_"+suffix)
	default:
		return filepath.Join(os.TempDir(), "mpv_socket_"+suffix)
	}
}

func createLaunchLogPath() (string, error) {
	file, err := os.CreateTemp("", "seanime-mpv-*.log")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func formatLaunchExitError(waitErr error, logPath string) error {
	if waitErr == nil {
		return nil
	}
	message := readLaunchErrorMessage(logPath)
	if message == "" {
		return waitErr
	}
	return fmt.Errorf("mpv: %s: %w", message, waitErr)
}

func readLaunchErrorMessage(logPath string) string {
	if logPath == "" {
		return ""
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}
	return extractLaunchErrorMessage(string(data))
}

func extractLaunchErrorMessage(content string) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	priority := []string{
		"Cannot open file",
		"Failed to open ",
		"Could not bind IPC socket",
	}

	for _, needle := range priority {
		for i := len(lines) - 1; i >= 0; i-- {
			message := cleanLaunchLogLine(lines[i])
			if strings.Contains(message, needle) {
				return message
			}
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.Contains(line, "][e][") {
			continue
		}
		message := cleanLaunchLogLine(line)
		if message != "" {
			return message
		}
	}

	return ""
}

func cleanLaunchLogLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}
	if idx := strings.LastIndex(trimmed, "] "); idx >= 0 {
		return strings.TrimSpace(trimmed[idx+2:])
	}
	return trimmed
}

// createCmd returns a new exec.Cmd instance.
func (m *Mpv) createCmd(filePath string, args ...string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	// Add user-defined arguments
	if m.Args != "" {
		userArgs, err := parseArgs(m.Args)
		if err != nil {
			m.Logger.Warn().Err(err).Msg("mpv: Failed to parse user arguments, using simple split")
			userArgs = strings.Fields(m.Args)
		}
		args = append(args, userArgs...)
	}

	if m.launchLogPath != "" && !containsLogFileArg(args) {
		args = append(args, "--log-file="+m.launchLogPath)
	}

	if filePath != "" {
		// escapedFilePath := url.PathEscape(filePath)
		args = append(args, filePath)
	}

	binaryPath := "mpv"
	switch m.AppPath {
	case "":
	default:
		binaryPath = m.AppPath
	}

	cmd = util.NewCmdCtx(cmdCtx, binaryPath, args...)

	m.Logger.Trace().Msgf("mpv: Command: %s", strings.Join(cmd.Args, " "))

	return cmd, nil
}

func containsLogFileArg(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == "--log-file" || strings.HasPrefix(args[i], "--log-file=") {
			return true
		}
	}
	return false
}

func (m *Mpv) takeLaunchError() error {
	if m.launchErrCh == nil {
		return nil
	}
	select {
	case err := <-m.launchErrCh:
		m.launchErrCh = nil
		return err
	default:
		return nil
	}
}

func (m *Mpv) applyPlaybackEvent(event *mpvipc.Event) {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	if m.Playback == nil {
		m.Playback = &Playback{}
	}

	switch event.Name {
	case "start-file":
		m.isFileLoaded = false
		m.startPlaybackSwitchLocked()
	case "file-loaded":
		m.isFileLoaded = true
		m.startPlaybackSwitchLocked()
	}

	if event.Data == nil {
		return
	}

	m.Playback.IsRunning = true

	switch event.ID {
	case 43:
		m.Playback.Paused = event.Data.(bool)
	case 42:
		m.Playback.Position = event.Data.(float64)
		m.freshPosition = true
	case 44:
		m.Playback.Duration = event.Data.(float64)
		m.freshDuration = true
	case 45:
		m.Playback.Filename = event.Data.(string)
	case 46:
		m.Playback.Filepath = event.Data.(string)
	}

	if m.playbackSwitch && m.freshPosition && m.freshDuration {
		m.playbackSwitch = false
	}
}

func (m *Mpv) startPlaybackSwitchLocked() {
	m.playbackSwitch = true
	m.freshPosition = false
	m.freshDuration = false
	if m.Playback != nil {
		m.Playback.Position = 0
	}
}

func (m *Mpv) snapshotPlaybackLocked() *Playback {
	playback := *m.Playback
	if m.playbackSwitch {
		playback.Position = 0
	}
	return &playback
}

func (m *Mpv) resetPlaybackStatus() {
	m.playbackMu.Lock()
	m.Logger.Trace().Msg("mpv: Resetting playback status")
	m.playbackSwitch = false
	m.isFileLoaded = false
	m.freshPosition = false
	m.freshDuration = false
	m.Playback.Filename = ""
	m.Playback.Filepath = ""
	m.Playback.Paused = false
	m.Playback.Position = 0
	m.Playback.Duration = 0
	m.Playback.IsRunning = false
	m.playbackMu.Unlock()
}

func (m *Mpv) publishDone() {
	defer func() {
		if r := recover(); r != nil {
			m.Logger.Warn().Msgf("mpv: Connection already closed")
		}
	}()
	m.subscribers.Range(func(key string, sub *Subscriber) bool {
		go func() {
			sub.closedCh <- struct{}{}
		}()
		return true
	})
}
