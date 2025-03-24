package plugin_ui

import (
	"strings"
	"sync"

	"github.com/dop251/goja"
)

type ScreenManager struct {
	ctx *Context
	mu  sync.RWMutex
}

func NewScreenManager(ctx *Context) *ScreenManager {
	return &ScreenManager{
		ctx: ctx,
	}
}

// bind binds 'screen' to the ctx object
//
//	Example:
//	ctx.screen.navigateTo("/entry?id=21");
func (s *ScreenManager) bind(ctxObj *goja.Object) {
	screenObj := s.ctx.vm.NewObject()
	_ = screenObj.Set("onNavigate", s.jsOnNavigate)
	_ = screenObj.Set("navigateTo", s.jsNavigateTo)
	_ = screenObj.Set("reload", s.jsReload)
	_ = screenObj.Set("loadCurrent", s.jsLoadCurrent)

	_ = ctxObj.Set("screen", screenObj)
}

// jsNavigateTo navigates to a new screen
//
//	Example:
//	ctx.screen.navigateTo("/entry?id=21");
func (s *ScreenManager) jsNavigateTo(path string) {
	if !strings.HasPrefix(path, "/") {
		return
	}

	s.ctx.SendEventToClient(ServerScreenNavigateToEvent, ServerScreenNavigateToEventPayload{
		Path: path,
	})
}

// jsReload reloads the current screen
func (s *ScreenManager) jsReload() {
	s.ctx.SendEventToClient(ServerScreenReloadEvent, ServerScreenReloadEventPayload{})
}

// jsLoadCurrent calls onNavigate with the current screen data
func (s *ScreenManager) jsLoadCurrent() {
	s.ctx.SendEventToClient(ServerScreenGetCurrentEvent, ServerScreenGetCurrentEventPayload{})
}

// jsOnNavigate registers a callback to be called when the current screen changes
//
//	Example:
//	const onNavigate = (event) => {
//		console.log(event.screen);
//	};
//	ctx.screen.onNavigate(onNavigate);
func (s *ScreenManager) jsOnNavigate(callback goja.Callable) {
	eventListener := s.ctx.RegisterEventListener(ClientScreenChangedEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientScreenChangedEventPayload
		if event.ParsePayloadAs(ClientScreenChangedEvent, &payload) {
			s.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(payload))
				return err
			})
		}
	})

	// go func(payload ClientScreenChangedEventPayload) {
	// 	for event := range eventListener.Channel {
	// 		if event.ParsePayloadAs(ClientScreenChangedEvent, &payload) {
	// 			s.ctx.scheduler.ScheduleAsync(func() error {
	// 				_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(payload))
	// 				if err != nil {
	// 					s.ctx.logger.Error().Err(err).Msg("error running screen navigation callback")
	// 				}
	// 				return err
	// 			})
	// 		}
	// 	}
	// }(payload)
}
