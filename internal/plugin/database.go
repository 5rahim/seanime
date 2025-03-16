package plugin

import (
	"errors"
	"seanime/internal/database/db_bridge"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	util "seanime/internal/util"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Database struct {
	ctx    *AppContextImpl
	logger *zerolog.Logger
	ext    *extension.Extension
}

// BindDatabase binds the database module to the Goja runtime.
// Permissions needed: databases
func (a *AppContextImpl) BindDatabase(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	dbLogger := logger.With().Str("id", ext.ID).Logger()
	db := &Database{
		ctx:    a,
		logger: &dbLogger,
		ext:    ext,
	}
	dbObj := vm.NewObject()

	localFilesObj := vm.NewObject()
	localFilesObj.Set("getAll", db.getAllLocalFiles)
	localFilesObj.Set("findBy", db.findLocalFilesBy)
	localFilesObj.Set("save", db.saveLocalFiles)
	localFilesObj.Set("insert", db.insertLocalFiles)

	anilistObj := vm.NewObject()
	anilistObj.Set("getToken", db.getAnilistToken)
	anilistObj.Set("getUsername", db.getAnilistUsername)

	dbObj.Set("localFiles", localFilesObj)
	dbObj.Set("anilist", anilistObj)

	_ = vm.Set("$database", dbObj)
}

func (d *Database) getAllLocalFiles() ([]*anime.LocalFile, error) {
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

func (d *Database) findLocalFilesBy(filterFn func(*anime.LocalFile) bool) ([]*anime.LocalFile, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	files, _, err := db_bridge.GetLocalFiles(db)
	if err != nil {
		return nil, err
	}

	filteredFiles := make([]*anime.LocalFile, 0)
	for _, file := range files {
		if filterFn(file) {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, nil
}

func (d *Database) saveLocalFiles(filesToSave []*anime.LocalFile) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	lfs, lfsId, err := db_bridge.GetLocalFiles(db)
	if err != nil {
		return err
	}

	filesToSaveMap := make(map[string]*anime.LocalFile)
	for _, file := range filesToSave {
		filesToSaveMap[util.NormalizePath(file.Path)] = file
	}

	for i := range lfs {
		if fileToSave, ok := filesToSaveMap[util.NormalizePath(lfs[i].Path)]; !ok {
			lfs[i] = fileToSave
		}
	}

	_, err = db_bridge.SaveLocalFiles(db, lfsId, lfs)
	if err != nil {
		return err
	}

	ws, ok := d.ctx.wsEventManager.Get()
	if ok {
		ws.SendEvent(events.InvalidateQueries, []string{events.GetLocalFilesEndpoint, events.GetAnimeEntryEndpoint, events.GetLibraryCollectionEndpoint, events.GetMissingEpisodesEndpoint})
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

	ws, ok := d.ctx.wsEventManager.Get()
	if ok {
		ws.SendEvent(events.InvalidateQueries, []string{events.GetLocalFilesEndpoint, events.GetAnimeEntryEndpoint, events.GetLibraryCollectionEndpoint, events.GetMissingEpisodesEndpoint})
	}

	return lfs, nil
}

func (d *Database) getAnilistToken() (string, error) {
	if d.ext.Plugin == nil || len(d.ext.Plugin.Permissions) == 0 {
		return "", errors.New("permission denied")
	}
	if !util.Contains(d.ext.Plugin.Permissions, extension.PluginPermissionAnilistToken) {
		return "", errors.New("permission denied")
	}
	db, ok := d.ctx.database.Get()
	if !ok {
		return "", errors.New("database not initialized")
	}
	return db.GetAnilistToken(), nil
}

func (d *Database) getAnilistUsername() (string, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return "", errors.New("database not initialized")
	}

	acc, err := db.GetAccount()
	if err != nil {
		return "", nil
	}

	return acc.Username, nil
}
