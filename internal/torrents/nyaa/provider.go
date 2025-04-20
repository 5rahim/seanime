package nyaa

import (
	"bytes"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/habari"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	NyaaProviderName = "nyaa"
)

type Provider struct {
	logger   *zerolog.Logger
	category string

	baseUrl string
}

func NewProvider(logger *zerolog.Logger, category string) hibiketorrent.AnimeProvider {
	return &Provider{
		logger:   logger,
		category: category,
	}
}

func (n *Provider) SetSavedUserConfig(config extension.SavedUserConfig) {
	n.baseUrl, _ = config.Values["apiUrl"]
}

func (n *Provider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: true,
		SmartSearchFilters: []hibiketorrent.AnimeProviderSmartSearchFilter{
			hibiketorrent.AnimeProviderSmartSearchFilterBatch,
			hibiketorrent.AnimeProviderSmartSearchFilterEpisodeNumber,
			hibiketorrent.AnimeProviderSmartSearchFilterResolution,
			hibiketorrent.AnimeProviderSmartSearchFilterQuery,
		},
		SupportsAdult: false,
	}
}

func (n *Provider) GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	fp := gofeed.NewParser()

	url, err := buildURL(n.baseUrl, BuildURLOptions{
		Provider: "nyaa",
		Query:    "",
		Category: n.category,
		SortBy:   "seeders",
		Filter:   "",
	})
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

	ret = torrentSliceToAnimeTorrentSlice(res, NyaaProviderName)

	return
}

func (n *Provider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	fp := gofeed.NewParser()

	n.logger.Trace().Str("query", opts.Query).Msg("nyaa: Search query")

	url, err := buildURL(n.baseUrl, BuildURLOptions{
		Provider: "nyaa",
		Query:    opts.Query,
		Category: n.category,
		SortBy:   "seeders",
		Filter:   "",
	})
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

	ret = torrentSliceToAnimeTorrentSlice(res, NyaaProviderName)

	return
}

func (n *Provider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {

	queries, ok := buildSmartSearchQueries(&opts)
	if !ok {
		return nil, fmt.Errorf("could not build queries")
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, query := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			fp := gofeed.NewParser()
			n.logger.Trace().Str("query", query).Msg("nyaa: Smart search query")
			url, err := buildURL(n.baseUrl, BuildURLOptions{
				Provider: "nyaa",
				Query:    query,
				Category: n.category,
				SortBy:   "seeders",
				Filter:   "",
			})
			if err != nil {
				return
			}
			n.logger.Trace().Str("url", url).Msg("nyaa: Smart search url")
			// get content
			feed, err := fp.ParseURL(url)
			if err != nil {
				return
			}
			// parse content
			res := convertRSS(feed)

			mu.Lock()
			ret = torrentSliceToAnimeTorrentSlice(res, NyaaProviderName)
			mu.Unlock()
		}(query)
	}
	wg.Wait()

	// remove duplicates
	lo.UniqBy(ret, func(i *hibiketorrent.AnimeTorrent) string {
		return i.Link
	})

	if !opts.Batch {
		// Single-episode search
		// If the episode number is provided, we can filter the results
		ret = lo.Filter(ret, func(i *hibiketorrent.AnimeTorrent, _ int) bool {
			relEp := i.EpisodeNumber
			if relEp == -1 {
				return false
			}
			absEp := opts.Media.AbsoluteSeasonOffset + opts.EpisodeNumber

			return opts.EpisodeNumber == relEp || absEp == relEp
		})
	}

	return
}

func (n *Provider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.InfoHash, nil
}

func (n *Provider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return TorrentMagnet(torrent.Link)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ADVANCED SEARCH
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// buildSmartSearchQueries will return a slice of queries for nyaa.si.
// The second index of the returned slice is the absolute episode query.
// If the function returns false, the query could not be built.
// BuildSearchQueryOptions.Title will override the constructed title query but not other parameters.
func buildSmartSearchQueries(opts *hibiketorrent.AnimeSmartSearchOptions) ([]string, bool) {

	romTitle := opts.Media.RomajiTitle
	engTitle := opts.Media.EnglishTitle

	allTitles := []*string{&romTitle, engTitle}
	for _, synonym := range opts.Media.Synonyms {
		allTitles = append(allTitles, &synonym)
	}

	season := 0
	part := 0

	// create titles by extracting season/part info
	titles := make([]string, 0)

	// Build titles if no query provided
	if opts.Query == "" {
		for _, title := range allTitles {
			if title == nil {
				continue
			}
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
			s, _ := util.ExtractSeasonNumber(synonym)
			if s != 0 && season == 0 {
				season = s
			}
		}

		// no season or part got parsed, meaning there is no "cleaned" title,
		// add romaji and english titles to the title list
		if season == 0 && part == 0 {
			titles = append(titles, romTitle)
			if engTitle != nil {
				if len(*engTitle) > 0 {
					titles = append(titles, *engTitle)
				}
			}
		}

		// convert III and II to season
		// these will get cleaned later
		if season == 0 && (strings.Contains(strings.ToLower(romTitle), " iii")) {
			season = 3
		}
		if season == 0 && (strings.Contains(strings.ToLower(romTitle), " ii")) {
			season = 2
		}
		if engTitle != nil {
			if season == 0 && (strings.Contains(strings.ToLower(*engTitle), " iii")) {
				season = 3
			}
			if season == 0 && (strings.Contains(strings.ToLower(*engTitle), " ii")) {
				season = 2
			}
		}

		// also, split romaji title by colon,
		// if first part is long enough, add it to the title list
		// DEVNOTE maybe we should only do that if the season IS found
		split := strings.Split(romTitle, ":")
		if len(split) > 1 && len(split[0]) > 8 {
			titles = append(titles, split[0])
		}
		if engTitle != nil {
			split := strings.Split(*engTitle, ":")
			if len(split) > 1 && len(split[0]) > 8 {
				titles = append(titles, split[0])
			}
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
	} else {
		titles = append(titles, strings.ToLower(opts.Query))
	}

	//
	// Parameters
	//

	// can batch if media stopped airing
	canBatch := false
	if opts.Media.Status == string(anilist.MediaStatusFinished) && opts.Media.EpisodeCount > 0 {
		canBatch = true
	}

	normalBuff := bytes.NewBufferString("")

	// Batch section - empty unless:
	// 1. If the media is finished and has more than 1 episode
	// 2. If the media is not a movie
	// 3. If the media is not a single episode
	batchBuff := bytes.NewBufferString("")
	if opts.Batch && canBatch && !(opts.Media.Format == string(anilist.MediaFormatMovie) && opts.Media.EpisodeCount == 1) {
		if season != 0 {
			batchBuff.WriteString(buildSeasonString(season))
		}
		if part != 0 {
			batchBuff.WriteString(buildPartString(part))
		}
		batchBuff.WriteString(buildBatchString(&opts.Media))

	} else {

		normalBuff.WriteString(buildSeasonString(season))
		if part != 0 {
			normalBuff.WriteString(buildPartString(part))
		}

		if !(opts.Media.Format == string(anilist.MediaFormatMovie) && opts.Media.EpisodeCount == 1) {
			normalBuff.WriteString(buildEpisodeString(opts.EpisodeNumber))
		}

	}

	titleStr := buildTitleString(titles)
	batchStr := batchBuff.String()
	normalStr := normalBuff.String()

	// Replace titleStr if user provided one
	if opts.Query != "" {
		titleStr = fmt.Sprintf(`(%s)`, opts.Query)
	}

	//println(spew.Sdump(titleStr, batchStr, normalStr))

	query := fmt.Sprintf("%s%s%s", titleStr, batchStr, normalStr)
	if opts.Resolution != "" {
		query = fmt.Sprintf("%s(%s)", query, opts.Resolution)
	} else {
		query = fmt.Sprintf("%s(%s)", query, strings.Join([]string{"360", "480", "720", "1080"}, "|"))
	}
	query2 := ""

	// Absolute episode addition
	if !opts.Batch && opts.Media.AbsoluteSeasonOffset > 0 && !(opts.Media.Format == string(anilist.MediaFormatMovie) && opts.Media.EpisodeCount == 1) {
		query2 = fmt.Sprintf("%s", buildAbsoluteGroupString(titleStr, opts.Resolution, opts)) // e.g. jujutsu kaisen 25
	}

	ret := []string{query}
	if query2 != "" {
		ret = append(ret, query2)
	}

	return ret, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//func sanitizeTitle(t string) string {
//	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(t, "!", ""), ":", ""), "[", ""), "]", ""), ".", "")
//}

// (title)
// ("jjk"|"jujutsu kaisen")
func buildTitleString(titles []string) string {
	// Single titles are not wrapped in quotes
	if len(titles) == 1 {
		return fmt.Sprintf(`(%s)`, titles[0])
	}

	return fmt.Sprintf("(%s)", strings.Join(lo.Map(titles, func(item string, _ int) string {
		return fmt.Sprintf(`"%s"`, item)
	}), "|"))
}

func buildAbsoluteGroupString(title, resolution string, opts *hibiketorrent.AnimeSmartSearchOptions) string {
	return fmt.Sprintf("%s(%d)(%s)", title, opts.EpisodeNumber+opts.Media.AbsoluteSeasonOffset, resolution)
}

// (s01e01)
func buildSeasonAndEpisodeGroup(season int, ep int) string {
	if season == 0 {
		season = 1
	}
	return fmt.Sprintf(`"s%se%s"`, zeropad(season), zeropad(ep))
}

// (01|e01|e01v|ep01|ep1)
func buildEpisodeString(ep int) string {
	pEp := zeropad(ep)
	//return fmt.Sprintf(`("%s"|"e%s"|"e%sv"|"%sv"|"ep%s"|"ep%d")`, pEp, pEp, pEp, pEp, pEp, ep)
	return fmt.Sprintf(`(%s|e%s|e%sv|%sv|ep%s|ep%d)`, pEp, pEp, pEp, pEp, pEp, ep)
}

// (season 1|season 01|s1|s01)
func buildSeasonString(season int) string {
	// Season section
	seasonBuff := bytes.NewBufferString("")
	// e.g. S1, season 1, season 01
	if season != 0 {
		seasonBuff.WriteString(fmt.Sprintf(`("%s%d"|`, "season ", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s"|`, "season ", zeropad(season)))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%d"|`, "s", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s")`, "s", zeropad(season)))
	}
	return seasonBuff.String()
}

func buildPartString(part int) string {
	partBuff := bytes.NewBufferString("")
	if part != 0 {
		partBuff.WriteString(fmt.Sprintf(`("%s%d")`, "part ", part))
	}
	return partBuff.String()
}

func buildBatchString(m *hibiketorrent.Media) string {

	buff := bytes.NewBufferString("")
	buff.WriteString("(")
	// e.g. 01-12
	s1 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " - ", zeropad(m.EpisodeCount))
	buff.WriteString(s1)
	buff.WriteString("|")
	// e.g. 01~12
	s2 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " ~ ", zeropad(m.EpisodeCount))
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func convertRSS(feed *gofeed.Feed) []Torrent {
	var res []Torrent

	for _, item := range feed.Items {
		res = append(
			res,
			Torrent{
				Name:        item.Title,
				Link:        item.Link,
				Date:        item.Published,
				Description: item.Description,
				GUID:        item.GUID,
				Comments:    item.Extensions["nyaa"]["comments"][0].Value,
				IsTrusted:   item.Extensions["nyaa"]["trusted"][0].Value,
				IsRemake:    item.Extensions["nyaa"]["remake"][0].Value,
				Size:        item.Extensions["nyaa"]["size"][0].Value,
				Seeders:     item.Extensions["nyaa"]["seeders"][0].Value,
				Leechers:    item.Extensions["nyaa"]["leechers"][0].Value,
				Downloads:   item.Extensions["nyaa"]["downloads"][0].Value,
				Category:    item.Extensions["nyaa"]["category"][0].Value,
				CategoryID:  item.Extensions["nyaa"]["categoryId"][0].Value,
				InfoHash:    item.Extensions["nyaa"]["infoHash"][0].Value,
			},
		)
	}
	return res
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func torrentSliceToAnimeTorrentSlice(torrents []Torrent, providerName string) []*hibiketorrent.AnimeTorrent {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	ret := make([]*hibiketorrent.AnimeTorrent, 0)
	for _, torrent := range torrents {
		wg.Add(1)
		go func(torrent Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, torrent.toAnimeTorrent(providerName))
			mu.Unlock()
		}(torrent)
	}
	wg.Wait()

	return ret
}

func (t *Torrent) toAnimeTorrent(providerName string) *hibiketorrent.AnimeTorrent {
	metadata := habari.Parse(t.Name)

	seeders, _ := strconv.Atoi(t.Seeders)
	leechers, _ := strconv.Atoi(t.Leechers)
	downloads, _ := strconv.Atoi(t.Downloads)

	formattedDate := ""
	parsedDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", t.Date)
	if err == nil {
		formattedDate = parsedDate.Format(time.RFC3339)
	}

	ret := &hibiketorrent.AnimeTorrent{
		Name:          t.Name,
		Date:          formattedDate,
		Size:          t.GetSizeInBytes(),
		FormattedSize: t.Size,
		Seeders:       seeders,
		Leechers:      leechers,
		DownloadCount: downloads,
		Link:          t.GUID,
		DownloadUrl:   t.Link,
		InfoHash:      t.InfoHash,
		MagnetLink:    "",    // Should be scraped
		Resolution:    "",    // Should be parsed
		IsBatch:       false, // Should be parsed
		EpisodeNumber: -1,    // Should be parsed
		ReleaseGroup:  "",    // Should be parsed
		Provider:      providerName,
		IsBestRelease: false,
		Confirmed:     false,
	}

	isBatchByGuess := false
	episode := -1

	if len(metadata.EpisodeNumber) > 1 || comparison.ValueContainsBatchKeywords(t.Name) {
		isBatchByGuess = true
	}
	if len(metadata.EpisodeNumber) == 1 {
		episode = util.StringToIntMust(metadata.EpisodeNumber[0])
	}

	ret.Resolution = metadata.VideoResolution
	ret.ReleaseGroup = metadata.ReleaseGroup

	// Only change batch status if it wasn't already 'true'
	if ret.IsBatch == false && isBatchByGuess {
		ret.IsBatch = true
	}

	ret.EpisodeNumber = episode

	return ret
}
