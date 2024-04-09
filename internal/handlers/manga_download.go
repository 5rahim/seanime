package handlers

import (
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/manga/providers"
)

// HandleDownloadMangaChapters
//
//	POST /api/v1/manga/download-chapters
func HandleDownloadMangaChapters(c *RouteCtx) error {

	type body struct {
		MediaId    int                      `json:"mediaId"`
		Provider   manga_providers.Provider `json:"provider"`
		ChapterIds []string                 `json:"chapterIds"`
		Start      bool                     `json:"start"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	for _, chapterId := range b.ChapterIds {
		err := c.App.MangaDownloader.DownloadChapter(manga.DownloadChapterOptions{
			Provider:  b.Provider,
			MediaId:   b.MediaId,
			ChapterId: chapterId,
		})
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(true)
}

// HandleGetMangaDownloadData
//
//	POST /api/v1/manga/download-data
func HandleGetMangaDownloadData(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	data, err := c.App.MangaDownloader.GetMediaDownloads(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)
}
