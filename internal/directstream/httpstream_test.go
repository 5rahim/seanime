package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func newHTTPStreamTestManager() *Manager {
	return &Manager{
		Logger:      util.NewLogger(),
		playbackCtx: context.Background(),
	}
}

func newTestNakamaStream(manager *Manager, streamURL string, token string) *Nakama {
	return &Nakama{
		httpBaseStream: httpBaseStream{
			streamUrl: streamURL,
			requestHeaders: http.Header{
				"X-Seanime-Nakama-Token": []string{token},
			},
			headResponseHeaders: http.Header{
				"X-Seanime-Nakama-Token": []string{token},
			},
			BaseStream: BaseStream{
				manager:               manager,
				logger:                manager.Logger,
				subtitleEventCache:    result.NewMap[string, *mkvparser.SubtitleEvent](),
				activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
			},
		},
	}
}

func TestNakamaLoadContentTypeUsesSharedRequestHeaders(t *testing.T) {
	const token = "nakama-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, token, r.Header.Get("X-Seanime-Nakama-Token"))
		require.Equal(t, http.MethodHead, r.Method)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "6")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := newHTTPStreamTestManager()
	stream := newTestNakamaStream(manager, server.URL+"/video.mp4", token)

	require.Equal(t, "video/mp4", stream.LoadContentType())
	require.Equal(t, int64(6), stream.contentLength)
}

func TestNakamaGetStreamHandlerPreservesHeadResponseToken(t *testing.T) {
	const token = "nakama-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, token, r.Header.Get("X-Seanime-Nakama-Token"))
		require.Equal(t, http.MethodHead, r.Method)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "6")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := newHTTPStreamTestManager()
	stream := newTestNakamaStream(manager, server.URL+"/video.mp4", token)
	stream.playbackInfo = &mediacore.PlaybackInfo{MkvMetadataParser: mo.None[*mkvparser.MetadataParser]()}

	require.Equal(t, "video/mp4", stream.LoadContentType())

	req := httptest.NewRequest(http.MethodHead, "/api/v1/directstream/stream", nil)
	rec := httptest.NewRecorder()

	stream.GetStreamHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, token, rec.Header().Get("X-Seanime-Nakama-Token"))
	require.Equal(t, "video/mp4", rec.Header().Get("Content-Type"))
	require.Equal(t, "6", rec.Header().Get("Content-Length"))
}

func TestNakamaGetStreamHandlerProxiesWithSharedRequestHeaders(t *testing.T) {
	const token = "nakama-secret"
	payload := []byte("abcdef")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, token, r.Header.Get("X-Seanime-Nakama-Token"))

		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			require.Equal(t, "bytes=0-3", r.Header.Get("Range"))
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Length", "4")
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-3/%d", len(payload)))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(payload[:4])
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	manager := newHTTPStreamTestManager()
	stream := newTestNakamaStream(manager, server.URL+"/video.mp4", token)
	stream.playbackInfo = &mediacore.PlaybackInfo{MkvMetadataParser: mo.None[*mkvparser.MetadataParser]()}

	require.Equal(t, "video/mp4", stream.LoadContentType())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/directstream/stream", nil)
	req.Header.Set("Range", "bytes=0-3")
	rec := httptest.NewRecorder()

	stream.GetStreamHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusPartialContent, rec.Code)
	require.Equal(t, "abcd", rec.Body.String())
	require.Equal(t, "video/mp4", rec.Header().Get("Content-Type"))
}

func TestNakamaMetadataReaderCarriesHeadersAcrossRangeRequests(t *testing.T) {
	const token = "nakama-secret"
	payload := []byte("abcdef")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, token, r.Header.Get("X-Seanime-Nakama-Token"))

		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			w.WriteHeader(http.StatusOK)
			return
		}

		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			_, _ = w.Write(payload)
			return
		}

		startStr := strings.TrimPrefix(rangeHeader, "bytes=")
		startStr = strings.TrimSuffix(startStr, "-")
		start, err := strconv.Atoi(startStr)
		require.NoError(t, err)

		w.Header().Set("Content-Length", strconv.Itoa(len(payload)-start))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(payload)-1, len(payload)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload[start:])
	}))
	defer server.Close()

	manager := newHTTPStreamTestManager()
	stream := newTestNakamaStream(manager, server.URL+"/video.mkv", token)

	reader, err := stream.newMetadataReader()
	require.NoError(t, err)
	defer reader.Close()

	first := make([]byte, 3)
	_, err = io.ReadFull(reader, first)
	require.NoError(t, err)
	require.Equal(t, "abc", string(first))

	_, err = reader.Seek(2, io.SeekStart)
	require.NoError(t, err)

	second := make([]byte, 3)
	_, err = io.ReadFull(reader, second)
	require.NoError(t, err)
	require.Equal(t, "cde", string(second))
}
