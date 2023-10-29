package db

// Should not import from scanner/localfile
import (
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(lfs).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return lfs, nil
}

func (db *Database) InsertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Create(lfs).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return lfs, nil
}

func (db *Database) GetLatestLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Last(lfs).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return lfs, nil
}
