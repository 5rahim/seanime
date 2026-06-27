package torrentstream

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/util/torrentutil"
	"strconv"
	"time"
)

var _ = http.Handler(&handler{})

type (
	// handler serves the torrent stream
	handler struct {
		repository *Repository
	}
)

func newHandler(repository *Repository) *handler {
	return &handler{
		repository: repository,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, found := h.repository.client.currentFile.Get()
	if !found || h.repository.client.currentTorrent.IsAbsent() {
		h.repository.logger.Error().Msg("torrentstream: No torrent to stream")
		http.Error(w, "No torrent to stream", http.StatusNotFound)
		return
	}

	h.repository.logger.Debug().
		Str("range", r.Header.Get("Range")).
		Int64("bytesCompleted", file.BytesCompleted()).
		Int64("fileSize", file.Length()).
		Msg("torrentstream: Stream endpoint hit")

	if r.Method == http.MethodHead {
		length := file.Length()
		filePath := file.DisplayPath()
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
		w.Header().Set("Content-Disposition", "inline; filename="+filePath)
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		return
	}

	var tr io.ReadSeekCloser
	useCompletedFile := false
	downloadDir := h.repository.GetDownloadDir()
	torrent := h.repository.client.currentTorrent.MustGet()
	if downloadDir != "" && torrent != nil && file.Length() > 0 {
		completedPath := filepath.Join(downloadDir, torrent.InfoHash().HexString(), filepath.FromSlash(file.Path()))
		if info, err := os.Stat(completedPath); err == nil && !info.IsDir() && info.Size() == file.Length() {
			if reader, err := os.Open(completedPath); err == nil {
				tr = reader
				useCompletedFile = true
				h.repository.logger.Debug().Str("path", completedPath).Msg("torrentstream: Using completed file from disk")
			}
		}
	}

	if !useCompletedFile {
		h.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: New reader")
		tr = torrentutil.NewReadSeeker(torrent, file, h.repository.logger)

		// If this is a range request for a later part of the file, prioritize those pieces initially
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Attempt to prioritize the pieces requested in the range
			torrentutil.PrioritizeRangeRequestPieces(rangeHeader, torrent, file, h.repository.logger)
		}
	}

	defer func() {
		h.repository.logger.Trace().Msg("torrentstream: Closing reader")
		_ = tr.Close()
	}()

	h.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: Serving file content")
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeContent(
		w,
		r,
		file.DisplayPath(),
		time.Now(),
		tr,
	)
	h.repository.logger.Trace().Msg("torrentstream: File content served")
}
