package torrentstream

import (
	"net/http"
	"seanime/internal/util/torrentutil"
	"strconv"
	"time"

	"github.com/anacrolix/torrent"
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
	h.repository.logger.Trace().Str("range", r.Header.Get("Range")).Msg("torrentstream: Stream endpoint hit")

	if h.repository.client.currentFile.IsAbsent() || h.repository.client.currentTorrent.IsAbsent() {
		h.repository.logger.Error().Msg("torrentstream: No torrent to stream")
		http.Error(w, "No torrent to stream", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodHead {
		r.Response.Header.Set("Content-Type", "video/mp4")
		r.Response.Header.Set("Content-Length", strconv.Itoa(int(h.repository.client.currentFile.MustGet().Length())))
		r.Response.Header.Set("Content-Disposition", "inline; filename="+h.repository.client.currentFile.MustGet().DisplayPath())
		r.Response.Header.Set("Accept-Ranges", "bytes")
		r.Response.Header.Set("Cache-Control", "no-cache")
		r.Response.Header.Set("Pragma", "no-cache")
		r.Response.Header.Set("Expires", "0")
		r.Response.Header.Set("X-Content-Type-Options", "nosniff")

		// No content, just headers
		w.WriteHeader(http.StatusOK)
		return
	}

	file := h.repository.client.currentFile.MustGet()
	h.repository.logger.Trace().Str("file", file.DisplayPath()).Msg("torrentstream: New reader")
	tr := file.NewReader()
	defer func(tr torrent.Reader) {
		h.repository.logger.Trace().Msg("torrentstream: Closing reader")
		_ = tr.Close()
	}(tr)

	tr.SetResponsive()
	// Read ahead 5MB for better streaming performance
	// DEVNOTE: Not sure if dynamic prioritization overwrites this but whatever
	tr.SetReadahead(5 * 1024 * 1024)

	// If this is a range request for a later part of the file, prioritize those pieces
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" && h.repository.client.currentTorrent.IsPresent() {
		t := h.repository.client.currentTorrent.MustGet()
		// Attempt to prioritize the pieces requested in the range
		torrentutil.PrioritizeRangeRequestPieces(rangeHeader, t, file, h.repository.logger)
	}

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
