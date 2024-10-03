package core

import (
	"embed"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// NewFiberApp creates a new fiber app instance
// and sets up the static file server for the web interface.
func NewFiberApp(app *App, webFS *embed.FS) *fiber.App {
	// Create a new fiber app
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})

	//
	// Serve the embedded web interface
	//

	distFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	fiberApp.Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(distFS),
		Browse: true,
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			if strings.HasPrefix(path, "/api") ||
				strings.HasPrefix(path, "/events") ||
				strings.HasPrefix(path, "/assets") ||
				strings.HasPrefix(path, "/manga-downloads") ||
				strings.HasPrefix(path, "/offline-assets") {
				return true // Continue to the next handler
			}
			if !strings.HasSuffix(path, ".html") && filepath.Ext(path) == "" {
				if strings.Contains(path, "?") {
					// Split the path into the actual path and the query string
					parts := strings.SplitN(path, "?", 2)
					actualPath := parts[0]
					queryString := parts[1]
					// Add ".html" to the actual path
					actualPath += ".html"
					// Reassemble the path with the query string
					path = actualPath + "?" + queryString
				} else {
					path += ".html"
				}
			}
			if path == "/.html" {
				path = "/index.html"
			}
			c.Path(path)
			return false // Continue to the filesystem handler
		},
	}))

	app.Logger.Info().Msgf("app: Serving embedded web interface")

	// Serve the web assets
	app.Logger.Info().Msgf("app: Web assets path: %s", app.Config.Web.AssetDir)
	fiberApp.Static("/assets", app.Config.Web.AssetDir, fiber.Static{
		Index:    "index.html",
		Compress: false,
	})

	// Serve the manga downloads
	if app.Config.Manga.DownloadDir != "" {
		app.Logger.Info().Msgf("app: Manga downloads path: %s", app.Config.Manga.DownloadDir)
		fiberApp.Static("/manga-downloads", app.Config.Manga.DownloadDir, fiber.Static{
			Index:    "index.html",
			Compress: false,
		})
	}

	// Serve the offline assets
	app.Logger.Info().Msgf("app: Offline assets path: %s", app.Config.Offline.AssetDir)
	fiberApp.Static("/offline-assets", app.Config.Offline.AssetDir, fiber.Static{
		Index:    "index.html",
		Compress: false,
	})

	return fiberApp
}

// RunServer starts the server
func RunServer(app *App, fiberApp *fiber.App) {
	app.Logger.Info().Msgf("app: Server Address: %s", app.Config.GetServerAddr())

	// DEVNOTE: Crashes self-update loop
	//app.Cleanups = append(app.Cleanups, func() {
	//	_ = fiberApp.ShutdownWithTimeout(time.Millisecond)
	//})

	// Start the server
	go func() {
		log.Fatal(fiberApp.Listen(app.Config.GetServerAddr()))
	}()

	app.Logger.Info().Msg("app: Seanime started at " + app.Config.GetServerURI())
}

func (a *App) Cleanup() {
	for _, f := range a.Cleanups {
		f()
	}
}
