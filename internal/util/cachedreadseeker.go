package util

import (
	"fmt"
	"io"
)

// CachedReadSeeker wraps an io.ReadSeekCloser and caches bytes as they are read.
// It implements io.ReadSeeker, allowing seeking within the already-cached
// range without hitting the underlying reader again.
// Additional reads beyond the cache will append to the cache automatically.
type CachedReadSeeker struct {
	src   io.ReadSeekCloser // underlying source
	cache []byte            // bytes read so far
	pos   int64             // current read position
}

func (c *CachedReadSeeker) Close() error {
	return c.src.Close()
}

var _ io.ReadSeekCloser = (*CachedReadSeeker)(nil)

// NewCachedReadSeeker constructs a new CachedReadSeeker wrapping a io.ReadSeekCloser.
func NewCachedReadSeeker(r io.ReadSeekCloser) *CachedReadSeeker {
	return &CachedReadSeeker{src: r}
}

// Read reads up to len(p) bytes into p. It first serves from cache
// if possible, then reads any remaining bytes from the underlying source,
// appending them to the cache.
func (c *CachedReadSeeker) Read(p []byte) (n int, err error) {
	// Check if any part of the request can be served from cache
	if c.pos < int64(len(c.cache)) {
		// Calculate how much we can read from cache
		available := int64(len(c.cache)) - c.pos
		toRead := int64(len(p))
		if available >= toRead {
			// Can serve entirely from cache
			n = copy(p, c.cache[c.pos:c.pos+toRead])
			c.pos += int64(n)
			return n, nil
		}
		// Read what we can from cache
		n = copy(p, c.cache[c.pos:])
		c.pos += int64(n)
		if n == len(p) {
			return n, nil
		}
		// Read the rest from source
		m, err := c.readFromSrc(p[n:])
		n += m
		return n, err
	}

	// Nothing in cache, read from source
	return c.readFromSrc(p)
}

// readFromSrc reads from the underlying source at the current position,
// appends those bytes to cache, and updates the current position.
func (c *CachedReadSeeker) readFromSrc(p []byte) (n int, err error) {
	// Seek to the current position in the source
	if _, err = c.src.Seek(c.pos, io.SeekStart); err != nil {
		return 0, err
	}

	// Read the requested data
	n, err = c.src.Read(p)
	if n > 0 {
		// If reading sequentially or within small gap of cache, append to cache
		if c.pos <= int64(len(c.cache)) {
			c.cache = append(c.cache, p[:n]...)
		}
		c.pos += int64(n)
	}
	return n, err
}

// Seek sets the read position for subsequent Read calls. Seeking within the
// cached range simply updates the position. Seeking beyond will position
// Read to fetch new data from the underlying source (and cache it).
func (c *CachedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var target int64
	switch whence {
	case io.SeekStart:
		target = offset
	case io.SeekCurrent:
		target = c.pos + offset
	case io.SeekEnd:
		// determine end by seeking underlying
		end, err := c.src.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}
		target = end + offset
		// Cache the end position for future SeekEnd calls
		if int64(len(c.cache)) < end {
			c.cache = append(c.cache, make([]byte, end-int64(len(c.cache)))...)
		}
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if target < 0 {
		return 0, fmt.Errorf("negative position: %d", target)
	}

	c.pos = target
	return c.pos, nil
}
