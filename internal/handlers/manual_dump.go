package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/library/scanner"
	"seanime/internal/util/limiter"

	"github.com/labstack/echo/v4"
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
func (h *Handler) HandleTestDump(c echo.Context) error {

	body := new(RequestBody)
	if err := c.Bind(body); err != nil {
		return h.RespondWithError(c, err)
	}

	localFiles, err := scanner.GetLocalFilesFromDir(body.Dir, h.App.Logger)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	completeAnimeCache := anilist.NewCompleteAnimeCache()

	mc, err := scanner.NewMediaFetcher(c.Request().Context(), &scanner.MediaFetcherOptions{
		Enhanced:               false,
		Platform:               h.App.AnilistPlatform,
		MetadataProvider:       h.App.MetadataProvider,
		LocalFiles:             localFiles,
		CompleteAnimeCache:     completeAnimeCache,
		Logger:                 h.App.Logger,
		AnilistRateLimiter:     limiter.NewAnilistLimiter(),
		DisableAnimeCollection: false,
		ScanLogger:             nil,
	})

	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, mc.AllMedia)
}
