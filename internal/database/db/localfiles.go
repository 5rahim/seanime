package db

import (
	"seanime/internal/database/models"

	"gorm.io/gorm/clause"
)

// TrimLocalFileEntries will trim the local file entries if there are more than 10 entries.
// This now uses the cleanup manager to avoid concurrent access issues.
func (db *Database) TrimLocalFileEntries() {
	// Use the cleanup manager to avoid concurrent access issues
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
