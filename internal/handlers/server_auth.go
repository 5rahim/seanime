package handlers

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) OptionalAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h.App.Config.Server.Password == "" {
			return next(c)
		}

		path := c.Request().URL.Path
		password := c.Request().Header.Get("X-Seanime-Password")

		if path == "/api/v1/auth/login" ||
			path == "/api/v1/auth/logout" ||
			path == "/api/v1/status" ||
			path == "/events" {

			if path == "/api/v1/status" {
				if password != h.App.Config.Server.Password {
					c.Set("unauthenticated", true)
				}
			}

			return next(c)
		}

		if password == h.App.Config.Server.Password {
			return next(c)
		}

		// Handle Nakama client connections
		if h.App.Settings.GetNakama().Enabled && h.App.Settings.GetNakama().IsHost {
			// Verify the Nakama host password in the client request
			nakamaPassword := c.Request().Header.Get("X-Seanime-Nakama-Password")

			// Allow WebSocket connections for peer-to-host communication
			if path == "/api/v1/nakama/ws" {
				if nakamaPassword == h.App.Settings.GetNakama().HostPassword {
					c.Response().Header().Set("X-Seanime-Nakama-Is-Client", "true")
					return next(c)
				}
			}

			// Only allow the following paths to be accessed by Nakama clients
			if strings.HasPrefix(path, "/api/v1/nakama/host/") {
				if nakamaPassword == h.App.Settings.GetNakama().HostPassword {
					c.Response().Header().Set("X-Seanime-Nakama-Is-Client", "true")
					return next(c)
				}
			}
			// Handle public Nakama paths (e.g. streaming endpoints)
			// For these public paths, we don't check the password header because they can be accessed by anyone
			// Instead we check if a query parameter is present
			if strings.HasPrefix(path, "/api/v1/nakama/public") {
				if c.QueryParam("nakama_password") == h.App.Settings.GetNakama().HostPassword {
					return next(c)
				}
			}
		}

		return h.RespondWithError(c, errors.New("UNAUTHENTICATED"))
	}
}
