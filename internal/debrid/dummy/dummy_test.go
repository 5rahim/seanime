package dummy

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testSettingsProvider struct {
	settings *models.DummyDebridSettings
}

func (p *testSettingsProvider) GetDummyDebridSettings() (*models.DummyDebridSettings, bool) {
	return p.settings, p.settings != nil
}

func TestDummyProviderTorrentAndUrls(t *testing.T) {
	provider, settings, data := newTestDummyProvider(t)
	defer provider.Close()

	info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: "magnet:?xt=urn:btih:ABC123",
		InfoHash:   "ABC123",
	})
	require.NoError(t, err)
	require.Equal(t, settings.ProfileName, info.Name)
	require.Equal(t, "ABC123", info.Hash)
	require.Len(t, info.Files, 2)
	require.Equal(t, "file-1", info.Files[0].ID)

	availability := provider.GetInstantAvailability([]string{"ABC123"})
	require.Contains(t, availability, "ABC123")
	require.Len(t, availability["ABC123"].CachedFiles, 2)

	torrentID, err := provider.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink:   "magnet:?xt=urn:btih:ABC123",
		InfoHash:     "ABC123",
		SelectFileId: "file-1",
	})
	require.NoError(t, err)
	require.Equal(t, "dummy-ABC123", torrentID)

	itemCh := make(chan debrid.TorrentItem, 2)
	streamURL, err := provider.GetTorrentStreamUrl(context.Background(), debrid.StreamTorrentOptions{
		ID:     torrentID,
		FileId: "file-1",
	}, itemCh)
	require.NoError(t, err)
	require.Contains(t, streamURL, "/dummy-debrid/files/file-1/")

	downloadURL, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
		ID:     torrentID,
		FileId: "file-1",
	})
	require.NoError(t, err)
	require.Equal(t, streamURL, downloadURL)

	item, err := provider.GetTorrent(torrentID)
	require.NoError(t, err)
	require.True(t, item.IsReady)
	require.Equal(t, 100, item.CompletionPercentage)

	req, err := http.NewRequest(http.MethodGet, streamURL, nil)
	require.NoError(t, err)
	req.Header.Set("Range", "bytes=3-9")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusPartialContent, resp.StatusCode)
	require.Equal(t, "bytes 3-9/64", resp.Header.Get("Content-Range"))
	require.Equal(t, data[3:10], body)

	headResp, err := http.Head(streamURL)
	require.NoError(t, err)
	defer headResp.Body.Close()
	require.Equal(t, http.StatusOK, headResp.StatusCode)
	require.Equal(t, "64", headResp.Header.Get("Content-Length"))

	req, err = http.NewRequest(http.MethodGet, streamURL, nil)
	require.NoError(t, err)
	req.Header.Set("Range", "bytes=64-80")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusRequestedRangeNotSatisfiable, resp.StatusCode)
	require.Equal(t, "bytes */64", resp.Header.Get("Content-Range"))

	require.NoError(t, provider.DeleteTorrent(torrentID))
	torrents, err := provider.GetTorrents()
	require.NoError(t, err)
	require.Empty(t, torrents)
}

func TestDummyProviderThrottlesResponseBody(t *testing.T) {
	provider, _, _ := newTestDummyProvider(t)
	defer provider.Close()

	provider.settingsProvider.(*testSettingsProvider).settings.FirstByteDelayMs = 0
	provider.settingsProvider.(*testSettingsProvider).settings.BandwidthBytesPerSecond = 100
	provider.settingsProvider.(*testSettingsProvider).settings.ChunkSize = 20

	streamURL, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{FileId: "file-1"})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, streamURL, nil)
	require.NoError(t, err)
	req.Header.Set("Range", "bytes=0-19")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.GreaterOrEqual(t, time.Since(start), 150*time.Millisecond)
}

func TestDummyProviderCancelsDuringFirstByteDelay(t *testing.T) {
	provider, _, _ := newTestDummyProvider(t)
	defer provider.Close()

	provider.settingsProvider.(*testSettingsProvider).settings.FirstByteDelayMs = 500
	provider.settingsProvider.(*testSettingsProvider).settings.BandwidthBytesPerSecond = 0

	streamURL, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{FileId: "file-1"})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
	require.Less(t, time.Since(start), 300*time.Millisecond)
}

func newTestDummyProvider(t *testing.T) (*Dummy, *models.DummyDebridSettings, []byte) {
	t.Helper()

	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i)
	}

	dir := t.TempDir()
	localFile := filepath.Join(dir, "fixture.mkv")
	require.NoError(t, os.WriteFile(localFile, data, 0644))

	settings := &models.DummyDebridSettings{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Enabled:                 true,
		ProfileName:             "test dummy",
		FallbackFilePath:        localFile,
		Cached:                  true,
		ReadyDelayMs:            0,
		ProgressIntervalMs:      1,
		FirstByteDelayMs:        0,
		BandwidthBytesPerSecond: 0,
		ChunkSize:               16,
		Files: models.DummyDebridFiles{
			{
				ID:            "file-1",
				Path:          "Batch/Show - S01E01.mkv",
				Name:          "Show - S01E01.mkv",
				EpisodeNumber: 1,
				LocalFilePath: localFile,
			},
			{
				ID:            "file-2",
				Path:          "Batch/Show - S01E02.mkv",
				Name:          "Show - S01E02.mkv",
				EpisodeNumber: 2,
				LocalFilePath: localFile,
			},
		},
	}

	provider := New(util.NewLogger(), &testSettingsProvider{settings: settings}).(*Dummy)
	require.NoError(t, provider.Authenticate(""))
	return provider, settings, data
}
