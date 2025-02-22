package plugin_ui

import (
	"seanime/internal/util/result"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/samber/mo"
)

type TrayManager struct {
	ctx  *Context
	tray mo.Option[*Tray]
}

type TrayRenderEvent struct {
	Components interface{} `json:"components"`
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

	if tray.renderFunc == nil {
		t.ctx.logger.Warn().Msg("plugin: Tray is registered but no render function is set")
		return
	}

	jsonValue := t.getComponentsJSON(tray)

	// Send the JSON value to the client
	t.ctx.SendEventToClient(TrayUpdatedEvent, TrayRenderEvent{
		Components: jsonValue,
	})
}

func (t *TrayManager) getComponentsJSON(tray *Tray) interface{} {
	if tray == nil {
		return nil
	}

	// Call the render function
	value := tray.renderFunc(goja.FunctionCall{})

	// Convert the value to a JSON string
	v, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	var ret interface{}
	err = json.Unmarshal(v, &ret)
	if err != nil {
		return nil
	}

	return ret
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Tray
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Tray struct {
	components  *result.Map[string, Component]
	renderFunc  func(goja.FunctionCall) goja.Value
	trayManager *TrayManager
}

type ComponentType string

const (
	ComponentTypeButton ComponentType = "button"
	ComponentTypeInput  ComponentType = "input"
	ComponentTypeFlex   ComponentType = "flex"
	ComponentTypeText   ComponentType = "text"
)

type Component struct {
	ID    string                 `json:"id"`
	Type  ComponentType          `json:"type"`
	Props map[string]interface{} `json:"props"`
	Key   string                 `json:"key,omitempty"`
}

// jsNewTray
//
//	Example:
//	const tray = ctx.newTray()
func (t *TrayManager) jsNewTray(goja.FunctionCall) goja.Value {
	tray := &Tray{
		components:  result.NewResultMap[string, Component](),
		renderFunc:  nil,
		trayManager: t,
	}

	t.tray = mo.Some(tray)

	// Create a new tray object
	trayObj := t.ctx.vm.NewObject()
	_ = trayObj.Set("render", tray.jsRender)
	_ = trayObj.Set("flex", tray.jsFlex)
	_ = trayObj.Set("text", tray.jsText)
	_ = trayObj.Set("button", tray.jsButton)
	_ = trayObj.Set("mount", tray.jsMount)
	_ = trayObj.Set("update", tray.jsUpdate)

	return trayObj
}

/////

// jsFlex
//
//	Example:
//	const flex = tray.flex([
//		tray.button({ label: "A button", onClick: "my-action" }),
//		true ? tray.text("Some text") : null,
//	])
//	// or
//	const flex = tray.flex({
//		items: [
//			tray.button({ label: "A button", onClick: "my-action" }),
//			true ? tray.text("Some text") : null,
//		]
//	})
//	tray.render(() => flex)
func (t *Tray) jsFlex(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.trayManager.ctx.vm.NewTypeError("flex requires at least one argument"))
	}

	var items []interface{}

	// Handle object argument with items property
	propsValue := call.Argument(0)
	if propsValue == goja.Undefined() {
		propsValue = t.trayManager.ctx.vm.NewObject()
	}

	props := propsValue.ToObject(t.trayManager.ctx.vm)

	itemsVal := props.Get("items")
	if itemsVal == goja.Undefined() {
		panic(t.trayManager.ctx.vm.NewTypeError("flex requires an items array"))
	}

	items = itemsVal.Export().([]interface{})

	components := make([]Component, 0)
	for _, item := range items {
		if item == nil {
			continue
		}

		// Try to convert the item to a Component
		if comp, ok := item.(Component); ok {
			components = append(components, comp)
		}
	}

	component := Component{
		ID:   uuid.New().String(),
		Type: ComponentTypeFlex,
		Props: map[string]interface{}{
			"items": components,
		},
	}

	t.components.Set(component.ID, component)

	return t.trayManager.ctx.vm.ToValue(component)
}

// jsText
//
//	Example:
//	const text = tray.text("Some text")
func (t *Tray) jsText(call goja.FunctionCall) goja.Value {
	return t.trayManager.ctx.vm.ToValue(Component{
		ID:   uuid.New().String(),
		Type: ComponentTypeText,
		Props: map[string]interface{}{
			"text": call.Argument(0).String(),
		},
	})
}

// jsButton
//
//	Example:
//	const button = tray.button({ label: "A button", onClick: "my-action" })
func (t *Tray) jsButton(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.trayManager.ctx.vm.NewTypeError("button requires a label"))
	}

	propsValue := call.Argument(0)
	if propsValue == goja.Undefined() {
		propsValue = t.trayManager.ctx.vm.NewObject()
	}

	props := propsValue.ToObject(t.trayManager.ctx.vm)

	label := props.Get("label").String()
	onClick := props.Get("onClick")

	component := Component{
		ID:   uuid.New().String(),
		Type: ComponentTypeButton,
		Props: map[string]interface{}{
			"label":   label,
			"onClick": onClick,
		},
	}

	t.components.Set(component.ID, component)

	return t.trayManager.ctx.vm.ToValue(component)
}

// jsRender
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
	//t.trayManager.ctx.PrintState()
	return goja.Undefined()
}

// jsMount should be called once to mount the tray once the application loads
//
//	Example:
//	tray.mount()
func (t *Tray) jsMount(call goja.FunctionCall) goja.Value {

	return goja.Undefined()
}
