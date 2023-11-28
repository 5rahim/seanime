package handlers

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/scanner"
)

type RequestBody struct {
	Dir      string `json:"dir"`
	Username string `json:"userName"`
}

// HandleManualDump is a test
func HandleManualDump(c *RouteCtx) error {

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
		Enhanced:       false,
		Username:       body.Username,
		AnilistClient:  c.App.AnilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
		Logger:         c.App.Logger,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(mc.AllMedia)

}
