package entities

import (
	"github.com/5rahim/tanuki"
	"github.com/seanime-app/seanime/internal/filesystem"
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
		Original     string   `json:"original"`               // Same as LocalFile.Name for LocalFile.ParsedData
		Title        string   `json:"title,omitempty"`        // Same as tanuki.Elements.AnimeTitle
		ReleaseGroup string   `json:"releaseGroup,omitempty"` // Same as tanuki.Elements.ReleaseGroup
		Season       string   `json:"season,omitempty"`       // First element of tanuki.Elements.AnimeSeason if not a range
		SeasonRange  []string `json:"seasonRange,omitempty"`  // Same as tanuki.Elements.AnimeSeason if range
		Part         string   `json:"part,omitempty"`         // First element of tanuki.Elements.AnimePart if not a range
		PartRange    []string `json:"partRange,omitempty"`    // Same tanuki.Elements.AnimePart if range
		Episode      string   `json:"episode,omitempty"`      // First element of tanuki.Elemenets.EpisodeNumber
		EpisodeRange []string `json:"episodeRange,omitempty"` // Same as tanuki.Elemenets.EpisodeNumber if range
		EpisodeTitle string   `json:"episodeTitle,omitempty"` // Same as tanuki.Elemenets.EpisodeTitle
		Year         string   `json:"year,omitempty"`         // Same as tanuki.Elemenets.AnimeYear
	}
)

// NewLocalFile creates and returns a reference to a new LocalFile struct from a path
func NewLocalFile(opath, dirPath string) *LocalFile {

	info := filesystem.SeparateFilePath(opath, dirPath)

	// Parse filename
	fElements := tanuki.Parse(info.Filename, tanuki.DefaultOptions)
	parsedInfo := newLocalFileParsedData(info.Filename, fElements)

	// Parse dirnames
	parsedFolderInfo := make([]*LocalFileParsedData, 0)
	for _, dirname := range info.Dirnames {
		if len(dirname) > 0 {
			pElements := tanuki.Parse(dirname, tanuki.DefaultOptions)
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
func newLocalFileParsedData(original string, elements *tanuki.Elements) *LocalFileParsedData {
	i := new(LocalFileParsedData)
	i.Original = original
	i.Title = elements.AnimeTitle
	i.ReleaseGroup = elements.ReleaseGroup
	i.EpisodeTitle = elements.EpisodeTitle
	i.Year = elements.AnimeYear

	if len(elements.AnimeSeason) > 0 {
		if len(elements.AnimeSeason) == 1 {
			i.Season = elements.AnimeSeason[0]
		} else {
			i.SeasonRange = elements.AnimeSeason
		}
	}

	if len(elements.EpisodeNumber) > 0 {
		if len(elements.EpisodeNumber) == 1 {
			i.Episode = elements.EpisodeNumber[0]
		} else {
			i.EpisodeRange = elements.EpisodeNumber
		}
	}

	if len(elements.AnimePart) > 0 {
		if len(elements.AnimePart) == 1 {
			i.Part = elements.AnimePart[0]
		} else {
			i.PartRange = elements.AnimePart
		}
	}

	return i
}
