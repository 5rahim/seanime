package handlers

import "seanime/internal/events"

func (h *Handler) HandleClientEvents(event *events.WebsocketClientEvent) {

	//h.App.Logger.Debug().Msgf("ws: message received: %+v", event)

	if h.App.WSEventManager != nil {
		h.App.WSEventManager.OnClientEvent(event)
	}
}
