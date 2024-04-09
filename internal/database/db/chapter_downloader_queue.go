package db

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
	"gorm.io/gorm"
)

func (db *Database) GetChapterDownloadQueue() ([]*models.ChapterDownloadQueueItem, error) {
	var res []*models.ChapterDownloadQueueItem
	err := db.gormdb.Find(&res).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to get chapter download queue")
		return nil, err
	}

	return res, nil
}

func (db *Database) GetNextChapterDownloadQueueItem() (*models.ChapterDownloadQueueItem, error) {
	var res models.ChapterDownloadQueueItem
	err := db.gormdb.Where("status = ?", "not_started").First(&res).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			db.logger.Error().Err(err).Msg("db: Failed to get next chapter download queue item")
		}
		return nil, nil
	}

	return &res, nil
}

func (db *Database) DequeueChapterDownloadQueueItem() (*models.ChapterDownloadQueueItem, error) {
	// Pop the first item from the queue
	var res models.ChapterDownloadQueueItem
	err := db.gormdb.Where("status = ?", "downloading").First(&res).Error
	if err != nil {
		return nil, err
	}

	err = db.gormdb.Delete(&res).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to delete chapter download queue item")
		return nil, err
	}

	return &res, nil
}

func (db *Database) InsertChapterDownloadQueueItem(item *models.ChapterDownloadQueueItem) error {

	// Check if the item already exists
	var existingItem models.ChapterDownloadQueueItem
	err := db.gormdb.Where("provider = ? AND media_id = ? AND chapter_id = ?", item.Provider, item.MediaID, item.ChapterID).First(&existingItem).Error
	if err == nil {
		db.logger.Debug().Msg("db: Chapter download queue item already exists")
		return nil
	}

	err = db.gormdb.Create(item).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to insert chapter download queue item")
		return err
	}
	return nil
}

func (db *Database) UpdateChapterDownloadQueueItemStatus(provider string, mId int, chapterId string, status string) error {
	err := db.gormdb.Model(&models.ChapterDownloadQueueItem{}).
		Where("provider = ? AND media_id = ? AND chapter_id = ?", provider, mId, chapterId).
		Update("status", status).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to update chapter download queue item status")
		return err
	}
	return nil
}

func (db *Database) GetMediaQueuedChapters(mediaId int) ([]*models.ChapterDownloadQueueItem, error) {
	var res []*models.ChapterDownloadQueueItem
	err := db.gormdb.Where("media_id = ?", mediaId).Find(&res).Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to get media queued chapters")
		return nil, err
	}

	return res, nil
}

func (db *Database) ClearAllChapterDownloadQueueItems() error {
	err := db.gormdb.
		Where("status = ? OR status = ? OR status = ?", "not_started", "downloading", "errored").
		Delete(&models.ChapterDownloadQueueItem{}).
		Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to clear all chapter download queue items")
		return err
	}
	return nil
}

func (db *Database) ResetErroredChapterDownloadQueueItems() error {
	err := db.gormdb.Model(&models.ChapterDownloadQueueItem{}).
		Where("status = ?", "errored").
		Update("status", "not_started").Error
	if err != nil {
		db.logger.Error().Err(err).Msg("db: Failed to reset errored chapter download queue items")
		return err
	}
	return nil
}
