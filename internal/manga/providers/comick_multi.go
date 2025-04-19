package manga_providers

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type (
	ComicKMulti struct {
		Url       string
		Client    *http.Client
		UserAgent string
		logger    *zerolog.Logger
	}
)

func NewComicKMulti(logger *zerolog.Logger) *ComicKMulti {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	//c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &ComicKMulti{
		Url:       "https://api.comick.fun",
		Client:    c,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

// DEVNOTE: Each chapter ID is a unique string provided by ComicK

func (c *ComicKMulti) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: true,
		SupportsMultiLanguage:  true,
	}
}

func (c *ComicKMulti) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {

	c.logger.Debug().Str("query", opts.Query).Msg("comick: Searching manga")

	searchUrl := fmt.Sprintf("%s/v1.0/search?q=%s&limit=25&page=1", c.Url, url.QueryEscape(opts.Query))
	if opts.Year != 0 {
		searchUrl += fmt.Sprintf("&from=%d&to=%d", opts.Year, opts.Year)
	}

	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", util.GetRandomUserAgent())

	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var data []*ComicKResultItem
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to decode response")
		return nil, fmt.Errorf("failed to reach API: %w", err)
	}

	results := make([]*hibikemanga.SearchResult, 0)
	for _, result := range data {

		// Skip fan-colored manga
		if strings.Contains(result.Slug, "fan-colored") {
			continue
		}

		var coverURL string
		if len(result.MdCovers) > 0 && result.MdCovers[0].B2Key != "" {
			coverURL = "https://meo.comick.pictures/" + result.MdCovers[0].B2Key
		}

		altTitles := make([]string, len(result.MdTitles))
		for j, title := range result.MdTitles {
			altTitles[j] = title.Title
		}

		// DEVNOTE: We don't compare to alt titles because ComicK's synonyms aren't good
		compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, []*string{&result.Title})

		results = append(results, &hibikemanga.SearchResult{
			ID:           result.HID,
			Title:        cmp.Or(result.Title, result.Slug),
			Synonyms:     altTitles,
			Image:        coverURL,
			Year:         result.Year,
			SearchRating: compRes.Rating,
			Provider:     ComickProvider,
		})
	}

	if len(results) == 0 {
		c.logger.Warn().Msg("comick: No results found")
		return nil, ErrNoChapters
	}

	c.logger.Info().Int("count", len(results)).Msg("comick: Found results")

	return results, nil
}

func (c *ComicKMulti) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	ret := make([]*hibikemanga.ChapterDetails, 0)

	// c.logger.Debug().Str("mangaId", id).Msg("comick: Fetching chapters")

	uri := fmt.Sprintf("%s/comic/%s/chapters?page=0&limit=1000000&chap-order=1", c.Url, id)
	c.logger.Debug().Str("mangaId", id).Str("uri", uri).Msg("comick: Fetching chapters")
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", util.GetRandomUserAgent())

	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Chapters []*ComicChapter `json:"chapters"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	chapters := make([]*hibikemanga.ChapterDetails, 0)
	chaptersCountMap := make(map[string]int)
	for _, chapter := range data.Chapters {
		if chapter.Chap == "" {
			continue
		}
		title := "Chapter " + chapter.Chap + " "

		if title == "" {
			if chapter.Title == "" {
				title = "Oneshot"
			} else {
				title = chapter.Title
			}
		}
		title = strings.TrimSpace(title)

		groupName := ""
		if len(chapter.GroupName) > 0 {
			groupName = chapter.GroupName[0]
		}

		count, ok := chaptersCountMap[groupName]
		if !ok {
			chaptersCountMap[groupName] = 0
			count = 0
		}
		chapters = append(chapters, &hibikemanga.ChapterDetails{
			Provider:  ComickProvider,
			ID:        chapter.HID,
			Title:     title,
			Language:  chapter.Lang,
			Index:     uint(count),
			URL:       fmt.Sprintf("%s/chapter/%s", c.Url, chapter.HID),
			Chapter:   chapter.Chap,
			Scanlator: groupName,
			Rating:    0,
			UpdatedAt: chapter.UpdatedAt,
		})
		chaptersCountMap[groupName]++
	}

	// Sort chapters by index
	slices.SortStableFunc(chapters, func(i, j *hibikemanga.ChapterDetails) int {
		return cmp.Compare(i.Index, j.Index)
	})

	ret = append(ret, chapters...)

	if len(ret) == 0 {
		c.logger.Warn().Msg("comick: No chapters found")
		return nil, ErrNoChapters
	}

	c.logger.Info().Int("count", len(ret)).Msg("comick: Found chapters")

	return ret, nil
}

func (c *ComicKMulti) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {
	ret := make([]*hibikemanga.ChapterPage, 0)

	c.logger.Debug().Str("chapterId", id).Msg("comick: Finding chapter pages")

	uri := fmt.Sprintf("%s/chapter/%s", c.Url, id)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", util.GetRandomUserAgent())

	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Chapter *ComicChapter `json:"chapter"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if data.Chapter == nil {
		c.logger.Error().Msg("comick: Chapter not found")
		return nil, fmt.Errorf("chapter not found")
	}

	for index, image := range data.Chapter.MdImages {
		ret = append(ret, &hibikemanga.ChapterPage{
			Provider: ComickProvider,
			URL:      fmt.Sprintf("https://meo.comick.pictures/%s", image.B2Key),
			Index:    index,
			Headers:  make(map[string]string),
		})
	}

	if len(ret) == 0 {
		c.logger.Warn().Msg("comick: No pages found")
		return nil, ErrNoPages
	}

	c.logger.Info().Int("count", len(ret)).Msg("comick: Found pages")

	return ret, nil

}
