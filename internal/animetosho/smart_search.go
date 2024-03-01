package animetosho

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/seanime-parser"
	"github.com/sourcegraph/conc/pool"
	"net/url"
	"strings"
	"time"
)

type (
	BuildSearchQueryOptions struct {
		Title          *string
		Media          *anilist.BaseMedia
		Batch          *bool
		EpisodeNumber  *int
		AbsoluteOffset *int
		Resolution     *string
		Cache          *SearchCache
	}
)

func SearchQuery(opts *BuildSearchQueryOptions) (torrents []*Torrent, err error) {

	romTitle := opts.Media.GetRomajiTitleSafe()
	engTitle := opts.Media.GetTitleSafe()
	isBatch := opts.Batch != nil && *opts.Batch

	season := 0

	// create titles by extracting season/part info
	titles := make([]string, 0)
	for _, title := range opts.Media.GetAllTitles() {
		s, cTitle := util.ExtractSeasonNumber(*title)
		if s != 0 { // update season if it got parsed
			season = s
		}
		if cTitle != "" { // add "cleaned" titles
			titles = append(titles, cTitle)
		}
	}

	// Check season from synonyms, only update season if it's still 0
	for _, synonym := range opts.Media.Synonyms {
		s, _ := util.ExtractSeasonNumber(*synonym)
		if s != 0 && season == 0 {
			season = s
		}
	}

	// add romaji and english titles to the title list
	titles = append(titles, romTitle)
	if len(engTitle) > 0 {
		titles = append(titles, engTitle)
	}

	// convert III and II to season
	// these will get cleaned later
	if season == 0 && (strings.Contains(strings.ToLower(romTitle), " iii") || strings.Contains(strings.ToLower(engTitle), " iii")) {
		season = 3
	}
	if season == 0 && (strings.Contains(strings.ToLower(romTitle), " ii") || strings.Contains(strings.ToLower(engTitle), " ii")) {
		season = 2
	}

	// also, split romaji title by colon,
	// if first part is long enough, add it to the title list
	// DEVNOTE maybe we should only do that if the season IS found
	split := strings.Split(romTitle, ":")
	if len(split) > 1 && len(split[0]) > 8 {
		titles = append(titles, split[0])
	}

	// clean titles
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

	shortestTitle := ""
	for _, title := range titles {
		if shortestTitle == "" || len(title) < len(shortestTitle) {
			shortestTitle = title
		}
	}

	/////////////////////// Season
	seasonBuff := bytes.NewBufferString("")
	if season > 0 {
		// (season 1|season 01|s1|s01)
		// Season section
		// e.g. S1, season 1, season 01
		seasonBuff.WriteString(fmt.Sprintf(`"%s %s%d" | `, shortestTitle, "season ", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s %s%s" | `, shortestTitle, "season ", zeropad(season)))
		seasonBuff.WriteString(fmt.Sprintf(`"%s %s%d" | `, shortestTitle, "s", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s %s%s"`, shortestTitle, "s", zeropad(season)))
	}

	//////////////////////// Episode
	episodeStr := ""
	if opts.EpisodeNumber != nil {
		pEp := zeropad(*opts.EpisodeNumber)

		episodeStr = fmt.Sprintf(`(%s|e%d)`, pEp, *opts.EpisodeNumber)
	}

	queryStr := ""

	qTitles := "("
	for idx, title := range titles {
		qTitles += "\"" + title + "\"" + " | "
		if idx == len(titles)-1 {
			qTitles = qTitles[:len(qTitles)-3]
		}
	}
	qTitles += seasonBuff.String()
	qTitles += ")"

	// 1. Add titles
	queryStr += qTitles
	// 2. Add episodes
	if episodeStr != "" {
		queryStr += " " + episodeStr
	}
	// 3. Add resolution
	if opts.Resolution != nil {
		queryStr += " " + *opts.Resolution
	}

	finalStr := queryStr

	metadata := seanime_parser.Parse(opts.Media.GetRomajiTitleSafe())

	// 1. Add title
	absoluteEpStr := metadata.Title
	if opts.AbsoluteOffset != nil {
		// 2. Add episodes
		ep := *opts.EpisodeNumber + *opts.AbsoluteOffset
		absoluteEpStr += fmt.Sprintf(` (%d|e%d|ep%d)`, ep, ep, ep)
		// 3. Add resolution
		if opts.Resolution != nil {
			absoluteEpStr += " " + *opts.Resolution
		}
		finalStr = fmt.Sprintf("(%s) | (%s)", absoluteEpStr, queryStr) // Override finalStr
	}

	format := "?only_tor=1&q=%s&qx=1&filter[0][t]=nyaa_class&order=size-a"
	query := fmt.Sprintf(format, url.QueryEscape(finalStr))

	if isBatch {
		// Batch search
		// e.g. "Title [1080p][Batch]"
		finalStr = fmt.Sprintf(`("%s" | "%s")`, engTitle, romTitle)
		finalStr += " " + getBatchGroup(opts.Media)
		if opts.Resolution != nil {
			finalStr += " " + *opts.Resolution
		}
		query = fmt.Sprintf("?only_tor=1&q=%s&qx=1&filter[0][t]=nyaa_class&order=size-d", url.QueryEscape(finalStr))
	}

	//check cache
	if opts.Cache != nil {
		cacheRes, found := opts.Cache.Get(finalStr)
		if found {
			return cacheRes, nil
		}
	}

	torrents, err = fetchTorrents(query)
	if err != nil {
		return nil, err
	}

	// Filter torrents
	p := pool.NewWithResults[*Torrent]()
	for _, torrent := range torrents {
		p.Go(func() *Torrent {
			m := seanime_parser.Parse(torrent.Title)
			if isBatch {
				if len(m.EpisodeNumber) < 2 && !comparison.ValueContainsBatchKeywords(torrent.Title) {
					return nil
				}
			} else if opts.EpisodeNumber != nil {
				if len(m.EpisodeNumber) == 1 {
					ep, _ := util.StringToInt(m.EpisodeNumber[0])
					if ep != *opts.EpisodeNumber && ep != *opts.EpisodeNumber+*opts.AbsoluteOffset {
						return nil
					}
				} else {
					return nil
				}
			}
			return torrent
		})
	}
	res := p.Wait()
	torrents = lo.Filter(res, func(i *Torrent, _ int) bool {
		return i != nil
	})

	// Add to the cache
	opts.Cache.SetT(finalStr, torrents, time.Minute)

	return
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
