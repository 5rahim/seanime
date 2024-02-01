package db

import (
	"github.com/seanime-app/seanime/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) GetMalInfo() (*models.Mal, error) {
	// Get the first entry
	var res models.Mal
	err := db.gormdb.First(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (db *Database) UpsertMalInfo(info *models.Mal) (*models.Mal, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(info).Error

	if err != nil {
		return nil, err
	}
	return info, nil
}

func (db *Database) InsertMalInfo(info *models.Mal) (*models.Mal, error) {
	err := db.gormdb.Create(info).Error

	if err != nil {
		return nil, err
	}
	return info, nil
}

func (db *Database) DeleteMalInfo() error {
	err := db.gormdb.Delete(&models.Mal{}, 1).Error

	if err != nil {
		return err
	}
	return nil
}
