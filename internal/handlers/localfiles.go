package handlers

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/entities"
)

func HandleGetLocalFiles(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleUpdateLocalFileData
// POST
func HandleUpdateLocalFileData(c *RouteCtx) error {

	type body struct {
		Path     string                      `json:"path"`
		Metadata *entities.LocalFileMetadata `json:"metadata"`
		Locked   bool                        `json:"locked"`
		Ignored  bool                        `json:"ignored"`
		MediaId  int                         `json:"mediaId"`
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

	lf, found := lo.Find(lfs, func(i *entities.LocalFile) bool {
		return i.Path == p.Path
	})
	if !found {
		return c.RespondWithError(errors.New("local file not found"))
	}
	lf.Metadata = p.Metadata
	lf.Locked = p.Locked
	lf.Ignored = p.Ignored
	lf.MediaId = p.MediaId

	// Save the local files
	retLfs, err := saveLocalFilesInDB(c.App.Database, dbId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}
