package plugin

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpvipc"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Playback struct {
	ctx       *AppContextImpl
	vm        *goja.Runtime
	logger    *zerolog.Logger
	ext       *extension.Extension
	scheduler *goja_util.Scheduler
}

// BindPlayback
//
// $playback interacts with Seanime's own playback and tracking system.
// It is used to play local files and stream media.
//
//	$playback.playUsingMediaPlayer(path) // Starts playback and tracking of a local file.
//	$playback.registerEventListener(id, callback) // Registers a callback for playback events.
//
// $mpv is used to interact with the MPV media player outside of Seanime's tracking system.
func (a *AppContextImpl) BindPlayback(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {
	p := &Playback{
		ctx:       a,
		vm:        vm,
		logger:    logger,
		ext:       ext,
		scheduler: scheduler,
	}

	playbackObj := vm.NewObject()
	_ = playbackObj.Set("playUsingMediaPlayer", p.playUsingMediaPlayer)
	_ = playbackObj.Set("streamUsingMediaPlayer", p.streamUsingMediaPlayer)
	_ = playbackObj.Set("registerEventListener", p.registerEventListener)
	vm.Set("$playback", playbackObj)

	// MPV
	mpvObj := vm.NewObject()
	_ = mpvObj.Set("newConnection", p.mpvNewConnection)
	_ = mpvObj.Set("registerEventListener", p.mpvRegisterEventListener)
	_ = vm.Set("$mpv", mpvObj)
}

func (a *AppContextImpl) BindPlaybackToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {
	p := &Playback{
		ctx:       a,
		vm:        vm,
		logger:    logger,
		ext:       ext,
		scheduler: scheduler,
	}

	playbackObj := vm.NewObject()
	_ = playbackObj.Set("playUsingMediaPlayer", p.playUsingMediaPlayer)
	_ = playbackObj.Set("streamUsingMediaPlayer", p.streamUsingMediaPlayer)
	_ = playbackObj.Set("registerEventListener", p.registerEventListener)
	_ = obj.Set("playback", playbackObj)

	// MPV
	mpvObj := vm.NewObject()
	_ = mpvObj.Set("newConnection", p.mpvNewConnection)
	_ = mpvObj.Set("registerEventListener", p.mpvRegisterEventListener)
	_ = obj.Set("mpv", mpvObj)
}

type PlaybackEvent struct {
	IsVideoStarted    bool `json:"isVideoStarted"`
	IsVideoStopped    bool `json:"isVideoStopped"`
	IsVideoCompleted  bool `json:"isVideoCompleted"`
	IsStreamStarted   bool `json:"isStreamStarted"`
	IsStreamStopped   bool `json:"isStreamStopped"`
	IsStreamCompleted bool `json:"isStreamCompleted"`
	StartedEvent      *struct {
		Filename string `json:"filename"`
	} `json:"startedEvent"`
	StoppedEvent *struct {
		Filename string `json:"filename"`
	} `json:"stoppedEvent"`
	CompletedEvent *struct {
		Filename string `json:"filename"`
	} `json:"completedEvent"`
	State  *playbackmanager.PlaybackState `json:"state"`
	Status *mediaplayer.PlaybackStatus    `json:"status"`
}

// playUsingMediaPlayer starts playback of a local file using the media player specified in the settings.
func (p *Playback) playUsingMediaPlayer(payload string) error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}

	return playbackManager.StartPlayingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
		Payload: payload,
	})
}

// streamUsingMediaPlayer starts streaming a video using the media player specified in the settings.
func (p *Playback) streamUsingMediaPlayer(windowTitle string, payload string, media *anilist.BaseAnime, aniDbEpisode string) error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}

	return playbackManager.StartStreamingUsingMediaPlayer(windowTitle, &playbackmanager.StartPlayingOptions{
		Payload: payload,
	}, media, aniDbEpisode)
}

////////////////////////////////////
// MPV
////////////////////////////////////

// mpvNewConnection creates a new MPV connection.
//
//	Example:
//	const conn = $mpv.newConnection("/tmp/mpv-socket")
func (p *Playback) mpvNewConnection(socketName string) (*mpvipc.Connection, error) {
	conn := mpvipc.NewConnection(socketName)
	return conn, nil
}

// mpvRegisterEventListener registers an event listener for the MPV connection.
//
//	Example:
//	const conn = $mpv.newConnection("/tmp/mpv-socket")
//	const cancel = $mpv.registerEventListener(conn, (event) => {
//		console.log(event)
//	})
//	ctx.setTimeout(() => {
//		cancel()
//		m.conn.close()
//	}, 1000)
func (p *Playback) mpvRegisterEventListener(conn *mpvipc.Connection, callback func(event *mpvipc.Event)) (func(), error) {
	if conn == nil || conn.IsClosed() {
		return nil, errors.New("mpv is not running")
	}
	events, stopListening := conn.NewEventListener()

	// rateLimit := 200 * time.Millisecond
	// lastEvent := time.Now()

	go func() {
		for event := range events {
			// if time.Since(lastEvent) < rateLimit {
			// 	continue
			// }
			// lastEvent = time.Now()
			p.scheduler.ScheduleAsync(func() error {
				callback(event)
				return nil
			})
		}
	}()

	cancel := func() {
		select {
		case stopListening <- struct{}{}:
		default:
			return
		}
	}

	return cancel, nil
}

// registerEventListener registers a subscriber for playback events.
//
//	Example:
//	$playback.registerEventListener("mySubscriber", (event) => {
//		console.log(event)
//	});
func (p *Playback) registerEventListener(id string, callback func(event *PlaybackEvent)) (func(), error) {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return nil, errors.New("playback manager not found")
	}

	id = p.ext.ID + "_" + id

	subscriber := playbackManager.SubscribeToPlaybackStatus(id)

	go func() {
		for ret := range subscriber.VideoStartedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsVideoStarted: true,
					StartedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.VideoStoppedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsVideoStopped: true,
					StoppedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.VideoCompletedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsVideoCompleted: true,
					CompletedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.StreamStartedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsStreamStarted: true,
					StartedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.StreamStoppedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsStreamStopped: true,
					StoppedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.StreamCompletedCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					IsStreamCompleted: true,
					CompletedEvent: &struct {
						Filename string `json:"filename"`
					}{
						Filename: ret,
					},
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.PlaybackStateCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					State: &ret,
				})
				return nil
			})
		}
	}()

	go func() {
		for ret := range subscriber.PlaybackStatusCh {
			p.scheduler.ScheduleAsync(func() error {
				callback(&PlaybackEvent{
					Status: &ret,
				})
				return nil
			})
		}
	}()

	cancelFn := func() {
		playbackManager.UnsubscribeFromPlaybackStatus(id)
	}

	return cancelFn, nil
}
