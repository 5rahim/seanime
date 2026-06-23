package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/player"
	"seanime/internal/util/result"
	"seanime/internal/util/torrentutil"

	"github.com/anacrolix/torrent"
	"github.com/google/uuid"
	"github.com/samber/mo"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*TorrentStream)(nil)

// TorrentStream is a stream that is a torrent.
type TorrentStream struct {
	BaseStream
	torrent       *torrent.Torrent
	file          *torrent.File
	downloadDir   string
	onTerminate   func()
	streamReadyCh chan struct{} // Closed by the initiator when the stream is ready
}

func (s *TorrentStream) Type() player.PlaybackType {
	return player.PlaybackTypeTorrent
}

func (s *TorrentStream) completedFilePath() (string, bool) {
	if s.downloadDir == "" || s.torrent == nil || s.file == nil || s.file.Length() <= 0 {
		return "", false
	}

	filePath := filepath.Join(s.downloadDir, s.torrent.InfoHash().HexString(), filepath.FromSlash(s.file.Path()))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() || info.Size() != s.file.Length() {
		return "", false
	}

	return filePath, true
}

func (s *TorrentStream) openCompletedFile() (io.ReadSeekCloser, bool) {
	filePath, ok := s.completedFilePath()
	if !ok {
		return nil, false
	}

	reader, err := os.Open(filePath)
	if err != nil {
		s.logger.Warn().Err(err).Str("path", filePath).Msg("directstream(torrent): Failed to open completed torrent file")
		return nil, false
	}

	s.logger.Trace().Str("path", filePath).Msg("directstream(torrent): Using completed torrent file")
	return reader, true
}

func (s *TorrentStream) hasCompletedFile() bool {
	_, ok := s.completedFilePath()
	return ok
}

func (s *TorrentStream) newReader() io.ReadSeekCloser {
	if reader, ok := s.openCompletedFile(); ok {
		return reader
	}

	return torrentutil.NewReadSeeker(s.torrent, s.file, s.logger)
}

func (s *TorrentStream) newMetadataReader() io.ReadSeekCloser {
	if reader, ok := s.openCompletedFile(); ok {
		return reader
	}

	reader := s.file.NewReader()
	reader.SetResponsive()
	reader.SetReadahead(0)
	return reader
}

func (s *TorrentStream) newSubtitleReader() io.ReadSeekCloser {
	if reader, ok := s.openCompletedFile(); ok {
		return reader
	}

	return torrentutil.NewReadSeeker(s.torrent, s.file, s.logger)
}

func (s *TorrentStream) LoadContentType() string {
	s.contentTypeOnce.Do(func() {
		if !s.shouldProcessMediaOnServer() {
			s.contentType = loadContentType(s.file.DisplayPath())
			if s.contentType == "" {
				s.contentType = "application/octet-stream"
			}
			return
		}
		r := s.newMetadataReader()
		defer r.Close()
		s.contentType = loadContentType(s.file.DisplayPath(), r)
	})

	return s.contentType
}

func (s *TorrentStream) LoadPlaybackInfo() (ret *player.PlaybackInfo, err error) {
	s.playbackInfoOnce.Do(func() {
		if s.file == nil || s.torrent == nil {
			ret = &player.PlaybackInfo{}
			err = fmt.Errorf("torrent is not set")
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

		streamURL := "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id + s.manager.GetHMACTokenQueryParam("/api/v1/directstream/stream", "&")
		playbackInfo := player.PlaybackInfo{
			ID:                id,
			PlaybackType:      s.Type(),
			PlaybackURI:       streamURL,
			StreamPath:        s.file.Path(),
			MimeType:          s.LoadContentType(),
			StreamURL:         streamURL,
			ContentLength:     s.file.Length(),
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
			Episode:           s.episode,
			Media:             s.media,
			EntryListData:     entryListData,
		}

		// VideoCore needs server-side MKV metadata and subtitle extraction.
		// MpvCore reads the proxied torrent bytes and lets libmpv demux them.
		if s.shouldProcessMediaOnServer() && isEbmlContent(s.LoadContentType()) {
			reader := s.newMetadataReader()
			defer reader.Close()
			parser := mkvparser.NewMetadataParser(reader, s.logger)
			metadataCtx := s.manager.playbackCtx
			if metadataCtx == nil {
				metadataCtx = context.Background()
			}
			metadata := parser.GetMetadata(metadataCtx)
			if metadata.Error != nil {
				err = fmt.Errorf("failed to get metadata: %w", metadata.Error)
				s.logger.Error().Err(metadata.Error).Msg("directstream(torrent): Failed to get metadata")
				s.playbackInfoErr = err
				return
			}

			// Add subtitle tracks from subtitle files in the torrent
			s.AppendSubtitleFile(s.torrent, s.file, metadata)

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

func (s *TorrentStream) GetStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//s.logger.Trace().Str("range", r.Header.Get("Range")).Str("method", r.Method).Msg("directstream(torrent): Stream endpoint hit")

		if s.file == nil || s.torrent == nil {
			s.logger.Error().Msg("directstream(torrent): No torrent to stream")
			http.Error(w, "No torrent to stream", http.StatusNotFound)
			return
		}

		size := s.file.Length()
		contentType := s.LoadContentType()
		name := s.file.DisplayPath()

		// Handle HEAD requests explicitly to provide file size information
		if r.Method == http.MethodHead {
			s.logger.Trace().Msg("directstream(torrent): Handling HEAD request")
			// Set the content length from torrent file
			w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", name))
			w.WriteHeader(http.StatusOK)
			return
		}

		if isThumbnailRequest(r) {
			reader := s.newReader()
			defer reader.Close()
			ra, ok := handleRange(w, r, reader, name, size)
			if !ok {
				return
			}
			serveContentRange(w, r, r.Context(), reader, name, size, contentType, ra)
			return
		}

		s.logger.Trace().Str("file", name).Msg("directstream(torrent): New reader")
		tr := s.newReader()
		defer func() {
			s.logger.Trace().Msg("directstream(torrent): Closing reader")
			_ = tr.Close()
		}()

		playbackCtx := s.manager.playbackCtx
		if playbackCtx == nil {
			playbackCtx = r.Context()
		}
		serveCtx, cancelServe := context.WithCancel(playbackCtx)
		stopRequestCancel := context.AfterFunc(r.Context(), cancelServe)
		defer func() {
			stopRequestCancel()
			cancelServe()
		}()

		ra, ok := handleRange(w, r, tr, name, size)
		if !ok {
			return
		}

		serveContentRange(w, r, serveCtx, tr, name, size, s.LoadContentType(), ra)
	})
}

// Terminate overrides BaseStream.Terminate to also terminate the torrent stream.
func (s *TorrentStream) Terminate() {
	s.onTerminate()

	// Call the base implementation
	s.BaseStream.Terminate()
}

type PlayTorrentStreamOptions struct {
	ClientId      string
	EpisodeNumber int
	AnidbEpisode  string
	Media         *anilist.BaseAnime
	Torrent       *torrent.Torrent
	File          *torrent.File
	DownloadDir   string
	OnTerminate   func()
}

// PlayTorrentStream is used by a module to load a new torrent stream.
func (m *Manager) PlayTorrentStream(ctx context.Context, opts PlayTorrentStreamOptions) (chan struct{}, error) {
	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:       nil,
		Media:               opts.Media,
		MetadataProviderRef: m.metadataProviderRef,
		Logger:              m.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
	}

	episode, ok := episodeCollection.FindEpisodeByAniDB(opts.AnidbEpisode)
	if !ok {
		return nil, fmt.Errorf("cannot play torrent stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &TorrentStream{
		torrent:     opts.Torrent,
		file:        opts.File,
		downloadDir: opts.DownloadDir,
		onTerminate: opts.OnTerminate,
		BaseStream: BaseStream{
			manager:               m,
			logger:                m.Logger,
			clientId:              opts.ClientId,
			media:                 opts.Media,
			filename:              filepath.Base(opts.File.DisplayPath()),
			episode:               episode,
			episodeCollection:     episodeCollection,
			subtitleEventCache:    result.NewMap[string, *mkvparser.SubtitleEvent](),
			activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
		},
		streamReadyCh: make(chan struct{}),
	}

	go func() {
		<-stream.streamReadyCh
		m.loadStream(stream)
	}()

	return stream.streamReadyCh, nil
}

// AppendSubtitleFile finds the subtitle file for the torrent and appends it as a track to the metadata
//   - If there's only one subtitle file, use it
//   - If there are multiple subtitle files, use the one that matches the name of the selected torrent file
//   - If there are no subtitle files, do nothing
//
// If the subtitle file is not ASS/SSA, it will be converted to ASS/SSA.
func (s *TorrentStream) AppendSubtitleFile(t *torrent.Torrent, file *torrent.File, metadata *mkvparser.Metadata) {

}
