package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime-server/internal/core"
	"log"
)

func websocketUpgradeMiddleware(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func createWebSocketHandler(app *core.App) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {

		app.WebsocketConn = c

		var (
			_   int
			msg []byte
			err error
		)
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				app.Logger.Err(err).Msg("ws: Failed to read message")
				break
			}
			log.Printf("recv: %s", msg)

			if err = c.WriteJSON(msg); err != nil {
				app.Logger.Err(err).Msg("ws: Failed to send message")
				break
			}
		}
	})
}
