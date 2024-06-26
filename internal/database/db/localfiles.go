package db

import (
	"github.com/goccy/go-json"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
	"gorm.io/gorm/clause"
)

// GetLocalFiles will return the latest local files and the id of the entry.
func (db *Database) GetLocalFiles() ([]*anime.LocalFile, uint, error) {

	if db.currLocalFiles.IsPresent() {
		return db.currLocalFiles.MustGet(), db.currLocalFilesDbId, nil
	}

	// Get the latest entry
	var res models.LocalFiles
	err := db.gormdb.Last(&res).Error
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal the local files
	lfsBytes := res.Value
	var lfs []*anime.LocalFile
	if err := json.Unmarshal(lfsBytes, &lfs); err != nil {
		return nil, 0, err
	}

	db.logger.Debug().Msg("db: Local files retrieved")

	db.currLocalFiles = mo.Some(lfs)
	db.currLocalFilesDbId = res.ID

	return lfs, res.ID, nil
}

// SaveLocalFiles will save the local files in the database at the given id.
func (db *Database) SaveLocalFiles(lfsId uint, lfs []*anime.LocalFile) ([]*anime.LocalFile, error) {
	// Marshal the local files
	marshaledLfs, err := json.Marshal(lfs)
	if err != nil {
		return nil, err
	}

	// Save the local files
	ret, err := db.upsertLocalFiles(&models.LocalFiles{
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

	db.currLocalFiles = mo.Some(retLfs)
	db.currLocalFilesDbId = ret.ID

	return retLfs, nil
}

// InsertLocalFiles will insert the local files in the database at a new entry.
func (db *Database) InsertLocalFiles(lfs []*anime.LocalFile) ([]*anime.LocalFile, error) {

	// Marshal the local files
	bytes, err := json.Marshal(lfs)
	if err != nil {
		return nil, err
	}

	// Save the local files to the database
	ret, err := db.insertLocalFiles(&models.LocalFiles{
		Value: bytes,
	})

	if err != nil {
		return nil, err
	}

	db.currLocalFiles = mo.Some(lfs)
	db.currLocalFilesDbId = ret.ID

	return lfs, nil

}

// TrimLocalFileEntries will trim the local file entries if there are more than 10 entries.
// This is run in a goroutine.
func (db *Database) TrimLocalFileEntries() {
	go func() {
		var count int64
		err := db.gormdb.Model(&models.LocalFiles{}).Count(&count).Error
		if err != nil {
			db.logger.Error().Err(err).Msg("Failed to count local file entries")
			return
		}
		if count > 10 {
			// Leave 5 entries
			err = db.gormdb.Delete(&models.LocalFiles{}, "id IN (SELECT id FROM local_files ORDER BY id ASC LIMIT ?)", count-5).Error
			if err != nil {
				db.logger.Error().Err(err).Msg("Failed to delete old local file entries")
				return
			}
		}
	}()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) upsertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(lfs).Error

	if err != nil {
		return nil, err
	}
	return lfs, nil
}

func (db *Database) insertLocalFiles(lfs *models.LocalFiles) (*models.LocalFiles, error) {
	err := db.gormdb.Create(lfs).Error

	if err != nil {
		return nil, err
	}
	return lfs, nil
}
