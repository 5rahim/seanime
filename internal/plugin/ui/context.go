package plugin_ui

import (
	"context"
	"fmt"
	"reflect"
	"seanime/internal/events"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Context manages the entire plugin UI during its lifecycle
type Context struct {
	ui *UI

	extensionID    string
	logger         *zerolog.Logger
	wsEventManager events.WSEventManagerInterface

	mu       sync.RWMutex
	fetchSem chan struct{} // Semaphore for concurrent fetch requests

	vm               *goja.Runtime
	states           *result.Map[string, *State]
	stateSubscribers []chan *State
	scheduler        *Scheduler // Schedule VM executions concurrently and execute them in order.
	wsSubscriber     *events.ClientEventSubscriber
	eventListeners   *result.Map[string, *EventListener] // Event listeners registered by plugin functions

	exceptionCount int                    // Number of exceptions that have occurred
	effectStack    map[string]bool        // Track currently executing effects to prevent infinite loops
	effectCalls    map[string][]time.Time // Track effect calls within time window

	webviewManager *WebviewManager // UNUSED
	screenManager  *ScreenManager  // Listen for screen events, send screen actions
	trayManager    *TrayManager    // Register and manage tray
	formManager    *FormManager    // Register and manage forms
	toastManager   *ToastManager   // Register and manage toasts
}

type State struct {
	ID    string
	Value goja.Value
}

// EventListener is a struct that contains the event type to listen for, and the channel for the event payload
// - It is registered by plugin functions
// - It is used to listen for events from the client
type EventListener struct {
	ID       string
	ListenTo []ClientEventType       // Optional event type to listen for
	Channel  chan *ClientPluginEvent // Channel for the event payload
}

func NewContext(ui *UI, extensionID string, logger *zerolog.Logger, vm *goja.Runtime, wsEventManager events.WSEventManagerInterface) *Context {
	ret := &Context{
		ui:               ui,
		extensionID:      extensionID,
		logger:           logger,
		vm:               vm,
		states:           result.NewResultMap[string, *State](),
		fetchSem:         make(chan struct{}, MAX_CONCURRENT_FETCH_REQUESTS),
		stateSubscribers: make([]chan *State, 0),
		eventListeners:   result.NewResultMap[string, *EventListener](),
		wsEventManager:   wsEventManager,
		effectStack:      make(map[string]bool),
		effectCalls:      make(map[string][]time.Time),
	}

	ret.scheduler = NewScheduler(ret)

	ret.trayManager = NewTrayManager(ret)
	ret.webviewManager = NewWebviewManager(ret)
	ret.screenManager = NewScreenManager(ret)
	ret.formManager = NewFormManager(ret)
	ret.toastManager = NewToastManager(ret)

	return ret
}

func (c *Context) bind(vm *goja.Runtime, obj *goja.Object) {
	_ = obj.Set("newTray", c.trayManager.jsNewTray)
	_ = obj.Set("newForm", c.formManager.jsNewForm)
	c.toastManager.BindToast(obj)

	_ = obj.Set("state", c.jsState)
	_ = obj.Set("setTimeout", c.jsSetTimeout)
	_ = obj.Set("sleep", c.jsSleep)
	_ = obj.Set("setInterval", c.jsSetInterval)
	_ = obj.Set("effect", c.jsEffect)
	_ = obj.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return c.vm.ToValue(c.jsFetch(call))
	})
	_ = obj.Set("registerEventHandler", c.jsRegisterEventHandler)

	_ = vm.Set("__ctx", obj)
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

// SendEventToClient sends an event to the client
// It always passes the extension ID
func (c *Context) SendEventToClient(eventType ServerEventType, payload interface{}) {
	c.wsEventManager.SendEvent(string(events.PluginEvent), &ServerPluginEvent{
		ExtensionID: c.extensionID,
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
	if c.exceptionCount >= MAX_EXCEPTIONS {
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
			Error: "The plugin has encountered a fatal error. It has been terminated.",
		})
	}
	c.ui.ClearInterrupt()
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
					c.publishStateUpdate(id)
				}
			} else {
				c.states.Set(id, &State{
					ID:    id,
					Value: arg,
				})
				c.publishStateUpdate(id)
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

	go func(fn goja.Callable) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delay) * time.Millisecond):
			if err := c.scheduler.Schedule(func() error {
				_, err := fn(goja.Undefined())
				return err
			}); err != nil {
				c.logger.Error().Err(err).Msg("plugin: Error running timeout callback")
			}
		}
	}(fn)

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

	ctx, cancel := context.WithCancel(context.Background())
	go func(fn goja.Callable) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(delay) * time.Millisecond):
				if err := c.scheduler.Schedule(func() error {
					_, err := fn(goja.Undefined())
					return err
				}); err != nil {
					c.logger.Error().Err(err).Msg("plugin: Error running interval callback")
				}
			}
		}
	}(fn)

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	}

	return c.vm.ToValue(cancelFunc)
}

// jsSleep
//
//	Example:
//	ctx.sleep(1000);
func (c *Context) jsSleep(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		c.HandleTypeError("sleep requires a delay")
	}

	delayValue := call.Argument(0)
	delay, ok := delayValue.Export().(int64)
	if !ok {
		c.HandleTypeError("delay must be a number")
	}

	time.Sleep(time.Duration(delay) * time.Millisecond)

	return goja.Undefined()
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
	if !ok {
		c.HandleTypeError("second argument to effect must be an array")
	}

	// Generate unique ID for this effect
	effectID := uuid.New().String()

	// Prepare dependencies and their old values
	lengthVal := depsObj.Get("length")
	depsLen := int(lengthVal.ToInteger())

	// If dependency array is empty, execute effect once and return
	if depsLen == 0 {
		if err := c.scheduler.Schedule(func() error {
			_, err := effectFn(goja.Undefined())
			return err
		}); err != nil {
			c.logger.Error().Err(err).Msg("plugin: Error running effect")
		}
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

	// Subscribe to state updates
	subChan := c.subscribeStateUpdates()
	ctxEffect, cancel := context.WithCancel(context.Background())
	go func(effectFn *goja.Callable) {
		for {
			select {
			case <-ctxEffect.Done():
				return
			case updatedState := <-subChan:
				if effectFn != nil {
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
								if callsInWindow >= MAX_EFFECT_CALLS_PER_WINDOW {
									c.mu.Unlock()
									c.fatalError(fmt.Errorf("effect %s exceeded rate limit with %d calls in %dms window", effectID, callsInWindow, EFFECT_TIME_WINDOW))
									return
								}

								// Track this call
								c.effectStack[effectID] = true
								c.effectCalls[effectID] = append(c.effectCalls[effectID], time.Now())
								c.mu.Unlock()

								if err := c.scheduler.Schedule(func() error {
									_, err := (*effectFn)(goja.Undefined())
									c.mu.Lock()
									c.effectStack[effectID] = false
									c.mu.Unlock()
									return err
								}); err != nil {
									c.logger.Error().Err(err).Msg("plugin: Error running effect")
								}
							}
						}
					}
				}
			}
		}
	}(&effectFn)

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

	go func() {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientEventHandlerTriggeredEvent, &payload) {
				if payload.HandlerName == handlerName {
					if err := c.scheduler.Schedule(func() error {
						// Trigger the callback with the event payload
						_, err := handlerCallback(goja.Undefined(), c.vm.ToValue(payload.Event))
						return err
					}); err != nil {
						c.logger.Error().Err(err).Str("handlerName", handlerName).Msg("plugin: Error running event handler")
					}
				}
			}
		}
	}()

	return goja.Undefined()
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

func safeEffectCall(fn *goja.Callable) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in effect: %v", r)
		}
	}()
	fmt.Println("safeEffect", fn)
	_, err = (*fn)(goja.Undefined())
	return
}

func (c *Context) cleanupOldEffectCalls(effectID string) {
	now := time.Now()
	window := time.Duration(EFFECT_TIME_WINDOW) * time.Millisecond
	var validCalls []time.Time

	for _, t := range c.effectCalls[effectID] {
		if now.Sub(t) <= window {
			validCalls = append(validCalls, t)
		}
	}

	c.effectCalls[effectID] = validCalls
}
