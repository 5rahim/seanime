package core

import (
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"seanime/internal/constants"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoApp(app *App, webFS *embed.FS) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	e.JSONSerializer = &CustomJSONSerializer{}
	e.StdLogger = log.Default()

	distFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	if app.Config.Server.Tls.Enabled {
		app.Logger.Debug().Msg("app: TLS is enabled, adding security middleware")
		e.Use(middleware.Secure())
	}

	if !constants.IsRspackFrontend {
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
	} else {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cUrl := c.Request().URL.RequestURI()

				if strings.HasPrefix(cUrl, "/api") ||
					strings.HasPrefix(cUrl, "/events") ||
					strings.HasPrefix(cUrl, "/assets") ||
					strings.HasPrefix(cUrl, "/manga-downloads") ||
					strings.HasPrefix(cUrl, "/offline-assets") {
					return next(c)
				}

				c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
				c.Response().Header().Set("Cross-Origin-Embedder-Policy", "credentialless")

				return next(c)
			}
		})

		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Filesystem: http.FS(distFS),
			HTML5:      true,
			Skipper: func(c echo.Context) bool {
				cUrl := c.Request().URL
				if strings.HasPrefix(cUrl.RequestURI(), "/api") ||
					strings.HasPrefix(cUrl.RequestURI(), "/events") ||
					strings.HasPrefix(cUrl.RequestURI(), "/assets") ||
					strings.HasPrefix(cUrl.RequestURI(), "/manga-downloads") ||
					strings.HasPrefix(cUrl.RequestURI(), "/offline-assets") {
					return true
				}
				return false
			},
		}))
	}

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
	serverAddr := app.Config.GetServerAddr()
	app.Logger.Info().Msgf("app: Server Address: %s", serverAddr)

	// Start the server
	go func() {
		if app.Config.Server.Tls.Enabled {
			certFile := app.Config.Server.Tls.CertPath
			keyFile := app.Config.Server.Tls.KeyPath

			// Generate certs if they don't exist
			if err := generateSelfSignedCert(certFile, keyFile, app.Logger); err != nil {
				app.Logger.Fatal().Err(err).Msg("app: Could not generate TLS certificates")
			}

			app.Logger.Info().Msg("app: Starting server with TLS enabled")
			if err := e.StartTLS(serverAddr, certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.Logger.Fatal().Err(err).Msg("app: Could not start TLS server")
			}
		} else {
			if err := e.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.Logger.Fatal().Err(err).Msg("app: Could not start server")
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	app.Logger.Info().Msg("app: Seanime started at " + app.Config.GetServerURI())
}
