package plugin_ui

import (
	"fmt"
	"seanime/internal/util/result"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

const (
	MaxActionsPerType = 3 // A plugin can only at most X actions of a certain type
)

// ActionManager
//
// Actions are buttons, dropdown items, and context menu items that are displayed in certain places in the UI.
// They are defined in the plugin code and are used to trigger events.
//
// The ActionManager is responsible for registering, rendering, and handling events for actions.
type ActionManager struct {
	ctx *Context

	animePageButtons          *result.Map[string, *AnimePageButton]
	animePageDropdownItems    *result.Map[string, *AnimePageDropdownMenuItem]
	animeLibraryDropdownItems *result.Map[string, *AnimeLibraryDropdownMenuItem]
	mangaPageButtons          *result.Map[string, *MangaPageButton]
	mediaCardContextMenuItems *result.Map[string, *MediaCardContextMenuItem]
}

type BaseActionProps struct {
	ID    string            `json:"id"`
	Label string            `json:"label"`
	Style map[string]string `json:"style,omitempty"`
}

// Base action struct that all action types embed
type BaseAction struct {
	BaseActionProps
}

// GetProps returns the base action properties
func (a *BaseAction) GetProps() BaseActionProps {
	return a.BaseActionProps
}

// SetProps sets the base action properties
func (a *BaseAction) SetProps(props BaseActionProps) {
	a.BaseActionProps = props
}

// UnmountAll unmounts all actions
// It should be called
func (a *ActionManager) UnmountAll() {

	if a.animePageButtons.ClearN() > 0 {
		a.renderAnimePageButtons()
	}
	if a.animePageDropdownItems.ClearN() > 0 {
		a.renderAnimePageDropdownItems()
	}
	if a.animeLibraryDropdownItems.ClearN() > 0 {
		a.renderAnimeLibraryDropdownItems()
	}
	if a.mangaPageButtons.ClearN() > 0 {
		a.renderMangaPageButtons()
	}
	if a.mediaCardContextMenuItems.ClearN() > 0 {
		a.renderMediaCardContextMenuItems()
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimePageButton struct {
	BaseAction
	Intent string `json:"intent,omitempty"`
}

func (a *AnimePageButton) CreateObject(actionManager *ActionManager) *goja.Object {
	obj := actionManager.ctx.vm.NewObject()
	actionManager.bindSharedToObject(obj, a)

	_ = obj.Set("setIntent", func(intent string) {
		a.Intent = intent
	})

	return obj
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MangaPageButton struct {
	BaseAction
	Intent string `json:"intent,omitempty"`
}

func (a *MangaPageButton) CreateObject(actionManager *ActionManager) *goja.Object {
	obj := actionManager.ctx.vm.NewObject()
	actionManager.bindSharedToObject(obj, a)

	_ = obj.Set("setIntent", func(intent string) {
		a.Intent = intent
	})

	return obj
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimePageDropdownMenuItem struct {
	BaseAction
}

func (a *AnimePageDropdownMenuItem) CreateObject(actionManager *ActionManager) *goja.Object {
	obj := actionManager.ctx.vm.NewObject()
	actionManager.bindSharedToObject(obj, a)
	return obj
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimeLibraryDropdownMenuItem struct {
	BaseAction
}

func (a *AnimeLibraryDropdownMenuItem) CreateObject(actionManager *ActionManager) *goja.Object {
	obj := actionManager.ctx.vm.NewObject()
	actionManager.bindSharedToObject(obj, a)
	return obj
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaCardContextMenuItemFor string

const (
	MediaCardContextMenuItemForAnime MediaCardContextMenuItemFor = "anime"
	MediaCardContextMenuItemForManga MediaCardContextMenuItemFor = "manga"
	MediaCardContextMenuItemForBoth  MediaCardContextMenuItemFor = "both"
)

type MediaCardContextMenuItem struct {
	BaseAction
	For MediaCardContextMenuItemFor `json:"for"` // anime, manga, both
}

func (a *MediaCardContextMenuItem) CreateObject(actionManager *ActionManager) *goja.Object {
	obj := actionManager.ctx.vm.NewObject()
	actionManager.bindSharedToObject(obj, a)

	_ = obj.Set("setFor", func(_for MediaCardContextMenuItemFor) {
		a.For = _for
	})

	return obj
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewActionManager(ctx *Context) *ActionManager {
	return &ActionManager{
		ctx: ctx,

		animePageButtons:          result.NewResultMap[string, *AnimePageButton](),
		animeLibraryDropdownItems: result.NewResultMap[string, *AnimeLibraryDropdownMenuItem](),
		animePageDropdownItems:    result.NewResultMap[string, *AnimePageDropdownMenuItem](),
		mangaPageButtons:          result.NewResultMap[string, *MangaPageButton](),
		mediaCardContextMenuItems: result.NewResultMap[string, *MediaCardContextMenuItem](),
	}
}

// renderAnimePageButtons is called when the client requests the buttons to display on the anime page.
func (a *ActionManager) renderAnimePageButtons() {
	buttons := make([]*AnimePageButton, 0)
	a.animePageButtons.Range(func(key string, value *AnimePageButton) bool {
		buttons = append(buttons, value)
		return true
	})

	a.ctx.SendEventToClient(ServerActionRenderAnimePageButtonsEvent, ServerActionRenderAnimePageButtonsEventPayload{
		Buttons: buttons,
	})
}

func (a *ActionManager) renderAnimePageDropdownItems() {
	items := make([]*AnimePageDropdownMenuItem, 0)
	a.animePageDropdownItems.Range(func(key string, value *AnimePageDropdownMenuItem) bool {
		items = append(items, value)
		return true
	})

	a.ctx.SendEventToClient(ServerActionRenderAnimePageDropdownItemsEvent, ServerActionRenderAnimePageDropdownItemsEventPayload{
		Items: items,
	})
}

func (a *ActionManager) renderAnimeLibraryDropdownItems() {
	items := make([]*AnimeLibraryDropdownMenuItem, 0)
	a.animeLibraryDropdownItems.Range(func(key string, value *AnimeLibraryDropdownMenuItem) bool {
		items = append(items, value)
		return true
	})

	a.ctx.SendEventToClient(ServerActionRenderAnimeLibraryDropdownItemsEvent, ServerActionRenderAnimeLibraryDropdownItemsEventPayload{
		Items: items,
	})
}

func (a *ActionManager) renderMangaPageButtons() {
	buttons := make([]*MangaPageButton, 0)
	a.mangaPageButtons.Range(func(key string, value *MangaPageButton) bool {
		buttons = append(buttons, value)
		return true
	})

	a.ctx.SendEventToClient(ServerActionRenderMangaPageButtonsEvent, ServerActionRenderMangaPageButtonsEventPayload{
		Buttons: buttons,
	})
}

func (a *ActionManager) renderMediaCardContextMenuItems() {
	items := make([]*MediaCardContextMenuItem, 0)
	a.mediaCardContextMenuItems.Range(func(key string, value *MediaCardContextMenuItem) bool {
		items = append(items, value)
		return true
	})

	a.ctx.SendEventToClient(ServerActionRenderMediaCardContextMenuItemsEvent, ServerActionRenderMediaCardContextMenuItemsEventPayload{
		Items: items,
	})
}

// bind binds 'action' to the ctx object
//
//	Example:
//	ctx.action.newAnimePageButton(...)
func (a *ActionManager) bind(ctxObj *goja.Object) {
	actionObj := a.ctx.vm.NewObject()
	_ = actionObj.Set("newAnimePageButton", a.jsNewAnimePageButton)
	_ = actionObj.Set("newAnimePageDropdownItem", a.jsNewAnimePageDropdownItem)
	_ = actionObj.Set("newAnimeLibraryDropdownItem", a.jsNewAnimeLibraryDropdownItem)
	_ = actionObj.Set("newMediaCardContextMenuItem", a.jsNewMediaCardContextMenuItem)
	_ = actionObj.Set("newMangaPageButton", a.jsNewMangaPageButton)
	_ = ctxObj.Set("action", actionObj)
}

////////////////////////////////////////////////////////////////////////////////////////////////
// Actions
////////////////////////////////////////////////////////////////////////////////////////////////

// jsNewAnimePageButton
//
//	Example:
//	const downloadButton = ctx.newAnimePageButton({
//		label: "Download",
//		intent: "primary",
//		onClick: "download-button-clicked",
//	})
func (a *ActionManager) jsNewAnimePageButton(call goja.FunctionCall) goja.Value {
	// Create a new action
	action := &AnimePageButton{}

	// Get the props
	a.unmarshalProps(call, action)
	action.ID = uuid.New().String()

	// Create the object
	obj := action.CreateObject(a)
	return obj
}

// jsNewAnimePageDropdownItem
//
//	Example:
//	const downloadButton = ctx.newAnimePageDropdownItem({
//		label: "Download",
//		onClick: "download-button-clicked",
//	})
func (a *ActionManager) jsNewAnimePageDropdownItem(call goja.FunctionCall) goja.Value {
	// Create a new action
	action := &AnimePageDropdownMenuItem{}

	// Get the props
	a.unmarshalProps(call, action)
	action.ID = uuid.New().String()

	// Create the object
	obj := action.CreateObject(a)
	return obj
}

// jsNewAnimeLibraryDropdownItem
//
//	Example:
//	const downloadButton = ctx.newAnimeLibraryDropdownItem({
//		label: "Download",
//		onClick: "download-button-clicked",
//	})
func (a *ActionManager) jsNewAnimeLibraryDropdownItem(call goja.FunctionCall) goja.Value {
	// Create a new action
	action := &AnimeLibraryDropdownMenuItem{}

	// Get the props
	a.unmarshalProps(call, action)
	action.ID = uuid.New().String()

	// Create the object
	obj := action.CreateObject(a)
	return obj
}

// jsNewMediaCardContextMenuItem
//
//	Example:
//	const downloadButton = ctx.newMediaCardContextMenuItem({
//		label: "Download",
//		onClick: "download-button-clicked",
//	})
func (a *ActionManager) jsNewMediaCardContextMenuItem(call goja.FunctionCall) goja.Value {
	// Create a new action
	action := &MediaCardContextMenuItem{}

	// Get the props
	a.unmarshalProps(call, action)
	action.ID = uuid.New().String()

	// Create the object
	obj := action.CreateObject(a)
	return obj
}

// jsNewMangaPageButton
//
//	Example:
//	const downloadButton = ctx.newMangaPageButton({
//		label: "Download",
//		onClick: "download-button-clicked",
//	})
func (a *ActionManager) jsNewMangaPageButton(call goja.FunctionCall) goja.Value {
	// Create a new action
	action := &MangaPageButton{}

	// Get the props
	a.unmarshalProps(call, action)
	action.ID = uuid.New().String()

	// Create the object
	obj := action.CreateObject(a)
	return obj
}

// ///////////////////////////////////////////////////////////////////////////////////
// Shared
// ///////////////////////////////////////////////////////////////////////////////////
// bindSharedToObject binds shared methods to action objects
//
//	Example:
//	const downloadButton = ctx.newAnimePageButton(...)
//	downloadButton.mount()
//	downloadButton.unmount()
//	downloadButton.setLabel("Downloading...")
func (a *ActionManager) bindSharedToObject(obj *goja.Object, action interface{}) {
	var id string
	var props BaseActionProps
	var mapToUse interface{}

	switch act := action.(type) {
	case *AnimePageButton:
		id = act.ID
		props = act.GetProps()
		mapToUse = a.animePageButtons
	case *MangaPageButton:
		id = act.ID
		props = act.GetProps()
		mapToUse = a.mangaPageButtons
	case *AnimePageDropdownMenuItem:
		id = act.ID
		props = act.GetProps()
		mapToUse = a.animePageDropdownItems
	case *AnimeLibraryDropdownMenuItem:
		id = act.ID
		props = act.GetProps()
		mapToUse = a.animeLibraryDropdownItems
	case *MediaCardContextMenuItem:
		id = act.ID
		props = act.GetProps()
		mapToUse = a.mediaCardContextMenuItems
	}

	_ = obj.Set("mount", func() {
		switch m := mapToUse.(type) {
		case *result.Map[string, *AnimePageButton]:
			if btn, ok := action.(*AnimePageButton); ok {
				m.Set(id, btn)
			}
		case *result.Map[string, *MangaPageButton]:
			if btn, ok := action.(*MangaPageButton); ok {
				m.Set(id, btn)
			}
		case *result.Map[string, *AnimePageDropdownMenuItem]:
			if item, ok := action.(*AnimePageDropdownMenuItem); ok {
				m.Set(id, item)
			}
		case *result.Map[string, *AnimeLibraryDropdownMenuItem]:
			if item, ok := action.(*AnimeLibraryDropdownMenuItem); ok {
				m.Set(id, item)
			}
		case *result.Map[string, *MediaCardContextMenuItem]:
			if item, ok := action.(*MediaCardContextMenuItem); ok {
				if item.For == "" {
					item.For = MediaCardContextMenuItemForBoth
				}
				m.Set(id, item)
			}
		}
		a.renderAnimePageButtons()
	})

	_ = obj.Set("unmount", func() {
		switch m := mapToUse.(type) {
		case *result.Map[string, *AnimePageButton]:
			m.Delete(id)
		case *result.Map[string, *MangaPageButton]:
			m.Delete(id)
		case *result.Map[string, *AnimePageDropdownMenuItem]:
			m.Delete(id)
		case *result.Map[string, *AnimeLibraryDropdownMenuItem]:
			m.Delete(id)
		case *result.Map[string, *MediaCardContextMenuItem]:
			m.Delete(id)
		}
		a.renderAnimePageButtons()
	})

	_ = obj.Set("setLabel", func(label string) {
		newProps := props
		newProps.Label = label

		switch act := action.(type) {
		case *AnimePageButton:
			act.SetProps(newProps)
		case *MangaPageButton:
			act.SetProps(newProps)
		case *AnimePageDropdownMenuItem:
			act.SetProps(newProps)
		case *AnimeLibraryDropdownMenuItem:
			act.SetProps(newProps)
		case *MediaCardContextMenuItem:
			act.SetProps(newProps)
		}

		a.renderAnimePageButtons()
	})

	_ = obj.Set("setStyle", func(style map[string]string) {
		newProps := props
		newProps.Style = style

		switch act := action.(type) {
		case *AnimePageButton:
			act.SetProps(newProps)
		case *MangaPageButton:
			act.SetProps(newProps)
		case *AnimePageDropdownMenuItem:
			act.SetProps(newProps)
		case *AnimeLibraryDropdownMenuItem:
			act.SetProps(newProps)
		case *MediaCardContextMenuItem:
			act.SetProps(newProps)
		}

		a.renderAnimePageButtons()
	})

	_ = obj.Set("onClick", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			a.ctx.handleTypeError("onClick requires a callback function")
		}

		callback, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			a.ctx.handleTypeError("onClick requires a callback function")
		}

		eventListener := a.ctx.RegisterEventListener(ClientActionClickedEvent)

		eventListener.SetCallback(func(event *ClientPluginEvent) {
			payload := ClientActionClickedEventPayload{}
			if event.ParsePayloadAs(ClientActionClickedEvent, &payload) && payload.ActionID == id {
				a.ctx.scheduler.ScheduleAsync(func() error {
					_, err := callback(goja.Undefined(), a.ctx.vm.ToValue(payload.Event))
					return err
				})
			}
		})

		// go func() {
		// 	for event := range eventListener.Channel {
		// 		if event.ParsePayloadAs(ClientActionClickedEvent, &payload) && payload.ActionID == id {
		// 			a.ctx.scheduler.ScheduleAsync(func() error {
		// 				_, err := callback(goja.Undefined(), a.ctx.vm.ToValue(payload.Event))
		// 				if err != nil {
		// 					a.ctx.logger.Error().Err(err).Msg("plugin: Error running action click callback")
		// 				}
		// 				return err
		// 			})
		// 		}
		// 	}
		// }()

		return goja.Undefined()
	})
}

/////////////////////////////////////////////////////////////////////////////////////
// Utils
/////////////////////////////////////////////////////////////////////////////////////

func (a *ActionManager) unmarshalProps(call goja.FunctionCall, ret interface{}) {
	if len(call.Arguments) < 1 {
		a.ctx.handleException(fmt.Errorf("expected 1 argument"))
	}

	props := call.Arguments[0].Export()
	if props == nil {
		a.ctx.handleException(fmt.Errorf("expected props object"))
	}

	marshaled, err := json.Marshal(props)
	if err != nil {
		a.ctx.handleException(err)
	}

	err = json.Unmarshal(marshaled, ret)
	if err != nil {
		a.ctx.handleException(err)
	}
}
