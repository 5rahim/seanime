package onlinestream_providers

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	"seanime/internal/onlinestream/sources"
	"seanime/internal/util"
	"strconv"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

type Gogoanime struct {
	BaseURL   string
	AjaxURL   string
	Client    http.Client
	UserAgent string
	logger    *zerolog.Logger
}

func NewGogoanime(logger *zerolog.Logger) hibikeonlinestream.Provider {
	return &Gogoanime{
		BaseURL:   "https://anitaku.to",
		AjaxURL:   "https://ajax.gogocdn.net",
		Client:    http.Client{},
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (g *Gogoanime) GetSettings() hibikeonlinestream.Settings {
	return hibikeonlinestream.Settings{
		EpisodeServers: []string{GogocdnServer, VidstreamingServer},
		SupportsDub:    true,
	}
}

func (g *Gogoanime) Search(opts hibikeonlinestream.SearchOptions) ([]*hibikeonlinestream.SearchResult, error) {
	var results []*hibikeonlinestream.SearchResult

	query := opts.Query
	dubbed := opts.Dub

	g.logger.Debug().Str("query", query).Bool("dubbed", dubbed).Msg("gogoanime: Searching anime")

	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	c.OnHTML(".last_episodes > ul > li", func(e *colly.HTMLElement) {
		id := ""
		idParts := strings.Split(e.ChildAttr("p.name > a", "href"), "/")
		if len(idParts) > 2 {
			id = idParts[2]
		}
		title := e.ChildText("p.name > a")
		url := g.BaseURL + e.ChildAttr("p.name > a", "href")
		subOrDub := hibikeonlinestream.Sub
		if strings.Contains(strings.ToLower(e.ChildText("p.name > a")), "dub") {
			subOrDub = hibikeonlinestream.Dub
		}
		results = append(results, &hibikeonlinestream.SearchResult{
			ID:       id,
			Title:    title,
			URL:      url,
			SubOrDub: subOrDub,
		})
	})

	searchURL := g.BaseURL + "/search.html?keyword=" + url.QueryEscape(query)
	if dubbed {
		searchURL += "%20(Dub)"
	}

	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	g.logger.Debug().Int("count", len(results)).Msg("gogoanime: Fetched anime")

	return results, nil
}

func (g *Gogoanime) FindEpisodes(id string) ([]*hibikeonlinestream.EpisodeDetails, error) {
	var episodes []*hibikeonlinestream.EpisodeDetails

	g.logger.Debug().Str("id", id).Msg("gogoanime: Fetching episodes")

	if !strings.Contains(id, "gogoanime") {
		id = fmt.Sprintf("%s/category/%s", g.BaseURL, id)
	}

	c := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	var epStart, epEnd, movieID, alias string

	c.OnHTML("#episode_page > li > a", func(e *colly.HTMLElement) {
		if epStart == "" {
			epStart = e.Attr("ep_start")
		}
		epEnd = e.Attr("ep_end")
	})

	c.OnHTML("#movie_id", func(e *colly.HTMLElement) {
		movieID = e.Attr("value")
	})

	c.OnHTML("#alias", func(e *colly.HTMLElement) {
		alias = e.Attr("value")
	})

	err := c.Visit(id)
	if err != nil {
		g.logger.Error().Err(err).Msg("gogoanime: Failed to fetch episodes")
		return nil, err
	}

	c2 := colly.NewCollector(
		colly.UserAgent(g.UserAgent),
	)

	c2.OnHTML("#episode_related > li", func(e *colly.HTMLElement) {
		episodeIDParts := strings.Split(e.ChildAttr("a", "href"), "/")
		if len(episodeIDParts) < 2 {
			return
		}
		episodeID := strings.TrimSpace(episodeIDParts[1])
		episodeNumberStr := strings.TrimPrefix(e.ChildText("div.name"), "EP ")
		episodeNumber, err := strconv.Atoi(episodeNumberStr)
		if err != nil {
			g.logger.Error().Err(err).Str("episodeID", episodeID).Msg("failed to parse episode number")
			return
		}
		episodes = append(episodes, &hibikeonlinestream.EpisodeDetails{
			Provider: GogoanimeProvider,
			ID:       episodeID,
			Number:   episodeNumber,
			URL:      g.BaseURL + "/" + episodeID,
		})
	})

	ajaxURL := fmt.Sprintf("%s/ajax/load-list-episode", g.AjaxURL)
	ajaxParams := url.Values{
		"ep_start":   {epStart},
		"ep_end":     {epEnd},
		"id":         {movieID},
		"alias":      {alias},
		"default_ep": {"0"},
	}
	ajaxURLWithParams := fmt.Sprintf("%s?%s", ajaxURL, ajaxParams.Encode())

	err = c2.Visit(ajaxURLWithParams)
	if err != nil {
		g.logger.Error().Err(err).Msg("gogoanime: Failed to fetch episodes")
		return nil, err
	}

	g.logger.Debug().Int("count", len(episodes)).Msg("gogoanime: Fetched episodes")

	return episodes, nil
}

func (g *Gogoanime) FindEpisodeServer(episodeInfo *hibikeonlinestream.EpisodeDetails, server string) (*hibikeonlinestream.EpisodeServer, error) {
	var source *hibikeonlinestream.EpisodeServer

	if server == DefaultServer {
		server = GogocdnServer
	}
	g.logger.Debug().Str("server", string(server)).Str("episodeID", episodeInfo.ID).Msg("gogoanime: Fetching server sources")

	c := colly.NewCollector()

	switch server {
	case VidstreamingServer:
		c.OnHTML(".anime_muti_link > ul > li.vidcdn > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			gogocdn := onlinestream_sources.NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episodeInfo.ID,
					},
					VideoSources: videoSources,
				}
			}
		})
	case GogocdnServer, "":
		c.OnHTML("#load_anime > div > div > iframe", func(e *colly.HTMLElement) {
			src := e.Attr("src")
			gogocdn := onlinestream_sources.NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episodeInfo.ID,
					},
					VideoSources: videoSources,
				}
			}
		})
	case StreamSBServer:
		c.OnHTML(".anime_muti_link > ul > li.streamsb > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			streamsb := onlinestream_sources.NewStreamSB()
			videoSources, err := streamsb.Extract(src)
			if err == nil {
				source = &hibikeonlinestream.EpisodeServer{
					Provider: GogoanimeProvider,
					Server:   server,
					Headers: map[string]string{
						"Referer":    g.BaseURL + "/" + episodeInfo.ID,
						"watchsb":    "streamsb",
						"User-Agent": g.UserAgent,
					},
					VideoSources: videoSources,
				}
			}
		})
	}

	err := c.Visit(g.BaseURL + "/" + episodeInfo.ID)
	if err != nil {
		return nil, err
	}

	if source == nil {
		g.logger.Warn().Str("server", server).Msg("gogoanime: No sources found")
		return nil, ErrSourceNotFound
	}

	g.logger.Debug().Str("server", server).Int("videoSources", len(source.VideoSources)).Msg("gogoanime: Fetched server sources")

	return source, nil

}
