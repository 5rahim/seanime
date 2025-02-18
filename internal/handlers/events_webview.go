package handlers

func (h *Handler) HandleWebviewEvents(event *WebsocketClientEvent) {

	h.App.Logger.Debug().Msgf("ws: message received: %+v", event)
}
