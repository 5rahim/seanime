package manga_providers

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strings"
	"time"
)

type (
	Manganato struct {
		Url       string
		Client    *http.Client
		UserAgent string
		logger    *zerolog.Logger
	}

	ManganatoSearchResult struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		NameUnsigned string `json:"nameunsigned"`
		LastChapter  string `json:"lastchapter"`
		Image        string `json:"image"`
		Author       string `json:"author"`
		StoryLink    string `json:"story_link"`
	}
)

func NewManganato(logger *zerolog.Logger) *Manganato {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	return &Manganato{
		Url:       "https://manganato.com",
		Client:    c,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (mp *Manganato) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (mp *Manganato) Search(opts hibikemanga.SearchOptions) (ret []*hibikemanga.SearchResult, err error) {
	ret = make([]*hibikemanga.SearchResult, 0)

	mp.logger.Debug().Str("query", opts.Query).Msg("manganato: Searching manga")

	q := opts.Query
	q = strings.ReplaceAll(q, " ", "_")
	q = strings.ToLower(q)
	q = strings.TrimSpace(q)
	q = url.QueryEscape(q)
	uri := fmt.Sprintf("https://manganato.com/search/story/%s", q)

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.OnHTML("div.search-story-item", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		result := &hibikemanga.SearchResult{
			Provider: string(ManganatoProvider),
		}
		result.ID = e.DOM.Find("a.item-title").AttrOr("href", "")
		splitHref := strings.Split(result.ID, "/")

		if strings.Contains(e.DOM.Find("a.item-title").AttrOr("href", ""), "chapmanganato") {
			result.ID = "chapmanganato$"
		} else {
			result.ID = "manganato$"
		}

		result.ID += splitHref[3]
		result.Title = e.DOM.Find("a.item-title").Text()
		result.Image = e.DOM.Find("img").AttrOr("src", "")

		compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, []*string{&result.Title})
		result.SearchRating = compRes.Rating
		ret = append(ret, result)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to visit")
		return nil, err
	}

	if len(ret) == 0 {
		mp.logger.Error().Str("query", opts.Query).Msg("manganato: No results found")
		return nil, ErrNoResults
	}

	mp.logger.Info().Int("count", len(ret)).Msg("manganato: Found results")

	return ret, nil
}

func (mp *Manganato) FindChapters(id string) (ret []*hibikemanga.ChapterDetails, err error) {
	ret = make([]*hibikemanga.ChapterDetails, 0)

	mp.logger.Debug().Str("mangaId", id).Msg("manganato: Finding chapters")

	splitId := strings.Split(id, "$")
	if len(splitId) != 2 {
		mp.logger.Error().Str("mangaId", id).Msg("manganato: Invalid manga ID")
		return nil, ErrNoChapters
	}

	uri := ""
	if splitId[0] == "manganato" {
		uri = fmt.Sprintf("https://manganato.com/%s", splitId[1])
	} else if splitId[0] == "chapmanganato" {
		uri = fmt.Sprintf("https://chapmanganato.to/%s", splitId[1])
	}

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.OnHTML("li.a-h", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		name := e.DOM.Find("a").Text()
		if strings.HasPrefix(name, "Vol.") {
			split := strings.Split(name, " ")
			name = strings.Join(split[1:], " ")
		}
		chStr := strings.TrimSpace(strings.Split(name, " ")[1])
		chStr = strings.TrimSuffix(chStr, ":")
		href := e.ChildAttr("a", "href")
		id := strings.Split(href, "/")[4]
		chapter := &hibikemanga.ChapterDetails{
			Provider: string(ManganatoProvider),
			ID:       splitId[1] + "$" + id,
			URL:      href,
			Title:    strings.TrimSpace(name),
			Chapter:  chStr,
		}
		ret = append(ret, chapter)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to visit")
		return nil, err
	}

	slices.Reverse(ret)
	for i, chapter := range ret {
		chapter.Index = uint(i)
	}

	if len(ret) == 0 {
		mp.logger.Error().Str("mangaId", id).Msg("manganato: No chapters found")
		return nil, ErrNoChapters
	}

	mp.logger.Info().Int("count", len(ret)).Msg("manganato: Found chapters")

	return ret, nil
}

func (mp *Manganato) FindChapterPages(id string) (ret []*hibikemanga.ChapterPage, err error) {
	ret = make([]*hibikemanga.ChapterPage, 0)

	mp.logger.Debug().Str("chapterId", id).Msg("manganato: Finding chapter pages")

	splitId := strings.Split(id, "$")
	if len(splitId) != 2 {
		mp.logger.Error().Str("chapterId", id).Msg("manganato: Invalid chapter ID")
		return nil, ErrNoPages
	}

	uri := fmt.Sprintf("https://chapmanganato.to/%s/%s", splitId[0], splitId[1])

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.OnHTML(".container-chapter-reader img", func(e *colly.HTMLElement) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		if e.Attr("src") == "" {
			return
		}
		page := &hibikemanga.ChapterPage{
			Provider: string(ManganatoProvider),
			URL:      e.Attr("src"),
			Index:    len(ret),
			Headers: map[string]string{
				"Referer": "https://chapmanganato.to",
			},
		}
		ret = append(ret, page)
	})

	err = c.Visit(uri)
	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to visit")
		return nil, err
	}

	if len(ret) == 0 {
		mp.logger.Error().Str("chapterId", id).Msg("manganato: No pages found")
		return nil, ErrNoPages
	}

	mp.logger.Info().Int("count", len(ret)).Msg("manganato: Found pages")

	return ret, nil

}
