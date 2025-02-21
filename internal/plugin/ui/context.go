package plugin_ui

import (
	"context"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Context struct {
	logger *zerolog.Logger
	vm     *goja.Runtime
	states *result.Map[string, *State]

	mu sync.RWMutex
}

type State struct {
	ID    string
	Value goja.Value
}

func NewContext(logger *zerolog.Logger, vm *goja.Runtime) *Context {
	ret := &Context{
		logger: logger,
		vm:     vm,
		states: result.NewResultMap[string, *State](),
	}

	return ret
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
				// Get the current state
				prevState, ok := c.states.Get(id)
				if ok {
					// Call the callback with the current state
					newVal, _ := callback(goja.Undefined(), prevState.Value)
					// Set the new state
					c.states.Set(id, &State{
						ID:    id,
						Value: newVal,
					})
				}
			} else {
				// Set the new state
				c.states.Set(id, &State{
					ID:    id,
					Value: arg,
				})
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
	// Use single assignment from AssertFunction
	jsDynamicDefFunc, ok := goja.AssertFunction(jsDynamicDefFuncValue)
	if !ok {
		panic("dynamic definition is not a function")
	}

	jsDynamicState, err := jsDynamicDefFunc(goja.Undefined(), stateObj, jsGetStateVal, jsSetStateVal)
	if err != nil {
		panic(err)
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

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delay) * time.Millisecond):
			fn(goja.Undefined())
		}
	}()

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	}

	cancelFuncVal := c.vm.ToValue(cancelFunc)

	return cancelFuncVal
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
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(delay) * time.Millisecond):
				fn(goja.Undefined())
			}
		}
	}()

	cancelFunc := func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	}

	cancelFuncVal := c.vm.ToValue(cancelFunc)

	return cancelFuncVal
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
