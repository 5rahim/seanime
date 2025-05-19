package mkvparser

import (
	"time"

	"github.com/at-wat/ebml-go" // Added for ebml.Block
)

// EBML Header (simplified)
type EBMLHeader struct {
	EBMLVersion        uint64 `ebml:"EBMLVersion"`
	EBMLReadVersion    uint64 `ebml:"EBMLReadVersion"`
	EBMLMaxIDLength    uint64 `ebml:"EBMLMaxIDLength"`
	EBMLMaxSizeLength  uint64 `ebml:"EBMLMaxSizeLength"`
	DocType            string `ebml:"EBMLDocType"`
	DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
	DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
}

// Segment element (top-level, focused on metadata headers)
type Segment struct {
	Info        []Info       `ebml:"Info,size=unknown"`
	Tracks      *Tracks      `ebml:"Tracks,size=unknown"`
	Chapters    *Chapters    `ebml:"Chapters,size=unknown"`
	Attachments *Attachments `ebml:"Attachments,size=unknown"`
	//Cues        *Cues        `ebml:"Cues,size=unknown"`
	Tags    []Tag     `ebml:"Tag,omitempty"`     // Keep Tags as they can contain metadata
	Cluster []Cluster `ebml:"Cluster,omitempty"` // Added for subtitles
}

// MKVRoot represents the top-level structure of an MKV file for parsing.
type MKVRoot struct {
	Header  EBMLHeader `ebml:"EBML"`                 // The EBML Header element
	Segment Segment    `ebml:"Segment,size=unknown"` // The main Segment element
}

// Info element and its children
type Info struct {
	SegmentUID         []byte             `ebml:"SegmentUID,omitempty"`
	SegmentFilename    string             `ebml:"SegmentFilename,omitempty"`
	PrevUID            []byte             `ebml:"PrevUID,omitempty"`
	PrevFilename       string             `ebml:"PrevFilename,omitempty"`
	NextUID            []byte             `ebml:"NextUID,omitempty"`
	NextFilename       string             `ebml:"NextFilename,omitempty"`
	SegmentFamily      []byte             `ebml:"SegmentFamily,omitempty"`
	ChapterTranslate   []ChapterTranslate `ebml:"ChapterTranslate,omitempty"`
	TimecodeScale      uint64             `ebml:"TimecodeScale"` // Default is 1,000,000 ns, handled in logic if missing
	Duration           float64            `ebml:"Duration,omitempty"`
	DateUTC            time.Time          `ebml:"DateUTC,omitempty"`
	Title              string             `ebml:"Title,omitempty"`
	MuxingApp          string             `ebml:"MuxingApp"`
	WritingApp         string             `ebml:"WritingApp"`
	TrackTimecodeScale float64            `ebml:"TrackTimestampScale"` // Corrected tag name. Default 1.0
	MaxBlockAdditionID uint64             `ebml:"MaxBlockAdditionID"`  // Default 0
	Name               string             `ebml:"Name,omitempty"`
	Language           string             `ebml:"Language"` // Default 'eng'
	LanguageIETF       string             `ebml:"LanguageIETF,omitempty"`
	CodecID            string             `ebml:"CodecID"`
	CodecPrivate       []byte             `ebml:"CodecPrivate,omitempty"`
	CodecName          string             `ebml:"CodecName,omitempty"`
	AttachmentLink     uint64             `ebml:"AttachmentLink,omitempty"`
	CodecDecodeAll     uint64             `ebml:"CodecDecodeAll"` // Default 1
	TrackOverlay       uint64             `ebml:"TrackOverlay,omitempty"`
	CodecDelay         uint64             `ebml:"CodecDelay,omitempty"` // Nanoseconds
	SeekPreRoll        uint64             `ebml:"SeekPreRoll"`          // Nanoseconds, Default 0
	TrackTranslate     []TrackTranslate   `ebml:"TrackTranslate,omitempty"`
	Video              *VideoTrack        `ebml:"Video,omitempty"`
	Audio              *AudioTrack        `ebml:"Audio,omitempty"`
	TrackOperation     *TrackOperation    `ebml:"TrackOperation,omitempty"`
	ContentEncodings   *ContentEncodings  `ebml:"ContentEncodings,omitempty"` // Added for subtitle decompression
}

type ChapterTranslate struct {
	ChapterTranslateEditionUID []uint64 `ebml:"ChapterTranslateEditionUID,omitempty"`
	ChapterTranslateCodec      uint64   `ebml:"ChapterTranslateCodec"`
	ChapterTranslateID         []byte   `ebml:"ChapterTranslateID"`
}

// Tracks element and its children
type Tracks struct {
	TrackEntry []TrackEntry `ebml:"TrackEntry"`
}

type TrackEntry struct {
	TrackNumber                 uint64            `ebml:"TrackNumber"`
	TrackUID                    uint64            `ebml:"TrackUID"`
	TrackType                   uint64            `ebml:"TrackType"`
	FlagEnabled                 uint64            `ebml:"FlagEnabled"` // Default 1 handled in logic
	FlagDefault                 uint64            `ebml:"FlagDefault"` // Default 1 handled in logic
	FlagForced                  uint64            `ebml:"FlagForced"`  // Default 0 handled in logic
	FlagLacing                  uint64            `ebml:"FlagLacing"`  // Default 1, deprecated
	MinCache                    uint64            `ebml:"MinCache"`    // Default 0
	MaxCache                    uint64            `ebml:"MaxCache,omitempty"`
	DefaultDuration             uint64            `ebml:"DefaultDuration,omitempty"`             // nanoseconds
	DefaultDecodedFieldDuration uint64            `ebml:"DefaultDecodedFieldDuration,omitempty"` // nanoseconds
	TrackTimecodeScale          float64           `ebml:"TrackTimestampScale"`                   // Default 1.0
	MaxBlockAdditionID          uint64            `ebml:"MaxBlockAdditionID"`                    // Default 0
	Name                        string            `ebml:"Name,omitempty"`
	Language                    string            `ebml:"Language"` // Default 'eng'
	LanguageIETF                string            `ebml:"LanguageIETF,omitempty"`
	CodecID                     string            `ebml:"CodecID"`
	CodecPrivate                []byte            `ebml:"CodecPrivate,omitempty"`
	CodecName                   string            `ebml:"CodecName,omitempty"`
	AttachmentLink              uint64            `ebml:"AttachmentLink,omitempty"`
	CodecDecodeAll              uint64            `ebml:"CodecDecodeAll"` // Default 1
	TrackOverlay                uint64            `ebml:"TrackOverlay,omitempty"`
	CodecDelay                  uint64            `ebml:"CodecDelay,omitempty"` // Nanoseconds
	SeekPreRoll                 uint64            `ebml:"SeekPreRoll"`          // Nanoseconds, Default 0
	TrackTranslate              []TrackTranslate  `ebml:"TrackTranslate,omitempty"`
	Video                       *VideoTrack       `ebml:"Video,omitempty"`
	Audio                       *AudioTrack       `ebml:"Audio,omitempty"`
	TrackOperation              *TrackOperation   `ebml:"TrackOperation,omitempty"`
	ContentEncodings            *ContentEncodings `ebml:"ContentEncodings,omitempty"` // Added for subtitle decompression
}

type VideoTrack struct {
	FlagInterlaced  uint64      `ebml:"FlagInterlaced"`       // Default 0
	StereoMode      uint64      `ebml:"StereoMode,omitempty"` // Deprecated: Use StereoMode in BlockAdditions
	AlphaMode       uint64      `ebml:"AlphaMode"`            // Default 0
	PixelWidth      uint64      `ebml:"PixelWidth"`
	PixelHeight     uint64      `ebml:"PixelHeight"`
	PixelCropBottom uint64      `ebml:"PixelCropBottom"`         // Default 0
	PixelCropTop    uint64      `ebml:"PixelCropTop"`            // Default 0
	PixelCropLeft   uint64      `ebml:"PixelCropLeft"`           // Default 0
	PixelCropRight  uint64      `ebml:"PixelCropRight"`          // Default 0
	DisplayWidth    uint64      `ebml:"DisplayWidth,omitempty"`  // In pixels
	DisplayHeight   uint64      `ebml:"DisplayHeight,omitempty"` // In pixels
	DisplayUnit     uint64      `ebml:"DisplayUnit"`             // Default 0
	AspectRatioType uint64      `ebml:"AspectRatioType"`         // Default 0
	ColourSpace     []byte      `ebml:"ColourSpace,omitempty"`   // Deprecated
	Colour          *Color      `ebml:"Colour,omitempty"`
	Projection      *Projection `ebml:"Projection,omitempty"`
}

type Color struct {
	MatrixCoefficients      uint64             `ebml:"MatrixCoefficients,omitempty"` // Deprecated
	BitsPerChannel          uint64             `ebml:"BitsPerChannel,omitempty"`     // Deprecated
	ChromaSubsamplingHorz   uint64             `ebml:"ChromaSubsamplingHorz,omitempty"`
	ChromaSubsamplingVert   uint64             `ebml:"ChromaSubsamplingVert,omitempty"`
	CbSubsamplingHorz       uint64             `ebml:"CbSubsamplingHorz,omitempty"`
	CbSubsamplingVert       uint64             `ebml:"CbSubsamplingVert,omitempty"`
	ChromaSitingHorz        uint64             `ebml:"ChromaSitingHorz,omitempty"`        // Deprecated
	ChromaSitingVert        uint64             `ebml:"ChromaSitingVert,omitempty"`        // Deprecated
	Range                   uint64             `ebml:"Range,omitempty"`                   // Deprecated
	TransferCharacteristics uint64             `ebml:"TransferCharacteristics,omitempty"` // Deprecated
	Primaries               uint64             `ebml:"Primaries,omitempty"`               // Deprecated
	MaxCLL                  uint64             `ebml:"MaxCLL,omitempty"`
	MaxFALL                 uint64             `ebml:"MaxFALL,omitempty"`
	MasteringMetadata       *MasteringMetadata `ebml:"MasteringMetadata,omitempty"`
}

type MasteringMetadata struct {
	PrimaryRChromaticityX   float64 `ebml:"PrimaryRChromaticityX,omitempty"`
	PrimaryRChromaticityY   float64 `ebml:"PrimaryRChromaticityY,omitempty"`
	PrimaryGChromaticityX   float64 `ebml:"PrimaryGChromaticityX,omitempty"`
	PrimaryGChromaticityY   float64 `ebml:"PrimaryGChromaticityY,omitempty"`
	PrimaryBChromaticityX   float64 `ebml:"PrimaryBChromaticityX,omitempty"`
	PrimaryBChromaticityY   float64 `ebml:"PrimaryBChromaticityY,omitempty"`
	WhitePointChromaticityX float64 `ebml:"WhitePointChromaticityX,omitempty"`
	WhitePointChromaticityY float64 `ebml:"WhitePointChromaticityY,omitempty"`
	LuminanceMax            float64 `ebml:"LuminanceMax,omitempty"`
	LuminanceMin            float64 `ebml:"LuminanceMin,omitempty"`
}

type Projection struct {
	ProjectionType      uint64  `ebml:"ProjectionType"` // Default 0
	ProjectionPrivate   []byte  `ebml:"ProjectionPrivate,omitempty"`
	ProjectionPoseYaw   float64 `ebml:"ProjectionPoseYaw"`   // Default 0.0
	ProjectionPosePitch float64 `ebml:"ProjectionPosePitch"` // Default 0.0
	ProjectionPoseRoll  float64 `ebml:"ProjectionPoseRoll"`  // Default 0.0
}

type AudioTrack struct {
	SamplingFrequency       float64 `ebml:"SamplingFrequency"` // Default 8000.0
	OutputSamplingFrequency float64 `ebml:"OutputSamplingFrequency,omitempty"`
	Channels                uint64  `ebml:"Channels"` // Default 1
	BitDepth                uint64  `ebml:"BitDepth,omitempty"`
}

type TrackTranslate struct {
	TrackTranslateEditionUID []uint64 `ebml:"TrackTranslateEditionUID,omitempty"`
	TrackTranslateCodec      uint64   `ebml:"TrackTranslateCodec"`
	TrackTranslateTrackID    []byte   `ebml:"TrackTranslateTrackID"`
}

type TrackOperation struct {
	TrackCombinePlanes *TrackCombinePlanes `ebml:"TrackCombinePlanes,omitempty"`
	TrackJoinBlocks    *TrackJoinBlocks    `ebml:"TrackJoinBlocks,omitempty"`
}

type TrackCombinePlanes struct {
	TrackPlane []TrackPlane `ebml:"TrackPlane"`
}

type TrackPlane struct {
	TrackPlaneUID  uint64 `ebml:"TrackPlaneUID"`
	TrackPlaneType uint64 `ebml:"TrackPlaneType"`
}

type TrackJoinBlocks struct {
	TrackJoinUID []uint64 `ebml:"TrackJoinUID"`
}

type ContentEncodings struct {
	ContentEncoding []ContentEncoding `ebml:"ContentEncoding"`
}

type ContentEncoding struct {
	ContentEncodingOrder uint64              `ebml:"ContentEncodingOrder"` // Default 0
	ContentEncodingScope uint64              `ebml:"ContentEncodingScope"` // Default 1
	ContentEncodingType  uint64              `ebml:"ContentEncodingType"`  // Default 0
	ContentCompression   *ContentCompression `ebml:"ContentCompression,omitempty"`
	ContentEncryption    *ContentEncryption  `ebml:"ContentEncryption,omitempty"`
}

type ContentCompression struct {
	ContentCompAlgo     uint64 `ebml:"ContentCompAlgo"` // Default 0
	ContentCompSettings []byte `ebml:"ContentCompSettings,omitempty"`
}

type ContentEncryption struct {
	ContentEncAlgo        uint64                 `ebml:"ContentEncAlgo,omitempty"`        // Default 0, deprecated
	ContentEncKeyID       []byte                 `ebml:"ContentEncKeyID,omitempty"`       // Deprecated
	ContentEncAESSettings *ContentEncAESSettings `ebml:"ContentEncAESSettings,omitempty"` // Deprecated
	ContentSignature      []byte                 `ebml:"ContentSignature,omitempty"`      // Deprecated
	ContentSigKeyID       []byte                 `ebml:"ContentSigKeyID,omitempty"`       // Deprecated
	ContentSigAlgo        uint64                 `ebml:"ContentSigAlgo,omitempty"`        // Deprecated
	ContentSigHashAlgo    uint64                 `ebml:"ContentSigHashAlgo,omitempty"`    // Deprecated
}

type ContentEncAESSettings struct { // Deprecated
	AESSettingsCipherMode uint64 `ebml:"AESSettingsCipherMode,omitempty"`
}

// Chapters element and its children
type Chapters struct {
	EditionEntry []EditionEntry `ebml:"EditionEntry"`
}

type EditionEntry struct {
	EditionUID         uint64        `ebml:"EditionUID,omitempty"`
	EditionFlagHidden  uint64        `ebml:"EditionFlagHidden"`  // Default 0
	EditionFlagDefault uint64        `ebml:"EditionFlagDefault"` // Default 0
	EditionFlagOrdered uint64        `ebml:"EditionFlagOrdered"` // Default 0
	ChapterAtom        []ChapterAtom `ebml:"ChapterAtom"`
}

type ChapterAtom struct {
	ChapterUID               uint64           `ebml:"ChapterUID"`
	ChapterStringUID         string           `ebml:"ChapterStringUID,omitempty"`
	ChapterTimeStart         uint64           `ebml:"ChapterTimeStart"`
	ChapterTimeEnd           uint64           `ebml:"ChapterTimeEnd,omitempty"`
	ChapterFlagHidden        uint64           `ebml:"ChapterFlagHidden"`  // Default 0
	ChapterFlagEnabled       uint64           `ebml:"ChapterFlagEnabled"` // Default 1
	ChapterSegmentUID        []byte           `ebml:"ChapterSegmentUID,omitempty"`
	ChapterSegmentEditionUID uint64           `ebml:"ChapterSegmentEditionUID,omitempty"`
	ChapterPhysicalEquiv     uint64           `ebml:"ChapterPhysicalEquiv,omitempty"`
	ChapterTrack             *ChapterTrack    `ebml:"ChapterTrack,omitempty"`
	ChapterDisplay           []ChapterDisplay `ebml:"ChapterDisplay,omitempty"`
	ChapProcess              []ChapProcess    `ebml:"ChapProcess,omitempty"`
}

type ChapterTrack struct {
	ChapterTrackUID uint64 `ebml:"ChapterTrackUID"` // Corrected name and type
}

type ChapterDisplay struct {
	ChapString       string   `ebml:"ChapString"`
	ChapLanguage     []string `ebml:"ChapLanguage"` // Default 'eng'
	ChapLanguageIETF []string `ebml:"ChapLanguageIETF,omitempty"`
	ChapCountry      []string `ebml:"ChapCountry,omitempty"`
}

type ChapProcess struct {
	ChapProcessCodecID uint64               `ebml:"ChapProcessCodecID"`
	ChapProcessPrivate []byte               `ebml:"ChapProcessPrivate,omitempty"`
	ChapProcessCommand []ChapProcessCommand `ebml:"ChapProcessCommand,omitempty"`
}

type ChapProcessCommand struct {
	ChapProcessTime uint64 `ebml:"ChapProcessTime"`
	ChapProcessData []byte `ebml:"ChapProcessData"`
}

// Attachments element and its children
type Attachments struct {
	AttachedFile []AttachedFile `ebml:"AttachedFile"`
}

type AttachedFile struct {
	FileDescription string `ebml:"FileDescription,omitempty"`
	FileName        string `ebml:"FileName"`
	FileMimeType    string `ebml:"FileMimeType"`
	FileData        []byte `ebml:"FileData"`
	FileUID         uint64 `ebml:"FileUID"`
}

// Tags and Tag elements
type Tag struct {
	Targets   *Targets    `ebml:"Targets"`
	SimpleTag []SimpleTag `ebml:"SimpleTag"`
}

type Targets struct {
	TargetTypeValue  uint64 `ebml:"TargetTypeValue,omitempty"` // Default 50
	TargetType       string `ebml:"TargetType,omitempty"`
	TagTrackUID      uint64 `ebml:"TagTrackUID,omitempty"`
	TagEditionUID    uint64 `ebml:"TagEditionUID,omitempty"`
	TagChapterUID    uint64 `ebml:"TagChapterUID,omitempty"`
	TagAttachmentUID uint64 `ebml:"TagAttachmentUID,omitempty"`
}

type SimpleTag struct {
	TagName         string      `ebml:"TagName"`
	TagLanguage     string      `ebml:"TagLanguage"` // Default 'und'
	TagLanguageIETF string      `ebml:"TagLanguageIETF,omitempty"`
	TagDefault      uint64      `ebml:"TagDefault"` // Default 1
	TagString       string      `ebml:"TagString,omitempty"`
	TagBinary       []byte      `ebml:"TagBinary,omitempty"`
	SimpleTag       []SimpleTag `ebml:"SimpleTag,omitempty"` // Nested tags
}

// Cluster element and its children
type Cluster struct {
	Timecode    uint64       `ebml:"Timecode"`              // Cluster's absolute timecode in TimecodeScale units
	SimpleBlock []ebml.Block `ebml:"SimpleBlock,omitempty"` // Use ebml.Block for SimpleBlock parsing
	BlockGroup  []BlockGroup `ebml:"BlockGroup,omitempty"`
	// Other elements like SilentTracks, Position, PrevSize can be added if needed
}

// BlockGroup element
type BlockGroup struct {
	Block         ebml.Block `ebml:"Block"`                   // Parsed Block data
	BlockDuration uint64     `ebml:"BlockDuration,omitempty"` // Duration in TimecodeScale units
	// ReferenceBlock can be added if needed for B-frames, etc.
	// Other elements like Slices, CodecState, etc., can be added if needed
}

// CuePoint struct
type CuePoint struct {
	Time      uint64             `ebml:"CueTime"`
	Positions []CueTrackPosition `ebml:"CueTrackPositions"`
}

// CueTrackPosition struct
type CueTrackPosition struct {
	Track           uint64 `ebml:"CueTrack"`
	ClusterPosition uint64 `ebml:"CueClusterPosition"`
	BlockNumber     uint64 `ebml:"CueBlockNumber,omitempty"`
}

// Cues struct
type Cues struct {
	CuePoints []CuePoint `ebml:"CuePoint"`
}

// IDs for common Matroska elements (from ebml-go/matroska/ids.go and spec)
const (
	// Segment Information
	idInfo          = 0x1549a966
	idTimecodeScale = 0x2ad7b1
	idDuration      = 0x4489
	idDateUTC       = 0x4461
	idTitle         = 0x7ba9
	idMuxingApp     = 0x4d80
	idWritingApp    = 0x5741

	// Tracks
	idTracks            = 0x1654ae6b
	idTrackEntry        = 0xae
	idTrackNumber       = 0xd7
	idTrackUID          = 0x73c5
	idTrackType         = 0x83
	idFlagEnabled       = 0xb9
	idFlagDefault       = 0x88
	idFlagForced        = 0x55aa
	idCodecID           = 0x86
	idCodecPrivate      = 0x63a2
	idCodecName         = 0x258688
	idName              = 0x536e
	idLanguage          = 0x22b59c // Deprecated, use LanguageIETF
	idLanguageIETF      = 0x22b59d
	idVideo             = 0xe0
	idPixelWidth        = 0xb0
	idPixelHeight       = 0xba
	idAudio             = 0xe1
	idSamplingFrequency = 0xb5
	idChannels          = 0x9f
	idBitDepth          = 0x6264

	// Chapters
	idChapters         = 0x1043a770
	idEditionEntry     = 0x45b9
	idChapterAtom      = 0xb6
	idChapterUID       = 0x73c4
	idChapterTimeStart = 0x91
	idChapterTimeEnd   = 0x92
	idChapterDisplay   = 0x80
	idChapString       = 0x85
	idChapLanguage     = 0x437c // Deprecated, use ChapLanguageIETF
	idChapLanguageIETF = 0x437d

	// Attachments
	idAttachments  = 0x1941a469
	idAttachedFile = 0x61a7
	idFileName     = 0x466e
	idFileMimeType = 0x4660
	idFileData     = 0x465c
	idFileUID      = 0x46ae

	// Cluster (ID might be needed for other logic even if not parsed)
	idCluster       = 0x1f43b675
	idTimecode      = 0xe7 // Element ID for Timecode within Cluster
	idBlockGroup    = 0xa0 // Element ID for BlockGroup
	idBlock         = 0xa1 // Element ID for Block within BlockGroup
	idBlockDuration = 0x9b // Element ID for BlockDuration within BlockGroup
	idSimpleBlock   = 0xa3 // Element ID for SimpleBlock within Cluster

	// Content Encoding
	idContentEncodings     = 0x6d80
	idContentEncoding      = 0x6240
	idContentEncodingOrder = 0x5031
	idContentEncodingScope = 0x5032
	idContentEncodingType  = 0x5033
	idContentCompression   = 0x5034
	idContentCompAlgo      = 0x4254
	idContentCompSettings  = 0x4255
	idContentEncryption    = 0x5035 // Note: ContentEncryption and its children are complex and often not needed for basic subtitle text extraction
	// idContentEncAlgo (0x47e1) and other encryption related IDs omitted for brevity unless specifically needed
)
