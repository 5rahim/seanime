package plugin_ui

import (
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
}

func (u *UI) Unload() {
	u.logger.Debug().Msg("plugin: Stopping UI")
	u.mu.Lock()
	defer u.mu.Unlock()
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

	u.logger.Debug().Msg("plugin: Stopped UI")
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
	mLogger := options.Logger.With().Str("id", options.Extension.ID).Logger()
	ui := &UI{
		ext:            options.Extension,
		vm:             options.VM,
		logger:         &mLogger,
		wsEventManager: options.WSManager,
		appContext:     plugin.GlobalAppContext, // Get the app context from the global hook manager
		scheduler:      options.Scheduler,
	}
	ui.context = NewContext(ui)
	ui.context.scheduler.SetOnException(func(err error) {
		ui.context.HandleException(err)
	})

	return ui
}

// Register a UI
// This is the main entry point for the UI
// - It is called once when the plugin is loaded and registers all necessary modules
func (u *UI) Register(callback string) {
	defer util.HandlePanicInModuleThen("plugin_ui/Register", func() {
		u.logger.Error().Msg("plugin: Panic in Register")
	})

	u.mu.Lock()
	defer u.mu.Unlock()

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
			if event.Type == events.PluginEvent {

				if payload, ok := event.Payload.(map[string]interface{}); ok {
					clientEvent := NewClientPluginEvent(payload)
					// If the extension ID is not set, or the extension ID is the same as the current plugin, send the event to the listeners
					if clientEvent.ExtensionID == "" || clientEvent.ExtensionID == u.ext.ID {

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

						case ClientRenderCommandPaletteEvent: // Client wants to render the command palette
							u.context.commandPaletteManager.renderCommandPaletteScheduled()

						case ClientListCommandPalettesEvent: // Client wants to list all command palettes
							u.context.commandPaletteManager.sendInfoToClient()

						default:
							u.context.eventListeners.Range(func(key string, listener *EventListener) bool {
								//util.SpewMany("Event to listeners", event.Payload)
								if len(listener.ListenTo) > 0 {
									// Check if the event type is in the listener's list of event types
									for _, eventType := range listener.ListenTo {
										if eventType == clientEvent.Type {
											listener.Channel <- clientEvent // Only send the event to the listener if the event type is in the list
										}
									}
								} else {
									listener.Channel <- clientEvent
								}
								return true
							})
						}
					}

				}
			}
		}
		u.logger.Debug().Msg("plugin: Event goroutine stopped")
	}()

	u.context.createAndBindContextObject(u.vm)

	// Execute the callback
	_, err := u.vm.RunString(`(` + callback + `).call(undefined, __ctx)`)
	if err != nil {
		u.logger.Error().Err(err).Msg("plugin: Failed to register UI")
		return
	}

	// Send events to the client
	u.context.trayManager.renderTrayScheduled()
	u.context.trayManager.sendIconToClient()
	u.context.actionManager.renderAnimePageButtons()
	u.context.actionManager.renderAnimePageDropdownItems()
	u.context.actionManager.renderAnimeLibraryDropdownItems()
	u.context.actionManager.renderMangaPageButtons()
	u.context.actionManager.renderMediaCardContextMenuItems()
	u.context.commandPaletteManager.renderCommandPaletteScheduled()
	u.context.commandPaletteManager.sendInfoToClient()
}
