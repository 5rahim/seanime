package directstream

import (
	"context"
	"fmt"
	"net/http"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util/result"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*Nakama)(nil)

// Nakama is a stream that is a torrent.
type Nakama struct {
	httpBaseStream
	torrent       *hibiketorrent.AnimeTorrent
	streamReadyCh chan struct{} // Closed by the initiator when the stream is ready
}

func (s *Nakama) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeNakama
}

func (s *Nakama) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return s.httpBaseStream.loadPlaybackInfo(s.Type())
}

func (s *Nakama) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

func (s *Nakama) GetStreamHandler() http.Handler {
	return s.httpBaseStream.getStreamHandler(s)
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
func (m *Manager) PlayNakamaStream(ctx context.Context, opts PlayNakamaStreamOptions) (err error) {
	if !m.BeginOpen(opts.ClientId, "Loading stream...", nil) {
		return fmt.Errorf("stream opening was cancelled")
	}
	defer func() {
		if err != nil {
			m.AbortOpen(opts.ClientId, err)
		}
	}()

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
		httpBaseStream: httpBaseStream{
			streamUrl: opts.StreamUrl,
			requestHeaders: http.Header{
				"X-Seanime-Nakama-Token": []string{opts.NakamaHostPassword},
			},
			headResponseHeaders: http.Header{
				"X-Seanime-Nakama-Token": []string{opts.NakamaHostPassword},
			},
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
