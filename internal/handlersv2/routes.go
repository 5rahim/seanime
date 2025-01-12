package handlersv2

import (
	"net/http"
	"path/filepath"
	"seanime/internal/core"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	App *core.App
}

func InitRoutes(app *core.App, e *echo.Echo) {
	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	lechoLogger := lecho.From(*app.Logger)

	urisToSkip := []string{
		"/internal/metrics",
		"/_next",
		"/icons",
		"/events",
		"/api/v1/image-proxy",
		"/api/v1/mediastream/transcode/",
		"/api/v1/torrent-client/list",
		"/api/v1/proxy",
	}

	// Logging middleware
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lechoLogger,
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.RequestURI()
			if filepath.Ext(c.Request().URL.Path) == ".txt" ||
				filepath.Ext(c.Request().URL.Path) == ".png" ||
				filepath.Ext(c.Request().URL.Path) == ".ico" {
				return true
			}
			for _, uri := range urisToSkip {
				if uri == path || strings.HasPrefix(path, uri) {
					return true
				}
			}
			return false
		},
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add which file the request came from
			return logger.Str("file", c.Path())
		},
	}))

	// Recovery middleware
	e.Use(middleware.Recover())

	// Client ID middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if the client has a UUID cookie
			cookie, err := c.Cookie("Seanime-Client-Id")

			if err != nil || cookie.Value == "" {
				// Generate a new UUID for the client
				u := uuid.New().String()

				// Create a cookie with the UUID
				newCookie := new(http.Cookie)
				newCookie.Name = "Seanime-Client-Id"
				newCookie.Value = u
				newCookie.HttpOnly = false // Make the cookie accessible via JS
				newCookie.Expires = time.Now().Add(24 * time.Hour)

				// Set the cookie
				c.SetCookie(newCookie)

				// Store the UUID in the context for use in the request
				c.Set("Seanime-Client-Id", u)
			} else {
				// Store the existing UUID in the context for use in the request
				c.Set("Seanime-Client-Id", cookie.Value)
			}

			return next(c)
		}
	})

	h := &Handler{App: app}

	e.GET("/events", h.webSocketEventHandler)

	v1 := e.Group("/api").Group("/v1") // Commented out for now, will be used later

	v1.GET("/status", h.HandleGetStatus)
	v1.GET("/log/*", h.HandleGetLogContent)
	v1.GET("/logs/filenames", h.HandleGetLogFilenames)
	v1.DELETE("/logs", h.HandleDeleteLogs)
}

func (h *Handler) JSON(c echo.Context, code int, i interface{}) error {
	return c.JSON(code, i)
}

func (h *Handler) RespondWithData(c echo.Context, data interface{}) error {
	return c.JSON(200, NewDataResponse(data))
}

func (h *Handler) RespondWithError(c echo.Context, err error) error {
	return c.JSON(500, NewErrorResponse(err))
}
