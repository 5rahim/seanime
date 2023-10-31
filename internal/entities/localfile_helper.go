package entities

import (
	"bytes"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"slices"
	"strings"
)

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
	groupedLfs := lop.GroupBy(lfs, func(item *LocalFile) int {
		return item.MediaId
	})

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
