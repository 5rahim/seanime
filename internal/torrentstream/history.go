package torrentstream

import (
	"seanime/internal/database/db_bridge"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/util"

	"github.com/5rahim/habari"
)

type BatchHistoryResponse struct {
	Torrent  *hibiketorrent.AnimeTorrent `json:"torrent"`
	Metadata *habari.Metadata            `json:"metadata"`
}

func (r *Repository) GetBatchHistory(mId int) (ret *BatchHistoryResponse) {
	defer util.HandlePanicInModuleThen("torrentstream/GetBatchHistory", func() {
		ret = &BatchHistoryResponse{}
	})

	torrent, err := db_bridge.GetTorrentstreamHistory(r.db, mId)
	if err != nil {
		return &BatchHistoryResponse{}
	}

	metadata := habari.Parse(torrent.Name)

	return &BatchHistoryResponse{
		torrent,
		metadata,
	}
}

func (r *Repository) AddBatchHistory(mId int, torrent *hibiketorrent.AnimeTorrent) {
	go func() {
		defer util.HandlePanicInModuleThen("torrentstream/AddBatchHistory", func() {})

		_ = db_bridge.InsertTorrentstreamHistory(r.db, mId, torrent)
	}()
}
