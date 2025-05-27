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

type AttachmentType string

const (
	AttachmentTypeFont     AttachmentType = "font"
	AttachmentTypeSubtitle AttachmentType = "subtitle"
	AttachmentTypeOther    AttachmentType = "other"
)

// TrackInfo holds extracted information about a media track.
type TrackInfo struct {
	Number       int64     `json:"number"`
	UID          int64     `json:"uid"`
	Type         TrackType `json:"type"` // "video", "audio", "subtitle", etc.
	CodecID      string    `json:"codecID"`
	Name         string    `json:"name,omitempty"`
	Language     string    `json:"language,omitempty"`     // Best effort language code
	LanguageIETF string    `json:"languageIETF,omitempty"` // IETF language code
	Default      bool      `json:"default"`
	Forced       bool      `json:"forced"`
	Enabled      bool      `json:"enabled"`
	CodecPrivate string    `json:"codecPrivate,omitempty"` // Raw CodecPrivate data, often used for subtitle headers (e.g., ASS/SSA styles)

	// Video specific
	Video *VideoTrack `json:"video,omitempty"`
	// Audio specific
	Audio *AudioTrack `json:"audio,omitempty"`
	// Internal fields
	contentEncodings *ContentEncodings `json:"-"`
	defaultDuration  uint64            `json:"-"` // in ns
}

// ChapterInfo holds extracted information about a chapter.
type ChapterInfo struct {
	UID           uint64   `json:"uid"`
	Start         float64  `json:"start"`         // Start time in seconds
	End           float64  `json:"end,omitempty"` // End time in seconds
	Text          string   `json:"text,omitempty"`
	Languages     []string `json:"languages,omitempty"`     // Legacy 3-letter language codes
	LanguagesIETF []string `json:"languagesIETF,omitempty"` // IETF language tags
}

// AttachmentInfo holds extracted information about an attachment.
type AttachmentInfo struct {
	UID          uint64         `json:"uid"`
	Filename     string         `json:"filename"`
	Mimetype     string         `json:"mimetype"`
	Size         int            `json:"size"`
	Description  string         `json:"description,omitempty"`
	Type         AttachmentType `json:"type,omitempty"`
	Data         []byte         `json:"-"` // Data loaded into memory
	IsCompressed bool           `json:"-"` // Whether the data is compressed
}

// Metadata holds all extracted metadata.
type Metadata struct {
	Title          string            `json:"title,omitempty"`
	Duration       float64           `json:"duration"`      // Duration in seconds
	TimecodeScale  float64           `json:"timecodeScale"` // Original timecode scale from Info
	MuxingApp      string            `json:"muxingApp,omitempty"`
	WritingApp     string            `json:"writingApp,omitempty"`
	Tracks         []*TrackInfo      `json:"tracks"`
	VideoTracks    []*TrackInfo      `json:"videoTracks"`
	AudioTracks    []*TrackInfo      `json:"audioTracks"`
	SubtitleTracks []*TrackInfo      `json:"subtitleTracks"`
	Chapters       []*ChapterInfo    `json:"chapters"`
	Attachments    []*AttachmentInfo `json:"attachments"`
	MimeCodec      string            `json:"mimeCodec,omitempty"` // RFC 6381 codec string
	Error          error             `json:"-"`
}

func (m *Metadata) GetTrackByNumber(num int64) *TrackInfo {
	for _, track := range m.Tracks {
		if track.Number == num {
			return track
		}
	}
	return nil
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
