package db

import (
	"fmt"
	"seanime/internal/database/models"
	"seanime/internal/util/result"
)

var mangaMappingCache = result.NewMap[string, *models.MangaMapping]()

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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var mangaChapterContainerCache = result.NewMap[string, *models.MangaChapterContainer]()

func formatMangaChapterContainerCacheKey(provider string, mediaId int, chapterId string) string {
	return fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)
}

func (db *Database) GetMangaChapterContainer(provider string, mediaId int, chapterId string) (*models.MangaChapterContainer, bool) {

	if res, ok := mangaChapterContainerCache.Get(formatMangaChapterContainerCacheKey(provider, mediaId, chapterId)); ok {
		return res, true
	}

	var res models.MangaChapterContainer
	err := db.gormdb.Where("provider = ? AND media_id = ? AND chapter_id = ?", provider, mediaId, chapterId).First(&res).Error
	if err != nil {
		return nil, false
	}

	mangaChapterContainerCache.Set(formatMangaChapterContainerCacheKey(provider, mediaId, chapterId), &res)

	return &res, true
}

func (db *Database) InsertMangaChapterContainer(provider string, mediaId int, chapterId string, chapterContainer []byte) error {
	container := models.MangaChapterContainer{
		Provider:  provider,
		MediaID:   mediaId,
		ChapterID: chapterId,
		Data:      chapterContainer,
	}

	mangaChapterContainerCache.Set(formatMangaChapterContainerCacheKey(provider, mediaId, chapterId), &container)

	return db.gormdb.Save(&container).Error
}

func (db *Database) DeleteMangaChapterContainer(provider string, mediaId int, chapterId string) error {
	err := db.gormdb.Where("provider = ? AND media_id = ? AND chapter_id = ?", provider, mediaId, chapterId).Delete(&models.MangaChapterContainer{}).Error
	if err != nil {
		return err
	}

	mangaChapterContainerCache.Delete(formatMangaChapterContainerCacheKey(provider, mediaId, chapterId))
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
