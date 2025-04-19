package vendor_habari

type Metadata struct {
	SeasonNumber        []string `json:"season_number,omitempty"`
	PartNumber          []string `json:"part_number,omitempty"`
	Title               string   `json:"title,omitempty"`
	FormattedTitle      string   `json:"formatted_title,omitempty"`
	AnimeType           []string `json:"anime_type,omitempty"`
	Year                string   `json:"year,omitempty"`
	AudioTerm           []string `json:"audio_term,omitempty"`
	DeviceCompatibility []string `json:"device_compatibility,omitempty"`
	EpisodeNumber       []string `json:"episode_number,omitempty"`
	OtherEpisodeNumber  []string `json:"other_episode_number,omitempty"`
	EpisodeNumberAlt    []string `json:"episode_number_alt,omitempty"`
	EpisodeTitle        string   `json:"episode_title,omitempty"`
	FileChecksum        string   `json:"file_checksum,omitempty"`
	FileExtension       string   `json:"file_extension,omitempty"`
	FileName            string   `json:"file_name,omitempty"`
	Language            []string `json:"language,omitempty"`
	ReleaseGroup        string   `json:"release_group,omitempty"`
	ReleaseInformation  []string `json:"release_information,omitempty"`
	ReleaseVersion      []string `json:"release_version,omitempty"`
	Source              []string `json:"source,omitempty"`
	Subtitles           []string `json:"subtitles,omitempty"`
	VideoResolution     string   `json:"video_resolution,omitempty"`
	VideoTerm           []string `json:"video_term,omitempty"`
	VolumeNumber        []string `json:"volume_number,omitempty"`
}