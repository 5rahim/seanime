package handlers

import (
	"net/http"
	"seanime/internal/events"

	"github.com/goccy/go-json"
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
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				h.App.Logger.Debug().Str("id", id).Msg("ws: Client disconnected")
			} else {
				h.App.Logger.Debug().Str("id", id).Msg("ws: Client disconnection")
			}
			h.App.WSEventManager.RemoveConn(id)
			break
		}

		event, err := UnmarshalWebsocketClientEvent(msg)
		if err != nil {
			h.App.Logger.Error().Err(err).Msg("ws: Failed to unmarshal message sent from webview")
			continue
		}

		// Handle ping messages
		if event.Type == "ping" {
			timestamp := int64(0)
			if payload, ok := event.Payload.(map[string]interface{}); ok {
				if ts, ok := payload["timestamp"]; ok {
					if tsFloat, ok := ts.(float64); ok {
						timestamp = int64(tsFloat)
					} else if tsInt, ok := ts.(int64); ok {
						timestamp = tsInt
					}
				}
			}

			// Send pong response back to the same client
			h.App.WSEventManager.SendEventTo(event.ClientID, "pong", map[string]int64{"timestamp": timestamp})
			continue // Skip further processing for ping messages
		}

		h.HandleClientEvents(event)

		// h.App.Logger.Debug().Msgf("ws: message received: %+v", msg)

		// // Echo the message back
		// if err = ws.WriteMessage(messageType, msg); err != nil {
		// 	h.App.Logger.Err(err).Msg("ws: Failed to send message")
		// 	break
		// }
	}

	return nil
}

func UnmarshalWebsocketClientEvent(msg []byte) (*events.WebsocketClientEvent, error) {
	var event events.WebsocketClientEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
