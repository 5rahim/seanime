package torrentstream

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/internal/database/db_bridge"
	"seanime/internal/util"
)

type BatchHistoryResponse struct {
	Torrent *hibiketorrent.AnimeTorrent `json:"torrent"`
}

func (r *Repository) GetBatchHistory(mId int) *BatchHistoryResponse {
	torrent, err := db_bridge.GetTorrentstreamHistory(r.db, mId)
	if err != nil {
		return &BatchHistoryResponse{}
	}

	return &BatchHistoryResponse{
		torrent,
	}
}

func (r *Repository) addBatchHistory(mId int, torrent *hibiketorrent.AnimeTorrent) {
	go func() {
		defer util.HandlePanicInModuleThen("torrentstream/addBatchHistory", func() {})

		_ = db_bridge.InsertTorrentstreamHistory(r.db, mId, torrent)
	}()
}
