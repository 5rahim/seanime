package db

import (
	"seanime/internal/database/models"
)

func (db *Database) UpsertNakamaSettings(nakamaSettings *models.NakamaSettings) (*models.NakamaSettings, error) {

	// Get current settings
	currentSettings, err := db.GetSettings()
	if err != nil {
		return nil, err
	}

	// Update the settings
	*(currentSettings.Nakama) = *nakamaSettings

	_, err = db.UpsertSettings(currentSettings)
	if err != nil {
		return nil, err
	}

	return nakamaSettings, nil
}
