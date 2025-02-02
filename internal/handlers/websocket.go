package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// webSocketEventHandler creates a new websocket handler for real-time event communication
func (h *Handler) webSocketEventHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// Get connection ID from query parameter
	id := c.QueryParam("id")
	if id == "" {
		id = "0"
	}

	// Add connection to manager
	h.App.WSEventManager.AddConn(id, ws)
	h.App.Logger.Debug().Str("id", id).Msg("ws: Client connected")

	for {
		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				h.App.Logger.Debug().Str("id", id).Msg("ws: Client disconnected")
			} else {
				h.App.Logger.Debug().Str("id", id).Msg("ws: Client disconnection")
			}
			h.App.WSEventManager.RemoveConn(id)
			break
		}

		h.App.Logger.Debug().Msgf("ws: message received: %+v", msg)

		// Echo the message back
		if err = ws.WriteMessage(messageType, msg); err != nil {
			h.App.Logger.Err(err).Msg("ws: Failed to send message")
			break
		}
	}

	return nil
}
