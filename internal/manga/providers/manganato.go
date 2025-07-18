package manga_providers

import (
	"bytes"
	"fmt"
	"net/url"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

type (
	Manganato struct {
		Url    string
		Client *req.Client
		logger *zerolog.Logger
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
	client := req.C().
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36").
		SetTimeout(60 * time.Second).
		EnableInsecureSkipVerify().
		ImpersonateSafari()

	return &Manganato{
		Url:    "https://natomanga.com",
		Client: client,
		logger: logger,
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
	uri := fmt.Sprintf("https://natomanga.com/search/story/%s", q)

	resp, err := mp.Client.R().
		SetHeader("User-Agent", util.GetRandomUserAgent()).
		Get(uri)

	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to send request")
		return nil, err
	}

	if !resp.IsSuccessState() {
		mp.logger.Error().Str("status", resp.Status).Str("uri", uri).Msg("manganato: Request failed")
		return nil, fmt.Errorf("failed to fetch search results: status %s", resp.Status)
	}

	bodyBytes := resp.Bytes()

	//mp.logger.Debug().Str("body", string(bodyBytes)).Msg("manganato: Response body")

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		mp.logger.Error().Err(err).Msg("manganato: Failed to parse HTML")
		return nil, err
	}

	doc.Find("div.story_item").Each(func(i int, s *goquery.Selection) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		result := &hibikemanga.SearchResult{
			Provider: string(ManganatoProvider),
		}

		href, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		if !strings.HasPrefix(href, "https://natomanga.com/") &&
			!strings.HasPrefix(href, "https://www.natomanga.com/") &&
			!strings.HasPrefix(href, "https://www.chapmanganato.com/") &&
			!strings.HasPrefix(href, "https://chapmanganato.com/") {
			return
		}

		result.ID = href
		splitHref := strings.Split(result.ID, "/")

		if strings.Contains(href, "chapmanganato") {
			result.ID = "chapmanganato$"
		} else {
			result.ID = "manganato$"
		}

		if len(splitHref) > 4 {
			result.ID += splitHref[4]
		}

		result.Title = s.Find("h3.story_name").Text()
		result.Title = strings.TrimSpace(result.Title)
		result.Image, _ = s.Find("img").Attr("src")

		compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, []*string{&result.Title})
		result.SearchRating = compRes.Rating
		ret = append(ret, result)
	})

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
		uri = fmt.Sprintf("https://natomanga.com/manga/%s", splitId[1])
	} else if splitId[0] == "chapmanganato" {
		uri = fmt.Sprintf("https://chapmanganato.com/manga/%s", splitId[1])
	}

	resp, err := mp.Client.R().
		SetHeader("User-Agent", util.GetRandomUserAgent()).
		Get(uri)

	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to send request")
		return nil, err
	}

	if !resp.IsSuccessState() {
		mp.logger.Error().Str("status", resp.Status).Str("uri", uri).Msg("manganato: Request failed")
		return nil, fmt.Errorf("failed to fetch chapters: status %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		mp.logger.Error().Err(err).Msg("manganato: Failed to parse HTML")
		return nil, err
	}

	doc.Find(".chapter-list .row").Each(func(i int, s *goquery.Selection) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		name := s.Find("a").Text()
		if strings.HasPrefix(name, "Vol.") {
			split := strings.Split(name, " ")
			name = strings.Join(split[1:], " ")
		}

		chStr := strings.TrimSpace(strings.Split(name, " ")[1])
		chStr = strings.TrimSuffix(chStr, ":")

		href, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		hrefParts := strings.Split(href, "/")
		if len(hrefParts) < 6 {
			return
		}

		chapterId := hrefParts[5]
		chapter := &hibikemanga.ChapterDetails{
			Provider: string(ManganatoProvider),
			ID:       splitId[1] + "$" + chapterId,
			URL:      href,
			Title:    strings.TrimSpace(name),
			Chapter:  chStr,
		}
		ret = append(ret, chapter)
	})

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

	uri := fmt.Sprintf("https://natomanga.com/manga/%s/%s", splitId[0], splitId[1])

	resp, err := mp.Client.R().
		SetHeader("User-Agent", util.GetRandomUserAgent()).
		SetHeader("Referer", "https://natomanga.com/").
		Get(uri)

	if err != nil {
		mp.logger.Error().Err(err).Str("uri", uri).Msg("manganato: Failed to send request")
		return nil, err
	}

	if !resp.IsSuccessState() {
		mp.logger.Error().Str("status", resp.Status).Str("uri", uri).Msg("manganato: Request failed")
		return nil, fmt.Errorf("failed to fetch chapter pages: status %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		mp.logger.Error().Err(err).Msg("manganato: Failed to parse HTML")
		return nil, err
	}

	doc.Find(".container-chapter-reader img").Each(func(i int, s *goquery.Selection) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		src, exists := s.Attr("src")
		if !exists || src == "" {
			return
		}

		page := &hibikemanga.ChapterPage{
			Provider: string(ManganatoProvider),
			URL:      src,
			Index:    len(ret),
			Headers: map[string]string{
				"Referer": "https://natomanga.com/",
			},
		}
		ret = append(ret, page)
	})

	if len(ret) == 0 {
		mp.logger.Error().Str("chapterId", id).Msg("manganato: No pages found")
		return nil, ErrNoPages
	}

	mp.logger.Info().Int("count", len(ret)).Msg("manganato: Found pages")

	return ret, nil

}
