package directstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util"
	httputil "seanime/internal/util/http"
	"seanime/internal/util/result"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

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
			ContentLength:     size,
			MkvMetadata:       nil,
			MkvMetadataParser: mo.None[*mkvparser.MetadataParser](),
			Episode:           s.episode,
			Media:             s.media,
			EntryListData:     entryListData,
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

func ServeLocalFile(w http.ResponseWriter, r *http.Request, lfStream *LocalFileStream) {
	if lfStream.serveContentCancelFunc != nil {
		lfStream.serveContentCancelFunc()
	}

	ct, cancel := context.WithCancel(lfStream.manager.playbackCtx)
	lfStream.serveContentCancelFunc = cancel

	reader, err := lfStream.newReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	playbackInfo, err := lfStream.LoadPlaybackInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	size := playbackInfo.ContentLength
	w.Header().Set("Content-Length", fmt.Sprint(size))

	// No Range header → let Go handle it
	rangeHdr := r.Header.Get("Range")
	if rangeHdr == "" {
		http.ServeContent(w, r, lfStream.localFile.Path, time.Now(), reader)
		return
	}

	// Parse the range header
	ranges, err := httputil.ParseRange(rangeHdr, size)
	if err != nil && !errors.Is(err, httputil.ErrNoOverlap) {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
		return
	} else if err != nil && errors.Is(err, httputil.ErrNoOverlap) {
		// Let Go handle overlap
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.ServeContent(w, r, lfStream.localFile.Path, time.Now(), reader)
		return
	}

	if _, ok := playbackInfo.MkvMetadataParser.Get(); ok {
		// Start a subtitle stream from the current position
		subReader, err := lfStream.newReader()
		if err != nil {
			lfStream.logger.Error().Err(err).Msg("directstream: Failed to create subtitle reader")
			http.Error(w, "Failed to create subtitle reader", http.StatusInternalServerError)
			return
		}
		lfStream.StartSubtitleStream(lfStream, lfStream.manager.playbackCtx, subReader, ranges[0].Start)
	}

	serveContentRange(w, r, ct, reader, lfStream.localFile.Path, size, playbackInfo.MimeType, ranges)
}

type PlayLocalFileOptions struct {
	ClientId   string
	Path       string
	LocalFiles []*anime.LocalFile
}

// PlayLocalFile is used by a module to load a new torrent stream.
func (m *Manager) PlayLocalFile(ctx context.Context, opts PlayLocalFileOptions) error {
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

	episodeCollection, err := anime.NewEpisodeCollectionFromLocalFiles(ctx, anime.NewEpisodeCollectionFromLocalFilesOptions{
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
			manager:               m,
			logger:                m.Logger,
			clientId:              opts.ClientId,
			filename:              filepath.Base(lf.Path),
			media:                 media,
			episode:               episode,
			episodeCollection:     episodeCollection,
			subtitleEventCache:    result.NewResultMap[string, *mkvparser.SubtitleEvent](),
			activeSubtitleStreams: result.NewResultMap[string, *SubtitleStream](),
		},
	}

	m.loadStream(stream)

	return nil
}
