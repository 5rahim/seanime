package torrentutil

import (
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
)

// +-----------------------+
// +   anacrolix/torrent   +
// +-----------------------+

// PrioritizeDownloadPieces sets sets high priority for the first 3% of pieces and the last few pieces to ensure faster loading.
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
