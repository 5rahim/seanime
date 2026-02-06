package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	httputil "seanime/internal/util/http"
	"seanime/internal/util/result"
	"sync"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*Nakama)(nil)

// Nakama is a stream that is a torrent.
type Nakama struct {
	BaseStream
	streamUrl          string
	contentLength      int64
	torrent            *hibiketorrent.AnimeTorrent
	streamReadyCh      chan struct{}        // Closed by the initiator when the stream is ready
	httpStream         *httputil.FileStream // Shared file-backed cache for multiple readers
	cacheMu            sync.RWMutex         // Protects httpStream access
	nakamaHostPassword string
}

func (s *Nakama) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeNakama
}

func (s *Nakama) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		s.cacheMu.RLock()
		if s.httpStream == nil {
			s.cacheMu.RUnlock()
			_ = s.initializeStream()
		} else {
			s.cacheMu.RUnlock()
		}

		info, ok := s.manager.FetchStreamInfo(s.streamUrl)
		if !ok {
			s.logger.Warn().Str("url", s.streamUrl).Msg("directstream(nakama): Failed to fetch stream info for content type")
			return
		}
		s.logger.Debug().Str("url", s.streamUrl).Str("contentType", info.ContentType).Int64("contentLength", info.ContentLength).Msg("directstream(nakama): Fetched content type and length")
		s.contentType = info.ContentType
		if s.contentType == "application/force-download" {
			s.contentType = "application/octet-stream"
		}
		s.contentLength = info.ContentLength
	})

	return s.contentType
}

// Close cleanup the HTTP cache and other resources
func (s *Nakama) Close() error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.logger.Debug().Msg("directstream(nakama): Closing HTTP cache")

	if s.httpStream != nil {
		if err := s.httpStream.Close(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(nakama): Failed to close HTTP cache")
			return err
		}
		s.httpStream = nil
	}

	s.logger.Debug().Msg("directstream(nakama): HTTP cache closed successfully")

	return nil
}

// Terminate overrides BaseStream.Terminate to also clean up the HTTP cache
func (s *Nakama) Terminate() {
	// Clean up HTTP cache first
	if err := s.Close(); err != nil {
		s.logger.Error().Err(err).Msg("directstream(nakama): Failed to clean up HTTP cache during termination")
	}

	// Call the base implementation
	s.BaseStream.Terminate()
}

func (s *Nakama) LoadPlaybackInfo() (ret *nativeplayer.PlaybackInfo, err error) {
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
			StreamPath:        "",
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
				s.logger.Error().Err(err).Msg("directstream(nakama): Failed to create reader for stream url")
				s.playbackInfoErr = err
				return
			}
			defer reader.Close() // Close this specific reader instance

			_, _ = reader.Seek(0, io.SeekStart)
			s.logger.Trace().Msgf(
				"directstream(nakama): Loading metadata for stream url: %s",
				s.streamUrl,
			)

			parser := mkvparser.NewMetadataParser(reader, s.logger)
			metadata := parser.GetMetadata(context.Background())
			if metadata.Error != nil {
				err = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				s.logger.Error().Err(metadata.Error).Msg("directstream(nakama): Failed to get metadata")
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

func (s *Nakama) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

func (s *Nakama) GetStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//s.logger.Trace().Str("range", r.Header.Get("Range")).Str("method", r.Method).Msg("directstream(nakama): Stream endpoint hit")

		if s.streamUrl == "" {
			s.logger.Error().Msg("directstream(nakama): No URL to stream")
			http.Error(w, "No URL to stream", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodHead {
			s.logger.Trace().Msg("directstream(nakama): Handling HEAD request")

			fileSize := s.contentLength
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
			w.Header().Set("Content-Type", s.LoadContentType())
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("X-Seanime-Nakama-Token", s.nakamaHostPassword)
			w.WriteHeader(http.StatusOK)
			return
		}

		rangeHeader := r.Header.Get("Range")

		if err := s.initializeStream(); err != nil {
			s.logger.Error().Err(err).Msg("directstream(nakama): Failed to initialize FileStream")
			http.Error(w, "Failed to initialize FileStream", http.StatusInternalServerError)
			return
		}

		reader, err := s.getReader()
		if err != nil {
			s.logger.Error().Err(err).Msg("directstream(nakama): Failed to create reader for stream url")
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
				s.logger.Error().Err(err).Msg("directstream(nakama): Failed to create subtitle reader for stream url")
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

		req.Header.Set("X-Seanime-Nakama-Token", s.nakamaHostPassword)

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

type PlayNakamaStreamOptions struct {
	StreamUrl          string
	MediaId            int
	AnidbEpisode       string // Animap episode
	Media              *anilist.BaseAnime
	NakamaHostPassword string
	ClientId           string
}

// PlayNakamaStream is used by a module to load a new nakama stream.
func (m *Manager) PlayNakamaStream(ctx context.Context, opts PlayNakamaStreamOptions) error {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:       nil,
		Media:               opts.Media,
		MetadataProviderRef: m.metadataProviderRef,
		Logger:              m.Logger,
	})
	if err != nil {
		return fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
	}

	episode, ok := episodeCollection.FindEpisodeByAniDB(opts.AnidbEpisode)
	if !ok {
		return fmt.Errorf("cannot play nakama stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &Nakama{
		streamUrl:          opts.StreamUrl,
		nakamaHostPassword: opts.NakamaHostPassword,
		BaseStream: BaseStream{
			manager:               m,
			logger:                m.Logger,
			clientId:              opts.ClientId,
			media:                 opts.Media,
			filename:              "",
			episode:               episode,
			episodeCollection:     episodeCollection,
			subtitleEventCache:    result.NewMap[string, *mkvparser.SubtitleEvent](),
			activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
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
func (s *Nakama) initializeStream() error {
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
		info, ok := s.manager.FetchStreamInfo(s.streamUrl)
		if !ok {
			return fmt.Errorf("failed to fetch stream info")
		}
		s.contentLength = info.ContentLength
	}

	s.logger.Debug().Msgf("directstream(nakama): Initializing FileStream for stream URL: %s", s.streamUrl)

	// Create a file-backed stream with the known content length
	cache, err := httputil.NewFileStream(s.manager.playbackCtx, s.logger, s.contentLength)
	if err != nil {
		return fmt.Errorf("failed to create FileStream: %w", err)
	}

	s.httpStream = cache

	s.logger.Debug().Msgf("directstream(nakama): FileStream initialized")

	return nil
}

func (s *Nakama) getReader() (io.ReadSeekCloser, error) {
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
