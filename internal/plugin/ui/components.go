package plugin_ui

import (
	"github.com/dop251/goja"
)

const (
	MAX_FIELD_REFS = 20
)

// ComponentManager is used to register components.
// Any higher-order UI system must use this to register components. (Tray)
type ComponentManager struct {
	ctx *Context

	fieldRefCount int
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
		{
			Name:             "items",
			Type:             "array",
			OptionalFirstArg: true,
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
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
		{
			Name:             "items",
			Type:             "array",
			OptionalFirstArg: true,
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
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
		{
			Name:             "items",
			Type:             "array",
			OptionalFirstArg: true,
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
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
		{
			Name:             "text",
			Type:             "string",
			Required:         true,
			OptionalFirstArg: true,
			Validate:         validateType("string"),
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
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
		{
			Name:             "label",
			Type:             "string",
			Required:         true,
			OptionalFirstArg: true,
			Validate:         validateType("string"),
		},
		{
			Name:     "onClick",
			Type:     "string",
			Required: false,
			Validate: validateType("string"),
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
	})
}

////////////////////////////////////////////
// Fields
////////////////////////////////////////////

// jsRegisterFieldRef allows to dynamically handle the value of a field outside the rendering context
//
//	Example:
//	const fieldRef = ctx.registerFieldRef("my-field")
//	fieldRef.setValue("Hello World!")
//	fieldRef.current // "Hello World!"
//
//	tray.render(() => tray.input({ fieldRef: "my-field" }))
func (c *ComponentManager) jsRegisterFieldRef(call goja.FunctionCall) goja.Value {
	fieldRefObj := c.ctx.vm.NewObject()

	if c.fieldRefCount >= MAX_FIELD_REFS {
		c.ctx.HandleTypeError("Too many field refs registered")
		return goja.Undefined()
	}

	c.fieldRefCount++

	fieldRefName, ok := call.Argument(0).Export().(string)
	if !ok {
		c.ctx.HandleTypeError("registerFieldRef requires a field name")
	}

	fieldRefObj.Set("setValue", func(call goja.FunctionCall) goja.Value {
		value := call.Argument(0).Export()
		if value == nil {
			c.ctx.HandleTypeError("setValue requires a value")
		}

		c.ctx.SendEventToClient(ServerFieldRefSetValueEvent, ServerFieldRefSetValueEventPayload{
			FieldRef: fieldRefName,
			Value:    value,
		})

		fieldRefObj.Set("current", value)

		return goja.Undefined()
	})

	fieldRefObj.Set("current", goja.Undefined())

	eventListener := c.ctx.RegisterEventListener(ClientFieldRefSendValueEvent)
	payload := ClientFieldRefSendValueEventPayload{}

	go func() {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientFieldRefSendValueEvent, &payload) {
				if payload.Value != nil {
					c.ctx.scheduler.Schedule(func() error {
						fieldRefObj.Set("current", payload.Value)
						return nil
					})
				}
			}
		}
	}()

	return fieldRefObj
}

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
		{
			Name:             "placeholder",
			Type:             "string",
			Required:         false,
			OptionalFirstArg: true,
			Validate:         validateType("string"),
		},
		{
			Name:     "value",
			Type:     "string",
			Required: false,
			Default:  "",
			Validate: validateType("string"),
		},
		{
			Name:     "onChange",
			Type:     "string",
			Required: false,
			Validate: validateType("string"),
		},
		{
			Name:     "fieldRef",
			Type:     "string",
			Required: false,
			Validate: validateType("string"),
		},
		{
			Name:     "style",
			Type:     "object",
			Required: false,
			Validate: validateType("object"),
		},
	})
}
