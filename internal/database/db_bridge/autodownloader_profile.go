package db_bridge

import (
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"

	"github.com/goccy/go-json"
)

func GetAutoDownloaderProfiles(db *db.Database) ([]*anime.AutoDownloaderProfile, error) {

	var res []*models.AutoDownloaderProfile
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	var profiles []*anime.AutoDownloaderProfile
	for _, r := range res {
		smBytes := r.Value
		var sm anime.AutoDownloaderProfile
		if err := json.Unmarshal(smBytes, &sm); err != nil {
			return nil, err
		}
		sm.DbID = r.ID
		profiles = append(profiles, &sm)
	}

	return profiles, nil
}

func GetAutoDownloaderProfile(db *db.Database, id uint) (*anime.AutoDownloaderProfile, error) {
	var res models.AutoDownloaderProfile
	err := db.Gorm().First(&res, id).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	smBytes := res.Value
	var sm anime.AutoDownloaderProfile
	if err := json.Unmarshal(smBytes, &sm); err != nil {
		return nil, err
	}
	sm.DbID = res.ID

	return &sm, nil
}

func InsertAutoDownloaderProfile(db *db.Database, sm *anime.AutoDownloaderProfile) error {

	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.Gorm().Create(&models.AutoDownloaderProfile{
		Value: bytes,
	}).Error
}

func DeleteAutoDownloaderProfile(db *db.Database, id uint) error {

	return db.Gorm().Delete(&models.AutoDownloaderProfile{}, id).Error
}

func UpdateAutoDownloaderProfile(db *db.Database, id uint, sm *anime.AutoDownloaderProfile) error {

	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.Gorm().Model(&models.AutoDownloaderProfile{}).Where("id = ?", id).Update("value", bytes).Error
}
