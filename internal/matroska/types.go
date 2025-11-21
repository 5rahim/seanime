// Package matroska defines all the data structures and constants used in the Matroska/EBML project.
//
// Matroska is an extensible open standard Audio/video container format. This package provides
// the core types and constants needed to parse, create, and manipulate Matroska files.
//
// The package includes definitions for:
//   - Compression types used in Matroska tracks
//   - Track types (video, audio, subtitle)
//   - Tag target types
//   - Seek types for navigation
//   - Packet flags for demuxed data
//   - Core data structures like Packet, TrackInfo, SegmentInfo, Attachment, Chapter, Cue, and Tag
//
// These types are used throughout the project to represent different aspects of Matroska files
// and serve as the central location for all data type definitions used by other files in the project.
package matroska

// Matroska compression types
//
// These constants define the compression algorithms that can be applied to Matroska tracks.
const (
	// CompZlib indicates zlib compression (RFC 1950).
	CompZlib = 0
	// CompBzip indicates bzip2 compression.
	CompBzip = 1
	// CompLZO1X indicates LZO1X compression.
	CompLZO1X = 2
	// CompPrepend indicates header stripping compression where data is prepended during decompression.
	CompPrepend = 3
)

// Track types
//
// These constants define the different types of tracks that can be present in a Matroska file.
const (
	// TypeVideo indicates a video track.
	TypeVideo = 1
	// TypeAudio indicates an audio track.
	TypeAudio = 2
	// TypeSubtitle indicates a subtitle track.
	TypeSubtitle = 17
)

// Tag target types
//
// These constants define the different types of targets that Matroska tags can be applied to.
const (
	// TargetTrack indicates that the tag applies to a specific track.
	TargetTrack = 0
	// TargetChapter indicates that the tag applies to a specific chapter.
	TargetChapter = 1
	// TargetAttachment indicates that the tag applies to a specific attachment.
	TargetAttachment = 2
	// TargetEdition indicates that the tag applies to a specific edition.
	TargetEdition = 3
)

// Seek types
//
// These constants define the different seeking behaviors that can be used when navigating
// through a Matroska file. They control how the player handles seeking operations.
const (
	// SeekToPrevKeyFrame seeks to the previous key frame before the requested position.
	SeekToPrevKeyFrame = 1
	// SeekToPrevKeyFrameStrict seeks to the previous key frame before the requested position,
	// but does not allow going beyond that point even if it means displaying nothing.
	SeekToPrevKeyFrameStrict = 2
)

// Packet flags
//
// These constants define flags that can be returned with a matroska.Packet in its Flags member.
// They provide additional information about the packet's properties and characteristics.
const (
	// UnknownStart indicates that the packet starts at an unknown position.
	UnknownStart = 0x00000001
	// UnknownEnd indicates that the packet ends at an unknown position.
	UnknownEnd = 0x00000002
	// KF indicates that the packet is a key frame.
	KF = 0x00000004
	// GAP indicates that the packet is a gap packet, which should be skipped during playback.
	GAP = 0x00800000
	// StreamMask is a bitmask used to extract the stream number from the Flags field.
	StreamMask = 0xff000000
	// StreamShift is the number of bits to shift right to extract the stream number from the Flags field.
	StreamShift = 24
)

// Packet contains a demuxed packet from a Matroska file.
//
// A Packet represents a single unit of media data that has been extracted (demuxed) from
// a Matroska container. It contains all the necessary information to process the media data,
// including timing information, track association, and the actual data payload.
type Packet struct {
	// Track is the track number this packet belongs to.
	// This corresponds to the TrackInfo.Number of the track.
	Track uint8
	// StartTime is the start time of this packet in nanoseconds.
	// This is the timestamp when the packet should be presented.
	StartTime uint64
	// EndTime is the end time of this packet in nanoseconds.
	// This is the timestamp when the packet should stop being presented.
	EndTime uint64
	// FilePos is the position in the input stream where this packet is located.
	// This can be useful for seeking or debugging purposes.
	FilePos uint64
	// Data contains the actual packet data.
	// This is the raw media data that needs to be decoded by the appropriate codec.
	Data []byte
	// Flags contains any packet flags. See the packet flag constants for details.
	// These flags provide additional information about the packet's properties.
	Flags uint32
	// Discard indicates whether this packet can be discarded.
	// A non-zero value suggests that the packet can be safely discarded without affecting playback.
	Discard int64
}

// TrackInfo contains information about a track in a Matroska file.
//
// A TrackInfo structure holds all metadata and configuration information for a single
// track within a Matroska file. This includes general track properties, codec information,
// and type-specific settings for video, audio, or subtitle tracks.
type TrackInfo struct {
	// Number is the track number used to identify this track within the Matroska file.
	// Track numbers are unique within a segment and are used to associate packets with tracks.
	Number uint8
	// Type is the track type. See the track type constants (TypeVideo, TypeAudio, TypeSubtitle).
	Type uint8
	// TrackOverlay specifies whether this track should be overlaid on another track.
	// This is typically used for subtitle or menu tracks that need to be displayed over video.
	TrackOverlay uint8
	// UID is a unique identifier for this track.
	// This allows tracks to be referenced even if their numbers change.
	UID uint64
	// MinCache is the minimum amount of frames a player should keep around to
	// be able to play back this file properly, e.g. the min DPB (Decoded Picture Buffer) size.
	MinCache uint64
	// MaxCache is the largest possible size a player could need for its cache
	// in order to play back this file smoothly.
	MaxCache uint64
	// DefaultDuration is the track's default duration in nanoseconds, which can be used,
	// for example, to calculate the duration of the last packet.
	DefaultDuration uint64
	// CodecDelay is any inherent delay required by the codec in nanoseconds.
	// This is used to ensure proper audio/video synchronization.
	CodecDelay uint64
	// SeekPreRoll is any pre-roll that must be applied after seeking for this codec.
	// This ensures that the decoder has enough data to start playback correctly after a seek.
	SeekPreRoll uint64
	// TimecodeScale is the timescale for this track's timecodes.
	// This is used to convert between the track's internal timecode and actual time.
	TimecodeScale float64
	// CodecPrivate contains codec-specific private data that should be passed to decoders.
	// This typically includes initialization data required by the codec.
	CodecPrivate []byte
	// CompMethod is the track compression method. See the compression method constants.
	CompMethod uint32
	// CompMethodPrivate contains any private data that should be passed to the decompressor
	// used to decompress the track.
	CompMethodPrivate []byte
	// MaxBlockAdditionID is the maximum ID of the BlockAdditional elements for this track.
	// This is used to identify additional data blocks associated with the track.
	MaxBlockAdditionID uint32

	// Enabled indicates whether this track is enabled and should be played.
	Enabled bool
	// Default indicates whether this track is on by default.
	// If true, the track should be enabled unless the user explicitly disables it.
	Default bool
	// Forced indicates whether this track is forced on.
	// Forced tracks are typically used for subtitles that must be displayed regardless of user preferences.
	Forced bool
	// Lacing indicates whether this track uses lacing.
	// Lacing is a method of reducing overhead by storing multiple small blocks in a single frame.
	Lacing bool
	// DecodeAll indicates whether this track has Error Resilience capabilities.
	// If true, the player should attempt to decode all frames even if some are corrupted.
	DecodeAll bool
	// CompEnabled indicates whether this track has compression enabled.
	CompEnabled bool

	// Video contains video-specific information. Only valid if the track is a video track.
	Video struct {
		// StereoMode is the stereo 3D mode used, if any.
		// This defines how the video should be displayed for 3D playback.
		StereoMode uint8
		// DisplayUnit is the unit used for DisplayWidth and DisplayHeight.
		// This defines whether the display dimensions are in pixels, centimeters, or inches.
		DisplayUnit uint8
		// AspectRatioType defines what type of resizing is needed for the aspect ratio:
		//     0 = free resizing
		//     1 = keep aspect ratio
		//     2 = fixed
		AspectRatioType uint8
		// PixelWidth is the width of the video in pixels.
		PixelWidth uint32
		// PixelHeight is the height of the video in pixels.
		PixelHeight uint32
		// DisplayWidth is the width at which the video should be displayed.
		// This may differ from PixelWidth if the video needs to be scaled.
		DisplayWidth uint32
		// DisplayHeight is the height at which the video should be displayed.
		// This may differ from PixelHeight if the video needs to be scaled.
		DisplayHeight uint32
		// CropL is the number of pixels to crop from the left side of the video.
		CropL uint32
		// CropT is the number of pixels to crop from the top of the video.
		CropT uint32
		// CropR is the number of pixels to crop from the right side of the video.
		CropR uint32
		// CropB is the number of pixels to crop from the bottom of the video.
		CropB uint32
		// ColourSpace is the colorspace of the video, similar to biCompression from BITMAPINFOHEADER.
		ColourSpace uint32
		// GammaValue is the gamma value to use for color adjustment.
		GammaValue float64
		// Colour contains detailed color information for the video.
		Colour struct {
			// MatrixCoefficients defines the matrix coefficients used for the video.
			// See: ISO/IEC 23091-4/ITU-T H.273 for standard values.
			MatrixCoefficients uint32
			// BitsPerChannel is the number of bits per color channel.
			BitsPerChannel uint32
			// ChromaSubsamplingHorz is the base 2 logarithm of horizontal chroma subsampling.
			ChromaSubsamplingHorz uint32
			// ChromaSubsamplingVert is the base 2 logarithm of vertical chroma subsampling.
			ChromaSubsamplingVert uint32
			// CbSubsamplingHorz is the amount of pixels to remove in the Cb channel for every pixel
			// not removed horizontally. This is additive with ChromaSubsamplingHorz.
			CbSubsamplingHorz uint32
			// CbSubsamplingVert is the amount of pixels to remove in the Cb channel for every pixel
			// not removed vertically. This is additive with ChromaSubsamplingVert.
			CbSubsamplingVert uint32
			// ChromaSitingHorz is the horizontal chroma position:
			//     0 = unspecified,
			//     1 = left collocated
			//     2 = half
			ChromaSitingHorz uint32
			// ChromaSitingVert is the vertical chroma position:
			//     0 = unspecified
			//     1 = left collocated
			//     2 = half
			ChromaSitingVert uint32
			// Range defines the color range:
			//     0 = unspecified
			//     1 = broadcast range (16-235)
			//     2 = full range (0-255)
			//     3 = defined by MatrixCoefficients / TransferCharacteristics
			Range uint32
			// TransferCharacteristics defines the transfer characteristics of the video.
			// See: ISO/IEC 23091-4/ITU-T H.273 for standard values.
			TransferCharacteristics uint32
			// Primaries defines the color primaries of the video.
			// See: ISO/IEC 23091-4/ITU-T H.273 for standard values.
			Primaries uint32
			// MaxCLL is the maximum content light level in nits.
			// This is used for HDR content to indicate the brightest point in the content.
			MaxCLL uint32
			// MaxFALL is the maximum frame-average light level in nits.
			// This is used for HDR content to indicate the average brightness of the brightest frame.
			MaxFALL uint32
			// MasteringMetadata contains mastering display metadata for HDR content.
			MasteringMetadata struct {
				// PrimaryRChromaticityX is the X chromaticity coordinate of the red primary.
				PrimaryRChromaticityX float32
				// PrimaryRChromaticityY is the Y chromaticity coordinate of the red primary.
				PrimaryRChromaticityY float32
				// PrimaryGChromaticityX is the X chromaticity coordinate of the green primary.
				PrimaryGChromaticityX float32
				// PrimaryGChromaticityY is the Y chromaticity coordinate of the green primary.
				PrimaryGChromaticityY float32
				// PrimaryBChromaticityX is the X chromaticity coordinate of the blue primary.
				PrimaryBChromaticityX float32
				// PrimaryBChromaticityY is the Y chromaticity coordinate of the blue primary.
				PrimaryBChromaticityY float32
				// WhitePointChromaticityX is the X chromaticity coordinate of the white point.
				WhitePointChromaticityX float32
				// WhitePointChromaticityY is the Y chromaticity coordinate of the white point.
				WhitePointChromaticityY float32
				// LuminanceMax is the maximum luminance of the display in nits.
				LuminanceMax float32
				// LuminanceMin is the minimum luminance of the display in nits.
				LuminanceMin float32
			}
		}
		// Interlaced indicates whether the video is interlaced.
		// If true, the video consists of interlaced fields rather than progressive frames.
		Interlaced bool
	}
	// Audio contains audio-specific information. Only valid if the track is an audio track.
	Audio struct {
		// SamplingFreq is the sampling frequency of the audio in Hz.
		SamplingFreq float64
		// OutputSamplingFreq is the sampling frequency to output during playback in Hz.
		// This may differ from SamplingFreq if resampling is required.
		OutputSamplingFreq float64
		// Channels is the number of audio channels.
		Channels uint8
		// BitDepth is the bit depth of the audio samples.
		BitDepth uint8
	}

	// Name is the human-readable name of the track.
	// This can be displayed to users to identify the track.
	Name string
	// Language is the language code of the track.
	// This follows the ISO 639-2 language codes (e.g., "eng" for English).
	Language string
	// LanguageIETF is the IETF language tag of the track.
	LanguageIETF string
	// CodecID is the identifier for the codec used by this track.
	// This is a string that identifies the codec, such as "V_MPEG4/ISO/AVC" for H.264 video.
	CodecID string
}

// SegmentInfo contains file-level (segment) information about a Matroska stream.
//
// A SegmentInfo structure holds metadata about the entire Matroska file or segment.
// This includes general information like title, duration, creation date, and
// relationships to other files in a sequence.
type SegmentInfo struct {
	// UID is the top-level unique identifier for this segment.
	// This is a 128-bit UUID that uniquely identifies this segment.
	UID [16]byte
	// PrevUID is the UID of any files which should be played back before this one.
	// This is used to create a sequence of related files.
	PrevUID [16]byte
	// NextUID is the UID of any files which should be played back after this one.
	// This is used to create a sequence of related files.
	NextUID [16]byte
	// Filename is the filename of this segment.
	// This is a human-readable name for the file.
	Filename string
	// PrevFilename is the filename of any files which should be played back before this one.
	// This corresponds to the file with UID equal to PrevUID.
	PrevFilename string
	// NextFilename is the filename of any files which should be played back after this one.
	// This corresponds to the file with UID equal to NextUID.
	NextFilename string
	// Title is the title of the segment.
	// This is a human-readable title for the content.
	Title string
	// MuxingApp is the name of the application that muxed this file.
	// This is useful for debugging and compatibility purposes.
	MuxingApp string
	// WritingApp is the name of the library that muxed this file.
	// This is useful for debugging and compatibility purposes.
	WritingApp string
	// TimecodeScale is the timescale of any timecodes in the segment.
	// This is used to convert timecodes to nanoseconds. The default is 1000000.
	TimecodeScale uint64
	// Duration is the file's duration in nanoseconds. May be 0 if unknown.
	Duration uint64
	// DateUTC is the date the file was created on, in nanoseconds since the Unix epoch.
	// This can be used to determine when the file was created.
	DateUTC int64
	// DateUTCValid indicates whether or not DateUTC can be considered valid.
	// If false, the DateUTC value should not be used.
	DateUTCValid bool
}

// Attachment contains information about a Matroska attachment.
//
// Matroska files can contain attached files, such as fonts, images, or other metadata.
// The Attachment structure holds information about these attached files, including
// their location, size, and metadata.
type Attachment struct {
	// Position is the attachment's position within the stream.
	// This is the byte offset where the attachment data begins.
	Position uint64
	// Length is the attachment's length in bytes.
	// This is the size of the attachment data.
	Length uint64
	// UID is the attachment's unique identifier.
	// This allows the attachment to be referenced by other elements.
	UID uint64
	// Name is the name of the attachment.
	// This is a human-readable name for the attached file.
	Name string
	// Description is a description of the attachment.
	// This provides additional information about the attached file.
	Description string
	// MimeType is the attachment's MIME type.
	// This identifies the type of the attached file, such as "font/ttf" or "image/jpeg".
	MimeType string
	// Data is the attachment's data.
	Data []byte
}

// ChapterDisplay contains display information for a given Chapter.
//
// A ChapterDisplay structure holds the human-readable information for a chapter,
// including its name, language, and country association. This allows chapters
// to be displayed in multiple languages and with country-specific variations.
type ChapterDisplay struct {
	// String is the display string for the chapter, usually the chapter name.
	// This is the human-readable text that will be displayed to users.
	String string
	// Language is the language code for this chapter display.
	// This follows the ISO 639-2 language codes (e.g., "eng" for English).
	Language string
	// Country is the country this chapter is associated with.
	// This is used when there may be language dialects that vary by country.
	Country string
}

// ChapterCommand represents a command associated with a chapter.
//
// Chapter commands are used to execute specific actions at certain times during
// chapter playback. This can be used for interactive content or special effects.
type ChapterCommand struct {
	// Time is the time when the command should be executed.
	// This is relative to the start of the chapter.
	Time uint32
	// Command contains the actual command data.
	// The format and meaning of this data depends on the chapter codec.
	Command []byte
}

// ChapterProcess contains processing information for a chapter.
//
// Chapter processes are used to apply special processing to chapters,
// such as codec-specific processing or command execution.
type ChapterProcess struct {
	// CodecID is the identifier for the codec used by this process.
	// This determines how the process data should be interpreted.
	CodecID uint32
	// CodecPrivate contains any private data for this process.
	// This is codec-specific data that may be needed for processing.
	CodecPrivate []byte
	// Commands contains all associated commands for this process.
	// These commands will be executed as part of the chapter processing.
	Commands []ChapterCommand
}

// Chapter contains all information about a Matroska chapter.
//
// A Chapter structure represents a chapter or section within a Matroska file.
// Chapters can be nested to create a hierarchical structure, and can contain
// display information, processing commands, and timing information.
type Chapter struct {
	// UID is the chapter's unique identifier.
	// This allows chapters to be referenced by other elements.
	UID uint64
	// Start is the start time for the chapter in nanoseconds.
	// This is relative to the beginning of the segment.
	Start uint64
	// End is the end time for the chapter in nanoseconds.
	// This is relative to the beginning of the segment.
	End uint64

	// Tracks contains the list of track UIDs this chapter pertains to.
	// If empty, the chapter applies to all tracks.
	Tracks []uint64
	// Display contains display information for this chapter.
	// This allows the chapter to be displayed in multiple languages.
	Display []ChapterDisplay
	// Children contains any child chapters for this chapter.
	// This allows for a hierarchical chapter structure.
	Children []*Chapter
	// Process contains the set of processes for this chapter.
	// These processes can be used for special chapter handling.
	Process []ChapterProcess

	// SegmentUID is the segment UID this chapter relates to.
	// This allows chapters to reference specific segments.
	SegmentUID [16]byte

	// Hidden indicates whether this chapter is hidden.
	// Hidden chapters are not displayed in chapter lists.
	Hidden bool
	// Enabled indicates whether this chapter is enabled.
	// Disabled chapters are ignored during playback.
	Enabled bool

	// Default indicates whether this Edition is the default.
	// If true, this chapter edition should be used by default.
	Default bool
	// Ordered indicates whether this chapter is ordered.
	// Ordered chapters must be played in a specific sequence.
	Ordered bool
}

// Cue contains all information about a Matroska cue.
//
// Cues are indexing points in a Matroska file that allow for efficient seeking.
// A Cue structure contains the timing and position information needed to locate
// specific points in the media stream.
type Cue struct {
	// Time is the cue's start time in nanoseconds.
	// This is the timestamp of the cue point.
	Time uint64
	// Duration is the cue's duration in nanoseconds.
	// This may be 0 if the duration is unknown.
	Duration uint64
	// Position is the cue's position in the stream.
	// This is the byte offset of the cluster containing the cue.
	Position uint64
	// RelativePosition is the cue's position relative to the cluster.
	// This is the byte offset within the cluster.
	RelativePosition uint64
	// Block is the block number within the cluster.
	// This identifies the specific block containing the cue.
	Block uint64
	// Track is the track which this cue covers.
	// This identifies which track the cue point belongs to.
	Track uint8
}

// Target contains information about a tag's target.
//
// A Target structure identifies what a Matroska tag applies to.
// Tags can be applied to tracks, chapters, attachments, or editions.
type Target struct {
	// UID is the target's unique identifier.
	// This identifies the specific element that the tag applies to.
	UID uint64
	// Type is the target type. See the tag target type constants.
	// This determines what kind of element the tag applies to.
	Type uint32
}

// SimpleTag contains a simple Matroska tag.
//
// A SimpleTag structure represents a single key-value metadata tag.
// These tags can be used to store information like title, artist, album, etc.
type SimpleTag struct {
	// Name is the tag name.
	// This is the key part of the key-value pair.
	Name string
	// Value is the tag value.
	// This is the value part of the key-value pair.
	Value string
	// Language is the tag language.
	// This follows the ISO 639-2 language codes (e.g., "eng" for English).
	Language string
	// Default indicates whether this tag is applied by default.
	// If true, this tag should be used unless the user explicitly selects another language.
	Default bool
}

// Tag contains all information relating to a Matroska tag.
//
// A Tag structure represents a collection of metadata tags that can be applied
// to various targets within a Matroska file. Tags are used to store metadata
// like titles, descriptions, and other information about the content.
type Tag struct {
	// Targets is a list of associated targets.
	// This specifies what elements in the file the tags apply to.
	Targets []Target
	// SimpleTags is a list of associated simple tags.
	// These are the actual key-value metadata pairs.
	SimpleTags []SimpleTag
}
