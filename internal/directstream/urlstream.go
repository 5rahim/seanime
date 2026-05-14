package directstream

import (
	"context"
	"fmt"
	"net/http"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util/result"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// URL Stream
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ Stream = (*UrlStream)(nil)

// UrlStream is an HTTP-proxied stream sourced from an arbitrary URL (e.g. from a plugin).
type UrlStream struct {
	httpBaseStream
}

func (s *UrlStream) Type() nativeplayer.StreamType {
	return nativeplayer.StreamTypeURL
}

func (s *UrlStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return s.httpBaseStream.loadPlaybackInfo(s.Type())
}

func (s *UrlStream) GetStreamHandler() http.Handler {
	return s.httpBaseStream.getStreamHandler(s)
}

func (s *UrlStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return getAttachmentByName(s.manager.playbackCtx, s, filename)
}

type PlayUrlStreamOptions struct {
	StreamUrl    string
	AnidbEpisode string
	Media        *anilist.BaseAnime
	ClientId     string
}

// PlayUrlStream starts built-in player playback for an arbitrary HTTP URL with progress tracking.
func (m *Manager) PlayUrlStream(ctx context.Context, opts PlayUrlStreamOptions) (err error) {
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
		return fmt.Errorf("cannot play URL stream, could not create episode collection: %w", err)
	}

	episode, ok := episodeCollection.FindEpisodeByAniDB(opts.AnidbEpisode)
	if !ok {
		return fmt.Errorf("cannot play URL stream, could not find episode: %s", opts.AnidbEpisode)
	}

	stream := &UrlStream{
		httpBaseStream: httpBaseStream{
			streamUrl: opts.StreamUrl,
			filepath:  "",
			BaseStream: BaseStream{
				manager:               m,
				logger:                m.Logger,
				clientId:              opts.ClientId,
				media:                 opts.Media,
				episode:               episode,
				episodeCollection:     episodeCollection,
				subtitleEventCache:    result.NewMap[string, *mkvparser.SubtitleEvent](),
				activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
			},
		},
	}

	go m.loadStream(stream)

	return nil
}
