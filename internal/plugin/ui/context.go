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
	"github.com/samber/mo"
)

// Constants for event batching
const (
	maxEventBatchSize       = 20 // Maximum number of events in a batch
	eventBatchFlushInterval = 10 // Flush interval in milliseconds
)

// BatchedPluginEvents represents a collection of plugin events to be sent together
type BatchedPluginEvents struct {
	Events []*ServerPluginEvent `json:"events"`
}

// BatchedEvents represents a collection of events to be sent together
type BatchedEvents struct {
	Events []events.WebsocketClientEvent `json:"events"`
}

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
	eventBus         *result.Map[ClientEventType, *result.Map[string, *EventListener]] // map[string]map[string]*EventListener (event -> listenerID -> listener)
	contextObj       *goja.Object

	fieldRefCount  int                    // Number of field refs registered
	exceptionCount int                    // Number of exceptions that have occurred
	effectStack    map[string]bool        // Track currently executing effects to prevent infinite loops
	effectCalls    map[string][]time.Time // Track effect calls within time window

	// State update batching
	updateBatchMu       sync.Mutex
	pendingStateUpdates map[string]struct{} // Set of state IDs with pending updates
	updateBatchTimer    *time.Timer         // Timer for flushing batched updates

	// Event batching system
	eventBatchMu        sync.Mutex
	pendingClientEvents []*ServerPluginEvent // Queue of pending events to send to client
	eventBatchTimer     *time.Timer          // Timer for flushing batched events
	eventBatchSize      int                  // Current size of the event batch

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
	domManager            *DOMManager            // DOM manipulation manager
	notificationManager   *NotificationManager   // Register and manage notifications

	atomicCleanupCounter atomic.Int64
	onCleanupFns         *result.Map[int64, func()]
	cron                 mo.Option[*plugin.Cron]

	registeredInlineEventHandlers *result.Map[string, *EventListener]
}

type State struct {
	ID    string
	Value goja.Value
}

// EventListener is used by Goja methods to listen for events from the client
type EventListener struct {
	ID       string
	ListenTo []ClientEventType        // Optional event type to listen for
	queue    []*ClientPluginEvent     // Queue for event payloads
	callback func(*ClientPluginEvent) // Callback function to process events
	closed   bool
	mu       sync.Mutex
}

func NewContext(ui *UI) *Context {
	ret := &Context{
		ui:                            ui,
		ext:                           ui.ext,
		logger:                        ui.logger,
		vm:                            ui.vm,
		states:                        result.NewResultMap[string, *State](),
		fetchSem:                      make(chan struct{}, MaxConcurrentFetchRequests),
		stateSubscribers:              make([]chan *State, 0),
		eventBus:                      result.NewResultMap[ClientEventType, *result.Map[string, *EventListener]](),
		wsEventManager:                ui.wsEventManager,
		effectStack:                   make(map[string]bool),
		effectCalls:                   make(map[string][]time.Time),
		pendingStateUpdates:           make(map[string]struct{}),
		lastUIUpdateAt:                time.Now().Add(-time.Hour), // Initialize to a time in the past
		atomicCleanupCounter:          atomic.Int64{},
		onCleanupFns:                  result.NewResultMap[int64, func()](),
		cron:                          mo.None[*plugin.Cron](),
		registeredInlineEventHandlers: result.NewResultMap[string, *EventListener](),
		pendingClientEvents:           make([]*ServerPluginEvent, 0, maxEventBatchSize),
		eventBatchSize:                0,
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
	ret.domManager = NewDOMManager(ret)
	ret.notificationManager = NewNotificationManager(ret)

	// Initialize the event batch timer
	ret.eventBatchTimer = time.AfterFunc(eventBatchFlushInterval*time.Millisecond, func() {
		ret.flushEventBatch()
	})
	ret.eventBatchTimer.Stop()

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
	_ = obj.Set("eventHandler", c.jsEventHandler)
	_ = obj.Set("fieldRef", c.jsfieldRef)

	c.bindFetch(obj)
	// Bind screen manager
	c.screenManager.bind(obj)
	// Bind action manager
	c.actionManager.bind(obj)
	// Bind toast manager
	c.toastManager.bind(obj)
	// Bind DOM manager
	c.domManager.BindToObj(vm, obj)
	// Bind manga
	plugin.GlobalAppContext.BindMangaToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind anime
	plugin.GlobalAppContext.BindAnimeToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind continuity
	plugin.GlobalAppContext.BindContinuityToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind filler manager
	plugin.GlobalAppContext.BindFillerManagerToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind auto downloader
	plugin.GlobalAppContext.BindAutoDownloaderToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind auto scanner
	plugin.GlobalAppContext.BindAutoScannerToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind external player link
	plugin.GlobalAppContext.BindExternalPlayerLinkToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind onlinestream
	plugin.GlobalAppContext.BindOnlinestreamToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
	// Bind mediastream
	plugin.GlobalAppContext.BindMediastreamToContextObj(vm, obj, c.logger, c.ext, c.scheduler)

	if c.ext.Plugin != nil {
		for _, permission := range c.ext.Plugin.Permissions.Scopes {
			switch permission {
			case extension.PluginPermissionPlayback:
				// Bind playback to the context object
				plugin.GlobalAppContext.BindPlaybackToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
			case extension.PluginPermissionSystem:
				plugin.GlobalAppContext.BindDownloaderToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
			case extension.PluginPermissionCron:
				// Bind cron to the context object
				cron := plugin.GlobalAppContext.BindCronToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
				c.cron = mo.Some(cron)
			case extension.PluginPermissionNotification:
				// Bind notification to the context object
				c.notificationManager.bind(obj)
			case extension.PluginPermissionDiscord:
				// Bind discord to the context object
				plugin.GlobalAppContext.BindDiscordToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
			case extension.PluginPermissionTorrentClient:
				// Bind torrent client to the context object
				plugin.GlobalAppContext.BindTorrentClientToContextObj(vm, obj, c.logger, c.ext, c.scheduler)
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
		queue:    make([]*ClientPluginEvent, 0),
		closed:   false,
	}

	// Register the listener for each event type
	for _, event := range events {
		if !c.eventBus.Has(event) {
			c.eventBus.Set(event, result.NewResultMap[string, *EventListener]())
		}
		listeners, _ := c.eventBus.Get(event)
		listeners.Set(id, listener)
	}

	return listener
}

func (c *Context) UnregisterEventListener(id string) {
	c.eventBus.Range(func(key ClientEventType, listenerMap *result.Map[string, *EventListener]) bool {
		listener, ok := listenerMap.Get(id)
		if !ok {
			return true
		}

		// Close the listener first before removing it
		listener.Close()

		listenerMap.Delete(id)

		return true
	})
}

func (c *Context) UnregisterEventListenerE(e *EventListener) {
	if e == nil {
		return
	}

	for _, event := range e.ListenTo {
		listeners, ok := c.eventBus.Get(event)
		if !ok {
			continue
		}

		listener, ok := listeners.Get(e.ID)
		if !ok {
			continue
		}

		// Close the listener first before removing it
		listener.Close()

		listeners.Delete(e.ID)
	}
}

func (e *EventListener) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return
	}
	e.closed = true
	e.queue = nil // Clear the queue
}

func (e *EventListener) Send(event *ClientPluginEvent) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("plugin: Error sending event %s\n", event.Type)
		}
	}()

	e.mu.Lock()

	if e.closed {
		e.mu.Unlock()
		return
	}

	// Add event to queue
	e.queue = append(e.queue, event)
	hasCallback := e.callback != nil

	e.mu.Unlock()

	// Process immediately if callback is set - call after releasing the lock
	if hasCallback {
		go e.processEvents()
	}
}

// SetCallback sets a function to call when events are received
func (e *EventListener) SetCallback(callback func(*ClientPluginEvent)) {
	e.mu.Lock()

	e.callback = callback
	hasEvents := len(e.queue) > 0 && !e.closed

	e.mu.Unlock()

	// Process any existing events in the queue - call after releasing the lock
	if hasEvents {
		go e.processEvents()
	}
}

// processEvents processes all events in the queue
func (e *EventListener) processEvents() {
	var _events []*ClientPluginEvent
	var callback func(*ClientPluginEvent)

	e.mu.Lock()
	if e.closed || e.callback == nil {
		e.mu.Unlock()
		return
	}

	// Get all _events from the queue and the callback
	_events = make([]*ClientPluginEvent, len(e.queue))
	copy(_events, e.queue)
	e.queue = e.queue[:0] // Clear the queue
	callback = e.callback // Make a copy of the callback

	e.mu.Unlock()

	// Process _events outside the lock with the copied callback
	for _, event := range _events {
		// Wrap each callback in a recover to prevent one bad event from stopping all processing
		func(evt *ClientPluginEvent) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("plugin: Error processing event: %v\n", r)
				}
			}()
			callback(evt)
		}(event)
	}
}

// SendEventToClient sends an event to the client
// It always passes the extension ID
func (c *Context) SendEventToClient(eventType ServerEventType, payload interface{}) {
	c.queueEventToClient("", eventType, payload)
}

// SendEventToClientWithClientID sends an event to the client with a specific client ID
func (c *Context) SendEventToClientWithClientID(clientID string, eventType ServerEventType, payload interface{}) {
	c.wsEventManager.SendEventTo(clientID, string(events.PluginEvent), &ServerPluginEvent{
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

// handleTypeError interrupts the UI the first time we encounter a type error.
// Interrupting early is better to catch wrong usage of the API.
func (c *Context) handleTypeError(msg string) {
	c.logger.Error().Err(fmt.Errorf(msg)).Msg("plugin: Type error")
	// c.fatalError(fmt.Errorf(msg))
	panic(c.vm.NewTypeError(msg))
}

// handleException interrupts the UI after a certain number of exceptions have occurred.
// As opposed to HandleTypeError, this is more-so for unexpected errors and not wrong usage of the API.
func (c *Context) handleException(err error) {
	// c.mu.Lock()
	// defer c.mu.Unlock()

	c.wsEventManager.SendEvent(events.ConsoleWarn, fmt.Sprintf("plugin(%s): Exception: %s", c.ext.ID, err.Error()))
	c.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("plugin(%s): Exception: %s", c.ext.ID, err.Error()))

	c.exceptionCount++
	if c.exceptionCount >= MaxExceptions {
		newErr := fmt.Errorf("plugin(%s): Encountered too many exceptions, last error: %w", c.ext.ID, err)
		c.logger.Error().Err(newErr).Msg("plugin: Encountered too many exceptions, interrupting plugin")
		c.fatalError(newErr)
	}
}

func (c *Context) fatalError(err error) {
	c.logger.Error().Err(err).Msg("plugin: Encountered fatal error, interrupting plugin")
	c.wsEventManager.SendEvent(events.ConsoleWarn, fmt.Sprintf("plugin(%s): Encountered fatal error, interrupting plugin", c.ext.ID))
	c.ui.lastException = err.Error()

	c.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("plugin(%s): Fatal error: %s", c.ext.ID, err.Error()))
	c.wsEventManager.SendEvent(events.ConsoleWarn, fmt.Sprintf("plugin(%s): Fatal error: %s", c.ext.ID, err.Error()))

	// Unload the UI and signal the Plugin that it's been terminated
	c.ui.Unload(true)
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
		c.handleTypeError(err.Error())
	}
	jsDynamicDefFunc, ok := goja.AssertFunction(jsDynamicDefFuncValue)
	if !ok {
		c.handleTypeError("dynamic definition is not a function")
	}

	jsDynamicState, err := jsDynamicDefFunc(goja.Undefined(), stateObj, jsGetStateVal, jsSetStateVal)
	if err != nil {
		c.handleTypeError(err.Error())
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
		c.handleTypeError("setTimeout requires a function and a delay")
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		c.handleTypeError("setTimeout requires a function")
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		c.handleTypeError("delay must be a number")
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
		c.handleTypeError("setInterval requires a function and a delay")
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		c.handleTypeError("setInterval requires a function")
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		c.handleTypeError("delay must be a number")
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

	c.registerOnCleanup(func() {
		cancel()
	})

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
		c.handleTypeError("effect requires a function and an array of dependencies")
	}

	effectFn, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		c.handleTypeError("first argument to effect must be a function")
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
			c.handleTypeError("dependency is not an object")
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

	c.registerOnCleanup(func() {
		cancel()
	})

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
		c.handleTypeError("registerEventHandler requires a handler name and a function")
	}

	handlerName := call.Argument(0).String()
	handlerCallback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		c.handleTypeError("second argument to registerEventHandler must be a function")
	}

	eventListener := c.RegisterEventListener(ClientEventHandlerTriggeredEvent)
	payload := ClientEventHandlerTriggeredEventPayload{}

	globalObj := c.vm.GlobalObject()

	eventListener.SetCallback(func(event *ClientPluginEvent) {
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
	})

	return goja.Undefined()
}

// jsEventHandler - inline event handler
//
//	Example:
//	tray.render(() => tray.button("Click me", {
//		onClick: ctx.eventHandler("unique-key", (e) => {
//			console.log("Button clicked", e);
//		})
//	}));
func (c *Context) jsEventHandler(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.handleTypeError("eventHandler requires a function")
	}

	uniqueKey := call.Argument(0).String()
	if existingListener, ok := c.registeredInlineEventHandlers.Get(uniqueKey); ok {
		c.UnregisterEventListenerE(existingListener)
	}

	handlerCallback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		c.handleTypeError("second argument to eventHandler must be a function")
	}

	id := "__eventHandler__" + uuid.New().String()

	eventListener := c.RegisterEventListener(ClientEventHandlerTriggeredEvent)
	payload := ClientEventHandlerTriggeredEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientEventHandlerTriggeredEvent, &payload) {
			// Check if the handler name matches
			if payload.HandlerName == id {
				c.scheduler.ScheduleAsync(func() error {
					// Trigger the callback with the event payload
					_, err := handlerCallback(goja.Undefined(), c.vm.ToValue(payload.Event))
					return err
				})
			}
		}
	})

	c.registeredInlineEventHandlers.Set(uniqueKey, eventListener)

	return c.vm.ToValue(id)
}

// jsfieldRef allows to dynamically handle the value of a field outside the rendering context
//
//	Example:
//	const fieldRef = ctx.fieldRef("defaultValue")
//	fieldRef.current // "defaultValue"
//	fieldRef.setValue("Hello World!") // Triggers an immediate update on the client
//	fieldRef.current // "Hello World!"
//
//	tray.render(() => tray.input({ fieldRef: "my-field" }))
func (c *Context) jsfieldRef(call goja.FunctionCall) goja.Value {
	fieldRefObj := c.vm.NewObject()

	if c.fieldRefCount >= MAX_FIELD_REFS {
		c.handleTypeError("Too many field refs registered")
		return goja.Undefined()
	}

	id := uuid.New().String()
	fieldRefObj.Set("__ID", id)

	c.fieldRefCount++

	var valueRef interface{}
	var onChangeCallback func(value interface{})

	// Handle default value if provided
	if len(call.Arguments) > 0 {
		valueRef = call.Argument(0).Export()
		fieldRefObj.Set("current", valueRef)
	} else {
		fieldRefObj.Set("current", goja.Undefined())
	}

	fieldRefObj.Set("setValue", func(call goja.FunctionCall) goja.Value {
		value := call.Argument(0).Export()
		if value == nil {
			c.handleTypeError("setValue requires a value")
		}

		c.SendEventToClient(ServerFieldRefSetValueEvent, ServerFieldRefSetValueEventPayload{
			FieldRef: id,
			Value:    value,
		})

		valueRef = value
		fieldRefObj.Set("current", value)

		return goja.Undefined()
	})

	fieldRefObj.Set("onValueChange", func(call goja.FunctionCall) goja.Value {
		callback, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			c.handleTypeError("onValueChange requires a function")
		}

		onChangeCallback = func(value interface{}) {
			_, err := callback(goja.Undefined(), c.vm.ToValue(value))
			if err != nil {
				c.handleTypeError(err.Error())
			}
		}

		return goja.Undefined()
	})

	// Listen for changes from the client
	eventListener := c.RegisterEventListener(ClientFieldRefSendValueEvent, ClientRenderTrayEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		payload := ClientFieldRefSendValueEventPayload{}
		renderPayload := ClientRenderTrayEventPayload{}
		if event.ParsePayloadAs(ClientFieldRefSendValueEvent, &payload) && payload.FieldRef == id {
			valueRef = payload.Value
			// Schedule the update of the object
			if payload.Value != nil {
				c.scheduler.ScheduleAsync(func() error {
					fieldRefObj.Set("current", payload.Value)
					return nil
				})
				if onChangeCallback != nil {
					c.scheduler.ScheduleAsync(func() error {
						onChangeCallback(payload.Value)
						return nil
					})
				}
			}
		}

		// Check if the client is requesting a render
		// If it is, we send the current value to the client
		if event.ParsePayloadAs(ClientRenderTrayEvent, &renderPayload) {
			c.SendEventToClient(ServerFieldRefSetValueEvent, ServerFieldRefSetValueEventPayload{
				FieldRef: id,
				Value:    valueRef,
			})
		}
	})

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

// Cleanup is called when the UI is being unloaded
func (c *Context) Cleanup() {
	// Flush any pending state updates
	c.flushStateUpdates()

	// Flush any pending events
	c.flushEventBatch()
}

// Stop is called when the UI is being unloaded
func (c *Context) Stop() {
	c.logger.Debug().Msg("plugin: Stopping context")

	if c.updateBatchTimer != nil {
		c.logger.Trace().Msg("plugin: Stopping update batch timer")
		c.updateBatchTimer.Stop()
	}

	if c.eventBatchTimer != nil {
		c.logger.Trace().Msg("plugin: Stopping event batch timer")
		c.eventBatchTimer.Stop()
	}

	// Stop the scheduler
	c.logger.Trace().Msg("plugin: Stopping scheduler")
	c.scheduler.Stop()

	// Stop the cron
	if cron, hasCron := c.cron.Get(); hasCron {
		c.logger.Trace().Msg("plugin: Stopping cron")
		cron.Stop()
	}

	// Stop all event listeners
	c.logger.Trace().Msg("plugin: Stopping event listeners")
	eventListenersToClose := make([]*EventListener, 0)

	// First collect all listeners to avoid modification during iteration
	c.eventBus.Range(func(_ ClientEventType, listenerMap *result.Map[string, *EventListener]) bool {
		listenerMap.Range(func(_ string, listener *EventListener) bool {
			eventListenersToClose = append(eventListenersToClose, listener)
			return true
		})
		return true
	})

	// Then close them all outside the locks
	for _, listener := range eventListenersToClose {
		func(l *EventListener) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Err(fmt.Errorf("%v", r)).Msg("plugin: Error stopping event listener")
				}
			}()
			l.Close()
		}(listener)
	}

	// Finally clear the maps
	c.eventBus.Range(func(_ ClientEventType, listenerMap *result.Map[string, *EventListener]) bool {
		listenerMap.Clear()
		return true
	})
	c.eventBus.Clear()

	// Stop all state subscribers
	c.logger.Trace().Msg("plugin: Stopping state subscribers")
	for _, sub := range c.stateSubscribers {
		go func(sub chan *State) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Err(fmt.Errorf("%v", r)).Msg("plugin: Error stopping state subscriber")
				}
			}()
			close(sub)
		}(sub)
	}

	// Run all cleanup functions
	c.onCleanupFns.Range(func(key int64, fn func()) bool {
		fn()
		return true
	})
	c.onCleanupFns.Clear()

	c.actionManager.UnmountAll()
	c.actionManager.renderAnimePageButtons()

	c.logger.Debug().Msg("plugin: Stopped context")
}

// queueEventToClient adds an event to the batch queue for sending to the client
func (c *Context) queueEventToClient(clientID string, eventType ServerEventType, payload interface{}) {
	c.eventBatchMu.Lock()
	defer c.eventBatchMu.Unlock()

	// Create the plugin event
	event := &ServerPluginEvent{
		ExtensionID: c.ext.ID,
		Type:        eventType,
		Payload:     payload,
	}

	// Add to pending events
	c.pendingClientEvents = append(c.pendingClientEvents, event)
	c.eventBatchSize++

	// If this is the first event, start the timer
	if c.eventBatchSize == 1 {
		c.eventBatchTimer.Reset(eventBatchFlushInterval * time.Millisecond)
	}

	// If we've reached max batch size, flush immediately
	if c.eventBatchSize >= maxEventBatchSize {
		// Use goroutine to avoid deadlock since we're already holding the lock
		go c.flushEventBatch()
	}
}

// flushEventBatch sends all pending events as a batch to the client
func (c *Context) flushEventBatch() {
	c.eventBatchMu.Lock()

	// If there are no events, just unlock and return
	if c.eventBatchSize == 0 {
		c.eventBatchMu.Unlock()
		return
	}

	// Stop the timer
	c.eventBatchTimer.Stop()

	// Create a copy of the pending events
	allEvents := make([]*ServerPluginEvent, len(c.pendingClientEvents))
	copy(allEvents, c.pendingClientEvents)

	// Clear the pending events
	c.pendingClientEvents = c.pendingClientEvents[:0]
	c.eventBatchSize = 0

	c.eventBatchMu.Unlock()

	// If only one event, send it directly to maintain compatibility with current system
	if len(allEvents) == 1 {
		// c.wsEventManager.SendEvent("plugin", allEvents[0])
		c.wsEventManager.SendEvent(string(events.PluginEvent), &ServerPluginEvent{
			ExtensionID: c.ext.ID,
			Type:        allEvents[0].Type,
			Payload:     allEvents[0].Payload,
		})
		return
	}

	// Send events as a batch
	batchPayload := &BatchedPluginEvents{
		Events: allEvents,
	}

	// Send the batch
	c.wsEventManager.SendEvent(string(events.PluginEvent), &ServerPluginEvent{
		ExtensionID: c.ext.ID,
		Type:        "plugin:batch-events",
		Payload:     batchPayload,
	})
}
