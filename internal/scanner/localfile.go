package scanner

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/filesystem"
	"github.com/seanime-app/tanuki"
	"strings"
)

type AugmentedLocalFile struct {
	LocalFile  *LocalFile
	Media      any
	AnimeTitle string
}

type LocalFile struct {
	Path             string                 `json:"path"`
	Name             string                 `json:"name"`
	ParsedData       *LocalFileParsedData   `json:"parsedInfo"`
	ParsedFolderData []*LocalFileParsedData `json:"parsedFolderInfo"`
	Metadata         *LocalFileMetadata     `json:"metadata"`
	Locked           bool                   `json:"locked"`
	Ignored          bool                   `json:"ignored"`
	MediaId          int                    `json:"mediaId"`
}

type LocalFileParsedData struct {
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

type LocalFileMetadata struct {
	Episode      int    `json:"episode"`
	AniDBEpisode string `json:"aniDBEpisode"`
	IsVersion    bool   `json:"isVersion"`
	IsSpecial    bool   `json:"isSpecial"`
	IsNC         bool   `json:"isNC"`
}

// LocalFileWithMedia Same as LocalFile but contains the fetched Media
type LocalFileWithMedia struct {
	*LocalFile
	Media any
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// LocalFile methods

func (f *LocalFile) GetParsedData() *LocalFileParsedData {
	return f.ParsedData
}

// GetParsedTitle returns the parsed title. Prefers the last parsed folder title if available.
func (f *LocalFile) GetParsedTitle() string {
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		title := f.ParsedFolderData[len(f.ParsedFolderData)-1].Title
		if len(title) > 0 {
			return title
		}
	}
	if len(f.ParsedData.Title) > 0 {
		return f.ParsedData.Title
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetLocalFilesFromDir creates a new LocalFile for each video file
func GetLocalFilesFromDir(dirPath string, logger *zerolog.Logger) ([]*LocalFile, error) {
	paths, err := filesystem.GetVideoFilePathsFromDir(dirPath)

	logger.Debug().
		Any("dirPath", dirPath).
		Msg("[localfile] Retrieving local files")

	// Concurrently populate localFiles
	localFiles := lop.Map(paths, func(path string, index int) *LocalFile {
		return NewLocalFile(path, dirPath)
	})

	logger.Debug().
		Any("count", len(localFiles)).
		Msg("[localfile] Retrieved local files")

	return localFiles, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewLocalFile creates and returns a reference to a new LocalFile struct from a path
func NewLocalFile(opath, dirPath string) *LocalFile {

	info := filesystem.SeparateFilePath(opath, dirPath)

	// Parse filename
	fElements := tanuki.Parse(info.Filename, tanuki.DefaultOptions)
	parsedInfo := NewLocalFileParsedData(info.Filename, fElements)

	// Parse dirnames
	parsedFolderInfo := make([]*LocalFileParsedData, 0)
	for _, dirname := range info.Dirnames {
		if len(dirname) > 0 {
			pElements := tanuki.Parse(dirname, tanuki.DefaultOptions)
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
			IsVersion:    false,
			IsSpecial:    false,
			IsNC:         false,
		},
		Locked:  false,
		Ignored: false,
		MediaId: 0,
	}

	return localFile

}

// NewLocalFileParsedData Converts tanuki.Elements into LocalFileParsedData.
//
// This is used by NewLocalFile
func NewLocalFileParsedData(original string, elements *tanuki.Elements) *LocalFileParsedData {
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetUniqueAnimeTitles returns all parsed anime titles without duplicates
func GetUniqueAnimeTitles(localFiles []*LocalFile) []string {
	// Concurrently get title from each local file
	titles := lop.Map(localFiles, func(file *LocalFile, index int) string {
		title := file.GetParsedTitle()
		// Some rudimentary exclusions
		for _, i := range []string{"SPECIALS", "SPECIAL", "EXTRA", "NC", "OP", "MOVIE", "MOVIES"} {
			if strings.ToUpper(title) == i {
				return ""
			}
		}
		return title
	})
	// Keep unique title and filter out empty ones
	titles = lo.Filter(lo.Uniq(titles), func(item string, index int) bool {
		return len(item) > 0
	})
	return titles
}
