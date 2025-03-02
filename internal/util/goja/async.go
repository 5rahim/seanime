package goja_util

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// BindAwait binds the $await function to the Goja runtime.
// Hooks don't wait for promises to resolve, so $await is used to wrap a promise and wait for it to resolve.
func BindAwait(vm *goja.Runtime) {
	vm.Set("$await", func(promise goja.Value) (goja.Value, error) {
		if promise, ok := promise.Export().(*goja.Promise); ok {
			doneCh := make(chan struct{})

			// Wait for the promise to resolve
			go func() {
				for promise.State() == goja.PromiseStatePending {
					time.Sleep(10 * time.Millisecond)
				}
				close(doneCh)
			}()

			<-doneCh

			// If the promise is rejected, return the error
			if promise.State() == goja.PromiseStateRejected {
				err := promise.Result()
				return nil, fmt.Errorf("promise rejected: %v", err)
			}

			// If the promise is fulfilled, return the result
			res := promise.Result()
			return res, nil
		}

		// If the promise is not a Goja promise, return the value as is
		return promise, nil
	})
}
