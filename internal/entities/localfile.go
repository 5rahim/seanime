package entities

import (
	"github.com/seanime-app/seanime/internal/filesystem"
	seanime_parser "github.com/seanime-app/seanime/seanime-parser"
)

const (
	LocalFileTypeMain    LocalFileType = "main"
	LocalFileTypeSpecial LocalFileType = "special"
	LocalFileTypeNC      LocalFileType = "nc"
)

type (
	LocalFileType string
	LocalFile     struct {
		Path             string                 `json:"path"`
		Name             string                 `json:"name"`
		ParsedData       *LocalFileParsedData   `json:"parsedInfo"`
		ParsedFolderData []*LocalFileParsedData `json:"parsedFolderInfo"`
		Metadata         *LocalFileMetadata     `json:"metadata"`
		Locked           bool                   `json:"locked"`
		Ignored          bool                   `json:"ignored"`
		MediaId          int                    `json:"mediaId"`
	}

	LocalFileMetadata struct {
		Episode      int           `json:"episode"`
		AniDBEpisode string        `json:"aniDBEpisode"`
		Type         LocalFileType `json:"type"`
	}

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

// NewLocalFile creates and returns a reference to a new LocalFile struct from a path
func NewLocalFile(opath, dirPath string) *LocalFile {

	info := filesystem.SeparateFilePath(opath, dirPath)

	// Parse filename
	fElements := seanime_parser.Parse(info.Filename)
	parsedInfo := newLocalFileParsedData(info.Filename, fElements)

	// Parse dirnames
	parsedFolderInfo := make([]*LocalFileParsedData, 0)
	for _, dirname := range info.Dirnames {
		if len(dirname) > 0 {
			pElements := seanime_parser.Parse(dirname)
			parsed := newLocalFileParsedData(dirname, pElements)
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

// newLocalFileParsedData Converts tanuki.Elements into LocalFileParsedData.
//
// This is used by NewLocalFile
func newLocalFileParsedData(original string, elements *seanime_parser.Metadata) *LocalFileParsedData {
	i := new(LocalFileParsedData)
	i.Original = original
	i.Title = elements.Title
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
