package db

import (
	"seanime/internal/database/models"
)

func (db *Database) TrimTorrentstreamHistory() {
	go func() {
		var count int64
		err := db.gormdb.Model(&models.TorrentstreamHistory{}).Count(&count).Error
		if err != nil {
			db.Logger.Error().Err(err).Msg("database: Failed to count torrent stream history entries")
			return
		}
		if count > 50 {
			// Leave 40 entries
			err = db.gormdb.Delete(&models.TorrentstreamHistory{}, "id IN (SELECT id FROM torrentstream_histories ORDER BY updated_at ASC LIMIT ?)", 10).Error
			if err != nil {
				db.Logger.Error().Err(err).Msg("database: Failed to delete old torrent stream history entries")
				return
			}
		}
	}()
}
