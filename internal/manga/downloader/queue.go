package chapter_downloader

import (
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"sync"
	"time"
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
		logger         *zerolog.Logger
		mu             sync.Mutex
		db             *db.Database
		current        *QueueInfo
		runCh          chan *QueueInfo // Channel to tell downloader to run the next item
		active         bool
		wsEventManager events.WSEventManagerInterface
	}

	QueueStatus string

	// QueueInfo stores details about the download progress of a chapter.
	QueueInfo struct {
		DownloadID
		Pages          []*hibikemanga.ChapterPage
		DownloadedUrls []string
		Status         QueueStatus
	}
)

func NewQueue(db *db.Database, logger *zerolog.Logger, wsEventManager events.WSEventManagerInterface, runCh chan *QueueInfo) *Queue {
	return &Queue{
		logger:         logger,
		db:             db,
		runCh:          runCh,
		wsEventManager: wsEventManager,
	}
}

// Add adds a chapter to the download queue.
// It tells the queue to download the next item if possible.
func (q *Queue) Add(id DownloadID, pages []*hibikemanga.ChapterPage, runNext bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	marshalled, err := json.Marshal(pages)
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to marshal pages for id %v", id)
		return err
	}

	err = q.db.InsertChapterDownloadQueueItem(&models.ChapterDownloadQueueItem{
		BaseModel:     models.BaseModel{},
		Provider:      id.Provider,
		MediaID:       id.MediaId,
		ChapterNumber: id.ChapterNumber,
		ChapterID:     id.ChapterId,
		PageData:      marshalled,
		Status:        string(QueueStatusNotStarted),
	})
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to insert chapter download queue item for id %v", id)
		return err
	}

	q.logger.Info().Msgf("chapter downloader: Added chapter to download queue: %s", id.ChapterId)

	q.wsEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)

	if runNext && q.active {
		// Tells queue to run next if possible
		go q.runNext()
	}

	return nil
}

func (q *Queue) HasCompleted(queueInfo *QueueInfo) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if queueInfo.Status == QueueStatusErrored {
		q.logger.Warn().Msgf("chapter downloader: Errored %s", queueInfo.DownloadID.ChapterId)
		// Update the status of the current item in the database.
		_ = q.db.UpdateChapterDownloadQueueItemStatus(q.current.DownloadID.Provider, q.current.DownloadID.MediaId, q.current.DownloadID.ChapterId, string(QueueStatusErrored))
	} else {
		q.logger.Debug().Msgf("chapter downloader: Dequeueing %s", queueInfo.DownloadID.ChapterId)
		// Dequeue the item from the database.
		_, err := q.db.DequeueChapterDownloadQueueItem()
		if err != nil {
			q.logger.Error().Err(err).Msgf("Failed to dequeue chapter download queue item for id %v", queueInfo.DownloadID)
			return
		}
	}

	q.wsEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)
	q.wsEventManager.SendEvent(events.RefreshedMangaDownloadData, nil)

	// Reset current item
	q.current = nil

	if q.active {
		// Tells queue to run next if possible
		q.runNext()
	}
}

// Run activates the queue and invokes runNext
func (q *Queue) Run() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.active {
		q.logger.Debug().Msg("chapter downloader: Starting queue")
	}

	q.active = true

	// Tells queue to run next if possible
	q.runNext()
}

// Stop deactivates the queue
func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.active {
		q.logger.Debug().Msg("chapter downloader: Stopping queue")
	}

	q.active = false
}

// runNext runs the next item in the queue.
//   - Checks if there is a current item, if so, it returns.
//   - If nothing is running, it gets the next item (QueueInfo) from the database, sets it as current and sends it to the downloader.
func (q *Queue) runNext() {

	q.logger.Debug().Msg("chapter downloader: Processing next item in queue")

	// Catch panic in runNext, so it doesn't bubble up and stop goroutines.
	defer util.HandlePanicInModuleThen("internal/manga/downloader/runNext", func() {
		q.logger.Error().Msg("chapter downloader: Panic in 'runNext'")
	})

	if q.current != nil {
		q.logger.Debug().Msg("chapter downloader: Current item is not nil")
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
		Provider:      next.Provider,
		MediaId:       next.MediaID,
		ChapterId:     next.ChapterID,
		ChapterNumber: next.ChapterNumber,
	}

	q.logger.Debug().Msgf("chapter downloader: Preparing next item in queue: %s", id.ChapterId)

	q.wsEventManager.SendEvent(events.ChapterDownloadQueueUpdated, nil)
	// Update status
	_ = q.db.UpdateChapterDownloadQueueItemStatus(id.Provider, id.MediaId, id.ChapterId, string(QueueStatusDownloading))

	// Set the current item.
	q.current = &QueueInfo{
		DownloadID:     id,
		DownloadedUrls: make([]string, 0),
		Status:         QueueStatusDownloading,
	}

	// Unmarshal the page data.
	err := json.Unmarshal(next.PageData, &q.current.Pages)
	if err != nil {
		q.logger.Error().Err(err).Msgf("Failed to unmarshal pages for id %v", id)
		_ = q.db.UpdateChapterDownloadQueueItemStatus(id.Provider, id.MediaId, id.ChapterId, string(QueueStatusNotStarted))
		return
	}

	// TODO: This is a temporary fix to prevent the downloader from running too fast.
	time.Sleep(5 * time.Second)

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
