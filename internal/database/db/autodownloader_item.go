package db

import (
	"seanime/internal/database/models"
)

func (db *Database) GetAutoDownloaderItems() ([]*models.AutoDownloaderItem, error) {
	var res []*models.AutoDownloaderItem
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) GetAutoDownloaderItem(id uint) (*models.AutoDownloaderItem, error) {
	var res models.AutoDownloaderItem
	err := db.gormdb.First(&res, id).Error
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *Database) GetAutoDownloaderItemByMediaId(mId int) ([]*models.AutoDownloaderItem, error) {
	var res []*models.AutoDownloaderItem
	err := db.gormdb.Where("media_id = ?", mId).Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) GetDelayedAutoDownloaderItems() ([]*models.AutoDownloaderItem, error) {
	var res []*models.AutoDownloaderItem
	err := db.gormdb.Where("is_delayed = ?", true).Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *Database) InsertAutoDownloaderItem(item *models.AutoDownloaderItem) error {
	err := db.gormdb.Create(item).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteAutoDownloaderItem(id uint) error {
	return db.gormdb.Delete(&models.AutoDownloaderItem{}, id).Error
}

// DeleteDownloadedAutoDownloaderItems will delete all the downloaded queued items from the database.
func (db *Database) DeleteDownloadedAutoDownloaderItems() error {
	return db.gormdb.Where("downloaded = ?", true).Delete(&models.AutoDownloaderItem{}).Error
}

func (db *Database) UpdateAutoDownloaderItem(id uint, item *models.AutoDownloaderItem) error {
	// Save the data
	return db.gormdb.Save(item).Error
}
