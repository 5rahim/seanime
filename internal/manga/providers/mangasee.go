package manga_providers

import (
	"errors"
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"net/http"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strconv"
	"strings"
	"time"
)

type (
	Mangasee struct {
		Url       string
		UserAgent string
		Client    *http.Client
		logger    *zerolog.Logger
	}

	MangaseeResultItem struct {
		S  string   `json:"s"`  // Title
		I  string   `json:"i"`  // Slug
		G  []string `json:"g"`  // Genres
		A  []string `json:"a"`  // Synonyms
		AL []string `json:"al"` // Synonyms
		PS string   `json:"ps"` // Ongoing
		O  string   `json:"o"`
		SS string   `json:"ss"`
		T  string   `json:"t"`
		V  string   `json:"v"`
		VM string   `json:"vm"`
		Y  string   `json:"y"`
		L  string   `json:"l"`
		LT string   `json:"lt"`
		H  string   `json:"h"`
	}
)

func NewMangasee(logger *zerolog.Logger) *Mangasee {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &Mangasee{
		Url:       "https://mangasee123.com",
		Client:    c,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

// DEVNOTE: The ID returned by the Search function is the slug of the manga
// DEVNOTE: Each chapter has an ID in the format: {slug}${chapter_number} -- e.g. Jujutsu-Kaisen$0001
// This ID is split by the $ character to reconstruct the chapter URL for subsequent requests

func (m *Mangasee) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {

	m.logger.Debug().Str("query", opts.Query).Msg("mangasee: Searching manga")

	searchUrl := fmt.Sprintf("%s/_search.php", m.Url)
	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to create request")
		return nil, err
	}
	req.Header.Set("Referer", m.Url)
	req.Header.Set("User-Agent", m.UserAgent)

	res, err := m.Client.Do(req)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to send request")
		return nil, err
	}
	defer res.Body.Close()

	var result []*MangaseeResultItem
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to decode response")
		return nil, err
	}

	var searchResults []*hibikemanga.SearchResult
	for _, item := range result {
		titles := make([]*string, 0)
		titles = append(titles, &item.S)
		for _, s := range item.A {
			titles = append(titles, &s)
		}
		compRes, ok := comparison.FindBestMatchWithSorensenDice(&opts.Query, titles)
		if !ok {
			continue
		}
		if compRes.Rating < 0.6 {
			continue
		}

		searchResults = append(searchResults, &hibikemanga.SearchResult{
			ID:           item.I,
			Title:        item.S,
			Synonyms:     item.A,
			Year:         0,
			Image:        "",
			Provider:     MangaseeProvider,
			SearchRating: compRes.Rating,
		})
	}

	if len(searchResults) == 0 {
		m.logger.Error().Msg("mangasee: No results found")
		return nil, ErrNoResults
	}

	m.logger.Info().Int("count", len(searchResults)).Msg("mangasee: Found results")

	return searchResults, nil
}

func (m *Mangasee) FindChapters(slug string) ([]*hibikemanga.ChapterDetails, error) {

	m.logger.Debug().Str("mangaId", slug).Msg("mangasee: Fetching chapters")

	chapterUrl := fmt.Sprintf("%s/manga/%s", m.Url, slug)

	c := colly.NewCollector(
		colly.UserAgent(m.UserAgent),
	)

	c.WithTransport(m.Client.Transport)

	var chapterData []struct {
		Chapter     string `json:"Chapter"`
		Type        string `json:"Type"`
		Date        string `json:"Date"`
		ChapterName string `json:"ChapterName"`
	}
	c.OnHTML("body > script:nth-child(16)", func(e *colly.HTMLElement) {
		m.getChapterData(e.Text, 0, &chapterData)
	})

	err := c.Visit(chapterUrl)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to visit chapter url")
		return nil, err
	}

	if chapterData == nil || len(chapterData) == 0 {
		m.logger.Error().Msg("mangasee: Failed to find chapter data")
		return nil, errors.New("failed to find chapter data")
	}

	slices.Reverse(chapterData)

	ret := make([]*hibikemanga.ChapterDetails, len(chapterData))
	for i, chapter := range chapterData {
		chStr := getChapterNumber(chapter.Chapter)

		unpaddedChStr := strings.TrimLeft(chStr, "0")
		if unpaddedChStr == "" {
			unpaddedChStr = "0"
		}

		ret[i] = &hibikemanga.ChapterDetails{
			Provider: MangaseeProvider,
			ID:       slug + "$" + chStr, // e.g. One-Piece$0001
			Title:    fmt.Sprintf("Chapter %s", unpaddedChStr),
			URL:      fmt.Sprintf("%s/read-online/%s-chapter-%s-page-1.html", m.Url, slug, chStr),
			Chapter:  unpaddedChStr,
			Index:    uint(i),
		}
	}

	if len(ret) == 0 {
		m.logger.Error().Msg("mangasee: No chapters found")
		return nil, ErrNoChapters
	}

	m.logger.Info().Int("count", len(ret)).Msg("mangasee: Found chapters")

	return ret, nil
}

func (m *Mangasee) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {

	if !strings.Contains(id, "$") {
		m.logger.Error().Str("chapterId", id).Msg("mangasee: Invalid chapter id")
		return nil, errors.New("invalid chapter id")
	}

	info := strings.Split(id, "$")
	if len(info) != 2 {
		m.logger.Error().Str("chapterId", id).Msg("mangasee: Invalid chapter id")
		return nil, errors.New("invalid chapter id")
	}

	slug := info[0]
	chapter := info[1]
	uri := fmt.Sprintf("%s/read-online/%s-chapter-%s-page-1.html", m.Url, slug, chapter)

	pages := make([]*hibikemanga.ChapterPage, 0)

	c := colly.NewCollector(
		colly.UserAgent(m.UserAgent),
	)

	c.WithTransport(m.Client.Transport)

	var curChapter struct {
		Chapter     string `json:"Chapter"`
		Type        string `json:"Type"`
		Date        string `json:"Date"`
		ChapterName string `json:"ChapterName"`
		Page        string `json:"Page"`
	}
	var curPathname string

	c.OnHTML("body > script:nth-child(19)", func(e *colly.HTMLElement) {
		m.getChapterData(e.Text, 1, &curChapter)
		m.getChapterData(e.Text, 2, &curPathname)
	})

	err := c.Visit(uri)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to visit chapter url")
		return nil, err
	}

	if curChapter.Chapter == "" {
		m.logger.Error().Msg("mangasee: Failed to find current chapter data")
		return nil, errors.New("failed to find current chapter data")
	}
	if curPathname == "" {
		m.logger.Error().Msg("mangasee: Failed to find pathname")
		return nil, errors.New("failed to find pathname")
	}

	pageCount, err := strconv.Atoi(curChapter.Page)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to convert page count")
		return nil, errors.New("failed to convert page count")
	}

	for i := 0; i < pageCount; i++ {
		pageNum := strings.Repeat("0", 3-len(strconv.Itoa(i+1))) + strconv.Itoa(i+1)
		ch := getChapterForImageUrl(getChapterNumber(curChapter.Chapter))

		pages = append(pages, &hibikemanga.ChapterPage{
			Provider: string(MangaseeProvider),
			URL:      fmt.Sprintf("https://%s/manga/%s/%s-%s.png", curPathname, slug, ch, pageNum),
			Index:    i,
			Headers:  map[string]string{"Referer": m.Url},
		})
	}

	if len(pages) == 0 {
		m.logger.Error().Msg("mangasee: No pages found")
		return nil, ErrNoPages
	}

	m.logger.Info().Int("count", len(pages)).Msg("mangasee: Found pages")

	return pages, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Mangasee) getChapterData(script string, i int, ret interface{}) {
	variable := "vm.Chapters = "
	if i == 1 {
		variable = "vm.CurChapter ="
	} else if i == 2 {
		variable = "vm.CurPathName ="
	}
	chopFront := script[strings.Index(script, variable)+len(variable):]
	semicolonIndex := strings.Index(chopFront, ";")
	if semicolonIndex == -1 {
		return
	}
	chapters := chopFront[:semicolonIndex]

	err := json.Unmarshal([]byte(chapters), &ret)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: Failed to unmarshal chapter data")
	}
}

func getChapterNumber(ch string) string {
	if len(ch) == 0 {
		return ch
	}

	decimal := ch[len(ch)-1:]
	if len(ch) > 1 {
		ch = ch[1 : len(ch)-1]
	}

	if decimal == "0" {
		return ch
	}

	if strings.HasPrefix(ch, "0") {
		ch = ch[1:]
	}

	return ch + "." + decimal
}

func getChapterForImageUrl(chapter string) string {
	if !strings.Contains(chapter, ".") {
		return strings.Repeat("0", 4-len(chapter)) + chapter
	}

	values := strings.Split(chapter, ".")
	pad := strings.Repeat("0", 4-len(values[0])) + values[0]

	return pad + "." + values[1]
}
