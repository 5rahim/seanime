package httputil

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type piece struct {
	start int64
	end   int64
}

// FileStream saves a HTTP file being streamed to disk.
// It allows multiple readers to read the file concurrently.
// It works by being fed the stream from the HTTP response body. It will simultaneously write to disk and to the HTTP writer.
type FileStream struct {
	length    int64
	file      *os.File
	closed    bool
	mu        sync.Mutex
	pieces    map[int64]*piece
	readers   []FileStreamReader
	readersMu sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zerolog.Logger
}

type FileStreamReader interface {
	io.ReadSeekCloser
}

// NewFileStream creates a new FileStream instance with a temporary file
func NewFileStream(ctx context.Context, logger *zerolog.Logger, contentLength int64) (*FileStream, error) {
	file, err := os.CreateTemp("", "filestream_*.tmp")
	if err != nil {
		return nil, err
	}

	// Pre-allocate the file to the expected content length
	if contentLength > 0 {
		if err := file.Truncate(contentLength); err != nil {
			_ = file.Close()
			_ = os.Remove(file.Name())
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(ctx)

	return &FileStream{
		file:   file,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
		pieces: make(map[int64]*piece),
		length: contentLength,
	}, nil
}

// WriteAndFlush writes the stream to the file at the given offset and flushes it to the HTTP writer
func (fs *FileStream) WriteAndFlush(src io.Reader, dst io.Writer, offset int64) error {
	fs.mu.Lock()
	if fs.closed {
		fs.mu.Unlock()
		return io.ErrClosedPipe
	}
	fs.mu.Unlock()

	buffer := make([]byte, 32*1024) // 32KB buffer
	currentOffset := offset

	for {
		select {
		case <-fs.ctx.Done():
			return fs.ctx.Err()
		default:
		}

		n, readErr := src.Read(buffer)
		if n > 0 {
			// Write to file
			fs.mu.Lock()
			if !fs.closed {
				if _, err := fs.file.WriteAt(buffer[:n], currentOffset); err != nil {
					fs.mu.Unlock()
					return err
				}

				// Update pieces map
				pieceEnd := currentOffset + int64(n) - 1
				fs.updatePieces(currentOffset, pieceEnd)
			}
			fs.mu.Unlock()

			// Write to HTTP response
			if _, err := dst.Write(buffer[:n]); err != nil {
				return err
			}

			// Flush if possible
			if flusher, ok := dst.(interface{ Flush() }); ok {
				flusher.Flush()
			}

			currentOffset += int64(n)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	// Sync file to ensure data is written
	fs.mu.Lock()
	if !fs.closed {
		_ = fs.file.Sync()
	}
	fs.mu.Unlock()

	return nil
}

// updatePieces merges the new piece with existing pieces
func (fs *FileStream) updatePieces(start, end int64) {
	newPiece := &piece{start: start, end: end}

	// Find overlapping or adjacent pieces
	var toMerge []*piece
	var toDelete []int64

	for key, p := range fs.pieces {
		if p.start <= end+1 && p.end >= start-1 {
			toMerge = append(toMerge, p)
			toDelete = append(toDelete, key)
		}
	}

	// Merge all overlapping pieces
	for _, p := range toMerge {
		if p.start < newPiece.start {
			newPiece.start = p.start
		}
		if p.end > newPiece.end {
			newPiece.end = p.end
		}
	}

	// Delete old pieces
	for _, key := range toDelete {
		delete(fs.pieces, key)
	}

	// Add the merged piece
	fs.pieces[newPiece.start] = newPiece
}

// isRangeAvailable checks if a given range is completely downloaded
func (fs *FileStream) isRangeAvailable(start, end int64) bool {
	for _, p := range fs.pieces {
		if p.start <= start && p.end >= end {
			return true
		}
	}
	return false
}

// NewReader creates a new FileStreamReader for concurrent reading
func (fs *FileStream) NewReader() (FileStreamReader, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.closed {
		return nil, io.ErrClosedPipe
	}

	reader := &fileStreamReader{
		fs:     fs,
		file:   fs.file,
		offset: 0,
	}

	fs.readersMu.Lock()
	fs.readers = append(fs.readers, reader)
	fs.readersMu.Unlock()

	return reader, nil
}

// Close closes the FileStream and cleans up resources
func (fs *FileStream) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.closed {
		return nil
	}

	fs.closed = true
	fs.cancel()

	// Close all readers
	fs.readersMu.Lock()
	for _, reader := range fs.readers {
		go reader.Close()
	}
	fs.readers = nil
	fs.readersMu.Unlock()

	// Remove the temp file and close
	fileName := fs.file.Name()
	_ = fs.file.Close()
	_ = os.Remove(fileName)

	return nil
}

// Length returns the current length of the stream
func (fs *FileStream) Length() int64 {
	return fs.length
}

// fileStreamReader implements FileStreamReader interface
type fileStreamReader struct {
	fs     *FileStream
	file   *os.File
	offset int64
	closed bool
	mu     sync.Mutex
}

// Read reads data from the file stream, blocking if data is not yet available
func (r *fileStreamReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.ErrClosedPipe
	}

	readSize := int64(len(p))
	readEnd := r.offset + readSize - 1

	if readEnd >= r.fs.length {
		readEnd = r.fs.length - 1
		readSize = r.fs.length - r.offset
		if readSize <= 0 {
			return 0, io.EOF
		}
	}

	for {
		select {
		case <-r.fs.ctx.Done():
			return 0, r.fs.ctx.Err()
		default:
		}

		r.fs.mu.Lock()
		streamClosed := r.fs.closed

		// Check if the range we want to read is available
		available := r.fs.isRangeAvailable(r.offset, readEnd)

		// If not fully available, check what we can read
		var actualReadSize int64 = readSize
		if !available {
			// Find the largest available chunk starting from our offset
			var maxRead int64 = 0
			for _, piece := range r.fs.pieces {
				if piece.start <= r.offset && piece.end >= r.offset {
					chunkEnd := piece.end
					if chunkEnd >= readEnd {
						maxRead = readSize
					} else {
						maxRead = chunkEnd - r.offset + 1
					}
					break
				}
			}
			actualReadSize = maxRead
		}
		r.fs.mu.Unlock()

		// If we have some data to read, or if stream is closed, attempt the read
		if available || actualReadSize > 0 || streamClosed {
			var n int
			var err error

			if actualReadSize > 0 {
				n, err = r.file.ReadAt(p[:actualReadSize], r.offset)
			} else if streamClosed {
				// If stream is closed and no data available, try reading anyway to get proper EOF
				n, err = r.file.ReadAt(p[:readSize], r.offset)
			}

			if n > 0 {
				r.offset += int64(n)
			}

			// If we read less than requested and stream is closed, return EOF
			if n < len(p) && streamClosed && r.offset >= r.fs.length {
				if err == nil {
					err = io.EOF
				}
			}

			// If no data was read and stream is closed, return EOF
			if n == 0 && streamClosed {
				return 0, io.EOF
			}

			// Return what we got, even if it's 0 bytes (this prevents hanging)
			return n, err
		}

		// Wait a bit before checking again
		r.mu.Unlock()
		select {
		case <-r.fs.ctx.Done():
			r.mu.Lock()
			return 0, r.fs.ctx.Err()
		case <-time.After(10 * time.Millisecond):
			r.mu.Lock()
		}
	}
}

// Seek sets the offset for the next Read
func (r *fileStreamReader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.ErrClosedPipe
	}

	switch whence {
	case io.SeekStart:
		r.offset = offset
	case io.SeekCurrent:
		r.offset += offset
	case io.SeekEnd:
		r.fs.mu.Lock()
		r.offset = r.fs.length + offset
		r.fs.mu.Unlock()
	default:
		return 0, errors.New("invalid whence")
	}

	if r.offset < 0 {
		r.offset = 0
	}

	return r.offset, nil
}

// Close closes the reader
func (r *fileStreamReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	r.fs.readersMu.Lock()
	for i, reader := range r.fs.readers {
		if reader == r {
			r.fs.readers = append(r.fs.readers[:i], r.fs.readers[i+1:]...)
			break
		}
	}
	r.fs.readersMu.Unlock()

	return nil
}
