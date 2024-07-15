package db

import (
	"github.com/seanime-app/seanime/internal/database/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertSettings(settings *models.Settings) (*models.Settings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save settings in the database")
		return nil, err
	}

	db.Logger.Debug().Msg("db: Settings saved")
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) GetLibraryPathFromSettings() (string, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.Library.LibraryPath, nil
}

func (db *Database) AutoUpdateProgressIsEnabled() (bool, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return false, err
	}
	return settings.Library.AutoUpdateProgress, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) UpsertMediastreamSettings(settings *models.MediastreamSettings) (*models.MediastreamSettings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save media streaming settings in the database")
		return nil, err
	}

	db.Logger.Debug().Msg("db: Media streaming settings saved")
	return settings, nil

}

func (db *Database) GetMediastreamSettings() (*models.MediastreamSettings, bool) {
	var settings models.MediastreamSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error

	if err != nil {
		return nil, false
	}
	return &settings, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) UpsertTorrentstreamSettings(settings *models.TorrentstreamSettings) (*models.TorrentstreamSettings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save torrent streaming settings in the database")
		return nil, err
	}

	db.Logger.Debug().Msg("db: Torrent streaming settings saved")
	return settings, nil
}

func (db *Database) GetTorrentstreamSettings() (*models.TorrentstreamSettings, bool) {
	var settings models.TorrentstreamSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error

	if err != nil {
		return nil, false
	}
	return &settings, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
