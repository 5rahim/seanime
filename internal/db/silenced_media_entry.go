package db

import (
	"github.com/seanime-app/seanime/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) GetSilencedMediaEntries() ([]*models.SilencedMediaEntry, error) {
	var res []*models.SilencedMediaEntry
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetSilencedMediaEntryIds returns the ids of all silenced media entries.
// It returns an empty slice if there is an error.
func (db *Database) GetSilencedMediaEntryIds() ([]int, error) {
	var res []*models.SilencedMediaEntry
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return make([]int, 0), err
	}

	if len(res) == 0 {
		return make([]int, 0), nil
	}

	mIds := make([]int, len(res))
	for i, v := range res {
		mIds[i] = v.MediaId
	}

	return mIds, nil
}

func (db *Database) GetSilencedMediaEntry(id uint) (*models.SilencedMediaEntry, error) {
	var res models.SilencedMediaEntry
	err := db.gormdb.First(&res, id).Error
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *Database) GetSilencedMediaEntryByMediaId(mId uint) (*models.SilencedMediaEntry, error) {
	var res models.SilencedMediaEntry
	err := db.gormdb.Where("media_id = ?", mId).First(&res).Error
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *Database) InsertSilencedMediaEntry(entry *models.SilencedMediaEntry) error {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "media_id"}},
		UpdateAll: true,
	}).Create(entry).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteSilencedMediaEntry(id uint) error {
	err := db.gormdb.Delete(&models.SilencedMediaEntry{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteSilencedMediaEntryByMediaId(mId uint) error {
	err := db.gormdb.Where("media_id = ?", mId).Delete(&models.SilencedMediaEntry{}).Error
	if err != nil {
		return err
	}
	return nil
}
