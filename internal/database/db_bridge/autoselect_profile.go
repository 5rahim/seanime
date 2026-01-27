package db_bridge

import (
	"errors"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"

	"github.com/goccy/go-json"
	"gorm.io/gorm"
)

func FindAutoSelectProfile(db *db.Database) (*anime.AutoSelectProfile, bool) {
	profile, err := GetAutoSelectProfile(db)
	return profile, err == nil && profile.DbID != 0
}

// GetAutoSelectProfile returns the single autoselect profile if it exists
func GetAutoSelectProfile(db *db.Database) (*anime.AutoSelectProfile, error) {
	var res models.AutoSelectProfile
	err := db.Gorm().First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &anime.AutoSelectProfile{}, nil
		}
		return nil, err
	}

	// Unmarshal the data
	var profile anime.AutoSelectProfile
	if err := json.Unmarshal(res.Value, &profile); err != nil {
		return nil, err
	}
	profile.DbID = res.ID

	return &profile, nil
}

// SaveAutoSelectProfile saves or updates the autoselect profile
// Since there's only one profile at all time, this will create or update it
func SaveAutoSelectProfile(db *db.Database, profile *anime.AutoSelectProfile) error {
	// Marshal the data
	bytes, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	// Check if a profile already exists
	var existing models.AutoSelectProfile
	err = db.Gorm().First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new profile
		return db.Gorm().Create(&models.AutoSelectProfile{
			Value: bytes,
		}).Error
	} else if err != nil {
		return err
	}

	// Update existing profile
	return db.Gorm().Model(&models.AutoSelectProfile{}).
		Where("id = ?", existing.ID).
		Update("value", bytes).Error
}

// DeleteAutoSelectProfile deletes the autoselect profile
func DeleteAutoSelectProfile(db *db.Database) error {
	return db.Gorm().Where("1 = 1").Delete(&models.AutoSelectProfile{}).Error
}
