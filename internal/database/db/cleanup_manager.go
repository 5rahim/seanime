package db

import (
	"seanime/internal/database/models"
	"sync"
	"time"
)

// CleanupManager manages database cleanup operations to prevent concurrent access issues
type CleanupManager struct {
	db    *Database
	mutex sync.Mutex
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(db *Database) *CleanupManager {
	return &CleanupManager{
		db: db,
	}
}

// RunAllCleanupOperations runs all cleanup operations sequentially to avoid database locks
func (cm *CleanupManager) RunAllCleanupOperations() {
	go func() {
		cm.mutex.Lock()
		defer cm.mutex.Unlock()

		cm.db.Logger.Debug().Msg("database: Starting cleanup operations")

		// Run cleanup operations sequentially with small delays
		cm.trimScanSummaryEntries()
		time.Sleep(100 * time.Millisecond)

		cm.trimLocalFileEntries()
		time.Sleep(100 * time.Millisecond)

		cm.trimTorrentstreamHistory()

		cm.db.Logger.Debug().Msg("database: Cleanup operations completed")
	}()
}

// trimScanSummaryEntries trims scan summary entries (internal implementation)
func (cm *CleanupManager) trimScanSummaryEntries() {
	var count int64
	err := cm.db.gormdb.Model(&models.ScanSummary{}).Count(&count).Error
	if err != nil {
		cm.db.Logger.Error().Err(err).Msg("database: Failed to count scan summary entries")
		return
	}
	if count > 10 {
		// Use a more efficient DELETE approach without subquery
		var idsToDelete []uint
		err = cm.db.gormdb.Model(&models.ScanSummary{}).
			Select("id").
			Order("id ASC").
			Limit(int(count-5)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.db.Logger.Error().Err(err).Msg("database: Failed to get scan summary IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			// Batch delete to avoid "too many SQL variables" error
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.db.gormdb.Delete(&models.ScanSummary{}, batch).Error
				if err != nil {
					cm.db.Logger.Error().Err(err).Msg("database: Failed to delete old scan summary entries")
					return // Exit on first error
				}
			}
			cm.db.Logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old scan summary entries")
		}
	}
}

// trimLocalFileEntries trims local file entries (internal implementation)
func (cm *CleanupManager) trimLocalFileEntries() {
	var count int64
	err := cm.db.gormdb.Model(&models.LocalFiles{}).Count(&count).Error
	if err != nil {
		cm.db.Logger.Error().Err(err).Msg("database: Failed to count local file entries")
		return
	}
	if count > 10 {
		// Use a more efficient DELETE approach without subquery
		var idsToDelete []uint
		err = cm.db.gormdb.Model(&models.LocalFiles{}).
			Select("id").
			Order("id ASC").
			Limit(int(count-5)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.db.Logger.Error().Err(err).Msg("database: Failed to get local file IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			// Batch delete to avoid "too many SQL variables" error
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.db.gormdb.Delete(&models.LocalFiles{}, batch).Error
				if err != nil {
					cm.db.Logger.Error().Err(err).Msg("database: Failed to delete old local file entries")
					return // Exit on first error
				}
			}
			cm.db.Logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old local file entries")
		}
	}
}

// trimTorrentstreamHistory trims torrent stream history entries (internal implementation)
func (cm *CleanupManager) trimTorrentstreamHistory() {
	var count int64
	err := cm.db.gormdb.Model(&models.TorrentstreamHistory{}).Count(&count).Error
	if err != nil {
		cm.db.Logger.Error().Err(err).Msg("database: Failed to count torrent stream history entries")
		return
	}
	if count > 50 {
		// Use a more efficient DELETE approach without subquery
		var idsToDelete []uint
		err = cm.db.gormdb.Model(&models.TorrentstreamHistory{}).
			Select("id").
			Order("updated_at ASC").
			Limit(int(count-40)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.db.Logger.Error().Err(err).Msg("database: Failed to get torrent stream history IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			// Batch delete to avoid "too many SQL variables" error
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.db.gormdb.Delete(&models.TorrentstreamHistory{}, batch).Error
				if err != nil {
					cm.db.Logger.Error().Err(err).Msg("database: Failed to delete old torrent stream history entries")
					return // Exit on first error
				}
			}
			cm.db.Logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old torrent stream history entries")
		}
	}
}
