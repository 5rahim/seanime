package manga_providers

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
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
	ComicK struct {
		Url    string
		Client *http.Client
		logger *zerolog.Logger
	}

	ComicKResultItem struct {
		ID                   int     `json:"id"`
		HID                  string  `json:"hid"`
		Slug                 string  `json:"slug"`
		Title                string  `json:"title"`
		Country              string  `json:"country"`
		Rating               string  `json:"rating"`
		BayesianRating       string  `json:"bayesian_rating"`
		RatingCount          int     `json:"rating_count"`
		FollowCount          int     `json:"follow_count"`
		Description          string  `json:"desc"`
		Status               int     `json:"status"`
		LastChapter          float64 `json:"last_chapter"`
		TranslationCompleted bool    `json:"translation_completed"`
		ViewCount            int     `json:"view_count"`
		ContentRating        string  `json:"content_rating"`
		Demographic          int     `json:"demographic"`
		UploadedAt           string  `json:"uploaded_at"`
		Genres               []int   `json:"genres"`
		CreatedAt            string  `json:"created_at"`
		UserFollowCount      int     `json:"user_follow_count"`
		Year                 int     `json:"year"`
		MuComics             struct {
			Year int `json:"year"`
		} `json:"mu_comics"`
		MdTitles []struct {
			Title string `json:"title"`
		} `json:"md_titles"`
		MdCovers []struct {
			W     int    `json:"w"`
			H     int    `json:"h"`
			B2Key string `json:"b2key"`
		} `json:"md_covers"`
		Highlight string `json:"highlight"`
	}
)

func NewComicK(logger *zerolog.Logger) *ComicK {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	//c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &ComicK{
		Url:    "https://api.comick.fun",
		Client: c,
		logger: logger,
	}
}

// DEVNOTE: Each chapter ID is a unique string provided by ComicK

func (c *ComicK) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (c *ComicK) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	searchUrl := fmt.Sprintf("%s/v1.0/search?q=%s&limit=25&page=1", c.Url, url.QueryEscape(opts.Query))
	if opts.Year != 0 {
		searchUrl += fmt.Sprintf("&from=%d&to=%d", opts.Year, opts.Year)
	}

	c.logger.Debug().Str("searchUrl", searchUrl).Msg("comick: Searching manga")

	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	//req.Header.Set("User-Agent", util.GetRandomUserAgent())
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data []*ComicKResultItem
	if err := json.Unmarshal(body, &data); err != nil {
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

func (c *ComicK) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	ret := make([]*hibikemanga.ChapterDetails, 0)

	c.logger.Debug().Str("mangaId", id).Msg("comick: Fetching chapters")

	uri := fmt.Sprintf("%s/comic/%s/chapters?lang=en&page=0&limit=1000000&chap-order=1", c.Url, id)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		c.logger.Error().Err(err).Msg("comick: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	//req.Header.Set("User-Agent", util.GetRandomUserAgent())
	req.Header.Set("User-Agent", "Mozilla/5.0")

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
	chaptersMap := make(map[string]*hibikemanga.ChapterDetails)
	count := 0
	for _, chapter := range data.Chapters {
		if chapter.Chap == "" || chapter.Lang != "en" {
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

		prev, ok := chaptersMap[chapter.Chap]
		rating := chapter.UpCount - chapter.DownCount

		if !ok || rating > prev.Rating {
			if !ok {
				count++
			}
			chaptersMap[chapter.Chap] = &hibikemanga.ChapterDetails{
				Provider:  ComickProvider,
				ID:        chapter.HID,
				Title:     title,
				Index:     uint(count),
				URL:       fmt.Sprintf("%s/chapter/%s", c.Url, chapter.HID),
				Chapter:   chapter.Chap,
				Rating:    rating,
				UpdatedAt: chapter.UpdatedAt,
			}
		}
	}

	for _, chapter := range chaptersMap {
		chapters = append(chapters, chapter)
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

func (c *ComicK) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Comic struct {
	ID      int    `json:"id"`
	HID     string `json:"hid"`
	Title   string `json:"title"`
	Country string `json:"country"`
	Status  int    `json:"status"`
	Links   struct {
		AL  string `json:"al"`
		AP  string `json:"ap"`
		BW  string `json:"bw"`
		KT  string `json:"kt"`
		MU  string `json:"mu"`
		AMZ string `json:"amz"`
		CDJ string `json:"cdj"`
		EBJ string `json:"ebj"`
		MAL string `json:"mal"`
		RAW string `json:"raw"`
	} `json:"links"`
	LastChapter          interface{}   `json:"last_chapter"`
	ChapterCount         int           `json:"chapter_count"`
	Demographic          int           `json:"demographic"`
	Hentai               bool          `json:"hentai"`
	UserFollowCount      int           `json:"user_follow_count"`
	FollowRank           int           `json:"follow_rank"`
	CommentCount         int           `json:"comment_count"`
	FollowCount          int           `json:"follow_count"`
	Description          string        `json:"desc"`
	Parsed               string        `json:"parsed"`
	Slug                 string        `json:"slug"`
	Mismatch             interface{}   `json:"mismatch"`
	Year                 int           `json:"year"`
	BayesianRating       interface{}   `json:"bayesian_rating"`
	RatingCount          int           `json:"rating_count"`
	ContentRating        string        `json:"content_rating"`
	TranslationCompleted bool          `json:"translation_completed"`
	RelateFrom           []interface{} `json:"relate_from"`
	Mies                 interface{}   `json:"mies"`
	MdTitles             []struct {
		Title string `json:"title"`
	} `json:"md_titles"`
	MdComicMdGenres []struct {
		MdGenres struct {
			Name  string      `json:"name"`
			Type  interface{} `json:"type"`
			Slug  string      `json:"slug"`
			Group string      `json:"group"`
		} `json:"md_genres"`
	} `json:"md_comic_md_genres"`
	MuComics struct {
		LicensedInEnglish interface{} `json:"licensed_in_english"`
		MuComicCategories []struct {
			MuCategories struct {
				Title string `json:"title"`
				Slug  string `json:"slug"`
			} `json:"mu_categories"`
			PositiveVote int `json:"positive_vote"`
			NegativeVote int `json:"negative_vote"`
		} `json:"mu_comic_categories"`
	} `json:"mu_comics"`
	MdCovers []struct {
		Vol   interface{} `json:"vol"`
		W     int         `json:"w"`
		H     int         `json:"h"`
		B2Key string      `json:"b2key"`
	} `json:"md_covers"`
	Iso6391    string `json:"iso639_1"`
	LangName   string `json:"lang_name"`
	LangNative string `json:"lang_native"`
}

type ComicChapter struct {
	ID        int      `json:"id"`
	Chap      string   `json:"chap"`
	Title     string   `json:"title"`
	Vol       string   `json:"vol,omitempty"`
	Lang      string   `json:"lang"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	UpCount   int      `json:"up_count"`
	DownCount int      `json:"down_count"`
	GroupName []string `json:"group_name"`
	HID       string   `json:"hid"`
	MdImages  []struct {
		Name  string `json:"name"`
		W     int    `json:"w"`
		H     int    `json:"h"`
		S     int    `json:"s"`
		B2Key string `json:"b2key"`
	} `json:"md_images"`
}
