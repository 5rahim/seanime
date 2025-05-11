package directstream

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mediastream/mkvparser"
	"seanime/internal/mediastream/nativeplayer"
	"seanime/internal/util"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/samber/mo"
)

// Stream is the common interface for all stream types.
type Stream interface {
	// Type returns the type of the stream.
	Type() nativeplayer.StreamType
	// ClientId returns the client ID of the current stream.
	ClientId() string
	// Media returns the media of the current stream.
	Media() *anilist.BaseAnime
	// Episode returns the episode of the current stream.
	Episode() *anime.Episode
	// EpisodeCollection returns the episode collection for the media of the current stream.
	EpisodeCollection() *anime.EpisodeCollection
	// PlaybackInfo loads and returns the playback info if it already exists.
	PlaybackInfo() *nativeplayer.PlaybackInfo
	// GetAttachmentByName returns the attachment by name for the stream.
	// It is used to serve fonts and other attachments.
	GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool)
}

func (m *Manager) loadStream(stream Stream) {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	// Cancel the previous playback
	if m.playbackCancelFunc != nil {
		m.playbackCancelFunc()
	}

	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	m.playbackCancelFunc = cancel

	_ = ctx

	m.streamLoop(ctx, stream)

	m.currentStream = mo.Some(stream)
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
				m.Logger.Debug().Msgf("directstream: Stream event: %s", event)
			}
		}
	}()
}
func (m *Manager) unloadStream() {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()
	m.currentStream = mo.None[Stream]()
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BaseStream struct {
	clientId          string
	episode           *anime.Episode
	media             *anilist.BaseAnime
	episodeCollection *anime.EpisodeCollection
	playbackInfo      *nativeplayer.PlaybackInfo
	playbackInfoOnce  sync.Once
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

func getAttachmentByName(stream Stream, filename string) (*mkvparser.AttachmentInfo, bool) {
	filename, _ = url.PathUnescape(filename)

	container := stream.PlaybackInfo()

	metadata, ok := container.OptionalMkvMetadata.Get()
	if !ok {
		return nil, false
	}

	attachment, ok := metadata.GetAttachmentByName(filename)
	if !ok {
		return nil, false
	}

	return attachment, true
}

func isEbmlExtension(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".mkv" || ext == ".m4v" || ext == ".mp4"
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local File
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*LocalFileStream)(nil)

// LocalFileStream is a stream that is a local file.
type LocalFileStream struct {
	BaseStream
	localFile             *anime.LocalFile
	localFileWrapperEntry *anime.LocalFileWrapperEntry
}

func (s *LocalFileStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeFile
}

func (s *LocalFileStream) PlaybackInfo() *nativeplayer.PlaybackInfo {
	s.playbackInfoOnce.Do(func() {

	})

	return s.playbackInfo
}

func (s *LocalFileStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s, filename)
}

type PlayLocalFileOptions struct {
	ClientId   string
	Path       string
	Media      *anilist.BaseAnime
	LocalFiles []*anime.LocalFile
}

// PlayLocalFile is used by a module to load a new torrent stream.
func (m *Manager) PlayLocalFile(opts PlayLocalFileOptions) error {
	animeCollection, ok := m.animeCollection.Get()
	if !ok {
		return fmt.Errorf("cannot play local file, anime collection is not set")
	}

	episodeCollection, err := anime.NewEpisodeCollectionFromLocalFiles(anime.NewEpisodeCollectionFromLocalFilesOptions{
		LocalFiles:       opts.LocalFiles,
		Media:            opts.Media,
		AnimeCollection:  animeCollection,
		Platform:         m.platform,
		MetadataProvider: m.metadataProvider,
		Logger:           m.Logger,
	})
	if err != nil {
		return fmt.Errorf("cannot play local file, could not create episode collection: %w", err)
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
			clientId:          opts.ClientId,
			media:             opts.Media,
			episode:           episode,
			episodeCollection: episodeCollection,
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

func (s *TorrentStream) PlaybackInfo() *nativeplayer.PlaybackInfo {
	s.playbackInfoOnce.Do(func() {
		if s.file == nil || s.torrent == nil {
			s.playbackInfo = &nativeplayer.PlaybackInfo{}
			return
		}

		playbackInfo := nativeplayer.PlaybackInfo{
			StreamType:          s.Type(),
			StreamUrl:           "{{SERVER_URL}}/api/v1/directstream",
			MkvMetadata:         nil,
			OptionalMkvMetadata: mo.None[*mkvparser.Metadata](),
		}

		if isEbmlExtension(s.file.DisplayPath()) {
			playbackInfo.MkvMetadata = &mkvparser.Metadata{}
		}

		s.playbackInfo = &playbackInfo
	})

	return s.playbackInfo
}

func (s *TorrentStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s, filename)
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
			clientId:          opts.ClientId,
			media:             opts.Media,
			episode:           episode,
			episodeCollection: episodeCollection,
		},
	}

	m.loadStream(stream)

	return nil
}
