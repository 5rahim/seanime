package plugin_ui

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
)

const (
	MAX_FIELD_REFS = 100
)

// ComponentManager is used to register components.
// Any higher-order UI system must use this to register components. (Tray)
type ComponentManager struct {
	ctx *Context

	// Last rendered components
	lastRenderedComponents interface{}
}

// jsDiv
//
//	Example:
//	const div = tray.div({
//		items: [
//			tray.text("Some text"),
//		]
//	})
func (c *ComponentManager) jsDiv(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "div", []ComponentProp{
		{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	})
}

// jsFlex
//
//	Example:
//	const flex = tray.flex({
//		items: [
//			tray.button({ label: "A button", onClick: "my-action" }),
//			true ? tray.text("Some text") : null,
//		]
//	})
//	tray.render(() => flex)
func (c *ComponentManager) jsFlex(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "flex", []ComponentProp{
		{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "gap", Type: "number", Required: false, Default: 2, Validate: validateType("number")},
		{Name: "direction", Type: "string", Required: false, Default: "row", Validate: validateType("string")},
	})
}

// jsStack
//
//	Example:
//	const stack = tray.stack({
//		items: [
//			tray.text("Some text"),
//		]
//	})
func (c *ComponentManager) jsStack(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "stack", []ComponentProp{
		{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "gap", Type: "number", Required: false, Default: 2, Validate: validateType("number")},
	})
}

// jsText
//
//	Example:
//	const text = tray.text("Some text")
//	// or
//	const text = tray.text({ text: "Some text" })
func (c *ComponentManager) jsText(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "text", []ComponentProp{
		{Name: "text", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	})
}

// jsButton
//
//	Example:
//	const button = tray.button("Click me")
//	// or
//	const button = tray.button({ label: "Click me", onClick: "my-action" })
func (c *ComponentManager) jsButton(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "button", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "intent", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "loading", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	})
}

////////////////////////////////////////////
// Fields
////////////////////////////////////////////

// jsInput
//
//	Example:
//	const input = tray.input("Enter your name") // placeholder as shorthand
//	// or
//	const input = tray.input({
//		placeholder: "Enter your name",
//		value: "John",
//		onChange: "input-changed"
//	})
func (c *ComponentManager) jsInput(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "input", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "placeholder", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "value", Type: "string", Required: false, Default: "", Validate: validateType("string")},
		{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	})
}

func validateOptions(v interface{}) error {
	if v == nil {
		return errors.New("options must be an array of objects")
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(v.([]byte), &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		return nil
	}
	for _, option := range arr {
		if _, ok := option["label"]; !ok {
			return errors.New("options must be an array of objects with a label property")
		}
		if _, ok := option["value"]; !ok {
			return errors.New("options must be an array of objects with a value property")
		}
	}
	return nil
}

// jsSelect
//
//	Example:
//	const select = tray.select("Select an item", {
//		options: [{ label: "Item 1", value: "item1" }, { label: "Item 2", value: "item2" }],
//		onChange: "select-changed"
//	})
//	// or
//	const select = tray.select({
//		placeholder: "Select an item",
//		options: [{ label: "Item 1", value: "item1" }, { label: "Item 2", value: "item2" }],
//		value: "Item 1",
//		onChange: "select-changed"
//	})
func (c *ComponentManager) jsSelect(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "select", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "placeholder", Type: "string", Required: false, Validate: validateType("string")},
		{
			Name:     "options",
			Type:     "array",
			Required: true,
			Validate: validateOptions,
		},
		{Name: "value", Type: "string", Required: false, Default: "", Validate: validateType("string")},
		{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	})
}

// jsCheckbox
//
//	Example:
//	const checkbox = tray.checkbox("I agree to the terms and conditions")
//	// or
//	const checkbox = tray.checkbox({ label: "I agree to the terms and conditions", value: true })
func (c *ComponentManager) jsCheckbox(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "checkbox", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "value", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	})
}

// jsRadioGroup
//
//	Example:
//	const radioGroup = tray.radioGroup({
//		options: [{ label: "Item 1", value: "item1" }, { label: "Item 2", value: "item2" }],
//		onChange: "radio-group-changed"
//	})
func (c *ComponentManager) jsRadioGroup(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "radioGroup", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "value", Type: "string", Required: false, Default: "", Validate: validateType("string")},
		{
			Name:     "options",
			Type:     "array",
			Required: true,
			Validate: validateOptions,
		},
		{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	})
}

// jsSwitch
//
//	Example:
//	const switch = tray.switch({
//		label: "Toggle me",
//		value: true
//	})
func (c *ComponentManager) jsSwitch(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "switch", []ComponentProp{
		{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
		{Name: "value", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
		{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
		{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
		{Name: "side", Type: "string", Required: false, Validate: validateType("string")},
	})
}
