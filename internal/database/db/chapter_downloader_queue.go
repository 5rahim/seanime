package db

import (
	"errors"
	"gorm.io/gorm"
	"seanime/internal/database/models"
)

func (db *Database) GetChapterDownloadQueue() ([]*models.ChapterDownloadQueueItem, error) {
	var res []*models.ChapterDownloadQueueItem
	err := db.gormdb.Find(&res).Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to get chapter download queue")
		return nil, err
	}

	return res, nil
}

func (db *Database) GetNextChapterDownloadQueueItem() (*models.ChapterDownloadQueueItem, error) {
	var res models.ChapterDownloadQueueItem
	err := db.gormdb.Where("status = ?", "not_started").First(&res).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			db.Logger.Error().Err(err).Msg("db: Failed to get next chapter download queue item")
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
		db.Logger.Error().Err(err).Msg("db: Failed to delete chapter download queue item")
		return nil, err
	}

	return &res, nil
}

func (db *Database) InsertChapterDownloadQueueItem(item *models.ChapterDownloadQueueItem) error {

	// Check if the item already exists
	var existingItem models.ChapterDownloadQueueItem
	err := db.gormdb.Where("provider = ? AND media_id = ? AND chapter_id = ?", item.Provider, item.MediaID, item.ChapterID).First(&existingItem).Error
	if err == nil {
		db.Logger.Debug().Msg("db: Chapter download queue item already exists")
		return errors.New("chapter is already in the download queue")
	}

	if item.ChapterID == "" {
		return errors.New("chapter ID is empty")
	}
	if item.Provider == "" {
		return errors.New("provider is empty")
	}
	if item.MediaID == 0 {
		return errors.New("media ID is empty")
	}
	if item.ChapterNumber == "" {
		return errors.New("chapter number is empty")
	}

	err = db.gormdb.Create(item).Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to insert chapter download queue item")
		return err
	}
	return nil
}

func (db *Database) UpdateChapterDownloadQueueItemStatus(provider string, mId int, chapterId string, status string) error {
	err := db.gormdb.Model(&models.ChapterDownloadQueueItem{}).
		Where("provider = ? AND media_id = ? AND chapter_id = ?", provider, mId, chapterId).
		Update("status", status).Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to update chapter download queue item status")
		return err
	}
	return nil
}

func (db *Database) GetMediaQueuedChapters(mediaId int) ([]*models.ChapterDownloadQueueItem, error) {
	var res []*models.ChapterDownloadQueueItem
	err := db.gormdb.Where("media_id = ?", mediaId).Find(&res).Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to get media queued chapters")
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
		db.Logger.Error().Err(err).Msg("db: Failed to clear all chapter download queue items")
		return err
	}
	return nil
}

func (db *Database) ResetErroredChapterDownloadQueueItems() error {
	err := db.gormdb.Model(&models.ChapterDownloadQueueItem{}).
		Where("status = ?", "errored").
		Update("status", "not_started").Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to reset errored chapter download queue items")
		return err
	}
	return nil
}

func (db *Database) ResetDownloadingChapterDownloadQueueItems() error {
	err := db.gormdb.Model(&models.ChapterDownloadQueueItem{}).
		Where("status = ?", "downloading").
		Update("status", "not_started").Error
	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to reset downloading chapter download queue items")
		return err
	}
	return nil
}
