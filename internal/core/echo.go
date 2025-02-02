package core

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4/middleware"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

func NewEchoApp(app *App, webFS *embed.FS) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	e.JSONSerializer = &CustomJSONSerializer{}

	distFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(distFS),
		Browse:     true,
		HTML5:      true,
		Skipper: func(c echo.Context) bool {
			cUrl := c.Request().URL
			if strings.HasPrefix(cUrl.RequestURI(), "/api") ||
				strings.HasPrefix(cUrl.RequestURI(), "/events") ||
				strings.HasPrefix(cUrl.RequestURI(), "/assets") ||
				strings.HasPrefix(cUrl.RequestURI(), "/manga-downloads") ||
				strings.HasPrefix(cUrl.RequestURI(), "/offline-assets") {
				return true // Continue to the next handler
			}
			if !strings.HasSuffix(cUrl.Path, ".html") && filepath.Ext(cUrl.Path) == "" {
				cUrl.Path = cUrl.Path + ".html"
			}
			if cUrl.Path == "/.html" {
				cUrl.Path = "/index.html"
			}
			return false // Continue to the filesystem handler
		},
	}))

	app.Logger.Info().Msgf("app: Serving embedded web interface")

	// Serve web assets
	app.Logger.Info().Msgf("app: Web assets path: %s", app.Config.Web.AssetDir)
	e.Static("/assets", app.Config.Web.AssetDir)

	// Serve manga downloads
	if app.Config.Manga.DownloadDir != "" {
		app.Logger.Info().Msgf("app: Manga downloads path: %s", app.Config.Manga.DownloadDir)
		e.Static("/manga-downloads", app.Config.Manga.DownloadDir)
	}

	// Serve offline assets
	app.Logger.Info().Msgf("app: Offline assets path: %s", app.Config.Offline.AssetDir)
	e.Static("/offline-assets", app.Config.Offline.AssetDir)

	return e
}

type CustomJSONSerializer struct{}

func (j *CustomJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	return enc.Encode(i)
}

func (j *CustomJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	dec := json.NewDecoder(c.Request().Body)
	return dec.Decode(i)
}

func RunEchoServer(app *App, e *echo.Echo) {
	app.Logger.Info().Msgf("app: Server Address: %s", app.Config.GetServerAddr())

	// Start the server
	go func() {
		log.Fatal(e.Start(app.Config.GetServerAddr()))
	}()

	time.Sleep(100 * time.Millisecond)
	app.Logger.Info().Msg("app: Seanime started at " + app.Config.GetServerURI())
}

//fileServer := http.FileServer(http.FS(distFS))

//e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
//	return func(c echo.Context) error {
//		path := c.Request().URL.RequestURI()
//
//		if path == "/" {
//			path = "/index.html"
//		}
//
//		// Skip certain paths
//		if strings.HasPrefix(path, "/api") ||
//			strings.HasPrefix(path, "/events") ||
//			strings.HasPrefix(path, "/assets") ||
//			strings.HasPrefix(path, "/manga-downloads") ||
//			strings.HasPrefix(path, "/offline-assets") {
//			return next(c)
//		}
//
//		// Handle HTML extension and index.html
//		if !strings.HasSuffix(path, ".html") && filepath.Ext(path) == "" {
//			if strings.Contains(path, "?") {
//				// Split the path into the actual path and the query string
//				parts := strings.SplitN(path, "?", 2)
//				actualPath := parts[0]
//				queryString := parts[1]
//				// Add ".html" to the actual path
//				actualPath += ".html"
//				// Reassemble the path with the query string
//				c.Request().URL.Path = actualPath + "?" + queryString
//			} else {
//				c.Request().URL.Path = path + ".html"
//			}
//		}
//
//		if c.Request().URL.Path == "/.html" {
//			c.Request().URL.Path = "/index.html"
//		}
//
//		fileServer.ServeHTTP(c.Response(), c.Request())
//		return nil
//	}
//})

// Handle static files from embedded filesystem
//e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	path := r.URL.Path
//
//	if path == "/" {
//		path = "/index.html"
//	}
//
//	// Handle HTML extension and index.html
//	if !strings.HasSuffix(path, ".html") && filepath.Ext(path) == "" {
//		if strings.Contains(path, "?") {
//			// Split the path into the actual path and the query string
//			parts := strings.SplitN(path, "?", 2)
//			actualPath := parts[0]
//			queryString := parts[1]
//			// Add ".html" to the actual path
//			actualPath += ".html"
//			// Reassemble the path with the query string
//			r.URL.Path = actualPath + "?" + queryString
//		} else {
//			r.URL.Path = path + ".html"
//		}
//	}
//
//	if r.URL.Path == "/.html" {
//		r.URL.Path = "/index.html"
//	}
//
//	fileServer.ServeHTTP(w, r)
//})))
