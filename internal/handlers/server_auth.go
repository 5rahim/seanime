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
		passwordHash := c.Request().Header.Get("X-Seanime-Token")

		// Allow the following paths to be accessed by anyone
		if path == "/api/v1/auth/login" || // for auth
			path == "/api/v1/auth/logout" || // for auth
			path == "/api/v1/status" || // for interface
			path == "/events" || // for server events
			strings.HasPrefix(path, "/api/v1/directstream") || // used by media players
			// strings.HasPrefix(path, "/api/v1/mediastream") || // used by media players // NODE: DO NOT
			strings.HasPrefix(path, "/api/v1/mediastream/att/") || // used by media players
			strings.HasPrefix(path, "/api/v1/mediastream/direct") || // used by media players
			strings.HasPrefix(path, "/api/v1/mediastream/transcode/") || // used by media players
			strings.HasPrefix(path, "/api/v1/mediastream/subs/") || // used by media players
			strings.HasPrefix(path, "/api/v1/image-proxy") || // used by img tag
			strings.HasPrefix(path, "/api/v1/proxy") || // used by video players
			strings.HasPrefix(path, "/api/v1/manga/local-page") || // used by img tag
			strings.HasPrefix(path, "/api/v1/torrentstream/stream/") || // accessible by media players
			strings.HasPrefix(path, "/api/v1/nakama/stream") { // accessible by media players

			if path == "/api/v1/status" {
				// allow status requests by anyone but mark as unauthenticated
				// so we can filter out critical info like settings
				if passwordHash != h.App.ServerPasswordHash {
					c.Set("unauthenticated", true)
				}
			}

			return next(c)
		}

		if passwordHash == h.App.ServerPasswordHash {
			return next(c)
		}

		// Check HMAC token in query parameter
		token := c.Request().URL.Query().Get("token")
		if token != "" {
			hmacAuth := h.App.GetServerPasswordHMACAuth()
			_, err := hmacAuth.ValidateToken(token, path)
			if err == nil {
				return next(c)
			} else {
				h.App.Logger.Debug().Err(err).Str("path", path).Msg("server auth: HMAC token validation failed")
			}
		}

		// Handle Nakama client connections
		if h.App.Settings.GetNakama().Enabled && h.App.Settings.GetNakama().IsHost {
			// Verify the Nakama host password in the client request
			nakamaPasswordHeader := c.Request().Header.Get("X-Seanime-Nakama-Token")

			// Allow WebSocket connections for peer-to-host communication
			if path == "/api/v1/nakama/ws" {
				if nakamaPasswordHeader == h.App.Settings.GetNakama().HostPassword {
					c.Response().Header().Set("X-Seanime-Nakama-Is-Client", "true")
					return next(c)
				}
			}

			// Only allow the following paths to be accessed by Nakama clients
			if strings.HasPrefix(path, "/api/v1/nakama/host/") {
				if nakamaPasswordHeader == h.App.Settings.GetNakama().HostPassword {
					c.Response().Header().Set("X-Seanime-Nakama-Is-Client", "true")
					return next(c)
				}
			}
		}

		return h.RespondWithError(c, errors.New("UNAUTHENTICATED"))
	}
}
