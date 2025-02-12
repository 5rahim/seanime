package manga_providers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"

	"seanime/internal/util"
	"seanime/internal/util/comparison"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

// WeebCentral implements the manga provider for WeebCentral
// It uses goquery to scrape search results, chapter lists, and chapter pages.

type WeebCentral struct {
	Url       string
	UserAgent string
	Client    *http.Client
	logger    *zerolog.Logger
}

// NewWeebCentral initializes and returns a new WeebCentral provider instance.
func NewWeebCentral(logger *zerolog.Logger) *WeebCentral {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	// Optionally, add transport modifications if necessary.
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

	req, err := http.NewRequest("POST", searchUrl, strings.NewReader(form.Encode()))
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to create search request")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Trigger", "quick-search-input")
	req.Header.Set("HX-Trigger-Name", "text")
	req.Header.Set("HX-Target", "quick-search-result")
	req.Header.Set("HX-Current-URL", w.Url+"/")
	req.Header.Set("User-Agent", w.UserAgent)

	res, err := w.Client.Do(req)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to send search request")
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to read search response")
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
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
	req, err := http.NewRequest("GET", chapterUrl, nil)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to create chapter list request")
		return nil, err
	}

	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "chapter-list")
	req.Header.Set("HX-Current-URL", fmt.Sprintf("%s/series/%s", w.Url, mangaId))
	req.Header.Set("Referer", fmt.Sprintf("%s/series/%s", w.Url, mangaId))
	req.Header.Set("User-Agent", w.UserAgent)

	res, err := w.Client.Do(req)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to fetch chapter list")
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to read chapter list response")
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to create chapter pages request")
		return nil, err
	}

	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Current-URL", fmt.Sprintf("%s/chapters/%s", w.Url, chapterId))
	req.Header.Set("Referer", fmt.Sprintf("%s/chapters/%s", w.Url, chapterId))
	req.Header.Set("User-Agent", w.UserAgent)

	res, err := w.Client.Do(req)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to fetch chapter pages")
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.logger.Error().Err(err).Msg("weebcentral: Failed to read chapter pages response")
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
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
