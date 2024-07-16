package db

import (
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/summary"
	"time"
)

type ScanSummaryItem struct {
	CreatedAt   time.Time            `json:"createdAt"`
	ScanSummary *summary.ScanSummary `json:"scanSummary"`
}

func (db *Database) TrimScanSummaryEntries() {
	go func() {
		var count int64
		err := db.gormdb.Model(&models.ScanSummary{}).Count(&count).Error
		if err != nil {
			db.Logger.Error().Err(err).Msg("Failed to count scan summary entries")
			return
		}
		if count > 10 {
			// Leave 5 entries
			err = db.gormdb.Delete(&models.ScanSummary{}, "id IN (SELECT id FROM scan_summaries ORDER BY id ASC LIMIT ?)", count-5).Error
			if err != nil {
				db.Logger.Error().Err(err).Msg("Failed to delete old scan summary entries")
				return
			}
		}
	}()
}
