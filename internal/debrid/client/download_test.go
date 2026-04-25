package debrid_client

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/hook_resolver"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type recordingWSEventManager struct {
	*events.MockWSEventManager
	mu     sync.Mutex
	events []events.MockWSEvent
}

func newRecordingWSEventManager(logger *zerolog.Logger) *recordingWSEventManager {
	return &recordingWSEventManager{
		MockWSEventManager: events.NewMockWSEventManager(logger),
	}
}

func (m *recordingWSEventManager) SendEvent(t string, payload interface{}) {
	m.mu.Lock()
	m.events = append(m.events, events.MockWSEvent{Type: t, Payload: payload})
	m.mu.Unlock()
}

func (m *recordingWSEventManager) countDownloadStatus(status string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, event := range m.events {
		if event.Type != events.DebridDownloadProgress {
			continue
		}
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			continue
		}
		if payload["status"] == status {
			count++
		}
	}
	return count
}

func TestDebridDownloadFileFullZipWithContentLength(t *testing.T) {
	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	repo := &Repository{logger: logger, wsEventManager: ws}
	zipBytes := makeDownloadTestZip(t, "episode.txt", "complete")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		writeZipResponseHeaders(w, len(zipBytes), "")
		_, _ = w.Write(zipBytes)
	}))
	t.Cleanup(server.Close)

	destination := t.TempDir()
	ok := repo.downloadFile(context.Background(), "torrent-1", server.URL, destination, result.NewMap[string, downloadStatus]())

	require.True(t, ok)
	require.FileExists(t, filepath.Join(destination, "episode.txt"))
	content, err := os.ReadFile(filepath.Join(destination, "episode.txt"))
	require.NoError(t, err)
	require.Equal(t, "complete", string(content))
}

func TestDebridDownloadFileResumesAfterTruncatedResponse(t *testing.T) {
	withFastDebridDownloadRetries(t, 3)

	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	repo := &Repository{logger: logger, wsEventManager: ws}
	zipBytes := makeDownloadTestZip(t, "episode.txt", "resumed")
	splitAt := len(zipBytes) / 2
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNumber := requests.Add(1)
		if requestNumber == 1 {
			writeZipResponseHeaders(w, len(zipBytes), "")
			_, _ = w.Write(zipBytes[:splitAt])
			return
		}

		require.Equal(t, fmt.Sprintf("bytes=%d-", splitAt), r.Header.Get("Range"))
		writeZipResponseHeaders(w, len(zipBytes)-splitAt, fmt.Sprintf("bytes %d-%d/%d", splitAt, len(zipBytes)-1, len(zipBytes)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(zipBytes[splitAt:])
	}))
	t.Cleanup(server.Close)

	destination := t.TempDir()
	ok := repo.downloadFile(context.Background(), "torrent-1", server.URL, destination, result.NewMap[string, downloadStatus]())

	require.True(t, ok)
	require.Equal(t, int32(2), requests.Load())
	content, err := os.ReadFile(filepath.Join(destination, "episode.txt"))
	require.NoError(t, err)
	require.Equal(t, "resumed", string(content))
}

func TestDebridDownloadFileRestartsWhenRangeIgnored(t *testing.T) {
	withFastDebridDownloadRetries(t, 3)

	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	repo := &Repository{logger: logger, wsEventManager: ws}
	zipBytes := makeDownloadTestZip(t, "episode.txt", "restarted")
	splitAt := len(zipBytes) / 2
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNumber := requests.Add(1)
		writeZipResponseHeaders(w, len(zipBytes), "")
		if requestNumber == 1 {
			_, _ = w.Write(zipBytes[:splitAt])
			return
		}
		require.NotEmpty(t, r.Header.Get("Range"))
		_, _ = w.Write(zipBytes)
	}))
	t.Cleanup(server.Close)

	destination := t.TempDir()
	ok := repo.downloadFile(context.Background(), "torrent-1", server.URL, destination, result.NewMap[string, downloadStatus]())

	require.True(t, ok)
	require.Equal(t, int32(2), requests.Load())
	content, err := os.ReadFile(filepath.Join(destination, "episode.txt"))
	require.NoError(t, err)
	require.Equal(t, "restarted", string(content))
}

func TestDebridDownloadFileRetriesTransientStatusResponses(t *testing.T) {
	withFastDebridDownloadRetries(t, 2)

	for _, statusCode := range []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		t.Run(fmt.Sprintf("status_%d", statusCode), func(t *testing.T) {
			logger := util.NewLogger()
			ws := newRecordingWSEventManager(logger)
			repo := &Repository{logger: logger, wsEventManager: ws}
			zipBytes := makeDownloadTestZip(t, "episode.txt", "retried")
			var requests atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestNumber := requests.Add(1)
				if requestNumber == 1 {
					w.WriteHeader(statusCode)
					_, _ = w.Write([]byte(http.StatusText(statusCode)))
					return
				}

				writeZipResponseHeaders(w, len(zipBytes), "")
				_, _ = w.Write(zipBytes)
			}))
			t.Cleanup(server.Close)

			destination := t.TempDir()
			ok := repo.downloadFile(context.Background(), "torrent-1", server.URL, destination, result.NewMap[string, downloadStatus]())

			require.True(t, ok)
			require.Equal(t, int32(2), requests.Load())
			content, err := os.ReadFile(filepath.Join(destination, "episode.txt"))
			require.NoError(t, err)
			require.Equal(t, "retried", string(content))
		})
	}
}

func TestDebridDownloadFileFailsFastPermanentStatusResponse(t *testing.T) {
	withFastDebridDownloadRetries(t, 3)

	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	repo := &Repository{logger: logger, wsEventManager: ws}
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(http.StatusText(http.StatusNotFound)))
	}))
	t.Cleanup(server.Close)

	ok := repo.downloadFile(context.Background(), "torrent-1", server.URL, t.TempDir(), result.NewMap[string, downloadStatus]())

	require.False(t, ok)
	require.Equal(t, int32(1), requests.Load())
}

func TestDebridDownloadTorrentDoesNotCompleteAfterTruncatedRetries(t *testing.T) {
	withFastDebridDownloadRetries(t, 3)

	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	zipBytes := makeDownloadTestZip(t, "episode.txt", "truncated")
	splitAt := len(zipBytes) / 2
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		writeZipResponseHeaders(w, len(zipBytes), "")
		_, _ = w.Write(zipBytes[:splitAt])
	}))
	t.Cleanup(server.Close)

	completed := make(chan struct{}, 1)
	withDebridDownloadCompletedHook(t, logger, completed)
	repo := newDownloadTestRepository(logger, ws, server.URL)

	err := repo.downloadTorrentItem("torrent-1", "test torrent", t.TempDir())
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return requests.Load() >= int32(debridDownloadMaxAttempts) && ws.countDownloadStatus("cancelled") > 0
	}, time.Second, 10*time.Millisecond)

	require.Never(t, func() bool {
		return ws.countDownloadStatus("completed") > 0 || len(completed) > 0
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestDebridDownloadTorrentAggregatesMultiURLFailures(t *testing.T) {
	withFastDebridDownloadRetries(t, 2)

	logger := util.NewLogger()
	ws := newRecordingWSEventManager(logger)
	successZip := makeDownloadTestZip(t, "episode.txt", "success")
	failedZip := makeDownloadTestZip(t, "failed.txt", "failed")
	var failedRequests atomic.Int32

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeZipResponseHeaders(w, len(successZip), "")
		_, _ = w.Write(successZip)
	}))
	t.Cleanup(successServer.Close)

	failedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failedRequests.Add(1)
		writeZipResponseHeaders(w, len(failedZip), "")
		_, _ = w.Write(failedZip[:len(failedZip)/2])
	}))
	t.Cleanup(failedServer.Close)

	completed := make(chan struct{}, 1)
	withDebridDownloadCompletedHook(t, logger, completed)
	downloadURL := strings.Join([]string{successServer.URL, failedServer.URL}, ",")
	repo := newDownloadTestRepository(logger, ws, downloadURL)

	err := repo.downloadTorrentItem("torrent-1", "test torrent", t.TempDir())
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return failedRequests.Load() >= int32(debridDownloadMaxAttempts) && ws.countDownloadStatus("cancelled") > 0
	}, time.Second, 10*time.Millisecond)

	require.Never(t, func() bool {
		return ws.countDownloadStatus("completed") > 0 || len(completed) > 0
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func makeDownloadTestZip(t *testing.T, name string, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	require.NoError(t, err)
	_, err = w.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func writeZipResponseHeaders(w http.ResponseWriter, contentLength int, contentRange string) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="batch.zip"`)
	w.Header().Set("Content-Length", strconv.Itoa(contentLength))
	if contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
	}
}

func withFastDebridDownloadRetries(t *testing.T, attempts int) {
	t.Helper()

	oldAttempts := debridDownloadMaxAttempts
	oldInitialBackoff := debridDownloadInitialBackoff
	oldMaxBackoff := debridDownloadMaxBackoff
	debridDownloadMaxAttempts = attempts
	debridDownloadInitialBackoff = time.Millisecond
	debridDownloadMaxBackoff = time.Millisecond
	t.Cleanup(func() {
		debridDownloadMaxAttempts = oldAttempts
		debridDownloadInitialBackoff = oldInitialBackoff
		debridDownloadMaxBackoff = oldMaxBackoff
	})
}

func withDebridDownloadCompletedHook(t *testing.T, logger *zerolog.Logger, completed chan<- struct{}) {
	t.Helper()

	oldManager := hook.GlobalHookManager
	manager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(manager)
	t.Cleanup(func() {
		hook.SetGlobalHookManager(oldManager)
	})

	manager.OnDebridLocalDownloadCompleted().BindFunc(func(e hook_resolver.Resolver) error {
		completed <- struct{}{}
		return e.Next()
	})
}

func newDownloadTestRepository(logger *zerolog.Logger, ws events.WSEventManagerInterface, downloadURL string) *Repository {
	return &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{downloadURL: func(opts debrid.DownloadTorrentOptions) (string, error) {
			return downloadURL, nil
		}}),
		logger:         logger,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}
}
