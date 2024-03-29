package manga_providers

import (
	"cmp"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"net/http"
	"strconv"
	"strings"
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
	return &Mangasee{
		Url:       "https://mangasee123.com",
		Client:    &http.Client{},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		logger:    logger,
	}
}

func (m *Mangasee) Search(opts SearchOptions) ([]*SearchResult, error) {

	searchUrl := fmt.Sprintf("%s/_search.php", m.Url)
	req, err := http.NewRequest("POST", searchUrl, nil)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: failed to create request")
		return nil, err
	}
	req.Header.Set("Referer", m.Url)

	res, err := m.Client.Do(req)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: failed to send request")
		return nil, err
	}
	defer res.Body.Close()

	var result []*MangaseeResultItem
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		m.logger.Error().Err(err).Msg("mangasee: failed to decode response")
		return nil, err
	}

	var searchResults []*SearchResult
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

		searchResults = append(searchResults, &SearchResult{
			ID:           item.I,
			Title:        item.S,
			Synonyms:     item.A,
			Year:         0,
			Image:        "",
			Provider:     MangaseeProvider,
			SearchRating: compRes.Rating,
		})
	}

	return searchResults, nil
}

func (m *Mangasee) FindChapters(slug string) ([]*ChapterDetails, error) {
	chapterUrl := fmt.Sprintf("%s/manga/%s", m.Url, slug)

	c := colly.NewCollector(
		colly.UserAgent(m.UserAgent),
	)

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
		m.logger.Error().Err(err).Msg("mangasee: failed to visit chapter url")
		return nil, err
	}

	if chapterData == nil || len(chapterData) == 0 {
		m.logger.Error().Msg("mangasee: failed to find chapter data")
		return nil, errors.New("failed to find chapter data")
	}

	ret := make([]*ChapterDetails, len(chapterData))
	for i, chapter := range chapterData {
		intChapterNum, _ := strconv.Atoi(getChapterNumber(chapter.Chapter))
		ret[i] = &ChapterDetails{
			Provider:  MangaseeProvider,
			Slug:      slug, // e.g. One-Piece
			Title:     cmp.Or(chapter.ChapterName, fmt.Sprintf("Chapter %d", intChapterNum)),
			URL:       fmt.Sprintf("%s/read-online/%s-chapter-%d-page-1.html", m.Url, slug, intChapterNum),
			Number:    intChapterNum,
			Rating:    0,
			UpdatedAt: 0,
		}
	}

	return ret, nil
}

func (m *Mangasee) FindChapterPages(info *ChapterDetails) ([]*ChapterPage, error) {

	pages := make([]*ChapterPage, 0)

	c := colly.NewCollector(
		colly.UserAgent(m.UserAgent),
	)

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

	err := c.Visit(info.URL)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: failed to visit chapter url")
		return nil, err
	}

	if curChapter.Chapter == "" {
		m.logger.Error().Msg("mangasee: failed to find current chapter data")
		return nil, errors.New("failed to find current chapter data")
	}
	if curPathname == "" {
		m.logger.Error().Msg("mangasee: failed to find pathname")
		return nil, errors.New("failed to find pathname")
	}

	pageCount, err := strconv.Atoi(curChapter.Page)
	if err != nil {
		m.logger.Error().Err(err).Msg("mangasee: failed to convert page count")
		return nil, errors.New("mangasee: failed to convert page count")
	}

	for i := 1; i <= pageCount; i++ {
		pageNum := strings.Repeat("0", 3-len(strconv.Itoa(i))) + strconv.Itoa(i)
		ch := getChapterForImageUrl(getChapterNumber(curChapter.Chapter))

		pages = append(pages, &ChapterPage{
			Provider: MangaseeProvider,
			Url:      fmt.Sprintf("https://%s/manga/%s/%s-%s.png", curPathname, info.Slug, ch, pageNum),
			Index:    i,
			Headers:  map[string]string{"Referer": info.URL},
		})

	}

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
		m.logger.Error().Err(err).Msg("mangasee: failed to unmarshal chapter data")
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
