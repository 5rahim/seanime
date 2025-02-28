package plugin

import (
	"errors"
	"seanime/internal/database/db_bridge"
	"seanime/internal/extension"
	"seanime/internal/library/anime"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Database struct {
	ctx    *AppContextImpl
	logger *zerolog.Logger
}

// BindDatabase binds the database module to the Goja runtime.
// Permissions needed: databases
func (a *AppContextImpl) BindDatabase(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	dbLogger := logger.With().Str("id", ext.ID).Logger()
	db := &Database{
		ctx:    a,
		logger: &dbLogger,
	}
	dbObj := vm.NewObject()

	dbObj.Set("getLocalFiles", db.getLocalFiles)
	dbObj.Set("saveLocalFiles", db.saveLocalFiles)
	dbObj.Set("insertLocalFiles", db.insertLocalFiles)

	_ = vm.Set("$database", dbObj)
}

func (d *Database) getLocalFiles() ([]*anime.LocalFile, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	files, _, err := db_bridge.GetLocalFiles(db)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (d *Database) saveLocalFiles(files []*anime.LocalFile) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	_, lfsId, err := db_bridge.GetLocalFiles(db)
	if err != nil {
		return err
	}

	_, err = db_bridge.SaveLocalFiles(db, lfsId, files)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) insertLocalFiles(files []*anime.LocalFile) ([]*anime.LocalFile, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	lfs, err := db_bridge.InsertLocalFiles(db, files)
	if err != nil {
		return nil, err
	}

	return lfs, nil
}
