package db_bridge

import (
	"errors"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"

	"github.com/goccy/go-json"
	"github.com/samber/mo"
	"gorm.io/gorm"
)

var CurrLocalFilesDbId uint
var CurrLocalFiles mo.Option[[]*anime.LocalFile]

// GetLocalFiles will return the latest local files and the id of the entry.
func GetLocalFiles(db *db.Database) ([]*anime.LocalFile, uint, error) {

	if CurrLocalFiles.IsPresent() {
		return CurrLocalFiles.MustGet(), CurrLocalFilesDbId, nil
	}

	// Get the latest entry
	var res models.LocalFiles
	err := db.Gorm().Last(&res).Error
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal the local files
	lfsBytes := res.Value
	var lfs []*anime.LocalFile
	if err := json.Unmarshal(lfsBytes, &lfs); err != nil {
		return nil, 0, err
	}

	db.Logger.Debug().Msg("db: Local files retrieved")

	CurrLocalFiles = mo.Some(lfs)
	CurrLocalFilesDbId = res.ID

	return lfs, res.ID, nil
}

// SaveLocalFiles will save the local files in the database at the given id.
func SaveLocalFiles(db *db.Database, lfsId uint, lfs []*anime.LocalFile) ([]*anime.LocalFile, error) {
	// Marshal the local files
	marshaledLfs, err := json.Marshal(lfs)
	if err != nil {
		return nil, err
	}

	// Save the local files
	ret, err := db.UpsertLocalFiles(&models.LocalFiles{
		BaseModel: models.BaseModel{
			ID: lfsId,
		},
		Value: marshaledLfs,
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal the saved local files
	var retLfs []*anime.LocalFile
	if err := json.Unmarshal(ret.Value, &retLfs); err != nil {
		return lfs, nil
	}

	CurrLocalFiles = mo.Some(retLfs)
	CurrLocalFilesDbId = ret.ID

	return retLfs, nil
}

// InsertLocalFiles will insert the local files in the database at a new entry.
func InsertLocalFiles(db *db.Database, lfs []*anime.LocalFile) ([]*anime.LocalFile, error) {

	// Marshal the local files
	bytes, err := json.Marshal(lfs)
	if err != nil {
		return nil, err
	}

	// Save the local files to the database
	ret, err := db.InsertLocalFiles(&models.LocalFiles{
		Value: bytes,
	})

	if err != nil {
		return nil, err
	}

	CurrLocalFiles = mo.Some(lfs)
	CurrLocalFilesDbId = ret.ID

	return lfs, nil

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func GetShelvedLocalFiles(db *db.Database) ([]*anime.LocalFile, error) {
	var res models.ShelvedLocalFiles
	err := db.Gorm().Last(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	lfsBytes := res.Value
	var lfs []*anime.LocalFile
	if err := json.Unmarshal(lfsBytes, &lfs); err != nil {
		return nil, err
	}

	db.Logger.Debug().Msg("db: Shelved local files retrieved")

	return lfs, nil
}

func SaveShelvedLocalFiles(db *db.Database, lfs []*anime.LocalFile) error {
	// Marshal the local files
	marshaledLfs, err := json.Marshal(lfs)
	if err != nil {
		return err
	}

	// Save the local files
	ret, err := db.UpsertShelvedLocalFiles(&models.ShelvedLocalFiles{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Value: marshaledLfs,
	})
	if err != nil {
		return err
	}

	// Unmarshal the saved local files
	var retLfs []*anime.LocalFile
	if err := json.Unmarshal(ret.Value, &retLfs); err != nil {
		return nil
	}

	return nil
}
