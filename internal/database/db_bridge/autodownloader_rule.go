package db_bridge

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
)

func GetAutoDownloaderRules(db *db.Database) ([]*anime.AutoDownloaderRule, error) {
	var res []*models.AutoDownloaderRule
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	var rules []*anime.AutoDownloaderRule
	for _, r := range res {
		smBytes := r.Value
		var sm anime.AutoDownloaderRule
		if err := json.Unmarshal(smBytes, &sm); err != nil {
			return nil, err
		}
		sm.DbID = r.ID
		rules = append(rules, &sm)
	}

	return rules, nil
}

func GetAutoDownloaderRule(db *db.Database, id uint) (*anime.AutoDownloaderRule, error) {
	var res models.AutoDownloaderRule
	err := db.Gorm().First(&res, id).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	smBytes := res.Value
	var sm anime.AutoDownloaderRule
	if err := json.Unmarshal(smBytes, &sm); err != nil {
		return nil, err
	}
	sm.DbID = res.ID

	return &sm, nil
}

func InsertAutoDownloaderRule(db *db.Database, sm *anime.AutoDownloaderRule) error {
	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.Gorm().Create(&models.AutoDownloaderRule{
		Value: bytes,
	}).Error
}

func DeleteAutoDownloaderRule(db *db.Database, id uint) error {
	return db.Gorm().Delete(&models.AutoDownloaderRule{}, id).Error
}

func UpdateAutoDownloaderRule(db *db.Database, id uint, sm *anime.AutoDownloaderRule) error {
	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.Gorm().Model(&models.AutoDownloaderRule{}).Where("id = ?", id).Update("value", bytes).Error
}
