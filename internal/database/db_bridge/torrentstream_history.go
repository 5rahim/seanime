package db_bridge

import (
	"github.com/goccy/go-json"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
)

func GetTorrentstreamHistory(db *db.Database, mId int) (*hibiketorrent.AnimeTorrent, error) {
	var history models.TorrentstreamHistory
	if err := db.Gorm().Where("media_id = ?", mId).First(&history).Error; err != nil {
		return nil, err
	}

	var torrent hibiketorrent.AnimeTorrent
	if err := json.Unmarshal(history.Torrent, &torrent); err != nil {
		return nil, err
	}
	return &torrent, nil
}

func InsertTorrentstreamHistory(db *db.Database, mId int, torrent *hibiketorrent.AnimeTorrent) error {
	if torrent == nil {
		return nil
	}

	// Marshal the data
	bytes, err := json.Marshal(torrent)
	if err != nil {
		return err
	}

	// Get current history
	var history models.TorrentstreamHistory
	if err := db.Gorm().Where("media_id = ?", mId).First(&history).Error; err == nil {
		// Update the history
		history.Torrent = bytes
		return db.Gorm().Save(&history).Error
	}

	return db.Gorm().Create(&models.TorrentstreamHistory{
		MediaId: mId,
		Torrent: bytes,
	}).Error
}
