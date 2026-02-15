package filler

import (
	"fmt"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adrg/strutil/metrics"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

type (
	SearchOptions struct {
		Titles []string
	}

	SearchResult struct {
		Slug  string
		Title string
	}

	API interface {
		Search(opts SearchOptions) (*SearchResult, error)
		FindFillerData(slug string) (*Data, error)
	}

	Data struct {
		FillerEpisodes []string `json:"fillerEpisodes"`
	}
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	AnimeFillerList struct {
		baseUrl string
		client  *req.Client
		logger  *zerolog.Logger
	}
)

func NewAnimeFillerList(logger *zerolog.Logger) *AnimeFillerList {
	return &AnimeFillerList{
		baseUrl: "https://www.animefillerlist.com",
		client: req.C().
			SetTimeout(10 * time.Second).
			ImpersonateChrome(),
		logger: logger,
	}
}

func (af *AnimeFillerList) Search(opts SearchOptions) (result *SearchResult, err error) {

	defer util.HandlePanicInModuleWithError("api/metadata/filler/Search", &err)

	ret := make([]*SearchResult, 0)

	resp, err := af.client.R().Get(fmt.Sprintf("%s/shows", af.baseUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("div.Group > ul > li > a").Each(func(i int, s *goquery.Selection) {
		ret = append(ret, &SearchResult{
			Slug:  s.AttrOr("href", ""),
			Title: s.Text(),
		})
	})

	if len(ret) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	lev := metrics.NewLevenshtein()
	lev.CaseSensitive = false

	compResults := make([]struct {
		OriginalValue string
		Value         string
		Distance      int
	}, 0)

	for _, result := range ret {
		firstTitle := result.Title
		secondTitle := ""

		// Check if a second title exists between parentheses
		if strings.LastIndex(firstTitle, " (") != -1 && strings.LastIndex(firstTitle, ")") != -1 {
			secondTitle = firstTitle[strings.LastIndex(firstTitle, " (")+2 : strings.LastIndex(firstTitle, ")")]
			if !util.IsMostlyLatinString(secondTitle) {
				secondTitle = ""
			}
		}

		if secondTitle != "" {
			firstTitle = firstTitle[:strings.LastIndex(firstTitle, " (")]
		}

		for _, mediaTitle := range opts.Titles {
			compResults = append(compResults, struct {
				OriginalValue string
				Value         string
				Distance      int
			}{
				OriginalValue: result.Title,
				Value:         firstTitle,
				Distance:      lev.Distance(mediaTitle, firstTitle),
			})
			if secondTitle != "" {
				compResults = append(compResults, struct {
					OriginalValue string
					Value         string
					Distance      int
				}{
					OriginalValue: result.Title,
					Value:         secondTitle,
					Distance:      lev.Distance(mediaTitle, secondTitle),
				})
			}
		}
	}

	// Find the best match
	bestResult := struct {
		OriginalValue string
		Value         string
		Distance      int
	}{}

	for _, result := range compResults {
		if bestResult.OriginalValue == "" || result.Distance <= bestResult.Distance {
			if bestResult.OriginalValue != "" && result.Distance == bestResult.Distance && len(result.OriginalValue) > len(bestResult.OriginalValue) {
				continue
			}
			bestResult = result
		}
	}

	if bestResult.OriginalValue == "" {
		return nil, fmt.Errorf("no results found")
	}

	if bestResult.Distance > 10 {
		return nil, fmt.Errorf("no results found")
	}

	// Get the result
	for _, r := range ret {
		if r.Title == bestResult.OriginalValue {
			return r, nil
		}
	}

	return
}

func (af *AnimeFillerList) FindFillerData(slug string) (ret *Data, err error) {

	defer util.HandlePanicInModuleWithError("api/metadata/filler/FindFillerEpisodes", &err)

	ret = &Data{
		FillerEpisodes: make([]string, 0),
	}

	resp, err := af.client.R().Get(fmt.Sprintf("%s%s", af.baseUrl, slug))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	fillerEps := make([]string, 0)
	doc.Find("tr.filler").Each(func(i int, s *goquery.Selection) {
		fillerEps = append(fillerEps, s.Find("td.Number").Text())
	})

	ret.FillerEpisodes = fillerEps

	return
}
