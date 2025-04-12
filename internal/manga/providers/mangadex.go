package manga_providers

import (
	"cmp"
	"encoding/json"
	"fmt"
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
	Mangadex struct {
		Url       string
		BaseUrl   string
		UserAgent string
		Client    *http.Client
		logger    *zerolog.Logger
	}

	MangadexManga struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		Attributes    MangadexMangeAttributes
		Relationships []MangadexMangaRelationship `json:"relationships"`
	}

	MangadexMangeAttributes struct {
		AltTitles []map[string]string `json:"altTitles"`
		Title     map[string]string   `json:"title"`
		Year      int                 `json:"year"`
	}

	MangadexMangaRelationship struct {
		ID         string                 `json:"id"`
		Type       string                 `json:"type"`
		Related    string                 `json:"related"`
		Attributes map[string]interface{} `json:"attributes"`
	}

	MangadexErrorResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}

	MangadexChapterData struct {
		ID         string                    `json:"id"`
		Attributes MangadexChapterAttributes `json:"attributes"`
	}

	MangadexChapterAttributes struct {
		Title     string `json:"title"`
		Volume    string `json:"volume"`
		Chapter   string `json:"chapter"`
		UpdatedAt string `json:"updatedAt"`
	}
)

// DEVNOTE: Each chapter ID is a unique string provided by Mangadex

func NewMangadex(logger *zerolog.Logger) *Mangadex {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &Mangadex{
		Url:       "https://api.mangadex.org",
		BaseUrl:   "https://mangadex.org",
		Client:    c,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (md *Mangadex) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (md *Mangadex) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	ret := make([]*hibikemanga.SearchResult, 0)

	retManga := make([]*MangadexManga, 0)

	for i := range 1 {
		uri := fmt.Sprintf("%s/manga?title=%s&limit=25&offset=%d&order[relevance]=desc&contentRating[]=safe&contentRating[]=suggestive&includes[]=cover_art", md.Url, url.QueryEscape(opts.Query), 25*i)

		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			md.logger.Error().Err(err).Msg("mangadex: Failed to create request")
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", md.UserAgent)
		req.Header.Set("Referer", "https://google.com")

		resp, err := md.Client.Do(req)
		if err != nil {
			return nil, err
		}

		var data struct {
			Data []*MangadexManga `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			md.logger.Error().Err(err).Msg("mangadex: Failed to decode response")
			_ = resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		_ = resp.Body.Close()

		retManga = append(retManga, data.Data...)
	}

	for _, manga := range retManga {
		var altTitles []string
		for _, title := range manga.Attributes.AltTitles {
			altTitle, ok := title["en"]
			if ok {
				altTitles = append(altTitles, altTitle)
			}
			altTitle, ok = title["jp"]
			if ok {
				altTitles = append(altTitles, altTitle)
			}
			altTitle, ok = title["ja"]
			if ok {
				altTitles = append(altTitles, altTitle)
			}
		}
		t := getTitle(manga.Attributes)

		var img string
		for _, relation := range manga.Relationships {
			if relation.Type == "cover_art" {
				fn, ok := relation.Attributes["fileName"].(string)
				if ok {
					img = fmt.Sprintf("%s/covers/%s/%s.512.jpg", md.BaseUrl, manga.ID, fn)
				} else {
					img = fmt.Sprintf("%s/covers/%s/%s.jpg.512.jpg", md.BaseUrl, manga.ID, relation.ID)
				}
			}
		}

		format := strings.ToUpper(manga.Type)
		if format == "ADAPTATION" {
			format = "MANGA"
		}

		compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, []*string{&t})

		result := &hibikemanga.SearchResult{
			ID:           manga.ID,
			Title:        t,
			Synonyms:     altTitles,
			Image:        img,
			Year:         manga.Attributes.Year,
			SearchRating: compRes.Rating,
			Provider:     string(MangadexProvider),
		}

		ret = append(ret, result)
	}

	if len(ret) == 0 {
		md.logger.Error().Msg("mangadex: No results found")
		return nil, ErrNoResults
	}

	md.logger.Info().Int("count", len(ret)).Msg("mangadex: Found results")

	return ret, nil
}
func (md *Mangadex) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	ret := make([]*hibikemanga.ChapterDetails, 0)

	md.logger.Debug().Str("mangaId", id).Msg("mangadex: Finding chapters")

	for page := 0; page <= 1; page++ {
		uri := fmt.Sprintf("%s/manga/%s/feed?limit=500&translatedLanguage%%5B%%5D=en&includes[]=scanlation_group&includes[]=user&order[volume]=desc&order[chapter]=desc&offset=%d&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic", md.Url, id, 500*page)

		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}

		var data struct {
			Result string                  `json:"result"`
			Errors []MangadexErrorResponse `json:"errors"`
			Data   []MangadexChapterData   `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			md.logger.Error().Err(err).Msg("mangadex: Failed to decode response")
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if data.Result == "error" {
			md.logger.Error().Str("error", data.Errors[0].Title).Str("detail", data.Errors[0].Detail).Msg("mangadex: Could not find chapters")
			return nil, fmt.Errorf("could not find chapters: %s", data.Errors[0].Detail)
		}

		slices.Reverse(data.Data)

		chapterMap := make(map[string]*hibikemanga.ChapterDetails)
		idx := uint(len(ret))
		for _, chapter := range data.Data {

			if chapter.Attributes.Chapter == "" {
				continue
			}

			title := "Chapter " + fmt.Sprintf("%s", chapter.Attributes.Chapter) + " "

			if _, ok := chapterMap[chapter.Attributes.Chapter]; ok {
				continue
			}

			chapterMap[chapter.Attributes.Chapter] = &hibikemanga.ChapterDetails{
				ID:        chapter.ID,
				Title:     title,
				Index:     idx,
				Chapter:   chapter.Attributes.Chapter,
				UpdatedAt: chapter.Attributes.UpdatedAt,
				Provider:  string(MangadexProvider),
			}
			idx++
		}

		chapters := make([]*hibikemanga.ChapterDetails, 0, len(chapterMap))
		for _, chapter := range chapterMap {
			chapters = append(chapters, chapter)
		}

		slices.SortStableFunc(chapters, func(i, j *hibikemanga.ChapterDetails) int {
			return cmp.Compare(i.Index, j.Index)
		})

		if len(chapters) > 0 {
			ret = append(ret, chapters...)
		} else {
			break
		}
	}

	if len(ret) == 0 {
		md.logger.Error().Msg("mangadex: No chapters found")
		return nil, ErrNoChapters
	}

	md.logger.Info().Int("count", len(ret)).Msg("mangadex: Found chapters")

	return ret, nil

}
func (md *Mangadex) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {
	ret := make([]*hibikemanga.ChapterPage, 0)

	md.logger.Debug().Str("chapterId", id).Msg("mangadex: Finding chapter pages")

	uri := fmt.Sprintf("%s/at-home/server/%s", md.Url, id)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		md.logger.Error().Err(err).Msg("mangadex: Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", util.GetRandomUserAgent())

	resp, err := md.Client.Do(req)
	if err != nil {
		md.logger.Error().Err(err).Msg("mangadex: Failed to get chapter pages")
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		BaseUrl string `json:"baseUrl"`
		Chapter struct {
			Hash string   `json:"hash"`
			Data []string `json:"data"`
		}
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		md.logger.Error().Err(err).Msg("mangadex: Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	for i, page := range data.Chapter.Data {
		ret = append(ret, &hibikemanga.ChapterPage{
			Provider: string(MangadexProvider),
			URL:      fmt.Sprintf("%s/data/%s/%s", data.BaseUrl, data.Chapter.Hash, page),
			Index:    i,
			Headers: map[string]string{
				"Referer": "https://mangadex.org",
			},
		})
	}

	if len(ret) == 0 {
		md.logger.Error().Msg("mangadex: No pages found")
		return nil, ErrNoPages
	}

	md.logger.Info().Int("count", len(ret)).Msg("mangadex: Found pages")

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getTitle(attributes MangadexMangeAttributes) string {
	altTitles := attributes.AltTitles
	title := attributes.Title

	enTitle := title["en"]
	if enTitle != "" {
		return enTitle
	}

	var enAltTitle string
	for _, altTitle := range altTitles {
		if value, ok := altTitle["en"]; ok {
			enAltTitle = value
			break
		}
	}

	if enAltTitle != "" && util.IsMostlyLatinString(enAltTitle) {
		return enAltTitle
	}

	// Check for other language titles
	if jaRoTitle, ok := title["ja-ro"]; ok {
		return jaRoTitle
	}
	if jpRoTitle, ok := title["jp-ro"]; ok {
		return jpRoTitle
	}
	if jpTitle, ok := title["jp"]; ok {
		return jpTitle
	}
	if jaTitle, ok := title["ja"]; ok {
		return jaTitle
	}
	if koTitle, ok := title["ko"]; ok {
		return koTitle
	}

	// Check alt titles for other languages
	for _, altTitle := range altTitles {
		if value, ok := altTitle["ja-ro"]; ok {
			return value
		}
	}
	for _, altTitle := range altTitles {
		if value, ok := altTitle["jp-ro"]; ok {
			return value
		}
	}
	for _, altTitle := range altTitles {
		if value, ok := altTitle["jp"]; ok {
			return value
		}
	}
	for _, altTitle := range altTitles {
		if value, ok := altTitle["ja"]; ok {
			return value
		}
	}
	for _, altTitle := range altTitles {
		if value, ok := altTitle["ko"]; ok {
			return value
		}
	}

	return ""
}
