package directstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	httputil "seanime/internal/util/http"
	"time"

	"github.com/neilotoole/streamcache"
)

func ServeLocalFile(w http.ResponseWriter, r *http.Request, lfStream *LocalFileStream) {
	if lfStream.serveContentCancelFunc != nil {
		lfStream.serveContentCancelFunc()
	}

	ct, cancel := context.WithCancel(lfStream.manager.playbackCtx)
	lfStream.serveContentCancelFunc = cancel

	reader, err := lfStream.newReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	playbackInfo, err := lfStream.LoadPlaybackInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	size := playbackInfo.ContentLength
	w.Header().Set("Content-Length", fmt.Sprint(size))

	// No Range header → let Go handle it
	rangeHdr := r.Header.Get("Range")
	if rangeHdr == "" {
		http.ServeContent(w, r, lfStream.localFile.Path, time.Now(), reader)
		return
	}

	// Parse the range header
	ranges, err := httputil.ParseRange(rangeHdr, size)
	if err != nil && !errors.Is(err, httputil.ErrNoOverlap) {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
		return
	} else if err != nil && errors.Is(err, httputil.ErrNoOverlap) {
		// Let Go handle overlap
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.ServeContent(w, r, lfStream.localFile.Path, time.Now(), reader)
		return
	}

	// If we have a range, stream subtitles
	// if len(ranges) > 0 {
	// 	lfStream.ServeSubtitles(ranges[0].Start)
	// 	lfStream.logger.Trace().Msgf("directstream > Serving subtitles for range %s", ranges[0].ContentRange(size))
	// }

	serveContentRange(w, r, ct, reader, lfStream.localFile.Path, size, playbackInfo.MimeType, ranges)
}

func serveContentRange(w http.ResponseWriter, r *http.Request, ctx context.Context, reader io.ReadSeekCloser, name string, size int64, contentType string, ranges []httputil.Range) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Connection", "keep-alive") // Explicitly use keep-alive
	w.Header().Set("Cache-Control", "no-store")

	// Only handle the first range for now (multiples are rare)
	ra := ranges[0]

	// Validate range
	if ra.Start >= size || ra.Start < 0 || ra.Length <= 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Set response headers for partial content
	w.Header().Set("Content-Range", ra.ContentRange(size))
	w.Header().Set("Content-Length", fmt.Sprint(ra.Length))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to the requested position
	_, err := reader.Seek(ra.Start, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	written, err := copyWithContext(ctx, w, reader, ra.Length)

	if err != nil && err != io.EOF && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		//log.Error().Msgf("ERR - directstream > Error copying data: %v (wrote %d of %d bytes)",
		//	err, written, ra.Length)
		_ = written
	}
}

// copyWithContext copies n bytes from src to dst, respecting context cancellation
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, n int64) (int64, error) {
	// Use a reasonably sized buffer
	buf := make([]byte, 32*1024) // 32KB buffer

	var written int64
	for written < n {
		// Check if context is done before each read
		select {
		case <-ctx.Done():
			fmt.Println("directstream > Context done")
			return written, ctx.Err()
		default:
		}

		// Calculate how much to read this iteration
		toRead := int64(len(buf))
		if n-written < toRead {
			toRead = n - written
		}

		// Read from source
		nr, readErr := io.LimitReader(src, toRead).Read(buf)
		if nr > 0 {
			// Write to destination
			nw, writeErr := dst.Write(buf[:nr])
			if nw < nr {
				return written + int64(nw), writeErr
			}
			written += int64(nr)

			// Handle write error
			if writeErr != nil {
				return written, writeErr
			}
		}

		// Handle read error or EOF
		if readErr != nil {
			if readErr == io.EOF {
				if written >= n {
					return written, nil // Successfully read everything requested
				}
			}
			return written, readErr
		}
	}

	return written, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////

type StreamCacheReadSeekCloser struct {
	stream         *streamcache.Stream
	streamReader   *streamcache.Reader
	originalReader io.ReadSeekCloser
}

var _ io.ReadSeekCloser = (*StreamCacheReadSeekCloser)(nil)

func NewStreamCacheReadSeekCloser(ctx context.Context, reader io.ReadSeekCloser) StreamCacheReadSeekCloser {
	stream := streamcache.New(reader)
	return StreamCacheReadSeekCloser{
		stream:         stream,
		streamReader:   stream.NewReader(ctx),
		originalReader: reader,
	}
}

func (s StreamCacheReadSeekCloser) Read(p []byte) (n int, err error) {
	return s.streamReader.Read(p)
}

func (s StreamCacheReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return s.originalReader.Seek(offset, whence)
}

func (s StreamCacheReadSeekCloser) Close() error {
	return s.originalReader.Close()
}
