package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime/internal/core"
)

// newWebSocketEventHandler creates a new websocket handler for real-time event communication
func newWebSocketEventHandler(app *core.App) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {

		// Attach the websocket connection to the app instance, so it is available to other handlers
		//app.WSEventManager.Conn = c

		id := c.Locals("id").(string)

		app.WSEventManager.AddConn(id, c)
		app.Logger.Debug().Str("id", id).Msg("ws: Client connected")

		var (
			_   int
			msg []byte
			err error
		)
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					app.Logger.Debug().Msg("ws: Client disconnected")
					app.WSEventManager.RemoveConn(c.Locals("id").(string))
				} else {
					app.Logger.Debug().Msg("ws: Client disconnection")
					app.WSEventManager.RemoveConn(c.Locals("id").(string))
				}
				break
			}
			app.Logger.Debug().Msgf("ws: message received: %+v", msg)

			if err = c.WriteJSON(msg); err != nil {
				app.Logger.Err(err).Msg("ws: Failed to send message")
				break
			}
		}
	})
}

func websocketUpgradeMiddleware(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		c.Locals("userAgent", c.Get("User-Agent"))
		c.Locals("id", c.Query("id", "0"))
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
