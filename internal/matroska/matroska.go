// Package matroska provides a pure Go implementation for parsing and demuxing Matroska/EBML files.
//
// Matroska is an open standard, free container format, a file format that can hold an unlimited
// number of video, audio, picture, or subtitle tracks in one file. It is intended to serve as
// a universal format for storing common multimedia content, like movies or TV shows.
//
// This package provides a high-level API for reading Matroska files, extracting track information,
// and reading media packets. The main entry point is the Demuxer struct, which can be created
// using either NewDemuxer for seekable inputs or NewStreamingDemuxer for non-seekable streams.
//
// Basic usage:
//
//	// Open a Matroska file
//	file, err := os.Open("video.mkv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Create a new demuxer
//	demuxer, err := matroska.NewDemuxer(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer demuxer.Close()
//
//	// Get file information
//	fileInfo, err := demuxer.GetFileInfo()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get track information
//	numTracks, err := demuxer.GetNumTracks()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for i := uint(0); i < numTracks; i++ {
//	    trackInfo, err := demuxer.GetTrackInfo(i)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Track %d: %s\n", i, trackInfo.Name)
//	}
//
//	// Read packets
//	for {
//	    packet, err := demuxer.ReadPacket()
//	    if err != nil {
//	        if err == io.EOF {
//	            break
//	        }
//	        log.Fatal(err)
//	    }
//	    // Process packet...
//	}
package matroska

import (
	"fmt"
	"io"
)

// Demuxer is a Matroska demuxer using pure Go implementation.
//
// The Demuxer struct provides the main interface for parsing and reading Matroska files.
// It encapsulates the underlying parser and provides methods for accessing track information,
// file metadata, and reading media packets. The Demuxer can work with both seekable and
// non-seekable input streams.
//
// For seekable inputs, use NewDemuxer. For non-seekable streams (like network streams),
// use NewStreamingDemuxer.
type Demuxer struct {
	parser *MatroskaParser
	reader io.ReadSeeker
}

// NewDemuxer creates a new Matroska demuxer from r.
//
// NewDemuxer creates a new Matroska demuxer from a seekable input stream.
// The input reader must implement both io.Reader and io.Seeker interfaces.
//
// This function initializes the underlying Matroska parser and returns a Demuxer
// instance that can be used to extract track information, file metadata, and read
// media packets from the Matroska file.
//
// Example:
//
//	file, err := os.Open("video.mkv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	demuxer, err := matroska.NewDemuxer(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer demuxer.Close()
//
// Parameters:
//   - r: An io.ReadSeeker that provides access to the Matroska file data.
//
// Returns:
//   - *Demuxer: A new Demuxer instance for the given input.
//   - error: An error if the demuxer could not be created.
func NewDemuxer(r io.ReadSeeker, elements ...uint32) (*Demuxer, error) {
	parser, err := NewMatroskaParser(r, false, elements...)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	return &Demuxer{
		parser: parser,
		reader: r,
	}, nil
}

// NewStreamingDemuxer creates a new Matroska demuxer from an
// io.Reader that has no ability to seek on the input stream.
//
// NewStreamingDemuxer creates a new Matroska demuxer from a non-seekable input stream.
// This function is designed for streaming scenarios where the input reader only implements
// the io.Reader interface and cannot seek backwards in the stream, such as network streams
// or pipes.
//
// The function wraps the non-seekable reader with a fakeSeeker that provides a minimal
// seeking interface, allowing the Matroska parser to work with streams that don't support
// random access. Note that some operations, like seeking to specific timecodes or accessing
// cues, may not work as efficiently or may not be available when using a streaming demuxer.
//
// Example:
//
//	resp, err := http.Get("http://example.com/video.mkv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer resp.Body.Close()
//
//	demuxer, err := matroska.NewStreamingDemuxer(resp.Body)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer demuxer.Close()
//
// Parameters:
//   - r: An io.Reader that provides access to the Matroska stream data.
//
// Returns:
//   - *Demuxer: A new Demuxer instance for the given input stream.
//   - error: An error if the demuxer could not be created.
func NewStreamingDemuxer(r io.Reader) (*Demuxer, error) {
	fs := &fakeSeeker{r: r}
	parser, err := NewMatroskaParser(fs, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming parser: %w", err)
	}

	return &Demuxer{
		parser: parser,
		reader: fs,
	}, nil
}

// Close closes a demuxer.
//
// Close releases any resources associated with the Demuxer.
// In this pure Go implementation, no explicit cleanup is required as
// Go's garbage collector handles memory management automatically.
// However, calling Close is still recommended for consistency and
// to allow for future implementations that might require cleanup.
//
// Example:
//
//	demuxer, err := matroska.NewDemuxer(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer demuxer.Close()
//
//	// Use demuxer...
func (d *Demuxer) Close() {
	// Pure Go implementation doesn't need explicit cleanup
}

// GetNumTracks gets the number of tracks available to a given demuxer.
//
// This function returns the total number of tracks (video, audio, subtitle, etc.)
// contained in the Matroska file. Track indices range from 0 to the returned
// value minus one, and can be used with GetTrackInfo to retrieve detailed
// information about each track.
//
// Example:
//
//	numTracks, err := demuxer.GetNumTracks()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("File contains %d tracks\n", numTracks)
//
// Returns:
//   - uint: The number of tracks in the Matroska file.
//   - error: An error if the track count could not be retrieved.
func (d *Demuxer) GetNumTracks() (uint, error) {
	return d.parser.GetNumTracks(), nil
}

// GetTrackInfo returns all track-level information available for a given track,
// where track is less than what is returned by GetNumTracks.
//
// This function retrieves detailed information about a specific track, including
// its type (video, audio, subtitle), codec, language, and other metadata.
// The track parameter must be a valid track index between 0 and GetNumTracks()-1.
//
// Example:
//
//	numTracks, err := demuxer.GetNumTracks()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for i := uint(0); i < numTracks; i++ {
//	    trackInfo, err := demuxer.GetTrackInfo(i)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Track %d: %s (%s)\n", i, trackInfo.Name, trackInfo.CodecID)
//	}
//
// Parameters:
//   - track: The index of the track to retrieve information for.
//
// Returns:
//   - *TrackInfo: Detailed information about the track.
//   - error: An error if the track information could not be retrieved or if the track index is invalid.
func (d *Demuxer) GetTrackInfo(track uint) (*TrackInfo, error) {
	trackInfo := d.parser.GetTrackInfo(track)
	if trackInfo == nil {
		return nil, fmt.Errorf("track %d not found", track)
	}
	return trackInfo, nil
}

// GetFileInfo gets all top-level (whole file) info available for a given
// demuxer.
//
// This function retrieves metadata about the Matroska file itself, including
// title, duration, muxing application, writing application, and other
// file-level information. This information is stored in the SegmentInfo
// structure.
//
// Example:
//
//	fileInfo, err := demuxer.GetFileInfo()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Title: %s\n", fileInfo.Title)
//	fmt.Printf("Duration: %f seconds\n", float64(fileInfo.Duration)/1000000000)
//	fmt.Printf("Muxing App: %s\n", fileInfo.MuxingApp)
//
// Returns:
//   - *SegmentInfo: File-level metadata about the Matroska file.
//   - error: An error if the file information could not be retrieved.
func (d *Demuxer) GetFileInfo() (*SegmentInfo, error) {
	fileInfo := d.parser.GetFileInfo()
	if fileInfo == nil {
		return nil, fmt.Errorf("no file info available")
	}
	return fileInfo, nil
}

// GetAttachments returns information on all available attachments
// for a given demuxer. The returned slice may be of length 0.
//
// This function retrieves information about any attachments embedded in the
// Matroska file, such as fonts, images, or other files. Attachments are
// typically used for fonts required by subtitle tracks or cover art.
//
// Example:
//
//	attachments := demuxer.GetAttachments()
//	for _, attachment := range attachments {
//	    fmt.Printf("Attachment: %s (%s)\n", attachment.Name, attachment.MimeType)
//	}
//
// Returns:
//   - []*Attachment: A slice of attachment information. May be empty if no attachments are present.
func (d *Demuxer) GetAttachments() []*Attachment {
	return d.parser.GetAttachments()
}

// GetChapters returns all chapters for a given demuxer. The returned slice may
// be of length 0.
//
// This function retrieves information about the chapter structure of the Matroska file.
// Chapters can be used to navigate through the content, similar to DVD chapters.
// The returned chapters may include nested chapters (editions).
//
// Example:
//
//	chapters := demuxer.GetChapters()
//	for _, chapter := range chapters {
//	    fmt.Printf("Chapter: %s (%d-%d)\n", chapter.Display.String, chapter.Start, chapter.End)
//	}
//
// Returns:
//   - []*Chapter: A slice of chapter information. May be empty if no chapters are present.
func (d *Demuxer) GetChapters() []*Chapter {
	return d.parser.GetChapters()
}

// GetTags returns all tags for a given demuxer. The returned slice may be of
// length 0.
//
// This function retrieves metadata tags associated with the Matroska file or
// specific tracks within the file. Tags can contain information such as title,
// artist, album, genre, and other descriptive metadata.
//
// Example:
//
//	tags := demuxer.GetTags()
//	for _, tag := range tags {
//	    for _, simpleTag := range tag.SimpleTags {
//	        fmt.Printf("Tag: %s = %s\n", simpleTag.Name, simpleTag.Value)
//	    }
//	}
//
// Returns:
//   - []*Tag: A slice of tag information. May be empty if no tags are present.
func (d *Demuxer) GetTags() []*Tag {
	return d.parser.GetTags()
}

// GetCues returns all cues for a given demuxer. The returned slice may be
// of length 0.
//
// This function retrieves cue points from the Matroska file. Cues are indexing
// information that allows for efficient seeking to specific timecodes in the file.
// They contain mappings between timecodes and their positions in the file.
//
// Example:
//
//	cues := demuxer.GetCues()
//	for _, cue := range cues {
//	    fmt.Printf("Cue at time %d, position %d\n", cue.Time, cue.Position)
//	}
//
// Returns:
//   - []*Cue: A slice of cue information. May be empty if no cues are present.
func (d *Demuxer) GetCues() []*Cue {
	return d.parser.GetCues()
}

// GetSegment returns the position of the segment.
//
// This function returns the file position (offset) where the Matroska segment
// begins. The segment is the main container for all tracks, chapters, tags,
// and other metadata in the file.
//
// Example:
//
//	segmentPos := demuxer.GetSegment()
//	fmt.Printf("Segment starts at position %d\n", segmentPos)
//
// Returns:
//   - uint64: The file position where the segment begins.
func (d *Demuxer) GetSegment() uint64 {
	return d.parser.GetSegment()
}

// GetSegmentTop returns the position of the next byte after the segment.
//
// This function returns the file position (offset) immediately after the end
// of the Matroska segment. This can be useful for determining the size of
// the segment or for locating other elements that may follow it in the file.
//
// Example:
//
//	segmentStart := demuxer.GetSegment()
//	segmentEnd := demuxer.GetSegmentTop()
//	segmentSize := segmentEnd - segmentStart
//	fmt.Printf("Segment size: %d bytes\n", segmentSize)
//
// Returns:
//   - uint64: The file position after the end of the segment.
func (d *Demuxer) GetSegmentTop() uint64 {
	return d.parser.GetSegmentTop()
}

// GetCuesPos returns the position of the cues in the stream.
//
// This function returns the file position (offset) where the cues element
// begins in the Matroska file. The cues element contains indexing information
// that allows for efficient seeking.
//
// Example:
//
//	cuesPos := demuxer.GetCuesPos()
//	fmt.Printf("Cues element at position %d\n", cuesPos)
//
// Returns:
//   - uint64: The file position where the cues element begins.
func (d *Demuxer) GetCuesPos() uint64 {
	return d.parser.GetCuesPos()
}

// GetCuesTopPos returns the position of the byte after the end of the cues.
//
// This function returns the file position (offset) immediately after the end
// of the cues element. This can be useful for determining the size of the
// cues element.
//
// Example:
//
//	cuesStart := demuxer.GetCuesPos()
//	cuesEnd := demuxer.GetCuesTopPos()
//	cuesSize := cuesEnd - cuesStart
//	fmt.Printf("Cues element size: %d bytes\n", cuesSize)
//
// Returns:
//   - uint64: The file position after the end of the cues element.
func (d *Demuxer) GetCuesTopPos() uint64 {
	return d.parser.GetCuesTopPos()
}

// Seek seeks to a given timecode.
//
// Flags here may be: 0 (normal seek), matroska.SeekToPrevKeyFrame,
// or matoska.SeekToPrevKeyFrameStrict
//
// This function moves the playback position to the specified
// timecode in the Matroska file. The flags parameter can be used to control
// the seeking behavior, such as whether to seek to the previous keyframe
// or to perform a strict seek.
//
// Parameters:
//   - timecode: The target timecode to seek to, in nanoseconds.
//   - flags: Seek behavior flags. May be 0 (normal seek), SeekToPrevKeyFrame,
//     or SeekToPrevKeyFrameStrict.
func (d *Demuxer) Seek(timecode uint64, flags uint32) {
	if d.parser.noSeeking {
		return
	}
	_ = d.parser.Seek(timecode, flags)
}

// SeekCueAware seeks to a given timecode while taking cues into account
//
// Flags here may be: 0 (normal seek), matroska.SeekToPrevKeyFrame,
// or matoska.SeekToPrevKeyFrameStrict
//
// fuzzy defines whether a fuzzy seek will be used or not.
//
// This function moves the playback position to the specified
// timecode in the Matroska file, using the cue information for more accurate
// seeking. The fuzzy parameter controls whether a fuzzy seek (approximate
// position) is acceptable if an exact match cannot be found.
//
// Parameters:
//   - timecode: The target timecode to seek to, in nanoseconds.
//   - flags: Seek behavior flags. May be 0 (normal seek), SeekToPrevKeyFrame,
//     or SeekToPrevKeyFrameStrict.
//   - fuzzy: Whether to allow fuzzy seeking (approximate positions).
func (d *Demuxer) SeekCueAware(timecode uint64, flags uint32, fuzzy bool) {
	// fuzzy is not supported yet, just call normal seek
	d.Seek(timecode, flags)
}

// SkipToKeyframe skips to the next keyframe in a stream.
//
// This function advances the playback position to the next
// keyframe in the current track. Keyframes are frames that can be decoded
// without reference to previous frames, making them ideal starting points
// for seeking or resuming playback.
func (d *Demuxer) SkipToKeyframe() {
	d.parser.SkipToKeyframe()
}

// GetLowestQTimecode returns the lowest queued timecode in the demuxer.
//
// This function returns the timecode of the earliest packet
// that is currently queued in the demuxer. This can be useful for buffering
// and synchronization purposes.
//
// Returns:
//   - uint64: The timecode of the lowest queued packet.
func (d *Demuxer) GetLowestQTimecode() uint64 {
	if d.parser.fileInfo == nil {
		return 0
	}
	return d.parser.clusterTimestamp * d.parser.fileInfo.TimecodeScale
}

// SetTrackMask sets the demuxer's track mask; that is, it tells the demuxer
// which tracks to skip, and which to use. Any tracks with ones in their bit
// positions will be ignored.
//
// Calling this withh cause all parsed and queued frames to be discarded.
//
// This function allows filtering of tracks during playback or
// processing. The mask is a bitmask where each bit corresponds to a track
// index. Setting a bit to 1 will cause that track to be ignored.
//
// Parameters:
//   - mask: A bitmask specifying which tracks to ignore. A bit set to 1 at
//     position N will cause track N to be ignored.
func (d *Demuxer) SetTrackMask(mask uint64) {
	d.parser.SetTrackMask(mask)
}

// ReadPacketMask is the same as ReadPacket except with a track mask.
//
// This function is intended to read the next packet from the demuxer while
// respecting the track mask specified by the mask parameter. Currently,
// the mask parameter is ignored and the function behaves identically to
// ReadPacket.
//
// Parameters:
//   - mask: A bitmask specifying which tracks to ignore. Currently ignored.
//
// Returns:
//   - *Packet: The next packet from the demuxer.
//   - error: An error if a packet could not be read.
func (d *Demuxer) ReadPacketMask(mask uint64) (*Packet, error) {
	// For now, ignore mask and read next packet
	return d.parser.ReadPacket()
}

// ReadPacket returns the next packet from a demuxer.
//
// This function reads and returns the next media packet from the Matroska file.
// Packets contain the actual media data (video frames, audio samples, etc.)
// along with metadata such as the track number, timecode, and flags.
//
// Example:
//
//	for {
//	    packet, err := demuxer.ReadPacket()
//	    if err != nil {
//	        if err == io.EOF {
//	            break
//	        }
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Packet: track=%d, time=%d, size=%d\n",
//	        packet.Track, packet.StartTime, len(packet.Data))
//	    // Process packet data...
//	}
//
// Returns:
//   - *Packet: The next packet from the demuxer.
//   - error: An error if a packet could not be read, or io.EOF if the end of the file has been reached.
func (d *Demuxer) ReadPacket() (*Packet, error) {
	return d.parser.ReadPacket()
}
