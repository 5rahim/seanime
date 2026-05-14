package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/util/result"
	"seanime/internal/videocore"
	"strconv"
	"strings"
	"sync"
	"time"

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
		m.playbackMu.Lock()
		stream, ok := m.currentStream.Get()
		m.playbackMu.Unlock()
		if !ok {
			http.Error(w, "no stream", http.StatusInternalServerError)
			return
		}

		playbackInfo, err := stream.LoadPlaybackInfo()
		if err != nil || playbackInfo == nil {
			http.Error(w, "stream is not ready", http.StatusInternalServerError)
			return
		}

		requestStreamID := r.URL.Query().Get("id")
		if requestStreamID == "" || requestStreamID != playbackInfo.ID {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}

		stream.GetStreamHandler().ServeHTTP(w, r)
	})
}

func (m *Manager) BeginOpen(clientId string, step string, onCancel func()) bool {
	// if there's a current stream, stop it
	m.playbackMu.Lock()
	replacedPlaybackId := m.currentPlaybackId
	replacedPlaybackClient := m.currentPlaybackClient
	previousStream, cancelPlayback, _ := m.releaseCurrentStreamLocked(nil)
	if previousStream != nil && replacedPlaybackId != "" {
		m.replacedPlaybackId = replacedPlaybackId
		m.replacedPlaybackClient = replacedPlaybackClient
	}
	m.clearPreparationLocked()
	m.clearCurrentPlaybackIdentityLocked()
	m.playbackMu.Unlock()

	m.cancelAndTerminateStream(previousStream, cancelPlayback)

	m.playbackMu.Lock()
	m.preparingClientID = clientId
	m.preparationCanceled = false
	m.preparationCancelFunc = onCancel
	ok := m.updateOpenStepLocked(clientId, step)
	m.playbackMu.Unlock()

	return ok
}

func (m *Manager) UpdateOpenStep(clientId string, step string) bool {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	return m.updateOpenStepLocked(clientId, step)
}

func (m *Manager) IsOpenActive(clientId string) bool {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	if m.preparingClientID == "" {
		return true
	}

	if clientId != "" && m.preparingClientID != clientId {
		return true
	}

	return !m.preparationCanceled
}

func (m *Manager) CancelOpen(clientId string) bool {
	m.playbackMu.Lock()
	cancelFunc, ok := m.cancelPreparationLocked(clientId, true)
	m.playbackMu.Unlock()
	if !ok {
		return false
	}
	if cancelFunc != nil {
		cancelFunc()
	}
	return true
}

func (m *Manager) CloseOpen(clientId string) bool {
	m.playbackMu.Lock()
	if m.preparingClientID == "" {
		m.playbackMu.Unlock()
		return false
	}
	if clientId != "" && m.preparingClientID != clientId {
		m.playbackMu.Unlock()
		return false
	}

	targetClientID := m.preparingClientID
	if clientId != "" {
		targetClientID = clientId
	}
	_, _ = m.cancelPreparationLocked(targetClientID, true)
	m.playbackMu.Unlock()

	m.nativePlayer.AbortOpen(targetClientID, "")
	return true
}

func (m *Manager) ResetOpenState(clientId string) {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	if clientId != "" && m.preparingClientID != "" && m.preparingClientID != clientId {
		return
	}

	m.clearPreparationLocked()
}

func (m *Manager) GetCurrentPlaybackIdentity() (playbackID string, clientID string, ok bool) {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()

	if m.currentPlaybackId == "" || m.currentPlaybackClient == "" {
		return "", "", false
	}

	return m.currentPlaybackId, m.currentPlaybackClient, true
}

func (m *Manager) PrepareNewStream(clientId string, step string) {
	m.BeginOpen(clientId, step, nil)
}

func (m *Manager) StreamError(err error) {
	// Clear the current stream if it exists
	if stream, ok := m.currentStream.Get(); ok {
		m.Logger.Warn().Err(err).Msgf("directstream: Terminating stream with error")
		stream.StreamError(err)
	}
}

// AbortOpen stops the stream preparation
func (m *Manager) AbortOpen(clientId string, err error) {
	m.playbackMu.Lock()
	previousStream, cancelPlayback, _ := m.releaseCurrentStreamLocked(nil)
	m.clearPreparationLocked()
	m.clearCurrentPlaybackIdentityLocked()
	m.playbackMu.Unlock()

	m.cancelAndTerminateStream(previousStream, cancelPlayback)

	m.Logger.Debug().Msgf("directstream: Signaling native player to abort stream preparation, reason: %s", err.Error())
	m.nativePlayer.AbortOpen(clientId, err.Error())
}

func (m *Manager) updateOpenStepLocked(clientId string, step string) bool {
	if m.preparingClientID == clientId && m.preparationCanceled {
		m.Logger.Debug().Str("clientId", clientId).Msg("directstream: Skipping open step for cancelled preparation")
		return false
	}

	if m.preparingClientID == "" {
		m.preparingClientID = clientId
	}

	m.Logger.Debug().Msgf("directstream: Signaling native player that a new stream is starting")
	m.nativePlayer.OpenAndAwait(clientId, step)
	return true
}

func (m *Manager) clearPreparationLocked() {
	m.preparingClientID = ""
	m.preparationCanceled = false
	m.preparationCancelFunc = nil
}

func (m *Manager) clearCurrentPlaybackIdentityLocked() {
	m.currentPlaybackId = ""
	m.currentPlaybackClient = ""
}

// releaseCurrentStreamLocked
func (m *Manager) releaseCurrentStreamLocked(target Stream) (stream Stream, cancel context.CancelFunc, ok bool) {
	currentStream, hasCurrentStream := m.currentStream.Get()
	if target != nil {
		if !hasCurrentStream || currentStream != target {
			return nil, nil, false
		}
	}

	cancel = m.playbackCtxCancelFunc
	m.playbackCtx = nil
	m.playbackCtxCancelFunc = nil

	if hasCurrentStream {
		m.currentStream = mo.None[Stream]()
		return currentStream, cancel, true
	}

	return nil, cancel, target == nil
}

func (m *Manager) cancelAndTerminateStream(stream Stream, cancel context.CancelFunc) {
	if cancel != nil {
		m.Logger.Trace().Msg("directstream: Cancelling playback context")
		cancel()
	}

	if stream != nil {
		m.Logger.Debug().Msg("directstream: Terminating current stream")
		stream.Terminate()
	}
}

func (m *Manager) isCurrentStreamLocked(stream Stream) bool {
	currentStream, ok := m.currentStream.Get()
	return ok && currentStream == stream
}

func (m *Manager) clearStreamLoadingState(stream Stream) {
	m.playbackMu.Lock()
	_, cancelPlayback, ok := m.releaseCurrentStreamLocked(stream)
	m.playbackMu.Unlock()
	if !ok {
		return
	}
	if cancelPlayback != nil {
		cancelPlayback()
	}
}

func (m *Manager) shouldHandleTerminatedEventLocked(event *videocore.VideoTerminatedEvent, stream Stream) bool {
	if event.GetClientId() != "" && event.GetClientId() != stream.ClientId() {
		return false
	}

	if m.currentPlaybackId == "" {
		return true
	}

	if event.GetPlaybackId() == "" {
		return false
	}

	return event.GetPlaybackId() == m.currentPlaybackId
}

func (m *Manager) shouldIgnoreReplacedTerminationLocked(event *videocore.VideoTerminatedEvent) bool {
	if m.replacedPlaybackId == "" || event.GetPlaybackId() != m.replacedPlaybackId {
		return false
	}

	if m.replacedPlaybackClient != "" && event.GetClientId() != "" && event.GetClientId() != m.replacedPlaybackClient {
		return false
	}

	m.Logger.Debug().Str("playbackId", event.GetPlaybackId()).Msg("directstream: Ignoring terminated event from replaced stream")
	m.replacedPlaybackId = ""
	m.replacedPlaybackClient = ""
	return true
}

func (m *Manager) cancelPreparationLocked(clientId string, clearCancelFunc bool) (func(), bool) {
	if clientId != "" && m.preparingClientID != "" && m.preparingClientID != clientId {
		return nil, false
	}

	if m.preparingClientID == "" {
		m.preparingClientID = clientId
	}
	if clientId != "" {
		m.preparingClientID = clientId
	}
	m.preparationCanceled = true

	cancelFunc := m.preparationCancelFunc
	if clearCancelFunc {
		m.preparationCancelFunc = nil
	}

	return cancelFunc, true
}

func (m *Manager) shouldStopOpeningLocked(clientId string) bool {
	return m.preparingClientID == clientId && m.preparationCanceled
}

func (m *Manager) discardCurrentStreamLocked(stream Stream) {
	if currentStream, ok := m.currentStream.Get(); ok && currentStream == stream {
		m.currentStream = mo.None[Stream]()
		m.playbackCtx = nil
		m.playbackCtxCancelFunc = nil
	}
}

// loadStream loads a new stream and keeps the control paths responsive while metadata is being prepared.
func (m *Manager) loadStream(stream Stream) {
	if !m.UpdateOpenStep(stream.ClientId(), "Loading stream...") {
		return
	}

	m.Logger.Debug().Msgf("directstream: Loading stream")

	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	if setter, ok := stream.(interface{ setPlaybackCancelFunc(context.CancelFunc) }); ok {
		setter.setPlaybackCancelFunc(cancel)
	}

	m.playbackMu.Lock()
	m.currentStream = mo.Some(stream)
	m.playbackCtx = ctx
	m.playbackCtxCancelFunc = cancel
	m.clearCurrentPlaybackIdentityLocked()
	m.playbackMu.Unlock()

	m.Logger.Debug().Msgf("directstream: Loading content type")
	if !m.UpdateOpenStep(stream.ClientId(), "Loading metadata...") {
		m.clearStreamLoadingState(stream)
		return
	}
	// Load the content type
	contentType := stream.LoadContentType()
	if contentType == "" {
		m.Logger.Error().Msg("directstream: Failed to load content type")
		m.preStreamError(stream, fmt.Errorf("failed to load content type"))
		return
	}
	m.playbackMu.Lock()
	shouldStopOpening := ctx.Err() != nil || m.shouldStopOpeningLocked(stream.ClientId()) || !m.isCurrentStreamLocked(stream)
	m.playbackMu.Unlock()
	if shouldStopOpening {
		m.clearStreamLoadingState(stream)
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
	m.playbackMu.Lock()
	shouldStopOpening = ctx.Err() != nil || m.shouldStopOpeningLocked(stream.ClientId()) || !m.isCurrentStreamLocked(stream)
	if shouldStopOpening {
		m.playbackMu.Unlock()
		m.clearStreamLoadingState(stream)
		return
	}
	m.currentPlaybackId = playbackInfo.ID
	m.currentPlaybackClient = stream.ClientId()
	m.clearPreparationLocked()
	m.playbackMu.Unlock()

	// Shut the mkv parser logger
	//parser, ok := playbackInfo.MkvMetadataParser.Get()
	//if ok {
	//	parser.SetLoggerEnabled(false)
	//}

	m.Logger.Debug().Msgf("directstream: Signaling native player that stream is ready")
	m.nativePlayer.Watch(stream.ClientId(), playbackInfo)
}

func (m *Manager) listenToPlayerEvents() {
	go func() {
		defer func() {
			m.Logger.Trace().Msg("directstream: Stream loop goroutine exited")
		}()

		for {
			select {
			case event := <-m.videoCoreSubscriber.Events():
				if !event.IsNativePlayer() {
					continue
				}

				m.playbackMu.Lock()
				terminatedEvent, isTerminated := event.(*videocore.VideoTerminatedEvent)
				if isTerminated && m.shouldIgnoreReplacedTerminationLocked(terminatedEvent) {
					m.playbackMu.Unlock()
					continue
				}

				cs, ok := m.currentStream.Get()
				if !ok {
					var cancelFunc func()
					shouldCancel := false
					if isTerminated {
						cancelFunc, shouldCancel = m.cancelPreparationLocked(event.GetClientId(), true)
					}
					m.playbackMu.Unlock()
					if shouldCancel {
						if cancelFunc != nil {
							cancelFunc()
						}
					}
					continue
				}
				if isTerminated {
					if !m.shouldHandleTerminatedEventLocked(terminatedEvent, cs) {
						m.playbackMu.Unlock()
						continue
					}
					m.clearPreparationLocked()
					m.playbackMu.Unlock()

					m.Logger.Debug().Msgf("directstream: Video terminated")
					m.unloadStream(cs)
					continue
				}

				m.playbackMu.Unlock()

				if event.GetClientId() != cs.ClientId() {
					continue
				}

				playbackInfo, err := cs.LoadPlaybackInfo()
				if err != nil || playbackInfo == nil {
					continue
				}
				if playbackInfo.ID != "" && event.GetPlaybackId() != "" && event.GetPlaybackId() != playbackInfo.ID {
					continue
				}

				switch event := event.(type) {
				case *videocore.VideoLoadedMetadataEvent:
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
				case *videocore.VideoSeekedEvent:
					m.Logger.Trace().Float64("currentTime", event.CurrentTime).Msg("directstream: Video seeked, refreshing subtitle stream")
					go m.startSubtitleStreamForTime(cs, playbackInfo, event.CurrentTime, event.Duration)
				case *videocore.VideoErrorEvent:
					m.Logger.Debug().Msgf("directstream: Video error, Error: %s", event.Error)
					cs.StreamError(fmt.Errorf("%s", event.Error))
				case *videocore.SubtitleFileUploadedEvent:
					m.Logger.Debug().Msgf("directstream: Subtitle file uploaded, Filename: %s", event.Filename)
					cs.OnSubtitleFileUploaded(event.Filename, event.Content)
				case *videocore.VideoCompletedEvent:
					m.Logger.Debug().Msgf("directstream: Video completed")

					if baseStream, ok := cs.(*BaseStream); ok {
						baseStream.updateProgress.Do(func() {
							mediaId := baseStream.media.GetID()
							epNum := baseStream.episode.GetProgressNumber()
							totalEpisodes := baseStream.media.GetTotalEpisodeCount() // total episode count or -1

							_ = baseStream.manager.platformRef.Get().UpdateEntryProgress(context.Background(), mediaId, epNum, &totalEpisodes)
						})
					}
				}
			}
		}
	}()
}

func (m *Manager) unloadStream(targets ...Stream) {
	m.Logger.Debug().Msg("directstream: Unloading current stream")

	var target Stream
	if len(targets) > 0 {
		target = targets[0]
	}
	m.playbackMu.Lock()
	stream, cancelPlayback, ok := m.releaseCurrentStreamLocked(target)
	m.playbackMu.Unlock()
	if ok {
		m.cancelAndTerminateStream(stream, cancelPlayback)
	}
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
	playbackCancelFunc     context.CancelFunc
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

func (s *BaseStream) setPlaybackCancelFunc(cancel context.CancelFunc) {
	s.playbackCancelFunc = cancel
}

func (s *BaseStream) Terminate() {
	s.terminateOnce.Do(func() {
		// Cancel the playback context
		// This will snowball and cancel other stuff
		if s.playbackCancelFunc != nil {
			s.playbackCancelFunc()
		}

		// Cancel all active subtitle streams
		s.activeSubtitleStreams.Range(func(_ string, s *SubtitleStream) bool {
			s.Stop(s.completed)
			return true
		})
		s.activeSubtitleStreams.Clear()

		s.subtitleEventCache.Clear()
	})
}

func (s *BaseStream) StreamError(err error) {
	s.logger.Error().Err(err).Msg("directstream: Stream error occurred")
	s.manager.playbackMu.Lock()
	if !s.manager.isCurrentStreamLocked(s) {
		s.manager.playbackMu.Unlock()
		return
	}
	s.manager.playbackMu.Unlock()

	s.manager.nativePlayer.Error(s.clientId, err)
	s.manager.unloadStream(s)
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
	m.playbackMu.Lock()
	if !m.isCurrentStreamLocked(stream) {
		m.playbackMu.Unlock()
		return
	}
	m.clearPreparationLocked()
	m.playbackMu.Unlock()

	m.nativePlayer.Error(stream.ClientId(), err)
	m.unloadStream(stream)
}

func overrideHeaders(dst http.Header, src http.Header) {
	if len(src) == 0 {
		return
	}

	for key, values := range src {
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func (m *Manager) getContentTypeAndLength(url string) (string, int64, error) {
	return m.getContentTypeAndLengthWithHeaders(url, nil)
}

func (m *Manager) getContentTypeAndLengthWithHeaders(url string, headers http.Header) (string, int64, error) {
	m.Logger.Trace().Msg("directstream(debrid): Fetching content type and length using HEAD request")

	// Create client with timeout for HEAD request (faster timeout since it's just headers)
	headClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try HEAD request first
	headReq, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create HEAD request: %w", err)
	}
	overrideHeaders(headReq.Header, headers)

	resp, err := headClient.Do(headReq)
	if err == nil {
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		contentLengthStr := resp.Header.Get("Content-Length")

		// Parse content length
		var length int64
		if contentLengthStr != "" {
			length, err = strconv.ParseInt(contentLengthStr, 10, 64)
			if err != nil {
				m.Logger.Error().Err(err).Str("contentType", contentType).Str("contentLength", contentLengthStr).
					Msg("directstream(debrid): Failed to parse content length from header")
				return "", 0, fmt.Errorf("failed to parse content length: %w", err)
			}
		}

		// If we have content type from a successful response, return early
		if contentType != "" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return contentType, length, nil
		}

		m.Logger.Trace().Int("status", resp.StatusCode).Msg("directstream(debrid): HEAD response not usable, falling back to GET")
	} else {
		m.Logger.Trace().Err(err).Msg("directstream(debrid): HEAD request failed")
	}

	// Fall back to GET with Range request (either HEAD failed or no content type in headers)
	m.Logger.Trace().Msg("directstream(debrid): Falling back to GET request")

	// Create client with longer timeout for GET request (downloading content)
	getClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create GET request: %w", err)
	}

	overrideHeaders(req.Header, headers)
	req.Header.Set("Range", "bytes=0-511")

	resp, err = getClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse total length from Content-Range header (for Range requests)
	// Format: "bytes 0-511/1234567" where 1234567 is the total size
	var length int64
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		// Extract total size from Content-Range header
		if idx := strings.LastIndex(contentRange, "/"); idx != -1 {
			totalSizeStr := contentRange[idx+1:]
			if totalSizeStr != "*" { // "*" means unknown size
				length, err = strconv.ParseInt(totalSizeStr, 10, 64)
				if err != nil {
					m.Logger.Warn().Err(err).Str("contentRange", contentRange).
						Msg("directstream(debrid): Failed to parse total size from Content-Range")
				}
			}
		}
	} else if contentLengthStr := resp.Header.Get("Content-Length"); contentLengthStr != "" {
		// Fallback to Content-Length if Content-Range not present (server might not support ranges)
		length, err = strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			m.Logger.Warn().Err(err).Str("contentLength", contentLengthStr).
				Msg("directstream(debrid): Failed to parse content length from GET response")
		}
	}

	// Check if server provided Content-Type in GET response
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		return contentType, length, nil
	}

	// Read only what's needed for content type detection
	buf := make([]byte, 512)
	n, err := io.ReadFull(resp.Body, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", 0, fmt.Errorf("failed to read response body: %w", err)
	}

	contentType = http.DetectContentType(buf[:n])

	return contentType, length, nil
}

type StreamInfo struct {
	ContentType   string
	ContentLength int64
}

func (m *Manager) FetchStreamInfo(streamUrl string) (info *StreamInfo, canStream bool) {
	return m.FetchStreamInfoWithHeaders(streamUrl, nil)
}

func (m *Manager) FetchStreamInfoWithHeaders(streamUrl string, headers http.Header) (info *StreamInfo, canStream bool) {
	_, isArchive := IsArchive(streamUrl)

	m.Logger.Debug().Str("url", streamUrl).Msg("directstream(http): Fetching stream info")

	// If we were able to verify that the stream URL is an archive, we can't stream it
	if isArchive {
		m.Logger.Warn().Str("url", streamUrl).Msg("directstream(http): Stream URL is an archive, cannot stream")
		return nil, false
	}

	// If the stream URL has an extension, we can stream it
	// if hasExtension {
	// 	// Strip query params before checking extension
	// 	cleanUrl := streamUrl
	// 	if idx := strings.IndexByte(cleanUrl, '?'); idx != -1 {
	// 		cleanUrl = cleanUrl[:idx]
	// 	}
	// 	ext := filepath.Ext(cleanUrl)
	// 	// If not a valid video extension, we can't stream it
	// 	if !util.IsValidVideoExtension(ext) {
	// 		m.Logger.Warn().Str("url", streamUrl).Str("ext", ext).Msg("directstream(http): Stream URL has an invalid video extension, cannot stream")
	// 		return nil, false
	// 	}
	// }

	// We'll fetch headers to get the info
	// If the headers are not available, we can't stream it

	contentType, contentLength, err := m.getContentTypeAndLengthWithHeaders(streamUrl, headers)
	if err != nil {
		m.Logger.Error().Err(err).Str("url", streamUrl).Msg("directstream(http): Failed to fetch content type and length")
		return nil, false
	}

	// If not a video content type, we can't stream it
	if !strings.HasPrefix(contentType, "video/") && contentType != "application/octet-stream" && contentType != "application/force-download" {
		m.Logger.Warn().Str("url", streamUrl).Str("contentType", contentType).Msg("directstream(http): Stream URL has an invalid content type, cannot stream")
		return nil, false
	}

	return &StreamInfo{
		ContentType:   contentType,
		ContentLength: contentLength,
	}, true
}

func IsArchive(streamUrl string) (hasExtension bool, isArchive bool) {
	// Strip query params before checking extension
	u := streamUrl
	if idx := strings.IndexByte(u, '?'); idx != -1 {
		u = u[:idx]
	}
	ext := filepath.Ext(u)
	if ext == ".zip" || ext == ".rar" {
		return true, true
	}

	if ext != "" {
		return true, false
	}

	return false, false
}
