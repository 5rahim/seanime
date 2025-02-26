package db

import (
	"fmt"
	"log"
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
}

func (db *Database) Gorm() *gorm.DB {
	return db.gormdb
}

func NewDatabase(appDataDir, dbName string, logger *zerolog.Logger) (*Database, error) {

	// Set the SQLite database path
	var sqlitePath string
	if os.Getenv("TEST_ENV") == "true" {
		sqlitePath = ":memory:"
	} else {
		sqlitePath = filepath.Join(appDataDir, dbName+".db")
	}

	// Connect to the SQLite database
	db, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
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

	// Migrate tables
	err = migrateTables(db)
	if err != nil {
		logger.Fatal().Err(err).Msg("db: Failed to perform auto migration")
		return nil, err
	}

	logger.Info().Str("name", fmt.Sprintf("%s.db", dbName)).Msg("db: Database instantiated")

	return &Database{
		gormdb:           db,
		Logger:           logger,
		CurrMediaFillers: mo.None[map[int]*MediaFillerItem](),
	}, nil
}

// MigrateTables performs auto migration on the database
func migrateTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.LocalFiles{},
		&models.Settings{},
		&models.Account{},
		&models.Mal{},
		&models.ScanSummary{},
		&models.AutoDownloaderRule{},
		&models.AutoDownloaderItem{},
		&models.SilencedMediaEntry{},
		&models.Theme{},
		&models.PlaylistEntry{},
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
		//&models.MangaChapterContainer{},
	)
	if err != nil {

		return err
	}

	return nil
}
