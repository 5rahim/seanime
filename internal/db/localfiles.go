package db

// Should not import from scanner/localfile
import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func UpsertLocalFiles(db *Database, token *models.LocalFiles, logger *zerolog.Logger) (*models.LocalFiles, error) {

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(token).Error

	if err != nil {
		logger.Error().Err(err).Msg("Failed to save local files in the database")
		return nil, err
	}
	return token, nil
}
