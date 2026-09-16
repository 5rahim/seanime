package torbox

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/rs/zerolog"
	"seanime/internal/debrid/debrid"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestTorBox(t *testing.T, respond func(*http.Request) (any, error)) *TorBox {
	t.Helper()
	logger := zerolog.Nop()
	provider := NewTorBox(&logger).(*TorBox)
	_ = provider.Authenticate("test-api-key")
	provider.client.Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		data, err := respond(r)
		if err != nil {
			return nil, err
		}
		body, err := json.Marshal(map[string]any{"success": true, "data": data})
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
	})
	return provider
}

func TestTorrentInfoMatchesHashCase(t *testing.T) {
	const hash = "0123456789abcdef0123456789abcdef01234567"
	for _, returnedHash := range []string{hash, strings.ToUpper(hash)} {
		t.Run(returnedHash, func(t *testing.T) {
			provider := newTestTorBox(t, func(r *http.Request) (any, error) {
				if !strings.HasSuffix(r.URL.Path, "/checkcached") {
					t.Fatalf("unexpected metadata fallback: %s", r.URL.Path)
				}
				if r.URL.Query().Get("hash") != hash {
					t.Fatalf("hash was not normalized: %s", r.URL.Query().Get("hash"))
				}
				return map[string]any{returnedHash: &TorrentInfo{Hash: hash, Files: []*TorrentInfoFile{{Name: "Batch/Episode.mkv"}}}}, nil
			})
			info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{InfoHash: " " + strings.ToUpper(hash) + " "})
			if err != nil || len(info.Files) != 1 {
				t.Fatalf("cached video should be returned: info=%v err=%v", info, err)
			}
		})
	}
}

func TestTorrentInfoZipOnlyDoesNotFallBack(t *testing.T) {
	provider := newTestTorBox(t, func(r *http.Request) (any, error) {
		if !strings.HasSuffix(r.URL.Path, "/checkcached") {
			t.Fatalf("ZIP cache must not fall back to original file metadata: %s", r.URL.Path)
		}
		return map[string]any{"abc": &TorrentInfo{Files: []*TorrentInfoFile{{Name: "Batch/Batch.ZIP"}, {Name: "Batch/readme.txt"}}}}, nil
	})
	info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{InfoHash: "ABC"})
	if info != nil || !errors.Is(err, errZipOnly) {
		t.Fatalf("expected an archive-only error, got info=%v err=%v", info, err)
	}
}

func TestTorrentInfoUncachedFallbackAndMixedFiles(t *testing.T) {
	requests := 0
	provider := newTestTorBox(t, func(r *http.Request) (any, error) {
		requests++
		if strings.HasSuffix(r.URL.Path, "/checkcached") {
			return map[string]any{}, nil
		}
		if !strings.HasSuffix(r.URL.Path, "/torrentinfo") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		return &TorrentInfo{Files: []*TorrentInfoFile{{Name: "Extras.zip"}, {Name: "Episode.mkv"}}}, nil
	})
	info, err := provider.GetTorrentInfo(debrid.GetTorrentInfoOptions{InfoHash: "ABC"})
	if err != nil || len(info.Files) != 2 || requests != 2 {
		t.Fatalf("video alongside a ZIP should remain available: info=%v requests=%d err=%v", info, requests, err)
	}
}

func TestStreamMissingFileStopsAfterFiveAttempts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		provider := newTestTorBox(t, func(r *http.Request) (any, error) {
			if !strings.HasSuffix(r.URL.Path, "/mylist") {
				t.Fatalf("missing file must not request a download: %s", r.URL.Path)
			}
			calls++
			return &Torrent{DownloadPresent: true, Progress: 1, Files: []*File{{ShortName: "Other.mkv"}}}, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := provider.GetTorrentStreamUrl(ctx, debrid.StreamTorrentOptions{ID: "1", FileId: "Episode.mkv"}, nil)
		if err == nil || !strings.Contains(err.Error(), "after 5 attempts") || !strings.Contains(err.Error(), "file not found") || calls != 10 {
			t.Fatalf("expected bounded filename failures, got calls=%d err=%v", calls, err)
		}
	})
}

func TestStreamLinkFailuresRemainBoundedAcrossStatusErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		statusCalls, linkCalls := 0, 0
		linkFailure := errors.New("download-link service unavailable")
		provider := newTestTorBox(t, func(r *http.Request) (any, error) {
			if strings.HasSuffix(r.URL.Path, "/mylist") {
				statusCalls++
				if statusCalls%3 == 0 {
					return nil, errors.New("temporary status failure")
				}
				return &Torrent{DownloadPresent: true, Files: []*File{{ID: 2, ShortName: "Episode.mkv"}}}, nil
			}
			linkCalls++
			return nil, linkFailure
		})
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		_, err := provider.GetTorrentStreamUrl(ctx, debrid.StreamTorrentOptions{ID: "1", FileId: "Episode.mkv"}, nil)
		if linkCalls != 5 || statusCalls != 14 || !errors.Is(err, linkFailure) {
			t.Fatalf("link failure limit/cause lost: status=%d links=%d err=%v", statusCalls, linkCalls, err)
		}
	})
}

func TestStreamTransientLinkFailureCanRecover(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		linkCalls := 0
		provider := newTestTorBox(t, func(r *http.Request) (any, error) {
			if strings.HasSuffix(r.URL.Path, "/mylist") {
				return &Torrent{DownloadPresent: true, Files: []*File{{ID: 2, ShortName: "Episode.mkv"}}}, nil
			}
			linkCalls++
			if linkCalls < 3 {
				return nil, errors.New("temporary link failure")
			}
			return "https://example.invalid/Episode.mkv", nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		url, err := provider.GetTorrentStreamUrl(ctx, debrid.StreamTorrentOptions{ID: "1", FileId: "Episode.mkv"}, nil)
		if err != nil || url != "https://example.invalid/Episode.mkv" || linkCalls != 3 {
			t.Fatalf("transient retry failed: url=%s calls=%d err=%v", url, linkCalls, err)
		}
	})
}

func TestStreamStaleZipSelectionFailsImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		provider := newTestTorBox(t, func(r *http.Request) (any, error) {
			if !strings.HasSuffix(r.URL.Path, "/mylist") {
				t.Fatalf("ZIP-only movie selection must not request a link: %s", r.URL.Path)
			}
			calls++
			return &Torrent{DownloadPresent: true, Files: []*File{{ID: 0, Name: "Batch/Batch.zip", ShortName: "Batch.zip", MimeType: "application/zip"}}}, nil
		})
		_, err := provider.GetTorrentStreamUrl(context.Background(), debrid.StreamTorrentOptions{ID: "1", FileId: "Movie.mkv"}, nil)
		if !errors.Is(err, errZipOnly) || calls != 2 {
			t.Fatalf("expected immediate ZIP-only failure: calls=%d err=%v", calls, err)
		}
	})
}

func TestStreamCancellationReachesInFlightRequests(t *testing.T) {
	for _, cancelAt := range []string{"status", "fileLookup", "downloadLink"} {
		t.Run(cancelAt, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				started := make(chan struct{})
				calls := 0
				provider := newTestTorBox(t, func(r *http.Request) (any, error) {
					calls++
					block := cancelAt == "status" || cancelAt == "fileLookup" && calls == 2 || cancelAt == "downloadLink" && strings.HasSuffix(r.URL.Path, "/requestdl")
					if block {
						close(started)
						<-r.Context().Done()
						return nil, r.Context().Err()
					}
					return &Torrent{DownloadPresent: true, Files: []*File{{ID: 2, ShortName: "Episode.mkv"}}}, nil
				})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				finished := make(chan error, 1)
				go func() {
					_, err := provider.GetTorrentStreamUrl(ctx, debrid.StreamTorrentOptions{ID: "1", FileId: "Episode.mkv"}, nil)
					finished <- err
				}()
				<-started
				cancel()
				if err := <-finished; !errors.Is(err, context.Canceled) {
					t.Fatalf("expected cancellation, got %v", err)
				}
			})
		})
	}
}

func TestArchiveDownloadsRemainSupported(t *testing.T) {
	for _, fileID := range []string{"", "Batch.zip"} {
		t.Run("file="+fileID, func(t *testing.T) {
			provider := newTestTorBox(t, func(r *http.Request) (any, error) {
				if strings.HasSuffix(r.URL.Path, "/mylist") {
					return &Torrent{Files: []*File{{ID: 0, ShortName: "Batch.zip"}}}, nil
				}
				if fileID != "" && r.URL.Query().Get("file_id") != "0" {
					t.Fatal("lost numeric ID zero")
				}
				if fileID == "" && r.URL.Query().Get("zip_link") != "true" {
					t.Fatal("lost archive download flag")
				}
				return "https://example.invalid/Batch.zip", nil
			})
			url, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{ID: "1", FileId: fileID})
			if err != nil || url == "" {
				t.Fatalf("explicit archive download must remain supported: %v", err)
			}
		})
	}
}
