package mkvparser

import (
	"bytes"
	"io"
)

// ReadIsMkvOrWebm reads the first 1KB of the stream to determine if it is a Matroska or WebM file.
// It returns the mime type and a boolean indicating if it is a Matroska or WebM file.
// It seeks to the beginning of the stream before and after reading.
func ReadIsMkvOrWebm(r io.ReadSeeker) (string, bool) {
	// Go to the beginning of the stream
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return "", false
	}
	defer r.Seek(0, io.SeekStart)

	return isMkvOrWebm(r)
}

func isMkvOrWebm(r io.Reader) (string, bool) {
	header := make([]byte, 1024) // Read the first 1KB to be safe
	n, err := r.Read(header)
	if err != nil {
		return "", false
	}

	// Check for EBML magic bytes
	if !bytes.HasPrefix(header, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return "", false
	}

	// Look for the DocType tag (0x42 82) and check the string
	docTypeTag := []byte{0x42, 0x82}
	idx := bytes.Index(header, docTypeTag)
	if idx == -1 || idx+3 >= n {
		return "", false
	}

	size := int(header[idx+2]) // Size of DocType field
	if idx+3+size > n {
		return "", false
	}

	docType := string(header[idx+3 : idx+3+size])
	switch docType {
	case "matroska":
		return "video/x-matroska", true
	case "webm":
		return "video/webm", true
	default:
		return "", false
	}
}
