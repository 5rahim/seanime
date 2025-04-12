package anime

import (
	"seanime/internal/library/filesystem"

	"github.com/5rahim/habari"
)

const (
	LocalFileTypeMain    LocalFileType = "main"    // Main episodes that are trackable
	LocalFileTypeSpecial LocalFileType = "special" // OVA, ONA, etc.
	LocalFileTypeNC      LocalFileType = "nc"      // Opening, ending, etc.
)

type (
	LocalFileType string
	// LocalFile represents a media file on the local filesystem.
	// It is used to store information about and state of the file, such as its path, name, and parsed data.
	LocalFile struct {
		Path             string                 `json:"path"`
		Name             string                 `json:"name"`
		ParsedData       *LocalFileParsedData   `json:"parsedInfo"`
		ParsedFolderData []*LocalFileParsedData `json:"parsedFolderInfo"`
		Metadata         *LocalFileMetadata     `json:"metadata"`
		Locked           bool                   `json:"locked"`
		Ignored          bool                   `json:"ignored"` // Unused for now
		MediaId          int                    `json:"mediaId"`
	}

	// LocalFileMetadata holds metadata related to a media episode.
	LocalFileMetadata struct {
		Episode      int           `json:"episode"`
		AniDBEpisode string        `json:"aniDBEpisode"`
		Type         LocalFileType `json:"type"`
	}

	// LocalFileParsedData holds parsed data from a media file's name.
	// This data is used to identify the media file during the scanning process.
	LocalFileParsedData struct {
		Original     string   `json:"original"`
		Title        string   `json:"title,omitempty"`
		ReleaseGroup string   `json:"releaseGroup,omitempty"`
		Season       string   `json:"season,omitempty"`
		SeasonRange  []string `json:"seasonRange,omitempty"`
		Part         string   `json:"part,omitempty"`
		PartRange    []string `json:"partRange,omitempty"`
		Episode      string   `json:"episode,omitempty"`
		EpisodeRange []string `json:"episodeRange,omitempty"`
		EpisodeTitle string   `json:"episodeTitle,omitempty"`
		Year         string   `json:"year,omitempty"`
	}
)

// NewLocalFileS creates and returns a reference to a new LocalFile struct.
// It will parse the file's name and its directory names to extract necessary information.
//   - opath: The full path to the file.
//   - dirPaths: The full paths to the directories that may contain the file. (Library root paths)
func NewLocalFileS(opath string, dirPaths []string) *LocalFile {
	info := filesystem.SeparateFilePathS(opath, dirPaths)
	return newLocalFile(opath, info)
}

// NewLocalFile creates and returns a reference to a new LocalFile struct.
// It will parse the file's name and its directory names to extract necessary information.
//   - opath: The full path to the file.
//   - dirPath: The full path to the directory containing the file. (The library root path)
func NewLocalFile(opath, dirPath string) *LocalFile {
	info := filesystem.SeparateFilePath(opath, dirPath)
	return newLocalFile(opath, info)
}

func newLocalFile(opath string, info *filesystem.SeparatedFilePath) *LocalFile {
	// Parse filename
	fElements := habari.Parse(info.Filename)
	parsedInfo := NewLocalFileParsedData(info.Filename, fElements)

	// Parse dir names
	parsedFolderInfo := make([]*LocalFileParsedData, 0)
	for _, dirname := range info.Dirnames {
		if len(dirname) > 0 {
			pElements := habari.Parse(dirname)
			parsed := NewLocalFileParsedData(dirname, pElements)
			parsedFolderInfo = append(parsedFolderInfo, parsed)
		}
	}

	localFile := &LocalFile{
		Path:             opath,
		Name:             info.Filename,
		ParsedData:       parsedInfo,
		ParsedFolderData: parsedFolderInfo,
		Metadata: &LocalFileMetadata{
			Episode:      0,
			AniDBEpisode: "",
			Type:         "",
		},
		Locked:  false,
		Ignored: false,
		MediaId: 0,
	}

	return localFile
}

// NewLocalFileParsedData Converts habari.Metadata into LocalFileParsedData, which is more suitable.
func NewLocalFileParsedData(original string, elements *habari.Metadata) *LocalFileParsedData {
	i := new(LocalFileParsedData)
	i.Original = original
	i.Title = elements.FormattedTitle
	i.ReleaseGroup = elements.ReleaseGroup
	i.EpisodeTitle = elements.EpisodeTitle
	i.Year = elements.Year

	if len(elements.SeasonNumber) > 0 {
		if len(elements.SeasonNumber) == 1 {
			i.Season = elements.SeasonNumber[0]
		} else {
			i.SeasonRange = elements.SeasonNumber
		}
	}

	if len(elements.EpisodeNumber) > 0 {
		if len(elements.EpisodeNumber) == 1 {
			i.Episode = elements.EpisodeNumber[0]
		} else {
			i.EpisodeRange = elements.EpisodeNumber
		}
	}

	if len(elements.PartNumber) > 0 {
		if len(elements.PartNumber) == 1 {
			i.Part = elements.PartNumber[0]
		} else {
			i.PartRange = elements.PartNumber
		}
	}

	return i
}
