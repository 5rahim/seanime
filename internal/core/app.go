package core

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/util"
	"log"
	"os"
)

type App struct {
	Config        *Config
	Database      *Database
	AnilistClient *anilist.Client
	Logger        *zerolog.Logger
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

	cfg, err := NewConfig(opts.Config)
	if err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	logger.Info().Msg("Loaded config from " + cfg.Data.AppDataDir)

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	logger.Info().Msg("Connected to database " + cfg.Database.Name)

	anilistClient := anilist.NewAuthedClient("")

	return &App{
		Config:        cfg,
		Database:      db,
		AnilistClient: anilistClient,
		Logger:        logger,
	}
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

///////////////////////

func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClient = anilist.NewAuthedClient(token)
}
