package db

import (
	"seanime/internal/database/models"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// CleanupManager manages database cleanup operations to prevent concurrent access issues
type CleanupManager struct {
	gormdb *gorm.DB
	logger *zerolog.Logger
}

func NewCleanupManager(gormdb *gorm.DB, logger *zerolog.Logger) *CleanupManager {
	return &CleanupManager{
		gormdb: gormdb,
		logger: logger,
	}
}

func (cm *CleanupManager) RunAllCleanupOperations() {
	cm.logger.Debug().Msg("database: Starting cleanup operations")

	cm.trimScanSummaryEntries()
	cm.trimLocalFileEntries()
	cm.trimTorrentstreamHistory()

	cm.logger.Debug().Msg("database: Cleanup operations completed")
}

// trimScanSummaryEntries trims scan summary entries
func (cm *CleanupManager) trimScanSummaryEntries() {
	var count int64
	err := cm.gormdb.Model(&models.ScanSummary{}).Count(&count).Error
	if err != nil {
		cm.logger.Error().Err(err).Msg("database: Failed to count scan summary entries")
		return
	}
	if count > 10 {
		var idsToDelete []uint
		err = cm.gormdb.Model(&models.ScanSummary{}).
			Select("id").
			Order("id ASC").
			Limit(int(count-5)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.logger.Error().Err(err).Msg("database: Failed to get scan summary IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.gormdb.Delete(&models.ScanSummary{}, batch).Error
				if err != nil {
					cm.logger.Error().Err(err).Msg("database: Failed to delete old scan summary entries")
					return // Exit on first error
				}
			}
			cm.logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old scan summary entries")
		}
	}
}

// trimLocalFileEntries trims local file entries
func (cm *CleanupManager) trimLocalFileEntries() {
	var count int64
	err := cm.gormdb.Model(&models.LocalFiles{}).Count(&count).Error
	if err != nil {
		cm.logger.Error().Err(err).Msg("database: Failed to count local file entries")
		return
	}
	if count > 10 {
		var idsToDelete []uint
		err = cm.gormdb.Model(&models.LocalFiles{}).
			Select("id").
			Order("id ASC").
			Limit(int(count-5)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.logger.Error().Err(err).Msg("database: Failed to get local file IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.gormdb.Delete(&models.LocalFiles{}, batch).Error
				if err != nil {
					cm.logger.Error().Err(err).Msg("database: Failed to delete old local file entries")
					return // Exit on first error
				}
			}
			cm.logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old local file entries")
		}
	}
}

// trimTorrentstreamHistory trims torrent stream history entries
func (cm *CleanupManager) trimTorrentstreamHistory() {
	var count int64
	err := cm.gormdb.Model(&models.TorrentstreamHistory{}).Count(&count).Error
	if err != nil {
		cm.logger.Error().Err(err).Msg("database: Failed to count torrent stream history entries")
		return
	}
	if count > 50 {
		var idsToDelete []uint
		err = cm.gormdb.Model(&models.TorrentstreamHistory{}).
			Select("id").
			Order("updated_at ASC").
			Limit(int(count-40)).
			Pluck("id", &idsToDelete).Error
		if err != nil {
			cm.logger.Error().Err(err).Msg("database: Failed to get torrent stream history IDs to delete")
			return
		}

		if len(idsToDelete) > 0 {
			batchSize := 900
			for i := 0; i < len(idsToDelete); i += batchSize {
				end := i + batchSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}
				batch := idsToDelete[i:end]
				err = cm.gormdb.Delete(&models.TorrentstreamHistory{}, batch).Error
				if err != nil {
					cm.logger.Error().Err(err).Msg("database: Failed to delete old torrent stream history entries")
					return // Exit on first error
				}
			}
			cm.logger.Debug().Int("deleted", len(idsToDelete)).Msg("database: Deleted old torrent stream history entries")
		}
	}
}
