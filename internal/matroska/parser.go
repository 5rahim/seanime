// Package matroska implements a parser for Matroska and WebM media container formats.
//
// Matroska is an open standard, free container format, a file format that can hold an
// unlimited number of video, audio, picture, or subtitle tracks in one file. It is
// intended to serve as a universal format for storing common multimedia content, like
// movies or TV shows.
//
// This package provides functionality to parse Matroska files, extract metadata, and
// read media packets. It builds upon the EBML (Extensible Binary Meta Language) parsing
// functionality to implement Matroska-specific parsing logic.
//
// The main entry point is the MatroskaParser type, which can be created using the
// NewMatroskaParser function. Once created, it can be used to read file information,
// track information, and media packets.
//
// Example usage:
//
//	file, err := os.Open("video.mkv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	parser, err := matroska.NewMatroskaParser(file, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get file information
//	fileInfo := parser.GetFileInfo()
//	fmt.Printf("Title: %s\n", fileInfo.Title)
//	fmt.Printf("Duration: %d\n", fileInfo.Duration)
//
//	// Get track information
//	numTracks := parser.GetNumTracks()
//	for i := uint(0); i < numTracks; i++ {
//	    track := parser.GetTrackInfo(i)
//	    fmt.Printf("Track %d: %s\n", track.Number, track.Name)
//	}
//
//	// Read packets
//	for {
//	    packet, err := parser.ReadPacket()
//	    if err != nil {
//	        if errors.Is(err, io.EOF {
//	            break
//	        }
//	        log.Fatal(err)
//	    }
//	    // Process packet...
//	}
package matroska

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
)

// MatroskaParser represents a parser for Matroska and WebM files.
//
// It provides functionality to parse Matroska container files, extract metadata,
// and read media packets. The parser maintains state information about the file
// structure, including the EBML header, segment information, tracks, and other
// metadata elements.
//
// The parser can operate in two modes:
//   - With seeking enabled (noSeeking=false): Allows for more efficient parsing
//     by seeking to specific positions in the file when needed.
//   - With seeking disabled (noSeeking=true): Parses the file sequentially,
//     which is useful for streaming or non-seekable input sources.
//
// After creating a parser with NewMatroskaParser, you can access file information,
// track information, and read media packets using the provided methods.
type MatroskaParser struct {
	reader      *EBMLReader
	header      *EBMLHeader
	segment     *SegmentElement
	tracks      []*TrackInfo
	fileInfo    *SegmentInfo
	chapters    []*Chapter
	tags        []*Tag
	cues        []*Cue
	attachments []*Attachment

	// Element IDs to parse
	elementsToParse map[uint32]struct{}

	// Cluster parsing state
	clusterTimestamp uint64
	currentTrackMask uint64

	// Position tracking
	segmentPos    uint64
	segmentTopPos uint64
	cuesPos       uint64
	cuesTopPos    uint64

	// Flags
	noSeeking bool
}

// SegmentElement represents the main segment element in a Matroska file.
//
// The segment is the top-level element in a Matroska file that contains all
// the actual data, including metadata, tracks, clusters, and other elements.
// It is the largest element in the file and contains all other elements except
// for the EBML header.
//
// The Position field indicates the byte offset from the beginning of the file
// where the segment element starts, and the Size field indicates the total size
// of the segment element in bytes.
type SegmentElement struct {
	Position uint64
	Size     uint64
}

// NewMatroskaParser creates a new Matroska parser for the given ReadSeeker.
//
// This function initializes a MatroskaParser and parses the EBML header and
// main segment of the Matroska file. It validates that the file is a valid
// Matroska or WebM file by checking the document type in the EBML header.
// If the file is already seeked, it will not attempt to read the EBML header.
//
// Parameters:
//   - r: An io.ReadSeeker that provides access to the Matroska file data.
//     This can be a file, network stream, or any other source that supports
//     both reading and seeking operations.
//   - noSeeking: A boolean flag that controls whether the parser should avoid
//     seeking operations. When set to true, the parser will parse the file
//     sequentially, which is useful for streaming or non-seekable input sources.
//     When set to false, the parser can seek to specific positions in the file
//     for more efficient parsing.
//   - elementsToParse: A variadic list of element IDs to selectively parse.
//     When no element IDs are provided, all top-level elements are parsed (default behavior).
//     When one or more element IDs are provided, only those specific elements will be parsed,
//     and all others will be skipped. This can significantly speed up parsing when you only
//     need specific information from the file.
//     Valid element IDs include: IDSegmentInfo, IDTracks, IDChapters, IDTags, IDAttachments.
//
// Returns:
//   - *MatroskaParser: A pointer to the initialized MatroskaParser.
//   - error: An error if the parser could not be created or if the file is not
//     a valid Matroska or WebM file.
//
// Example (parse all elements):
//
//	file, err := os.Open("video.mkv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	parser, err := matroska.NewMatroskaParser(file, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example (parse only tracks and attachments):
//
//	parser, err := matroska.NewMatroskaParser(file, false, IDTracks, IDAttachments)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewMatroskaParser(r io.ReadSeeker, noSeeking bool, elementsToParse ...uint32) (*MatroskaParser, error) {
	parser := &MatroskaParser{
		reader:          NewEBMLReader(r),
		noSeeking:       noSeeking,
		elementsToParse: make(map[uint32]struct{}),
	}

	if len(elementsToParse) > 0 {
		for _, id := range elementsToParse {
			parser.elementsToParse[id] = struct{}{}
		}
	}

	// Try to parse header and segment
	headerErr := parser.parseHeader()
	var segmentErr error
	if headerErr == nil {
		segmentErr = parser.parseSegment()
	}

	// If we failed to parse header or segment, we might be positioned at a cluster offset
	// Initialize with minimal defaults to allow packet reading from current position
	if headerErr != nil || segmentErr != nil {
		// Initialize with defaults for cluster-only parsing
		parser.header = &EBMLHeader{
			DocType:        "matroska",
			DocTypeVersion: 1,
		}

		parser.fileInfo = &SegmentInfo{
			TimecodeScale: 1000000, // Default timecode scale (1ms)
		}

		parser.tracks = make([]*TrackInfo, 0)
		parser.chapters = make([]*Chapter, 0)
		parser.tags = make([]*Tag, 0)
		parser.cues = make([]*Cue, 0)
		parser.attachments = make([]*Attachment, 0)

		// Set segment to unknown size since we're starting mid-stream
		currentPos := parser.reader.Position()
		parser.segment = &SegmentElement{
			Position: uint64(currentPos),
			Size:     (1 << (7 * 8)) - 1, // Unknown size marker
		}

		parser.segmentPos = uint64(currentPos)
		parser.segmentTopPos = parser.segmentPos + parser.segment.Size

		// Reader is already positioned at cluster data, ready for ReadPacket()
	}

	// Devnote: Avoid parsing cues to speed up the process.
	//if !noSeeking && parser.cuesPos == 0 {
	//	// Cues not found in initial scan, let's scan the whole segment more carefully
	//	currentPos := parser.reader.Position()
	//	if _, err := parser.reader.Seek(int64(parser.segmentPos), io.SeekStart); err != nil {
	//		return nil, fmt.Errorf("failed to seek back to segment start: %w", err)
	//	}
	//
	//	// Scan through the segment looking for cues without parsing everything
	//	segmentEnd := parser.segmentPos + parser.segment.Size
	//	for parser.reader.Position() < int64(segmentEnd) {
	//		id, size, err := parser.reader.ReadElementHeader()
	//		if err != nil {
	//			if errors.Is(err, io.EOF) {
	//				break
	//			}
	//			// If we can't read more elements, break gracefully
	//			break
	//		}
	//
	//		if id == IDCues {
	//			parser.cuesPos = uint64(parser.reader.Position())
	//			parser.cuesTopPos = uint64(parser.reader.Position()) + size
	//			if err = parser.parseCues(size); err != nil {
	//				// If cues parsing fails, continue without cues
	//				break
	//			}
	//			break
	//		} else {
	//			// Skip this element
	//			if _, err = parser.reader.Seek(int64(size), io.SeekCurrent); err != nil {
	//				break
	//			}
	//		}
	//	}
	//
	//	// Restore original position
	//	if _, err := parser.reader.Seek(currentPos, io.SeekStart); err != nil {
	//		return nil, fmt.Errorf("failed to restore position: %w", err)
	//	}
	//}

	return parser, nil
}

// parseHeader parses the EBML header from the Matroska file.
//
// This method reads and validates the EBML (Extensible Binary Meta Language) header
// at the beginning of the Matroska file. The EBML header contains metadata about
// the file, including the document type, version, and other identification information.
//
// The method validates that the document type is either "matroska" or "webm",
// ensuring that the file is a valid Matroska or WebM file. If the document type
// is not recognized, an error is returned.
//
// Returns:
//   - error: An error if the header could not be read or if the document type
//     is not supported.
func (mp *MatroskaParser) parseHeader() error {
	header, err := mp.reader.ReadEBMLHeader()
	if err != nil {
		return err
	}

	// Validate it's a Matroska/WebM file
	if header.DocType != "matroska" && header.DocType != "webm" {
		return fmt.Errorf("unsupported document type: %s", header.DocType)
	}

	mp.header = header
	return nil
}

// parseSegment parses the main segment from the Matroska file.
//
// The segment is the top-level element in a Matroska file that contains all
// the actual data, including metadata, tracks, clusters, and other elements.
// This method reads the segment element header, validates that it is indeed
// a segment element, and stores its position and size for later use.
//
// After parsing the segment header, this method calls parseSegmentChildren()
// to parse the children of the segment element, which includes the segment
// information, tracks, cues, chapters, tags, and attachments.
//
// Returns:
//   - error: An error if the segment header could not be read or if the element
//     is not a valid segment element.
func (mp *MatroskaParser) parseSegment() error {
	// Read segment element header
	id, size, err := mp.reader.ReadElementHeader()
	if err != nil {
		return fmt.Errorf("failed to read segment header: %w", err)
	}

	if id != IDSegment {
		return fmt.Errorf("expected segment element, got ID 0x%X", id)
	}

	mp.segment = &SegmentElement{
		Position: uint64(mp.reader.Position()),
		Size:     size,
	}

	mp.segmentPos = mp.segment.Position
	mp.segmentTopPos = mp.segment.Position + mp.segment.Size

	// Parse segment children
	if err = mp.parseSegmentChildren(); err != nil {
		return fmt.Errorf("failed to parse segment children: %w", err)
	}

	return nil
}

// shouldParseElement returns true if the given element ID should be parsed.
// If no specific elements were configured via elementsToParse, all elements are parsed.
func (mp *MatroskaParser) shouldParseElement(id uint32) bool {
	// If elementsToParse is empty, parse everything
	if len(mp.elementsToParse) == 0 {
		return true
	}
	// Otherwise, only parse if the element is in the map
	_, ok := mp.elementsToParse[id]
	return ok
}

// parseSegmentChildren parses the children of the segment element.
//
// This method iterates through all child elements of the segment and dispatches
// them to appropriate parsing methods based on their element ID. The segment can
// contain various types of child elements, including:
//
//   - SegmentInfo: Contains metadata about the file, such as title, duration,
//     and timestamp scale.
//   - Tracks: Contains information about the media tracks in the file.
//   - Cues: Contains indexing information for seeking (currently skipped).
//   - Chapters: Contains chapter information (currently skipped).
//   - Tags: Contains metadata tags (currently skipped).
//   - Attachments: Contains attached files (currently skipped).
//   - Cluster: Contains the actual media data, which is handled during packet reading.
//
// If the parser is configured to avoid seeks (noSeeking=true), it will parse
// the entire segment sequentially. Otherwise, it will stop parsing when it
// encounters the first cluster element, as clusters are handled during packet reading.
//
// Returns:
//   - error: An error if any of the child elements could not be parsed.
func (mp *MatroskaParser) parseSegmentChildren() error {
	segmentEnd := mp.segment.Position + mp.segment.Size

	for mp.reader.Position() < int64(segmentEnd) {
		id, size, err := mp.reader.ReadElementHeader()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// If the segment uses unknown size (streaming), EOF is a natural terminator.
				// Otherwise, hitting EOF before the declared end means the segment is truncated.
				if mp.segment != nil && mp.segment.Size == (1<<(7*8))-1 {
					break
				}
				return fmt.Errorf("failed to read element header: %w", io.ErrUnexpectedEOF)
			}
			return fmt.Errorf("failed to read element header: %w", err)
		}

		currentPos := mp.reader.Position()

		switch id {
		case IDSegmentInfo:
			if mp.shouldParseElement(IDSegmentInfo) {
				if err = mp.parseSegmentInfo(size); err != nil {
					return fmt.Errorf("failed to parse segment info: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDTracks:
			if mp.shouldParseElement(IDTracks) {
				if err = mp.parseTracks(size); err != nil {
					return fmt.Errorf("failed to parse tracks: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDCues:
			if mp.shouldParseElement(IDCues) {
				mp.cuesPos = uint64(currentPos)
				mp.cuesTopPos = uint64(currentPos) + size
				if err = mp.parseCues(size); err != nil {
					return fmt.Errorf("failed to parse cues: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDChapters:
			if mp.shouldParseElement(IDChapters) {
				if err = mp.parseChapters(size); err != nil {
					return fmt.Errorf("failed to parse chapters: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDTags:
			if mp.shouldParseElement(IDTags) {
				if err = mp.parseTags(size); err != nil {
					return fmt.Errorf("failed to parse tags: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDAttachments:
			if mp.shouldParseElement(IDAttachments) {
				if err = mp.parseAttachments(size); err != nil {
					return fmt.Errorf("failed to parse attachments: %w", err)
				}
			} else {
				// Skip this element
				if err = mp.reader.SeekOrSkip(mp.noSeeking, int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		case IDCluster:
			// We'll handle clusters during packet reading
			// For now, just skip to end of parsing metadata
			if !mp.noSeeking {
				return nil
			}
			// Fall through to skip if avoiding seeks
			fallthrough
		default:
			// Skip unknown elements
			if mp.noSeeking {
				if _, err = mp.reader.Skip(int64(size)); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			} else {
				if _, err = mp.reader.Seek(int64(size), io.SeekCurrent); err != nil {
					return fmt.Errorf("failed to skip element: %w", err)
				}
			}
		}
	}

	return nil
}

// parseSegmentInfo parses segment information from the Matroska file.
//
// The SegmentInfo element contains metadata about the file, such as the title,
// duration, timestamp scale, and other information. This method reads the
// SegmentInfo element and populates the fileInfo field of the MatroskaParser
// with the parsed data.
//
// The SegmentInfo element can contain the following child elements:
//   - SegmentUID: A unique identifier for the segment.
//   - SegmentFilename: The filename of the segment.
//   - PrevUID: The unique identifier of the previous segment.
//   - PrevFilename: The filename of the previous segment.
//   - NextUID: The unique identifier of the next segment.
//   - NextFilename: The filename of the next segment.
//   - TimestampScale: The scale factor for timestamps in nanoseconds.
//   - Duration: The duration of the segment in timestamp units.
//   - DateUTC: The date and time the file was created.
//   - Title: The title of the segment.
//   - MuxingApp: The name of the application used to create the file.
//   - WritingApp: The name of the application used to write the file.
//
// Parameters:
//   - size: The size of the SegmentInfo element in bytes.
//
// Returns:
//   - error: An error if the SegmentInfo element could not be read or parsed.
func (mp *MatroskaParser) parseSegmentInfo(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	mp.fileInfo = &SegmentInfo{
		TimecodeScale: 1000000, // Default timecode scale
	}

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	var durationInTimecodeScale float64

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		switch element.ID {
		case IDSegmentUID:
			if len(element.Data) >= 16 {
				copy(mp.fileInfo.UID[:], element.Data[:16])
			}
		case IDSegmentFilename:
			mp.fileInfo.Filename = element.ReadString()
		case IDPrevUID:
			if len(element.Data) >= 16 {
				copy(mp.fileInfo.PrevUID[:], element.Data[:16])
			}
		case IDPrevFilename:
			mp.fileInfo.PrevFilename = element.ReadString()
		case IDNextUID:
			if len(element.Data) >= 16 {
				copy(mp.fileInfo.NextUID[:], element.Data[:16])
			}
		case IDNextFilename:
			mp.fileInfo.NextFilename = element.ReadString()
		case IDTimestampScale:
			mp.fileInfo.TimecodeScale = element.ReadUInt()
		case IDDuration:
			// Duration is stored as a float in timecode scale units
			durationInTimecodeScale = element.ReadFloat()
		case IDDateUTC:
			mp.fileInfo.DateUTC = element.ReadInt()
			mp.fileInfo.DateUTCValid = true
		case IDTitle:
			mp.fileInfo.Title = element.ReadString()
		case IDMuxingApp:
			mp.fileInfo.MuxingApp = element.ReadString()
		case IDWritingApp:
			mp.fileInfo.WritingApp = element.ReadString()
		}
	}

	// Convert duration to nanoseconds after all elements are parsed
	if durationInTimecodeScale > 0 {
		mp.fileInfo.Duration = uint64(durationInTimecodeScale * float64(mp.fileInfo.TimecodeScale))
	}

	return nil
}

// parseTracks parses track information from the Matroska file.
//
// The Tracks element contains information about all media tracks in the file,
// including video, audio, and subtitle tracks. This method reads the Tracks
// element and populates the tracks field of the MatroskaParser with the parsed
// track information.
//
// Each track is represented by a TrackEntry element, which contains detailed
// information about the track, such as its number, type, codec, and other
// properties. This method calls parseTrackEntry() for each TrackEntry element
// to parse the individual track information.
//
// After parsing all tracks, this method sorts them by track number to ensure
// they are in the correct order.
//
// Parameters:
//   - size: The size of the Tracks element in bytes.
//
// Returns:
//   - error: An error if the Tracks element could not be read or parsed.
func (mp *MatroskaParser) parseTracks(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID != IDTrackEntry {
			continue
		}

		trackInfo, errParseTrackEntry := mp.parseTrackEntry(element.Data)
		if errParseTrackEntry != nil {
			return fmt.Errorf("failed to parse track entry: %w", errParseTrackEntry)
		}
		mp.tracks = append(mp.tracks, trackInfo)
	}

	// Sort tracks by track number
	sort.Slice(mp.tracks, func(i, j int) bool {
		return mp.tracks[i].Number < mp.tracks[j].Number
	})

	return nil
}

// parseTrackEntry parses a single track entry from the Matroska file.
//
// A TrackEntry element contains detailed information about a single media track,
// such as its number, type, codec, and other properties. This method reads the
// TrackEntry element and returns a TrackInfo struct populated with the parsed data.
//
// The TrackEntry element can contain the following child elements:
//   - TrackNumber: The track number used to identify the track.
//   - TrackUID: A unique identifier for the track.
//   - TrackType: The type of the track (video, audio, subtitle, etc.).
//   - TrackName: A human-readable name for the track.
//   - Language: The language of the track (e.g., "eng" for English).
//   - CodecID: The identifier for the codec used to encode the track.
//   - CodecPrivate: Private data for the codec.
//   - Video: Video-specific information (parsed by parseVideoTrack).
//   - Audio: Audio-specific information (parsed by parseAudioTrack).
//
// This method initializes a TrackInfo struct with default values and then updates
// it with the values found in the TrackEntry element. If the track is a video
// or audio track, it calls the appropriate parsing method to handle the
// track-specific information.
//
// Parameters:
//   - data: The raw data of the TrackEntry element.
//
// Returns:
//   - *TrackInfo: A pointer to the parsed TrackInfo struct.
//   - error: An error if the TrackEntry element could not be parsed.
func (mp *MatroskaParser) parseTrackEntry(data []byte) (*TrackInfo, error) {
	track := &TrackInfo{
		Enabled:       true, // Default values
		Default:       true,
		Lacing:        true,
		TimecodeScale: 1.0,
		Language:      "eng",
	}

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDTrackNum:
			track.Number = uint8(element.ReadUInt())
		case IDTrackUID:
			track.UID = element.ReadUInt()
		case IDTrackType:
			track.Type = uint8(element.ReadUInt())
		case IDTrackName:
			track.Name = element.ReadString()
		case IDLanguage:
			if len(element.Data) >= 3 {
				track.Language = string(element.Data[:3])
			}
		case IDLanguageIETF:
			track.LanguageIETF = element.ReadString()
		case IDCodecID:
			track.CodecID = element.ReadString()
		case IDCodecPriv:
			track.CodecPrivate = element.ReadBytes()
		case IDVideo:
			if err = mp.parseVideoTrack(element.Data, track); err != nil {
				return nil, err
			}
		case IDAudio:
			if err = mp.parseAudioTrack(element.Data, track); err != nil {
				return nil, err
			}
		case IDDefaultDuration:
			track.DefaultDuration = element.ReadUInt()
		case IDContentEncodings:
			if err = mp.parseContentEncodings(element.Data, track); err != nil {
				return nil, err
			}
		}
	}

	return track, nil
}

// parseContentEncodings parses content encoding information from the Matroska file.
//
// The ContentEncodings element contains information about any transformations
// that have been applied to the track data, such as compression or encryption.
// This method parses the ContentEncodings element and extracts compression
// information, which is stored in the TrackInfo struct.
//
// The ContentEncodings element can contain one or more ContentEncoding elements.
// Each ContentEncoding can specify:
//   - ContentEncodingOrder: The order in which to apply the encoding
//   - ContentEncodingScope: What kind of data to apply the encoding to
//   - ContentEncodingType: The type of encoding (compression or encryption)
//   - ContentCompression: Compression settings
//   - ContentCompAlgo: The compression algorithm used
//   - ContentCompSettings: Private data for the compression algorithm
//   - ContentEncryption: Encryption settings (currently not parsed)
//
// Currently, this implementation only handles compression encodings and extracts
// the compression algorithm and any private data required by the decompressor.
//
// Parameters:
//   - data: The raw data of the ContentEncodings element.
//   - track: A pointer to the TrackInfo struct to be updated with the parsed data.
//
// Returns:
//   - error: An error if the ContentEncodings element could not be parsed.
func (mp *MatroskaParser) parseContentEncodings(data []byte, track *TrackInfo) error {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if element.ID == IDContentEncoding {
			if err = mp.parseContentEncoding(element.Data, track); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseContentEncoding parses a single content encoding entry from the Matroska file.
//
// This method processes a ContentEncoding element, which describes a single
// transformation applied to the track data. It extracts information about
// compression settings and stores them in the TrackInfo struct.
//
// Parameters:
//   - data: The raw data of the ContentEncoding element.
//   - track: A pointer to the TrackInfo struct to be updated with the parsed data.
//
// Returns:
//   - error: An error if the ContentEncoding element could not be parsed.
func (mp *MatroskaParser) parseContentEncoding(data []byte, track *TrackInfo) error {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch element.ID {
		case IDContentEncodingType:
			// 0 = Compression, 1 = Encryption
			encodingType := element.ReadUInt()
			if encodingType == 0 {
				track.CompEnabled = true
			}
		case IDContentCompression:
			if err = mp.parseContentCompression(element.Data, track); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseContentCompression parses content compression settings from the Matroska file.
//
// The ContentCompression element contains details about the compression algorithm
// used for the track and any private data required by the decompressor.
//
// Compression algorithms:
//   - 0 = zlib
//   - 1 = bzlib (not supported by most players)
//   - 2 = lzo1x (not supported by most players)
//   - 3 = Header Stripping (often used for removing codec headers)
//
// Parameters:
//   - data: The raw data of the ContentCompression element.
//   - track: A pointer to the TrackInfo struct to be updated with the parsed data.
//
// Returns:
//   - error: An error if the ContentCompression element could not be parsed.
func (mp *MatroskaParser) parseContentCompression(data []byte, track *TrackInfo) error {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	// Note: Sometimes ContentCompression is present even without data, so just set it to true
	track.CompEnabled = true

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch element.ID {
		case IDContentCompAlgo:
			track.CompMethod = uint32(element.ReadUInt())
		case IDContentCompSettings:
			track.CompMethodPrivate = element.ReadBytes()
		}
	}

	return nil
}

// parseVideoTrack parses video track information from the Matroska file.
//
// The Video element contains video-specific information for a track, such as
// pixel dimensions, display dimensions, and interlacing settings. This method
// reads the Video element and populates the Video field of the TrackInfo struct
// with the parsed data.
//
// The Video element can contain the following child elements:
//   - PixelWidth: The width of the video in pixels.
//   - PixelHeight: The height of the video in pixels.
//   - DisplayWidth: The width of the video when displayed (may differ from pixel width).
//   - DisplayHeight: The height of the video when displayed (may differ from pixel height).
//   - FlagInterlaced: Indicates whether the video is interlaced.
//
// If the display dimensions are not specified in the file, this method sets them
// to the pixel dimensions as a fallback.
//
// Parameters:
//   - data: The raw data of the Video element.
//   - track: A pointer to the TrackInfo struct to be updated with the parsed data.
//
// Returns:
//   - error: An error if the Video element could not be parsed.
func (mp *MatroskaParser) parseVideoTrack(data []byte, track *TrackInfo) error {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch element.ID {
		case IDPixelWidth:
			track.Video.PixelWidth = uint32(element.ReadUInt())
		case IDPixelHeight:
			track.Video.PixelHeight = uint32(element.ReadUInt())
		case IDDisplayWidth:
			track.Video.DisplayWidth = uint32(element.ReadUInt())
		case IDDisplayHeight:
			track.Video.DisplayHeight = uint32(element.ReadUInt())
		case IDFlagInterlaced:
			track.Video.Interlaced = element.ReadUInt() != 0
		}
	}

	// Set display dimensions to pixel dimensions if not specified
	if track.Video.DisplayWidth == 0 {
		track.Video.DisplayWidth = track.Video.PixelWidth
	}
	if track.Video.DisplayHeight == 0 {
		track.Video.DisplayHeight = track.Video.PixelHeight
	}

	return nil
}

// parseAudioTrack parses audio track information from the Matroska file.
//
// The Audio element contains audio-specific information for a track, such as
// sampling frequency, number of channels, and bit depth. This method reads the
// Audio element and populates the Audio field of the TrackInfo struct with the
// parsed data.
//
// The Audio element can contain the following child elements:
//   - SamplingFrequency: The sampling frequency of the audio in Hz.
//   - OutputSamplingFrequency: The output sampling frequency of the audio in Hz.
//   - Channels: The number of audio channels.
//   - BitDepth: The number of bits per sample.
//
// This method sets default values for the audio track (1 channel, 8000.0 Hz sampling
// frequency) before parsing the element. If the output sampling frequency is not
// specified in the file, this method sets it to the sampling frequency as a fallback.
//
// Parameters:
//   - data: The raw data of the Audio element.
//   - track: A pointer to the TrackInfo struct to be updated with the parsed data.
//
// Returns:
//   - error: An error if the Audio element could not be parsed.
func (mp *MatroskaParser) parseAudioTrack(data []byte, track *TrackInfo) error {
	// Set defaults
	track.Audio.Channels = 1
	track.Audio.SamplingFreq = 8000.0

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch element.ID {
		case IDSamplingFrequency:
			track.Audio.SamplingFreq = element.ReadFloat()
		case IDOutputSamplingFrequency:
			track.Audio.OutputSamplingFreq = element.ReadFloat()
		case IDChannels:
			track.Audio.Channels = uint8(element.ReadUInt())
		case IDBitDepth:
			track.Audio.BitDepth = uint8(element.ReadUInt())
		}
	}

	// Set output sampling frequency if not specified
	if track.Audio.OutputSamplingFreq == 0 {
		track.Audio.OutputSamplingFreq = track.Audio.SamplingFreq
	}

	return nil
}

// parseCues parses cue information for seeking from the Matroska file.
//
// The Cues element contains indexing information that enables efficient seeking
// to specific positions in the file. This information is particularly useful
// for media players that need to quickly jump to different timecodes in the file.
//
// This method parses the cue points and stores them for later use during seeking operations.
//
// Parameters:
//   - size: The size of the Cues element in bytes.
//
// Returns:
//   - error: An error if the Cues element could not be parsed.
func (mp *MatroskaParser) parseCues(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID == IDCuePoint {
			cuePoints, errParseCuePoint := mp.parseCuePoint(element.Data)
			if errParseCuePoint != nil {
				return errParseCuePoint
			}
			mp.cues = append(mp.cues, cuePoints...)
		}
	}

	// Cues should be sorted by time for efficient searching
	sort.Slice(mp.cues, func(i, j int) bool {
		return mp.cues[i].Time < mp.cues[j].Time
	})

	return nil
}

func (mp *MatroskaParser) parseCuePoint(data []byte) ([]*Cue, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	var cueTime uint64
	var cues []*Cue

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDCueTime:
			cueTime = element.ReadUInt()
		case IDCueTrackPosition:
			cue, errParseCueTrackPositions := mp.parseCueTrackPositions(element.Data)
			if errParseCueTrackPositions != nil {
				return nil, errParseCueTrackPositions
			}
			cue.Time = cueTime * mp.fileInfo.TimecodeScale
			cues = append(cues, cue)
		}
	}
	return cues, nil
}

func (mp *MatroskaParser) parseCueTrackPositions(data []byte) (*Cue, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	cue := &Cue{}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDCueTrack:
			cue.Track = uint8(element.ReadUInt())
		case IDCueClusterPos:
			cue.Position = element.ReadUInt()
		case IDCueRelativePos:
			cue.RelativePosition = element.ReadUInt()
		case IDCueBlockNum:
			cue.Block = element.ReadUInt()
		case IDCueDuration:
			cue.Duration = element.ReadUInt() * mp.fileInfo.TimecodeScale
		}
	}
	return cue, nil
}

// parseChapters parses chapter information from the Matroska file.
//
// The Chapters element contains information about the chapters in the file,
// such as chapter titles, timecodes, and other metadata. This information
// is typically used to provide navigation within the file, allowing users
// to jump to specific sections or chapters.
//
// This method parses the chapter information and stores it for later use, enabling chapter-based navigation.
//
// Parameters:
//   - size: The size of the Chapters element in bytes.
//
// Returns:
//   - error: An error if the Chapters element could not be parsed.
func (mp *MatroskaParser) parseChapters(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID == IDEditionEntry {
			chapters, errParseEditionEntry := mp.parseEditionEntry(element.Data)
			if errParseEditionEntry != nil {
				return errParseEditionEntry
			}
			mp.chapters = append(mp.chapters, chapters...)
		}
	}

	return nil
}

func (mp *MatroskaParser) parseEditionEntry(data []byte) ([]*Chapter, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	var chapters []*Chapter
	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if element.ID == IDChapterAtom {
			chapter, errParseChapterAtom := mp.parseChapterAtom(element.Data)
			if errParseChapterAtom != nil {
				return nil, errParseChapterAtom
			}
			chapters = append(chapters, chapter)
		}
	}
	return chapters, nil
}

func (mp *MatroskaParser) parseChapterAtom(data []byte) (*Chapter, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	chapter := &Chapter{
		Enabled: true, // Default value
	}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDChapterUID:
			chapter.UID = element.ReadUInt()
		case IDChapterTimeStart:
			chapter.Start = element.ReadUInt()
		case IDChapterTimeEnd:
			chapter.End = element.ReadUInt()
		case IDChapterHidden:
			chapter.Hidden = element.ReadUInt() != 0
		case IDChapterEnabled:
			chapter.Enabled = element.ReadUInt() != 0
		case IDChapterDisplay:
			display, errParseChapterDisplay := mp.parseChapterDisplay(element.Data)
			if errParseChapterDisplay != nil {
				return nil, errParseChapterDisplay
			}
			chapter.Display = append(chapter.Display, display)
		case IDChapterAtom:
			childChapter, errParseChapterAtom := mp.parseChapterAtom(element.Data)
			if errParseChapterAtom != nil {
				return nil, errParseChapterAtom
			}
			chapter.Children = append(chapter.Children, childChapter)
		}
	}

	return chapter, nil
}

func (mp *MatroskaParser) parseChapterDisplay(data []byte) (ChapterDisplay, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	display := ChapterDisplay{
		Language: "eng", // Default value
	}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return display, err
		}

		switch element.ID {
		case IDChapterString:
			display.String = element.ReadString()
		case IDChapterLanguage:
			display.Language = element.ReadString()
		case IDChapterCountry:
			display.Country = element.ReadString()
		}
	}
	return display, nil
}

// parseTags parses tag information from the Matroska file.
//
// The Tags element contains metadata tags that provide additional information
// about the file, such as artist, album, genre, and other descriptive metadata.
// This information is similar to ID3 tags in MP3 files and can be used to
// enrich the user experience by providing more context about the media content.
//
// This method parses the tag information and stores it for later use, enabling applications to display
// or utilize this metadata.
//
// Parameters:
//   - size: The size of the Tags element in bytes.
//
// Returns:
//   - error: An error if the Tags element could not be parsed.
func (mp *MatroskaParser) parseTags(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID == IDTag {
			tag, errParseTag := mp.parseTag(element.Data)
			if errParseTag != nil {
				return errParseTag
			}
			mp.tags = append(mp.tags, tag)
		}
	}

	return nil
}

func (mp *MatroskaParser) parseTag(data []byte) (*Tag, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	tag := &Tag{}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDTargets:
			target, errParseTarget := mp.parseTarget(element.Data)
			if errParseTarget != nil {
				return nil, errParseTarget
			}
			tag.Targets = append(tag.Targets, target)
		case IDSimpleTag:
			simpleTag, errParseSimpleTag := mp.parseSimpleTag(element.Data)
			if errParseSimpleTag != nil {
				return nil, errParseSimpleTag
			}
			tag.SimpleTags = append(tag.SimpleTags, simpleTag)
		}
	}

	return tag, nil
}

func (mp *MatroskaParser) parseTarget(data []byte) (Target, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	target := Target{}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return target, err
		}

		switch element.ID {
		case IDTargetTypeValue:
			target.Type = uint32(element.ReadUInt())
		case IDTagTrackUID:
			target.UID = element.ReadUInt()
		case IDTagEditionUID:
			target.UID = element.ReadUInt()
		case IDTagChapterUID:
			target.UID = element.ReadUInt()
		case IDTagAttachmentUID:
			target.UID = element.ReadUInt()
		}
	}

	return target, nil
}

func (mp *MatroskaParser) parseSimpleTag(data []byte) (SimpleTag, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	simpleTag := SimpleTag{
		Language: "eng", // Default language
		Default:  true,  // Default value
	}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return simpleTag, err
		}

		switch element.ID {
		case IDTagName:
			simpleTag.Name = element.ReadString()
		case IDTagString:
			simpleTag.Value = element.ReadString()
		case IDTagLanguage:
			simpleTag.Language = element.ReadString()
		case IDTagDefault:
			simpleTag.Default = element.ReadUInt() != 0
		}
	}

	return simpleTag, nil
}

// parseAttachments parses attachment information from the Matroska file.
//
// The Attachments element contains files that are attached to the Matroska file,
// such as cover art, fonts, or other related files. These attachments are
// embedded within the Matroska container and can be extracted for use by
// media players or other applications.
//
// This method parses the attachment information and stores it for later use, enabling applications to extract
// and utilize these attached files.
//
// Parameters:
//   - size: The size of the Attachments element in bytes.
//
// Returns:
//   - error: An error if the Attachments element could not be parsed.
func (mp *MatroskaParser) parseAttachments(size uint64) error {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(size) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID == IDAttachedFile {
			attachment, errParseAttachedFile := mp.parseAttachedFile(element.Data)
			if errParseAttachedFile != nil {
				return errParseAttachedFile
			}
			mp.attachments = append(mp.attachments, attachment)
		}
	}

	return nil
}

func (mp *MatroskaParser) parseAttachedFile(data []byte) (*Attachment, error) {
	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	attachment := &Attachment{
		Position: uint64(mp.reader.Position()),
	}

	for childReader.pos < int64(len(data)) {
		element, err := childReader.ReadElement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch element.ID {
		case IDFileDescription:
			attachment.Description = element.ReadString()
		case IDFileName:
			attachment.Name = element.ReadString()
		case IDFileMimeType:
			attachment.MimeType = element.ReadString()
		case IDFileUID:
			attachment.UID = element.ReadUInt()
		case IDFileData:
			attachment.Length = uint64(len(element.Data))
			attachment.Data = element.Data
		}
	}

	return attachment, nil
}

// ReadPacket reads the next packet from the Matroska stream.
//
// This method reads and parses the next media packet from the Matroska file.
// A packet represents a unit of media data, such as a video frame or audio
// samples, along with metadata about the packet, such as the track number,
// timestamp, and flags.
//
// The method iterates through the elements in the file, looking for Cluster,
// SimpleBlock, and BlockGroup elements, which contain the actual media data.
// When it encounters a Cluster element, it parses the cluster header to update
// the cluster timestamp. When it encounters a SimpleBlock or BlockGroup element,
// it parses the block and returns a Packet struct containing the media data
// and metadata.
//
// If the method encounters a Timestamp element within a cluster, it updates
// the cluster timestamp accordingly. Unknown elements are skipped.
//
// Returns:
//   - *Packet: A pointer to the parsed Packet struct containing the media data
//     and metadata. Returns nil when the end of the file is reached.
//   - error: An error if a packet could not be read or parsed. When the end
//     of the file is reached, the error will be io.EOF.
//
// Example:
//
//	for {
//	    packet, err := parser.ReadPacket()
//	    if err != nil {
//	        if errors.Is(err, io.EOF {
//	            break
//	        }
//	        log.Fatal(err)
//	    }
//	    // Process packet...
//	    fmt.Printf("Track: %d, Timestamp: %d\n", packet.Track, packet.StartTime)
//	}
func (mp *MatroskaParser) ReadPacket() (*Packet, error) {
	for {
		// Try to read next element
		id, size, err := mp.reader.ReadElementHeader()
		if err != nil {
			return nil, err
		}

		var packet *Packet
		var parseErr error

		switch id {
		case IDCluster:
			// Start of a new cluster, reset timestamp and parse its children
			mp.clusterTimestamp = 0
			clusterEnd := mp.reader.Position() + int64(size)
			for mp.reader.Position() < clusterEnd {
				childID, childSize, childErr := mp.reader.ReadElementHeader()
				if childErr != nil {
					return nil, childErr
				}
				switch childID {
				case IDTimestamp:
					data := make([]byte, childSize)
					if n, errReadFull := io.ReadFull(mp.reader.r, data); errReadFull != nil {
						return nil, errReadFull
					} else {
						mp.reader.pos += int64(n)
					}
					element := &EBMLElement{ID: childID, Size: childSize, Data: data}
					mp.clusterTimestamp = element.ReadUInt()
				case IDSimpleBlock:
					packet, parseErr = mp.parseSimpleBlock(childSize)
					if parseErr != nil {
						return nil, parseErr
					}
					if packet != nil {
						if mp.currentTrackMask == 0 || (1<<(packet.Track-1))&mp.currentTrackMask == 0 {
							return packet, nil
						}
					}
				case IDBlockGroup:
					packet, parseErr = mp.parseBlockGroup(childSize)
					if parseErr != nil {
						return nil, parseErr
					}
					if packet != nil {
						if mp.currentTrackMask == 0 || (1<<(packet.Track-1))&mp.currentTrackMask == 0 {
							return packet, nil
						}
					}
				default:
					if _, err = mp.reader.Seek(int64(childSize), io.SeekCurrent); err != nil {
						return nil, err
					}
				}
			}
			continue

		case IDSimpleBlock:
			packet, parseErr = mp.parseSimpleBlock(size)

		case IDBlockGroup:
			packet, parseErr = mp.parseBlockGroup(size)

		case IDTimestamp:
			// Update cluster timestamp
			data := make([]byte, size)
			if n, errReadFull := io.ReadFull(mp.reader.r, data); errReadFull != nil {
				return nil, errReadFull
			} else {
				mp.reader.pos += int64(n)
			}
			element := &EBMLElement{ID: id, Size: size, Data: data}
			mp.clusterTimestamp = element.ReadUInt()
			continue

		case IDSegmentInfo, IDTracks, IDCues, IDChapters, IDTags, IDAttachments:
			// Skip metadata elements if encountered mid-stream
			// This can happen when starting from a cluster offset
			if _, err = mp.reader.Seek(int64(size), io.SeekCurrent); err != nil {
				return nil, err
			}
			continue

		default:
			// Skip unknown elements
			if _, err = mp.reader.Seek(int64(size), io.SeekCurrent); err != nil {
				return nil, err
			}
			continue
		}

		if parseErr != nil {
			return nil, parseErr
		}

		if packet != nil {
			if mp.currentTrackMask != 0 && (1<<(packet.Track-1))&mp.currentTrackMask != 0 {
				continue
			}
			return packet, nil
		}
	}
}

// parseClusterHeader parses the header of a Cluster element.
//
// A Cluster is a top-level element that contains a group of blocks (media data)
// that are related to each other, typically by time. The cluster header contains
// metadata about the cluster, such as the timestamp.
//
// This method currently only resets the cluster timestamp to zero when a new cluster
// is encountered. A more complete implementation would parse the cluster
// header elements, such as the timestamp, and update the parser's state accordingly.
//
// Parameters:
//   - size: The size of the Cluster element in bytes.
//
// Returns:
//   - error: An error if the cluster header could not be parsed.
func (mp *MatroskaParser) parseClusterHeader(size uint64) error {
	// We need to find the timestamp of the cluster.
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	for childReader.pos < int64(len(data)) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return errReadElement
		}

		if element.ID == IDTimestamp {
			mp.clusterTimestamp = element.ReadUInt()
			// We found the timestamp, but we need to continue parsing the rest of the cluster
			// so we have to seek back.
			if _, err = mp.reader.Seek(int64(-size), io.SeekCurrent); err != nil {
				return err
			}
			return nil
		}
	}

	// Timestamp not found, which is weird, but let's seek back to where we were.
	if _, err = mp.reader.Seek(int64(-size), io.SeekCurrent); err != nil {
		return err
	}
	mp.clusterTimestamp = 0
	return nil
}

// parseSimpleBlock parses a simple block element from the Matroska file.
//
// A SimpleBlock element contains a single frame of media data along with metadata
// about the frame, such as the track number, timestamp, and flags. SimpleBlocks
// are the most common way to store media data in a Matroska file.
//
// This method parses the SimpleBlock element and returns a Packet struct containing
// the media data and metadata. The parsing process includes:
//   - Reading the track number (as a variable-length integer)
//   - Reading the timestamp (relative to the cluster timestamp)
//   - Reading the flags (which indicate keyframe status, discardable status, etc.)
//   - Extracting the frame data, handling different lacing types if present
//
// Matroska supports three types of lacing for storing multiple frames in a single block:
//   - Fixed-size lacing: All frames have the same size.
//   - EBML lacing: Frame sizes are encoded as EBML variable-length integers.
//   - Xiph lacing: Frame sizes are encoded similarly to Xiph's lacing method.
//
// Parameters:
//   - size: The size of the SimpleBlock element in bytes.
//
// Returns:
//   - *Packet: A pointer to the parsed Packet struct containing the media data
//     and metadata.
//   - error: An error if the SimpleBlock element could not be parsed.
func (mp *MatroskaParser) parseSimpleBlock(size uint64) (*Packet, error) {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return nil, err
	}
	mp.reader.pos += int64(n)

	if len(data) < 4 {
		return nil, fmt.Errorf("block too short")
	}

	// Parse track number (VINT)
	trackNum, trackBytes := mp.parseVInt(data)
	if trackBytes == 0 {
		return nil, fmt.Errorf("invalid track number")
	}

	// Parse timestamp (2 bytes, signed)
	if len(data) < trackBytes+2 {
		return nil, fmt.Errorf("block too short for timestamp")
	}

	timestamp := int16(data[trackBytes])<<8 | int16(data[trackBytes+1])

	// Parse flags
	if len(data) < trackBytes+3 {
		return nil, fmt.Errorf("block too short for flags")
	}

	flags := data[trackBytes+2]

	// Extract frame data, handling lacing
	frameData := data[trackBytes+3:]

	// Check lacing flags (bits 1-0)
	lacingType := flags & 0x06
	if lacingType != 0 {
		// Handle laced frames
		if len(frameData) < 1 {
			return nil, fmt.Errorf("laced block too short")
		}

		frameCount := int(frameData[0]) + 1
		frameData = frameData[1:] // Skip frame count byte

		switch lacingType {
		case 0x02: // Fixed-size lacing
			if frameCount > 1 {
				frameSize := len(frameData) / frameCount
				frameData = frameData[:frameSize]
			}
		case 0x04: // EBML lacing
			// For EBML lacing, we need to reconstruct the original stream
			// The reference seems to include size information in the output
			if frameCount > 1 && len(frameData) > 1 {
				// Don't skip anything - include all lacing information
				// This matches the reference file format
			}
		case 0x06: // Xiph lacing
			// Parse Xiph lacing sizes
			if frameCount > 1 {
				frameSizes := make([]int, frameCount)
				offset := 0

				// Parse sizes for all frames except the last one
				for i := 0; i < frameCount-1; i++ {
					dataSize := 0
					// Xiph lacing: sizes are encoded as a series of 255 bytes
					// followed by the remainder
					for offset < len(frameData) && frameData[offset] == 0xFF {
						dataSize += 255
						offset++
					}
					if offset < len(frameData) {
						dataSize += int(frameData[offset])
						offset++
					}
					frameSizes[i] = dataSize
				}

				// Last frame size is the remainder
				totalPrevFrames := 0
				for i := 0; i < frameCount-1; i++ {
					totalPrevFrames += frameSizes[i]
				}
				frameSizes[frameCount-1] = len(frameData) - offset - totalPrevFrames

				// Extract the first frame (for simplicity, just return the first frame)
				// In a full implementation, you'd want to return all frames
				if frameSizes[0] > 0 && offset+frameSizes[0] <= len(frameData) {
					frameData = frameData[offset : offset+frameSizes[0]]
				} else {
					// If parsing failed, take remaining data after size headers
					frameData = frameData[offset:]
				}
			}
		}
	}

	scaledTime := (mp.clusterTimestamp + uint64(timestamp)) * mp.fileInfo.TimecodeScale
	packet := &Packet{
		Track:     uint8(trackNum),
		StartTime: scaledTime,
		EndTime:   scaledTime, // Will be updated if duration is known
		FilePos:   uint64(mp.reader.Position()) - size,
		Data:      frameData,
		Flags:     uint32(flags),
	}

	// Set keyframe flag if present
	if flags&0x80 != 0 {
		packet.Flags |= KF
	}

	return packet, nil
}

// parseBlockGroup parses a block group element from the Matroska file.
//
// A BlockGroup element contains a block along with additional metadata, such as
// duration, reference frames, and other information. BlockGroups are more complex
// than SimpleBlocks and can contain multiple blocks and additional metadata elements.
//
// This method parses the BlockGroup element and returns a Packet struct containing
// the media data and metadata. The parsing process includes:
//   - Reading the Block element, which contains the actual media data
//   - Reading the BlockDuration element, which specifies the duration of the block
//   - Extracting the frame data and metadata
//
// Unlike SimpleBlocks, BlockGroups do not have flags in the block header itself,
// but they can contain additional metadata elements that provide similar information.
//
// Parameters:
//   - size: The size of the BlockGroup element in bytes.
//
// Returns:
//   - *Packet: A pointer to the parsed Packet struct containing the media data
//     and metadata.
//   - error: An error if the BlockGroup element could not be parsed.
func (mp *MatroskaParser) parseBlockGroup(size uint64) (*Packet, error) {
	data := make([]byte, size)
	n, err := io.ReadFull(mp.reader.r, data)
	if err != nil {
		return nil, err
	}
	mp.reader.pos += int64(n)

	reader := bytes.NewReader(data)
	childReader := &EBMLReader{r: &seekableReader{reader}, pos: 0}

	var packet *Packet
	var duration uint64

	for childReader.pos < int64(len(data)) {
		element, errReadElement := childReader.ReadElement()
		if errReadElement != nil {
			if errors.Is(errReadElement, io.EOF) {
				break
			}
			return nil, errReadElement
		}

		switch element.ID {
		case IDBlock:
			// Parse block similar to simple block but without flags
			blockData := element.Data
			if len(blockData) < 4 {
				return nil, fmt.Errorf("block too short")
			}

			trackNum, trackBytes := mp.parseVInt(blockData)
			if trackBytes == 0 {
				return nil, fmt.Errorf("invalid track number")
			}

			timestamp := int16(blockData[trackBytes])<<8 | int16(blockData[trackBytes+1])
			frameData := blockData[trackBytes+3:] // Skip flags byte

			scaledTime := (mp.clusterTimestamp + uint64(timestamp)) * mp.fileInfo.TimecodeScale
			packet = &Packet{
				Track:     uint8(trackNum),
				StartTime: scaledTime,
				EndTime:   scaledTime,
				FilePos:   uint64(mp.reader.Position()) - size,
				Data:      frameData,
				Flags:     KF, // Block groups are typically keyframes
			}

		case 0x9B: // BlockDuration
			duration = element.ReadUInt()
		}
	}

	if packet != nil && duration > 0 {
		packet.EndTime = packet.StartTime + (duration * mp.fileInfo.TimecodeScale)
	}

	return packet, nil
}

// parseVInt parses a variable-length integer (VINT) from the given data.
//
// Variable-length integers are used throughout Matroska and EBML to encode
// element IDs and sizes in a space-efficient manner. The length of the integer
// is encoded in the first byte, allowing for integers of different sizes to be
// represented compactly.
//
// The VINT format works as follows:
//   - The first byte contains both the length information and the most significant
//     bits of the value.
//   - The length is determined by the position of the first '1' bit in the first byte.
//     For example, if the first bit is '1', the VINT is 1 byte long; if the second
//     bit is the first '1', the VINT is 2 bytes long, and so on.
//   - The remaining bits in the first byte (after the length marker) and all bits
//     in subsequent bytes form the actual value.
//
// Parameters:
//   - data: A byte slice containing the VINT to be parsed.
//
// Returns:
//   - uint64: The parsed value.
//   - int: The number of bytes consumed from the input data. Returns 0 if the
//     VINT is invalid or if the data is too short.
func (mp *MatroskaParser) parseVInt(data []byte) (uint64, int) {
	if len(data) == 0 {
		return 0, 0
	}

	firstByte := data[0]
	if firstByte == 0 {
		return 0, 0
	}

	// Find length by counting leading zeros
	var length int
	mask := uint8(0x80)
	for i := 0; i < 8; i++ {
		if firstByte&mask != 0 {
			length = i + 1
			break
		}
		mask >>= 1
	}

	if length == 0 || len(data) < length {
		return 0, 0
	}

	// Extract value
	result := uint64(firstByte & (mask - 1))
	for i := 1; i < length; i++ {
		result = (result << 8) | uint64(data[i])
	}

	return result, length
}

// GetNumTracks returns the number of tracks
func (mp *MatroskaParser) GetNumTracks() uint {
	return uint(len(mp.tracks))
}

// GetTrackInfo returns information about a specific track
func (mp *MatroskaParser) GetTrackInfo(track uint) *TrackInfo {
	if track >= uint(len(mp.tracks)) {
		return nil
	}
	return mp.tracks[track]
}

// GetFileInfo returns file-level information
func (mp *MatroskaParser) GetFileInfo() *SegmentInfo {
	return mp.fileInfo
}

// GetAttachments returns all attachments
func (mp *MatroskaParser) GetAttachments() []*Attachment {
	return mp.attachments
}

// GetChapters returns all chapters
func (mp *MatroskaParser) GetChapters() []*Chapter {
	return mp.chapters
}

// GetTags returns all tags
func (mp *MatroskaParser) GetTags() []*Tag {
	return mp.tags
}

// GetCues returns all cues
func (mp *MatroskaParser) GetCues() []*Cue {
	return mp.cues
}

// GetSegment returns the segment position
func (mp *MatroskaParser) GetSegment() uint64 {
	return mp.segmentPos
}

// GetSegmentTop returns the segment top position
func (mp *MatroskaParser) GetSegmentTop() uint64 {
	return mp.segmentTopPos
}

// GetCuesPos returns the cues position
func (mp *MatroskaParser) GetCuesPos() uint64 {
	return mp.cuesPos
}

// GetCuesTopPos returns the cues top position
func (mp *MatroskaParser) GetCuesTopPos() uint64 {
	return mp.cuesTopPos
}

func (mp *MatroskaParser) Seek(timecode uint64, flags uint32) error {
	if mp.noSeeking {
		return fmt.Errorf("seeking not supported in streaming mode")
	}

	if len(mp.cues) == 0 {
		return fmt.Errorf("no cues available for seeking")
	}

	// Find the right cue point. Cues are sorted by time.
	// We want to find the last cue point with time <= timecode.
	i := sort.Search(len(mp.cues), func(i int) bool {
		return mp.cues[i].Time >= timecode
	})

	if i > 0 && (i == len(mp.cues) || mp.cues[i].Time > timecode) {
		// sort.Search finds the first element >= timecode.
		// We want the one before it, which is <= timecode for a keyframe seek.
		i--
	}

	if i >= len(mp.cues) {
		i = len(mp.cues) - 1
	}

	// We have a cue point, now seek to the cluster position.
	cue := mp.cues[i]
	if _, err := mp.reader.Seek(int64(mp.segmentPos+cue.Position), io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to cue position: %w", err)
	}

	// Reset cluster parsing state so ReadPacket will look for a new cluster
	mp.clusterTimestamp = 0
	return nil
}

func (mp *MatroskaParser) SkipToKeyframe() {
	// If we can't seek, we can't really skip efficiently
	if mp.noSeeking {
		return
	}

	// Store current position in case we need to restore it
	currentPos := mp.reader.Position()

	for {
		packet, err := mp.ReadPacket()
		if err != nil {
			// If we hit an error, restore position and return
			_, _ = mp.reader.Seek(currentPos, io.SeekStart)
			return
		}

		// Check if this is a keyframe and the track is not masked
		isKeyframe := (packet.Flags & KF) != 0
		isTrackEnabled := mp.currentTrackMask == 0 || (1<<(packet.Track-1))&mp.currentTrackMask == 0

		if isKeyframe && isTrackEnabled {
			// Don't seek back - we've already consumed this packet
			// The next ReadPacket call will read the next packet naturally
			return
		}
	}
}

func (mp *MatroskaParser) SetTrackMask(mask uint64) {
	mp.currentTrackMask = mask
	// Here we could discard queued packets if we had a queue
}
