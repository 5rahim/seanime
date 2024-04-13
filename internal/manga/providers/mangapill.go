package manga_providers

import (
	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util"
	"net/http"
	"time"
)

type (
	Mangapill struct {
		Url       string
		Client    *http.Client
		UserAgent string
		logger    *zerolog.Logger
	}
)

func NewMangapill(logger *zerolog.Logger) *Mangapill {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &Mangapill{
		Url:       "https://mangapill.com",
		Client:    c,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		logger:    logger,
	}
}

func (mp *Mangapill) Search(opts SearchOptions) ([]*SearchResult, error) {
	results := make([]*SearchResult, 0)

	mp.logger.Debug().Str("query", opts.Query).Msg("mangapill: Searching manga")

	c := colly.NewCollector(
		colly.UserAgent(mp.UserAgent),
	)

	c.WithTransport(mp.Client.Transport)

	// code

	if len(results) == 0 {
		mp.logger.Error().Str("query", opts.Query).Msg("mangapill: No results found")
		return nil, ErrNoResults
	}

	mp.logger.Info().Int("count", len(results)).Msg("mangapill: Found results")

	return results, nil
}

func (mp *Mangapill) FindChapters(id string) ([]*ChapterDetails, error) {
	ret := make([]*ChapterDetails, 0)

	mp.logger.Debug().Str("mangaId", id).Msg("mangapill: Finding chapters")

	// code

	if len(ret) == 0 {
		mp.logger.Error().Str("mangaId", id).Msg("mangapill: No chapters found")
		return nil, ErrNoChapters
	}

	mp.logger.Info().Int("count", len(ret)).Msg("mangapill: Found chapters")

	return ret, nil
}

func (mp *Mangapill) FindChapterPages(id string) ([]*ChapterPage, error) {
	ret := make([]*ChapterPage, 0)

	mp.logger.Debug().Str("chapterId", id).Msg("mangapill: Finding chapter pages")

	// code

	if len(ret) == 0 {
		mp.logger.Error().Str("chapterId", id).Msg("mangapill: No pages found")
		return nil, ErrNoPages
	}

	mp.logger.Info().Int("count", len(ret)).Msg("mangapill: Found pages")

	return ret, nil

}
