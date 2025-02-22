package plugin_ui

import "github.com/goccy/go-json"

// ClientPluginEvent is an event sent from the client
type ClientPluginEvent struct {
	ExtensionID string      `json:"extensionId,omitempty"` // If not set, the event is sent to all plugins
	Type        EventType   `json:"type"`
	Payload     interface{} `json:"payload"`
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
		Type:        EventType(eventType),
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

func (e *ClientPluginEvent) ParsePayloadAs(t EventType, ret interface{}) bool {
	if e.Type != t {
		return false
	}
	return e.ParsePayload(ret)
}

type EventType string

const (
	TrayCustomEventTriggered EventType = "trayCustomEventTriggered" // When a custom event registered by the tray is triggered
	TrayFormSubmitted        EventType = "trayFormSubmitted"        // When a form registered by the tray is submitted
	ScreenChanged            EventType = "screenChanged"            // When the current screen changes
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TrayCustomEventPayload struct {
	EventName string `json:"eventName"`
}

type TrayFormSubmittedPayload struct {
	FormName string                 `json:"formName"`
	Data     map[string]interface{} `json:"data"`
}

type ScreenChangedPayload struct {
	Pathname string `json:"pathname"`
	Query    string `json:"query"`
}
