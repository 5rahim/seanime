package util

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"
)

// mockSlowReader simulates a slow reader (like network or disk) by adding artificial delay
type mockSlowReader struct {
	data    []byte
	pos     int64
	delay   time.Duration
	readCnt int // count of actual reads from source
}

func newMockSlowReader(data []byte, delay time.Duration) *mockSlowReader {
	return &mockSlowReader{
		data:  data,
		delay: delay,
	}
}

func (m *mockSlowReader) Read(p []byte) (n int, err error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}

	// Simulate latency
	time.Sleep(m.delay)

	m.readCnt++ // track actual reads from source
	n = copy(p, m.data[m.pos:])
	m.pos += int64(n)
	return n, nil
}

func (m *mockSlowReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = m.pos + offset
	case io.SeekEnd:
		abs = int64(len(m.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	m.pos = abs
	return abs, nil
}

func (m *mockSlowReader) Close() error {
	return nil
}

func TestCachedReadSeeker_CachingBehavior(t *testing.T) {
	data := []byte("Hello, this is test data for streaming!")
	delay := 10 * time.Millisecond

	mock := newMockSlowReader(data, delay)
	cached := NewCachedReadSeeker(mock)

	// First read - should hit the source
	buf1 := make([]byte, 5)
	n, err := cached.Read(buf1)
	if err != nil || n != 5 || string(buf1) != "Hello" {
		t.Errorf("First read failed: got %q, want %q", buf1, "Hello")
	}

	// Seek back to start - should not hit source
	_, err = cached.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("Seek failed: %v", err)
	}

	// Second read of same data - should be from cache
	readCntBefore := mock.readCnt
	buf2 := make([]byte, 5)
	n, err = cached.Read(buf2)
	if err != nil || n != 5 || string(buf2) != "Hello" {
		t.Errorf("Second read failed: got %q, want %q", buf2, "Hello")
	}
	if mock.readCnt != readCntBefore {
		t.Error("Second read hit source when it should have used cache")
	}
}

func TestCachedReadSeeker_Performance(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 1000) // ~26KB of data
	delay := 10 * time.Millisecond

	t.Run("Without Cache", func(t *testing.T) {
		mock := newMockSlowReader(data, delay)
		start := time.Now()

		// Read entire data
		if _, err := io.ReadAll(mock); err != nil {
			t.Fatal(err)
		}

		// Seek back and read again
		mock.Seek(0, io.SeekStart)
		if _, err := io.ReadAll(mock); err != nil {
			t.Fatal(err)
		}

		uncachedDuration := time.Since(start)
		t.Logf("Without cache duration: %v", uncachedDuration)
	})

	t.Run("With Cache", func(t *testing.T) {
		mock := newMockSlowReader(data, delay)
		cached := NewCachedReadSeeker(mock)
		start := time.Now()

		// Read entire data
		if _, err := io.ReadAll(cached); err != nil {
			t.Fatal(err)
		}

		// Seek back and read again
		cached.Seek(0, io.SeekStart)
		if _, err := io.ReadAll(cached); err != nil {
			t.Fatal(err)
		}

		cachedDuration := time.Since(start)
		t.Logf("With cache duration: %v", cachedDuration)
	})
}

func TestCachedReadSeeker_SeekBehavior(t *testing.T) {
	data := []byte("0123456789")
	mock := newMockSlowReader(data, 0)
	cached := NewCachedReadSeeker(mock)

	tests := []struct {
		name        string
		offset      int64
		whence      int
		wantPos     int64
		wantRead    string
		readBufSize int
	}{
		{"SeekStart", 3, io.SeekStart, 3, "3456", 4},
		{"SeekCurrent", 2, io.SeekCurrent, 9, "9", 4},
		{"SeekEnd", -5, io.SeekEnd, 5, "56789", 5},
		{"SeekStartZero", 0, io.SeekStart, 0, "0123", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := cached.Seek(tt.offset, tt.whence)
			if err != nil {
				t.Errorf("Seek failed: %v", err)
				return
			}
			if pos != tt.wantPos {
				t.Errorf("Seek position = %d, want %d", pos, tt.wantPos)
			}

			buf := make([]byte, tt.readBufSize)
			n, err := cached.Read(buf)
			if err != nil && err != io.EOF {
				t.Errorf("Read failed: %v", err)
				return
			}
			got := string(buf[:n])
			if got != tt.wantRead {
				t.Errorf("Read after seek = %q, want %q", got, tt.wantRead)
			}
		})
	}
}

func TestCachedReadSeeker_LargeReads(t *testing.T) {
	// Test with larger data to simulate real streaming scenarios
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 1000) // ~26KB
	mock := newMockSlowReader(data, 0)
	cached := NewCachedReadSeeker(mock)

	// Read in chunks
	chunkSize := 1024
	buf := make([]byte, chunkSize)

	var totalRead int
	for {
		n, err := cached.Read(buf)
		totalRead += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
	}

	if totalRead != len(data) {
		t.Errorf("Total read = %d, want %d", totalRead, len(data))
	}

	// Verify cache by seeking back and reading again
	cached.Seek(0, io.SeekStart)
	readCntBefore := mock.readCnt

	totalRead = 0
	for {
		n, err := cached.Read(buf)
		totalRead += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Second read error: %v", err)
		}
	}

	if mock.readCnt != readCntBefore {
		t.Error("Second read hit source when it should have used cache")
	}
}

func TestCachedReadSeeker_ChunkedReadsAndSeeks(t *testing.T) {
	// Create ~1MB of test data
	data := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz0123456789"), 30_000)
	delay := 300 * time.Millisecond // 10ms delay per read to simulate network/disk latency

	// Define read patterns to simulate real-world streaming
	type readOp struct {
		seekOffset int64
		seekWhence int
		readSize   int
		desc       string
	}

	// Simulate typical streaming behavior with repeated reads
	ops := []readOp{
		{0, io.SeekStart, 10 * 1024 * 1024, "initial header"},        // Read first 10MB (headers)
		{500_000, io.SeekStart, 15 * 1024 * 1024, "middle preview"},  // Seek to middle, read 15MB
		{0, io.SeekStart, len(data), "full read after random seeks"}, // Read entire file
		{0, io.SeekStart, len(data), "re-read entire file"},          // Re-read entire file (should be cached)
	}

	var uncachedDuration, cachedDuration time.Duration
	var uncachedReads, cachedReads int

	runTest := func(name string, useCache bool) {
		t.Run(name, func(t *testing.T) {
			mock := newMockSlowReader(data, delay)
			var reader io.ReadSeekCloser = mock
			if useCache {
				reader = NewCachedReadSeeker(mock)
			}

			start := time.Now()
			var totalRead int64

			for i, op := range ops {
				pos, err := reader.Seek(op.seekOffset, op.seekWhence)
				if err != nil {
					t.Fatalf("op %d (%s) - seek failed: %v", i, op.desc, err)
				}

				buf := make([]byte, op.readSize)
				n, err := io.ReadFull(reader, buf)
				if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
					t.Fatalf("op %d (%s) - read failed: %v", i, op.desc, err)
				}

				totalRead += int64(n)
				t.Logf("op %d (%s) - seek to %d, read %d bytes", i, op.desc, pos, n)
			}

			duration := time.Since(start)
			t.Logf("Total bytes read: %d", totalRead)
			t.Logf("Total time: %v", duration)
			t.Logf("Read count from source: %d", mock.readCnt)

			if useCache {
				cachedDuration = duration
				cachedReads = mock.readCnt
			} else {
				uncachedDuration = duration
				uncachedReads = mock.readCnt
			}
		})
	}

	// Run both tests
	runTest("Without Cache", false)
	runTest("With Cache", true)

	// Report performance comparison
	t.Logf("\nPerformance comparison:")
	t.Logf("Uncached: %v (%d reads from source)", uncachedDuration, uncachedReads)
	t.Logf("Cached:   %v (%d reads from source)", cachedDuration, cachedReads)
	t.Logf("Speed improvement: %.2fx", float64(uncachedDuration)/float64(cachedDuration))
	t.Logf("Read reduction: %.2fx", float64(uncachedReads)/float64(cachedReads))
}
