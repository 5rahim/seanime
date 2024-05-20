package animetosho

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
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
		Logger         *zerolog.Logger
	}
)

func SearchQuery(opts *BuildSearchQueryOptions) (torrents []*Torrent, err error) {
	format := "?only_tor=1&q=%s&qx=1&filter[0][t]=nyaa_class&order=size-a"

	isBatch := opts.Batch != nil && *opts.Batch
	hasSingleEpisode := opts.Media.IsMovieOrSingleEpisode()

	urlQuery := ""
	finalQueryStr := "" // Final search query string, used for caching

	switch hasSingleEpisode {
	case true:

		queryStr := ""

		allTitles := opts.Media.GetAllTitles()
		// 1. Build a query string
		qTitles := "("
		for _, title := range allTitles {
			qTitles += fmt.Sprintf("%s | ", *title)
		}
		qTitles = qTitles[:len(qTitles)-3] + ")"

		queryStr += qTitles

		// 2. Add resolution
		if opts.Resolution != nil {
			queryStr += " " + *opts.Resolution
		}

		finalQueryStr = queryStr
	case false:
		switch isBatch {
		// Single episode search
		case false:

			qTitles := opts.buildTitles()
			qEpisodes := opts.buildEpisodes()

			queryStr := ""
			// 1. Add titles
			queryStr += qTitles
			// 2. Add episodes
			if qEpisodes != "" {
				queryStr += " " + qEpisodes
			}
			// 3. Add resolution
			if opts.Resolution != nil {
				queryStr += " " + *opts.Resolution
			}

			finalQueryStr = queryStr

			// If we can also search for absolute episodes (there is an offset)
			if opts.AbsoluteOffset != nil && *opts.AbsoluteOffset > 0 {
				// Parse a good title
				metadata := seanime_parser.Parse(opts.Media.GetRomajiTitleSafe())
				// 1. Start building a new query string
				absoluteQueryStr := metadata.Title
				// 2. Add episodes
				ep := *opts.EpisodeNumber + *opts.AbsoluteOffset
				absoluteQueryStr += fmt.Sprintf(` ("%d"|"e%d"|"ep%d")`, ep, ep, ep)
				// 3. Add resolution
				if opts.Resolution != nil {
					absoluteQueryStr += " " + *opts.Resolution
				}
				// Overwrite finalQueryStr by adding the absolute query string
				finalQueryStr = fmt.Sprintf("(%s) | (%s)", absoluteQueryStr, queryStr)
			}

		case true:
			// Batch search
			// e.g. "Title [1080p][Batch]"
			romTitle := opts.Media.GetRomajiTitleSafe()
			engTitle := opts.Media.GetTitleSafe()
			finalQueryStr = fmt.Sprintf(`(%s | %s)`, engTitle, romTitle)
			finalQueryStr += " " + getBatchGroup(opts.Media)
			if opts.Resolution != nil {
				finalQueryStr += " " + *opts.Resolution
			}
		}
	}

	cacheKey := finalQueryStr + map[bool]string{true: "+batch", false: ""}[*opts.Batch]

	// Check cache
	if opts.Cache != nil {
		cacheRes, found := opts.Cache.Get(cacheKey)
		if found {
			opts.Logger.Debug().Str("query", finalQueryStr).Msgf("animetosho: Cache HIT")
			return cacheRes, nil
		}
	}

	opts.Logger.Debug().Str("query", finalQueryStr).Msgf("animetosho: Cache MISS")

	urlQuery = fmt.Sprintf(format, url.QueryEscape(finalQueryStr))
	torrents, err = fetchTorrents(urlQuery)
	torrentMap := make(map[string]*Torrent)
	for _, t := range torrents {
		torrentMap[t.Title] = t
	}

	urlQuery = fmt.Sprintf(format, url.QueryEscape(finalQueryStr+" -S0"))
	other, _ := fetchTorrents(urlQuery)
	for _, t := range other {
		if _, ok := torrentMap[t.Title]; !ok {
			torrents = append(torrents, t)
		}
	}

	opts.Logger.Debug().Msgf("animetosho: Fetched %d torrents", len(torrents))
	if err != nil {
		return nil, err
	}

	// Filter torrents
	p := pool.NewWithResults[*Torrent]()
	for _, torrent := range torrents {
		p.Go(func() *Torrent {
			m := seanime_parser.Parse(torrent.Title)
			// When we're looking for batches
			// we only want to return torrents that are actually batches (more than 1 episode or contain batch keywords)
			if isBatch {
				if len(m.EpisodeNumber) < 2 && !comparison.ValueContainsBatchKeywords(torrent.Title) {
					return nil
				}
			} else if opts.EpisodeNumber != nil { // We're looking for a single episode
				// If more than one episode, skip it
				if len(m.EpisodeNumber) > 1 || comparison.ValueContainsBatchKeywords(torrent.Title) {
					return nil
				}
				if len(m.EpisodeNumber) == 1 {
					// If the episode number is not the one we're looking for, skip it
					ep, ok := util.StringToInt(m.EpisodeNumber[0])
					if ok && ep != *opts.EpisodeNumber && ep != *opts.EpisodeNumber+*opts.AbsoluteOffset {
						return nil
					}
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
	opts.Cache.SetT(cacheKey, torrents, time.Minute)

	return
}

func (opts *BuildSearchQueryOptions) buildEpisodes() string {
	episodeStr := ""
	if opts.EpisodeNumber != nil {
		pEp := zeropad(*opts.EpisodeNumber)

		episodeStr = fmt.Sprintf(`("%s"|"e%d") -S0`, pEp, *opts.EpisodeNumber)
	}
	return episodeStr
}

func (opts *BuildSearchQueryOptions) buildTitles() string {

	romTitle := opts.Media.GetRomajiTitleSafe()
	engTitle := opts.Media.GetTitleSafe()

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
	split = strings.Split(engTitle, ":")
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

	qTitles := "("
	for idx, title := range titles {
		qTitles += "\"" + title + "\"" + " | "
		if idx == len(titles)-1 {
			qTitles = qTitles[:len(qTitles)-3]
		}
	}
	qTitles += seasonBuff.String()
	qTitles += ")"

	return qTitles
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
