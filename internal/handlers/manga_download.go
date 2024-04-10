package handlers

import (
	"github.com/seanime-app/seanime/internal/events"
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
		StartNow   bool                     `json:"startNow"`
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
			StartNow:  b.StartNow,
		})
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(true)
}

// HandleGetMangaDownloadData returns the download data (manga.MediaDownloadData) for a specific media.
// This is used to display information about the downloaded and queued chapters.
//
//	POST /api/v1/manga/download-data
func HandleGetMangaDownloadData(c *RouteCtx) error {

	type body struct {
		MediaId int  `json:"mediaId"`
		Cached  bool `json:"cached"` // If false, it will refresh the data
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	data, err := c.App.MangaDownloader.GetMediaDownloads(b.MediaId, b.Cached)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)
}

// HandleGetMangaDownloadQueue is used to display the current download queue.
//
//	GET /api/v1/manga/download-queue
func HandleGetMangaDownloadQueue(c *RouteCtx) error {

	data, err := c.App.Database.GetChapterDownloadQueue()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)
}

// HandleStartMangaDownloadQueue
//
//	POST /api/v1/manga/download-queue/start
func HandleStartMangaDownloadQueue(c *RouteCtx) error {

	c.App.MangaDownloader.RunChapterDownloadQueue()

	return c.RespondWithData(true)
}

// HandleStopMangaDownloadQueue
//
//	POST /api/v1/manga/download-queue/stop
func HandleStopMangaDownloadQueue(c *RouteCtx) error {

	c.App.MangaDownloader.StopChapterDownloadQueue()

	return c.RespondWithData(true)

}

// HandleRefreshMangaDownloadData
// FIXME NOT USED
//
//	POST /api/v1/manga/download-data/refresh
func HandleRefreshMangaDownloadData(c *RouteCtx) error {

	data := c.App.MangaDownloader.RefreshMediaMap()

	return c.RespondWithData(data)
}

// HandleClearAllChapterDownloadQueue
//
//	DELETE /api/v1/manga/download-queue
func HandleClearAllChapterDownloadQueue(c *RouteCtx) error {

	err := c.App.Database.ClearAllChapterDownloadQueueItems()
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)

	return c.RespondWithData(true)
}

// HandleResetErroredChapterDownloadQueue
//
//	POST /api/v1/manga/download-queue/reset-errored
func HandleResetErroredChapterDownloadQueue(c *RouteCtx) error {

	err := c.App.Database.ResetErroredChapterDownloadQueueItems()
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)

	return c.RespondWithData(true)
}

// HandleDeleteMangaChapterDownload
//
//	DELETE /api/v1/manga/download-chapter
func HandleDeleteMangaChapterDownload(c *RouteCtx) error {

	type body struct {
		MediaId       int    `json:"mediaId"`
		Provider      string `json:"provider"`
		ChapterId     string `json:"chapterId"`
		ChapterNumber string `json:"chapterNumber"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.MangaDownloader.DeleteChapter(b.Provider, b.MediaId, b.ChapterId, b.ChapterNumber)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleGetMangaDownloadsList is used to display the list of downloaded manga.
// It returns a list of manga.DownloadListItem. The media data might be nil.
//
//	GET /api/v1/manga/downloads
func HandleGetMangaDownloadsList(c *RouteCtx) error {

	mangaCollection, err := c.App.GetMangaCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	res, err := c.App.MangaDownloader.NewDownloadList(&manga.NewDownloadListOptions{
		MangaCollection: mangaCollection,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(res)
}
