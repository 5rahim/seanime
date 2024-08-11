package db

import (
	"gorm.io/gorm/clause"
	"seanime/internal/database/models"
)

func (db *Database) UpsertToken(token *models.Token) (*models.Token, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(token).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("Failed to save token in the database")
		return nil, err
	}
	return token, nil

}
