package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	// StreamError is called when an error occurs while streaming.
	// This is used to notify the native player that an error occurred.
	// It will close the stream.
	StreamError(err error)
	// Terminate ends the stream.
	// Once this is called, the stream should not be used anymore.
	Terminate()
	//
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

	// Start the subtitle stream
	stream.ServeSubtitles(0)

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

				fmt.Printf("event, %T\n", event)

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
					// Make sure this event is for the current stream
					if cs, ok := m.currentStream.Get(); ok && event.GetClientId() == cs.ClientId() {
						// Request subtitles from the new time
						// cs.RequestSubtitlesFrom(event.CurrentTime)
					}
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

type subtitleHead struct {
	doNotExceed int64 // in bytes
}

type BaseStream struct {
	logger                   *zerolog.Logger
	clientId                 string
	contentType              string
	contentTypeOnce          sync.Once
	episode                  *anime.Episode
	media                    *anilist.BaseAnime
	episodeCollection        *anime.EpisodeCollection
	playbackInfo             *nativeplayer.PlaybackInfo
	playbackInfoErr          error
	playbackInfoOnce         sync.Once
	subtitleStreamCancelFunc context.CancelFunc
	subtitleEventCache       *result.Map[string, *mkvparser.SubtitleEvent]
	subtitleHeads            *result.Map[int64, *subtitleHead]
	terminateOnce            sync.Once
	serveContentCancelFunc   context.CancelFunc

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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local File
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*LocalFileStream)(nil)

// LocalFileStream is a stream that is a local file.
type LocalFileStream struct {
	BaseStream
	localFile *anime.LocalFile
}

func (s *LocalFileStream) newReader() (io.ReadSeeker, error) {
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
	// Cancel the previous subtitle stream
	// if s.subtitleStreamCancelFunc != nil {
	// 	s.subtitleStreamCancelFunc()
	// }

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

	// s.subtitleStreamCancelFunc = cancel

	// Open the file
	fr, err := s.newReader()
	if err != nil {
		s.logger.Error().Err(err).Msg("directstream(file): Failed to open local file")
		return
	}

	subReader, ok := fr.(io.ReadSeeker)
	if !ok {
		s.logger.Error().Msg("directstream(file): File is not seekable")
		return
	}

	// Offset the reader
	n, err := subReader.Seek(start, io.SeekStart)
	if err != nil {
		s.logger.Error().Err(err).Msg("directstream(file): Failed to seek to start")
		return
	}

	s.logger.Debug().Msgf("directstream(file): Seeked %d bytes to start", n)

	if start > 0 {
		s.subtitleHeads.Range(func(key int64, value *subtitleHead) bool {
			// if the doNotExceed value of this head is less than the start, set it to the start
			// this will interrupt the subtitle stream to avoid overlap
			if value.doNotExceed > start {
				value.doNotExceed = start
			}
			return true
		})
	}

	s.subtitleHeads.Set(start, &subtitleHead{
		doNotExceed: s.playbackInfo.ContentLength,
	})
	defer s.subtitleHeads.Delete(start)

	go func(ctx context.Context, cancel context.CancelFunc) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Get reader position
				pos, err := subReader.Seek(0, io.SeekCurrent)
				if err != nil {
					s.logger.Error().Err(err).Msg("directstream(file): Failed to get reader position")
					return
				}

				// Cancel the subtitle stream if the reader position is greater than the doNotExceed value of any head
				head, ok := s.subtitleHeads.Get(start)
				if !ok {
					continue
				}

				if pos >= head.doNotExceed {
					s.logger.Debug().Int64("pos", pos).Int64("doNotExceed", head.doNotExceed).Msg("directstream(file): Reader position is greater than doNotExceed value, cancelling subtitle stream")
					if cancel != nil {
						cancel()
					}
					return
				}
			}
		}

		// If the reader position is greater than the content length, set the doNotExceed value to the content length
	}(ctx, cancel)

	s.manager.streamSubtitles(ctx, s, parser, subReader, cancel)
}

// func (s *LocalFileStream) RequestSubtitlesFrom(currentTime float64) {
// 	s.logger.Debug().Float64("currentTime", currentTime).Msg("directstream(file): Requesting subtitles from")
// 	// Cancel the previous subtitle stream
// 	if s.subtitleStreamCancelFunc != nil {
// 		s.subtitleStreamCancelFunc()
// 	}

// 	container, err := s.LoadPlaybackInfo()
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("directstream(file): Failed to load playback info for subtitle request")
// 		return
// 	}

// 	parser, ok := container.MkvMetadataParser.Get()
// 	if !ok {
// 		s.logger.Warn().Msg("directstream(file): No mkv metadata parser available for subtitle request")
// 		return
// 	}

// 	ctx, cancel := context.WithCancel(s.manager.playbackCtx)
// 	s.subtitleStreamCancelFunc = cancel

// 	fr, err := s.newReader()
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("directstream(file): Failed to open local file for subtitle request")
// 		return
// 	}

// 	reader, ok := fr.(io.ReadSeeker)
// 	if !ok {
// 		s.logger.Error().Msg("directstream(file): File is not seekable for subtitle request")
// 		if closer, ok := fr.(io.Closer); ok {
// 			_ = closer.Close()
// 		}
// 		return
// 	}

// 	s.manager.streamSubtitles(ctx, s, parser, reader, currentTime, false)
// }

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
			subtitleHeads:      result.NewResultMap[int64, *subtitleHead](),
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
	// Cancel the previous subtitle stream
	if s.subtitleStreamCancelFunc != nil {
		s.subtitleStreamCancelFunc()
	}

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
	s.subtitleStreamCancelFunc = cancel

	s.manager.streamSubtitles(ctx, s, parser, s.file.NewReader(), nil)
}

// func (s *TorrentStream) RequestSubtitlesFrom(currentTime float64) {
// 	s.logger.Debug().Float64("currentTime", currentTime).Msg("directstream(torrent): Requesting subtitles from")
// 	// Cancel the previous subtitle stream
// 	if s.subtitleStreamCancelFunc != nil {
// 		s.subtitleStreamCancelFunc()
// 	}

// 	container, err := s.LoadPlaybackInfo()
// 	if err != nil {
// 		s.logger.Error().Err(err).Msg("directstream(torrent): Failed to load playback info for subtitle request")
// 		return
// 	}

// 	parser, ok := container.MkvMetadataParser.Get()
// 	if !ok {
// 		s.logger.Warn().Msg("directstream(torrent): No mkv metadata parser available for subtitle request")
// 		return
// 	}

// 	ctx, cancel := context.WithCancel(s.manager.playbackCtx)
// 	s.subtitleStreamCancelFunc = cancel

// 	s.manager.streamSubtitles(ctx, s, parser, s.file.NewReader(), currentTime, false)
// }

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
			subtitleHeads:      result.NewResultMap[int64, *subtitleHead](),
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
func loadContentType(path string, reader ...io.ReadSeeker) string {
	ext := filepath.Ext(path)

	switch ext {
	case ".mp4":
		return "video/mp4"
	//case ".mkv":
	//	return "video/x-matroska"
	// Note: .mkv will be treated as a webm file
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
//
// It should be given a pre-offset reader each time. The reader will be closed when the stream is done if it implements io.Closer.
func (m *Manager) streamSubtitles(ctx context.Context, stream Stream, parser *mkvparser.MetadataParser, offsetReader io.ReadSeeker, cancelFunc context.CancelFunc) {
	subtitleCh, errCh := parser.ExtractSubtitles(ctx, offsetReader)

	go func() {
		defer func(reader io.ReadSeeker) {
			// Close the reader if it implements io.Closer
			if closer, ok := reader.(io.Closer); ok {
				_ = closer.Close()
			}
			m.Logger.Trace().Msg("directstream(file): Closing subtitle stream goroutine")
		}(offsetReader)
		defer func() {
			if cancelFunc != nil {
				cancelFunc()
			}
		}()

		// Keep track if channels are active to manage loop termination
		subtitleChannelActive := true
		errorChannelActive := true

		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
			select {
			case <-ctx.Done():
				m.Logger.Info().Msg("directstream: Subtitle streaming cancelled by context")
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
					m.Logger.Error().Err(err).Msg("directstream: Error streaming subtitles")
					// Send a toast notification only if there's an actual error.
					m.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Error streaming subtitles: %s", err.Error()))
				} else {
					m.Logger.Info().Msg("directstream: Subtitle streaming completed by parser.")
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

func getAttachmentByName(ctx context.Context, stream Stream, filename string) (*mkvparser.AttachmentInfo, bool) {
	filename, _ = url.PathUnescape(filename)

	container, err := stream.LoadPlaybackInfo()
	if err != nil {
		return nil, false
	}

	parser, ok := container.MkvMetadataParser.Get()
	if !ok {
		return nil, false
	}

	attachment, ok := parser.GetMetadata(ctx).GetAttachmentByName(filename)
	if !ok {
		return nil, false
	}

	return attachment, true
}

func isEbmlExtension(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".mkv" || ext == ".m4v" || ext == ".mp4"
}

func isEbmlContent(mimeType string) bool {
	return mimeType == "video/x-matroska" || mimeType == "video/webm"
}
