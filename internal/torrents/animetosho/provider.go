package animetosho

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/util"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/habari"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

var (
	JsonFeedUrl  = util.Decode("aHR0cHM6Ly9mZWVkLmFuaW1ldG9zaG8ub3JnL2pzb24=")
	ProviderName = "animetosho"
)

type (
	Provider struct {
		logger         *zerolog.Logger
		sneedexNyaaIDs map[int]struct{}
	}
)

func NewProvider(logger *zerolog.Logger) hibiketorrent.AnimeProvider {
	ret := &Provider{
		logger:         logger,
		sneedexNyaaIDs: make(map[int]struct{}),
	}

	go ret.loadSneedex()

	return ret
}

func (at *Provider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: true,
		SmartSearchFilters: []hibiketorrent.AnimeProviderSmartSearchFilter{
			hibiketorrent.AnimeProviderSmartSearchFilterBatch,
			hibiketorrent.AnimeProviderSmartSearchFilterEpisodeNumber,
			hibiketorrent.AnimeProviderSmartSearchFilterResolution,
			hibiketorrent.AnimeProviderSmartSearchFilterBestReleases,
		},
		SupportsAdult: false,
	}
}

// GetLatest returns all the latest torrents currently visible on the site
func (at *Provider) GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	at.logger.Debug().Msg("animetosho: Fetching latest torrents")
	query := "?q="
	torrents, err := at.fetchTorrents(query)
	if err != nil {
		return nil, err
	}

	ret = at.torrentSliceToAnimeTorrentSlice(torrents, false, &hibiketorrent.Media{})

	return ret, nil
}

func (at *Provider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	at.logger.Debug().Str("query", opts.Query).Msg("animetosho: Searching for torrents")
	query := fmt.Sprintf("?q=%s", url.QueryEscape(sanitizeTitle(opts.Query)))
	atTorrents, err := at.fetchTorrents(query)
	if err != nil {
		return nil, err
	}

	ret = at.torrentSliceToAnimeTorrentSlice(atTorrents, false, &opts.Media)

	return ret, nil
}

func (at *Provider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	if opts.BestReleases {
		return at.smartSearchBestReleases(&opts)
	}
	if opts.Batch {
		return at.smartSearchBatch(&opts)
	}
	return at.smartSearchSingleEpisode(&opts)
}

func (at *Provider) smartSearchSingleEpisode(opts *hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	ret = make([]*hibiketorrent.AnimeTorrent, 0)

	at.logger.Debug().Int("aid", opts.AnidbAID).Msg("animetosho: Searching batches by Episode ID")

	foundByID := false

	atTorrents := make([]*Torrent, 0)

	if opts.AnidbEID > 0 {
		// Get all torrents by Episode ID
		atTorrents, err = at.searchByEID(opts.AnidbEID, opts.Resolution)
		if err != nil {
			return nil, err
		}

		foundByID = true
	}

	if foundByID {
		// Get all torrents with only 1 file
		atTorrents = lo.Filter(atTorrents, func(t *Torrent, _ int) bool {
			return t.NumFiles == 1
		})
		ret = at.torrentSliceToAnimeTorrentSlice(atTorrents, true, &opts.Media)
		return
	}

	at.logger.Debug().Msg("animetosho: Searching batches by query")

	// If we couldn't find batches by AniDB Episode ID, use query builder
	queries := buildSmartSearchQueries(opts)

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, query := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()

			at.logger.Trace().Str("query", query).Msg("animetosho: Searching by query")
			torrents, err := at.fetchTorrents(fmt.Sprintf("?only_tor=1&q=%s&qx=1", url.QueryEscape(query)))
			if err != nil {
				return
			}
			for _, t := range torrents {
				// Skip if torrent has more than 1 file
				if t.NumFiles > 1 && !(opts.Media.Format == string(anilist.MediaFormatMovie) && opts.Media.EpisodeCount == 1) {
					continue
				}
				mu.Lock()
				ret = append(ret, t.toAnimeTorrent(&opts.Media))
				mu.Unlock()
			}
		}(query)
	}

	wg.Wait()

	// Remove duplicates
	lo.UniqBy(ret, func(t *hibiketorrent.AnimeTorrent) string {
		return t.Link
	})

	return
}

func (at *Provider) smartSearchBatch(opts *hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	ret = make([]*hibiketorrent.AnimeTorrent, 0)

	at.logger.Debug().Int("aid", opts.AnidbAID).Msg("animetosho: Searching batches by Anime ID")

	foundByID := false

	atTorrents := make([]*Torrent, 0)

	if opts.AnidbAID > 0 {
		// Get all torrents by Anime ID
		atTorrents, err = at.searchByAID(opts.AnidbAID, opts.Resolution)
		if err != nil {
			return nil, err
		}

		// Retain batches ONLY if the media is NOT a movie or single-episode
		// i.e. if the media is a movie or single-episode return all torrents
		if !(opts.Media.Format == string(anilist.MediaFormatMovie) || opts.Media.EpisodeCount == 1) {
			batchTorrents := lo.Filter(atTorrents, func(t *Torrent, _ int) bool {
				return t.NumFiles > 1
			})
			if len(batchTorrents) > 0 {
				atTorrents = batchTorrents
			}
		}

		if len(atTorrents) > 0 {
			foundByID = true
		}
	}

	if foundByID {
		ret = at.torrentSliceToAnimeTorrentSlice(atTorrents, true, &opts.Media)
		return
	}

	at.logger.Debug().Msg("animetosho: Searching batches by query")

	// If we couldn't find batches by AniDB Anime ID, use query builder
	queries := buildSmartSearchQueries(opts)

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, query := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()

			at.logger.Trace().Str("query", query).Msg("animetosho: Searching by query")
			torrents, err := at.fetchTorrents(fmt.Sprintf("?only_tor=1&q=%s&order=size-d", url.QueryEscape(query)))
			if err != nil {
				return
			}
			for _, t := range torrents {
				// Skip if not batch only if the media is not a movie or single-episode
				if t.NumFiles == 1 && !(opts.Media.Format == string(anilist.MediaFormatMovie) && opts.Media.EpisodeCount == 1) {
					continue
				}
				mu.Lock()
				ret = append(ret, t.toAnimeTorrent(&opts.Media))
				mu.Unlock()
			}
		}(query)
	}

	wg.Wait()

	// Remove duplicates
	lo.UniqBy(ret, func(t *hibiketorrent.AnimeTorrent) string {
		return t.Link
	})

	return
}

type sneedexItem struct {
	NyaaIDs []int  `json:"nyaaIDs"`
	EntryID string `json:"entryID"`
}

func (at *Provider) loadSneedex() {
	// Load Sneedex Nyaa IDs
	resp, err := http.Get("https://sneedex.moe/api/public/nyaa")
	if err != nil {
		at.logger.Error().Err(err).Msg("animetosho: Failed to fetch Sneedex Nyaa IDs")
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		at.logger.Error().Err(err).Msg("animetosho: Failed to read Sneedex Nyaa IDs response")
		return
	}

	var sneedexItems []*sneedexItem
	if err := json.Unmarshal(b, &sneedexItems); err != nil {
		at.logger.Error().Err(err).Msg("animetosho: Failed to unmarshal Sneedex Nyaa IDs")
		return
	}

	for _, item := range sneedexItems {
		for _, nyaaID := range item.NyaaIDs {
			at.sneedexNyaaIDs[nyaaID] = struct{}{}
		}
	}

	at.logger.Debug().Int("count", len(at.sneedexNyaaIDs)).Msg("animetosho: Loaded Sneedex Nyaa IDs")
}

func (at *Provider) smartSearchBestReleases(opts *hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	return at.findSneedexBestReleases(opts)
}

func (at *Provider) findSneedexBestReleases(opts *hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	ret := make([]*hibiketorrent.AnimeTorrent, 0)

	at.logger.Debug().Int("aid", opts.AnidbAID).Msg("animetosho: Searching best releases by Anime ID")

	if opts.AnidbAID > 0 {
		// Get all torrents by Anime ID
		atTorrents, err := at.searchByAID(opts.AnidbAID, opts.Resolution)
		if err != nil {
			return nil, err
		}

		// Filter by Sneedex Nyaa IDs
		atTorrents = lo.Filter(atTorrents, func(t *Torrent, _ int) bool {
			_, found := at.sneedexNyaaIDs[t.NyaaId]
			return found
		})

		ret = at.torrentSliceToAnimeTorrentSlice(atTorrents, true, &opts.Media)
	}

	return ret, nil
}

//--------------------------------------------------------------------------------------------------------------------------------------------------//

func (at *Provider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.InfoHash, nil
}

func (at *Provider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.MagnetLink, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func buildSmartSearchQueries(opts *hibiketorrent.AnimeSmartSearchOptions) (ret []string) {

	hasSingleEpisode := opts.Media.EpisodeCount == 1 || opts.Media.Format == string(anilist.MediaFormatMovie)

	var queryStr []string // Final search query string, used for caching
	allTitles := []string{opts.Media.RomajiTitle}
	if opts.Media.EnglishTitle != nil {
		allTitles = append(allTitles, *opts.Media.EnglishTitle)
	}
	for _, title := range opts.Media.Synonyms {
		allTitles = append(allTitles, title)
	}

	//
	// Media only has 1 episode
	//
	if hasSingleEpisode {
		str := ""
		// 1. Build a query string
		qTitles := "("
		for _, title := range allTitles {
			qTitles += fmt.Sprintf("%s | ", sanitizeTitle(title))
		}
		qTitles = qTitles[:len(qTitles)-3] + ")"

		str += qTitles
		// 2. Add resolution
		if opts.Resolution != "" {
			str += " " + opts.Resolution
		}

		// e.g. (Attack on Titan|Shingeki no Kyojin) 1080p
		queryStr = []string{str}

	} else {

		//
		// Media has multiple episodes
		//
		if !opts.Batch { // Single episode search

			qTitles := buildTitleString(opts)
			qEpisodes := buildEpisodeString(opts)

			str := ""
			// 1. Add titles
			str += qTitles
			// 2. Add episodes
			if qEpisodes != "" {
				str += " " + qEpisodes
			}
			// 3. Add resolution
			if opts.Resolution != "" {
				str += " " + opts.Resolution
			}

			queryStr = append(queryStr, str)

			// If we can also search for absolute episodes (there is an offset)
			if opts.Media.AbsoluteSeasonOffset > 0 {
				// Parse a good title
				metadata := habari.Parse(opts.Media.RomajiTitle)
				// 1. Start building a new query string
				absoluteQueryStr := metadata.Title
				// 2. Add episodes
				ep := opts.EpisodeNumber + opts.Media.AbsoluteSeasonOffset
				absoluteQueryStr += fmt.Sprintf(` ("%d"|"e%d"|"ep%d")`, ep, ep, ep)
				// 3. Add resolution
				if opts.Resolution != "" {
					absoluteQueryStr += " " + opts.Resolution
				}
				// Overwrite queryStr by adding the absolute query string
				queryStr = append(queryStr, fmt.Sprintf("(%s) | (%s)", absoluteQueryStr, str))
			}

		} else {

			// Batch search
			// e.g. "(Shingeki No Kyojin | Attack on Titan) ("Batch"|"Complete Series") 1080"
			str := fmt.Sprintf(`(%s)`, opts.Media.RomajiTitle)
			if opts.Media.EnglishTitle != nil {
				str = fmt.Sprintf(`(%s | %s)`, opts.Media.RomajiTitle, *opts.Media.EnglishTitle)
			}
			str += " " + buildBatchGroup(&opts.Media)
			if opts.Resolution != "" {
				str += " " + opts.Resolution
			}

			queryStr = []string{str}
		}

	}

	for _, q := range queryStr {
		ret = append(ret, q)
		ret = append(ret, q+" -S0")
	}

	return
}

// searches for torrents by Anime ID
func (at *Provider) searchByAID(aid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatQuality(quality))
	query := fmt.Sprintf(`?order=size-d&aid=%d&q=%s`, aid, q)
	return at.fetchTorrents(query)
}

// searches for torrents by Episode ID
func (at *Provider) searchByEID(eid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatQuality(quality))
	query := fmt.Sprintf(`?eid=%d&q=%s`, eid, q)
	return at.fetchTorrents(query)
}

func (at *Provider) fetchTorrents(suffix string) (torrents []*Torrent, err error) {
	furl := JsonFeedUrl + suffix

	at.logger.Debug().Str("url", furl).Msg("animetosho: Fetching torrents")

	resp, err := http.Get(furl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch torrents, %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the feed
	var ret []*Torrent
	if err := json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}

	for _, t := range ret {
		if t.Seeders > 100000 {
			t.Seeders = 0
		}
		if t.Leechers > 100000 {
			t.Leechers = 0
		}
	}

	return ret, nil
}

func formatQuality(quality string) string {
	if quality == "" {
		return ""
	}
	quality = strings.TrimSuffix(quality, "p")
	return fmt.Sprintf(`%s`, quality)
}

// sanitizeTitle removes characters that impact the search query
func sanitizeTitle(t string) string {
	// Replace hyphens with spaces
	t = strings.ReplaceAll(t, "-", " ")
	// Remove everything except alphanumeric characters, spaces.
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	t = re.ReplaceAllString(t, "")

	// Trim large spaces
	re2 := regexp.MustCompile(`\s+`)
	t = re2.ReplaceAllString(t, " ")

	// return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(t, "!", ""), ":", ""), "[", ""), "]", "")
	return t
}

func getAllTitles(media *hibiketorrent.Media) []string {
	titles := make([]string, 0)
	titles = append(titles, media.RomajiTitle)
	if media.EnglishTitle != nil {
		titles = append(titles, *media.EnglishTitle)
	}
	for _, title := range media.Synonyms {
		titles = append(titles, title)
	}
	return titles
}

// ("01"|"e01") -S0
func buildEpisodeString(opts *hibiketorrent.AnimeSmartSearchOptions) string {
	episodeStr := ""
	if opts.EpisodeNumber != -1 {
		pEp := zeropad(opts.EpisodeNumber)
		episodeStr = fmt.Sprintf(`("%s"|"e%d") -S0`, pEp, opts.EpisodeNumber)
	}
	return episodeStr
}

func buildTitleString(opts *hibiketorrent.AnimeSmartSearchOptions) string {

	romTitle := sanitizeTitle(opts.Media.RomajiTitle)
	engTitle := ""
	if opts.Media.EnglishTitle != nil {
		engTitle = sanitizeTitle(*opts.Media.EnglishTitle)
	}

	season := 0

	// create titles by extracting season/part info
	titles := make([]string, 0)
	for _, title := range getAllTitles(&opts.Media) {
		s, cTitle := util.ExtractSeasonNumber(title)
		if s != 0 { // update season if it got parsed
			season = s
		}
		if cTitle != "" { // add "cleaned" titles
			titles = append(titles, sanitizeTitle(cTitle))
		}
	}

	// Check season from synonyms, only update season if it's still 0
	for _, synonym := range opts.Media.Synonyms {
		s, _ := util.ExtractSeasonNumber(synonym)
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
	if season == 0 && strings.Contains(strings.ToLower(romTitle), " iii") {
		season = 3
	}
	if season == 0 && strings.Contains(strings.ToLower(romTitle), " ii") {
		season = 2
	}

	if engTitle != "" {
		if season == 0 && strings.Contains(strings.ToLower(engTitle), " iii") {
			season = 3
		}
		if season == 0 && strings.Contains(strings.ToLower(engTitle), " ii") {
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
	if engTitle != "" {
		split = strings.Split(engTitle, ":")
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

func buildBatchGroup(m *hibiketorrent.Media) string {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (at *Provider) torrentSliceToAnimeTorrentSlice(torrents []*Torrent, confirmed bool, media *hibiketorrent.Media) []*hibiketorrent.AnimeTorrent {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	ret := make([]*hibiketorrent.AnimeTorrent, 0)
	for _, torrent := range torrents {
		wg.Add(1)
		go func(torrent *Torrent) {
			defer wg.Done()
			t := torrent.toAnimeTorrent(media)
			_, isBest := at.sneedexNyaaIDs[torrent.NyaaId]
			t.IsBestRelease = isBest
			t.Confirmed = confirmed
			mu.Lock()
			ret = append(ret, t)
			mu.Unlock()
		}(torrent)
	}
	wg.Wait()

	return ret
}

func (t *Torrent) toAnimeTorrent(media *hibiketorrent.Media) *hibiketorrent.AnimeTorrent {
	metadata := habari.Parse(t.Title)

	formattedDate := ""
	parsedDate := time.Unix(int64(t.Timestamp), 0)
	formattedDate = parsedDate.Format(time.RFC3339)

	ret := &hibiketorrent.AnimeTorrent{
		Name:          t.Title,
		Date:          formattedDate,
		Size:          t.TotalSize,
		FormattedSize: util.Bytes(uint64(t.TotalSize)),
		Seeders:       t.Seeders,
		Leechers:      t.Leechers,
		DownloadCount: t.TorrentDownloadCount,
		Link:          t.Link,
		DownloadUrl:   t.TorrentUrl,
		MagnetLink:    t.MagnetUri,
		InfoHash:      t.InfoHash,
		Resolution:    metadata.VideoResolution,
		IsBatch:       t.NumFiles > 1,
		EpisodeNumber: 0,
		ReleaseGroup:  metadata.ReleaseGroup,
		Provider:      ProviderName,
		IsBestRelease: false,
		Confirmed:     false,
	}

	episode := -1

	if len(metadata.EpisodeNumber) == 1 {
		episode = util.StringToIntMust(metadata.EpisodeNumber[0])
	}

	// Force set episode number to 1 if it's a movie or single-episode and the torrent isn't a batch
	if !ret.IsBatch && episode == -1 && (media.EpisodeCount == 1 || media.Format == string(anilist.MediaFormatMovie)) {
		episode = 1
	}

	ret.EpisodeNumber = episode

	return ret
}
