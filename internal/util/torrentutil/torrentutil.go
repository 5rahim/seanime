package torrentutil

import (
	"fmt"
	"io"
	"sync"
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
)

// readerInfo tracks information about an active reader
type readerInfo struct {
	id         string
	position   int64
	lastAccess time.Time
}

// priorityManager manages piece priorities for multiple readers on the same file
type priorityManager struct {
	mu      sync.RWMutex
	readers map[string]*readerInfo
	torrent *torrent.Torrent
	file    *torrent.File
	logger  *zerolog.Logger
}

// global map to track priority managers per torrent+file combination
var (
	priorityManagers   = make(map[string]*priorityManager)
	priorityManagersMu sync.RWMutex
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
		readers: make(map[string]*readerInfo),
		torrent: t,
		file:    file,
		logger:  logger,
	}
	priorityManagers[key] = pm

	// Start cleanup goroutine for the first manager
	if len(priorityManagers) == 1 {
		go pm.cleanupStaleReaders()
	}

	return pm
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
	pieceLength := t.Info().PieceLength

	if pieceLength == 0 {
		return
	}

	numTorrentPieces := int64(t.NumPieces())
	fileFirstPieceIdx := file.Offset() / pieceLength
	fileLastPieceIdx := (file.Offset() + file.Length() - 1) / pieceLength

	for idx := fileFirstPieceIdx; idx <= fileLastPieceIdx; idx++ {
		if idx >= 0 && idx < numTorrentPieces {
			piece := t.Piece(int(idx))
			if piece.State().Priority != torrent.PiecePriorityNone {
				piece.SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}
}

// cleanupStaleReaders periodically removes readers that haven't been accessed recently
func (pm *priorityManager) cleanupStaleReaders() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pm.mu.Lock()
		cutoff := time.Now().Add(-2 * time.Minute)

		for id, reader := range pm.readers {
			if reader.lastAccess.Before(cutoff) {
				delete(pm.readers, id)
				if pm.logger != nil {
					pm.logger.Debug().Msgf("torrentutil: Cleaned up stale reader %s", id)
				}
			}
		}

		// Update priorities after cleanup
		if len(pm.readers) > 0 {
			pm.updatePriorities()
		}

		pm.mu.Unlock()
	}
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

	pm := getPriorityManager(t, file, loggerPtr)

	rs := &ReadSeeker{
		id:              fmt.Sprintf("reader_%d_%d", time.Now().UnixNano(), len(pm.readers)),
		torrent:         t,
		file:            file,
		reader:          tr,
		priorityManager: pm,
		logger:          loggerPtr,
	}

	// Register this reader with the priority manager
	pm.registerReader(rs.id, 0)

	return rs
}

func (rs *ReadSeeker) Read(p []byte) (n int, err error) {
	return rs.reader.Read(p)
}

func (rs *ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := rs.reader.Seek(offset, whence)
	if err != nil {
		if rs.logger != nil {
			rs.logger.Error().Err(err).Int64("offset", offset).Int("whence", whence).Msg("torrentutil: ReadSeeker seek error")
		}
		return newOffset, err
	}

	// Update this reader's position in the priority manager
	rs.priorityManager.updateReaderPosition(rs.id, newOffset)

	return newOffset, nil
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

// PrioritizeDownloadPieces sets high priority for the first 3% of pieces and the last few pieces to ensure faster loading.
func PrioritizeDownloadPieces(t *torrent.Torrent, file *torrent.File, logger *zerolog.Logger) {
	// Calculate file's pieces
	firstPieceIdx := file.Offset() * int64(t.NumPieces()) / t.Length()
	endPieceIdx := (file.Offset() + file.Length()) * int64(t.NumPieces()) / t.Length()

	// Prioritize more pieces at the beginning for faster initial loading (3% for beginning)
	numPiecesForStart := (endPieceIdx - firstPieceIdx + 1) * 3 / 100
	if logger != nil {
		logger.Debug().Msgf("torrentuil: Setting high priority for first 3%% - pieces %d to %d (total %d)",
			firstPieceIdx, firstPieceIdx+numPiecesForStart, numPiecesForStart)
	}
	for idx := firstPieceIdx; idx <= firstPieceIdx+numPiecesForStart; idx++ {
		t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
	}

	// Also prioritize the last few pieces
	numPiecesForEnd := (endPieceIdx - firstPieceIdx + 1) * 1 / 100
	if logger != nil {
		logger.Debug().Msgf("torrentuil: Setting priority for last pieces %d to %d (total %d)",
			endPieceIdx-numPiecesForEnd, endPieceIdx, numPiecesForEnd)
	}
	for idx := endPieceIdx - numPiecesForEnd; idx <= endPieceIdx; idx++ {
		if idx >= 0 && int(idx) < t.NumPieces() {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
	}
}

// PrioritizeRangeRequestPieces attempts to prioritize pieces needed for the range request.
func PrioritizeRangeRequestPieces(rangeHeader string, t *torrent.Torrent, file *torrent.File, logger *zerolog.Logger) {
	// Parse the range header (format: bytes=START-END)
	var start int64
	_, _ = fmt.Sscanf(rangeHeader, "bytes=%d-", &start)

	if start >= 0 {
		// Calculate file's pieces range
		fileOffset := file.Offset()
		fileLength := file.Length()

		// Calculate the total range of pieces for this file
		firstFilePieceIdx := fileOffset * int64(t.NumPieces()) / t.Length()
		endFilePieceIdx := (fileOffset + fileLength) * int64(t.NumPieces()) / t.Length()

		// Calculate the piece index for this seek offset with small padding
		// Subtract a small amount to ensure we don't miss the beginning of a needed piece
		seekPosition := start
		if seekPosition >= 1024*1024 { // If we're at least 1MB in, add some padding
			seekPosition -= 1024 * 512 // Subtract 512KB to ensure we get the right piece
		}
		seekPieceIdx := (fileOffset + seekPosition) * int64(t.NumPieces()) / t.Length()

		// Prioritize the next several pieces from this point
		// This is especially important for seeking
		numPiecesToPrioritize := int64(10) // Prioritize next 10 pieces, adjust as needed

		if seekPieceIdx+numPiecesToPrioritize > endFilePieceIdx {
			numPiecesToPrioritize = endFilePieceIdx - seekPieceIdx
		}

		if logger != nil {
			logger.Debug().Msgf("torrentutil: Prioritizing range request pieces %d to %d",
				seekPieceIdx, seekPieceIdx+numPiecesToPrioritize)
		}

		// Set normal priority for pieces far from our current position
		// This allows background downloading while still prioritizing the seek point
		for idx := firstFilePieceIdx; idx <= endFilePieceIdx; idx++ {
			if idx >= 0 && int(idx) < t.NumPieces() {
				// Don't touch the beginning pieces which should maintain their high priority
				// for the next potential restart, and don't touch pieces near our seek point
				if idx > firstFilePieceIdx+100 && idx < seekPieceIdx-100 ||
					idx > seekPieceIdx+numPiecesToPrioritize+100 {
					// Set to normal priority - allow background downloading
					t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNormal)
				}
			}
		}

		// Now set the highest priority for the pieces we need right now
		for idx := seekPieceIdx; idx < seekPieceIdx+numPiecesToPrioritize; idx++ {
			if idx >= 0 && int(idx) < t.NumPieces() {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
			}
		}

		// Also prioritize a small buffer before the seek point to handle small rewinds
		// This is useful for MPV's default rewind behavior
		bufferBeforeCount := int64(5) // 5 pieces buffer before seek point
		if seekPieceIdx > firstFilePieceIdx+bufferBeforeCount {
			for idx := seekPieceIdx - bufferBeforeCount; idx < seekPieceIdx; idx++ {
				if idx >= 0 && int(idx) < t.NumPieces() {
					t.Piece(int(idx)).SetPriority(torrent.PiecePriorityHigh)
				}
			}
		}

		// Also prioritize the next readahead segment after our immediate needs
		// This helps prepare for continued playback
		nextReadStart := seekPieceIdx + numPiecesToPrioritize
		nextReadCount := int64(100) // 100 additional pieces for nextRead
		if nextReadStart+nextReadCount > endFilePieceIdx {
			nextReadCount = endFilePieceIdx - nextReadStart
		}

		if nextReadCount > 0 {
			if logger != nil {
				logger.Debug().Msgf("torrentutil: Setting next priority for pieces %d to %d",
					nextReadStart, nextReadStart+nextReadCount)
			}
			for idx := nextReadStart; idx < nextReadStart+nextReadCount; idx++ {
				if idx >= 0 && int(idx) < t.NumPieces() {
					t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNext)
				}
			}
		}

		// Also prioritize the next readahead segment after our immediate needs
		// This helps prepare for continued playback
		readAheadCount := int64(100)
		if nextReadStart+readAheadCount > endFilePieceIdx {
			readAheadCount = endFilePieceIdx - nextReadStart
		}

		if readAheadCount > 0 {
			if logger != nil {
				logger.Debug().Msgf("torrentutil: Setting read ahead priority for pieces %d to %d",
					nextReadStart, nextReadStart+readAheadCount)
			}
			for idx := nextReadStart; idx < nextReadStart+readAheadCount; idx++ {
				if idx >= 0 && int(idx) < t.NumPieces() {
					t.Piece(int(idx)).SetPriority(torrent.PiecePriorityReadahead)
				}
			}
		}
	}
}
