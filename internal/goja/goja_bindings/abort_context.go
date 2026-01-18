package goja_bindings

import (
	"context"
	gojautil "seanime/internal/util/goja"
	"sync"

	"github.com/dop251/goja"
)

type AbortSignal struct {
	mu        sync.RWMutex
	aborted   bool
	reason    goja.Value
	ctx       context.Context
	cancel    context.CancelFunc
	listeners []goja.Callable
	vm        *goja.Runtime
	scheduler *gojautil.Scheduler
}

func newAbortSignal(vm *goja.Runtime, scheduler *gojautil.Scheduler) *AbortSignal {
	ctx, cancel := context.WithCancel(context.Background())
	return &AbortSignal{
		ctx:       ctx,
		cancel:    cancel,
		vm:        vm,
		reason:    goja.Undefined(),
		scheduler: scheduler,
	}
}

func (s *AbortSignal) abort(reason goja.Value) {
	s.mu.Lock()
	if s.aborted {
		s.mu.Unlock()
		return
	}

	s.aborted = true
	if reason == nil || goja.IsUndefined(reason) {
		s.reason = s.vm.NewGoError(func() error { return context.Canceled }())
	} else {
		s.reason = reason
	}

	// create a snapshot of listeners to call outside the lock
	listeners := make([]goja.Callable, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	s.cancel()

	// fire listeners
	for _, fn := range listeners {
		s.scheduler.ScheduleAsync(func() error {
			_, _ = fn(goja.Undefined())
			return nil
		})
	}
}

func (s *AbortSignal) toObject() *goja.Object {
	obj := s.vm.NewObject()

	_ = obj.DefineAccessorProperty("aborted", s.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.vm.ToValue(s.aborted)
	}), nil, goja.FLAG_TRUE, goja.FLAG_TRUE)

	_ = obj.DefineAccessorProperty("reason", s.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.reason
	}), nil, goja.FLAG_TRUE, goja.FLAG_TRUE)

	_ = obj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		eventName := call.Argument(0).String()
		callback, ok := goja.AssertFunction(call.Argument(1))

		if eventName == "abort" && ok {
			s.mu.Lock()
			// If already aborted, the spec says we should fire immediately (or schedule it)
			if s.aborted {
				s.mu.Unlock()
				s.scheduler.ScheduleAsync(func() error {
					_, _ = callback(goja.Undefined())
					return nil
				})
				return goja.Undefined()
			}
			s.listeners = append(s.listeners, callback)
			s.mu.Unlock()
		}
		return goja.Undefined()
	})

	// Internal helper for fetch bindings
	_ = obj.Set("_getContext", func(call goja.FunctionCall) goja.Value {
		return s.vm.ToValue(s.ctx)
	})

	return obj
}

// BindAbortContext binds the AbortContext to the VM
func BindAbortContext(vm *goja.Runtime, scheduler *gojautil.Scheduler) {
	_ = vm.Set("AbortContext", func(call goja.ConstructorCall) *goja.Object {
		signal := newAbortSignal(vm, scheduler)
		instance := vm.NewObject()

		_ = instance.Set("signal", signal.toObject())

		_ = instance.Set("abort", func(call goja.FunctionCall) goja.Value {
			var reason goja.Value
			if len(call.Arguments) > 0 {
				reason = call.Argument(0)
			}
			signal.abort(reason)
			return goja.Undefined()
		})

		return instance
	})
}
