package plugin_ui

import "github.com/goccy/go-json"

type ClientEventType string

const (
	RenderTrayEvent           ClientEventType = "tray:render"            // Client wants to render the tray
	RenderTraysEvent          ClientEventType = "tray:render-all"        // Client wants to render the tray
	TrayHandlerTriggeredEvent ClientEventType = "tray:handler-triggered" // When a custom event registered by the tray is triggered
	TrayFormSubmittedEvent    ClientEventType = "tray:form-submitted"    // When a form registered by the tray is submitted
	ScreenChangedEvent        ClientEventType = "screen:changed"         // When the current screen changes
)

type RenderTrayEventPayload struct{}
type RenderTraysEventPayload struct{}

type TrayHandlerTriggeredEventPayload struct {
	EventName string `json:"eventName"`
}

type TrayFormSubmittedEventPayload struct {
	FormName string                 `json:"formName"`
	Data     map[string]interface{} `json:"data"`
}

type ScreenChangedEventPayload struct {
	Pathname string `json:"pathname"`
	Query    string `json:"query"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ServerEventType string

const (
	TrayUpdatedEvent ServerEventType = "tray:updated" // When the trays are updated
)

type TrayUpdatedEventPayload struct {
	Components interface{} `json:"components"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ClientPluginEvent is an event received from the client
type ClientPluginEvent struct {
	// ExtensionID is the "sent to"
	// If not set, the event is being sent to all plugins
	ExtensionID string          `json:"extensionId,omitempty"`
	Type        ClientEventType `json:"type"`
	Payload     interface{}     `json:"payload"`
}

// ServerPluginEvent is an event sent to the client
type ServerPluginEvent struct {
	ExtensionID string          `json:"extensionId"` // Extension ID must be set
	Type        ServerEventType `json:"type"`
	Payload     interface{}     `json:"payload"`
}

func NewClientPluginEvent(data map[string]interface{}) *ClientPluginEvent {
	extensionID, ok := data["extensionId"].(string)
	if !ok {
		extensionID = ""
	}

	eventType, ok := data["type"].(string)
	if !ok {
		return nil
	}

	payload, ok := data["payload"]
	if !ok {
		return nil
	}

	return &ClientPluginEvent{
		ExtensionID: extensionID,
		Type:        ClientEventType(eventType),
		Payload:     payload,
	}
}

func (e *ClientPluginEvent) ParsePayload(ret interface{}) bool {
	data, err := json.Marshal(e.Payload)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(data, &ret); err != nil {
		return false
	}
	return true
}

func (e *ClientPluginEvent) ParsePayloadAs(t ClientEventType, ret interface{}) bool {
	if e.Type != t {
		return false
	}
	return e.ParsePayload(ret)
}
