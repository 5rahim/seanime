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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const shortSeekBytes = 1024

// A HttpReadSeeker reads from a http.Response.Body. It can Seek
// by doing range requests.
type HttpReadSeeker struct {
	c   *http.Client
	req *http.Request
	res *http.Response
	ctx context.Context
	r   io.ReadCloser
	pos int64

	Requests int
}

var _ io.ReadCloser = (*HttpReadSeeker)(nil)
var _ io.Seeker = (*HttpReadSeeker)(nil)

var (
	// ErrNoContentLength is returned by Seek when the initial http response did not include a Content-Length header
	ErrNoContentLength = errors.New("Content-Length was not set")
	// ErrRangeRequestsNotSupported is returned by Seek and Read
	// when the remote server does not allow range requests (Accept-Ranges was not set)
	ErrRangeRequestsNotSupported = errors.New("range requests are not supported by the remote server")
	// ErrInvalidRange is returned by Read when trying to read past the end of the file
	ErrInvalidRange = errors.New("invalid range")
	// ErrContentHasChanged is returned by Read when the content has changed since the first request
	ErrContentHasChanged = errors.New("content has changed since first request")
)

// NewHttpReadSeeker returns a HttpReadSeeker, using the http.Response and, optionaly, the http.Client
// that needs to be used for future range requests. If no http.Client is given, http.DefaultClient will
// be used.
//
// res.Request will be reused for range requests, headers may be added/removed
func NewHttpReadSeeker(res *http.Response, client ...*http.Client) *HttpReadSeeker {
	r := &HttpReadSeeker{
		req: res.Request,
		ctx: res.Request.Context(),
		res: res,
		r:   res.Body,
	}
	if len(client) > 0 {
		r.c = client[0]
	} else {
		r.c = http.DefaultClient
	}
	return r
}

// Clone clones the reader to enable parallel downloads of ranges.
// The new HttpReadSeeker will have its own independent request object,
// sharing the original response metadata and HTTP client.
// The context from the original seeker is also carried over.
func (r *HttpReadSeeker) Clone() *HttpReadSeeker {
	clonedReq := r.req.Clone(r.ctx)
	return &HttpReadSeeker{
		c:   r.c,
		req: clonedReq,
		res: r.res,
		ctx: r.ctx, // Carry over the context
		r:   nil,   // Cloned seeker starts with no active reader stream
		// pos and Requests will be zero-initialized, which is the existing behavior.
	}
}

// Read reads from the response body. It does a range request if Seek was called before.
//
// May return ErrRangeRequestsNotSupported, ErrInvalidRange or ErrContentHasChanged
func (r *HttpReadSeeker) Read(p []byte) (n int, err error) {
	if r.r == nil {
		err = r.rangeRequest()
	}
	if r.r != nil {
		n, err = r.r.Read(p)
		r.pos += int64(n)
	}
	return
}

// ReadAt reads from the response body starting at offset off.
//
// May return ErrRangeRequestsNotSupported, ErrInvalidRange or ErrContentHasChanged
func (r *HttpReadSeeker) ReadAt(p []byte, off int64) (n int, err error) {
	var nn int

	r.Seek(off, 0)

	for n < len(p) && err == nil {
		nn, err = r.Read(p[n:])
		n += nn
	}
	return
}

// Close closes the response body
func (r *HttpReadSeeker) Close() error {
	if r.r != nil {
		return r.r.Close()
	}
	return nil
}

// Seek moves the reader position to a new offset.
//
// It does not send http requests, allowing for multiple seeks without overhead.
// The http request will be sent by the next Read call.
//
// May return ErrNoContentLength or ErrRangeRequestsNotSupported
func (r *HttpReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var err error
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.pos
	case io.SeekEnd:
		if r.res.ContentLength <= 0 {
			return 0, ErrNoContentLength
		}
		offset = r.res.ContentLength + offset
	}
	if r.r != nil {
		// Try to read, which is cheaper than doing a request
		if r.pos < offset && offset-r.pos <= shortSeekBytes {
			_, err := io.CopyN(io.Discard, r, offset-r.pos)
			if err != nil {
				return 0, err
			}
		}

		if r.pos != offset {
			err = r.r.Close()
			r.r = nil
		}
	}
	r.pos = offset
	return r.pos, err
}

func (r *HttpReadSeeker) newRequest() *http.Request {
	newreq := r.req.WithContext(r.ctx) // includes shallow copies of maps, but okay
	if r.req.ContentLength == 0 {
		newreq.Body = nil // Issue 16036: nil Body for http.Transport retries
	}
	newreq.Header = r.req.Header.Clone()
	return newreq
}

func (r *HttpReadSeeker) rangeRequest() error {
	r.req = r.newRequest()
	r.req.Header.Set("Range", fmt.Sprintf("bytes=%d-", r.pos))
	etag, last := r.res.Header.Get("ETag"), r.res.Header.Get("Last-Modified")
	switch {
	case last != "":
		r.req.Header.Set("If-Range", last)
	case etag != "":
		r.req.Header.Set("If-Range", etag)
	}

	r.Requests++

	res, err := r.c.Do(r.req)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusRequestedRangeNotSatisfiable:
		return ErrInvalidRange
	case http.StatusOK:
		// some servers return 200 OK for bytes=0-
		if r.pos > 0 ||
			(etag != "" && etag != res.Header.Get("ETag")) {
			return ErrContentHasChanged
		}
		fallthrough
	case http.StatusPartialContent:
		r.r = res.Body
		return nil
	}
	return ErrRangeRequestsNotSupported
}
