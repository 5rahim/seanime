package plugin_ui

import (
	"context"
	"fmt"
	"reflect"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/plugin"
	goja_util "seanime/internal/util/goja"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Context manages the entire plugin UI during its lifecycle
type Context struct {
	ui *UI

	ext            *extension.Extension
	logger         *zerolog.Logger
	wsEventManager events.WSEventManagerInterface

	mu       sync.RWMutex
	fetchSem chan struct{} // Semaphore for concurrent fetch requests

	vm               *goja.Runtime
	states           *result.Map[string, *State]
	stateSubscribers []chan *State
	scheduler        *goja_util.Scheduler // Schedule VM executions concurrently and execute them in order.
	wsSubscriber     *events.ClientEventSubscriber
	eventListeners   *result.Map[string, *EventListener] // Event listeners registered by plugin functions
	contextObj       *goja.Object

	fieldRefCount  int                    // Number of field refs registered
	exceptionCount int                    // Number of exceptions that have occurred
	effectStack    map[string]bool        // Track currently executing effects to prevent infinite loops
	effectCalls    map[string][]time.Time // Track effect calls within time window

	// State update batching
	updateBatchMu       sync.Mutex
	pendingStateUpdates map[string]struct{} // Set of state IDs with pending updates
	updateBatchTimer    *time.Timer         // Timer for flushing batched updates

	// UI update rate limiting
	lastUIUpdateAt time.Time
	uiUpdateMu     sync.Mutex

	webviewManager        *WebviewManager        // UNUSED
	screenManager         *ScreenManager         // Listen for screen events, send screen actions
	trayManager           *TrayManager           // Register and manage tray
	actionManager         *ActionManager         // Register and manage actions
	formManager           *FormManager           // Register and manage forms
	toastManager          *ToastManager          // Register and manage toasts
	commandPaletteManager *CommandPaletteManager // Register and manage command palette

	atomicCleanupCounter atomic.Int64
	onCleanupFns         *result.Map[int64, func()]
}

type State struct {
	ID    string
	Value goja.Value
}

// EventListener is used by Goja methods to listen for events from the client
type EventListener struct {
	ID       string
	ListenTo []ClientEventType       // Optional event type to listen for
	Channel  chan *ClientPluginEvent // Channel for the event payload
	closed   bool
}

func NewContext(ui *UI) *Context {
	ret := &Context{
		ui:                   ui,
		ext:                  ui.ext,
		logger:               ui.logger,
		vm:                   ui.vm,
		states:               result.NewResultMap[string, *State](),
		fetchSem:             make(chan struct{}, MaxConcurrentFetchRequests),
		stateSubscribers:     make([]chan *State, 0),
		eventListeners:       result.NewResultMap[string, *EventListener](),
		wsEventManager:       ui.wsEventManager,
		effectStack:          make(map[string]bool),
		effectCalls:          make(map[string][]time.Time),
		pendingStateUpdates:  make(map[string]struct{}),
		lastUIUpdateAt:       time.Now().Add(-time.Hour), // Initialize to a time in the past
		atomicCleanupCounter: atomic.Int64{},
		onCleanupFns:         result.NewResultMap[int64, func()](),
	}

	ret.scheduler = ui.scheduler
	ret.updateBatchTimer = time.AfterFunc(time.Duration(StateUpdateBatchInterval)*time.Millisecond, ret.flushStateUpdates)
	ret.updateBatchTimer.Stop() // Start in stopped state

	ret.trayManager = NewTrayManager(ret)
	ret.actionManager = NewActionManager(ret)
	ret.webviewManager = NewWebviewManager(ret)
	ret.screenManager = NewScreenManager(ret)
	ret.formManager = NewFormManager(ret)
	ret.toastManager = NewToastManager(ret)
	ret.commandPaletteManager = NewCommandPaletteManager(ret)

	return ret
}

func (c *Context) createAndBindContextObject(vm *goja.Runtime) {
	obj := vm.NewObject()

	_ = obj.Set("newTray", c.trayManager.jsNewTray)
	_ = obj.Set("newForm", c.formManager.jsNewForm)

	_ = obj.Set("newCommandPalette", c.commandPaletteManager.jsNewCommandPalette)

	_ = obj.Set("state", c.jsState)
	_ = obj.Set("setTimeout", c.jsSetTimeout)
	_ = obj.Set("setInterval", c.jsSetInterval)
	_ = obj.Set("effect", c.jsEffect)
	_ = obj.Set("registerEventHandler", c.jsRegisterEventHandler)
	_ = obj.Set("registerFieldRef", c.jsRegisterFieldRef)

	c.bindFetch(obj)
	// Bind screen manager
	c.screenManager.bind(obj)
	// Bind action manager
	c.actionManager.bind(obj)
	// Bind toast manager
	c.toastManager.bind(obj)

	if c.ext.Plugin != nil {
		for _, permission := range c.ext.Plugin.Permissions {
			switch permission {
			case extension.PluginPermissionPlayback:
				// Bind playback to the context object
				plugin.GlobalAppContext.BindPlaybackToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
			case extension.PluginPermissionCron:
				// Bind cron to the context object
				plugin.GlobalAppContext.BindCronToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
			}
		}
	}

	_ = vm.Set("__ctx", obj)

	c.contextObj = obj
}

// RegisterEventListener is used to register a new event listener in a Goja function
func (c *Context) RegisterEventListener(events ...ClientEventType) *EventListener {
	id := uuid.New().String()
	listener := &EventListener{
		ID:       id,
		ListenTo: events,
		Channel:  make(chan *ClientPluginEvent),
	}
	c.eventListeners.Set(id, listener)
	return listener
}

func (c *Context) UnregisterEventListener(id string) {
	listener, ok := c.eventListeners.Get(id)
	if !ok {
		return
	}
	close(listener.Channel)
	c.eventListeners.Delete(id)
}

// SendEventToClient sends an event to the client
// It always passes the extension ID
func (c *Context) SendEventToClient(eventType ServerEventType, payload interface{}) {
	c.wsEventManager.SendEvent(string(events.PluginEvent), &ServerPluginEvent{
		ExtensionID: c.ext.ID,
		Type:        eventType,
		Payload:     payload,
	})
}

// PrintState prints all states to the logger
func (c *Context) PrintState() {
	c.states.Range(func(key string, state *State) bool {
		c.logger.Info().Msgf("State %s = %+v", key, state.Value)
		return true
	})
}

func (c *Context) GetContextObj() (*goja.Object, bool) {
	return c.contextObj, c.contextObj != nil
}

// HandleTypeError interrupts the UI the first time we encounter a type error.
// Interrupting early is better to catch wrong usage of the API.
func (c *Context) HandleTypeError(msg string) {
	// c.mu.Lock()
	// defer c.mu.Unlock()

	c.logger.Error().Err(fmt.Errorf(msg)).Msg("plugin: Type error, interrupting UI")
	c.fatalError(fmt.Errorf(msg))
	// panic(c.vm.NewTypeError(msg))
}

// HandleException interrupts the UI after a certain number of exceptions have occurred.
// As opposed to HandleTypeError, this is more-so for unexpected errors and not wrong usage of the API.
func (c *Context) HandleException(err error) {
	// c.mu.Lock()
	// defer c.mu.Unlock()

	c.exceptionCount++
	if c.exceptionCount >= MaxExceptions {
		c.logger.Error().Err(err).Msg("plugin: Too many errors, interrupting UI")
		c.fatalError(err)
	}
}

func (c *Context) fatalError(err error) {
	c.logger.Error().Err(err).Msg("plugin: Fatal error, interrupting UI")
	if err != nil {
		c.SendEventToClient(ServerFatalErrorEvent, ServerFatalErrorEventPayload{
			Error: err.Error(),
		})
	} else {
		c.SendEventToClient(ServerFatalErrorEvent, ServerFatalErrorEventPayload{
			Error: fmt.Sprintf("plugin '%s' has encountered a fatal error and has been terminated.", c.ext.Name),
		})
	}

	c.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("Plugin: '%s' has encountered a fatal error and has been terminated.", c.ext.Name))

	c.ui.Unload()
}

func (c *Context) registerOnCleanup(fn func()) {
	c.atomicCleanupCounter.Add(1)
	c.onCleanupFns.Set(c.atomicCleanupCounter.Load(), fn)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// jsState is used to create a new state object
//
//	Example:
//	const text = ctx.state("Hello, world!");
//	text.set("Button clicked");
//	text.get(); // "Button clicked"
//	text.length; // 15
//	text.set(p => p + "!!!!");
//	text.get(); // "Button clicked!!!!"
//	text.length; // 19
func (c *Context) jsState(call goja.FunctionCall) goja.Value {
	id := uuid.New().String()
	initial := goja.Undefined()
	if len(call.Arguments) > 0 {
		initial = call.Argument(0)
	}

	state := &State{
		ID:    id,
		Value: initial,
	}

	// Store the initial state
	c.states.Set(id, state)

	// Create a new JS object to represent the state
	stateObj := c.vm.NewObject()

	// Define getter and setter functions that interact with the Go-managed state
	jsGetState := func(call goja.FunctionCall) goja.Value {
		res, _ := c.states.Get(id)
		return res.Value
	}
	jsSetState := func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			arg := call.Argument(0)
			// e.g. state.set(prev => prev + "!!!")
			if callback, ok := goja.AssertFunction(arg); ok {
				prevState, ok := c.states.Get(id)
				if ok {
					newVal, _ := callback(goja.Undefined(), prevState.Value)
					c.states.Set(id, &State{
						ID:    id,
						Value: newVal,
					})
					c.queueStateUpdate(id)
				}
			} else {
				c.states.Set(id, &State{
					ID:    id,
					Value: arg,
				})
				c.queueStateUpdate(id)
			}
		}
		return goja.Undefined()
	}

	jsGetStateVal := c.vm.ToValue(jsGetState)
	jsSetStateVal := c.vm.ToValue(jsSetState)

	// Define a dynamic state object that includes a 'value' property, get(), set(), and length
	jsDynamicDefFuncValue, err := c.vm.RunString(`(function(obj, getter, setter) {
	Object.defineProperty(obj, 'value', {
		get: getter,
		set: setter,
		enumerable: true,
		configurable: true
	});
	obj.get = function() { return this.value; };
	obj.set = function(val) { this.value = val; return val; };
	Object.defineProperty(obj, 'length', {
		get: function() {
			var val = this.value;
			return (typeof val === 'string' ? val.length : undefined);
		},
		enumerable: true,
		configurable: true
	});
	return obj;
})`)
	if err != nil {
		c.HandleTypeError(err.Error())
	}
	jsDynamicDefFunc, ok := goja.AssertFunction(jsDynamicDefFuncValue)
	if !ok {
		c.HandleTypeError("dynamic definition is not a function")
	}

	jsDynamicState, err := jsDynamicDefFunc(goja.Undefined(), stateObj, jsGetStateVal, jsSetStateVal)
	if err != nil {
		c.HandleTypeError(err.Error())
	}

	// Attach hidden state ID for subscription
	if obj, ok := jsDynamicState.(*goja.Object); ok {
		_ = obj.Set("__stateId", id)
	}

	return jsDynamicState
}

// jsSetTimeout
//
//	Example:
//	const cancel = ctx.setTimeout(() => {
//		console.log("Printing after 1 second");
//	}, 1000);
//	cancel(); // cancels the timeout
func (c *Context) jsSetTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		c.HandleTypeError("setTimeout requires a function and a delay")
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		c.HandleTypeError("setTimeout requires a function")
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		c.HandleTypeError("delay must be a number")
	}

	ctx, cancel := context.WithCancel(context.Background())

	globalObj := c.vm.GlobalObject()

	go func(fn goja.Callable, globalObj goja.Value) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delay) * time.Millisecond):
			c.scheduler.ScheduleAsync(func() error {
				_, err := fn(globalObj)
				return err
			})
		}
	}(fn, globalObj)

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	}

	return c.vm.ToValue(cancelFunc)
}

// jsSetInterval
//
//	Example:
//	const cancel = ctx.setInterval(() => {
//		console.log("Printing every second");
//	}, 1000);
//	cancel(); // cancels the interval
func (c *Context) jsSetInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.HandleTypeError("setInterval requires a function and a delay")
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		c.HandleTypeError("setInterval requires a function")
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		c.HandleTypeError("delay must be a number")
	}

	globalObj := c.vm.GlobalObject()

	ctx, cancel := context.WithCancel(context.Background())
	go func(fn goja.Callable, globalObj goja.Value) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(delay) * time.Millisecond):
				c.scheduler.ScheduleAsync(func() error {
					_, err := fn(globalObj)
					return err
				})
			}
		}
	}(fn, globalObj)

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	}

	return c.vm.ToValue(cancelFunc)
}

// jsEffect
//
//	Example:
//	const text = ctx.state("Hello, world!");
//	ctx.effect(() => {
//		console.log("Text changed");
//	}, [text]);
//	text.set("Hello, world!"); // This will not trigger the effect
//	text.set("Hello, world! 2"); // This will trigger the effect
func (c *Context) jsEffect(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.HandleTypeError("effect requires a function and an array of dependencies")
	}

	effectFn, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		c.HandleTypeError("first argument to effect must be a function")
	}

	depsObj, ok := call.Argument(1).(*goja.Object)
	// If no dependencies, execute effect once and return
	if !ok {
		c.scheduler.ScheduleAsync(func() error {
			_, err := effectFn(goja.Undefined())
			return err
		})
		return c.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return goja.Undefined()
		})
	}

	// Generate unique ID for this effect
	effectID := uuid.New().String()

	// Prepare dependencies and their old values
	lengthVal := depsObj.Get("length")
	depsLen := int(lengthVal.ToInteger())

	// If dependency array is empty, execute effect once and return
	if depsLen == 0 {
		c.scheduler.ScheduleAsync(func() error {
			_, err := effectFn(goja.Undefined())
			return err
		})
		return c.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return goja.Undefined()
		})
	}

	deps := make([]*goja.Object, depsLen)
	oldValues := make([]goja.Value, depsLen)
	dropIDs := make([]string, depsLen) // to store state IDs of dependencies
	for i := 0; i < depsLen; i++ {
		depVal := depsObj.Get(fmt.Sprintf("%d", i))
		depObj, ok := depVal.(*goja.Object)
		if !ok {
			c.HandleTypeError("dependency is not an object")
		}
		deps[i] = depObj
		oldValues[i] = depObj.Get("value")

		idVal := depObj.Get("__stateId")
		exported := idVal.Export()
		idStr, ok := exported.(string)
		if !ok {
			idStr = fmt.Sprintf("%v", exported)
		}
		dropIDs[i] = idStr
	}

	globalObj := c.vm.GlobalObject()

	// Subscribe to state updates
	subChan := c.subscribeStateUpdates()
	ctxEffect, cancel := context.WithCancel(context.Background())
	go func(effectFn *goja.Callable, globalObj goja.Value) {
		for {
			select {
			case <-ctxEffect.Done():
				return
			case updatedState := <-subChan:
				if effectFn != nil && updatedState != nil {
					// Check if the updated state is one of our dependencies by matching __stateId
					for i, depID := range dropIDs {
						if depID == updatedState.ID {
							newVal := deps[i].Get("value")
							if !reflect.DeepEqual(oldValues[i].Export(), newVal.Export()) {
								oldValues[i] = newVal

								// Check for infinite loops
								c.mu.Lock()
								if c.effectStack[effectID] {
									c.logger.Warn().Msgf("Detected potential infinite loop in effect %s, skipping execution", effectID)
									c.mu.Unlock()
									continue
								}

								// Clean up old calls and check rate
								c.cleanupOldEffectCalls(effectID)
								callsInWindow := len(c.effectCalls[effectID])
								if callsInWindow >= MaxEffectCallsPerWindow {
									c.mu.Unlock()
									c.fatalError(fmt.Errorf("effect %s exceeded rate limit with %d calls in %dms window", effectID, callsInWindow, EffectTimeWindow))
									return
								}

								// Track this call
								c.effectStack[effectID] = true
								c.effectCalls[effectID] = append(c.effectCalls[effectID], time.Now())
								c.mu.Unlock()

								c.scheduler.ScheduleAsync(func() error {
									_, err := (*effectFn)(globalObj)
									c.mu.Lock()
									c.effectStack[effectID] = false
									c.mu.Unlock()
									return err
								})
							}
						}
					}
				}
			}
		}
	}(&effectFn, globalObj)

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		c.mu.Lock()
		delete(c.effectCalls, effectID)
		delete(c.effectStack, effectID)
		c.mu.Unlock()
		return goja.Undefined()
	}

	return c.vm.ToValue(cancelFunc)
}

// jsRegisterEventHandler
//
//	Example:
//	ctx.registerEventHandler("button-clicked", (e) => {
//		console.log("Button clicked", e);
//	});
func (c *Context) jsRegisterEventHandler(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.HandleTypeError("registerEventHandler requires a handler name and a function")
	}

	handlerName := call.Argument(0).String()
	handlerCallback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		c.HandleTypeError("second argument to registerEventHandler must be a function")
	}

	eventListener := c.RegisterEventListener(ClientEventHandlerTriggeredEvent)
	payload := ClientEventHandlerTriggeredEventPayload{}

	globalObj := c.vm.GlobalObject()

	go func(handlerCallback goja.Callable, globalObj goja.Value) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientEventHandlerTriggeredEvent, &payload) {
				// Check if the handler name matches
				if payload.HandlerName == handlerName {
					c.scheduler.ScheduleAsync(func() error {
						// Trigger the callback with the event payload
						_, err := handlerCallback(globalObj, c.vm.ToValue(payload.Event))
						return err
					})
				}
			}
		}
	}(handlerCallback, globalObj)

	return goja.Undefined()
}

// jsRegisterFieldRef allows to dynamically handle the value of a field outside the rendering context
//
//	Example:
//	const fieldRef = ctx.registerFieldRef("my-field")
//	fieldRef.setValue("Hello World!") // Triggers an immediate update on the client
//	fieldRef.current // "Hello World!"
//
//	tray.render(() => tray.input({ fieldRef: "my-field" }))
func (c *Context) jsRegisterFieldRef(call goja.FunctionCall) goja.Value {
	fieldRefObj := c.vm.NewObject()

	if c.fieldRefCount >= MAX_FIELD_REFS {
		c.HandleTypeError("Too many field refs registered")
		return goja.Undefined()
	}

	c.fieldRefCount++

	var valueRef interface{}

	fieldRefName, ok := call.Argument(0).Export().(string)
	if !ok {
		c.HandleTypeError("registerFieldRef requires a field name")
	}

	fieldRefObj.Set("setValue", func(call goja.FunctionCall) goja.Value {
		value := call.Argument(0).Export()
		if value == nil {
			c.HandleTypeError("setValue requires a value")
		}

		c.SendEventToClient(ServerFieldRefSetValueEvent, ServerFieldRefSetValueEventPayload{
			FieldRef: fieldRefName,
			Value:    value,
		})

		valueRef = value
		fieldRefObj.Set("current", value)

		return goja.Undefined()
	})

	valueRef = nil
	fieldRefObj.Set("current", goja.Undefined())

	// Listen for changes from the client
	eventListener := c.RegisterEventListener(ClientFieldRefSendValueEvent, ClientRenderTrayEvent)
	payload := ClientFieldRefSendValueEventPayload{}
	renderPayload := ClientRenderTrayEventPayload{}

	globalObj := c.vm.GlobalObject()

	go func(eventListener *EventListener, globalObj goja.Value) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientFieldRefSendValueEvent, &payload) {
				if payload.Value != nil {
					// Schedule the update of the object
					c.scheduler.ScheduleAsync(func() error {
						fieldRefObj.Set("current", payload.Value)
						return nil
					})
				}
			}
			// Check if the client is requesting a render
			// If it is, we send the current value to the client
			if event.ParsePayloadAs(ClientRenderTrayEvent, &renderPayload) {
				c.SendEventToClient(ServerFieldRefSetValueEvent, ServerFieldRefSetValueEventPayload{
					FieldRef: fieldRefName,
					Value:    valueRef,
				})
			}
		}
	}(eventListener, globalObj)

	return fieldRefObj
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *Context) subscribeStateUpdates() chan *State {
	ch := make(chan *State, 10)
	c.mu.Lock()
	c.stateSubscribers = append(c.stateSubscribers, ch)
	c.mu.Unlock()
	return ch
}

func (c *Context) publishStateUpdate(id string) {
	state, ok := c.states.Get(id)
	if !ok {
		return
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, sub := range c.stateSubscribers {
		select {
		case sub <- state:
		default:
		}
	}
}

func (c *Context) cleanupOldEffectCalls(effectID string) {
	now := time.Now()
	window := time.Duration(EffectTimeWindow) * time.Millisecond
	var validCalls []time.Time

	for _, t := range c.effectCalls[effectID] {
		if now.Sub(t) <= window {
			validCalls = append(validCalls, t)
		}
	}

	c.effectCalls[effectID] = validCalls
}

// queueStateUpdate adds a state update to the batch queue
func (c *Context) queueStateUpdate(id string) {
	c.updateBatchMu.Lock()
	defer c.updateBatchMu.Unlock()

	// Add to pending updates
	c.pendingStateUpdates[id] = struct{}{}

	// Start the timer if it's not running
	if !c.updateBatchTimer.Stop() {
		select {
		case <-c.updateBatchTimer.C:
			// Timer already fired, drain the channel
		default:
			// Timer was already stopped
		}
	}
	c.updateBatchTimer.Reset(time.Duration(StateUpdateBatchInterval) * time.Millisecond)
}

// flushStateUpdates processes all pending state updates
func (c *Context) flushStateUpdates() {
	c.updateBatchMu.Lock()

	// Get all pending updates
	pendingUpdates := make([]string, 0, len(c.pendingStateUpdates))
	for id := range c.pendingStateUpdates {
		pendingUpdates = append(pendingUpdates, id)
	}

	// Clear the pending updates
	c.pendingStateUpdates = make(map[string]struct{})

	c.updateBatchMu.Unlock()

	// Process all updates
	for _, id := range pendingUpdates {
		c.publishStateUpdate(id)
	}

	// Trigger UI update after state changes
	c.triggerUIUpdate()
}

// triggerUIUpdate schedules a UI update after state changes
func (c *Context) triggerUIUpdate() {
	c.uiUpdateMu.Lock()
	defer c.uiUpdateMu.Unlock()

	// Rate limit UI updates
	if time.Since(c.lastUIUpdateAt) < time.Millisecond*time.Duration(UIUpdateRateLimit) {
		return
	}

	c.lastUIUpdateAt = time.Now()

	// Trigger tray update if available
	if c.trayManager != nil {
		c.trayManager.renderTrayScheduled()
	}
}

// Cleanup stops the update batch timer and performs any necessary cleanup
func (c *Context) Cleanup() {
	if c.updateBatchTimer != nil {
		c.updateBatchTimer.Stop()
	}

	// Flush any remaining updates
	c.flushStateUpdates()
}

func (c *Context) Stop() {
	c.logger.Debug().Msg("plugin: Stopping context")

	if c.updateBatchTimer != nil {
		c.updateBatchTimer.Stop()
	}

	// Stop the scheduler
	c.scheduler.Stop()

	// Stop all event listeners
	for _, listener := range c.eventListeners.Values() {
		go func(listener *EventListener) {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			listener.closed = true
			close(listener.Channel)
		}(listener)
	}

	// Stop all state subscribers
	for _, sub := range c.stateSubscribers {
		go func(sub chan *State) {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			close(sub)
		}(sub)
	}

	c.onCleanupFns.Range(func(key int64, fn func()) bool {
		fn()
		return true
	})
	c.onCleanupFns.Clear()

	c.logger.Debug().Msg("plugin: Stopped context")
}
