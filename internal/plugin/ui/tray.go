package plugin_ui

import (
	"github.com/dop251/goja"
	"github.com/samber/mo"
)

type TrayManager struct {
	ctx  *Context
	tray mo.Option[*Tray]
	// Store the last rendered component tree for diffing
	lastRenderedComponents interface{}
}

func NewTrayManager(ctx *Context) *TrayManager {
	return &TrayManager{
		ctx:  ctx,
		tray: mo.None[*Tray](),
	}
}

// renderTray is called when the client wants to render the tray
func (t *TrayManager) renderTray() {
	tray, registered := t.tray.Get()
	if !registered {
		return
	}

	newComponents, err := renderComponents(tray.renderFunc, t.lastRenderedComponents)
	if err != nil {
		t.ctx.logger.Error().Err(err).Msg("plugin: Failed to render tray")
		return
	}

	// Store for next render
	t.lastRenderedComponents = newComponents

	// Send the JSON value to the client
	t.ctx.SendEventToClient(ServerTrayUpdatedEvent, ServerTrayUpdatedEventPayload{
		Components: newComponents,
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Tray
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Tray struct {
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
func (t *TrayManager) jsNewTray(goja.FunctionCall) goja.Value {
	tray := &Tray{
		renderFunc:  nil,
		trayManager: t,
	}

	t.tray = mo.Some(tray)

	cm := &ComponentManager{
		ctx: t.ctx,
	}

	// Create a new tray object
	trayObj := t.ctx.vm.NewObject()
	_ = trayObj.Set("render", tray.jsRender)
	_ = trayObj.Set("div", cm.jsDiv)
	_ = trayObj.Set("flex", cm.jsFlex)
	_ = trayObj.Set("stack", cm.jsStack)
	_ = trayObj.Set("text", cm.jsText)
	_ = trayObj.Set("button", cm.jsButton)
	_ = trayObj.Set("input", cm.jsInput)
	_ = trayObj.Set("update", tray.jsUpdate)
	_ = trayObj.Set("onOpen", tray.jsOnOpen)
	_ = trayObj.Set("onClose", tray.jsOnClose)

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
		panic(t.trayManager.ctx.vm.NewTypeError("render requires a function"))
	}

	// Set the render function
	t.renderFunc = funcRes

	return goja.Undefined()
}

// jsUpdate takes the current state and schedules a re-render on the client
//
//	Example:
//	tray.update()
func (t *Tray) jsUpdate(call goja.FunctionCall) goja.Value {
	t.trayManager.renderTray()
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
		panic(t.trayManager.ctx.vm.NewTypeError("onOpen requires a callback function"))
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		panic(t.trayManager.ctx.vm.NewTypeError("onOpen requires a callback function"))
	}

	eventListener := t.trayManager.ctx.RegisterEventListener(ClientTrayOpenedEvent)
	payload := ClientTrayOpenedEventPayload{}

	go func() {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientTrayOpenedEvent, &payload) {
				if err := t.trayManager.ctx.scheduler.Schedule(func() error {
					_, err := callback(goja.Undefined(), t.trayManager.ctx.vm.ToValue(payload))
					return err
				}); err != nil {
					t.trayManager.ctx.logger.Error().Err(err).Msg("error running tray open callback")
				}
			}
		}
	}()
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
		panic(t.trayManager.ctx.vm.NewTypeError("onClose requires a callback function"))
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		panic(t.trayManager.ctx.vm.NewTypeError("onClose requires a callback function"))
	}

	eventListener := t.trayManager.ctx.RegisterEventListener(ClientTrayClosedEvent)
	payload := ClientTrayClosedEventPayload{}

	go func() {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientTrayClosedEvent, &payload) {
				if err := t.trayManager.ctx.scheduler.Schedule(func() error {
					_, err := callback(goja.Undefined(), t.trayManager.ctx.vm.ToValue(payload))
					return err
				}); err != nil {
					t.trayManager.ctx.logger.Error().Err(err).Msg("error running tray close callback")
				}
			}
		}
	}()

	return goja.Undefined()
}
