package mkvparser

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"image/png"
	"io"
	"path/filepath"
	"seanime/internal/matroska"
	"seanime/internal/pgs"
	"seanime/internal/util"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	defaultTimecodeScale   = 1_000_000 // 1ms
	clusterSearchChunkSize = 8192      // 8KB
)

var matroskaClusterID = []byte{0x1F, 0x43, 0xB6, 0x75}
var subtitleExtensions = map[string]struct{}{".ass": {}, ".ssa": {}, ".srt": {}, ".vtt": {}, ".txt": {}}
var fontExtensions = map[string]struct{}{".ttf": {}, ".ttc": {}, ".woff": {}, ".woff2": {}, ".bdf": {}, ".otf": {}, ".cff": {}, ".otc": {}, ".pfa": {}, ".pfb": {}, ".pcf": {}, ".pfr": {}, ".fnt": {}, ".eot": {}}

// SubtitleEvent holds information for a single subtitle entry.
type SubtitleEvent struct {
	TrackNumber uint64            `json:"trackNumber"`
	Text        string            `json:"text"`      // Content
	StartTime   float64           `json:"startTime"` // Start time in seconds
	Duration    float64           `json:"duration"`  // Duration in seconds
	CodecID     string            `json:"codecID"`   // e.g., "S_TEXT/ASS", "S_TEXT/UTF8"
	ExtraData   map[string]string `json:"extraData,omitempty"`

	HeadPos int64 `json:"-"` // Position in the stream
}

// GetSubtitleEventKey stringifies the subtitle event to serve as a key
func GetSubtitleEventKey(se *SubtitleEvent) string {
	marshaled, err := json.Marshal(se)
	if err != nil {
		return ""
	}
	return string(marshaled)
}

// MetadataParser parses Matroska metadata from a file.
type MetadataParser struct {
	reader       io.ReadSeeker
	logger       *zerolog.Logger
	realLogger   *zerolog.Logger
	parseErr     error
	parseOnce    sync.Once
	metadataOnce sync.Once

	// Internal state for parsing
	demuxer *matroska.Demuxer

	// Result
	extractedMetadata *Metadata
}

// NewMetadataParser creates a new MetadataParser.
func NewMetadataParser(reader io.ReadSeeker, logger *zerolog.Logger) *MetadataParser {
	return &MetadataParser{
		reader:            reader,
		logger:            logger,
		realLogger:        logger,
		extractedMetadata: nil,
	}
}

func (mp *MetadataParser) SetLoggerEnabled(enabled bool) {
	if !enabled {
		mp.logger = new(zerolog.Nop())
	} else {
		mp.logger = mp.realLogger
	}
}

// convertTrackType converts Matroska track type uint to a string representation.
func convertTrackType(trackType uint8) TrackType {
	switch trackType {
	case matroska.TypeVideo:
		return TrackTypeVideo
	case matroska.TypeAudio:
		return TrackTypeAudio
	case matroska.TypeSubtitle:
		return TrackTypeSubtitle
	default:
		return TrackTypeUnknown
	}
}

func getLanguageCode(track *TrackInfo) string {
	if track.LanguageIETF != "" {
		return track.LanguageIETF
	}
	if track.Language != "" && track.Language != "und" {
		return track.Language
	}
	return "eng"
}

func getSubtitleTrackType(codecID string) string {
	switch codecID {
	case "S_TEXT/ASS":
		return "ASS"
	case "S_TEXT/SSA":
		return "SSA"
	case "S_TEXT/UTF8":
		return "TEXT"
	case "S_HDMV/PGS":
		return "PGS"
	}
	return "unknown"
}

// parseMetadataOnce performs the actual parsing of the file stream.
func (mp *MetadataParser) parseMetadataOnce(ctx context.Context) {
	mp.parseOnce.Do(func() {
		mp.logger.Debug().Msg("mkvparser: Starting metadata parsing")

		_, _ = mp.reader.Seek(0, io.SeekStart)

		// Create demuxer
		demuxer, err := matroska.NewDemuxer(mp.reader,
			matroska.IDSegmentInfo,
			matroska.IDSegment,
			matroska.IDTracks,
			matroska.IDChapters,
			matroska.IDAttachments,
		)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to create demuxer")
			mp.parseErr = fmt.Errorf("mkvparser: Failed to create demuxer: %w", err)
			return
		}

		mp.demuxer = demuxer
		mp.logger.Debug().Msg("mkvparser: Metadata parsing completed")
	})
}

// GetMetadata extracts all relevant metadata from the file.
func (mp *MetadataParser) GetMetadata(ctx context.Context) *Metadata {
	mp.parseMetadataOnce(ctx)

	mp.metadataOnce.Do(func() {
		if mp.parseErr != nil {
			mp.extractedMetadata = &Metadata{
				Error: mp.parseErr,
			}
			return
		}

		result := &Metadata{
			VideoTracks:    make([]*TrackInfo, 0),
			AudioTracks:    make([]*TrackInfo, 0),
			SubtitleTracks: make([]*TrackInfo, 0),
			Tracks:         make([]*TrackInfo, 0),
			Chapters:       make([]*ChapterInfo, 0),
			Attachments:    make([]*AttachmentInfo, 0),
		}

		// Get file info
		fileInfo, err := mp.demuxer.GetFileInfo()
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to get file info")
			result.Error = err
			mp.extractedMetadata = result
			return
		}

		result.Title = fileInfo.Title
		result.MuxingApp = fileInfo.MuxingApp
		result.WritingApp = fileInfo.WritingApp
		result.TimecodeScale = float64(fileInfo.TimecodeScale)
		if fileInfo.Duration > 0 {
			// Duration in matroska-go is in nanoseconds, convert to seconds
			result.Duration = float64(fileInfo.Duration) / 1e9
		}

		// Get tracks
		numTracks, err := mp.demuxer.GetNumTracks()
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to get number of tracks")
			result.Error = err
			mp.extractedMetadata = result
			return
		}

		for i := uint(0); i < numTracks; i++ {
			trackInfo, err := mp.demuxer.GetTrackInfo(i)
			if err != nil {
				mp.logger.Error().Err(err).Uint("track", i).Msg("mkvparser: Failed to get track info")
				continue
			}

			track := &TrackInfo{
				Number:       int64(trackInfo.Number),
				UID:          trackInfo.UID,
				Type:         convertTrackType(trackInfo.Type),
				CodecID:      trackInfo.CodecID,
				Name:         trackInfo.Name,
				Language:     trackInfo.Language,
				LanguageIETF: trackInfo.LanguageIETF,
				Default:      trackInfo.Default,
				Forced:       trackInfo.Forced,
				Enabled:      trackInfo.Enabled,
				CodecPrivate: string(trackInfo.CodecPrivate),
			}

			// Convert video info
			if trackInfo.Type == matroska.TypeVideo {
				track.Video = &VideoTrack{
					PixelWidth:  uint64(trackInfo.Video.PixelWidth),
					PixelHeight: uint64(trackInfo.Video.PixelHeight),
				}
			}

			// Convert audio info
			if trackInfo.Type == matroska.TypeAudio {
				track.Audio = &AudioTrack{
					SamplingFrequency: trackInfo.Audio.SamplingFreq,
					Channels:          uint64(trackInfo.Audio.Channels),
					BitDepth:          uint64(trackInfo.Audio.BitDepth),
				}
			}

			// Store compression info
			if trackInfo.CompEnabled {
				track.contentEncodings = &ContentEncodings{
					ContentEncoding: []ContentEncoding{
						{
							ContentCompression: &ContentCompression{
								ContentCompAlgo: uint64(trackInfo.CompMethod),
							},
						},
					},
				}
			}

			track.defaultDuration = trackInfo.DefaultDuration

			result.Tracks = append(result.Tracks, track)

			switch track.Type {
			case TrackTypeVideo:
				result.VideoTracks = append(result.VideoTracks, track)
			case TrackTypeAudio:
				result.AudioTracks = append(result.AudioTracks, track)
			case TrackTypeSubtitle:
				// Fix missing fields
				track.Name = lo.If(track.Name != "", track.Name).Else(strings.ToUpper(lo.If(track.Language != "", track.Language).Else(track.LanguageIETF)))
				track.Language = getLanguageCode(track)
				result.SubtitleTracks = append(result.SubtitleTracks, track)
			}
		}

		// Get chapters
		chapters := mp.demuxer.GetChapters()
		for _, chapter := range chapters {
			chapterInfo := &ChapterInfo{
				UID:   chapter.UID,
				Start: float64(chapter.Start) / 1e9, // Convert nanoseconds to seconds
				End:   float64(chapter.End) / 1e9,
			}

			if len(chapter.Display) > 0 {
				chapterInfo.Text = chapter.Display[0].String
				for _, display := range chapter.Display {
					if display.Language != "" {
						chapterInfo.Languages = append(chapterInfo.Languages, display.Language)
					}
				}
			}

			result.Chapters = append(result.Chapters, chapterInfo)
		}

		// Get attachments
		attachments := mp.demuxer.GetAttachments()
		for _, attachment := range attachments {
			attachmentInfo := &AttachmentInfo{
				UID:         attachment.UID,
				Filename:    attachment.Name,
				Mimetype:    attachment.MimeType,
				Size:        int(attachment.Length),
				Description: attachment.Description,
				Data:        attachment.Data,
			}

			//// Extract attachment data from file
			//data, err := mp.extractAttachmentData(attachment.Position, attachment.Length)
			//if err != nil {
			//	mp.logger.Error().Err(err).Str("filename", attachment.Name).Msg("mkvparser: Failed to extract attachment data")
			//} else {
			//	attachmentInfo.Data = data
			//	attachmentInfo.Size = len(data)
			//}

			// Determine attachment type
			fileExt := strings.ToLower(filepath.Ext(attachment.Name))
			if _, ok := fontExtensions[fileExt]; ok {
				attachmentInfo.Type = AttachmentTypeFont
			} else if _, ok := subtitleExtensions[fileExt]; ok {
				attachmentInfo.Type = AttachmentTypeSubtitle
			} else {
				attachmentInfo.Type = AttachmentTypeOther
			}

			result.Attachments = append(result.Attachments, attachmentInfo)
		}

		mp.logger.Debug().
			Int("tracks", len(result.Tracks)).
			Int("chapters", len(result.Chapters)).
			Int("attachments", len(result.Attachments)).
			Msg("mkvparser: Metadata parsing complete")

		// Generate MimeCodec string
		result.MimeCodec = mp.generateMimeCodec(result)

		mp.extractedMetadata = result
	})

	return mp.extractedMetadata
}

// extractAttachmentData reads attachment data from the file at the given position
//func (mp *MetadataParser) extractAttachmentData(position uint64, length uint64) ([]byte, error) {
//	// Seek to attachment position
//	_, err := mp.reader.Seek(int64(position), io.SeekStart)
//	if err != nil {
//		return nil, fmt.Errorf("failed to seek to attachment: %w", err)
//	}
//
//	// Read attachment data
//	data := make([]byte, length)
//	n, err := io.ReadFull(mp.reader, data)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read attachment data: %w", err)
//	}
//
//	if n != int(length) {
//		return nil, fmt.Errorf("attachment data size mismatch: expected %d, got %d", length, n)
//	}
//
//	// Try to decompress if it's zlib compressed
//	if len(data) > 2 && data[0] == 0x78 { // zlib magic number
//		zlibReader, err := zlib.NewReader(bytes.NewReader(data))
//		if err == nil {
//			decompressed, err := io.ReadAll(zlibReader)
//			_ = zlibReader.Close()
//			if err == nil {
//				return decompressed, nil
//			}
//		}
//	}
//
//	return data, nil
//}

// generateMimeCodec generates RFC 6381 codec string
func (mp *MetadataParser) generateMimeCodec(metadata *Metadata) string {
	var codecStrings []string
	seenCodecs := make(map[string]bool)

	if len(metadata.VideoTracks) > 0 {
		firstVideoTrack := metadata.VideoTracks[0]
		var videoCodecStr string
		switch firstVideoTrack.CodecID {
		case "V_MPEGH/ISO/HEVC":
			videoCodecStr = "hvc1"
		case "V_MPEG4/ISO/AVC":
			videoCodecStr = "avc1"
		case "V_AV1":
			videoCodecStr = "av01"
		case "V_VP9":
			videoCodecStr = "vp09"
		case "V_VP8":
			videoCodecStr = "vp8"
		default:
			if firstVideoTrack.CodecID != "" {
				videoCodecStr = strings.ToLower(strings.ReplaceAll(firstVideoTrack.CodecID, "/", "."))
			}
		}
		if videoCodecStr != "" && !seenCodecs[videoCodecStr] {
			codecStrings = append(codecStrings, videoCodecStr)
			seenCodecs[videoCodecStr] = true
		}
	}

	for _, audioTrack := range metadata.AudioTracks {
		var audioCodecStr string
		switch audioTrack.CodecID {
		case "A_AAC":
			audioCodecStr = "mp4a.40.2"
		case "A_AC3":
			audioCodecStr = "ac-3"
		case "A_EAC3":
			audioCodecStr = "ec-3"
		case "A_OPUS":
			audioCodecStr = "opus"
		case "A_DTS":
			audioCodecStr = "dts"
		case "A_FLAC":
			audioCodecStr = "flac"
		case "A_TRUEHD":
			audioCodecStr = "mlp"
		default:
			if audioTrack.CodecID != "" {
				audioCodecStr = strings.ToLower(strings.ReplaceAll(audioTrack.CodecID, "/", "."))
			}
		}
		if audioCodecStr != "" && !seenCodecs[audioCodecStr] {
			codecStrings = append(codecStrings, audioCodecStr)
			seenCodecs[audioCodecStr] = true
		}
	}

	ret := ""
	if len(codecStrings) > 0 {
		ret = fmt.Sprintf("video/x-matroska; codecs=\"%s\"", strings.Join(codecStrings, ", "))
	} else {
		ret = "video/x-matroska"
	}

	return ret
}

// ExtractSubtitles extracts subtitles from a streaming source by reading it as a continuous flow.
func (mp *MetadataParser) ExtractSubtitles(ctx context.Context, newReader io.ReadSeekCloser, offset int64, backoffBytes int64) (<-chan *SubtitleEvent, <-chan error, <-chan struct{}) {
	subtitleCh := make(chan *SubtitleEvent)
	errCh := make(chan error, 1)
	startedCh := make(chan struct{})

	var closeOnce sync.Once
	closeChannels := func(err error) {
		closeOnce.Do(func() {
			select {
			case errCh <- err:
			default:
			}
			close(subtitleCh)
			close(errCh)
		})
	}

	extractCtx, cancel := context.WithCancel(ctx)

	if offset > 0 {
		mp.logger.Debug().Int64("offset", offset).Msg("mkvparser: Attempting to find cluster near offset")

		clusterSeekOffset, err := findNextClusterOffset(newReader, offset, backoffBytes)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				mp.logger.Error().Err(err).Msg("mkvparser: Failed to seek to offset for subtitle extraction")
			}
			cancel()
			closeChannels(err)
			return subtitleCh, errCh, startedCh
		}

		close(startedCh)

		mp.logger.Debug().Int64("clusterSeekOffset", clusterSeekOffset).Msg("mkvparser: Found cluster near offset")

		_, err = newReader.Seek(clusterSeekOffset, io.SeekStart)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to seek to cluster offset")
			cancel()
			closeChannels(err)
			return subtitleCh, errCh, startedCh
		}
	} else {
		close(startedCh)
	}

	go func() {
		defer util.HandlePanicInModuleThen("mkvparser2/ExtractSubtitles", func() {
			closeChannels(fmt.Errorf("subtitle extraction goroutine panic"))
		})
		defer cancel()
		defer mp.logger.Trace().Msgf("mkvparser: Subtitle extraction goroutine finished.")

		sampler := new(mp.logger.Sample(&zerolog.BasicSampler{N: 500}))

		// First, ensure metadata is parsed to get track information
		mp.parseMetadataOnce(extractCtx)

		if mp.parseErr != nil && !errors.Is(mp.parseErr, io.EOF) {
			mp.logger.Error().Err(mp.parseErr).Msg("mkvparser: ExtractSubtitles cannot proceed due to initial metadata parsing error")
			closeChannels(fmt.Errorf("initial metadata parse failed: %w", mp.parseErr))
			return
		}

		// Get metadata to know subtitle tracks
		metadata := mp.GetMetadata(extractCtx)
		if metadata.Error != nil {
			mp.logger.Error().Err(metadata.Error).Msg("mkvparser: Failed to get metadata for subtitle extraction")
			closeChannels(metadata.Error)
			return
		}

		// Create a map of subtitle tracks for quick lookup
		subtitleTracks := make(map[uint8]*TrackInfo)
		for _, track := range metadata.SubtitleTracks {
			subtitleTracks[uint8(track.Number)] = track
		}

		if len(subtitleTracks) == 0 {
			mp.logger.Info().Msg("mkvparser: No subtitle tracks found for streaming")
			closeChannels(nil)
			return
		}

		// Create a new demuxer for reading packets
		demuxer, err := matroska.NewDemuxer(newReader, matroska.IDSegmentInfo)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to create streaming demuxer")
			closeChannels(err)
			return
		}
		defer demuxer.Close()

		lastSubtitleEvents := make(map[uint8]*SubtitleEvent)
		pgsDecoders := make(map[uint8]*pgs.PgsDecoder)

		// Read packets and extract subtitles
		for {
			select {
			case <-extractCtx.Done():
				mp.logger.Debug().Msg("mkvparser: Subtitle extraction cancelled by context")
				closeChannels(nil)
				return
			default:
			}

			packet, err := demuxer.ReadPacket()
			if err != nil {
				if errors.Is(err, io.EOF) {
					mp.logger.Debug().Msg("mkvparser: Reached end of stream")
					closeChannels(nil)
					return
				}
				mp.logger.Error().Err(err).Msg("mkvparser: Error reading packet")
				closeChannels(err)
				return
			}

			track, isSubtitle := subtitleTracks[packet.Track]
			if !isSubtitle {
				continue
			}

			// Skip unknown subtitle types
			if getSubtitleTrackType(track.CodecID) == "unknown" {
				continue
			}

			// Process subtitle packet
			subtitleData := packet.Data

			// Decompress if needed
			if track.contentEncodings != nil {
				if zr, err := zlib.NewReader(bytes.NewReader(subtitleData)); err == nil {
					if buf, err := io.ReadAll(zr); err == nil {
						subtitleData = buf
					}
					_ = zr.Close()
				}
			}

			milliseconds := float64(packet.StartTime) / 1e6 // Convert nanoseconds to milliseconds
			duration := float64(packet.EndTime-packet.StartTime) / 1e6

			// If duration is 0, try to use default duration
			if duration == 0 && track.defaultDuration > 0 {
				duration = float64(track.defaultDuration) / 1e6
			}

			subtitleEvent := mp.processSubtitleData(packet.Track, track, subtitleData, milliseconds, duration, packet.FilePos, sampler, lastSubtitleEvents, subtitleCh, extractCtx, pgsDecoders)
			if subtitleEvent != nil {
				eventCopy := *subtitleEvent
				lastSubtitleEvents[packet.Track] = &eventCopy
			}
		}
	}()

	return subtitleCh, errCh, startedCh
}

var ssaKeys = []string{"readorder", "layer", "style", "name", "marginl", "marginr", "marginv", "effect"}

// processSubtitleData processes subtitle data and sends events to the channel
func (mp *MetadataParser) processSubtitleData(
	trackNum uint8,
	track *TrackInfo,
	subtitleData []byte,
	milliseconds, duration float64,
	headPos uint64,
	sampler *zerolog.Logger,
	lastSubtitleEvents map[uint8]*SubtitleEvent,
	subtitleCh chan<- *SubtitleEvent,
	ctx context.Context,
	pgsDecoders map[uint8]*pgs.PgsDecoder,
) *SubtitleEvent {
	initialText := string(subtitleData)
	subtitleEvent := &SubtitleEvent{
		TrackNumber: uint64(trackNum),
		Text:        initialText,
		StartTime:   milliseconds,
		Duration:    duration,
		CodecID:     track.CodecID,
		ExtraData:   make(map[string]string),
		HeadPos:     int64(headPos),
	}

	// Handling for ASS/SSA format
	switch {
	case track.CodecID == "S_TEXT/ASS" || track.CodecID == "S_TEXT/SSA":
		values := strings.Split(initialText, ",")
		if len(values) < 9 {
			return nil
		}
		startIndex := 1
		if track.CodecID == "S_TEXT/SSA" {
			startIndex = 2
		}
		for i := startIndex; i < 8 && i < len(values); i++ {
			if i < len(ssaKeys) {
				subtitleEvent.ExtraData[ssaKeys[i]] = values[i]
			}
		}
		if len(values) > 8 {
			text := strings.Join(values[8:], ",")
			subtitleEvent.Text = text
		}
	case track.CodecID == "S_TEXT/UTF8":
		subtitleEvent.Text = UTF8ToASSText(initialText)
		subtitleEvent.CodecID = "S_TEXT/ASS"
		subtitleEvent.ExtraData["readorder"] = "0"
		subtitleEvent.ExtraData["layer"] = "0"
		subtitleEvent.ExtraData["style"] = "Default"
		subtitleEvent.ExtraData["name"] = "Default"
		subtitleEvent.ExtraData["marginl"] = "0"
		subtitleEvent.ExtraData["marginr"] = "0"
	case track.CodecID == "S_HDMV/PGS":
		// Initialize decoder if not exists
		if _, exists := pgsDecoders[trackNum]; !exists {
			pgsDecoders[trackNum] = pgs.NewPgsDecoder()
		}
		decoder := pgsDecoders[trackNum]

		segments := pgs.ListPgsSegments(subtitleData)
		sampler.Debug().
			Uint8("track", trackNum).
			Int("dataLen", len(subtitleData)).
			Strs("segments", segments).
			Msg("mkvparser: Processing PGS packet")

		// Decode PGS packet
		img, err := decoder.DecodePacket(subtitleData)
		if err != nil {
			mp.logger.Warn().Err(err).Uint8("track", trackNum).Msg("mkvparser: Failed to decode PGS packet")
			return nil
		}

		// Check if this is a clear command (no objects to display)
		if decoder.IsClearCommand() {
			sampler.Debug().
				Uint8("track", trackNum).
				Float64("clearTime", milliseconds).
				Msg("mkvparser: PGS clear command received")

			// Update the previous subtitle's duration to end at this clear command
			if lastEvent, exists := lastSubtitleEvents[trackNum]; exists {
				calculatedDuration := milliseconds - lastEvent.StartTime
				if calculatedDuration > 0 {
					updatedLastEvent := *lastEvent
					updatedLastEvent.Duration = calculatedDuration

					sampler.Debug().
						Uint8("trackNum", trackNum).
						Float64("previousStartTime", lastEvent.StartTime).
						Float64("clearDuration", calculatedDuration).
						Msg("mkvparser: Updated PGS subtitle duration from clear command")

					select {
					case subtitleCh <- &updatedLastEvent:
					case <-ctx.Done():
						return nil
					}

					// Remove the last event as it's now been properly terminated
					delete(lastSubtitleEvents, trackNum)
				}
			}

			return nil
		}

		// Only process if we got an image
		if img != nil {
			sampler.Debug().
				Uint8("track", trackNum).
				Int("width", img.Bounds().Dx()).
				Int("height", img.Bounds().Dy()).
				Msg("mkvparser: PGS image decoded successfully")

			// If there's a buffered event, send it now with duration up to this new subtitle
			if bufferedEvent, exists := lastSubtitleEvents[trackNum]; exists {
				calculatedDuration := milliseconds - bufferedEvent.StartTime
				if calculatedDuration > 0 {
					bufferedEvent.Duration = calculatedDuration

					sampler.Debug().
						Uint8("trackNum", trackNum).
						Float64("startTime", bufferedEvent.StartTime).
						Float64("duration", calculatedDuration).
						Msg("mkvparser: Sending previous PGS subtitle (replaced by new one)")

					select {
					case subtitleCh <- bufferedEvent:
					case <-ctx.Done():
						return nil
					}
				}
			}

			// Encode image to base64 PNG
			encodedImage, err := pgs.EncodePgsImageToBase64PNG(img, png.BestSpeed)
			if err != nil {
				mp.logger.Warn().Err(err).Uint8("track", trackNum).Msg("mkvparser: Failed to encode PGS image")
				return nil
			}

			subtitleEvent.Text = encodedImage
			subtitleEvent.ExtraData["type"] = "image"
			subtitleEvent.ExtraData["width"] = fmt.Sprintf("%d", img.Bounds().Dx())
			subtitleEvent.ExtraData["height"] = fmt.Sprintf("%d", img.Bounds().Dy())

			// Don't set duration yet, will be calculated when clear command or next subtitle arrives
			subtitleEvent.Duration = 0

			// Add composition information if available
			if comp := decoder.GetCurrentComposition(); comp != nil {
				subtitleEvent.ExtraData["canvas_width"] = fmt.Sprintf("%d", comp.Width)
				subtitleEvent.ExtraData["canvas_height"] = fmt.Sprintf("%d", comp.Height)

				// Add position information from first composition object
				if len(comp.Objects) > 0 {
					obj := comp.Objects[0]
					subtitleEvent.ExtraData["x"] = fmt.Sprintf("%d", obj.X)
					subtitleEvent.ExtraData["y"] = fmt.Sprintf("%d", obj.Y)

					if obj.Cropped {
						subtitleEvent.ExtraData["crop_x"] = fmt.Sprintf("%d", obj.CropX)
						subtitleEvent.ExtraData["crop_y"] = fmt.Sprintf("%d", obj.CropY)
						subtitleEvent.ExtraData["crop_width"] = fmt.Sprintf("%d", obj.CropWidth)
						subtitleEvent.ExtraData["crop_height"] = fmt.Sprintf("%d", obj.CropHeight)
					}
				}
			}

			// Buffer this event, wait for clear command or next subtitle
			// Return nil to skip the normal send logic
			eventCopy := *subtitleEvent
			lastSubtitleEvents[trackNum] = &eventCopy
			return nil
		} else {
			// Packet processed but no complete image yet (multi-segment)
			return nil
		}
	}

	// Handle previous subtitle event duration
	// For non-PGS subs, calculate the duration from the previous subtitle event
	if track.CodecID != "S_HDMV/PGS" {
		if lastEvent, exists := lastSubtitleEvents[trackNum]; exists {
			if lastEvent.Duration == 0 {
				calculatedDuration := milliseconds - lastEvent.StartTime
				if calculatedDuration > 0 {
					updatedLastEvent := *lastEvent
					updatedLastEvent.Duration = calculatedDuration

					sampler.Trace().
						Uint8("trackNum", trackNum).
						Float64("previousStartTime", lastEvent.StartTime).
						Float64("calculatedDuration", calculatedDuration).
						Msg("mkvparser: Updated previous subtitle event duration")

					select {
					case subtitleCh <- &updatedLastEvent:
					case <-ctx.Done():
						return nil
					}
				}
			}
		}
	}

	sampler.Trace().
		Uint8("trackNum", trackNum).
		Float64("startTime", milliseconds).
		Float64("duration", duration).
		Str("codecId", track.CodecID).
		Msg("mkvparser: Subtitle event")

	// Send current event
	// PGS subtitles are buffered and sent from within their handling code
	// Other subtitles are sent if they have a duration
	if track.CodecID != "S_HDMV/PGS" && duration > 0 {
		select {
		case subtitleCh <- subtitleEvent:
		case <-ctx.Done():
			return nil
		}
	}

	return subtitleEvent
}

// findNextClusterOffset searches for the Matroska Cluster ID in the ReadSeeker
func findNextClusterOffset(rs io.ReadSeeker, seekOffset, backoffBytes int64) (int64, error) {
	if seekOffset > backoffBytes {
		seekOffset -= backoffBytes
	} else {
		seekOffset = 0
	}

	absPosOfNextRead, err := rs.Seek(seekOffset, io.SeekStart)
	if err != nil {
		return -1, fmt.Errorf("initial seek to %d failed: %w", seekOffset, err)
	}

	mainBuf := make([]byte, clusterSearchChunkSize)
	searchBuf := make([]byte, (len(matroskaClusterID)-1)+clusterSearchChunkSize)

	lenOverlapCarried := 0

	for {
		n, readErr := rs.Read(mainBuf)

		if n == 0 && readErr == io.EOF {
			return -1, fmt.Errorf("cluster ID not found, EOF reached before reading new data")
		}
		if readErr != nil && readErr != io.EOF {
			return -1, fmt.Errorf("error reading file: %w", readErr)
		}

		copy(searchBuf[lenOverlapCarried:], mainBuf[:n])
		currentSearchWindow := searchBuf[:lenOverlapCarried+n]

		idx := bytes.Index(currentSearchWindow, matroskaClusterID)
		if idx != -1 {
			foundAtAbsoluteOffset := (absPosOfNextRead - int64(lenOverlapCarried)) + int64(idx)

			_, seekErr := rs.Seek(foundAtAbsoluteOffset, io.SeekStart)
			if seekErr != nil {
				return -1, fmt.Errorf("failed to seek to found cluster position %d: %w", foundAtAbsoluteOffset, seekErr)
			}
			return foundAtAbsoluteOffset, nil
		}

		if readErr == io.EOF {
			return -1, io.EOF
		}

		if len(currentSearchWindow) >= len(matroskaClusterID)-1 {
			lenOverlapCarried = len(matroskaClusterID) - 1
			copy(searchBuf[:lenOverlapCarried], currentSearchWindow[len(currentSearchWindow)-lenOverlapCarried:])
		} else {
			lenOverlapCarried = len(currentSearchWindow)
			copy(searchBuf[:lenOverlapCarried], currentSearchWindow)
		}

		absPosOfNextRead += int64(n)
	}
}
