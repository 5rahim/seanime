package onlinestream_providers

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"net/http"
	"net/url"
	"seanime/internal/onlinestream/sources"
	"seanime/internal/util"
	"strconv"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

type Zoro struct {
	BaseURL   string
	Client    *http.Client
	UserAgent string
	logger    *zerolog.Logger
}

func NewZoro(logger *zerolog.Logger) hibikeonlinestream.Provider {
	return &Zoro{
		BaseURL:   "https://hianime.to",
		UserAgent: util.GetRandomUserAgent(),
		Client:    &http.Client{},
		logger:    logger,
	}
}

func (z *Zoro) GetSettings() hibikeonlinestream.Settings {
	return hibikeonlinestream.Settings{
		EpisodeServers: []string{VidcloudServer, VidstreamingServer},
		SupportsDub:    true,
	}
}

func (z *Zoro) Search(opts hibikeonlinestream.SearchOptions) ([]*hibikeonlinestream.SearchResult, error) {
	var results []*hibikeonlinestream.SearchResult

	query := opts.Query
	dubbed := opts.Dub

	z.logger.Debug().Str("query", query).Bool("dubbed", dubbed).Msg("zoro: Searching anime")

	c := colly.NewCollector()

	c.OnHTML(".flw-item", func(e *colly.HTMLElement) {
		id := strings.Split(strings.Split(e.ChildAttr(".film-name a", "href"), "/")[1], "?")[0]
		title := e.ChildText(".film-name a")
		url := strings.Split(z.BaseURL+e.ChildAttr(".film-name a", "href"), "?")[0]
		subOrDub := hibikeonlinestream.Sub
		foundSub := false
		foundDub := false
		if e.ChildText(".tick-item.tick-dub") != "" {
			foundDub = true
		}
		if e.ChildText(".tick-item.tick-sub") != "" {
			foundSub = true
		}
		if foundSub && foundDub {
			subOrDub = hibikeonlinestream.SubAndDub
		} else if foundDub {
			subOrDub = hibikeonlinestream.Dub
		}
		results = append(results, &hibikeonlinestream.SearchResult{
			ID:       id,
			Title:    title,
			URL:      url,
			SubOrDub: subOrDub,
		})
	})

	searchURL := z.BaseURL + "/search?keyword=" + url.QueryEscape(query)

	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	if dubbed {
		results = lo.Filter(results, func(r *hibikeonlinestream.SearchResult, _ int) bool {
			return r.SubOrDub == hibikeonlinestream.Dub || r.SubOrDub == hibikeonlinestream.SubAndDub
		})
	}

	z.logger.Debug().Int("count", len(results)).Msg("zoro: Fetched anime")

	return results, nil
}

func (z *Zoro) FindEpisodes(id string) ([]*hibikeonlinestream.EpisodeDetails, error) {
	var episodes []*hibikeonlinestream.EpisodeDetails

	z.logger.Debug().Str("id", id).Msg("zoro: Fetching episodes")

	c := colly.NewCollector()

	subOrDub := hibikeonlinestream.Sub

	c.OnHTML("div.film-stats > div.tick", func(e *colly.HTMLElement) {
		if e.ChildText(".tick-item.tick-dub") != "" {
			subOrDub = hibikeonlinestream.Dub
		}
		if e.ChildText(".tick-item.tick-sub") != "" {
			if subOrDub == hibikeonlinestream.Dub {
				subOrDub = hibikeonlinestream.SubAndDub
			}
		}
	})

	watchUrl := fmt.Sprintf("%s/watch/%s", z.BaseURL, id)
	err := c.Visit(watchUrl)
	if err != nil {
		z.logger.Error().Err(err).Msg("zoro: Failed to fetch episodes")
		return nil, err
	}

	// Get episodes

	splitId := strings.Split(id, "-")
	idNum := splitId[len(splitId)-1]
	ajaxUrl := fmt.Sprintf("%s/ajax/v2/episode/list/%s", z.BaseURL, idNum)

	c2 := colly.NewCollector(
		colly.UserAgent(z.UserAgent),
	)

	c2.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Referer", watchUrl)
	})

	c2.OnResponse(func(r *colly.Response) {
		var jsonResponse map[string]interface{}
		err = json.Unmarshal(r.Body, &jsonResponse)
		if err != nil {
			return
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResponse["html"].(string)))
		if err != nil {
			return
		}
		content := doc.Find(".detail-infor-content")
		content.Find("a").Each(func(i int, s *goquery.Selection) {
			id := s.AttrOr("href", "")
			if id == "" {
				return
			}
			hrefParts := strings.Split(s.AttrOr("href", ""), "/")
			if len(hrefParts) < 2 {
				return
			}
			if subOrDub == hibikeonlinestream.SubAndDub {
				subOrDub = "both"
			}
			id = fmt.Sprintf("%s$%s", strings.Replace(hrefParts[2], "?ep=", "$episode$", 1), subOrDub)
			epNumber, _ := strconv.Atoi(s.AttrOr("data-number", ""))
			url := z.BaseURL + s.AttrOr("href", "")
			title := s.AttrOr("title", "")
			episodes = append(episodes, &hibikeonlinestream.EpisodeDetails{
				Provider: ZoroProvider,
				ID:       id,
				Number:   epNumber,
				URL:      url,
				Title:    title,
			})
		})
	})

	err = c2.Visit(ajaxUrl)
	if err != nil {
		z.logger.Error().Err(err).Msg("zoro: Failed to fetch episodes")
		return nil, err
	}

	z.logger.Debug().Int("count", len(episodes)).Msg("zoro: Fetched episodes")

	return episodes, nil
}

func (z *Zoro) FindEpisodeServer(episodeInfo *hibikeonlinestream.EpisodeDetails, server string) (*hibikeonlinestream.EpisodeServer, error) {
	var source *hibikeonlinestream.EpisodeServer

	if server == DefaultServer {
		server = VidcloudServer
	}

	z.logger.Debug().Str("server", server).Str("episodeID", episodeInfo.ID).Msg("zoro: Fetching server sources")

	episodeParts := strings.Split(episodeInfo.ID, "$")

	if len(episodeParts) < 3 {
		return nil, errors.New("invalid episode id")
	}

	episodeID := fmt.Sprintf("%s?ep=%s", episodeParts[0], episodeParts[2])
	subOrDub := hibikeonlinestream.Sub
	if episodeParts[len(episodeParts)-1] == "dub" {
		subOrDub = hibikeonlinestream.Dub
	}

	// Get server

	var serverId string

	c := colly.NewCollector(
		colly.UserAgent(z.UserAgent),
	)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
	})

	c.OnResponse(func(r *colly.Response) {
		var jsonResponse map[string]interface{}
		err := json.Unmarshal(r.Body, &jsonResponse)
		if err != nil {
			return
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResponse["html"].(string)))
		if err != nil {
			return
		}

		switch server {
		case VidcloudServer:
			serverId = z.findServerId(doc, 4, subOrDub)
		case VidstreamingServer:
			serverId = z.findServerId(doc, 4, subOrDub)
		case StreamSBServer:
			serverId = z.findServerId(doc, 4, subOrDub)
		case StreamtapeServer:
			serverId = z.findServerId(doc, 4, subOrDub)
		}
	})

	ajaxEpisodeUrl := fmt.Sprintf("%s/ajax/v2/episode/servers?episodeId=%s", z.BaseURL, strings.Split(episodeID, "?ep=")[1])
	if err := c.Visit(ajaxEpisodeUrl); err != nil {
		return nil, err
	}

	if serverId == "" {
		return nil, ErrServerNotFound
	}

	c2 := colly.NewCollector(
		colly.UserAgent(z.UserAgent),
	)
	c2.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
	})

	c2.OnResponse(func(r *colly.Response) {
		var jsonResponse map[string]interface{}
		err := json.Unmarshal(r.Body, &jsonResponse)
		if err != nil {
			return
		}
		if _, ok := jsonResponse["link"].(string); !ok {
			return
		}
		switch server {
		case VidcloudServer, VidstreamingServer:
			megacloud := onlinestream_sources.NewMegaCloud()
			sources, err := megacloud.Extract(jsonResponse["link"].(string))
			if err != nil {
				return
			}
			source = &hibikeonlinestream.EpisodeServer{
				Provider:     ZoroProvider,
				Server:       server,
				Headers:      map[string]string{},
				VideoSources: sources,
			}
		case StreamtapeServer:
			streamtape := onlinestream_sources.NewStreamtape()
			sources, err := streamtape.Extract(jsonResponse["link"].(string))
			if err != nil {
				return
			}
			source = &hibikeonlinestream.EpisodeServer{
				Provider: ZoroProvider,
				Server:   server,
				Headers: map[string]string{
					"Referer":    jsonResponse["link"].(string),
					"User-Agent": z.UserAgent,
				},
				VideoSources: sources,
			}
		case StreamSBServer:
			streamsb := onlinestream_sources.NewStreamSB()
			sources, err := streamsb.Extract(jsonResponse["link"].(string))
			if err != nil {
				return
			}
			source = &hibikeonlinestream.EpisodeServer{
				Provider: ZoroProvider,
				Server:   server,
				Headers: map[string]string{
					"Referer":    jsonResponse["link"].(string),
					"watchsb":    "streamsb",
					"User-Agent": z.UserAgent,
				},
				VideoSources: sources,
			}
		}
	})

	// Get sources
	serverSourceUrl := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%s", z.BaseURL, serverId)
	if err := c2.Visit(serverSourceUrl); err != nil {
		return nil, err
	}

	if source == nil {
		z.logger.Warn().Str("server", server).Msg("zoro: No sources found")
		return nil, ErrSourceNotFound
	}

	z.logger.Debug().Str("server", server).Int("videoSources", len(source.VideoSources)).Msg("zoro: Fetched server sources")

	return source, nil
}

func (z *Zoro) findServerId(doc *goquery.Document, idx int, subOrDub hibikeonlinestream.SubOrDub) string {
	var serverId string
	doc.Find(fmt.Sprintf("div.ps_-block.ps_-block-sub.servers-%s > div.ps__-list > div", subOrDub)).Each(func(i int, s *goquery.Selection) {
		_serverId := s.AttrOr("data-server-id", "")
		if serverId == "" {
			if _serverId == strconv.Itoa(idx) {
				serverId = s.AttrOr("data-id", "")
			}
		}
	})
	return serverId
}
