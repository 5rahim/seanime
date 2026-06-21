package mediacore

import "seanime/internal/player"

type Target = player.Target

const (
	TargetVideoCore = player.TargetVideoCore
	TargetMpvCore   = player.TargetMpvCore
)

type Renderer = player.Renderer

const (
	RendererWeb    = player.RendererWeb
	RendererNative = player.RendererNative
	RendererMpv    = player.RendererMpv
)

type PlaybackType = player.PlaybackType

const (
	PlaybackTypeLocalFile    = player.PlaybackTypeLocalFile
	PlaybackTypeTorrent      = player.PlaybackTypeTorrent
	PlaybackTypeDebrid       = player.PlaybackTypeDebrid
	PlaybackTypeNakama       = player.PlaybackTypeNakama
	PlaybackTypeOnlinestream = player.PlaybackTypeOnlinestream
	PlaybackTypeURL          = player.PlaybackTypeURL
)

type SessionKey = player.SessionKey
type SubtitleTrack = player.SubtitleTrack
type VideoSource = player.VideoSource
type InitialState = player.InitialState
type SkipInterval = player.SkipInterval
type SkipDataEntry = player.SkipDataEntry
type OnlinestreamParams = player.OnlinestreamParams
type SkipData = player.SkipData
type PlaybackInfo = player.PlaybackInfo
type PlaybackState = player.PlaybackState
type PlaybackStatus = player.PlaybackStatus
type PlaylistState = player.PlaylistState
type Event = player.Event
type BaseEvent = player.BaseEvent

type PlaybackLoadedEvent = player.PlaybackLoadedEvent
type LoadedMetadataEvent = player.LoadedMetadataEvent
type CanPlayEvent = player.CanPlayEvent
type PausedEvent = player.PausedEvent
type ResumedEvent = player.ResumedEvent
type StatusEvent = player.StatusEvent
type SeekedEvent = player.SeekedEvent
type CompletedEvent = player.CompletedEvent
type EndedEvent = player.EndedEvent
type ErrorEvent = player.ErrorEvent
type TerminatedEvent = player.TerminatedEvent
type FullscreenChangedEvent = player.FullscreenChangedEvent
type PipChangedEvent = player.PipChangedEvent
type AudioTrackChangedEvent = player.AudioTrackChangedEvent
type SubtitleTrackChangedEvent = player.SubtitleTrackChangedEvent
type SubtitleFileUploadedEvent = player.SubtitleFileUploadedEvent
type PlaylistStateEvent = player.PlaylistStateEvent
type SkipDataEvent = player.SkipDataEvent

type CommandType = player.CommandType

const (
	CommandPause                       = player.CommandPause
	CommandResume                      = player.CommandResume
	CommandSeek                        = player.CommandSeek
	CommandSeekTo                      = player.CommandSeekTo
	CommandSetFullscreen               = player.CommandSetFullscreen
	CommandSetPip                      = player.CommandSetPip
	CommandSetAudioTrack               = player.CommandSetAudioTrack
	CommandSetSubtitleTrack            = player.CommandSetSubtitleTrack
	CommandAddSubtitleTrack            = player.CommandAddSubtitleTrack
	CommandAddExternalSubtitleTrack    = player.CommandAddExternalSubtitleTrack
	CommandSetMediaCaptionTrack        = player.CommandSetMediaCaptionTrack
	CommandAddMediaCaptionTrack        = player.CommandAddMediaCaptionTrack
	CommandPlayPlaylistEpisode         = player.CommandPlayPlaylistEpisode
	CommandShowMessage                 = player.CommandShowMessage
	CommandSetSkipData                 = player.CommandSetSkipData
	CommandClearSkipData               = player.CommandClearSkipData
	CommandStartOnlinestreamWatchParty = player.CommandStartOnlinestreamWatchParty
)

type ShowMessagePayload = player.ShowMessagePayload
type Command = player.Command
