package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/directstream"
	"seanime/internal/extension"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/player"
	gojautil "seanime/internal/util/goja"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// the API is called VideoCore but controls MediaCore (VideoCore+MpvCore)

type VideoCore struct {
	ctx                 *AppContextImpl
	vm                  *goja.Runtime
	logger              *zerolog.Logger
	ext                 *extension.Extension
	scheduler           *gojautil.Scheduler
	listeners           *result.Map[string, *VideoCoreEventListener]
	mediacoreSubscriber *mediacore.Subscriber
	unsubscribeOnce     sync.Once
}

type VideoCoreEventListener struct {
	eventId    string
	listenerCh chan player.Event
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
	_ = vcObj.Set("setSkipData", p.setSkipData)
	_ = vcObj.Set("clearSkipData", p.clearSkipData)

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
	_ = vcObj.Set("getSkipData", p.getSkipData)

	// Initiate playback
	_ = vcObj.Set("playStream", p.playStream)
	_ = vcObj.Set("playLocalFile", p.playLocalFile)

	_ = obj.Set("videoCore", vcObj)
}

func (p *VideoCore) getDenshiClientId() string {
	wsEventManager, ok := p.ctx.WSEventManager().Get()
	if ok {
		ids := wsEventManager.GetClientIds()
		for _, id := range ids {
			platform := wsEventManager.GetClientPlatform(id)
			if platform == "denshi" {
				return id
			}
		}
	}
	return ""
}

func (p *VideoCore) playStream(streamUrl string, anidbEpisode string, media *anilist.BaseAnime) goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	dsManager, ok := p.ctx.DirectStreamManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("directstream manager not available")))
		return p.vm.ToValue(promise)
	}

	if streamUrl == "" || anidbEpisode == "" || media == nil {
		reject(p.vm.NewGoError(errors.New("playStream: streamUrl, anidbEpisode, and media are required")))
		return p.vm.ToValue(promise)
	}

	go func() {
		clientId := p.getDenshiClientId()

		opts := directstream.PlayUrlStreamOptions{
			ClientId:     clientId,
			StreamUrl:    streamUrl,
			AnidbEpisode: anidbEpisode,
			Media:        media,
		}
		playErr := dsManager.PlayUrlStream(context.Background(), opts)
		p.scheduler.ScheduleAsync(func() error {
			if playErr != nil {
				reject(p.vm.NewGoError(playErr))
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}

func (p *VideoCore) playLocalFile(path string) goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	dsManager, ok := p.ctx.DirectStreamManager().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("directstream manager not available")))
		return p.vm.ToValue(promise)
	}

	db, ok := p.ctx.Database().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("database not available")))
		return p.vm.ToValue(promise)
	}

	if path == "" {
		reject(p.vm.NewGoError(errors.New("playLocalFile: path is required")))
		return p.vm.ToValue(promise)
	}

	go func() {
		clientId := p.getDenshiClientId()

		lfs, _, err := db_bridge.GetLocalFiles(db)
		if err != nil {
			p.scheduler.ScheduleAsync(func() error {
				reject(p.vm.NewGoError(err))
				return nil
			})
			return
		}

		playErr := dsManager.PlayLocalFile(context.Background(), directstream.PlayLocalFileOptions{
			ClientId:   clientId,
			Path:       path,
			LocalFiles: lfs,
		})
		p.scheduler.ScheduleAsync(func() error {
			if playErr != nil {
				reject(p.vm.NewGoError(playErr))
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}

func (p *VideoCore) getEventType(event player.Event) string {
	switch event.(type) {
	case *player.PlaybackLoadedEvent:
		return "video-loaded"
	case *player.LoadedMetadataEvent:
		return "video-loaded-metadata"
	case *player.CanPlayEvent:
		return "video-can-play"
	case *player.PausedEvent:
		return "video-paused"
	case *player.ResumedEvent:
		return "video-resumed"
	case *player.StatusEvent:
		return "video-status"
	case *player.CompletedEvent:
		return "video-completed"
	case *player.FullscreenChangedEvent:
		return "video-fullscreen"
	case *player.PipChangedEvent:
		return "video-pip"
	case *player.SubtitleTrackChangedEvent:
		return "video-subtitle-track"
	case *player.AudioTrackChangedEvent:
		return "video-audio-track"
	case *player.EndedEvent:
		return "video-ended"
	case *player.SeekedEvent:
		return "video-seeked"
	case *player.ErrorEvent:
		return "video-error"
	case *player.TerminatedEvent:
		return "video-terminated"
	case *player.SubtitleFileUploadedEvent:
		return "subtitle-file-uploaded"
	case *player.PlaylistStateEvent:
		return "video-playlist"
	default:
		return ""
	}
}

func (p *VideoCore) convertEventToJSObject(event player.Event) goja.Value {
	// extract base metadata first
	session := event.GetSessionKey()

	playerType := "web"
	playbackType := "localfile"
	playbackId := session.PlaybackID
	clientId := session.ClientID

	if coordinator, ok := p.ctx.MediacoreCoordinator().Get(); ok {
		if info, ok := coordinator.GetActivePlaybackInfo(); ok && info != nil {
			playerType = string(info.Renderer)
			playbackType = string(info.PlaybackType)
			if playbackId == "" {
				playbackId = info.ID
			}
		}
	}

	if playerType == "" || playerType == "web" || playerType == "native" || playerType == "mpv" {
		if session.Target == "mpvcore" {
			playerType = "mpv"
		} else {
			wsEventManager, ok := p.ctx.WSEventManager().Get()
			if ok && clientId != "" {
				platform := wsEventManager.GetClientPlatform(clientId)
				if platform == "denshi" {
					playerType = "native"
				} else {
					playerType = "web"
				}
			} else {
				if session.Target == "videocore" {
					// default to web if not explicitly denshi client
					playerType = "web"
				}
			}
		}
	}

	// map playerType format
	switch playerType {
	case "web":
		playerType = "web"
	case "native":
		playerType = "native"
	case "mpv":
		playerType = "mpv"
	}

	base := baseVideoEvent{
		PlayerType:   playerType,
		PlaybackType: playbackType,
		PlaybackID:   playbackId,
		ClientID:     clientId,
	}

	var jsEvent interface{}

	switch ev := event.(type) {
	case *player.PlaybackLoadedEvent:
		jsEvent = &struct {
			baseVideoEvent
			State player.PlaybackState `json:"state"`
		}{
			baseVideoEvent: base,
			State:          ev.State,
		}
	case *player.LoadedMetadataEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
			Paused      bool    `json:"paused"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
			Paused:         ev.Paused,
		}
	case *player.CanPlayEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
			Paused      bool    `json:"paused"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
			Paused:         ev.Paused,
		}
	case *player.PausedEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
		}
	case *player.ResumedEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
		}
	case *player.StatusEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
			Paused      bool    `json:"paused"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
			Paused:         ev.Paused,
		}
	case *player.CompletedEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
		}
	case *player.FullscreenChangedEvent:
		jsEvent = &struct {
			baseVideoEvent
			Fullscreen bool `json:"fullscreen"`
		}{
			baseVideoEvent: base,
			Fullscreen:     ev.Fullscreen,
		}
	case *player.PipChangedEvent:
		jsEvent = &struct {
			baseVideoEvent
			Pip bool `json:"pip"`
		}{
			baseVideoEvent: base,
			Pip:            ev.Pip,
		}
	case *player.SubtitleTrackChangedEvent:
		trackNum := 0
		if n, ok := ev.TrackID.(int); ok {
			trackNum = n
		} else if f, ok := ev.TrackID.(float64); ok {
			trackNum = int(f)
		}
		kind := "event"
		if trackNum >= 1000 {
			kind = "file"
		}
		jsEvent = &struct {
			baseVideoEvent
			TrackNumber int    `json:"trackNumber"`
			Kind        string `json:"kind"`
		}{
			baseVideoEvent: base,
			TrackNumber:    trackNum,
			Kind:           kind,
		}
	case *player.AudioTrackChangedEvent:
		trackNum := 0
		if n, ok := ev.TrackID.(int); ok {
			trackNum = n
		} else if f, ok := ev.TrackID.(float64); ok {
			trackNum = int(f)
		}
		jsEvent = &struct {
			baseVideoEvent
			TrackNumber int  `json:"trackNumber"`
			IsHLS       bool `json:"isHLS"`
		}{
			baseVideoEvent: base,
			TrackNumber:    trackNum,
			IsHLS:          playbackType == "onlinestream",
		}
	case *player.EndedEvent:
		jsEvent = &struct {
			baseVideoEvent
			AutoNext bool `json:"autoNext"`
		}{
			baseVideoEvent: base,
			AutoNext:       ev.AutoNext,
		}
	case *player.SeekedEvent:
		jsEvent = &struct {
			baseVideoEvent
			CurrentTime float64 `json:"currentTime"`
			Duration    float64 `json:"duration"`
			Paused      bool    `json:"paused"`
		}{
			baseVideoEvent: base,
			CurrentTime:    ev.CurrentTime,
			Duration:       ev.Duration,
			Paused:         ev.Paused,
		}
	case *player.ErrorEvent:
		jsEvent = &struct {
			baseVideoEvent
			Error string `json:"error"`
		}{
			baseVideoEvent: base,
			Error:          ev.Error,
		}
	case *player.TerminatedEvent:
		jsEvent = &struct {
			baseVideoEvent
		}{
			baseVideoEvent: base,
		}
	case *player.SubtitleFileUploadedEvent:
		jsEvent = &struct {
			baseVideoEvent
			Filename string `json:"filename"`
			Content  string `json:"content"`
		}{
			baseVideoEvent: base,
			Filename:       ev.Filename,
			Content:        ev.Content,
		}
	case *player.PlaylistStateEvent:
		jsEvent = &struct {
			baseVideoEvent
			Playlist *player.PlaylistState `json:"playlist"`
		}{
			baseVideoEvent: base,
			Playlist:       ev.Playlist,
		}
	default:
		jsEvent = event
	}

	return p.vm.ToValue(jsEvent)
}

func (p *VideoCore) subscribeToEvents() {
	p.unsubscribeOnce = sync.Once{}
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return
	}
	sub := coordinator.Subscribe("__plugin_videocore_subscriber__" + p.ext.ID)
	p.mediacoreSubscriber = sub
	go func() {
		for event := range sub.Events() {
			p.listeners.Range(func(eventId string, listener *VideoCoreEventListener) bool {
				if listener.closed.Load() {
					return true
				}

				eventType := p.getEventType(event)
				if eventType == "" || eventType != listener.eventId {
					return true
				}

				select {
				case listener.listenerCh <- event:
				default:
				}
				return true
			})
		}
	}()
}

func (p *VideoCore) addEventListener(call goja.FunctionCall) goja.Value {
	_, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		panic(p.vm.NewTypeError("mediacore coordinator not found"))
	}

	eventId := gojautil.ExpectStringArg(p.vm, call, 0)
	callback := gojautil.ExpectFunctionArg(p.vm, call, 1)

	listener := &VideoCoreEventListener{
		eventId:    eventId,
		listenerCh: make(chan player.Event, 100),
	}

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

func (p *VideoCore) removeEventListener(call goja.FunctionCall) goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		panic(p.vm.NewTypeError("mediacore coordinator not found"))
	}

	eventId := gojautil.ExpectStringArg(p.vm, call, 0)

	if listener, ok := p.listeners.Pop(eventId); ok {
		listener.closed.Store(true)
		listener.closeOnce.Do(func() {
			close(listener.listenerCh)
		})
	}

	listenerCount := len(p.listeners.Keys())
	if listenerCount == 0 {
		p.unsubscribeOnce.Do(func() {
			if p.mediacoreSubscriber != nil {
				coordinator.Unsubscribe(p.mediacoreSubscriber.GetID())
				p.mediacoreSubscriber = nil
			}
		})
	}

	return goja.Undefined()
}

type baseVideoEvent struct {
	PlayerType   string `json:"playerType"`
	PlaybackType string `json:"playbackType"`
	PlaybackID   string `json:"playbackId"`
	ClientID     string `json:"clientId"`
}

func (p *VideoCore) pause() error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandPause})
	}
	return nil
}

func (p *VideoCore) resume() error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandResume})
	}
	return nil
}

func (p *VideoCore) seek(seconds float64) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSeek, Payload: seconds})
	}
	return nil
}

func (p *VideoCore) seekTo(seconds float64) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSeekTo, Payload: seconds})
	}
	return nil
}

func (p *VideoCore) terminate() error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		coordinator.Terminate(session)
	}
	return nil
}

func (p *VideoCore) setFullscreen(fullscreen bool) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSetFullscreen, Payload: fullscreen})
	}
	return nil
}

func (p *VideoCore) setPip(pip bool) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSetPip, Payload: pip})
	}
	return nil
}

func (p *VideoCore) showMessage(call goja.FunctionCall) goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return goja.Undefined()
	}

	message := gojautil.ExpectStringArg(p.vm, call, 0)
	duration := 3000
	durationVal := call.Argument(1)
	if !goja.IsUndefined(durationVal) && !goja.IsNull(durationVal) {
		duration = int(durationVal.ToInteger())
	}

	if session, ok := coordinator.GetSession(); ok {
		_ = coordinator.Execute(session, player.Command{
			Type: player.CommandShowMessage,
			Payload: player.ShowMessagePayload{
				Message:  message,
				Duration: duration,
			},
		})
	}
	return goja.Undefined()
}

func (p *VideoCore) setSkipData(call goja.FunctionCall) goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		panic(p.vm.NewTypeError("mediacore coordinator not found"))
	}

	arg := call.Argument(0)
	if goja.IsUndefined(arg) || goja.IsNull(arg) {
		coordinator.ClearSkipData()
		return goja.Undefined()
	}

	marshaled, err := json.Marshal(arg.Export())
	if err != nil {
		panic(p.vm.NewTypeError("invalid skip data payload"))
	}

	var skipData player.SkipData
	if err := json.Unmarshal(marshaled, &skipData); err != nil {
		panic(p.vm.NewTypeError("invalid skip data payload"))
	}

	coordinator.SetSkipData(&skipData)
	return goja.Undefined()
}

func (p *VideoCore) clearSkipData() error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	coordinator.ClearSkipData()
	return nil
}

func (p *VideoCore) setSubtitleTrack(trackNumber int) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSetSubtitleTrack, Payload: trackNumber})
	}
	return nil
}

func (p *VideoCore) addSubtitleTrack(track mkvparser.TrackInfo) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandAddSubtitleTrack, Payload: &track})
	}
	return nil
}

func (p *VideoCore) addExternalSubtitleTrack(track player.SubtitleTrack) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandAddExternalSubtitleTrack, Payload: &track})
	}
	return nil
}

func (p *VideoCore) setMediaCaptionTrack(trackIndex int) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSetMediaCaptionTrack, Payload: trackIndex})
	}
	return nil
}

func (p *VideoCore) addMediaCaptionTrack(track interface{}) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandAddMediaCaptionTrack, Payload: track})
	}
	return nil
}

func (p *VideoCore) setAudioTrack(trackNumber int) error {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return nil
	}
	if session, ok := coordinator.GetActiveSession(); ok {
		return coordinator.Execute(session, player.Command{Type: player.CommandSetAudioTrack, Payload: trackNumber})
	}
	return nil
}

func (p *VideoCore) sendGetFullscreen() error {
	return nil
}

func (p *VideoCore) sendGetPip() error {
	return nil
}

func (p *VideoCore) sendGetAnime4K() error {
	return nil
}

func (p *VideoCore) sendGetSubtitleTrack() error {
	return nil
}

func (p *VideoCore) sendGetAudioTrack() error {
	return nil
}

func (p *VideoCore) sendGetMediaCaptionTrack() error {
	return nil
}

func (p *VideoCore) sendGetPlaybackState() error {
	return nil
}

func (p *VideoCore) getPlaybackStatus() goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return goja.Undefined()
	}

	status, ok := coordinator.GetActivePlaybackStatus()
	if !ok {
		return goja.Undefined()
	}

	return p.vm.ToValue(status)
}

func (p *VideoCore) getPlaybackState() goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return goja.Undefined()
	}

	state, ok := coordinator.GetActivePlaybackState()
	if !ok {
		return goja.Undefined()
	}

	if state.PlayerType == "" {
		if state.PlaybackInfo != nil {
			switch state.PlaybackInfo.Renderer {
			case player.RendererWeb:
				state.PlayerType = "web"
			case player.RendererNative:
				state.PlayerType = "native"
			case player.RendererMpv:
				state.PlayerType = "mpv"
			default:
				state.PlayerType = string(state.PlaybackInfo.Renderer)
			}
		} else {
			// fallback if playback info is nil
			wsEventManager, ok := p.ctx.WSEventManager().Get()
			if ok && state.ClientID != "" {
				platform := wsEventManager.GetClientPlatform(state.ClientID)
				if platform == "denshi" {
					state.PlayerType = "native"
				} else {
					state.PlayerType = "web"
				}
			} else {
				state.PlayerType = "web"
			}
		}
	}

	return p.vm.ToValue(&state)
}

func (p *VideoCore) getCurrentPlaybackInfo() goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return goja.Undefined()
	}

	info, ok := coordinator.GetActivePlaybackInfo()
	if !ok || info == nil {
		return goja.Undefined()
	}

	return p.vm.ToValue(info)
}

func (p *VideoCore) getCurrentMedia() goja.Value {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return goja.Undefined()
	}

	info, ok := coordinator.GetActivePlaybackInfo()
	if !ok || info == nil || info.Media == nil {
		return goja.Undefined()
	}

	return p.vm.ToValue(info.Media)
}

func (p *VideoCore) getCurrentClientId() string {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return ""
	}

	session, ok := coordinator.GetActiveSession()
	if !ok {
		return ""
	}

	return session.ClientID
}

func (p *VideoCore) getCurrentPlayerType() string {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return ""
	}

	info, ok := coordinator.GetActivePlaybackInfo()
	if !ok || info == nil {
		return ""
	}

	switch info.Renderer {
	case player.RendererWeb:
		return "web"
	case player.RendererNative:
		return "native"
	case player.RendererMpv:
		return "mpv"
	default:
		return string(info.Renderer)
	}
}

func (p *VideoCore) getCurrentPlaybackType() string {
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		return ""
	}

	info, ok := coordinator.GetActivePlaybackInfo()
	if !ok || info == nil {
		return ""
	}

	return string(info.PlaybackType)
}

func (p *VideoCore) getSkipData() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("mediacore coordinator not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		skipData, ok := coordinator.GetSkipData()
		p.scheduler.ScheduleAsync(func() error {
			if ok && skipData != nil {
				resolve(p.vm.ToValue(skipData))
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

	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("mediacore coordinator not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		playlist, ok := coordinator.GetPlaylist()
		p.scheduler.ScheduleAsync(func() error {
			if ok && playlist != nil {
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
	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		panic(p.vm.NewTypeError("mediacore coordinator not found"))
	}

	which := gojautil.ExpectStringArg(p.vm, call, 0)
	if session, ok := coordinator.GetActiveSession(); ok {
		_ = coordinator.Execute(session, player.Command{Type: player.CommandPlayPlaylistEpisode, Payload: which})
	}

	return goja.Undefined()
}

func (p *VideoCore) pullStatus() goja.Value {
	promise, resolve, reject := p.vm.NewPromise()

	coordinator, ok := p.ctx.MediacoreCoordinator().Get()
	if !ok {
		reject(p.vm.NewGoError(errors.New("mediacore coordinator not found")))
		return p.vm.ToValue(promise)
	}

	go func() {
		status, ok := coordinator.PullStatus()
		p.scheduler.ScheduleAsync(func() error {
			if ok {
				resolve(p.vm.ToValue(status))
			} else {
				resolve(goja.Undefined())
			}
			return nil
		})
	}()

	return p.vm.ToValue(promise)
}

func (p *VideoCore) getTextTracks() goja.Value {
	promise, resolve, _ := p.vm.NewPromise()
	resolve(goja.Undefined())
	return p.vm.ToValue(promise)
}
