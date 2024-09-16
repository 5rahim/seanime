package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/library/scanner"
	"seanime/internal/util/limiter"
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

	mc, err := scanner.NewMediaFetcher(&scanner.MediaFetcherOptions{
		Enhanced:               false,
		Platform:               c.App.AnilistPlatform,
		MetadataProvider:       c.App.MetadataProvider,
		LocalFiles:             localFiles,
		CompleteAnimeCache:     completeAnimeCache,
		Logger:                 c.App.Logger,
		AnilistRateLimiter:     limiter.NewAnilistLimiter(),
		DisableAnimeCollection: false,
		ScanLogger:             nil,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mc.AllMedia)

}
