package chapter_downloader

import (
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util"
	"sync"
)

const (
	QueueStatusNotStarted  QueueStatus = "not_started"
	QueueStatusDownloading QueueStatus = "downloading"
	QueueStatusErrored     QueueStatus = "errored"
)

type (
	// Queue is used to manage the download queue.
	// If feeds the downloader with the next item in the queue.
	Queue struct {
		logger  *zerolog.Logger
		mu      sync.Mutex
		db      *db.Database
		current *QueueInfo
		runCh   chan *QueueInfo // Channel to tell downloader to run the next item
	}

	QueueStatus string

	// QueueInfo stores details about the download progress of a chapter.
	QueueInfo struct {
		DownloadID
		Pages          []*manga_providers.ChapterPage
		DownloadedUrls []string    `json:"downloadedUrls"`
		Status         QueueStatus `json:"status"`
	}
)

func NewQueue(db *db.Database, logger *zerolog.Logger, runCh chan *QueueInfo) *Queue {
	return &Queue{
		logger: logger,
		db:     db,
		runCh:  runCh,
	}
}

// Add adds a chapter to the download queue.
// It tells the queue to download the next item if possible.
func (q *Queue) Add(id DownloadID, pages []*manga_providers.ChapterPage, runNext bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	marshalled, err := json.Marshal(pages)
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to marshal pages for id %v", id)
		return err
	}

	err = q.db.InsertChapterDownloadQueueItem(&models.ChapterDownloadQueueItem{
		BaseModel: models.BaseModel{},
		Provider:  id.Provider,
		MediaID:   id.MediaId,
		ChapterID: id.ChapterId,
		PageData:  marshalled,
		Status:    string(QueueStatusNotStarted),
	})
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to insert chapter download queue item for id %v", id)
		return err
	}

	q.logger.Info().Msgf("chapter downloader: Added chapter to download queue: %s", id.ChapterId)

	if runNext {
		// Tells queue to run next if possible
		go q.runNext()
	}

	return nil
}

func (q *Queue) HasCompleted(id DownloadID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current.Status == QueueStatusErrored {
		// Update the status of the current item in the database.
		_ = q.db.UpdateChapterDownloadQueueItemStatus(q.current.DownloadID.Provider, q.current.DownloadID.MediaId, q.current.DownloadID.ChapterId, string(QueueStatusErrored))
	} else {
		// Dequeue the item from the database.
		_, err := q.db.DequeueChapterDownloadQueueItem()
		if err != nil {
			q.logger.Error().Err(err).Msgf("Failed to dequeue chapter download queue item for id %v", id)
			return
		}
	}

	// Reset current item
	q.current = nil

	// Tells queue to run next if possible
	q.runNext()
}

// Run invokes runNext
func (q *Queue) Run() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Tells queue to run next if possible
	q.runNext()
}

// runNext runs the next item in the queue.
//   - Checks if there is a current item, if so, it returns.
//   - If nothing is running, it gets the next item (QueueInfo) from the database, sets it as current and sends it to the downloader.
func (q *Queue) runNext() {

	// Catch panic in runNext, so it doesn't bubble up and stop goroutines.
	defer util.HandlePanicInModuleThen("internal/manga/downloader/runNext", func() {
		q.logger.Error().Msg("chapter downloader: Panic in 'runNext'")
	})

	if q.current != nil {
		return
	}

	q.logger.Debug().Msg("chapter downloader: Checking next item in queue")

	// Get next item from the database.
	next, _ := q.db.GetNextChapterDownloadQueueItem()
	if next == nil {
		q.logger.Debug().Msg("chapter downloader: No next item in queue")
		return
	}

	id := DownloadID{
		Provider:  next.Provider,
		MediaId:   next.MediaID,
		ChapterId: next.ChapterID,
	}

	// Set the current item.
	q.current = &QueueInfo{
		DownloadID:     id,
		DownloadedUrls: make([]string, 0),
		Status:         QueueStatusNotStarted,
	}

	// Unmarshal the page data.
	err := json.Unmarshal(next.PageData, &q.current.Pages)
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to unmarshal pages for id %v", id)
		return
	}

	q.logger.Info().Msgf("chapter downloader: Running next item in queue: %s", id.ChapterId)

	// Tell Downloader to run
	q.runCh <- q.current
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (q *Queue) GetCurrent() (qi *QueueInfo, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current == nil {
		return nil, false
	}

	return q.current, true
}
