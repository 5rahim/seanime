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
func (s *ScreenManager) bind(vm *goja.Runtime, ctxObj *goja.Object) {
	screenObj := vm.NewObject()
	_ = screenObj.Set("onNavigate", s.jsOnNavigate)
	_ = screenObj.Set("navigateTo", s.jsNavigateTo)

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

// jsOnNavigate registers a callback to be called when the current screen changes
//
//	Example:
//	const onNavigate = (event) => {
//		console.log(event.screen);
//	};
//	ctx.screen.onNavigate(onNavigate);
func (s *ScreenManager) jsOnNavigate(callback goja.Callable) {
	eventListener := s.ctx.RegisterEventListener()
	var payload ClientScreenChangedEventPayload

	go func(payload ClientScreenChangedEventPayload) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientScreenChangedEvent, &payload) {
				s.ctx.scheduler.ScheduleAsync(func() error {
					_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(payload))
					if err != nil {
						s.ctx.logger.Error().Err(err).Msg("error running screen navigation callback")
					}
					return err
				})
			}
		}
	}(payload)
}
