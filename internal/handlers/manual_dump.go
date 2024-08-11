package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/library/scanner"
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

	completeAnimeCache := anilist.NewCompleteAnimeCache()
	anizipCache := anizip.NewCache()

	mc, err := scanner.NewMediaFetcher(&scanner.MediaFetcherOptions{
		Enhanced:           false,
		Platform:           c.App.AnilistPlatform,
		LocalFiles:         localFiles,
		CompleteAnimeCache: completeAnimeCache,
		AnizipCache:        anizipCache,
		Logger:             c.App.Logger,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mc.AllMedia)

}
