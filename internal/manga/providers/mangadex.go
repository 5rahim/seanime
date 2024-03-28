package manga_providers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"net/http"
	"net/url"
	"slices"
	"strings"
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

func NewMangadex(logger *zerolog.Logger) *Mangadex {
	c := &http.Client{}
	c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &Mangadex{
		Url:       "https://api.mangadex.org",
		BaseUrl:   "https://mangadex.org",
		Client:    c,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (md *Mangadex) Search(opts SearchOptions) ([]*SearchResult, error) {
	ret := make([]*SearchResult, 0)

	retManga := make([]*MangadexManga, 0)

	for i := range 1 {
		uri := fmt.Sprintf("%s/manga?title=%s&limit=25&offset=%d&order[relevance]=desc&contentRating[]=safe&contentRating[]=suggestive&includes[]=cover_art", md.Url, url.QueryEscape(opts.Query), 25*i)

		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			md.logger.Error().Err(err).Msg("mangadex: failed to create request")
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
			md.logger.Error().Err(err).Msg("mangadex: failed to decode response")
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

		result := &SearchResult{
			ID:           manga.ID,
			Title:        t,
			Synonyms:     altTitles,
			Image:        img,
			Year:         manga.Attributes.Year,
			SearchRating: compRes.Rating,
			Provider:     MangadexProvider,
		}

		ret = append(ret, result)
	}

	return ret, nil
}
func (md *Mangadex) FindChapters(id string) ([]*ChapterDetails, error) {

	ret := make([]*ChapterDetails, 0)

	for page := 0; ; page++ {
		uri := fmt.Sprintf("%s/manga/%s/feed?limit=500&translatedLanguage%%5B%%5D=en&includes[]=scanlation_group&includes[]=user&order[volume]=desc&order[chapter]=desc&offset=%d&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic", md.Url, id, 500*page)

		fmt.Println(uri)
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
			md.logger.Error().Err(err).Msg("mangadex: failed to decode response")
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if data.Result == "error" {
			md.logger.Error().Str("error", data.Errors[0].Title).Str("detail", data.Errors[0].Detail).Msg("mangadex: could not find chapters")
			return nil, fmt.Errorf("could not find chapters: %s", data.Errors[0].Detail)
		}

		slices.Reverse(data.Data)

		chapters := make([]*ChapterDetails, 0)
		idx := uint(len(ret))
		for _, chapter := range data.Data {
			var title string

			if chapter.Attributes.Volume != "" {
				title += "Vol. " + fmt.Sprintf("%03s", chapter.Attributes.Volume) + " "
			}
			if chapter.Attributes.Chapter != "" {
				title += "Ch. " + fmt.Sprintf("%s", chapter.Attributes.Chapter) + " "
			}

			if title == "" {
				if chapter.Attributes.Title == "" {
					title = "Oneshot"
				} else {
					title = chapter.Attributes.Title
				}
			}

			canPush := true
			for _, ch := range chapters {
				if ch.Title == title {
					canPush = false
					break
				}
			}

			if canPush {
				chapters = append(chapters, &ChapterDetails{
					ID:        chapter.ID,
					Title:     title,
					Index:     idx,
					UpdatedAt: chapter.Attributes.UpdatedAt,
					Provider:  MangadexProvider,
				})
				idx++
			}
		}

		if len(chapters) > 0 {
			ret = append(ret, chapters...)
		} else {
			break
		}
	}

	return ret, nil

}
func (md *Mangadex) FindChapterPages(info *ChapterDetails) ([]*ChapterPage, error) {
	panic("not implemented")
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
