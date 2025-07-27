package directstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	httputil "seanime/internal/util/http"

	"github.com/neilotoole/streamcache"
)

func serveContentRange(w http.ResponseWriter, r *http.Request, ctx context.Context, reader io.ReadSeekCloser, name string, size int64, contentType string, ranges []httputil.Range) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", contentType)
	//w.Header().Set("Content-Type", "video/webm")
	//w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "keep-alive")
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
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
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

func serveTorrent(w http.ResponseWriter, r *http.Request, ctx context.Context, reader io.ReadSeekCloser, name string, size int64, contentType string, ranges []httputil.Range) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", contentType)
	//w.Header().Set("Content-Type", "video/webm")
	//w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-store")

	// Only handle the first range for now (multiples are rare)
	ra := ranges[0]

	// Validate range
	if ra.Start >= size || ra.Start < 0 || ra.Length <= 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	//if ra.Start == 0 {
	//	// Treat “bytes=0-” exactly like “no Range” => 200 OK
	//	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	//	w.WriteHeader(http.StatusOK)
	//
	//	// Seek back to 0
	//	if _, seekErr := reader.Seek(0, io.SeekStart); seekErr != nil {
	//		http.Error(w, seekErr.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//	copyWithFlush(ctx, w, reader, size)
	//	return
	//}

	// Set response headers for partial content
	w.Header().Set("Content-Range", ra.ContentRange(size))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ra.Length))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to the requested position
	_, err := reader.Seek(ra.Start, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		fmt.Printf("directstream > Served content range: %s\n", ra.ContentRange(size))
	}()

	//http.ServeContent(w, r, name, time.Now(), reader)
	copyWithContext(ctx, w, reader, ra.Length)
}

func copyWithFlush(ctx context.Context, w http.ResponseWriter, rdr io.Reader, totalBytes int64) {
	const flushThreshold = 1 * 1024 * 1024 // 1 MiB
	buf := make([]byte, 32*1024)           // 32 KiB buffer
	var written int64
	var sinceLastFlush int64
	flusher, _ := w.(http.Flusher)

	for written < totalBytes {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// B) Decide how many bytes to read this iteration
		toRead := int64(len(buf))
		if totalBytes-written < toRead {
			toRead = totalBytes - written
		}

		lr := io.LimitReader(rdr, toRead)
		nr, readErr := lr.Read(buf)
		if nr > 0 {
			nw, writeErr := w.Write(buf[:nr])
			written += int64(nw)
			sinceLastFlush += int64(nw)

			if flusher != nil && sinceLastFlush >= flushThreshold {
				flusher.Flush()
				sinceLastFlush = 0
			}
			if writeErr != nil {
				return
			}
			if nw < nr {
				// Client closed or truncated write
				return
			}
		}
		if readErr != nil {
			// EOF or any other read error → stop streaming
			return
		}
	}
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
