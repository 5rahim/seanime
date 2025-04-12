package torrentstream

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anacrolix/torrent"
)

type (
	// serverManager manages the streaming server
	serverManager struct {
		repository    *Repository
		lastUsed      time.Time // Used to track the last time the server was used
		serverRunning bool      // Whether the server is running
	}
)

//// ref: torrserver
//func dnsResolve() {
//	addrs, _ := net.LookupHost("www.google.com")
//	if len(addrs) == 0 {
//		//fmt.Println("Check dns failed", addrs, err)
//
//		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
//			d := net.Dialer{}
//			return d.DialContext(ctx, "udp", "1.1.1.1:53")
//		}
//
//		net.DefaultResolver = &net.Resolver{
//			Dial: fn,
//		}
//
//		addrs, _ = net.LookupHost("www.google.com")
//		//fmt.Println("Check cloudflare dns", addrs, err)
//	} else {
//		//fmt.Println("Check dns OK", addrs, err)
//	}
//}

// newServerManager is called once during the lifetime of the application.
func newServerManager(repository *Repository) *serverManager {
	ret := &serverManager{
		repository: repository,
	}

	//dnsResolve()

	return ret
}

func (s *serverManager) initializeServer() {
	// no-op
}

func (s *serverManager) createServer() {
	// no-op
}

func (s *serverManager) startServer() {
	// no-op
}

// stopServer stops the streaming server.
func (s *serverManager) stopServer() {
	// no-op
}

func (s *serverManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.lastUsed = time.Now()
	s.repository.logger.Trace().Str("range", r.Header.Get("Range")).Msg("torrentstream: Stream endpoint hit")

	if s.repository.client.currentFile.IsAbsent() || s.repository.client.currentTorrent.IsAbsent() {
		s.repository.logger.Error().Msg("torrentstream: No torrent to stream")
		http.Error(w, "No torrent to stream", http.StatusNotFound)
		return
	}

	file := s.repository.client.currentFile.MustGet()
	s.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: New reader")
	tr := file.NewReader()
	defer func(tr torrent.Reader) {
		s.repository.logger.Trace().Msg("torrentstream: Closing reader")
		_ = tr.Close()
	}(tr)

	tr.SetResponsive()
	// Read ahead 5MB for better streaming performance
	// DEVNOTE: Not sure if dynamic prioritization overwrites this but whatever
	tr.SetReadahead(5 * 1024 * 1024)

	// If this is a range request for a later part of the file, prioritize those pieces
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" && s.repository.client.currentTorrent.IsPresent() {
		t := s.repository.client.currentTorrent.MustGet()
		// Attempt to prioritize the pieces requested in the range
		s.prioritizeRangeRequestPieces(rangeHeader, file, t)
	}

	s.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: Serving file content")
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeContent(
		w,
		r,
		file.DisplayPath(),
		time.Now(),
		tr,
	)
	s.repository.logger.Trace().Msg("torrentstream: File content served")
}

// prioritizeRangeRequestPieces attempts to prioritize pieces needed for the range request
func (s *serverManager) prioritizeRangeRequestPieces(rangeHeader string, file *torrent.File, t *torrent.Torrent) {
	// Parse the range header (format: bytes=START-END)
	var start int64
	fmt.Sscanf(rangeHeader, "bytes=%d-", &start)

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

		s.repository.logger.Debug().Msgf("torrentstream: Prioritizing range request pieces %d to %d",
			seekPieceIdx, seekPieceIdx+numPiecesToPrioritize)

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
			s.repository.logger.Debug().Msgf("torrentstream: Setting next priority for pieces %d to %d",
				nextReadStart, nextReadStart+nextReadCount)
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
			s.repository.logger.Debug().Msgf("torrentstream: Setting read ahead priority for pieces %d to %d",
				nextReadStart, nextReadStart+readAheadCount)
			for idx := nextReadStart; idx < nextReadStart+readAheadCount; idx++ {
				if idx >= 0 && int(idx) < t.NumPieces() {
					t.Piece(int(idx)).SetPriority(torrent.PiecePriorityReadahead)
				}
			}
		}
	}
}
