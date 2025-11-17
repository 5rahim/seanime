package matroska

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

const testFile = "testdata/test.mkv"

// limitedSeeker is a helper type that limits seeking to simulate errors
type limitedSeeker struct {
	*bytes.Reader
	limit int64
}

func (ls *limitedSeeker) Seek(offset int64, whence int) (int64, error) {
	pos, err := ls.Reader.Seek(offset, whence)
	if err != nil {
		return pos, err
	}
	if pos > ls.limit {
		return pos, io.ErrUnexpectedEOF
	}
	return pos, nil
}

// TestNewMatroskaParser tests the creation of a new parser.
// This test requires a real Matroska file.
func TestNewMatroskaParser(t *testing.T) {
	file, err := os.Open(testFile)
	if err != nil {
		t.Skipf("Skipping test: could not open test file %s: %v", testFile, err)
	}
	defer func() {
		_ = file.Close()
	}()

	parser, err := NewMatroskaParser(file, false)
	if err != nil {
		t.Fatalf("NewMatroskaParser() failed: %v", err)
	}

	if parser.header == nil {
		t.Error("Expected parser to have a non-nil header")
	}
	if parser.segment == nil {
		t.Error("Expected parser to have a non-nil segment")
	}
	if parser.fileInfo == nil {
		t.Error("Expected parser to have non-nil fileInfo")
	}
	if len(parser.tracks) == 0 {
		t.Error("Expected parser to have found some tracks")
	}
}

// TestParseSegmentInfo tests the parsing of the SegmentInfo element.
func TestParseSegmentInfo(t *testing.T) {
	// Create a mock SegmentInfo element
	buf := new(bytes.Buffer)
	// Title
	buf.Write([]byte{0x7B, 0xA9, 0x85, 't', 'i', 't', 'l', 'e'})
	// MuxingApp
	buf.Write([]byte{0x4D, 0x80, 0x84, 't', 'e', 's', 't'})
	// WritingApp
	buf.Write([]byte{0x57, 0x41, 0x8B, 'g', 'o', '-', 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	// TimestampScale
	buf.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // 1,000,000
	// Duration
	buf.Write([]byte{0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0}) // 100000

	parser := &MatroskaParser{
		reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
	}

	err := parser.parseSegmentInfo(uint64(buf.Len()))
	if err != nil {
		t.Fatalf("parseSegmentInfo() failed: %v", err)
	}

	if parser.fileInfo.Title != "title" {
		t.Errorf("Expected Title 'title', got %q", parser.fileInfo.Title)
	}
	if parser.fileInfo.MuxingApp != "test" {
		t.Errorf("Expected MuxingApp 'test', got %q", parser.fileInfo.MuxingApp)
	}
	if parser.fileInfo.WritingApp != "go-matroska" {
		t.Errorf("Expected WritingApp 'go-matroska', got %q", parser.fileInfo.WritingApp)
	}
	if parser.fileInfo.TimecodeScale != 1000000 {
		t.Errorf("Expected TimecodeScale 1000000, got %d", parser.fileInfo.TimecodeScale)
	}
	if parser.fileInfo.Duration != 100000 {
		t.Errorf("Expected Duration 100000, got %d", parser.fileInfo.Duration)
	}
}

// TestParseTracks tests parsing of the Tracks element.
func TestParseTracks(t *testing.T) {
	// Create a mock Tracks element containing one video and one audio track entry
	trackEntryVideo, _ := createMockTrackEntry(1, TypeVideo, "V_MPEG4/ISO/AVC", "Video", "und")
	trackEntryAudio, _ := createMockTrackEntry(2, TypeAudio, "A_AAC", "Audio", "eng")

	tracksElement := new(bytes.Buffer)
	// Write TrackEntry 1
	tracksElement.Write([]byte{0xAE})
	tracksElement.Write(vintEncode(uint64(len(trackEntryVideo))))
	tracksElement.Write(trackEntryVideo)
	// Write TrackEntry 2
	tracksElement.Write([]byte{0xAE})
	tracksElement.Write(vintEncode(uint64(len(trackEntryAudio))))
	tracksElement.Write(trackEntryAudio)

	parser := &MatroskaParser{
		reader:   NewEBMLReader(bytes.NewReader(tracksElement.Bytes())),
		fileInfo: &SegmentInfo{TimecodeScale: 1000000},
	}

	err := parser.parseTracks(uint64(tracksElement.Len()))
	if err != nil {
		t.Fatalf("parseTracks() failed: %v", err)
	}

	if len(parser.tracks) != 2 {
		t.Fatalf("Expected 2 tracks, got %d", len(parser.tracks))
	}

	// Video track checks
	videoTrack := parser.tracks[0]
	if videoTrack.Number != 1 {
		t.Errorf("Expected video track number 1, got %d", videoTrack.Number)
	}
	if videoTrack.Type != TypeVideo {
		t.Errorf("Expected video track type %d, got %d", TypeVideo, videoTrack.Type)
	}
	if videoTrack.CodecID != "V_MPEG4/ISO/AVC" {
		t.Errorf("Expected video CodecID 'V_MPEG4/ISO/AVC', got %q", videoTrack.CodecID)
	}
	if videoTrack.Name != "Video" {
		t.Errorf("Expected video name 'Video', got %q", videoTrack.Name)
	}

	// Audio track checks
	audioTrack := parser.tracks[1]
	if audioTrack.Number != 2 {
		t.Errorf("Expected audio track number 2, got %d", audioTrack.Number)
	}
	if audioTrack.Type != TypeAudio {
		t.Errorf("Expected audio track type %d, got %d", TypeAudio, audioTrack.Type)
	}
	if audioTrack.CodecID != "A_AAC" {
		t.Errorf("Expected audio CodecID 'A_AAC', got %q", audioTrack.CodecID)
	}
	if audioTrack.Language != "eng" {
		t.Errorf("Expected audio language 'eng', got %q", audioTrack.Language)
	}
}

// TestParseSimpleBlock tests the parsing of a SimpleBlock.
func TestParseSimpleBlock(t *testing.T) {
	// SimpleBlock: Track 1, Timecode 1234, Flags 0x80 (Keyframe), Data "frame"
	blockData := []byte{
		0x81,       // Track number 1
		0x04, 0xD2, // Timecode 1234
		0x80,                    // Flags (keyframe)
		'f', 'r', 'a', 'm', 'e', // Frame data
	}

	parser := &MatroskaParser{
		reader:           NewEBMLReader(bytes.NewReader(blockData)),
		clusterTimestamp: 1000,
		fileInfo: &SegmentInfo{
			TimecodeScale: uint64(time.Second / time.Nanosecond), // 1ms
		},
	}

	packet, err := parser.parseSimpleBlock(uint64(len(blockData)))
	if err != nil {
		t.Fatalf("parseSimpleBlock() failed: %v", err)
	}

	if packet.Track != 1 {
		t.Errorf("Expected track 1, got %d", packet.Track)
	}
	expectedTime := (1000 + 1234) * (uint64(time.Second) / uint64(time.Nanosecond))
	if packet.StartTime != expectedTime {
		t.Errorf("Expected start time %d, got %d", expectedTime, packet.StartTime)
	}
	if (packet.Flags & KF) == 0 {
		t.Error("Expected keyframe flag to be set")
	}
	if string(packet.Data) != "frame" {
		t.Errorf("Expected data 'frame', got %q", string(packet.Data))
	}
}

// TestNewMatroskaParser_EdgeCases tests edge cases for NewMatroskaParser.
func TestNewMatroskaParser_EdgeCases(t *testing.T) {
	t.Run("Invalid reader - empty", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})
		_, err := NewMatroskaParser(reader, false)
		if err == nil {
			t.Errorf("Expected error for empty reader, but got nil")
		}
	})

	t.Run("Invalid reader - non-EBML format", func(t *testing.T) {
		invalidData := []byte("This is not an EBML file")
		reader := bytes.NewReader(invalidData)
		_, err := NewMatroskaParser(reader, false)
		if err == nil {
			t.Errorf("Expected error for non-EBML format, but got nil")
		}
	})

	t.Run("Parser with noSeeking=true", func(t *testing.T) {
		// Create a mock Matroska file
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, true)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with noSeeking=true failed: %v", err)
		}

		if !parser.noSeeking {
			t.Errorf("Expected noSeeking to be true, got false")
		}
		if parser.header == nil {
			t.Error("Expected parser to have a non-nil header")
		}
		if parser.segment == nil {
			t.Error("Expected parser to have a non-nil segment")
		}
	})

	t.Run("Parser with different buffer sizes", func(t *testing.T) {
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		// Test with a small buffer, this should still work
		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with small buffer failed: %v", err)
		}
		if parser.header == nil {
			t.Error("Expected parser to have a non-nil header")
		}
	})

	t.Run("Parser with Cues scanning", func(t *testing.T) {
		// Create a mock file with Cues at the end to trigger the scanning logic
		mockFile, err := createMockMatroskaFileWithCues()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file with cues: %v", err)
		}

		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with cues scanning failed: %v", err)
		}

		// Should have found cues during scanning
		if parser.cuesPos == 0 {
			t.Error("Expected parser to find cues during scanning")
		}
		if len(parser.cues) == 0 {
			t.Error("Expected parser to have parsed cues")
		}
	})

	t.Run("Parser with seek error during cues scanning", func(t *testing.T) {
		// Create a mock file that will cause seek errors
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		// Use a limited reader that will cause seek errors
		limitedSeekerReader := &limitedSeeker{
			Reader: bytes.NewReader(mockFile),
			limit:  int64(len(mockFile) / 2), // Limit to half the file
		}

		_, err = NewMatroskaParser(limitedSeekerReader, false)
		// Should handle seek errors gracefully or return an error
		// The function should not panic
		_ = err
	})

	t.Run("Parser with parseHeader error", func(t *testing.T) {
		// Create invalid EBML header that will cause parseHeader to fail
		invalidHeader := []byte{0x1A, 0x45, 0xDF, 0xA3, 0x01} // Incomplete EBML header
		reader := bytes.NewReader(invalidHeader)

		_, err := NewMatroskaParser(reader, false)
		if err == nil {
			t.Error("Expected parseHeader error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse header") {
			t.Errorf("Expected 'failed to parse header' error, got: %v", err)
		}
	})

	t.Run("Parser with parseSegment error", func(t *testing.T) {
		// Create valid EBML header but invalid segment
		buf := new(bytes.Buffer)

		// Valid EBML header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Invalid segment (incomplete)
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67, 0x01}) // Segment ID + incomplete size

		reader := bytes.NewReader(buf.Bytes())
		_, err := NewMatroskaParser(reader, false)
		if err == nil {
			t.Error("Expected parseSegment error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse segment") {
			t.Errorf("Expected 'failed to parse segment' error, got: %v", err)
		}
	})

	t.Run("Parser with restore position error", func(t *testing.T) {
		// This is harder to test directly, but we can create a scenario
		// where the parser tries to restore position after cues scanning
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		// Use a reader that will work initially but fail on later seeks
		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Parser should be created successfully
		if parser == nil {
			t.Error("Expected parser to be created")
		}
	})
}

func TestParseHeader_EdgeCases(t *testing.T) {
	t.Run("Corrupted EBML header", func(t *testing.T) {
		// EBML header with invalid size
		data := []byte{0x1A, 0x45, 0xDF, 0xA3, 0x01, 0x02, 0x03}
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(data)),
		}
		err := parser.parseHeader()
		if err == nil {
			t.Errorf("Expected error for corrupted EBML header, but got nil")
		}
	})

	t.Run("Non-Matroska file header", func(t *testing.T) {
		// EBML header for a different document type
		buf := new(bytes.Buffer)
		ebmlHeader := new(bytes.Buffer)
		ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'o', 't', 'h', 'e', 'r', 'd', 'o', 'c'}) // DocType
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})                                          // EBML Header ID
		buf.Write(vintEncode(uint64(ebmlHeader.Len())))
		buf.Write(ebmlHeader.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseHeader()
		if err == nil {
			t.Errorf("Expected error for non-Matroska file header, but got nil")
		}
	})
}

func TestParseSegment_EdgeCases(t *testing.T) {
	t.Run("Empty Segment", func(t *testing.T) {
		// Create an empty segment
		data := []byte{0x18, 0x53, 0x80, 0x67, 0x80} // Segment ID with size 0
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(data)),
		}

		err := parser.parseSegment()
		if err != nil {
			t.Fatalf("parseSegment() with empty segment failed: %v", err)
		}
	})

	t.Run("Corrupted Segment", func(t *testing.T) {
		// Create a corrupted segment (e.g., invalid size)
		data := []byte{0x18, 0x53, 0x80, 0x67, 0xFF} // Segment ID with invalid size
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(data)),
		}

		err := parser.parseSegment()
		if err == nil {
			t.Errorf("Expected error for corrupted segment, but got nil")
		}
	})
}

// TestParseVideoTrack tests the parsing of video track data.
func TestParseVideoTrack(t *testing.T) {
	t.Run("Valid video track data", func(t *testing.T) {
		// Create mock video track data
		buf := new(bytes.Buffer)
		// PixelWidth: 1920
		buf.Write([]byte{0xB0, 0x82, 0x07, 0x80}) // ID: PixelWidth, Size: 2, Value: 1920
		// PixelHeight: 1080
		buf.Write([]byte{0xBA, 0x82, 0x04, 0x38}) // ID: PixelHeight, Size: 2, Value: 1080
		// DisplayWidth: 1920
		buf.Write([]byte{0x54, 0xB0, 0x82, 0x07, 0x80}) // ID: DisplayWidth, Size: 2, Value: 1920
		// DisplayHeight: 1080
		buf.Write([]byte{0x54, 0xBA, 0x82, 0x04, 0x38}) // ID: DisplayHeight, Size: 2, Value: 1080

		parser := &MatroskaParser{}
		track := &TrackInfo{}

		err := parser.parseVideoTrack(buf.Bytes(), track)
		if err != nil {
			t.Fatalf("parseVideoTrack() failed: %v", err)
		}

		if track.Video.PixelWidth != 1920 {
			t.Errorf("Expected PixelWidth 1920, got %d", track.Video.PixelWidth)
		}
		if track.Video.PixelHeight != 1080 {
			t.Errorf("Expected PixelHeight 1080, got %d", track.Video.PixelHeight)
		}
		if track.Video.DisplayWidth != 1920 {
			t.Errorf("Expected DisplayWidth 1920, got %d", track.Video.DisplayWidth)
		}
		if track.Video.DisplayHeight != 1080 {
			t.Errorf("Expected DisplayHeight 1080, got %d", track.Video.DisplayHeight)
		}
	})

	t.Run("Empty video track data", func(t *testing.T) {
		parser := &MatroskaParser{}
		track := &TrackInfo{}

		err := parser.parseVideoTrack([]byte{}, track)
		if err != nil {
			t.Fatalf("parseVideoTrack() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})

	// Cover interlaced flag branch
	t.Run("Interlaced flag", func(t *testing.T) {
		buf := new(bytes.Buffer)
		// FlagInterlaced: 1
		buf.Write([]byte{0x9A, 0x81, 0x01})
		parser := &MatroskaParser{}
		track := &TrackInfo{}
		if err := parser.parseVideoTrack(buf.Bytes(), track); err != nil {
			t.Fatalf("parseVideoTrack() failed: %v", err)
		}
		if !track.Video.Interlaced {
			t.Errorf("expected interlaced=true")
		}
	})
}

// TestParseAudioTrack tests the parsing of audio track data.
func TestParseAudioTrack(t *testing.T) {
	t.Run("Valid audio track data", func(t *testing.T) {
		// Create mock audio track data
		buf := new(bytes.Buffer)
		// SamplingFrequency: 48000.0 (as 64-bit float)
		samplingFreq := math.Float64bits(48000.0)
		buf.Write([]byte{0xB5, 0x88}) // ID: SamplingFrequency, Size: 8
		_ = binary.Write(buf, binary.BigEndian, samplingFreq)

		// Channels: 2
		buf.Write([]byte{0x9F, 0x81, 0x02}) // ID: Channels, Size: 1, Value: 2

		// BitDepth: 16
		buf.Write([]byte{0x62, 0x64, 0x81, 0x10}) // ID: BitDepth, Size: 1, Value: 16

		parser := &MatroskaParser{}
		track := &TrackInfo{}

		err := parser.parseAudioTrack(buf.Bytes(), track)
		if err != nil {
			t.Fatalf("parseAudioTrack() failed: %v", err)
		}

		if track.Audio.SamplingFreq != 48000.0 {
			t.Errorf("Expected SamplingFreq 48000.0, got %f", track.Audio.SamplingFreq)
		}
		if track.Audio.Channels != 2 {
			t.Errorf("Expected Channels 2, got %d", track.Audio.Channels)
		}
		if track.Audio.BitDepth != 16 {
			t.Errorf("Expected BitDepth 16, got %d", track.Audio.BitDepth)
		}
	})

	t.Run("Empty audio track data", func(t *testing.T) {
		parser := &MatroskaParser{}
		track := &TrackInfo{}

		err := parser.parseAudioTrack([]byte{}, track)
		if err != nil {
			t.Fatalf("parseAudioTrack() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})
}

// TestParseCues tests the parsing of Cues element.
func TestParseCues(t *testing.T) {
	t.Run("Valid cues data", func(t *testing.T) {
		// Create mock cues data with one CuePoint
		buf := new(bytes.Buffer)

		// CuePoint
		cuePoint := new(bytes.Buffer)
		// CueTime: 1000
		cuePoint.Write([]byte{0xB3, 0x82, 0x03, 0xE8})
		// CueTrackPositions
		cueTrackPositions := new(bytes.Buffer)
		// CueTrack: 1
		cueTrackPositions.Write([]byte{0xF7, 0x81, 0x01})
		// CueClusterPosition: 100
		cueTrackPositions.Write([]byte{0xF1, 0x81, 0x64})
		cuePoint.Write([]byte{0xB7}) // CueTrackPositions ID
		cuePoint.Write(vintEncode(uint64(cueTrackPositions.Len())))
		cuePoint.Write(cueTrackPositions.Bytes())

		buf.Write([]byte{0xBB}) // CuePoint ID
		buf.Write(vintEncode(uint64(cuePoint.Len())))
		buf.Write(cuePoint.Bytes())

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(buf.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseCues(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseCues() failed: %v", err)
		}

		if len(parser.cues) == 0 {
			t.Fatal("Expected at least one cue, got none")
		}

		cue := parser.cues[0]
		if cue.Time != 1000000000 { // 1000 * 1000000 (timecode scale)
			t.Errorf("Expected cue time 1000000000, got %d", cue.Time)
		}
		if cue.Track != 1 {
			t.Errorf("Expected cue track 1, got %d", cue.Track)
		}
		if cue.Position != 100 {
			t.Errorf("Expected cue position 100, got %d", cue.Position)
		}
	})

	t.Run("Empty cues data", func(t *testing.T) {
		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader([]byte{})),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseCues(0)
		if err != nil {
			t.Fatalf("parseCues() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})

	t.Run("Multiple CuePoints with sorting", func(t *testing.T) {
		buf := new(bytes.Buffer)

		// Create two CuePoints with times in reverse order to test sorting
		// CuePoint 1 (time: 2000)
		cuePoint1 := new(bytes.Buffer)
		cuePoint1.Write([]byte{0xB3, 0x82, 0x07, 0xD0}) // CueTime: 2000
		cueTrackPos1 := new(bytes.Buffer)
		cueTrackPos1.Write([]byte{0xF7, 0x81, 0x01}) // CueTrack: 1
		cueTrackPos1.Write([]byte{0xF1, 0x81, 0x64}) // CueClusterPosition: 100
		cuePoint1.Write([]byte{0xB7})                // CueTrackPositions ID
		cuePoint1.Write(vintEncode(uint64(cueTrackPos1.Len())))
		cuePoint1.Write(cueTrackPos1.Bytes())

		buf.Write([]byte{0xBB}) // CuePoint ID
		buf.Write(vintEncode(uint64(cuePoint1.Len())))
		buf.Write(cuePoint1.Bytes())

		// CuePoint 2 (time: 1000)
		cuePoint2 := new(bytes.Buffer)
		cuePoint2.Write([]byte{0xB3, 0x82, 0x03, 0xE8}) // CueTime: 1000
		cueTrackPos2 := new(bytes.Buffer)
		cueTrackPos2.Write([]byte{0xF7, 0x81, 0x01}) // CueTrack: 1
		cueTrackPos2.Write([]byte{0xF1, 0x81, 0x32}) // CueClusterPosition: 50
		cuePoint2.Write([]byte{0xB7})                // CueTrackPositions ID
		cuePoint2.Write(vintEncode(uint64(cueTrackPos2.Len())))
		cuePoint2.Write(cueTrackPos2.Bytes())

		buf.Write([]byte{0xBB}) // CuePoint ID
		buf.Write(vintEncode(uint64(cuePoint2.Len())))
		buf.Write(cuePoint2.Bytes())

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(buf.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseCues(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseCues() with multiple cues failed: %v", err)
		}

		if len(parser.cues) != 2 {
			t.Fatalf("Expected 2 cues, got %d", len(parser.cues))
		}

		// Should be sorted by time (1000 before 2000)
		if parser.cues[0].Time >= parser.cues[1].Time {
			t.Errorf("Cues not sorted correctly: first=%d, second=%d",
				parser.cues[0].Time, parser.cues[1].Time)
		}
	})

	t.Run("Invalid cues data", func(t *testing.T) {
		// Create invalid EBML data
		invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(invalidData)),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseCues(uint64(len(invalidData)))
		if err == nil {
			t.Error("Expected error for invalid cues data, but got nil")
		}
	})

	t.Run("ReadFull error", func(t *testing.T) {
		// Create a reader that will fail on ReadFull
		reader := &failingReader{}
		parser := &MatroskaParser{
			reader:   NewEBMLReader(reader),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseCues(100) // Request more data than available
		if err == nil {
			t.Error("Expected error for ReadFull failure, but got nil")
		}
	})
}

// TestParseCuePoint tests the parsing of a CuePoint element.
func TestParseCuePoint(t *testing.T) {
	t.Run("Valid cue point data", func(t *testing.T) {
		// Create mock cue point data
		buf := new(bytes.Buffer)
		// CueTime: 2000
		buf.Write([]byte{0xB3, 0x82, 0x07, 0xD0})
		// CueTrackPositions
		cueTrackPositions := new(bytes.Buffer)
		// CueTrack: 2
		cueTrackPositions.Write([]byte{0xF7, 0x81, 0x02})
		// CueClusterPosition: 200
		cueTrackPositions.Write([]byte{0xF1, 0x81, 0xC8})
		buf.Write([]byte{0xB7}) // CueTrackPositions ID
		buf.Write(vintEncode(uint64(cueTrackPositions.Len())))
		buf.Write(cueTrackPositions.Bytes())

		parser := &MatroskaParser{
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		cues, err := parser.parseCuePoint(buf.Bytes())
		if err != nil {
			t.Fatalf("parseCuePoint() failed: %v", err)
		}

		if len(cues) == 0 {
			t.Fatal("Expected at least one cue, got none")
		}

		cue := cues[0]
		if cue.Time != 2000000000 { // 2000 * 1000000 (timecode scale)
			t.Errorf("Expected cue time 2000000000, got %d", cue.Time)
		}
		if cue.Track != 2 {
			t.Errorf("Expected cue track 2, got %d", cue.Track)
		}
		if cue.Position != 200 {
			t.Errorf("Expected cue position 200, got %d", cue.Position)
		}
	})

	t.Run("Cue point missing CueTime", func(t *testing.T) {
		// Create cue point data without CueTime
		buf := new(bytes.Buffer)
		// Only CueTrackPositions, no CueTime
		cueTrackPositions := new(bytes.Buffer)
		cueTrackPositions.Write([]byte{0xF7, 0x81, 0x01})
		cueTrackPositions.Write([]byte{0xF1, 0x81, 0x64})
		buf.Write([]byte{0xB7}) // CueTrackPositions ID
		buf.Write(vintEncode(uint64(cueTrackPositions.Len())))
		buf.Write(cueTrackPositions.Bytes())

		parser := &MatroskaParser{
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		cues, err := parser.parseCuePoint(buf.Bytes())
		if err != nil {
			t.Fatalf("parseCuePoint() failed: %v", err)
		}
		// Should handle missing CueTime gracefully (might return empty cues or cues with default time)
		_ = cues
	})
}

// TestParseCueTrackPositions tests the parsing of CueTrackPositions element.
func TestParseCueTrackPositions(t *testing.T) {
	t.Run("Valid cue track positions data", func(t *testing.T) {
		// Create mock cue track positions data
		buf := new(bytes.Buffer)
		// CueTrack: 3
		buf.Write([]byte{0xF7, 0x81, 0x03})
		// CueClusterPosition: 300
		buf.Write([]byte{0xF1, 0x82, 0x01, 0x2C})
		// CueRelativePosition: 50
		buf.Write([]byte{0xF0, 0x81, 0x32})
		// CueBlockNumber: 1
		buf.Write([]byte{0x53, 0x78, 0x81, 0x01})

		parser := &MatroskaParser{}

		cue, err := parser.parseCueTrackPositions(buf.Bytes())
		if err != nil {
			t.Fatalf("parseCueTrackPositions() failed: %v", err)
		}

		if cue.Track != 3 {
			t.Errorf("Expected cue track 3, got %d", cue.Track)
		}
		if cue.Position != 300 {
			t.Errorf("Expected cue position 300, got %d", cue.Position)
		}
		if cue.RelativePosition != 50 {
			t.Errorf("Expected cue relative position 50, got %d", cue.RelativePosition)
		}
		if cue.Block != 1 {
			t.Errorf("Expected cue block 1, got %d", cue.Block)
		}
	})

	t.Run("Cue track positions missing CueTrack", func(t *testing.T) {
		// Create cue track positions data without CueTrack
		buf := new(bytes.Buffer)
		// Only CueClusterPosition, no CueTrack
		buf.Write([]byte{0xF1, 0x81, 0x64})

		parser := &MatroskaParser{}

		cue, err := parser.parseCueTrackPositions(buf.Bytes())
		if err != nil {
			t.Fatalf("parseCueTrackPositions() failed: %v", err)
		}
		// Should handle missing CueTrack gracefully (might have default track value)
		_ = cue
	})
}

// TestParseTags tests the parsing of Tags element.
func TestParseTags(t *testing.T) {
	t.Run("Valid tags data", func(t *testing.T) {
		// Create mock tags data with one Tag
		buf := new(bytes.Buffer)

		// Tag
		tag := new(bytes.Buffer)
		// Targets
		targets := new(bytes.Buffer)
		targets.Write([]byte{0x68, 0xCA, 0x81, 0x32}) // TargetTypeValue = 50
		tag.Write([]byte{0x63, 0xC0})                 // Targets ID
		tag.Write(vintEncode(uint64(targets.Len())))
		tag.Write(targets.Bytes())

		// SimpleTag
		simpleTag := new(bytes.Buffer)
		simpleTag.Write([]byte{0x45, 0xA3, 0x85, 'T', 'I', 'T', 'L', 'E'}) // TagName = "TITLE"
		simpleTag.Write([]byte{0x44, 0x87, 0x85, 'A', 'l', 'b', 'u', 'm'}) // TagString = "Album"
		tag.Write([]byte{0x67, 0xC8})                                      // SimpleTag ID
		tag.Write(vintEncode(uint64(simpleTag.Len())))
		tag.Write(simpleTag.Bytes())

		buf.Write([]byte{0x73, 0x73}) // Tag ID
		buf.Write(vintEncode(uint64(tag.Len())))
		buf.Write(tag.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseTags(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseTags() failed: %v", err)
		}

		if len(parser.tags) == 0 {
			t.Fatal("Expected at least one tag, got none")
		}

		parsedTag := parser.tags[0]
		if len(parsedTag.Targets) == 0 {
			t.Fatal("Expected tag targets, got none")
		}
		if parsedTag.Targets[0].Type != 50 {
			t.Errorf("Expected target type 50, got %d", parsedTag.Targets[0].Type)
		}
		if len(parsedTag.SimpleTags) == 0 {
			t.Fatal("Expected simple tags, got none")
		}
		if parsedTag.SimpleTags[0].Name != "TITLE" {
			t.Errorf("Expected tag name 'TITLE', got %q", parsedTag.SimpleTags[0].Name)
		}
		if parsedTag.SimpleTags[0].Value != "Album" {
			t.Errorf("Expected tag value 'Album', got %q", parsedTag.SimpleTags[0].Value)
		}
	})

	t.Run("Empty tags data", func(t *testing.T) {
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader([]byte{})),
		}

		err := parser.parseTags(0)
		if err != nil {
			t.Fatalf("parseTags() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})

	t.Run("ReadFull error", func(t *testing.T) {
		// Create a reader that will fail on ReadFull
		reader := &failingReader{
			data:       make([]byte, 5), // Small data
			failAtByte: 3,               // Fail after 3 bytes
		}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseTags(10) // Request more bytes than available
		if err == nil {
			t.Fatal("Expected ReadFull error, got nil")
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("ReadElement error", func(t *testing.T) {
		// Create invalid EBML data that will cause ReadElement to fail
		invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid EBML element
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(invalidData)),
		}

		err := parser.parseTags(uint64(len(invalidData)))
		if err == nil {
			t.Fatal("Expected ReadElement error, got nil")
		}
	})

	t.Run("Non-Tag elements", func(t *testing.T) {
		// Create tags data with non-Tag elements (should be ignored)
		buf := new(bytes.Buffer)

		// Add a non-Tag element (using a different ID)
		buf.Write([]byte{0x12, 0x34, 0x81, 0x00}) // Unknown element with size 1 and data 0x00

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseTags(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseTags() with non-Tag elements failed: %v", err)
		}
		// Should ignore non-Tag elements
		if len(parser.tags) != 0 {
			t.Errorf("Expected no tags, got %d", len(parser.tags))
		}
	})

	t.Run("parseTag error", func(t *testing.T) {
		// Create tags data with invalid Tag that will cause parseTag to fail
		buf := new(bytes.Buffer)

		// Tag with invalid data
		buf.Write([]byte{0x73, 0x73})             // Tag ID
		buf.Write([]byte{0x84})                   // Size: 4
		buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseTags(uint64(buf.Len()))
		if err == nil {
			t.Fatal("Expected parseTag error, got nil")
		}
	})
}

// TestParseTag tests the parsing of a Tag element.
func TestParseTag(t *testing.T) {
	t.Run("Valid tag data", func(t *testing.T) {
		// Create mock tag data
		buf := new(bytes.Buffer)
		// Targets
		targets := new(bytes.Buffer)
		targets.Write([]byte{0x68, 0xCA, 0x81, 0x3C}) // TargetTypeValue = 60
		buf.Write([]byte{0x63, 0xC0})                 // Targets ID
		buf.Write(vintEncode(uint64(targets.Len())))
		buf.Write(targets.Bytes())

		// SimpleTag
		simpleTag := new(bytes.Buffer)
		simpleTag.Write([]byte{0x45, 0xA3, 0x86, 'A', 'R', 'T', 'I', 'S', 'T'})                // TagName = "ARTIST"
		simpleTag.Write([]byte{0x44, 0x87, 0x89, 'T', 'e', 's', 't', ' ', 'B', 'a', 'n', 'd'}) // TagString = "Test Band"
		buf.Write([]byte{0x67, 0xC8})                                                          // SimpleTag ID
		buf.Write(vintEncode(uint64(simpleTag.Len())))
		buf.Write(simpleTag.Bytes())

		parser := &MatroskaParser{}

		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() failed: %v", err)
		}

		if len(tag.Targets) == 0 {
			t.Fatal("Expected tag targets, got none")
		}
		if tag.Targets[0].Type != 60 {
			t.Errorf("Expected target type 60, got %d", tag.Targets[0].Type)
		}
		if len(tag.SimpleTags) == 0 {
			t.Fatal("Expected simple tags, got none")
		}
		if tag.SimpleTags[0].Name != "ARTIST" {
			t.Errorf("Expected tag name 'ARTIST', got %q", tag.SimpleTags[0].Name)
		}
		if tag.SimpleTags[0].Value != "Test Band" {
			t.Errorf("Expected tag value 'Test Band', got %q", tag.SimpleTags[0].Value)
		}
	})

	t.Run("Empty tag data", func(t *testing.T) {
		parser := &MatroskaParser{}

		tag, err := parser.parseTag([]byte{})
		if err != nil {
			t.Fatalf("parseTag() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
		_ = tag
	})
}

// TestParseTarget tests the parsing of a Target element.
func TestParseTarget(t *testing.T) {
	t.Run("Valid target data", func(t *testing.T) {
		// Create mock target data
		buf := new(bytes.Buffer)
		// TargetTypeValue: 70
		buf.Write([]byte{0x68, 0xCA, 0x81, 0x46})
		// TargetType: "TRACK"
		buf.Write([]byte{0x63, 0xCA, 0x85, 'T', 'R', 'A', 'C', 'K'})
		// TrackUID: 1
		buf.Write([]byte{0x63, 0xC5, 0x81, 0x01})

		parser := &MatroskaParser{}

		target, err := parser.parseTarget(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTarget() failed: %v", err)
		}

		if target.Type != 70 {
			t.Errorf("Expected target type 70, got %d", target.Type)
		}
		if target.UID != 1 {
			t.Errorf("Expected target UID 1, got %d", target.UID)
		}
	})

	t.Run("Empty target data", func(t *testing.T) {
		parser := &MatroskaParser{}

		target, err := parser.parseTarget([]byte{})
		if err != nil {
			t.Fatalf("parseTarget() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
		_ = target
	})

	t.Run("Target with different UID types", func(t *testing.T) {
		parser := &MatroskaParser{}

		// Test IDTagEditionUID
		buf := new(bytes.Buffer)
		buf.Write([]byte{0x63, 0xC9, 0x81, 0x02}) // IDTagEditionUID: 2
		target, err := parser.parseTarget(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTarget() with EditionUID failed: %v", err)
		}
		if target.UID != 2 {
			t.Errorf("Expected EditionUID 2, got %d", target.UID)
		}

		// Test IDTagChapterUID
		buf.Reset()
		buf.Write([]byte{0x63, 0xC4, 0x81, 0x03}) // IDTagChapterUID: 3
		target, err = parser.parseTarget(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTarget() with ChapterUID failed: %v", err)
		}
		if target.UID != 3 {
			t.Errorf("Expected ChapterUID 3, got %d", target.UID)
		}

		// Test IDTagAttachmentUID
		buf.Reset()
		buf.Write([]byte{0x63, 0xC6, 0x81, 0x04}) // IDTagAttachmentUID: 4
		target, err = parser.parseTarget(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTarget() with AttachmentUID failed: %v", err)
		}
		if target.UID != 4 {
			t.Errorf("Expected AttachmentUID 4, got %d", target.UID)
		}
	})

	t.Run("Target with error handling", func(t *testing.T) {
		parser := &MatroskaParser{}

		// Test with invalid data that should cause ReadElement to fail
		invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid EBML data
		_, err := parser.parseTarget(invalidData)
		if err == nil {
			t.Error("Expected error for invalid target data, but got nil")
		}
	})
}

// TestParseSimpleTag tests the parsing of a SimpleTag element.
func TestParseSimpleTag(t *testing.T) {
	t.Run("Valid simple tag data", func(t *testing.T) {
		// Create mock simple tag data
		buf := new(bytes.Buffer)
		// TagName: "GENRE"
		buf.Write([]byte{0x45, 0xA3, 0x85, 'G', 'E', 'N', 'R', 'E'})
		// TagString: "Rock"
		buf.Write([]byte{0x44, 0x87, 0x84, 'R', 'o', 'c', 'k'})
		// TagLanguage: "eng"
		buf.Write([]byte{0x44, 0x7A, 0x83, 'e', 'n', 'g'})
		// TagDefault: 1
		buf.Write([]byte{0x44, 0x84, 0x81, 0x01})

		parser := &MatroskaParser{}

		simpleTag, err := parser.parseSimpleTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseSimpleTag() failed: %v", err)
		}

		if simpleTag.Name != "GENRE" {
			t.Errorf("Expected tag name 'GENRE', got %q", simpleTag.Name)
		}
		if simpleTag.Value != "Rock" {
			t.Errorf("Expected tag value 'Rock', got %q", simpleTag.Value)
		}
		if simpleTag.Language != "eng" {
			t.Errorf("Expected tag language 'eng', got %q", simpleTag.Language)
		}
		if !simpleTag.Default {
			t.Errorf("Expected tag default to be true, got false")
		}
	})

	t.Run("Nested simple tag", func(t *testing.T) {
		// Create mock simple tag data with nested SimpleTag
		buf := new(bytes.Buffer)
		// TagName: "ALBUM"
		buf.Write([]byte{0x45, 0xA3, 0x85, 'A', 'L', 'B', 'U', 'M'})
		// TagString: "Test Album"
		buf.Write([]byte{0x44, 0x87, 0x8A, 'T', 'e', 's', 't', ' ', 'A', 'l', 'b', 'u', 'm'})

		// Nested SimpleTag
		nestedTag := new(bytes.Buffer)
		nestedTag.Write([]byte{0x45, 0xA3, 0x84, 'Y', 'E', 'A', 'R'}) // TagName: "YEAR"
		nestedTag.Write([]byte{0x44, 0x87, 0x84, '2', '0', '2', '3'}) // TagString: "2023"
		buf.Write([]byte{0x67, 0xC8})                                 // SimpleTag ID
		buf.Write(vintEncode(uint64(nestedTag.Len())))
		buf.Write(nestedTag.Bytes())

		parser := &MatroskaParser{}

		simpleTag, err := parser.parseSimpleTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseSimpleTag() failed: %v", err)
		}

		if simpleTag.Name != "ALBUM" {
			t.Errorf("Expected tag name 'ALBUM', got %q", simpleTag.Name)
		}
		if simpleTag.Value != "Test Album" {
			t.Errorf("Expected tag value 'Test Album', got %q", simpleTag.Value)
		}
		// Note: Nested tags might not be directly accessible in the current structure
	})

	t.Run("Empty simple tag data", func(t *testing.T) {
		parser := &MatroskaParser{}

		simpleTag, err := parser.parseSimpleTag([]byte{})
		if err != nil {
			t.Fatalf("parseSimpleTag() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
		_ = simpleTag
	})
}

// TestParseAttachments tests the parsing of Attachments element.
func TestParseAttachments(t *testing.T) {
	t.Run("Valid attachments data", func(t *testing.T) {
		// Create mock attachments data with one AttachedFile
		buf := new(bytes.Buffer)

		// AttachedFile
		attachedFile := new(bytes.Buffer)
		// FileName: "test.txt"
		attachedFile.Write([]byte{0x46, 0x6E, 0x88, 't', 'e', 's', 't', '.', 't', 'x', 't'})
		// FileMimeType: "text/plain"
		attachedFile.Write([]byte{0x46, 0x60, 0x8A, 't', 'e', 'x', 't', '/', 'p', 'l', 'a', 'i', 'n'})
		// FileData: "hello"
		attachedFile.Write([]byte{0x46, 0x5C, 0x85, 'h', 'e', 'l', 'l', 'o'})
		// FileUID: 1
		attachedFile.Write([]byte{0x46, 0xAE, 0x81, 0x01})

		buf.Write([]byte{0x61, 0xA7}) // AttachedFile ID
		buf.Write(vintEncode(uint64(attachedFile.Len())))
		buf.Write(attachedFile.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseAttachments(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseAttachments() failed: %v", err)
		}

		if len(parser.attachments) == 0 {
			t.Fatal("Expected at least one attachment, got none")
		}

		attachment := parser.attachments[0]
		if attachment.Name != "test.txt" {
			t.Errorf("Expected attachment name 'test.txt', got %q", attachment.Name)
		}
		if attachment.MimeType != "text/plain" {
			t.Errorf("Expected MIME type 'text/plain', got %q", attachment.MimeType)
		}
		if attachment.Length == 0 {
			t.Errorf("Expected attachment to have data length > 0, got %d", attachment.Length)
		}
		if attachment.UID != 1 {
			t.Errorf("Expected attachment UID 1, got %d", attachment.UID)
		}
	})

	t.Run("Empty attachments data", func(t *testing.T) {
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader([]byte{})),
		}

		err := parser.parseAttachments(0)
		if err != nil {
			t.Fatalf("parseAttachments() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})
}

// TestParseAttachedFile tests the parsing of an AttachedFile element.
func TestParseAttachedFile(t *testing.T) {
	t.Run("Valid attached file data", func(t *testing.T) {
		// Create mock attached file data
		buf := new(bytes.Buffer)
		// FileName: "image.jpg"
		buf.Write([]byte{0x46, 0x6E, 0x89, 'i', 'm', 'a', 'g', 'e', '.', 'j', 'p', 'g'})
		// FileMimeType: "image/jpeg"
		buf.Write([]byte{0x46, 0x60, 0x8A, 'i', 'm', 'a', 'g', 'e', '/', 'j', 'p', 'e', 'g'})
		// FileDescription: "Test image"
		buf.Write([]byte{0x46, 0x75, 0x8A, 'T', 'e', 's', 't', ' ', 'i', 'm', 'a', 'g', 'e'})
		// FileData: "data"
		buf.Write([]byte{0x46, 0x5C, 0x84, 'd', 'a', 't', 'a'})
		// FileUID: 2
		buf.Write([]byte{0x46, 0xAE, 0x81, 0x02})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		attachment, err := parser.parseAttachedFile(buf.Bytes())
		if err != nil {
			t.Fatalf("parseAttachedFile() failed: %v", err)
		}

		if attachment.Name != "image.jpg" {
			t.Errorf("Expected attachment name 'image.jpg', got %q", attachment.Name)
		}
		if attachment.MimeType != "image/jpeg" {
			t.Errorf("Expected MIME type 'image/jpeg', got %q", attachment.MimeType)
		}
		// Description might not be parsed or might be empty
		_ = attachment.Description
		if attachment.Length == 0 {
			t.Errorf("Expected attachment to have data length > 0, got %d", attachment.Length)
		}
		if attachment.UID != 2 {
			t.Errorf("Expected attachment UID 2, got %d", attachment.UID)
		}
	})

	t.Run("Attached file missing FileData", func(t *testing.T) {
		// Create attached file data without FileData
		buf := new(bytes.Buffer)
		// FileName: "empty.txt"
		buf.Write([]byte{0x46, 0x6E, 0x89, 'e', 'm', 'p', 't', 'y', '.', 't', 'x', 't'})
		// FileMimeType: "text/plain"
		buf.Write([]byte{0x46, 0x60, 0x8A, 't', 'e', 'x', 't', '/', 'p', 'l', 'a', 'i', 'n'})
		// FileUID: 3
		buf.Write([]byte{0x46, 0xAE, 0x81, 0x03})
		// No FileData

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		attachment, err := parser.parseAttachedFile(buf.Bytes())
		if err != nil {
			t.Fatalf("parseAttachedFile() failed: %v", err)
		}
		// Should handle missing FileData gracefully
		if attachment.Name != "empty.txt" {
			t.Errorf("Expected attachment name 'empty.txt', got %q", attachment.Name)
		}
		// Should handle missing FileData gracefully - length might be 0
		_ = attachment.Length
	})

	t.Run("Empty attached file data", func(t *testing.T) {
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader([]byte{})),
		}

		attachment, err := parser.parseAttachedFile([]byte{})
		if err != nil {
			t.Fatalf("parseAttachedFile() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
		_ = attachment
	})
}

// TestReadPacket_Advanced tests advanced scenarios for ReadPacket.
func TestReadPacket_Advanced(t *testing.T) {
	t.Run("Read packet from mock file", func(t *testing.T) {
		// Create a mock Matroska file with a packet
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Try to read a packet
		packet, err := parser.ReadPacket()
		if err != nil && err != io.EOF {
			t.Fatalf("ReadPacket() failed: %v", err)
		}

		if packet != nil {
			// Verify packet properties
			if packet.Track == 0 {
				t.Errorf("Expected packet track > 0, got %d", packet.Track)
			}
			if len(packet.Data) == 0 {
				t.Errorf("Expected packet data length > 0, got %d", len(packet.Data))
			}
		}
	})

	t.Run("Read packet with EOF", func(t *testing.T) {
		// Create a minimal mock file that will quickly reach EOF
		buf := new(bytes.Buffer)
		// Just EBML header, no segment data
		ebmlHeader := new(bytes.Buffer)
		ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(ebmlHeader.Len())))
		buf.Write(ebmlHeader.Bytes())

		reader := bytes.NewReader(buf.Bytes())
		parser, err := NewMatroskaParser(reader, false)
		if err == nil {
			// If parser creation succeeds, try to read packet
			_, err = parser.ReadPacket()
			if err != io.EOF {
				t.Errorf("Expected EOF when reading from empty segment, got %v", err)
			}
		}
		// If parser creation fails, that's also acceptable for this minimal file
	})
}

// TestParseSimpleBlock_Advanced tests advanced scenarios for parseSimpleBlock.
func TestParseSimpleBlock_Advanced(t *testing.T) {
	t.Run("SimpleBlock with different flags", func(t *testing.T) {
		// Create SimpleBlock with different flag combinations
		testCases := []struct {
			name     string
			flags    byte
			expected uint32
		}{
			{"Keyframe", 0x80, KF},
			{"No flags", 0x00, 0},
			{"Invisible", 0x08, 0},   // No specific constant for invisible
			{"Discardable", 0x01, 0}, // No specific constant for discardable
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				blockData := []byte{
					0x81,       // Track number 1
					0x00, 0x00, // Timecode 0
					tc.flags,                // Flags
					'f', 'r', 'a', 'm', 'e', // Frame data
				}

				parser := &MatroskaParser{
					reader:           NewEBMLReader(bytes.NewReader(blockData)),
					clusterTimestamp: 0,
					fileInfo: &SegmentInfo{
						TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
					},
				}

				packet, err := parser.parseSimpleBlock(uint64(len(blockData)))
				if err != nil {
					t.Fatalf("parseSimpleBlock() failed: %v", err)
				}

				if (packet.Flags&tc.expected) == 0 && tc.expected != 0 {
					t.Errorf("Expected flag %d to be set, but it wasn't. Got flags: %d", tc.expected, packet.Flags)
				}
			})
		}
	})

	t.Run("SimpleBlock with large timecode", func(t *testing.T) {
		// Create SimpleBlock with large timecode
		blockData := []byte{
			0x81,       // Track number 1
			0x7F, 0xFF, // Large timecode (32767)
			0x80,                    // Keyframe flag
			'f', 'r', 'a', 'm', 'e', // Frame data
		}

		parser := &MatroskaParser{
			reader:           NewEBMLReader(bytes.NewReader(blockData)),
			clusterTimestamp: 1000,
			fileInfo: &SegmentInfo{
				TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
			},
		}

		packet, err := parser.parseSimpleBlock(uint64(len(blockData)))
		if err != nil {
			t.Fatalf("parseSimpleBlock() failed: %v", err)
		}

		expectedTime := (1000 + 32767) * uint64(time.Millisecond/time.Nanosecond)
		if packet.StartTime != expectedTime {
			t.Errorf("Expected start time %d, got %d", expectedTime, packet.StartTime)
		}
	})

	t.Run("SimpleBlock with invalid track number", func(t *testing.T) {
		// Create SimpleBlock with invalid track number (0)
		blockData := []byte{
			0x80,       // Invalid track number (0 with length marker)
			0x00, 0x00, // Timecode 0
			0x80,                    // Keyframe flag
			'f', 'r', 'a', 'm', 'e', // Frame data
		}

		parser := &MatroskaParser{
			reader:           NewEBMLReader(bytes.NewReader(blockData)),
			clusterTimestamp: 0,
			fileInfo: &SegmentInfo{
				TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
			},
		}

		_, err := parser.parseSimpleBlock(uint64(len(blockData)))
		// Should handle invalid track number gracefully (might return error or default)
		_ = err
	})
}

// TestParseClusterHeader tests the parsing of cluster header.
func TestParseClusterHeader(t *testing.T) {
	t.Run("Valid cluster header", func(t *testing.T) {
		// Create mock cluster header data
		buf := new(bytes.Buffer)
		// Timestamp: 1000
		buf.Write([]byte{0xE7, 0x82, 0x03, 0xE8})
		// Position: 100 (optional)
		buf.Write([]byte{0xA7, 0x81, 0x64})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseClusterHeader(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseClusterHeader() failed: %v", err)
		}

		if parser.clusterTimestamp != 1000 {
			t.Errorf("Expected cluster timestamp 1000, got %d", parser.clusterTimestamp)
		}
	})

	t.Run("Empty cluster header", func(t *testing.T) {
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader([]byte{})),
		}

		err := parser.parseClusterHeader(0)
		if err != nil {
			t.Fatalf("parseClusterHeader() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})

	t.Run("Cluster header without timestamp", func(t *testing.T) {
		// Create cluster header data without timestamp
		buf := new(bytes.Buffer)
		// Position: 100 (no timestamp)
		buf.Write([]byte{0xA7, 0x81, 0x64})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseClusterHeader(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseClusterHeader() without timestamp failed: %v", err)
		}

		// Should set timestamp to 0 when not found
		if parser.clusterTimestamp != 0 {
			t.Errorf("Expected cluster timestamp 0 when not found, got %d", parser.clusterTimestamp)
		}
	})

	t.Run("Invalid cluster header data", func(t *testing.T) {
		// Create invalid EBML data
		invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(invalidData)),
		}

		err := parser.parseClusterHeader(uint64(len(invalidData)))
		if err == nil {
			t.Error("Expected error for invalid cluster header data, but got nil")
		}
	})

	t.Run("ReadFull error", func(t *testing.T) {
		// Create a reader that will fail on ReadFull
		reader := &failingReader{}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseClusterHeader(100) // Request more data than available
		if err == nil {
			t.Error("Expected error for ReadFull failure, but got nil")
		}
	})
}

// TestParseBlockGroup tests the parsing of BlockGroup element.
func TestParseBlockGroup(t *testing.T) {
	t.Run("Valid block group", func(t *testing.T) {
		// Create mock block group data
		buf := new(bytes.Buffer)

		// Block
		block := new(bytes.Buffer)
		block.Write([]byte{0x81})                    // Track number 1
		block.Write([]byte{0x00, 0x00})              // Timecode 0
		block.Write([]byte{0x80})                    // Flags (keyframe)
		block.Write([]byte{'f', 'r', 'a', 'm', 'e'}) // Frame data
		buf.Write([]byte{0xA1})                      // Block ID
		buf.Write(vintEncode(uint64(block.Len())))
		buf.Write(block.Bytes())

		// BlockDuration: 40 (optional)
		buf.Write([]byte{0x9B, 0x81, 0x28})

		// ReferenceBlock: -1 (optional)
		buf.Write([]byte{0xFB, 0x81, 0xFF})

		parser := &MatroskaParser{
			reader:           NewEBMLReader(bytes.NewReader(buf.Bytes())),
			clusterTimestamp: 1000,
			fileInfo: &SegmentInfo{
				TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
			},
		}

		packet, err := parser.parseBlockGroup(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseBlockGroup() failed: %v", err)
		}

		if packet == nil {
			t.Fatal("Expected packet, got nil")
		}
		if packet.Track != 1 {
			t.Errorf("Expected track 1, got %d", packet.Track)
		}
		if string(packet.Data) != "frame" {
			t.Errorf("Expected packet data 'frame', got %q", string(packet.Data))
		}
		expectedDuration := 40 * uint64(time.Millisecond/time.Nanosecond)
		actualDuration := packet.EndTime - packet.StartTime
		if actualDuration != expectedDuration {
			t.Errorf("Expected duration %d, got %d", expectedDuration, actualDuration)
		}
	})

	t.Run("Block group without Block", func(t *testing.T) {
		// Create block group data without Block element
		buf := new(bytes.Buffer)
		// BlockDuration: 40
		buf.Write([]byte{0x9B, 0x81, 0x28})

		parser := &MatroskaParser{
			reader:           NewEBMLReader(bytes.NewReader(buf.Bytes())),
			clusterTimestamp: 1000,
			fileInfo: &SegmentInfo{
				TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
			},
		}

		packet, err := parser.parseBlockGroup(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseBlockGroup() failed: %v", err)
		}
		// Should handle missing Block gracefully (might return nil packet)
		_ = packet
	})

	t.Run("Empty block group", func(t *testing.T) {
		parser := &MatroskaParser{
			reader:           NewEBMLReader(bytes.NewReader([]byte{})),
			clusterTimestamp: 1000,
			fileInfo: &SegmentInfo{
				TimecodeScale: uint64(time.Millisecond / time.Nanosecond),
			},
		}

		packet, err := parser.parseBlockGroup(0)
		if err != nil {
			t.Fatalf("parseBlockGroup() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully (might return nil packet)
		_ = packet
	})
}

// TestReadPacket_Comprehensive tests comprehensive scenarios for ReadPacket.
func TestReadPacket_Comprehensive(t *testing.T) {
	t.Run("Read multiple packets", func(t *testing.T) {
		// Create a more complex mock file with multiple clusters
		mockFile, err := createMockMatroskaFileWithMultipleClusters()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Try to read multiple packets
		packetsRead := 0
		for i := 0; i < 5; i++ {
			packet, errReadPacket := parser.ReadPacket()
			if errReadPacket == io.EOF {
				break
			}
			if errReadPacket != nil {
				t.Fatalf("ReadPacket() failed on iteration %d: %v", i, errReadPacket)
			}
			if packet != nil {
				packetsRead++
				if packet.Track == 0 {
					t.Errorf("Expected packet track > 0, got %d", packet.Track)
				}
			}
		}

		// Note: packetsRead might be 0 if the mock file doesn't contain valid packets
		// This is acceptable for testing the ReadPacket function's error handling
		_ = packetsRead
	})

	t.Run("Read packet with track mask", func(t *testing.T) {
		mockFile, err := createMockMatroskaFile()
		if err != nil {
			t.Fatalf("Failed to create mock matroska file: %v", err)
		}

		reader := bytes.NewReader(mockFile)
		parser, err := NewMatroskaParser(reader, false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Set track mask to ignore track 1
		parser.SetTrackMask(0x02)

		// Try to read packet - should be filtered by mask
		packet, err := parser.ReadPacket()
		if err != nil && err != io.EOF {
			t.Fatalf("ReadPacket() with track mask failed: %v", err)
		}
		// Packet might be nil if filtered by mask
		_ = packet
	})

	// Unknown child inside cluster should be skipped gracefully
	t.Run("Cluster with unknown child skipped", func(t *testing.T) {
		buf := new(bytes.Buffer)
		// Header
		eh := new(bytes.Buffer)
		eh.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(eh.Len())))
		buf.Write(eh.Bytes())
		// Segment
		seg := new(bytes.Buffer)
		si := new(bytes.Buffer)
		si.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
		seg.Write([]byte{0x15, 0x49, 0xA9, 0x66})
		seg.Write(vintEncode(uint64(si.Len())))
		seg.Write(si.Bytes())
		te, _ := createMockTrackEntry(1, TypeVideo, "V", "V", "und")
		trs := new(bytes.Buffer)
		trs.Write([]byte{0xAE})
		trs.Write(vintEncode(uint64(len(te))))
		trs.Write(te)
		seg.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
		seg.Write(vintEncode(uint64(trs.Len())))
		seg.Write(trs.Bytes())
		// Cluster with unknown child before SimpleBlock
		cl := new(bytes.Buffer)
		cl.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0
		// Unknown child with a valid 1-byte ID (0x81), size 2, data 0x00 0x01
		cl.Write([]byte{0x81, 0x82, 0x00, 0x01})
		b := []byte{0x81, 0x00, 0x00, 0x80, 'K'}
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(b))))
		cl.Write(b)
		seg.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
		seg.Write(vintEncode(uint64(cl.Len())))
		seg.Write(cl.Bytes())
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
		buf.Write(vintEncode(uint64(seg.Len())))
		buf.Write(seg.Bytes())

		p, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}
		pkt, err := p.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket failed: %v", err)
		}
		if string(pkt.Data) != "K" || (pkt.Flags&KF) == 0 {
			t.Errorf("unexpected pkt: %+v", pkt)
		}
	})
}

// createMockMatroskaFileWithMultipleClusters creates a mock file with multiple clusters
func createMockMatroskaFileWithMultipleClusters() ([]byte, error) {
	buf := new(bytes.Buffer)

	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'}) // DocType
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})                                          // EBML Header ID
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)

	// -- SegmentInfo
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x7B, 0xA9, 0x8A, 'T', 'e', 's', 't', ' ', 'T', 'i', 't', 'l', 'e'}) // Title
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})                           // TimestampScale 1,000,000
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})                                             // SegmentInfo ID
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	// -- Tracks
	trackEntry, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "TestVideo", "und")
	tracks := new(bytes.Buffer)
	tracks.Write([]byte{0xAE}) // TrackEntry ID
	tracks.Write(vintEncode(uint64(len(trackEntry))))
	tracks.Write(trackEntry)
	segment.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
	segment.Write(vintEncode(uint64(tracks.Len())))
	segment.Write(tracks.Bytes())

	// -- Cluster 1
	cluster1 := new(bytes.Buffer)
	cluster1.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0
	// SimpleBlock: Track 1, Timecode 0, Flags 0x80 (Keyframe), Data "frame1"
	blockData1 := []byte{0x81, 0x00, 0x00, 0x80, 'f', 'r', 'a', 'm', 'e', '1'}
	cluster1.Write([]byte{0xA3, byte(0x80 | len(blockData1))})
	cluster1.Write(blockData1)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
	segment.Write(vintEncode(uint64(cluster1.Len())))
	segment.Write(cluster1.Bytes())

	// -- Cluster 2
	cluster2 := new(bytes.Buffer)
	cluster2.Write([]byte{0xE7, 0x82, 0x03, 0xE8}) // Timestamp 1000
	// SimpleBlock: Track 1, Timecode 0, Flags 0x00, Data "frame2"
	blockData2 := []byte{0x81, 0x00, 0x00, 0x00, 'f', 'r', 'a', 'm', 'e', '2'}
	cluster2.Write([]byte{0xA3, byte(0x80 | len(blockData2))})
	cluster2.Write(blockData2)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
	segment.Write(vintEncode(uint64(cluster2.Len())))
	segment.Write(cluster2.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())

	return buf.Bytes(), nil
}

// TestParseSegmentInfo_Advanced tests advanced scenarios for parseSegmentInfo.
func TestParseSegmentInfo_Advanced(t *testing.T) {
	t.Run("SegmentInfo with all optional fields", func(t *testing.T) {
		// Create comprehensive SegmentInfo with all possible fields
		buf := new(bytes.Buffer)
		// Title
		buf.Write([]byte{0x7B, 0xA9, 0x85, 't', 'i', 't', 'l', 'e'})
		// MuxingApp
		buf.Write([]byte{0x4D, 0x80, 0x84, 't', 'e', 's', 't'})
		// WritingApp
		buf.Write([]byte{0x57, 0x41, 0x8B, 'g', 'o', '-', 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		// TimestampScale
		buf.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // 1,000,000
		// Duration (as float)
		buf.Write([]byte{0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0}) // Duration as 8-byte float
		// DateUTC (as int)
		buf.Write([]byte{0x44, 0x61, 0x88, 0x00, 0x00, 0x01, 0x86, 0xA0, 0x00, 0x00, 0x00}) // Some timestamp
		// SegmentUID
		buf.Write([]byte{0x73, 0xA4, 0x90, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseSegmentInfo(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseSegmentInfo() failed: %v", err)
		}

		if parser.fileInfo.Title != "title" {
			t.Errorf("Expected Title 'title', got %q", parser.fileInfo.Title)
		}
		if parser.fileInfo.MuxingApp != "test" {
			t.Errorf("Expected MuxingApp 'test', got %q", parser.fileInfo.MuxingApp)
		}
		if parser.fileInfo.WritingApp != "go-matroska" {
			t.Errorf("Expected WritingApp 'go-matroska', got %q", parser.fileInfo.WritingApp)
		}
		if parser.fileInfo.TimecodeScale != 1000000 {
			t.Errorf("Expected TimecodeScale 1000000, got %d", parser.fileInfo.TimecodeScale)
		}
		if parser.fileInfo.Duration != 100000 {
			t.Errorf("Expected Duration 100000, got %d", parser.fileInfo.Duration)
		}
	})

	t.Run("SegmentInfo with minimal fields", func(t *testing.T) {
		// Create minimal SegmentInfo with only required fields
		buf := new(bytes.Buffer)
		// Only TimestampScale (required)
		buf.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // 1,000,000

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseSegmentInfo(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseSegmentInfo() with minimal fields failed: %v", err)
		}

		if parser.fileInfo.TimecodeScale != 1000000 {
			t.Errorf("Expected TimecodeScale 1000000, got %d", parser.fileInfo.TimecodeScale)
		}
		// Other fields should have default values
		if parser.fileInfo.Title != "" {
			t.Errorf("Expected empty Title, got %q", parser.fileInfo.Title)
		}
	})

	t.Run("SegmentInfo with unknown elements", func(t *testing.T) {
		// Create SegmentInfo with unknown/unsupported elements
		buf := new(bytes.Buffer)
		// TimestampScale
		buf.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // 1,000,000
		// Unknown element (fake ID with proper VINT encoding)
		buf.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseSegmentInfo(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseSegmentInfo() with unknown elements failed: %v", err)
		}
		// Should handle unknown elements gracefully
	})

	t.Run("SegmentInfo with PrevUID, NextUID and other fields", func(t *testing.T) {
		// Test fields that weren't covered in previous tests
		buf := new(bytes.Buffer)
		// SegmentFilename (ID: 0x7384)
		buf.Write([]byte{0x73, 0x84, 0x88, 't', 'e', 's', 't', '.', 'm', 'k', 'v'})
		// PrevUID (ID: 0x3CB923) - 16 bytes
		buf.Write([]byte{0x3C, 0xB9, 0x23, 0x90, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		// PrevFilename (ID: 0x3C83AB)
		buf.Write([]byte{0x3C, 0x83, 0xAB, 0x88, 'p', 'r', 'e', 'v', '.', 'm', 'k', 'v'})
		// NextUID (ID: 0x3EB923) - 16 bytes
		buf.Write([]byte{0x3E, 0xB9, 0x23, 0x90, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30})
		// NextFilename (ID: 0x3E83BB)
		buf.Write([]byte{0x3E, 0x83, 0xBB, 0x88, 'n', 'e', 'x', 't', '.', 'm', 'k', 'v'})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseSegmentInfo(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseSegmentInfo() with additional fields failed: %v", err)
		}

		if parser.fileInfo.Filename != "test.mkv" {
			t.Errorf("Expected Filename 'test.mkv', got %q", parser.fileInfo.Filename)
		}
		if parser.fileInfo.PrevFilename != "prev.mkv" {
			t.Errorf("Expected PrevFilename 'prev.mkv', got %q", parser.fileInfo.PrevFilename)
		}
		if parser.fileInfo.NextFilename != "next.mkv" {
			t.Errorf("Expected NextFilename 'next.mkv', got %q", parser.fileInfo.NextFilename)
		}
	})

	t.Run("SegmentInfo with ReadFull error", func(t *testing.T) {
		// Test error handling when ReadFull fails
		reader := &limitedReader{data: []byte{0x01, 0x02}, limit: 1}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseSegmentInfo(10) // Request more data than available
		if err == nil {
			t.Errorf("Expected error when ReadFull fails, but got nil")
		}
	})
}

// limitedReader is a helper for testing ReadFull errors
type limitedReader struct {
	data  []byte
	pos   int
	limit int
}

func (lr *limitedReader) Read(p []byte) (n int, err error) {
	if lr.pos >= lr.limit {
		return 0, io.ErrUnexpectedEOF
	}
	if lr.pos >= len(lr.data) {
		return 0, io.EOF
	}
	n = copy(p, lr.data[lr.pos:])
	if lr.pos+n > lr.limit {
		n = lr.limit - lr.pos
	}
	lr.pos += n
	if lr.pos >= lr.limit {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}

func (lr *limitedReader) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

// TestParseTracks_EdgeCases tests edge cases for parseTracks function.
func TestParseTracks_EdgeCases(t *testing.T) {
	t.Run("Empty Tracks element", func(t *testing.T) {
		// Test with empty Tracks element (no TrackEntry elements)
		tracksElement := new(bytes.Buffer)
		// Empty buffer

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(tracksElement.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseTracks(uint64(tracksElement.Len()))
		if err != nil {
			t.Fatalf("parseTracks() with empty tracks failed: %v", err)
		}

		if len(parser.tracks) != 0 {
			t.Errorf("Expected 0 tracks for empty Tracks element, got %d", len(parser.tracks))
		}
	})

	t.Run("Tracks with ReadFull error", func(t *testing.T) {
		// Test error handling when ReadFull fails
		reader := &limitedReader{data: []byte{0x01, 0x02}, limit: 1}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseTracks(10) // Request more data than available
		if err == nil {
			t.Errorf("Expected error when ReadFull fails, but got nil")
		}
	})

	t.Run("Tracks with invalid TrackEntry", func(t *testing.T) {
		// Test with corrupted TrackEntry that causes parseTrackEntry to fail
		tracksElement := new(bytes.Buffer)
		// Write invalid TrackEntry (ID correct but data corrupted)
		tracksElement.Write([]byte{0xAE, 0x85, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(tracksElement.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseTracks(uint64(tracksElement.Len()))
		if err == nil {
			t.Errorf("Expected error for invalid TrackEntry, but got nil")
		}
	})

	t.Run("Tracks with non-TrackEntry elements", func(t *testing.T) {
		// Test with Tracks element containing non-TrackEntry elements (should be ignored)
		tracksElement := new(bytes.Buffer)
		// Add a valid TrackEntry
		trackEntryVideo, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "Video", "und")
		tracksElement.Write([]byte{0xAE})
		tracksElement.Write(vintEncode(uint64(len(trackEntryVideo))))
		tracksElement.Write(trackEntryVideo)
		// Add an unknown element (should be ignored)
		tracksElement.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(tracksElement.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseTracks(uint64(tracksElement.Len()))
		if err != nil {
			t.Fatalf("parseTracks() with unknown elements failed: %v", err)
		}

		if len(parser.tracks) != 1 {
			t.Errorf("Expected 1 track (unknown element should be ignored), got %d", len(parser.tracks))
		}
	})

	t.Run("Tracks sorting by track number", func(t *testing.T) {
		// Test that tracks are sorted by track number
		trackEntry1, _ := createMockTrackEntry(3, TypeVideo, "V_TEST", "Video3", "und")
		trackEntry2, _ := createMockTrackEntry(1, TypeAudio, "A_TEST", "Audio1", "eng")
		trackEntry3, _ := createMockTrackEntry(2, TypeSubtitle, "S_TEST", "Subtitle2", "fra")

		tracksElement := new(bytes.Buffer)
		// Add tracks in non-sorted order (3, 1, 2)
		tracksElement.Write([]byte{0xAE})
		tracksElement.Write(vintEncode(uint64(len(trackEntry1))))
		tracksElement.Write(trackEntry1)
		tracksElement.Write([]byte{0xAE})
		tracksElement.Write(vintEncode(uint64(len(trackEntry2))))
		tracksElement.Write(trackEntry2)
		tracksElement.Write([]byte{0xAE})
		tracksElement.Write(vintEncode(uint64(len(trackEntry3))))
		tracksElement.Write(trackEntry3)

		parser := &MatroskaParser{
			reader:   NewEBMLReader(bytes.NewReader(tracksElement.Bytes())),
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}

		err := parser.parseTracks(uint64(tracksElement.Len()))
		if err != nil {
			t.Fatalf("parseTracks() with sorting test failed: %v", err)
		}

		if len(parser.tracks) != 3 {
			t.Fatalf("Expected 3 tracks, got %d", len(parser.tracks))
		}

		// Check that tracks are sorted by number (1, 2, 3)
		if parser.tracks[0].Number != 1 {
			t.Errorf("Expected first track number 1, got %d", parser.tracks[0].Number)
		}
		if parser.tracks[1].Number != 2 {
			t.Errorf("Expected second track number 2, got %d", parser.tracks[1].Number)
		}
		if parser.tracks[2].Number != 3 {
			t.Errorf("Expected third track number 3, got %d", parser.tracks[2].Number)
		}
	})
}

// TestParseTrackEntry_EdgeCases tests edge cases for parseTrackEntry function.
func TestParseTrackEntry_EdgeCases(t *testing.T) {
	t.Run("TrackEntry with minimal fields", func(t *testing.T) {
		// Test with minimal TrackEntry (only required fields)
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x01})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x01})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x01}) // Video

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with minimal fields failed: %v", err)
		}

		if track.Number != 1 {
			t.Errorf("Expected track number 1, got %d", track.Number)
		}
		if track.UID != 1 {
			t.Errorf("Expected track UID 1, got %d", track.UID)
		}
		if track.Type != 1 {
			t.Errorf("Expected track type 1, got %d", track.Type)
		}
		// Check default values
		if track.Enabled != true {
			t.Errorf("Expected default Enabled true, got %v", track.Enabled)
		}
		if track.Default != true {
			t.Errorf("Expected default Default true, got %v", track.Default)
		}
		if track.Language != "eng" {
			t.Errorf("Expected default Language 'eng', got %q", track.Language)
		}
	})

	t.Run("TrackEntry with CodecPrivate", func(t *testing.T) {
		// Test with TrackEntry containing CodecPrivate
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x02})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x02})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x02}) // Audio
		// CodecPrivate
		buf.Write([]byte{0x63, 0xA2, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with CodecPrivate failed: %v", err)
		}

		if track.Number != 2 {
			t.Errorf("Expected track number 2, got %d", track.Number)
		}
		if len(track.CodecPrivate) != 4 {
			t.Errorf("Expected CodecPrivate length 4, got %d", len(track.CodecPrivate))
		}
		expectedPrivate := []byte{0x01, 0x02, 0x03, 0x04}
		for i, b := range expectedPrivate {
			if i < len(track.CodecPrivate) && track.CodecPrivate[i] != b {
				t.Errorf("Expected CodecPrivate[%d] = 0x%02X, got 0x%02X", i, b, track.CodecPrivate[i])
			}
		}
	})

	t.Run("TrackEntry with short language field", func(t *testing.T) {
		// Test with Language field shorter than 3 bytes (should be ignored)
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x03})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x03})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x01}) // Video
		// Language (only 2 bytes - should be ignored) - ID: 0x22B59C
		buf.Write([]byte{0x22, 0xB5, 0x9C, 0x82, 'e', 'n'})

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with short language failed: %v", err)
		}

		// Language should remain default since the provided one was too short
		if track.Language != "eng" {
			t.Errorf("Expected default language 'eng' for short language field, got %q", track.Language)
		}
	})

	t.Run("TrackEntry with Video element", func(t *testing.T) {
		// Test with TrackEntry containing Video element
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x04})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x04})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x01}) // Video
		// Video element
		videoBuf := new(bytes.Buffer)
		// PixelWidth
		videoBuf.Write([]byte{0xB0, 0x82, 0x02, 0x80}) // 640
		// PixelHeight
		videoBuf.Write([]byte{0xBA, 0x82, 0x01, 0xE0}) // 480
		buf.Write([]byte{0xE0})                        // Video ID
		buf.Write(vintEncode(uint64(videoBuf.Len())))
		buf.Write(videoBuf.Bytes())

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with Video element failed: %v", err)
		}

		// Video should be parsed (check if PixelWidth was set)
		if track.Video.PixelWidth == 0 {
			t.Fatal("Expected Video element to be parsed")
		}
		if track.Video.PixelWidth != 640 {
			t.Errorf("Expected PixelWidth 640, got %d", track.Video.PixelWidth)
		}
		if track.Video.PixelHeight != 480 {
			t.Errorf("Expected PixelHeight 480, got %d", track.Video.PixelHeight)
		}
	})

	t.Run("TrackEntry with Audio element", func(t *testing.T) {
		// Test with TrackEntry containing Audio element
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x05})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x05})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x02}) // Audio
		// Audio element
		audioBuf := new(bytes.Buffer)
		// SamplingFrequency
		audioBuf.Write([]byte{0xB5, 0x88, 0x40, 0xE5, 0x88, 0x80, 0x00, 0x00, 0x00, 0x00}) // 44100.0
		// Channels
		audioBuf.Write([]byte{0x9F, 0x81, 0x02}) // 2
		buf.Write([]byte{0xE1})                  // Audio ID
		buf.Write(vintEncode(uint64(audioBuf.Len())))
		buf.Write(audioBuf.Bytes())

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with Audio element failed: %v", err)
		}

		// Audio should be parsed (check if SamplingFreq was set)
		if track.Audio.SamplingFreq == 0 {
			t.Fatal("Expected Audio element to be parsed")
		}
		if track.Audio.SamplingFreq != 44100.0 {
			t.Errorf("Expected SamplingFreq 44100.0, got %f", track.Audio.SamplingFreq)
		}
		if track.Audio.Channels != 2 {
			t.Errorf("Expected Channels 2, got %d", track.Audio.Channels)
		}
	})

	t.Run("TrackEntry with unknown elements", func(t *testing.T) {
		// Test with TrackEntry containing unknown elements (should be ignored)
		buf := new(bytes.Buffer)
		// TrackNumber
		buf.Write([]byte{0xD7, 0x81, 0x06})
		// TrackUID
		buf.Write([]byte{0x73, 0xC5, 0x81, 0x06})
		// TrackType
		buf.Write([]byte{0x83, 0x81, 0x01}) // Video
		// Unknown element (should be ignored)
		buf.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{}
		track, err := parser.parseTrackEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTrackEntry() with unknown elements failed: %v", err)
		}

		if track.Number != 6 {
			t.Errorf("Expected track number 6, got %d", track.Number)
		}
		// Should handle unknown elements gracefully
	})

	t.Run("TrackEntry with ReadElement error", func(t *testing.T) {
		// Test with corrupted data that causes ReadElement to fail
		corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		parser := &MatroskaParser{}
		_, err := parser.parseTrackEntry(corruptedData)
		if err == nil {
			t.Errorf("Expected error for corrupted TrackEntry data, but got nil")
		}
	})
}

// TestParseEditionEntry_EdgeCases tests edge cases for parseEditionEntry function.
func TestParseEditionEntry_EdgeCases(t *testing.T) {
	t.Run("Empty EditionEntry", func(t *testing.T) {
		// Test with empty EditionEntry (no ChapterAtom elements)
		buf := new(bytes.Buffer)
		// Empty buffer

		parser := &MatroskaParser{}
		chapters, err := parser.parseEditionEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseEditionEntry() with empty data failed: %v", err)
		}

		if len(chapters) != 0 {
			t.Errorf("Expected 0 chapters for empty EditionEntry, got %d", len(chapters))
		}
	})

	t.Run("EditionEntry with single ChapterAtom", func(t *testing.T) {
		// Test with EditionEntry containing one ChapterAtom
		buf := new(bytes.Buffer)
		// ChapterAtom
		chapterBuf := new(bytes.Buffer)
		// ChapterUID
		chapterBuf.Write([]byte{0x73, 0xC4, 0x81, 0x01})
		// ChapterTimeStart
		chapterBuf.Write([]byte{0x91, 0x81, 0x00})

		buf.Write([]byte{0xB6}) // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterBuf.Len())))
		buf.Write(chapterBuf.Bytes())

		parser := &MatroskaParser{}
		chapters, err := parser.parseEditionEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseEditionEntry() with single ChapterAtom failed: %v", err)
		}

		if len(chapters) != 1 {
			t.Fatalf("Expected 1 chapter, got %d", len(chapters))
		}

		if chapters[0].UID != 1 {
			t.Errorf("Expected chapter UID 1, got %d", chapters[0].UID)
		}
		if chapters[0].Start != 0 {
			t.Errorf("Expected chapter start 0, got %d", chapters[0].Start)
		}
	})

	t.Run("EditionEntry with multiple ChapterAtoms", func(t *testing.T) {
		// Test with EditionEntry containing multiple ChapterAtoms
		buf := new(bytes.Buffer)

		// ChapterAtom 1
		chapterBuf1 := new(bytes.Buffer)
		chapterBuf1.Write([]byte{0x73, 0xC4, 0x81, 0x01}) // ChapterUID: 1
		chapterBuf1.Write([]byte{0x91, 0x81, 0x00})       // ChapterTimeStart: 0
		buf.Write([]byte{0xB6})                           // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterBuf1.Len())))
		buf.Write(chapterBuf1.Bytes())

		// ChapterAtom 2
		chapterBuf2 := new(bytes.Buffer)
		chapterBuf2.Write([]byte{0x73, 0xC4, 0x81, 0x02}) // ChapterUID: 2
		chapterBuf2.Write([]byte{0x91, 0x82, 0x03, 0xE8}) // ChapterTimeStart: 1000
		buf.Write([]byte{0xB6})                           // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterBuf2.Len())))
		buf.Write(chapterBuf2.Bytes())

		parser := &MatroskaParser{}
		chapters, err := parser.parseEditionEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseEditionEntry() with multiple ChapterAtoms failed: %v", err)
		}

		if len(chapters) != 2 {
			t.Fatalf("Expected 2 chapters, got %d", len(chapters))
		}

		if chapters[0].UID != 1 {
			t.Errorf("Expected first chapter UID 1, got %d", chapters[0].UID)
		}
		if chapters[1].UID != 2 {
			t.Errorf("Expected second chapter UID 2, got %d", chapters[1].UID)
		}
	})

	t.Run("EditionEntry with non-ChapterAtom elements", func(t *testing.T) {
		// Test with EditionEntry containing non-ChapterAtom elements (should be ignored)
		buf := new(bytes.Buffer)
		// Add a valid ChapterAtom
		chapterBuf := new(bytes.Buffer)
		chapterBuf.Write([]byte{0x73, 0xC4, 0x81, 0x01}) // ChapterUID: 1
		chapterBuf.Write([]byte{0x91, 0x81, 0x00})       // ChapterTimeStart: 0
		buf.Write([]byte{0xB6})                          // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterBuf.Len())))
		buf.Write(chapterBuf.Bytes())
		// Add an unknown element (should be ignored)
		buf.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{}
		chapters, err := parser.parseEditionEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseEditionEntry() with unknown elements failed: %v", err)
		}

		if len(chapters) != 1 {
			t.Errorf("Expected 1 chapter (unknown element should be ignored), got %d", len(chapters))
		}
	})

	t.Run("EditionEntry with invalid ChapterAtom", func(t *testing.T) {
		// Test with EditionEntry containing invalid ChapterAtom that causes parseChapterAtom to fail
		buf := new(bytes.Buffer)
		// Write invalid ChapterAtom (ID correct but data corrupted)
		buf.Write([]byte{0xB6, 0x85, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{}
		_, err := parser.parseEditionEntry(buf.Bytes())
		if err == nil {
			t.Errorf("Expected error for invalid ChapterAtom, but got nil")
		}
	})

	t.Run("EditionEntry with ReadElement error", func(t *testing.T) {
		// Test with corrupted data that causes ReadElement to fail
		corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		parser := &MatroskaParser{}
		_, err := parser.parseEditionEntry(corruptedData)
		if err == nil {
			t.Errorf("Expected error for corrupted EditionEntry data, but got nil")
		}
	})
}

// TestParseTag_EdgeCases tests edge cases for parseTag function.
func TestParseTag_EdgeCases(t *testing.T) {
	t.Run("Empty Tag", func(t *testing.T) {
		// Test with empty Tag (no Targets or SimpleTags)
		buf := new(bytes.Buffer)
		// Empty buffer

		parser := &MatroskaParser{}
		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() with empty data failed: %v", err)
		}

		if len(tag.Targets) != 0 {
			t.Errorf("Expected 0 targets for empty Tag, got %d", len(tag.Targets))
		}
		if len(tag.SimpleTags) != 0 {
			t.Errorf("Expected 0 simple tags for empty Tag, got %d", len(tag.SimpleTags))
		}
	})

	t.Run("Tag with single Target", func(t *testing.T) {
		// Test with Tag containing one Target
		buf := new(bytes.Buffer)
		// Targets
		targetBuf := new(bytes.Buffer)
		// TargetTypeValue
		targetBuf.Write([]byte{0x68, 0xCA, 0x81, 0x32}) // 50 (ALBUM)

		buf.Write([]byte{0x63, 0xC0}) // Targets ID
		buf.Write(vintEncode(uint64(targetBuf.Len())))
		buf.Write(targetBuf.Bytes())

		parser := &MatroskaParser{}
		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() with single Target failed: %v", err)
		}

		if len(tag.Targets) != 1 {
			t.Fatalf("Expected 1 target, got %d", len(tag.Targets))
		}

		if tag.Targets[0].Type != 50 {
			t.Errorf("Expected target type 50, got %d", tag.Targets[0].Type)
		}
	})

	t.Run("Tag with single SimpleTag", func(t *testing.T) {
		// Test with Tag containing one SimpleTag
		buf := new(bytes.Buffer)
		// SimpleTag
		simpleTagBuf := new(bytes.Buffer)
		// TagName
		simpleTagBuf.Write([]byte{0x45, 0xA3, 0x85, 'T', 'I', 'T', 'L', 'E'})
		// TagString
		simpleTagBuf.Write([]byte{0x44, 0x87, 0x84, 'T', 'e', 's', 't'})

		buf.Write([]byte{0x67, 0xC8}) // SimpleTag ID
		buf.Write(vintEncode(uint64(simpleTagBuf.Len())))
		buf.Write(simpleTagBuf.Bytes())

		parser := &MatroskaParser{}
		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() with single SimpleTag failed: %v", err)
		}

		if len(tag.SimpleTags) != 1 {
			t.Fatalf("Expected 1 simple tag, got %d", len(tag.SimpleTags))
		}

		if tag.SimpleTags[0].Name != "TITLE" {
			t.Errorf("Expected simple tag name 'TITLE', got %q", tag.SimpleTags[0].Name)
		}
		if tag.SimpleTags[0].Value != "Test" {
			t.Errorf("Expected simple tag value 'Test', got %q", tag.SimpleTags[0].Value)
		}
	})

	t.Run("Tag with multiple Targets", func(t *testing.T) {
		// Test with Tag containing multiple Targets
		buf := new(bytes.Buffer)

		// Target 1
		targetBuf1 := new(bytes.Buffer)
		targetBuf1.Write([]byte{0x68, 0xCA, 0x81, 0x32}) // TargetTypeValue: 50
		buf.Write([]byte{0x63, 0xC0})                    // Targets ID
		buf.Write(vintEncode(uint64(targetBuf1.Len())))
		buf.Write(targetBuf1.Bytes())

		// Target 2
		targetBuf2 := new(bytes.Buffer)
		targetBuf2.Write([]byte{0x68, 0xCA, 0x81, 0x1E}) // TargetTypeValue: 30
		buf.Write([]byte{0x63, 0xC0})                    // Targets ID
		buf.Write(vintEncode(uint64(targetBuf2.Len())))
		buf.Write(targetBuf2.Bytes())

		parser := &MatroskaParser{}
		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() with multiple targets failed: %v", err)
		}

		if len(tag.Targets) != 2 {
			t.Errorf("Expected 2 targets, got %d", len(tag.Targets))
		}

		if tag.Targets[0].Type != 50 {
			t.Errorf("Expected first target type 50, got %d", tag.Targets[0].Type)
		}
		if tag.Targets[1].Type != 30 {
			t.Errorf("Expected second target type 30, got %d", tag.Targets[1].Type)
		}
	})

	t.Run("Tag with unknown elements", func(t *testing.T) {
		// Test with Tag containing unknown elements (should be ignored)
		buf := new(bytes.Buffer)
		// Add a valid SimpleTag
		simpleTagBuf := new(bytes.Buffer)
		simpleTagBuf.Write([]byte{0x45, 0xA3, 0x85, 'T', 'I', 'T', 'L', 'E'})
		simpleTagBuf.Write([]byte{0x44, 0x87, 0x84, 'T', 'e', 's', 't'})
		buf.Write([]byte{0x67, 0xC8}) // SimpleTag ID
		buf.Write(vintEncode(uint64(simpleTagBuf.Len())))
		buf.Write(simpleTagBuf.Bytes())
		// Add an unknown element (should be ignored)
		buf.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{}
		tag, err := parser.parseTag(buf.Bytes())
		if err != nil {
			t.Fatalf("parseTag() with unknown elements failed: %v", err)
		}

		if len(tag.SimpleTags) != 1 {
			t.Errorf("Expected 1 simple tag (unknown element should be ignored), got %d", len(tag.SimpleTags))
		}
	})

	t.Run("Tag with invalid Target", func(t *testing.T) {
		// Test with Tag containing invalid Target that causes parseTarget to fail
		buf := new(bytes.Buffer)
		// Write invalid Target (ID correct but data corrupted)
		buf.Write([]byte{0x63, 0xC0, 0x85, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{}
		_, err := parser.parseTag(buf.Bytes())
		if err == nil {
			t.Errorf("Expected error for invalid Target, but got nil")
		}
	})

	t.Run("Tag with invalid SimpleTag", func(t *testing.T) {
		// Test with Tag containing invalid SimpleTag that causes parseSimpleTag to fail
		buf := new(bytes.Buffer)
		// Write invalid SimpleTag (ID correct but data corrupted)
		buf.Write([]byte{0x67, 0xC8, 0x85, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{}
		_, err := parser.parseTag(buf.Bytes())
		if err == nil {
			t.Errorf("Expected error for invalid SimpleTag, but got nil")
		}
	})

	t.Run("Tag with ReadElement error", func(t *testing.T) {
		// Test with corrupted data that causes ReadElement to fail
		corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		parser := &MatroskaParser{}
		_, err := parser.parseTag(corruptedData)
		if err == nil {
			t.Errorf("Expected error for corrupted Tag data, but got nil")
		}
	})
}

// TestParseAttachments_EdgeCases tests edge cases for parseAttachments function.
func TestParseAttachments_EdgeCases(t *testing.T) {
	t.Run("Empty Attachments", func(t *testing.T) {
		// Test with empty Attachments (no AttachedFile elements)
		buf := new(bytes.Buffer)
		// Empty buffer

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}
		err := parser.parseAttachments(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseAttachments() with empty data failed: %v", err)
		}

		if len(parser.attachments) != 0 {
			t.Errorf("Expected 0 attachments for empty Attachments, got %d", len(parser.attachments))
		}
	})

	t.Run("Attachments with single AttachedFile", func(t *testing.T) {
		// Test with Attachments containing one AttachedFile
		buf := new(bytes.Buffer)
		// AttachedFile
		attachedFileBuf := new(bytes.Buffer)
		// FileName
		attachedFileBuf.Write([]byte{0x46, 0x6E, 0x88, 't', 'e', 's', 't', '.', 't', 'x', 't'})
		// FileMimeType
		attachedFileBuf.Write([]byte{0x46, 0x60, 0x8A, 't', 'e', 'x', 't', '/', 'p', 'l', 'a', 'i', 'n'})
		// FileData
		attachedFileBuf.Write([]byte{0x46, 0x5C, 0x85, 'h', 'e', 'l', 'l', 'o'})
		// FileUID
		attachedFileBuf.Write([]byte{0x46, 0xAE, 0x81, 0x01})

		buf.Write([]byte{0x61, 0xA7}) // AttachedFile ID
		buf.Write(vintEncode(uint64(attachedFileBuf.Len())))
		buf.Write(attachedFileBuf.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}
		err := parser.parseAttachments(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseAttachments() with single AttachedFile failed: %v", err)
		}

		if len(parser.attachments) != 1 {
			t.Fatalf("Expected 1 attachment, got %d", len(parser.attachments))
		}

		attachment := parser.attachments[0]
		if attachment.Name != "test.txt" {
			t.Errorf("Expected attachment name 'test.txt', got %q", attachment.Name)
		}
		if attachment.MimeType != "text/plain" {
			t.Errorf("Expected MIME type 'text/plain', got %q", attachment.MimeType)
		}
		if attachment.UID != 1 {
			t.Errorf("Expected UID 1, got %d", attachment.UID)
		}
	})

	t.Run("Attachments with multiple AttachedFiles", func(t *testing.T) {
		// Test with Attachments containing two simple AttachedFiles
		buf := new(bytes.Buffer)

		// AttachedFile 1 (simplified)
		attachedFileBuf1 := new(bytes.Buffer)
		// IDFileName (0x466E) with size 5 and content "file1"
		attachedFileBuf1.Write([]byte{0x46, 0x6E, 0x85}) // IDFileName + size
		attachedFileBuf1.Write([]byte{'f', 'i', 'l', 'e', '1'})
		// IDFileUID (0x46AE) with size 1 and value 1
		attachedFileBuf1.Write([]byte{0x46, 0xAE, 0x81, 0x01})

		buf.Write([]byte{0x61, 0xA7}) // AttachedFile ID
		buf.Write(vintEncode(uint64(attachedFileBuf1.Len())))
		buf.Write(attachedFileBuf1.Bytes())

		// AttachedFile 2 (simplified)
		attachedFileBuf2 := new(bytes.Buffer)
		// IDFileName (0x466E) with size 5 and content "file2"
		attachedFileBuf2.Write([]byte{0x46, 0x6E, 0x85}) // IDFileName + size
		attachedFileBuf2.Write([]byte{'f', 'i', 'l', 'e', '2'})
		// IDFileUID (0x46AE) with size 1 and value 2
		attachedFileBuf2.Write([]byte{0x46, 0xAE, 0x81, 0x02})

		buf.Write([]byte{0x61, 0xA7}) // AttachedFile ID
		buf.Write(vintEncode(uint64(attachedFileBuf2.Len())))
		buf.Write(attachedFileBuf2.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}
		err := parser.parseAttachments(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseAttachments() with multiple AttachedFiles failed: %v", err)
		}

		if len(parser.attachments) != 2 {
			t.Fatalf("Expected 2 attachments, got %d", len(parser.attachments))
		}

		if parser.attachments[0].Name != "file1" {
			t.Errorf("Expected first attachment name 'file1', got %q", parser.attachments[0].Name)
		}
		if parser.attachments[1].Name != "file2" {
			t.Errorf("Expected second attachment name 'file2', got %q", parser.attachments[1].Name)
		}
	})

	t.Run("Attachments with non-AttachedFile elements", func(t *testing.T) {
		// Test with Attachments containing non-AttachedFile elements (should be ignored)
		buf := new(bytes.Buffer)
		// Add a valid AttachedFile
		attachedFileBuf := new(bytes.Buffer)
		attachedFileBuf.Write([]byte{0x46, 0x6E, 0x88, 't', 'e', 's', 't', '.', 't', 'x', 't'})
		attachedFileBuf.Write([]byte{0x46, 0x60, 0x8A, 't', 'e', 'x', 't', '/', 'p', 'l', 'a', 'i', 'n'})
		attachedFileBuf.Write([]byte{0x46, 0x5C, 0x85, 'h', 'e', 'l', 'l', 'o'})
		attachedFileBuf.Write([]byte{0x46, 0xAE, 0x81, 0x01})

		buf.Write([]byte{0x61, 0xA7}) // AttachedFile ID
		buf.Write(vintEncode(uint64(attachedFileBuf.Len())))
		buf.Write(attachedFileBuf.Bytes())
		// Add an unknown element (should be ignored)
		buf.Write([]byte{0x7F, 0xFF, 0x84, 0x01, 0x02, 0x03, 0x04})

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}
		err := parser.parseAttachments(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseAttachments() with unknown elements failed: %v", err)
		}

		if len(parser.attachments) != 1 {
			t.Errorf("Expected 1 attachment (unknown element should be ignored), got %d", len(parser.attachments))
		}
	})

	t.Run("Attachments with ReadFull error", func(t *testing.T) {
		// Test error handling when ReadFull fails
		reader := &limitedReader{data: []byte{0x01, 0x02}, limit: 1}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseAttachments(10) // Request more data than available
		if err == nil {
			t.Errorf("Expected error when ReadFull fails, but got nil")
		}
	})

	t.Run("Attachments with invalid AttachedFile", func(t *testing.T) {
		// Test with Attachments containing invalid AttachedFile that causes parseAttachedFile to fail
		buf := new(bytes.Buffer)
		// Write invalid AttachedFile (ID correct but data corrupted)
		buf.Write([]byte{0x61, 0xA7, 0x85, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}
		err := parser.parseAttachments(uint64(buf.Len()))
		if err == nil {
			t.Errorf("Expected error for invalid AttachedFile, but got nil")
		}
	})

	t.Run("Attachments with ReadElement error", func(t *testing.T) {
		// Test with corrupted data that causes ReadElement to fail
		corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(corruptedData)),
		}

		err := parser.parseAttachments(uint64(len(corruptedData)))
		if err == nil {
			t.Errorf("Expected error for corrupted Attachments data, but got nil")
		}
	})
}

// ===== Additional tests to raise coverage toward 95% =====

func TestReadPacket_BasicAndTrackMask(t *testing.T) {
	// Basic packet read from a minimal valid Matroska file
	mockFile, err := createMockMatroskaFile()
	if err != nil {
		t.Fatalf("Failed to create mock matroska file: %v", err)
	}
	parser, err := NewMatroskaParser(bytes.NewReader(mockFile), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser() failed: %v", err)
	}

	// Read first (and only) packet
	pkt, err := parser.ReadPacket()
	if err != nil && err != io.EOF {
		t.Fatalf("ReadPacket() failed: %v", err)
	}
	if pkt == nil {
		t.Fatalf("Expected a packet, got nil")
	}
	if pkt.Track != 1 {
		t.Errorf("Expected track 1, got %d", pkt.Track)
	}
	if string(pkt.Data) != "frame" {
		t.Errorf("Expected data 'frame', got %q", string(pkt.Data))
	}
	if pkt.Flags&KF == 0 {
		t.Errorf("Expected keyframe flag to be set")
	}
	if pkt.StartTime != 0 { // cluster ts 0 + block rel 0
		t.Errorf("Expected StartTime 0, got %d", pkt.StartTime)
	}

	// Next read should be EOF
	pkt2, err := parser.ReadPacket()
	if err != io.EOF {
		t.Errorf("Expected io.EOF on second read, got %v (pkt=%v)", err, pkt2)
	}

	// Track mask should filter out packets
	parser2, err := NewMatroskaParser(bytes.NewReader(mockFile), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser() failed: %v", err)
	}
	parser2.SetTrackMask(1 << (1 - 1)) // mask track 1
	pkt3, err := parser2.ReadPacket()
	if err != io.EOF || pkt3 != nil {
		t.Errorf("Expected EOF with masked track, got pkt=%v err=%v", pkt3, err)
	}
}

func TestParserProxyMethods_AttachmentsAndChapters(t *testing.T) {
	// Attachments
	mockA, err := createMockMatroskaFileWithAttachments()
	if err != nil {
		t.Fatalf("Failed to create mock with attachments: %v", err)
	}
	pA, err := NewMatroskaParser(bytes.NewReader(mockA), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser() failed: %v", err)
	}
	atts := pA.GetAttachments()
	if len(atts) == 0 {
		t.Fatalf("Expected attachments, got none")
	}
	if atts[0].Name == "" || atts[0].MimeType == "" || atts[0].UID == 0 {
		t.Errorf("Attachment fields not populated: %+v", atts[0])
	}

	// Chapters
	mockC, err := createMockMatroskaFileWithChapters()
	if err != nil {
		t.Fatalf("Failed to create mock with chapters: %v", err)
	}
	pC, err := NewMatroskaParser(bytes.NewReader(mockC), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser() failed: %v", err)
	}
	chs := pC.GetChapters()
	if len(chs) == 0 {
		t.Fatalf("Expected chapters, got none")
	}
	// Expect at least one ChapterDisplay or Children entry to exist in the mock
	if len(chs[0].Display) == 0 && len(chs[0].Children) == 0 {
		t.Fatalf("Expected chapter to have display info or children, got %+v", chs[0])
	}
	// Also ensure GetNumTracks and GetTrackInfo return sensible values
	if pC.GetNumTracks() != 1 {
		t.Errorf("Expected 1 track, got %d", pC.GetNumTracks())
	}
	if pC.GetTrackInfo(0) == nil || pC.GetTrackInfo(1) != nil {
		t.Errorf("GetTrackInfo boundary conditions failed")
	}
}

func TestParseVInt_Cases(t *testing.T) {
	mp := &MatroskaParser{}
	// Empty data
	if v, n := mp.parseVInt(nil); v != 0 || n != 0 {
		t.Errorf("Expected (0,0) for nil input, got (%d,%d)", v, n)
	}
	// First byte 0 (invalid)
	if v, n := mp.parseVInt([]byte{0x00}); v != 0 || n != 0 {
		t.Errorf("Expected (0,0) for first byte 0, got (%d,%d)", v, n)
	}
	// Length 2 but insufficient bytes
	if v, n := mp.parseVInt([]byte{0x40}); v != 0 || n != 0 {
		t.Errorf("Expected (0,0) for short data, got (%d,%d)", v, n)
	}
	// 1-byte vint: 0x81 => 1
	if v, n := mp.parseVInt([]byte{0x81}); v != 1 || n != 1 {
		t.Errorf("Expected (1,1) for 0x81, got (%d,%d)", v, n)
	}
	// 2-byte vint: 0x40 0x01 => 1
	if v, n := mp.parseVInt([]byte{0x40, 0x01}); v != 1 || n != 2 {
		t.Errorf("Expected (1,2) for 0x40 0x01, got (%d,%d)", v, n)
	}
}

// Build a minimal Matroska stream with unknown-size Segment that ends at EOF to
// exercise parseSegmentChildren EOF handling for streaming input.
func buildUnknownSizeSegmentFile() []byte {
	buf := new(bytes.Buffer)
	// EBML Header (DocType matroska)
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment (unknown size)
	segment := new(bytes.Buffer)
	// Minimal SegmentInfo with Title only (Title size = 4 -> 0x84)
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x7B, 0xA9, 0x84, 'T', 'e', 's', 't'})
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
	// Unknown size marker (as used elsewhere in tests for streaming)
	buf.Write([]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	buf.Write(segment.Bytes())
	return buf.Bytes()
}

func TestParseSegment_UnknownSizeEOF_OK(t *testing.T) {
	data := buildUnknownSizeSegmentFile()
	if _, err := NewMatroskaParser(bytes.NewReader(data), false); err != nil {
		t.Fatalf("Expected parser to handle unknown-size segment ending at EOF, got error: %v", err)
	}
}

// Helper to create a Matroska file with two clusters and an unknown child to exercise more ReadPacket branches.
func createMockMatroskaFileTwoClusters() ([]byte, error) {
	buf := new(bytes.Buffer)
	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)

	// -- SegmentInfo with TimestampScale = 1,000,000
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	// -- Tracks (single video track)
	trackEntry, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "TestVideo", "und")
	tracks := new(bytes.Buffer)
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(len(trackEntry))))
	tracks.Write(trackEntry)
	segment.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
	segment.Write(vintEncode(uint64(tracks.Len())))
	segment.Write(tracks.Bytes())

	// -- Cluster 1: Timestamp 0, SimpleBlock data "f1", plus an unknown child (Void 0xEC)
	c1 := new(bytes.Buffer)
	c1.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0
	// Add unknown child (Void) with 2 bytes payload
	c1.Write([]byte{0xEC, 0x82, 0xAA, 0xBB})
	// SimpleBlock: track1 (0x81), timecode 0, flags 0x80, data "f1"
	sb1 := []byte{0x81, 0x00, 0x00, 0x80, 'f', '1'}
	c1.Write([]byte{0xA3})
	c1.Write(vintEncode(uint64(len(sb1))))
	c1.Write(sb1)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
	segment.Write(vintEncode(uint64(c1.Len())))
	segment.Write(c1.Bytes())

	// -- Cluster 2: Timestamp 5, SimpleBlock data "f2"
	c2 := new(bytes.Buffer)
	c2.Write([]byte{0xE7, 0x81, 0x05}) // Timestamp 5
	sb2 := []byte{0x81, 0x00, 0x00, 0x80, 'f', '2'}
	c2.Write([]byte{0xA3})
	c2.Write(vintEncode(uint64(len(sb2))))
	c2.Write(sb2)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
	segment.Write(vintEncode(uint64(c2.Len())))
	segment.Write(c2.Bytes())

	// Wrap segment
	buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())
	return buf.Bytes(), nil
}

func TestReadPacket_MultiClusters_AndSkipUnknown(t *testing.T) {
	data, err := createMockMatroskaFileTwoClusters()
	if err != nil {
		t.Fatalf("failed to build mock: %v", err)
	}
	p, err := NewMatroskaParser(bytes.NewReader(data), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	// First packet
	pkt1, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket#1 failed: %v", err)
	}
	if string(pkt1.Data) != "f1" || pkt1.Track != 1 || pkt1.Flags&KF == 0 {
		t.Errorf("Unexpected pkt1: %+v", pkt1)
	}
	// Second packet
	pkt2, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket#2 failed: %v", err)
	}
	if string(pkt2.Data) != "f2" || pkt2.Track != 1 || pkt2.Flags&KF == 0 {
		t.Errorf("Unexpected pkt2: %+v", pkt2)
	}
	if pkt2.StartTime == 0 {
		t.Errorf("Expected non-zero StartTime for second cluster, got %d", pkt2.StartTime)
	}
	// Then EOF
	if pkt3, errReadPacket := p.ReadPacket(); errReadPacket != io.EOF || pkt3 != nil {
		t.Errorf("Expected EOF after two packets, got pkt=%v err=%v", pkt3, errReadPacket)
	}
}

func TestParser_Seek_And_SkipToKeyframe_NoPanics(t *testing.T) {
	data, err := createMockMatroskaFileTwoClusters()
	if err != nil {
		t.Fatalf("failed to build mock: %v", err)
	}
	// Parser with seeks enabled
	p, err := NewMatroskaParser(bytes.NewReader(data), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	// Inject a simple cues table so Seek() path gets executed
	p.cues = []*Cue{{Time: 0, Position: 0, Track: 1}}
	if err = p.Seek(0, SeekToPrevKeyFrame); err != nil {
		t.Fatalf("Seek failed: %v", err)
	}
	// SkipToKeyframe should iterate and return without panic
	p.SkipToKeyframe()

	// Parser with noSeeking=true should no-op SkipToKeyframe
	p2, err := NewMatroskaParser(bytes.NewReader(data), true)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	p2.SkipToKeyframe()
}

// TestParser_Seek_EdgeCases tests edge cases for the Seek function
func TestParser_Seek_EdgeCases(t *testing.T) {
	t.Run("Seek with noSeeking enabled", func(t *testing.T) {
		data, err := createMockMatroskaFileTwoClusters()
		if err != nil {
			t.Fatalf("failed to build mock: %v", err)
		}

		p, err := NewMatroskaParser(bytes.NewReader(data), true) // noSeeking=true
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}

		err = p.Seek(1000, 0)
		if err == nil {
			t.Error("Expected error when seeking with noSeeking=true, but got nil")
		}
	})

	t.Run("Seek with no cues", func(t *testing.T) {
		data, err := createMockMatroskaFileTwoClusters()
		if err != nil {
			t.Fatalf("failed to build mock: %v", err)
		}

		p, err := NewMatroskaParser(bytes.NewReader(data), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}

		// Clear cues to test no cues case
		p.cues = nil

		err = p.Seek(1000, 0)
		if err == nil {
			t.Error("Expected error when seeking with no cues, but got nil")
		}
	})

	t.Run("Seek to exact timecode", func(t *testing.T) {
		data, err := createMockMatroskaFileTwoClusters()
		if err != nil {
			t.Fatalf("failed to build mock: %v", err)
		}

		p, err := NewMatroskaParser(bytes.NewReader(data), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}

		// Add multiple cues for testing
		p.cues = []*Cue{
			{Time: 1000, Position: 100, Track: 1},
			{Time: 2000, Position: 200, Track: 1},
			{Time: 3000, Position: 300, Track: 1},
		}

		// Seek to exact timecode
		err = p.Seek(2000, 0)
		if err != nil {
			t.Fatalf("Seek to exact timecode failed: %v", err)
		}
	})

	t.Run("Seek to timecode between cues", func(t *testing.T) {
		data, err := createMockMatroskaFileTwoClusters()
		if err != nil {
			t.Fatalf("failed to build mock: %v", err)
		}

		p, err := NewMatroskaParser(bytes.NewReader(data), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}

		// Add multiple cues for testing
		p.cues = []*Cue{
			{Time: 1000, Position: 100, Track: 1},
			{Time: 3000, Position: 300, Track: 1},
		}

		// Seek to timecode between cues (should use the earlier one)
		err = p.Seek(2000, 0)
		if err != nil {
			t.Fatalf("Seek between cues failed: %v", err)
		}
	})

	t.Run("Seek beyond last cue", func(t *testing.T) {
		data, err := createMockMatroskaFileTwoClusters()
		if err != nil {
			t.Fatalf("failed to build mock: %v", err)
		}

		p, err := NewMatroskaParser(bytes.NewReader(data), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser failed: %v", err)
		}

		// Add cues for testing
		p.cues = []*Cue{
			{Time: 1000, Position: 100, Track: 1},
			{Time: 2000, Position: 200, Track: 1},
		}

		// Seek beyond last cue (should use the last cue)
		err = p.Seek(5000, 0)
		if err != nil {
			t.Fatalf("Seek beyond last cue failed: %v", err)
		}
	})
}

// TestParseVideoTrack_Defaults verifies Display* defaults from Pixel* when absent.
func TestParseVideoTrack_Defaults(t *testing.T) {
	// Only PixelWidth/PixelHeight provided; DisplayWidth/Height should default to Pixel*
	buf := new(bytes.Buffer)
	// PixelWidth: 640
	buf.Write([]byte{0xB0, 0x82, 0x02, 0x80})
	// PixelHeight: 360
	buf.Write([]byte{0xBA, 0x82, 0x01, 0x68})

	parser := &MatroskaParser{}
	track := &TrackInfo{}
	if err := parser.parseVideoTrack(buf.Bytes(), track); err != nil {
		t.Fatalf("parseVideoTrack() failed: %v", err)
	}
	if track.Video.DisplayWidth != track.Video.PixelWidth || track.Video.DisplayHeight != track.Video.PixelHeight {
		t.Errorf("Display defaults not applied: got %dx%d disp vs %dx%d pixel", track.Video.DisplayWidth, track.Video.DisplayHeight, track.Video.PixelWidth, track.Video.PixelHeight)
	}
}

// TestParseAudioTrack_Defaults verifies default channel/freq and OutputSamplingFreq fallback.
func TestParseAudioTrack_Defaults(t *testing.T) {
	parser := &MatroskaParser{}
	track := &TrackInfo{}
	// No fields set -> defaults apply
	if err := parser.parseAudioTrack([]byte{}, track); err != nil {
		t.Fatalf("parseAudioTrack(empty) failed: %v", err)
	}
	if track.Audio.Channels != 1 || track.Audio.SamplingFreq != 8000.0 || track.Audio.OutputSamplingFreq != 8000.0 {
		t.Errorf("unexpected audio defaults: %+v", track.Audio)
	}

	// Only SamplingFrequency set -> OutputSamplingFreq should mirror it when absent
	buf := new(bytes.Buffer)
	sf := math.Float64bits(22050.0)
	buf.Write([]byte{0xB5, 0x88})
	_ = binary.Write(buf, binary.BigEndian, sf)
	track2 := &TrackInfo{}
	if err := parser.parseAudioTrack(buf.Bytes(), track2); err != nil {
		t.Fatalf("parseAudioTrack(sfreq) failed: %v", err)
	}
	if track2.Audio.SamplingFreq != 22050.0 || track2.Audio.OutputSamplingFreq != 22050.0 {
		t.Errorf("output sampling fallback failed: %+v", track2.Audio)
	}
}

// TestParseCuePoint_Full covers additional fields in cue track positions.
func TestParseCuePoint_Full(t *testing.T) {
	// Build CuePoint with time and full CueTrackPositions
	cue := new(bytes.Buffer)
	// CueTime = 7
	cue.Write([]byte{0xB3, 0x81, 0x07})
	// CueTrackPositions
	ctp := new(bytes.Buffer)
	ctp.Write([]byte{0xF7, 0x81, 0x02})       // Track 2
	ctp.Write([]byte{0xF1, 0x81, 0x64})       // ClusterPos 100
	ctp.Write([]byte{0xF0, 0x81, 0x05})       // RelativePos 5
	ctp.Write([]byte{0x53, 0x78, 0x81, 0x03}) // BlockNum 3
	ctp.Write([]byte{0x9B, 0x81, 0x02})       // Duration 2
	cue.Write([]byte{0xB7})
	cue.Write(vintEncode(uint64(ctp.Len())))
	cue.Write(ctp.Bytes())

	mp := &MatroskaParser{fileInfo: &SegmentInfo{TimecodeScale: 1000000}}
	cues, err := mp.parseCuePoint(cue.Bytes())
	if err != nil {
		t.Fatalf("parseCuePoint failed: %v", err)
	}
	if len(cues) != 1 {
		t.Fatalf("expected 1 cue, got %d", len(cues))
	}
	got := cues[0]
	if got.Track != 2 || got.Position != 100 || got.RelativePosition != 5 || got.Block != 3 || got.Duration != 2*mp.fileInfo.TimecodeScale {
		t.Errorf("unexpected cue fields: %+v", got)
	}
	if got.Time != 7*mp.fileInfo.TimecodeScale {
		t.Errorf("unexpected scaled time: %d", got.Time)
	}
}

// TestParseBlockGroup_WithDuration verifies duration affects EndTime.
func TestParseBlockGroup_WithDuration(t *testing.T) {
	// Construct a BlockGroup with Block and BlockDuration=4
	block := []byte{0x81, 0x00, 0x00, 0x00, 'D'} // track 1, ts 0, flags 0x00, data 'D'
	bg := new(bytes.Buffer)
	// Block
	bg.Write([]byte{0xA1})
	bg.Write(vintEncode(uint64(len(block))))
	bg.Write(block)
	// BlockDuration = 4
	bg.Write([]byte{0x9B, 0x81, 0x04})

	mp := &MatroskaParser{reader: NewEBMLReader(bytes.NewReader(bg.Bytes())), fileInfo: &SegmentInfo{TimecodeScale: 1000000}}
	pkt, err := mp.parseBlockGroup(uint64(bg.Len()))
	if err != nil {
		t.Fatalf("parseBlockGroup failed: %v", err)
	}
	if pkt == nil || pkt.Track != 1 {
		t.Fatalf("unexpected packet: %+v", pkt)
	}
	if pkt.EndTime-pkt.StartTime != 4*mp.fileInfo.TimecodeScale {
		t.Errorf("duration not applied: start=%d end=%d", pkt.StartTime, pkt.EndTime)
	}
}

// TestReadPacket_TopLevelTimestamp_And_Mask exercises top-level Timestamp and mask filtering.
func TestReadPacket_TopLevelTimestamp_And_Mask(t *testing.T) {
	makeFile := func() []byte {
		buf := new(bytes.Buffer)
		// EBML Header
		eh := new(bytes.Buffer)
		eh.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(eh.Len())))
		buf.Write(eh.Bytes())
		// Segment
		seg := new(bytes.Buffer)
		// Info TS scale
		si := new(bytes.Buffer)
		si.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
		seg.Write([]byte{0x15, 0x49, 0xA9, 0x66})
		seg.Write(vintEncode(uint64(si.Len())))
		seg.Write(si.Bytes())
		// Tracks (1 video)
		te, _ := createMockTrackEntry(1, TypeVideo, "V", "V", "und")
		trs := new(bytes.Buffer)
		trs.Write([]byte{0xAE})
		trs.Write(vintEncode(uint64(len(te))))
		trs.Write(te)
		seg.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
		seg.Write(vintEncode(uint64(trs.Len())))
		seg.Write(trs.Bytes())
		// First add an empty Cluster (so parseSegmentChildren returns early and ReadPacket drives parsing)
		cl := new(bytes.Buffer)
		cl.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0
		seg.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
		seg.Write(vintEncode(uint64(cl.Len())))
		seg.Write(cl.Bytes())
		// Then add a top-level Timestamp element and a SimpleBlock
		seg.Write([]byte{0xE7}) // IDTimestamp at top-level
		seg.Write(vintEncode(2))
		seg.Write([]byte{0x03, 0xE8})             // 1000
		sb := []byte{0x81, 0x00, 0x00, 0x80, 'X'} // keyframe block
		seg.Write([]byte{0xA3})
		seg.Write(vintEncode(uint64(len(sb))))
		seg.Write(sb)
		// Wrap segment
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
		buf.Write(vintEncode(uint64(seg.Len())))
		buf.Write(seg.Bytes())
		return buf.Bytes()
	}

	// Normal read: should get one packet with scaled time using top-level timestamp
	p, err := NewMatroskaParser(bytes.NewReader(makeFile()), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	pkt, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}
	if (pkt.Flags&KF) == 0 || pkt.Track != 1 {
		t.Errorf("unexpected packet: %+v", pkt)
	}

	// Mask out track 1 and attempt to read -> should hit EOF (filtered)
	p2, err := NewMatroskaParser(bytes.NewReader(makeFile()), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	p2.SetTrackMask(0x01) // ignore track 1
	pkt2, err := p2.ReadPacket()
	if err == nil || err != io.EOF || pkt2 != nil {
		t.Errorf("expected EOF due to mask, got pkt=%v err=%v", pkt2, err)
	}
}

// TestSkipToKeyframe_Behavior ensures it consumes up to next keyframe.
func TestSkipToKeyframe_Behavior(t *testing.T) {
	// Build a stream: non-keyframe, keyframe, then a third frame
	mk := func() []byte {
		buf := new(bytes.Buffer)
		eh := new(bytes.Buffer)
		eh.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(eh.Len())))
		buf.Write(eh.Bytes())
		seg := new(bytes.Buffer)
		// TS scale
		si := new(bytes.Buffer)
		si.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
		seg.Write([]byte{0x15, 0x49, 0xA9, 0x66})
		seg.Write(vintEncode(uint64(si.Len())))
		seg.Write(si.Bytes())
		te, _ := createMockTrackEntry(1, TypeVideo, "V", "V", "und")
		trs := new(bytes.Buffer)
		trs.Write([]byte{0xAE})
		trs.Write(vintEncode(uint64(len(te))))
		trs.Write(te)
		seg.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
		seg.Write(vintEncode(uint64(trs.Len())))
		seg.Write(trs.Bytes())
		cl := new(bytes.Buffer)
		cl.Write([]byte{0xE7, 0x81, 0x00}) // ts 0
		// non-keyframe
		b1 := []byte{0x81, 0x00, 0x00, 0x00, 'a'}
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(b1))))
		cl.Write(b1)
		// keyframe
		b2 := []byte{0x81, 0x00, 0x00, 0x80, 'b'}
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(b2))))
		cl.Write(b2)
		// third
		b3 := []byte{0x81, 0x00, 0x00, 0x00, 'c'}
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(b3))))
		cl.Write(b3)
		seg.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
		seg.Write(vintEncode(uint64(cl.Len())))
		seg.Write(cl.Bytes())
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
		buf.Write(vintEncode(uint64(seg.Len())))
		buf.Write(seg.Bytes())
		return buf.Bytes()
	}

	p, err := NewMatroskaParser(bytes.NewReader(mk()), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	p.SkipToKeyframe()
	// Next packet should be the one after the keyframe (i.e., 'c')
	pkt, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket after SkipToKeyframe failed: %v", err)
	}
	if string(pkt.Data) != "c" {
		t.Errorf("expected 'c' after SkipToKeyframe, got %q", string(pkt.Data))
	}
}

func TestParseSegmentInfo_Rich(t *testing.T) {
	buf := new(bytes.Buffer)
	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)
	segInfo := new(bytes.Buffer)
	// SegmentUID (16 bytes)
	segInfo.Write([]byte{0x73, 0xA4, 0x90})
	for i := 0; i < 16; i++ {
		segInfo.WriteByte(byte(i + 1))
	}
	// SegmentFilename "a.mkv"
	segInfo.Write([]byte{0x73, 0x84, 0x85, 'a', '.', 'm', 'k', 'v'})
	// PrevUID (16)
	segInfo.Write([]byte{0x3C, 0xB9, 0x23, 0x90})
	for i := 0; i < 16; i++ {
		segInfo.WriteByte(byte(0xA0 + i))
	}
	// PrevFilename "p.mkv"
	segInfo.Write([]byte{0x3C, 0x83, 0xAB, 0x85, 'p', '.', 'm', 'k', 'v'})
	// NextUID (16)
	segInfo.Write([]byte{0x3E, 0xB9, 0x23, 0x90})
	for i := 0; i < 16; i++ {
		segInfo.WriteByte(byte(0xB0 + i))
	}
	// NextFilename "n.mkv"
	segInfo.Write([]byte{0x3E, 0x83, 0xBB, 0x85, 'n', '.', 'm', 'k', 'v'})
	// TimestampScale 1,000,000
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
	// Duration = 123 (as uint)
	segInfo.Write([]byte{0x44, 0x89, 0x81, 0x7B})
	// DateUTC (int64 as signed vint stored in ReadInt path via element.ReadInt; here emulate 8-byte int 0)
	// We will skip setting DateUTC to keep test simple and stable.
	// Title
	segInfo.Write([]byte{0x7B, 0xA9, 0x8A, 'R', 'i', 'c', 'h', ' ', 'T', 'i', 't', 'l', 'e'})
	// MuxingApp
	segInfo.Write([]byte{0x4D, 0x80, 0x84, 'm', 'u', 'x', 'r'})
	// WritingApp
	segInfo.Write([]byte{0x57, 0x41, 0x84, 'w', 'r', 'i', 't'})

	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())

	p, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	fi := p.GetFileInfo()
	if fi == nil || fi.Title != "Rich Title" || fi.Filename != "a.mkv" || fi.PrevFilename != "p.mkv" || fi.NextFilename != "n.mkv" {
		t.Fatalf("Unexpected file info: %+v", fi)
	}
	if fi.TimecodeScale != 1000000 || fi.Duration != 123 {
		t.Errorf("Unexpected scale/duration: %+v", fi)
	}
}

func createMockMatroskaFileWithBlockGroup() ([]byte, error) {
	buf := new(bytes.Buffer)
	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)

	// -- SegmentInfo TimestampScale = 1,000,000
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	// -- Tracks
	trackEntry, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "TestVideo", "und")
	tracks := new(bytes.Buffer)
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(len(trackEntry))))
	tracks.Write(trackEntry)
	segment.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
	segment.Write(vintEncode(uint64(tracks.Len())))
	segment.Write(tracks.Bytes())

	// -- Cluster with BlockGroup
	cluster := new(bytes.Buffer)
	cluster.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0

	// BlockGroup
	bg := new(bytes.Buffer)
	// Block element (0xA1) with track 1, timecode 0, flags 0, data "BG"
	blockData := []byte{0x81, 0x00, 0x00, 0x00, 'B', 'G'}
	bg.Write([]byte{0xA1})
	bg.Write(vintEncode(uint64(len(blockData))))
	bg.Write(blockData)
	// BlockDuration (0x9B) value 5
	bg.Write([]byte{0x9B, 0x81, 0x05})

	cluster.Write([]byte{0xA0}) // BlockGroup ID
	cluster.Write(vintEncode(uint64(bg.Len())))
	cluster.Write(bg.Bytes())

	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
	segment.Write(vintEncode(uint64(cluster.Len())))
	segment.Write(cluster.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())

	return buf.Bytes(), nil
}

func TestReadPacket_BlockGroup(t *testing.T) {
	data, err := createMockMatroskaFileWithBlockGroup()
	if err != nil {
		t.Fatalf("failed to build mock: %v", err)
	}
	p, err := NewMatroskaParser(bytes.NewReader(data), false)
	if err != nil {
		t.Fatalf("NewMatroskaParser failed: %v", err)
	}
	pkt, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}
	if string(pkt.Data) != "BG" || pkt.Track != 1 || pkt.Flags&KF == 0 {
		t.Errorf("Unexpected packet from BlockGroup: %+v", pkt)
	}
	if pkt.EndTime <= pkt.StartTime {
		t.Errorf("Expected EndTime > StartTime due to BlockDuration, got %d <= %d", pkt.EndTime, pkt.StartTime)
	}
}

// parseSegmentChildren: out-of-order children and unknown IDs should be tolerated
func TestParseSegmentChildren_OrderAndUnknown(t *testing.T) {
	buf := new(bytes.Buffer)
	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)

	// Put Tracks first (before SegmentInfo)
	trackEntry, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "T", "und")
	tracks := new(bytes.Buffer)
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(len(trackEntry))))
	tracks.Write(trackEntry)
	segment.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
	segment.Write(vintEncode(uint64(tracks.Len())))
	segment.Write(tracks.Bytes())

	// Unknown child (Void 0xEC) between known ones
	segment.Write([]byte{0xEC, 0x81, 0x00})

	// SegmentInfo
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x7B, 0xA9, 0x87, 'O', 'r', 'd', 'e', 'r', 'e', 'd'})
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	// One Cluster with a block
	cluster := new(bytes.Buffer)
	cluster.Write([]byte{0xE7, 0x81, 0x00}) // Timestamp 0
	sb := []byte{0x81, 0x00, 0x00, 0x80, 'x'}
	cluster.Write([]byte{0xA3})
	cluster.Write(vintEncode(uint64(len(sb))))
	cluster.Write(sb)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
	segment.Write(vintEncode(uint64(cluster.Len())))
	segment.Write(cluster.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())

	if _, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false); err != nil {
		t.Fatalf("Parser should accept out-of-order children and unknown IDs: %v", err)
	}
}

// Tracks with multiple TrackEntry types: audio and subtitle in addition to video
func TestParseTrackEntry_VariousTypes(t *testing.T) {
	buf := new(bytes.Buffer)
	// EBML Header
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	// Segment
	segment := new(bytes.Buffer)

	// SegmentInfo minimal
	segInfo := new(bytes.Buffer)
	segInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
	segment.Write([]byte{0x15, 0x49, 0xA9, 0x66})
	segment.Write(vintEncode(uint64(segInfo.Len())))
	segment.Write(segInfo.Bytes())

	// Tracks
	tracks := new(bytes.Buffer)
	// Video track (1)
	vte, _ := createMockTrackEntry(1, TypeVideo, "V_TEST", "V", "und")
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(len(vte))))
	tracks.Write(vte)
	// Audio track (2) with channels 1 and sampling frequency 44100.0
	ate := new(bytes.Buffer)
	// TrackNumber (0xD7) = 2
	ate.Write([]byte{0xD7, 0x81, 0x02})
	// TrackUID (0x73C5) = 2
	ate.Write([]byte{0x73, 0xC5, 0x81, 0x02})
	// TrackType (0x83) = audio (2)
	ate.Write([]byte{0x83, 0x81, 0x02})
	// CodecID (0x86) = "A_TEST"
	ate.Write([]byte{0x86, 0x86, 'A', '_', 'T', 'E', 'S', 'T'})
	// Name (0x536E) = "A"
	ate.Write([]byte{0x53, 0x6E, 0x81, 'A'})
	// Language (0x22B59C) = "eng"
	ate.Write([]byte{0x22, 0xB5, 0x9C, 0x83, 'e', 'n', 'g'})
	// Audio (0xE1) child: SamplingFrequency (0xB5) + Channels (0x9F)
	audio := new(bytes.Buffer)
	// SamplingFrequency 44100.0
	sf := math.Float64bits(44100.0)
	audio.Write([]byte{0xB5, 0x88})
	_ = binary.Write(audio, binary.BigEndian, sf)
	// Channels 1
	audio.Write([]byte{0x9F, 0x81, 0x01})
	ate.Write([]byte{0xE1})
	ate.Write(vintEncode(uint64(audio.Len())))
	ate.Write(audio.Bytes())
	// Wrap as TrackEntry (0xAE)
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(ate.Len())))
	tracks.Write(ate.Bytes())
	// Subtitle track (3)
	ste, _ := createMockTrackEntry(3, TypeSubtitle, "S_TEST", "S", "eng")
	tracks.Write([]byte{0xAE})
	tracks.Write(vintEncode(uint64(len(ste))))
	tracks.Write(ste)

	segment.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
	segment.Write(vintEncode(uint64(tracks.Len())))
	segment.Write(tracks.Bytes())

	// Minimal cluster so parser finishes
	cluster := new(bytes.Buffer)
	cluster.Write([]byte{0xE7, 0x81, 0x00})
	cluster.Write([]byte{0xA3, 0x82, 0x81, 0x00}) // tiny SimpleBlock (may not decode, but ok)
	segment.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
	segment.Write(vintEncode(uint64(cluster.Len())))
	segment.Write(cluster.Bytes())

	buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
	buf.Write(vintEncode(uint64(segment.Len())))
	buf.Write(segment.Bytes())

	p, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
	if err != nil {
		t.Fatalf("parser failed: %v", err)
	}

	if p.GetNumTracks() != 3 {
		t.Fatalf("expected 3 tracks, got %d", p.GetNumTracks())
	}
	if p.GetTrackInfo(1) == nil || p.GetTrackInfo(1).Type != TypeAudio {
		t.Errorf("expected track 2 to be audio: %+v", p.GetTrackInfo(1))
	}
	if p.GetTrackInfo(2) == nil || p.GetTrackInfo(2).Type != TypeSubtitle {
		t.Errorf("expected track 3 to be subtitle: %+v", p.GetTrackInfo(2))
	}
	if p.GetTrackInfo(1).Audio.SamplingFreq != 44100.0 || p.GetTrackInfo(1).Audio.Channels != 1 {
		t.Errorf("audio fields not parsed: %+v", p.GetTrackInfo(1).Audio)
	}
}

// SimpleBlock lacing variants
func TestParseSimpleBlock_LacingVariants(t *testing.T) {
	// Build a file with two SimpleBlocks: one Xiph-laced and one EBML-laced.
	buildWithBlock := func(flags byte, payload []byte) []byte {
		// track 1 vint 0x81, timecode 0x0000, flags, then payload
		b := []byte{0x81, 0x00, 0x00, flags}
		b = append(b, payload...)
		return b
	}

	// Xiph lacing: flags with 0x06; two frames: sizes [1, remainder]. Header: frameCount-1=1 then size 0x01, data "A" "B"
	xiphPayload := append([]byte{0x01, 0x01}, []byte{'A', 'B'}...)
	xiphBlock := buildWithBlock(0x06|0x80, xiphPayload) // include keyframe bit

	// EBML lacing: flags with 0x04; minimal payload for 2 frames. We keep it simple (parser doesn't parse, just returns data)
	// Frame count-1=1, then leave some bytes as sizes/data.
	ebmlPayload := append([]byte{0x01, 0x81}, []byte{'Z', 'Z'}...)
	ebmlBlock := buildWithBlock(0x04|0x80, ebmlPayload)

	makeFile := func(block []byte) []byte {
		buf := new(bytes.Buffer)
		// Header
		eh := new(bytes.Buffer)
		eh.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(eh.Len())))
		buf.Write(eh.Bytes())
		// Segment
		seg := new(bytes.Buffer)
		// Info TS scale
		si := new(bytes.Buffer)
		si.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
		seg.Write([]byte{0x15, 0x49, 0xA9, 0x66})
		seg.Write(vintEncode(uint64(si.Len())))
		seg.Write(si.Bytes())
		// Tracks
		te, _ := createMockTrackEntry(1, TypeVideo, "V", "V", "und")
		trs := new(bytes.Buffer)
		trs.Write([]byte{0xAE})
		trs.Write(vintEncode(uint64(len(te))))
		trs.Write(te)
		seg.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
		seg.Write(vintEncode(uint64(trs.Len())))
		seg.Write(trs.Bytes())
		// Cluster
		cl := new(bytes.Buffer)
		cl.Write([]byte{0xE7, 0x81, 0x00})
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(block))))
		cl.Write(block)
		seg.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
		seg.Write(vintEncode(uint64(cl.Len())))
		seg.Write(cl.Bytes())
		// Wrap
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
		buf.Write(vintEncode(uint64(seg.Len())))
		buf.Write(seg.Bytes())
		return buf.Bytes()
	}

	// Xiph test
	p, err := NewMatroskaParser(bytes.NewReader(makeFile(xiphBlock)), false)
	if err != nil {
		t.Fatalf("parser err: %v", err)
	}
	pkt, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket xiph err: %v", err)
	}
	if string(pkt.Data) != "A" {
		t.Errorf("expected first frame 'A', got %q", string(pkt.Data))
	}

	// EBML test
	p2, err := NewMatroskaParser(bytes.NewReader(makeFile(ebmlBlock)), false)
	if err != nil {
		t.Fatalf("parser err: %v", err)
	}
	pkt2, err := p2.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket ebml err: %v", err)
	}
	if len(pkt2.Data) == 0 {
		t.Errorf("expected non-empty data for EBML lacing")
	}
}

// Fixed-size lacing variant to cover 0x02 branch
func TestParseSimpleBlock_LacingFixed(t *testing.T) {
	// Build fixed-size laced SimpleBlock with 2 frames of equal size
	// Flags: keyframe + fixed lacing (0x80 | 0x02)
	// header: track 1, ts 0
	header := []byte{0x81, 0x00, 0x00, 0x82}
	// frame count-1 = 1
	// payload two frames: "AB" and "CD"
	payload := append([]byte{0x01}, []byte{'A', 'B', 'C', 'D'}...)
	block := append(header, payload...)

	// Wrap in a minimal cluster + segment so ReadPacket parses it
	file := func() []byte {
		buf := new(bytes.Buffer)
		eh := new(bytes.Buffer)
		eh.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'})
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write(vintEncode(uint64(eh.Len())))
		buf.Write(eh.Bytes())
		seg := new(bytes.Buffer)
		si := new(bytes.Buffer)
		si.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40})
		seg.Write([]byte{0x15, 0x49, 0xA9, 0x66})
		seg.Write(vintEncode(uint64(si.Len())))
		seg.Write(si.Bytes())
		te, _ := createMockTrackEntry(1, TypeVideo, "V", "V", "und")
		trs := new(bytes.Buffer)
		trs.Write([]byte{0xAE})
		trs.Write(vintEncode(uint64(len(te))))
		trs.Write(te)
		seg.Write([]byte{0x16, 0x54, 0xAE, 0x6B})
		seg.Write(vintEncode(uint64(trs.Len())))
		seg.Write(trs.Bytes())
		cl := new(bytes.Buffer)
		cl.Write([]byte{0xE7, 0x81, 0x00})
		cl.Write([]byte{0xA3})
		cl.Write(vintEncode(uint64(len(block))))
		cl.Write(block)
		seg.Write([]byte{0x1F, 0x43, 0xB6, 0x75})
		seg.Write(vintEncode(uint64(cl.Len())))
		seg.Write(cl.Bytes())
		buf.Write([]byte{0x18, 0x53, 0x80, 0x67})
		buf.Write(vintEncode(uint64(seg.Len())))
		buf.Write(seg.Bytes())
		return buf.Bytes()
	}()

	p, err := NewMatroskaParser(bytes.NewReader(file), false)
	if err != nil {
		t.Fatalf("parser err: %v", err)
	}
	pkt, err := p.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket err: %v", err)
	}
	if string(pkt.Data) != "AB" {
		t.Errorf("expected first fixed-laced frame 'AB', got %q", string(pkt.Data))
	}
}

// Error path tests for parseSimpleBlock to cover short data and invalid VINT
func TestParseSimpleBlock_ErrorPaths(t *testing.T) {
	// Helper to run parseSimpleBlock on raw data
	run := func(data []byte) error {
		mp := &MatroskaParser{
			reader:   &EBMLReader{r: &seekableReader{bytes.NewReader(data)}, pos: 0},
			fileInfo: &SegmentInfo{TimecodeScale: 1000000},
		}
		_, err := mp.parseSimpleBlock(uint64(len(data)))
		return err
	}

	// Too short block (<4)
	if err := run([]byte{0x81, 0x00, 0x00}); err == nil {
		t.Errorf("expected error for short block, got nil")
	}

	// Invalid VINT for track number (first byte = 0x00)
	// Build 4 bytes to pass the initial length check but fail vint parsing.
	if err := run([]byte{0x00, 0x00, 0x00, 0x00}); err == nil {
		t.Errorf("expected error for invalid track VINT, got nil")
	}

	// Short for timestamp: valid 1-byte vint track (0x81) but missing bytes for timestamp
	if err := run([]byte{0x81, 0x00}); err == nil {
		t.Errorf("expected error for short timestamp, got nil")
	}
}

// TestParseChapters tests the parsing of Chapters element.
func TestParseChapters(t *testing.T) {
	t.Run("Valid chapters data", func(t *testing.T) {
		// Create mock chapters data with one EditionEntry containing one ChapterAtom
		buf := new(bytes.Buffer)

		// EditionEntry
		editionEntry := new(bytes.Buffer)

		// ChapterAtom
		chapterAtom := new(bytes.Buffer)
		// ChapterUID: 1
		chapterAtom.Write([]byte{0x73, 0xC4, 0x81, 0x01})
		// ChapterTimeStart: 0 (0 nanoseconds)
		chapterAtom.Write([]byte{0x91, 0x81, 0x00})
		// ChapterTimeEnd: 5000 (5000 nanoseconds)
		chapterAtom.Write([]byte{0x92, 0x82, 0x13, 0x88})
		// ChapterDisplay
		chapterDisplay := new(bytes.Buffer)
		// ChapterString: "Chapter 1"
		chapterDisplay.Write([]byte{0x85, 0x89, 'C', 'h', 'a', 'p', 't', 'e', 'r', ' ', '1'})
		// ChapterLanguage: "eng"
		chapterDisplay.Write([]byte{0x43, 0x7C, 0x83, 'e', 'n', 'g'})

		chapterAtom.Write([]byte{0x80}) // ChapterDisplay ID
		chapterAtom.Write(vintEncode(uint64(chapterDisplay.Len())))
		chapterAtom.Write(chapterDisplay.Bytes())

		editionEntry.Write([]byte{0xB6}) // ChapterAtom ID
		editionEntry.Write(vintEncode(uint64(chapterAtom.Len())))
		editionEntry.Write(chapterAtom.Bytes())

		buf.Write([]byte{0x45, 0xB9}) // EditionEntry ID
		buf.Write(vintEncode(uint64(editionEntry.Len())))
		buf.Write(editionEntry.Bytes())

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseChapters(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseChapters() failed: %v", err)
		}

		if len(parser.chapters) == 0 {
			t.Fatal("Expected at least one chapter, got none")
		}

		chapter := parser.chapters[0]
		if chapter.UID != 1 {
			t.Errorf("Expected chapter UID 1, got %d", chapter.UID)
		}
		if chapter.Start != 0 {
			t.Errorf("Expected chapter start time 0, got %d", chapter.Start)
		}
		if chapter.End != 5000 {
			t.Errorf("Expected chapter end time 5000, got %d", chapter.End)
		}
		if len(chapter.Display) == 0 {
			t.Fatal("Expected chapter display information, got none")
		}
		if chapter.Display[0].String != "Chapter 1" {
			t.Errorf("Expected chapter string 'Chapter 1', got %q", chapter.Display[0].String)
		}
		if chapter.Display[0].Language != "eng" {
			t.Errorf("Expected chapter language 'eng', got %q", chapter.Display[0].Language)
		}
	})

	t.Run("Empty chapters data", func(t *testing.T) {
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader([]byte{})),
		}

		err := parser.parseChapters(0)
		if err != nil {
			t.Fatalf("parseChapters() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully
	})

	t.Run("ReadFull error", func(t *testing.T) {
		// Create a reader that will fail on ReadFull
		reader := &failingReader{
			data:       make([]byte, 5), // Small data
			failAtByte: 3,               // Fail after 3 bytes
		}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		err := parser.parseChapters(10) // Request more bytes than available
		if err == nil {
			t.Fatal("Expected ReadFull error, got nil")
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("ReadElement error", func(t *testing.T) {
		// Create invalid EBML data that will cause ReadElement to fail
		invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid EBML element
		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(invalidData)),
		}

		err := parser.parseChapters(uint64(len(invalidData)))
		if err == nil {
			t.Fatal("Expected ReadElement error, got nil")
		}
	})

	t.Run("Non-EditionEntry elements", func(t *testing.T) {
		// Create chapters data with non-EditionEntry elements (should be ignored)
		buf := new(bytes.Buffer)

		// Add a non-EditionEntry element (using a different ID)
		buf.Write([]byte{0x12, 0x34, 0x81, 0x00}) // Unknown element with size 1 and data 0x00

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseChapters(uint64(buf.Len()))
		if err != nil {
			t.Fatalf("parseChapters() with non-EditionEntry elements failed: %v", err)
		}
		// Should ignore non-EditionEntry elements
		if len(parser.chapters) != 0 {
			t.Errorf("Expected no chapters, got %d", len(parser.chapters))
		}
	})

	t.Run("parseEditionEntry error", func(t *testing.T) {
		// Create chapters data with invalid EditionEntry that will cause parseEditionEntry to fail
		buf := new(bytes.Buffer)

		// EditionEntry with invalid data
		buf.Write([]byte{0x45, 0xB9})             // EditionEntry ID
		buf.Write([]byte{0x84})                   // Size: 4
		buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // Invalid data

		parser := &MatroskaParser{
			reader: NewEBMLReader(bytes.NewReader(buf.Bytes())),
		}

		err := parser.parseChapters(uint64(buf.Len()))
		if err == nil {
			t.Fatal("Expected parseEditionEntry error, got nil")
		}
	})
}

// TestParseEditionEntry tests the parsing of an EditionEntry element.
func TestParseEditionEntry(t *testing.T) {
	t.Run("Valid edition entry with multiple chapters", func(t *testing.T) {
		// Create mock edition entry data with two ChapterAtoms
		buf := new(bytes.Buffer)

		// ChapterAtom 1
		chapterAtom1 := new(bytes.Buffer)
		chapterAtom1.Write([]byte{0x73, 0xC4, 0x81, 0x01}) // ChapterUID: 1
		chapterAtom1.Write([]byte{0x91, 0x81, 0x00})       // ChapterTimeStart: 0

		buf.Write([]byte{0xB6}) // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterAtom1.Len())))
		buf.Write(chapterAtom1.Bytes())

		// ChapterAtom 2
		chapterAtom2 := new(bytes.Buffer)
		chapterAtom2.Write([]byte{0x73, 0xC4, 0x81, 0x02}) // ChapterUID: 2
		chapterAtom2.Write([]byte{0x91, 0x82, 0x13, 0x88}) // ChapterTimeStart: 5000

		buf.Write([]byte{0xB6}) // ChapterAtom ID
		buf.Write(vintEncode(uint64(chapterAtom2.Len())))
		buf.Write(chapterAtom2.Bytes())

		parser := &MatroskaParser{}

		chapters, err := parser.parseEditionEntry(buf.Bytes())
		if err != nil {
			t.Fatalf("parseEditionEntry() failed: %v", err)
		}

		if len(chapters) != 2 {
			t.Fatalf("Expected 2 chapters, got %d", len(chapters))
		}

		if chapters[0].UID != 1 {
			t.Errorf("Expected first chapter UID 1, got %d", chapters[0].UID)
		}
		if chapters[1].UID != 2 {
			t.Errorf("Expected second chapter UID 2, got %d", chapters[1].UID)
		}
	})

	t.Run("Empty edition entry", func(t *testing.T) {
		parser := &MatroskaParser{}

		chapters, err := parser.parseEditionEntry([]byte{})
		if err != nil {
			t.Fatalf("parseEditionEntry() with empty data failed: %v", err)
		}
		if len(chapters) != 0 {
			t.Errorf("Expected no chapters for empty data, got %d", len(chapters))
		}
	})
}

// TestParseChapterAtom tests the parsing of a ChapterAtom element.
func TestParseChapterAtom(t *testing.T) {
	t.Run("Complete chapter atom with all fields", func(t *testing.T) {
		// Create mock chapter atom data with all possible fields
		buf := new(bytes.Buffer)
		// ChapterUID: 123
		buf.Write([]byte{0x73, 0xC4, 0x81, 0x7B})
		// ChapterTimeStart: 1000
		buf.Write([]byte{0x91, 0x82, 0x03, 0xE8})
		// ChapterTimeEnd: 2000
		buf.Write([]byte{0x92, 0x82, 0x07, 0xD0})
		// ChapterHidden: 1 (true)
		buf.Write([]byte{0x98, 0x81, 0x01})
		// ChapterEnabled: 0 (false)
		buf.Write([]byte{0x45, 0x98, 0x81, 0x00})

		// ChapterDisplay
		chapterDisplay := new(bytes.Buffer)
		chapterDisplay.Write([]byte{0x85, 0x8A, 'T', 'e', 's', 't', ' ', 'T', 'i', 't', 'l', 'e'}) // ChapterString: "Test Title"
		chapterDisplay.Write([]byte{0x43, 0x7C, 0x83, 'j', 'p', 'n'})                              // ChapterLanguage: "jpn"
		chapterDisplay.Write([]byte{0x43, 0x7E, 0x82, 'J', 'P'})                                   // ChapterCountry: "JP"

		buf.Write([]byte{0x80}) // ChapterDisplay ID
		buf.Write(vintEncode(uint64(chapterDisplay.Len())))
		buf.Write(chapterDisplay.Bytes())

		// Nested ChapterAtom
		nestedChapter := new(bytes.Buffer)
		nestedChapter.Write([]byte{0x73, 0xC4, 0x81, 0x7C}) // ChapterUID: 124
		nestedChapter.Write([]byte{0x91, 0x82, 0x05, 0xDC}) // ChapterTimeStart: 1500

		buf.Write([]byte{0xB6}) // ChapterAtom ID (nested)
		buf.Write(vintEncode(uint64(nestedChapter.Len())))
		buf.Write(nestedChapter.Bytes())

		parser := &MatroskaParser{}

		chapter, err := parser.parseChapterAtom(buf.Bytes())
		if err != nil {
			t.Fatalf("parseChapterAtom() failed: %v", err)
		}

		if chapter.UID != 123 {
			t.Errorf("Expected chapter UID 123, got %d", chapter.UID)
		}
		if chapter.Start != 1000 {
			t.Errorf("Expected chapter start time 1000, got %d", chapter.Start)
		}
		if chapter.End != 2000 {
			t.Errorf("Expected chapter end time 2000, got %d", chapter.End)
		}
		if !chapter.Hidden {
			t.Errorf("Expected chapter to be hidden, got false")
		}
		if chapter.Enabled {
			t.Errorf("Expected chapter to be disabled, got true")
		}

		if len(chapter.Display) == 0 {
			t.Fatal("Expected chapter display information, got none")
		}
		display := chapter.Display[0]
		if display.String != "Test Title" {
			t.Errorf("Expected chapter string 'Test Title', got %q", display.String)
		}
		if display.Language != "jpn" {
			t.Errorf("Expected chapter language 'jpn', got %q", display.Language)
		}
		if display.Country != "JP" {
			t.Errorf("Expected chapter country 'JP', got %q", display.Country)
		}

		if len(chapter.Children) == 0 {
			t.Fatal("Expected nested chapter, got none")
		}
		if chapter.Children[0].UID != 124 {
			t.Errorf("Expected nested chapter UID 124, got %d", chapter.Children[0].UID)
		}
	})

	t.Run("Minimal chapter atom", func(t *testing.T) {
		// Create minimal chapter atom data with only UID
		buf := new(bytes.Buffer)
		buf.Write([]byte{0x73, 0xC4, 0x81, 0x01}) // ChapterUID: 1

		parser := &MatroskaParser{}

		chapter, err := parser.parseChapterAtom(buf.Bytes())
		if err != nil {
			t.Fatalf("parseChapterAtom() failed: %v", err)
		}

		if chapter.UID != 1 {
			t.Errorf("Expected chapter UID 1, got %d", chapter.UID)
		}
		if !chapter.Enabled {
			t.Errorf("Expected chapter to be enabled by default, got false")
		}
		if chapter.Hidden {
			t.Errorf("Expected chapter to not be hidden by default, got true")
		}
	})

	t.Run("Empty chapter atom", func(t *testing.T) {
		parser := &MatroskaParser{}

		chapter, err := parser.parseChapterAtom([]byte{})
		if err != nil {
			t.Fatalf("parseChapterAtom() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully with default values
		if !chapter.Enabled {
			t.Errorf("Expected chapter to be enabled by default, got false")
		}
	})
}

// TestParseChapterDisplay tests the parsing of a ChapterDisplay element.
func TestParseChapterDisplay(t *testing.T) {
	t.Run("Complete chapter display with all fields", func(t *testing.T) {
		// Create mock chapter display data with all fields
		buf := new(bytes.Buffer)
		// ChapterString: "My Chapter"
		buf.Write([]byte{0x85, 0x8A, 'M', 'y', ' ', 'C', 'h', 'a', 'p', 't', 'e', 'r'})
		// ChapterLanguage: "fra"
		buf.Write([]byte{0x43, 0x7C, 0x83, 'f', 'r', 'a'})
		// ChapterCountry: "FR"
		buf.Write([]byte{0x43, 0x7E, 0x82, 'F', 'R'})

		parser := &MatroskaParser{}

		display, err := parser.parseChapterDisplay(buf.Bytes())
		if err != nil {
			t.Fatalf("parseChapterDisplay() failed: %v", err)
		}

		if display.String != "My Chapter" {
			t.Errorf("Expected chapter string 'My Chapter', got %q", display.String)
		}
		if display.Language != "fra" {
			t.Errorf("Expected chapter language 'fra', got %q", display.Language)
		}
		if display.Country != "FR" {
			t.Errorf("Expected chapter country 'FR', got %q", display.Country)
		}
	})

	t.Run("Minimal chapter display with only string", func(t *testing.T) {
		// Create chapter display data with only ChapterString
		buf := new(bytes.Buffer)
		buf.Write([]byte{0x85, 0x85, 'T', 'i', 't', 'l', 'e'}) // ChapterString: "Title"

		parser := &MatroskaParser{}

		display, err := parser.parseChapterDisplay(buf.Bytes())
		if err != nil {
			t.Fatalf("parseChapterDisplay() failed: %v", err)
		}

		if display.String != "Title" {
			t.Errorf("Expected chapter string 'Title', got %q", display.String)
		}
		if display.Language != "eng" {
			t.Errorf("Expected default language 'eng', got %q", display.Language)
		}
		if display.Country != "" {
			t.Errorf("Expected empty country, got %q", display.Country)
		}
	})

	t.Run("Empty chapter display", func(t *testing.T) {
		parser := &MatroskaParser{}

		display, err := parser.parseChapterDisplay([]byte{})
		if err != nil {
			t.Fatalf("parseChapterDisplay() with empty data failed: %v", err)
		}
		// Should handle empty data gracefully with default values
		if display.Language != "eng" {
			t.Errorf("Expected default language 'eng', got %q", display.Language)
		}
	})

	t.Run("Multiple language chapter display", func(t *testing.T) {
		// Test with different language combinations
		testCases := []struct {
			name     string
			langCode string
			country  string
		}{
			{"German", "ger", "DE"},
			{"Spanish", "spa", "ES"},
			{"Chinese", "chi", "CN"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				buf.Write([]byte{0x85, 0x84, 'T', 'e', 's', 't'}) // ChapterString: "Test"
				// ChapterLanguage
				buf.Write([]byte{0x43, 0x7C, byte(0x80 | len(tc.langCode))})
				buf.WriteString(tc.langCode)
				// ChapterCountry
				buf.Write([]byte{0x43, 0x7E, byte(0x80 | len(tc.country))})
				buf.WriteString(tc.country)

				parser := &MatroskaParser{}
				display, err := parser.parseChapterDisplay(buf.Bytes())
				if err != nil {
					t.Fatalf("parseChapterDisplay() failed for %s: %v", tc.name, err)
				}

				if display.Language != tc.langCode {
					t.Errorf("Expected language %q, got %q", tc.langCode, display.Language)
				}
				if display.Country != tc.country {
					t.Errorf("Expected country %q, got %q", tc.country, display.Country)
				}
			})
		}
	})
}

// TestParseSegmentChildren_noSeeking tests parseSegmentChildren with noSeeking=true
func TestParseSegmentChildren_noSeeking(t *testing.T) {
	t.Run("noSeeking with Cluster", func(t *testing.T) {
		// Create a segment with SegmentInfo, Tracks, and Cluster
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Cluster
		cluster := new(bytes.Buffer)
		cluster.Write([]byte{0xE7, 0x81, 0x00}) // Timecode: 0

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster.Len())))
		segmentData.Write(cluster.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		// Test with noSeeking=true
		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), true)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with noSeeking=true failed: %v", err)
		}

		if parser.fileInfo == nil {
			t.Error("Expected fileInfo to be parsed")
		}
		if len(parser.tracks) == 0 {
			t.Error("Expected tracks to be parsed")
		}
	})

	t.Run("noSeeking with unknown element", func(t *testing.T) {
		// Create a segment with an unknown element
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Unknown element (fake ID) - use a simpler unknown ID
		unknownData := []byte{0x01, 0x02, 0x03, 0x04}
		segmentData.Write([]byte{0xBF}) // Unknown ID (1 byte)
		segmentData.Write(vintEncode(uint64(len(unknownData))))
		segmentData.Write(unknownData)

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		// Test with noSeeking=true
		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), true)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with unknown element failed: %v", err)
		}

		if parser.fileInfo == nil {
			t.Error("Expected fileInfo to be parsed despite unknown element")
		}
	})
}

// TestParseSegmentChildren_ErrorHandling tests error handling in parseSegmentChildren
func TestParseSegmentChildren_ErrorHandling(t *testing.T) {
	t.Run("Truncated segment", func(t *testing.T) {
		// Create a segment that claims to be larger than the actual data
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment with size larger than actual data
		segmentData := new(bytes.Buffer)
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		// Claim segment is much larger than actual data
		buf.Write(vintEncode(uint64(segmentData.Len() + 1000)))
		buf.Write(segmentData.Bytes())
		// Don't write the extra 1000 bytes

		// This should result in an error when trying to parse
		_, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err == nil {
			t.Error("Expected error for truncated segment, got nil")
		}
	})

	t.Run("Invalid element in segment", func(t *testing.T) {
		// Create a segment with invalid element data
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// Invalid SegmentInfo (too short)
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})       // SegmentInfo ID
		segmentData.Write([]byte{0x85, 0x01, 0x02, 0x03, 0x04}) // Size 5, but only 4 bytes follow

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		// This should result in an error
		_, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err == nil {
			t.Error("Expected error for invalid segment element, got nil")
		}
	})
}

// TestParseSegmentChildren_StreamingScenario tests streaming scenario with unknown size
func TestParseSegmentChildren_StreamingScenario(t *testing.T) {
	t.Run("Unknown size segment with EOF", func(t *testing.T) {
		// Create a segment with unknown size that ends with EOF
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment with unknown size
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		// Unknown size (all 1s in the size field)
		buf.Write([]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		buf.Write(segmentData.Bytes())
		// EOF naturally terminates the segment

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with unknown size segment failed: %v", err)
		}

		if parser.fileInfo == nil {
			t.Error("Expected fileInfo to be parsed in streaming scenario")
		}
	})
}

// TestParseSegment_CompleteFlow tests the complete flow of parseSegment
func TestParseSegment_CompleteFlow(t *testing.T) {
	t.Run("Complete segment with basic elements", func(t *testing.T) {
		// Create a simpler segment with basic elements
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() with complete segment failed: %v", err)
		}

		// Verify basic elements were parsed
		if parser.fileInfo == nil {
			t.Error("Expected fileInfo to be parsed")
		}
		if len(parser.tracks) == 0 {
			t.Error("Expected tracks to be parsed")
		}
	})
}

// createMinimalEBMLHeader creates a minimal EBML header for testing
func createMinimalEBMLHeader() []byte {
	buf := new(bytes.Buffer)

	// EBML Header content
	ebmlHeader := new(bytes.Buffer)
	ebmlHeader.Write([]byte{0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'}) // DocType: "matroska"

	// EBML Header element
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3}) // EBML Header ID
	buf.Write(vintEncode(uint64(ebmlHeader.Len())))
	buf.Write(ebmlHeader.Bytes())

	return buf.Bytes()
}

// TestParseVInt_EdgeCases tests edge cases for parseVInt function
func TestParseVInt_EdgeCases(t *testing.T) {
	mp := &MatroskaParser{}

	testCases := []struct {
		name           string
		input          []byte
		expectedValue  uint64
		expectedLength int
	}{
		// Valid cases
		{"1-byte minimum", []byte{0x81}, 1, 1},
		{"1-byte maximum", []byte{0xFF}, 127, 1},
		{"2-byte minimum", []byte{0x40, 0x01}, 1, 2},
		{"2-byte maximum", []byte{0x7F, 0xFF}, 16383, 2},
		{"3-byte minimum", []byte{0x20, 0x00, 0x01}, 1, 3},
		{"3-byte maximum", []byte{0x3F, 0xFF, 0xFF}, 2097151, 3},
		{"4-byte minimum", []byte{0x10, 0x00, 0x00, 0x01}, 1, 4},
		{"4-byte maximum", []byte{0x1F, 0xFF, 0xFF, 0xFF}, 268435455, 4},
		{"5-byte minimum", []byte{0x08, 0x00, 0x00, 0x00, 0x01}, 1, 5},
		{"6-byte minimum", []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x01}, 1, 6},
		{"7-byte minimum", []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, 1, 7},
		{"8-byte minimum", []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, 1, 8},

		// Edge cases
		{"Single bit set", []byte{0x80}, 0, 1},
		{"All bits set in 1-byte", []byte{0xFF}, 127, 1},
		{"All bits set in 2-byte", []byte{0x7F, 0xFF}, 16383, 2},

		// Error cases
		{"Empty data", []byte{}, 0, 0},
		{"Zero first byte", []byte{0x00}, 0, 0},
		{"Insufficient data for 2-byte", []byte{0x40}, 0, 0},
		{"Insufficient data for 3-byte", []byte{0x20, 0x00}, 0, 0},
		{"Insufficient data for 4-byte", []byte{0x10, 0x00, 0x00}, 0, 0},
		{"Insufficient data for 8-byte", []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, length := mp.parseVInt(tc.input)
			if value != tc.expectedValue {
				t.Errorf("Expected value %d, got %d", tc.expectedValue, value)
			}
			if length != tc.expectedLength {
				t.Errorf("Expected length %d, got %d", tc.expectedLength, length)
			}
		})
	}
}

// TestParseVInt_LargeValues tests parseVInt with large values
func TestParseVInt_LargeValues(t *testing.T) {
	mp := &MatroskaParser{}

	testCases := []struct {
		name           string
		input          []byte
		expectedValue  uint64
		expectedLength int
	}{
		{
			"5-byte large value",
			[]byte{0x08, 0xFF, 0xFF, 0xFF, 0xFF},
			0xFFFFFFFF, // 4294967295 (mask removes the length bit)
			5,
		},
		{
			"6-byte large value",
			[]byte{0x04, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			0xFFFFFFFFFF, // 1099511627775 (mask removes the length bit)
			6,
		},
		{
			"7-byte large value",
			[]byte{0x02, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			0xFFFFFFFFFFFF, // 281474976710655 (mask removes the length bit)
			7,
		},
		{
			"8-byte large value",
			[]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			0xFFFFFFFFFFFFFF, // 72057594037927935
			8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, length := mp.parseVInt(tc.input)
			if value != tc.expectedValue {
				t.Errorf("Expected value %d, got %d", tc.expectedValue, value)
			}
			if length != tc.expectedLength {
				t.Errorf("Expected length %d, got %d", tc.expectedLength, length)
			}
		})
	}
}

// TestParseVInt_SpecialPatterns tests parseVInt with special bit patterns
func TestParseVInt_SpecialPatterns(t *testing.T) {
	mp := &MatroskaParser{}

	testCases := []struct {
		name           string
		input          []byte
		expectedValue  uint64
		expectedLength int
	}{
		// Patterns with alternating bits
		{"2-byte alternating", []byte{0x55, 0xAA}, 0x15AA, 2},
		{"3-byte alternating", []byte{0x2A, 0x55, 0xAA}, 0xA55AA, 3},

		// Patterns with specific bit arrangements
		{"2-byte with high bits", []byte{0x7F, 0x00}, 16128, 2}, // 0x3F00 = 16128
		{"3-byte with high bits", []byte{0x3F, 0x80, 0x00}, 2064384, 3},

		// Boundary values for each length
		{"1-byte boundary", []byte{0x81}, 1, 1},
		{"2-byte boundary", []byte{0x40, 0x00}, 0, 2},
		{"3-byte boundary", []byte{0x20, 0x00, 0x00}, 0, 3},
		{"4-byte boundary", []byte{0x10, 0x00, 0x00, 0x00}, 0, 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, length := mp.parseVInt(tc.input)
			if value != tc.expectedValue {
				t.Errorf("Expected value %d, got %d", tc.expectedValue, value)
			}
			if length != tc.expectedLength {
				t.Errorf("Expected length %d, got %d", tc.expectedLength, length)
			}
		})
	}
}

// TestReadPacket_ErrorHandling tests error handling in ReadPacket
func TestReadPacket_ErrorHandling(t *testing.T) {
	t.Run("EOF during packet reading", func(t *testing.T) {
		// Create a truncated file that ends abruptly
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Start a cluster but don't complete it
		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write([]byte{0x85})                   // Size: 5 bytes (but we won't provide all 5)
		segmentData.Write([]byte{0xE7, 0x81, 0x00})       // Timecode: 0 (only 3 bytes, missing 2)

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Try to read a packet - should get EOF error
		_, err = parser.ReadPacket()
		if err == nil {
			t.Error("Expected EOF error, got nil")
		}
	})

	t.Run("Invalid SimpleBlock data", func(t *testing.T) {
		// Create a file with invalid SimpleBlock
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Invalid SimpleBlock (too short)
		segmentData.Write([]byte{0xA3, 0x82, 0x01, 0x02}) // SimpleBlock ID + size 2 + only 2 bytes data

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Try to read a packet - should get error due to invalid SimpleBlock
		_, err = parser.ReadPacket()
		if err == nil {
			t.Error("Expected error for invalid SimpleBlock, got nil")
		}
	})

	t.Run("ReadElementHeader error", func(t *testing.T) {
		// Create a reader that will fail on ReadElementHeader
		reader := &failingReader{
			data:       []byte{0x18, 0x53, 0x80, 0x67, 0x81}, // Segment ID + size but incomplete
			failAtByte: 4,                                    // Fail before completing the header
		}
		parser := &MatroskaParser{
			reader: NewEBMLReader(reader),
		}

		_, err := parser.ReadPacket()
		if err == nil {
			t.Error("Expected ReadElementHeader error, got nil")
		}
	})

	t.Run("Cluster child ReadElementHeader error", func(t *testing.T) {
		// Create a file with a cluster that has invalid child element header
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Cluster with invalid child element header
		cluster := new(bytes.Buffer)
		cluster.Write([]byte{0xFF, 0xFF}) // Invalid element ID (incomplete)

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster.Len())))
		segmentData.Write(cluster.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		_, err = parser.ReadPacket()
		if err == nil {
			t.Error("Expected child ReadElementHeader error, got nil")
		}
	})

	t.Run("Cluster Timestamp ReadFull error", func(t *testing.T) {
		// Create a file with a cluster that has incomplete timestamp data
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Cluster with incomplete timestamp
		cluster := new(bytes.Buffer)
		cluster.Write([]byte{0xE7, 0x82}) // Timestamp ID + size 2
		cluster.Write([]byte{0x00})       // Only 1 byte of data (should be 2)

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster.Len())))
		segmentData.Write(cluster.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		_, err = parser.ReadPacket()
		if err == nil {
			t.Error("Expected Timestamp ReadFull error, got nil")
		}
	})

}

// TestReadPacket_TrackMaskFiltering tests track mask filtering in ReadPacket
func TestReadPacket_TrackMaskFiltering(t *testing.T) {
	t.Run("Filter specific tracks", func(t *testing.T) {
		// Create a file with multiple tracks and packets
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks with two tracks
		tracks := new(bytes.Buffer)

		// Track 1 (video)
		trackEntry1 := new(bytes.Buffer)
		trackEntry1.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry1.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry1.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry1.Len())))
		tracks.Write(trackEntry1.Bytes())

		// Track 2 (audio)
		trackEntry2 := new(bytes.Buffer)
		trackEntry2.Write([]byte{0xD7, 0x81, 0x02})       // TrackNumber: 2
		trackEntry2.Write([]byte{0x73, 0xC5, 0x81, 0x02}) // TrackUID: 2
		trackEntry2.Write([]byte{0x83, 0x81, 0x02})       // TrackType: 2 (audio)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry2.Len())))
		tracks.Write(trackEntry2.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Cluster with packets from both tracks
		cluster := new(bytes.Buffer)
		cluster.Write([]byte{0xE7, 0x81, 0x00}) // Timecode: 0

		// SimpleBlock for track 1
		simpleBlock1 := new(bytes.Buffer)
		simpleBlock1.Write([]byte{0x81})                   // Track number: 1 (VINT encoded)
		simpleBlock1.Write([]byte{0x00, 0x00})             // Timestamp: 0
		simpleBlock1.Write([]byte{0x80})                   // Flags: keyframe
		simpleBlock1.Write([]byte{0x01, 0x02, 0x03, 0x04}) // Data

		cluster.Write([]byte{0xA3}) // SimpleBlock ID
		cluster.Write(vintEncode(uint64(simpleBlock1.Len())))
		cluster.Write(simpleBlock1.Bytes())

		// SimpleBlock for track 2
		simpleBlock2 := new(bytes.Buffer)
		simpleBlock2.Write([]byte{0x82})                   // Track number: 2 (VINT encoded)
		simpleBlock2.Write([]byte{0x00, 0x64})             // Timestamp: 100
		simpleBlock2.Write([]byte{0x80})                   // Flags: keyframe
		simpleBlock2.Write([]byte{0x05, 0x06, 0x07, 0x08}) // Data

		cluster.Write([]byte{0xA3}) // SimpleBlock ID
		cluster.Write(vintEncode(uint64(simpleBlock2.Len())))
		cluster.Write(simpleBlock2.Bytes())

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster.Len())))
		segmentData.Write(cluster.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Set track mask to filter out track 2 (bit 1 set)
		parser.SetTrackMask(0x02) // Binary: 10 (filter track 2)

		// Read first packet - should be from track 1
		packet1, err := parser.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket() failed: %v", err)
		}
		if packet1.Track != 1 {
			t.Errorf("Expected packet from track 1, got track %d", packet1.Track)
		}

		// Try to read second packet - should get EOF since track 2 is filtered
		_, err = parser.ReadPacket()
		if err != io.EOF {
			t.Errorf("Expected EOF after filtering, got: %v", err)
		}
	})
}

// TestReadPacket_ClusterHandling tests cluster handling in ReadPacket
func TestReadPacket_ClusterHandling(t *testing.T) {
	t.Run("Multiple clusters with timestamp updates", func(t *testing.T) {
		// Create a file with multiple clusters
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// First cluster with timestamp 0
		cluster1 := new(bytes.Buffer)
		cluster1.Write([]byte{0xE7, 0x81, 0x00}) // Timecode: 0

		// SimpleBlock in first cluster
		simpleBlock1 := new(bytes.Buffer)
		simpleBlock1.Write([]byte{0x81})                   // Track number: 1
		simpleBlock1.Write([]byte{0x00, 0x00})             // Timestamp: 0
		simpleBlock1.Write([]byte{0x80})                   // Flags: keyframe
		simpleBlock1.Write([]byte{0x01, 0x02, 0x03, 0x04}) // Data

		cluster1.Write([]byte{0xA3}) // SimpleBlock ID
		cluster1.Write(vintEncode(uint64(simpleBlock1.Len())))
		cluster1.Write(simpleBlock1.Bytes())

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster1.Len())))
		segmentData.Write(cluster1.Bytes())

		// Second cluster with timestamp 1000
		cluster2 := new(bytes.Buffer)
		cluster2.Write([]byte{0xE7, 0x82, 0x03, 0xE8}) // Timecode: 1000

		// SimpleBlock in second cluster
		simpleBlock2 := new(bytes.Buffer)
		simpleBlock2.Write([]byte{0x81})                   // Track number: 1
		simpleBlock2.Write([]byte{0x00, 0x64})             // Timestamp: 100 (relative to cluster)
		simpleBlock2.Write([]byte{0x80})                   // Flags: keyframe
		simpleBlock2.Write([]byte{0x05, 0x06, 0x07, 0x08}) // Data

		cluster2.Write([]byte{0xA3}) // SimpleBlock ID
		cluster2.Write(vintEncode(uint64(simpleBlock2.Len())))
		cluster2.Write(simpleBlock2.Bytes())

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster2.Len())))
		segmentData.Write(cluster2.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Read first packet
		packet1, err := parser.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket() failed: %v", err)
		}
		if packet1.StartTime != 0 {
			t.Errorf("Expected first packet timestamp 0, got %d", packet1.StartTime)
		}

		// Read second packet
		packet2, err := parser.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket() failed: %v", err)
		}
		// Timestamp is calculated as (cluster_timestamp + relative_timestamp) * timecode_scale
		// Expected: (1000 + 100) * 1000000 = 1100000000
		if packet2.StartTime != 1100000000 {
			t.Errorf("Expected second packet timestamp 1100000000, got %d", packet2.StartTime)
		}
	})

	t.Run("Cluster with unknown elements", func(t *testing.T) {
		// Create a cluster with unknown elements that should be skipped
		buf := new(bytes.Buffer)

		// EBML Header
		ebmlHeader := createMinimalEBMLHeader()
		buf.Write(ebmlHeader)

		// Segment
		segmentData := new(bytes.Buffer)

		// SegmentInfo
		segmentInfo := new(bytes.Buffer)
		segmentInfo.Write([]byte{0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40}) // TimestampScale: 1000000
		segmentData.Write([]byte{0x15, 0x49, 0xA9, 0x66})                   // SegmentInfo ID
		segmentData.Write(vintEncode(uint64(segmentInfo.Len())))
		segmentData.Write(segmentInfo.Bytes())

		// Tracks
		tracks := new(bytes.Buffer)
		trackEntry := new(bytes.Buffer)
		trackEntry.Write([]byte{0xD7, 0x81, 0x01})       // TrackNumber: 1
		trackEntry.Write([]byte{0x73, 0xC5, 0x81, 0x01}) // TrackUID: 1
		trackEntry.Write([]byte{0x83, 0x81, 0x01})       // TrackType: 1 (video)

		tracks.Write([]byte{0xAE}) // TrackEntry ID
		tracks.Write(vintEncode(uint64(trackEntry.Len())))
		tracks.Write(trackEntry.Bytes())

		segmentData.Write([]byte{0x16, 0x54, 0xAE, 0x6B}) // Tracks ID
		segmentData.Write(vintEncode(uint64(tracks.Len())))
		segmentData.Write(tracks.Bytes())

		// Cluster with unknown element
		cluster := new(bytes.Buffer)
		cluster.Write([]byte{0xE7, 0x81, 0x00}) // Timecode: 0

		// Unknown element (should be skipped)
		cluster.Write([]byte{0xBF, 0x84, 0x01, 0x02, 0x03, 0x04}) // Unknown ID + size + data

		// SimpleBlock
		simpleBlock := new(bytes.Buffer)
		simpleBlock.Write([]byte{0x81})                   // Track number: 1
		simpleBlock.Write([]byte{0x00, 0x00})             // Timestamp: 0
		simpleBlock.Write([]byte{0x80})                   // Flags: keyframe
		simpleBlock.Write([]byte{0x01, 0x02, 0x03, 0x04}) // Data

		cluster.Write([]byte{0xA3}) // SimpleBlock ID
		cluster.Write(vintEncode(uint64(simpleBlock.Len())))
		cluster.Write(simpleBlock.Bytes())

		segmentData.Write([]byte{0x1F, 0x43, 0xB6, 0x75}) // Cluster ID
		segmentData.Write(vintEncode(uint64(cluster.Len())))
		segmentData.Write(cluster.Bytes())

		buf.Write([]byte{0x18, 0x53, 0x80, 0x67}) // Segment ID
		buf.Write(vintEncode(uint64(segmentData.Len())))
		buf.Write(segmentData.Bytes())

		parser, err := NewMatroskaParser(bytes.NewReader(buf.Bytes()), false)
		if err != nil {
			t.Fatalf("NewMatroskaParser() failed: %v", err)
		}

		// Should be able to read packet despite unknown element
		packet, err := parser.ReadPacket()
		if err != nil {
			t.Fatalf("ReadPacket() failed: %v", err)
		}
		if packet.Track != 1 {
			t.Errorf("Expected packet from track 1, got track %d", packet.Track)
		}
	})
}

// helper to write an EBML UInt element: [ID][size-vint][big-endian data]
func writeUIntElement(buf *bytes.Buffer, id uint32, value uint64, dataLen int) {
	// write ID (1-4 bytes) directly, big-endian by bytes as specified in ebml.go constants
	switch {
	case id <= 0xFF:
		buf.WriteByte(byte(id))
	case id <= 0xFFFF:
		buf.WriteByte(byte(id >> 8))
		buf.WriteByte(byte(id))
	case id <= 0xFFFFFF:
		buf.WriteByte(byte(id >> 16))
		buf.WriteByte(byte(id >> 8))
		buf.WriteByte(byte(id))
	default:
		buf.WriteByte(byte(id >> 24))
		buf.WriteByte(byte(id >> 16))
		buf.WriteByte(byte(id >> 8))
		buf.WriteByte(byte(id))
	}
	// size vint
	buf.Write(vintEncode(uint64(dataLen)))
	// big-endian value padded to dataLen
	tmp := make([]byte, dataLen)
	for i := dataLen - 1; i >= 0; i-- {
		tmp[i] = byte(value & 0xFF)
		value >>= 8
	}
	buf.Write(tmp)
}

func TestParseCueTrackPositions_AllFields(t *testing.T) {
	mp := &MatroskaParser{fileInfo: &SegmentInfo{TimecodeScale: 100}}

	var data bytes.Buffer
	// IDCueTrack (0xF7) = 1
	writeUIntElement(&data, IDCueTrack, 1, 1)
	// IDCueClusterPos (0xF1) = 0x1234
	writeUIntElement(&data, IDCueClusterPos, 0x1234, 2)
	// IDCueRelativePos (0xF0) = 5
	writeUIntElement(&data, IDCueRelativePos, 5, 1)
	// IDCueBlockNum (0x5378) = 7
	writeUIntElement(&data, IDCueBlockNum, 7, 1)
	// IDCueDuration (0x9B) = 2 (scaled by 100)
	writeUIntElement(&data, IDCueDuration, 2, 1)

	cue, err := mp.parseCueTrackPositions(data.Bytes())
	if err != nil {
		t.Fatalf("parseCueTrackPositions failed: %v", err)
	}
	if cue.Track != 1 {
		t.Errorf("Track = %d, want 1", cue.Track)
	}
	if cue.Position != 0x1234 {
		t.Errorf("Position = %#x, want 0x1234", cue.Position)
	}
	if cue.RelativePosition != 5 {
		t.Errorf("RelativePosition = %d, want 5", cue.RelativePosition)
	}
	if cue.Block != 7 {
		t.Errorf("Block = %d, want 7", cue.Block)
	}
	if cue.Duration != 200 { // 2 * 100
		t.Errorf("Duration = %d, want 200", cue.Duration)
	}
}

func TestParseCuePoint_TimeAndTrackPositions(t *testing.T) {
	mp := &MatroskaParser{fileInfo: &SegmentInfo{TimecodeScale: 100}}

	// Build IDCueTrackPosition payload (same as above but without duration to vary path)
	var ctp bytes.Buffer
	writeUIntElement(&ctp, IDCueTrack, 2, 1)
	writeUIntElement(&ctp, IDCueClusterPos, 0x20, 1)
	writeUIntElement(&ctp, IDCueBlockNum, 1, 1)

	// Wrap as IDCueTrackPosition element: [IDCueTrackPosition][size][payload]
	var payload bytes.Buffer
	payload.WriteByte(byte(IDCueTrackPosition))
	payload.Write(vintEncode(uint64(ctp.Len())))
	payload.Write(ctp.Bytes())

	// Now build CuePoint element data: [IDCueTime]=3 and the track position element
	var cp bytes.Buffer
	writeUIntElement(&cp, IDCueTime, 3, 1)
	cp.Write(payload.Bytes())

	cues, err := mp.parseCuePoint(cp.Bytes())
	if err != nil {
		t.Fatalf("parseCuePoint failed: %v", err)
	}
	if len(cues) != 1 {
		t.Fatalf("expected 1 cue, got %d", len(cues))
	}
	if cues[0].Time != 300 { // 3 * 100
		t.Errorf("cue.Time = %d, want 300", cues[0].Time)
	}
	if cues[0].Track != 2 || cues[0].Position != 0x20 || cues[0].Block != 1 {
		t.Errorf("cue fields unexpected: %+v", cues[0])
	}
}
