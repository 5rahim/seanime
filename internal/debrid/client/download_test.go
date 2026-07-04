package debrid_client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func initTestDownload(t *testing.T, attempts int, delay func(int) time.Duration) {
	t.Helper()

	oldAttempts := DownloadAttempts
	oldDelay := DownloadRetryDelay
	DownloadAttempts = attempts
	DownloadRetryDelay = delay
	t.Cleanup(func() {
		DownloadAttempts = oldAttempts
		DownloadRetryDelay = oldDelay
	})
}

func initTestDownloadManager(t *testing.T) {
	t.Helper()

	oldManager := hook.GlobalHookManager
	manager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: util.NewLogger()})
	hook.SetGlobalHookManager(manager)
	t.Cleanup(func() { hook.SetGlobalHookManager(oldManager) })
}

func hasDebridDownloadStatus(ws *events.MockWSEventManager, status string) bool {
	for _, event := range ws.Events() {
		if event.Type != events.DebridDownloadProgress {
			continue
		}
		payload, ok := event.Payload.(map[string]interface{})
		if ok && payload["status"] == status {
			return true
		}
	}

	return false
}

func setMobileDownload(t *testing.T, mobile bool) {
	t.Helper()

	old := isMobileDownload
	isMobileDownload = func() bool { return mobile }
	t.Cleanup(func() { isMobileDownload = old })
}

func TestCreateDownloadTempDirUsesAppTempOnMobile(t *testing.T) {
	setMobileDownload(t, true)

	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	destination := t.TempDir()

	tmpDir, err := createDownloadTempDir(destination)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	require.Equal(t, tempRoot, filepath.Dir(tmpDir))
	require.NotEqual(t, destination, filepath.Dir(tmpDir))
	require.Contains(t, filepath.Base(tmpDir), "seanime-debrid-")
}

func TestCreateDownloadTempDirUsesDestinationOffMobile(t *testing.T) {
	setMobileDownload(t, false)

	destination := t.TempDir()

	tmpDir, err := createDownloadTempDir(destination)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	require.Equal(t, destination, filepath.Dir(tmpDir))
	require.Contains(t, filepath.Base(tmpDir), ".tmp-")
}

func TestDownloadFileRetriesOnPartialRead(t *testing.T) {
	initTestDownload(t, 2, func(int) time.Duration { return 0 })

	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()
	var getCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "video/x-matroska")
			return
		}

		if getCalls.Add(1) == 1 {
			w.Header().Set("Content-Length", "10")
			_, _ = w.Write([]byte("short"))
			return
		}

		w.Header().Set("Content-Type", "video/x-matroska")
		_, _ = w.Write([]byte("complete"))
	}))
	t.Cleanup(server.Close)

	repo := &Repository{logger: logger, wsEventManager: ws}
	ok := repo.downloadFileR(context.Background(), "torrent-1", server.URL+"/episode.mkv", destination, result.NewMap[string, downloadStatus]())

	require.True(t, ok)
	require.Equal(t, int32(2), getCalls.Load())
	_, err := os.Stat(filepath.Join(destination, "episode.mkv"))
	require.NoError(t, err)
}

func TestTorrentDownloadCancellationOnFailure(t *testing.T) {
	initTestDownload(t, 1, func(int) time.Duration { return 0 })
	initTestDownloadManager(t)

	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write([]byte("not a zip"))
	}))
	t.Cleanup(server.Close)

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
			return server.URL + "/bad.zip", nil
		}}),
		logger:         logger,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}

	require.NoError(t, repo.downloadTorrentItem("torrent-1", "bad zip", destination))
	require.Eventually(t, func() bool {
		return hasDebridDownloadStatus(ws, "cancelled")
	}, time.Second, 10*time.Millisecond)
	require.Never(t, func() bool {
		return hasDebridDownloadStatus(ws, "completed")
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestRDDownload(t *testing.T) {
	initTestDownload(t, 1, func(int) time.Duration { return 0 })
	initTestDownloadManager(t)

	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()
	body := []byte("rd data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Disposition", `attachment; filename="Episode 01.mkv"`)
			w.Header().Set("Content-Type", "application/force-download")
			return
		}

		// rd usually gives us the real filename in headers, not in the content type
		w.Header().Set("Content-Type", "application/force-download")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
			return server.URL + "/rd", nil
		}}),
		logger:         logger,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}

	require.NoError(t, repo.downloadTorrentItem("torrent-1", "rd", destination))
	require.Eventually(t, func() bool {
		return hasDebridDownloadStatus(ws, "completed")
	}, time.Second, 10*time.Millisecond)

	var data []byte
	require.Eventually(t, func() bool {
		var err error
		data, err = os.ReadFile(filepath.Join(destination, "Episode 01.mkv"))
		return err == nil
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, string(body), string(data))
}

func TestTorBoxZip(t *testing.T) {
	initTestDownload(t, 1, func(int) time.Duration { return 0 })
	initTestDownloadManager(t)

	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	destination := t.TempDir()
	archivePath := writeZipFixture(t, map[string]string{
		"Anime/Episode 01.mkv": "torbox data",
	})
	archive, err := os.ReadFile(archivePath)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// torbox zip links are detected from the response type and extracted locally
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(archive)
	}))
	t.Cleanup(server.Close)

	repo := &Repository{
		provider: mo.Some[debrid.Provider](&fakeDebridProvider{getTorrentDownloadUrl: func(opts debrid.DownloadTorrentOptions) (string, error) {
			return server.URL + "/requestdl", nil
		}}),
		logger:         logger,
		wsEventManager: ws,
		ctxMap:         result.NewMap[string, context.CancelFunc](),
	}

	require.NoError(t, repo.downloadTorrentItem("torrent-1", "torbox", destination))
	require.Eventually(t, func() bool {
		return hasDebridDownloadStatus(ws, "completed")
	}, time.Second, 10*time.Millisecond)

	var data []byte
	require.Eventually(t, func() bool {
		var err error
		data, err = os.ReadFile(filepath.Join(destination, "Episode 01.mkv"))
		return err == nil
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, "torbox data", string(data))
}
