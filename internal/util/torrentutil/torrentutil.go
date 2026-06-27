package torrentutil

import (
	"fmt"
	"io"
	httputil "seanime/internal/util/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
)

// +-----------------------+
// +   anacrolix/torrent   +
// +-----------------------+

const (
	piecesForNow        = int64(5)
	piecesForHighBefore = int64(2)
	piecesForNext       = int64(30)
	piecesForReadahead  = int64(30)
	defaultUpdateStride = int64(512 * 1024)
)

// readerInfo tracks information about an active reader
type readerInfo struct {
	id         string
	position   int64
	lastAccess time.Time
}

// priorityManager manages piece priorities for multiple readers on the same file
type priorityManager struct {
	mu        sync.RWMutex
	readers   map[string]*readerInfo
	torrent   *torrent.Torrent
	file      *torrent.File
	logger    *zerolog.Logger
	createdAt time.Time
}

// global map to track priority managers per torrent+file combination
var (
	priorityManagers    = make(map[string]*priorityManager)
	priorityManagersMu  sync.RWMutex
	priorityCleanupOnce sync.Once
)

// getPriorityManager gets or creates a priority manager for a torrent+file combination
func getPriorityManager(t *torrent.Torrent, file *torrent.File, logger *zerolog.Logger) *priorityManager {
	key := fmt.Sprintf("%s:%s", t.InfoHash().String(), file.Path())

	priorityManagersMu.Lock()
	defer priorityManagersMu.Unlock()

	if pm, exists := priorityManagers[key]; exists {
		return pm
	}

	pm := &priorityManager{
		readers:   make(map[string]*readerInfo),
		torrent:   t,
		file:      file,
		logger:    new(logger.Sample(&zerolog.BasicSampler{N: 20})),
		createdAt: time.Now(),
	}
	priorityManagers[key] = pm

	priorityCleanupOnce.Do(func() {
		go cleanupStalePriorityManagers()
	})

	return pm
}

func cleanupStalePriorityManagers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for now := range ticker.C {
		priorityManagersMu.RLock()
		snapshot := make(map[string]*priorityManager, len(priorityManagers))
		for key, pm := range priorityManagers {
			snapshot[key] = pm
		}
		priorityManagersMu.RUnlock()

		emptyKeys := make([]string, 0)
		for key, pm := range snapshot {
			if pm.cleanupStaleReaders(now) {
				emptyKeys = append(emptyKeys, key)
			}
		}

		if len(emptyKeys) == 0 {
			continue
		}

		priorityManagersMu.Lock()
		for _, key := range emptyKeys {
			pm, ok := priorityManagers[key]
			if !ok {
				continue
			}
			if pm.readerCount() == 0 {
				delete(priorityManagers, key)
			}
		}
		priorityManagersMu.Unlock()
	}
}

// registerReader registers a new reader with the priority manager
func (pm *priorityManager) registerReader(readerID string, position int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.readers[readerID] = &readerInfo{
		id:         readerID,
		position:   position,
		lastAccess: time.Now(),
	}

	pm.updatePriorities()
}

// updateReaderPosition updates a reader's position and recalculates priorities
func (pm *priorityManager) updateReaderPosition(readerID string, position int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if reader, exists := pm.readers[readerID]; exists {
		reader.position = position
		reader.lastAccess = time.Now()
		pm.updatePriorities()
	}
}

func (pm *priorityManager) touchReader(readerID string, position int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if reader, exists := pm.readers[readerID]; exists {
		reader.position = position
		reader.lastAccess = time.Now()
	}
}

// unregisterReader removes a reader from tracking
func (pm *priorityManager) unregisterReader(readerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.readers, readerID)

	// If no more readers, clean up and recalculate priorities
	if len(pm.readers) == 0 {
		pm.resetAllPriorities()
	} else {
		pm.updatePriorities()
	}
}

// updatePriorities recalculates piece priorities based on all active readers
func (pm *priorityManager) updatePriorities() {
	if pm.torrent == nil || pm.file == nil || pm.torrent.Info() == nil {
		return
	}

	t := pm.torrent
	file := pm.file
	pieceLength := t.Info().PieceLength

	if pieceLength == 0 {
		if pm.logger != nil {
			pm.logger.Warn().Msg("torrentutil: piece length is zero, cannot prioritize")
		}
		return
	}

	numTorrentPieces := int64(t.NumPieces())
	if numTorrentPieces == 0 {
		if pm.logger != nil {
			pm.logger.Warn().Msg("torrentutil: torrent has zero pieces, cannot prioritize")
		}
		return
	}

	// Calculate file piece range
	fileFirstPieceIdx := file.Offset() / pieceLength
	fileLastPieceIdx := (file.Offset() + file.Length() - 1) / pieceLength

	// Collect all needed piece ranges from all active readers
	neededPieces := make(map[int64]torrent.PiecePriority)

	for _, reader := range pm.readers {
		position := reader.position
		// Remove 1MB from the position (for subtitle cluster)
		position -= 1 * 1024 * 1024
		if position < 0 {
			position = 0
		}
		if position < 0 {
			position = 0
		}

		currentGlobalSeekPieceIdx := (file.Offset() + position) / pieceLength

		// Pieces needed NOW (immediate)
		for i := int64(0); i < piecesForNow; i++ {
			idx := currentGlobalSeekPieceIdx + i
			if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
				if current, exists := neededPieces[idx]; !exists || current < torrent.PiecePriorityNow {
					neededPieces[idx] = torrent.PiecePriorityNow
				}
			}
		}

		// Pieces needed HIGH (before current position for rewinds)
		for i := int64(1); i <= piecesForHighBefore; i++ {
			idx := currentGlobalSeekPieceIdx - i
			if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx >= 0 {
				if current, exists := neededPieces[idx]; !exists || current < torrent.PiecePriorityHigh {
					neededPieces[idx] = torrent.PiecePriorityHigh
				}
			}
		}

		// Pieces needed NEXT (immediate readahead)
		nextStartIdx := currentGlobalSeekPieceIdx + piecesForNow
		for i := int64(0); i < piecesForNext; i++ {
			idx := nextStartIdx + i
			if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
				if current, exists := neededPieces[idx]; !exists || current < torrent.PiecePriorityNext {
					neededPieces[idx] = torrent.PiecePriorityNext
				}
			}
		}

		// Pieces needed for READAHEAD (further readahead)
		readaheadStartIdx := nextStartIdx + piecesForNext
		for i := int64(0); i < piecesForReadahead; i++ {
			idx := readaheadStartIdx + i
			if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
				if current, exists := neededPieces[idx]; !exists || current < torrent.PiecePriorityReadahead {
					neededPieces[idx] = torrent.PiecePriorityReadahead
				}
			}
		}
	}

	// Reset pieces that are no longer needed by any reader
	isStartup := time.Since(pm.createdAt) < 60*time.Second
	headEndPieceIdx := fileFirstPieceIdx + (8 * 1024 * 1024 / pieceLength)
	tailStartPieceIdx := fileLastPieceIdx - (4 * 1024 * 1024 / pieceLength)

	for idx := fileFirstPieceIdx; idx <= fileLastPieceIdx; idx++ {
		if idx < 0 || idx >= numTorrentPieces {
			continue
		}

		piece := t.Piece(int(idx))
		currentPriority := piece.State().Priority

		if neededPriority, needed := neededPieces[idx]; needed {
			// Set to the highest priority needed by any reader
			if currentPriority != neededPriority {
				piece.SetPriority(neededPriority)
			}
		} else {
			// Don't downgrade head and tail pieces during startup if they are already higher than Normal
			if isStartup && (idx <= headEndPieceIdx || idx >= tailStartPieceIdx) {
				continue
			}
			// Only reset to normal if not completely unwanted and not already at highest priority
			if currentPriority != torrent.PiecePriorityNone && currentPriority != torrent.PiecePriorityNow {
				piece.SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}

	if pm.logger != nil {
		pm.logger.Debug().Msgf("torrentutil: Updated priorities for %d readers, %d pieces prioritized", len(pm.readers), len(neededPieces))
	}
}

// resetAllPriorities resets all file pieces to normal priority
func (pm *priorityManager) resetAllPriorities() {
	if pm.torrent == nil || pm.file == nil || pm.torrent.Info() == nil {
		return
	}

	t := pm.torrent
	file := pm.file
	pieceLen := t.Info().PieceLength

	if pieceLen == 0 {
		return
	}

	numTorrentPieces := int64(t.NumPieces())
	fileFirstPieceIdx := file.Offset() / pieceLen
	fileLastPieceIdx := (file.Offset() + file.Length() - 1) / pieceLen

	for idx := fileFirstPieceIdx; idx <= fileLastPieceIdx; idx++ {
		if idx >= 0 && idx < numTorrentPieces {
			piece := t.Piece(int(idx))
			if piece.State().Priority != torrent.PiecePriorityNone {
				piece.SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}
}

// cleanupStaleReaders removes readers that haven't been accessed recently.
// It returns true when the manager no longer tracks any readers.
func (pm *priorityManager) cleanupStaleReaders(now time.Time) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cutoff := now.Add(-2 * time.Minute)
	for id, reader := range pm.readers {
		if reader.lastAccess.Before(cutoff) {
			delete(pm.readers, id)
			if pm.logger != nil {
				pm.logger.Debug().Msgf("torrentutil: Cleaned up stale reader %s", id)
			}
		}
	}

	if len(pm.readers) == 0 {
		return true
	}

	pm.updatePriorities()
	return false
}

func (pm *priorityManager) readerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.readers)
}

// ReadSeeker implements io.ReadSeekCloser for a torrent file being streamed.
// It allows dynamic prioritization of pieces when seeking, optimized for streaming
// and supports multiple concurrent readers on the same file.
type ReadSeeker struct {
	id              string
	torrent         *torrent.Torrent
	file            *torrent.File
	reader          torrent.Reader
	priorityManager *priorityManager
	logger          *zerolog.Logger
	position        atomic.Int64
	lastPriorityPos atomic.Int64
	updateStride    int64
}

var _ io.ReadSeekCloser = &ReadSeeker{}

func NewReadSeeker(t *torrent.Torrent, file *torrent.File, logger ...*zerolog.Logger) io.ReadSeekCloser {
	tr := file.NewReader()
	tr.SetResponsive()
	// Read ahead 5MB for better streaming performance
	// DEVNOTE: Not sure if dynamic prioritization overwrites this but whatever
	tr.SetReadahead(5 * 1024 * 1024)

	var loggerPtr *zerolog.Logger
	if len(logger) > 0 {
		loggerPtr = logger[0]
	}

	if loggerPtr == nil {
		loggerPtr = new(zerolog.Nop())
	}

	pm := getPriorityManager(t, file, loggerPtr)

	rs := &ReadSeeker{
		id:              fmt.Sprintf("reader_%d_%d", time.Now().UnixNano(), len(pm.readers)),
		torrent:         t,
		file:            file,
		reader:          tr,
		priorityManager: pm,
		logger:          loggerPtr,
		updateStride:    getPriorityUpdateStride(t),
	}

	// Register this reader with the priority manager
	pm.registerReader(rs.id, 0)

	return rs
}

func (rs *ReadSeeker) Read(p []byte) (n int, err error) {
	n, err = rs.reader.Read(p)
	if n > 0 {
		newOffset := rs.position.Add(int64(n))
		if rs.shouldRefreshPriority(newOffset) {
			rs.lastPriorityPos.Store(newOffset)
			rs.priorityManager.updateReaderPosition(rs.id, newOffset)
		} else {
			rs.priorityManager.touchReader(rs.id, newOffset)
		}
	}
	return n, err
}

func (rs *ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := rs.reader.Seek(offset, whence)
	if err != nil {
		if rs.logger != nil {
			rs.logger.Error().Err(err).Int64("offset", offset).Int("whence", whence).Msg("torrentutil: ReadSeeker seek error")
		}
		return newOffset, err
	}

	rs.position.Store(newOffset)
	rs.lastPriorityPos.Store(newOffset)
	// Update this reader's position in the priority manager
	rs.priorityManager.updateReaderPosition(rs.id, newOffset)

	return newOffset, nil
}

func (rs *ReadSeeker) shouldRefreshPriority(offset int64) bool {
	lastOffset := rs.lastPriorityPos.Load()
	if offset < lastOffset {
		return true
	}
	return offset-lastOffset >= rs.updateStride
}

func getPriorityUpdateStride(t *torrent.Torrent) int64 {
	if t == nil || t.Info() == nil || t.Info().PieceLength <= 0 {
		return defaultUpdateStride
	}

	stride := int64(t.Info().PieceLength) / 2
	if stride < defaultUpdateStride {
		return defaultUpdateStride
	}

	return stride
}

// Close closes the underlying torrent file reader and unregisters from priority manager.
// This makes ReadSeeker implement io.ReadSeekCloser.
func (rs *ReadSeeker) Close() error {
	// Unregister from priority manager
	rs.priorityManager.unregisterReader(rs.id)

	if rs.reader != nil {
		return rs.reader.Close()
	}
	return nil
}

// PrioritizeDownloadPieces sets piece priorities for the initial download windows.
// First 8 MiB: Immediate (torrent.PiecePriorityNow)
// Next 24 MiB: Readahead (torrent.PiecePriorityReadahead)
// Last 4 MiB: High (torrent.PiecePriorityHigh)
func PrioritizeDownloadPieces(t *torrent.Torrent, file *torrent.File, logger *zerolog.Logger) {
	if t == nil || file == nil || t.Info() == nil {
		return
	}
	pieceLength := t.Info().PieceLength
	if pieceLength <= 0 {
		return
	}

	fileOffset := file.Offset()
	fileLength := file.Length()
	numTorrentPieces := int64(t.NumPieces())

	firstPieceIdx := fileOffset / pieceLength
	endPieceIdx := (fileOffset + fileLength - 1) / pieceLength

	getPieceIdx := func(offset int64) int64 {
		if offset < 0 {
			offset = 0
		}
		if offset > fileLength {
			offset = fileLength
		}
		return (fileOffset + offset) / pieceLength
	}

	immediateEndIdx := getPieceIdx(8 * 1024 * 1024)
	readaheadEndIdx := getPieceIdx((8 + 24) * 1024 * 1024)
	finalStartIdx := getPieceIdx(fileLength - 4*1024*1024)

	if logger != nil {
		logger.Debug().Msgf("torrentutil: Prioritizing pieces for file %s. Immediate: [%d-%d], Readahead: [%d-%d], Final: [%d-%d]",
			file.DisplayPath(), firstPieceIdx, immediateEndIdx, immediateEndIdx+1, readaheadEndIdx, finalStartIdx, endPieceIdx)
	}

	for idx := firstPieceIdx; idx <= endPieceIdx; idx++ {
		if idx >= 0 && idx < numTorrentPieces {
			piece := t.Piece(int(idx))
			if idx <= immediateEndIdx {
				piece.SetPriority(torrent.PiecePriorityNow)
			} else if idx <= readaheadEndIdx {
				piece.SetPriority(torrent.PiecePriorityReadahead)
			} else if idx >= finalStartIdx {
				piece.SetPriority(torrent.PiecePriorityHigh)
			} else {
				piece.SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}
}

// PrioritizeRangeRequestPieces attempts to prioritize pieces needed for the range request.
func PrioritizeRangeRequestPieces(rangeHeader string, t *torrent.Torrent, file *torrent.File, logger *zerolog.Logger) {
	if t == nil || file == nil || t.Info() == nil {
		return
	}
	pieceLen := t.Info().PieceLength
	if pieceLen <= 0 {
		return
	}

	ranges, err := httputil.ParseRange(rangeHeader, file.Length())
	if err != nil || len(ranges) == 0 {
		return
	}
	start := ranges[0].Start

	// Calculate file's pieces range
	fileOffset := file.Offset()
	fileLength := file.Length()
	numTorrentPieces := int64(t.NumPieces())

	firstFilePieceIdx := fileOffset / pieceLen
	endFilePieceIdx := (fileOffset + fileLength - 1) / pieceLen

	// Calculate the piece index for this seek offset with small padding
	// Subtract a small amount to ensure we don't miss the beginning of a needed piece
	seekPosition := start
	if seekPosition >= 1024*1024 { // If we're at least 1MB in, add some padding
		seekPosition -= 1024 * 512
	}
	seekPieceIdx := (fileOffset + seekPosition) / pieceLen

	// Prioritize the next several pieces from this point
	numPiecesToPrioritize := int64(10)

	if seekPieceIdx+numPiecesToPrioritize > endFilePieceIdx {
		numPiecesToPrioritize = endFilePieceIdx - seekPieceIdx + 1
	}
	if numPiecesToPrioritize <= 0 {
		numPiecesToPrioritize = 1
	}

	if logger != nil {
		logger.Debug().Msgf("torrentutil: Prioritizing range request pieces %d to %d",
			seekPieceIdx, seekPieceIdx+numPiecesToPrioritize-1)
	}

	// Set normal priority for pieces far from current position
	for idx := firstFilePieceIdx; idx <= endFilePieceIdx; idx++ {
		if idx >= 0 && idx < numTorrentPieces {
			// Don't touch the beginning pieces and don't touch pieces near the seek point
			if idx > firstFilePieceIdx+100 && idx < seekPieceIdx-100 ||
				idx > seekPieceIdx+numPiecesToPrioritize+100 {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}

	// Set the highest priority for the pieces we need right now
	for idx := seekPieceIdx; idx < seekPieceIdx+numPiecesToPrioritize; idx++ {
		if idx >= 0 && idx < numTorrentPieces {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
	}

	// Also prioritize a small buffer before the seek point to handle small rewinds
	bufferBeforeCount := int64(5)
	if seekPieceIdx > firstFilePieceIdx+bufferBeforeCount {
		for idx := seekPieceIdx - bufferBeforeCount; idx < seekPieceIdx; idx++ {
			if idx >= 0 && idx < numTorrentPieces {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityHigh)
			}
		}
	}

	nextReadStart := seekPieceIdx + numPiecesToPrioritize
	nextReadCount := int64(100) // 100 additional pieces for nextRead
	if nextReadStart+nextReadCount > endFilePieceIdx {
		nextReadCount = endFilePieceIdx - nextReadStart + 1
	}

	if nextReadCount > 0 {
		if logger != nil {
			logger.Debug().Msgf("torrentutil: Setting next priority for pieces %d to %d",
				nextReadStart, nextReadStart+nextReadCount-1)
		}
		for idx := nextReadStart; idx < nextReadStart+nextReadCount; idx++ {
			if idx >= 0 && idx < numTorrentPieces {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNext)
			}
		}
	}

	readAheadCount := int64(100)
	if nextReadStart+readAheadCount > endFilePieceIdx {
		readAheadCount = endFilePieceIdx - nextReadStart + 1
	}

	if readAheadCount > 0 {
		if logger != nil {
			logger.Debug().Msgf("torrentutil: Setting read ahead priority for pieces %d to %d",
				nextReadStart, nextReadStart+readAheadCount-1)
		}
		for idx := nextReadStart; idx < nextReadStart+readAheadCount; idx++ {
			if idx >= 0 && idx < numTorrentPieces {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityReadahead)
			}
		}
	}
}
