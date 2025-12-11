package nativeplayer

import (
	"seanime/internal/mkvparser"
)

type ServerEvent string

const (
	ServerEventOpenAndAwait     ServerEvent = "open-and-await"
	ServerEventAbortOpen        ServerEvent = "abort-open"
	ServerEventWatch            ServerEvent = "watch"
	ServerEventSubtitleEvent    ServerEvent = "subtitle-event"
	ServerEventSetTracks        ServerEvent = "set-tracks"
	ServerEventError            ServerEvent = "error"
	ServerEventAddSubtitleTrack ServerEvent = "add-subtitle-track"
	ServerEventTerminate        ServerEvent = "terminate"
)

// OpenAndAwait opens the player and waits for the client to send the watch event.
func (p *NativePlayer) OpenAndAwait(clientId string, loadingState string) {
	p.sendPlayerEventTo(clientId, string(ServerEventOpenAndAwait), loadingState)
}

// AbortOpen closes the player
func (p *NativePlayer) AbortOpen(clientId string, reason string) {
	p.sendPlayerEventTo(clientId, string(ServerEventAbortOpen), reason)
}

// Watch sends the watch event to the client.
func (p *NativePlayer) Watch(clientId string, playbackInfo *PlaybackInfo) {
	// Store the playback info
	p.sendPlayerEventTo(clientId, string(ServerEventWatch), playbackInfo, true)
}

// SubtitleEvent sends the subtitle event to the client.
func (p *NativePlayer) SubtitleEvent(clientId string, event *mkvparser.SubtitleEvent) {
	p.sendPlayerEventTo(clientId, string(ServerEventSubtitleEvent), event, true)
}

// SetTracks sends the set tracks event to the client.
func (p *NativePlayer) SetTracks(clientId string, tracks []*mkvparser.TrackInfo) {
	p.sendPlayerEventTo(clientId, string(ServerEventSetTracks), tracks)
}

// Error stops the playback and displays an error message.
func (p *NativePlayer) Error(clientId string, err error) {
	p.sendPlayerEventTo(clientId, string(ServerEventError), struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

// AddSubtitleTrack sends the subtitle track added event to the client.
func (p *NativePlayer) AddSubtitleTrack(clientId string, track *mkvparser.TrackInfo) {
	p.sendPlayerEventTo(clientId, string(ServerEventAddSubtitleTrack), track)
}

// Stop emits a VideoTerminatedEvent to all subscribers.
// It should only be called by a module.
func (p *NativePlayer) Stop() {
	if p.playbackInfo == nil {
		return
	}
	p.logger.Debug().Msg("nativeplayer: Stopping playback, notifying subscribers")
	p.sendPlayerEvent(string(ServerEventTerminate), nil)
}
