package nyaa

import (
	"bytes"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/result"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"strings"
	"time"
)

// https://github.com/irevenko/go-nyaa

type (
	SearchOptions struct {
		Provider string
		Query    string
		Category string
		SortBy   string
		Filter   string
		Cache    *SearchCache // optional
	}

	SearchMultipleOptions struct {
		Provider string
		Query    []string
		Category string
		SortBy   string
		Filter   string
		Cache    *SearchCache // optional
	}
	BuildSearchQueryOptions struct {
		Title          *string
		Media          *anilist.BaseMedia
		Batch          *bool
		EpisodeNumber  *int
		AbsoluteOffset *int
		Resolution     *string
	}

	SearchCache struct {
		*result.Cache[string, []*DetailedTorrent]
	}
)

func NewSearchCache() *SearchCache {
	return &SearchCache{result.NewCache[string, []*DetailedTorrent]()}
}

func Search(opts SearchOptions) ([]*DetailedTorrent, error) {

	fp := gofeed.NewParser()

	if opts.Cache != nil {
		//check cache
		cacheRes, found := opts.Cache.Get(opts.Query)
		if found {
			return cacheRes, nil
		}
	}

	// create search url
	url, err := buildURL(opts)
	if err != nil {
		return nil, err
	}

	// get content
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	// parse content
	res := convertRSS(feed)

	ret := make([]*DetailedTorrent, 0)
	for _, torrent := range res {
		ret = append(ret, torrent.toDetailedTorrent())
	}

	// add to the cache
	opts.Cache.SetT(opts.Query, ret, time.Minute)

	return ret, nil
}

func SearchMultiple(opts SearchMultipleOptions) ([]*DetailedTorrent, error) {

	fp := gofeed.NewParser()

	p := pool.NewWithResults[[]*DetailedTorrent]()
	for _, query := range opts.Query {
		query := query
		p.Go(func() []*DetailedTorrent {

			//check cache
			if opts.Cache != nil {
				cacheRes, found := opts.Cache.Get(query)
				if found {
					return cacheRes
				}
			}

			// create the url
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

			// get content
			feed, err := fp.ParseURL(url)
			if err != nil {
				return nil
			}

			// parse content
			conv := convertRSS(feed)

			// convert to detailed torrents
			ret := make([]*DetailedTorrent, 0)
			for _, torrent := range conv {
				ret = append(ret, torrent.toDetailedTorrent())
			}

			// add to the cache
			opts.Cache.SetT(query, ret, time.Minute)

			return ret
		})
	}
	slicesSlice := p.Wait()
	slicesSlice = lo.Filter(slicesSlice, func(i []*DetailedTorrent, _ int) bool {
		return i != nil
	})
	res := lo.Flatten(slicesSlice)

	return res, nil
}

// BuildSearchQuery will return a slice of queries for nyaa.si.
// The second index of the returned slice is the absolute episode query.
// If the function returns false, the query could not be built.
// BuildSearchQueryOptions.Title will override the constructed title query but not other parameters.
func BuildSearchQuery(opts *BuildSearchQueryOptions) ([]string, bool) {

	if opts.Media == nil || opts.Batch == nil || opts.EpisodeNumber == nil || opts.AbsoluteOffset == nil || opts.Resolution == nil {
		return make([]string, 0), false
	}

	_ = *opts.EpisodeNumber
	romTitle := opts.Media.GetRomajiTitleSafe()
	engTitle := opts.Media.GetTitleSafe()

	season := 0
	part := 0

	// create titles by extracting season/part info
	titles := make([]string, 0)
	for _, title := range opts.Media.GetAllTitles() {
		s, cTitle := util.ExtractSeasonNumber(*title)
		p, cTitle := util.ExtractPartNumber(cTitle)
		if s != 0 { // update season if it got parsed
			season = s
		}
		if p != 0 { // update part if it got parsed
			part = p
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

	// no season or part got parsed, meaning there is no "cleaned" title,
	// add romaji and english titles to the title list
	if season == 0 && part == 0 {
		titles = append(titles, romTitle)
		if len(engTitle) > 0 {
			titles = append(titles, engTitle)
		}
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

	//
	// Parameters
	//

	// can batch if media stopped airing
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

	//println(spew.Sdump(titleStr, batchStr, normalStr))

	query := fmt.Sprintf("%s%s%s%s", titleStr, batchStr, normalStr, *opts.Resolution)
	query2 := ""

	// Absolute episode addition
	if !*opts.Batch && *opts.AbsoluteOffset > 0 {
		query2 = fmt.Sprintf("%s%s", getAbsoluteGroup(titleStr, opts), *opts.Resolution) // e.g. jujutsu kaisen 25
	}

	// Movie addition
	// We add this because the first query might be invalid because of inclusion of "ep01" etc...
	if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
		query2 = titleStr
	}

	ret := []string{query}
	if query2 != "" {
		ret = append(ret, query2)
	}

	return ret, true
}
