package core

import (
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Database = gorm.DB

func NewDatabase(cfg *Config, logger *zerolog.Logger) (*Database, error) {
	// Get the app data directory from the configuration

	// Set the SQLite database path
	var sqlitePath string
	if os.Getenv("TEST_ENV") == "true" {
		sqlitePath = ":memory:"
	} else {
		sqlitePath = filepath.Join(cfg.Data.AppDataDir, cfg.Database.Name+".db")
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
		logger.Fatal().Err(err).Msg("Failed to connect to the SQLite database")
		return nil, err
	}

	// Migrate tables
	err = migrateTables(db, logger)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// MigrateTables performs auto migration on the database
func migrateTables(db *Database, logger *zerolog.Logger) error {
	err := db.AutoMigrate(
		&models.Token{},
		&models.LocalFiles{},
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to perform auto migration")
		return err
	}

	logger.Info().Msg("Performed auto migration on the database")

	return nil
}
