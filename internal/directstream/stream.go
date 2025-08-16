package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/continuity"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util/result"
	"sync"

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
	// ListEntryData returns the list entry data for the current stream.
	ListEntryData() *anime.EntryListData
	// EpisodeCollection returns the episode collection for the media of the current stream.
	EpisodeCollection() *anime.EpisodeCollection
	// LoadPlaybackInfo loads and returns the playback info.
	LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error)
	// GetAttachmentByName returns the attachment by name for the stream.
	// It is used to serve fonts and other attachments.
	GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool)
	// GetStreamHandler returns the stream handler.
	GetStreamHandler() http.Handler
	// StreamError is called when an error occurs while streaming.
	// This is used to notify the native player that an error occurred.
	// It will close the stream.
	StreamError(err error)
	// Terminate ends the stream.
	// Once this is called, the stream should not be used anymore.
	Terminate()
	// GetSubtitleEventCache accesses the subtitle event cache.
	GetSubtitleEventCache() *result.Map[string, *mkvparser.SubtitleEvent]
	// OnSubtitleFileUploaded is called when a subtitle file is uploaded.
	OnSubtitleFileUploaded(filename string, content string)
}

func (m *Manager) getStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream, ok := m.currentStream.Get()
		if !ok {
			http.Error(w, "no stream", http.StatusInternalServerError)
			return
		}
		stream.GetStreamHandler().ServeHTTP(w, r)
	})
}

func (m *Manager) PrepareNewStream(clientId string, step string) {
	m.prepareNewStream(clientId, step)
}

func (m *Manager) prepareNewStream(clientId string, step string) {
	// Cancel the previous playback
	if m.playbackCtxCancelFunc != nil {
		m.Logger.Trace().Msgf("directstream: Cancelling previous playback")
		m.playbackCtxCancelFunc()
		m.playbackCtxCancelFunc = nil
	}

	// Clear the current stream if it exists
	if stream, ok := m.currentStream.Get(); ok {
		m.Logger.Debug().Msgf("directstream: Terminating previous stream before preparing new stream")
		stream.Terminate()
		m.currentStream = mo.None[Stream]()
	}

	m.Logger.Debug().Msgf("directstream: Signaling native player that a new stream is starting")
	// Signal the native player that a new stream is starting
	m.nativePlayer.OpenAndAwait(clientId, step)
}

// loadStream loads a new stream and cancels the previous one.
// Caller should use mutex to lock the manager.
func (m *Manager) loadStream(stream Stream) {
	m.prepareNewStream(stream.ClientId(), "Loading stream...")

	m.Logger.Debug().Msgf("directstream: Loading stream")
	m.currentStream = mo.Some(stream)

	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	m.playbackCtx = ctx
	m.playbackCtxCancelFunc = cancel

	m.Logger.Debug().Msgf("directstream: Loading content type")
	m.nativePlayer.OpenAndAwait(stream.ClientId(), "Loading metadata...")
	// Load the content type
	contentType := stream.LoadContentType()
	if contentType == "" {
		m.Logger.Error().Msg("directstream: Failed to load content type")
		m.preStreamError(stream, fmt.Errorf("failed to load content type"))
		return
	}

	m.Logger.Debug().Msgf("directstream: Signaling native player that metadata is being loaded")

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
}

func (m *Manager) listenToNativePlayerEvents() {
	go func() {
		defer func() {
			m.Logger.Trace().Msg("directstream: Stream loop goroutine exited")
		}()

		for {
			select {
			case event := <-m.nativePlayerSubscriber.Events():
				cs, ok := m.currentStream.Get()
				if !ok {
					continue
				}

				if event.GetClientId() != "" && event.GetClientId() != cs.ClientId() {
					continue
				}
				switch event := event.(type) {
				case *nativeplayer.VideoPausedEvent:
					m.Logger.Debug().Msgf("directstream: Video paused")

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.UpdateAnimeActivity(int(event.CurrentTime), int(event.Duration), true)
					}
				case *nativeplayer.VideoResumedEvent:
					m.Logger.Debug().Msgf("directstream: Video resumed")

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.UpdateAnimeActivity(int(event.CurrentTime), int(event.Duration), false)
					}
				case *nativeplayer.VideoEndedEvent:
					m.Logger.Debug().Msgf("directstream: Video ended")

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.Close()
					}
				case *nativeplayer.VideoSeekedEvent:
					m.Logger.Debug().Msgf("directstream: Video seeked, CurrentTime: %f", event.CurrentTime)
					// Convert video timestamp to byte offset for subtitle extraction
					// if event.CurrentTime > 0 {
					// 	cs.ServeSubtitlesFromTime(event.CurrentTime)
					// }
				case *nativeplayer.VideoLoadedMetadataEvent:
					m.Logger.Debug().Msgf("directstream: Video loaded metadata")
					// Start subtitle extraction from the beginning
					// cs.ServeSubtitlesFromTime(0.0)
					if lfStream, ok := cs.(*LocalFileStream); ok {
						subReader, err := lfStream.newReader()
						if err != nil {
							m.Logger.Error().Err(err).Msg("directstream: Failed to create subtitle reader")
							cs.StreamError(fmt.Errorf("failed to create subtitle reader: %w", err))
							return
						}
						lfStream.StartSubtitleStream(lfStream, m.playbackCtx, subReader, 0)
					} else if ts, ok := cs.(*TorrentStream); ok {
						subReader := ts.file.NewReader()
						subReader.SetResponsive()
						ts.StartSubtitleStream(ts, m.playbackCtx, subReader, 0)
					}

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
							ID:            cs.Media().GetID(),
							Title:         cs.Media().GetPreferredTitle(),
							Image:         cs.Media().GetCoverImageSafe(),
							IsMovie:       cs.Media().IsMovie(),
							EpisodeNumber: cs.Episode().ProgressNumber,
							Progress:      int(event.CurrentTime),
							Duration:      int(event.Duration),
						})
					}
				case *nativeplayer.VideoErrorEvent:
					m.Logger.Debug().Msgf("directstream: Video error, Error: %s", event.Error)
					cs.StreamError(fmt.Errorf(event.Error))

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.Close()
					}
				case *nativeplayer.SubtitleFileUploadedEvent:
					m.Logger.Debug().Msgf("directstream: Subtitle file uploaded, Filename: %s", event.Filename)
					cs.OnSubtitleFileUploaded(event.Filename, event.Content)
				case *nativeplayer.VideoTerminatedEvent:
					m.Logger.Debug().Msgf("directstream: Video terminated")
					cs.Terminate()

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.Close()
					}
				case *nativeplayer.VideoStatusEvent:
					_ = m.continuityManager.UpdateWatchHistoryItem(&continuity.UpdateWatchHistoryItemOptions{
						CurrentTime:   event.Status.CurrentTime,
						Duration:      event.Status.Duration,
						MediaId:       cs.Media().GetID(),
						EpisodeNumber: cs.Episode().GetEpisodeNumber(),
						Kind:          continuity.MediastreamKind,
					})

					// Discord
					if m.discordPresence != nil && !*m.isOffline {
						go m.discordPresence.UpdateAnimeActivity(int(event.Status.CurrentTime), int(event.Status.Duration), event.Status.Paused)
					}
				case *nativeplayer.VideoCompletedEvent:
					m.Logger.Debug().Msgf("directstream: Video completed")

					if baseStream, ok := cs.(*BaseStream); ok {
						baseStream.updateProgress.Do(func() {
							mediaId := baseStream.media.GetID()
							epNum := baseStream.episode.GetProgressNumber()
							totalEpisodes := baseStream.media.GetTotalEpisodeCount() // total episode count or -1

							_ = baseStream.manager.platform.UpdateEntryProgress(context.Background(), mediaId, epNum, &totalEpisodes)
						})
					}
				}
			}
		}
	}()
}

func (m *Manager) unloadStream() {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	m.Logger.Debug().Msg("directstream: Unloading current stream")

	// Cancel any existing playback context first
	if m.playbackCtxCancelFunc != nil {
		m.Logger.Trace().Msg("directstream: Cancelling playback context")
		m.playbackCtxCancelFunc()
		m.playbackCtxCancelFunc = nil
	}

	// Clear the current stream
	if stream, ok := m.currentStream.Get(); ok {
		m.Logger.Debug().Msg("directstream: Terminating current stream")
		stream.Terminate()
	}

	m.currentStream = mo.None[Stream]()
	m.Logger.Debug().Msg("directstream: Stream unloaded successfully")
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BaseStream struct {
	logger                 *zerolog.Logger
	clientId               string
	contentType            string
	contentTypeOnce        sync.Once
	episode                *anime.Episode
	media                  *anilist.BaseAnime
	listEntryData          *anime.EntryListData
	episodeCollection      *anime.EpisodeCollection
	playbackInfo           *nativeplayer.PlaybackInfo
	playbackInfoErr        error
	playbackInfoOnce       sync.Once
	subtitleEventCache     *result.Map[string, *mkvparser.SubtitleEvent]
	terminateOnce          sync.Once
	serveContentCancelFunc context.CancelFunc
	filename               string // Name of the file being streamed, if applicable

	// Subtitle stream management
	activeSubtitleStreams *result.Map[string, *SubtitleStream]

	manager        *Manager
	updateProgress sync.Once
}

var _ Stream = (*BaseStream)(nil)

func (s *BaseStream) GetAttachmentByName(filename string) (*mkvparser.AttachmentInfo, bool) {
	return nil, false
}

func (s *BaseStream) GetStreamHandler() http.Handler {
	return nil
}

func (s *BaseStream) LoadContentType() string {
	return s.contentType
}

func (s *BaseStream) LoadPlaybackInfo() (*nativeplayer.PlaybackInfo, error) {
	return s.playbackInfo, s.playbackInfoErr
}

func (s *BaseStream) Type() nativeplayer.StreamType {
	return ""
}

func (s *BaseStream) Media() *anilist.BaseAnime {
	return s.media
}

func (s *BaseStream) Episode() *anime.Episode {
	return s.episode
}

func (s *BaseStream) ListEntryData() *anime.EntryListData {
	return s.listEntryData
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
		s.activeSubtitleStreams.Range(func(_ string, s *SubtitleStream) bool {
			s.cleanupFunc()
			return true
		})
		s.activeSubtitleStreams.Clear()

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
	case ".mkv":
		//return "video/x-matroska"
		return "video/webm"
	case ".webm", ".m4v":
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

func (m *Manager) preStreamError(stream Stream, err error) {
	stream.Terminate()
	m.nativePlayer.Error(stream.ClientId(), err)
	m.unloadStream()
}
