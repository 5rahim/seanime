package nyaa

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mmcdole/gofeed"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/sourcegraph/conc/pool"
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
	SearchMultipleOptions struct {
		Provider string
		Query    []string
		Category string
		SortBy   string
		Filter   string
	}
	BuildSearchQueryOptions struct {
		Title          *string
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

	println(url)

	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	res := convertRSS(feed)

	return res, nil
}

func SearchMultiple(opts SearchMultipleOptions) ([]Torrent, error) {

	fp := gofeed.NewParser()

	p := pool.NewWithResults[[]Torrent]()
	for _, query := range opts.Query {
		p.Go(func() []Torrent {
			url, err := buildURL(SearchOptions{
				Provider: opts.Provider,
				Query:    query,
				Category: opts.Category,
				SortBy:   opts.SortBy,
				Filter:   opts.Filter,
			})
			if err != nil {
				return nil
			}
			feed, err := fp.ParseURL(url)
			if err != nil {
				return nil
			}
			return convertRSS(feed)
		})
	}
	slicesSlice := p.Wait()
	slicesSlice = lo.Filter(slicesSlice, func(i []Torrent, _ int) bool {
		return i != nil
	})
	res := lo.Flatten(slicesSlice)

	return res, nil
}

func BuildSearchQuery(opts *BuildSearchQueryOptions) ([]string, bool) {

	if opts.Media == nil || opts.Batch == nil || opts.EpisodeNumber == nil || opts.AbsoluteOffset == nil || opts.Quality == nil {
		return make([]string, 0), false
	}

	_ = *opts.EpisodeNumber
	romTitle := opts.Media.GetRomajiTitleSafe()
	engTitle := opts.Media.GetTitleSafe()

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

	if season == 0 && (strings.Contains(strings.ToLower(romTitle), " iii") || strings.Contains(strings.ToLower(engTitle), " iii")) {
		season = 3
	}
	if season == 0 && (strings.Contains(strings.ToLower(romTitle), " ii") || strings.Contains(strings.ToLower(engTitle), " ii")) {
		season = 2
	}

	split := strings.Split(romTitle, ":")
	if len(split) > 1 && len(split[0]) > 8 {
		titles = append(titles, split[0])
	}

	for i, title := range titles {
		titles[i] = strings.TrimSpace(strings.ReplaceAll(title, ":", " "))
		titles[i] = strings.TrimSpace(strings.ReplaceAll(titles[i], "-", " "))
		titles[i] = strings.Join(strings.Fields(titles[i]), " ")
		titles[i] = strings.ToLower(titles[i])
		if season != 0 {
			titles[i] = strings.ReplaceAll(titles[i], " iii", "")
			titles[i] = strings.ReplaceAll(titles[i], " ii", "")
		}
	}
	titles = lo.Uniq(titles)

	//
	// Parameters
	//

	canBatch := false
	if *opts.Media.GetStatus() == anilist.MediaStatusFinished && opts.Media.GetTotalEpisodeCount() > 0 {
		canBatch = true
	}

	normalBuff := bytes.NewBufferString("")

	// Batch section - empty unless:
	// 1. If the media is finished and has more than 1 episode
	// 2. If the media is not a movie
	// 3. If the media is not a single episode
	batchBuff := bytes.NewBufferString("")
	if *opts.Batch && canBatch && *opts.Media.GetFormat() != anilist.MediaFormatMovie && opts.Media.GetTotalEpisodeCount() != 1 {
		if season != 0 {
			batchBuff.WriteString(getSeasonGroup(season))
		}
		if part != 0 {
			batchBuff.WriteString(getPartGroup(part))
		}
		batchBuff.WriteString(getBatchGroup(opts.Media))

	} else {

		normalBuff.WriteString(getSeasonGroup(season))
		if part != 0 {
			normalBuff.WriteString(getPartGroup(part))
		}
		normalBuff.WriteString(getEpisodeGroup(*opts.EpisodeNumber))

	}

	titleStr := getTitleGroup(titles)
	batchStr := batchBuff.String()
	normalStr := normalBuff.String()

	// Replace titleStr if user provided one
	if opts.Title != nil && *opts.Title != "" {
		titleStr = fmt.Sprintf(`(%s)`, *opts.Title)
	}

	println(spew.Sdump(titleStr, batchStr, normalStr))

	query := fmt.Sprintf("%s%s%s%s", titleStr, batchStr, normalStr, *opts.Quality)
	query2 := ""

	// Absolute episode addition
	if !*opts.Batch && *opts.AbsoluteOffset > 0 {
		query2 = fmt.Sprintf("%s%s", getAbsoluteGroup(titleStr, opts), *opts.Quality) // e.g. jujutsu kaisen 25
	}

	println(spew.Sdump(query, query2))

	ret := []string{query}
	if query2 != "" {
		ret = append(ret, query2)
	}

	return ret, true
}

// (jjk|jujutsu kaisen)
func getTitleGroup(titles []string) string {
	return fmt.Sprintf("(%s)", strings.Join(titles, "|"))
}

func getAbsoluteGroup(title string, opts *BuildSearchQueryOptions) string {
	return fmt.Sprintf("(%s(%d))", title, *opts.EpisodeNumber+*opts.AbsoluteOffset)
}

// (s01e01)
func getSeasonAndEpisodeGroup(season int, ep int) string {
	if season == 0 {
		season = 1
	}
	return fmt.Sprintf(`"s%se%s"`, zeropad(season), zeropad(ep))
}

// (01|e01|e01v|ep01|ep1)
func getEpisodeGroup(ep int) string {
	pEp := zeropad(ep)
	//return fmt.Sprintf(`("%s"|"e%s"|"e%sv"|"%sv"|"ep%s"|"ep%d")`, pEp, pEp, pEp, pEp, pEp, ep)
	return fmt.Sprintf(`(%s|e%s|e%sv|%sv|ep%s|ep%d)`, pEp, pEp, pEp, pEp, pEp, ep)
}

// (season 1|season 01|s1|s01)
func getSeasonGroup(season int) string {
	// Season section
	seasonBuff := bytes.NewBufferString("")
	// e.g. S1, season 1, season 01
	if season != 0 {
		seasonBuff.WriteString(fmt.Sprintf(`("%s%d"|`, "season ", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s"|`, "season ", zeropad(season)))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%d"|`, "s", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s")`, "s", zeropad(season)))
		//seasonBuff.WriteString(fmt.Sprintf(`(%s%d|`, "season ", season))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%s|`, "season ", zeropad(season)))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%d|`, "s", season))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%s)`, "s", zeropad(season)))
	}
	return seasonBuff.String()
}
func getPartGroup(part int) string {
	partBuff := bytes.NewBufferString("")
	if part != 0 {
		partBuff.WriteString(fmt.Sprintf(`("%s%d")`, "part ", part))
	}
	return partBuff.String()
}

func getBatchGroup(m *anilist.BaseMedia) string {
	buff := bytes.NewBufferString("")
	buff.WriteString("(")
	// e.g. 01-12
	s1 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " - ", zeropad(m.GetTotalEpisodeCount()))
	buff.WriteString(s1)
	buff.WriteString("|")
	// e.g. 01~12
	s2 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " ~ ", zeropad(m.GetTotalEpisodeCount()))
	buff.WriteString(s2)
	buff.WriteString("|")
	// e.g. 01~12
	buff.WriteString(`"Batch"|`)
	buff.WriteString(`"Complete"|`)
	buff.WriteString(`"+ OVA"|`)
	buff.WriteString(`"+ Specials"|`)
	buff.WriteString(`"+ Special"|`)
	buff.WriteString(`"Seasons"|`)
	buff.WriteString(`"Parts"`)
	buff.WriteString(")")
	return buff.String()
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
