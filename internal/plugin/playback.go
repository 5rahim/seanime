package plugin

import (
	"errors"
	"seanime/internal/extension"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
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

	_ = playbackObj.Set("registerSubscriber", p.registerSubscriber)

	// MPV
	mpvObj := vm.NewObject()
	_ = mpvObj.Set("play", p.mpvPlay)
	_ = playbackObj.Set("mpv", mpvObj)

	// VLC
	vlcObj := vm.NewObject()
	_ = vlcObj.Set("play", p.vlcPlay)
	_ = playbackObj.Set("vlc", vlcObj)

	// MPC-HC
	mpcHcObj := vm.NewObject()
	_ = mpcHcObj.Set("play", p.mpcHcPlay)
	_ = playbackObj.Set("mpcHc", mpcHcObj)

	vm.Set("$playback", playbackObj)
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

func (p *Playback) registerSubscriber(id string, callback func(event *PlaybackEvent)) error {
	playbackManager, ok := p.ctx.PlaybackManager().Get()
	if !ok {
		return errors.New("playback manager not found")
	}

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

	return nil
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

func (p *Playback) mpvPlay(payload string) error {
	mediaPlayerRepository, ok := p.ctx.MediaPlayerRepository().Get()
	if !ok {
		return errors.New("media player repository not found")
	}

	defaultPlayer := mediaPlayerRepository.Default
	mediaPlayerRepository.Default = "mpv"
	defer func() {
		mediaPlayerRepository.Default = defaultPlayer
	}()
	return mediaPlayerRepository.Play(payload)
}

func (p *Playback) vlcPlay(payload string) error {
	mediaPlayerRepository, ok := p.ctx.MediaPlayerRepository().Get()
	if !ok {
		return errors.New("media player repository not found")
	}

	defaultPlayer := mediaPlayerRepository.Default
	mediaPlayerRepository.Default = "vlc"
	defer func() {
		mediaPlayerRepository.Default = defaultPlayer
	}()
	return mediaPlayerRepository.Play(payload)
}

func (p *Playback) mpcHcPlay(payload string) error {
	mediaPlayerRepository, ok := p.ctx.MediaPlayerRepository().Get()
	if !ok {
		return errors.New("media player repository not found")
	}

	defaultPlayer := mediaPlayerRepository.Default
	mediaPlayerRepository.Default = "mpc-hc"
	defer func() {
		mediaPlayerRepository.Default = defaultPlayer
	}()
	return mediaPlayerRepository.Play(payload)
}
