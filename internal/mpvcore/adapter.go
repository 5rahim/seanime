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
	eventsCh chan player.Event
}

var _ mediacore.Backend = (*Adapter)(nil)

func NewAdapter(mc *MpvCore) *Adapter {
	id := "mpvcore:adapter:" + uuid.NewString()
	sub := mc.Subscribe(id)

	a := &Adapter{
		mc:       mc,
		sub:      sub,
		eventsCh: make(chan player.Event, 100),
	}

	a.startEventLoop()
	return a
}

func (a *Adapter) Target() player.Target {
	return player.TargetMpvCore
}

func (a *Adapter) OpenAndAwait(clientID, state string) {
	a.mc.OpenAndAwait(clientID, state)
}

func (a *Adapter) AbortOpen(clientID, reason string) {
	a.mc.AbortOpen(clientID, reason)
}

func (a *Adapter) Watch(clientID string, info *player.PlaybackInfo) {
	if info != nil {
		a.mc.Watch(clientID, info)
	}
}

func (a *Adapter) Error(clientID string, err error) {
	a.mc.Error(clientID, err)
}

func (a *Adapter) Execute(session player.SessionKey, cmd player.Command) error {
	switch cmd.Type {
	case player.CommandPause:
		a.mc.Pause()
	case player.CommandResume:
		a.mc.Resume()
	case player.CommandSeek:
		if sec, ok := cmd.Payload.(float64); ok {
			a.mc.Seek(sec)
		} else {
			return errors.New("invalid payload type for Seek")
		}
	case player.CommandSeekTo:
		if sec, ok := cmd.Payload.(float64); ok {
			a.mc.SeekTo(sec)
		} else {
			return errors.New("invalid payload type for SeekTo")
		}
	case player.CommandSetFullscreen:
		if val, ok := cmd.Payload.(bool); ok {
			a.mc.SetFullscreen(val)
		} else {
			return errors.New("invalid payload type for SetFullscreen")
		}
	case player.CommandSetPip:
		if val, ok := cmd.Payload.(bool); ok {
			a.mc.SetPip(val)
		} else {
			return errors.New("invalid payload type for SetPip")
		}
	case player.CommandSetAudioTrack:
		a.mc.SetAudioTrack(cmd.Payload)
	case player.CommandSetSubtitleTrack:
		a.mc.SetSubtitleTrack(cmd.Payload)
	case player.CommandAddSubtitleTrack, player.CommandAddExternalSubtitleTrack:
		if val, ok := cmd.Payload.(*player.SubtitleTrack); ok {
			a.mc.AddSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for AddSubtitleTrack")
		}
	case player.CommandPlayPlaylistEpisode:
		if val, ok := cmd.Payload.(string); ok {
			a.mc.PlayPlaylistEpisode(val)
		} else {
			return errors.New("invalid payload type for PlayPlaylistEpisode")
		}
	case player.CommandShowMessage:
		if val, ok := cmd.Payload.(player.ShowMessagePayload); ok {
			a.mc.ShowMessage(val.Message, val.Duration)
		} else {
			return errors.New("invalid payload type for ShowMessage")
		}
	case player.CommandSetSkipData:
		if val, ok := cmd.Payload.(*player.SkipData); ok {
			a.mc.SetSkipData(val)
		} else {
			return errors.New("invalid payload type for SetSkipData")
		}
	case player.CommandClearSkipData:
		a.mc.ClearSkipData()
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Type)
	}
	return nil
}

func (a *Adapter) Terminate(session player.SessionKey) {
	a.mc.Terminate()
}

func (a *Adapter) Events() <-chan player.Event {
	return a.eventsCh
}

func (a *Adapter) Close() error {
	a.mc.Unsubscribe(a.sub.GetID())
	return nil
}

func (a *Adapter) PullStatus() (player.PlaybackStatus, bool) {
	status, ok := a.mc.PullStatus()
	if !ok {
		return player.PlaybackStatus{}, false
	}
	return player.PlaybackStatus{
		ID:          status.PlaybackID,
		ClientID:    status.ClientID,
		Paused:      status.Paused,
		CurrentTime: status.CurrentTime,
		Duration:    status.Duration,
	}, true
}

func (a *Adapter) GetPlaylist() (*player.PlaylistState, bool) {
	playlist, ok := a.mc.GetPlaylist()
	if !ok || playlist == nil {
		return nil, false
	}
	return playlist, true
}

func (a *Adapter) GetSkipData() (*player.SkipData, bool) {
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

func (a *Adapter) mapEvent(ev VideoEvent) player.Event {
	session := player.SessionKey{
		Target:     player.TargetMpvCore,
		ClientID:   ev.GetClientID(),
		PlaybackID: ev.GetPlaybackID(),
	}

	base := player.BaseEvent{Session: session}

	switch e := ev.(type) {
	case *PlaybackLoadedEvent:
		return &player.PlaybackLoadedEvent{
			BaseEvent: base,
			State: player.PlaybackState{
				ClientID:     e.ClientID,
				PlaybackInfo: toMediaCorePlaybackInfo(e.State.PlaybackInfo),
			},
		}
	case *LoadedMetadataEvent:
		return &player.LoadedMetadataEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *CanPlayEvent:
		return &player.CanPlayEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *PausedEvent:
		return &player.PausedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *ResumedEvent:
		return &player.ResumedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *StatusEvent:
		return &player.StatusEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *SeekedEvent:
		return &player.SeekedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *CompletedEvent:
		return &player.CompletedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *EndedEvent:
		return &player.EndedEvent{
			BaseEvent: base,
			AutoNext:  e.AutoNext,
		}
	case *ErrorEvent:
		return &player.ErrorEvent{
			BaseEvent: base,
			Error:     e.Error,
		}
	case *TerminatedEvent:
		return &player.TerminatedEvent{
			BaseEvent: base,
		}
	case *FullscreenChangedEvent:
		return &player.FullscreenChangedEvent{
			BaseEvent:  base,
			Fullscreen: e.Fullscreen,
		}
	case *PipChangedEvent:
		return &player.PipChangedEvent{
			BaseEvent: base,
			Pip:       e.Pip,
		}
	case *AudioTrackChangedEvent:
		return &player.AudioTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackID,
		}
	case *SubtitleTrackChangedEvent:
		return &player.SubtitleTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackID,
		}
	case *PlaylistStateEvent:
		return &player.PlaylistStateEvent{
			BaseEvent: base,
			Playlist:  e.Playlist,
		}
	case *SkipDataEvent:
		return &player.SkipDataEvent{
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
