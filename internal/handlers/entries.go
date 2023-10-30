package handlers

import "github.com/seanime-app/seanime-server/internal/entities"

type mediaEntryQuery struct {
	MediaId int `query:"mediaId" json:"mediaId"`
}

func HandleGetMediaEntry(c *RouteCtx) error {

	p := new(mediaEntryQuery)
	if err := c.Fiber.QueryParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media entry
	entry, err := entities.NewMediaEntry(&entities.NewMediaEntryOptions{
		MediaId:     p.MediaId,
		LocalFiles:  lfs,
		AnizipCache: c.App.AnizipCache,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(entry)
}
