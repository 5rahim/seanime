package db

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func UpsertToken(db *Database, token *models.Token, logger *zerolog.Logger) (*models.Token, error) {

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(token).Error

	if err != nil {
		logger.Error().Err(err).Msg("Failed to save token in the database")
		return nil, err
	}
	return token, nil

}

func GetToken(db *Database, logger *zerolog.Logger) string {
	var token models.Token
	err := db.Where("id = ?", 1).First(&token).Error
	if err != nil {
		logger.Error().Err(err).Msg("failed to get token")
		return ""
	}
	return token.Value
}
