package handlers

import (
	"errors"

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

		return h.RespondWithError(c, errors.New("UNAUTHENTICATED"))
	}
}
