package db

import (
	"fmt"
	"seanime/internal/database/models"
	"seanime/internal/util/result"
)

var onlinestreamMappingCache = result.NewMap[string, *models.OnlinestreamMapping]()

func formatOnlinestreamMappingCacheKey(provider string, mediaId int) string {
	return fmt.Sprintf("%s$%d", provider, mediaId)
}

func (db *Database) GetOnlinestreamMapping(provider string, mediaId int) (*models.OnlinestreamMapping, bool) {

	if res, ok := onlinestreamMappingCache.Get(formatOnlinestreamMappingCacheKey(provider, mediaId)); ok {
		return res, true
	}

	var res models.OnlinestreamMapping
	err := db.gormdb.Where("provider = ? AND media_id = ?", provider, mediaId).First(&res).Error
	if err != nil {
		return nil, false
	}

	onlinestreamMappingCache.Set(formatOnlinestreamMappingCacheKey(provider, mediaId), &res)

	return &res, true
}

func (db *Database) InsertOnlinestreamMapping(provider string, mediaId int, animeId string) error {
	mapping := models.OnlinestreamMapping{
		Provider: provider,
		MediaID:  mediaId,
		AnimeID:  animeId,
	}

	onlinestreamMappingCache.Set(formatOnlinestreamMappingCacheKey(provider, mediaId), &mapping)

	return db.gormdb.Save(&mapping).Error
}

func (db *Database) DeleteOnlinestreamMapping(provider string, mediaId int) error {
	err := db.gormdb.Where("provider = ? AND media_id = ?", provider, mediaId).Delete(&models.OnlinestreamMapping{}).Error
	if err != nil {
		return err
	}

	onlinestreamMappingCache.Delete(formatOnlinestreamMappingCacheKey(provider, mediaId))
	return nil
}
