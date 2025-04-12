package plugin_ui

import (
	goja_util "seanime/internal/util/goja"
	"seanime/internal/util/result"
	"slices"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// CommandPaletteManager is a manager for the command palette.
// Unlike the Tray, command palette items are not reactive to state changes.
// They are only rendered when the setItems function is called or the refresh function is called.
type CommandPaletteManager struct {
	ctx              *Context
	updateMutex      sync.Mutex
	lastUpdated      time.Time
	componentManager *ComponentManager

	placeholder      string
	keyboardShortcut string

	// registered is true if the command palette has been registered
	registered bool

	items         *result.Map[string, *commandItem]
	renderedItems []*CommandItemJSON // Store rendered items when setItems is called
}

type (
	commandItem struct {
		index        int
		id           string
		label        string
		value        string
		filterType   string // "includes" or "startsWith" or ""
		heading      string
		renderFunc   func(goja.FunctionCall) goja.Value
		onSelectFunc func(goja.FunctionCall) goja.Value
	}

	// CommandItemJSON is the JSON representation of a command item.
	// It is used to send the command item to the client.
	CommandItemJSON struct {
		Index      int         `json:"index"`
		ID         string      `json:"id"`
		Label      string      `json:"label"`
		Value      string      `json:"value"`
		FilterType string      `json:"filterType"`
		Heading    string      `json:"heading"`
		Components interface{} `json:"components"`
	}
)

func NewCommandPaletteManager(ctx *Context) *CommandPaletteManager {
	return &CommandPaletteManager{
		ctx:              ctx,
		componentManager: &ComponentManager{ctx: ctx},
		items:            result.NewResultMap[string, *commandItem](),
		renderedItems:    make([]*CommandItemJSON, 0),
	}
}

type NewCommandPaletteOptions struct {
	Placeholder      string `json:"placeholder,omitempty"`
	KeyboardShortcut string `json:"keyboardShortcut,omitempty"`
}

// sendInfoToClient sends the command palette info to the client after it's been requested.
func (c *CommandPaletteManager) sendInfoToClient() {
	if c.registered {
		c.ctx.SendEventToClient(ServerCommandPaletteInfoEvent, ServerCommandPaletteInfoEventPayload{
			Placeholder:      c.placeholder,
			KeyboardShortcut: c.keyboardShortcut,
		})
	}
}

func (c *CommandPaletteManager) jsNewCommandPalette(options NewCommandPaletteOptions) goja.Value {
	c.registered = true
	c.keyboardShortcut = options.KeyboardShortcut
	c.placeholder = options.Placeholder

	cmdObj := c.ctx.vm.NewObject()

	_ = cmdObj.Set("setItems", func(items []interface{}) {
		c.items.Clear()

		for idx, item := range items {
			itemMap := item.(map[string]interface{})
			id := uuid.New().String()
			label, _ := itemMap["label"].(string)
			value, ok := itemMap["value"].(string)
			if !ok {
				c.ctx.handleTypeError("value must be a string")
				return
			}
			filterType, _ := itemMap["filterType"].(string)
			if filterType != "includes" && filterType != "startsWith" && filterType != "" {
				c.ctx.handleTypeError("filterType must be 'includes', 'startsWith'")
				return
			}
			heading, _ := itemMap["heading"].(string)
			renderFunc, ok := itemMap["render"].(func(goja.FunctionCall) goja.Value)
			if len(label) == 0 && !ok {
				c.ctx.handleTypeError("label or render function must be provided")
				return
			}
			onSelectFunc, ok := itemMap["onSelect"].(func(goja.FunctionCall) goja.Value)
			if !ok {
				c.ctx.handleTypeError("onSelect must be a function")
				return
			}

			c.items.Set(id, &commandItem{
				index:        idx,
				id:           id,
				label:        label,
				value:        value,
				filterType:   filterType,
				heading:      heading,
				renderFunc:   renderFunc,
				onSelectFunc: onSelectFunc,
			})
		}

		// Convert the items to JSON
		itemsJSON := make([]*CommandItemJSON, 0)
		c.items.Range(func(key string, value *commandItem) bool {
			itemsJSON = append(itemsJSON, value.ToJSON(c.ctx, c.componentManager, c.ctx.scheduler))
			return true
		})
		// Store the converted items
		c.renderedItems = itemsJSON

		c.renderCommandPaletteScheduled()
	})

	_ = cmdObj.Set("refresh", func() {
		// Convert the items to JSON
		itemsJSON := make([]*CommandItemJSON, 0)
		c.items.Range(func(key string, value *commandItem) bool {
			itemsJSON = append(itemsJSON, value.ToJSON(c.ctx, c.componentManager, c.ctx.scheduler))
			return true
		})

		c.renderedItems = itemsJSON

		c.renderCommandPaletteScheduled()
	})

	_ = cmdObj.Set("setPlaceholder", func(placeholder string) {
		c.placeholder = placeholder
		c.renderCommandPaletteScheduled()
	})

	_ = cmdObj.Set("open", func() {
		c.ctx.SendEventToClient(ServerCommandPaletteOpenEvent, ServerCommandPaletteOpenEventPayload{})
	})

	_ = cmdObj.Set("close", func() {
		c.ctx.SendEventToClient(ServerCommandPaletteCloseEvent, ServerCommandPaletteCloseEventPayload{})
	})

	_ = cmdObj.Set("setInput", func(input string) {
		c.ctx.SendEventToClient(ServerCommandPaletteSetInputEvent, ServerCommandPaletteSetInputEventPayload{
			Value: input,
		})
	})

	_ = cmdObj.Set("getInput", func() string {
		c.ctx.SendEventToClient(ServerCommandPaletteGetInputEvent, ServerCommandPaletteGetInputEventPayload{})

		eventListener := c.ctx.RegisterEventListener(ClientCommandPaletteInputEvent)
		defer c.ctx.UnregisterEventListener(eventListener.ID)

		timeout := time.After(1500 * time.Millisecond)
		input := make(chan string)

		eventListener.SetCallback(func(event *ClientPluginEvent) {
			payload := ClientCommandPaletteInputEventPayload{}
			if event.ParsePayloadAs(ClientCommandPaletteInputEvent, &payload) {
				input <- payload.Value
			}
		})

		// go func() {
		// 	for event := range eventListener.Channel {
		// 		if event.ParsePayloadAs(ClientCommandPaletteInputEvent, &payload) {
		// 			input <- payload.Value
		// 		}
		// 	}
		// }()

		select {
		case <-timeout:
			return ""
		case input := <-input:
			return input
		}
	})

	// jsOnOpen
	//
	//	Example:
	//	commandPalette.onOpen(() => {
	//		console.log("command palette opened by the user")
	//	})
	_ = cmdObj.Set("onOpen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			c.ctx.handleTypeError("onOpen requires a callback function")
		}

		callback, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			c.ctx.handleTypeError("onOpen requires a callback function")
		}

		eventListener := c.ctx.RegisterEventListener(ClientCommandPaletteOpenedEvent)

		eventListener.SetCallback(func(event *ClientPluginEvent) {
			payload := ClientCommandPaletteOpenedEventPayload{}
			if event.ParsePayloadAs(ClientCommandPaletteOpenedEvent, &payload) {
				c.ctx.scheduler.ScheduleAsync(func() error {
					_, err := callback(goja.Undefined(), c.ctx.vm.ToValue(map[string]interface{}{}))
					return err
				})
			}
		})

		// go func() {
		// 	for event := range eventListener.Channel {
		// 		if event.ParsePayloadAs(ClientCommandPaletteOpenedEvent, &payload) {
		// 			c.ctx.scheduler.ScheduleAsync(func() error {
		// 				_, err := callback(goja.Undefined(), c.ctx.vm.ToValue(map[string]interface{}{}))
		// 				if err != nil {
		// 					c.ctx.logger.Error().Err(err).Msg("plugin: Error running command palette open callback")
		// 				}
		// 				return err
		// 			})
		// 		}
		// 	}
		// }()
		return goja.Undefined()
	})

	// jsOnClose
	//
	//	Example:
	//	commandPalette.onClose(() => {
	//		console.log("command palette closed by the user")
	//	})
	_ = cmdObj.Set("onClose", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			c.ctx.handleTypeError("onClose requires a callback function")
		}

		callback, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			c.ctx.handleTypeError("onClose requires a callback function")
		}

		eventListener := c.ctx.RegisterEventListener(ClientCommandPaletteClosedEvent)

		eventListener.SetCallback(func(event *ClientPluginEvent) {
			payload := ClientCommandPaletteClosedEventPayload{}
			if event.ParsePayloadAs(ClientCommandPaletteClosedEvent, &payload) {
				c.ctx.scheduler.ScheduleAsync(func() error {
					_, err := callback(goja.Undefined(), c.ctx.vm.ToValue(map[string]interface{}{}))
					return err
				})
			}
		})

		// go func() {
		// 	for event := range eventListener.Channel {
		// 		if event.ParsePayloadAs(ClientCommandPaletteClosedEvent, &payload) {
		// 			c.ctx.scheduler.ScheduleAsync(func() error {
		// 				_, err := callback(goja.Undefined(), c.ctx.vm.ToValue(map[string]interface{}{}))
		// 				if err != nil {
		// 					c.ctx.logger.Error().Err(err).Msg("plugin: Error running command palette close callback")
		// 				}
		// 				return err
		// 			})
		// 		}
		// 	}
		// }()
		return goja.Undefined()
	})

	eventListener := c.ctx.RegisterEventListener(ClientCommandPaletteItemSelectedEvent)
	eventListener.SetCallback(func(event *ClientPluginEvent) {
		payload := ClientCommandPaletteItemSelectedEventPayload{}
		if event.ParsePayloadAs(ClientCommandPaletteItemSelectedEvent, &payload) {
			c.ctx.scheduler.ScheduleAsync(func() error {
				item, found := c.items.Get(payload.ItemID)
				if found {
					_ = item.onSelectFunc(goja.FunctionCall{})
				}
				return nil
			})
		}
	})
	// go func() {
	// 	eventListener := c.ctx.RegisterEventListener(ClientCommandPaletteItemSelectedEvent)
	// 	payload := ClientCommandPaletteItemSelectedEventPayload{}

	// 	for event := range eventListener.Channel {
	// 		if event.ParsePayloadAs(ClientCommandPaletteItemSelectedEvent, &payload) {
	// 			item, found := c.items.Get(payload.ItemID)
	// 			if found {
	// 				c.ctx.scheduler.ScheduleAsync(func() error {
	// 					_ = item.onSelectFunc(goja.FunctionCall{})
	// 					return nil
	// 				})
	// 			}
	// 		}
	// 	}
	// }()

	// Register components
	_ = cmdObj.Set("div", c.componentManager.jsDiv)
	_ = cmdObj.Set("flex", c.componentManager.jsFlex)
	_ = cmdObj.Set("stack", c.componentManager.jsStack)
	_ = cmdObj.Set("text", c.componentManager.jsText)
	_ = cmdObj.Set("button", c.componentManager.jsButton)

	return cmdObj
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *commandItem) ToJSON(ctx *Context, componentManager *ComponentManager, scheduler *goja_util.Scheduler) *CommandItemJSON {

	var components interface{}
	if c.renderFunc != nil {
		var err error
		components, err = componentManager.renderComponents(c.renderFunc)
		if err != nil {
			ctx.logger.Error().Err(err).Msg("plugin: Failed to render command palette item")
			ctx.handleException(err)
			return nil
		}
	}

	// Reset the last rendered components, we don't care about diffing
	componentManager.lastRenderedComponents = nil

	return &CommandItemJSON{
		Index:      c.index,
		ID:         c.id,
		Label:      c.label,
		Value:      c.value,
		FilterType: c.filterType,
		Heading:    c.heading,
		Components: components,
	}
}

func (c *CommandPaletteManager) renderCommandPaletteScheduled() {
	c.updateMutex.Lock()
	defer c.updateMutex.Unlock()

	if !c.registered {
		return
	}

	slices.SortFunc(c.renderedItems, func(a, b *CommandItemJSON) int {
		return a.Index - b.Index
	})

	c.ctx.SendEventToClient(ServerCommandPaletteUpdatedEvent, ServerCommandPaletteUpdatedEventPayload{
		Placeholder: c.placeholder,
		Items:       c.renderedItems,
	})
}
