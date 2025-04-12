package anime

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
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
	if f.Metadata == nil {
		return -1
	}
	return f.Metadata.Episode
}

func (f *LocalFile) GetParsedEpisodeTitle() string {
	if f.ParsedData == nil {
		return ""
	}
	return f.ParsedData.EpisodeTitle
}

// HasBeenWatched returns whether the episode has been watched.
// This only applies to main episodes.
func (f *LocalFile) HasBeenWatched(progress int) bool {
	if f.Metadata == nil {
		return false
	}
	if f.GetEpisodeNumber() == 0 && progress == 0 {
		return false
	}
	return progress >= f.GetEpisodeNumber()
}

// GetType returns the metadata type.
// This requires the LocalFile to be hydrated.
func (f *LocalFile) GetType() LocalFileType {
	return f.Metadata.Type
}

// IsMain returns true if the metadata type is LocalFileTypeMain
func (f *LocalFile) IsMain() bool {
	return f.Metadata.Type == LocalFileTypeMain
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

// GetNormalizedPath returns the lowercase path of the LocalFile.
// Use this for comparison.
func (f *LocalFile) GetNormalizedPath() string {
	return util.NormalizePath(f.Path)
}

func (f *LocalFile) GetPath() string {
	return f.Path
}

func (f *LocalFile) HasSamePath(path string) bool {
	return f.GetNormalizedPath() == util.NormalizePath(path)
}

// IsInDir returns true if the LocalFile is in the given directory.
func (f *LocalFile) IsInDir(dirPath string) bool {
	dirPath = util.NormalizePath(dirPath)
	if !filepath.IsAbs(dirPath) {
		return false
	}
	return strings.HasPrefix(f.GetNormalizedPath(), dirPath)
}

// IsAtRootOf returns true if the LocalFile is at the root of the given directory.
func (f *LocalFile) IsAtRootOf(dirPath string) bool {
	dirPath = strings.TrimSuffix(util.NormalizePath(dirPath), "/")
	return filepath.ToSlash(filepath.Dir(f.GetNormalizedPath())) == dirPath
}

func (f *LocalFile) Equals(lf *LocalFile) bool {
	return util.NormalizePath(f.Path) == util.NormalizePath(lf.Path)
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

// buildTitle concatenates the given strings into a single string.
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

// GetUniqueAnimeTitlesFromLocalFiles returns all parsed anime titles without duplicates, from a slice of LocalFile's.
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

// GetMediaIdsFromLocalFiles returns all media ids from a slice of LocalFile's.
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

// GetLocalFilesFromMediaId returns all local files with the given media id.
func GetLocalFilesFromMediaId(lfs []*LocalFile, mId int) []*LocalFile {

	return lo.Filter(lfs, func(item *LocalFile, _ int) bool {
		return item.MediaId == mId
	})

}

// GroupLocalFilesByMediaID returns a map of media id to local files.
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
	folderTitles := make([]string, 0)
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		// Go through each folder data and keep the ones with a title
		data := lo.Filter(f.ParsedFolderData, func(fpd *LocalFileParsedData, _ int) bool {
			return len(fpd.Title) > 0
		})
		if len(data) == 0 {
			return ""
		}
		// Get the titles
		for _, v := range data {
			folderTitles = append(folderTitles, v.Title)
		}
		// If there are multiple titles, return the one closest to the end
		return folderTitles[len(folderTitles)-1]
	}

	return ""
}

// GetTitleVariations is used for matching.
func (f *LocalFile) GetTitleVariations() []*string {

	folderSeason := 0

	// Get the season from the folder data
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

	part := 0

	// Get the part from the folder data
	if f.ParsedFolderData != nil && len(f.ParsedFolderData) > 0 {
		v, found := lo.Find(f.ParsedFolderData, func(fpd *LocalFileParsedData) bool {
			return len(fpd.Part) > 0
		})
		if found {
			if res, ok := util.StringToInt(v.Season); ok {
				part = res
			}
		}
	}

	// Get the part from the filename
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

	bothTitles := len(f.ParsedData.Title) > 0 && len(folderTitle) > 0                    // Both titles are present (filename and folder)
	noSeasonsOrParts := folderSeason == 0 && season == 0 && part == 0                    // No seasons or parts are present
	bothTitlesSimilar := bothTitles && strings.Contains(folderTitle, f.ParsedData.Title) // The folder title contains the filename title
	eitherSeason := folderSeason > 0 || season > 0                                       // Either season is present
	eitherSeasonFirst := folderSeason == 1 || season == 1                                // Either season is 1

	// Part
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

	// Title, no seasons, no parts, or season 1
	// e.g. "Bungou Stray Dogs"
	// e.g. "Bungou Stray Dogs Season 1"
	if noSeasonsOrParts || eitherSeasonFirst {
		if len(f.ParsedData.Title) > 0 { // Add filename title
			titleVariations = append(titleVariations, f.ParsedData.Title)
		}
		if len(folderTitle) > 0 { // Both titles are present and similar, add folder title
			titleVariations = append(titleVariations, folderTitle)
		}
	}

	// Part & Season
	// e.g. "Spy x Family Season 1 Part 2"
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

	// Season is present
	if eitherSeason {
		arr := make([]string, 0)

		seas := folderSeason // Default to folder parsed season
		if season > 0 {      // Use filename parsed season if present
			seas = season
		}

		// Both titles are present
		if bothTitles {
			// Add both titles
			arr = append(arr, f.ParsedData.Title)
			arr = append(arr, folderTitle)
			if !bothTitlesSimilar { // Combine both titles if they are not similar
				arr = append(arr, fmt.Sprintf("%s %s", folderTitle, f.ParsedData.Title))
			}
		} else if len(folderTitle) > 0 { // Only folder title is present

			arr = append(arr, folderTitle)

		} else if len(f.ParsedData.Title) > 0 { // Only filename title is present

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

	// If there are no title variations, use the folder title or the parsed title
	if len(titleVariations) == 0 {
		if len(folderTitle) > 0 {
			titleVariations = append(titleVariations, folderTitle)
		}
		if len(f.ParsedData.Title) > 0 {
			titleVariations = append(titleVariations, f.ParsedData.Title)
		}
	}

	return lo.ToSlicePtr(titleVariations)

}
