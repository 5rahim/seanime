package offline

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"time"
)

type database struct {
	gormdb *gorm.DB
	logger *zerolog.Logger
}

func newDatabase(appDataDir, dbName string, logger *zerolog.Logger, isOffline bool) (*database, error) {

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
		if isOffline {
			logger.Fatal().Err(err).Msg("offline hub: Failed to perform auto migration")
		}
		return nil, err
	}

	if isOffline {
		logger.Info().Str("name", fmt.Sprintf("%s.db", dbName)).Msg("offline hub: Database instantiated")
	}

	return &database{
		gormdb: db,
		logger: logger,
	}, nil
}

// MigrateTables performs auto migration on the database
func migrateTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&SnapshotEntry{},
		&SnapshotMediaEntry{},
	)
	if err != nil {

		return err
	}

	return nil
}
