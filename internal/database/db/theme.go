package db

import (
	"seanime/internal/database/models"

	"github.com/goccy/go-json"
	"gorm.io/gorm/clause"
)

var themeCache *models.Theme

func (db *Database) GetTheme() (*models.Theme, error) {

	if themeCache != nil {
		return themeCache, nil
	}

	var theme models.Theme
	err := db.gormdb.Where("id = ?", 1).Find(&theme).Error

	if err != nil {
		return nil, err
	}

	themeCache = &theme

	return &theme, nil
}

var themeCopyCache *models.Theme

// GetThemeCopy returns a copy of the theme settings.
// The copy will have the HomeItems removed.
func (db *Database) GetThemeCopy() (*models.Theme, error) {

	if themeCopyCache != nil {
		return themeCopyCache, nil
	}

	theme, err := db.GetTheme()
	if err != nil {
		return nil, err
	}

	marshaledTheme, err := json.Marshal(theme)
	if err != nil {
		return nil, err
	}

	var themeCopy models.Theme
	err = json.Unmarshal(marshaledTheme, &themeCopy)
	if err != nil {
		return nil, err
	}

	themeCopyCache = &themeCopy

	return &themeCopy, nil
}

// UpsertTheme updates the theme settings.
func (db *Database) UpsertTheme(settings *models.Theme) (*models.Theme, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save theme in the database")
		return nil, err
	}

	db.Logger.Debug().Msg("db: Theme saved")

	themeCache = settings
	themeCopyCache = nil

	return settings, nil

}
