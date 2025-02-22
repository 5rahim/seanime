package plugin_ui

import (
	"seanime/internal/events"
	"seanime/internal/util"
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
}

func (u *UI) ClearInterrupt() {
	u.vm.ClearInterrupt()
	u.context.scheduler.Stop()
	if u.context.wsSubscriber != nil {
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
	u.context.wsSubscriber = u.wsEventManager.SubscribeToClientEvents("plugin-" + u.extensionID)

	// Listen for client events and send them to the event listeners
	go func() {
		for event := range u.context.wsSubscriber.Channel {
			//u.logger.Trace().Msgf("Received event %s", event.Type)
			if event.Type == events.PluginEvent {
				//u.logger.Trace().Msgf("Dispatching event %s to all listeners", event.Type)
				u.context.eventListeners.Range(func(key string, listener *EventListener) bool {
					util.SpewMany("Event to listeners", event.Payload)
					if payload, ok := event.Payload.(map[string]interface{}); ok {
						clientEvent := NewClientPluginEvent(payload)
						//u.logger.Trace().Msgf("Dispatching event %s to listener %s with payload %+v", event.Type, key, payload)
						// If the extension ID is not set, or the extension ID is the same as the current plugin, send the event to the listener
						if clientEvent.ExtensionID == "" || clientEvent.ExtensionID == u.extensionID {
							//u.logger.Trace().Msgf("Dispatching event %s to listener %s with payload %+v", event.Type, key, payload)
							listener.Channel <- clientEvent
						}
					}
					return true
				})
			}
		}
		// Close the listener channels when the subscriber is closed
		u.context.eventListeners.Range(func(key string, listener *EventListener) bool {
			close(listener.Channel)
			return true
		})
	}()

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

	// Webview object
	webviewObj := u.vm.NewObject()
	_ = contextObj.Set("webview", webviewObj)

	u.context.screenManager.bind(u.vm, contextObj)

	// Execute the callback
	_, err := u.vm.RunString(`(` + callback + `).call(undefined, __ctx)`)
	if err != nil {
		u.logger.Error().Err(err).Msg("Failed to run UI callback")
		return
	}
}
