package torrentstream

import "seanime/internal/events"

const (
	eventLoading               = "loading"
	eventLoadingFailed         = "loading-failed"
	eventTorrentLoaded         = "loaded"
	eventTorrentStartedPlaying = "started-playing"
	eventTorrentStatus         = "status"
	eventTorrentStopped        = "stopped"
)

type TorrentLoadingStatusState string

const (
	TLSStateLoading                    TorrentLoadingStatusState = "LOADING"
	TLSStateSearchingTorrents          TorrentLoadingStatusState = "SEARCHING_TORRENTS"
	TLSStateCheckingTorrent            TorrentLoadingStatusState = "CHECKING_TORRENT"
	TLSStateAddingTorrent              TorrentLoadingStatusState = "ADDING_TORRENT"
	TLSStateSelectingFile              TorrentLoadingStatusState = "SELECTING_FILE"
	TLSStateStartingServer             TorrentLoadingStatusState = "STARTING_SERVER"
	TLSStateSendingStreamToMediaPlayer TorrentLoadingStatusState = "SENDING_STREAM_TO_MEDIA_PLAYER"
)

type TorrentStreamState struct {
	State string `json:"state"`
}

func (r *Repository) sendStateEvent(event string, data ...interface{}) {
	var dataToSend interface{}

	if len(data) > 0 {
		dataToSend = data[0]
	}
	r.wsEventManager.SendEvent(events.TorrentStreamState, struct {
		State string      `json:"state"`
		Data  interface{} `json:"data"`
	}{
		State: event,
		Data:  dataToSend,
	})
}

//func (r *Repository) sendTorrentLoadingStatus(event TorrentLoadingStatusState, checking string) {
//	r.wsEventManager.SendEvent(eventTorrentLoadingStatus, &TorrentLoadingStatus{
//		TorrentBeingChecked: checking,
//		State:               event,
//	})
//}
