package onlinestream_providers

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Gogoanime struct {
	BaseURL   string
	AjaxURL   string
	Client    http.Client
	UserAgent string
	logger    *zerolog.Logger
}

func NewGogoanime(logger *zerolog.Logger) *Gogoanime {
	return &Gogoanime{
		BaseURL:   "https://anitaku.to",
		AjaxURL:   "https://ajax.gogocdn.net",
		Client:    http.Client{},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		logger:    logger,
	}
}

func (g *Gogoanime) Search(query string, dubbed bool) ([]*SearchResult, error) {
	var results []*SearchResult

	c := colly.NewCollector()

	c.OnHTML(".last_episodes > ul > li", func(e *colly.HTMLElement) {
		id := ""
		idParts := strings.Split(e.ChildAttr("p.name > a", "href"), "/")
		if len(idParts) > 2 {
			id = idParts[2]
		}
		title := e.ChildText("p.name > a")
		url := g.BaseURL + e.ChildAttr("p.name > a", "href")
		subOrDub := Sub
		if strings.Contains(strings.ToLower(e.ChildText("p.name > a")), "dub") {
			subOrDub = Dub
		}
		results = append(results, &SearchResult{
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

	return results, nil
}

func (g *Gogoanime) FindEpisodes(id string) ([]*ProviderEpisode, error) {
	var episodes []*ProviderEpisode

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
		g.logger.Error().Err(err).Msg("failed to fetch episodes")
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
		spew.Dump(episodeID)
		episodes = append(episodes, &ProviderEpisode{
			ID:     episodeID,
			Number: episodeNumber,
			URL:    g.BaseURL + "/" + episodeID,
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
		g.logger.Error().Err(err).Msg("failed to fetch episodes")
		return nil, err
	}

	return episodes, nil
}

func (g *Gogoanime) FindEpisodeSources(episode *ProviderEpisode, server Server) (*ProviderEpisodeSource, error) {
	var source *ProviderEpisodeSource

	c := colly.NewCollector()

	switch server {
	case VidstreamingServer:
		c.OnHTML(".anime_muti_link > ul > li.vidcdn > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			gogocdn := onlinestream_sources.NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &ProviderEpisodeSource{
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episode.ID,
					},
					Sources: videoSources,
				}
			}
		})
	case GogocdnServer, "":
		c.OnHTML("#load_anime > div > div > iframe", func(e *colly.HTMLElement) {
			src := e.Attr("src")
			gogocdn := onlinestream_sources.NewGogoCDN()
			videoSources, err := gogocdn.Extract(src)
			if err == nil {
				source = &ProviderEpisodeSource{
					Headers: map[string]string{
						"Referer": g.BaseURL + "/" + episode.ID,
					},
					Sources: videoSources,
				}
			}
		})
	case StreamSBServer:
		c.OnHTML(".anime_muti_link > ul > li.streamsb > a", func(e *colly.HTMLElement) {
			src := e.Attr("data-video")
			streamsb := onlinestream_sources.NewStreamSB()
			videoSources, err := streamsb.Extract(src)
			if err == nil {
				source = &ProviderEpisodeSource{
					Headers: map[string]string{
						"Referer":    g.BaseURL + "/" + episode.ID,
						"watchsb":    "streamsb",
						"User-Agent": g.UserAgent,
					},
					Sources: videoSources,
				}
			}
		})
	}

	err := c.Visit(g.BaseURL + "/" + episode.ID)
	if err != nil {
		return nil, err
	}

	if source == nil {
		return nil, ErrSourceNotFound
	}

	return source, nil

}
