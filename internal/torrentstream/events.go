package torrentstream

const (
	eventTorrentLoading        = "torrentstream-torrent-loading"
	eventTorrentLoadingFailed  = "torrentstream-torrent-loading-failed"
	eventTorrentLoadingStatus  = "torrentstream-torrent-loading-status"
	eventTorrentLoaded         = "torrentstream-torrent-loaded"
	eventTorrentStartedPlaying = "torrentstream-torrent-started-playing"
	eventTorrentStatus         = "torrentstream-torrent-status"
	eventTorrentStopped        = "torrentstream-torrent-stopped"
)

type TorrentLoadingStatus struct {
	TorrentBeingChecked string                    `json:"torrentBeingChecked"`
	State               TorrentLoadingStatusState `json:"state"`
}

type TorrentLoadingStatusState string

const (
	TLSStateSearchingTorrents          TorrentLoadingStatusState = "SEARCHING_TORRENTS"
	TLSStateCheckingTorrent            TorrentLoadingStatusState = "CHECKING_TORRENT"
	TLSStateAddingTorrent              TorrentLoadingStatusState = "ADDING_TORRENT"
	TLSStateSelectingFile              TorrentLoadingStatusState = "SELECTING_FILE"
	TLSStateStartingServer             TorrentLoadingStatusState = "STARTING_SERVER"
	TLSStateSendingStreamToMediaPlayer TorrentLoadingStatusState = "SENDING_STREAM_TO_MEDIA_PLAYER"
)

func (r *Repository) sendTorrentLoadingStatus(event TorrentLoadingStatusState, checking string) {
	r.wsEventManager.SendEvent(eventTorrentLoadingStatus, &TorrentLoadingStatus{
		TorrentBeingChecked: checking,
		State:               event,
	})
}
