// Package matroska provides utilities for working with Matroska/EBML format.
// This file contains auxiliary utility functions and types that support
// the main functionality of the Matroska/EBML parser.
package matroska

import (
	"fmt"
	"io"
)

// fakeSeeker wraps an io.Reader to implement the io.ReadSeeker interface.
// It provides a way to pass an io.Reader to functions that require an io.ReadSeeker,
// when the seeking functionality will never actually be used.
//
// This is particularly useful in scenarios where you have stream input that doesn't
// support seeking, but you need to pass it to a function that expects a ReadSeeker
// for compatibility reasons, with the guarantee that Seek() will never be called.
//
// Warning: This is a workaround and should only be used when you are certain
// that the Seek method will not be invoked by the receiving function.
// If Seek() is called, it will always return an error.
//
// Example:
//
//	reader := bytes.NewReader([]byte("some data"))
//	fakeSeeker := &fakeSeeker{r: reader}
//
//	// Pass fakeSeeker to a function that expects io.ReadSeeker
//	// but will only call Read()
//	processData(fakeSeeker)
type fakeSeeker struct {
	r io.Reader // The underlying reader that provides the actual data
}

// Read implements the io.Reader interface by delegating to the underlying reader.
// It reads up to len(p) bytes into p and returns the number of bytes read
// and any error encountered.
//
// This method simply forwards the call to the underlying io.Reader's Read method.
//
// Parameters:
//   - p: The byte slice to read data into
//
// Returns:
//   - int: The number of bytes read
//   - error: Any error encountered during reading
func (f *fakeSeeker) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

// Seek implements the io.Seeker interface but always returns an error.
// This method is included only to satisfy the io.ReadSeeker interface requirement.
//
// Since fakeSeeker is designed for streams that don't support actual seeking,
// calling this method will always result in an error. This type should only
// be used in contexts where Seek() is guaranteed not to be called.
//
// Parameters:
//   - offset: The seek offset (ignored)
//   - whence: The seek origin (ignored)
//
// Returns:
//   - int64: Always returns -1
//   - error: Always returns an error indicating this is a fake seeker
func (f *fakeSeeker) Seek(offset int64, whence int) (int64, error) {
	return -1, fmt.Errorf("this is a fake seeker")
}
