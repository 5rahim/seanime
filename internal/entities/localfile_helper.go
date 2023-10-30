package entities

import (
	"bytes"
	"github.com/samber/lo"
	"github.com/samber/lo/parallel"
	"strings"
)

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

// GetUniqueAnimeTitles returns all parsed anime titles without duplicates
func GetUniqueAnimeTitles(localFiles []*LocalFile) []string {
	// Concurrently get title from each local file
	titles := parallel.Map(localFiles, func(file *LocalFile, index int) string {
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
