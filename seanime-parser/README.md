<p align="center">
<img src="../docs/images/logo.png" alt="preview" width="75px"/>
</p>

<h2 align="center"><b>Seanime Parser</b></h2>

<h4 align="center">Smarter anime video filename parser written in Go.</h4>

# Features

- Fixes some of the issues with Anitomy-based parsers by using new parsing algorithms.
- Support for part numbers.
- Better parsing of keywords.

 _                                                                                       | Seanime Parser | Anitomy | Anitogo 
:----------------------------------------------------------------------------------------|:---------------|:--------|:--------
 Standard filename formatting                                                            | ✅              | ✅       | ✅       
 [Mobile Suit Gundam Seed Destiny HD REMASTER][07][Big5][720p][AVC_AAC][encoded by SEED] | ✅              | ❌       | ❌       
 [ReDone] Memories Off 3.5 - 04 (DVD 10-bit)                                             | ✅              | ❌       | ❌       
 Byousoku 5 Centimeter [Blu-Ray][1920x1080 H.264][2.0ch AAC][SOFTSUBS]                   | ✅              | ❌       | ❌       
 37 [Ruberia]_Death_Note_-_37v2_[FINAL]_[XviD][6FA7D273]                                 | ✅              | ❌       | ❌       
 [Judas] Aharen-san wa Hakarenai - S01E06v2.mkv                                          | ✅              | ✅       | ❌       
 S01E05 - Episode title.mkv                                                              | ✅              | ❔       | ❌       

# Use

```go
package main

import (
	parser "github.com/5rahim/seanime/seanime-parser"
)

func main() {
	data := parser.Parse("Hyouka (2012) S1-2 [BD 1080p HEVC OPUS] [Dual-Audio]")
	println(data.Title)           // Hyouka
	println(data.FormattedTitle)  // Hyouka (2012)
	println(data.Year)            // 2012
	println(data.SeasonNumber)    // []string{"1", "2"}
	println(data.VideoResolution) // 1080p
}
```

```go
package main

type Metadata struct {
	Title               string   `json:"title,omitempty"`
	FormattedTitle      string   `json:"formatted_title,omitempty"`
	SeasonNumber        []string `json:"season_number,omitempty"`
	PartNumber          []string `json:"part_number,omitempty"`
	VolumeNumber        []string `json:"volume_number,omitempty"`
	EpisodeNumber       []string `json:"episode_number,omitempty"`
	EpisodeNumberAlt    []string `json:"episode_number_alt,omitempty"`
	OtherEpisodeNumber  []string `json:"other_episode_number,omitempty"`
	AnimeType           []string `json:"anime_type,omitempty"`
	Year                string   `json:"year,omitempty"`
	AudioTerm           []string `json:"audio_term,omitempty"`
	DeviceCompatibility []string `json:"device_compatibility,omitempty"`
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
}

```

# Examples

```json
{
  "file_name": "Howl's_Moving_Castle_(2004)_[1080p,BluRay,flac,dts,x264]_-_THORA v2.mkv",
  "title": "Howl's Moving Castle",
  "formatted_title": "Howl's Moving Castle (2004)",
  "year": "2004",
  "audio_term": [
    "flac",
    "dts"
  ],
  "file_extension": "mkv",
  "release_group": "THORA",
  "release_version": [
    "2"
  ],
  "source": [
    "BluRay"
  ],
  "video_resolution": "1080p",
  "video_term": [
    "x264"
  ]
}
```

```json
{
  "file_name": "[TV-J] Kidou Senshi Gundam UC Unicorn - episode.02 [BD 1920x1080 h264+AAC(5.1ch JP+EN) +Sub(JP-EN-SP-FR-CH) Chap].mp4",
  "title": "Kidou Senshi Gundam UC Unicorn",
  "formatted_title": "Kidou Senshi Gundam UC Unicorn",
  "episode_number": [
    "02"
  ],
  "file_extension": "mp4",
  "release_group": "TV-J",
  "source": [
    "BD"
  ],
  "audio_term": [
    "AAC",
    "5.1ch"
  ],
  "video_resolution": "1920x1080",
  "video_term": [
    "h264"
  ],
  "subtitles": [
    "Sub"
  ],
  "language": [
    "JP",
    "EN",
    "JP",
    "EN",
    "FR",
    "CH"
  ]
}
```

## Acknowledgements

- [Anitogo](https://github.com/nssteinbrenner/anitogo) - Inspiration for this project
