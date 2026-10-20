package plugin_ui

import "github.com/dop251/goja"

type ModalOptions struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	ClassName   string `json:"className,omitempty"`
}

type ModalManager struct {
	ctx              *Context
	componentManager *ComponentManager
	options          ModalOptions
	renderFunc       func(goja.FunctionCall) goja.Value
	open             bool
}

func NewModalManager(ctx *Context) *ModalManager {
	m := &ModalManager{ctx: ctx, componentManager: &ComponentManager{ctx: ctx}}
	listener := ctx.RegisterEventListener(ClientModalClosedEvent)
	listener.SetCallback(func(event *ClientPluginEvent) {
		ctx.scheduler.ScheduleAsync(func() error {
			m.open = false
			return nil
		})
	})
	return m
}

func (m *ModalManager) jsNewModal(options ModalOptions) goja.Value {
	m.options = options
	obj := m.ctx.vm.NewObject()
	_ = obj.Set("render", func(fn func(goja.FunctionCall) goja.Value) {
		m.renderFunc = fn
		m.renderScheduled()
	})
	_ = obj.Set("open", func() {
		if m.renderFunc == nil {
			m.ctx.handleTypeError("modal render function is not set")
		}
		m.open = true
		if err := m.render("open"); err != nil {
			m.ctx.handleException(err)
		}
	})
	_ = obj.Set("close", func() {
		m.open = false
		m.ctx.SendEventToClient(ServerModalUpdatedEvent, ServerModalUpdatedEventPayload{Action: "close"})
	})
	_ = obj.Set("update", m.renderScheduled)
	c := m.componentManager
	_ = obj.Set("div", c.jsDiv)
	_ = obj.Set("flex", c.jsFlex)
	_ = obj.Set("stack", c.jsStack)
	_ = obj.Set("text", c.jsText)
	_ = obj.Set("button", c.jsButton)
	_ = obj.Set("anchor", c.jsAnchor)
	_ = obj.Set("input", c.jsInput)
	_ = obj.Set("radioGroup", c.jsRadioGroup)
	_ = obj.Set("switch", c.jsSwitch)
	_ = obj.Set("checkbox", c.jsCheckbox)
	_ = obj.Set("select", c.jsSelect)
	_ = obj.Set("css", c.jsCSS)
	_ = obj.Set("tooltip", c.jsTooltip)
	_ = obj.Set("modal", c.jsModal)
	_ = obj.Set("dropdownMenu", c.jsDropdownMenu)
	_ = obj.Set("dropdownMenuItem", c.jsDropdownMenuItem)
	_ = obj.Set("dropdownMenuSeparator", c.jsDropdownMenuSeparator)
	_ = obj.Set("dropdownMenuLabel", c.jsDropdownMenuLabel)
	_ = obj.Set("popover", c.jsPopover)
	_ = obj.Set("a", c.jsA)
	_ = obj.Set("p", c.jsP)
	_ = obj.Set("alert", c.jsAlert)
	_ = obj.Set("tabs", c.jsTabs)
	_ = obj.Set("tabsList", c.jsTabsList)
	_ = obj.Set("tabsTrigger", c.jsTabsTrigger)
	_ = obj.Set("tabsContent", c.jsTabsContent)
	_ = obj.Set("badge", c.jsBadge)
	_ = obj.Set("span", c.jsSpan)
	_ = obj.Set("img", c.jsImg)
	return obj
}

func (m *ModalManager) renderScheduled() {
	m.ctx.scheduler.ScheduleAsync(func() error {
		if !m.open {
			return nil
		}
		return m.render("update")
	})
}

func (m *ModalManager) render(action string) error {
	components, err := m.componentManager.renderComponents(m.renderFunc)
	if err != nil {
		return err
	}
	m.ctx.SendEventToClient(ServerModalUpdatedEvent, ServerModalUpdatedEventPayload{
		Action:      action,
		Title:       m.options.Title,
		Description: m.options.Description,
		ClassName:   m.options.ClassName,
		Components:  components,
	})
	return nil
}
