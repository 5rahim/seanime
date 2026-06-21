package mpvcore

import (
	"errors"
	"fmt"
	"seanime/internal/mediacore"
	"seanime/internal/player"

	"github.com/google/uuid"
)

type Adapter struct {
	mc       *MpvCore
	sub      *Subscriber
	eventsCh chan mediacore.Event
}

var _ mediacore.Backend = (*Adapter)(nil)

func NewAdapter(mc *MpvCore) *Adapter {
	id := "mpvcore:adapter:" + uuid.NewString()
	sub := mc.Subscribe(id)

	a := &Adapter{
		mc:       mc,
		sub:      sub,
		eventsCh: make(chan mediacore.Event, 100),
	}

	a.startEventLoop()
	return a
}

func (a *Adapter) Target() mediacore.Target {
	return mediacore.TargetMpvCore
}

func (a *Adapter) OpenAndAwait(clientID, state string) {
	a.mc.OpenAndAwait(clientID, state)
}

func (a *Adapter) AbortOpen(clientID, reason string) {
	a.mc.AbortOpen(clientID, reason)
}

func (a *Adapter) Watch(clientID string, info *mediacore.PlaybackInfo) {
	if info != nil {
		a.mc.Watch(clientID, info)
	}
}

func (a *Adapter) Error(clientID string, err error) {
	a.mc.Error(clientID, err)
}

func (a *Adapter) Execute(session mediacore.SessionKey, cmd mediacore.Command) error {
	switch cmd.Type {
	case mediacore.CommandPause:
		a.mc.Pause()
	case mediacore.CommandResume:
		a.mc.Resume()
	case mediacore.CommandSeek:
		if sec, ok := cmd.Payload.(float64); ok {
			a.mc.Seek(sec)
		} else {
			return errors.New("invalid payload type for Seek")
		}
	case mediacore.CommandSeekTo:
		if sec, ok := cmd.Payload.(float64); ok {
			a.mc.SeekTo(sec)
		} else {
			return errors.New("invalid payload type for SeekTo")
		}
	case mediacore.CommandSetFullscreen:
		if val, ok := cmd.Payload.(bool); ok {
			a.mc.SetFullscreen(val)
		} else {
			return errors.New("invalid payload type for SetFullscreen")
		}
	case mediacore.CommandSetPip:
		if val, ok := cmd.Payload.(bool); ok {
			a.mc.SetPip(val)
		} else {
			return errors.New("invalid payload type for SetPip")
		}
	case mediacore.CommandSetAudioTrack:
		a.mc.SetAudioTrack(cmd.Payload)
	case mediacore.CommandSetSubtitleTrack:
		a.mc.SetSubtitleTrack(cmd.Payload)
	case mediacore.CommandAddSubtitleTrack, mediacore.CommandAddExternalSubtitleTrack:
		if val, ok := cmd.Payload.(*mediacore.SubtitleTrack); ok {
			a.mc.AddSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for AddSubtitleTrack")
		}
	case mediacore.CommandPlayPlaylistEpisode:
		if val, ok := cmd.Payload.(string); ok {
			a.mc.PlayPlaylistEpisode(val)
		} else {
			return errors.New("invalid payload type for PlayPlaylistEpisode")
		}
	case mediacore.CommandShowMessage:
		if val, ok := cmd.Payload.(mediacore.ShowMessagePayload); ok {
			a.mc.ShowMessage(val.Message, val.Duration)
		} else {
			return errors.New("invalid payload type for ShowMessage")
		}
	case mediacore.CommandSetSkipData:
		if val, ok := cmd.Payload.(*mediacore.SkipData); ok {
			a.mc.SetSkipData(val)
		} else {
			return errors.New("invalid payload type for SetSkipData")
		}
	case mediacore.CommandClearSkipData:
		a.mc.ClearSkipData()
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Type)
	}
	return nil
}

func (a *Adapter) Terminate(session mediacore.SessionKey) {
	a.mc.Terminate()
}

func (a *Adapter) Events() <-chan mediacore.Event {
	return a.eventsCh
}

func (a *Adapter) Close() error {
	a.mc.Unsubscribe(a.sub.GetID())
	return nil
}

func (a *Adapter) PullStatus() (mediacore.PlaybackStatus, bool) {
	status, ok := a.mc.PullStatus()
	if !ok {
		return mediacore.PlaybackStatus{}, false
	}
	return mediacore.PlaybackStatus{
		ID:          status.PlaybackID,
		ClientID:    status.ClientID,
		Paused:      status.Paused,
		CurrentTime: status.CurrentTime,
		Duration:    status.Duration,
	}, true
}

func (a *Adapter) GetPlaylist() (*mediacore.PlaylistState, bool) {
	playlist, ok := a.mc.GetPlaylist()
	if !ok || playlist == nil {
		return nil, false
	}
	return playlist, true
}

func (a *Adapter) GetSkipData() (*mediacore.SkipData, bool) {
	skipData, ok := a.mc.GetSkipData()
	if !ok || skipData == nil {
		return nil, false
	}
	return skipData, true
}

func (a *Adapter) startEventLoop() {
	go func() {
		defer close(a.eventsCh)
		for ev := range a.sub.Events() {
			mapped := a.mapEvent(ev)
			if mapped != nil {
				a.eventsCh <- mapped
			}
		}
	}()
}

func (a *Adapter) mapEvent(ev VideoEvent) mediacore.Event {
	session := mediacore.SessionKey{
		Target:     mediacore.TargetMpvCore,
		ClientID:   ev.GetClientID(),
		PlaybackID: ev.GetPlaybackID(),
	}

	base := mediacore.BaseEvent{Session: session}

	switch e := ev.(type) {
	case *PlaybackLoadedEvent:
		return &mediacore.PlaybackLoadedEvent{
			BaseEvent: base,
			State: mediacore.PlaybackState{
				ClientID:     e.ClientID,
				PlaybackInfo: toMediaCorePlaybackInfo(e.State.PlaybackInfo),
			},
		}
	case *LoadedMetadataEvent:
		return &mediacore.LoadedMetadataEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *CanPlayEvent:
		return &mediacore.CanPlayEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *PausedEvent:
		return &mediacore.PausedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *ResumedEvent:
		return &mediacore.ResumedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *StatusEvent:
		return &mediacore.StatusEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *SeekedEvent:
		return &mediacore.SeekedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *CompletedEvent:
		return &mediacore.CompletedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *EndedEvent:
		return &mediacore.EndedEvent{
			BaseEvent: base,
			AutoNext:  e.AutoNext,
		}
	case *ErrorEvent:
		return &mediacore.ErrorEvent{
			BaseEvent: base,
			Error:     e.Error,
		}
	case *TerminatedEvent:
		return &mediacore.TerminatedEvent{
			BaseEvent: base,
		}
	case *FullscreenChangedEvent:
		return &mediacore.FullscreenChangedEvent{
			BaseEvent:  base,
			Fullscreen: e.Fullscreen,
		}
	case *PipChangedEvent:
		return &mediacore.PipChangedEvent{
			BaseEvent: base,
			Pip:       e.Pip,
		}
	case *AudioTrackChangedEvent:
		return &mediacore.AudioTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackID,
		}
	case *SubtitleTrackChangedEvent:
		return &mediacore.SubtitleTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackID,
		}
	case *PlaylistStateEvent:
		return &mediacore.PlaylistStateEvent{
			BaseEvent: base,
			Playlist:  e.Playlist,
		}
	case *SkipDataEvent:
		return &mediacore.SkipDataEvent{
			BaseEvent: base,
			SkipData:  e.SkipData,
		}
	}
	return nil
}

func toMediaCorePlaybackInfo(info *player.PlaybackInfo) *player.PlaybackInfo {
	if info == nil {
		return nil
	}
	copied := *info
	copied.Target = player.TargetMpvCore
	copied.Renderer = player.RendererMpv
	return &copied
}
