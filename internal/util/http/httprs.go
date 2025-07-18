package httputil

// Original source: https://github.com/jfbus/httprs/tree/master

/*
Package httprs provides a ReadSeeker for http.Response.Body.

Usage :

	resp, err := http.Get(url)
	rs := httprs.NewHttpReadSeeker(resp)
	defer rs.Close()
	io.ReadFull(rs, buf) // reads the first bytes from the response body
	rs.Seek(1024, 0) // moves the position, but does no range request
	io.ReadFull(rs, buf) // does a range request and reads from the response body

If you want to use a specific http.Client for additional range requests :

	rs := httprs.NewHttpReadSeeker(resp, client)
*/

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// HttpReadSeeker implements io.ReadSeeker for HTTP responses
// It allows seeking within an HTTP response by using HTTP Range requests
type HttpReadSeeker struct {
	url        string         // The URL of the resource
	client     *http.Client   // HTTP client to use for requests
	resp       *http.Response // Current response
	offset     int64          // Current offset in the resource
	size       int64          // Size of the resource, -1 if unknown
	readBuf    []byte         // Buffer for reading
	readOffset int            // Current offset in readBuf
	mu         sync.Mutex     // Mutex for thread safety
}

// NewHttpReadSeeker creates a new HttpReadSeeker from an http.Response
func NewHttpReadSeeker(resp *http.Response) *HttpReadSeeker {
	url := ""
	if resp.Request != nil {
		url = resp.Request.URL.String()
	}

	size := int64(-1)
	if resp.ContentLength > 0 {
		size = resp.ContentLength
	}

	return &HttpReadSeeker{
		url:        url,
		client:     http.DefaultClient,
		resp:       resp,
		offset:     0,
		size:       size,
		readBuf:    nil,
		readOffset: 0,
	}
}

// Read implements io.Reader
func (hrs *HttpReadSeeker) Read(p []byte) (n int, err error) {
	hrs.mu.Lock()
	defer hrs.mu.Unlock()

	// If we have buffered data, read from it first
	if hrs.readBuf != nil && hrs.readOffset < len(hrs.readBuf) {
		n = copy(p, hrs.readBuf[hrs.readOffset:])
		hrs.readOffset += n
		hrs.offset += int64(n)

		// Clear buffer if we've read it all
		if hrs.readOffset >= len(hrs.readBuf) {
			hrs.readBuf = nil
			hrs.readOffset = 0
		}

		return n, nil
	}

	// If we don't have a response or it's been closed, get a new one
	if hrs.resp == nil {
		if err := hrs.makeRangeRequest(); err != nil {
			return 0, err
		}
	}

	// Read from the response body
	n, err = hrs.resp.Body.Read(p)
	hrs.offset += int64(n)

	return n, err
}

// Seek implements io.Seeker
func (hrs *HttpReadSeeker) Seek(offset int64, whence int) (int64, error) {
	hrs.mu.Lock()
	defer hrs.mu.Unlock()

	var newOffset int64

	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = hrs.offset + offset
	case io.SeekEnd:
		if hrs.size < 0 {
			// If we don't know the size, we need to determine it
			if err := hrs.determineSize(); err != nil {
				return hrs.offset, err
			}
		}
		newOffset = hrs.size + offset
	default:
		return hrs.offset, fmt.Errorf("httprs: invalid whence %d", whence)
	}

	if newOffset < 0 {
		return hrs.offset, fmt.Errorf("httprs: negative position")
	}

	// If we're just moving the offset without reading, we can skip the request
	// We'll make a new request when Read is called
	if hrs.resp != nil {
		hrs.resp.Body.Close()
		hrs.resp = nil
	}

	hrs.offset = newOffset
	hrs.readBuf = nil
	hrs.readOffset = 0

	return hrs.offset, nil
}

// Close closes the underlying response body
func (hrs *HttpReadSeeker) Close() error {
	hrs.mu.Lock()
	defer hrs.mu.Unlock()

	if hrs.resp != nil {
		err := hrs.resp.Body.Close()
		hrs.resp = nil
		return err
	}

	return nil
}

// makeRangeRequest makes a new HTTP request with the Range header
func (hrs *HttpReadSeeker) makeRangeRequest() error {
	req, err := http.NewRequest("GET", hrs.url, nil)
	if err != nil {
		return err
	}

	// Set Range header from current offset
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", hrs.offset))

	// Make the request
	resp, err := hrs.client.Do(req)
	if err != nil {
		return err
	}

	// Check if the server supports range requests
	if resp.StatusCode != http.StatusPartialContent && hrs.offset > 0 {
		resp.Body.Close()
		return fmt.Errorf("httprs: server does not support range requests")
	}

	// Update our response and offset
	if hrs.resp != nil {
		hrs.resp.Body.Close()
	}
	hrs.resp = resp

	// Update the size if we get it from Content-Range
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		// Format: bytes <start>-<end>/<size>
		parts := strings.Split(contentRange, "/")
		if len(parts) > 1 && parts[1] != "*" {
			if size, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				hrs.size = size
			}
		}
	} else if resp.ContentLength > 0 {
		// If we don't have a Content-Range header but we do have Content-Length,
		// then the size is the current offset plus the content length
		hrs.size = hrs.offset + resp.ContentLength
	}

	return nil
}

// determineSize makes a HEAD request to determine the size of the resource
func (hrs *HttpReadSeeker) determineSize() error {
	req, err := http.NewRequest("HEAD", hrs.url, nil)
	if err != nil {
		return err
	}

	resp, err := hrs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.ContentLength > 0 {
		hrs.size = resp.ContentLength
	} else {
		// If we still don't know the size, return an error
		return fmt.Errorf("httprs: unable to determine resource size")
	}

	return nil
}

// ReadAt implements io.ReaderAt
func (hrs *HttpReadSeeker) ReadAt(p []byte, off int64) (n int, err error) {
	// Save current offset
	currentOffset := hrs.offset

	// Seek to the requested offset
	if _, err := hrs.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}

	// Read the data
	n, err = hrs.Read(p)

	// Restore the original offset
	if _, seekErr := hrs.Seek(currentOffset, io.SeekStart); seekErr != nil {
		// If we can't restore the offset, return that error instead
		if err == nil {
			err = seekErr
		}
	}

	return n, err
}

// Size returns the size of the resource, or -1 if unknown
func (hrs *HttpReadSeeker) Size() int64 {
	hrs.mu.Lock()
	defer hrs.mu.Unlock()

	if hrs.size < 0 {
		// Try to determine the size
		_ = hrs.determineSize()
	}

	return hrs.size
}

// WithClient returns a new HttpReadSeeker with the specified client
func (hrs *HttpReadSeeker) WithClient(client *http.Client) *HttpReadSeeker {
	hrs.mu.Lock()
	defer hrs.mu.Unlock()

	hrs.client = client
	return hrs
}
