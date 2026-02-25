package anime

import (
	"bytes"
	"cmp"
	"fmt"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
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

// getRegexCompFilename returns a normalized filename without words that could mess up regex comparisons
func (f *LocalFile) getRegexCompFilename() (ret string) {
	ret = f.ParsedData.Original
	if f.ParsedData.EpisodeTitle != "" {
		ret = strings.Replace(ret, f.ParsedData.EpisodeTitle, "PLACEHOLDER", 1)
	}
	if f.ParsedData.ReleaseGroup != "" {
		ret = strings.Replace(ret, f.ParsedData.ReleaseGroup, "PLACEHOLDER", 1)
	}
	return
}

func (f *LocalFile) IsProbablyNC() bool {
	return comparison.ValueContainsNC(f.getRegexCompFilename())
}

func (f *LocalFile) IsProbablySpecial() bool {
	return comparison.ValueContainsSpecial(f.getRegexCompFilename())
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
	return cmp.Or(f.ParsedData.Title, f.GetFolderTitle())
}

// GetFolderTitle returns the parsed title of the closest folder to the file.
// It will ignore any folder titles that are simply keywords like "specials", "extras", etc.
// If all is true, it will return the title regardless of whether it's a keyword or not.
func (f *LocalFile) GetFolderTitle(all ...bool) string {
	folderTitles := make([]string, 0)
	if len(f.ParsedFolderData) > 0 {
		// Go through each folder data and keep the ones with a title
		data := lo.Filter(f.ParsedFolderData, func(fpd *LocalFileParsedData, _ int) bool {
			// remove non-anime titles
			cleanTitle := strings.TrimSpace(strings.ToLower(fpd.Title))
			if len(all) == 0 || !all[0] {
				if _, ok := comparison.IgnoredFilenames[cleanTitle]; ok {
					return false
				}
				// Also check the original folder name for ignored keywords
				if comparison.ValueContainsIgnoredKeywords(fpd.Original) {
					return false
				}
			}
			return len(cleanTitle) > 0
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

// GetAllFolderTitles returns all valid folder titles (not just the closest one).
func (f *LocalFile) GetAllFolderTitles() []string {
	if len(f.ParsedFolderData) == 0 {
		return nil
	}
	titles := make([]string, 0, len(f.ParsedFolderData))
	for _, fpd := range f.ParsedFolderData {
		cleanTitle := strings.TrimSpace(strings.ToLower(fpd.Title))
		if len(cleanTitle) == 0 {
			continue
		}
		if _, ok := comparison.IgnoredFilenames[cleanTitle]; ok {
			continue
		}
		// Also check the original folder name for ignored keywords
		if comparison.ValueContainsIgnoredKeywords(fpd.Original) {
			continue
		}
		titles = append(titles, fpd.Title)
	}
	return titles
}

// GetTitleVariations returns title variations for the local file.
func (f *LocalFile) GetTitleVariations() []*string {

	folderSeason := 0

	// Get the season from the folder data
	if len(f.ParsedFolderData) > 0 {
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
	if len(f.ParsedFolderData) > 0 {
		v, found := lo.Find(f.ParsedFolderData, func(fpd *LocalFileParsedData) bool {
			return len(fpd.Part) > 0
		})
		if found {
			if res, ok := util.StringToInt(v.Part); ok {
				part = res
			}
		}
	}

	folderTitle := f.GetFolderTitle()

	// Collect all valid folder titles (not just the closest one)
	// This ensures parent folder titles like "Re Zero kara Hajimeru Isekai Seikatsu" are used
	// for token-based candidate matching even when the closest folder has a shorter name like "ReZero"
	allFolderTitles := f.GetAllFolderTitles()

	// shortcircuit if there are no titles
	if len(f.ParsedData.Title) == 0 && len(folderTitle) == 0 && len(allFolderTitles) == 0 {
		return make([]*string, 0)
	}

	titleVariations := make([]string, 0, 10)

	bothTitles := len(f.ParsedData.Title) > 0 && len(folderTitle) > 0
	noSeasonsOrParts := folderSeason == 0 && season == 0 && part == 0
	bothTitlesSimilar := bothTitles && strings.Contains(folderTitle, f.ParsedData.Title)
	eitherSeason := folderSeason > 0 || season > 0
	eitherSeasonFirst := folderSeason == 1 || season == 1

	// Collect base titles to use
	// Primary titles: filename and closest folder (used for season/part variations)
	// Extra titles: parent folder titles (used as raw base titles for candidate lookup and matching with lower weight)
	primaryTitles := make([]string, 0, 2)
	if len(f.ParsedData.Title) > 0 {
		primaryTitles = append(primaryTitles, f.ParsedData.Title)
	}
	if len(folderTitle) > 0 && folderTitle != f.ParsedData.Title {
		primaryTitles = append(primaryTitles, folderTitle)
	}

	// Extra folder titles from parent folders (not the closest folder)
	extraTitles := make([]string, 0, len(allFolderTitles))
	for _, ft := range allFolderTitles {
		alreadyAdded := false
		for _, bt := range primaryTitles {
			if ft == bt {
				alreadyAdded = true
				break
			}
		}
		if !alreadyAdded {
			extraTitles = append(extraTitles, ft)
		}
	}

	// All base titles combined for raw title addition and candidate lookup
	baseTitles := make([]string, 0, len(primaryTitles)+len(extraTitles))
	baseTitles = append(baseTitles, primaryTitles...)
	baseTitles = append(baseTitles, extraTitles...)

	// Add primary titles as raw base title variations
	for _, t := range primaryTitles {
		titleVariations = append(titleVariations, t)
	}
	// Add parent folder titles (far away titles)
	// We'll optionally penalize them later during scoring to avoid false positives.
	for _, t := range extraTitles {
		titleVariations = append(titleVariations, t)
	}

	// Part variations (only from primary titles, not parent folder titles)
	if part > 0 {
		for _, t := range primaryTitles {
			titleVariations = append(titleVariations,
				buildTitle(t, "Part", strconv.Itoa(part)),
				buildTitle(t, "Part", util.IntegerToOrdinal(part)),
				buildTitle(t, "Cour", strconv.Itoa(part)),
			)
			// Roman numerals for parts 1-3
			if partRoman := intToRoman(part); partRoman != "" {
				titleVariations = append(titleVariations, buildTitle(t, "Part", partRoman))
			}
		}
	}

	// Season variations (only from primary titles, not parent folder titles)
	if eitherSeason {
		seas := folderSeason
		if season > 0 {
			seas = season
		}

		for _, t := range primaryTitles {
			// Standard formats
			titleVariations = append(titleVariations,
				buildTitle(t, "Season", strconv.Itoa(seas)),          // "Title Season 2"
				buildTitle(t, "S"+strconv.Itoa(seas)),                // "Title S2"
				buildTitle(t, fmt.Sprintf("S%02d", seas)),            // "Title S02"
				buildTitle(t, util.IntegerToOrdinal(seas), "Season"), // "Title 2nd Season"
				fmt.Sprintf("%s %d", t, seas),                        // "Title 2" (common pattern)
			)
		}

		// Combined with part
		if part > 0 {
			for _, t := range primaryTitles {
				titleVariations = append(titleVariations,
					buildTitle(t, "Season", strconv.Itoa(seas), "Part", strconv.Itoa(part)),
					buildTitle(t, fmt.Sprintf("S%d", seas), fmt.Sprintf("Part %d", part)),
				)
			}
		}
	}

	// Season 1 or no season info. base titles already added
	if noSeasonsOrParts || eitherSeasonFirst {
		// Already added base titles above
		// For season 1, also add without the "Season 1" suffix as many first seasons
		// don't have season indicators in their official titles
	}

	// Combined folder + filename title variations
	// e.g. "Anime/S02/Episode.mkv"
	if bothTitles && !bothTitlesSimilar {
		combined := fmt.Sprintf("%s %s", folderTitle, f.ParsedData.Title)
		titleVariations = append(titleVariations, combined)
	}

	// Deduplicate
	titleVariations = lo.Uniq(titleVariations)

	// If there are still no title variations, fall back to raw titles
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

// intToRoman converts small integers (1-10) to Roman numerals
func intToRoman(n int) string {
	romans := map[int]string{
		1: "I", 2: "II", 3: "III", 4: "IV", 5: "V",
		6: "VI", 7: "VII", 8: "VIII", 9: "IX", 10: "X",
	}
	if r, ok := romans[n]; ok {
		return r
	}
	return ""
}
