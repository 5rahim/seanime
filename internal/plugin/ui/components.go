package plugin_ui

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
)

const (
	MaxFieldRefs = 100
)

// ComponentManager is used to register components.
// Any higher-order UI system must use this to register components. (Tray)
type ComponentManager struct {
	ctx *Context

	// Last rendered components
	lastRenderedComponents interface{}
}

func getComponentPropNames(componentProps []ComponentProp) []string {
	names := make([]string, len(componentProps))
	for i, prop := range componentProps {
		names[i] = prop.Name
	}
	return names
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var divComponentProps = []ComponentProp{
	{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
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
	return defineComponent(c.ctx.vm, call, "div", divComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var flexComponentProps = []ComponentProp{
	{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "gap", Type: "number", Required: false, Default: 2, Validate: validateType("number")},
	{Name: "direction", Type: "string", Required: false, Default: "row", Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
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
	return defineComponent(c.ctx.vm, call, "flex", flexComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var stackComponentProps = []ComponentProp{
	{Name: "items", Type: "array", Required: false, OptionalFirstArg: true},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "gap", Type: "number", Required: false, Default: 2, Validate: validateType("number")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
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
	return defineComponent(c.ctx.vm, call, "stack", stackComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var textComponentProps = []ComponentProp{
	{Name: "text", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsText
//
//	Example:
//	const text = tray.text("Some text")
//	// or
//	const text = tray.text({ text: "Some text" })
func (c *ComponentManager) jsText(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "text", textComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var buttonComponentProps = []ComponentProp{
	{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "intent", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "loading", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsButton
//
//	Example:
//	const button = tray.button("Click me")
//	// or
//	const button = tray.button({ label: "Click me", onClick: "my-action" })
func (c *ComponentManager) jsButton(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "button", buttonComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var anchorComponentProps = []ComponentProp{
	{Name: "text", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "href", Type: "string", Required: true, Validate: validateType("string")},
	{Name: "target", Type: "string", Required: false, Default: "_blank", Validate: validateType("string")},
	{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsAnchor
//
//	Example:
//	const anchor = tray.anchor("Click here", { href: "https://example.com" })
//	// or
//	const anchor = tray.anchor({ text: "Click here", href: "https://example.com" })
func (c *ComponentManager) jsAnchor(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "anchor", anchorComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fields
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var inputComponentProps = []ComponentProp{
	{Name: "label", Type: "string", Required: false, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "placeholder", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "value", Type: "string", Required: false, Default: "", Validate: validateType("string")},
	{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "onSelect", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "textarea", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
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
	return defineComponent(c.ctx.vm, call, "input", inputComponentProps)
}

func validateOptions(v interface{}) error {
	if v == nil {
		return errors.New("options must be an array of objects")
	}
	marshaled, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(marshaled, &arr); err != nil {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var selectComponentProps = []ComponentProp{
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
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
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
	return defineComponent(c.ctx.vm, call, "select", selectComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var checkboxComponentProps = []ComponentProp{
	{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "value", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsCheckbox
//
//	Example:
//	const checkbox = tray.checkbox("I agree to the terms and conditions")
//	// or
//	const checkbox = tray.checkbox({ label: "I agree to the terms and conditions", value: true })
func (c *ComponentManager) jsCheckbox(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "checkbox", checkboxComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var radioGroupComponentProps = []ComponentProp{
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
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsRadioGroup
//
//	Example:
//	const radioGroup = tray.radioGroup({
//		options: [{ label: "Item 1", value: "item1" }, { label: "Item 2", value: "item2" }],
//		onChange: "radio-group-changed"
//	})
func (c *ComponentManager) jsRadioGroup(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "radio-group", radioGroupComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var switchComponentProps = []ComponentProp{
	{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "value", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "onChange", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "fieldRef", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "size", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "side", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsSwitch
//
//	Example:
//	const switch = tray.switch({
//		label: "Toggle me",
//		value: true
//	})
func (c *ComponentManager) jsSwitch(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "switch", switchComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var cssComponentProps = []ComponentProp{
	{Name: "css", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
}

// jsCSS
//
//	Example:
//	const switch = tray.css({
//		css: `
//			.my-class {
//				color: red;
//			}
//		`
//	})
func (c *ComponentManager) jsCSS(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "css", cssComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var tooltipComponentProps = []ComponentProp{
	{Name: "text", Type: "string", Required: true, Validate: validateType("string")},
	{Name: "item", Type: "any", Required: true, OptionalFirstArg: true},
	{Name: "side", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "sideOffset", Type: "number", Required: false, Validate: validateType("number")},
}

// jsTooltip
//
//	Example:
//	const switch = tray.tooltip([tray.button("Click me")], { text: "This is a tooltip" })
func (c *ComponentManager) jsTooltip(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "tooltip", tooltipComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var modalComponentProps = []ComponentProp{
	{Name: "trigger", Type: "any", Required: true},
	{Name: "title", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "description", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "items", Type: "array", Required: false},
	{Name: "footer", Type: "array", Required: false},
	{Name: "open", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "onOpenChange", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsModal
//
//	Example:
//	const modal = tray.modal({
//		trigger: tray.button("Open Modal"),
//		title: "Modal Title",
//		items: [tray.text("Modal content")],
//	})
func (c *ComponentManager) jsModal(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "modal", modalComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var dropdownMenuComponentProps = []ComponentProp{
	{Name: "trigger", Type: "any", Required: true},
	{Name: "items", Type: "array", Required: true},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsDropdownMenu
//
//	Example:
//	const menu = tray.dropdownMenu({
//		trigger: tray.button("Open Menu"),
//		items: [
//			tray.dropdownMenuItem({ label: "Item 1", onClick: "item-1" }),
//			tray.dropdownMenuSeparator(),
//			tray.dropdownMenuItem({ label: "Item 2", onClick: "item-2" }),
//		]
//	})
func (c *ComponentManager) jsDropdownMenu(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "dropdown-menu", dropdownMenuComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var dropdownMenuItemComponentProps = []ComponentProp{
	{Name: "item", Type: "any", Required: true, OptionalFirstArg: true},
	{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "disabled", Type: "boolean", Required: false, Default: false, Validate: validateType("boolean")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsDropdownMenuItem
//
//	Example:
//	const item = tray.dropdownMenuItem({ item: tray.span("Item 1"), onClick: "item-1" })
func (c *ComponentManager) jsDropdownMenuItem(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "dropdown-menu-item", dropdownMenuItemComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var dropdownMenuSeparatorComponentProps []ComponentProp

// jsDropdownMenuSeparator
//
//	Example:
//	const separator = tray.dropdownMenuSeparator()
func (c *ComponentManager) jsDropdownMenuSeparator(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "dropdown-menu-separator", dropdownMenuSeparatorComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var dropdownMenuLabelComponentProps = []ComponentProp{
	{Name: "label", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsDropdownMenuLabel
//
//	Example:
//	const label = tray.dropdownMenuLabel("Section Title")
func (c *ComponentManager) jsDropdownMenuLabel(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "dropdown-menu-label", dropdownMenuLabelComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var popoverComponentProps = []ComponentProp{
	{Name: "trigger", Type: "any", Required: true},
	{Name: "items", Type: "array", Required: true},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsPopover
//
//	Example:
//	const popover = tray.popover({
//		trigger: tray.button("Open Popover"),
//		items: [tray.text("Popover content")],
//	})
func (c *ComponentManager) jsPopover(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "popover", popoverComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var aComponentProps = []ComponentProp{
	{Name: "href", Type: "string", Required: true, Validate: validateType("string")},
	{Name: "items", Type: "array", Required: true, OptionalFirstArg: true},
	{Name: "target", Type: "string", Required: false, Default: "_blank", Validate: validateType("string")},
	{Name: "onClick", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsA
//
//	Example:
//	const link = tray.a({ href: "https://example.com", items: [tray.text("Click here")] })
func (c *ComponentManager) jsA(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "a", aComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var pComponentProps = []ComponentProp{
	{Name: "items", Type: "array", Required: true, OptionalFirstArg: true},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsP
//
//	Example:
//	const paragraph = tray.p({ items: [tray.text("Some text")] })
func (c *ComponentManager) jsP(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "p", pComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var alertComponentProps = []ComponentProp{
	{Name: "title", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "description", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "intent", Type: "string", Required: false, Default: "info", Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsAlert
//
//	Example:
//	const alert = tray.alert({ title: "Alert", description: "This is an alert", intent: "info" })
func (c *ComponentManager) jsAlert(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "alert", alertComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var tabsComponentProps = []ComponentProp{
	{Name: "defaultValue", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "items", Type: "array", Required: true, OptionalFirstArg: true},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsTabs
//
//	Example:
//	const tabs = tray.tabs({
//		items: [
//			tray.tabsList
//		]
//	})
func (c *ComponentManager) jsTabs(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "tabs", tabsComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var tabsTriggerComponentProps = []ComponentProp{
	{Name: "item", Type: "any", Required: true, OptionalFirstArg: true},
	{Name: "value", Type: "string", Required: true, Validate: validateType("string")},
}

// jsTabsTrigger
//
//	Example:
//	const trigger = tray.tabsTrigger({ value: "tab1", item: "Tab 1" })
func (c *ComponentManager) jsTabsTrigger(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "tabs-trigger", tabsTriggerComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var tabsContentComponentProps = []ComponentProp{
	{Name: "value", Type: "string", Required: true, Validate: validateType("string")},
	{Name: "items", Type: "array", Required: true, OptionalFirstArg: true},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsTabsContent
//
//	Example:
//	const content = tray.tabsContent({ value: "tab1", items: [tray.text("Content")] })
func (c *ComponentManager) jsTabsContent(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "tabs-content", tabsContentComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var badgeComponentProps = []ComponentProp{
	{Name: "text", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "intent", Type: "string", Required: false, Default: "gray", Validate: validateType("string")},
	{Name: "size", Type: "string", Required: false, Default: "md", Validate: validateType("string")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsBadge
//
//	Example:
//	const badge = tray.badge("New")
//	// or
//	const badge = tray.badge({ text: "New", intent: "success" })
func (c *ComponentManager) jsBadge(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "badge", badgeComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var tabsListComponentProps = []ComponentProp{
	{Name: "items", Type: "array", Required: true, OptionalFirstArg: true},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsTabsList
//
//	Example:
//	const list = tray.tabsList({ items: [tray.tabsTrigger({ value: "tab1", label: "Tab 1" })] })
func (c *ComponentManager) jsTabsList(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "tabs-list", tabsListComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var spanComponentProps = []ComponentProp{
	{Name: "text", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "items", Type: "array", Required: false},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsSpan
//
//	Example:
//	const span = tray.span({ items: [tray.text("Some text")] })
func (c *ComponentManager) jsSpan(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "span", spanComponentProps)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var imgComponentProps = []ComponentProp{
	{Name: "src", Type: "string", Required: true, OptionalFirstArg: true, Validate: validateType("string")},
	{Name: "alt", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "width", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "height", Type: "string", Required: false, Validate: validateType("string")},
	{Name: "style", Type: "object", Required: false, Validate: validateType("object")},
	{Name: "className", Type: "string", Required: false, Validate: validateType("string")},
}

// jsImg
//
//	Example:
//	const img = tray.img("https://example.com/image.png")
//	// or
//	const img = tray.img({ src: "https://example.com/image.png", alt: "Description" })
func (c *ComponentManager) jsImg(call goja.FunctionCall) goja.Value {
	return defineComponent(c.ctx.vm, call, "img", imgComponentProps)
}
