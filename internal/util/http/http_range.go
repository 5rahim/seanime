package httputil

import (
	"errors"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

// Range specifies the byte range to be sent to the client.
type Range struct {
	Start  int64
	Length int64
}

// ContentRange returns Content-Range header value.
func (r Range) ContentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.Start, r.Start+r.Length-1, size)
}

var (
	// ErrNoOverlap is returned by ParseRange if first-byte-pos of
	// all of the byte-range-spec values is greater than the content size.
	ErrNoOverlap = errors.New("invalid range: failed to overlap")

	// ErrInvalid is returned by ParseRange on invalid input.
	ErrInvalid = errors.New("invalid range")
)

// ParseRange parses a Range header string as per RFC 7233.
// ErrNoOverlap is returned if none of the ranges overlap.
// ErrInvalid is returned if s is invalid range.
func ParseRange(s string, size int64) ([]Range, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, ErrInvalid
	}
	var ranges []Range
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, ErrInvalid
		}
		start, end := textproto.TrimString(ra[:i]), textproto.TrimString(ra[i+1:])
		var r Range
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				return nil, ErrInvalid
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, ErrInvalid
			}
			if i > size {
				i = size
			}
			r.Start = size - i
			r.Length = size - r.Start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, ErrInvalid
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.Start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.Length = size - r.Start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.Start > i {
					return nil, ErrInvalid
				}
				if i >= size {
					i = size - 1
				}
				r.Length = i - r.Start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, ErrNoOverlap
	}
	return ranges, nil
}
