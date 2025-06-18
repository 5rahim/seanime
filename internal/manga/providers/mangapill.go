package manga_providers

import (
	"fmt"
	"net/url"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

type (
	Mangapill struct {
		Url       string
		Client    *req.Client
		UserAgent string
		logger    *zerolog.Logger
	}
)

func NewMangapill(logger *zerolog.Logger) *Mangapill {
	client := req.C().
		SetUserAgent(util.GetRandomUserAgent()).
		SetTimeout(60 * time.Second).
		EnableInsecureSkipVerify().
		ImpersonateChrome()

	return &Mangapill{
		Url:       "https://mangapill.com",
		Client:    client,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

// DEVNOTE: Unique ID
// Each chapter ID has this format: {number}${slug} -- e.g. 6502-10004000$gokurakugai-chapter-4
// The chapter ID is split by the $ character to reconstruct the chapter URL for subsequent requests

func (mp *Mangapill) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (mp *Mangapill) Search(opts hibikemanga.SearchOptions) (ret []*hibikemanga.SearchResult, err error) {
	ret = make([]*hibikemanga.SearchResult, 0)

	mp.logger.Debug().Str("query", opts.Query).Msg("mangapill: Searching manga")

	uri := fmt.Sprintf("%s/search?q=%s", mp.Url, url.QueryEscape(opts.Query))

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.WithTransport(mp.Client.Transport)

	c.OnHTML("div.container div.my-3.justify-end > div", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		result := &hibikemanga.SearchResult{
			Provider: string(MangapillProvider),
		}

		result.ID = strings.Split(e.ChildAttr("a", "href"), "/manga/")[1]
		result.ID = strings.Replace(result.ID, "/", "$", -1)

		title := e.DOM.Find("div > a > div.mt-3").Text()
		result.Title = strings.TrimSpace(title)

		altTitles := e.DOM.Find("div > a > div.text-xs.text-secondary").Text()
		if altTitles != "" {
			result.Synonyms = []string{strings.TrimSpace(altTitles)}
		}

		compTitles := []*string{&result.Title}
		if len(result.Synonyms) > 0 {
			compTitles = append(compTitles, &result.Synonyms[0])
		}
		compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, compTitles)
		result.SearchRating = compRes.Rating

		result.Image = e.ChildAttr("a img", "data-src")

		yearStr := e.DOM.Find("div > div.flex > div").Eq(1).Text()
		year, err := strconv.Atoi(strings.TrimSpace(yearStr))
		if err != nil {
			result.Year = 0
		} else {
			result.Year = year
		}

		ret = append(ret, result)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Msg("mangapill: Failed to visit")
		return nil, err
	}

	// code

	if len(ret) == 0 {
		mp.logger.Error().Str("query", opts.Query).Msg("mangapill: No results found")
		return nil, ErrNoResults
	}

	mp.logger.Info().Int("count", len(ret)).Msg("mangapill: Found results")

	return ret, nil
}

func (mp *Mangapill) FindChapters(id string) (ret []*hibikemanga.ChapterDetails, err error) {
	ret = make([]*hibikemanga.ChapterDetails, 0)

	mp.logger.Debug().Str("mangaId", id).Msg("mangapill: Finding chapters")

	uriId := strings.Replace(id, "$", "/", -1)
	uri := fmt.Sprintf("%s/manga/%s", mp.Url, uriId)

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.WithTransport(mp.Client.Transport)

	c.OnHTML("div.container div.border-border div#chapters div.grid-cols-1 a", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		chapter := &hibikemanga.ChapterDetails{
			Provider: MangapillProvider,
		}

		chapter.ID = strings.Split(e.Attr("href"), "/chapters/")[1]
		chapter.ID = strings.Replace(chapter.ID, "/", "$", -1)

		chapter.Title = strings.TrimSpace(e.Text)

		splitTitle := strings.Split(chapter.Title, "Chapter ")
		if len(splitTitle) < 2 {
			return
		}
		chapter.Chapter = splitTitle[1]

		ret = append(ret, chapter)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Msg("mangapill: Failed to visit")
		return nil, err
	}

	if len(ret) == 0 {
		mp.logger.Error().Str("mangaId", id).Msg("mangapill: No chapters found")
		return nil, ErrNoChapters
	}

	slices.Reverse(ret)

	for i, chapter := range ret {
		chapter.Index = uint(i)
	}

	mp.logger.Info().Int("count", len(ret)).Msg("mangapill: Found chapters")

	return ret, nil
}

func (mp *Mangapill) FindChapterPages(id string) (ret []*hibikemanga.ChapterPage, err error) {
	ret = make([]*hibikemanga.ChapterPage, 0)

	mp.logger.Debug().Str("chapterId", id).Msg("mangapill: Finding chapter pages")

	uriId := strings.Replace(id, "$", "/", -1)
	uri := fmt.Sprintf("%s/chapters/%s", mp.Url, uriId)

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.WithTransport(mp.Client.Transport)

	c.OnHTML("chapter-page", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		page := &hibikemanga.ChapterPage{}

		page.URL = e.DOM.Find("div picture img").AttrOr("data-src", "")
		if page.URL == "" {
			return
		}
		indexStr := e.DOM.Find("div[data-summary] > div").Text()
		index, _ := strconv.Atoi(strings.Split(strings.Split(indexStr, "page ")[1], "/")[0])
		page.Index = index - 1

		page.Headers = map[string]string{
			"Referer": "https://mangapill.com/",
		}

		ret = append(ret, page)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Msg("mangapill: Failed to visit")
		return nil, err
	}

	if len(ret) == 0 {
		mp.logger.Error().Str("chapterId", id).Msg("mangapill: No pages found")
		return nil, ErrNoPages
	}

	mp.logger.Info().Int("count", len(ret)).Msg("mangapill: Found pages")

	return ret, nil

}
