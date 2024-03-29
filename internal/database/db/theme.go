package db

import (
	"github.com/seanime-app/seanime/internal/database/models"
	"gorm.io/gorm/clause"
)

func (db *Database) GetTheme() (*models.Theme, error) {
	var theme models.Theme
	err := db.gormdb.Where("id = ?", 1).Find(&theme).Error

	if err != nil {
		return nil, err
	}
	return &theme, nil
}

// UpsertTheme updates the theme settings.
func (db *Database) UpsertTheme(settings *models.Theme) (*models.Theme, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to save theme in the database")
		return nil, err
	}

	db.logger.Debug().Msg("db: Theme saved")
	return settings, nil

}
