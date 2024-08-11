package db_bridge

import (
	"github.com/goccy/go-json"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/summary"
)

func GetScanSummaries(database *db.Database) ([]*db.ScanSummaryItem, error) {
	var res []*models.ScanSummary
	err := database.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	var items []*db.ScanSummaryItem
	for _, r := range res {
		smBytes := r.Value
		var sm summary.ScanSummary
		if err := json.Unmarshal(smBytes, &sm); err != nil {
			return nil, err
		}
		items = append(items, &db.ScanSummaryItem{
			CreatedAt:   r.CreatedAt,
			ScanSummary: &sm,
		})
	}

	return items, nil
}

func InsertScanSummary(db *db.Database, sm *summary.ScanSummary) error {
	if sm == nil {
		return nil
	}

	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.Gorm().Create(&models.ScanSummary{
		Value: bytes,
	}).Error
}
