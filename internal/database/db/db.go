package db

import (
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Database struct {
	gormdb           *gorm.DB
	Logger           *zerolog.Logger
	CurrMediaFillers mo.Option[map[int]*MediaFillerItem]
	cleanupManager   *CleanupManager
}

func (db *Database) Gorm() *gorm.DB {
	return db.gormdb
}

func NewDatabase(appDataDir, dbName string, logger *zerolog.Logger) (*Database, error) {

	// Set the SQLite database path
	var sqlitePath string
	if os.Getenv("TEST_ENV") == "true" || appDataDir == "" {
		sqlitePath = ":memory:"
	} else {
		sqlitePath = filepath.Join(appDataDir, dbName+".db")
	}

	// Connect to the SQLite database with optimized settings
	db, err := gorm.Open(sqlite.Open(sqlitePath+"?_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=on"), &gorm.Config{
		Logger: gormlogger.New(
			logger,
			gormlogger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gormlogger.Error,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      false,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Migrate tables
	err = migrateTables(db)
	if err != nil {
		logger.Fatal().Err(err).Msg("db: Failed to perform auto migration")
		return nil, err
	}

	logger.Info().Str("name", fmt.Sprintf("%s.db", dbName)).Msg("db: Database instantiated")

	database := &Database{
		gormdb:           db,
		Logger:           logger,
		CurrMediaFillers: mo.None[map[int]*MediaFillerItem](),
	}

	// Initialize cleanup manager
	database.cleanupManager = NewCleanupManager(database.gormdb, database.Logger)

	return database, nil
}

// MigrateTables performs auto migration on the database
func migrateTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.LocalFiles{},
		&models.ShelvedLocalFiles{},
		&models.Settings{},
		&models.Account{},
		&models.Mal{},
		&models.ScanSummary{},
		&models.AutoSelectProfile{},
		&models.AutoDownloaderRule{},
		&models.AutoDownloaderProfile{},
		&models.AutoDownloaderItem{},
		&models.SilencedMediaEntry{},
		&models.Theme{},
		&models.PlaylistEntry{}, // Legacy playlists
		&models.Playlist{},
		&models.ChapterDownloadQueueItem{},
		&models.TorrentstreamSettings{},
		&models.TorrentstreamHistory{},
		&models.MediastreamSettings{},
		&models.MediaFiller{},
		&models.MangaMapping{},
		&models.OnlinestreamMapping{},
		&models.DebridSettings{},
		&models.DebridTorrentItem{},
		&models.PluginData{},
		&models.CustomSourceCollection{},
		&models.CustomSourceIdentifier{},
		&models.MediaMetadataParent{},
		//&models.MangaChapterContainer{},
	)
	if err != nil {

		return err
	}

	return nil
}

// RunDatabaseCleanup runs all database cleanup operations
func (db *Database) RunDatabaseCleanup() {
	db.cleanupManager.RunAllCleanupOperations()
}
