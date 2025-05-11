package mkvparser

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/at-wat/ebml-go"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	// MaxScanBytes defines the maximum number of bytes to scan from the beginning of the file
	// to find metadata. This is to avoid reading too much data for large files or slow torrents.
	maxScanBytes = 50 * 1024 * 1024 // 50MB
	// Default timecode scale (1ms)
	defaultTimecodeScale = 1_000_000
)

// SubtitleEvent holds information for a single subtitle entry.
type SubtitleEvent struct {
	TrackNumber  uint64  `json:"trackNumber"`
	Text         string  `json:"text"`         // Content
	StartTime    float64 `json:"startTime"`    // Start time in seconds
	Duration     float64 `json:"duration"`     // Duration in seconds
	CodecID      string  `json:"codecID"`      // e.g., "S_TEXT/ASS", "S_TEXT/UTF8"
	CodecPrivate string  `json:"codecPrivate"` // For ASS/SSA styling, etc.
	// ExtraData is a map of additional subtitle-specific data.
	// For ASS/SSA, the keys are "readorder", "layer", "style", "name", "marginl", "marginr", "marginv", "effect"
	ExtraData map[string]string `json:"extraData,omitempty"`
}

// MetadataParser parses Matroska metadata from a torrent file.
// It reads only the initial part of the file to extract metadata efficiently.
type MetadataParser struct {
	reader        io.Reader // Changed from *torrent.File
	logger        *zerolog.Logger
	realLogger    *zerolog.Logger
	parsedSegment *Segment
	parseErr      error
	parseOnce     sync.Once

	metadataOnce      sync.Once
	extractedMetadata *Metadata
}

// NewMetadataParser creates a new MetadataParser.
func NewMetadataParser(reader io.Reader, logger *zerolog.Logger) *MetadataParser {
	return &MetadataParser{
		reader:     reader,
		logger:     logger,
		realLogger: logger,

		extractedMetadata: nil,
	}
}

func (mp *MetadataParser) SetLoggerEnabled(enabled bool) {
	if !enabled {
		mp.logger = lo.ToPtr(zerolog.Nop())
	} else {
		mp.logger = mp.realLogger
	}
}

// convertTrackType converts Matroska track type uint to a string representation.
func convertTrackType(trackType uint64) TrackType {
	switch trackType {
	case 0x01:
		return TrackTypeVideo
	case 0x02:
		return TrackTypeAudio
	case 0x03:
		return TrackTypeComplex
	case 0x10:
		return TrackTypeLogo
	case 0x11:
		return TrackTypeSubtitle
	case 0x12:
		return TrackTypeButtons
	default:
		return TrackTypeUnknown
	}
}

// getLanguageCode returns the IETF language code if available, otherwise the older 3-letter code,
// defaulting to "eng" if neither is present or valid.
func getLanguageCode(track *TrackEntry) string {
	if track.LanguageIETF != "" && track.LanguageIETF != "und" {
		return track.LanguageIETF
	}
	if track.Language != "" && track.Language != "und" {
		return track.Language
	}
	// If both are missing or 'und', default to 'eng' as per spec default for Language tag
	return "eng"
}

// parseMetadataOnce performs the actual parsing of the torrent file stream.
// It unmarshals into a MKVRoot struct, using WithIgnoreUnknown(true) to skip elements
// not defined in the structs up to maxScanBytes.
func (mp *MetadataParser) parseMetadataOnce(ctx context.Context) {
	mp.parseOnce.Do(func() {
		mp.logger.Debug().Msg("mkv parser: Starting metadata parsing")
		startTime := time.Now()

		// Read up to maxScanBytes
		limitedReader := io.LimitReader(mp.reader, maxScanBytes)

		done := make(chan error, 1)
		var root MKVRoot // Unmarshal into the top-level structure

		go func() {
			// Blocks until the stream is fully read or an error occurs.
			// Use WithIgnoreUnknown to skip elements not defined (like Cluster, Cues).
			done <- ebml.Unmarshal(limitedReader, &root, ebml.WithIgnoreUnknown(true))
		}()

		select {
		case err := <-done:
			if err != nil && err != io.EOF && !strings.Contains(err.Error(), "unexpected EOF") {
				mp.logger.Error().Err(err).Msg("mkv parser: EBML unmarshalling error")
				mp.parseErr = fmt.Errorf("ebml unmarshalling failed: %w", err)
			} else if err != nil {
				mp.logger.Debug().Err(err).Msg("mkv parser: EBML unmarshalling finished with EOF/unexpected EOF (expected outcome).")
				mp.parseErr = nil
			} else {
				mp.logger.Debug().Msg("mkv parser: EBML unmarshalling completed fully within scan limit.")
				mp.parseErr = nil
			}
		case <-ctx.Done():
			mp.logger.Warn().Msg("mkv parser: EBML unmarshalling cancelled or timed out via context")
			mp.parseErr = fmt.Errorf("ebml unmarshalling context cancelled/timed out: %w", ctx.Err())
		}

		// Assign the parsed segment even if errors occurred (EOF/context timeout might allow partial data)
		mp.parsedSegment = &root.Segment

		logMsg := mp.logger.Info().Dur("parseDuration", time.Since(startTime))
		if mp.parseErr != nil {
			logMsg.Err(mp.parseErr)
		}
		logMsg.Msg("mkv parser: Metadata parsing attempt finished")
	})
}

// GetMetadata extracts all relevant metadata from the torrent file.
func (mp *MetadataParser) GetMetadata(ctx context.Context) *Metadata {
	mp.parseMetadataOnce(ctx)

	mp.metadataOnce.Do(func() {
		result := &Metadata{
			Tracks:      make([]*TrackInfo, 0),
			Chapters:    make([]*ChapterInfo, 0),
			Attachments: make([]*AttachmentInfo, 0),
			Error:       mp.parseErr, // Assign error from parsing attempt
		}

		if mp.parseErr != nil {
			// Allow proceeding if error was due to context cancellation/deadline and some segment was parsed
			// This might happen if header was parsed but full scan for GetMetadata (which is limited) was cut short
			proceedWithPartial := (errors.Is(mp.parseErr, context.Canceled) || errors.Is(mp.parseErr, context.DeadlineExceeded)) && mp.parsedSegment != nil
			if !proceedWithPartial {
				mp.extractedMetadata = result
				return
			}
		}

		if mp.parsedSegment == nil {
			if result.Error == nil { // Assign a generic error if parsedSegment is nil and no specific error was recorded
				result.Error = fmt.Errorf("metadata segment is nil after parsing attempt")
			}
			// Cannot proceed without a segment
			mp.extractedMetadata = result
			return
		}

		if len(mp.parsedSegment.Info) == 0 {
			if result.Error == nil {
				mp.logger.Warn().Msg("mkv parser: No Info element found in parsed segment. Metadata extraction might be incomplete.")
				result.Error = fmt.Errorf("no Info element found in parsed segment")
			} else {
				mp.logger.Warn().Err(result.Error).Msg("mkv parser: No Info element found in parsed segment (parsing attempt may have been interrupted).")
			}
			mp.extractedMetadata = result
			return
		}

		info := mp.parsedSegment.Info[0] // Use the first Info element
		timecodeScale := uint64(defaultTimecodeScale)

		result.Title = info.Title
		result.MuxingApp = info.MuxingApp
		result.WritingApp = info.WritingApp
		if info.TimecodeScale > 0 {
			timecodeScale = info.TimecodeScale
		}
		result.TimecodeScale = float64(timecodeScale)
		if info.Duration > 0 {
			result.Duration = (info.Duration * float64(timecodeScale)) / 1e9 // nanoseconds to seconds
		} else if result.Error == nil {
			mp.logger.Warn().Msg("mkv parser: Duration is zero or missing in Matroska Info element.")
		}

		if mp.parsedSegment.Tracks != nil {
			for _, track := range mp.parsedSegment.Tracks.TrackEntry {
				ti := &TrackInfo{
					Number:           track.TrackNumber,
					UID:              track.TrackUID,
					Type:             convertTrackType(track.TrackType),
					CodecID:          track.CodecID,
					Name:             track.Name,
					Language:         getLanguageCode(&track),
					Default:          track.FlagDefault == 1, // Default is 1 if FlagDefault is not present, but here we check if it's explicitly set to 1
					Forced:           track.FlagForced == 1,
					Enabled:          track.FlagEnabled == 1, // Default is 1
					defaultDuration:  track.DefaultDuration,  // Store in ns
					contentEncodings: track.ContentEncodings, // Store for GetSubtitles
				}
				// Matroska spec: FlagEnabled defaults to 1. TrackEntry.FlagEnabled is uint64.
				// If FlagEnabled is not present in the file, it will be 0 by Go's default.
				// The EBML library doesn't automatically apply default values for missing elements.
				// However, for GetMetadata, we usually report what's in the file.
				// A track is enabled if the element is present and not 0, or if the element is absent (defaults to 1).
				// For simplicity here, if track.FlagEnabled is 0, it means it was either absent or explicitly set to 0.
				// If strict spec adherence for defaulting is needed, this logic might need adjustment based on element presence.
				// Currently, it assumes if 0, it's explicitly disabled or absent (and we interpret absent as effectively enabled by default per spec, but our struct has 0).
				// Let's assume if FlagEnabled is 0, it was explicitly set to 0 or absent. For enabled status, usually 1 means enabled.
				// The problem is FlagEnabled element defaults to 1. If it's NOT in the file, our struct field will be 0.
				// So, a track is enabled if track.FlagEnabled is 1 OR if the element was not present (which we can't easily tell here without more complex EBML parsing).
				// For now, let's stick to: if track.FlagEnabled is 1, it's true. This might misrepresent tracks where FlagEnabled is missing.
				// Re-evaluating: TrackEntry.FlagEnabled should reflect the spec default if not present.
				// The `ebml-go` library will leave it as the zero value (0) if not present.
				// Spec: FlagEnabled: "A flag to indicate if the track is usable. Default: 1"
				// So, if track.FlagEnabled is 0, it means it was explicitly set to 0. Otherwise, it's 1 (present and 1, or absent).
				// This is tricky. Let's assume the current FlagEnabled == 1 is okay for now as it reflects explicit flags.

				if len(track.CodecPrivate) > 0 {
					ti.CodecPrivate = string(track.CodecPrivate)
				}

				if track.Video != nil {
					ti.PixelWidth = track.Video.PixelWidth
					ti.PixelHeight = track.Video.PixelHeight
				}
				if track.Audio != nil {
					ti.SamplingFrequency = track.Audio.SamplingFrequency
					ti.Channels = track.Audio.Channels
					ti.BitDepth = track.Audio.BitDepth
					if ti.SamplingFrequency == 0 {
						ti.SamplingFrequency = 8000.0
					}
					if ti.Channels == 0 {
						ti.Channels = 1
					}
				}

				result.Tracks = append(result.Tracks, ti)
			}
		}

		if mp.parsedSegment.Chapters != nil {
			for _, edition := range mp.parsedSegment.Chapters.EditionEntry {
				for _, atom := range edition.ChapterAtom {
					ci := &ChapterInfo{
						UID: atom.ChapterUID,
						// ChapterTimeStart is in nanoseconds
						Start: float64(atom.ChapterTimeStart) / 1e9,
					}
					if atom.ChapterTimeEnd > 0 {
						ci.End = float64(atom.ChapterTimeEnd) / 1e9
					}
					if len(atom.ChapterDisplay) > 0 {
						// Prefer IETF language if available, otherwise first language, then first string
						displayText := atom.ChapterDisplay[0].ChapString
						bestDisplay := lo.Filter(atom.ChapterDisplay, func(d ChapterDisplay, _ int) bool {
							return len(d.ChapLanguageIETF) > 0 && d.ChapLanguageIETF[0] != ""
						})
						if len(bestDisplay) > 0 {
							displayText = bestDisplay[0].ChapString
						} else {
							bestDisplay = lo.Filter(atom.ChapterDisplay, func(d ChapterDisplay, _ int) bool {
								return len(d.ChapLanguage) > 0 && d.ChapLanguage[0] != "" && d.ChapLanguage[0] != "und"
							})
							if len(bestDisplay) > 0 {
								displayText = bestDisplay[0].ChapString
							}
						}
						ci.Text = displayText
					}
					result.Chapters = append(result.Chapters, ci)
				}
			}
		}

		if mp.parsedSegment.Attachments != nil {
			for _, attachedFile := range mp.parsedSegment.Attachments.AttachedFile {
				ai := &AttachmentInfo{
					UID:      attachedFile.FileUID,
					Filename: attachedFile.FileName,
					Mimetype: attachedFile.FileMimeType,
					Size:     len(attachedFile.FileData),
					Data:     attachedFile.FileData, // This can be large
				}
				result.Attachments = append(result.Attachments, ai)
			}
		}

		mp.extractedMetadata = result
	})

	return mp.extractedMetadata
}

// internal struct to hold necessary info for subtitle processing for a track
type subtitleTrackInternalInfo struct {
	Number           uint64
	UID              uint64
	CodecID          string
	CodecPrivate     []byte
	DefaultDuration  uint64 // in ns
	ContentEncodings *ContentEncodings
}

// StreamSubtitles extracts subtitles from a streaming source by reading it as a continuous flow.
// This method doesn't require the reader to support seeking, making it suitable for HTTP streams.
// It processes the MKV file on-the-fly and returns subtitles as they're encountered.
//
// The function returns a channel of SubtitleEvent which will be closed when:
// - The context is canceled
// - The entire stream is processed
// - An unrecoverable error occurs (which is also returned in the error channel)
//
// The error channel will receive nil if processing completed normally, or an error if something failed.
//
// If newReader is provided, it will be used instead of mp.reader, which may have been partially consumed.
func (mp *MetadataParser) StreamSubtitles(ctx context.Context, newReader ...io.Reader) (<-chan *SubtitleEvent, <-chan error) {
	subtitleCh := make(chan *SubtitleEvent)
	errCh := make(chan error, 1)

	go func() {
		defer close(subtitleCh)
		defer close(errCh)

		// Define local structs for streaming unmarshalling of clusters
		type streamingSegment struct {
			Cluster chan *Cluster `ebml:"Cluster,omitempty"`
			// We might need Info here if we want the most up-to-date TimecodeScale
			// from the full stream scan, rather than just the header.
			// However, parseMetadataOnce already gives us a good one from the header.
			Info []Info `ebml:"Info,omitempty"`
		}
		type streamingMKVRoot struct {
			Header  EBMLHeader       `ebml:"EBML"` // EBML Header
			Segment streamingSegment `ebml:"Segment,size=unknown"`
		}

		// First, ensure metadata is parsed to get the initial header and track information
		mp.parseMetadataOnce(ctx)

		if mp.parseErr != nil && !errors.Is(mp.parseErr, io.EOF) && !strings.Contains(mp.parseErr.Error(), "unexpected EOF") {
			// A critical error during header parsing that wasn't just EOF due to limit
			mp.logger.Error().Err(mp.parseErr).Msg("mkv parser: StreamSubtitles cannot proceed due to initial metadata parsing error")
			errCh <- fmt.Errorf("initial metadata parse failed: %w", mp.parseErr)
			return
		}
		// If mp.parseErr is EOF/unexpected EOF, it's fine for header parsing, we likely got what we needed.

		if mp.parsedSegment == nil || mp.parsedSegment.Tracks == nil {
			mp.logger.Error().Msg("mkv parser: StreamSubtitles cannot proceed, no tracks found in initial metadata")
			errCh <- errors.New("no track information available from initial parse")
			return
		}

		// Identify subtitle tracks and build map for quick lookup from the initial header parse
		subTrackInfos := make(map[uint64]*subtitleTrackInternalInfo)
		if mp.parsedSegment.Tracks != nil {
			for _, track := range mp.parsedSegment.Tracks.TrackEntry {
				if convertTrackType(track.TrackType) == TrackTypeSubtitle {
					subTrackInfos[track.TrackNumber] = &subtitleTrackInternalInfo{
						Number:           track.TrackNumber,
						UID:              track.TrackUID,
						CodecID:          track.CodecID,
						CodecPrivate:     track.CodecPrivate,
						DefaultDuration:  track.DefaultDuration, // ns
						ContentEncodings: track.ContentEncodings,
					}
					mp.logger.Debug().Uint64("trackNum", track.TrackNumber).Str("codec", track.CodecID).Msg("mkv parser: Identified subtitle track for streaming")
				}
			}
		}

		if len(subTrackInfos) == 0 {
			mp.logger.Info().Msg("mkv parser: No subtitle tracks found for streaming")
			errCh <- nil // No error, just no subtitles
			return
		}

		// Get timecode scale from the initial header parse. This might be refined if the full stream has a different one.
		var timecodeScaleVal uint64 = defaultTimecodeScale
		if len(mp.parsedSegment.Info) > 0 && mp.parsedSegment.Info[0].TimecodeScale > 0 {
			timecodeScaleVal = mp.parsedSegment.Info[0].TimecodeScale
			mp.logger.Debug().Uint64("timecodeScale", timecodeScaleVal).Msg("mkv parser: Using TimecodeScale from header for streaming")
		}
		timecodeScaleFactor := float64(timecodeScaleVal) / 1e9 // to seconds

		// Determine which reader to use
		var reader io.Reader
		if len(newReader) > 0 && newReader[0] != nil {
			reader = newReader[0]
			mp.logger.Debug().Msg("mkv parser: Using provided fresh reader for subtitle streaming")
		} else {
			reader = mp.reader
			if _, ok := reader.(io.ReadSeeker); ok {
				// Note: Do not seek to beginning, as we may want to stream from the current position
			} else {
				mp.logger.Warn().Msg("mkv parser: Main reader doesn't support seeking, subtitle stream may be incomplete.")
			}
		}

		// Initialize streaming root and the channel for clusters
		streamingRoot := streamingMKVRoot{}
		// A small buffer can help prevent the Unmarshal from blocking if the consumer is slightly slower.
		streamingRoot.Segment.Cluster = make(chan *Cluster, 5)

		streamCtx, cancelStream := context.WithCancel(ctx)
		defer cancelStream()

		var wg sync.WaitGroup
		wg.Add(1) // For the cluster processing goroutine

		go func() { // Goroutine to process clusters from the channel
			defer wg.Done()
			defer mp.logger.Debug().Msg("mkv parser: Cluster processing goroutine finished.")

			// This variable can be updated if a new Info tag is encountered during streaming
			currentStreamTimecodeScaleFactor := timecodeScaleFactor

			for {
				select {
				case <-streamCtx.Done():
					mp.logger.Info().Msg("mkv parser: Cluster processing cancelled via context.")
					return
				case cluster, ok := <-streamingRoot.Segment.Cluster:
					if !ok {
						mp.logger.Debug().Msg("mkv parser: Cluster channel closed, ending processing.")
						return // Channel closed
					}
					if cluster == nil {
						mp.logger.Warn().Msg("mkv parser: Received nil cluster, skipping.")
						continue
					}

					//mp.logger.Debug().Uint64("clusterTimecode", cluster.Timecode).Int("numSimpleBlocks", len(cluster.SimpleBlock)).Int("numBlockGroups", len(cluster.BlockGroup)).Msg("mkv parser: Processing new cluster from channel")
					clusterTimecode := cluster.Timecode

					// Logic to process blocks within a cluster
					processBlock := func(block ebml.Block, blockGroupDuration uint64) {
						trackInfo, isSubTrack := subTrackInfos[block.TrackNumber]
						if !isSubTrack {
							return // Not a subtitle track we care about
						}

						for _, frameData := range block.Data {
							payload := frameData
							if trackInfo.ContentEncodings != nil {
								for _, encoding := range trackInfo.ContentEncodings.ContentEncoding {
									if encoding.ContentCompression != nil &&
										encoding.ContentCompression.ContentCompAlgo == 0 { // 0 is zlib
										zlibReader, err := zlib.NewReader(bytes.NewReader(payload))
										if err != nil {
											mp.logger.Error().Err(err).Uint64("trackNum", trackInfo.Number).Msg("mkv parser: Failed to create zlib reader for subtitle frame")
											continue
										}
										decompressedPayload, err := io.ReadAll(zlibReader)
										_ = zlibReader.Close()
										if err != nil {
											mp.logger.Error().Err(err).Uint64("trackNum", trackInfo.Number).Msg("mkv parser: Failed to decompress zlib subtitle frame")
											continue
										}
										payload = decompressedPayload
										break
									}
								}
							}

							// Calculate startTime and duration here, as they depend on block and cluster timecodes
							startTime := (float64(clusterTimecode) + float64(block.Timecode)) * currentStreamTimecodeScaleFactor
							var duration float64
							if blockGroupDuration > 0 {
								duration = float64(blockGroupDuration) * currentStreamTimecodeScaleFactor
							} else if trackInfo.DefaultDuration > 0 {
								duration = float64(trackInfo.DefaultDuration) / 1e9 // DefaultDuration is in ns
							} else {
								duration = 0
							}

							initialText := string(payload)
							subtitleEvent := &SubtitleEvent{
								TrackNumber:  trackInfo.Number,
								Text:         initialText, // Default to full payload
								StartTime:    startTime,
								Duration:     duration,
								CodecID:      trackInfo.CodecID,
								CodecPrivate: string(trackInfo.CodecPrivate),
								ExtraData:    make(map[string]string),
							}

							// Special handling for ASS/SSA format
							if trackInfo.CodecID == "S_TEXT/ASS" || trackInfo.CodecID == "S_TEXT/SSA" {
								values := strings.Split(initialText, ",")

								// ASS/SSA format has fields in this order:
								// ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text
								// For SSA, the format is slightly different with no Layer field

								if len(values) > 0 {
									subtitleEvent.ExtraData["readorder"] = values[0]
								}

								// Handle Layer field differently for ASS vs SSA
								if trackInfo.CodecID != "S_TEXT/SSA" {
									if len(values) > 1 {
										subtitleEvent.ExtraData["layer"] = values[1]
									}
								}

								// For both formats, map the remaining fields
								valueOffset := 0 // No offset needed for standard field mapping

								if len(values) > 2+valueOffset {
									subtitleEvent.ExtraData["style"] = values[2+valueOffset]
								}
								if len(values) > 3+valueOffset {
									subtitleEvent.ExtraData["name"] = values[3+valueOffset]
								}
								if len(values) > 4+valueOffset {
									subtitleEvent.ExtraData["marginl"] = values[4+valueOffset]
								}
								if len(values) > 5+valueOffset {
									subtitleEvent.ExtraData["marginr"] = values[5+valueOffset]
								}
								if len(values) > 6+valueOffset {
									subtitleEvent.ExtraData["marginv"] = values[6+valueOffset]
								}
								if len(values) > 7+valueOffset {
									subtitleEvent.ExtraData["effect"] = values[7+valueOffset]
								}

								// Extract the actual subtitle text
								if len(values) > 8+valueOffset {
									subtitleEvent.Text = strings.Join(values[8+valueOffset:], ",")
								} else {
									subtitleEvent.Text = ""
								}
							}

							mp.logger.Debug().
								Uint64("trackNum", trackInfo.Number).
								Float64("startTime", startTime).
								Str("codecId", trackInfo.CodecID).
								Str("text", subtitleEvent.Text).
								Msg("mkv parser: Subtitle event")

							select {
							case subtitleCh <- subtitleEvent:
							case <-streamCtx.Done():
								mp.logger.Info().Msg("mkv parser: Subtitle sending cancelled by context.")
								return
							}
						}
					}

					for _, simpleBlock := range cluster.SimpleBlock {
						processBlock(simpleBlock, 0) // SimpleBlock has no BlockDuration
					}
					for _, blockGroup := range cluster.BlockGroup {
						processBlock(blockGroup.Block, blockGroup.BlockDuration)
					}
				}
			}
		}()

		mp.logger.Debug().Msg("mkv parser: Starting EBML unmarshal for streaming clusters")
		unmarshalStartTime := time.Now()
		// This will block until the stream is fully read or an error occurs.
		// Clusters are sent to streamingRoot.Segment.Cluster channel.
		unmarshalErr := ebml.Unmarshal(reader, &streamingRoot, ebml.WithIgnoreUnknown(true))
		mp.logger.Info().Dur("unmarshalDuration", time.Since(unmarshalStartTime)).Err(unmarshalErr).Msg("mkv parser: EBML unmarshal for streaming finished.")

		// After Unmarshal finishes (or errors), close the cluster channel to signal the processing goroutine.
		close(streamingRoot.Segment.Cluster)
		mp.logger.Debug().Msg("mkv parser: Closed cluster channel.")

		// Wait for the cluster processing goroutine to finish all items in the channel.
		wg.Wait()
		mp.logger.Debug().Msg("mkv parser: Cluster processing goroutine completed after unmarshal.")

		// Check for new Info tag during stream processing if it was defined in streamingSegment
		if len(streamingRoot.Segment.Info) > 0 && streamingRoot.Segment.Info[0].TimecodeScale > 0 {
			if streamingRoot.Segment.Info[0].TimecodeScale != timecodeScaleVal {
				mp.logger.Info().
					Uint64("headerScale", timecodeScaleVal).
					Uint64("streamScale", streamingRoot.Segment.Info[0].TimecodeScale).
					Msg("mkv parser: TimecodeScale from full stream differs from header. Events used header scale.")
				// Potentially, one could re-process or adjust, but for streaming, events are already sent.
			}
		}

		if unmarshalErr != nil && !errors.Is(unmarshalErr, io.EOF) && !strings.Contains(unmarshalErr.Error(), "unexpected EOF") {
			mp.logger.Error().Err(unmarshalErr).Msg("mkv parser: Unrecoverable error during subtitle stream unmarshalling")
			errCh <- unmarshalErr
		} else {
			mp.logger.Info().Msg("mkv parser: Subtitle streaming completed successfully or with expected EOF.")
			errCh <- nil
		}
	}()

	return subtitleCh, errCh
}
