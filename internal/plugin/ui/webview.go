package plugin_ui

import (
	"seanime/internal/util/result"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type WebviewSlot string

const (
	// The webview has its own screen/page
	ScreenSlot WebviewSlot = "screen"
	// The webview is rendered above all elements
	FixedSlot WebviewSlot = "fixed"
	// The webview is rendered after the home screen toolbar
	AfterHomeScreenToolbarSlot      WebviewSlot = "after-home-screen-toolbar"
	HomeScreenBottomSlot            WebviewSlot = "home-screen-bottom"
	ScheduleScreenTopSlot           WebviewSlot = "schedule-screen-top"
	ScheduleScreenBottomSlot        WebviewSlot = "schedule-screen-bottom"
	AnimeEntryScreenBottomSlot      WebviewSlot = "anime-screen-bottom"
	AfterAnimeEntryEpisodeListSlot  WebviewSlot = "after-anime-entry-episode-list"
	BeforeAnimeEntryEpisodeListSlot WebviewSlot = "before-anime-entry-episode-list"
	MangaScreenBottomSlot           WebviewSlot = "manga-screen-bottom"
	MangaEntryScreenBottomSlot      WebviewSlot = "manga-entry-screen-bottom"
	AfterMangaEntryChapterListSlot  WebviewSlot = "after-manga-entry-chapter-list"
	AfterDiscoverScreenHeaderSlot   WebviewSlot = "after-discover-screen-header"
	AfterMediaEntryDetailsSlot      WebviewSlot = "after-media-entry-details"
	AfterMediaEntryFormSlot         WebviewSlot = "after-media-entry-form"
)

var WebviewSlots = []WebviewSlot{
	ScreenSlot, FixedSlot, AfterHomeScreenToolbarSlot, HomeScreenBottomSlot, ScheduleScreenTopSlot, ScheduleScreenBottomSlot,
	AnimeEntryScreenBottomSlot, AfterAnimeEntryEpisodeListSlot, BeforeAnimeEntryEpisodeListSlot, MangaScreenBottomSlot,
	MangaEntryScreenBottomSlot, AfterMangaEntryChapterListSlot, AfterDiscoverScreenHeaderSlot, AfterMediaEntryDetailsSlot, AfterMediaEntryFormSlot,
}

type WebviewManager struct {
	ctx         *Context
	webviews    *result.Map[WebviewSlot, *Webview]
	updateMutex sync.Mutex

	componentManager *ComponentManager
}

type WebviewChannel struct {
	webview      *Webview
	syncedStates *result.Map[string, string] // map[key]stateID
}

func NewWebviewManager(ctx *Context) *WebviewManager {
	return &WebviewManager{
		ctx:              ctx,
		webviews:         result.NewMap[WebviewSlot, *Webview](),
		componentManager: &ComponentManager{ctx: ctx},
	}
}

//// renderWebviewScheduled renders the new component tree of the webview at the given slot.
//// This function is unsafe because it is not thread-safe and should be scheduled.
//func (t *WebviewManager) renderWebviewScheduled(slots ...WebviewSlot) {
//	t.updateMutex.Lock()
//	defer t.updateMutex.Unlock()
//
//	shouldMount := false
//
//	// renderWebviewScheduled can be called without slots (when states are updated)
//	if len(slots) == 0 {
//		slots = WebviewSlots
//	} else {
//		// Set the webview as mounted if renderWebviewScheduled has been called WITH a slot
//		shouldMount = true
//	}
//
//	for _, slot := range slots {
//		webview, ok := t.webviews.Get(slot)
//		if !ok {
//			continue
//		}
//
//		if webview.renderFunc == nil {
//			continue
//		}
//
//		// Make sure it's mounted
//		if shouldMount && !webview.mounted.Load() {
//			webview.mounted.Store(true)
//		}
//
//		// Ignore if it's not mounted
//		// renderWebviewScheduled can be called without slots, in this case it will render already mounted webviews
//		if !webview.mounted.Load() {
//			continue
//		}
//
//		webview.lastUpdatedAt = time.Now()
//
//		t.ctx.scheduler.ScheduleAsync(func() error {
//			newComponents, err := t.componentManager.renderComponents(webview.renderFunc)
//			if err != nil {
//				t.ctx.logger.Error().Err(err).Msg("plugin: Failed to render webview")
//				t.ctx.handleException(err)
//				return nil
//			}
//
//			// t.ctx.logger.Trace().Msg("plugin: Sending webview update to client")
//			// Send the JSON value to the client
//			t.ctx.SendEventToClient(ServerWebviewUpdatedEvent, ServerWebviewUpdatedEventPayload{
//				Slot:       string(slot),
//				Components: newComponents,
//			})
//			return nil
//		})
//	}
//}

// renderWebviewIframe
func (t *WebviewManager) renderWebviewIframe(slots ...WebviewSlot) {
	t.updateMutex.Lock()
	defer t.updateMutex.Unlock()

	// renderWebviewScheduled can be called without slots (when states are updated)
	if len(slots) == 0 {
		slots = WebviewSlots
	}

	for _, slot := range slots {
		webview, ok := t.webviews.Get(slot)
		if !ok {
			continue
		}

		if webview.contentFunc == nil {
			continue
		}

		webview.lastUpdatedAt = time.Now()

		t.ctx.scheduler.ScheduleAsync(func() error {
			str := t.componentManager.executeContentFunc(webview.contentFunc)

			t.ctx.SendEventToClient(ServerWebviewIframeEvent, ServerWebviewIframeEventPayload{
				Slot:    string(slot),
				Content: str,
				ID:      webview.id,
				Options: webview.options,
			})
			return nil
		})
	}
}

func (t *WebviewManager) renderWebviewSidebar() {
	t.webviews.Range(func(_ WebviewSlot, webview *Webview) bool {
		if !webview.options.hasSidebar() {
			return true
		}
		if containsDangerousHTML(webview.options.Sidebar.Icon) {
			t.ctx.logger.Warn().Msg("plugin: Sidebar icon contains dangerous HTML, it will not be rendered")
			return true
		}
		t.ctx.SendEventToClient(ServerWebviewSidebarEvent, ServerWebviewSidebarEventPayload{
			Label: webview.options.Sidebar.Label,
			Icon:  webview.options.Sidebar.Icon,
		})
		return true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Webview
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Webview struct {
	webviewManager *WebviewManager
	id             string
	channel        *WebviewChannel

	// Server-side rendering
	renderFunc func(goja.FunctionCall) goja.Value
	mounted    atomic.Bool

	// HTML content rendering (iframe)
	contentFunc func(goja.FunctionCall) goja.Value

	lastUpdatedAt time.Time // for renderFunc

	Slot WebviewSlot

	// Styling and window management options
	options *WebviewOptions
}

type WebviewOptions struct {
	// Styling
	ClassName string `json:"className,omitempty"`
	Style     string `json:"style,omitempty"`
	Width     string `json:"width,omitempty"`
	Height    string `json:"height,omitempty"`
	MaxWidth  string `json:"maxWidth,omitempty"`
	MaxHeight string `json:"maxHeight,omitempty"`
	ZIndex    int    `json:"zIndex,omitempty"`

	Window WebviewWindowOptions `json:"window,omitempty"`

	// Responsiveness
	AutoHeight bool `json:"autoHeight,omitempty"`
	FullWidth  bool `json:"fullWidth,omitempty"`

	Sidebar WebviewSidebarOptions `json:"sidebar,omitempty"`

	Hidden bool `json:"hidden,omitempty"`
}

type WebviewWindowOptions struct {
	// Window management
	Draggable bool `json:"draggable,omitempty"`
	//Resizable bool `json:"resizable,omitempty"`
	//Closable  bool `json:"closable,omitempty"`
	DefaultX        int    `json:"defaultX,omitempty"`
	DefaultY        int    `json:"defaultY,omitempty"`
	Frameless       bool   `json:"frameless,omitempty"`
	DefaultPosition string `json:"defaultPosition,omitempty"`
}

type WebviewSidebarOptions struct {
	Label string `json:"label,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

func (o *WebviewOptions) hasSidebar() bool {
	return o.Sidebar.Label != "" || o.Sidebar.Icon != ""
}

// jsNewWebview
//
//	Example:
//	const webview = ctx.newWebview({ slot: "" })
func (t *WebviewManager) jsNewWebview(call goja.FunctionCall) goja.Value {
	webview := &Webview{
		webviewManager: t,
		id:             uuid.NewString(),
		renderFunc:     nil,
		contentFunc:    nil,
		lastUpdatedAt:  time.Time{},
		Slot:           "",
		channel: &WebviewChannel{
			syncedStates: result.NewMap[string, string](),
		},
		options: &WebviewOptions{},
	}

	// Set the webview reference in the channel
	webview.channel.webview = webview

	props := call.Arguments
	if len(props) > 0 {
		propsObj := props[0].Export().(map[string]interface{})
		if propsObj["slot"] != nil {
			s, _ := propsObj["slot"].(string)
			webview.Slot = WebviewSlot(s)
		}

		// Parse styling options
		if propsObj["className"] != nil {
			webview.options.ClassName, _ = propsObj["className"].(string)
		}
		if propsObj["style"] != nil {
			webview.options.Style, _ = propsObj["style"].(string)
		}
		if propsObj["width"] != nil {
			webview.options.Width, _ = propsObj["width"].(string)
		}
		if propsObj["height"] != nil {
			webview.options.Height, _ = propsObj["height"].(string)
		}
		if propsObj["maxWidth"] != nil {
			webview.options.MaxWidth, _ = propsObj["maxWidth"].(string)
		}
		if propsObj["maxHeight"] != nil {
			webview.options.MaxHeight, _ = propsObj["maxHeight"].(string)
		}
		if propsObj["zIndex"] != nil {
			if zIdx, ok := propsObj["zIndex"].(int64); ok {
				webview.options.ZIndex = int(zIdx)
			} else if zIdx, ok := propsObj["zIndex"].(float64); ok {
				webview.options.ZIndex = int(zIdx)
			}
		}

		if propsObj["window"] != nil {
			if windowObj, ok := propsObj["window"].(map[string]interface{}); ok {
				// Parse window management options
				if windowObj["draggable"] != nil {
					webview.options.Window.Draggable, _ = windowObj["draggable"].(bool)
				}
				if windowObj["frameless"] != nil {
					webview.options.Window.Frameless, _ = windowObj["frameless"].(bool)
				}
				if windowObj["defaultPosition"] != nil {
					webview.options.Window.DefaultPosition, _ = windowObj["defaultPosition"].(string)
				}
				//if windowObj["resizable"] != nil {
				//	webview.options.Resizable, _ = windowObj["resizable"].(bool)
				//}
				//if windowObj["closable"] != nil {
				//	webview.options.Closable, _ = windowObj["closable"].(bool)
				//}
				if windowObj["defaultX"] != nil {
					if x, ok := windowObj["defaultX"].(int64); ok {
						webview.options.Window.DefaultX = int(x)
					} else if x, ok := windowObj["defaultX"].(float64); ok {
						webview.options.Window.DefaultX = int(x)
					}
				}
				if windowObj["defaultY"] != nil {
					if y, ok := windowObj["defaultY"].(int64); ok {
						webview.options.Window.DefaultY = int(y)
					} else if y, ok := windowObj["defaultY"].(float64); ok {
						webview.options.Window.DefaultY = int(y)
					}
				}
			}
		}

		if propsObj["sidebar"] != nil {
			if sidebarObj, ok := propsObj["sidebar"].(map[string]interface{}); ok {
				if sidebarObj["label"] != nil {
					webview.options.Sidebar.Label, _ = sidebarObj["label"].(string)
				}
				if sidebarObj["icon"] != nil {
					webview.options.Sidebar.Icon, _ = sidebarObj["icon"].(string)
				}
			}
		}

		// Parse responsiveness options
		if propsObj["autoHeight"] != nil {
			webview.options.AutoHeight, _ = propsObj["autoHeight"].(bool)
		}
		if propsObj["fullWidth"] != nil {
			webview.options.FullWidth, _ = propsObj["fullWidth"].(bool)
		}
		if propsObj["hidden"] != nil {
			webview.options.Hidden, _ = propsObj["hidden"].(bool)
		}
	}

	if webview.Slot == "" {
		t.ctx.handleTypeError("newWebview requires a slot name")
	}

	t.webviews.Set(webview.Slot, webview)

	// Create a new webview object
	webviewObj := t.ctx.vm.NewObject()
	_ = webviewObj.Set("render", webview.jsRender)
	_ = webviewObj.Set("setContent", webview.jsSetContent)
	_ = webviewObj.Set("update", webview.jsUpdate)
	_ = webviewObj.Set("setOptions", webview.jsSetOptions)
	_ = webviewObj.Set("close", webview.jsClose)
	_ = webviewObj.Set("show", webview.jsShow)
	_ = webviewObj.Set("hide", webview.jsHide)
	_ = webviewObj.Set("onMount", webview.jsOnMount)
	_ = webviewObj.Set("onLoad", webview.jsOnLoad)
	_ = webviewObj.Set("onUnmount", webview.jsOnUnmount)
	_ = webviewObj.Set("getScreenPath", webview.jsGetScreenPath)
	_ = webviewObj.Set("isHidden", webview.jsIsHidden)
	//_ = webviewObj.Set("setPosition", webview.jsSetPosition)

	// Create a new webview object
	channelObj := t.ctx.vm.NewObject()
	_ = channelObj.Set("sync", webview.channel.jsSync)
	_ = channelObj.Set("on", webview.channel.jsOn)
	_ = channelObj.Set("send", webview.channel.jsSend)

	_ = webviewObj.Set("channel", channelObj)

	// Register components
	_ = webviewObj.Set("div", t.componentManager.jsDiv)
	_ = webviewObj.Set("flex", t.componentManager.jsFlex)
	_ = webviewObj.Set("stack", t.componentManager.jsStack)
	_ = webviewObj.Set("text", t.componentManager.jsText)
	_ = webviewObj.Set("button", t.componentManager.jsButton)
	_ = webviewObj.Set("anchor", t.componentManager.jsAnchor)
	_ = webviewObj.Set("input", t.componentManager.jsInput)
	_ = webviewObj.Set("radioGroup", t.componentManager.jsRadioGroup)
	_ = webviewObj.Set("switch", t.componentManager.jsSwitch)
	_ = webviewObj.Set("checkbox", t.componentManager.jsCheckbox)
	_ = webviewObj.Set("select", t.componentManager.jsSelect)
	_ = webviewObj.Set("css", t.componentManager.jsCSS)
	_ = webviewObj.Set("tooltip", t.componentManager.jsTooltip)
	_ = webviewObj.Set("modal", t.componentManager.jsModal)
	_ = webviewObj.Set("dropdownMenu", t.componentManager.jsDropdownMenu)
	_ = webviewObj.Set("dropdownMenuItem", t.componentManager.jsDropdownMenuItem)
	_ = webviewObj.Set("dropdownMenuSeparator", t.componentManager.jsDropdownMenuSeparator)
	_ = webviewObj.Set("dropdownMenuLabel", t.componentManager.jsDropdownMenuLabel)
	_ = webviewObj.Set("popover", t.componentManager.jsPopover)
	_ = webviewObj.Set("a", t.componentManager.jsA)
	_ = webviewObj.Set("p", t.componentManager.jsP)
	_ = webviewObj.Set("alert", t.componentManager.jsAlert)
	_ = webviewObj.Set("tabs", t.componentManager.jsTabs)
	_ = webviewObj.Set("tabsList", t.componentManager.jsTabsList)
	_ = webviewObj.Set("tabsTrigger", t.componentManager.jsTabsTrigger)
	_ = webviewObj.Set("tabsContent", t.componentManager.jsTabsContent)
	_ = webviewObj.Set("badge", t.componentManager.jsBadge)
	_ = webviewObj.Set("span", t.componentManager.jsSpan)
	_ = webviewObj.Set("img", t.componentManager.jsImg)

	// Listen to mount events in order to return the webview object
	listener := t.ctx.RegisterEventListener(ClientWebviewMountedEvent)
	t.ctx.registerOnCleanup(func() {
		t.ctx.UnregisterEventListenerE(listener)
	})
	listener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientWebviewMountedEventPayload
		if event.ParsePayloadAs(ClientWebviewMountedEvent, &payload) && payload.Slot == string(webview.Slot) {
			webview.mounted.Store(true)
			t.renderWebviewIframe(webview.Slot)
		}
	})
	// Listen to mount events in order to return the webview object
	unmountListener := t.ctx.RegisterEventListener(ClientWebviewUnmountedEvent)
	t.ctx.registerOnCleanup(func() {
		t.ctx.UnregisterEventListenerE(unmountListener)
	})
	unmountListener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientWebviewUnmountedEventPayload
		if event.ParsePayloadAs(ClientWebviewUnmountedEvent, &payload) && payload.Slot == string(webview.Slot) {
			webview.mounted.Store(false)
		}
	})

	sidebarListener := t.ctx.RegisterEventListener(ClientWebviewSidebarMountedEvent)
	t.ctx.registerOnCleanup(func() {
		t.ctx.UnregisterEventListenerE(sidebarListener)
	})
	sidebarListener.SetCallback(func(event *ClientPluginEvent) {
		if !webview.options.hasSidebar() {
			return
		}
		var payload ClientWebviewSidebarMountedEventPayload
		if event.ParsePayloadAs(ClientWebviewSidebarMountedEvent, &payload) {
			t.ctx.scheduler.ScheduleAsync(func() error {
				// Return the sidebar object to the client
				t.renderWebviewSidebar()
				return nil
			})
		}
	})

	return webviewObj
}

/////

// jsSetContent registers the HTML content to be rendered in the iframe
// Communication with the Plugin UI context will be done using a bridge
//
//	Example:
//	webview.setContent(() => `<div>Hello World!</div>`)
func (w *Webview) jsSetContent(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		w.webviewManager.ctx.handleTypeError("render requires a function")
		return goja.Undefined()
	}

	// Set the render function
	w.contentFunc = funcRes
	w.renderFunc = nil

	return goja.Undefined()
}

// jsRender registers a function to be called when the webview is rendered/updated
//
//	Example:
//	webview.render(() => webview.stack([]))
func (w *Webview) jsRender(call goja.FunctionCall) goja.Value {

	funcRes, ok := call.Argument(0).Export().(func(goja.FunctionCall) goja.Value)
	if !ok {
		w.webviewManager.ctx.handleTypeError("render requires a function")
		return goja.Undefined()
	}

	// Set the render function
	w.renderFunc = funcRes
	w.contentFunc = nil

	return goja.Undefined()
}

// jsUpdate schedules a re-render on the client
//
//	Example:
//	webview.update()
func (w *Webview) jsUpdate(_ goja.FunctionCall) goja.Value {
	// Update the context's lastUIUpdateAt to prevent duplicate updates
	w.webviewManager.ctx.uiUpdateMu.Lock()
	w.webviewManager.ctx.lastUIUpdateAt = time.Now()
	w.webviewManager.ctx.uiUpdateMu.Unlock()

	//w.webviewManager.renderWebviewScheduled(w.Slot)
	w.webviewManager.renderWebviewIframe(w.Slot)
	return goja.Undefined()
}

// jsSetOptions updates the webview options dynamically
//
//	Example:
//	webview.setOptions({ width: "400px", height: "300px", draggable: true })
func (w *Webview) jsSetOptions(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		w.webviewManager.ctx.handleTypeError("setOptions requires an options object")
	}

	propsObj, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		w.webviewManager.ctx.handleTypeError("setOptions requires an options object")
		return goja.Undefined()
	}

	// Update options
	if propsObj["className"] != nil {
		w.options.ClassName, _ = propsObj["className"].(string)
	}
	if propsObj["style"] != nil {
		w.options.Style, _ = propsObj["style"].(string)
	}
	if propsObj["width"] != nil {
		w.options.Width, _ = propsObj["width"].(string)
	}
	if propsObj["height"] != nil {
		w.options.Height, _ = propsObj["height"].(string)
	}
	if propsObj["maxWidth"] != nil {
		w.options.MaxWidth, _ = propsObj["maxWidth"].(string)
	}
	if propsObj["maxHeight"] != nil {
		w.options.MaxHeight, _ = propsObj["maxHeight"].(string)
	}
	if propsObj["zIndex"] != nil {
		if zIdx, ok := propsObj["zIndex"].(int64); ok {
			w.options.ZIndex = int(zIdx)
		} else if zIdx, ok := propsObj["zIndex"].(float64); ok {
			w.options.ZIndex = int(zIdx)
		}
	}
	if propsObj["window"] != nil {
		if windowObj, ok := propsObj["window"].(map[string]interface{}); ok {
			// Parse window management options
			if windowObj["draggable"] != nil {
				w.options.Window.Draggable, _ = windowObj["draggable"].(bool)
			}
			if windowObj["frameless"] != nil {
				w.options.Window.Frameless, _ = windowObj["frameless"].(bool)
			}
			if windowObj["defaultPosition"] != nil {
				w.options.Window.DefaultPosition, _ = windowObj["defaultPosition"].(string)
			}
			if windowObj["defaultX"] != nil {
				if x, ok := windowObj["defaultX"].(int64); ok {
					w.options.Window.DefaultX = int(x)
				} else if x, ok := windowObj["defaultX"].(float64); ok {
					w.options.Window.DefaultX = int(x)
				}
			}
			if windowObj["defaultY"] != nil {
				if y, ok := windowObj["defaultY"].(int64); ok {
					w.options.Window.DefaultY = int(y)
				} else if y, ok := windowObj["defaultY"].(float64); ok {
					w.options.Window.DefaultY = int(y)
				}
			}
		}
	}
	if propsObj["autoHeight"] != nil {
		w.options.AutoHeight, _ = propsObj["autoHeight"].(bool)
	}
	if propsObj["fullWidth"] != nil {
		w.options.FullWidth, _ = propsObj["fullWidth"].(bool)
	}
	if propsObj["hidden"] != nil {
		w.options.Hidden, _ = propsObj["hidden"].(bool)
	}

	// Send update to client
	w.webviewManager.renderWebviewIframe(w.Slot)
	return goja.Undefined()
}

// jsClose closes/hides the webview
//
//	Example:
//	webview.close()
func (w *Webview) jsClose(_ goja.FunctionCall) goja.Value {
	w.webviewManager.ctx.SendEventToClient(ServerWebviewCloseEvent, ServerWebviewCloseEventPayload{
		WebviewID: w.GetID(),
	})
	return goja.Undefined()
}

// jsShow shows the webview (reverses hide)
//
//	Example:
//	webview.show()
func (w *Webview) jsShow(_ goja.FunctionCall) goja.Value {
	w.options.Hidden = false
	w.webviewManager.renderWebviewIframe(w.Slot)
	return goja.Undefined()
}

// jsHide hides the webview without closing it
//
//	Example:
//	webview.hide()
func (w *Webview) jsHide(_ goja.FunctionCall) goja.Value {
	w.options.Hidden = true
	w.webviewManager.renderWebviewIframe(w.Slot)
	return goja.Undefined()
}

// jsIsHidden returns whether the webview is hidden
//
//	Example:
//	webview.isHidden()
func (w *Webview) jsIsHidden(_ goja.FunctionCall) goja.Value {
	return w.webviewManager.ctx.vm.ToValue(w.options.Hidden)
}

// jsOnMount is called when the webview is mounted and before it is loaded
func (w *Webview) jsOnMount(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		w.webviewManager.ctx.handleTypeError("onMount requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		w.webviewManager.ctx.handleTypeError("onMount requires a callback function")
		return goja.Undefined()
	}

	eventListener := w.webviewManager.ctx.RegisterEventListener(ClientWebviewMountedEvent)
	payload := ClientWebviewMountedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientWebviewMountedEvent, &payload) && payload.Slot == string(w.Slot) {
			w.webviewManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), w.webviewManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					w.webviewManager.ctx.logger.Error().Err(err).Msg("plugin: Error running webview on mount callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsOnLoad is called when the webview is loaded after it's been mounted
func (w *Webview) jsOnLoad(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		w.webviewManager.ctx.handleTypeError("onLoad requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		w.webviewManager.ctx.handleTypeError("onLoad requires a callback function")
		return goja.Undefined()
	}

	eventListener := w.webviewManager.ctx.RegisterEventListener(ClientWebviewLoadedEvent)
	payload := ClientWebviewLoadedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientWebviewLoadedEvent, &payload) && payload.Slot == string(w.Slot) {
			w.webviewManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), w.webviewManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					w.webviewManager.ctx.logger.Error().Err(err).Msg("plugin: Error running webview on load callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsOnUnmount is called after the webview has been unmounted
func (w *Webview) jsOnUnmount(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		w.webviewManager.ctx.handleTypeError("onUnmount requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		w.webviewManager.ctx.handleTypeError("onUnmount requires a callback function")
		return goja.Undefined()
	}

	eventListener := w.webviewManager.ctx.RegisterEventListener(ClientWebviewUnmountedEvent)
	payload := ClientWebviewUnmountedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientWebviewUnmountedEvent, &payload) && payload.Slot == string(w.Slot) {
			w.webviewManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), w.webviewManager.ctx.vm.ToValue(map[string]interface{}{}))
				if err != nil {
					w.webviewManager.ctx.logger.Error().Err(err).Msg("plugin: Error running webview on unmount callback")
				}
				return err
			})
		}
	})

	return goja.Undefined()
}
func (w *Webview) jsGetScreenPath(_ goja.FunctionCall) goja.Value {
	return w.webviewManager.ctx.vm.ToValue("/webview?id=" + w.GetID())
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// WebviewChannel
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// jsSync syncs a state with the webview iframe
// When the state updates, the new value is sent to the iframe
//
//	Example:
//	const count = ctx.state(0)
//	webview.channel.sync("count", count)
func (c *WebviewChannel) jsSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("sync requires a key and a state")
	}

	key, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("sync: first argument must be a string key")
		return goja.Undefined()
	}

	stateObj := call.Argument(1).ToObject(c.webview.webviewManager.ctx.vm)
	if stateObj == nil {
		c.webview.webviewManager.ctx.handleTypeError("sync: second argument must be a state object")
		return goja.Undefined()
	}

	stateIDVal := stateObj.Get("__stateId")
	if stateIDVal == nil {
		c.webview.webviewManager.ctx.handleTypeError("sync: state object must have an id")
		return goja.Undefined()
	}

	stateID := stateIDVal.String()

	// Store the mapping
	c.syncedStates.Set(key, stateID)

	// Subscribe to state changes
	state, ok := c.webview.webviewManager.ctx.states.Get(stateID)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("sync: state not found")
		return goja.Undefined()
	}

	// Send initial value
	c.sendStateToWebview(key, state.Value.Export())

	// Listen for state changes and sync to webview
	stateCh := c.webview.webviewManager.ctx.subscribeStateUpdates()

	// Start a goroutine to listen for this specific state's updates
	go func() {
		for newState := range stateCh {
			if newState.ID == stateID {
				c.sendStateToWebview(key, newState.Value.Export())
			}
		}
	}()

	eventListener := c.webview.webviewManager.ctx.RegisterEventListener(ClientWebviewLoadedEvent)
	payload := ClientWebviewLoadedEventPayload{}

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		if event.ParsePayloadAs(ClientWebviewLoadedEvent, &payload) && payload.Slot == string(c.webview.Slot) {
			state, ok := c.webview.webviewManager.ctx.states.Get(stateID)
			if !ok {
				c.webview.webviewManager.ctx.handleTypeError("sync: state not found")
				return
			}
			c.sendStateToWebview(key, state.Value.Export())
			// send the value again just in case
			go func() {
				state, ok := c.webview.webviewManager.ctx.states.Get(stateID)
				if !ok {
					c.webview.webviewManager.ctx.handleTypeError("sync: state not found")
					return
				}
				time.Sleep(1000 * time.Millisecond)
				c.sendStateToWebview(key, state.Value.Export())
			}()
		}
	})

	return goja.Undefined()
}

// jsOn registers an event handler for messages from the webview
// This is called from the server-side to listen to events triggered by webview.trigger()
//
//	Example:
//	webview.channel.on("customEvent", (data) => {
//		console.log(data)
//	})
func (c *WebviewChannel) jsOn(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("on requires an event name and a callback")
	}

	eventName, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("on: first argument must be a string event name")
		return goja.Undefined()
	}

	callback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("on: second argument must be a callback function")
		return goja.Undefined()
	}

	// Register event handler to listen for messages from the webview
	eventListener := c.webview.webviewManager.ctx.RegisterEventListener(ClientWebviewPostMessageEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		// Parse the payload
		var payload ClientWebviewPostMessageEventPayload
		if event.ParsePayloadAs(ClientWebviewPostMessageEvent, &payload) && payload.Slot == string(c.webview.Slot) && payload.EventName == eventName {
			c.webview.webviewManager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), c.webview.webviewManager.ctx.vm.ToValue(payload.Event))
				if err != nil {
					c.webview.webviewManager.ctx.logger.Error().Err(err).Msgf("plugin: Error running webview channel.on callback for event %s", eventName)
				}
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsSend sends a message to the webview iframe
// This is used to send arbitrary data from the server to the iframe
//
//	Example:
//	webview.channel.send("messageType", { data: "hello" })
func (c *WebviewChannel) jsSend(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.webview.webviewManager.ctx.handleTypeError("send requires a key and a value")
	}

	key, ok := call.Argument(0).Export().(string)
	if !ok {
		c.webview.webviewManager.ctx.handleTypeError("send: first argument must be a string key")
		return goja.Undefined()
	}

	value := call.Argument(1).Export()

	c.sendStateToWebview(key, value)

	return goja.Undefined()
}

// sendStateToWebview sends a state value to the webview iframe
func (c *WebviewChannel) sendStateToWebview(key string, value interface{}) {
	webviewId := c.webview.GetID()

	// Security: Iframe won't receive anilist token
	if str, ok := value.(string); ok {
		strings.ReplaceAll(str, c.webview.webviewManager.ctx.anilistToken, "[TOKEN]")
	} else {
		encoded, err := json.Marshal(value)
		if err == nil {
			if strings.Contains(string(encoded), c.webview.webviewManager.ctx.anilistToken) {
				return
			}
		}
	}

	// Get the token from the iframe (we'll need to update the iframe creation to store this)
	// For now, we'll send it without token verification on the receive side
	c.webview.webviewManager.ctx.SendEventToClient(ServerWebviewSyncStateEvent, ServerWebviewSyncStateEventPayload{
		WebviewID: webviewId,
		Key:       key,
		Value:     value,
		Token:     "", // Will be populated by the client-side handler
	})
}

func (w *Webview) GetID() string {
	return w.webviewManager.ctx.ext.ID + "-" + string(w.Slot)
}
