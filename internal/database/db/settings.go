package db

import (
	"seanime/internal/database/models"
	"seanime/internal/util"
	"strings"

	"gorm.io/gorm/clause"
)

var CurrSettings *models.Settings

func (db *Database) UpsertSettings(settings *models.Settings) (*models.Settings, error) {
	if settings != nil && settings.Torrent != nil {
		settings.Torrent.QBittorrentHost = strings.TrimSpace(strings.Trim(settings.Torrent.QBittorrentHost, "\""))
		settings.Torrent.TransmissionHost = strings.TrimSpace(strings.Trim(settings.Torrent.TransmissionHost, "\""))
		settings.Torrent.QBittorrentPath = strings.TrimSpace(strings.Trim(settings.Torrent.QBittorrentPath, "\""))
		settings.Torrent.TransmissionPath = strings.TrimSpace(strings.Trim(settings.Torrent.TransmissionPath, "\""))
	}
	dbSettings := CloneSettings(settings)
	VirtualizeSettingsPaths(dbSettings)

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(dbSettings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save settings in the database")
		return nil, err
	}

	ResolveSettingsPathsPhysical(settings)
	CurrSettings = settings

	db.Logger.Debug().Msg("db: Settings saved")
	return settings, nil
}

func (db *Database) GetSettings() (*models.Settings, error) {

	if CurrSettings != nil {
		return CurrSettings, nil
	}

	var settings models.Settings
	err := db.gormdb.Where("id = ?", 1).Find(&settings).Error

	if err != nil {
		return nil, err
	}

	ResolveSettingsPathsPhysical(&settings)

	CurrSettings = &settings
	return &settings, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *Database) GetLibraryPathFromSettings() (string, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return "", err
	}
	return util.ResolvePhysicalPath(settings.Library.LibraryPath), nil
}

func (db *Database) GetAdditionalLibraryPathsFromSettings() ([]string, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return []string{}, err
	}
	resolved := make([]string, len(settings.Library.LibraryPaths))
	for i, p := range settings.Library.LibraryPaths {
		resolved[i] = util.ResolvePhysicalPath(p)
	}
	return resolved, nil
}

func (db *Database) GetAllLibraryPathsFromSettings() ([]string, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return []string{}, err
	}
	if settings.Library == nil {
		return []string{}, nil
	}
	r := append([]string{settings.Library.LibraryPath}, settings.Library.LibraryPaths...)
	resolved := make([]string, len(r))
	for i, p := range r {
		resolved[i] = util.ResolvePhysicalPath(p)
	}
	return resolved, nil
}

func (db *Database) AllLibraryPathsFromSettings(settings *models.Settings) *[]string {
	if settings.Library == nil {
		return &[]string{}
	}
	r := append([]string{settings.Library.LibraryPath}, settings.Library.LibraryPaths...)
	resolved := make([]string, len(r))
	for i, p := range r {
		resolved[i] = util.ResolvePhysicalPath(p)
	}
	return &resolved
}

func (db *Database) AutoUpdateProgressIsEnabled() (bool, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return false, err
	}
	return settings.Library.AutoUpdateProgress, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var CurrMediastreamSettings *models.MediastreamSettings

func (db *Database) UpsertMediastreamSettings(settings *models.MediastreamSettings) (*models.MediastreamSettings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save media streaming settings in the database")
		return nil, err
	}

	CurrMediastreamSettings = settings

	db.Logger.Debug().Msg("db: Media streaming settings saved")
	return settings, nil

}

func (db *Database) GetMediastreamSettings() (*models.MediastreamSettings, bool) {

	if CurrMediastreamSettings != nil {
		return CurrMediastreamSettings, true
	}

	var settings models.MediastreamSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error

	if err != nil {
		return nil, false
	}
	return &settings, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var CurrTorrentstreamSettings *models.TorrentstreamSettings

func (db *Database) UpsertTorrentstreamSettings(settings *models.TorrentstreamSettings) (*models.TorrentstreamSettings, error) {

	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save torrent streaming settings in the database")
		return nil, err
	}

	CurrTorrentstreamSettings = settings

	db.Logger.Debug().Msg("db: Torrent streaming settings saved")
	return settings, nil
}

func (db *Database) GetTorrentstreamSettings() (*models.TorrentstreamSettings, bool) {

	if CurrTorrentstreamSettings != nil {
		return CurrTorrentstreamSettings, true
	}

	var settings models.TorrentstreamSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error

	if err != nil {
		return nil, false
	}
	return &settings, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var CurrentDebridSettings *models.DebridSettings

func (db *Database) UpsertDebridSettings(settings *models.DebridSettings) (*models.DebridSettings, error) {
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save debrid settings in the database")
		return nil, err
	}

	CurrentDebridSettings = settings

	db.Logger.Debug().Msg("db: Debrid settings saved")
	return settings, nil
}

func (db *Database) GetDebridSettings() (*models.DebridSettings, bool) {

	if CurrentDebridSettings != nil {
		return CurrentDebridSettings, true
	}

	var settings models.DebridSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error
	if err != nil {
		return nil, false
	}
	return &settings, true
}

var CurrentDummyDebridSettings *models.DummyDebridSettings

func (db *Database) UpsertDummyDebridSettings(settings *models.DummyDebridSettings) (*models.DummyDebridSettings, error) {
	settings.ID = 1
	err := db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(settings).Error

	if err != nil {
		db.Logger.Error().Err(err).Msg("db: Failed to save dummy debrid settings in the database")
		return nil, err
	}

	CurrentDummyDebridSettings = settings

	db.Logger.Debug().Msg("db: Dummy debrid settings saved")
	return settings, nil
}

func (db *Database) GetDummyDebridSettings() (*models.DummyDebridSettings, bool) {
	if CurrentDummyDebridSettings != nil {
		return CurrentDummyDebridSettings, true
	}

	var settings models.DummyDebridSettings
	err := db.gormdb.Where("id = ?", 1).First(&settings).Error
	if err != nil {
		return nil, false
	}
	CurrentDummyDebridSettings = &settings
	return &settings, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ResolveSettingsPathsPhysical(settings *models.Settings) {
	if settings == nil {
		return
	}
	if settings.Library != nil {
		settings.Library.LibraryPath = util.ResolvePhysicalPath(settings.Library.LibraryPath)
		for i, p := range settings.Library.LibraryPaths {
			settings.Library.LibraryPaths[i] = util.ResolvePhysicalPath(p)
		}
	}
	if settings.Manga != nil {
		settings.Manga.LocalSourceDirectory = util.ResolvePhysicalPath(settings.Manga.LocalSourceDirectory)
	}
}

func VirtualizeSettingsPaths(settings *models.Settings) {
	if !util.IsIOS() || settings == nil {
		return
	}
	if settings.Library != nil {
		settings.Library.LibraryPath = util.ResolveVirtualPath(settings.Library.LibraryPath)
		for i, p := range settings.Library.LibraryPaths {
			settings.Library.LibraryPaths[i] = util.ResolveVirtualPath(p)
		}
	}
	if settings.Manga != nil {
		settings.Manga.LocalSourceDirectory = util.ResolveVirtualPath(settings.Manga.LocalSourceDirectory)
	}
}

func CloneSettings(settings *models.Settings) *models.Settings {
	if settings == nil {
		return nil
	}
	clone := *settings
	if settings.Library != nil {
		lib := *settings.Library
		if settings.Library.LibraryPaths != nil {
			lib.LibraryPaths = append([]string{}, settings.Library.LibraryPaths...)
		} else {
			lib.LibraryPaths = []string{}
		}
		clone.Library = &lib
	}
	if settings.Manga != nil {
		clone.Manga = new(*settings.Manga)
	}
	return &clone
}
