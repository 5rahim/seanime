package plugin_ui

import "github.com/goccy/go-json"

/////////////////////////////////////////////////////////////////////////////////////
// Client to server
/////////////////////////////////////////////////////////////////////////////////////

type ClientEventType string

// ClientPluginEvent is an event received from the client
type ClientPluginEvent struct {
	// ExtensionID is the "sent to"
	// If not set, the event is being sent to all plugins
	ExtensionID string          `json:"extensionId,omitempty"`
	Type        ClientEventType `json:"type"`
	Payload     interface{}     `json:"payload"`
}

const (
	ClientRenderTrayEvent           ClientEventType = "tray:render"            // Client wants to render the tray
	ClientRenderTraysEvent          ClientEventType = "tray:render-all"        // Client wants to render the tray
	ClientTrayHandlerTriggeredEvent ClientEventType = "tray:handler-triggered" // When a custom event registered by the tray is triggered
	ClientTrayOpenedEvent           ClientEventType = "tray:opened"            // When the tray is opened
	ClientTrayClosedEvent           ClientEventType = "tray:closed"            // When the tray is closed
	ClientFormSubmittedEvent        ClientEventType = "form:submitted"         // When the form registered by the tray is submitted
	ClientScreenChangedEvent        ClientEventType = "screen:changed"         // When the current screen changes
)

type ClientRenderTrayEventPayload struct{}
type ClientRenderTraysEventPayload struct{}
type ClientTrayOpenedEventPayload struct{}
type ClientTrayClosedEventPayload struct{}

type ClientTrayHandlerTriggeredEventPayload struct {
	EventName string `json:"eventName"`
}

type ClientFormSubmittedEventPayload struct {
	FormName string                 `json:"formName"`
	Data     map[string]interface{} `json:"data"`
}

type ClientScreenChangedEventPayload struct {
	Pathname string `json:"pathname"`
	Query    string `json:"query"`
}

/////////////////////////////////////////////////////////////////////////////////////
// Server to client
/////////////////////////////////////////////////////////////////////////////////////

type ServerEventType string

// ServerPluginEvent is an event sent to the client
type ServerPluginEvent struct {
	ExtensionID string          `json:"extensionId"` // Extension ID must be set
	Type        ServerEventType `json:"type"`
	Payload     interface{}     `json:"payload"`
}

const (
	ServerTrayUpdatedEvent ServerEventType = "tray:updated" // When the trays are updated
	ServerFormResetEvent   ServerEventType = "form:reset"
	ServerFatalErrorEvent  ServerEventType = "fatal-error" // When the UI encounters a fatal error
)

type ServerTrayUpdatedEventPayload struct {
	Components interface{} `json:"components"`
}

type ServerFormResetEventPayload struct {
	FormName     string `json:"formName"`
	FieldToReset string `json:"fieldToReset"` // If not set, the form will be reset
}

type ServerFatalErrorEventPayload struct {
	Error string `json:"error"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
