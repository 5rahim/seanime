package entities

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/util"
	"slices"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------------------------------------------------

func (f *LocalFile) GetEpisodeNumber() int {
	return f.Metadata.Episode
}
func (f *LocalFile) GetAniDBEpisode() string {
	return f.Metadata.AniDBEpisode
}

//----------------------------------------------------------------------------------------------------------------------

func buildTitle(vals ...string) string {
	buf := bytes.NewBuffer([]byte{})
	for i, v := range vals {
		buf.WriteString(v)
		if i != len(vals)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

// GetUniqueAnimeTitlesFromLocalFiles returns all parsed anime titles without duplicates
func GetUniqueAnimeTitlesFromLocalFiles(lfs []*LocalFile) []string {
	// Concurrently get title from each local file
	titles := lop.Map(lfs, func(file *LocalFile, index int) string {
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

func GetMediaIdsFromLocalFiles(lfs []*LocalFile) []int {

	// Group local files by media id
	groupedLfs := GetGroupedLocalFiles(lfs)

	// Get slice of media ids from local files
	mIds := make([]int, len(groupedLfs))
	for key := range groupedLfs {
		if !slices.Contains(mIds, key) {
			mIds = append(mIds, key)
		}
	}

	return mIds

}

func GetLocalFilesFromMediaId(lfs []*LocalFile, mId int) []*LocalFile {

	return lo.Filter(lfs, func(item *LocalFile, _ int) bool {
		return item.MediaId == mId
	})

}

func GetGroupedLocalFiles(lfs []*LocalFile) (groupedLfs map[int][]*LocalFile) {
	groupedLfs = lop.GroupBy(lfs, func(item *LocalFile) int {
		return item.MediaId
	})

	return
}

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

// GetTitleVariations is used for matching.
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
