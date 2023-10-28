package db

// Should not import from scanner/localfile
import (
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertLocalFiles(token *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(token).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return token, nil
}

func (db *Database) InsertLocalFiles(token *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Create(token).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return token, nil
}

func (db *Database) GetLatestLocalFiles(token *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Last(token).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return token, nil
}
