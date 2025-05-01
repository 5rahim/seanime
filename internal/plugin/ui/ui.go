package plugin_ui

import (
	"errors"
	"fmt"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/plugin"
	"seanime/internal/util"
	goja_util "seanime/internal/util/goja"
	"sync"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

var (
	ErrTooManyExceptions = errors.New("plugin: Too many exceptions")
	ErrFatalError        = errors.New("plugin: Fatal error")
)

const (
	MaxExceptions              = 5    // Maximum number of exceptions that can be thrown before the UI is interrupted
	MaxConcurrentFetchRequests = 10   // Maximum number of concurrent fetch requests
	MaxEffectCallsPerWindow    = 100  // Maximum number of effect calls allowed in time window
	EffectTimeWindow           = 1000 // Time window in milliseconds to track effect calls
	StateUpdateBatchInterval   = 10   // Time in milliseconds to batch state updates
	UIUpdateRateLimit          = 120  // Time in milliseconds to rate limit UI updates
)

// UI registry, unique to a plugin and VM
type UI struct {
	ext            *extension.Extension
	context        *Context
	mu             sync.RWMutex
	vm             *goja.Runtime // VM executing the UI
	logger         *zerolog.Logger
	wsEventManager events.WSEventManagerInterface
	appContext     plugin.AppContext
	scheduler      *goja_util.Scheduler

	lastException string

	// Channel to signal the UI has been unloaded
	// This is used to interrupt the Plugin when the UI is stopped
	destroyedCh chan struct{}
	destroyed   bool
}

type NewUIOptions struct {
	Logger    *zerolog.Logger
	VM        *goja.Runtime
	WSManager events.WSEventManagerInterface
	Database  *db.Database
	Scheduler *goja_util.Scheduler
	Extension *extension.Extension
}

func NewUI(options NewUIOptions) *UI {
	ui := &UI{
		ext:            options.Extension,
		vm:             options.VM,
		logger:         options.Logger,
		wsEventManager: options.WSManager,
		appContext:     plugin.GlobalAppContext, // Get the app context from the global hook manager
		scheduler:      options.Scheduler,
		destroyedCh:    make(chan struct{}),
	}
	ui.context = NewContext(ui)
	ui.context.scheduler.SetOnException(func(err error) {
		ui.logger.Error().Err(err).Msg("plugin: Encountered exception in asynchronous task")
		ui.context.handleException(err)
	})

	return ui
}

// Called by the Plugin when it's being unloaded
func (u *UI) Unload(signalDestroyed bool) {
	u.logger.Debug().Msg("plugin: Stopping UI")

	u.UnloadFromInside(signalDestroyed)

	u.logger.Debug().Msg("plugin: Stopped UI")
}

// UnloadFromInside is called by the UI module itself when it's being unloaded
func (u *UI) UnloadFromInside(signalDestroyed bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.destroyed {
		return
	}
	// Stop the VM
	u.vm.ClearInterrupt()
	// Unsubscribe from client all events
	if u.context.wsSubscriber != nil {
		u.wsEventManager.UnsubscribeFromClientEvents("plugin-" + u.ext.ID)
	}
	// Clean up the context (all modules)
	if u.context != nil {
		u.context.Stop()
	}

	// Send the plugin unloaded event to the client
	u.wsEventManager.SendEvent(events.PluginUnloaded, u.ext.ID)

	if signalDestroyed {
		u.signalDestroyed()
	}
}

// Destroyed returns a channel that is closed when the UI is destroyed
func (u *UI) Destroyed() <-chan struct{} {
	return u.destroyedCh
}

// signalDestroyed tells the plugin that the UI has been destroyed.
// This is used to interrupt the Plugin when the UI is stopped
// TODO: FIX
func (u *UI) signalDestroyed() {
	defer func() {
		if r := recover(); r != nil {
			u.logger.Error().Msgf("plugin: Panic in signalDestroyed: %v", r)
		}
	}()

	if u.destroyed {
		return
	}
	u.destroyed = true
	close(u.destroyedCh)
}

// Register a UI
// This is the main entry point for the UI
// - It is called once when the plugin is loaded and registers all necessary modules
func (u *UI) Register(callback string) error {
	defer util.HandlePanicInModuleThen("plugin_ui/Register", func() {
		u.logger.Error().Msg("plugin: Panic in Register")
	})

	u.mu.Lock()

	// Create a wrapper JavaScript function that calls the provided callback
	callback = `function(ctx) { return (` + callback + `).call(undefined, ctx); }`
	// Compile the callback into a Goja program
	// pr := goja.MustCompile("", "{("+callback+").apply(undefined, __ctx)}", true)

	// Subscribe the plugin to client events
	u.context.wsSubscriber = u.wsEventManager.SubscribeToClientEvents("plugin-" + u.ext.ID)

	u.logger.Debug().Msg("plugin: Registering UI")

	// Listen for client events and send them to the event listeners
	go func() {
		for event := range u.context.wsSubscriber.Channel {
			//u.logger.Trace().Msgf("Received event %s", event.Type)
			u.HandleWSEvent(event)
		}
		u.logger.Debug().Msg("plugin: Event goroutine stopped")
	}()

	u.context.createAndBindContextObject(u.vm)

	// Execute the callback
	_, err := u.vm.RunString(`(` + callback + `).call(undefined, __ctx)`)
	if err != nil {
		u.mu.Unlock()
		u.logger.Error().Err(err).Msg("plugin: Encountered exception in UI handler, unloading plugin")
		u.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("plugin(%s): Encountered exception in UI handler: %s", u.ext.ID, err.Error()))
		u.wsEventManager.SendEvent(events.ConsoleLog, fmt.Sprintf("plugin(%s): Encountered exception in UI handler: %s", u.ext.ID, err.Error()))
		// Unload the UI and signal the Plugin that it's been terminated
		u.UnloadFromInside(true)
		return fmt.Errorf("plugin: Encountered exception in UI handler: %w", err)
	}

	// Send events to the client
	u.context.trayManager.renderTrayScheduled()
	u.context.trayManager.sendIconToClient()
	u.context.actionManager.renderAnimePageButtons()
	u.context.actionManager.renderAnimePageDropdownItems()
	u.context.actionManager.renderAnimeLibraryDropdownItems()
	u.context.actionManager.renderMangaPageButtons()
	u.context.actionManager.renderMediaCardContextMenuItems()
	u.context.actionManager.renderEpisodeCardContextMenuItems()
	u.context.actionManager.renderEpisodeGridItemMenuItems()
	u.context.commandPaletteManager.renderCommandPaletteScheduled()
	u.context.commandPaletteManager.sendInfoToClient()

	u.wsEventManager.SendEvent(events.PluginLoaded, u.ext.ID)

	u.mu.Unlock()
	return nil
}

// Add this new type to handle batched events from the client
type BatchedClientEvents struct {
	Events []map[string]interface{} `json:"events"`
}

// HandleWSEvent handles a websocket event from the client
func (u *UI) HandleWSEvent(event *events.WebsocketClientEvent) {
	defer util.HandlePanicInModuleThen("plugin/HandleWSEvent", func() {})

	u.mu.RLock()
	defer u.mu.RUnlock()

	// Ignore if UI is destroyed
	if u.destroyed {
		return
	}

	if event.Type == events.PluginEvent {
		// Extract the event payload
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			u.logger.Error().Str("payload", fmt.Sprintf("%+v", event.Payload)).Msg("plugin/ui: Failed to parse plugin event payload")
			return
		}

		// Check if this is a batch event
		eventType, _ := payload["type"].(string)
		if eventType == "client:batch-events" {
			u.handleBatchedClientEvents(event.ClientID, payload)
			return
		}

		// Process normal event
		clientEvent := NewClientPluginEvent(payload)
		if clientEvent == nil {
			u.logger.Error().Interface("payload", payload).Msg("plugin/ui: Failed to create client plugin event")
			return
		}

		// If the event is for this plugin
		if clientEvent.ExtensionID == u.ext.ID || clientEvent.ExtensionID == "" {
			// Process the event based on type
			u.dispatchClientEvent(clientEvent)
		}
	}
}

// dispatchClientEvent handles a client event based on its type
func (u *UI) dispatchClientEvent(clientEvent *ClientPluginEvent) {
	switch clientEvent.Type {
	case ClientRenderTrayEvent: // Client wants to render the tray
		u.context.trayManager.renderTrayScheduled()

	case ClientListTrayIconsEvent: // Client wants to list all tray icons from all plugins
		u.context.trayManager.sendIconToClient()

	case ClientActionRenderAnimePageButtonsEvent: // Client wants to update the anime page buttons
		u.context.actionManager.renderAnimePageButtons()

	case ClientActionRenderAnimePageDropdownItemsEvent: // Client wants to update the anime page dropdown items
		u.context.actionManager.renderAnimePageDropdownItems()

	case ClientActionRenderAnimeLibraryDropdownItemsEvent: // Client wants to update the anime library dropdown items
		u.context.actionManager.renderAnimeLibraryDropdownItems()

	case ClientActionRenderMangaPageButtonsEvent: // Client wants to update the manga page buttons
		u.context.actionManager.renderMangaPageButtons()

	case ClientActionRenderMediaCardContextMenuItemsEvent: // Client wants to update the media card context menu items
		u.context.actionManager.renderMediaCardContextMenuItems()

	case ClientActionRenderEpisodeCardContextMenuItemsEvent: // Client wants to update the episode card context menu items
		u.context.actionManager.renderEpisodeCardContextMenuItems()

	case ClientActionRenderEpisodeGridItemMenuItemsEvent: // Client wants to update the episode grid item menu items
		u.context.actionManager.renderEpisodeGridItemMenuItems()

	case ClientRenderCommandPaletteEvent: // Client wants to render the command palette
		u.context.commandPaletteManager.renderCommandPaletteScheduled()

	case ClientListCommandPalettesEvent: // Client wants to list all command palettes
		u.context.commandPaletteManager.sendInfoToClient()

	default:
		// Send to registered event listeners
		eventListeners, ok := u.context.eventBus.Get(clientEvent.Type)
		if !ok {
			return
		}
		eventListeners.Range(func(key string, listener *EventListener) bool {
			listener.Send(clientEvent)
			return true
		})
	}
}

// handleBatchedClientEvents processes a batch of client events
func (u *UI) handleBatchedClientEvents(clientID string, payload map[string]interface{}) {
	if eventPayload, ok := payload["payload"].(map[string]interface{}); ok {
		if eventsRaw, ok := eventPayload["events"].([]interface{}); ok {
			// Process each event in the batch
			for _, eventRaw := range eventsRaw {
				if eventMap, ok := eventRaw.(map[string]interface{}); ok {
					// Create a synthetic event object
					syntheticPayload := map[string]interface{}{
						"type":        eventMap["type"],
						"extensionId": eventMap["extensionId"],
						"payload":     eventMap["payload"],
					}

					// Create and dispatch the event
					clientEvent := NewClientPluginEvent(syntheticPayload)
					if clientEvent == nil {
						u.logger.Error().Interface("payload", syntheticPayload).Msg("plugin/ui: Failed to create client plugin event from batch")
						continue
					}

					// If the event is for this plugin
					if clientEvent.ExtensionID == u.ext.ID || clientEvent.ExtensionID == "" {
						// Process the event
						u.dispatchClientEvent(clientEvent)
					}
				}
			}
		}
	}
}
