package mkvparser

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/remko/go-mkvparse"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	maxScanBytes = 35 * 1024 * 1024 // 35MB
	// Default timecode scale (1ms)
	defaultTimecodeScale = 1_000_000

	defaultClusterSearchChunkSize = 8192             // 8KB
	defaultClusterSearchDepth     = 10 * 1024 * 1024 // 1MB
)

var matroskaClusterID = []byte{0x1F, 0x43, 0xB6, 0x75}

// SubtitleEvent holds information for a single subtitle entry.
type SubtitleEvent struct {
	TrackNumber uint64            `json:"trackNumber"`
	Text        string            `json:"text"`      // Content
	StartTime   float64           `json:"startTime"` // Start time in seconds
	Duration    float64           `json:"duration"`  // Duration in seconds
	CodecID     string            `json:"codecID"`   // e.g., "S_TEXT/ASS", "S_TEXT/UTF8"
	ExtraData   map[string]string `json:"extraData,omitempty"`

	readerPos int64 `json:"-"` // Position in the stream
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
	timecodeScale uint64
	currentTrack  *TrackInfo
	tracks        []*TrackInfo
	info          *Info
	chapters      []*ChapterInfo
	attachments   []*AttachmentInfo

	// Result
	extractedMetadata *Metadata

	// subtitleStreamTails stores the start offsets of each subtitle stream
	subtitleStreamTails *result.Map[int64, bool]
}

// NewMetadataParser creates a new MetadataParser.
func NewMetadataParser(reader io.ReadSeeker, logger *zerolog.Logger) *MetadataParser {
	return &MetadataParser{
		reader:              reader,
		logger:              logger,
		realLogger:          logger,
		timecodeScale:       defaultTimecodeScale,
		tracks:              make([]*TrackInfo, 0),
		chapters:            make([]*ChapterInfo, 0),
		attachments:         make([]*AttachmentInfo, 0),
		info:                &Info{},
		extractedMetadata:   nil,
		subtitleStreamTails: result.NewResultMap[int64, bool](),
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

func getLanguageCode(track *TrackInfo) string {
	if track.LanguageIETF != "" {
		return track.LanguageIETF
	}
	if track.Language != "" && track.Language != "und" {
		return track.Language
	}
	return "eng"
}

// parseMetadataOnce performs the actual parsing of the file stream.
func (mp *MetadataParser) parseMetadataOnce(ctx context.Context) {
	mp.parseOnce.Do(func() {
		mp.logger.Debug().Msg("mkvparser: Starting metadata parsing")
		startTime := time.Now()

		// Create a handler for parsing
		handler := &metadataHandler{
			mp:     mp,
			ctx:    ctx,
			logger: mp.logger,
		}

		_, _ = mp.reader.Seek(0, io.SeekStart)

		limitedReader, err := util.NewLimitedReadSeeker(mp.reader, maxScanBytes)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to create limited reader")
			mp.parseErr = fmt.Errorf("mkvparser: Failed to create limited reader: %w", err)
			return
		}

		// Parse the MKV file
		err = mkvparse.ParseSections(limitedReader, handler,
			mkvparse.InfoElement,
			mkvparse.AttachmentsElement,
			mkvparse.TracksElement,
			mkvparse.SegmentElement,
			mkvparse.ChaptersElement,
		)
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "unexpected EOF") {
			mp.logger.Error().Err(err).Msg("mkvparser: MKV parsing error")
			mp.parseErr = fmt.Errorf("mkv parsing failed: %w", err)
		} else if err != nil {
			mp.logger.Debug().Err(err).Msg("mkvparser: MKV parsing finished with EOF/unexpected EOF (expected outcome).")
			mp.parseErr = nil
		} else {
			mp.logger.Debug().Msg("mkvparser: MKV parsing completed fully within scan limit.")
			mp.parseErr = nil
		}

		logMsg := mp.logger.Info().Dur("parseDuration", time.Since(startTime))
		if mp.parseErr != nil {
			logMsg.Err(mp.parseErr)
		}
		logMsg.Msg("mkvparser: Metadata parsing attempt finished")
	})
}

// Handler for parsing metadata
type metadataHandler struct {
	mkvparse.DefaultHandler
	mp     *MetadataParser
	ctx    context.Context
	logger *zerolog.Logger

	// Track parsing state
	inTrackEntry bool
	currentTrack *TrackInfo
	inVideo      bool
	inAudio      bool

	// Chapter parsing state
	inEditionEntry   bool
	inChapterAtom    bool
	currentChapter   *ChapterInfo
	inChapterDisplay bool
	currentLanguages []string // Temporary storage for chapter languages
	currentIETF      []string // Temporary storage for chapter IETF languages

	// Attachment parsing state
	isAttachment      bool
	currentAttachment *AttachmentInfo
}

func (h *metadataHandler) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	switch id {
	case mkvparse.SegmentElement:
		return true, nil // Parse Segment and its children
	case mkvparse.TracksElement:
		return true, nil // Parse Track metadata
	case mkvparse.TrackEntryElement:
		h.inTrackEntry = true
		h.currentTrack = &TrackInfo{
			Default: false,
			Enabled: true,
		}
		return true, nil
	case mkvparse.VideoElement:
		h.inVideo = true
		if h.currentTrack != nil && h.currentTrack.Video == nil {
			h.currentTrack.Video = &VideoTrack{}
		}
		return true, nil
	case mkvparse.AudioElement:
		h.inAudio = true
		if h.currentTrack != nil && h.currentTrack.Audio == nil {
			h.currentTrack.Audio = &AudioTrack{}
		}
		return true, nil
	case mkvparse.InfoElement:
		if h.mp.info == nil {
			h.mp.info = &Info{}
		}
		return true, nil
	case mkvparse.ChaptersElement:
		return true, nil
	case mkvparse.EditionEntryElement:
		h.inEditionEntry = true
		return true, nil
	case mkvparse.ChapterAtomElement:
		h.inChapterAtom = true
		h.currentChapter = &ChapterInfo{}
		return true, nil
	case mkvparse.ChapterDisplayElement:
		h.inChapterDisplay = true
		h.currentLanguages = make([]string, 0)
		h.currentIETF = make([]string, 0)
		return true, nil
	case mkvparse.AttachmentsElement:
		return true, nil
	case mkvparse.AttachedFileElement:
		h.isAttachment = true
		h.currentAttachment = &AttachmentInfo{}
		return true, nil
	case mkvparse.ContentEncodingsElement:
		if h.currentTrack != nil && h.currentTrack.contentEncodings == nil {
			h.currentTrack.contentEncodings = &ContentEncodings{
				ContentEncoding: make([]ContentEncoding, 0),
			}
		} else if h.isAttachment && h.currentAttachment != nil {
			// Handle content encoding for attachments
			h.currentAttachment.IsCompressed = true
		}
		return true, nil
	}
	return false, nil
}

func (h *metadataHandler) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.TrackEntryElement:
		if h.currentTrack != nil {
			h.mp.tracks = append(h.mp.tracks, h.currentTrack)
		}
		h.inTrackEntry = false
		h.currentTrack = nil
	case mkvparse.VideoElement:
		h.inVideo = false
	case mkvparse.AudioElement:
		h.inAudio = false
	case mkvparse.EditionEntryElement:
		h.inEditionEntry = false
	case mkvparse.ChapterAtomElement:
		if h.currentChapter != nil && h.inEditionEntry {
			h.mp.chapters = append(h.mp.chapters, h.currentChapter)
		}
		h.inChapterAtom = false
		h.currentChapter = nil
	case mkvparse.ChapterDisplayElement:
		if h.currentChapter != nil {
			h.currentChapter.Languages = h.currentLanguages
			h.currentChapter.LanguagesIETF = h.currentIETF
		}
		h.inChapterDisplay = false
		h.currentLanguages = nil
		h.currentIETF = nil
	case mkvparse.AttachedFileElement:
		if h.currentAttachment != nil {
			// Handle compressed attachments if needed
			if h.currentAttachment.Data != nil && h.currentAttachment.IsCompressed {
				zlibReader, err := zlib.NewReader(bytes.NewReader(h.currentAttachment.Data))
				if err != nil {
					h.logger.Error().Err(err).Str("filename", h.currentAttachment.Filename).Msg("mkvparser: Failed to create zlib reader for attachment")
				} else {
					decompressedData, err := io.ReadAll(zlibReader)
					_ = zlibReader.Close()
					if err != nil {
						h.logger.Error().Err(err).Str("filename", h.currentAttachment.Filename).Msg("mkvparser: Failed to decompress attachment")
					} else {
						h.currentAttachment.Data = decompressedData
						h.currentAttachment.Size = len(decompressedData)
					}
				}
			}
			fileExt := strings.ToLower(filepath.Ext(h.currentAttachment.Filename))
			if fileExt == ".ttf" || fileExt == ".woff2" || fileExt == ".woff" || fileExt == ".otf" {
				h.currentAttachment.Type = AttachmentTypeFont
			} else if fileExt == ".ass" || fileExt == ".ssa" || fileExt == ".srt" || fileExt == ".vtt" {
				h.currentAttachment.Type = AttachmentTypeSubtitle
			}
			h.mp.attachments = append(h.mp.attachments, h.currentAttachment)
		}
		h.isAttachment = false
		h.currentAttachment = nil
	}
	return nil
}

func (h *metadataHandler) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.CodecIDElement:
		if h.currentTrack != nil {
			h.currentTrack.CodecID = value
		}
	case mkvparse.LanguageElement:
		if h.currentTrack != nil {
			h.currentTrack.Language = value
		} else if h.inChapterDisplay {
			h.currentLanguages = append(h.currentLanguages, value)
		}
	case mkvparse.LanguageIETFElement:
		if h.currentTrack != nil {
			h.currentTrack.LanguageIETF = value
		} else if h.inChapterDisplay {
			h.currentIETF = append(h.currentIETF, value)
		}
	case mkvparse.NameElement:
		if h.currentTrack != nil {
			h.currentTrack.Name = value
		}
	case mkvparse.TitleElement:
		if h.mp.info != nil {
			h.mp.info.Title = value
		}
	case mkvparse.MuxingAppElement:
		if h.mp.info != nil {
			h.mp.info.MuxingApp = value
		}
	case mkvparse.WritingAppElement:
		if h.mp.info != nil {
			h.mp.info.WritingApp = value
		}
	case mkvparse.ChapStringElement:
		if h.inChapterDisplay && h.currentChapter != nil {
			h.currentChapter.Text = value
		}
	case mkvparse.FileDescriptionElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Description = value
		}
	case mkvparse.FileNameElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Filename = value
		}
	case mkvparse.FileMimeTypeElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Mimetype = value
		}
	}
	return nil
}

func (h *metadataHandler) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.TimecodeScaleElement:
		h.mp.timecodeScale = uint64(value)
		if h.mp.info != nil {
			h.mp.info.TimecodeScale = uint64(value)
		}
	case mkvparse.TrackNumberElement:
		if h.currentTrack != nil {
			h.currentTrack.Number = value
		}
	case mkvparse.TrackUIDElement:
		if h.currentTrack != nil {
			h.currentTrack.UID = value
		}
	case mkvparse.TrackTypeElement:
		if h.currentTrack != nil {
			h.currentTrack.Type = convertTrackType(uint64(value))
		}
	case mkvparse.DefaultDurationElement:
		if h.currentTrack != nil {
			h.currentTrack.defaultDuration = uint64(value)
		}
	// case mkvparse.DurationElement:
	// 	if h.currentTrack != nil {
	// 		h.currentTrack.Duration = uint64(value)
	// 	}
	case mkvparse.FlagDefaultElement:
		if h.currentTrack != nil {
			h.currentTrack.Default = value == 1
		}
	case mkvparse.FlagForcedElement:
		if h.currentTrack != nil {
			h.currentTrack.Forced = value == 1
		}
	case mkvparse.FlagEnabledElement:
		if h.currentTrack != nil {
			h.currentTrack.Enabled = value == 1
		}
	case mkvparse.PixelWidthElement:
		if h.currentTrack != nil && h.currentTrack.Video != nil {
			h.currentTrack.Video.PixelWidth = uint64(value)
		}
	case mkvparse.PixelHeightElement:
		if h.currentTrack != nil && h.currentTrack.Video != nil {
			h.currentTrack.Video.PixelHeight = uint64(value)
		}
	case mkvparse.ChannelsElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.Channels = uint64(value)
		}
	case mkvparse.BitDepthElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.BitDepth = uint64(value)
		}
	case mkvparse.ChapterTimeStartElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.Start = float64(value) * float64(h.mp.timecodeScale) / 1e9
		}
	case mkvparse.ChapterTimeEndElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.End = float64(value) * float64(h.mp.timecodeScale) / 1e9
		}
	case mkvparse.ChapterUIDElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.UID = uint64(value)
		}
	case mkvparse.FileUIDElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.UID = uint64(value)
		}
	}
	return nil
}

func (h *metadataHandler) HandleFloat(id mkvparse.ElementID, value float64, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.DurationElement:
		if h.mp.info != nil {
			h.mp.info.Duration = value
		}
	case mkvparse.SamplingFrequencyElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.SamplingFrequency = value
		}
	}
	return nil
}

func (h *metadataHandler) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.CodecPrivateElement:
		if h.currentTrack != nil {
			h.currentTrack.CodecPrivate = string(value)
			h.currentTrack.CodecPrivate = strings.ReplaceAll(h.currentTrack.CodecPrivate, "\r\n", "\n")
		}
	case mkvparse.FileDataElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Data = value
			h.currentAttachment.Size = len(value)
		}
	}
	return nil
}

// GetMetadata extracts all relevant metadata from the file.
func (mp *MetadataParser) GetMetadata(ctx context.Context) *Metadata {
	mp.parseMetadataOnce(ctx)

	mp.metadataOnce.Do(func() {
		result := &Metadata{
			VideoTracks:    make([]*TrackInfo, 0),
			AudioTracks:    make([]*TrackInfo, 0),
			SubtitleTracks: make([]*TrackInfo, 0),
			Tracks:         mp.tracks,
			Chapters:       mp.chapters,
			Attachments:    mp.attachments,
			Error:          mp.parseErr,
		}

		if mp.parseErr != nil {
			if !(errors.Is(mp.parseErr, context.Canceled) || errors.Is(mp.parseErr, context.DeadlineExceeded)) {
				mp.extractedMetadata = result
				return
			}
		}

		if mp.info != nil {
			result.Title = mp.info.Title
			result.MuxingApp = mp.info.MuxingApp
			result.WritingApp = mp.info.WritingApp
			result.TimecodeScale = float64(mp.timecodeScale)
			if mp.info.Duration > 0 {
				result.Duration = (mp.info.Duration * float64(mp.timecodeScale)) / 1e9
			}
		}

		mp.logger.Debug().
			Int("tracks", len(mp.tracks)).
			Int("chapters", len(mp.chapters)).
			Int("attachments", len(mp.attachments)).
			Msg("mkvparser: Metadata parsing complete")

		if len(mp.chapters) == 0 {
			mp.logger.Debug().Msg("mkvparser: No chapters found")
		}
		if len(mp.attachments) == 0 {
			mp.logger.Debug().Msg("mkvparser: No attachments found")
		}

		for _, track := range mp.tracks {
			switch track.Type {
			case TrackTypeVideo:
				result.VideoTracks = append(result.VideoTracks, track)
			case TrackTypeAudio:
				result.AudioTracks = append(result.AudioTracks, track)
			case TrackTypeSubtitle:
				result.SubtitleTracks = append(result.SubtitleTracks, track)
			}
		}

		// Generate MimeCodec string
		var codecStrings []string
		seenCodecs := make(map[string]bool)

		if len(result.VideoTracks) > 0 {
			firstVideoTrack := result.VideoTracks[0]
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

		for _, audioTrack := range result.AudioTracks {
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
			case "A_MS/ACM":
				if strings.Contains(strings.ToLower(audioTrack.Name), "vorbis") {
					audioCodecStr = "vorbis"
				} else if audioTrack.CodecID != "" {
					audioCodecStr = strings.ToLower(strings.ReplaceAll(audioTrack.CodecID, "/", "."))
				}
			case "A_VORBIS":
				audioCodecStr = "vorbis"
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

		if len(codecStrings) > 0 {
			result.MimeCodec = fmt.Sprintf("video/x-matroska; codecs=\"%s\"", strings.Join(codecStrings, ", "))
		} else {
			result.MimeCodec = "video/x-matroska"
		}

		mp.extractedMetadata = result
	})

	return mp.extractedMetadata
}

// ExtractSubtitles extracts subtitles from a streaming source by reading it as a continuous flow.
// If an offset is provided, it will seek to the cluster near the offset and start parsing from there.
//
// The function returns a channel of SubtitleEvent which will be closed when:
// - The context is canceled
// - The entire stream is processed
// - An unrecoverable error occurs (which is also returned in the error channel)
func (mp *MetadataParser) ExtractSubtitles(ctx context.Context, newReader io.ReadSeekCloser, offset int64) (<-chan *SubtitleEvent, <-chan error) {
	subtitleCh := make(chan *SubtitleEvent)
	errCh := make(chan error, 1)

	var closeOnce sync.Once
	closeChannels := func(err error) {
		closeOnce.Do(func() {
			select {
			case errCh <- err:
			default: // Channel might be full or closed, ignore
			}
			close(subtitleCh)
			close(errCh)
		})
	}

	// Create a cancellable context for coordination between goroutines
	extractCtx, cancel := context.WithCancel(ctx)

	// Add the stream offset to the table
	_, exists := mp.subtitleStreamTails.Get(offset)
	if !exists {
		mp.subtitleStreamTails.Set(offset, false) // false -> not complete
	}

	if offset > 0 {
		mp.logger.Debug().Int64("offset", offset).Msg("mkvparser: Attempting to find cluster near offset")

		clusterSeekOffset, err := findNextClusterOffset(newReader, offset)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to seek to offset for subtitle extraction")
			cancel()
			closeChannels(err)
			return subtitleCh, errCh
		}

		mp.logger.Debug().Int64("clusterSeekOffset", clusterSeekOffset).Msg("mkvparser: Found cluster near offset")

		_, err = newReader.Seek(clusterSeekOffset, io.SeekStart)
		if err != nil {
			mp.logger.Error().Err(err).Msg("mkvparser: Failed to seek to cluster offset")
			cancel()
			closeChannels(err)
			return subtitleCh, errCh
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				mp.logger.Error().Msgf("mkvparser: Subtitle extraction goroutine panicked: %v", r)
				closeChannels(fmt.Errorf("subtitle extraction goroutine panic: %v", r))
			}
		}()
		defer cancel() // Ensure context is cancelled when main goroutine exits
		defer mp.logger.Trace().Msgf("mkvparser: Subtitle extraction goroutine finished.")

		sampler := lo.ToPtr(mp.logger.Sample(&zerolog.BasicSampler{N: 10}))

		// First, ensure metadata is parsed to get track information
		mp.parseMetadataOnce(extractCtx)

		if mp.parseErr != nil && !errors.Is(mp.parseErr, io.EOF) && !strings.Contains(mp.parseErr.Error(), "unexpected EOF") {
			mp.logger.Error().Err(mp.parseErr).Msg("mkvparser: ExtractSubtitles cannot proceed due to initial metadata parsing error")
			closeChannels(fmt.Errorf("initial metadata parse failed: %w", mp.parseErr))
			return
		}

		// Create a map of subtitle tracks for quick lookup
		subtitleTracks := make(map[uint64]*TrackInfo)
		for _, track := range mp.tracks {
			if track.Type == TrackTypeSubtitle {
				subtitleTracks[uint64(track.Number)] = track
			}
		}

		if len(subtitleTracks) == 0 {
			mp.logger.Info().Msg("mkvparser: No subtitle tracks found for streaming")
			closeChannels(nil)
			return
		}

		// Create a handler for subtitle extraction
		handler := &subtitleHandler{
			mp:             mp,
			ctx:            extractCtx, // use extraction context instead of original context
			logger:         mp.logger,
			sampler:        sampler,
			subtitleCh:     subtitleCh,
			subtitleTracks: subtitleTracks,
			timecodeScale:  mp.timecodeScale,
			clusterTime:    0,
			reader:         newReader,
		}

		// Start monitoring goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					mp.logger.Error().Msgf("mkvparser: Subtitle monitoring goroutine panicked: %v", r)
				}
			}()

			tick := time.NewTicker(1 * time.Second)
			defer tick.Stop()
			for {
				select {
				case <-extractCtx.Done():
					return
				case <-tick.C:
					shouldStop := false
					handler.headPosMu.RLock()
					mp.subtitleStreamTails.Range(func(tail int64, value bool) bool {
						// e.g. this stream's start offset is 1000, subtitleStreamTails contains {1000, 2500, 4000}
						// the stream should stop if the head position is >= 2500 or >= 4000
						if offset != tail && handler.headPos >= tail {
							shouldStop = true
							return false
						}
						return true
					})
					handler.headPosMu.RUnlock()
					if shouldStop {
						mp.logger.Debug().Msg("mkvparser: Subtitle stream interrupted")
						closeChannels(nil)
						return
					}
				}
			}
		}()

		// Parse the MKV file for subtitles
		err := mkvparse.Parse(newReader, handler)
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "unexpected EOF") {
			mp.logger.Error().Err(err).Msg("mkvparser: Unrecoverable error during subtitle stream parsing")
			closeChannels(err)
		} else {
			mp.logger.Info().Msg("mkvparser: Subtitle streaming completed successfully or with expected EOF.")
			closeChannels(nil)
		}
	}()

	return subtitleCh, errCh
}

// Handler for subtitle extraction
type subtitleHandler struct {
	mkvparse.DefaultHandler
	mp                   *MetadataParser
	ctx                  context.Context
	logger               *zerolog.Logger
	sampler              *zerolog.Logger
	subtitleCh           chan<- *SubtitleEvent
	subtitleTracks       map[uint64]*TrackInfo
	timecodeScale        uint64
	clusterTime          uint64
	currentBlockDuration uint64
	reader               io.ReadSeekCloser
	headPosMu            sync.RWMutex
	headPos              int64
}

func (h *subtitleHandler) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	switch id {
	case mkvparse.SegmentElement:
		return true, nil
	case mkvparse.ClusterElement:
		return true, nil
	case mkvparse.BlockGroupElement:
		return true, nil
	}
	return false, nil
}

func (h *subtitleHandler) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	if id == mkvparse.TimecodeElement {
		h.clusterTime = uint64(value)
	} else if id == mkvparse.BlockDurationElement {
		h.currentBlockDuration = uint64(value)
	}
	return nil
}

func (h *subtitleHandler) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.SimpleBlockElement, mkvparse.BlockElement:
		if len(value) < 4 {
			return nil
		}

		trackNum := uint64(value[0] & 0x7F)
		track, isSubtitle := h.subtitleTracks[trackNum]
		if !isSubtitle {
			return nil
		}

		pos, err := h.reader.Seek(0, io.SeekCurrent)
		if err == nil {
			h.headPosMu.Lock()
			h.headPos = pos
			h.headPosMu.Unlock()
		}

		blockTimecode := int16(value[1])<<8 | int16(value[2])
		absoluteTimeScaled := h.clusterTime + uint64(blockTimecode)
		timestampNs := absoluteTimeScaled * h.timecodeScale
		milliseconds := float64(timestampNs) / 1e6

		// Calculate duration in milliseconds
		var duration float64
		if h.currentBlockDuration > 0 {
			duration = float64(h.currentBlockDuration*h.timecodeScale) / 1e6 // ms
		} else if track.defaultDuration > 0 {
			duration = float64(track.defaultDuration) / 1e6 // ms
		}

		subtitleData := value[4:]

		// Handle content encoding (compression)
		if track.contentEncodings != nil {
			for _, encoding := range track.contentEncodings.ContentEncoding {
				if encoding.ContentCompression != nil && encoding.ContentCompression.ContentCompAlgo == 0 {
					zlibReader, err := zlib.NewReader(bytes.NewReader(subtitleData))
					if err != nil {
						h.logger.Error().Err(err).Uint64("trackNum", trackNum).Msg("mkvparser: Failed to create zlib reader for subtitle frame")
						continue
					}
					decompressedData, err := io.ReadAll(zlibReader)
					_ = zlibReader.Close()
					if err != nil {
						h.logger.Error().Err(err).Uint64("trackNum", trackNum).Msg("mkvparser: Failed to decompress zlib subtitle frame")
						continue
					}
					subtitleData = decompressedData
					break
				}
			}
		}

		initialText := string(subtitleData)
		subtitleEvent := &SubtitleEvent{
			TrackNumber: trackNum,
			Text:        initialText,
			StartTime:   milliseconds,
			Duration:    duration,
			CodecID:     track.CodecID,
			ExtraData:   make(map[string]string),
		}

		// Special handling for ASS/SSA format
		if track.CodecID == "S_TEXT/ASS" || track.CodecID == "S_TEXT/SSA" {
			values := strings.Split(initialText, ",")
			if len(values) < 9 {
				//h.logger.Warn().
				//	Str("text", initialText).
				//	Int("fields", len(values)).
				//	Msg("mkvparser: Invalid ASS/SSA subtitle format, not enough fields")
				return nil
			}

			// Format: ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text
			startIndex := 1
			if track.CodecID == "S_TEXT/SSA" {
				startIndex = 2 // Skip both ReadOrder and Layer for SSA
			}

			subtitleEvent.ExtraData["readorder"] = values[0]
			if track.CodecID == "S_TEXT/ASS" {
				subtitleEvent.ExtraData["layer"] = values[1]
			}

			subtitleEvent.ExtraData["style"] = values[startIndex+1]
			subtitleEvent.ExtraData["name"] = values[startIndex+2]
			subtitleEvent.ExtraData["marginl"] = values[startIndex+3]
			subtitleEvent.ExtraData["marginr"] = values[startIndex+4]
			subtitleEvent.ExtraData["marginv"] = values[startIndex+5]
			subtitleEvent.ExtraData["effect"] = values[startIndex+6]

			text := strings.Join(values[startIndex+7:], ",")
			// Remove leading comma if present
			text = strings.TrimPrefix(text, ",")
			subtitleEvent.Text = strings.TrimSpace(text)
		}

		h.sampler.Trace().
			Uint64("trackNum", trackNum).
			Float64("startTime", milliseconds).
			Float64("duration", duration).
			Str("codecId", track.CodecID).
			Str("text", subtitleEvent.Text).
			Interface("data", subtitleEvent.ExtraData).
			Msg("mkvparser: Subtitle event")

		// Check context before sending to avoid sending on closed channel
		select {
		case h.subtitleCh <- subtitleEvent:
			// Successfully sent
		case <-h.ctx.Done():
			h.logger.Debug().Msg("mkvparser: Subtitle sending cancelled by context.")
			return h.ctx.Err()
		}

		// Reset the block duration for the next block
		h.currentBlockDuration = 0
	}
	return nil
}

// findNextClusterOffset searches for the Matroska Cluster ID in the ReadSeeker rs,
// starting from seekOffset. It returns the absolute file offset of the found Cluster ID,
// or an error. If found, the ReadSeeker's position is set to the start of the Cluster ID.
func findNextClusterOffset(rs io.ReadSeeker, seekOffset int64) (int64, error) {

	// DEVNOTE: findNextClusterOffset is faster than findPrecedingOrCurrentClusterOffset
	// however it's not ideal so we'll offset the offset by 2MB to avoid missing a cluster
	toRemove := int64(2 * 1024 * 1024) // 2MB
	if seekOffset > toRemove {
		seekOffset -= toRemove
	} else {
		seekOffset = 0
	}

	// Seek to the starting position
	absPosOfNextRead, err := rs.Seek(seekOffset, io.SeekStart)
	if err != nil {
		return -1, fmt.Errorf("initial seek to %d failed: %w", seekOffset, err)
	}

	mainBuf := make([]byte, defaultClusterSearchChunkSize)
	searchBuf := make([]byte, (len(matroskaClusterID)-1)+defaultClusterSearchChunkSize)

	lenOverlapCarried := 0 // Length of overlap data copied into searchBuf's start from previous iteration

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
			return -1, fmt.Errorf("cluster ID not found, EOF after final search of remaining data")
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

// findPrecedingOrCurrentClusterOffset searches for the Matroska Cluster ID in the ReadSeeker rs,
// looking for a cluster that starts at or before targetFileOffset. It searches backwards from targetFileOffset.
// It returns the absolute file offset of the found Cluster ID, or an error.
// If found, the ReadSeeker's position is set to the start of the Cluster ID.
func findPrecedingOrCurrentClusterOffset(rs io.ReadSeeker, targetFileOffset int64) (int64, error) {
	mainBuf := make([]byte, defaultClusterSearchChunkSize)
	searchBuf := make([]byte, (len(matroskaClusterID)-1)+defaultClusterSearchChunkSize)

	// Start from targetFileOffset and work backwards
	currentReadEndPos := targetFileOffset + int64(len(matroskaClusterID))
	lenOverlapCarried := 0

	for {
		// Calculate read position and size
		readStartPos := currentReadEndPos - defaultClusterSearchChunkSize
		if readStartPos < 0 {
			readStartPos = 0
		}
		bytesToRead := currentReadEndPos - readStartPos

		// Check if we have enough data to potentially find a cluster
		if bytesToRead < int64(len(matroskaClusterID)) {
			return -1, fmt.Errorf("cluster ID not found at or before offset %d", targetFileOffset)
		}

		// Seek and read
		_, err := rs.Seek(readStartPos, io.SeekStart)
		if err != nil {
			return -1, fmt.Errorf("seek to %d failed: %w", readStartPos, err)
		}

		n, readErr := rs.Read(mainBuf[:bytesToRead])
		if readErr != nil && readErr != io.EOF {
			return -1, fmt.Errorf("error reading file: %w", readErr)
		}
		if n == 0 {
			return -1, fmt.Errorf("no data read at offset %d", readStartPos)
		}

		// Copy data to search buffer
		copy(searchBuf[lenOverlapCarried:], mainBuf[:n])
		currentSearchWindow := searchBuf[:lenOverlapCarried+n]

		// Search for cluster ID in current window
		for i := len(currentSearchWindow) - len(matroskaClusterID); i >= 0; i-- {
			if bytes.Equal(currentSearchWindow[i:i+len(matroskaClusterID)], matroskaClusterID) {
				foundOffset := readStartPos + int64(i)
				if foundOffset <= targetFileOffset {
					_, seekErr := rs.Seek(foundOffset, io.SeekStart)
					if seekErr != nil {
						return -1, fmt.Errorf("failed to seek to found cluster at %d: %w", foundOffset, seekErr)
					}
					return foundOffset, nil
				}
			}
		}

		// Check search depth limit
		if (targetFileOffset - readStartPos) >= defaultClusterSearchDepth {
			return -1, fmt.Errorf("cluster ID not found within search depth %dMB", defaultClusterSearchDepth/1024/1024)
		}

		// If we've reached the start of the file, we're done
		if readStartPos == 0 {
			return -1, fmt.Errorf("cluster ID not found, reached start of file")
		}

		// Prepare for next iteration
		// Keep overlap from start of current window for next search
		if len(currentSearchWindow) >= len(matroskaClusterID)-1 {
			lenOverlapCarried = len(matroskaClusterID) - 1
			copy(searchBuf[:lenOverlapCarried], currentSearchWindow[:lenOverlapCarried])
		} else {
			lenOverlapCarried = len(currentSearchWindow)
			copy(searchBuf[:lenOverlapCarried], currentSearchWindow)
		}

		currentReadEndPos = readStartPos + int64(lenOverlapCarried)
	}
}
