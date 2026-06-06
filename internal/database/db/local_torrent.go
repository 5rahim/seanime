package db

import (
	"seanime/internal/database/models"

	"gorm.io/gorm/clause"
)

func (db *Database) GetLocalTorrents() ([]*models.LocalTorrent, error) {
	var torrents []*models.LocalTorrent
	err := db.gormdb.Order("queue_index asc, created_at asc").Find(&torrents).Error
	return torrents, err
}

func (db *Database) UpsertLocalTorrent(torrent *models.LocalTorrent) error {
	return db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hash"}},
		UpdateAll: true,
	}).Create(torrent).Error
}

func (db *Database) UpdateLocalTorrent(hash string, values map[string]interface{}) error {
	return db.gormdb.Model(&models.LocalTorrent{}).Where("hash = ?", hash).Updates(values).Error
}

func (db *Database) DeleteLocalTorrent(hash string) error {
	return db.gormdb.Where("hash = ?", hash).Delete(&models.LocalTorrent{}).Error
}
