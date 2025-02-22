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

type Context struct {
	extensionID    string
	logger         *zerolog.Logger
	wsEventManager events.WSEventManagerInterface
	mu             sync.RWMutex

	vm               *goja.Runtime
	states           *result.Map[string, *State]
	stateSubscribers []chan *State
	scheduler        *Scheduler
	wsSubscriber     *events.ClientEventSubscriber
	eventListeners   *result.Map[string, *EventListener] // Event listeners added

	webviewManager *WebviewManager
	screenManager  *ScreenManager
	trayManager    *TrayManager
}

type State struct {
	ID    string
	Value goja.Value
}

type EventListener struct {
	ID      string
	Channel chan *ClientPluginEvent // Channel for the event payload
}

func NewContext(extensionID string, logger *zerolog.Logger, vm *goja.Runtime, wsEventManager events.WSEventManagerInterface) *Context {
	ret := &Context{
		extensionID:      extensionID,
		logger:           logger,
		vm:               vm,
		states:           result.NewResultMap[string, *State](),
		stateSubscribers: make([]chan *State, 0),
		eventListeners:   result.NewResultMap[string, *EventListener](),
		scheduler:        NewScheduler(),
		wsEventManager:   wsEventManager,
	}

	ret.trayManager = NewTrayManager(ret)
	ret.webviewManager = NewWebviewManager(ret)
	ret.screenManager = NewScreenManager(ret)

	return ret
}

func (c *Context) RegisterEventListener() *EventListener {
	id := uuid.New().String()
	listener := &EventListener{
		ID:      id,
		Channel: make(chan *ClientPluginEvent),
	}
	c.eventListeners.Set(id, listener)
	return listener
}

func (c *Context) SendEventToClient(eventType ServerEventType, payload interface{}) {
	c.wsEventManager.SendEvent(string(events.PluginEvent), &ServerPluginEvent{
		ExtensionID: c.extensionID,
		Type:        eventType,
		Payload:     payload,
	})
}

func (c *Context) PrintState() {
	c.states.Range(func(key string, state *State) bool {
		c.logger.Info().Msgf("State %s = %+v", key, state.Value)
		return true
	})
}

// jsState
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
		panic(err)
	}
	jsDynamicDefFunc, ok := goja.AssertFunction(jsDynamicDefFuncValue)
	if !ok {
		panic("dynamic definition is not a function")
	}

	jsDynamicState, err := jsDynamicDefFunc(goja.Undefined(), stateObj, jsGetStateVal, jsSetStateVal)
	if err != nil {
		panic(err)
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
		panic(c.vm.NewTypeError("setTimeout requires a function and a delay"))
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		panic(c.vm.NewTypeError("setTimeout requires a function"))
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		panic(c.vm.NewTypeError("delay must be a number"))
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
				c.logger.Error().Err(err).Msg("error running timeout callback")
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
		panic(c.vm.NewTypeError("setInterval requires a function and a delay"))
	}

	fnValue := call.Argument(0)
	delayValue := call.Argument(1)

	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		panic(c.vm.NewTypeError("setInterval requires a function"))
	}

	delay, ok := delayValue.Export().(int64)
	if !ok {
		panic(c.vm.NewTypeError("delay must be a number"))
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
					c.logger.Error().Err(err).Msg("error running interval callback")
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
		panic(c.vm.NewTypeError("sleep requires a delay"))
	}

	delayValue := call.Argument(0)
	delay, ok := delayValue.Export().(int64)
	if !ok {
		panic(c.vm.NewTypeError("delay must be a number"))
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
		panic(c.vm.NewTypeError("effect requires a function and an array of dependencies"))
	}

	effectFn, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		panic(c.vm.NewTypeError("first argument to effect must be a function"))
	}

	depsObj, ok := call.Argument(1).(*goja.Object)
	if !ok {
		panic(c.vm.NewTypeError("second argument to effect must be an array"))
	}

	// Prepare dependencies and their old values
	lengthVal := depsObj.Get("length")
	depsLen := int(lengthVal.ToInteger())
	deps := make([]*goja.Object, depsLen)
	oldValues := make([]goja.Value, depsLen)
	dropIDs := make([]string, depsLen) // to store state IDs of dependencies
	for i := 0; i < depsLen; i++ {
		depVal := depsObj.Get(fmt.Sprintf("%d", i))
		depObj, ok := depVal.(*goja.Object)
		if !ok {
			panic(c.vm.NewTypeError("dependency is not an object"))
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
								if err := c.scheduler.Schedule(func() error {
									_, err := (*effectFn)(goja.Undefined())
									return err
								}); err != nil {
									c.logger.Error().Err(err).Msg("error running effect")
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
		return goja.Undefined()
	}

	return c.vm.ToValue(cancelFunc)
}

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
