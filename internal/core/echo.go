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

	basePath := app.Config.GetBaseURLPath()
	patchedIndexHTML := []byte(nil)
	if basePath != "/" {
		if b, readErr := fs.ReadFile(distFS, "index.html"); readErr == nil {
			patchedIndexHTML = []byte(patchIndexHTMLBasePath(string(b), basePath))
		}
	}

	if app.Config.Server.Tls.Enabled {
		app.Logger.Debug().Msg("app: TLS is enabled, adding security middleware")
		e.Use(middleware.Secure())
	}

	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if basePath == "/" {
				return next(c)
			}

			req := c.Request()
			path := req.URL.Path
			if path == "" {
				path = "/"
			}

			trimPrefix := func(prefix string) bool {
				if prefix == "" || prefix == "/" {
					return false
				}

				if path == prefix {
					req.URL.Path = "/"
					req.RequestURI = "/"
					return true
				}
				withSlash := prefix + "/"
				if strings.HasPrefix(path, withSlash) {
					req.URL.Path = "/" + strings.TrimPrefix(path, withSlash)
					if req.URL.RawQuery != "" {
						req.RequestURI = req.URL.Path + "?" + req.URL.RawQuery
					} else {
						req.RequestURI = req.URL.Path
					}
					return true
				}

				return false
			}

			if trimPrefix(basePath) {
				return next(c)
			}

			forwardedPrefix := NormalizeBaseURLPath(req.Header.Get("X-Forwarded-Prefix"))
			if forwardedPrefix != "/" {
				_ = trimPrefix(forwardedPrefix)
			}

			return next(c)
		}
	})

	if !constants.IsRspackFrontend {
		if basePath != "/" && len(patchedIndexHTML) > 0 {
			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead {
						return next(c)
					}

					if shouldServePatchedIndex(c.Request().URL.Path) {
						return c.Blob(http.StatusOK, echo.MIMETextHTMLCharsetUTF8, patchedIndexHTML)
					}

					return next(c)
				}
			})
		}

		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Filesystem: http.FS(distFS),
			Browse:     true,
			HTML5:      true,
			Skipper: func(c echo.Context) bool {
				cURL := c.Request().URL
				if isReservedServerPath(cURL.RequestURI()) {
					return true // Continue to the next handler
				}
				if !strings.HasSuffix(cURL.Path, ".html") && filepath.Ext(cURL.Path) == "" {
					cURL.Path = cURL.Path + ".html"
				}
				if cURL.Path == "/.html" {
					cURL.Path = "/index.html"
				}
				return false // Continue to the filesystem handler
			},
		}))
	} else {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cURL := c.Request().URL.RequestURI()

				if isReservedServerPath(cURL) {
					return next(c)
				}

				c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
				c.Response().Header().Set("Cross-Origin-Embedder-Policy", "credentialless")

				return next(c)
			}
		})

		if basePath != "/" && len(patchedIndexHTML) > 0 {
			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead {
						return next(c)
					}

					if shouldServePatchedIndex(c.Request().URL.Path) {
						return c.Blob(http.StatusOK, echo.MIMETextHTMLCharsetUTF8, patchedIndexHTML)
					}

					return next(c)
				}
			})
		}

		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Filesystem: http.FS(distFS),
			HTML5:      true,
			Skipper: func(c echo.Context) bool {
				cURL := c.Request().URL
				if isReservedServerPath(cURL.RequestURI()) {
					return true
				}
				return false
			},
		}))
	}

	if basePath == "/" {
		app.Logger.Info().Msgf("app: Serving embedded web interface")
	} else {
		app.Logger.Info().Msgf("app: Serving embedded web interface at base URL %s", basePath)
	}

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

func patchIndexHTMLBasePath(indexHTML string, basePath string) string {
	if basePath == "/" {
		return indexHTML
	}

	runtimeScript := `<script>window.__SEANIME_BASE_URL__ = "` + basePath + `";</script>`
	if strings.Contains(indexHTML, "<head>") {
		indexHTML = strings.Replace(indexHTML, "<head>", "<head>\n    "+runtimeScript, 1)
	} else {
		indexHTML = runtimeScript + indexHTML
	}

	indexHTML = strings.ReplaceAll(indexHTML, "href=\"/", "href=\""+basePath+"/")
	indexHTML = strings.ReplaceAll(indexHTML, "src=\"/", "src=\""+basePath+"/")
	return indexHTML
}

func isReservedServerPath(path string) bool {
	return strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/events") ||
		strings.HasPrefix(path, "/assets") ||
		strings.HasPrefix(path, "/manga-downloads") ||
		strings.HasPrefix(path, "/offline-assets")
}

func shouldServePatchedIndex(path string) bool {
	if isReservedServerPath(path) {
		return false
	}

	if path == "" || path == "/" {
		return true
	}

	if strings.HasSuffix(path, ".html") {
		return true
	}

	return filepath.Ext(path) == ""
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
