package mkvparser

import (
	"time"
)

// Info element and its children
type Info struct {
	Title         string
	MuxingApp     string
	WritingApp    string
	TimecodeScale uint64
	Duration      float64
	DateUTC       time.Time
}

// TrackEntry represents a track in the MKV file
type TrackEntry struct {
	TrackNumber      uint64
	TrackUID         uint64
	TrackType        uint64
	FlagEnabled      uint64
	FlagDefault      uint64
	FlagForced       uint64
	DefaultDuration  uint64
	Name             string
	Language         string
	LanguageIETF     string
	CodecID          string
	CodecPrivate     []byte
	Video            *VideoTrack
	Audio            *AudioTrack
	ContentEncodings *ContentEncodings
}

// VideoTrack contains video-specific track data
type VideoTrack struct {
	PixelWidth  uint64
	PixelHeight uint64
}

// AudioTrack contains audio-specific track data
type AudioTrack struct {
	SamplingFrequency float64
	Channels          uint64
	BitDepth          uint64
}

// ContentEncodings contains information about how the track data is encoded
type ContentEncodings struct {
	ContentEncoding []ContentEncoding
}

// ContentEncoding describes a single encoding applied to the track data
type ContentEncoding struct {
	ContentEncodingOrder uint64
	ContentEncodingScope uint64
	ContentEncodingType  uint64
	ContentCompression   *ContentCompression
}

// ContentCompression describes how the track data is compressed
type ContentCompression struct {
	ContentCompAlgo     uint64
	ContentCompSettings []byte
}

// ChapterAtom represents a single chapter point
type ChapterAtom struct {
	ChapterUID       uint64
	ChapterTimeStart uint64
	ChapterTimeEnd   uint64
	ChapterDisplay   []ChapterDisplay
}

// ChapterDisplay contains displayable chapter information
type ChapterDisplay struct {
	ChapString       string
	ChapLanguage     []string
	ChapLanguageIETF []string
}

// AttachedFile represents a file attached to the MKV container
type AttachedFile struct {
	FileDescription string
	FileName        string
	FileMimeType    string
	FileData        []byte
	FileUID         uint64
}

// Block represents a data block in the MKV file
type Block struct {
	TrackNumber uint64
	Timecode    int16
	Data        [][]byte
}

// BlockGroup represents a group of blocks with additional information
type BlockGroup struct {
	Block         Block
	BlockDuration uint64
}

// Cluster represents a cluster of blocks in the MKV file
type Cluster struct {
	Timecode    uint64
	SimpleBlock []Block
	BlockGroup  []BlockGroup
}

// Tracks element and its children
type Tracks struct {
	TrackEntry []TrackEntry `ebml:"TrackEntry"`
}
