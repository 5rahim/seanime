package libmpv

import (
	"context"
	"errors"
	"fmt"
	"github.com/gen2brain/go-mpv"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/util"
	"sync"
)

type (
	LibMpv struct {
		logger      *zerolog.Logger
		playback    *Playback
		cancelFunc  context.CancelFunc
		instance    mo.Option[*mpv.Mpv]
		mu          *sync.Mutex
		playbackMu  *sync.RWMutex
		subscribers map[string]*Subscriber // Subscribers to the mpv events
	}

	Subscriber struct {
		ClosedCh chan struct{}
	}

	Playback struct {
		Filename  string
		Paused    bool
		Position  float64
		Duration  float64
		IsRunning bool
		Filepath  string
	}
)

func New(logger *zerolog.Logger) *LibMpv {
	return &LibMpv{
		logger:      logger,
		mu:          &sync.Mutex{},
		playbackMu:  &sync.RWMutex{},
		playback:    &Playback{},
		instance:    mo.None[*mpv.Mpv](),
		subscribers: make(map[string]*Subscriber),
	}
}

func (m *LibMpv) OpenAndPlay(filepath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Debug().Str("filepath", filepath).Msg("libmpv: Opening and playing media")

	// If mpv is already running, replace the file
	if m.isRunning() {
		m.logger.Debug().Msg("libmpv: mpv is already running, replacing file")
		err := m.replaceFile(filepath)
		if err != nil {
			return err
		}
		return nil
	}

	m.logger.Debug().Msg("libmpv: Creating new MPV instance")

	// Create a new mpv instance
	mpvInstance := mpv.New()
	m.instance = mo.Some[*mpv.Mpv](mpvInstance)

	// Initialize the mpv instance
	err := m.instance.MustGet().Initialize()
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Error initializing mpv instance")
		m.terminate()
		return err
	}

	var ctx context.Context
	ctx, m.cancelFunc = context.WithCancel(context.Background())

	// Load the file
	err = m.loadFile(filepath)
	if err != nil {
		m.terminate()
		return err
	}

	go m.listenToEvents(ctx)

	return nil
}

func (m *LibMpv) listenToEvents(ctx context.Context) {
	if !m.isRunning() {
		m.logger.Error().Msg("libmpv: Cannot listen to events, mpv is not running")
		return
	}
	defer func() {
		m.logger.Debug().Msg("libmpv: Stopping event listener")
		m.mu.Lock()
		defer m.mu.Unlock()
		m.terminate()
		m.logger.Debug().Msg("libmpv: Event listener stopped")
	}()

	m.logger.Debug().Msg("libmpv: Listening to events")

	_ = m.instance.MustGet().SetOption("osc", mpv.FormatFlag, true)

	err := m.instance.MustGet().ObserveProperty(0, "time-pos", mpv.FormatDouble)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe time-pos property")
		return
	}
	err = m.instance.MustGet().ObserveProperty(1, "pause", mpv.FormatFlag)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe pause property")
		return
	}
	err = m.instance.MustGet().ObserveProperty(2, "duration", mpv.FormatDouble)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe duration property")
		return
	}
	err = m.instance.MustGet().ObserveProperty(3, "filename", mpv.FormatString)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe filename property")
		return
	}
	err = m.instance.MustGet().ObserveProperty(4, "path", mpv.FormatString)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe path property")
		return
	}
	err = m.instance.MustGet().ObserveProperty(5, "playback-abort", mpv.FormatFlag)
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Failed to observe playback-abort property")
		return
	}

loop:
	for {
		select {
		case <-ctx.Done():
			m.logger.Debug().Msg("libmpv: Context cancelled, stopping event listener")
			break loop
		default:
			e := m.instance.MustGet().WaitEvent(10000)

			switch e.EventID {
			case mpv.EventPropertyChange:
				m.playback.IsRunning = true

				prop := e.Property()

				switch prop.Name {
				case "filename":
					if prop.Data != nil {
						ret := m.instance.MustGet().GetPropertyString("filename")
						m.playback.Filename = ret
					}
				case "path":
					if prop.Data != nil {
						ret := m.instance.MustGet().GetPropertyString("path")
						m.playback.Filepath = ret
					}
				case "duration":
					if prop.Data != nil {
						data := prop.Data.(float64)
						m.playback.Duration = data
					}
				case "time-pos":
					if prop.Data != nil {
						data := prop.Data.(float64)
						m.playback.Position = data
					}
				case "pause":
					if prop.Data != nil {
						data := prop.Data.(int)
						m.playback.Paused = data == 1
					}
				case "playback-abort":
					aborted := prop.Data.(int)
					if aborted == 1 && m.playback.Filename != "" {
						break loop
					}
				}
			}
		}
	}
}

func (m *LibMpv) replaceFile(filepath string) error {
	return m.loadFile(filepath)
}

func (m *LibMpv) loadFile(filepath string) error {
	m.logger.Trace().Str("filepath", filepath).Msg("libmpv: Loading file")

	if !m.isRunning() {
		m.logger.Error().Msg("libmpv: Cannot load file, mpv is not running")
		return fmt.Errorf("cannot load file, mpv is not running")
	}

	err := m.instance.MustGet().Command([]string{"loadfile", filepath})
	if err != nil {
		m.logger.Error().Err(err).Msg("libmpv: Error loading file")
		return err
	}

	m.logger.Trace().Msg("libmpv: File loaded")
	return nil
}

//////////// Subscribers

func (m *LibMpv) Subscribe(id string) *Subscriber {
	sub := &Subscriber{
		ClosedCh: make(chan struct{}),
	}
	m.subscribers[id] = sub
	return sub
}

func (m *LibMpv) Unsubscribe(id string) {
	delete(m.subscribers, id)
}

// Done returns a channel that will be closed when the mpv instance is terminated
func (s *Subscriber) Done() chan struct{} {
	return s.ClosedCh
}

//////////// Instance

func (m *LibMpv) isRunning() bool {
	return m.instance.IsPresent()
}

func (m *LibMpv) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terminate()
}

// Terminate should be called in a locked context
// It will terminate the mpv instance and destroy.
// It will also reset the playback status.
func (m *LibMpv) terminate() {
	defer util.HandlePanicInModuleThen("mediaplayers/libmpv/Terminate", func() {})

	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.cleanupWhenTerminated()
}

// cleanupWhenTerminated should be safe to call multiple times
func (m *LibMpv) cleanupWhenTerminated() {
	if m.instance.IsPresent() {
		m.instance.MustGet().TerminateDestroy()
		m.instance = mo.None[*mpv.Mpv]()
	}
	m.resetPlaybackStatus()
	m.publishDone()
}

func (m *LibMpv) publishDone() {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Warn().Msgf("mpv: Connection already closed")
		}
	}()
	for _, sub := range m.subscribers {
		close(sub.ClosedCh)
		sub.ClosedCh = make(chan struct{})
	}
}

//////////// Playback

func (m *LibMpv) resetPlaybackStatus() {
	m.playbackMu.Lock()
	m.logger.Debug().Msg("mpv: Resetting playback status")
	m.playback.Filename = ""
	m.playback.Filepath = ""
	m.playback.Paused = false
	m.playback.Position = 0
	m.playback.Duration = 0
	m.playbackMu.Unlock()
	return
}

func (m *LibMpv) GetPlaybackStatus() (*Playback, error) {
	m.playbackMu.RLock()
	defer m.playbackMu.RUnlock()
	if m.playback.IsRunning == false {
		return nil, errors.New("mpv is not running")
	}
	if m.playback == nil {
		return nil, errors.New("no playback status")
	}
	if m.playback.Filename == "" {
		return nil, errors.New("no media found")
	}
	return m.playback, nil
}
