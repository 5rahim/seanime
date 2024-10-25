package db

import (
	"seanime/internal/database/models"
)

func (db *Database) GetDebridTorrentItems() ([]*models.DebridTorrentItem, error) {
	var res []*models.DebridTorrentItem
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) GetDebridTorrentItemByDbId(dbId uint) (*models.DebridTorrentItem, error) {
	var res models.DebridTorrentItem
	err := db.gormdb.First(&res, dbId).Error
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *Database) GetDebridTorrentItemByTorrentItemId(tId string) (*models.DebridTorrentItem, error) {
	var res *models.DebridTorrentItem
	err := db.gormdb.Where("torrent_item_id = ?", tId).First(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) InsertDebridTorrentItem(item *models.DebridTorrentItem) error {
	err := db.gormdb.Create(item).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteDebridTorrentItemByDbId(dbId uint) error {
	return db.gormdb.Delete(&models.DebridTorrentItem{}, dbId).Error
}

func (db *Database) DeleteDebridTorrentItemByTorrentItemId(tId string) error {
	return db.gormdb.Where("torrent_item_id = ?", tId).Delete(&models.DebridTorrentItem{}).Error
}

func (db *Database) UpdateDebridTorrentItemByDbId(dbId uint, item *models.DebridTorrentItem) error {
	// Save the data
	return db.gormdb.Model(&models.DebridTorrentItem{}).Where("id = ?", dbId).Updates(item).Error
}
