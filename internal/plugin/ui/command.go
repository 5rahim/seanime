package plugin_ui

import (
	goja_util "seanime/internal/util/goja"
	"seanime/internal/util/result"
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

	placeholder  string
	shouldFilter bool
	filterType   string

	// registered is true if the command palette has been registered
	registered bool

	items         *result.Map[string, *commandItem]
	renderedItems []*CommandItemJSON // Store rendered items when setItems is called
}

type (
	commandItem struct {
		id           string
		label        string
		value        string
		renderFunc   func(goja.FunctionCall) goja.Value
		onSelectFunc func(goja.FunctionCall) goja.Value
	}

	// CommandItemJSON is the JSON representation of a command item.
	// It is used to send the command item to the client.
	CommandItemJSON struct {
		ID         string      `json:"id"`
		Label      string      `json:"label"`
		Value      string      `json:"value"`
		Components interface{} `json:"components"`
	}
)

func NewCommandPaletteManager(ctx *Context) *CommandPaletteManager {
	return &CommandPaletteManager{
		ctx:              ctx,
		componentManager: &ComponentManager{ctx: ctx},
		items:            result.NewResultMap[string, *commandItem](),
	}
}

type NewCommandPaletteOptions struct {
	Placeholder  string `json:"placeholder,omitempty"`
	ShouldFilter bool   `json:"shouldFilter,omitempty"`
	FilterType   string `json:"filterType,omitempty"`
}

func (c *CommandPaletteManager) jsNewCommandPalette(options NewCommandPaletteOptions) goja.Value {
	c.registered = true

	cmdObj := c.ctx.vm.NewObject()

	_ = cmdObj.Set("setItems", func(items []interface{}) {
		c.items.Clear()
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			id := uuid.New().String()
			label, _ := itemMap["label"].(string)
			value, ok := itemMap["value"].(string)
			if !ok {
				c.ctx.HandleTypeError("value must be a string")
				return
			}
			renderFunc, ok := itemMap["render"].(func(goja.FunctionCall) goja.Value)
			if len(label) == 0 && !ok {
				c.ctx.HandleTypeError("label or render function must be provided")
				return
			}
			onSelectFunc, ok := itemMap["onSelect"].(func(goja.FunctionCall) goja.Value)
			if !ok {
				c.ctx.HandleTypeError("onSelect must be a function")
				return
			}

			c.items.Set(id, &commandItem{
				id:           id,
				label:        label,
				value:        value,
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

	_ = cmdObj.Set("setShouldFilter", func(shouldFilter bool) {
		c.shouldFilter = shouldFilter
		c.renderCommandPaletteScheduled()
	})

	_ = cmdObj.Set("setFilterType", func(filterType string) {
		c.filterType = filterType
		c.renderCommandPaletteScheduled()
	})

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

	// t.ctx.logger.Trace().Msg("plugin: Rendering tray")
	components, err := componentManager.renderComponents(c.renderFunc)
	if err != nil {
		ctx.logger.Error().Err(err).Msg("plugin: Failed to render command palette item")
		ctx.HandleException(err)
		return nil
	}

	// Reset the last rendered components, we don't care about diffing
	componentManager.lastRenderedComponents = nil

	return &CommandItemJSON{
		ID:         c.id,
		Label:      c.label,
		Value:      c.value,
		Components: components,
	}
}

func (c *CommandPaletteManager) renderCommandPaletteScheduled() {
	c.updateMutex.Lock()
	defer c.updateMutex.Unlock()

	if !c.registered {
		return
	}

	c.ctx.SendEventToClient(ServerCommandPaletteUpdatedEvent, ServerCommandPaletteUpdatedEventPayload{
		Placeholder:  c.placeholder,
		ShouldFilter: c.shouldFilter,
		FilterType:   c.filterType,
		Items:        c.renderedItems,
	})
}
