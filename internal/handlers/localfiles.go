package handlers

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime-server/internal/db"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/models"
)

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

func HandleGetLocalFiles(c *RouteCtx) error {

	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(lfs)

}
