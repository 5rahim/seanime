package onlinestream_providers

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Zoro struct {
	BaseURL   string
	Logo      string
	Client    *http.Client
	UserAgent string
}

func NewZoro() *Zoro {
	return &Zoro{
		BaseURL:   "https://hianime.to",
		Logo:      "https://is3-ssl.mzstatic.com/image/thumb/Purple112/v4/7e/91/00/7e9100ee-2b62-0942-4cdc-e9b93252ce1c/source/512x512bb.jpg",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		Client:    &http.Client{},
	}
}

func (z *Zoro) Search(query string, dubbed bool) ([]*SearchResult, error) {
	var results []*SearchResult

	c := colly.NewCollector()

	c.OnHTML(".flw-item", func(e *colly.HTMLElement) {
		id := strings.Split(strings.Split(e.ChildAttr(".film-name a", "href"), "/")[1], "?")[0]
		title := e.ChildText(".film-name a")
		url := strings.Split(z.BaseURL+e.ChildAttr(".film-name a", "href"), "?")[0]
		subOrDub := Sub
		foundSub := false
		foundDub := false
		if e.ChildText(".tick-item.tick-dub") != "" {
			foundDub = true
		}
		if e.ChildText(".tick-item.tick-sub") != "" {
			foundSub = true
		}
		if foundSub && foundDub {
			subOrDub = SubAndDub
		} else if foundDub {
			subOrDub = Dub
		}
		results = append(results, &SearchResult{
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
		results = lo.Filter(results, func(r *SearchResult, _ int) bool {
			return r.SubOrDub == Dub || r.SubOrDub == SubAndDub
		})
	}

	return results, nil
}

func (z *Zoro) FetchEpisodes(id string) ([]*ProviderEpisode, error) {
	var episodes []*ProviderEpisode

	c := colly.NewCollector()

	subOrDub := Sub

	c.OnHTML("div.film-stats > div.tick", func(e *colly.HTMLElement) {
		if e.ChildText(".tick-item.tick-dub") != "" {
			subOrDub = Dub
		}
		if e.ChildText(".tick-item.tick-sub") != "" {
			if subOrDub == Dub {
				subOrDub = SubAndDub
			}
		}
	})

	watchUrl := fmt.Sprintf("%s/watch/%s", z.BaseURL, id)
	err := c.Visit(watchUrl)
	if err != nil {
		return nil, err
	}

	// Get episodes

	splitId := strings.Split(id, "-")
	idNum := splitId[len(splitId)-1]
	ajaxUrl := fmt.Sprintf("%s/ajax/v2/episode/list/%s", z.BaseURL, idNum)
	spew.Dump(ajaxUrl)

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
			if subOrDub == SubAndDub {
				subOrDub = "auto"
			}
			id = fmt.Sprintf("%s$%s", strings.Replace(hrefParts[2], "?ep=", "$episode$", 1), subOrDub)
			epNumber, _ := strconv.Atoi(s.AttrOr("data-number", ""))
			url := z.BaseURL + s.AttrOr("href", "")
			title := s.AttrOr("title", "")
			episodes = append(episodes, &ProviderEpisode{
				ID:     id,
				Number: epNumber,
				URL:    url,
				Title:  title,
			})
		})
	})

	err = c2.Visit(ajaxUrl)
	if err != nil {
		return nil, err
	}

	return episodes, nil
}

func (z *Zoro) FetchEpisodeSources(episode *ProviderEpisode, server Server) (*ProviderEpisodeSource, error) {
	var source *ProviderEpisodeSource

	if server == DefaultServer {
		server = VidcloudServer
	}

	episodeParts := strings.Split(episode.ID, "$")

	if len(episodeParts) < 3 {
		return nil, errors.New("invalid episode id")
	}

	episodeID := fmt.Sprintf("%s?ep=%s", episodeParts[0], episodeParts[2])
	subOrDub := Sub
	if episodeParts[len(episodeParts)-1] == "dub" {
		subOrDub = Dub
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
			serverId = z.findServerId(doc, 1, subOrDub)
		case VidstreamingServer:
			serverId = z.findServerId(doc, 4, subOrDub)
		case StreamSBServer:
			serverId = z.findServerId(doc, 5, subOrDub)
		case StreamtapeServer:
			serverId = z.findServerId(doc, 3, subOrDub)
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
			source = &ProviderEpisodeSource{
				Headers: map[string]string{},
				Sources: sources,
			}
		case StreamtapeServer:
			streamtape := onlinestream_sources.NewStreamtape()
			sources, err := streamtape.Extract(jsonResponse["link"].(string))
			if err != nil {
				return
			}
			source = &ProviderEpisodeSource{
				Headers: map[string]string{
					"Referer":    jsonResponse["link"].(string),
					"User-Agent": z.UserAgent,
				},
				Sources: sources,
			}
		case StreamSBServer:
			streamsb := onlinestream_sources.NewStreamSB()
			sources, err := streamsb.Extract(jsonResponse["link"].(string))
			if err != nil {
				return
			}
			source = &ProviderEpisodeSource{
				Headers: map[string]string{
					"Referer":    jsonResponse["link"].(string),
					"watchsb":    "streamsb",
					"User-Agent": z.UserAgent,
				},
				Sources: sources,
			}
		}
	})

	// Get sources
	serverSourceUrl := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%s", z.BaseURL, serverId)
	if err := c2.Visit(serverSourceUrl); err != nil {
		return nil, err
	}

	if source == nil {
		return nil, ErrSourceNotFound
	}

	return source, nil
}

func (z *Zoro) findServerId(doc *goquery.Document, idx int, subOrDub SubOrDub) string {
	var serverId string
	doc.Find(fmt.Sprintf(".ps_-block.ps_-block-sub.servers-%s > .ps__-list .server-item", subOrDub)).Each(func(i int, s *goquery.Selection) {
		_serverId := s.AttrOr("data-server-id", "")
		if serverId == "" {
			if _serverId == strconv.Itoa(idx) {
				serverId = s.AttrOr("data-id", "")
			}
		}
	})
	return serverId
}
