package handlers

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/db"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/models"
)

func HandleGetLocalFiles(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}

// HandleUpdateLocalFileMetadata
// POST
func HandleUpdateLocalFileMetadata(c *RouteCtx) error {

	type body struct {
		Path     string                      `json:"path"`
		Metadata *entities.LocalFileMetadata `json:"metadata"`
		Locked   bool                        `json:"locked"`
		Ignored  bool                        `json:"ignored"`
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

	// Save the local files
	retLfs, err := saveLocalFilesInDB(c.App.Database, dbId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

func getLocalFilesFromDB(db *db.Database) ([]*entities.LocalFile, error) {
	res, err := db.GetLatestLocalFiles(&models.LocalFiles{})
	if err != nil {
		return nil, err
	}

	lfsBytes := res.Value
	var lfs []*entities.LocalFile
	if err := json.Unmarshal(lfsBytes, &lfs); err != nil {
		return nil, err
	}

	return lfs, nil
}

func saveLocalFilesInDB(db *db.Database, id uint, lfs []*entities.LocalFile) ([]*entities.LocalFile, error) {
	// Marshal the local files
	marshaledLfs, err := json.Marshal(lfs)
	if err != nil {
		return nil, err
	}

	// Save the local files
	ret, err := db.UpsertLocalFiles(&models.LocalFiles{
		BaseModel: models.BaseModel{
			ID: id,
		},
		Value: marshaledLfs,
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal the saved local files
	var retLfs []*entities.LocalFile
	if err := json.Unmarshal(ret.Value, &retLfs); err != nil {
		return lfs, nil
	}

	return retLfs, nil
}

func getLocalFilesAndIdFromDB(db *db.Database) ([]*entities.LocalFile, uint, error) {
	res, err := db.GetLatestLocalFiles(&models.LocalFiles{})
	if err != nil {
		return nil, 0, err
	}

	lfsBytes := res.Value
	var lfs []*entities.LocalFile
	if err := json.Unmarshal(lfsBytes, &lfs); err != nil {
		return nil, 0, err
	}

	return lfs, res.ID, nil
}
