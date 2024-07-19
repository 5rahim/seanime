package db

import (
	"gorm.io/gorm/clause"
	"seanime/internal/database/models"
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
		mIds[i] = int(v.ID)
	}

	return mIds, nil
}

func (db *Database) GetSilencedMediaEntry(mId uint) (*models.SilencedMediaEntry, error) {
	var res models.SilencedMediaEntry
	err := db.gormdb.First(&res, mId).Error
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *Database) InsertSilencedMediaEntry(mId uint) error {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&models.SilencedMediaEntry{
		BaseModel: models.BaseModel{
			ID: mId,
		},
	}).Error
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
