package util

import (
	"bytes"
	"os"
	"sync"
	"testing"
)

func TestGlobalLogBufferConcurrentWrites(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})

	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer := logBufferWriter{}
			line := bytes.Repeat([]byte("x"), 512)
			for range 20000 {
				_, _ = writer.Write(line)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-done:
			WriteGlobalLogBufferToFile(file)
			return
		default:
			WriteGlobalLogBufferToFile(file)
		}
	}
}

func TestGlobalLogBufferTail(t *testing.T) {
	logBufferMutex.Lock()
	logBuffer.Reset()
	logBuffer.WriteString("first\nsecond\n")
	logBufferMutex.Unlock()

	if got := string(GlobalLogBufferTail(7)); got != "second\n" {
		t.Fatalf("unexpected tail %q", got)
	}
	if got := string(GlobalLogBufferTail(100)); got != "first\nsecond\n" {
		t.Fatalf("unexpected tail %q", got)
	}

	logBufferMutex.Lock()
	remaining := logBuffer.Len()
	logBuffer.Reset()
	logBufferMutex.Unlock()

	if remaining != len("first\nsecond\n") {
		t.Fatalf("tail consumed the buffer, %d bytes left", remaining)
	}
}
