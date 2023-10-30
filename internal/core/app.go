package core

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	_db "github.com/seanime-app/seanime-server/internal/db"
	"github.com/seanime-app/seanime-server/internal/models"
	"github.com/seanime-app/seanime-server/internal/mpchc"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/scanner"
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/seanime-app/seanime-server/internal/vlc"
	"log"
	"os"
)

type App struct {
	Config        *Config
	Database      *_db.Database
	AnilistClient *anilist.Client
	Logger        *zerolog.Logger
	MediaPlayer   struct {
		VLC   *vlc.VLC
		MpcHc *mpchc.MpcHc
	}
	QBittorrent       *qbittorrent.Client
	Watcher           *scanner.Watcher
	anilistCollection *anilist.AnimeCollection
	account           *models.Account
}

type ServerOptions struct {
	Config *ConfigOptions
}

// NewApp creates a new server instance
func NewApp(options *ServerOptions) *App {

	opts := *options

	// Set up a default config if none is provided
	if options.Config == nil {
		opts.Config = &ConfigOptions{
			DataDirPath: "",
		}
	}

	logger := util.NewLogger()

	// Load the config
	cfg, err := NewConfig(opts.Config)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
		os.Exit(1)
	}

	logger.Info().Msgf("app: Loaded config from \"%s\"", cfg.Data.AppDataDir)

	// Initialize the database
	db, err := _db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
		os.Exit(1)
	}

	logger.Info().Msgf("app: Connected to database \"%s.db\"", cfg.Database.Name)

	// Initialize Anilist client
	anilistClient := anilist.NewAuthedClient("")

	app := &App{
		Config:        cfg,
		Database:      db,
		AnilistClient: anilistClient,
		Logger:        logger,
	}

	app.InitOrRefreshDependencies()

	return app
}

func NewFiberApp(app *App) *fiber.App {
	// Create a new fiber app
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})

	// Set up a custom logger for fiber
	fiberLogger := fiberzerolog.New(fiberzerolog.Config{
		Logger:   app.Logger,
		SkipURIs: []string{"/internal/metrics"},
		Levels:   []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.TraceLevel},
	})
	fiberApp.Use(fiberLogger)

	return fiberApp
}

func RunServer(app *App, fiberApp *fiber.App) {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	// Start the server
	go func() {
		log.Fatal(fiberApp.Listen(addr))
	}()

	app.Logger.Info().Msg("Server started at http://" + addr)

	select {}
}
