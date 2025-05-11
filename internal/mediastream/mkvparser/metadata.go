package mkvparser

// TrackType represents the type of a Matroska track.
type TrackType string

const (
	TrackTypeVideo    TrackType = "video"
	TrackTypeAudio    TrackType = "audio"
	TrackTypeSubtitle TrackType = "subtitle"
	TrackTypeLogo     TrackType = "logo"
	TrackTypeButtons  TrackType = "buttons"
	TrackTypeComplex  TrackType = "complex"
	TrackTypeUnknown  TrackType = "unknown"
)

// TrackInfo holds extracted information about a media track.
type TrackInfo struct {
	Number       uint64    `json:"number"`
	UID          uint64    `json:"uid"`
	Type         TrackType `json:"type"` // "video", "audio", "subtitle", etc.
	CodecID      string    `json:"codecID"`
	Name         string    `json:"name,omitempty"`
	Language     string    `json:"language,omitempty"` // Best effort language code (IETF or 3-letter)
	Default      bool      `json:"default"`
	Forced       bool      `json:"forced"`
	Enabled      bool      `json:"enabled"`
	CodecPrivate string    `json:"codecPrivate,omitempty"` // Raw CodecPrivate data, often used for subtitle headers (e.g., ASS/SSA styles)

	// Video specific
	PixelWidth  uint64 `json:"pixelWidth,omitempty"`
	PixelHeight uint64 `json:"pixelHeight,omitempty"`

	// Audio specific
	SamplingFrequency float64 `json:"samplingFrequency,omitempty"`
	Channels          uint64  `json:"channels,omitempty"`
	BitDepth          uint64  `json:"bitDepth,omitempty"`

	// Internal fields
	contentEncodings *ContentEncodings `json:"-"`
	defaultDuration  uint64            `json:"-"` // in ns
}

// ChapterInfo holds extracted information about a chapter.
type ChapterInfo struct {
	UID   uint64  `json:"uid"`
	Start float64 `json:"start"`         // Start time in seconds
	End   float64 `json:"end,omitempty"` // End time in seconds
	Text  string  `json:"text,omitempty"`
}

// AttachmentInfo holds extracted information about an attachment.
type AttachmentInfo struct {
	UID      uint64 `json:"uid"`
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
	Size     int    `json:"size"`
	Data     []byte `json:"-"` // Data loaded into memory
}

// Metadata holds all extracted metadata.
type Metadata struct {
	Title         string            `json:"title,omitempty"`
	Duration      float64           `json:"duration"`      // Duration in seconds
	TimecodeScale float64           `json:"timecodeScale"` // Original timecode scale from Info
	MuxingApp     string            `json:"muxingApp,omitempty"`
	WritingApp    string            `json:"writingApp,omitempty"`
	Tracks        []*TrackInfo      `json:"tracks"`
	Chapters      []*ChapterInfo    `json:"chapters"`
	Attachments   []*AttachmentInfo `json:"attachments"`
	Error         error             `json:"error,omitempty"`
}

func (m *Metadata) GetTrackByNumber(num uint64) *TrackInfo {
	for _, track := range m.Tracks {
		if track.Number == num {
			return track
		}
	}
	return nil
}

func (m *Metadata) GetSubtitleTracks() (ret []*TrackInfo) {
	for _, track := range m.Tracks {
		if track.Type == TrackTypeSubtitle {
			ret = append(ret, track)
		}
	}
	return
}

func (m *Metadata) GetAudioTracks() (ret []*TrackInfo) {
	for _, track := range m.Tracks {
		if track.Type == TrackTypeAudio {
			ret = append(ret, track)
		}
	}
	return
}

func (m *Metadata) GetVideoTracks() (ret []*TrackInfo) {
	for _, track := range m.Tracks {
		if track.Type == TrackTypeVideo {
			ret = append(ret, track)
		}
	}
	return
}

func (m *Metadata) GetAttachmentByName(name string) (*AttachmentInfo, bool) {
	for _, attachment := range m.Attachments {
		if attachment.Filename == name {
			return attachment, true
		}
	}
	return nil, false
}

///////////////////////////////////////////////////////////////////////////////////////////

func (t *TrackInfo) IsAudioTrack() bool {
	return t.Type == TrackTypeAudio
}

func (t *TrackInfo) IsVideoTrack() bool {
	return t.Type == TrackTypeVideo
}

func (t *TrackInfo) IsSubtitleTrack() bool {
	return t.Type == TrackTypeSubtitle
}
