package handlers

import (
	"errors"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/constants"
	"github.com/seanime-app/seanime-server/internal/entities"
)

func HandleGetMediaEntry(c *RouteCtx) error {

	type query struct {
		MediaId int `query:"mediaId" json:"mediaId"`
	}

	p := new(query)
	if err := c.Fiber.QueryParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media entry
	entry, err := entities.NewMediaEntry(&entities.NewMediaEntryOptions{
		MediaId:           p.MediaId,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
		AnilistCollection: anilistCollection,
		AnilistClient:     c.App.AnilistClient,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	// Fetch media details in the background and send them via websocket
	go func() {
		details, err := c.App.AnilistClient.MediaDetailsByID(c.Fiber.Context(), &p.MediaId)
		if err == nil {
			c.App.WSEventManager.SendEvent(constants.EventMediaDetails, details)
		}
	}()

	return c.RespondWithData(entry)
}

func HandleToggleEntryLockedStatus(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, dbId, err := getLocalFilesAndIdFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by media id
	groupedLfs := lop.GroupBy(lfs, func(item *entities.LocalFile) int {
		return item.MediaId
	})

	selectLfs, ok := groupedLfs[p.MediaId]
	if !ok {
		return c.RespondWithError(errors.New("no local files found for media id"))
	}

	// Flip the locked status of all the local files for the given media
	allLocked := lo.EveryBy(selectLfs, func(item *entities.LocalFile) bool {
		return item.Locked
	})

	lfs = lop.Map(lfs, func(item *entities.LocalFile, _ int) *entities.LocalFile {
		if item.MediaId == p.MediaId && p.MediaId != 0 {
			item.Locked = !allLocked
		}
		return item
	})

	// Save the local files
	retLfs, err := saveLocalFilesInDB(c.App.Database, dbId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}
