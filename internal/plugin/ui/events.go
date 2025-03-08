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
	ClientRenderTrayEvent                            ClientEventType = "tray:render"                                 // Client wants to render the tray
	ClientListTrayIconsEvent                         ClientEventType = "tray:list-icons"                             // Client wants to list all icons from all plugins
	ClientTrayOpenedEvent                            ClientEventType = "tray:opened"                                 // When the tray is opened
	ClientTrayClosedEvent                            ClientEventType = "tray:closed"                                 // When the tray is closed
	ClientTrayClickedEvent                           ClientEventType = "tray:clicked"                                // When the tray is clicked
	ClientListCommandPalettesEvent                   ClientEventType = "command-palette:list"                        // When the client wants to list all command palettes
	ClientCommandPaletteOpenedEvent                  ClientEventType = "command-palette:opened"                      // When the client opens the command palette
	ClientCommandPaletteClosedEvent                  ClientEventType = "command-palette:closed"                      // When the client closes the command palette
	ClientRenderCommandPaletteEvent                  ClientEventType = "command-palette:render"                      // When the client requests the command palette to render
	ClientCommandPaletteInputEvent                   ClientEventType = "command-palette:input"                       // The client sends the current input of the command palette
	ClientCommandPaletteItemSelectedEvent            ClientEventType = "command-palette:item-selected"               // When the client selects an item from the command palette
	ClientActionRenderAnimePageButtonsEvent          ClientEventType = "action:anime-page-buttons:render"            // When the client requests the buttons to display on the anime page
	ClientActionRenderAnimePageDropdownItemsEvent    ClientEventType = "action:anime-page-dropdown-items:render"     // When the client requests the dropdown items to display on the anime page
	ClientActionRenderMangaPageButtonsEvent          ClientEventType = "action:manga-page-buttons:render"            // When the client requests the buttons to display on the manga page
	ClientActionRenderMediaCardContextMenuItemsEvent ClientEventType = "action:media-card-context-menu-items:render" // When the client requests the context menu items to display on the media card
	ClientActionRenderAnimeLibraryDropdownItemsEvent ClientEventType = "action:anime-library-dropdown-items:render"  // When the client requests the dropdown items to display on the anime library
	ClientActionClickedEvent                         ClientEventType = "action:clicked"                              // When the user clicks on an action
	ClientFormSubmittedEvent                         ClientEventType = "form:submitted"                              // When the form registered by the tray is submitted
	ClientScreenChangedEvent                         ClientEventType = "screen:changed"                              // When the current screen changes
	ClientEventHandlerTriggeredEvent                 ClientEventType = "handler:triggered"                           // When a custom event registered by the plugin is triggered
	ClientFieldRefSendValueEvent                     ClientEventType = "field-ref:send-value"                        // When the client sends the value of a field that has a ref
)

type ClientRenderTrayEventPayload struct{}
type ClientListTrayIconsEventPayload struct{}
type ClientTrayOpenedEventPayload struct{}
type ClientTrayClosedEventPayload struct{}
type ClientTrayClickedEventPayload struct{}
type ClientActionRenderAnimePageButtonsEventPayload struct{}
type ClientActionRenderAnimePageDropdownItemsEventPayload struct{}
type ClientActionRenderMangaPageButtonsEventPayload struct{}
type ClientActionRenderMediaCardContextMenuItemsEventPayload struct{}
type ClientActionRenderAnimeLibraryDropdownItemsEventPayload struct{}

type ClientListCommandPalettesEventPayload struct{}

type ClientCommandPaletteOpenedEventPayload struct{}

type ClientCommandPaletteClosedEventPayload struct{}

type ClientActionClickedEventPayload struct {
	ActionID string                 `json:"actionId"`
	Event    map[string]interface{} `json:"event"`
}

type ClientEventHandlerTriggeredEventPayload struct {
	HandlerName string                 `json:"handlerName"`
	Event       map[string]interface{} `json:"event"`
}

type ClientFormSubmittedEventPayload struct {
	FormName string                 `json:"formName"`
	Data     map[string]interface{} `json:"data"`
}

type ClientScreenChangedEventPayload struct {
	Pathname string `json:"pathname"`
	Query    string `json:"query"`
}

type ClientFieldRefSendValueEventPayload struct {
	FieldRef string      `json:"fieldRef"`
	Value    interface{} `json:"value"`
}

type ClientRenderCommandPaletteEventPayload struct{}

type ClientCommandPaletteItemSelectedEventPayload struct {
	ItemID string `json:"itemId"`
}

type ClientCommandPaletteInputEventPayload struct {
	Value string `json:"value"`
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
	ServerTrayUpdatedEvent                           ServerEventType = "tray:updated"                                 // When the trays are updated
	ServerTrayIconEvent                              ServerEventType = "tray:icon"                                    // When the tray sends its icon to the client
	ServerCommandPaletteInfoEvent                    ServerEventType = "command-palette:info"                         // When the command palette sends its state to the client
	ServerCommandPaletteUpdatedEvent                 ServerEventType = "command-palette:updated"                      // When the command palette is updated
	ServerCommandPaletteOpenEvent                    ServerEventType = "command-palette:open"                         // When the command palette is opened
	ServerCommandPaletteCloseEvent                   ServerEventType = "command-palette:close"                        // When the command palette is closed
	ServerCommandPaletteGetInputEvent                ServerEventType = "command-palette:get-input"                    // When the command palette requests the input from the client
	ServerCommandPaletteSetInputEvent                ServerEventType = "command-palette:set-input"                    // When the command palette sets the input
	ServerActionRenderAnimePageButtonsEvent          ServerEventType = "action:anime-page-buttons:updated"            // When the server renders the anime page buttons
	ServerActionRenderAnimePageDropdownItemsEvent    ServerEventType = "action:anime-page-dropdown-items:updated"     // When the server renders the anime page dropdown items
	ServerActionRenderMangaPageButtonsEvent          ServerEventType = "action:manga-page-buttons:updated"            // When the server renders the manga page buttons
	ServerActionRenderMediaCardContextMenuItemsEvent ServerEventType = "action:media-card-context-menu-items:updated" // When the server renders the media card context menu items
	ServerActionRenderAnimeLibraryDropdownItemsEvent ServerEventType = "action:anime-library-dropdown-items:updated"  // When the server renders the anime library dropdown items
	ServerFormResetEvent                             ServerEventType = "form:reset"
	ServerFormSetValuesEvent                         ServerEventType = "form:set-values"
	ServerFieldRefSetValueEvent                      ServerEventType = "field-ref:set-value" // Set the value of a field (not in a form)
	ServerFatalErrorEvent                            ServerEventType = "fatal-error"         // When the UI encounters a fatal error
	ServerScreenNavigateToEvent                      ServerEventType = "screen:navigate-to"  // Navigate to a new screen
	ServerScreenReloadEvent                          ServerEventType = "screen:reload"       // Reload the current screen
)

type ServerTrayUpdatedEventPayload struct {
	Components interface{} `json:"components"`
}

type ServerCommandPaletteUpdatedEventPayload struct {
	Placeholder string      `json:"placeholder"`
	Items       interface{} `json:"items"`
}

type ServerTrayIconEventPayload struct {
	IconURL     string `json:"iconUrl"`
	WithContent bool   `json:"withContent"`
	TooltipText string `json:"tooltipText"`
}

type ServerFormResetEventPayload struct {
	FormName     string `json:"formName"`
	FieldToReset string `json:"fieldToReset"` // If not set, the form will be reset
}

type ServerFormSetValuesEventPayload struct {
	FormName string                 `json:"formName"`
	Data     map[string]interface{} `json:"data"`
}

type ServerFieldRefSetValueEventPayload struct {
	FieldRef string      `json:"fieldRef"`
	Value    interface{} `json:"value"`
}

type ServerFieldRefGetValueEventPayload struct {
	FieldRef string `json:"fieldRef"`
}

type ServerFatalErrorEventPayload struct {
	Error string `json:"error"`
}

type ServerScreenNavigateToEventPayload struct {
	Path string `json:"path"`
}

type ServerActionRenderAnimePageButtonsEventPayload struct {
	Buttons interface{} `json:"buttons"`
}

type ServerActionRenderAnimePageDropdownItemsEventPayload struct {
	Items interface{} `json:"items"`
}

type ServerActionRenderMangaPageButtonsEventPayload struct {
	Buttons interface{} `json:"buttons"`
}

type ServerActionRenderMediaCardContextMenuItemsEventPayload struct {
	Items interface{} `json:"items"`
}

type ServerActionRenderAnimeLibraryDropdownItemsEventPayload struct {
	Items interface{} `json:"items"`
}

type ServerScreenReloadEventPayload struct{}

type ServerCommandPaletteInfoEventPayload struct {
	Placeholder      string `json:"placeholder"`
	KeyboardShortcut string `json:"keyboardShortcut"`
}

type ServerCommandPaletteOpenEventPayload struct{}

type ServerCommandPaletteCloseEventPayload struct{}

type ServerCommandPaletteGetInputEventPayload struct{}

type ServerCommandPaletteSetInputEventPayload struct {
	Value string `json:"value"`
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
