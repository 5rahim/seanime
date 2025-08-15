package directstream

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"seanime/internal/events"
	"seanime/internal/mkvparser"
	"seanime/internal/util"
	"strings"
	"sync"
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
	completed bool // ran until the EOF

	cleanupFunc func()
	stopOnce    sync.Once
}

func (s *SubtitleStream) Stop(completed bool) {
	s.stopOnce.Do(func() {
		s.logger.Debug().Int64("offset", s.offset).Msg("directstream: Stopping subtitle stream")
		s.completed = completed
		s.cleanupFunc()
	})
}

// StartSubtitleStreamP starts a subtitle stream for the given stream at the given offset with a specified backoff bytes.
func (s *BaseStream) StartSubtitleStreamP(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64, backoffBytes int64) {
	mkvMetadataParser, ok := s.playbackInfo.MkvMetadataParser.Get()
	if !ok {
		return
	}

	s.logger.Trace().Int64("offset", offset).Msg("directstream: Starting new subtitle stream")
	subtitleStream := &SubtitleStream{
		stream: stream,
		logger: s.logger,
		parser: mkvMetadataParser,
		reader: newReader,
		offset: offset,
	}

	// Check if we have a completed subtitle stream for this offset
	shouldContinue := true
	s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
		// If a stream is completed and its offset comes before this one, we don't need to start a new stream
		// |------------------------------->| other stream
		//                    |               this stream
		//                   ^^^ starting in an area the other stream has already completed
		if offset > 0 && value.offset <= offset && value.completed {
			shouldContinue = false
			return false
		}
		return true
	})

	if !shouldContinue {
		s.logger.Debug().Int64("offset", offset).Msg("directstream: Skipping subtitle stream, range already fulfilled")
		return
	}

	ctx, subtitleCtxCancel := context.WithCancel(playbackCtx)
	subtitleStream.cleanupFunc = subtitleCtxCancel

	subtitleStreamId := uuid.New().String()
	s.activeSubtitleStreams.Set(subtitleStreamId, subtitleStream)

	subtitleCh, errCh, _ := subtitleStream.parser.ExtractSubtitles(ctx, newReader, offset, backoffBytes)

	firstEventSentCh := make(chan struct{})
	closeFirstEventSentOnce := sync.Once{}

	onFirstEventSent := func() {
		closeFirstEventSentOnce.Do(func() {
			s.logger.Debug().Int64("offset", offset).Msg("directstream: First subtitle event sent")
			close(firstEventSentCh) // Notify that the first subtitle event has been sent
		})
	}

	var lastSubtitleEvent *mkvparser.SubtitleEvent
	lastSubtitleEventRWMutex := sync.RWMutex{}

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
				if lastSubtitleEvent == nil {
					continue
				}
				shouldEnd := false
				lastSubtitleEventRWMutex.RLock()
				s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
					if key != subtitleStreamId {
						// If the other stream is ahead of this stream
						// and the last subtitle event is after the other stream's offset
						// |--------------->                   this stream
						//                     |-------------> other stream
						//                    ^^^ stop this stream where it reached the tail of the other stream
						if offset > 0 && offset < value.offset && lastSubtitleEvent.HeadPos >= value.offset {
							shouldEnd = true
						}
					}
					return true
				})
				lastSubtitleEventRWMutex.RUnlock()
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
			subtitleStream.cleanupFunc()
		}()

		// Keep track if channels are active to manage loop termination
		subtitleChannelActive := true
		errorChannelActive := true

		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
			select {
			case <-ctx.Done():
				s.logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle streaming cancelled by context")
				return

			case subtitle, ok := <-subtitleCh:
				if !ok {
					subtitleCh = nil // Mark as exhausted
					subtitleChannelActive = false
					if !errorChannelActive { // If both channels are exhausted, exit
						return
					}
					continue // Continue to wait for errorChannel or ctx.Done()
				}
				if subtitle != nil {
					onFirstEventSent()
					s.manager.nativePlayer.SubtitleEvent(stream.ClientId(), subtitle)
					lastSubtitleEventRWMutex.Lock()
					lastSubtitleEvent = subtitle
					lastSubtitleEventRWMutex.Unlock()
				}

			case err, ok := <-errCh:
				if !ok {
					errCh = nil // Mark as exhausted
					errorChannelActive = false
					if !subtitleChannelActive { // If both channels are exhausted, exit
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
	// use 1MB as the cluster padding for subtitle streams
	s.StartSubtitleStreamP(stream, playbackCtx, newReader, offset, 1024*1024)
}

//// StartSubtitleStream is similar to BaseStream.StartSubtitleStream, but rate limits the requests to the external debrid server.
//// - There will only be one subtitle stream at a time.
//func (s *DebridStream) StartSubtitleStream(stream Stream, playbackCtx context.Context, newReader io.ReadSeekCloser, offset int64, end int64) {
//	mkvMetadataParser, ok := s.playbackInfo.MkvMetadataParser.Get()
//	if !ok {
//		return
//	}
//
//	s.logger.Trace().Int64("offset", offset).Msg("directstream(debrid): Starting new subtitle stream")
//	subtitleStream := &SubtitleStream{
//		stream: stream,
//		logger: s.logger,
//		parser: mkvMetadataParser,
//		reader: newReader,
//		offset: offset,
//	}
//
//	s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
//		value.Stop(true)
//		return true
//	})
//
//	ctx, subtitleCtxCancel := context.WithCancel(playbackCtx)
//	subtitleStream.cleanupFunc = subtitleCtxCancel
//
//	subtitleStreamId := uuid.New().String()
//	s.activeSubtitleStreams.Set(subtitleStreamId, subtitleStream)
//
//	subtitleCh, errCh, _ := subtitleStream.parser.ExtractSubtitles(ctx, newReader, offset)
//
//	firstEventSentCh := make(chan struct{})
//	closeFirstEventSentOnce := sync.Once{}
//
//	onFirstEventSent := func() {
//		closeFirstEventSentOnce.Do(func() {
//			s.logger.Debug().Int64("offset", offset).Msg("directstream: First subtitle event sent")
//			close(firstEventSentCh) // Notify that the first subtitle event has been sent
//		})
//	}
//
//	var lastSubtitleEvent *mkvparser.SubtitleEvent
//	lastSubtitleEventRWMutex := sync.RWMutex{}
//
//	// Check every second if we need to end this stream
//	go func() {
//		ticker := time.NewTicker(1 * time.Second)
//		defer ticker.Stop()
//		for {
//			select {
//			case <-ctx.Done():
//				subtitleStream.Stop(false)
//				return
//			case <-ticker.C:
//				if lastSubtitleEvent == nil {
//					continue
//				}
//				shouldEnd := false
//				lastSubtitleEventRWMutex.RLock()
//				s.activeSubtitleStreams.Range(func(key string, value *SubtitleStream) bool {
//					if key != subtitleStreamId {
//						// If the other stream is ahead of this stream
//						// and the last subtitle event is after the other stream's offset
//						// |--------------->                   this stream
//						//                     |-------------> other stream
//						//                    ^^^ stop this stream where it reached the tail of the other stream
//						if offset > 0 && offset < value.offset && lastSubtitleEvent.HeadPos >= value.offset {
//							shouldEnd = true
//						}
//					}
//					return true
//				})
//				lastSubtitleEventRWMutex.RUnlock()
//				if shouldEnd {
//					subtitleStream.Stop(false)
//					return
//				}
//			}
//		}
//	}()
//
//	go func() {
//		defer func(reader io.ReadSeekCloser) {
//			_ = reader.Close()
//			s.logger.Trace().Int64("offset", offset).Msg("directstream: Closing subtitle stream goroutine")
//		}(newReader)
//		defer func() {
//			onFirstEventSent()
//			subtitleStream.cleanupFunc()
//		}()
//
//		// Keep track if channels are active to manage loop termination
//		subtitleChannelActive := true
//		errorChannelActive := true
//
//		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
//			select {
//			case <-ctx.Done():
//				s.logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle streaming cancelled by context")
//				return
//
//			case subtitle, ok := <-subtitleCh:
//				if !ok {
//					subtitleCh = nil // Mark as exhausted
//					subtitleChannelActive = false
//					if !errorChannelActive { // If both channels are exhausted, exit
//						return
//					}
//					continue // Continue to wait for errorChannel or ctx.Done()
//				}
//				if subtitle != nil {
//					onFirstEventSent()
//					s.manager.nativePlayer.SubtitleEvent(stream.ClientId(), subtitle)
//					lastSubtitleEventRWMutex.Lock()
//					lastSubtitleEvent = subtitle
//					lastSubtitleEventRWMutex.Unlock()
//				}
//
//			case err, ok := <-errCh:
//				if !ok {
//					errCh = nil // Mark as exhausted
//					errorChannelActive = false
//					if !subtitleChannelActive { // If both channels are exhausted, exit
//						return
//					}
//					continue // Continue to wait for subtitleChannel or ctx.Done()
//				}
//				// A value (error or nil) was received from errCh.
//				// This is the terminal signal from the mkvparser's subtitle streaming process.
//				if err != nil {
//					s.logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Error streaming subtitles")
//				} else {
//					s.logger.Info().Int64("offset", offset).Msg("directstream: Subtitle streaming completed by parser.")
//					subtitleStream.Stop(true)
//				}
//				return // Terminate goroutine
//			}
//		}
//	}()
//}

//// streamSubtitles starts the subtitle stream.
//// It will stream the subtitles from all tracks to the client. The client should load the subtitles in an array.
//func (m *Manager) streamSubtitles(ctx context.Context, stream Stream, parser *mkvparser.MetadataParser, newReader io.ReadSeekCloser, offset int64, cleanupFunc func()) (firstEventSentCh chan struct{}) {
//	m.Logger.Debug().Int64("offset", offset).Str("clientId", stream.ClientId()).Msg("directstream: Starting subtitle extraction")
//
//	subtitleCh, errCh, _ := parser.ExtractSubtitles(ctx, newReader, offset)
//
//	firstEventSentCh = make(chan struct{})
//	closeFirstEventSentOnce := sync.Once{}
//
//	onFirstEventSent := func() {
//		closeFirstEventSentOnce.Do(func() {
//			m.Logger.Debug().Int64("offset", offset).Msg("directstream: First subtitle event sent")
//			close(firstEventSentCh) // Notify that the first subtitle event has been sent
//		})
//	}
//
//	go func() {
//		defer func(reader io.ReadSeekCloser) {
//			_ = reader.Close()
//			m.Logger.Trace().Int64("offset", offset).Msg("directstream: Closing subtitle stream goroutine")
//		}(newReader)
//		defer func() {
//			onFirstEventSent()
//			if cleanupFunc != nil {
//				cleanupFunc()
//			}
//		}()
//
//		// Keep track if channels are active to manage loop termination
//		subtitleChannelActive := true
//		errorChannelActive := true
//
//		for subtitleChannelActive || errorChannelActive { // Loop as long as at least one channel might still produce data or a final status
//			select {
//			case <-ctx.Done():
//				m.Logger.Debug().Int64("offset", offset).Msg("directstream: Subtitle streaming cancelled by context")
//				return
//
//			case subtitle, ok := <-subtitleCh:
//				if !ok {
//					subtitleCh = nil // Mark as exhausted
//					subtitleChannelActive = false
//					if !errorChannelActive { // If both channels are exhausted, exit
//						return
//					}
//					continue // Continue to wait for errorChannel or ctx.Done()
//				}
//				if subtitle != nil {
//					onFirstEventSent()
//					m.nativePlayer.SubtitleEvent(stream.ClientId(), subtitle)
//				}
//
//			case err, ok := <-errCh:
//				if !ok {
//					errCh = nil // Mark as exhausted
//					errorChannelActive = false
//					if !subtitleChannelActive { // If both channels are exhausted, exit
//						return
//					}
//					continue // Continue to wait for subtitleChannel or ctx.Done()
//				}
//				// A value (error or nil) was received from errCh.
//				// This is the terminal signal from the mkvparser's subtitle streaming process.
//				if err != nil {
//					m.Logger.Warn().Err(err).Int64("offset", offset).Msg("directstream: Error streaming subtitles")
//				} else {
//					m.Logger.Info().Int64("offset", offset).Msg("directstream: Subtitle streaming completed by parser.")
//				}
//				return // Terminate goroutine
//			}
//		}
//	}()
//
//	return
//}

// OnSubtitleFileUploaded adds a subtitle track, converts it to ASS if needed.
func (s *BaseStream) OnSubtitleFileUploaded(filename string, content string) {
	parser, ok := s.playbackInfo.MkvMetadataParser.Get()
	if !ok {
		s.logger.Error().Msg("directstream:A Failed to load playback info")
		return
	}

	ext := util.FileExt(filename)

	newContent := content
	if ext != ".ass" && ext != ".ssa" {
		var err error
		var from int
		switch ext {
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
	lang := cmp.Or(strings.TrimPrefix(probableLangExt, "."), name)
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
		UID:          num + 900,
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

	s.manager.nativePlayer.AddSubtitleTrack(s.clientId, track)
}
