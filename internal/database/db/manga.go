package db

import (
	"fmt"
	"seanime/internal/database/models"
	"seanime/internal/util/result"
)

var mangaMappingCache = result.NewResultMap[string, *models.MangaMapping]()

func formatMangaMappingCacheKey(provider string, mediaId int) string {
	return fmt.Sprintf("%s$%d", provider, mediaId)
}

func (db *Database) GetMangaMapping(provider string, mediaId int) (*models.MangaMapping, bool) {

	if res, ok := mangaMappingCache.Get(formatMangaMappingCacheKey(provider, mediaId)); ok {
		return res, true
	}

	var res models.MangaMapping
	err := db.gormdb.Where("provider = ? AND media_id = ?", provider, mediaId).First(&res).Error
	if err != nil {
		return nil, false
	}

	mangaMappingCache.Set(formatMangaMappingCacheKey(provider, mediaId), &res)

	return &res, true
}

func (db *Database) InsertMangaMapping(provider string, mediaId int, mangaId string) error {
	mapping := models.MangaMapping{
		Provider: provider,
		MediaID:  mediaId,
		MangaID:  mangaId,
	}

	mangaMappingCache.Set(formatMangaMappingCacheKey(provider, mediaId), &mapping)

	return db.gormdb.Save(&mapping).Error
}

func (db *Database) DeleteMangaMapping(provider string, mediaId int) error {
	err := db.gormdb.Where("provider = ? AND media_id = ?", provider, mediaId).Delete(&models.MangaMapping{}).Error
	if err != nil {
		return err
	}

	mangaMappingCache.Delete(formatMangaMappingCacheKey(provider, mediaId))
	return nil
}
