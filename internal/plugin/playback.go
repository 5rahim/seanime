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
	"github.com/google/uuid"
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
	_ = playbackObj.Set("pause", p.pause)
	_ = playbackObj.Set("resume", p.resume)
	_ = playbackObj.Set("seek", p.seek)
	_ = playbackObj.Set("cancel", p.cancel)
	_ = playbackObj.Set("getNextEpisode", p.getNextEpisode)
	_ = playbackObj.Set("playNextEpisode", p.playNextEpisode)
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

func (p *PlaybackMPV) openAndPlay(filePath string) goja.Value {
	promise, resolve, reject := p.playback.vm.NewPromise()

	go func() {
		err := p.mpv.OpenAndPlay(filePath)
		p.playback.scheduler.ScheduleAsync(func() error {
			if err != nil {
				jsErr := p.playback.vm.NewGoError(err)
				reject(jsErr)
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return p.playback.vm.ToValue(promise)
}

func (p *PlaybackMPV) onEvent(callback func(event *mpvipc.Event, closed bool)) (func(), error) {
	id := p.playback.ext.ID + "_mpv"
	sub := p.mpv.Subscribe(id)

	go func() {
		for event := range sub.Events() {
			p.playback.scheduler.ScheduleAsync(func() error {
				callback(event, false)
				return nil
			})
		}
	}()

	go func() {
		for range sub.Closed() {
			p.playback.scheduler.ScheduleAsync(func() error {
				callback(nil, true)
				return nil
			})
		}
	}()

	cancelFn := func() {
		p.mpv.Unsubscribe(id)
	}

	return cancelFn, nil
}

func (p *PlaybackMPV) stop() goja.Value {
	promise, resolve, _ := p.playback.vm.NewPromise()

	go func() {
		p.mpv.CloseAll()
		p.playback.scheduler.ScheduleAsync(func() error {
			resolve(goja.Undefined())
			return nil
		})
	}()

	return p.playback.vm.ToValue(promise)
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
func (p *Playback) registerEventListener(callback func(event *PlaybackEvent)) (func(), error) {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return nil, errors.New("playback manager not found")
	}

	id := uuid.New().String()

	subscriber := playbackManager.SubscribeToPlaybackStatus(id)

	go func() {
		for event := range subscriber.EventCh {
			switch e := event.(type) {
			case playbackmanager.PlaybackStatusChangedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						Status: &e.Status,
						State:  &e.State,
					})
					return nil
				})
			case playbackmanager.VideoStartedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsVideoStarted: true,
						StartedEvent: &struct {
							Filename string `json:"filename"`
						}{
							Filename: e.Filename,
						},
					})
					return nil
				})
			case playbackmanager.VideoStoppedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsVideoStopped: true,
						StoppedEvent: &struct {
							Reason string `json:"reason"`
						}{
							Reason: e.Reason,
						},
					})
					return nil
				})
			case playbackmanager.VideoCompletedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsVideoCompleted: true,
						CompletedEvent: &struct {
							Filename string `json:"filename"`
						}{
							Filename: e.Filename,
						},
					})
					return nil
				})
			case playbackmanager.StreamStateChangedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						State: &e.State,
					})
					return nil
				})
			case playbackmanager.StreamStatusChangedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						Status: &e.Status,
					})
					return nil
				})
			case playbackmanager.StreamStartedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsStreamStarted: true,
						StartedEvent: &struct {
							Filename string `json:"filename"`
						}{
							Filename: e.Filename,
						},
					})
					return nil
				})
			case playbackmanager.StreamStoppedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsStreamStopped: true,
						StoppedEvent: &struct {
							Reason string `json:"reason"`
						}{
							Reason: e.Reason,
						},
					})
					return nil
				})
			case playbackmanager.StreamCompletedEvent:
				p.scheduler.ScheduleAsync(func() error {
					callback(&PlaybackEvent{
						IsStreamCompleted: true,
						CompletedEvent: &struct {
							Filename string `json:"filename"`
						}{
							Filename: e.Filename,
						},
					})
					return nil
				})
			}
		}
	}()

	cancelFn := func() {
		playbackManager.UnsubscribeFromPlaybackStatus(id)
	}

	return cancelFn, nil
}

func (p *Playback) pause() error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.Pause()
}

func (p *Playback) resume() error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.Resume()
}

func (p *Playback) seek(seconds float64) error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.Seek(seconds)
}

func (p *Playback) cancel() error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.Cancel()
}

func (p *Playback) getNextEpisode() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		nextEpisode := playbackManager.GetNextEpisode()
		p.scheduler.ScheduleAsync(func() error {
			resolve(p.vm.ToValue(nextEpisode))
			return nil
		})
	}()
	return p.vm.ToValue(promise)
}

func (p *Playback) playNextEpisode() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		err := playbackManager.PlayNextEpisode()
		p.scheduler.ScheduleAsync(func() error {
			if err != nil {
				reject(p.vm.NewGoError(err))
			} else {
				resolve(goja.Undefined())
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}
