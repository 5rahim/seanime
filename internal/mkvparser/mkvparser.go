package mkvparser

import (
	"bytes"
	"cmp"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"seanime/internal/util"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/gomkv"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	maxScanBytes = 35 * 1024 * 1024 // 35MB
	// Default timecode scale (1ms)
	defaultTimecodeScale = 1_000_000

	clusterSearchChunkSize = 8192             // 8KB
	clusterSearchDepth     = 10 * 1024 * 1024 // 1MB
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
	timecodeScale uint64
	currentTrack  *TrackInfo
	tracks        []*TrackInfo
	info          *Info
	chapters      []*ChapterInfo
	attachments   []*AttachmentInfo

	// Result
	extractedMetadata *Metadata
}

// NewMetadataParser creates a new MetadataParser.
func NewMetadataParser(reader io.ReadSeeker, logger *zerolog.Logger) *MetadataParser {
	return &MetadataParser{
		reader:            reader,
		logger:            logger,
		realLogger:        logger,
		timecodeScale:     defaultTimecodeScale,
		tracks:            make([]*TrackInfo, 0),
		chapters:          make([]*ChapterInfo, 0),
		attachments:       make([]*AttachmentInfo, 0),
		info:              &Info{},
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
		return "SSA"
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
		startTime := time.Now()

		// Create a handler for parsing
		handler := &metadataHandler{
			mp:     mp,
			ctx:    ctx,
			logger: mp.logger,
		}

		_, _ = mp.reader.Seek(0, io.SeekStart)

		// Devnote: Don't limit the depth anymore
		//limitedReader, err := util.NewLimitedReadSeeker(mp.reader, maxScanBytes)
		//if err != nil {
		//	mp.logger.Error().Err(err).Msg("mkvparser: Failed to create limited reader")
		//	mp.parseErr = fmt.Errorf("mkvparser: Failed to create limited reader: %w", err)
		//	return
		//}

		// Parse the MKV file
		err := gomkv.ParseSections(mp.reader, handler,
			gomkv.InfoElement,
			gomkv.AttachmentsElement,
			gomkv.TracksElement,
			gomkv.SegmentElement,
			gomkv.ChaptersElement,
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
	gomkv.DefaultHandler
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

func (h *metadataHandler) HandleMasterBegin(id gomkv.ElementID, info gomkv.ElementInfo) (bool, error) {
	switch id {
	case gomkv.SegmentElement:
		return true, nil // Parse Segment and its children
	case gomkv.TracksElement:
		return true, nil // Parse Track metadata
	case gomkv.TrackEntryElement:
		h.inTrackEntry = true
		h.currentTrack = &TrackInfo{
			Default: false,
			Enabled: true,
		}
		return true, nil
	case gomkv.VideoElement:
		h.inVideo = true
		if h.currentTrack != nil && h.currentTrack.Video == nil {
			h.currentTrack.Video = &VideoTrack{}
		}
		return true, nil
	case gomkv.AudioElement:
		h.inAudio = true
		if h.currentTrack != nil && h.currentTrack.Audio == nil {
			h.currentTrack.Audio = &AudioTrack{}
		}
		return true, nil
	case gomkv.InfoElement:
		if h.mp.info == nil {
			h.mp.info = &Info{}
		}
		return true, nil
	case gomkv.ChaptersElement:
		return true, nil
	case gomkv.EditionEntryElement:
		h.inEditionEntry = true
		return true, nil
	case gomkv.ChapterAtomElement:
		h.inChapterAtom = true
		h.currentChapter = &ChapterInfo{}
		return true, nil
	case gomkv.ChapterDisplayElement:
		h.inChapterDisplay = true
		h.currentLanguages = make([]string, 0)
		h.currentIETF = make([]string, 0)
		return true, nil
	case gomkv.AttachmentsElement:
		return true, nil
	case gomkv.AttachedFileElement:
		h.isAttachment = true
		h.currentAttachment = &AttachmentInfo{}
		return true, nil
	case gomkv.ContentEncodingsElement:
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

func (h *metadataHandler) HandleMasterEnd(id gomkv.ElementID, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.TrackEntryElement:
		if h.currentTrack != nil {
			h.mp.tracks = append(h.mp.tracks, h.currentTrack)
		}
		h.inTrackEntry = false
		h.currentTrack = nil
	case gomkv.VideoElement:
		h.inVideo = false
	case gomkv.AudioElement:
		h.inAudio = false
	case gomkv.EditionEntryElement:
		h.inEditionEntry = false
	case gomkv.ChapterAtomElement:
		if h.currentChapter != nil && h.inEditionEntry {
			h.mp.chapters = append(h.mp.chapters, h.currentChapter)
		}
		h.inChapterAtom = false
		h.currentChapter = nil
	case gomkv.ChapterDisplayElement:
		if h.currentChapter != nil {
			h.currentChapter.Languages = h.currentLanguages
			h.currentChapter.LanguagesIETF = h.currentIETF
		}
		h.inChapterDisplay = false
		h.currentLanguages = nil
		h.currentIETF = nil
	case gomkv.AttachedFileElement:
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
			if _, ok := fontExtensions[fileExt]; ok {
				h.currentAttachment.Type = AttachmentTypeFont
			} else if _, ok := subtitleExtensions[fileExt]; ok {
				h.currentAttachment.Type = AttachmentTypeSubtitle
			} else {
				h.currentAttachment.Type = AttachmentTypeOther
			}
			h.mp.attachments = append(h.mp.attachments, h.currentAttachment)
		}
		h.isAttachment = false
		h.currentAttachment = nil
	}
	return nil
}

func (h *metadataHandler) HandleString(id gomkv.ElementID, value string, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.CodecIDElement:
		if h.currentTrack != nil {
			h.currentTrack.CodecID = value
		}
	case gomkv.LanguageElement:
		if h.currentTrack != nil {
			h.currentTrack.Language = value
		} else if h.inChapterDisplay {
			h.currentLanguages = append(h.currentLanguages, value)
		}
	case gomkv.LanguageIETFElement:
		if h.currentTrack != nil {
			h.currentTrack.LanguageIETF = value
		} else if h.inChapterDisplay {
			h.currentIETF = append(h.currentIETF, value)
		}
	case gomkv.NameElement:
		if h.currentTrack != nil {
			h.currentTrack.Name = value
		}
	case gomkv.TitleElement:
		if h.mp.info != nil {
			h.mp.info.Title = value
		}
	case gomkv.MuxingAppElement:
		if h.mp.info != nil {
			h.mp.info.MuxingApp = value
		}
	case gomkv.WritingAppElement:
		if h.mp.info != nil {
			h.mp.info.WritingApp = value
		}
	case gomkv.ChapStringElement:
		if h.inChapterDisplay && h.currentChapter != nil {
			h.currentChapter.Text = value
		}
	case gomkv.FileDescriptionElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Description = value
		}
	case gomkv.FileNameElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Filename = value
		}
	case gomkv.FileMimeTypeElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.Mimetype = value
		}
	}
	return nil
}

func (h *metadataHandler) HandleInteger(id gomkv.ElementID, value int64, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.TimecodeScaleElement:
		h.mp.timecodeScale = uint64(value)
		if h.mp.info != nil {
			h.mp.info.TimecodeScale = uint64(value)
		}
	case gomkv.TrackNumberElement:
		if h.currentTrack != nil {
			h.currentTrack.Number = value
		}
	case gomkv.TrackUIDElement:
		if h.currentTrack != nil {
			h.currentTrack.UID = value
		}
	case gomkv.TrackTypeElement:
		if h.currentTrack != nil {
			h.currentTrack.Type = convertTrackType(uint64(value))
		}
	case gomkv.DefaultDurationElement:
		if h.currentTrack != nil {
			h.currentTrack.defaultDuration = uint64(value)
		}
	case gomkv.FlagDefaultElement:
		if h.currentTrack != nil {
			h.currentTrack.Default = value == 1
		}
	case gomkv.FlagForcedElement:
		if h.currentTrack != nil {
			h.currentTrack.Forced = value == 1
		}
	case gomkv.FlagEnabledElement:
		if h.currentTrack != nil {
			h.currentTrack.Enabled = value == 1
		}
	case gomkv.PixelWidthElement:
		if h.currentTrack != nil && h.currentTrack.Video != nil {
			h.currentTrack.Video.PixelWidth = uint64(value)
		}
	case gomkv.PixelHeightElement:
		if h.currentTrack != nil && h.currentTrack.Video != nil {
			h.currentTrack.Video.PixelHeight = uint64(value)
		}
	case gomkv.ChannelsElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.Channels = uint64(value)
		}
	case gomkv.BitDepthElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.BitDepth = uint64(value)
		}
	case gomkv.ChapterTimeStartElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.Start = float64(value) * float64(h.mp.timecodeScale) / 1e9
		}
	case gomkv.ChapterTimeEndElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.End = float64(value) * float64(h.mp.timecodeScale) / 1e9
		}
	case gomkv.ChapterUIDElement:
		if h.inChapterAtom && h.currentChapter != nil {
			h.currentChapter.UID = uint64(value)
		}
	case gomkv.FileUIDElement:
		if h.isAttachment && h.currentAttachment != nil {
			h.currentAttachment.UID = uint64(value)
		}
	}
	return nil
}

func (h *metadataHandler) HandleFloat(id gomkv.ElementID, value float64, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.DurationElement:
		if h.mp.info != nil {
			h.mp.info.Duration = value
		}
	case gomkv.SamplingFrequencyElement:
		if h.currentTrack != nil && h.currentTrack.Audio != nil {
			h.currentTrack.Audio.SamplingFrequency = value
		}
	}
	return nil
}

func (h *metadataHandler) HandleBinary(id gomkv.ElementID, value []byte, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.CodecPrivateElement:
		if h.currentTrack != nil {
			h.currentTrack.CodecPrivate = string(value)
			h.currentTrack.CodecPrivate = strings.ReplaceAll(h.currentTrack.CodecPrivate, "\r\n", "\n")
		}
	case gomkv.FileDataElement:
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
				// Fix missing fields
				track.Name = cmp.Or(track.Name, strings.ToUpper(track.Language), strings.ToUpper(track.LanguageIETF))
				track.Language = getLanguageCode(track)
				result.SubtitleTracks = append(result.SubtitleTracks, track)
			}
		}

		// Group subtitle tracks by duplicate name
		groups := lo.GroupBy(result.SubtitleTracks, func(t *TrackInfo) string {
			return t.Name
		})
		for _, group := range groups {
			for _, track := range group {
				track.Name = fmt.Sprintf("%s", track.Name)
				if track.Language == "" {
					track.Language = getLanguageCode(track)
				}
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
func (mp *MetadataParser) ExtractSubtitles(ctx context.Context, newReader io.ReadSeekCloser, offset int64, backoffBytes int64) (<-chan *SubtitleEvent, <-chan error, <-chan struct{}) {
	subtitleCh := make(chan *SubtitleEvent)
	errCh := make(chan error, 1)
	startedCh := make(chan struct{})

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

	// coordination between extraction goroutines
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
		defer util.HandlePanicInModuleThen("mkvparser/ExtractSubtitles", func() {
			closeChannels(fmt.Errorf("subtitle extraction goroutine panic"))
		})
		defer cancel() // Ensure context is cancelled when main goroutine exits
		defer mp.logger.Trace().Msgf("mkvparser: Subtitle extraction goroutine finished.")

		sampler := lo.ToPtr(mp.logger.Sample(&zerolog.BasicSampler{N: 500}))

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

		handler := &subtitleHandler{
			mp:                 mp,
			ctx:                extractCtx, // use extraction context instead of original context
			logger:             mp.logger,
			sampler:            sampler,
			subtitleCh:         subtitleCh,
			subtitleTracks:     subtitleTracks,
			timecodeScale:      mp.timecodeScale,
			clusterTime:        0,
			reader:             newReader,
			lastSubtitleEvents: make(map[uint64]*SubtitleEvent),
			startedCh:          startedCh,
		}

		// Parse the stream for subtitles
		err := gomkv.Parse(newReader, handler)
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "unexpected EOF") {
			//mp.logger.Error().Err(err).Msg("mkvparser: Unrecoverable error during subtitle stream parsing")
			closeChannels(err)
		} else {
			mp.logger.Debug().Err(err).Msg("mkvparser: Subtitle streaming completed successfully or with expected EOF.")
			closeChannels(nil)
		}
	}()

	return subtitleCh, errCh, startedCh
}

// Handler for subtitle extraction
type subtitleHandler struct {
	gomkv.DefaultHandler
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
	lastSubtitleEvents   map[uint64]*SubtitleEvent // Track last subtitle event per track for duration calculation
	startedCh            chan struct{}
	// BlockGroup handling
	inBlockGroup bool
	pendingBlock *pendingSubtitleBlock
}

// pendingSubtitleBlock holds block data until we have complete information
type pendingSubtitleBlock struct {
	trackNum    uint64
	timecode    int16
	data        []byte
	duration    uint64
	hasBlock    bool
	hasDuration bool
	headPos     int64
}

func (h *subtitleHandler) HandleMasterBegin(id gomkv.ElementID, info gomkv.ElementInfo) (bool, error) {
	switch id {
	case gomkv.SegmentElement:
		return true, nil
	case gomkv.ClusterElement:
		return true, nil
	case gomkv.BlockGroupElement:
		h.inBlockGroup = true
		headPos, _ := h.reader.Seek(0, io.SeekCurrent)
		h.pendingBlock = &pendingSubtitleBlock{
			headPos: headPos,
		}
		return true, nil
	}
	return false, nil
}

func (h *subtitleHandler) HandleMasterEnd(id gomkv.ElementID, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.BlockGroupElement:
		// Process the pending block if we have complete information
		if h.pendingBlock != nil && h.pendingBlock.hasBlock {
			// If we have duration from BlockDurationElement, use it; otherwise use track default
			if h.pendingBlock.hasDuration {
				h.processPendingBlock(h.pendingBlock.duration)
			} else {
				h.processPendingBlock(0) // Will fall back to track defaultDuration
			}
		}
		h.inBlockGroup = false
		h.pendingBlock = nil
		h.currentBlockDuration = 0
	}
	return nil
}

func (h *subtitleHandler) HandleInteger(id gomkv.ElementID, value int64, info gomkv.ElementInfo) error {
	if id == gomkv.TimecodeElement {
		h.clusterTime = uint64(value)
	} else if id == gomkv.BlockDurationElement {
		if h.inBlockGroup && h.pendingBlock != nil {
			h.pendingBlock.duration = uint64(value)
			h.pendingBlock.hasDuration = true
		} else {
			h.currentBlockDuration = uint64(value)
		}
	}
	return nil
}

func (h *subtitleHandler) processPendingBlock(blockDuration uint64) {
	if h.pendingBlock == nil || !h.pendingBlock.hasBlock {
		return
	}

	track, isSubtitle := h.subtitleTracks[h.pendingBlock.trackNum]
	if !isSubtitle {
		return
	}

	// PGS subtitles are not supported
	if track.CodecID == "S_HDMV/PGS" || getSubtitleTrackType(track.CodecID) == "unknown" {
		return
	}

	absoluteTimeScaled := h.clusterTime + uint64(h.pendingBlock.timecode)
	timestampNs := absoluteTimeScaled * h.timecodeScale
	milliseconds := float64(timestampNs) / 1e6

	// Calculate duration in milliseconds
	var duration float64
	if blockDuration > 0 {
		duration = float64(blockDuration*h.timecodeScale) / 1e6 // ms
	} else if track.defaultDuration > 0 {
		duration = float64(track.defaultDuration) / 1e6 // ms
	}

	h.processSubtitleData(h.pendingBlock.trackNum, track, h.pendingBlock.data, milliseconds, duration, h.pendingBlock.headPos)
}

func (h *subtitleHandler) processSubtitleData(trackNum uint64, track *TrackInfo, subtitleData []byte, milliseconds, duration float64, headPos int64) {
	if getSubtitleTrackType(track.CodecID) == "PGS" || getSubtitleTrackType(track.CodecID) == "unknown" {
		return
	}

	if track.contentEncodings != nil {
		if zr, err := zlib.NewReader(bytes.NewReader(subtitleData)); err == nil {
			if buf, err := io.ReadAll(zr); err == nil {
				subtitleData = buf
			}
			_ = zr.Close()
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
		HeadPos:     headPos,
	}

	// Handling for ASS/SSA format
	if track.CodecID == "S_TEXT/ASS" || track.CodecID == "S_TEXT/SSA" {
		values := strings.Split(initialText, ",")
		if len(values) < 9 {
			//h.logger.Warn().
			//	Str("text", initialText).
			//	Int("fields", len(values)).
			//	Msg("mkvparser: Invalid ASS/SSA subtitle format, not enough fields")
			return
		}

		// SSA_KEYS = ['readOrder', 'layer', 'style', 'name', 'marginL', 'marginR', 'marginV', 'effect', 'text']
		// For ASS: ignore readOrder (start from index 1), extract indices 1-7, text from index 8
		// For SSA: ignore readOrder and layer (start from index 2), extract indices 2-7, text from index 8

		startIndex := 1
		if track.CodecID == "S_TEXT/SSA" {
			startIndex = 2
		}

		// Map values to ExtraData based on SSA_KEYS array
		ssaKeys := []string{"readorder", "layer", "style", "name", "marginl", "marginr", "marginv", "effect"}

		for i := startIndex; i < 8 && i < len(values); i++ {
			if i < len(ssaKeys) {
				subtitleEvent.ExtraData[ssaKeys[i]] = values[i]
			}
		}

		// Text is everything from index 8 onwards
		if len(values) > 8 {
			text := strings.Join(values[8:], ",")
			subtitleEvent.Text = strings.TrimSpace(text)
		}
	} else if track.CodecID == "S_TEXT/UTF8" {
		// Convert UTF8 to ASS format
		subtitleEvent.Text = UTF8ToASSText(initialText)

		subtitleEvent.CodecID = "S_TEXT/ASS"
		subtitleEvent.ExtraData = make(map[string]string)
		subtitleEvent.ExtraData["readorder"] = "0"
		subtitleEvent.ExtraData["layer"] = "0"
		subtitleEvent.ExtraData["style"] = "Default"
		subtitleEvent.ExtraData["name"] = "Default"
		subtitleEvent.ExtraData["marginl"] = "0"
		subtitleEvent.ExtraData["marginr"] = "0"
	}

	// Update the subtitle event duration after potential ASS/SSA calculation
	subtitleEvent.Duration = duration

	// Handle previous subtitle event duration according to Matroska spec:
	// If a subtitle has no duration, it should be displayed until the next subtitle is encountered
	if lastEvent, exists := h.lastSubtitleEvents[trackNum]; exists {
		// If the previous event had no duration, calculate it based on the time difference
		if lastEvent.Duration == 0 {
			calculatedDuration := milliseconds - lastEvent.StartTime
			if calculatedDuration > 0 {
				// Create a copy of the last event with updated duration
				updatedLastEvent := *lastEvent
				updatedLastEvent.Duration = calculatedDuration

				h.sampler.Trace().
					Uint64("trackNum", trackNum).
					Float64("previousStartTime", lastEvent.StartTime).
					Float64("calculatedDuration", calculatedDuration).
					Str("previousText", lastEvent.Text).
					Msg("mkvparser: Updated previous subtitle event duration")

				// Send the updated previous event with calculated duration
				select {
				case h.subtitleCh <- &updatedLastEvent:
					// Successfully sent updated previous event
				case <-h.ctx.Done():
					// h.logger.Debug().Msg("mkvparser: Subtitle sending cancelled by context.")
					return
				}
			}
		}
	}

	// Store current event as the last event for this track
	// Create a copy to avoid potential issues with pointer references
	eventCopy := *subtitleEvent
	h.lastSubtitleEvents[trackNum] = &eventCopy

	h.sampler.Trace().
		Uint64("trackNum", trackNum).
		Float64("startTime", milliseconds).
		Float64("duration", duration).
		Str("codecId", track.CodecID).
		Str("text", subtitleEvent.Text).
		Interface("data", subtitleEvent.ExtraData).
		Msg("mkvparser: Subtitle event")

	// Only send the current subtitle event if it has a duration > 0
	// Events without duration will be held and sent when their duration is calculated by the next event
	if duration > 0 {
		select {
		case h.subtitleCh <- subtitleEvent:
			// Successfully sent
		case <-h.ctx.Done():
			// h.logger.Debug().Msg("mkvparser: Subtitle sending cancelled by context.")
			return
		}
	}
}

func (h *subtitleHandler) HandleBinary(id gomkv.ElementID, value []byte, info gomkv.ElementInfo) error {
	switch id {
	case gomkv.SimpleBlockElement, gomkv.BlockElement:
		if len(value) < 4 {
			return nil
		}

		trackNum := uint64(value[0] & 0x7F)
		track, isSubtitle := h.subtitleTracks[trackNum]
		if !isSubtitle {
			return nil
		}

		blockTimecode := int16(value[1])<<8 | int16(value[2])
		subtitleData := value[4:]

		if h.inBlockGroup && h.pendingBlock != nil {
			// Store block data for later processing when BlockGroup ends
			h.pendingBlock.trackNum = trackNum
			h.pendingBlock.timecode = blockTimecode
			h.pendingBlock.data = make([]byte, len(subtitleData))
			copy(h.pendingBlock.data, subtitleData)
			h.pendingBlock.hasBlock = true
		} else {
			// Process immediately for SimpleBlock or standalone Block
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

			// Get current position for SimpleBlock/standalone Block
			headPos, _ := h.reader.Seek(0, io.SeekCurrent)
			h.processSubtitleData(trackNum, track, subtitleData, milliseconds, duration, headPos)

			// Reset the block duration for the next block
			h.currentBlockDuration = 0
		}
	}
	return nil
}

// findNextClusterOffset searches for the Matroska Cluster ID in the ReadSeeker rs,
// starting from seekOffset. It returns the absolute file offset of the found Cluster ID,
// or an error. If found, the ReadSeeker's position is set to the start of the Cluster ID.
func findNextClusterOffset(rs io.ReadSeeker, seekOffset, backoffBytes int64) (int64, error) {

	// DEVNOTE: findNextClusterOffset is faster than findPrecedingOrCurrentClusterOffset
	// however it's not ideal so we'll offset the offset by 1MB to avoid missing a cluster
	//toRemove := int64(1 * 1024 * 1024) // 1MB
	if seekOffset > backoffBytes {
		seekOffset -= backoffBytes
	} else {
		seekOffset = 0
	}

	// Seek to the starting position
	absPosOfNextRead, err := rs.Seek(seekOffset, io.SeekStart)
	if err != nil {
		return -1, fmt.Errorf("initial seek to %d failed: %w", seekOffset, err)
	}

	mainBuf := make([]byte, clusterSearchChunkSize)
	searchBuf := make([]byte, (len(matroskaClusterID)-1)+clusterSearchChunkSize)

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

// findPrecedingOrCurrentClusterOffset searches for the Matroska Cluster ID in the ReadSeeker rs,
// looking for a cluster that starts at or before targetFileOffset. It searches backwards from targetFileOffset.
// It returns the absolute file offset of the found Cluster ID, or an error.
// If found, the ReadSeeker's position is set to the start of the Cluster ID.
func findPrecedingOrCurrentClusterOffset(rs io.ReadSeeker, targetFileOffset int64) (int64, error) {
	mainBuf := make([]byte, clusterSearchChunkSize)
	searchBuf := make([]byte, (len(matroskaClusterID)-1)+clusterSearchChunkSize)

	// Start from targetFileOffset and work backwards
	currentReadEndPos := targetFileOffset + int64(len(matroskaClusterID))
	lenOverlapCarried := 0

	for {
		// Calculate read position and size
		readStartPos := currentReadEndPos - clusterSearchChunkSize
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
		if (targetFileOffset - readStartPos) >= clusterSearchDepth {
			return -1, fmt.Errorf("cluster ID not found within search depth %dMB", clusterSearchDepth/1024/1024)
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
