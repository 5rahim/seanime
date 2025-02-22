package plugin_ui

import (
	"seanime/internal/events"
	"sync"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// UI registry, unique to a plugin and VM
type UI struct {
	extensionID    string
	context        *Context
	mu             sync.RWMutex
	vm             *goja.Runtime // VM executing the UI
	logger         *zerolog.Logger
	wsEventManager events.WSEventManagerInterface
	subscriber     *events.ClientEventSubscriber
}

func (u *UI) ClearInterrupt() {
	u.vm.ClearInterrupt()
	u.context.scheduler.Stop()
	if u.subscriber != nil {
		u.wsEventManager.UnsubscribeFromClientEvents("plugin-" + u.extensionID)
	}
}

func NewUI(logger *zerolog.Logger, vm *goja.Runtime, wsEventManager events.WSEventManagerInterface) *UI {
	return &UI{
		context:        NewContext(logger, vm),
		vm:             vm,
		logger:         logger,
		wsEventManager: wsEventManager,
	}
}

// Register a UI
// This is the main entry point for the UI
// - It is called once when the plugin is loaded and registers all necessary modules
func (u *UI) Register(callback string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Create a wrapper JavaScript function that calls the provided callback
	callback = `function(ctx) { return (` + callback + `).call(undefined, ctx); }`
	// Compile the callback into a Goja program
	// pr := goja.MustCompile("", "{("+callback+").apply(undefined, __ctx)}", true)

	// Subscribe the plugin to client events
	u.subscriber = u.wsEventManager.SubscribeToClientEvents("plugin-" + u.extensionID)

	contextObj := u.vm.NewObject()
	_ = contextObj.Set("newTray", u.context.jsNewTray)
	_ = contextObj.Set("state", u.context.jsState)
	_ = contextObj.Set("setTimeout", u.context.jsSetTimeout)
	_ = contextObj.Set("sleep", u.context.jsSleep)
	_ = contextObj.Set("setInterval", u.context.jsSetInterval)
	_ = contextObj.Set("effect", u.context.jsEffect)
	_ = contextObj.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return u.vm.ToValue(u.context.jsFetch(call))
	})
	_ = u.vm.Set("__ctx", contextObj)

	// Execute the callback
	_, err := u.vm.RunString(`(` + callback + `).call(undefined, __ctx)`)
	if err != nil {
		u.logger.Error().Err(err).Msg("Failed to run UI callback")
		return
	}
}
