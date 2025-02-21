package plugin_ui

import (
	"sync"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// UI registry, unique to a plugin and VM
type UI struct {
	extensionID string
	context     *Context
	mu          sync.RWMutex
	vm          *goja.Runtime // VM executing the UI
	logger      *zerolog.Logger
}

func (u *UI) GetVM() *goja.Runtime {
	return u.vm
}

func NewUI(logger *zerolog.Logger, vm *goja.Runtime) *UI {
	return &UI{
		context: NewContext(logger, vm),
		vm:      vm,
		logger:  logger,
	}
}

// Register a UI
// This is the main entry point for the UI
// - It is called once when the plugin is loaded and registers all necessary things
func (u *UI) Register(callback string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Create a wrapper JavaScript function that calls the provided callback
	callback = `function(ctx) { return (` + callback + `).call(undefined, ctx); }`
	// Compile the callback into a Goja program
	// pr := goja.MustCompile("", "{("+callback+").apply(undefined, __ctx)}", true)

	contextObj := u.vm.NewObject()
	_ = contextObj.Set("newTray", u.context.jsNewTray)
	_ = contextObj.Set("state", u.context.jsState)
	_ = contextObj.Set("setTimeout", u.context.jsSetTimeout)
	_ = contextObj.Set("sleep", u.context.jsSleep)
	_ = contextObj.Set("setInterval", u.context.jsSetInterval)
	_ = u.vm.Set("__ctx", contextObj)

	// Execute the callback
	_, err := u.vm.RunString(`(` + callback + `).call(undefined, __ctx)`)
	if err != nil {
		u.logger.Error().Err(err).Msg("Failed to run UI callback")
		return
	}
}
