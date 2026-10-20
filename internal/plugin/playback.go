package plugin

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/mediaplayers/mpvipc"
	gojautil "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Playback struct {
	ctx       *AppContextImpl
	vm        *goja.Runtime
	logger    *zerolog.Logger
	ext       *extension.Extension
	scheduler *gojautil.Scheduler
}

type ManualTrackingOptions struct {
	ClientId      string
	MediaId       int
	EpisodeNumber int
}

type PlaybackMPV struct {
	mpv      *mpv.Mpv
	playback *Playback
}

func (a *AppContextImpl) BindPlaybackToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) {
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
	_ = playbackObj.Set("seekTo", p.seekTo)
	_ = playbackObj.Set("cancel", p.cancel)
	_ = playbackObj.Set("startManualTracking", p.startManualTracking)
	_ = playbackObj.Set("syncCurrentProgress", p.syncCurrentProgress)
	_ = playbackObj.Set("cancelManualTracking", p.cancelManualTracking)
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
func (p *Playback) playUsingMediaPlayer(payload string) goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		err := playbackManager.StartPlayingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
			Payload: payload,
		})
		p.scheduler.ScheduleAsync(func() error {
			if err != nil {
				jsErr := p.vm.NewGoError(err)
				reject(jsErr)
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}

// streamUsingMediaPlayer starts streaming a video using the media player specified in the settings.
func (p *Playback) streamUsingMediaPlayer(windowTitle string, payload string, media *anilist.BaseAnime, aniDbEpisode string) goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		err := playbackManager.StartStreamingUsingMediaPlayer(windowTitle, &playbackmanager.StartPlayingOptions{
			Payload: payload,
		}, media, aniDbEpisode)
		p.scheduler.ScheduleAsync(func() error {
			if err != nil {
				jsErr := p.vm.NewGoError(err)
				reject(jsErr)
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
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
//	ctx.playback.registerEventListener((event) => {
//		console.log(event)
//	});
func (p *Playback) registerEventListener(callback func(event *PlaybackEvent)) (func(), error) {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return nil, errors.New("playback manager not found")
	}

	id := uuid.New().String()

	subscriber := playbackManager.SubscribeToPlaybackStatus(id)

	go p.listen(subscriber, callback)

	cancelFn := func() {
		playbackManager.UnsubscribeFromPlaybackStatus(id)
	}

	return cancelFn, nil
}

func (p *Playback) listen(subscriber *playbackmanager.PlaybackStatusSubscriber, callback func(event *PlaybackEvent)) {
	var state *playbackmanager.PlaybackState
	var status *mediaplayer.PlaybackStatus

	for event := range subscriber.EventCh {
		data := &PlaybackEvent{State: state, Status: status}
		switch e := event.(type) {
		case playbackmanager.PlaybackStatusChangedEvent:
			state, status = &e.State, &e.Status
			data.State, data.Status = state, status

		case playbackmanager.VideoStartedEvent:
			data.IsVideoStarted = true
			data.StartedEvent = &struct {
				Filename string `json:"filename"`
			}{Filename: e.Filename}

		case playbackmanager.VideoStoppedEvent:
			data.IsVideoStopped = true
			data.StoppedEvent = &struct {
				Reason string `json:"reason"`
			}{Reason: e.Reason}

		case playbackmanager.VideoCompletedEvent:
			data.IsVideoCompleted = true
			data.CompletedEvent = &struct {
				Filename string `json:"filename"`
			}{Filename: e.Filename}

		case playbackmanager.StreamStartedEvent:
			data.IsStreamStarted = true
			data.StartedEvent = &struct {
				Filename string `json:"filename"`
			}{Filename: e.Filename}

		case playbackmanager.StreamStoppedEvent:
			data.IsStreamStopped = true
			data.StoppedEvent = &struct {
				Reason string `json:"reason"`
			}{Reason: e.Reason}

		case playbackmanager.StreamCompletedEvent:
			data.IsStreamCompleted = true
			data.CompletedEvent = &struct {
				Filename string `json:"filename"`
			}{Filename: e.Filename}

		default:
			continue
		}
		p.scheduler.ScheduleAsync(func() error {
			callback(data)
			return nil
		})
	}
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

func (p *Playback) seekTo(seconds float64) error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.SeekTo(seconds)
}

func (p *Playback) cancel() error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	return playbackManager.Cancel()
}

func (p *Playback) startManualTracking(call goja.FunctionCall) goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	if len(call.Arguments) == 0 || goja.IsUndefined(call.Argument(0)) || goja.IsNull(call.Argument(0)) {
		reject(p.vm.NewGoError(errors.New("manual tracking options are required")))
		return p.vm.ToValue(promise)
	}

	optsObj := call.Argument(0).ToObject(p.vm)
	opts := ManualTrackingOptions{
		ClientId:      p.ext.ID,
		MediaId:       int(optsObj.Get("mediaId").ToInteger()),
		EpisodeNumber: int(optsObj.Get("episodeNumber").ToInteger()),
	}
	if clientIdValue := optsObj.Get("clientId"); !goja.IsUndefined(clientIdValue) && !goja.IsNull(clientIdValue) {
		if clientId := clientIdValue.String(); clientId != "" {
			opts.ClientId = clientId
		}
	}

	if opts.MediaId <= 0 {
		reject(p.vm.NewGoError(errors.New("manual tracking requires a valid mediaId")))
		return p.vm.ToValue(promise)
	}
	if opts.EpisodeNumber <= 0 {
		reject(p.vm.NewGoError(errors.New("manual tracking requires a valid episodeNumber")))
		return p.vm.ToValue(promise)
	}

	go func() {
		err := playbackManager.StartManualProgressTracking(&playbackmanager.StartManualProgressTrackingOptions{
			ClientId:      opts.ClientId,
			MediaId:       opts.MediaId,
			EpisodeNumber: opts.EpisodeNumber,
		})
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

func (p *Playback) syncCurrentProgress() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("playback manager not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		err := playbackManager.SyncCurrentProgress()
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

func (p *Playback) cancelManualTracking() error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}
	playbackManager.CancelManualProgressTracking()
	return nil
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
