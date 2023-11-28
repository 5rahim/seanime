package entities

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/util"
	"slices"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------------------------------------------------

func (f *LocalFile) IsParsedEpisodeValid() bool {
	if f == nil || f.ParsedData == nil {
		return false
	}
	return len(f.ParsedData.Episode) > 0
}

// GetEpisodeNumber returns the metadata episode number.
// This requires the LocalFile to be hydrated.
func (f *LocalFile) GetEpisodeNumber() int {
	return f.Metadata.Episode
}

// GetType returns the metadata type.
// This requires the LocalFile to be hydrated.
func (f *LocalFile) GetType() LocalFileType {
	return f.Metadata.Type
}

// GetMetadata returns the file metadata.
// This requires the LocalFile to be hydrated.
func (f *LocalFile) GetMetadata() *LocalFileMetadata {
	return f.Metadata
}

// GetAniDBEpisode returns the metadata AniDB episode number.
// This requires the LocalFile to be hydrated.
func (f *LocalFile) GetAniDBEpisode() string {
	return f.Metadata.AniDBEpisode
}

func (f *LocalFile) IsLocked() bool {
	return f.Locked
}
func (f *LocalFile) IsIgnored() bool {
	return f.Ignored
}

// GetPath returns the lowercased path of the LocalFile.
// Use this for comparison.
func (f *LocalFile) GetPath() string {
	return strings.ToLower(f.Path)
}
func (f *LocalFile) HasSamePath(path string) bool {
	return strings.ToLower(f.Path) == strings.ToLower(path)
}

func (f *LocalFile) Equals(lf *LocalFile) bool {
	return strings.ToLower(f.Path) == strings.ToLower(lf.Path)
}

func (f *LocalFile) IsIncluded(lfs []*LocalFile) bool {
	for _, lf := range lfs {
		if f.Equals(lf) {
			return true
		}
	}
	return false
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
	groupedLfs := GroupLocalFilesByMediaID(lfs)

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

// GroupLocalFilesByMediaID groups local files by media id
func GroupLocalFilesByMediaID(lfs []*LocalFile) (groupedLfs map[int][]*LocalFile) {
	groupedLfs = lop.GroupBy(lfs, func(item *LocalFile) int {
		return item.MediaId
	})

	return
}

// IsLocalFileGroupValidEntry checks if there are any main episodes with valid episodes
func IsLocalFileGroupValidEntry(lfs []*LocalFile) bool {
	// Check if there are any main episodes with valid parsed data
	flag := false
	for _, lf := range lfs {
		if lf.GetType() == LocalFileTypeMain && lf.IsParsedEpisodeValid() {
			flag = true
			break
		}
	}
	return flag
}

// FindLatestLocalFileFromGroup returns the "main" episode with the highest episode number.
// Returns false if there are no episodes.
func FindLatestLocalFileFromGroup(lfs []*LocalFile) (*LocalFile, bool) {
	// Check if there are any main episodes with valid parsed data
	if !IsLocalFileGroupValidEntry(lfs) {
		return nil, false
	}
	if lfs == nil || len(lfs) == 0 {
		return nil, false
	}
	// Get the episode with the highest progress number
	latest, found := lo.Find(lfs, func(lf *LocalFile) bool {
		return lf.GetType() == LocalFileTypeMain && lf.IsParsedEpisodeValid()
	})
	if !found {
		return nil, false
	}
	for _, lf := range lfs {
		if lf.GetType() == LocalFileTypeMain && lf.GetEpisodeNumber() > latest.GetEpisodeNumber() {
			latest = lf
		}
	}
	if latest == nil || latest.GetType() != LocalFileTypeMain {
		return nil, false
	}
	return latest, true
}

func (f *LocalFile) GetParsedData() *LocalFileParsedData {
	return f.ParsedData
}

// GetParsedTitle returns the parsed title of the LocalFile. Falls back to the folder title if the file title is empty.
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
	//folderDepth := 0

	// Get the season from the folder data
	folderSeason := 0
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		//folderDepth = len(f.ParsedFolderData)

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

	if comparison.ValueContainsIgnoredKeywords(folderTitle) {
		folderTitle = ""
	}

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
		if bothTitles && !bothTitlesSimilar {
			titleVariations = append(titleVariations, fmt.Sprintf("%s %s", folderTitle, f.ParsedData.Title))
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

		// Both titles are present
		if bothTitles {
			// Add filename parsed title
			arr = append(arr, f.ParsedData.Title)
			arr = append(arr, folderTitle)
			if !bothTitlesSimilar {
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
