package videocore

import (
	"errors"
	"fmt"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"

	"github.com/google/uuid"
)

type Adapter struct {
	vc       *VideoCore
	np       *nativeplayer.NativePlayer
	sub      *Subscriber
	eventsCh chan mediacore.Event
}

var _ mediacore.Backend = (*Adapter)(nil)

func NewAdapter(vc *VideoCore, np *nativeplayer.NativePlayer) *Adapter {
	id := "videocore:adapter:" + uuid.NewString()
	sub := vc.Subscribe(id)

	a := &Adapter{
		vc:       vc,
		np:       np,
		sub:      sub,
		eventsCh: make(chan mediacore.Event, 100),
	}

	a.startEventLoop()
	return a
}

func (a *Adapter) Target() mediacore.Target {
	return mediacore.TargetVideoCore
}

func (a *Adapter) OpenAndAwait(clientID, state string) {
	if a.np != nil {
		a.np.OpenAndAwait(clientID, state)
	}
}

func (a *Adapter) AbortOpen(clientID, reason string) {
	if a.np != nil {
		a.np.AbortOpen(clientID, reason)
	}
}

func (a *Adapter) Watch(clientID string, info *mediacore.PlaybackInfo) {
	if a.np != nil && info != nil {
		a.np.Watch(clientID, toNativePlaybackInfo(info))
	}
}

func (a *Adapter) Error(clientID string, err error) {
	if a.np != nil {
		a.np.Error(clientID, err)
	} else {
		a.vc.Reset()
	}
}

func (a *Adapter) Execute(session mediacore.SessionKey, cmd mediacore.Command) error {
	switch cmd.Type {
	case mediacore.CommandPause:
		a.vc.Pause()
	case mediacore.CommandResume:
		a.vc.Resume()
	case mediacore.CommandSeek:
		if sec, ok := cmd.Payload.(float64); ok {
			a.vc.Seek(sec)
		} else {
			return errors.New("invalid payload type for Seek")
		}
	case mediacore.CommandSeekTo:
		if sec, ok := cmd.Payload.(float64); ok {
			a.vc.SeekTo(sec)
		} else {
			return errors.New("invalid payload type for SeekTo")
		}
	case mediacore.CommandSetFullscreen:
		if val, ok := cmd.Payload.(bool); ok {
			a.vc.SetFullscreen(val)
		} else {
			return errors.New("invalid payload type for SetFullscreen")
		}
	case mediacore.CommandSetPip:
		if val, ok := cmd.Payload.(bool); ok {
			a.vc.SetPip(val)
		} else {
			return errors.New("invalid payload type for SetPip")
		}
	case mediacore.CommandSetAudioTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetAudioTrack(val)
		} else {
			return errors.New("invalid payload type for SetAudioTrack")
		}
	case mediacore.CommandSetSubtitleTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for SetSubtitleTrack")
		}
	case mediacore.CommandAddSubtitleTrack:
		if val, ok := cmd.Payload.(*mkvparser.TrackInfo); ok {
			a.vc.AddSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for AddSubtitleTrack")
		}
	case mediacore.CommandAddExternalSubtitleTrack:
		if val, ok := cmd.Payload.(*mediacore.SubtitleTrack); ok {
			a.vc.AddExternalSubtitleTrack(toVideoSubtitleTrack(val))
		} else {
			return errors.New("invalid payload type for AddExternalSubtitleTrack")
		}
	case mediacore.CommandSetMediaCaptionTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetMediaCaptionTrack(val)
		} else {
			return errors.New("invalid payload type for SetMediaCaptionTrack")
		}
	case mediacore.CommandAddMediaCaptionTrack:
		a.vc.AddMediaCaptionTrack(cmd.Payload)
	case mediacore.CommandPlayPlaylistEpisode:
		if val, ok := cmd.Payload.(string); ok {
			a.vc.PlayPlaylistEpisode(val)
		} else {
			return errors.New("invalid payload type for PlayPlaylistEpisode")
		}
	case mediacore.CommandShowMessage:
		if val, ok := cmd.Payload.(mediacore.ShowMessagePayload); ok {
			a.vc.ShowMessage(val.Message, val.Duration)
		} else {
			return errors.New("invalid payload type for ShowMessage")
		}
	case mediacore.CommandSetSkipData:
		if val, ok := cmd.Payload.(*mediacore.SkipData); ok {
			a.vc.SetSkipData(toSkipData(val))
		} else {
			return errors.New("invalid payload type for SetSkipData")
		}
	case mediacore.CommandClearSkipData:
		a.vc.ClearSkipData()
	case mediacore.CommandStartOnlinestreamWatchParty:
		if val, ok := cmd.Payload.(*mediacore.OnlinestreamParams); ok {
			a.vc.StartOnlinestreamWatchParty(toOnlinestreamParams(val))
		} else {
			return errors.New("invalid payload type for StartOnlinestreamWatchParty")
		}
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Type)
	}
	return nil
}

func (a *Adapter) Terminate(session mediacore.SessionKey) {
	a.vc.Terminate()
}

func (a *Adapter) Events() <-chan mediacore.Event {
	return a.eventsCh
}

func (a *Adapter) Close() error {
	a.vc.Unsubscribe(a.sub.GetId())
	return nil
}

func (a *Adapter) PullStatus() (mediacore.PlaybackStatus, bool) {
	status, ok := a.vc.PullStatus()
	if !ok {
		return mediacore.PlaybackStatus{}, false
	}
	return mediacore.PlaybackStatus{
		ID:          status.PlaybackId,
		ClientID:    status.ClientId,
		Paused:      status.Paused,
		CurrentTime: status.CurrentTime,
		Duration:    status.Duration,
	}, true
}

func (a *Adapter) GetPlaylist() (*mediacore.PlaylistState, bool) {
	playlist, ok := a.vc.GetPlaylist()
	if !ok || playlist == nil {
		return nil, false
	}
	return toMediaCorePlaylistState(playlist), true
}

func (a *Adapter) GetSkipData() (*mediacore.SkipData, bool) {
	skipData, ok := a.vc.GetSkipData()
	if !ok || skipData == nil {
		return nil, false
	}
	return toMediaCoreSkipData(skipData), true
}

func (a *Adapter) startEventLoop() {
	go func() {
		defer close(a.eventsCh)
		for ev := range a.sub.Events() {
			mapped := a.toEvent(ev)
			if mapped != nil {
				a.eventsCh <- mapped
			}
		}
	}()
}

func (a *Adapter) toEvent(ev VideoEvent) mediacore.Event {
	session := mediacore.SessionKey{
		Target:     mediacore.TargetVideoCore,
		ClientID:   ev.GetClientId(),
		PlaybackID: ev.GetPlaybackId(),
	}

	base := mediacore.BaseEvent{Session: session}

	switch e := ev.(type) {
	case *VideoLoadedEvent:
		return &mediacore.PlaybackLoadedEvent{
			BaseEvent: base,
			State: mediacore.PlaybackState{
				ClientID:     e.ClientId,
				PlaybackInfo: toMediaCorePlaybackInfo(e.State.PlaybackInfo, e.PlayerType),
			},
		}
	case *VideoLoadedMetadataEvent:
		return &mediacore.LoadedMetadataEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoCanPlayEvent:
		return &mediacore.CanPlayEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoPausedEvent:
		return &mediacore.PausedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoResumedEvent:
		return &mediacore.ResumedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoStatusEvent:
		return &mediacore.StatusEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoSeekedEvent:
		return &mediacore.SeekedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoCompletedEvent:
		return &mediacore.CompletedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoEndedEvent:
		return &mediacore.EndedEvent{
			BaseEvent: base,
			AutoNext:  e.AutoNext,
		}
	case *VideoErrorEvent:
		return &mediacore.ErrorEvent{
			BaseEvent: base,
			Error:     e.Error,
		}
	case *VideoTerminatedEvent:
		return &mediacore.TerminatedEvent{
			BaseEvent: base,
		}
	case *VideoFullscreenEvent:
		return &mediacore.FullscreenChangedEvent{
			BaseEvent:  base,
			Fullscreen: e.Fullscreen,
		}
	case *VideoPipEvent:
		return &mediacore.PipChangedEvent{
			BaseEvent: base,
			Pip:       e.Pip,
		}
	case *VideoAudioTrackEvent:
		return &mediacore.AudioTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackNumber,
		}
	case *VideoSubtitleTrackEvent:
		return &mediacore.SubtitleTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackNumber,
		}
	case *SubtitleFileUploadedEvent:
		return &mediacore.SubtitleFileUploadedEvent{
			BaseEvent: base,
			Filename:  e.Filename,
			Content:   e.Content,
		}
	case *VideoPlaylistEvent:
		return &mediacore.PlaylistStateEvent{
			BaseEvent: base,
			Playlist:  toMediaCorePlaylistState(e.Playlist),
		}
	case *VideoSkipDataEvent:
		return &mediacore.SkipDataEvent{
			BaseEvent: base,
			SkipData:  toMediaCoreSkipData(e.SkipData),
		}
	}
	return nil
}

func toMediaCorePlaybackInfo(info *VideoPlaybackInfo, playerType PlayerType) *mediacore.PlaybackInfo {
	if info == nil {
		return nil
	}

	var renderer mediacore.Renderer
	if playerType == WebPlayer {
		renderer = mediacore.RendererWeb
	} else {
		renderer = mediacore.RendererNative
	}

	tracks := make([]*mediacore.SubtitleTrack, 0, len(info.SubtitleTracks))
	for _, track := range info.SubtitleTracks {
		if track == nil {
			continue
		}
		tracks = append(tracks, &mediacore.SubtitleTrack{
			Index:     track.Index,
			URI:       track.Src,
			SourceURL: track.Src,
			Content:   track.Content,
			Label:     track.Label,
			Language:  track.Language,
			Format:    track.Type,
			Default:   track.Default,
		})
	}

	sources := make([]*mediacore.VideoSource, 0, len(info.VideoSources))
	for _, src := range info.VideoSources {
		if src == nil {
			continue
		}
		sources = append(sources, &mediacore.VideoSource{
			Index:      src.Index,
			Resolution: src.Resolution,
			URL:        src.URL,
			Label:      src.Label,
			MoreInfo:   src.MoreInfo,
		})
	}

	var init *mediacore.InitialState
	if info.InitialState != nil {
		init = &mediacore.InitialState{
			CurrentTime: info.InitialState.CurrentTime,
			Paused:      info.InitialState.Paused,
		}
	}

	return &mediacore.PlaybackInfo{
		ID:                             info.Id,
		Target:                         mediacore.TargetVideoCore,
		Renderer:                       renderer,
		PlaybackType:                   mediacore.PlaybackType(info.PlaybackType),
		StreamURL:                      info.StreamURL,
		StreamPath:                     info.StreamPath,
		MkvMetadata:                    info.MkvMetadata,
		SubtitleTracks:                 tracks,
		VideoSources:                   sources,
		SelectedVideoSource:            info.SelectedVideoSource,
		PlaylistExternalEpisodeNumbers: info.PlaylistExternalEpisodeNumbers,
		DisableRestoreFromContinuity:   info.DisableRestoreFromContinuity,
		InitialState:                   init,
		Media:                          info.Media,
		Episode:                        info.Episode,
		LocalFile:                      info.LocalFile,
		OnlinestreamParams:             toMediaCoreOnlinestreamParams(info.OnlinestreamParams),
		IsNakamaWatchParty:             info.IsNakamaWatchParty,
	}
}

func toNativePlaybackInfo(info *mediacore.PlaybackInfo) *nativeplayer.PlaybackInfo {
	if info == nil {
		return nil
	}
	tracks := make([]*nativeplayer.VideoSubtitleTrack, 0, len(info.SubtitleTracks))
	for _, track := range info.SubtitleTracks {
		if track == nil {
			continue
		}
		tracks = append(tracks, &nativeplayer.VideoSubtitleTrack{
			Index:    track.Index,
			Src:      track.URI,
			Content:  track.Content,
			Label:    track.Label,
			Language: track.Language,
			Type:     track.Format,
			Default:  track.Default,
		})
	}
	return &nativeplayer.PlaybackInfo{
		ID:                 info.ID,
		StreamType:         nativeplayer.StreamType(info.PlaybackType),
		StreamPath:         info.StreamPath,
		MimeType:           info.MimeType,
		StreamUrl:          info.StreamURL,
		ContentLength:      info.ContentLength,
		MkvMetadata:        info.MkvMetadata,
		SubtitleTracks:     tracks,
		EntryListData:      info.EntryListData,
		Episode:            info.Episode,
		Media:              info.Media,
		IsNakamaWatchParty: info.IsNakamaWatchParty,
		LocalFile:          info.LocalFile,
		MkvMetadataParser:  info.MkvMetadataParser,
	}
}

func toVideoSubtitleTrack(track *mediacore.SubtitleTrack) *VideoSubtitleTrack {
	if track == nil {
		return nil
	}
	return &VideoSubtitleTrack{
		Index:    track.Index,
		Src:      track.URI,
		Content:  track.Content,
		Label:    track.Label,
		Language: track.Language,
		Type:     track.Format,
		Default:  track.Default,
	}
}

func toSkipData(data *mediacore.SkipData) *SkipData {
	if data == nil {
		return nil
	}
	var op, ed *SkipDataEntry
	if data.Op != nil {
		op = &SkipDataEntry{
			Interval: SkipInterval{
				StartTime: data.Op.Interval.StartTime,
				EndTime:   data.Op.Interval.EndTime,
			},
		}
	}
	if data.Ed != nil {
		ed = &SkipDataEntry{
			Interval: SkipInterval{
				StartTime: data.Ed.Interval.StartTime,
				EndTime:   data.Ed.Interval.EndTime,
			},
		}
	}
	return &SkipData{Op: op, Ed: ed}
}

func toMediaCoreSkipData(data *SkipData) *mediacore.SkipData {
	if data == nil {
		return nil
	}
	var op, ed *mediacore.SkipDataEntry
	if data.Op != nil {
		op = &mediacore.SkipDataEntry{
			Interval: mediacore.SkipInterval{
				StartTime: data.Op.Interval.StartTime,
				EndTime:   data.Op.Interval.EndTime,
			},
		}
	}
	if data.Ed != nil {
		ed = &mediacore.SkipDataEntry{
			Interval: mediacore.SkipInterval{
				StartTime: data.Ed.Interval.StartTime,
				EndTime:   data.Ed.Interval.EndTime,
			},
		}
	}
	return &mediacore.SkipData{Op: op, Ed: ed}
}

func toMediaCorePlaylistState(state *VideoPlaylistState) *mediacore.PlaylistState {
	if state == nil {
		return nil
	}
	return &mediacore.PlaylistState{
		Type:            mediacore.PlaybackType(state.Type),
		Episodes:        state.Episodes,
		PreviousEpisode: state.PreviousEpisode,
		NextEpisode:     state.NextEpisode,
		CurrentEpisode:  state.CurrentEpisode,
		AnimeEntry:      state.AnimeEntry,
	}
}

func toMediaCoreOnlinestreamParams(params *OnlinestreamParams) *mediacore.OnlinestreamParams {
	if params == nil {
		return nil
	}
	return &mediacore.OnlinestreamParams{
		MediaId:       params.MediaId,
		EpisodeNumber: params.EpisodeNumber,
		Provider:      params.Provider,
		Server:        params.Server,
		Quality:       params.Quality,
		Dubbed:        params.Dubbed,
	}
}

func toOnlinestreamParams(params *mediacore.OnlinestreamParams) *OnlinestreamParams {
	if params == nil {
		return nil
	}
	return &OnlinestreamParams{
		MediaId:       params.MediaId,
		EpisodeNumber: params.EpisodeNumber,
		Provider:      params.Provider,
		Server:        params.Server,
		Quality:       params.Quality,
		Dubbed:        params.Dubbed,
	}
}
