package db

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
)

func (db *Database) GetAutoDownloaderRules() ([]*anime.AutoDownloaderRule, error) {
	var res []*models.AutoDownloaderRule
	err := db.gormdb.Find(&res).Error
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

func (db *Database) GetAutoDownloaderRule(id uint) (*anime.AutoDownloaderRule, error) {
	var res models.AutoDownloaderRule
	err := db.gormdb.First(&res, id).Error
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

func (db *Database) InsertAutoDownloaderRule(sm *anime.AutoDownloaderRule) error {
	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.gormdb.Create(&models.AutoDownloaderRule{
		Value: bytes,
	}).Error
}

func (db *Database) DeleteAutoDownloaderRule(id uint) error {
	return db.gormdb.Delete(&models.AutoDownloaderRule{}, id).Error
}

func (db *Database) UpdateAutoDownloaderRule(id uint, sm *anime.AutoDownloaderRule) error {
	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.gormdb.Model(&models.AutoDownloaderRule{}).Where("id = ?", id).Update("value", bytes).Error
}
