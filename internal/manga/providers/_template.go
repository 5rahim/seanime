package manga_providers

import (
	"github.com/rs/zerolog"
	"net/http"
	"seanime/internal/util"
	"time"
)

type (
	Template struct {
		Url       string
		Client    *http.Client
		UserAgent string
		logger    *zerolog.Logger
	}
)

func NewTemplate(logger *zerolog.Logger) *Template {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.Transport = util.AddCloudFlareByPass(c.Transport)
	return &Template{
		Url:       "https://XXXXXX.com",
		Client:    c,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		logger:    logger,
	}
}

func (mp *Template) Search(opts SearchOptions) ([]*SearchResult, error) {
	results := make([]*SearchResult, 0)

	mp.logger.Debug().Str("query", opts.Query).Msg("XXXXXX: Searching manga")

	// code

	if len(results) == 0 {
		mp.logger.Error().Str("query", opts.Query).Msg("XXXXXX: No results found")
		return nil, ErrNoResults
	}

	mp.logger.Info().Int("count", len(results)).Msg("XXXXXX: Found results")

	return results, nil
}

func (mp *Template) FindChapters(id string) ([]*ChapterDetails, error) {
	ret := make([]*ChapterDetails, 0)

	mp.logger.Debug().Str("mangaId", id).Msg("XXXXXX: Finding chapters")

	// code

	if len(ret) == 0 {
		mp.logger.Error().Str("mangaId", id).Msg("XXXXXX: No chapters found")
		return nil, ErrNoChapters
	}

	mp.logger.Info().Int("count", len(ret)).Msg("XXXXXX: Found chapters")

	return ret, nil
}

func (mp *Template) FindChapterPages(id string) ([]*ChapterPage, error) {
	ret := make([]*ChapterPage, 0)

	mp.logger.Debug().Str("chapterId", id).Msg("XXXXXX: Finding chapter pages")

	// code

	if len(ret) == 0 {
		mp.logger.Error().Str("chapterId", id).Msg("XXXXXX: No pages found")
		return nil, ErrNoPages
	}

	mp.logger.Info().Int("count", len(ret)).Msg("XXXXXX: Found pages")

	return ret, nil

}
