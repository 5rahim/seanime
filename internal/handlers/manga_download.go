package handlers

import (
	"seanime/internal/events"
	"seanime/internal/manga"
	chapter_downloader "seanime/internal/manga/downloader"
	"time"
)

// HandleDownloadMangaChapters
//
//	@summary adds chapters to the download queue.
//	@route /api/v1/manga/download-chapters [POST]
//	@returns bool
func HandleDownloadMangaChapters(c *RouteCtx) error {

	type body struct {
		MediaId    int      `json:"mediaId"`
		Provider   string   `json:"provider"`
		ChapterIds []string `json:"chapterIds"`
		StartNow   bool     `json:"startNow"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent(events.InfoToast, "Adding chapters to download queue...")

	// Add chapters to the download queue
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
		time.Sleep(400 * time.Millisecond) // Sleep to avoid rate limiting
	}

	return c.RespondWithData(true)
}

// HandleGetMangaDownloadData
//
//	@summary returns the download data for a specific media.
//	@desc This is used to display information about the downloaded and queued chapters in the UI.
//	@desc If the 'cached' parameter is false, it will refresh the data by rescanning the download folder.
//	@route /api/v1/manga/download-data [POST]
//	@returns manga.MediaDownloadData
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

// HandleGetMangaDownloadQueue
//
//	@summary returns the items in the download queue.
//	@route /api/v1/manga/download-queue [GET]
//	@returns []models.ChapterDownloadQueueItem
func HandleGetMangaDownloadQueue(c *RouteCtx) error {

	data, err := c.App.Database.GetChapterDownloadQueue()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)
}

// HandleStartMangaDownloadQueue
//
//	@summary starts the download queue if it's not already running.
//	@desc This will start the download queue if it's not already running.
//	@desc Returns 'true' whether the queue was started or not.
//	@route /api/v1/manga/download-queue/start [POST]
//	@returns bool
func HandleStartMangaDownloadQueue(c *RouteCtx) error {

	c.App.MangaDownloader.RunChapterDownloadQueue()

	return c.RespondWithData(true)
}

// HandleStopMangaDownloadQueue
//
//	@summary stops the manga download queue.
//	@desc This will stop the manga download queue.
//	@desc Returns 'true' whether the queue was stopped or not.
//	@route /api/v1/manga/download-queue/stop [POST]
//	@returns bool
func HandleStopMangaDownloadQueue(c *RouteCtx) error {

	c.App.MangaDownloader.StopChapterDownloadQueue()

	return c.RespondWithData(true)

}

// HandleClearAllChapterDownloadQueue
//
//	@summary clears all chapters from the download queue.
//	@desc This will clear all chapters from the download queue.
//	@desc Returns 'true' whether the queue was cleared or not.
//	@desc This will also send a websocket event telling the client to refetch the download queue.
//	@route /api/v1/manga/download-queue [DELETE]
//	@returns bool
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
//	@summary resets the errored chapters in the download queue.
//	@desc This will reset the errored chapters in the download queue, so they can be re-downloaded.
//	@desc Returns 'true' whether the queue was reset or not.
//	@desc This will also send a websocket event telling the client to refetch the download queue.
//	@route /api/v1/manga/download-queue/reset-errored [POST]
//	@returns bool
func HandleResetErroredChapterDownloadQueue(c *RouteCtx) error {

	err := c.App.Database.ResetErroredChapterDownloadQueueItems()
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)

	return c.RespondWithData(true)
}

// HandleDeleteMangaDownloadedChapters
//
//	@summary deletes downloaded chapters.
//	@desc This will delete downloaded chapters from the filesystem.
//	@desc Returns 'true' whether the chapters were deleted or not.
//	@desc The client should refetch the download data after this.
//	@route /api/v1/manga/download-chapter [DELETE]
//	@returns bool
func HandleDeleteMangaDownloadedChapters(c *RouteCtx) error {

	type body struct {
		DownloadIds []chapter_downloader.DownloadID `json:"downloadIds"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.MangaDownloader.DeleteChapters(b.DownloadIds)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleGetMangaDownloadsList
//
//	@summary displays the list of downloaded manga.
//	@desc This analyzes the download folder and returns a well-formatted structure for displaying downloaded manga.
//	@desc It returns a list of manga.DownloadListItem where the media data might be nil if it's not in the AniList collection.
//	@route /api/v1/manga/downloads [GET]
//	@returns []manga.DownloadListItem
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
