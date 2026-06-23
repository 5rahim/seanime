package mediacore

import "seanime/internal/player"

type Backend interface {
	Target() player.Target
	OpenAndAwait(clientID, state string)
	AbortOpen(clientID, reason string)
	Watch(clientID string, info *player.PlaybackInfo)
	Error(clientID string, err error)
	Execute(session player.SessionKey, cmd player.Command) error
	Terminate(session player.SessionKey)
	Events() <-chan player.Event
	Close() error

	PullStatus() (player.PlaybackStatus, bool)
	GetPlaylist() (*player.PlaylistState, bool)
	GetSkipData() (*player.SkipData, bool)
}
