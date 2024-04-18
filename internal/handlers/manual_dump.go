package handlers

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/library/scanner"
)

// DUMMY HANDLER

type RequestBody struct {
	Dir      string `json:"dir"`
	Username string `json:"userName"`
}

// HandleTestDump
//
//	@summary this is a dummy handler for testing purposes.
//	@route /api/v1/test-dump [POST]
func HandleTestDump(c *RouteCtx) error {

	c.AcceptJSON()

	body := new(RequestBody)
	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	localFiles, err := scanner.GetLocalFilesFromDir(body.Dir, c.App.Logger)
	if err != nil {
		return c.RespondWithError(err)
	}

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()

	mc, err := scanner.NewMediaFetcher(&scanner.MediaFetcherOptions{
		Enhanced:             false,
		Username:             body.Username,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		LocalFiles:           localFiles,
		BaseMediaCache:       baseMediaCache,
		AnizipCache:          anizipCache,
		Logger:               c.App.Logger,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mc.AllMedia)

}
