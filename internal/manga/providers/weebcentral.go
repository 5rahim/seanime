package manga_providers

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

// WeebCentral implements the manga provider for WeebCentral
// It uses goquery to scrape search results, chapter lists, and chapter pages.

type WeebCentral struct {
	Url       string
	UserAgent string
	Client    *req.Client
	logger    *zerolog.Logger
}

// NewWeebCentral initializes and returns a new WeebCentral provider instance.
func NewWeebCentral(logger *zerolog.Logger) *WeebCentral {
	client := req.C().
		SetUserAgent(util.GetRandomUserAgent()).
		SetTimeout(60 * time.Second).
		EnableInsecureSkipVerify().
		ImpersonateChrome()

	return &WeebCentral{
		Url:       "https://weebcentral.com",
		UserAgent: util.GetRandomUserAgent(),
		Client:    client,
		logger:    logger,
	}
}

func (w *WeebCentral) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (w *WeebCentral) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	w.logger.Debug().Str("query", opts.Query).Msg("weebcentral: Searching manga")

	searchUrl := fmt.Sprintf("%s/search/simple?location=main", w.Url)
	form := url.Values{}
	form.Set("text", opts.Query)

	resp, err := w.Client.R().
		SetContentType("application/x-www-form-urlencoded").
		SetHeader("HX-Request", "true").
		SetHeader("HX-Trigger", "quick-search-input").
		SetHeader("HX-Trigger-Name", "text").
		SetHeader("HX-Target", "quick-search-result").
		SetHeader("HX-Current-URL", w.Url+"/").
		SetBody(form.Encode()).
		Post(searchUrl)

	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to send search request")
		return nil, err
	}

	if !resp.IsSuccessState() {
		w.logger.Error().Str("status", resp.Status).Msg("weebcentral: Search request failed")
		return nil, fmt.Errorf("search request failed: status %s", resp.Status)
	}

	body := resp.String()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to parse search HTML")
		return nil, err
	}

	var searchResults []*hibikemanga.SearchResult
	doc.Find("#quick-search-result > div > a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		title := strings.TrimSpace(s.Find(".flex-1").Text())

		var image string
		if s.Find("source").Length() > 0 {
			image, _ = s.Find("source").Attr("srcset")
		} else if s.Find("img").Length() > 0 {
			image, _ = s.Find("img").Attr("src")
		}

		// Extract manga id from link assuming the format contains '/series/{id}/'
		idPart := ""
		parts := strings.Split(link, "/series/")
		if len(parts) > 1 {
			subparts := strings.Split(parts[1], "/")
			idPart = subparts[0]
		}
		if idPart == "" {
			return
		}

		titleCopy := title
		titles := []*string{&titleCopy}
		compRes, ok := comparison.FindBestMatchWithSorensenDice(&opts.Query, titles)
		if !ok || compRes.Rating < 0.6 {
			return
		}

		searchResults = append(searchResults, &hibikemanga.SearchResult{
			ID:           idPart,
			Title:        title,
			Synonyms:     []string{},
			Year:         0,
			Image:        image,
			Provider:     WeebCentralProvider,
			SearchRating: compRes.Rating,
		})
	})

	if len(searchResults) == 0 {
		w.logger.Error().Msg("weebcentral: No search results found")
		return nil, errors.New("no results found")
	}

	w.logger.Info().Int("count", len(searchResults)).Msg("weebcentral: Found search results")
	return searchResults, nil
}

func (w *WeebCentral) FindChapters(mangaId string) ([]*hibikemanga.ChapterDetails, error) {
	w.logger.Debug().Str("mangaId", mangaId).Msg("weebcentral: Fetching chapters")

	chapterUrl := fmt.Sprintf("%s/series/%s/full-chapter-list", w.Url, mangaId)

	resp, err := w.Client.R().
		SetHeader("HX-Request", "true").
		SetHeader("HX-Target", "chapter-list").
		SetHeader("HX-Current-URL", fmt.Sprintf("%s/series/%s", w.Url, mangaId)).
		SetHeader("Referer", fmt.Sprintf("%s/series/%s", w.Url, mangaId)).
		Get(chapterUrl)

	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to fetch chapter list")
		return nil, err
	}

	if !resp.IsSuccessState() {
		w.logger.Error().Str("status", resp.Status).Msg("weebcentral: Chapter list request failed")
		return nil, fmt.Errorf("chapter list request failed: status %s", resp.Status)
	}

	body := resp.String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to parse chapter list HTML")
		return nil, err
	}

	var chapters []*hibikemanga.ChapterDetails
	volumeCounter := 1
	lastChapterNumber := 9999.0

	chapterRegex := regexp.MustCompile("(\\d+(?:\\.\\d+)?)")

	doc.Find("div.flex.items-center").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		chapterUrl, exists := a.Attr("href")
		if !exists {
			return
		}
		chapterTitle := strings.TrimSpace(a.Find("span.grow > span").First().Text())

		var chapterNumber string
		var parsedChapterNumber float64

		match := chapterRegex.FindStringSubmatch(chapterTitle)
		if len(match) > 1 {
			chapterNumber = w.cleanChapterNumber(match[1])
			if num, err := strconv.ParseFloat(chapterNumber, 64); err == nil {
				parsedChapterNumber = num
			}
		} else {
			chapterNumber = ""
		}

		if parsedChapterNumber > lastChapterNumber {
			volumeCounter++
		}
		if parsedChapterNumber != 0 {
			lastChapterNumber = parsedChapterNumber
		}

		// Extract chapter id from the URL assuming format contains '/chapters/{id}'
		chapterId := ""
		parts := strings.Split(chapterUrl, "/chapters/")
		if len(parts) > 1 {
			chapterId = parts[1]
		}

		chapters = append(chapters, &hibikemanga.ChapterDetails{
			ID:       chapterId,
			URL:      chapterUrl,
			Title:    chapterTitle,
			Chapter:  chapterNumber,
			Index:    uint(i),
			Provider: WeebCentralProvider,
		})
	})

	if len(chapters) == 0 {
		w.logger.Error().Msg("weebcentral: No chapters found")
		return nil, errors.New("no chapters found")
	}

	slices.Reverse(chapters)

	for i := range chapters {
		chapters[i].Index = uint(i)
	}

	w.logger.Info().Int("count", len(chapters)).Msg("weebcentral: Found chapters")
	return chapters, nil
}

func (w *WeebCentral) FindChapterPages(chapterId string) ([]*hibikemanga.ChapterPage, error) {
	url := fmt.Sprintf("%s/chapters/%s/images?is_prev=False&reading_style=long_strip", w.Url, chapterId)

	resp, err := w.Client.R().
		SetHeader("HX-Request", "true").
		SetHeader("HX-Current-URL", fmt.Sprintf("%s/chapters/%s", w.Url, chapterId)).
		SetHeader("Referer", fmt.Sprintf("%s/chapters/%s", w.Url, chapterId)).
		Get(url)

	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to fetch chapter pages")
		return nil, err
	}

	if !resp.IsSuccessState() {
		w.logger.Error().Str("status", resp.Status).Msg("weebcentral: Chapter pages request failed")
		return nil, fmt.Errorf("chapter pages request failed: status %s", resp.Status)
	}

	body := resp.String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to parse chapter pages HTML")
		return nil, err
	}

	var pages []*hibikemanga.ChapterPage
	totalImgs := doc.Find("img").Length()

	doc.Find("section.flex-1 img").Each(func(i int, s *goquery.Selection) {
		imageUrl, exists := s.Attr("src")
		if !exists || imageUrl == "" {
			return
		}
		pages = append(pages, &hibikemanga.ChapterPage{
			URL:      imageUrl,
			Index:    i,
			Headers:  map[string]string{"Referer": w.Url},
			Provider: WeebCentralProvider,
		})
	})

	if len(pages) == 0 && totalImgs > 0 {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			imageUrl, exists := s.Attr("src")
			if !exists || imageUrl == "" {
				return
			}
			pages = append(pages, &hibikemanga.ChapterPage{
				URL:      imageUrl,
				Index:    i,
				Headers:  map[string]string{"Referer": w.Url},
				Provider: WeebCentralProvider,
			})
		})
	}

	if len(pages) == 0 {
		w.logger.Error().Msg("weebcentral: No pages found")
		return nil, errors.New("no pages found")
	}

	w.logger.Info().Int("count", len(pages)).Msg("weebcentral: Found chapter pages")
	return pages, nil
}

func (w *WeebCentral) cleanChapterNumber(chapterStr string) string {
	cleaned := strings.TrimLeft(chapterStr, "0")
	if cleaned == "" {
		return "0"
	}
	return cleaned
}
