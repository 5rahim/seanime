package torrentstream

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

type BatchHistoryResponse struct {
	Torrent *hibiketorrent.AnimeTorrent `json:"torrent"`
}

func (r *Repository) GetBatchHistory(mId int) *BatchHistoryResponse {
	torrent, found := r.selectionHistoryMap.Get(mId)
	if !found {
		return &BatchHistoryResponse{}
	}

	return &BatchHistoryResponse{
		torrent,
	}
}
