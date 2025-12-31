package plugin_ui

import (
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type WebviewSlot string

const (
	FixedSlot WebviewSlot = "fixed"
)

var WebviewSlots = []WebviewSlot{FixedSlot}

type WebviewManager struct {
	ctx         *Context
	webviews    *result.Map[WebviewSlot, *Webview]
	updateMutex sync.Mutex

	componentManager *ComponentManager
}

type WebviewChannel struct {
	webview      *Webview
	syncedStates *result.Map[string, string] // map[key]stateID
}

func NewWebviewManager(ctx *Context) *WebviewManager {
	return &WebviewManager{
		ctx:              ctx,
		webviews:         result.NewMap[WebviewSlot, *Webview](),
		componentManager: &ComponentManager{ctx: ctx},
	}
}

// renderWebviewScheduled renders the new component tree of the webview at the given slot.
// This function is unsafe because it is not thread-safe and should be scheduled.
func (t *WebviewManager) renderWebviewScheduled(slots ...WebviewSlot) {
	t.updateMutex.Lock()
	defer t.updateMutex.Unlock()

	shouldMount := false

	// renderWebviewScheduled can be called without slots (when states are updated)
	if len(slots) == 0 {
		slots = WebviewSlots
	} else {
		// Set the webview as mounted if renderWebviewScheduled has been called WITH a slot
		shouldMount = true
	}

	for _, slot := range slots {
		webview, ok := t.webviews.Get(slot)
		if !ok {
			return
		}

		if webview.renderFunc == nil {
			return
		}

		// Make sure it's mounted
		if shouldMount && !webview.mounted.Load() {
			webview.mounted.Store(true)
		}

		// Ignore if it's not mounted
		// renderWebviewScheduled can be called without slots, in this case it will render already mounted webviews
		if !webview.mounted.Load() {
			return
		}

		webview.lastUpdatedAt = time.Now()

		t.ctx.scheduler.ScheduleAsync(func() error {
			newComponents, err := t.componentManager.renderComponents(webview.renderFunc)
			if err != nil {
				t.ctx.logger.Error().Err(err).Msg("plugin: Failed to render webview")
				t.ctx.handleException(err)
				return nil
			}

			// t.ctx.logger.Trace().Msg("plugin: Sending webview update to client")
			// Send the JSON value to the client
			t.ctx.SendEventToClient(ServerWebviewUpdatedEvent, ServerWebviewUpdatedEventPayload{
				Slot:       string(slot),
				Components: newComponents,
			})
			return nil
		})
	}
}

// renderWebviewIframeScheduled
func (t *WebviewManager) renderWebviewIframeScheduled(slot WebviewSlot) {
	t.updateMutex.Lock()
	defer t.updateMutex.Unlock()

	webview, ok := t.webviews.Get(slot)
	if !ok {
		return
	}

	if webview.contentFunc == nil {
		return
	}

	webview.lastUpdatedAt = time.Now()

	t.ctx.scheduler.ScheduleAsync(func() error {
		str := t.componentManager.executeContentFunc(webview.contentFunc)

		t.ctx.SendEventToClient(ServerWebviewIframeEvent, ServerWebviewIframeEventPayload{
			Slot:    string(slot),
			Content: str,
			ID:      webview.id,
		})
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Webview
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Webview struct {
	webviewManager *WebviewManager
	id             string
	channel        *WebviewChannel

	// Server-side rendering
	renderFunc func(goja.FunctionCall) goja.Value
	mounted    atomic.Bool

	// HTML content rendering (iframe)
	contentFunc func(goja.FunctionCall) goja.Value

	lastUpdatedAt time.Time // for renderFunc

	Slot WebviewSlot
}

// jsNewWebview
//
//	Example:
//	const webview = ctx.newWebview({ slot: "" })
func (t *WebviewManager) jsNewWebview(call goja.FunctionCall) goja.Value {
	webview := &Webview{
		webviewManager: t,
		id:             uuid.NewString(),
		renderFunc:     nil,
		contentFunc:    nil,
		lastUpdatedAt:  time.Time{},
		Slot:           "",
		channel: &WebviewChannel{
			syncedStates: result.NewMap[string, string](),
		},
	}

	// Set the webview reference in the channel
	webview.channel.webview = webview

	props := call.Arguments
	if len(props) > 0 {
		propsObj := props[0].Export().(map[string]interface{})
		if propsObj["slot"] != nil {
			s, _ := propsObj["slot"].(string)
			webview.Slot = WebviewSlot(s)
		}
	}

	if webview.Slot == "" {
		t.ctx.handleTypeError("newWebview requires a slot name")
	}

	t.webviews.Set(webview.Slot, webview)

	// Create a new webview object
	webviewObj := t.ctx.vm.NewObject()
	_ = webviewObj.Set("render", webview.jsRender)
	_ = webviewObj.Set("setContent", webview.jsSetContent)
	_ = webviewObj.Set("update", webview.jsUpdate)

	// Create a new webview object
	channelObj := t.ctx.vm.NewObject()
	_ = channelObj.Set("sync", webview.channel.jsSync)
	_ = channelObj.Set("on", webview.channel.jsOn)
	_ = channelObj.Set("send", webview.channel.jsSend)

	_ = webviewObj.Set("channel", channelObj)

	// Register components
	_ = webviewObj.Set("div", t.componentManager.jsDiv)
	_ = webviewObj.Set("flex", t.componentManager.jsFlex)
	_ = webviewObj.Set("stack", t.componentManager.jsStack)
	_ = webviewObj.Set("text", t.componentManager.jsText)
	_ = webviewObj.Set("button", t.componentManager.jsButton)
	_ = webviewObj.Set("anchor", t.componentManager.jsAnchor)
	_ = webviewObj.Set("input", t.componentManager.jsInput)
	_ = webviewObj.Set("radioGroup", t.componentManager.jsRadioGroup)
	_ = webviewObj.Set("switch", t.componentManager.jsSwitch)
	_ = webviewObj.Set("checkbox", t.componentManager.jsCheckbox)
	_ = webviewObj.Set("select", t.componentManager.jsSelect)
	_ = webviewObj.Set("css", t.componentManager.jsCSS)
	_ = webviewObj.Set("tooltip", t.componentManager.jsTooltip)

	// Listen to mount events in order to return the webview object
	listener := t.ctx.RegisterEventListener(ClientWebviewMountedEvent)
	t.ctx.registerOnCleanup(func() {
		t.ctx.UnregisterEventListenerE(listener)
	})
	listener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientWebviewMountedEventPayload
		if event.ParsePayloadAs(ClientWebviewMountedEvent, &payload) && payload.Slot == string(webview.Slot) {
			t.ctx.scheduler.ScheduleAsync(func() error {
				// Return the webview object to the client
				t.renderWebviewIframeScheduled(webview.Slot)
				return nil
			})
		}
	})

	return webviewObj
}

/////

// jsSetContent registers the HTML content to be rendered in the iframe
// Communication with the Plugin UI context will be done using a bridge
//
//	Example:
//	webview.setContent(() => `<div>Hello World!</div>`)
func (w *Webview) jsSetContent(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		w.webviewManager.ctx.handleTypeError("render requires a function")
	}

	// Set the render function
	w.contentFunc = funcRes
	w.renderFunc = nil

	return goja.Undefined()
}

// jsRender registers a function to be called when the webview is rendered/updated
//
//	Example:
//	webview.render(() => webview.stack([]))
func (w *Webview) jsRender(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		w.webviewManager.ctx.handleTypeError("render requires a function")
	}

	// Set the render function
	w.renderFunc = funcRes
	w.contentFunc = nil

	return goja.Undefined()
}

// jsUpdate schedules a re-render on the client
//
//	Example:
//	webview.update()
func (w *Webview) jsUpdate(call goja.FunctionCall) goja.Value {
	// Update the context's lastUIUpdateAt to prevent duplicate updates
	w.webviewManager.ctx.uiUpdateMu.Lock()
	w.webviewManager.ctx.lastUIUpdateAt = time.Now()
	w.webviewManager.ctx.uiUpdateMu.Unlock()

	w.webviewManager.renderWebviewScheduled(w.Slot)
	return goja.Undefined()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// WebviewChannel
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// jsSync syncs a state with the webview iframe
// When the state updates, the new value is sent to the iframe
//
//	Example:
//	const count = ctx.state(0)
//	webview.channel.sync("count", count)
func (c *WebviewChannel) jsSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("sync requires a key and a state")
	}

	key, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("sync: first argument must be a string key")
	}

	stateObj := call.Argument(1).ToObject(c.webview.webviewManager.ctx.vm)
	if stateObj == nil {
		c.webview.webviewManager.ctx.handleTypeError("sync: second argument must be a state object")
	}

	stateIDVal := stateObj.Get("__stateId")
	if stateIDVal == nil {
		c.webview.webviewManager.ctx.handleTypeError("sync: state object must have an id")
	}

	stateID := stateIDVal.String()

	// Store the mapping
	c.syncedStates.Set(key, stateID)

	// Subscribe to state changes
	state, ok := c.webview.webviewManager.ctx.states.Get(stateID)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("sync: state not found")
	}

	// Send initial value
	c.sendStateToWebview(key, state.Value.Export())

	// Listen for state changes and sync to webview
	stateCh := c.webview.webviewManager.ctx.subscribeStateUpdates()

	// Start a goroutine to listen for this specific state's updates
	go func() {
		for newState := range stateCh {
			if newState.ID == stateID {
				c.sendStateToWebview(key, newState.Value.Export())
			}
		}
	}()

	return goja.Undefined()
}

// jsOn registers an event handler for messages from the webview
// This is called from the server-side to listen to events triggered by webview.trigger()
//
//	Example:
//	webview.channel.on("customEvent", (data) => {
//		console.log(data)
//	})
func (c *WebviewChannel) jsOn(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("on requires an event name and a callback")
	}

	eventName, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("on: first argument must be a string event name")
	}

	callback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("on: second argument must be a callback function")
	}

	// Register event handler to listen for messages from the webview
	eventListener := c.webview.webviewManager.ctx.RegisterEventListener(ClientEventHandlerTriggeredEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		// Parse the payload
		var payload ClientEventHandlerTriggeredEventPayload
		if event.ParsePayloadAs(ClientEventHandlerTriggeredEvent, &payload) && payload.HandlerName == eventName {
			c.webview.webviewManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), c.webview.webviewManager.ctx.vm.ToValue(payload.Event))
				if err != nil {
					c.webview.webviewManager.ctx.logger.Error().Err(err).Msgf("plugin: Error running webview channel.on callback for event %s", eventName)
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsSend sends a message to the webview iframe
// This is used to send arbitrary data from the server to the iframe
//
//	Example:
//	webview.channel.send("messageType", { data: "hello" })
func (c *WebviewChannel) jsSend(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("send requires a key and a value")
	}

	key, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("send: first argument must be a string key")
	}

	value := call.Argument(1).Export()

	c.sendStateToWebview(key, value)

	return goja.Undefined()
}

// sendStateToWebview sends a state value to the webview iframe
func (c *WebviewChannel) sendStateToWebview(key string, value interface{}) {
	webviewId := c.webview.GetID()

	// Get the token from the iframe (we'll need to update the iframe creation to store this)
	// For now, we'll send it without token verification on the receive side
	c.webview.webviewManager.ctx.SendEventToClient(ServerWebviewSyncStateEvent, ServerWebviewSyncStateEventPayload{
		WebviewID: webviewId,
		Key:       key,
		Value:     value,
		Token:     "", // Will be populated by the client-side handler
	})
}

func (w *Webview) GetID() string {
	return w.webviewManager.ctx.ext.ID + "-" + string(w.Slot)
}
