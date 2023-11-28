package db

import (
	"github.com/seanime-app/seanime/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertSettings(settings *models.Settings) (*models.Settings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to save settings in the database")
		return nil, err
	}

	db.logger.Debug().Msg("db: Settings saved")
	return settings, nil

}

func (db *Database) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	err := db.gormdb.Where("id = ?", 1).Find(&settings).Error

	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (db *Database) GetLibraryPath() (string, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.Library.LibraryPath, nil
}
