package util

import (
	"errors"
	"io"
)

// Common errors that might occur during operations
var (
	ErrInvalidOffset = errors.New("invalid offset: negative or beyond limit")
	ErrInvalidWhence = errors.New("invalid whence value")
	ErrReadLimit     = errors.New("read would exceed limit")
)

// LimitedReadSeeker wraps an io.ReadSeeker and limits the number of bytes
// that can be read from it.
type LimitedReadSeeker struct {
	rs      io.ReadSeeker // The underlying ReadSeeker
	offset  int64         // Current read position relative to start
	limit   int64         // Maximum number of bytes that can be read
	basePos int64         // Original position in the underlying ReadSeeker
}

// NewLimitedReadSeeker creates a new LimitedReadSeeker from the provided
// io.ReadSeeker, starting at the current position and with the given limit.
// The limit parameter specifies the maximum number of bytes that can be
// read from the underlying ReadSeeker.
func NewLimitedReadSeeker(rs io.ReadSeeker, limit int64) (*LimitedReadSeeker, error) {
	if limit < 0 {
		return nil, errors.New("negative limit")
	}

	// Get the current position
	pos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return &LimitedReadSeeker{
		rs:      rs,
		offset:  0,
		limit:   limit,
		basePos: pos,
	}, nil
}

// Read implements the io.Reader interface.
func (lrs *LimitedReadSeeker) Read(p []byte) (n int, err error) {
	if lrs.offset >= lrs.limit {
		return 0, io.EOF
	}

	// Calculate how many bytes we can read
	maxToRead := lrs.limit - lrs.offset
	if int64(len(p)) > maxToRead {
		p = p[:maxToRead]
	}

	n, err = lrs.rs.Read(p)
	lrs.offset += int64(n)
	return
}

// Seek implements the io.Seeker interface.
func (lrs *LimitedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var absoluteOffset int64

	// Calculate the absolute offset based on whence
	switch whence {
	case io.SeekStart:
		absoluteOffset = offset
	case io.SeekCurrent:
		absoluteOffset = lrs.offset + offset
	case io.SeekEnd:
		absoluteOffset = lrs.limit + offset
	default:
		return 0, ErrInvalidWhence
	}

	// Check if the offset is valid
	if absoluteOffset < 0 || absoluteOffset > lrs.limit {
		return 0, ErrInvalidOffset
	}

	// Seek in the underlying ReadSeeker
	_, err := lrs.rs.Seek(lrs.basePos+absoluteOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	// Update our offset
	lrs.offset = absoluteOffset
	return absoluteOffset, nil
}

// Size returns the limit of this LimitedReadSeeker.
func (lrs *LimitedReadSeeker) Size() int64 {
	return lrs.limit
}

// Remaining returns the number of bytes that can still be read.
func (lrs *LimitedReadSeeker) Remaining() int64 {
	return lrs.limit - lrs.offset
}

// Reset resets the read position to the beginning of the limited section.
func (lrs *LimitedReadSeeker) Reset() error {
	_, err := lrs.Seek(0, io.SeekStart)
	return err
}
