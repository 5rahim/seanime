package mediacore

type Backend interface {
	Target() Target
	OpenAndAwait(clientID, state string)
	AbortOpen(clientID, reason string)
	Watch(clientID string, info *PlaybackInfo)
	Error(clientID string, err error)
	Execute(session SessionKey, cmd Command) error
	Terminate(session SessionKey)
	Events() <-chan Event
	Close() error

	PullStatus() (PlaybackStatus, bool)
	GetPlaylist() (*PlaylistState, bool)
	GetSkipData() (*SkipData, bool)
}
