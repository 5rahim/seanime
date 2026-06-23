package videocore

import (
	"errors"
	"fmt"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/player"

	"github.com/google/uuid"
)

type Adapter struct {
	vc       *VideoCore
	np       *nativeplayer.NativePlayer
	sub      *Subscriber
	eventsCh chan player.Event
}

var _ mediacore.Backend = (*Adapter)(nil)

func NewAdapter(vc *VideoCore, np *nativeplayer.NativePlayer) *Adapter {
	id := "videocore:adapter:" + uuid.NewString()
	sub := vc.Subscribe(id)

	a := &Adapter{
		vc:       vc,
		np:       np,
		sub:      sub,
		eventsCh: make(chan player.Event, 100),
	}

	a.startEventLoop()
	return a
}

func (a *Adapter) Target() player.Target {
	return player.TargetVideoCore
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

func (a *Adapter) Watch(clientID string, info *player.PlaybackInfo) {
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

func (a *Adapter) Execute(session player.SessionKey, cmd player.Command) error {
	switch cmd.Type {
	case player.CommandPause:
		a.vc.Pause()
	case player.CommandResume:
		a.vc.Resume()
	case player.CommandSeek:
		if sec, ok := cmd.Payload.(float64); ok {
			a.vc.Seek(sec)
		} else {
			return errors.New("invalid payload type for Seek")
		}
	case player.CommandSeekTo:
		if sec, ok := cmd.Payload.(float64); ok {
			a.vc.SeekTo(sec)
		} else {
			return errors.New("invalid payload type for SeekTo")
		}
	case player.CommandSetFullscreen:
		if val, ok := cmd.Payload.(bool); ok {
			a.vc.SetFullscreen(val)
		} else {
			return errors.New("invalid payload type for SetFullscreen")
		}
	case player.CommandSetPip:
		if val, ok := cmd.Payload.(bool); ok {
			a.vc.SetPip(val)
		} else {
			return errors.New("invalid payload type for SetPip")
		}
	case player.CommandSetAudioTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetAudioTrack(val)
		} else {
			return errors.New("invalid payload type for SetAudioTrack")
		}
	case player.CommandSetSubtitleTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for SetSubtitleTrack")
		}
	case player.CommandAddSubtitleTrack:
		if val, ok := cmd.Payload.(*mkvparser.TrackInfo); ok {
			a.vc.AddSubtitleTrack(val)
		} else {
			return errors.New("invalid payload type for AddSubtitleTrack")
		}
	case player.CommandAddExternalSubtitleTrack:
		if val, ok := cmd.Payload.(*player.SubtitleTrack); ok {
			a.vc.AddExternalSubtitleTrack(toVideoSubtitleTrack(val))
		} else {
			return errors.New("invalid payload type for AddExternalSubtitleTrack")
		}
	case player.CommandSetMediaCaptionTrack:
		if val, ok := cmd.Payload.(int); ok {
			a.vc.SetMediaCaptionTrack(val)
		} else {
			return errors.New("invalid payload type for SetMediaCaptionTrack")
		}
	case player.CommandAddMediaCaptionTrack:
		a.vc.AddMediaCaptionTrack(cmd.Payload)
	case player.CommandPlayPlaylistEpisode:
		if val, ok := cmd.Payload.(string); ok {
			a.vc.PlayPlaylistEpisode(val)
		} else {
			return errors.New("invalid payload type for PlayPlaylistEpisode")
		}
	case player.CommandShowMessage:
		if val, ok := cmd.Payload.(player.ShowMessagePayload); ok {
			a.vc.ShowMessage(val.Message, val.Duration)
		} else {
			return errors.New("invalid payload type for ShowMessage")
		}
	case player.CommandSetSkipData:
		if val, ok := cmd.Payload.(*player.SkipData); ok {
			a.vc.SetSkipData(toSkipData(val))
		} else {
			return errors.New("invalid payload type for SetSkipData")
		}
	case player.CommandClearSkipData:
		a.vc.ClearSkipData()
	case player.CommandStartOnlinestreamWatchParty:
		if val, ok := cmd.Payload.(*player.OnlinestreamParams); ok {
			a.vc.StartOnlinestreamWatchParty(toOnlinestreamParams(val))
		} else {
			return errors.New("invalid payload type for StartOnlinestreamWatchParty")
		}
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Type)
	}
	return nil
}

func (a *Adapter) Terminate(session player.SessionKey) {
	a.vc.Terminate()
}

func (a *Adapter) Events() <-chan player.Event {
	return a.eventsCh
}

func (a *Adapter) Close() error {
	a.vc.Unsubscribe(a.sub.GetId())
	return nil
}

func (a *Adapter) PullStatus() (player.PlaybackStatus, bool) {
	status, ok := a.vc.PullStatus()
	if !ok {
		return player.PlaybackStatus{}, false
	}
	return player.PlaybackStatus{
		ID:          status.PlaybackId,
		ClientID:    status.ClientId,
		Paused:      status.Paused,
		CurrentTime: status.CurrentTime,
		Duration:    status.Duration,
	}, true
}

func (a *Adapter) GetPlaylist() (*player.PlaylistState, bool) {
	playlist, ok := a.vc.GetPlaylist()
	if !ok || playlist == nil {
		return nil, false
	}
	return toMediaCorePlaylistState(playlist), true
}

func (a *Adapter) GetSkipData() (*player.SkipData, bool) {
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

func (a *Adapter) toEvent(ev VideoEvent) player.Event {
	session := player.SessionKey{
		Target:     player.TargetVideoCore,
		ClientID:   ev.GetClientId(),
		PlaybackID: ev.GetPlaybackId(),
	}

	base := player.BaseEvent{Session: session}

	switch e := ev.(type) {
	case *VideoLoadedEvent:
		return &player.PlaybackLoadedEvent{
			BaseEvent: base,
			State: player.PlaybackState{
				ClientID:     e.ClientId,
				PlaybackInfo: toMediaCorePlaybackInfo(e.State.PlaybackInfo, e.PlayerType),
			},
		}
	case *VideoPlaybackStateEvent:
		return &player.PlaybackLoadedEvent{
			BaseEvent: base,
			State: player.PlaybackState{
				ClientID:     e.ClientId,
				PlaybackInfo: toMediaCorePlaybackInfo(e.State.PlaybackInfo, e.PlayerType),
			},
		}
	case *VideoLoadedMetadataEvent:
		return &player.LoadedMetadataEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoCanPlayEvent:
		return &player.CanPlayEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoPausedEvent:
		return &player.PausedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoResumedEvent:
		return &player.ResumedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoStatusEvent:
		return &player.StatusEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoSeekedEvent:
		return &player.SeekedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
			Paused:      e.Paused,
		}
	case *VideoCompletedEvent:
		return &player.CompletedEvent{
			BaseEvent:   base,
			CurrentTime: e.CurrentTime,
			Duration:    e.Duration,
		}
	case *VideoEndedEvent:
		return &player.EndedEvent{
			BaseEvent: base,
			AutoNext:  e.AutoNext,
		}
	case *VideoErrorEvent:
		return &player.ErrorEvent{
			BaseEvent: base,
			Error:     e.Error,
		}
	case *VideoTerminatedEvent:
		return &player.TerminatedEvent{
			BaseEvent: base,
		}
	case *VideoFullscreenEvent:
		return &player.FullscreenChangedEvent{
			BaseEvent:  base,
			Fullscreen: e.Fullscreen,
		}
	case *VideoPipEvent:
		return &player.PipChangedEvent{
			BaseEvent: base,
			Pip:       e.Pip,
		}
	case *VideoAudioTrackEvent:
		return &player.AudioTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackNumber,
		}
	case *VideoSubtitleTrackEvent:
		return &player.SubtitleTrackChangedEvent{
			BaseEvent: base,
			TrackID:   e.TrackNumber,
		}
	case *SubtitleFileUploadedEvent:
		return &player.SubtitleFileUploadedEvent{
			BaseEvent: base,
			Filename:  e.Filename,
			Content:   e.Content,
		}
	case *VideoPlaylistEvent:
		return &player.PlaylistStateEvent{
			BaseEvent: base,
			Playlist:  toMediaCorePlaylistState(e.Playlist),
		}
	case *VideoSkipDataEvent:
		return &player.SkipDataEvent{
			BaseEvent: base,
			SkipData:  toMediaCoreSkipData(e.SkipData),
		}
	}
	return nil
}

func toMediaCorePlaybackInfo(info *VideoPlaybackInfo, playerType PlayerType) *player.PlaybackInfo {
	if info == nil {
		return nil
	}

	var renderer player.Renderer
	if playerType == WebPlayer {
		renderer = player.RendererWeb
	} else {
		renderer = player.RendererNative
	}

	tracks := make([]*player.SubtitleTrack, 0, len(info.SubtitleTracks))
	for _, track := range info.SubtitleTracks {
		if track == nil {
			continue
		}
		tracks = append(tracks, &player.SubtitleTrack{
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

	sources := make([]*player.VideoSource, 0, len(info.VideoSources))
	for _, src := range info.VideoSources {
		if src == nil {
			continue
		}
		sources = append(sources, &player.VideoSource{
			Index:      src.Index,
			Resolution: src.Resolution,
			URL:        src.URL,
			Label:      src.Label,
			MoreInfo:   src.MoreInfo,
		})
	}

	var init *player.InitialState
	if info.InitialState != nil {
		init = &player.InitialState{
			CurrentTime: info.InitialState.CurrentTime,
			Paused:      info.InitialState.Paused,
		}
	}

	return &player.PlaybackInfo{
		ID:                             info.Id,
		Target:                         player.TargetVideoCore,
		Renderer:                       renderer,
		PlaybackType:                   player.PlaybackType(info.PlaybackType),
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

func toNativePlaybackInfo(info *player.PlaybackInfo) *nativeplayer.PlaybackInfo {
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

func toVideoSubtitleTrack(track *player.SubtitleTrack) *VideoSubtitleTrack {
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

func toSkipData(data *player.SkipData) *SkipData {
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

func toMediaCoreSkipData(data *SkipData) *player.SkipData {
	if data == nil {
		return nil
	}
	var op, ed *player.SkipDataEntry
	if data.Op != nil {
		op = &player.SkipDataEntry{
			Interval: player.SkipInterval{
				StartTime: data.Op.Interval.StartTime,
				EndTime:   data.Op.Interval.EndTime,
			},
		}
	}
	if data.Ed != nil {
		ed = &player.SkipDataEntry{
			Interval: player.SkipInterval{
				StartTime: data.Ed.Interval.StartTime,
				EndTime:   data.Ed.Interval.EndTime,
			},
		}
	}
	return &player.SkipData{Op: op, Ed: ed}
}

func toMediaCorePlaylistState(state *VideoPlaylistState) *player.PlaylistState {
	if state == nil {
		return nil
	}
	return &player.PlaylistState{
		Type:            player.PlaybackType(state.Type),
		Episodes:        state.Episodes,
		PreviousEpisode: state.PreviousEpisode,
		NextEpisode:     state.NextEpisode,
		CurrentEpisode:  state.CurrentEpisode,
		AnimeEntry:      state.AnimeEntry,
	}
}

func toOnlinestreamParams(params *player.OnlinestreamParams) *OnlinestreamParams {
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

func toMediaCoreOnlinestreamParams(params *OnlinestreamParams) *player.OnlinestreamParams {
	if params == nil {
		return nil
	}
	return &player.OnlinestreamParams{
		MediaId:       params.MediaId,
		EpisodeNumber: params.EpisodeNumber,
		Provider:      params.Provider,
		Server:        params.Server,
		Quality:       params.Quality,
		Dubbed:        params.Dubbed,
	}
}
