package db

import (
	"seanime/internal/database/models"

	"gorm.io/gorm/clause"
)

// TrimLocalFileEntries will trim the local file entries if there are more than 10 entries.
// This is run in a goroutine.
func (db *Database) TrimLocalFileEntries() {
	go func() {
		var count int64
		err := db.gormdb.Model(&models.LocalFiles{}).Count(&count).Error
		if err != nil {
			db.Logger.Error().Err(err).Msg("database: Failed to count local file entries")
			return
		}
		if count > 10 {
			// Leave 5 entries
			err = db.gormdb.Delete(&models.LocalFiles{}, "id IN (SELECT id FROM local_files ORDER BY id ASC LIMIT ?)", count-5).Error
			if err != nil {
				db.Logger.Error().Err(err).Msg("database: Failed to delete old local file entries")
				return
			}
		}
	}()
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
