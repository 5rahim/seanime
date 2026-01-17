package plugin_ui

import (
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/samber/mo"
)

type TrayManager struct {
	ctx           *Context
	tray          mo.Option[*Tray]
	lastUpdatedAt time.Time
	updateMutex   sync.Mutex

	componentManager *ComponentManager
}

func NewTrayManager(ctx *Context) *TrayManager {
	return &TrayManager{
		ctx:              ctx,
		tray:             mo.None[*Tray](),
		componentManager: &ComponentManager{ctx: ctx},
	}
}

// renderTrayScheduled renders the new component tree.
// This function is unsafe because it is not thread-safe and should be scheduled.
func (t *TrayManager) renderTrayScheduled() {
	t.updateMutex.Lock()
	defer t.updateMutex.Unlock()

	tray, registered := t.tray.Get()
	if !registered {
		return
	}

	if !tray.WithContent {
		return
	}

	// Rate limit updates
	//if time.Since(t.lastUpdatedAt) < time.Millisecond*200 {
	//	return
	//}

	t.lastUpdatedAt = time.Now()

	t.ctx.scheduler.ScheduleAsync(func() error {
		// t.ctx.logger.Trace().Msg("plugin: Rendering tray")
		newComponents, err := t.componentManager.renderComponents(tray.renderFunc)
		if err != nil {
			t.ctx.logger.Error().Err(err).Msg("plugin: Failed to render tray")
			t.ctx.handleException(err)
			return nil
		}

		// t.ctx.logger.Trace().Msg("plugin: Sending tray update to client")
		// Send the JSON value to the client
		t.ctx.SendEventToClient(ServerTrayUpdatedEvent, ServerTrayUpdatedEventPayload{
			Components: newComponents,
		})
		return nil
	})
}

// sendIconToClient sends the tray icon to the client after it's been requested.
func (t *TrayManager) sendIconToClient() {
	if tray, registered := t.tray.Get(); registered {
		t.ctx.SendEventToClient(ServerTrayIconEvent, ServerTrayIconEventPayload{
			ExtensionID:   t.ctx.ext.ID,
			ExtensionName: t.ctx.ext.Name,
			IconURL:       tray.IconURL,
			WithContent:   tray.WithContent,
			TooltipText:   tray.TooltipText,
			BadgeNumber:   tray.BadgeNumber,
			BadgeIntent:   tray.BadgeIntent,
			Width:         tray.Width,
			MinHeight:     tray.MinHeight,
			IsDrawer:      tray.IsDrawer,
		})
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Tray
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Tray struct {
	// WithContent is used to determine if the tray has any content
	// If false, only the tray icon will be rendered and tray.render() will be ignored
	WithContent bool `json:"withContent"`

	IconURL     string `json:"iconUrl"`
	TooltipText string `json:"tooltipText"`
	BadgeNumber int    `json:"badgeNumber"`
	BadgeIntent string `json:"badgeIntent"`
	Width       string `json:"width,omitempty"`
	MinHeight   string `json:"minHeight,omitempty"`
	IsDrawer    bool   `json:"isDrawer,omitempty"`

	renderFunc  func(goja.FunctionCall) goja.Value
	trayManager *TrayManager
}

type Component struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Props map[string]interface{} `json:"props"`
	Key   string                 `json:"key,omitempty"`
}

// jsNewTray
//
//	Example:
//	const tray = ctx.newTray()
func (t *TrayManager) jsNewTray(call goja.FunctionCall) goja.Value {
	tray := &Tray{
		renderFunc:  nil,
		trayManager: t,
		WithContent: true,
	}

	props := call.Arguments
	if len(props) > 0 {
		propsObj := props[0].Export().(map[string]interface{})
		if propsObj["withContent"] != nil {
			tray.WithContent, _ = propsObj["withContent"].(bool)
		}
		if propsObj["iconUrl"] != nil {
			tray.IconURL, _ = propsObj["iconUrl"].(string)
		}
		if propsObj["tooltipText"] != nil {
			tray.TooltipText, _ = propsObj["tooltipText"].(string)
		}
		if propsObj["isDrawer"] != nil {
			tray.IsDrawer, _ = propsObj["isDrawer"].(bool)
		}
		if propsObj["width"] != nil {
			tray.Width, _ = propsObj["width"].(string)
		}
		if propsObj["minHeight"] != nil {
			tray.MinHeight, _ = propsObj["minHeight"].(string)
		}
	}

	t.tray = mo.Some(tray)

	// Create a new tray object
	trayObj := t.ctx.vm.NewObject()
	_ = trayObj.Set("render", tray.jsRender)
	_ = trayObj.Set("htm", tray.jsHtm)
	_ = trayObj.Set("update", tray.jsUpdate)
	_ = trayObj.Set("onOpen", tray.jsOnOpen)
	_ = trayObj.Set("onClose", tray.jsOnClose)
	_ = trayObj.Set("onClick", tray.jsOnClick)
	_ = trayObj.Set("open", tray.jsOpen)
	_ = trayObj.Set("close", tray.jsClose)
	_ = trayObj.Set("updateBadge", tray.jsUpdateBadge)

	// Register components
	_ = trayObj.Set("div", t.componentManager.jsDiv)
	_ = trayObj.Set("flex", t.componentManager.jsFlex)
	_ = trayObj.Set("stack", t.componentManager.jsStack)
	_ = trayObj.Set("text", t.componentManager.jsText)
	_ = trayObj.Set("button", t.componentManager.jsButton)
	_ = trayObj.Set("anchor", t.componentManager.jsAnchor)
	_ = trayObj.Set("input", t.componentManager.jsInput)
	_ = trayObj.Set("radioGroup", t.componentManager.jsRadioGroup)
	_ = trayObj.Set("switch", t.componentManager.jsSwitch)
	_ = trayObj.Set("checkbox", t.componentManager.jsCheckbox)
	_ = trayObj.Set("select", t.componentManager.jsSelect)
	_ = trayObj.Set("css", t.componentManager.jsCSS)
	_ = trayObj.Set("tooltip", t.componentManager.jsTooltip)
	_ = trayObj.Set("modal", t.componentManager.jsModal)
	_ = trayObj.Set("dropdownMenu", t.componentManager.jsDropdownMenu)
	_ = trayObj.Set("dropdownMenuItem", t.componentManager.jsDropdownMenuItem)
	_ = trayObj.Set("dropdownMenuSeparator", t.componentManager.jsDropdownMenuSeparator)
	_ = trayObj.Set("dropdownMenuLabel", t.componentManager.jsDropdownMenuLabel)
	_ = trayObj.Set("popover", t.componentManager.jsPopover)
	_ = trayObj.Set("a", t.componentManager.jsA)
	_ = trayObj.Set("p", t.componentManager.jsP)
	_ = trayObj.Set("alert", t.componentManager.jsAlert)
	_ = trayObj.Set("tabs", t.componentManager.jsTabs)
	_ = trayObj.Set("tabsList", t.componentManager.jsTabsList)
	_ = trayObj.Set("tabsTrigger", t.componentManager.jsTabsTrigger)
	_ = trayObj.Set("tabsContent", t.componentManager.jsTabsContent)
	_ = trayObj.Set("badge", t.componentManager.jsBadge)
	_ = trayObj.Set("span", t.componentManager.jsSpan)
	_ = trayObj.Set("img", t.componentManager.jsImg)

	return trayObj
}

/////

// jsRender registers a function to be called when the tray is rendered/updated
//
//	Example:
//	tray.render(() => flex)
func (t *Tray) jsRender(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		t.trayManager.ctx.handleTypeError("render requires a function")
		return goja.Undefined()
	}

	// Set the render function
	t.renderFunc = funcRes

	return goja.Undefined()
}

// jsHtm registers a function that returns an HTM string to be parsed and rendered
//
//	Example:
//	tray.htm(() => `<stack><text text="Hello" /></stack>`)
func (t *Tray) jsHtm(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		t.trayManager.ctx.handleTypeError("htm requires a function")
		return goja.Undefined()
	}

	// Create a wrapper function that parses the HTM string and returns components
	t.renderFunc = func(call goja.FunctionCall) goja.Value {
		// return htm string
		htmValue := funcRes(goja.FunctionCall{})
		htmString, ok := htmValue.Export().(string)
		if !ok {
			t.trayManager.ctx.handleTypeError("htm function must return a string")
			return goja.Undefined()
		}
		htmString = cleanHTMLString(htmString)

		// Parse the HTM string into components
		components, err := t.trayManager.componentManager.parseHTM(htmString)
		if err != nil {
			t.trayManager.ctx.logger.Error().Err(err).Msg("plugin: Failed to parse HTM string")
			return goja.Undefined()
		}

		// Convert back to goja.Value
		return t.trayManager.ctx.vm.ToValue(components)
	}

	return goja.Undefined()
}

// jsUpdate schedules a re-render on the client
//
//	Example:
//	tray.update()
func (t *Tray) jsUpdate(call goja.FunctionCall) goja.Value {
	// Update the context's lastUIUpdateAt to prevent duplicate updates
	t.trayManager.ctx.uiUpdateMu.Lock()
	t.trayManager.ctx.lastUIUpdateAt = time.Now()
	t.trayManager.ctx.uiUpdateMu.Unlock()

	t.trayManager.renderTrayScheduled()
	return goja.Undefined()
}

// jsOpen
//
//	Example:
//	tray.open()
func (t *Tray) jsOpen(call goja.FunctionCall) goja.Value {
	t.trayManager.ctx.SendEventToClient(ServerTrayOpenEvent, ServerTrayOpenEventPayload{
		ExtensionID: t.trayManager.ctx.ext.ID,
	})
	return goja.Undefined()
}

// jsClose
//
//	Example:
//	tray.close()
func (t *Tray) jsClose(call goja.FunctionCall) goja.Value {
	t.trayManager.ctx.SendEventToClient(ServerTrayCloseEvent, ServerTrayCloseEventPayload{
		ExtensionID: t.trayManager.ctx.ext.ID,
	})
	return goja.Undefined()
}

// jsUpdateBadge
//
//	Example:
//	tray.updateBadge({ number: 1, intent: "success" })
func (t *Tray) jsUpdateBadge(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		t.trayManager.ctx.handleTypeError("updateBadge requires a callback function")
		return goja.Undefined()
	}

	propsObj, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		t.trayManager.ctx.handleTypeError("updateBadge requires a callback function")
		return goja.Undefined()
	}

	number, ok := propsObj["number"].(int64)
	if !ok {
		t.trayManager.ctx.handleTypeError("updateBadge: number must be an integer")
		return goja.Undefined()
	}

	intent, ok := propsObj["intent"].(string)
	if !ok {
		intent = "info"
	}

	t.BadgeNumber = int(number)
	t.BadgeIntent = intent

	t.trayManager.ctx.SendEventToClient(ServerTrayBadgeUpdatedEvent, ServerTrayBadgeUpdatedEventPayload{
		BadgeNumber: t.BadgeNumber,
		BadgeIntent: t.BadgeIntent,
	})
	return goja.Undefined()
}

// jsOnOpen
//
//	Example:
//	tray.onOpen(() => {
//		console.log("tray opened by the user")
//	})
func (t *Tray) jsOnOpen(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		t.trayManager.ctx.handleTypeError("onOpen requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		t.trayManager.ctx.handleTypeError("onOpen requires a callback function")
		return goja.Undefined()
	}

	eventListener := t.trayManager.ctx.RegisterEventListener(ClientTrayOpenedEvent)
	payload := ClientTrayOpenedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientTrayOpenedEvent, &payload) {
			t.trayManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), t.trayManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					t.trayManager.ctx.logger.Error().Err(err).Msg("plugin: Error running tray open callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsOnClick
//
//	Example:
//	tray.onClick(() => {
//		console.log("tray clicked by the user")
//	})
func (t *Tray) jsOnClick(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		t.trayManager.ctx.handleTypeError("onClick requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		t.trayManager.ctx.handleTypeError("onClick requires a callback function")
		return goja.Undefined()
	}

	eventListener := t.trayManager.ctx.RegisterEventListener(ClientTrayClickedEvent)
	payload := ClientTrayClickedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientTrayClickedEvent, &payload) {
			t.trayManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), t.trayManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					t.trayManager.ctx.logger.Error().Err(err).Msg("plugin: Error running tray click callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsOnClose
//
//	Example:
//	tray.onClose(() => {
//		console.log("tray closed by the user")
//	})
func (t *Tray) jsOnClose(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		t.trayManager.ctx.handleTypeError("onClose requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		t.trayManager.ctx.handleTypeError("onClose requires a callback function")
		return goja.Undefined()
	}

	eventListener := t.trayManager.ctx.RegisterEventListener(ClientTrayClosedEvent)
	payload := ClientTrayClosedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientTrayClosedEvent, &payload) {
			t.trayManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), t.trayManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					t.trayManager.ctx.logger.Error().Err(err).Msg("plugin: Error running tray close callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}
