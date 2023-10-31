package entities

import (
	"fmt"
	"github.com/5rahim/tanuki"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/filesystem"
	"github.com/seanime-app/seanime-server/internal/util"
	"strconv"
	"strings"
)

type (
	LocalFile struct {
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
		Episode      int    `json:"episode"`
		AniDBEpisode string `json:"aniDBEpisode"`
		IsVersion    bool   `json:"isVersion"`
		IsSpecial    bool   `json:"isSpecial"`
		IsNC         bool   `json:"isNC"`
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

func (f *LocalFile) GetParsedData() *LocalFileParsedData {
	return f.ParsedData
}

// GetParsedTitle returns the parsed title. Prefers the last parsed folder title if available.
func (f *LocalFile) GetParsedTitle() string {
	if len(f.ParsedData.Title) > 0 {
		return f.ParsedData.Title
	}
	if len(f.GetFolderTitle()) > 0 {
		return f.GetFolderTitle()
	}
	return ""
}

func (f *LocalFile) GetFolderTitle() string {
	folderTitle := ""
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		v, found := lo.Find(f.ParsedFolderData, func(fpd *LocalFileParsedData) bool {
			return len(fpd.Title) > 0
		})
		if found {
			folderTitle = v.Title
		}
	}
	return folderTitle
}

func (f *LocalFile) GetTitleVariations() []*string {
	// Get the season from the folder data
	folderSeason := 0
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		v, found := lo.Find(f.ParsedFolderData, func(fpd *LocalFileParsedData) bool {
			return len(fpd.Season) > 0
		})
		if found {
			if res, ok := util.StringToInt(v.Season); ok {
				folderSeason = res
			}
		}
	}

	// Get the season from the filename
	season := 0
	if len(f.ParsedData.Season) > 0 {
		if res, ok := util.StringToInt(f.ParsedData.Season); ok {
			season = res
		}
	}

	// Get the part from the filename
	part := 0
	if len(f.ParsedData.Part) > 0 {
		if res, ok := util.StringToInt(f.ParsedData.Part); ok {
			part = res
		}
	}

	folderTitle := f.GetFolderTitle()

	if len(f.ParsedData.Title) == 0 && len(folderTitle) == 0 {
		return make([]*string, 0)
	}

	titleVariations := make([]string, 0)

	bothTitles := len(f.ParsedData.Title) > 0 && len(folderTitle) > 0
	noSeasonsOrParts := folderSeason == 0 && season == 0 && part == 0
	bothTitlesSimilar := bothTitles && strings.Contains(folderTitle, f.ParsedData.Title)
	eitherSeason := folderSeason > 0 || season > 0
	eitherSeasonFirst := folderSeason == 1 || season == 1

	if part > 0 {
		if len(folderTitle) > 0 {
			titleVariations = append(titleVariations,
				buildTitle(folderTitle, "Part", strconv.Itoa(part)),
				buildTitle(folderTitle, "Part", util.IntegerToOrdinal(part)),
				buildTitle(folderTitle, "Cour", strconv.Itoa(part)),
				buildTitle(folderTitle, "Cour", util.IntegerToOrdinal(part)),
			)
		}
		if len(f.ParsedData.Title) > 0 {
			titleVariations = append(titleVariations,
				buildTitle(f.ParsedData.Title, "Part", strconv.Itoa(part)),
				buildTitle(f.ParsedData.Title, "Part", util.IntegerToOrdinal(part)),
				buildTitle(f.ParsedData.Title, "Cour", strconv.Itoa(part)),
				buildTitle(f.ParsedData.Title, "Cour", util.IntegerToOrdinal(part)),
			)
		}
	}

	if noSeasonsOrParts || eitherSeasonFirst {
		if len(folderTitle) > 0 && bothTitlesSimilar {
			titleVariations = append(titleVariations, folderTitle)
		}
		if len(f.ParsedData.Title) > 0 {
			titleVariations = append(titleVariations, f.ParsedData.Title)
		}
	}

	if part > 0 && eitherSeason {
		if len(folderTitle) > 0 {
			if season > 0 {
				titleVariations = append(titleVariations,
					buildTitle(folderTitle, "Season", strconv.Itoa(season), "Part", strconv.Itoa(part)),
				)
			} else if folderSeason > 0 {
				titleVariations = append(titleVariations,
					buildTitle(folderTitle, "Season", strconv.Itoa(folderSeason), "Part", strconv.Itoa(part)),
				)
			}
		}
		if len(f.ParsedData.Title) > 0 {
			if season > 0 {
				titleVariations = append(titleVariations,
					buildTitle(f.ParsedData.Title, "Season", strconv.Itoa(season), "Part", strconv.Itoa(part)),
				)
			} else if folderSeason > 0 {
				titleVariations = append(titleVariations,
					buildTitle(f.ParsedData.Title, "Season", strconv.Itoa(folderSeason), "Part", strconv.Itoa(part)),
				)
			}
		}
	}

	if eitherSeason {
		arr := make([]string, 0)

		seas := folderSeason
		if season > 0 {
			seas = season
		}

		if bothTitles {
			arr = append(arr, f.ParsedData.Title)
			if bothTitlesSimilar {
				arr = append(arr, folderTitle)
			} else {
				arr = append(arr, fmt.Sprintf("%s %s", folderTitle, f.ParsedData.Title))
			}
		} else if len(folderTitle) > 0 {
			arr = append(arr, folderTitle)
		} else if len(f.ParsedData.Title) > 0 {
			arr = append(arr, f.ParsedData.Title)
		}

		for _, t := range arr {
			titleVariations = append(titleVariations,
				buildTitle(t, "Season", strconv.Itoa(seas)),
				buildTitle(t, "S"+strconv.Itoa(seas)),
				buildTitle(t, util.IntegerToOrdinal(seas), "Season"),
			)
		}
	}

	titleVariations = lo.Uniq(titleVariations)

	return lo.ToSlicePtr(titleVariations)

}

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
