package db

import (
	"seanime/internal/database/models"

	"gorm.io/gorm/clause"
)

// TrimLocalFileEntries will trim the local file entries if there are more than 10 entries.
func (db *Database) TrimLocalFileEntries() {
	db.cleanupManager.trimLocalFileEntries()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) UpsertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(lfs).Error

	if err != nil {
		return nil, err
	}
	return lfs, nil
}

func (db *Database) InsertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Create(lfs).Error

	if err != nil {
		return nil, err
	}
	return lfs, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) UpsertShelvedLocalFiles(lfs *models.ShelvedLocalFiles) (*models.ShelvedLocalFiles, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(lfs).Error

	if err != nil {
		return nil, err
	}
	return lfs, nil
}
