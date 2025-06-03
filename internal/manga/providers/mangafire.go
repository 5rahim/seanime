package manga_providers

import (
	"fmt"
	"net/url"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

// DEVNOTE: Shelved due to WAF captcha

type (
	Mangafire struct {
		Url       string
		Client    *req.Client
		UserAgent string
		logger    *zerolog.Logger
	}
)

func NewMangafire(logger *zerolog.Logger) *Mangafire {
	client := req.C().
		SetUserAgent(util.GetRandomUserAgent()).
		SetTimeout(60 * time.Second).
		EnableInsecureSkipVerify().
		ImpersonateChrome()

	return &Mangafire{
		Url:       "https://mangafire.to",
		Client:    client,
		UserAgent: util.GetRandomUserAgent(),
		logger:    logger,
	}
}

func (mf *Mangafire) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (mf *Mangafire) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	results := make([]*hibikemanga.SearchResult, 0)

	mf.logger.Debug().Str("query", opts.Query).Msg("mangafire: Searching manga")

	yearStr := ""
	if opts.Year > 0 {
		yearStr = fmt.Sprintf("&year=%%5B%%5D=%d", opts.Year)
	}
	uri := fmt.Sprintf("%s/filter?keyword=%s%s&sort=recently_updated", mf.Url, url.QueryEscape(opts.Query), yearStr)

	c := colly.NewCollector(
		colly.UserAgent(util.GetRandomUserAgent()),
	)

	c.WithTransport(mf.Client.Transport)

	type ToVisit struct {
		ID    string
		Title string
		Image string
	}
	toVisit := make([]ToVisit, 0)

	c.OnHTML("main div.container div.original div.unit", func(e *colly.HTMLElement) {
		id := e.ChildAttr("a", "href")
		if len(toVisit) >= 15 || id == "" {
			return
		}
		title := ""
		e.ForEachWithBreak("div.info a", func(i int, e *colly.HTMLElement) bool {
			if i == 0 && e.Text != "" {
				title = strings.TrimSpace(e.Text)
				return false
			}
			return true
		})
		obj := ToVisit{
			ID:    id,
			Title: title,
			Image: e.ChildAttr("img", "src"),
		}
		if obj.Title != "" && obj.ID != "" {
			toVisit = append(toVisit, obj)
		}
	})

	err := c.Visit(uri)
	if err != nil {
		mf.logger.Error().Err(err).Msg("mangafire: Failed to visit")
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(toVisit))

	for _, v := range toVisit {
		go func(tv ToVisit) {
			defer wg.Done()

			c2 := colly.NewCollector(
				colly.UserAgent(mf.UserAgent),
			)

			c2.WithTransport(mf.Client.Transport)

			result := &hibikemanga.SearchResult{
				Provider: MangafireProvider,
			}

			// Synonyms
			c2.OnHTML("main div#manga-page div.info h6", func(e *colly.HTMLElement) {
				parts := strings.Split(e.Text, "; ")
				for i, v := range parts {
					parts[i] = strings.TrimSpace(v)
				}
				syn := strings.Join(parts, "")
				if syn != "" {
					result.Synonyms = append(result.Synonyms, syn)
				}
			})

			// Year
			c2.OnHTML("main div#manga-page div.meta", func(e *colly.HTMLElement) {
				if result.Year != 0 || e.Text == "" {
					return
				}
				parts := strings.Split(e.Text, "Published: ")
				if len(parts) < 2 {
					return
				}
				parts2 := strings.Split(parts[1], " to")
				if len(parts2) < 2 {
					return
				}
				result.Year = util.StringToIntMust(strings.TrimSpace(parts2[0]))
			})

			result.ID = tv.ID
			result.Title = tv.Title
			result.Image = tv.Image

			err := c2.Visit(fmt.Sprintf("%s/%s", mf.Url, tv.ID))
			if err != nil {
				mf.logger.Error().Err(err).Str("id", tv.ID).Msg("mangafire: Failed to visit manga page")
				return
			}

			// Comparison
			compTitles := []*string{&result.Title}
			for _, syn := range result.Synonyms {
				if !util.IsMostlyLatinString(syn) {
					continue
				}
				compTitles = append(compTitles, &syn)
			}
			compRes, _ := comparison.FindBestMatchWithSorensenDice(&opts.Query, compTitles)

			result.SearchRating = compRes.Rating

			results = append(results, result)
		}(v)
	}

	wg.Wait()

	if len(results) == 0 {
		mf.logger.Error().Str("query", opts.Query).Msg("mangafire: No results found")
		return nil, ErrNoResults
	}

	mf.logger.Info().Int("count", len(results)).Msg("mangafire: Found results")

	return results, nil
}

func (mf *Mangafire) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	ret := make([]*hibikemanga.ChapterDetails, 0)

	mf.logger.Debug().Str("mangaId", id).Msg("mangafire: Finding chapters")

	// code

	if len(ret) == 0 {
		mf.logger.Error().Str("mangaId", id).Msg("mangafire: No chapters found")
		return nil, ErrNoChapters
	}

	mf.logger.Info().Int("count", len(ret)).Msg("mangafire: Found chapters")

	return ret, nil
}

func (mf *Mangafire) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {
	ret := make([]*hibikemanga.ChapterPage, 0)

	mf.logger.Debug().Str("chapterId", id).Msg("mangafire: Finding chapter pages")

	// code

	if len(ret) == 0 {
		mf.logger.Error().Str("chapterId", id).Msg("mangafire: No pages found")
		return nil, ErrNoPages
	}

	mf.logger.Info().Int("count", len(ret)).Msg("mangafire: Found pages")

	return ret, nil

}
