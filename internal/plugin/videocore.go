package plugin

import (
	"errors"
	"seanime/internal/extension"
	"seanime/internal/mkvparser"
	gojautil "seanime/internal/util/goja"
	"seanime/internal/util/result"
	"seanime/internal/videocore"
	"sync"
	"sync/atomic"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type VideoCore struct {
	ctx                 *AppContextImpl
	vm                  *goja.Runtime
	logger              *zerolog.Logger
	ext                 *extension.Extension
	scheduler           *gojautil.Scheduler
	listeners           *result.Map[string, *VideoCoreEventListener]
	videoCoreSubscriber *videocore.Subscriber
	unsubscribeOnce     sync.Once
}

type VideoCoreEventListener struct {
	eventId    string
	listenerCh chan videocore.VideoEvent
	closed     atomic.Bool
	closeOnce  sync.Once
}

func (a *AppContextImpl) BindVideoCoreToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) {
	p := &VideoCore{
		ctx:       a,
		vm:        vm,
		logger:    logger,
		ext:       ext,
		scheduler: scheduler,
		listeners: result.NewMap[string, *VideoCoreEventListener](),
	}

	vcObj := vm.NewObject()
	// Event listeners
	_ = vcObj.Set("addEventListener", p.addEventListener)
	_ = vcObj.Set("removeEventListener", p.removeEventListener)

	// Playback control
	_ = vcObj.Set("pause", p.pause)
	_ = vcObj.Set("resume", p.resume)
	_ = vcObj.Set("seek", p.seek)
	_ = vcObj.Set("seekTo", p.seekTo)
	_ = vcObj.Set("terminate", p.terminate)
	_ = vcObj.Set("playEpisodeFromPlaylist", p.playEpisodeFromPlaylist)

	// UI control
	_ = vcObj.Set("setFullscreen", p.setFullscreen)
	_ = vcObj.Set("setPip", p.setPip)
	_ = vcObj.Set("showMessage", p.showMessage)

	// Track control
	_ = vcObj.Set("setSubtitleTrack", p.setSubtitleTrack)
	_ = vcObj.Set("addSubtitleTrack", p.addSubtitleTrack)
	_ = vcObj.Set("addExternalSubtitleTrack", p.addExternalSubtitleTrack)
	_ = vcObj.Set("setMediaCaptionTrack", p.setMediaCaptionTrack)
	_ = vcObj.Set("addMediaCaptionTrack", p.addMediaCaptionTrack)
	_ = vcObj.Set("setAudioTrack", p.setAudioTrack)

	// State requests
	_ = vcObj.Set("sendGetFullscreen", p.sendGetFullscreen)
	_ = vcObj.Set("sendGetPip", p.sendGetPip)
	_ = vcObj.Set("sendGetAnime4K", p.sendGetAnime4K)
	_ = vcObj.Set("sendGetSubtitleTrack", p.sendGetSubtitleTrack)
	_ = vcObj.Set("sendGetAudioTrack", p.sendGetAudioTrack)
	_ = vcObj.Set("sendGetMediaCaptionTrack", p.sendGetMediaCaptionTrack)
	_ = vcObj.Set("sendGetPlaybackState", p.sendGetPlaybackState)

	// Async getters
	_ = vcObj.Set("getPlaylist", p.getPlaylist)
	_ = vcObj.Set("pullStatus", p.pullStatus)
	_ = vcObj.Set("getTextTracks", p.getTextTracks)

	// Sync getters
	_ = vcObj.Set("getPlaybackStatus", p.getPlaybackStatus)
	_ = vcObj.Set("getPlaybackState", p.getPlaybackState)
	_ = vcObj.Set("getCurrentPlaybackInfo", p.getCurrentPlaybackInfo)
	_ = vcObj.Set("getCurrentMedia", p.getCurrentMedia)
	_ = vcObj.Set("getCurrentClientId", p.getCurrentClientId)
	_ = vcObj.Set("getCurrentPlayerType", p.getCurrentPlayerType)
	_ = vcObj.Set("getCurrentPlaybackType", p.getCurrentPlaybackType)

	//_ = vcObj.Set("startOnlinestreamWatchParty", p.startOnlinestreamWatchParty)

	_ = obj.Set("videoCore", vcObj)

}

type VideoCoreEvent struct {
}

// getEventType maps a VideoEvent to its event type identifier
func (p *VideoCore) getEventType(event videocore.VideoEvent) string {
	switch event.(type) {
	case *videocore.VideoLoadedEvent:
		return string(videocore.PlayerEventVideoLoaded)
	case *videocore.VideoLoadedMetadataEvent:
		return string(videocore.PlayerEventVideoLoadedMetadata)
	case *videocore.VideoCanPlayEvent:
		return string(videocore.PlayerEventVideoCanPlay)
	case *videocore.VideoPausedEvent:
		return string(videocore.PlayerEventVideoPaused)
	case *videocore.VideoResumedEvent:
		return string(videocore.PlayerEventVideoResumed)
	case *videocore.VideoStatusEvent:
		return string(videocore.PlayerEventVideoStatus)
	case *videocore.VideoCompletedEvent:
		return string(videocore.PlayerEventVideoCompleted)
	case *videocore.VideoFullscreenEvent:
		return string(videocore.PlayerEventVideoFullscreen)
	case *videocore.VideoPipEvent:
		return string(videocore.PlayerEventVideoPip)
	case *videocore.VideoSubtitleTrackEvent:
		return string(videocore.PlayerEventVideoSubtitleTrack)
	case *videocore.VideoMediaCaptionTrackEvent:
		return string(videocore.PlayerEventMediaCaptionTrack)
	case *videocore.VideoAnime4KEvent:
		return string(videocore.PlayerEventAnime4K)
	case *videocore.VideoAudioTrackEvent:
		return string(videocore.PlayerEventVideoAudioTrack)
	case *videocore.VideoEndedEvent:
		return string(videocore.PlayerEventVideoEnded)
	case *videocore.VideoSeekedEvent:
		return string(videocore.PlayerEventVideoSeeked)
	case *videocore.VideoErrorEvent:
		return string(videocore.PlayerEventVideoError)
	case *videocore.VideoTerminatedEvent:
		return string(videocore.PlayerEventVideoTerminated)
	case *videocore.VideoPlaybackStateEvent:
		return string(videocore.PlayerEventVideoPlaybackState)
	case *videocore.SubtitleFileUploadedEvent:
		return string(videocore.PlayerEventSubtitleFileUploaded)
	case *videocore.VideoPlaylistEvent:
		return string(videocore.PlayerEventVideoPlaylist)
	case *videocore.VideoTextTracksEvent:
		return string(videocore.PlayerEventVideoTextTracks)
	default:
		return ""
	}
}

func (p *VideoCore) convertEventToJSObject(event videocore.VideoEvent) goja.Value {
	return p.vm.ToValue(event)
}

func (p *VideoCore) subscribeToEvents() {
	p.unsubscribeOnce = sync.Once{}
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return
	}
	p.videoCoreSubscriber = videoCore.Subscribe("__plugin_videocore_subscriber__" + p.ext.ID)
	go func() {
		for event := range p.videoCoreSubscriber.Events() {
			p.listeners.Range(func(eventId string, listener *VideoCoreEventListener) bool {
				if listener.closed.Load() {
					return true
				}

				// Filter events based on the event type the listener is subscribed to
				eventType := p.getEventType(event)
				if eventType == "" || eventType != listener.eventId {
					return true
				}

				select {
				case listener.listenerCh <- event:
				default:
					// Channel is full, drop the event
				}
				return true
			})
		}
	}()
}

// addEventListener registers a subscriber for playback events.
//
//	Example:
//	ctx.videoCore.addEventListener("video-loaded", (event) => {
//		console.log(event)
//	});
func (p *VideoCore) addEventListener(call goja.FunctionCall) goja.Value {
	_, ok := p.ctx.VideoCore().Get()
	if !ok {
		panic(p.vm.NewTypeError("videocore not found"))
	}

	eventId := gojautil.ExpectStringArg(p.vm, call, 0)
	callback := gojautil.ExpectFunctionArg(p.vm, call, 1)

	listener := &VideoCoreEventListener{
		eventId:    eventId,
		listenerCh: make(chan videocore.VideoEvent, 100),
	}

	// If it's the first listener, subscribe to the videocore events
	listenerCount := len(p.listeners.Keys())
	if listenerCount == 0 {
		p.subscribeToEvents()
	}

	p.listeners.Set(eventId, listener)

	go func() {
		for e := range listener.listenerCh {
			if listener.closed.Load() {
				return
			}
			p.scheduler.ScheduleAsync(func() error {
				eventObj := p.convertEventToJSObject(e)
				_, err := callback(goja.Undefined(), eventObj)
				if err != nil {
					p.logger.Error().Err(err).Msgf("plugin: Error calling videoCore event callback for event %s", eventId)
				}
				return nil
			})
		}
	}()

	return goja.Undefined()
}

// removeEventListener removes a playback event listener.
//
//	Example:
//	ctx.videoCore.removeEventListener("video-loaded");
func (p *VideoCore) removeEventListener(call goja.FunctionCall) goja.Value {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		panic(p.vm.NewTypeError("videocore not found"))
	}

	eventId := gojautil.ExpectStringArg(p.vm, call, 0)

	if listener, ok := p.listeners.Pop(eventId); ok {
		listener.closed.Store(true)
		listener.closeOnce.Do(func() {
			close(listener.listenerCh)
		})
	}

	// If it's the last listener, unsubscribe from the videocore events
	listenerCount := len(p.listeners.Keys())
	if listenerCount == 0 {
		p.unsubscribeOnce.Do(func() {
			if p.videoCoreSubscriber != nil {
				videoCore.Unsubscribe(p.videoCoreSubscriber.GetId())
				p.videoCoreSubscriber = nil
			}
		})
	}

	return goja.Undefined()
}

func (p *VideoCore) pause() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.Pause()
	return nil
}

func (p *VideoCore) resume() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.Resume()
	return nil
}

func (p *VideoCore) seek(seconds float64) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.Seek(seconds)
	return nil
}

func (p *VideoCore) seekTo(seconds float64) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SeekTo(seconds)
	return nil
}

func (p *VideoCore) terminate() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.Terminate()
	return nil
}

func (p *VideoCore) getTextTracks() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("videocore not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		ret, ok := videoCore.GetTextTracks()
		p.scheduler.ScheduleAsync(func() error {
			if ok {
				resolve(p.vm.ToValue(ret))
			} else {
				resolve(goja.Undefined())
			}
			return nil
		})
	}()
	return p.vm.ToValue(promise)
}

func (p *VideoCore) getPlaylist() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("videocore not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		playlist, ok := videoCore.GetPlaylist()
		p.scheduler.ScheduleAsync(func() error {
			if ok {
				resolve(p.vm.ToValue(playlist))
			} else {
				resolve(goja.Undefined())
			}
			return nil
		})
	}()
	return p.vm.ToValue(promise)
}

func (p *VideoCore) playEpisodeFromPlaylist(call goja.FunctionCall) goja.Value {

	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		panic(p.vm.NewTypeError("videocore not found"))
	}

	which := gojautil.ExpectStringArg(p.vm, call, 0)
	videoCore.PlayPlaylistEpisode(which)

	return goja.Undefined()
}

// UI control methods

func (p *VideoCore) setFullscreen(fullscreen bool) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SetFullscreen(fullscreen)
	return nil
}

func (p *VideoCore) setPip(pip bool) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SetPip(pip)
	return nil
}

func (p *VideoCore) showMessage(message string, duration int) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.ShowMessage(message, duration)
	return nil
}

// Track control methods

func (p *VideoCore) setSubtitleTrack(trackNumber int) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SetSubtitleTrack(trackNumber)
	return nil
}

func (p *VideoCore) addSubtitleTrack(track mkvparser.TrackInfo) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}

	videoCore.AddSubtitleTrack(&track)
	return nil
}

func (p *VideoCore) addExternalSubtitleTrack(track videocore.VideoSubtitleTrack) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}

	videoCore.AddExternalSubtitleTrack(&track)
	return nil
}

func (p *VideoCore) setMediaCaptionTrack(trackIndex int) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SetMediaCaptionTrack(trackIndex)
	return nil
}

func (p *VideoCore) addMediaCaptionTrack(track interface{}) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}

	videoCore.AddMediaCaptionTrack(track)
	return nil
}

func (p *VideoCore) setAudioTrack(trackNumber int) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SetAudioTrack(trackNumber)
	return nil
}

// State request methods

func (p *VideoCore) sendGetFullscreen() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetFullscreen()
	return nil
}

func (p *VideoCore) sendGetPip() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetPip()
	return nil
}

func (p *VideoCore) sendGetAnime4K() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetAnime4K()
	return nil
}

func (p *VideoCore) sendGetSubtitleTrack() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetSubtitleTrack()
	return nil
}

func (p *VideoCore) sendGetAudioTrack() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetAudioTrack()
	return nil
}

func (p *VideoCore) sendGetMediaCaptionTrack() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetMediaCaptionTrack()
	return nil
}

func (p *VideoCore) sendGetPlaybackState() error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}
	videoCore.SendGetPlaybackState()
	return nil
}

// Async getter methods

func (p *VideoCore) pullStatus() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("videocore not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		status, ok := videoCore.PullStatus()
		p.scheduler.ScheduleAsync(func() error {
			if ok {
				_ = resolve(p.vm.ToValue(status))
			} else {
				_ = resolve(goja.Undefined())
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}

// Sync getter methods

func (p *VideoCore) getPlaybackStatus() goja.Value {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return goja.Undefined()
	}

	status, ok := videoCore.GetPlaybackStatus()
	if !ok {
		return goja.Undefined()
	}

	return p.vm.ToValue(status)
}

func (p *VideoCore) getPlaybackState() goja.Value {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return goja.Undefined()
	}

	state, ok := videoCore.GetPlaybackState()
	if !ok {
		return goja.Undefined()
	}

	return p.vm.ToValue(state)
}

func (p *VideoCore) getCurrentPlaybackInfo() goja.Value {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return goja.Undefined()
	}

	info, ok := videoCore.GetCurrentPlaybackInfo()
	if !ok {
		return goja.Undefined()
	}

	return p.vm.ToValue(info)
}

func (p *VideoCore) getCurrentMedia() goja.Value {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return goja.Undefined()
	}

	media, ok := videoCore.GetCurrentMedia()
	if !ok {
		return goja.Undefined()
	}

	return p.vm.ToValue(media)
}

func (p *VideoCore) getCurrentClientId() string {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return ""
	}

	return videoCore.GetCurrentClientId()
}

func (p *VideoCore) getCurrentPlayerType() string {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return ""
	}

	playerType, ok := videoCore.GetCurrentPlayerType()
	if !ok {
		return ""
	}

	return string(playerType)
}

func (p *VideoCore) getCurrentPlaybackType() string {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return ""
	}

	playbackType, ok := videoCore.GetCurrentPlaybackType()
	if !ok {
		return ""
	}

	return string(playbackType)
}

// Special methods

func (p *VideoCore) startOnlinestreamWatchParty(params videocore.OnlinestreamParams) error {
	videoCore, ok := p.ctx.VideoCore().Get()
	if !ok {
		return errors.New("videocore not found")
	}

	videoCore.StartOnlinestreamWatchParty(&params)
	return nil
}
