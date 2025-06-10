package directstream

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	httputil "seanime/internal/util/http"
	"seanime/internal/util/result"
	"seanime/internal/util/torrentutil"
	"time"

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
	streamReadyCh chan struct{} // Closed by the initiator when the stream is ready
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

		var entryListData *anime.EntryListData
		if animeCollection, ok := s.manager.animeCollection.Get(); ok {
			if listEntry, ok := animeCollection.GetListEntryFromAnimeId(s.media.ID); ok {
				entryListData = anime.NewEntryListData(listEntry)
			}
		}

		playbackInfo := nativeplayer.PlaybackInfo{
			ID:                id,
			StreamType:        s.Type(),
			MimeType:          s.LoadContentType(),
			StreamUrl:         "{{SERVER_URL}}/api/v1/directstream/stream?id=" + id,
			ContentLength:     s.file.Length(),
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
			Episode:           s.episode,
			Media:             s.media,
			EntryListData:     entryListData,
		}

		// If the content type is an EBML content type, we can create a metadata parser
		if isEbmlContent(s.LoadContentType()) {
			reader := torrentutil.NewReadSeeker(s.torrent, s.file, s.logger)
			parser := mkvparser.NewMetadataParser(reader, s.logger)
			metadata := parser.GetMetadata(context.Background())
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
		s.logger.Trace().Str("range", r.Header.Get("Range")).Str("method", r.Method).Msg("directstream(torrent): Stream endpoint hit")

		if s.file == nil || s.torrent == nil {
			s.logger.Error().Msg("directstream(torrent): No torrent to stream")
			http.Error(w, "No torrent to stream", http.StatusNotFound)
			return
		}

		// Handle HEAD requests explicitly to provide file size information
		if r.Method == http.MethodHead {
			s.logger.Trace().Msg("directstream(torrent): Handling HEAD request")

			// Set the content length from torrent file
			fileSize := s.file.Length()
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
			w.Header().Set("Content-Type", s.LoadContentType())
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", s.file.DisplayPath()))
			w.WriteHeader(http.StatusOK)
			return
		}

		file := s.file
		s.logger.Trace().Str("file", file.DisplayPath()).Msg("directstream(torrent): New reader")
		tr := torrentutil.NewReadSeeker(s.torrent, file, s.logger)
		defer func() {
			s.logger.Trace().Msg("directstream(torrent): Closing reader")
			_ = tr.Close()
		}()

		// If this is a range request for a later part of the file, prioritize those pieces
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" && s.torrent != nil {
			// Attempt to prioritize the pieces requested in the range
			torrentutil.PrioritizeRangeRequestPieces(rangeHeader, s.torrent, file, s.logger)
		}

		// Parse the range header
		ranges, err := httputil.ParseRange(rangeHeader, file.Length())
		if err != nil && !errors.Is(err, httputil.ErrNoOverlap) {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", file.Length()))
			http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
			return
		} else if err != nil && errors.Is(err, httputil.ErrNoOverlap) {
			// Let Go handle overlap
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", file.Length()))
			http.ServeContent(w, r, file.DisplayPath(), time.Now(), tr)
			return
		}

		if _, ok := s.playbackInfo.MkvMetadataParser.Get(); ok {
			// Start a subtitle stream from the current position
			subReader := file.NewReader()
			subReader.SetResponsive()
			s.StartSubtitleStream(s, s.manager.playbackCtx, subReader, ranges[0].Start)
		}

		serveTorrent(w, r, s.manager.playbackCtx, tr, file.DisplayPath(), file.Length(), s.LoadContentType(), ranges)
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
func (m *Manager) PlayTorrentStream(ctx context.Context, opts PlayTorrentStreamOptions) (chan struct{}, error) {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	episodeCollection, err := anime.NewEpisodeCollection(anime.NewEpisodeCollectionOptions{
		AnimeMetadata:    nil,
		Media:            opts.Media,
		MetadataProvider: m.metadataProvider,
		Logger:           m.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
	}

	episode, ok := episodeCollection.FindEpisodeByAniDB(opts.AnidbEpisode)
	if !ok {
		return nil, fmt.Errorf("cannot play torrent stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &TorrentStream{
		torrent: opts.Torrent,
		file:    opts.File,
		BaseStream: BaseStream{
			manager:               m,
			logger:                m.Logger,
			clientId:              opts.ClientId,
			media:                 opts.Media,
			filename:              filepath.Base(opts.File.DisplayPath()),
			episode:               episode,
			episodeCollection:     episodeCollection,
			subtitleEventCache:    result.NewResultMap[string, *mkvparser.SubtitleEvent](),
			activeSubtitleStreams: result.NewResultMap[string, *SubtitleStream](),
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
