package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mediastream/mkvparser"
	"seanime/internal/mediastream/nativeplayer"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"seanime/internal/util/torrentutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

// Stream is the common interface for all stream types.
type Stream interface {
	// Type returns the type of the stream.
	Type() nativeplayer.StreamType
	// LoadContentType loads and returns the content type of the stream.
	// e.g. "video/mp4", "video/webm", "video/x-matroska"
	LoadContentType() string
	// ClientId returns the client ID of the current stream.
	ClientId() string
	// Media returns the media of the current stream.
	Media() *anilist.BaseAnime
	// Episode returns the episode of the current stream.
	Episode() *anime.Episode
	// EpisodeCollection returns the episode collection for the media of the current stream.
	EpisodeCollection() *anime.EpisodeCollection
	// LoadPlaybackInfo loads and returns the playback info.
	LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error)
	// GetAttachmentByName returns the attachment by name for the stream.
	// It is used to serve fonts and other attachments.
	GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool)
	// GetStreamHandler returns the stream handler.
	GetStreamHandler() http.Handler
	// ServeSubtitles starts the subtitle stream at the given position.
	// Only if the stream supports it.
	ServeSubtitles(start int64)
	// ServeSubtitlesFromTime starts subtitle extraction from a video timestamp.
	// This is the preferred method for video player events.
	ServeSubtitlesFromTime(timeSeconds float64)
	// StreamError is called when an error occurs while streaming.
	// This is used to notify the native player that an error occurred.
	// It will close the stream.
	StreamError(err error)
	// Terminate ends the stream.
	// Once this is called, the stream should not be used anymore.
	Terminate()
	// GetSubtitleEventCache accesses the subtitle event cache.
	GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent]
}

func (m *Manager) getStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		stream, ok := m.currentStream.Get()
		if !ok {
			http.Error(w, "no stream", http.StatusNotFound)
			return
		}
		m.Logger.Debug().Msgf("directstream: Getting stream handler for %s", stream.Episode().EpisodeTitle)
		stream.GetStreamHandler().ServeHTTP(w, r)
	})
}

// loadStream loads a new stream and cancels the previous one.
// Caller should use mutex to lock the manager.
func (m *Manager) loadStream(stream Stream) {
	// Cancel the previous playback
	if m.playbackCtxCancelFunc != nil {
		m.Logger.Trace().Msgf("directstream: Cancelling previous playback")
		m.playbackCtxCancelFunc()
		m.playbackCtxCancelFunc = nil
	}

	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	m.playbackCtx = ctx
	m.playbackCtxCancelFunc = cancel

	m.Logger.Debug().Msgf("directstream: Loading stream")
	m.currentStream = mo.Some(stream)

	m.Logger.Debug().Msgf("directstream: Signaling native player that a new stream is starting")
	// Signal the native player that a new stream is starting
	m.nativePlayer.OpenAndAwait(stream.ClientId(), "Checking content type...")

	m.Logger.Debug().Msgf("directstream: Loading content type")
	// Load the content type
	contentType := stream.LoadContentType()
	if contentType == "" {
		m.Logger.Error().Msg("directstream: Failed to load content type")
		m.preStreamError(stream, fmt.Errorf("failed to load content type"))
		return
	}

	m.Logger.Debug().Msgf("directstream: Signaling native player that metadata is being loaded")
	m.nativePlayer.OpenAndAwait(stream.ClientId(), "Loading metadata...")

	// Load the playback info
	// If EBML, it will block until the metadata is parsed
	playbackInfo, err := stream.LoadPlaybackInfo()
	if err != nil {
		m.Logger.Error().Err(err).Msg("directstream: Failed to load playback info")
		m.preStreamError(stream, fmt.Errorf("failed to load playback info: %w", err))
		return
	}

	// Shut the mkv parser logger
	//parser, ok := playbackInfo.MkvMetadataParser.Get()
	//if ok {
	//	parser.SetLoggerEnabled(false)
	//}

	m.Logger.Debug().Msgf("directstream: Signaling native player that stream is ready")
	m.nativePlayer.Watch(stream.ClientId(), playbackInfo)

	// Start the stream loop
	m.streamLoop(ctx, stream)
}

func (m *Manager) streamLoop(ctx context.Context, stream Stream) {
	go func() {
		defer func() {
			m.Logger.Trace().Msg("directstream: Stream loop goroutine exited")
		}()

		for {
			select {
			case <-ctx.Done():
				m.Logger.Debug().Msg("directstream: Stream loop cancelled")
				return
			case event := <-m.nativePlayerSubscriber.Events():
				if event.GetClientId() != stream.ClientId() {
					continue
				}

				cs, ok := m.currentStream.Get()
				if !ok {
					continue
				}

				_ = cs

				switch event := event.(type) {
				case *nativeplayer.VideoStartedEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: %s", event)
				case *nativeplayer.VideoPausedEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: %s", event)
				case *nativeplayer.VideoResumedEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: %s", event)
				case *nativeplayer.VideoEndedEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: %s", event)
				case *nativeplayer.VideoSeekedEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: VideoSeeked, CurrentTime: %f", event.CurrentTime)
					// Convert video timestamp to byte offset for subtitle extraction
					if event.CurrentTime > 0 {
						cs.ServeSubtitlesFromTime(event.CurrentTime)
					}
				case *nativeplayer.VideoLoadedMetadataEvent:
					m.Logger.Debug().Msgf("directstream: Stream event: VideoLoadedMetadata")
					// Start subtitle extraction from the beginning
					cs.ServeSubtitlesFromTime(0.0)
				}
			}
		}
	}()
}

func (m *Manager) unloadStream() {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	// Cancel any existing playback context
	if m.playbackCtxCancelFunc != nil {
		m.Logger.Trace().Msgf("directstream: Cancelling previous playback in unloadStream")
		m.playbackCtxCancelFunc()
	}

	// Clear the current stream
	if stream, ok := m.currentStream.Get(); ok {
		m.Logger.Debug().Msgf("directstream: Terminating current stream in unloadStream")
		stream.Terminate()
	}

	m.currentStream = mo.None[Stream]()
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BaseStream struct {
	logger                 *zerolog.Logger
	clientId               string
	contentType            string
	contentTypeOnce        sync.Once
	episode                *anime.Episode
	media                  *anilist.BaseAnime
	episodeCollection      *anime.EpisodeCollection
	playbackInfo           *nativeplayer.PlaybackInfo
	playbackInfoErr        error
	playbackInfoOnce       sync.Once
	subtitleEventCache     *result.Map[string, *mkvparser.SubtitleEvent]
	terminateOnce          sync.Once
	serveContentCancelFunc context.CancelFunc

	// Subtitle stream management
	subtitleStreamsMu     sync.RWMutex
	activeSubtitleStreams map[string]context.CancelFunc // key: stream identifier, value: cancel function

	manager *Manager
}

func (s *BaseStream) Media() *anilist.BaseAnime {
	return s.media
}

func (s *BaseStream) Episode() *anime.Episode {
	return s.episode
}

func (s *BaseStream) EpisodeCollection() *anime.EpisodeCollection {
	return s.episodeCollection
}

func (s *BaseStream) ClientId() string {
	return s.clientId
}

func (s *BaseStream) Terminate() {
	s.terminateOnce.Do(func() {
		// Cancel the playback context
		// This will snowball and cancel other stuff
		if s.manager.playbackCtxCancelFunc != nil {
			s.manager.playbackCtxCancelFunc()
		}

		// Cancel all active subtitle streams
		s.subtitleStreamsMu.Lock()
		for streamId, cancel := range s.activeSubtitleStreams {
			cancel()
			delete(s.activeSubtitleStreams, streamId)
		}
		s.subtitleStreamsMu.Unlock()

		s.subtitleEventCache.Clear()
	})
}

func (s *BaseStream) StreamError(err error) {
	s.logger.Error().Err(err).Msg("directstream: Stream error occurred")
	s.manager.nativePlayer.Error(s.clientId, err)
	s.Terminate()
	s.manager.unloadStream()
}

func (s *BaseStream) GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent] {
	return s.subtitleEventCache
}

// addSubtitleStream adds a new subtitle stream and returns its identifier
func (s *BaseStream) addSubtitleStream(startOffset int64, cancel context.CancelFunc) string {
	streamId := fmt.Sprintf("%s_%d_%d", s.clientId, startOffset, time.Now().UnixNano())

	s.subtitleStreamsMu.Lock()
	if s.activeSubtitleStreams == nil {
		s.activeSubtitleStreams = make(map[string]context.CancelFunc)
	}
	s.activeSubtitleStreams[streamId] = cancel
	streamCount := len(s.activeSubtitleStreams)
	s.subtitleStreamsMu.Unlock()

	s.manager.Logger.Debug().
		Str("streamId", streamId).
		Int64("startOffset", startOffset).
		Int("totalActiveStreams", streamCount).
		Msg("directstream: Added new subtitle stream")

	return streamId
}

// removeSubtitleStream removes a subtitle stream from tracking
func (s *BaseStream) removeSubtitleStream(streamId string) {
	s.subtitleStreamsMu.Lock()
	delete(s.activeSubtitleStreams, streamId)
	s.subtitleStreamsMu.Unlock()
}

// getActiveSubtitleStreamCount returns the number of currently active subtitle streams
func (s *BaseStream) getActiveSubtitleStreamCount() int {
	s.subtitleStreamsMu.RLock()
	defer s.subtitleStreamsMu.RUnlock()
	return len(s.activeSubtitleStreams)
}

// cancelAllSubtitleStreams cancels all active subtitle streams
// This is used when starting a new subtitle stream from a video event
func (s *BaseStream) cancelAllSubtitleStreams() {
	s.subtitleStreamsMu.Lock()
	defer s.subtitleStreamsMu.Unlock()

	if s.activeSubtitleStreams == nil {
		return
	}

	cancelledCount := len(s.activeSubtitleStreams)

	for streamId, cancel := range s.activeSubtitleStreams {
		s.manager.Logger.Debug().
			Str("streamId", streamId).
			Msg("directstream: Cancelling subtitle stream for new video event")
		cancel()
		delete(s.activeSubtitleStreams, streamId)
	}

	if cancelledCount > 0 {
		s.manager.Logger.Debug().
			Int("cancelledStreams", cancelledCount).
			Msg("directstream: Cancelled all subtitle streams for new video event")
	}
}

// cancelPreviousSubtitleStreams cancels subtitle streams that are no longer relevant
func (s *BaseStream) cancelPreviousSubtitleStreams(newStartOffset int64) {
	s.subtitleStreamsMu.Lock()
	defer s.subtitleStreamsMu.Unlock()

	if s.activeSubtitleStreams == nil {
		return
	}

	initialCount := len(s.activeSubtitleStreams)

	// Cancel streams that are significantly behind the new seek position
	// Keep streams that are close to the new position for smoother playback
	const offsetTolerance = 5 * 1024 * 1024 // 5MB tolerance

	cancelledStreams := 0
	for streamId, cancel := range s.activeSubtitleStreams {
		// Extract offset from streamId (format: "clientId_offset_timestamp")
		parts := strings.Split(streamId, "_")
		if len(parts) >= 3 {
			if oldOffset, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				// Cancel if the old stream is too far behind
				if newStartOffset-oldOffset > offsetTolerance {
					s.manager.Logger.Debug().
						Int64("oldOffset", oldOffset).
						Int64("newOffset", newStartOffset).
						Str("streamId", streamId).
						Msg("directstream: Cancelling old subtitle stream")
					cancel()
					delete(s.activeSubtitleStreams, streamId)
					cancelledStreams++
				}
			}
		}
	}

	if cancelledStreams > 0 {
		s.manager.Logger.Debug().
			Int("cancelledStreams", cancelledStreams).
			Int("remainingStreams", len(s.activeSubtitleStreams)).
			Int("initialStreams", initialCount).
			Int64("newOffset", newStartOffset).
			Msg("directstream: Subtitle stream cleanup completed")
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local File
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*LocalFileStream)(nil)

// LocalFileStream is a stream that is a local file.
type LocalFileStream struct {
	BaseStream
	localFile *anime.LocalFile
}

func (s *LocalFileStream) newReader() (io.ReadSeekCloser, error) {
	r, err := os.OpenFile(s.localFile.Path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *LocalFileStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeFile
}

func (s *LocalFileStream) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		// No need to pass a reader because we are not going to read the file
		// Get the mime type from the file extension
		s.contentType = loadContentType(s.localFile.Path)
	})

	return s.contentType
}

func (s *LocalFileStream) LoadPlaybackInfo() (ret *nativeplayer.PlaybackInfo, err error) {
	s.playbackInfoOnce.Do(func() {
		if s.localFile == nil {
			s.playbackInfo = &nativeplayer.PlaybackInfo{}
			err = fmt.Errorf("local file is not set")
			s.playbackInfoErr = err
			return
		}

		// Open the file
		fr, err := s.newReader()
		if err != nil {
			s.logger.Error().Err(err).Msg("directstream(file): Failed to open local file")
			s.manager.preStreamError(s, fmt.Errorf("cannot stream local file: %w", err))
			return
		}

		// Close the file when done
		defer func() {
			if closer, ok := fr.(io.Closer); ok {
				s.logger.Trace().Msg("directstream(file): Closing local file reader")
				_ = closer.Close()
			} else {
				s.logger.Trace().Msg("directstream(file): Local file reader does not implement io.Closer")
			}
		}()

		// Get the file size
		size, err := fr.Seek(0, io.SeekEnd)
		if err != nil {
			s.logger.Error().Err(err).Msg("directstream(file): Failed to get file size")
			s.manager.preStreamError(s, fmt.Errorf("failed to get file size: %w", err))
			return
		}
		_, _ = fr.Seek(0, io.SeekStart)

		id := uuid.New().String()

		playbackInfo := nativeplayer.PlaybackInfo{
			ID:                id,
			StreamType:        s.Type(),
			MimeType:          s.LoadContentType(),
			StreamUrl:         "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id,
			ContentLength:     size,
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
		}

		// If the content type is an EBML content type, we can create a metadata parser
		if isEbmlContent(s.LoadContentType()) {

			parserKey := util.Base64EncodeStr(s.localFile.Path)

			parser, ok := s.manager.parserCache.Get(parserKey)
			if !ok {
				parser = mkvparser.NewMetadataParser(fr, s.logger)
				s.manager.parserCache.SetT(parserKey, parser, 2*time.Hour)
			}

			metadata := parser.GetMetadata(context.Background())
			if metadata.Error != nil {
				s.logger.Error().Err(metadata.Error).Msg("directstream(torrent): Failed to get metadata")
				s.manager.preStreamError(s, fmt.Errorf("failed to get metadata: %w", metadata.Error))
				s.playbackInfoErr = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				return
			}

			playbackInfo.MkvMetadata = metadata
			playbackInfo.MkvMetadataParser = mo.Some(parser)
		}

		s.playbackInfo = &playbackInfo
	})

	return s.playbackInfo, s.playbackInfoErr
}

func (s *LocalFileStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

func (s *LocalFileStream) ServeSubtitles(start int64) {
	// Cancel streams that are no longer relevant to prevent resource exhaustion
	// Use byte-offset based cancellation since this might be called from HTTP ranges
	s.cancelPreviousSubtitleStreams(start)

	container, err := s.LoadPlaybackInfo()
	if err != nil {
		s.logger.Error().Err(err).Msg("directstream(file): Failed to load playback info")
		return
	}

	// Check that it has mkv metadata
	parser, ok := container.MkvMetadataParser.Get()
	if !ok {
		// No mkv metadata, nothing to stream
		return
	}

	// Create a new context
	ctx, cancel := context.WithCancel(s.manager.playbackCtx)

	// Register this subtitle stream
	streamId := s.addSubtitleStream(start, cancel)

	// Get a new reader
	subReader, err := s.newReader()
	if err != nil {
		s.logger.Error().Err(err).Msg("directstream(file): Failed to open local file")
		s.removeSubtitleStream(streamId)
		cancel()
		return
	}

	s.manager.streamSubtitles(ctx, s, parser, subReader, start, func() {
		s.removeSubtitleStream(streamId)
		cancel()
	})
}

func (s *LocalFileStream) ServeSubtitlesFromTime(timeSeconds float64) {
	// Cancel all previous subtitle streams since this is a new video event
	s.cancelAllSubtitleStreams()

	// Convert time to approximate byte offset
	// For now, if seeking to time 0, start from beginning
	// For non-zero times, we'll need to estimate or use metadata
	var byteOffset int64 = 0

	if timeSeconds > 0 {
		// For time-based seeking, we can estimate position based on file size and duration
		playbackInfo, err := s.LoadPlaybackInfo()
		if err == nil && playbackInfo.MkvMetadata != nil && playbackInfo.MkvMetadata.Duration > 0 && playbackInfo.ContentLength > 0 {
			// Estimate byte position based on time ratio
			// This is approximate but should get us close to the right cluster
			timeRatio := timeSeconds / playbackInfo.MkvMetadata.Duration

			byteOffset = int64(float64(playbackInfo.ContentLength) * timeRatio)
			s.logger.Debug().
				Float64("timeSeconds", timeSeconds).
				Float64("duration", playbackInfo.MkvMetadata.Duration).
				Int64("estimatedOffset", byteOffset).
				Msg("directstream: Converting time to byte offset for subtitle extraction")
		}
	}

	s.ServeSubtitles(byteOffset)
}

func (s *LocalFileStream) GetStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Trace().Str("method", r.Method).Msg("directstream: Received request")

		defer func() {
			s.logger.Trace().Msg("directstream: Request finished")
		}()

		if r.Method == http.MethodHead {
			// Get the file size
			fileInfo, err := os.Stat(s.localFile.Path)
			if err != nil {
				s.logger.Error().Msg("directstream: Failed to get file info")
				http.Error(w, "Failed to get file info", http.StatusInternalServerError)
				return
			}

			// Set the content length
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
			w.Header().Set("Content-Type", s.LoadContentType())
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", s.localFile.Path))
			w.WriteHeader(http.StatusOK)
		} else {
			ServeLocalFile(w, r, s)
		}
	})
}

type PlayLocalFileOptions struct {
	ClientId   string
	Path       string
	LocalFiles []*anime.LocalFile
}

// PlayLocalFile is used by a module to load a new torrent stream.
func (m *Manager) PlayLocalFile(opts PlayLocalFileOptions) error {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	animeCollection, ok := m.animeCollection.Get()
	if !ok {
		return fmt.Errorf("cannot play local file, anime collection is not set")
	}

	// Get the local file
	var lf *anime.LocalFile
	for _, l := range opts.LocalFiles {
		if util.NormalizePath(l.Path) == util.NormalizePath(opts.Path) {
			lf = l
			break
		}
	}

	if lf == nil {
		return fmt.Errorf("cannot play local file, could not find local file: %s", opts.Path)
	}

	if lf.MediaId == 0 {
		return fmt.Errorf("local file has not been matched to a media: %s", opts.Path)
	}

	mId := lf.MediaId
	var media *anilist.BaseAnime
	listEntry, ok := animeCollection.GetListEntryFromAnimeId(mId)
	if ok {
		media = listEntry.Media
	}

	if media == nil {
		return fmt.Errorf("media not found in anime collection: %d", mId)
	}

	episodeCollection, err := anime.NewEpisodeCollectionFromLocalFiles(anime.NewEpisodeCollectionFromLocalFilesOptions{
		LocalFiles:       opts.LocalFiles,
		Media:            media,
		AnimeCollection:  animeCollection,
		Platform:         m.platform,
		MetadataProvider: m.metadataProvider,
		Logger:           m.Logger,
	})
	if err != nil {
		return fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
	}

	var episode *anime.Episode
	for _, e := range episodeCollection.Episodes {
		if e.LocalFile != nil && util.NormalizePath(e.LocalFile.Path) == util.NormalizePath(lf.Path) {
			episode = e
			break
		}
	}

	if episode == nil {
		return fmt.Errorf("cannot play local file, could not find episode for local file: %s", opts.Path)
	}

	stream := &LocalFileStream{
		localFile: lf,
		BaseStream: BaseStream{
			manager:            m,
			logger:             m.Logger,
			clientId:           opts.ClientId,
			media:              media,
			episode:            episode,
			episodeCollection:  episodeCollection,
			subtitleEventCache: result.NewResultMap[string, *mkvparser.SubtitleEvent](),
		},
	}

	m.loadStream(stream)

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*TorrentStream)(nil)

// TorrentStream is a stream that is a torrent.
type TorrentStream struct {
	BaseStream
	torrent *torrent.Torrent
	file    *torrent.File
}

func (s *TorrentStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeTorrent
}

func (s *TorrentStream) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		r := s.file.NewReader()
		defer r.Close()
		s.contentType = loadContentType(s.file.DisplayPath(), r)
	})

	return s.contentType
}

func (s *TorrentStream) LoadPlaybackInfo() (ret *nativeplayer.PlaybackInfo, err error) {
	s.playbackInfoOnce.Do(func() {
		if s.file == nil || s.torrent == nil {
			ret = &nativeplayer.PlaybackInfo{}
			err = fmt.Errorf("torrent is not set")
			s.playbackInfoErr = err
			return
		}

		id := uuid.New().String()

		playbackInfo := nativeplayer.PlaybackInfo{
			ID:                id,
			StreamType:        s.Type(),
			MimeType:          s.LoadContentType(),
			StreamUrl:         "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id,
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
		}

		// If the content type is an EBML content type, we can create a metadata parser
		if isEbmlContent(s.LoadContentType()) {
			parser := mkvparser.NewMetadataParser(s.file.NewReader(), s.logger)
			metadata := parser.GetMetadata(context.Background())
			if metadata.Error != nil {
				s.logger.Error().Err(metadata.Error).Msg("directstream(torrent): Failed to get metadata")
				s.manager.preStreamError(s, fmt.Errorf("failed to get metadata: %w", metadata.Error))
				s.playbackInfoErr = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				return
			}

			playbackInfo.MkvMetadata = metadata
			playbackInfo.MkvMetadataParser = mo.Some(parser)
		}

		s.playbackInfo = &playbackInfo
	})

	return s.playbackInfo, s.playbackInfoErr
}

func (s *TorrentStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

func (s *TorrentStream) ServeSubtitles(start int64) {
	// Cancel streams that are no longer relevant to prevent resource exhaustion
	s.cancelPreviousSubtitleStreams(start)

	container, err := s.LoadPlaybackInfo()
	if err != nil {
		s.logger.Error().Err(err).Msg("directstream(torrent): Failed to load playback info")
		return
	}

	// Check that it has mkv metadata
	parser, ok := container.MkvMetadataParser.Get()
	if !ok {
		// No mkv metadata, nothing to stream
		return
	}

	// Create a new context
	ctx, cancel := context.WithCancel(s.manager.playbackCtx)

	// Register this subtitle stream
	streamId := s.addSubtitleStream(start, cancel)

	s.manager.streamSubtitles(ctx, s, parser, s.file.NewReader(), start, func() {
		s.removeSubtitleStream(streamId)
		cancel()
	})
}

func (s *TorrentStream) ServeSubtitlesFromTime(timeSeconds float64) {
	// Cancel all previous subtitle streams since this is a new video event
	s.cancelAllSubtitleStreams()

	// Convert time to approximate byte offset
	var byteOffset int64 = 0

	if timeSeconds > 0 {
		// For time-based seeking, estimate position based on file size and duration
		playbackInfo, err := s.LoadPlaybackInfo()
		if err == nil && playbackInfo.MkvMetadata != nil && playbackInfo.MkvMetadata.Duration > 0 {
			// Estimate byte position based on time ratio
			timeRatio := timeSeconds / playbackInfo.MkvMetadata.Duration
			byteOffset = int64(float64(s.file.Length()) * timeRatio)

			s.logger.Debug().
				Float64("timeSeconds", timeSeconds).
				Float64("duration", playbackInfo.MkvMetadata.Duration).
				Int64("estimatedOffset", byteOffset).
				Int64("fileLength", s.file.Length()).
				Msg("directstream: Converting time to byte offset for subtitle extraction")
		}
	}

	s.ServeSubtitles(byteOffset)
}

func (s *TorrentStream) GetStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Trace().Str("range", r.Header.Get("Range")).Msg("directstream(torrent): Stream endpoint hit")

		if s.file == nil || s.torrent == nil {
			s.logger.Error().Msg("directstream(torrent): No torrent to stream")
			http.Error(w, "No torrent to stream", http.StatusNotFound)
			return
		}

		file := s.file
		s.logger.Trace().Str("file", file.DisplayPath()).Msg("directstream(torrent): New reader")
		tr := file.NewReader()
		defer func(tr torrent.Reader) {
			s.logger.Trace().Msg("directstream(torrent): Closing reader")
			_ = tr.Close()
		}(tr)

		tr.SetResponsive()
		// Read ahead 5MB for better streaming performance
		// DEVNOTE: Not sure if dynamic prioritization overwrites this but whatever
		tr.SetReadahead(5 * 1024 * 1024)

		// If this is a range request for a later part of the file, prioritize those pieces
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" && s.torrent != nil {
			// Attempt to prioritize the pieces requested in the range
			torrentutil.PrioritizeRangeRequestPieces(rangeHeader, s.torrent, file, s.logger)
		}

		s.logger.Trace().Str("file", file.DisplayPath()).Msg("directstream(torrent): Serving file content")
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeContent(
			w,
			r,
			file.DisplayPath(),
			time.Now(),
			tr,
		)
		s.logger.Trace().Msg("directstream(torrent): File content served")
	})
}

type PlayTorrentStreamOptions struct {
	ClientId      string
	EpisodeNumber int
	AnidbEpisode  string
	Media         *anilist.BaseAnime
	Torrent       *torrent.Torrent
	File          *torrent.File
}

// PlayTorrentStream is used by a module to load a new torrent stream.
func (m *Manager) PlayTorrentStream(opts PlayTorrentStreamOptions) error {
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

	stream := &TorrentStream{
		torrent: opts.Torrent,
		file:    opts.File,
		BaseStream: BaseStream{
			manager:            m,
			logger:             m.Logger,
			clientId:           opts.ClientId,
			media:              opts.Media,
			episode:            episode,
			episodeCollection:  episodeCollection,
			subtitleEventCache: result.NewResultMap[string, *mkvparser.SubtitleEvent](),
		},
	}

	m.loadStream(stream)

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// loadContentType loads the content type of the file.
// If the content type cannot be determined from the file extension,
// the first reader will be used to determine the content type.
func loadContentType(path string, reader ...io.ReadSeekCloser) string {
	ext := filepath.Ext(path)

	switch ext {
	case ".mp4":
		return "video/mp4"
	//case ".mkv":
	//	return "video/x-matroska"
	// Note: .mkv will be treated as a webm file for playback purposes
	case ".webm", ".mkv":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".flv":
		return "video/x-flv"
	default:
	}

	// No extension found
	// Read the first 1KB to determine the content type
	if len(reader) > 0 {
		if mimeType, ok := mkvparser.ReadIsMkvOrWebm(reader[0]); ok {
			return mimeType
		}
	}

	return ""
}

// streamSubtitles starts the subtitle stream.
// It will stream the subtitles from all tracks to the client. The client should load the subtitles in an array.
func (m *Manager) streamSubtitles(ctx context.Context, stream Stream, parser *mkvparser.MetadataParser, newReader io.ReadSeekCloser, offset int64, cleanupFunc func()) {
	m.Logger.Debug().Int64("offset", offset).Str("clientId", stream.ClientId()).Msg("directstream: Starting subtitle extraction")

	subtitleCh, errCh := parser.ExtractSubtitles(ctx, newReader, offset)

	go func() {
		defer func(reader io.ReadSeekCloser) {
			_ = reader.Close()
			m.Logger.Trace().Int64("offset", offset).Msg("directstream: Closing subtitle stream goroutine")
		}(newReader)
		defer func() {
			if cleanupFunc != nil {
				cleanupFunc()
			}
		}()

		// Keep track if channels are active to manage loop termination
		subtitleChannelActive := true
		errorChannelActive := true

		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
			select {
			case <-ctx.Done():
				m.Logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle streaming cancelled by context")
				return

			case subtitle, ok := <-subtitleCh:
				if !ok {
					subtitleCh = nil // Mark as exhausted
					subtitleChannelActive = false
					if !errorChannelActive { // If both channels are exhausted, exit
						return
					}
					continue // Continue to wait for errorChannel or ctx.Done()
				}
				if subtitle != nil {
					m.nativePlayer.SubtitleEvent(stream.ClientId(), subtitle)
				}

			case err, ok := <-errCh:
				if !ok {
					errCh = nil // Mark as exhausted
					errorChannelActive = false
					if !subtitleChannelActive { // If both channels are exhausted, exit
						return
					}
					continue // Continue to wait for subtitleChannel or ctx.Done()
				}
				// A value (error or nil) was received from errCh.
				// This is the terminal signal from the mkvparser's subtitle streaming process.
				if err != nil {
					m.Logger.Error().Err(err).Int64("offset", offset).Msg("directstream: Error streaming subtitles")
					// Send a toast notification only if there's an actual error.
					m.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Error streaming subtitles: %s", err.Error()))
				} else {
					m.Logger.Info().Int64("offset", offset).Msg("directstream: Subtitle streaming completed by parser.")
				}
				return // Terminate goroutine
			}
		}
	}()
}

func (m *Manager) preStreamError(stream Stream, err error) {
	m.wsEventManager.SendEvent(events.ErrorToast, err.Error())
	stream.Terminate()
	m.unloadStream()
}
