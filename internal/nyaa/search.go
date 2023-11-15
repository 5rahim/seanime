package nyaa

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mmcdole/gofeed"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/util"
	"strings"
	"unicode"
)

// https://github.com/irevenko/go-nyaa

type (
	SearchOptions struct {
		Provider string
		Query    string
		Category string
		SortBy   string
		Filter   string
	}
	BuildSearchQueryOptions struct {
		Media          *anilist.BaseMedia
		Batch          *bool
		EpisodeNumber  *int
		AbsoluteOffset *int
		Quality        *string
	}
)

func Search(opts SearchOptions) ([]Torrent, error) {

	fp := gofeed.NewParser()

	url, err := buildURL(opts)
	if err != nil {
		return nil, err
	}

	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	res := convertRSS(feed)

	return res, nil
}

func BuildSearchQuery(opts *BuildSearchQueryOptions) (string, bool) {

	if opts.Media == nil || opts.Batch == nil || opts.EpisodeNumber == nil || opts.AbsoluteOffset == nil {
		return "", false
	}

	romTitle := opts.Media.GetRomajiTitleSafe()
	engTitle := opts.Media.GetTitleSafe()

	//episodes := []string{strconv.Itoa(*opts.EpisodeNumber)} FIXME remove
	//// We include the offsetted episode if it's within the total episode count
	//if *opts.AbsoluteOffset > 0 && opts.Media.GetCurrentEpisodeCount() > (*opts.EpisodeNumber+*opts.AbsoluteOffset) {
	//	episodes = append(episodes, strconv.Itoa(*opts.EpisodeNumber+*opts.AbsoluteOffset))
	//}

	//parsedRom :=

	season := 0
	part := 0

	titles := make([]string, 0)
	for _, title := range opts.Media.GetAllTitles() {
		s, cTitle := util.ExtractSeasonNumber(*title)
		p, cTitle := util.ExtractPartNumber(cTitle)
		if s != 0 {
			season = s
		}
		if p != 0 {
			part = p
		}
		if cTitle != "" {
			titles = append(titles, cTitle)
		}
	}

	// Check season from synonyms
	for _, synonym := range opts.Media.Synonyms {
		s, _ := util.ExtractSeasonNumber(*synonym)
		if s != 0 {
			season = s
		}
	}

	// no season or part got parsed, meaning there is no clean title,
	// add romaji and english titles to the title list
	if season == 0 && part == 0 {
		titles = append(titles, romTitle)
		if len(engTitle) > 0 {
			titles = append(titles, engTitle)
		}
	}

	//// use tanuki to try and get a clean title
	//parsedRom := tanuki.Parse(romTitle, tanuki.DefaultOptions)
	//if parsedRom.AnimeTitle != "" {
	//	titles = append(titles, parsedRom.AnimeTitle)
	//}
	//if engTitle != "" {
	//	parsedEng := tanuki.Parse(engTitle, tanuki.DefaultOptions)
	//	if parsedEng.AnimeTitle != "" {
	//		titles = append(titles, parsedEng.AnimeTitle)
	//	}
	//}

	for i, title := range titles {
		titles[i] = strings.TrimSpace(strings.ReplaceAll(title, ":", " "))
		titles[i] = strings.TrimSpace(strings.ReplaceAll(titles[i], "-", " "))
		titles[i] = strings.Join(strings.Fields(titles[i]), " ")
		titles[i] = strings.ToLower(titles[i])
	}
	titles = lo.Uniq(titles)

	//// Add some synonyms FIXME remove
	//for _, syn := range opts.Media.Synonyms {
	//	if len(*syn) > 4 && isMostlyLatinString(*syn) {
	//		titles = append(titles, *syn)
	//	}
	//}

	//
	// Parameters
	//

	canBatch := false
	if *opts.Media.GetStatus() == anilist.MediaStatusFinished && opts.Media.GetTotalEpisodeCount() > 0 {
		canBatch = true
	}

	// Batch section
	// 1. If the media is finished and has more than 1 episode
	// 2. If the media is not a movie
	// 3. If the media is not a single episode
	batchBuff := bytes.NewBufferString("") // this will be joined by |
	if *opts.Batch && canBatch && *opts.Media.GetFormat() != anilist.MediaFormatMovie && opts.Media.GetTotalEpisodeCount() != 1 {
		batchBuff.WriteString("(")
		// e.g. 01-12
		s1 := fmt.Sprintf("%s%s%s", zeropad("1"), " - ", zeropad(opts.Media.GetTotalEpisodeCount()))
		batchBuff.WriteString(s1)
		batchBuff.WriteString("|")
		// e.g. 01~12
		s2 := fmt.Sprintf("%s%s%s", zeropad("1"), " ~ ", zeropad(opts.Media.GetTotalEpisodeCount()))
		batchBuff.WriteString(s2)
		batchBuff.WriteString("|")
		// e.g. 01~12
		batchBuff.WriteString("Batch")
		batchBuff.WriteString("|")
		batchBuff.WriteString("Complete")
		batchBuff.WriteString(")")
	}
	batchStr := batchBuff.String()

	titleStr := fmt.Sprintf("(%s)", strings.Join(titles, "|"))

	println(spew.Sdump(titleStr, batchStr))

	return "", false
}

func zeropad(v interface{}) string {
	switch i := v.(type) {
	case int:
		return fmt.Sprintf("%02d", i)
	case string:
		return fmt.Sprintf("%02s", i)
	default:
		return ""
	}
}

func isMostlyLatinString(str string) bool {
	if len(str) <= 0 {
		return false
	}
	latinLength := 0
	nonLatinLength := 0
	for _, r := range str {
		if isLatinRune(r) {
			latinLength++
		} else {
			nonLatinLength++
		}
	}
	return latinLength > nonLatinLength
}

func isLatinRune(r rune) bool {
	return unicode.In(r, unicode.Latin)
}
