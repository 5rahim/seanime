package db

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/summary"
	"time"
)

type ScanSummaryItem struct {
	CreatedAt time.Time `json:"createdAt"`
	*summary.ScanSummary
}

func (db *Database) GetLastScanSummary() (*summary.ScanSummary, uint, error) {
	// Get the latest entry
	var res models.ScanSummary
	err := db.gormdb.Last(&res).Error
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal the data
	smBytes := res.Value
	var sm *summary.ScanSummary
	if err := json.Unmarshal(smBytes, &sm); err != nil {
		return nil, 0, err
	}

	return sm, res.ID, nil
}

func (db *Database) GetScanSummaries() ([]*ScanSummaryItem, error) {
	var res []*models.ScanSummary
	err := db.gormdb.Find(&res).Error
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

func (db *Database) InsertScanSummary(sm *summary.ScanSummary) error {
	if sm == nil {
		return nil
	}

	// Marshal the data
	bytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	// Save the data
	return db.gormdb.Create(&models.ScanSummary{
		Value: bytes,
	}).Error
}

func (db *Database) CleanUpScanSummaries() {
	go func() {
		var count int64
		err := db.gormdb.Model(&models.ScanSummary{}).Count(&count).Error
		if err != nil {
			db.logger.Error().Err(err).Msg("Failed to count scan summary entries")
			return
		}
		if count > 10 {
			// Leave 5 entries
			err = db.gormdb.Delete(&models.ScanSummary{}, "id IN (SELECT id FROM scan_summaries ORDER BY id ASC LIMIT ?)", count-5).Error
			if err != nil {
				db.logger.Error().Err(err).Msg("Failed to delete old scan summary entries")
				return
			}
		}
	}()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) insertScanSummary(sm *summary.ScanSummary) (*summary.ScanSummary, error) {
	err := db.gormdb.Create(sm).Error

	if err != nil {
		return nil, err
	}
	return sm, nil
}
