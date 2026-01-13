package db

import (
	"seanime/internal/database/models"
)

func (db *Database) GetMediaMetadataParent(mId int) (*models.MediaMetadataParent, error) {
	var ret models.MediaMetadataParent
	if err := db.Gorm().Where("media_id = ?", mId).First(&ret).Error; err != nil {
		return nil, err
	}
	return &ret, nil
}

func (db *Database) InsertMediaMetadataParent(m models.MediaMetadataParent) (*models.MediaMetadataParent, error) {
	_ = db.DeleteMediaMetadataParent(m.MediaId)
	err := db.gormdb.Save(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (db *Database) DeleteMediaMetadataParent(mId int) error {
	err := db.gormdb.Where("media_id = ?", mId).Delete(&models.MediaMetadataParent{}).Error
	if err != nil {
		return err
	}
	return nil
}
