package plugin_ui

import (
	"fmt"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type Tray struct {
	vm         *goja.Runtime
	context    *Context
	components *result.Map[string, Component]
	renderFunc func(goja.FunctionCall) goja.Value
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

func (c *Context) jsNewTray(goja.FunctionCall) goja.Value {
	tray := &Tray{
		context:    c,
		components: result.NewResultMap[string, Component](),
	}

	// Create a new tray object
	trayObj := c.vm.NewObject()
	trayObj.Set("render", tray.jsRender)
	trayObj.Set("flex", tray.jsFlex)
	trayObj.Set("text", tray.jsText)
	trayObj.Set("button", tray.jsButton)
	trayObj.Set("mount", tray.jsMount)
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
		panic(t.context.vm.NewTypeError("flex requires at least one argument"))
	}

	var items []interface{}

	// Handle object argument with items property
	propsValue := call.Argument(0)
	if propsValue == goja.Undefined() {
		propsValue = t.context.vm.NewObject()
	}

	props := propsValue.ToObject(t.context.vm)

	itemsVal := props.Get("items")
	if itemsVal == goja.Undefined() {
		panic(t.context.vm.NewTypeError("flex requires an items array"))
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

	return t.context.vm.ToValue(component)
}

// jsText
//
//	Example:
//	const text = tray.text("Some text")
func (t *Tray) jsText(call goja.FunctionCall) goja.Value {
	return t.context.vm.ToValue(Component{
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
		panic(t.context.vm.NewTypeError("button requires a label"))
	}

	propsValue := call.Argument(0)
	if propsValue == goja.Undefined() {
		propsValue = t.context.vm.NewObject()
	}

	props := propsValue.ToObject(t.context.vm)

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

	return t.context.vm.ToValue(component)
}

// jsRender
//
//	Example:
//	tray.render(() => flex)
func (t *Tray) jsRender(call goja.FunctionCall) goja.Value {

	fmt.Println(call.Argument(0).ExportType())

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		panic(t.context.vm.NewTypeError("render requires a function"))
	}

	// Set the render function
	t.renderFunc = funcRes

	return goja.Undefined()
}

// jsMount is a test function to see if the render function is working
//
//	Example:
//	tray.mount()
func (t *Tray) jsMount(call goja.FunctionCall) goja.Value {
	// Get json from calling renderFunc
	value := t.renderFunc(call)
	structValue := value.Export()

	util.Spew(structValue)

	t.context.PrintState()

	return goja.Undefined()
}
