package onlinestream_providers

import (
	"cmp"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	"regexp"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	onlinestream_sources "seanime/internal/onlinestream/sources"
	"seanime/internal/util"
	"sort"
	"strings"
	"sync"
)

type (
	Animepahe struct {
		BaseURL   string
		Client    http.Client
		UserAgent string
		logger    *zerolog.Logger
	}
	AnimepaheSearchResult struct {
		Data []struct {
			ID      int    `json:"id"`
			Title   string `json:"title"`
			Year    int    `json:"year"`
			Poster  string `json:"poster"`
			Type    string `json:"type"`
			Session string `json:"session"`
		} `json:"data"`
	}
)

func NewAnimepahe(logger *zerolog.Logger) hibikeonlinestream.Provider {
	return &Animepahe{
		BaseURL:   "https://animepahe.ru",
		Client:    http.Client{},
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (g *Animepahe) GetSettings() hibikeonlinestream.Settings {
	return hibikeonlinestream.Settings{
		EpisodeServers: []string{"animepahe"},
		SupportsDub:    false,
	}
}

func (g *Animepahe) Search(opts hibikeonlinestream.SearchOptions) ([]*hibikeonlinestream.SearchResult, error) {
	var results []*hibikeonlinestream.SearchResult

	query := opts.Query
	dubbed := opts.Dub

	g.logger.Debug().Str("query", query).Bool("dubbed", dubbed).Msg("animepahe: Searching anime")

	q := url.QueryEscape(query)
	request, err := http.NewRequest("GET", g.BaseURL+fmt.Sprintf("/api?m=search&q=%s", q), nil)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
		return nil, err
	}

	request.Header.Set("User-Agent", g.UserAgent)
	request.Header.Set("Cookie", "__ddg1_=;__ddg2_=;")

	response, err := g.Client.Do(request)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
		return nil, err
	}
	defer response.Body.Close()

	var searchResult AnimepaheSearchResult
	err = json.NewDecoder(response.Body).Decode(&searchResult)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to decode response")
		return nil, err
	}

	for _, data := range searchResult.Data {
		results = append(results, &hibikeonlinestream.SearchResult{
			ID:       cmp.Or(fmt.Sprintf("%d", data.ID), data.Session),
			Title:    data.Title,
			URL:      fmt.Sprintf("%s/anime/%d", g.BaseURL, data.ID),
			SubOrDub: hibikeonlinestream.Sub,
		})
	}

	return results, nil
}

func (g *Animepahe) FindEpisodes(id string) ([]*hibikeonlinestream.EpisodeDetails, error) {
	var episodes []*hibikeonlinestream.EpisodeDetails

	q1 := fmt.Sprintf("/anime/%s", id)
	if !strings.Contains(id, "-") {
		q1 = fmt.Sprintf("/a/%s", id)
	}
	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "__ddg1_=;__ddg2_=")
	})

	var tempId string
	c.OnHTML("head > meta[property='og:url']", func(e *colly.HTMLElement) {
		parts := strings.Split(e.Attr("content"), "/")
		tempId = parts[len(parts)-1]
	})

	err := c.Visit(g.BaseURL + q1)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to fetch episodes")
		return nil, err
	}

	// { last_page: number; data: { id: number; episode: number; title: string; snapshot: string; filler: number; created_at?: string }[] }
	type data struct {
		LastPage int `json:"last_page"`
		Data     []struct {
			ID        int    `json:"id"`
			Episode   int    `json:"episode"`
			Title     string `json:"title"`
			Snapshot  string `json:"snapshot"`
			Filler    int    `json:"filler"`
			Session   string `json:"session"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}

	q2 := fmt.Sprintf("/api?m=release&id=%s&sort=episode_asc&page=1", tempId)
	request, err := http.NewRequest("GET", g.BaseURL+q2, nil)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
		return nil, err
	}

	request.Header.Set("User-Agent", g.UserAgent)
	request.Header.Set("Cookie", "__ddg1_=;__ddg2_=")

	response, err := g.Client.Do(request)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
		return nil, err
	}
	defer response.Body.Close()

	var d data
	err = json.NewDecoder(response.Body).Decode(&d)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to decode response")
		return nil, err
	}

	for _, e := range d.Data {
		episodes = append(episodes, &hibikeonlinestream.EpisodeDetails{
			Provider: "animepahe",
			ID:       fmt.Sprintf("%d$%s", e.ID, id),
			Number:   e.Episode,
			URL:      fmt.Sprintf("%s/anime/%s/%d", g.BaseURL, id, e.Episode),
			Title:    cmp.Or(e.Title, "Episode "+fmt.Sprintf("%d", e.Episode)),
		})
	}

	var pageNumbers []int

	for i := 2; i <= d.LastPage; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pageNumbers))
	mu := sync.Mutex{}

	for _, p := range pageNumbers {
		go func(p int) {
			defer wg.Done()
			q2 := fmt.Sprintf("/api?m=release&id=%s&sort=episode_asc&page=%d", tempId, p)
			request, err := http.NewRequest("GET", g.BaseURL+q2, nil)
			if err != nil {
				g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
				return
			}

			request.Header.Set("User-Agent", g.UserAgent)
			request.Header.Set("Cookie", "__ddg1_=;__ddg2_=")

			response, err := g.Client.Do(request)
			if err != nil {
				g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
				return
			}
			defer response.Body.Close()

			var d data
			err = json.NewDecoder(response.Body).Decode(&d)
			if err != nil {
				g.logger.Error().Err(err).Msg("animepahe: Failed to decode response")
				return
			}

			mu.Lock()
			for _, e := range d.Data {
				episodes = append(episodes, &hibikeonlinestream.EpisodeDetails{
					Provider: "animepahe",
					ID:       fmt.Sprintf("%d$%s", e.ID, id),
					Number:   e.Episode,
					URL:      fmt.Sprintf("%s/anime/%s/%d", g.BaseURL, id, e.Episode),
					Title:    cmp.Or(e.Title, "Episode "+fmt.Sprintf("%d", e.Episode)),
				})
			}
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	g.logger.Debug().Int("count", len(episodes)).Msg("animepahe: Fetched episodes")

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found")
	}

	// Normalize episode numbers
	offset := episodes[0].Number + 1
	for i, e := range episodes {
		episodes[i].Number = e.Number - offset
	}

	return episodes, nil
}

func (g *Animepahe) FindEpisodeServer(episodeInfo *hibikeonlinestream.EpisodeDetails, server string) (*hibikeonlinestream.EpisodeServer, error) {
	var source *hibikeonlinestream.EpisodeServer

	parts := strings.Split(episodeInfo.ID, "$")
	if len(parts) < 2 {
		return nil, fmt.Errorf("animepahe: Invalid episode ID")
	}

	episodeID := parts[0]
	animeID := parts[1]

	q1 := fmt.Sprintf("/anime/%s", animeID)
	if !strings.Contains(animeID, "-") {
		q1 = fmt.Sprintf("/a/%s", animeID)
	}
	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	var reqUrl *url.URL

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "__ddg1_=;__ddg2_=")
	})

	c.OnResponse(func(r *colly.Response) {
		reqUrl = r.Request.URL
	})

	var tempId string
	c.OnHTML("head > meta[property='og:url']", func(e *colly.HTMLElement) {
		parts := strings.Split(e.Attr("content"), "/")
		tempId = parts[len(parts)-1]
	})

	err := c.Visit(g.BaseURL + q1)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to fetch episodes")
		return nil, err
	}

	var sessionId string
	// retain url without query
	reqUrlStr := reqUrl.Path
	reqUrlStrParts := strings.Split(reqUrlStr, "/anime/")
	sessionId = reqUrlStrParts[len(reqUrlStrParts)-1]

	// { last_page: number; data: { id: number; episode: number; title: string; snapshot: string; filler: number; created_at?: string }[] }
	type data struct {
		LastPage int `json:"last_page"`
		Data     []struct {
			ID      int    `json:"id"`
			Session string `json:"session"`
		} `json:"data"`
	}

	q2 := fmt.Sprintf("/api?m=release&id=%s&sort=episode_asc&page=1", tempId)
	request, err := http.NewRequest("GET", g.BaseURL+q2, nil)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
		return nil, err
	}

	request.Header.Set("User-Agent", g.UserAgent)
	request.Header.Set("Cookie", "__ddg1_=;__ddg2_=")

	response, err := g.Client.Do(request)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
		return nil, err
	}
	defer response.Body.Close()

	var d data
	err = json.NewDecoder(response.Body).Decode(&d)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to decode response")
		return nil, err
	}

	episodeSession := ""

	for _, e := range d.Data {
		if fmt.Sprintf("%d", e.ID) == episodeID {
			episodeSession = e.Session
			break
		}
	}

	var pageNumbers []int

	for i := 1; i <= d.LastPage; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	if episodeSession == "" {
		wg := sync.WaitGroup{}
		wg.Add(len(pageNumbers))
		mu := sync.Mutex{}

		for _, p := range pageNumbers {
			go func(p int) {
				defer wg.Done()
				q2 := fmt.Sprintf("/api?m=release&id=%s&sort=episode_asc&page=%d", tempId, p)
				request, err := http.NewRequest("GET", g.BaseURL+q2, nil)
				if err != nil {
					g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
					return
				}

				request.Header.Set("User-Agent", g.UserAgent)
				request.Header.Set("Cookie", "__ddg1_=;__ddg2_=")

				response, err := g.Client.Do(request)
				if err != nil {
					g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
					return
				}
				defer response.Body.Close()

				var d data
				err = json.NewDecoder(response.Body).Decode(&d)
				if err != nil {
					g.logger.Error().Err(err).Msg("animepahe: Failed to decode response")
					return
				}

				mu.Lock()
				for _, e := range d.Data {
					if fmt.Sprintf("%d", e.ID) == episodeID {
						episodeSession = e.Session
						break
					}
				}
				mu.Unlock()
			}(p)
		}

		wg.Wait()
	}

	if episodeSession == "" {
		return nil, fmt.Errorf("animepahe: Episode not found")
	}

	q3 := fmt.Sprintf("/play/%s/%s", sessionId, episodeSession)
	request2, err := http.NewRequest("GET", g.BaseURL+q3, nil)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to create request")
		return nil, err
	}

	request2.Header.Set("User-Agent", g.UserAgent)
	request2.Header.Set("Cookie", "__ddg1_=;__ddg2_=")

	response2, err := g.Client.Do(request2)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to send request")
		return nil, err
	}
	defer response2.Body.Close()

	htmlString := ""

	doc, err := goquery.NewDocumentFromReader(response2.Body)
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to parse response")
		return nil, err
	}

	htmlString = doc.Text()

	//const regex = /https:\/\/kwik\.si\/e\/\w+/g;
	//            const matches = watchReq.match(regex);
	//
	//            if (matches === null) return undefined;

	re := regexp.MustCompile(`https:\/\/kwik\.si\/e\/\w+`)
	matches := re.FindAllString(htmlString, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("animepahe: Failed to find episode source")
	}

	kwik := onlinestream_sources.NewKwik()
	videoSources, err := kwik.Extract(matches[0])
	if err != nil {
		g.logger.Error().Err(err).Msg("animepahe: Failed to extract video sources")
		return nil, fmt.Errorf("animepahe: Failed to extract video sources, %w", err)
	}

	source = &hibikeonlinestream.EpisodeServer{
		Provider:     "animepahe",
		Server:       KwikServer,
		Headers:      map[string]string{"Referer": "https://kwik.si/"},
		VideoSources: videoSources,
	}

	return source, nil

}
