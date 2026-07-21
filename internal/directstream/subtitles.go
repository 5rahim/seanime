package directstream

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"seanime/internal/events"
	"seanime/internal/mkvparser"
	"seanime/internal/player"
	"seanime/internal/util"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type SubtitleStream struct {
	stream    Stream
	logger    *zerolog.Logger
	parser    *mkvparser.MetadataParser
	reader    io.ReadSeekCloser
	offset    int64
	request   subtitleRequest
	completed atomic.Bool // ran until the EOF

	cleanupFunc func()
	onStop      func()
	stopOnce    sync.Once
}

type subtitleRequest struct {
	playbackID string
	generation int64
	seekTime   float64
}

const (
	subtitleBackoffBytes   int64 = 1024 * 1024
	streamDedupWindowBytes       = 1024 * 1024
)

type subtitleFlushConfig struct {
	flushInterval       time.Duration
	maxBatchSize        int
	sleepAfterFullBatch time.Duration
	minSendInterval     time.Duration
}

func subtitleFlushConfigFor(streamType player.PlaybackType, offset int64) subtitleFlushConfig {
	config := subtitleFlushConfig{
		flushInterval:       100 * time.Millisecond,
		maxBatchSize:        500,
		sleepAfterFullBatch: 200 * time.Millisecond,
	}

	if streamType == player.PlaybackTypeTorrent {
		config = subtitleFlushConfig{
			flushInterval:       250 * time.Millisecond,
			maxBatchSize:        25,
			sleepAfterFullBatch: 100 * time.Millisecond,
			minSendInterval:     100 * time.Millisecond,
		}
		if offset > 0 {
			config = subtitleFlushConfig{
				flushInterval:       100 * time.Millisecond,
				maxBatchSize:        35,
				sleepAfterFullBatch: 50 * time.Millisecond,
				minSendInterval:     75 * time.Millisecond,
			}
		}
	}

	if streamType == player.PlaybackTypeLocalFile {
		config = subtitleFlushConfig{
			flushInterval:       300 * time.Millisecond,
			maxBatchSize:        50,
			sleepAfterFullBatch: 500 * time.Millisecond,
		}
		// local resume/seek need to catch up very quickly
		if offset > 0 {
			config = subtitleFlushConfig{
				flushInterval:       75 * time.Millisecond,
				maxBatchSize:        200,
				sleepAfterFullBatch: 25 * time.Millisecond,
			}
		}
	}

	if streamType == player.PlaybackTypeDebrid || streamType == player.PlaybackTypeURL || streamType == player.PlaybackTypeNakama {
		if offset > 0 {
			config = subtitleFlushConfig{
				flushInterval:       10 * time.Millisecond,
				maxBatchSize:        1000,
				sleepAfterFullBatch: 0 * time.Millisecond,
			}
		}
	}

	return config
}

func subtitleEventId(event *mkvparser.SubtitleEvent) string {
	if event == nil {
		return ""
	}

	hash := fnv.New64a()
	_, _ = hash.Write([]byte(event.Text))

	if len(event.ExtraData) > 0 {
		keys := make([]string, 0, len(event.ExtraData))
		for key := range event.ExtraData {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			_, _ = hash.Write([]byte(key))
			_, _ = hash.Write([]byte{'='})
			_, _ = hash.Write([]byte(event.ExtraData[key]))
			_, _ = hash.Write([]byte{'|'})
		}
	}

	return fmt.Sprintf("%d:%s:%f:%f:%x", event.TrackNumber, event.CodecID, event.StartTime, event.Duration, hash.Sum64())
}

func (s *BaseStream) shouldSendSubtitleEvent(event *mkvparser.SubtitleEvent) bool {
	if event == nil {
		return false
	}
	if s.subtitleEventCache == nil {
		return true
	}

	_, loaded := s.subtitleEventCache.LoadOrStore(subtitleEventId(event), event)
	return !loaded
}

func (s *BaseStream) sendSubtitleEvents(ctx context.Context, stream Stream, events []*mkvparser.SubtitleEvent, config subtitleFlushConfig, request subtitleRequest) bool {
	if len(events) == 0 {
		return true
	}
	if ctx.Err() != nil || request.generation != s.subtitleGeneration.Load() {
		return false
	}

	s.manager.playbackMu.Lock()
	target := s.manager.currentPlaybackTarget
	s.manager.playbackMu.Unlock()
	if target != PlaybackTargetVideoCore || s.manager.nativePlayer == nil {
		// MpvCore lets libmpv demux and render embedded subtitles directly.
		return true
	}

	s.subtitleSendMu.Lock()
	defer s.subtitleSendMu.Unlock()

	if ctx.Err() != nil || request.generation != s.subtitleGeneration.Load() {
		return false
	}

	if !s.subtitleLastSent.IsZero() && s.subtitleLastSentGen == request.generation && config.minSendInterval > 0 {
		if !s.waitForSubtitleSend(ctx, config.minSendInterval) {
			return false
		}
		if request.generation != s.subtitleGeneration.Load() {
			return false
		}
	}

	s.manager.nativePlayer.SubtitleEventsWithGen(stream.ClientId(), events, request.playbackID, request.generation, request.seekTime)
	s.subtitleLastSentGen = request.generation
	s.subtitleLastSent = time.Now()
	return true
}

func (s *BaseStream) waitForSubtitleSend(ctx context.Context, minSendInterval time.Duration) bool {
	if s.subtitleLastSent.IsZero() {
		return true
	}

	wait := minSendInterval - time.Since(s.subtitleLastSent)
	if wait <= 0 {
		return true
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func subtitleOffsetForTime(playbackInfo *player.PlaybackInfo, currentTime float64, duration float64) int64 {
	if playbackInfo == nil || playbackInfo.ContentLength <= 0 || currentTime <= 0 {
		return 0
	}

	// Try to seek using Matroska cues if available
	if playbackInfo.MkvMetadata != nil && len(playbackInfo.MkvMetadata.Cues) > 0 {
		preroll := 10.0 // 10 seconds default for text formats
		for _, track := range playbackInfo.MkvMetadata.SubtitleTracks {
			if track.CodecID == "S_HDMV/PGS" {
				preroll = 30.0 // 30 seconds for PGS
				break
			}
		}

		targetTimeNs := uint64(math.Max(currentTime-preroll, 0) * 1e9)
		i := sort.Search(len(playbackInfo.MkvMetadata.Cues), func(i int) bool {
			return playbackInfo.MkvMetadata.Cues[i].Time >= targetTimeNs
		})

		if i > 0 && (i == len(playbackInfo.MkvMetadata.Cues) || playbackInfo.MkvMetadata.Cues[i].Time > targetTimeNs) {
			i--
		}

		if i >= len(playbackInfo.MkvMetadata.Cues) {
			i = len(playbackInfo.MkvMetadata.Cues) - 1
		}

		cue := playbackInfo.MkvMetadata.Cues[i]
		return int64(cue.Position)
	}

	effectiveDuration := duration
	if effectiveDuration <= 0 && playbackInfo.MkvMetadata != nil {
		effectiveDuration = playbackInfo.MkvMetadata.Duration
	}
	if effectiveDuration <= 0 {
		return 0
	}

	progress := currentTime / effectiveDuration
	if math.IsNaN(progress) || math.IsInf(progress, 0) {
		return 0
	}
	progress = min(max(progress, 0), 1)

	offset := int64(progress * float64(playbackInfo.ContentLength))
	maxOffset := max(playbackInfo.ContentLength-subtitleBackoffBytes, 0)
	return min(max(offset, 0), maxOffset)
}

func subtitleOffsetDistance(a int64, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

func (m *Manager) startSubtitleStreamForTime(stream Stream, playbackInfo *player.PlaybackInfo, currentTime float64, duration float64) {
	if playbackInfo == nil {
		return
	}
	if _, ok := playbackInfo.MkvMetadataParser.Get(); !ok {
		return
	}
	playbackCtx := m.playbackCtx
	if playbackCtx == nil {
		return
	}

	offset := subtitleOffsetForTime(playbackInfo, currentTime, duration)

	baseStream := stream.GetBaseStream()
	request := baseStream.beginSubtitleSeek(currentTime)

	switch s := stream.(type) {
	case *LocalFileStream:
		reader, err := s.newReader()
		if err != nil {
			m.Logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Failed to create subtitle reader after seek")
			return
		}
		s.startSubtitleStream(s, playbackCtx, reader, offset, request)
	case *TorrentStream:
		reader := s.newSubtitleReader()
		s.startSubtitleStream(s, playbackCtx, reader, offset, request)
	case *UrlStream:
		reader, err := s.newMetadataReader()
		if err != nil {
			m.Logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Failed to create subtitle reader after seek")
			return
		}
		s.startSubtitleStream(s, playbackCtx, reader, offset, request)
	case *DebridStream:
		reader, err := s.newMetadataReader()
		if err != nil {
			m.Logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Failed to create subtitle reader after seek")
			return
		}
		s.startSubtitleStream(s, playbackCtx, reader, offset, request)
	case *Nakama:
		reader, err := s.newMetadataReader()
		if err != nil {
			m.Logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Failed to create subtitle reader after seek")
			return
		}
		s.startSubtitleStream(s, playbackCtx, reader, offset, request)
	}
}

func (s *BaseStream) beginSubtitleSeek(seekTime float64) subtitleRequest {
	s.subtitleSeekMu.Lock()
	defer s.subtitleSeekMu.Unlock()

	request := subtitleRequest{
		generation: s.subtitleGeneration.Add(1),
		seekTime:   seekTime,
	}
	if s.playbackInfo != nil {
		request.playbackID = s.playbackInfo.ID
	}

	s.activeSubtitleStreams.Range(func(_ string, value *SubtitleStream) bool {
		value.Stop(false)
		return true
	})

	return request
}

func (s *SubtitleStream) Stop(completed bool) {
	s.stopOnce.Do(func() {
		s.logger.Debug().Int64("offset", s.offset).Msg("directstream: Stopping subtitle stream")
		s.completed.Store(completed)
		if s.onStop != nil {
			s.onStop()
		}
		if s.cleanupFunc != nil {
			s.cleanupFunc()
		}
	})
}

// StartSubtitleStreamP starts a subtitle stream for the given stream at the given offset with a specified backoff bytes.
func (s *BaseStream) StartSubtitleStreamP(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64, backoffBytes int64) {
	request := subtitleRequest{}
	if s.playbackInfo != nil {
		request.playbackID = s.playbackInfo.ID
	}
	s.startSubtitleStreamP(stream, playbackCtx, newReader, offset, backoffBytes, request)
}

func (s *BaseStream) startSubtitleStreamP(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64, backoffBytes int64, request subtitleRequest) {
	if playbackCtx == nil {
		_ = newReader.Close()
		return
	}
	if request.generation != s.subtitleGeneration.Load() {
		_ = newReader.Close()
		return
	}

	mkvMetadataParser, ok := s.playbackInfo.MkvMetadataParser.Get()
	if !ok {
		_ = newReader.Close()
		return
	}

	s.subtitleSeekMu.Lock()
	if request.generation != s.subtitleGeneration.Load() {
		s.subtitleSeekMu.Unlock()
		_ = newReader.Close()
		return
	}

	// Check if we have a completed subtitle stream for this offset
	shouldContinue := true
	skipReason := ""
	s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
		if value.request.generation != request.generation {
			return true
		}
		if subtitleOffsetDistance(value.offset, offset) <= streamDedupWindowBytes {
			skipReason = "nearby stream already active"
			shouldContinue = false
			return false
		}

		// If a stream is completed and its offset comes before this one, we don't need to start a new stream
		// |------------------------------->| other stream
		//                    |               this stream
		//                   ^^^ starting in an area the other stream has already completed
		if offset > 0 && value.offset <= offset && value.completed.Load() {
			skipReason = "range already fulfilled"
			shouldContinue = false
			return false
		}
		return true
	})

	if !shouldContinue {
		s.subtitleSeekMu.Unlock()
		s.logger.Debug().Int64("offset", offset).Str("reason", skipReason).Msg("directstream: Skipping subtitle stream")
		_ = newReader.Close()
		return
	}

	s.logger.Trace().Int64("offset", offset).Msg("directstream: Starting new subtitle stream")
	subtitleStream := &SubtitleStream{
		stream:  stream,
		logger:  s.logger,
		parser:  mkvMetadataParser,
		reader:  newReader,
		offset:  offset,
		request: request,
	}

	ctx, subtitleCtxCancel := context.WithCancel(playbackCtx)
	subtitleStream.cleanupFunc = subtitleCtxCancel

	subtitleStreamId := uuid.New().String()
	subtitleStream.onStop = func() {
		s.activeSubtitleStreams.Delete(subtitleStreamId)
	}
	s.activeSubtitleStreams.Set(subtitleStreamId, subtitleStream)
	s.subtitleSeekMu.Unlock()

	subtitleCh, errCh, _ := subtitleStream.parser.ExtractSubtitles(ctx, newReader, offset, backoffBytes, request.seekTime)

	firstEventSentCh := make(chan struct{}) // no-op
	closeFirstEventSentOnce := sync.Once{}

	onFirstEventSent := func() {
		closeFirstEventSentOnce.Do(func() {
			s.logger.Debug().Int64("offset", offset).Msg("directstream: First subtitle event sent")
			close(firstEventSentCh) // Notify that the first subtitle event has been sent
		})
	}

	var lastSubtitleEvent *mkvparser.SubtitleEvent
	lastSubtitleEventRWMutex := sync.RWMutex{}
	setLastSubtitleEvent := func(event *mkvparser.SubtitleEvent) {
		lastSubtitleEventRWMutex.Lock()
		lastSubtitleEvent = event
		lastSubtitleEventRWMutex.Unlock()
	}

	// Check every second if we need to end this stream
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				subtitleStream.Stop(false)
				return
			case <-ticker.C:
				lastSubtitleEventRWMutex.RLock()
				lastEvent := lastSubtitleEvent
				lastSubtitleEventRWMutex.RUnlock()
				if lastEvent == nil {
					continue
				}
				shouldEnd := false
				s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
					if key != subtitleStreamId {
						// If the other stream is ahead of this stream
						// and the last subtitle event is after the other stream's offset
						// |--------------->                   this stream
						//                     |-------------> other stream
						//                    ^^^ stop this stream where it reached the tail of the other stream
						if offset < value.offset && lastEvent.HeadPos >= value.offset {
							shouldEnd = true
						}
					}
					return true
				})
				if shouldEnd {
					subtitleStream.Stop(false)
					return
				}
			}
		}
	}()

	go func() {
		defer func(reader io.ReadSeekCloser) {
			_ = reader.Close()
			s.logger.Trace().Int64("offset", offset).Msg("directstream: Closing subtitle stream goroutine")
		}(newReader)
		defer func() {
			onFirstEventSent()
			subtitleStream.Stop(subtitleStream.completed.Load())
		}()

		// Keep track if channels are active to manage loop termination
		subtitleChannelActive := true
		errorChannelActive := true

		flushConfig := subtitleFlushConfigFor(stream.Type(), offset)
		flushInterval := flushConfig.flushInterval
		maxBatchSize := flushConfig.maxBatchSize
		sleepAfterFullBatch := flushConfig.sleepAfterFullBatch

		eventBatch := make([]*mkvparser.SubtitleEvent, 0, maxBatchSize)
		flushBatch := func(fullBatch bool) {
			if len(eventBatch) == 0 {
				return
			}
			if !s.sendSubtitleEvents(ctx, stream, eventBatch, flushConfig, request) {
				eventBatch = eventBatch[:0]
				return
			}

			eventBatch = eventBatch[:0]

			if fullBatch && sleepAfterFullBatch > 0 {
				// only slow down when the parser outruns the flush timer and fills a batch completely
				time.Sleep(sleepAfterFullBatch)
			}
		}

		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
			select {
			case <-ctx.Done():
				s.logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle streaming cancelled by context")
				return

			case <-ticker.C:
				flushBatch(false)

			case subtitle, ok := <-subtitleCh:
				if !ok {
					subtitleCh = nil // Mark as exhausted
					subtitleChannelActive = false
					if !errorChannelActive { // If both channels are exhausted, exit
						flushBatch(false)
						return
					}
					continue // Continue to wait for errorChannel or ctx.Done()
				}
				if subtitle != nil {
					onFirstEventSent()
					setLastSubtitleEvent(subtitle)
					if !s.shouldSendSubtitleEvent(subtitle) {
						continue
					}

					eventBatch = append(eventBatch, subtitle)

					isFirstBatch := false
					s.subtitleSendMu.Lock()
					if s.subtitleLastSent.IsZero() || s.subtitleLastSentGen != request.generation {
						isFirstBatch = true
					}
					s.subtitleSendMu.Unlock()

					if isFirstBatch || len(eventBatch) >= maxBatchSize {
						flushBatch(true)
					}
				}

			case err, ok := <-errCh:
				if !ok {
					errCh = nil // Mark as exhausted
					errorChannelActive = false
					if !subtitleChannelActive { // If both channels are exhausted, exit
						flushBatch(false)
						return
					}
					continue // Continue to wait for subtitleChannel or ctx.Done()
				}
				// A value (error or nil) was received from errCh.
				// This is the terminal signal from the mkvparser's subtitle streaming process.
				if err != nil {
					s.logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Error streaming subtitles")
				} else {
					s.logger.Info().Int64("offset", offset).Msg("directstream: Subtitle streaming completed by parser.")
					subtitleStream.Stop(true)
				}
				flushBatch(false)
				return // Terminate goroutine
			}
		}
	}()

	//// Then wait for first subtitle event or timeout to prevent indefinite stalling
	//if offset > 0 {
	//	// Wait for cluster to be found first
	//	<-startedCh
	//
	//	select {
	//	case <-firstEventSentCh:
	//		s.logger.Debug().Int64("offset", offset).Msg("directstream: First subtitle event received, continuing")
	//	case <-time.After(3 * time.Second):
	//		s.logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle timeout reached (3s), continuing without waiting")
	//	case <-ctx.Done():
	//		s.logger.Debug().Int64("offset", offset).Msg("directstream: Context cancelled while waiting for first subtitle")
	//		return
	//	}
	//}
}

// StartSubtitleStream starts a subtitle stream for the given stream at the given offset.
//
// If the media has no MKV metadata, this function will do nothing.
func (s *BaseStream) StartSubtitleStream(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64) {
	request := subtitleRequest{}
	if s.playbackInfo != nil {
		request.playbackID = s.playbackInfo.ID
	}
	s.startSubtitleStream(stream, playbackCtx, newReader, offset, request)
}

func (s *BaseStream) startSubtitleStream(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64, request subtitleRequest) {
	backoff := subtitleBackoffBytes
	if s.playbackInfo != nil && s.playbackInfo.MkvMetadata != nil && len(s.playbackInfo.MkvMetadata.Cues) > 0 {
		// If cues are available, offset is precise. No backoff needed.
		backoff = 0
	}
	s.startSubtitleStreamP(stream, playbackCtx, newReader, offset, backoff, request)
}

// OnSubtitleFileUploaded adds a subtitle track, converts it to ASS if needed.
func (s *BaseStream) OnSubtitleFileUploaded(filename string, content string) {
	parser, ok := s.playbackInfo.MkvMetadataParser.Get()
	if !ok {
		s.logger.Error().Msg("directstream:A Failed to load playback info")
		return
	}

	ext := util.FileExt(filename)

	newContent := content
	if ext != ".ass" {
		var err error
		var from int
		switch ext {
		case ".ssa":
			from = mkvparser.SubtitleTypeSSA
		case ".srt":
			from = mkvparser.SubtitleTypeSRT
		case ".vtt":
			from = mkvparser.SubtitleTypeWEBVTT
		case ".ttml":
			from = mkvparser.SubtitleTypeTTML
		case ".stl":
			from = mkvparser.SubtitleTypeSTL
		case ".txt":
			from = mkvparser.SubtitleTypeUnknown
		default:
			err = errors.New("unsupported subtitle format")
		}
		s.logger.Debug().
			Str("filename", filename).
			Str("ext", ext).
			Int("detected", from).
			Msg("directstream: Converting uploaded subtitle file")
		newContent, err = mkvparser.ConvertToASS(content, from)
		if err != nil {
			s.manager.wsEventManager.SendEventTo(s.clientId, events.ErrorToast, "Failed to convert subtitle file: "+err.Error())
			return
		}
	}

	metadata := parser.GetMetadata(context.Background())
	num := int64(len(metadata.Tracks)) + 1
	subtitleNum := int64(len(metadata.SubtitleTracks))

	// e.g. filename = "title.eng.srt" -> name = "title.eng"
	name := strings.TrimSuffix(filename, ext)
	// e.g. "title.eng" -> ".eng" or "title.eng"
	name = strings.Replace(name, strings.Replace(s.filename, util.FileExt(s.filename), "", -1), "", 1) // remove the filename from the subtitle name
	name = strings.TrimSpace(name)

	// e.g. name = "title.eng" -> probableLangExt = ".eng"
	probableLangExt := util.FileExt(name)

	// if probableLangExt is not empty, use it as the language
	lang := cmp.Or(strings.TrimPrefix(probableLangExt, "."), "unknown")
	// cleanup lang
	lang = strings.ReplaceAll(lang, "-", " ")
	lang = strings.ReplaceAll(lang, "_", " ")
	lang = strings.ReplaceAll(lang, ".", " ")
	lang = strings.ReplaceAll(lang, ",", " ")
	lang = cmp.Or(lang, fmt.Sprintf("Added track %d", num+1))

	if name == "PLACEHOLDER" {
		name = fmt.Sprintf("External (#%d)", subtitleNum+1)
		lang = "und"
	}

	track := &mkvparser.TrackInfo{
		Number:       num,
		UID:          uint64(num + 900),
		Type:         mkvparser.TrackTypeSubtitle,
		CodecID:      "S_TEXT/ASS",
		Name:         name,
		Language:     lang,
		LanguageIETF: lang,
		Default:      false,
		Forced:       false,
		Enabled:      true,
		CodecPrivate: newContent,
	}

	metadata.Tracks = append(metadata.Tracks, track)
	metadata.SubtitleTracks = append(metadata.SubtitleTracks, track)

	s.logger.Debug().
		Msg("directstream: Sending subtitle file to the client")

	s.manager.playbackMu.Lock()
	target := s.manager.currentPlaybackTarget
	s.manager.playbackMu.Unlock()
	if target == PlaybackTargetVideoCore && s.manager.videoCore != nil {
		s.manager.videoCore.AddSubtitleTrack(track)
	} else {
		session, ok := s.manager.mediacoreCoordinator.GetActiveSession()
		if ok {
			format := "ass"
			cmd := player.Command{
				Type: player.CommandAddSubtitleTrack,
				Payload: &player.SubtitleTrack{
					Index:    int(subtitleNum),
					Content:  &newContent,
					Label:    name,
					Language: lang,
					Format:   &format,
				},
			}
			_ = s.manager.mediacoreCoordinator.Execute(session, cmd)
		}
	}
}
