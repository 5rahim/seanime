package goja_bindings

import (
	"seanime/internal/util"
	gojautil "seanime/internal/util/goja"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestAbortContext(t *testing.T) {
	vm := goja.New()
	BindAbortContext(vm, gojautil.NewScheduler())

	t.Run("AbortContext basic functionality", func(t *testing.T) {
		script := `
			const controller = new AbortContext();
			const signal = controller.signal;
			
			let aborted = signal.aborted;
			controller.abort();
			
			({
				initialAborted: aborted,
				finalAborted: signal.aborted
			})
		`

		val, err := vm.RunString(script)
		assert.NoError(t, err)

		obj := val.ToObject(vm)
		util.Spew(obj.Export())
		initialAborted := obj.Get("initialAborted").ToBoolean()
		finalAborted := obj.Get("finalAborted").ToBoolean()

		assert.False(t, initialAborted, "Signal should not be aborted initially")
		assert.True(t, finalAborted, "Signal should be aborted after controller.abort()")
	})

	//t.Run("AbortSignal event listener", func(t *testing.T) {
	//	script := `
	//		const controller = new AbortContext();
	//		const signal = controller.signal;
	//
	//		let eventFired = false;
	//		signal.addEventListener('abort', () => {
	//			eventFired = true;
	//		});
	//
	//		controller.abort();
	//
	//		eventFired
	//	`
	//
	//	val, err := vm.RunString(script)
	//	require.NoError(t, err)
	//	assert.True(t, val.ToBoolean(), "Abort event should fire")
	//})

	t.Run("AbortSignal with reason", func(t *testing.T) {
		script := `
			const controller = new AbortContext();
			const signal = controller.signal;
			
			controller.abort('Custom reason');
			
			signal.reason
		`

		val, err := vm.RunString(script)
		assert.NoError(t, err)
		assert.Equal(t, "Custom reason", val.String())
	})
}

func TestAbortContextWithFetch(t *testing.T) {
	vm := goja.New()
	BindAbortContext(vm, gojautil.NewScheduler())
	fetch := BindFetch(vm)
	defer fetch.Close()

	// Start the response channel handler
	go func() {
		for fn := range fetch.ResponseChannel() {
			fn()
		}
	}()

	t.Run("Abort fetch immediately", func(t *testing.T) {
		script := `
			const controller = new AbortContext();
			controller.abort();
			
			fetch('https://api.github.com/users/github', {
				signal: controller.signal
			})
		`

		val, err := vm.RunString(script)
		assert.NoError(t, err)

		promise, ok := val.Export().(*goja.Promise)
		assert.True(t, ok, "fetch should return a promise")

		time.Sleep(100 * time.Millisecond)

		// Promise should be rejected
		assert.Equal(t, goja.PromiseStateRejected, promise.State())
	})
}
