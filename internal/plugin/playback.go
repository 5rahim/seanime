package plugin

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpv"
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

type PlaybackMPV struct {
	mpv      *mpv.Mpv
	playback *Playback
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
	mpv := mpv.New(logger, "", "")
	playbackMPV := &PlaybackMPV{
		mpv:      mpv,
		playback: p,
	}
	_ = mpvObj.Set("openAndPlay", playbackMPV.openAndPlay)
	_ = mpvObj.Set("onEvent", playbackMPV.onEvent)
	_ = mpvObj.Set("getConnection", playbackMPV.getConnection)
	_ = mpvObj.Set("stop", playbackMPV.stop)
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
		Reason string `json:"reason"`
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
func (p *PlaybackMPV) openAndPlay(filePath string) error {
	return p.mpv.OpenAndPlay(filePath)
}

func (p *PlaybackMPV) onEvent(callback func(event *mpvipc.Event, closed bool)) (func(), error) {
	id := p.playback.ext.ID + "_mpv"
	sub := p.mpv.Subscribe(id)

	go func() {
		for event := range sub.Events() {
			callback(event, false)
		}
	}()

	go func() {
		for range sub.Closed() {
			callback(nil, true)
		}
	}()

	cancelFn := func() {
		p.mpv.Unsubscribe(id)
	}

	return cancelFn, nil
}

func (p *PlaybackMPV) stop() error {
	p.mpv.CloseAll()
	return nil
}

func (p *PlaybackMPV) getConnection() goja.Value {
	conn, err := p.mpv.GetOpenConnection()
	if err != nil {
		return goja.Undefined()
	}
	return p.playback.vm.ToValue(conn)
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
						Reason string `json:"reason"`
					}{
						Reason: ret,
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
						Reason string `json:"reason"`
					}{
						Reason: ret,
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
