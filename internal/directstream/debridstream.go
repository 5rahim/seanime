package directstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util"
	httputil "seanime/internal/util/http"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*DebridStream)(nil)

// DebridStream is a stream that is a torrent.
type DebridStream struct {
	BaseStream
	streamUrl     string
	contentLength int64
	torrent       *hibiketorrent.AnimeTorrent
	streamReadyCh chan struct{}        // Closed by the initiator when the stream is ready
	httpStream    *httputil.FileStream // Shared file-backed cache for multiple readers
	cacheMu       sync.RWMutex         // Protects httpStream access
}

func (s *DebridStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeDebrid
}

func (s *DebridStream) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		s.cacheMu.RLock()
		if s.httpStream == nil {
			s.cacheMu.RUnlock()
			_ = s.initializeStream()
		} else {
			s.cacheMu.RUnlock()
		}

		info, ok := s.FetchStreamInfo(s.streamUrl)
		if !ok {
			s.logger.Warn().Str("url", s.streamUrl).Msg("directstream(debrid): Failed to fetch stream info for content type")
			return
		}
		s.logger.Debug().Str("url", s.streamUrl).Str("contentType", info.ContentType).Int64("contentLength", info.ContentLength).Msg("directstream(debrid): Fetched content type and length")
		s.contentType = info.ContentType
		if s.contentType == "application/force-download" {
			s.contentType = "application/octet-stream"
		}
		s.contentLength = info.ContentLength
	})

	return s.contentType
}

// Close cleanup the HTTP cache and other resources
func (s *DebridStream) Close() error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.logger.Debug().Msg("directstream(debrid): Closing HTTP cache")

	if s.httpStream != nil {
		if err := s.httpStream.Close(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(debrid): Failed to close HTTP cache")
			return err
		}
		s.httpStream = nil
	}

	s.logger.Debug().Msg("directstream(debrid): HTTP cache closed successfully")

	return nil
}

// Terminate overrides BaseStream.Terminate to also clean up the HTTP cache
func (s *DebridStream) Terminate() {
	// Clean up HTTP cache first
	if err := s.Close(); err != nil {
		s.logger.Error().Err(err).Msg("directstream(debrid): Failed to clean up HTTP cache during termination")
	}

	// Call the base implementation
	s.BaseStream.Terminate()
}

func (s *DebridStream) LoadPlaybackInfo() (ret *nativeplayer.PlaybackInfo, err error) {
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
			StreamType:        s.Type(),
			MimeType:          contentType,
			StreamUrl:         "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id,
			ContentLength:     s.contentLength, // loaded by LoadContentType
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
			Episode:           s.episode,
			Media:             s.media,
			EntryListData:     entryListData,
		}

		// If the content type is an EBML content type, we can create a metadata parser
		if isEbmlContent(s.LoadContentType()) {
			reader, err := httputil.NewHttpReadSeekerFromURL(s.streamUrl)
			//reader, err := s.getPriorityReader()
			if err != nil {
				err = fmt.Errorf("failed to create reader for stream url: %w", err)
				s.logger.Error().Err(err).Msg("directstream(debrid): Failed to create reader for stream url")
				s.playbackInfoErr = err
				return
			}
			defer reader.Close() // Close this specific reader instance

			_, _ = reader.Seek(0, io.SeekStart)
			s.logger.Trace().Msgf(
				"directstream(debrid): Loading metadata for stream url: %s",
				s.streamUrl,
			)

			parser := mkvparser.NewMetadataParser(reader, s.logger)
			metadata := parser.GetMetadata(context.Background())
			if metadata.Error != nil {
				err = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				s.logger.Error().Err(metadata.Error).Msg("directstream(debrid): Failed to get metadata")
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

func (s *DebridStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

var videoProxyClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false, // Fixes issues on Linux
	},
	Timeout: 60 * time.Second,
}

func (s *DebridStream) GetStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Trace().Str("range", r.Header.Get("Range")).Str("method", r.Method).Msg("directstream(debrid): Stream endpoint hit")

		if s.streamUrl == "" {
			s.logger.Error().Msg("directstream(debrid): No URL to stream")
			http.Error(w, "No URL to stream", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodHead {
			s.logger.Trace().Msg("directstream(debrid): Handling HEAD request")

			fileSize := s.contentLength
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
			w.Header().Set("Content-Type", s.LoadContentType())
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusOK)
			return
		}

		rangeHeader := r.Header.Get("Range")

		if err := s.initializeStream(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(debrid): Failed to initialize FileStream")
			http.Error(w, "Failed to initialize FileStream", http.StatusInternalServerError)
			return
		}

		reader, err := s.getReader()
		if err != nil {
			s.logger.Error().Err(err).Msg("directstream(debrid): Failed to create reader for stream url")
			http.Error(w, "Failed to create reader for stream url", http.StatusInternalServerError)
		}

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
				s.logger.Error().Err(err).Msg("directstream(debrid): Failed to create subtitle reader for stream url")
				http.Error(w, "Failed to create subtitle reader for stream url", http.StatusInternalServerError)
				return
			}
			if ra.Start < s.contentLength-1024*1024 {
				go s.StartSubtitleStreamP(s, s.manager.playbackCtx, subReader, ra.Start, 0)
			}
		}

		req, err := http.NewRequest(http.MethodGet, s.streamUrl, nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Accept", "*/*")
		req.Header.Set("Range", rangeHeader)

		// Copy original request headers to the proxied request
		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := videoProxyClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Set(key, value)
			}
		}

		w.Header().Set("Content-Type", s.LoadContentType()) // overwrite the type
		w.WriteHeader(resp.StatusCode)

		_ = s.httpStream.WriteAndFlush(resp.Body, w, ra.Start)
	})
}

//func (s *DebridStream) GetStreamHandler() http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		s.logger.Trace().Str("range", r.Header.Get("Range")).Str("method", r.Method).Msg("directstream(debrid): Stream endpoint hit")
//
//		if s.streamUrl == "" {
//			s.logger.Error().Msg("directstream(debrid): No URL to stream")
//			http.Error(w, "No URL to stream", http.StatusNotFound)
//			return
//		}
//
//		// Handle HEAD requests explicitly to provide file size information
//		if r.Method == http.MethodHead {
//			s.logger.Trace().Msg("directstream(debrid): Handling HEAD request")
//
//			// Set the content length from torrent file
//			fileSize := s.contentLength
//			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
//			w.Header().Set("Content-Type", s.LoadContentType())
//			w.Header().Set("Accept-Ranges", "bytes")
//			w.WriteHeader(http.StatusOK)
//			return
//		}
//
//		rangeHeader := r.Header.Get("Range")
//
//		// Parse the range header
//		ranges, err := httputil.ParseRange(rangeHeader, s.contentLength)
//		if err != nil && !errors.Is(err, httputil.ErrNoOverlap) {
//			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", s.contentLength))
//			http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
//			return
//		} else if err != nil && errors.Is(err, httputil.ErrNoOverlap) {
//			// Let Go handle overlap
//			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", s.contentLength))
//		}
//
//		// Initialize the FileStream
//		if err := s.initializeStream(); err != nil {
//			s.logger.Error().Err(err).Msg("directstream(debrid): Failed to initialize FileStream")
//			http.Error(w, "Failed to initialize FileStream", http.StatusInternalServerError)
//			return
//		}
//
//		// Determine the offset for the HTTP request and FileStream
//		var httpRequestOffset int64 = 0
//		var fileWriteOffset int64 = 0
//		var httpResponseOffset int64 = 0
//
//		if len(ranges) > 0 {
//			originalOffset := ranges[0].Start
//			// Start HTTP request 1MB earlier to ensure subtitle clusters are available
//			const bufferSize = 1024 * 1024 // 1MB
//			httpRequestOffset = originalOffset - bufferSize
//			if httpRequestOffset < 0 {
//				httpRequestOffset = 0
//			}
//			fileWriteOffset = httpRequestOffset
//			httpResponseOffset = originalOffset - httpRequestOffset
//		}
//
//		// Update the range header for the actual HTTP request
//		var actualRangeHeader string
//		if len(ranges) > 0 {
//			if httpRequestOffset != ranges[0].Start {
//				// Create a new range header starting from the earlier offset
//				endOffset := ranges[0].Start + ranges[0].Length - 1
//				if endOffset >= s.contentLength {
//					endOffset = s.contentLength - 1
//				}
//				actualRangeHeader = fmt.Sprintf("bytes=%d-%d", httpRequestOffset, endOffset)
//			} else {
//				actualRangeHeader = rangeHeader
//			}
//		}
//
//		// Create HTTP request for the range
//		req, err := http.NewRequest(http.MethodGet, s.streamUrl, nil)
//		if err != nil {
//			http.Error(w, "Failed to create request", http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", s.LoadContentType())
//		w.Header().Set("Accept-Ranges", "bytes")
//		w.Header().Set("Connection", "keep-alive")
//		w.Header().Set("Cache-Control", "no-store")
//
//		// Copy original request headers to the proxied request
//		for key, values := range r.Header {
//			for _, value := range values {
//				req.Header.Add(key, value)
//			}
//		}
//
//		req.Header.Set("Accept", "*/*")
//		req.Header.Set("Range", actualRangeHeader)
//
//		// Make the HTTP request
//		resp, err := videoProxyClient.Do(req)
//		if err != nil {
//			http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
//			return
//		}
//		defer resp.Body.Close()
//
//		if _, ok := s.playbackInfo.MkvMetadataParser.Get(); ok {
//			// Start a subtitle stream from the current position using normal reader (no prefetching)
//			subReader, err := s.getReader()
//			if err != nil {
//				s.logger.Error().Err(err).Msg("directstream(debrid): Failed to create subtitle reader for stream url")
//				http.Error(w, "Failed to create subtitle reader for stream url", http.StatusInternalServerError)
//				return
//			}
//			// Do not start stream if start if 1MB from the end
//			if len(ranges) > 0 && ranges[0].Start < s.contentLength-1024*1024 {
//				go s.StartSubtitleStream(s, s.manager.playbackCtx, subReader, ranges[0].Start)
//			}
//		}
//
//		// Copy response headers but adjust Content-Range if we modified the range
//		for key, values := range resp.Header {
//			if key == "Content-Type" {
//				continue
//			}
//			if key == "Content-Range" && httpResponseOffset > 0 {
//				// Adjust the Content-Range header to reflect the original request
//				continue // We'll set this manually below
//			}
//			if key == "Content-Length" && httpResponseOffset > 0 {
//				continue
//			}
//			for _, value := range values {
//				w.Header().Set(key, value)
//			}
//		}
//
//		// Set the correct Content-Range header for the original request
//		if len(ranges) > 0 && httpResponseOffset > 0 {
//			originalRange := ranges[0]
//			w.Header().Set("Content-Range", originalRange.ContentRange(s.contentLength))
//			w.Header().Set("Content-Length", fmt.Sprintf("%d", s.contentLength))
//		}
//
//		// Set the status code
//		w.WriteHeader(resp.StatusCode)
//
//		// Create a custom writer that skips the buffer bytes for HTTP response
//		httpWriter := &offsetWriter{
//			writer:    w,
//			skipBytes: httpResponseOffset,
//			skipped:   0,
//		}
//
//		// Use FileStream's WriteAndFlush to write all data to file but only desired range to HTTP response
//		err = s.httpStream.WriteAndFlush(resp.Body, httpWriter, fileWriteOffset)
//		if err != nil {
//			s.logger.Error().Err(err).Msg("directstream(debrid): Failed to stream response body")
//			http.Error(w, "Failed to stream response body", http.StatusInternalServerError)
//			return
//		}
//	})
//}

type PlayDebridStreamOptions struct {
	StreamUrl     string
	MediaId       int
	EpisodeNumber int    // RELATIVE Episode number to identify the file
	AnidbEpisode  string // Anizip episode
	Media         *anilist.BaseAnime
	Torrent       *hibiketorrent.AnimeTorrent // Selected torrent
	FileId        string                      // File ID or index
	UserAgent     string
	ClientId      string
	AutoSelect    bool
}

// PlayDebridStream is used by a module to load a new torrent stream.
func (m *Manager) PlayDebridStream(ctx context.Context, opts PlayDebridStreamOptions) error {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:    nil,
		Media:            opts.Media,
		MetadataProvider: m.metadataProvider,
		Logger:           m.Logger,
	})
	if err != nil {
		return fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
	}

	episode, ok := episodeCollection.FindEpisodeByAniDB(opts.AnidbEpisode)
	if !ok {
		return fmt.Errorf("cannot play torrent stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &DebridStream{
		streamUrl: opts.StreamUrl,
		torrent:   opts.Torrent,
		BaseStream: BaseStream{
			manager:               m,
			logger:                m.Logger,
			clientId:              opts.ClientId,
			media:                 opts.Media,
			filename:              "",
			episode:               episode,
			episodeCollection:     episodeCollection,
			subtitleEventCache:    result.NewResultMap[string, *mkvparser.SubtitleEvent](),
			activeSubtitleStreams: result.NewResultMap[string, *SubtitleStream](),
		},
		streamReadyCh: make(chan struct{}),
	}

	go func() {
		m.loadStream(stream)
	}()

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// initializeStream creates the HTTP cache for this stream if it doesn't exist
func (s *DebridStream) initializeStream() error {
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
		info, ok := s.FetchStreamInfo(s.streamUrl)
		if !ok {
			return fmt.Errorf("failed to fetch stream info")
		}
		s.contentLength = info.ContentLength
	}

	s.logger.Debug().Msgf("directstream(debrid): Initializing FileStream for stream URL: %s", s.streamUrl)

	// Create a file-backed stream with the known content length
	cache, err := httputil.NewFileStream(s.manager.playbackCtx, s.logger, s.contentLength)
	if err != nil {
		return fmt.Errorf("failed to create FileStream: %w", err)
	}

	s.httpStream = cache

	s.logger.Debug().Msgf("directstream(debrid): FileStream initialized")

	return nil
}

func (s *DebridStream) getReader() (io.ReadSeekCloser, error) {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// offsetWriter is a wrapper that skips a specified number of bytes before writing to the underlying writer
type offsetWriter struct {
	writer    io.Writer
	skipBytes int64
	skipped   int64
}

func (ow *offsetWriter) Write(p []byte) (n int, err error) {
	if ow.skipped < ow.skipBytes {
		// We still need to skip some bytes
		remaining := ow.skipBytes - ow.skipped
		if int64(len(p)) <= remaining {
			// Skip all of this write
			ow.skipped += int64(len(p))
			return len(p), nil
		} else {
			// Skip part of this write and write the rest
			skipCount := remaining
			ow.skipped = ow.skipBytes
			return ow.writer.Write(p[skipCount:])
		}
	}
	// No more skipping needed, write everything
	return ow.writer.Write(p)
}

func fetchContentLength(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch content length: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	if contentLength < 0 {
		return 0, errors.New("content length not provided")
	}

	return contentLength, nil
}

type StreamInfo struct {
	ContentType   string
	ContentLength int64
}

func (s *DebridStream) FetchStreamInfo(streamUrl string) (info *StreamInfo, canStream bool) {
	hasExtension, isArchive := IsArchive(streamUrl)

	// If we were able to verify that the stream URL is an archive, we can't stream it
	if isArchive {
		s.logger.Warn().Str("url", streamUrl).Msg("directstream(debrid): Stream URL is an archive, cannot stream")
		return nil, false
	}

	// If the stream URL has an extension, we can stream it
	if hasExtension {
		ext := filepath.Ext(streamUrl)
		// If not a valid video extension, we can't stream it
		if !util.IsValidVideoExtension(ext) {
			s.logger.Warn().Str("url", streamUrl).Str("ext", ext).Msg("directstream(debrid): Stream URL has an invalid video extension, cannot stream")
			return nil, false
		}
	}

	// We'll fetch headers to get the info
	// If the headers are not available, we can't stream it

	contentType, contentLength, err := s.GetContentTypeAndLength(streamUrl)
	if err != nil {
		s.logger.Error().Err(err).Str("url", streamUrl).Msg("directstream(debrid): Failed to fetch content type and length")
		return nil, false
	}

	// If not a video content type, we can't stream it
	if !strings.HasPrefix(contentType, "video/") && contentType != "application/octet-stream" && contentType != "application/force-download" {
		s.logger.Warn().Str("url", streamUrl).Str("contentType", contentType).Msg("directstream(debrid): Stream URL has an invalid content type, cannot stream")
		return nil, false
	}

	return &StreamInfo{
		ContentType:   contentType,
		ContentLength: contentLength,
	}, true
}

func IsArchive(streamUrl string) (hasExtension bool, isArchive bool) {
	ext := filepath.Ext(streamUrl)
	if ext == ".zip" || ext == ".rar" {
		return true, true
	}

	if ext != "" {
		return true, false
	}

	return false, false
}

func GetContentTypeAndLengthHead(url string) (string, string) {
	resp, err := http.Head(url)
	if err != nil {
		return "", ""
	}

	defer resp.Body.Close()

	return resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length")
}

func (s *DebridStream) GetContentTypeAndLength(url string) (string, int64, error) {
	// Try using HEAD request
	cType, cLength := GetContentTypeAndLengthHead(url)

	length, err := strconv.ParseInt(cLength, 10, 64)
	if err != nil && cLength != "" {
		s.logger.Error().Err(err).Str("contentType", cType).Str("contentLength", cLength).Msg("directstream(debrid): Failed to parse content length from header")
		return "", 0, fmt.Errorf("failed to parse content length: %w", err)
	}

	if cType != "" {
		return cType, length, nil
	}

	s.logger.Trace().Msg("directstream(debrid): Content type not found in headers, falling back to GET request")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}

	// Only read a small amount of data to determine the content type.
	req.Header.Set("Range", "bytes=0-511")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	// Read the first 512 bytes
	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return "", 0, err
	}

	// Detect content type based on the read bytes
	contentType := http.DetectContentType(buf[:n])

	return contentType, length, nil
}
