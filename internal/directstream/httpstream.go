package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	httputil "seanime/internal/util/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

// httpBaseStream holds shared state and logic for HTTP URL-based streams (debrid, URL, nakama).
type httpBaseStream struct {
	BaseStream
	streamUrl           string
	contentLength       int64
	filepath            string
	requestHeaders      http.Header
	headResponseHeaders http.Header
	httpStream          *httputil.FileStream // Shared file-backed cache for multiple readers
	cacheMu             sync.RWMutex         // Protects httpStream access
}

var videoProxyClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   false, // Fixes issues on Linux
	},
}

// Headers that should not be forwarded to the CDN
var proxyHopHeaders = map[string]bool{
	"Host":                true,
	"Accept":              true,
	"Accept-Encoding":     true,
	"Range":               true,
	"Connection":          true,
	"Proxy-Connection":    true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

func (s *httpBaseStream) applyReqHeaders(dst http.Header) {
	overrideHeaders(dst, s.requestHeaders)
}

func (s *httpBaseStream) applyHeadRespHeaders(dst http.Header) {
	overrideHeaders(dst, s.headResponseHeaders)
}

func (s *httpBaseStream) newMetadataReader() (io.ReadSeekCloser, error) {
	return httputil.NewHttpReadSeekerFromURLWithHeaders(s.streamUrl, s.requestHeaders)
}

func (s *httpBaseStream) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		s.cacheMu.RLock()
		if s.httpStream == nil {
			s.cacheMu.RUnlock()
			_ = s.initializeStream()
		} else {
			s.cacheMu.RUnlock()
		}

		info, ok := s.manager.FetchStreamInfoWithHeaders(s.streamUrl, s.requestHeaders)
		if !ok {
			s.logger.Warn().Str("url", s.streamUrl).Msg("directstream(http): Failed to fetch stream info for content type")
			return
		}
		s.logger.Debug().Str("url", s.streamUrl).Str("contentType", info.ContentType).Int64("contentLength", info.ContentLength).Msg("directstream(http): Fetched content type and length")
		s.contentType = info.ContentType
		if s.contentType == "application/force-download" {
			s.contentType = "application/octet-stream"
		}
		s.contentLength = info.ContentLength
	})

	return s.contentType
}

// Close cleans up the HTTP cache and other resources
func (s *httpBaseStream) Close() error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.logger.Debug().Msg("directstream(http): Closing HTTP cache")

	if s.httpStream != nil {
		if err := s.httpStream.Close(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(http): Failed to close HTTP cache")
			return err
		}
		s.httpStream = nil
	}

	s.logger.Debug().Msg("directstream(http): HTTP cache closed successfully")

	return nil
}

// Terminate overrides BaseStream.Terminate to also clean up the HTTP cache
func (s *httpBaseStream) Terminate() {
	// Clean up HTTP cache first
	if err := s.Close(); err != nil {
		s.logger.Error().Err(err).Msg("directstream(http): Failed to clean up HTTP cache during termination")
	}

	// Call the base implementation
	s.BaseStream.Terminate()
}

// loadPlaybackInfo is called by concrete types, passing their own StreamType.
func (s *httpBaseStream) loadPlaybackInfo(streamType nativeplayer.StreamType) (ret *nativeplayer.PlaybackInfo, err error) {
	s.playbackInfoOnce.Do(func() {
		if s.streamUrl == "" {
			ret = &nativeplayer.PlaybackInfo{}
			err = fmt.Errorf("stream url is not set")
			s.playbackInfoErr = err
			return
		}

		id := uuid.New().String()

		var entryListData *anime.EntryListData
		if animeCollection, ok := s.manager.animeCollection.Get(); ok {
			if listEntry, ok := animeCollection.GetListEntryFromAnimeId(s.media.ID); ok {
				entryListData = anime.NewEntryListData(listEntry)
			}
		}

		contentType := s.LoadContentType()

		playbackInfo := nativeplayer.PlaybackInfo{
			ID:                id,
			StreamType:        streamType,
			StreamPath:        s.filepath,
			MimeType:          contentType,
			StreamUrl:         "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id + s.manager.GetHMACTokenQueryParam("/api/v1/directstream/stream", "&"),
			ContentLength:     s.contentLength, // loaded by LoadContentType
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
			Episode:           s.episode,
			Media:             s.media,
			EntryListData:     entryListData,
		}

		// If the content type is an EBML content type, we can create a metadata parser
		if isEbmlContent(s.LoadContentType()) || s.LoadContentType() == "application/octet-stream" || s.LoadContentType() == "application/force-download" {
			reader, readErr := s.newMetadataReader()
			if readErr != nil {
				err = fmt.Errorf("failed to create reader for stream url: %w", readErr)
				s.logger.Error().Err(readErr).Msg("directstream(http): Failed to create reader for stream url")
				s.playbackInfoErr = err
				return
			}
			defer reader.Close() // Close this specific reader instance

			_, _ = reader.Seek(0, io.SeekStart)
			s.logger.Trace().Msgf("directstream(http): Loading metadata for stream url: %s", s.streamUrl)

			parser := mkvparser.NewMetadataParser(reader, s.logger)
			metadataCtx := s.manager.playbackCtx
			if metadataCtx == nil {
				metadataCtx = context.Background()
			}
			metadata := parser.GetMetadata(metadataCtx)
			if metadata.Error != nil {
				err = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				s.logger.Error().Err(metadata.Error).Msg("directstream(http): Failed to get metadata")
				s.playbackInfoErr = err
				return
			}

			playbackInfo.MkvMetadata = metadata
			playbackInfo.MkvMetadataParser = mo.Some(parser)
		}

		s.playbackInfo = &playbackInfo
	})

	return s.playbackInfo, s.playbackInfoErr
}

// getStreamHandler is called by concrete types, passing themselves as the Stream interface
// so that subtitle streaming uses the correct outer stream.
func (s *httpBaseStream) getStreamHandler(outer Stream) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.streamUrl == "" {
			s.logger.Error().Msg("directstream(http): No URL to stream")
			http.Error(w, "No URL to stream", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodHead {
			s.logger.Trace().Msg("directstream(http): Handling HEAD request")

			fileSize := s.contentLength
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
			w.Header().Set("Content-Type", s.LoadContentType())
			w.Header().Set("Accept-Ranges", "bytes")
			s.applyHeadRespHeaders(w.Header())
			w.WriteHeader(http.StatusOK)
			return
		}

		rangeHeader := r.Header.Get("Range")

		if err := s.initializeStream(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(http): Failed to initialize FileStream")
			http.Error(w, "Failed to initialize FileStream", http.StatusInternalServerError)
			return
		}

		reader, err := s.getReader()
		if err != nil {
			s.logger.Error().Err(err).Msg("directstream(http): Failed to create reader for stream url")
			http.Error(w, "Failed to create reader for stream url", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		if isThumbnailRequest(r) {
			ra, ok := handleRange(w, r, reader, s.filename, s.contentLength)
			if !ok {
				return
			}
			serveContentRange(w, r, r.Context(), reader, s.filename, s.contentLength, s.contentType, ra)
			return
		}

		ra, ok := handleRange(w, r, reader, s.filename, s.contentLength)
		if !ok {
			return
		}

		if _, ok := s.playbackInfo.MkvMetadataParser.Get(); ok {
			subReader, err := s.getReader()
			if err != nil {
				s.logger.Error().Err(err).Msg("directstream(http): Failed to create subtitle reader for stream url")
				http.Error(w, "Failed to create subtitle reader for stream url", http.StatusInternalServerError)
				return
			}
			if ra.Start < s.contentLength-1024*1024 {
				// subReader is closed inside the subtitle goroutine
				go s.StartSubtitleStreamP(outer, s.manager.playbackCtx, subReader, ra.Start, 0)
			} else {
				_ = subReader.Close()
			}
		}

		// Use the client's request context so the CDN request is cancelled when the client disconnects
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, s.streamUrl, nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Accept", "*/*")
		req.Header.Set("Range", rangeHeader)

		// Only forward safe headers to avoid conflicts with the CDN
		for key, values := range r.Header {
			if proxyHopHeaders[http.CanonicalHeaderKey(key)] {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		s.applyReqHeaders(req.Header)

		resp, err := videoProxyClient.Do(req)
		if err != nil {
			s.logger.Error().Err(err).Str("range", rangeHeader).Msg("directstream(http): CDN proxy request failed")
			http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Reject non-2xx CDN responses to avoid corrupting the file cache
		if resp.StatusCode >= 300 {
			s.logger.Error().Int("status", resp.StatusCode).Str("range", rangeHeader).Msg("directstream(http): CDN returned non-2xx status")
			http.Error(w, fmt.Sprintf("CDN error: %d", resp.StatusCode), resp.StatusCode)
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Set(key, value)
			}
		}

		w.Header().Set("Content-Type", s.LoadContentType()) // overwrite the type
		w.WriteHeader(resp.StatusCode)

		if err := s.httpStream.WriteAndFlush(resp.Body, w, ra.Start); err != nil {
			s.logger.Warn().Err(err).Str("range", rangeHeader).Msg("directstream(http): WriteAndFlush error")
		}
	})
}

// initializeStream creates the HTTP cache for this stream if it doesn't exist
func (s *httpBaseStream) initializeStream() error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if s.httpStream != nil {
		return nil // Already initialized
	}

	if s.streamUrl == "" {
		return fmt.Errorf("stream URL is not set")
	}

	// Get content length first
	if s.contentLength == 0 {
		info, ok := s.manager.FetchStreamInfoWithHeaders(s.streamUrl, s.requestHeaders)
		if !ok {
			return fmt.Errorf("failed to fetch stream info")
		}
		s.contentLength = info.ContentLength
	}

	s.logger.Debug().Msgf("directstream(http): Initializing FileStream for stream URL: %s", s.streamUrl)

	// Create a file-backed stream with the known content length
	cache, err := httputil.NewFileStream(s.manager.playbackCtx, s.logger, s.contentLength)
	if err != nil {
		return fmt.Errorf("failed to create FileStream: %w", err)
	}

	s.httpStream = cache

	s.logger.Debug().Msgf("directstream(http): FileStream initialized")

	return nil
}

func (s *httpBaseStream) getReader() (io.ReadSeekCloser, error) {
	if err := s.initializeStream(); err != nil {
		return nil, err
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if s.httpStream == nil {
		return nil, fmt.Errorf("FileStream not initialized")
	}

	return s.httpStream.NewReader()
}
