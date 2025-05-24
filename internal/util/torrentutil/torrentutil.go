package torrentutil

import (
	"fmt"
	"io"

	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
)

// +-----------------------+
// +   anacrolix/torrent   +
// +-----------------------+

const (
	piecesForNow        = int64(10)
	piecesForHighBefore = int64(5)
	piecesForNext       = int64(100)
	piecesForReadahead  = int64(100)
)

// ReadSeeker implements io.ReadSeekCloser for a torrent file being streamed.
// It allows dynamic prioritization of pieces when seeking, optimized for streaming.
type ReadSeeker struct {
	torrent *torrent.Torrent
	file    *torrent.File
	reader  torrent.Reader
	logger  *zerolog.Logger
}

var _ io.ReadSeekCloser = &ReadSeeker{}

func NewReadSeeker(t *torrent.Torrent, file *torrent.File, logger ...*zerolog.Logger) io.ReadSeekCloser {
	tr := file.NewReader()
	rs := &ReadSeeker{
		torrent: t,
		file:    file,
		reader:  tr,
	}
	if len(logger) > 0 {
		rs.logger = logger[0]
	}
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

	rs.prioritizeForOffset(newOffset)

	return newOffset, nil
}

// prioritizeForOffset adjusts piece priorities around a given offset within the file,
// making it more aggressive for streaming.
func (rs *ReadSeeker) prioritizeForOffset(seekOffsetInFile int64) {
	if rs.torrent == nil || rs.file == nil || rs.torrent.Info() == nil {
		return
	}

	t := rs.torrent
	file := rs.file

	pieceLength := t.Info().PieceLength
	if pieceLength == 0 { // Avoid division by zero if piece length is unknown or zero
		if rs.logger != nil {
			rs.logger.Warn().Msg("torrentutil: ReadSeeker piece length is zero, cannot prioritize")
		}
		return
	}

	numTorrentPieces := int64(t.NumPieces())
	if numTorrentPieces == 0 {
		if rs.logger != nil {
			rs.logger.Warn().Msg("torrentutil: ReadSeeker torrent has zero pieces, cannot prioritize")
		}
		return
	}

	// Calculate the global torrent piece indices for the start and end of the current file.
	fileFirstPieceIdx := file.Offset() / pieceLength
	fileLastPieceIdx := (file.Offset() + file.Length() - 1) / pieceLength

	// Calculate the global torrent piece index for the current seek position.
	currentGlobalSeekPieceIdx := (file.Offset() + seekOffsetInFile) / pieceLength

	if rs.logger != nil {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker adjusting priorities. Seek in file to offset %d (global piece %d). File spans global pieces %d-%d.",
			seekOffsetInFile, currentGlobalSeekPieceIdx, fileFirstPieceIdx, fileLastPieceIdx)
	}

	// Reset priorities for pieces within the file but far from the new immediate interest zone.
	resetFarBehindThreshold := currentGlobalSeekPieceIdx - piecesForHighBefore - 30
	resetFarAheadThreshold := currentGlobalSeekPieceIdx + piecesForNow + piecesForNext + piecesForReadahead + 30

	for idx := fileFirstPieceIdx; idx <= fileLastPieceIdx; idx++ {
		if idx < 0 || idx >= numTorrentPieces {
			continue
		}
		isFarBehind := idx < resetFarBehindThreshold
		isFarAhead := idx > resetFarAheadThreshold

		if isFarBehind || isFarAhead {
			// Only set to normal if not already None (completely unwanted)
			if t.Piece(int(idx)).State().Priority != torrent.PiecePriorityNone {
				t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNormal)
			}
		}
	}
	if rs.logger != nil {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker reset distant pieces (far_behind_threshold: %d, far_ahead_threshold: %d)", resetFarBehindThreshold, resetFarAheadThreshold)
	}

	// Prioritize pieces immediately needed
	prioritizedNowCount := 0
	for i := int64(0); i < piecesForNow; i++ {
		idx := currentGlobalSeekPieceIdx + i
		if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
			prioritizedNowCount++
		}
	}
	if rs.logger != nil && prioritizedNowCount > 0 {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker set %d pieces to NOW around global piece %d", prioritizedNowCount, currentGlobalSeekPieceIdx)
	}

	// Prioritize "High" before the current piece for small rewinds
	prioritizedHighBeforeCount := 0
	for i := int64(1); i <= piecesForHighBefore; i++ {
		idx := currentGlobalSeekPieceIdx - i
		if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx >= 0 {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityHigh)
			prioritizedHighBeforeCount++
		}
	}
	if rs.logger != nil && prioritizedHighBeforeCount > 0 {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker set %d pieces to HIGH BEFORE global piece %d", prioritizedHighBeforeCount, currentGlobalSeekPieceIdx)
	}

	// Prioritize next pieces for immediate readahead
	prioritizedNextCount := 0
	nextStartIdx := currentGlobalSeekPieceIdx + piecesForNow
	for i := int64(0); i < piecesForNext; i++ {
		idx := nextStartIdx + i
		if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNext)
			prioritizedNextCount++
		}
	}
	if rs.logger != nil && prioritizedNextCount > 0 {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker set %d pieces to NEXT from global piece %d", prioritizedNextCount, nextStartIdx)
	}

	// Prioritize readahead pieces for further readahead
	prioritizedReadaheadCount := 0
	readaheadStartIdx := nextStartIdx + piecesForNext
	for i := int64(0); i < piecesForReadahead; i++ {
		idx := readaheadStartIdx + i
		if idx >= fileFirstPieceIdx && idx <= fileLastPieceIdx && idx < numTorrentPieces {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityReadahead)
			prioritizedReadaheadCount++
		}
	}
	if rs.logger != nil && prioritizedReadaheadCount > 0 {
		rs.logger.Debug().Msgf("torrentutil: ReadSeeker set %d pieces to REAHEAD from global piece %d", prioritizedReadaheadCount, readaheadStartIdx)
	}
}

// Close closes the underlying torrent file reader.
// This makes ReadSeeker implement io.ReadSeekCloser.
func (rs *ReadSeeker) Close() error {
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
