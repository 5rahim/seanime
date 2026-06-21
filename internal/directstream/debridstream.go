package directstream

import (
	"context"
	"fmt"
	"net/http"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/util/result"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Debrid
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*DebridStream)(nil)

// DebridStream is a stream sourced from a debrid provider.
type DebridStream struct {
	httpBaseStream
	torrent       *hibiketorrent.AnimeTorrent
	streamReadyCh chan struct{} // Closed by the initiator when the stream URL is resolved
}

func (s *DebridStream) Type() mediacore.PlaybackType {
	return mediacore.PlaybackTypeDebrid
}

func (s *DebridStream) LoadPlaybackInfo() (*mediacore.PlaybackInfo, error) {
	return s.httpBaseStream.loadPlaybackInfo(s.Type())
}

func (s *DebridStream) GetStreamHandler() http.Handler {
	return s.httpBaseStream.getStreamHandler(s)
}

func (s *DebridStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

type PlayDebridStreamOptions struct {
	StreamUrl    string
	MediaId      int
	AnidbEpisode string // Anizip episode
	Media        *anilist.BaseAnime
	Torrent      *hibiketorrent.AnimeTorrent // Selected torrent
	FileId       string                      // File ID or index
	UserAgent    string
	ClientId     string
	AutoSelect   bool
}

// PlayDebridStream is used by a module to load a new debrid stream.
func (m *Manager) PlayDebridStream(ctx context.Context, filepath string, opts PlayDebridStreamOptions) error {
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
		return fmt.Errorf("cannot play debrid stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &DebridStream{
		torrent: opts.Torrent,
		httpBaseStream: httpBaseStream{
			streamUrl: opts.StreamUrl,
			filepath:  filepath,
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
		},
		streamReadyCh: make(chan struct{}),
	}

	go func() {
		m.loadStream(stream)
	}()

	return nil
}
