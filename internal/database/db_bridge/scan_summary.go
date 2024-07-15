package db_bridge

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/summary"
	"time"
)

type ScanSummaryItem struct {
	CreatedAt   time.Time            `json:"createdAt"`
	ScanSummary *summary.ScanSummary `json:"scanSummary"`
}

func GetScanSummaries(db *db.Database) ([]*ScanSummaryItem, error) {
	var res []*models.ScanSummary
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the data
	var items []*ScanSummaryItem
	for _, r := range res {
		smBytes := r.Value
		var sm summary.ScanSummary
		if err := json.Unmarshal(smBytes, &sm); err != nil {
			return nil, err
		}
		items = append(items, &ScanSummaryItem{
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
