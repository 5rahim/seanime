package core

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	_db "github.com/seanime-app/seanime-server/internal/db"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/models"
	"github.com/seanime-app/seanime-server/internal/mpchc"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/scanner"
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/seanime-app/seanime-server/internal/vlc"
	"log"
	"os"
	"strings"
)

type App struct {
	Config   *Config
	Database *_db.Database
	Logger   *zerolog.Logger

	MediaPlayer struct {
		VLC   *vlc.VLC
		MpcHc *mpchc.MpcHc
	}
	QBittorrent *qbittorrent.Client

	Watcher *scanner.Watcher

	// AnizipCache holds fetched AniZip media for 10 minutes.
	// It is used by route handlers.
	AnizipCache   *anizip.Cache
	AnilistClient *anilist.Client

	// anilistCollection holds the user's Anilist collection.
	// It is fetched when the server starts.
	anilistCollection *anilist.AnimeCollection
	// Similarly, account holds the user's Anilist account.
	account *models.Account

	// WSEventManager allows the server to send events to the client via websocket and receive events.
	// It is instanciated in the websocket route handler.
	WSEventManager *events.WSEventManager
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

	/*
	 Initialize the config
	 If the config file does not exist, it will be created
	*/
	cfg, err := NewConfig(opts.Config)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
		os.Exit(1)
	}

	logger.Info().Msgf("app: Loaded config from \"%s\"", cfg.Data.AppDataDir)

	/*
	 Initialize the database
	*/
	db, err := _db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
		os.Exit(1)
	}

	logger.Info().Msgf("app: Connected to database \"%s.db\"", cfg.Database.Name)

	app := &App{
		Config:         cfg,
		Database:       db,
		AnilistClient:  anilist.NewAuthedClient(""),
		AnizipCache:    anizip.NewCache(),
		WSEventManager: events.NewWSEventManager(logger),
		Logger:         logger,
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
func NewFiberWebApp() *fiber.App {
	// Create a new fiber app
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})
	// Set up a custom logger for fiber
	//fiberLogger := fiberzerolog.New(fiberzerolog.Config{
	//	Logger:   util.NewLogger(),
	//	SkipURIs: []string{"/internal/metrics"},
	//	Levels:   []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.TraceLevel},
	//})
	//fiberApp.Use(fiberLogger)
	//fiberApp.Use(cors.New(cors.Config{
	//	AllowOrigins: "*",
	//	AllowHeaders: "Origin, Content-Type, Accept",
	//}))

	fiberApp.Static("/", "./web")

	// Serve pages without the ".html" extension
	//fiberApp.Get("/:page", func(c *fiber.Ctx) error {
	//	page := c.Params("page") + ".html"
	//	return c.SendFile("web/" + page)
	//})

	fiberApp.Get("*", func(c *fiber.Ctx) error {
		path := c.OriginalURL()
		if !strings.HasSuffix(path, ".html") {
			path += ".html"
		}
		return c.SendFile("web" + path)
	})

	return fiberApp
}

func RunServer(app *App, fiberApp *fiber.App) {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	// Start the server
	go func() {
		log.Fatal(fiberApp.Listen(addr))
	}()

	app.Logger.Info().Msg("Server started at http://" + addr)

}

func RunWebApp(app *App, fiberWebApp *fiber.App) {
	webAddr := fmt.Sprintf("%s:%d", app.Config.Server.Host, 43211)

	go func() {
		log.Fatal(fiberWebApp.Listen(webAddr))
	}()

	app.Logger.Info().Msg("WebApp started at http://" + webAddr)

}
