package plugin

import (
	"errors"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	util "seanime/internal/util"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
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

	// Local files
	localFilesObj := vm.NewObject()
	localFilesObj.Set("getAll", db.getAllLocalFiles)
	localFilesObj.Set("findBy", db.findLocalFilesBy)
	localFilesObj.Set("save", db.saveLocalFiles)
	localFilesObj.Set("insert", db.insertLocalFiles)
	dbObj.Set("localFiles", localFilesObj)

	// Anilist
	anilistObj := vm.NewObject()
	anilistObj.Set("getToken", db.getAnilistToken)
	anilistObj.Set("getUsername", db.getAnilistUsername)
	dbObj.Set("anilist", anilistObj)

	// Auto downloader rules
	autoDownloaderRulesObj := vm.NewObject()
	autoDownloaderRulesObj.Set("getAll", db.getAllAutoDownloaderRules)
	autoDownloaderRulesObj.Set("get", db.getAutoDownloaderRule)
	autoDownloaderRulesObj.Set("getByMediaId", db.getAutoDownloaderRulesByMediaId)
	autoDownloaderRulesObj.Set("update", db.updateAutoDownloaderRule)
	autoDownloaderRulesObj.Set("insert", db.insertAutoDownloaderRule)
	autoDownloaderRulesObj.Set("remove", db.deleteAutoDownloaderRule)
	dbObj.Set("autoDownloaderRules", autoDownloaderRulesObj)

	// Auto downloader items
	autoDownloaderItemsObj := vm.NewObject()
	autoDownloaderItemsObj.Set("getAll", db.getAllAutoDownloaderItems)
	autoDownloaderItemsObj.Set("get", db.getAutoDownloaderItem)
	autoDownloaderItemsObj.Set("getByMediaId", db.getAutoDownloaderItemsByMediaId)
	autoDownloaderItemsObj.Set("insert", db.insertAutoDownloaderItem)
	autoDownloaderItemsObj.Set("remove", db.deleteAutoDownloaderItem)
	dbObj.Set("autoDownloaderItems", autoDownloaderItemsObj)

	// Silenced media entries
	silencedMediaEntriesObj := vm.NewObject()
	silencedMediaEntriesObj.Set("getAllIds", db.getAllSilencedMediaEntryIds)
	silencedMediaEntriesObj.Set("isSilenced", db.isSilenced)
	silencedMediaEntriesObj.Set("setSilenced", db.setSilenced)
	dbObj.Set("silencedMediaEntries", silencedMediaEntriesObj)

	// Media fillers
	mediaFillersObj := vm.NewObject()
	mediaFillersObj.Set("getAll", db.getAllMediaFillers)
	mediaFillersObj.Set("get", db.getMediaFiller)
	mediaFillersObj.Set("insert", db.insertMediaFiller)
	mediaFillersObj.Set("remove", db.deleteMediaFiller)
	dbObj.Set("mediaFillers", mediaFillersObj)

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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d *Database) getAnilistToken() (string, error) {
	if d.ext.Plugin == nil || len(d.ext.Plugin.Permissions.Scopes) == 0 {
		return "", errors.New("permission denied")
	}
	if !util.Contains(d.ext.Plugin.Permissions.Scopes, extension.PluginPermissionAnilistToken) {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d *Database) getAllAutoDownloaderRules() ([]*anime.AutoDownloaderRule, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	rules, err := db_bridge.GetAutoDownloaderRules(db)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

func (d *Database) getAutoDownloaderRule(id uint) (*anime.AutoDownloaderRule, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	rule, err := db_bridge.GetAutoDownloaderRule(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return rule, nil
}

func (d *Database) getAutoDownloaderRulesByMediaId(mediaId int) ([]*anime.AutoDownloaderRule, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	rules := db_bridge.GetAutoDownloaderRulesByMediaId(db, mediaId)

	return rules, nil
}

func (d *Database) updateAutoDownloaderRule(id uint, rule *anime.AutoDownloaderRule) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db_bridge.UpdateAutoDownloaderRule(db, id, rule)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) insertAutoDownloaderRule(rule *anime.AutoDownloaderRule) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db_bridge.InsertAutoDownloaderRule(db, rule)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) deleteAutoDownloaderRule(id uint) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db_bridge.DeleteAutoDownloaderRule(db, id)
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d *Database) getAllAutoDownloaderItems() ([]*models.AutoDownloaderItem, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	items, err := db.GetAutoDownloaderItems()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Database) getAutoDownloaderItem(id uint) (*models.AutoDownloaderItem, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	item, err := db.GetAutoDownloaderItem(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (d *Database) getAutoDownloaderItemsByMediaId(mediaId int) ([]*models.AutoDownloaderItem, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	items, err := db.GetAutoDownloaderItemByMediaId(mediaId)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Database) insertAutoDownloaderItem(item *models.AutoDownloaderItem) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db.InsertAutoDownloaderItem(item)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) deleteAutoDownloaderItem(id uint) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db.DeleteAutoDownloaderItem(id)
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d *Database) getAllSilencedMediaEntryIds() ([]int, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	ids, err := db.GetSilencedMediaEntryIds()
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (d *Database) isSilenced(mediaId int) (bool, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return false, errors.New("database not initialized")
	}

	entry, err := db.GetSilencedMediaEntry(uint(mediaId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return entry != nil, nil
}

func (d *Database) setSilenced(mediaId int, silenced bool) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	if silenced {
		err := db.InsertSilencedMediaEntry(uint(mediaId))
		if err != nil {
			return nil
		}
	} else {
		err := db.DeleteSilencedMediaEntry(uint(mediaId))
		if err != nil {
			return nil
		}
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d *Database) getAllMediaFillers() (map[int]*db.MediaFillerItem, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	fillers, err := db.GetCachedMediaFillers()
	if err != nil {
		return nil, err
	}

	return fillers, nil
}

func (d *Database) getMediaFiller(mediaId int) (*db.MediaFillerItem, error) {
	db, ok := d.ctx.database.Get()
	if !ok {
		return nil, errors.New("database not initialized")
	}

	filler, ok := db.GetMediaFillerItem(mediaId)
	if !ok {
		return nil, nil
	}

	return filler, nil
}

func (d *Database) insertMediaFiller(provider string, mediaId int, slug string, fillerEpisodes []string) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	err := db.InsertMediaFiller(provider, mediaId, slug, time.Now(), fillerEpisodes)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) deleteMediaFiller(mediaId int) error {
	db, ok := d.ctx.database.Get()
	if !ok {
		return errors.New("database not initialized")
	}

	return db.DeleteMediaFiller(mediaId)
}
